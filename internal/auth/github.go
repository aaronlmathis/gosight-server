package gosightauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/store/userstore"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/aaronlmathis/gosight-shared/utils"
	"golang.org/x/oauth2"
)

// GitHubAuth implements the AuthProvider interface for GitHub OAuth
type GitHubAuth struct {
	OAuthConfig *oauth2.Config
	Store       userstore.UserStore
}

// StartLogin initiates the GitHub OAuth flow
func (g *GitHubAuth) StartLogin(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/"
	}

	state := base64.URLEncoding.EncodeToString([]byte(next))
	url := g.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

// HandleCallback processes the GitHub OAuth callback
func (g *GitHubAuth) HandleCallback(w http.ResponseWriter, r *http.Request) (*usermodel.User, error) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")

	token, err := g.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := g.OAuthConfig.Client(ctx, token)

	// First get the user profile
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var githubUser struct {
		ID        int64  `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Login     string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, err
	}

	// If email is not public, fetch primary email from emails endpoint
	if githubUser.Email == "" {
		req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
			return nil, err
		}

		// Find primary email
		for _, email := range emails {
			if email.Primary && email.Verified {
				githubUser.Email = email.Email
				break
			}
		}
	}

	if githubUser.Email == "" {
		return nil, ErrUnauthorized
	}

	user, err := g.Store.GetUserByEmail(ctx, githubUser.Email)
	if err != nil {
		return nil, ErrUnauthorized
	}

	// If user has SSO info, verify match
	if user.SSOID != "" {
		if user.SSOID != string(githubUser.ID) || user.SSOProvider != "github" {
			utils.Warn("SSO mismatch for %s: expected %s/%s, got %s/github",
				user.Email, user.SSOProvider, user.SSOID, githubUser.ID)
			return nil, ErrUnauthorized
		}
	} else {
		// First-time login — store SSO info
		user.SSOID = string(githubUser.ID)
		user.SSOProvider = "github"
		utils.Info("First-time SSO link: %s → %s/%s", user.Email, user.SSOProvider, user.SSOID)
	}

	// Update or create user profile with SSO data
	profile, err := g.Store.GetUserProfile(ctx, user.ID)
	if err != nil {
		utils.Debug("Failed to get user profile for %s: %v", user.ID, err)
		// Create new profile if none exists
		profile = &usermodel.UserProfile{
			UserID: user.ID,
		}
	}

	// Update profile with GitHub data if not already set
	updated := false
	if profile.FullName == "" && githubUser.Name != "" {
		profile.FullName = githubUser.Name
		updated = true
	}
	if profile.AvatarURL == "" && githubUser.AvatarURL != "" {
		profile.AvatarURL = githubUser.AvatarURL
		updated = true
	}

	// Save profile if updated
	if updated {
		err = g.Store.CreateUserProfile(ctx, profile) // Uses UPSERT
		if err != nil {
			utils.Warn("Failed to update user profile for %s: %v", user.ID, err)
		} else {
			utils.Info("Updated profile for user %s with GitHub SSO data", user.Email)
		}
	}

	// Always update last_login
	user.LastLogin = time.Now()
	if err := g.Store.SaveUser(ctx, user); err != nil {
		utils.Warn("Failed to update user login metadata: %v", err)
	}

	// Authorized user
	return user, nil
}
