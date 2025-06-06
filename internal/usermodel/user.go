package usermodel

import (
	"encoding/json"
	"time"
)

type User struct {
	ID            string
	Username      string
	Email         string
	PasswordHash  string
	FirstName     string
	LastName      string
	MFAEnabled    bool
	MFAMethod     string // "totp", "webauthn"
	TOTPSecret    string
	WebAuthnCreds []byte
	SSOProvider   string
	SSOID         string
	Roles         []Role
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLogin     time.Time
	Scopes        map[string][]string
}
type SafeUser struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// UserProfile represents the user_profiles table
type UserProfile struct {
	UserID    string    `json:"user_id"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	AvatarURL string    `json:"avatar_url"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserSetting represents a single setting from the user_settings table
type UserSetting struct {
	UserID string          `json:"user_id"`
	Key    string          `json:"key"`
	Value  json.RawMessage `json:"value"`
}

// UserSettings represents a collection of user settings
type UserSettings map[string]json.RawMessage

// ProfileUpdateRequest represents the request payload for profile updates
type ProfileUpdateRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

// PasswordChangeRequest represents the request payload for password changes
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

// UserPreferencesRequest represents the request payload for user preferences
type UserPreferencesRequest struct {
	Theme         string                  `json:"theme"`
	Notifications NotificationPreferences `json:"notifications"`
	Dashboard     map[string]interface{}  `json:"dashboard"`
}

// NotificationPreferences represents detailed notification settings
type NotificationPreferences struct {
	EmailAlerts    bool   `json:"email_alerts"`
	PushAlerts     bool   `json:"push_alerts"`
	AlertFrequency string `json:"alert_frequency"`
}

// CompleteUser represents a user with their profile and settings
type CompleteUser struct {
	User     *User        `json:"user"`
	Profile  *UserProfile `json:"profile"`
	Settings UserSettings `json:"settings"`
}

type Role struct {
	ID          string
	Name        string
	Description string
	Permissions []Permission
}

type Permission struct {
	ID          string
	Name        string
	Description string
}

// CreateUserRequest represents the request payload for creating a new user
type CreateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// AdminPasswordChangeRequest represents the request payload for admin password changes
type AdminPasswordChangeRequest struct {
	NewPassword string `json:"new_password"`
}
