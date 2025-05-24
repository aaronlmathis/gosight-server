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

// AzureAuth implements the AuthProvider interface for Azure AD SSO
type AzureAuth struct {
	OAuthConfig *oauth2.Config
	Store       userstore.UserStore
}

// StartLogin initiates the Azure AD OAuth2 flow
func (a *AzureAuth) StartLogin(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/"
	}

	state := base64.URLEncoding.EncodeToString([]byte(next))
	url := a.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

// HandleCallback processes the Azure AD OAuth2 callback
func (a *AzureAuth) HandleCallback(w http.ResponseWriter, r *http.Request) (*usermodel.User, error) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")

	token, err := a.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := a.OAuthConfig.Client(ctx, token)
	// Microsoft Graph API endpoint for user profile
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"userPrincipalName"` // Azure AD uses userPrincipalName for email
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
		if user.SSOID != userInfo.ID || user.SSOProvider != "azure" {
			utils.Warn("SSO mismatch for %s: expected %s/%s, got %s/azure",
				user.Email, user.SSOProvider, user.SSOID, userInfo.ID)
			return nil, ErrUnauthorized
		}
	} else {
		// First-time login — store SSO info
		user.SSOID = userInfo.ID
		user.SSOProvider = "azure"
		utils.Info("First-time SSO link: %s → %s/%s", user.Email, user.SSOProvider, user.SSOID)
	}

	// Always update last_login
	user.LastLogin = time.Now()
	if err := a.Store.SaveUser(ctx, user); err != nil {
		utils.Warn("⚠Failed to update user login metadata: %v", err)
	}

	// Authorized user
	return user, nil
}
