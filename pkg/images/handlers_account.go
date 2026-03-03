package images

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
)

var accountStorage AccountImageStorage = NewDBAccountImageStorage()

// processUploadedProfileImage handles file validation and profile image processing
func processUploadedProfileImage(c *gin.Context) (*ProcessedImage, error) {
	file, err := c.FormFile("image")
	if err != nil {
		return nil, errors.New(ErrMsgNoImageProvided)
	}

	if file.Size > MaxUploadSize {
		return nil, fmt.Errorf("%w: file size exceeds maximum allowed (%d bytes)", ErrTooLarge, MaxUploadSize)
	}

	fileReader, err := file.Open()
	if err != nil {
		return nil, errors.New("failed to read uploaded file")
	}
	defer fileReader.Close()

	processed, err := ProcessProfileImageFromReader(fileReader)
	if err != nil {
		return nil, err
	}

	return processed, nil
}

// UploadMyProfileImage uploads or updates the profile image for the current user
// @Summary Upload or update profile image
// @Description Upload or update the profile image for the currently logged-in user
// @Security Bearer
// @Tags Account Images
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file (JPEG, PNG, or WebP, max 5MB)"
// @Success 200 {object} map[string]interface{} "Image uploaded successfully"
// @Failure 400 {object} map[string]string "Invalid request or image"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 413 {object} map[string]string "Payload too large"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/myaccount/image [post]
func UploadMyProfileImage(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "upload profile image: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	processed, err := processUploadedProfileImage(c)
	if err != nil {
		if errors.Is(err, ErrInvalidFormat) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image format. Only JPEG, PNG, and WebP are supported"})
			return
		}
		if errors.Is(err, ErrTooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File size exceeds maximum allowed"})
			return
		}
		if errors.Is(err, ErrCorrupted) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Corrupted or invalid image file"})
			return
		}
		if err.Error() == ErrMsgNoImageProvided {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
			return
		}
		helper.LogAndSanitize(err, "upload profile image: process image failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	err = accountStorage.Save(ctx, userID, processed.Data, processed.Metadata)
	if err != nil {
		helper.LogAndSanitize(err, "upload profile image: save image failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Profile image uploaded successfully",
		"mime_type": processed.Metadata.MimeType,
		"file_size": processed.Metadata.FileSize,
		"width":     processed.Metadata.Width,
		"height":    processed.Metadata.Height,
	})
}

// DeleteMyProfileImage deletes the profile image for the current user
// @Summary Delete profile image
// @Description Delete the profile image for the currently logged-in user
// @Security Bearer
// @Tags Account Images
// @Produce json
// @Success 200 {object} map[string]interface{} "Image deleted successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/myaccount/image [delete]
func DeleteMyProfileImage(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "delete profile image: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	err = accountStorage.Delete(ctx, userID)
	if err != nil {
		helper.LogAndSanitize(err, "delete profile image: delete image failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile image deleted successfully",
	})
}

// GetProfileImage retrieves the profile image for a given account
// @Summary Get profile image
// @Description Get the profile image for a user account
// @Tags Account Images
// @Produce image/jpeg
// @Param id path int true "Account ID"
// @Success 200 {file} image/jpeg "Profile image"
// @Failure 400 {string} string "Invalid account ID"
// @Failure 404 {string} string "No profile image found"
// @Failure 500 {string} string "Internal server error"
// @Router /v1/accounts/{id}/image [get]
func GetProfileImage(c *gin.Context) {
	ctx := c.Request.Context()

	accountID, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid account ID format")
		return
	}

	img, err := accountStorage.Get(ctx, accountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.String(http.StatusNotFound, "No profile image found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to retrieve image")
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, img.Metadata.MimeType, img.Data)
}
