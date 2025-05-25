// File: server/internal/http/handleFileAPI.go
// Description: This file contains the file upload handlers for the GoSight server including profile picture uploads.

package httpserver

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaronlmathis/gosight-server/internal/contextutil"
	"github.com/aaronlmathis/gosight-server/internal/usermodel"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

const (
	MaxFileSize = 5 * 1024 * 1024 // 5MB max file size
	AvatarSize  = 256             // Avatar dimensions in pixels
	UploadsDir  = "uploads"
	AvatarsDir  = "uploads/avatars"
)

// HandleUploadAvatar handles profile picture upload and processing
func (s *HttpServer) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
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
	avatarURL, err := s.processAvatar(file, userID)
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
	profile, err := s.Sys.Stores.Users.GetUserProfile(ctx, userID)
	if err != nil {
		// Create new profile if none exists
		profile = &usermodel.UserProfile{
			UserID: userID,
		}
	}

	profile.AvatarURL = avatarURL
	err = s.Sys.Stores.Users.CreateUserProfile(ctx, profile)
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

// HandleDeleteAvatar handles profile picture deletion
func (s *HttpServer) HandleDeleteAvatar(w http.ResponseWriter, r *http.Request) {
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
	profile, err := s.Sys.Stores.Users.GetUserProfile(ctx, userID)
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
		filePath := filepath.Join(".", profile.AvatarURL)
		if err := os.Remove(filePath); err != nil {
			utils.Warn("Failed to delete avatar file %s: %v", filePath, err)
		}
	}

	// Clear avatar URL from profile
	profile.AvatarURL = ""
	err = s.Sys.Stores.Users.CreateUserProfile(ctx, profile)
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

// processAvatar resizes and saves the uploaded image
func (s *HttpServer) processAvatar(file io.Reader, userID string) (string, error) {
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

	// Return the web-accessible URL
	avatarURL := "/" + strings.Replace(filePath, "\\", "/", -1)
	return avatarURL, nil
}

// HandleCropAvatar handles avatar cropping with specified coordinates
func (s *HttpServer) HandleCropAvatar(w http.ResponseWriter, r *http.Request) {
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
	profile, err := s.Sys.Stores.Users.GetUserProfile(ctx, userID)
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
	avatarPath := filepath.Join(".", profile.AvatarURL[1:]) // Remove leading slash
	file, err := os.Open(avatarPath)
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
	newAvatarURL, err := s.processCroppedAvatar(file, userID, cropData.X, cropData.Y, cropData.Width, cropData.Height)
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
	err = s.Sys.Stores.Users.CreateUserProfile(ctx, profile)
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

// processCroppedAvatar crops and resizes the uploaded image
func (s *HttpServer) processCroppedAvatar(file io.Reader, userID string, x, y, width, height int) (string, error) {
	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %v", err)
	}

	// Crop the image
	bounds := img.Bounds()

	// Validate crop coordinates
	if x < 0 || y < 0 || x+width > bounds.Dx() || y+height > bounds.Dy() {
		return "", fmt.Errorf("crop coordinates are out of bounds")
	}

	// Create a cropped image
	croppedImg := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(x, y, x+width, y+height))

	// Resize to avatar size
	resizedImg := resize.Resize(AvatarSize, AvatarSize, croppedImg, resize.Lanczos3)

	// Generate unique filename
	filename := fmt.Sprintf("%s_%s.jpg", userID, uuid.New().String()[:8])
	filePath := filepath.Join(AvatarsDir, filename)

	// Create the file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Encode as JPEG with high quality
	err = jpeg.Encode(outFile, resizedImg, &jpeg.Options{Quality: 90})
	if err != nil {
		return "", fmt.Errorf("failed to encode image: %v", err)
	}

	// Return the web-accessible URL
	return fmt.Sprintf("/uploads/avatars/%s", filename), nil
}

// Helper functions

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

// HandleGetUploadLimits returns upload configuration limits
func (s *HttpServer) HandleGetUploadLimits(w http.ResponseWriter, r *http.Request) {
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
