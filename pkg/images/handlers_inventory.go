package images

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/inventories"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
)

var inventoryStorage InventoryImageStorage = NewDBInventoryImageStorage()

// UploadInventoryItemImage uploads or updates an image for an inventory item
// @Summary Upload or update inventory item image
// @Description Upload or update an image for an inventory item. Only the item owner can upload images.
// @Security Bearer
// @Tags Inventory Images
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Inventory Item ID"
// @Param image formData file true "Image file (JPEG, PNG, or WebP, max 5MB)"
// @Success 200 {object} map[string]interface{} "Image uploaded successfully"
// @Failure 400 {object} map[string]string "Invalid request or image"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Item does not belong to user"
// @Failure 404 {object} map[string]string "Item not found"
// @Failure 413 {object} map[string]string "Payload too large"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/myinventory/{id}/image [post]
func UploadInventoryItemImage(c *gin.Context) {
	if !config.FeatureItemPicturesUpload {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Item picture upload is currently disabled"})
		return
	}

	ctx := c.Request.Context()

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "upload inventory image: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	itemID, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID format"})
		return
	}

	_, err = inventories.FindInventoryByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, inventories.ErrNoItemFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Inventory item not found"})
			return
		}
		helper.LogAndSanitize(err, "upload inventory image: find item failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	isOwner, err := inventories.CheckInventoryOwnership(ctx, itemID, userID)
	if err != nil {
		helper.LogAndSanitize(err, "upload inventory image: check ownership failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "This item does not belong to you"})
		return
	}

	processed, err := processUploadedInventoryItemImage(c)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidFormat):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image format. Only JPEG, PNG, and WebP are supported"})
		case errors.Is(err, ErrTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File size exceeds maximum allowed"})
		case errors.Is(err, ErrCorrupted):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Corrupted or invalid image file"})
		case err.Error() == ErrMsgNoImageProvided:
			c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		default:
			helper.LogAndSanitize(err, "upload inventory image: process image failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		}
		return
	}

	err = inventoryStorage.Save(ctx, itemID, processed.Data, processed.Metadata)
	if err != nil {
		helper.LogAndSanitize(err, "upload inventory image: save image failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Image uploaded successfully",
		"item_id":   itemID,
		"mime_type": processed.Metadata.MimeType,
		"file_size": processed.Metadata.FileSize,
		"width":     processed.Metadata.Width,
		"height":    processed.Metadata.Height,
	})
}

// GetInventoryItemImage retrieves an image for an inventory item
// @Summary Get inventory item image
// @Description Get the image for an inventory item. Only the item owner can view the image.
// @Security Bearer
// @Tags Inventory Images
// @Produce image/jpeg
// @Param id path int true "Inventory Item ID"
// @Success 200 {file} image/jpeg "Inventory item image"
// @Failure 400 {string} string "Invalid item ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Item not found or has no image"
// @Failure 500 {string} string "Internal server error"
// @Router /v1/myinventory/{id}/image [get]
func GetInventoryItemImage(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "get inventory image: extract token ID failed")
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	itemID, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid item ID format")
		return
	}

	_, err = inventories.FindInventoryByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, inventories.ErrNoItemFound) {
			c.String(http.StatusNotFound, "Inventory item not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to find item")
		return
	}

	isOwner, err := inventories.CheckInventoryOwnership(ctx, itemID, userID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to verify ownership")
		return
	}
	if !isOwner {
		c.String(http.StatusForbidden, "Forbidden")
		return
	}

	img, err := inventoryStorage.Get(ctx, itemID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.String(http.StatusNotFound, "Item has no image")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to retrieve image")
		return
	}

	c.Header("Cache-Control", "private, max-age=86400")
	c.Data(http.StatusOK, img.Metadata.MimeType, img.Data)
}

// DeleteInventoryItemImage deletes an image for an inventory item
// @Summary Delete inventory item image
// @Description Delete the image for an inventory item. Only the item owner can delete images.
// @Security Bearer
// @Tags Inventory Images
// @Produce json
// @Param id path int true "Inventory Item ID"
// @Success 200 {object} map[string]interface{} "Image deleted successfully"
// @Failure 400 {object} map[string]string "Invalid item ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Item does not belong to user"
// @Failure 404 {object} map[string]string "Item not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/myinventory/{id}/image [delete]
func DeleteInventoryItemImage(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "delete inventory image: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	itemID, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID format"})
		return
	}

	_, err = inventories.FindInventoryByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, inventories.ErrNoItemFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Inventory item not found"})
			return
		}
		helper.LogAndSanitize(err, "delete inventory image: find item failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	isOwner, err := inventories.CheckInventoryOwnership(ctx, itemID, userID)
	if err != nil {
		helper.LogAndSanitize(err, "delete inventory image: check ownership failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "This item does not belong to you"})
		return
	}

	err = inventoryStorage.Delete(ctx, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image deleted successfully",
		"item_id": itemID,
	})
}

// processUploadedInventoryItemImage handles file validation and image processing for inventory items
func processUploadedInventoryItemImage(c *gin.Context) (*ProcessedImage, error) {
	file, err := c.FormFile("image")
	if err != nil {
		return nil, errors.New(ErrMsgNoImageProvided)
	}

	if file.Size > MaxUploadSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed (%d bytes): %w", MaxUploadSize, ErrTooLarge)
	}

	fileReader, err := file.Open()
	if err != nil {
		return nil, errors.New("failed to read uploaded file")
	}
	defer fileReader.Close()

	processed, err := ProcessInventoryItemImageFromReader(fileReader)
	if err != nil {
		return nil, err
	}

	return processed, nil
}
