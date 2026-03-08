package images

import (
	"errors"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/packs"
	"github.com/gin-gonic/gin"
)

// GetPackItemImage retrieves an image for an item within a pack
// @Summary Get pack item image
// @Description Get the image for an inventory item within a pack. Public packs are accessible without authentication.
// @Tags Pack Images
// @Produce image/jpeg
// @Param id path int true "Pack ID"
// @Param itemId path int true "Item (Inventory) ID"
// @Success 200 {file} image/jpeg "Item image"
// @Failure 400 {string} string "Invalid pack ID or item ID"
// @Failure 404 {string} string "Pack not found, item not in pack, or has no image"
// @Failure 500 {string} string "Internal server error"
// @Router /v1/packs/{id}/items/{itemId}/image [get]
func GetPackItemImage(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pack ID from URL
	packID, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid pack ID format")
		return
	}

	// Parse item ID from URL
	itemID, err := helper.StringToUint(c.Param("itemId"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid item ID format")
		return
	}

	// Check if pack exists and if it's public
	pack, err := packs.FindPackByID(ctx, packID)
	if err != nil {
		if errors.Is(err, packs.ErrPackNotFound) {
			c.String(http.StatusNotFound, "Pack not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to retrieve pack")
		return
	}

	// Only allow access to public (shared) packs
	if pack.SharingCode == nil || *pack.SharingCode == "" {
		c.String(http.StatusNotFound, "Pack not found")
		return
	}

	// Check that the item is part of this pack
	inPack, err := packs.CheckItemInPack(ctx, packID, itemID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to verify item in pack")
		return
	}
	if !inPack {
		c.String(http.StatusNotFound, "Item not found in pack")
		return
	}

	// Retrieve image from inventory storage
	img, err := inventoryStorage.Get(ctx, itemID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.String(http.StatusNotFound, "Item has no image")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to retrieve image")
		return
	}

	// Set cache headers for public access
	c.Header("Cache-Control", "public, max-age=86400")

	// Return image binary
	c.Data(http.StatusOK, img.Metadata.MimeType, img.Data)
}
