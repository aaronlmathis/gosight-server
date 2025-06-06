/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

// Package handlers provides HTTP handlers for the GoSight API server.
package handlers

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/contextutil"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

const (
	MaxFileSize = 5 * 1024 * 1024 // 5MB max file size
	AvatarSize  = 256             // Avatar dimensions in pixels
	UploadsDir  = "uploads"
	AvatarsDir  = "uploads/avatars"
)

// UsersHandler provides handlers for user API endpoints
type UsersHandler struct {
	Sys *sys.SystemContext
}

// NewUsersHandler creates a new UsersHandler
func NewUsersHandler(sys *sys.SystemContext) *UsersHandler {
	return &UsersHandler{
		Sys: sys,
	}
}

// HandleUpdateUserProfile handles PUT /users/profile requests to update user profile
func (h *UsersHandler) HandleUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Not authenticated",
		})
		return
	}

	// Parse request body
	var profileRequest usermodel.ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&profileRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	// Get existing profile or create new one
	profile, err := h.Sys.Stores.Users.GetUserProfile(ctx, userID)
	if err != nil {
		utils.Error("Failed to get user profile %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get user profile",
		})
		return
	}

	// Update profile fields
	profile.UserID = userID
	profile.FullName = profileRequest.FullName
	profile.Phone = profileRequest.Phone

	// Save or update the profile
	if profile.AvatarURL == "" && profile.FullName == "" && profile.Phone == "" {
		// This is a new profile
		err = h.Sys.Stores.Users.CreateUserProfile(ctx, profile)
	} else {
		err = h.Sys.Stores.Users.UpdateUserProfile(ctx, profile)
	}

	if err != nil {
		utils.Error("Failed to save user profile %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to save user profile",
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Profile updated successfully",
		"profile": profile,
	})
}

// HandleUpdateUserPassword handles PUT /users/password requests to update user password
func (h *UsersHandler) HandleUpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Not authenticated",
		})
		return
	}

	// Parse request body
	var passwordRequest usermodel.PasswordChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&passwordRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	// Validate password requirements
	if passwordRequest.CurrentPassword == "" || passwordRequest.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Current password and new password are required",
		})
		return
	}

	// Validate password confirmation
	if passwordRequest.NewPassword != passwordRequest.ConfirmPassword {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "New password and confirm password do not match",
		})
		return
	}

	// Get current user to verify current password
	user, err := h.Sys.Stores.Users.GetUserByID(ctx, userID)
	if err != nil {
		utils.Error("Failed to get user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get user data",
		})
		return
	}

	// Verify current password
	if !gosightauth.CheckPasswordHash(passwordRequest.CurrentPassword, user.PasswordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Current password is incorrect",
		})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordRequest.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.Error("Failed to hash password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to process new password",
		})
		return
	}

	// Update password in database
	err = h.Sys.Stores.Users.UpdateUserPassword(ctx, userID, string(hashedPassword))
	if err != nil {
		utils.Error("Failed to update password for user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to update password",
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password updated successfully",
	})
}

// HandleGetUserPreferences handles GET /users/preferences requests to get user settings
func (h *UsersHandler) HandleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Not authenticated",
		})
		return
	}

	// Get user settings
	settings, err := h.Sys.Stores.Users.GetUserSettings(ctx, userID)
	if err != nil {
		utils.Error("Failed to get user settings %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get user preferences",
		})
		return
	}

	// Transform flattened settings back to nested structure for frontend
	response := make(map[string]interface{})
	notifications := make(map[string]interface{})
	dashboard := make(map[string]interface{})

	for key, rawValue := range settings {
		var value interface{}
		if err := json.Unmarshal(rawValue, &value); err != nil {
			utils.Warn("Failed to unmarshal setting %s: %v", key, err)
			continue
		}

		switch {
		case key == "theme":
			response["theme"] = value
		case strings.HasPrefix(key, "notifications."):
			subKey := strings.TrimPrefix(key, "notifications.")
			switch subKey {
			case "emailAlerts":
				notifications["email_alerts"] = value
			case "pushNotifications":
				notifications["push_alerts"] = value
			case "alertFrequency":
				notifications["alert_frequency"] = value
			}
		case strings.HasPrefix(key, "dashboard."):
			subKey := strings.TrimPrefix(key, "dashboard.")
			switch subKey {
			case "refreshInterval":
				dashboard["refresh_interval"] = value
			case "defaultTimeRange":
				dashboard["default_time_range"] = value
			case "showSystemMetrics":
				dashboard["show_system_metrics"] = value
			}
		}
	}

	// Only include nested objects if they have data
	if len(notifications) > 0 {
		response["notifications"] = notifications
	}
	if len(dashboard) > 0 {
		response["dashboard"] = dashboard
	}

	// Return structured settings
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateUserPreferences handles PUT /users/preferences requests to update user settings
func (h *UsersHandler) HandleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Not authenticated",
		})
		return
	}

	// Parse request body
	var preferencesRequest usermodel.UserPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&preferencesRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	// Flatten preferences into individual key-value pairs
	flatPreferences := make(map[string]interface{})

	// Add theme
	if preferencesRequest.Theme != "" {
		flatPreferences["theme"] = preferencesRequest.Theme
	}

	// Flatten notifications - handle detailed notification structure
	flatPreferences["notifications.emailAlerts"] = preferencesRequest.Notifications.EmailAlerts
	flatPreferences["notifications.pushNotifications"] = preferencesRequest.Notifications.PushAlerts
	flatPreferences["notifications.alertFrequency"] = preferencesRequest.Notifications.AlertFrequency

	// Flatten dashboard settings
	if preferencesRequest.Dashboard != nil {
		for key, value := range preferencesRequest.Dashboard {
			switch key {
			case "refresh_interval":
				flatPreferences["dashboard.refreshInterval"] = value
			case "default_time_range":
				flatPreferences["dashboard.defaultTimeRange"] = value
			case "show_system_metrics":
				flatPreferences["dashboard.showSystemMetrics"] = value
			}
		}
	}

	// Save each flattened preference
	for key, value := range flatPreferences {
		if value != nil {
			valueBytes, err := json.Marshal(value)
			if err != nil {
				utils.Error("Failed to marshal preference %s: %v", key, err)
				continue
			}

			err = h.Sys.Stores.Users.SetUserSetting(ctx, userID, key, valueBytes)
			if err != nil {
				utils.Error("Failed to save preference %s for user %s: %v", key, userID, err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Failed to save preferences",
				})
				return
			}
		}
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Preferences updated successfully",
	})
}

// HandleGetCompleteUser handles GET /users/me requests to get complete user data with profile and settings
func (h *UsersHandler) HandleGetCompleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Not authenticated",
		})
		return
	}

	// Get complete user data
	completeUser, err := h.Sys.Stores.Users.GetCompleteUser(ctx, userID)
	if err != nil {
		utils.Error("Failed to get complete user data %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get user data",
		})
		return
	}

	// Return user data (excluding sensitive info)
	response := map[string]interface{}{
		"id":          completeUser.User.ID,
		"username":    completeUser.User.Username,
		"email":       completeUser.User.Email,
		"first_name":  completeUser.User.FirstName,
		"last_name":   completeUser.User.LastName,
		"permissions": gosightauth.FlattenPermissions(completeUser.User.Roles),
		"profile":     completeUser.Profile,
		"settings":    completeUser.Settings,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleAPIUsers handles GET /users requests to list all users
func (h *UsersHandler) HandleAPIUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	// Get all users with permissions
	users, err := h.Sys.Stores.Users.GetAllUsersWithPermissions(ctx)
	if err != nil {
		utils.Error("Failed to get users: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get users",
		})
		return
	}

	// Return users (excluding sensitive info)
	var response []map[string]interface{}
	for _, user := range users {
		response = append(response, map[string]interface{}{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"permissions": gosightauth.FlattenPermissions(user.Roles),
			"created_at":  user.CreatedAt,
			"updated_at":  user.UpdatedAt,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleAPIUserCreate handles POST /users requests to create a new user
func (h *UsersHandler) HandleAPIUserCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	// Parse request body
	var userRequest usermodel.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&userRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	// Validate required fields
	if userRequest.Username == "" || userRequest.Email == "" || userRequest.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Username, email, and password are required",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.Error("Failed to hash password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to process password",
		})
		return
	}

	// Create user
	user := &usermodel.User{
		Username:     userRequest.Username,
		Email:        userRequest.Email,
		FirstName:    userRequest.FirstName,
		LastName:     userRequest.LastName,
		PasswordHash: string(hashedPassword),
	}

	err = h.Sys.Stores.Users.CreateUser(ctx, user)
	if err != nil {
		utils.Error("Failed to create user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to create user",
		})
		return
	}

	// Return created user (excluding sensitive info)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"created_at": user.CreatedAt,
	})
}

// HandleAPIUser handles GET /users/{id} requests to get a user by ID
func (h *UsersHandler) HandleAPIUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "User ID is required",
		})
		return
	}

	// Get user with permissions
	user, err := h.Sys.Stores.Users.GetUserWithPermissions(ctx, userID)
	if err != nil {
		utils.Error("Failed to get user %s: %v", userID, err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "User not found",
		})
		return
	}

	// Get user profile (optional)
	profile, _ := h.Sys.Stores.Users.GetUserProfile(ctx, userID)
	profileData := map[string]interface{}{}
	if profile != nil {
		profileData = map[string]interface{}{
			"full_name":  profile.FullName,
			"phone":      profile.Phone,
			"avatar_url": profile.AvatarURL,
		}
	}

	// Return user data (excluding sensitive info)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"first_name":  user.FirstName,
		"last_name":   user.LastName,
		"permissions": gosightauth.FlattenPermissions(user.Roles),
		"profile":     profileData,
		"created_at":  user.CreatedAt,
		"updated_at":  user.UpdatedAt,
	})
}

// HandleAPIUserUpdate handles PUT /users/{id} requests to update a user
func (h *UsersHandler) HandleAPIUserUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "User ID is required",
		})
		return
	}

	// Parse request body
	var updateRequest usermodel.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	// Get existing user
	user, err := h.Sys.Stores.Users.GetUserByID(ctx, userID)
	if err != nil {
		utils.Error("Failed to get user %s: %v", userID, err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "User not found",
		})
		return
	}

	// Update user fields
	if updateRequest.Username != "" {
		user.Username = updateRequest.Username
	}
	if updateRequest.Email != "" {
		user.Email = updateRequest.Email
	}
	if updateRequest.FirstName != "" {
		user.FirstName = updateRequest.FirstName
	}
	if updateRequest.LastName != "" {
		user.LastName = updateRequest.LastName
	}

	// Update user in database
	err = h.Sys.Stores.Users.UpdateUser(ctx, user)
	if err != nil {
		utils.Error("Failed to update user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to update user",
		})
		return
	}

	// Return updated user (excluding sensitive info)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"updated_at": user.UpdatedAt,
	})
}

// HandleAPIUserDelete handles DELETE /users/{id} requests to delete a user
func (h *UsersHandler) HandleAPIUserDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "User ID is required",
		})
		return
	}

	// Delete user
	err := h.Sys.Stores.Users.DeleteUser(ctx, userID)
	if err != nil {
		utils.Error("Failed to delete user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to delete user",
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User deleted successfully",
	})
}

// HandleAPIUserPasswordChange handles POST /users/{id}/password requests to change a user's password (admin)
func (h *UsersHandler) HandleAPIUserPasswordChange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "User ID is required",
		})
		return
	}

	// Parse request body
	var passwordRequest usermodel.AdminPasswordChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&passwordRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	// Validate password requirements
	if passwordRequest.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "New password is required",
		})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordRequest.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.Error("Failed to hash password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to process new password",
		})
		return
	}

	// Update password in database
	err = h.Sys.Stores.Users.UpdateUserPassword(ctx, userID, string(hashedPassword))
	if err != nil {
		utils.Error("Failed to update password for user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to update password",
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password updated successfully",
	})
}

// HandleAPIUserSettings handles GET /users/{id}/settings requests to get user settings by ID
func (h *UsersHandler) HandleAPIUserSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "User ID is required",
		})
		return
	}

	// Get user settings
	settings, err := h.Sys.Stores.Users.GetUserSettings(ctx, userID)
	if err != nil {
		utils.Error("Failed to get user settings %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get user settings",
		})
		return
	}

	// Transform flattened settings back to nested structure
	response := make(map[string]interface{})
	notifications := make(map[string]interface{})
	dashboard := make(map[string]interface{})

	for key, rawValue := range settings {
		var value interface{}
		if err := json.Unmarshal(rawValue, &value); err != nil {
			utils.Warn("Failed to unmarshal setting %s: %v", key, err)
			continue
		}

		switch {
		case key == "theme":
			response["theme"] = value
		case strings.HasPrefix(key, "notifications."):
			subKey := strings.TrimPrefix(key, "notifications.")
			notifications[subKey] = value
		case strings.HasPrefix(key, "dashboard."):
			subKey := strings.TrimPrefix(key, "dashboard.")
			dashboard[subKey] = value
		}
	}

	// Include nested objects if they have data
	if len(notifications) > 0 {
		response["notifications"] = notifications
	}
	if len(dashboard) > 0 {
		response["dashboard"] = dashboard
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleAPIUserSettingsUpdate handles PUT /users/{id}/settings requests to update user settings by ID
func (h *UsersHandler) HandleAPIUserSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "User ID is required",
		})
		return
	}

	// Parse request body
	var settingsRequest usermodel.UserPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&settingsRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	// Flatten settings into individual key-value pairs
	flatSettings := make(map[string]interface{})

	// Add theme
	if settingsRequest.Theme != "" {
		flatSettings["theme"] = settingsRequest.Theme
	}

	// Flatten notifications
	// Check if any notification field has a non-zero value
	notifications := settingsRequest.Notifications
	if notifications.EmailAlerts || notifications.PushAlerts || notifications.AlertFrequency != "" {
		flatSettings["notifications.emailAlerts"] = notifications.EmailAlerts
		flatSettings["notifications.pushNotifications"] = notifications.PushAlerts
		flatSettings["notifications.alertFrequency"] = notifications.AlertFrequency
	}

	// Flatten dashboard settings
	if settingsRequest.Dashboard != nil {
		for key, value := range settingsRequest.Dashboard {
			flatSettings["dashboard."+key] = value
		}
	}

	// Save each flattened setting
	for key, value := range flatSettings {
		if value != nil {
			valueBytes, err := json.Marshal(value)
			if err != nil {
				utils.Error("Failed to marshal setting %s: %v", key, err)
				continue
			}

			err = h.Sys.Stores.Users.SetUserSetting(ctx, userID, key, valueBytes)
			if err != nil {
				utils.Error("Failed to save setting %s for user %s: %v", key, userID, err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Failed to save settings",
				})
				return
			}
		}
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Settings updated successfully",
	})
}

// HandleUploadAvatar handles POST /users/avatar requests to upload user avatar
func (h *UsersHandler) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from context
	userID, ok := contextutil.GetUserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Authentication required",
		})
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(MaxFileSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to parse upload form",
		})
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("avatar")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "No file provided",
		})
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid file type. Only JPEG, PNG, and GIF are allowed",
		})
		return
	}

	// Validate file size
	if header.Size > MaxFileSize {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "File too large. Maximum size is 5MB",
		})
		return
	}

	// Create uploads directory if it doesn't exist
	if err := ensureUploadsDir(); err != nil {
		utils.Error("Failed to create uploads directory: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Server error",
		})
		return
	}

	// Process the image
	avatarURL, err := h.processAvatar(file, userID)
	if err != nil {
		utils.Error("Failed to process avatar for user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to process image",
		})
		return
	}

	// Update user profile with new avatar URL
	ctx := r.Context()
	profile, err := h.Sys.Stores.Users.GetUserProfile(ctx, userID)
	if err != nil {
		// Create new profile if none exists
		profile = &usermodel.UserProfile{
			UserID: userID,
		}
	}

	// Delete old avatar file if it exists and is a local upload
	if profile.AvatarURL != "" && strings.HasPrefix(profile.AvatarURL, "/uploads/") {
		// Strip query parameters (cache-busting) from URL before deleting
		oldAvatarPath := profile.AvatarURL
		if strings.Contains(oldAvatarPath, "?") {
			oldAvatarPath = strings.Split(oldAvatarPath, "?")[0]
		}
		oldFilePath := filepath.Join(".", oldAvatarPath)
		if err := os.Remove(oldFilePath); err != nil {
			utils.Warn("Failed to delete old avatar file %s: %v", oldFilePath, err)
		} else {
			utils.Info("Deleted old avatar file: %s", oldFilePath)
		}
	}

	profile.AvatarURL = avatarURL
	err = h.Sys.Stores.Users.CreateUserProfile(ctx, profile)
	if err != nil {
		utils.Error("Failed to update user profile with avatar: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to save avatar",
		})
		return
	}

	utils.Info("Avatar uploaded successfully for user %s: %s", userID, avatarURL)

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message":    "Avatar uploaded successfully",
		"avatar_url": avatarURL,
	})
}

// HandleCropAvatar handles POST /users/avatar/crop requests to crop user avatar
func (h *UsersHandler) HandleCropAvatar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from context
	userID, ok := contextutil.GetUserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Authentication required",
		})
		return
	}

	// Parse the crop data
	var cropData struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Width  int `json:"width"`
		Height int `json:"height"`
	}

	if err := json.NewDecoder(r.Body).Decode(&cropData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid crop data",
		})
		return
	}

	// Get current user profile to find existing avatar
	ctx := r.Context()
	profile, err := h.Sys.Stores.Users.GetUserProfile(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to get user profile",
		})
		return
	}

	// Check if user has an avatar to crop
	if profile.AvatarURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "No avatar to crop",
		})
		return
	}

	// For external URLs (SSO avatars), we can't crop them
	if !strings.HasPrefix(profile.AvatarURL, "/uploads/") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Cannot crop external avatar. Please upload a new image.",
		})
		return
	}

	// Open the existing avatar file
	// Strip query parameters (cache-busting) from URL before opening file
	avatarPath := profile.AvatarURL
	if strings.Contains(avatarPath, "?") {
		avatarPath = strings.Split(avatarPath, "?")[0]
	}
	avatarFilePath := filepath.Join(".", avatarPath[1:]) // Remove leading slash
	file, err := os.Open(avatarFilePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to open avatar file",
		})
		return
	}
	defer file.Close()

	// Process the cropped avatar
	newAvatarURL, err := h.processCroppedAvatar(file, userID, cropData.X, cropData.Y, cropData.Width, cropData.Height)
	if err != nil {
		utils.Error("Failed to process cropped avatar: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to process cropped avatar",
		})
		return
	}

	// Update user profile with new avatar URL
	profile.AvatarURL = newAvatarURL
	err = h.Sys.Stores.Users.CreateUserProfile(ctx, profile)
	if err != nil {
		utils.Error("Failed to update user profile: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to update profile",
		})
		return
	}

	utils.Info("Avatar cropped for user %s", userID)

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"avatar_url": newAvatarURL,
		"message":    "Avatar cropped successfully",
	})
}

// HandleDeleteAvatar handles DELETE /users/avatar requests to delete user avatar
func (h *UsersHandler) HandleDeleteAvatar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from context
	userID, ok := contextutil.GetUserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Authentication required",
		})
		return
	}

	ctx := r.Context()

	// Get current user profile
	profile, err := h.Sys.Stores.Users.GetUserProfile(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Profile not found",
		})
		return
	}

	// Delete the file if it's a local upload (not external URL)
	if profile.AvatarURL != "" && strings.HasPrefix(profile.AvatarURL, "/uploads/") {
		// Strip query parameters (cache-busting) from URL before deleting
		avatarPath := profile.AvatarURL
		if strings.Contains(avatarPath, "?") {
			avatarPath = strings.Split(avatarPath, "?")[0]
		}
		filePath := filepath.Join(".", avatarPath)
		if err := os.Remove(filePath); err != nil {
			utils.Warn("Failed to delete avatar file %s: %v", filePath, err)
		} else {
			utils.Info("Deleted avatar file: %s", filePath)
		}
	}

	// Clear avatar URL from profile
	profile.AvatarURL = ""
	err = h.Sys.Stores.Users.CreateUserProfile(ctx, profile)
	if err != nil {
		utils.Error("Failed to update user profile: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to delete avatar",
		})
		return
	}

	utils.Info("Avatar deleted for user %s", userID)

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Avatar deleted successfully",
	})
}

// HandleGetUploadLimits handles GET /upload/limits requests to get file upload limits
func (h *UsersHandler) HandleGetUploadLimits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limits := map[string]interface{}{
		"max_file_size":    MaxFileSize,
		"max_file_size_mb": MaxFileSize / (1024 * 1024),
		"avatar_size":      AvatarSize,
		"supported_types":  []string{"image/jpeg", "image/png", "image/gif"},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(limits)
}

// Helper functions for avatar processing

// processAvatar resizes and saves the uploaded image
func (h *UsersHandler) processAvatar(file io.Reader, userID string) (string, error) {
	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %v", err)
	}

	// Resize image to avatar size while maintaining aspect ratio
	resizedImg := resize.Resize(AvatarSize, AvatarSize, img, resize.Lanczos3)

	// Generate unique filename
	filename := fmt.Sprintf("%s_%s.jpg", userID, uuid.New().String()[:8])
	filePath := filepath.Join(AvatarsDir, filename)

	// Create the file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	// Encode as JPEG with high quality
	err = jpeg.Encode(outFile, resizedImg, &jpeg.Options{Quality: 90})
	if err != nil {
		return "", fmt.Errorf("failed to encode image: %v", err)
	}

	// Return the web-accessible URL with cache-busting parameter
	avatarURL := "/" + strings.Replace(filePath, "\\", "/", -1)
	// Add timestamp to prevent browser caching
	avatarURL += fmt.Sprintf("?v=%d", time.Now().Unix())
	return avatarURL, nil
}

// processCroppedAvatar crops and resizes the uploaded image
func (h *UsersHandler) processCroppedAvatar(file io.Reader, userID string, x, y, width, height int) (string, error) {
	utils.Info("Starting crop process for user %s with coords: x=%d, y=%d, w=%d, h=%d", userID, x, y, width, height)

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		utils.Error("Failed to decode image for cropping: %v", err)
		return "", fmt.Errorf("failed to decode image: %v", err)
	}

	// Get actual image bounds
	bounds := img.Bounds()
	actualWidth := bounds.Dx()
	actualHeight := bounds.Dy()
	utils.Debug("Actual image bounds: %dx%d", actualWidth, actualHeight)

	// Handle coordinate scaling and bounds checking
	// The crop coordinates might be based on the original image size before upload processing

	// First, ensure basic bounds are valid
	if x < 0 || y < 0 || width <= 0 || height <= 0 {
		utils.Error("Invalid crop coordinates: x=%d, y=%d, w=%d, h=%d", x, y, width, height)
		return "", fmt.Errorf("invalid crop coordinates")
	}

	// If coordinates are out of bounds, attempt proportional scaling
	if x >= actualWidth || y >= actualHeight || x+width > actualWidth || y+height > actualHeight {
		utils.Debug("Crop coordinates out of bounds, attempting proportional scaling")

		// Calculate the maximum possible scale to fit within bounds
		maxScaleX := float64(actualWidth) / float64(x+width)
		maxScaleY := float64(actualHeight) / float64(y+height)
		scale := math.Min(maxScaleX, maxScaleY)

		// Only scale if it would help and scale is reasonable (between 0.1 and 1.0)
		if scale > 0.1 && scale < 1.0 {
			x = int(float64(x) * scale)
			y = int(float64(y) * scale)
			width = int(float64(width) * scale)
			height = int(float64(height) * scale)
			utils.Debug("Applied scale %.3f: x=%d, y=%d, w=%d, h=%d", scale, x, y, width, height)
		} else {
			// Fallback: clamp coordinates to image bounds
			if x >= actualWidth {
				x = actualWidth - 1
			}
			if y >= actualHeight {
				y = actualHeight - 1
			}
			if x+width > actualWidth {
				width = actualWidth - x
			}
			if y+height > actualHeight {
				height = actualHeight - y
			}
			utils.Debug("Clamped coordinates: x=%d, y=%d, w=%d, h=%d", x, y, width, height)
		}
	}

	// Final validation after scaling/clamping
	if x < 0 || y < 0 || x >= actualWidth || y >= actualHeight ||
		width <= 0 || height <= 0 || x+width > actualWidth || y+height > actualHeight {
		utils.Error("Crop coordinates still invalid after adjustment: x=%d, y=%d, w=%d, h=%d, image_size=%dx%d",
			x, y, width, height, actualWidth, actualHeight)
		return "", fmt.Errorf("crop coordinates cannot be adjusted to fit image bounds")
	}

	// Create a cropped image
	subImager, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	})
	if !ok {
		utils.Error("Image does not support SubImage interface")
		return "", fmt.Errorf("image type does not support cropping")
	}

	croppedImg := subImager.SubImage(image.Rect(x, y, x+width, y+height))
	utils.Debug("Image cropped successfully")

	// Resize to avatar size
	resizedImg := resize.Resize(AvatarSize, AvatarSize, croppedImg, resize.Lanczos3)
	utils.Debug("Image resized to %dx%d", AvatarSize, AvatarSize)

	// Generate unique filename
	filename := fmt.Sprintf("%s_%s.jpg", userID, uuid.New().String()[:8])
	filePath := filepath.Join(AvatarsDir, filename)
	utils.Debug("Generated file path: %s", filePath)

	// Ensure avatars directory exists
	if err := ensureUploadsDir(); err != nil {
		utils.Error("Failed to ensure uploads directory: %v", err)
		return "", fmt.Errorf("failed to create uploads directory: %v", err)
	}

	// Create the file
	outFile, err := os.Create(filePath)
	if err != nil {
		utils.Error("Failed to create output file %s: %v", filePath, err)
		return "", fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Encode as JPEG with high quality
	err = jpeg.Encode(outFile, resizedImg, &jpeg.Options{Quality: 90})
	if err != nil {
		utils.Error("Failed to encode image to JPEG: %v", err)
		return "", fmt.Errorf("failed to encode image: %v", err)
	}

	// Return the web-accessible URL with cache-busting parameter
	avatarURL := fmt.Sprintf("/uploads/avatars/%s?v=%d", filename, time.Now().Unix())
	utils.Info("Crop process completed successfully for user %s, avatar URL: %s", userID, avatarURL)
	return avatarURL, nil
}

func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

func ensureUploadsDir() error {
	// Create uploads directory
	if err := os.MkdirAll(UploadsDir, 0755); err != nil {
		return err
	}

	// Create avatars subdirectory
	if err := os.MkdirAll(AvatarsDir, 0755); err != nil {
		return err
	}

	return nil
}

// AssignRoleRequest represents the request body for assigning roles to a user
type AssignRoleRequest struct {
	RoleIDs []uuid.UUID `json:"role_ids"`
}

// GetUserRoles returns all roles assigned to a user
func (h *UsersHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	roles, err := h.Sys.Stores.Users.GetUserRoles(ctx, userID.String())
	if err != nil {
		http.Error(w, "Failed to fetch user roles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

// AssignRolesToUser assigns roles to a user
func (h *UsersHandler) AssignRolesToUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Convert UUIDs to strings
	roleIDs := make([]string, len(req.RoleIDs))
	for i, roleID := range req.RoleIDs {
		roleIDs[i] = roleID.String()
	}

	// Use the UserStore method to assign roles
	err = h.Sys.Stores.Users.AssignRolesToUser(ctx, userID.String(), roleIDs)
	if err != nil {
		utils.Error("Failed to assign roles to user %s: %v", userID, err)
		http.Error(w, "Failed to assign roles", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
