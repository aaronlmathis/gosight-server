package gosightauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/store/userstore"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/aaronlmathis/gosight-shared/utils"
	"golang.org/x/oauth2"
)

// AWS Cognito SSO handler

// AWSAuth implements the AuthProvider interface for AWS Cognito SSO
type AWSAuth struct {
	OAuthConfig *oauth2.Config
	Store       userstore.UserStore
}

// StartLogin initiates the AWS Cognito OAuth2 flow
func (a *AWSAuth) StartLogin(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/"
	}

	state := base64.URLEncoding.EncodeToString([]byte(next))
	url := a.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

// HandleCallback processes the AWS Cognito OAuth2 callback
func (a *AWSAuth) HandleCallback(w http.ResponseWriter, r *http.Request) (*usermodel.User, error) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")

	token, err := a.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := a.OAuthConfig.Client(ctx, token)

	// Build the correct userinfo endpoint from the config
	// Extract the UserPoolID and Region from the OAuth config
	userInfoURL := ""
	if authURL := a.OAuthConfig.Endpoint.AuthURL; authURL != "" {
		// Parse the auth URL to extract the domain
		// Format: https://{userPoolDomain}.auth.{region}.amazoncognito.com/oauth2/authorize
		start := strings.Index(authURL, "://") + 3
		end := strings.Index(authURL[start:], "/")
		if end > 0 {
			domain := authURL[start : start+end]
			userInfoURL = "https://" + domain + "/oauth2/userInfo"
		}
	}

	if userInfoURL == "" {
		return nil, errors.New("unable to construct AWS Cognito userinfo endpoint")
	}

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		Sub           string `json:"sub"`         // AWS Cognito User ID
		Email         string `json:"email"`       // User's email
		Name          string `json:"name"`        // Full name
		GivenName     string `json:"given_name"`  // First name
		FamilyName    string `json:"family_name"` // Last name
		PreferredName string `json:"preferred_username"`
		Picture       string `json:"picture"` // Profile picture URL
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	user, err := a.Store.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, ErrUnauthorized
	}

	// If user has SSO info, verify match
	if user.SSOID != "" {
		if user.SSOID != userInfo.Sub || user.SSOProvider != "aws" {
			utils.Warn("SSO mismatch for %s: expected %s/%s, got %s/aws",
				user.Email, user.SSOProvider, user.SSOID, userInfo.Sub)
			return nil, ErrUnauthorized
		}
	} else {
		// First-time login — store SSO info
		user.SSOID = userInfo.Sub
		user.SSOProvider = "aws"
		utils.Info("First-time SSO link: %s → %s/%s", user.Email, user.SSOProvider, user.SSOID)
	}

	// Update or create user profile with SSO data
	profile, err := a.Store.GetUserProfile(ctx, user.ID)
	if err != nil {
		utils.Debug("Failed to get user profile for %s: %v", user.ID, err)
		// Create new profile if none exists
		profile = &usermodel.UserProfile{
			UserID: user.ID,
		}
	}

	// Update profile with AWS data if not already set
	updated := false
	if profile.FullName == "" {
		if userInfo.Name != "" {
			profile.FullName = userInfo.Name
			updated = true
		} else if userInfo.GivenName != "" || userInfo.FamilyName != "" {
			profile.FullName = strings.TrimSpace(userInfo.GivenName + " " + userInfo.FamilyName)
			updated = true
		}
	}
	if profile.AvatarURL == "" && userInfo.Picture != "" {
		profile.AvatarURL = userInfo.Picture
		updated = true
	}

	// Save profile if updated
	if updated {
		err = a.Store.CreateUserProfile(ctx, profile) // Uses UPSERT
		if err != nil {
			utils.Warn("Failed to update user profile for %s: %v", user.ID, err)
		} else {
			utils.Info("Updated profile for user %s with AWS SSO data", user.Email)
		}
	}

	// Always update last_login
	user.LastLogin = time.Now()
	if err := a.Store.SaveUser(ctx, user); err != nil {
		utils.Warn("⚠Failed to update user login metadata: %v", err)
	}

	// Authorized user
	return user, nil
}
