// File: server/internal/http/handleFileAPI.go
// Description: This file contains the file upload handlers for the GoSight server including profile picture uploads.

package httpserver

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

	// Return the web-accessible URL with cache-busting parameter
	avatarURL := "/" + strings.Replace(filePath, "\\", "/", -1)
	// Add timestamp to prevent browser caching
	avatarURL += fmt.Sprintf("?v=%d", time.Now().Unix())
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
