package images

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/packs"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
)

var storage ImageStorage = NewDBImageStorage()

// processUploadedImage handles file validation and image processing
func processUploadedImage(c *gin.Context) (*ProcessedImage, error) {
	// Get file from form data
	file, err := c.FormFile("image")
	if err != nil {
		return nil, errors.New("no image file provided")
	}

	// Check file size (5MB limit)
	if file.Size > MaxUploadSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed (%d bytes)", MaxUploadSize)
	}

	// Open file
	fileReader, err := file.Open()
	if err != nil {
		return nil, errors.New("failed to read uploaded file")
	}
	defer fileReader.Close()

	// Process image
	processed, err := ProcessImageFromReader(fileReader)
	if err != nil {
		return nil, err
	}

	return processed, nil
}

// UploadPackImage uploads or updates an image for a pack
// @Summary Upload or update pack image
// @Description Upload or update an image for a pack. Only the pack owner can upload images.
// @Security Bearer
// @Tags Pack Images
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Pack ID"
// @Param image formData file true "Image file (JPEG, PNG, or WebP, max 5MB)"
// @Success 200 {object} map[string]interface{} "Image uploaded successfully"
// @Failure 400 {object} map[string]string "Invalid request or image"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Pack does not belong to user"
// @Failure 404 {object} map[string]string "Pack not found"
// @Failure 413 {object} map[string]string "Payload too large"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/mypack/{id}/image [post]
func UploadPackImage(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract user ID from JWT
	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "upload pack image: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	// Parse pack ID from URL
	packID, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID format"})
		return
	}

	// Check if pack exists
	_, err = packs.FindPackByID(packID)
	if err != nil {
		if errors.Is(err, packs.ErrPackNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
		helper.LogAndSanitize(err, "upload pack image: find pack failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	// Check pack ownership
	isOwner, err := packs.CheckPackOwnership(packID, userID)
	if err != nil {
		helper.LogAndSanitize(err, "upload pack image: check pack ownership failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}

	// Process uploaded image
	processed, err := processUploadedImage(c)
	if err != nil {
		// Handle specific error types
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
		// Generic file upload errors
		if err.Error() == "no image file provided" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
			return
		}
		helper.LogAndSanitize(err, "upload pack image: process image failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	// Save to storage
	err = storage.Save(ctx, packID, processed.Data, processed.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Return success response with metadata
	c.JSON(http.StatusOK, gin.H{
		"message":   "Image uploaded successfully",
		"pack_id":   packID,
		"mime_type": processed.Metadata.MimeType,
		"file_size": processed.Metadata.FileSize,
		"width":     processed.Metadata.Width,
		"height":    processed.Metadata.Height,
	})
}

// GetPackImage retrieves an image for a pack
// @Summary Get pack image
// @Description Get the image for a pack. Public packs are accessible without authentication.
// @Tags Pack Images
// @Produce image/jpeg
// @Param id path int true "Pack ID"
// @Success 200 {file} image/jpeg "Pack image"
// @Failure 400 {string} string "Invalid pack ID"
// @Failure 401 {string} string "Unauthorized (for private packs)"
// @Failure 403 {string} string "Forbidden (for private packs not owned by user)"
// @Failure 404 {string} string "Pack not found or has no image"
// @Failure 500 {string} string "Internal server error"
// @Router /v1/packs/{id}/image [get]
func GetPackImage(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pack ID from URL
	packID, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid pack ID format")
		return
	}

	// Check if pack exists and if it's public
	pack, err := packs.FindPackByID(packID)
	if err != nil {
		if errors.Is(err, packs.ErrPackNotFound) {
			c.String(http.StatusNotFound, "Pack not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to retrieve pack")
		return
	}

	// Check access permissions for private packs
	if pack.SharingCode == nil || *pack.SharingCode == "" {
		// Private pack - require authentication and ownership
		userID, err := security.ExtractTokenID(c)
		if err != nil {
			c.String(http.StatusUnauthorized, "Unauthorized")
			return
		}

		isOwner, err := packs.CheckPackOwnership(packID, userID)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to verify ownership")
			return
		}
		if !isOwner {
			c.String(http.StatusForbidden, "Forbidden")
			return
		}
	}

	// Retrieve image from storage
	img, err := storage.Get(ctx, packID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.String(http.StatusNotFound, "Pack has no image")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to retrieve image")
		return
	}

	// Set cache headers based on stable metadata
	isPrivate := pack.SharingCode == nil || *pack.SharingCode == ""
	cacheControl := "public, max-age=86400" // 24 hours
	if isPrivate {
		cacheControl = "private, max-age=86400"
	}
	etag := fmt.Sprintf(`"%d-%d"`, packID, pack.UpdatedAt.Unix())
	c.Header("Cache-Control", cacheControl)
	c.Header("ETag", etag)
	c.Header("Last-Modified", pack.UpdatedAt.UTC().Format(http.TimeFormat))

	// Return image binary
	c.Data(http.StatusOK, img.Metadata.MimeType, img.Data)
}

// DeletePackImage deletes an image for a pack
// @Summary Delete pack image
// @Description Delete the image for a pack. Only the pack owner can delete images.
// @Security Bearer
// @Tags Pack Images
// @Produce json
// @Param id path int true "Pack ID"
// @Success 200 {object} map[string]interface{} "Image deleted successfully"
// @Failure 400 {object} map[string]string "Invalid pack ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Pack does not belong to user"
// @Failure 404 {object} map[string]string "Pack not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/mypack/{id}/image [delete]
func DeletePackImage(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract user ID from JWT
	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "delete pack image: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	// Parse pack ID from URL
	packID, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID format"})
		return
	}

	// Check if pack exists
	_, err = packs.FindPackByID(packID)
	if err != nil {
		if errors.Is(err, packs.ErrPackNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
		helper.LogAndSanitize(err, "delete pack image: find pack failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	// Check pack ownership
	isOwner, err := packs.CheckPackOwnership(packID, userID)
	if err != nil {
		helper.LogAndSanitize(err, "delete pack image: check pack ownership failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}

	// Delete image from storage (idempotent)
	err = storage.Delete(ctx, packID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image deleted successfully",
		"pack_id": packID,
	})
}
