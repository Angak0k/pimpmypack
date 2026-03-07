package images

import (
	"errors"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
)

var bannerStorage BannerImageStorage = NewDBBannerImageStorage()

// UploadMyBannerImage uploads or updates the banner image for the current user
// @Summary Upload or update banner image
// @Description Upload or update the banner/hero image for the currently logged-in user
// @Security Bearer
// @Tags Account Images
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file (JPEG, PNG, or WebP, max 5MB)"
// @Success 200 {object} map[string]interface{} "Banner image uploaded successfully"
// @Failure 400 {object} map[string]string "Invalid request or image"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 413 {object} map[string]string "Payload too large"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/myaccount/banner [post]
func UploadMyBannerImage(c *gin.Context) {
	uploadUserImage(c, bannerStorage, ProcessBannerImageFromReader,
		"upload banner image", "Banner image uploaded successfully")
}

// DeleteMyBannerImage deletes the banner image for the current user
// @Summary Delete banner image
// @Description Delete the banner/hero image for the currently logged-in user
// @Security Bearer
// @Tags Account Images
// @Produce json
// @Success 200 {object} map[string]interface{} "Banner image deleted successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/myaccount/banner [delete]
func DeleteMyBannerImage(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "delete banner image: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	err = bannerStorage.Delete(ctx, userID)
	if err != nil {
		helper.LogAndSanitize(err, "delete banner image: delete image failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Banner image deleted successfully",
	})
}

// GetBannerImage retrieves the banner image for a given account
// @Summary Get banner image
// @Description Get the banner/hero image for a user account
// @Tags Account Images
// @Produce image/jpeg
// @Param id path int true "Account ID"
// @Success 200 {file} image/jpeg "Banner image"
// @Failure 400 {string} string "Invalid account ID"
// @Failure 404 {string} string "No banner image found"
// @Failure 500 {string} string "Internal server error"
// @Router /v1/accounts/{id}/banner [get]
func GetBannerImage(c *gin.Context) {
	ctx := c.Request.Context()

	accountID, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid account ID format")
		return
	}

	img, err := bannerStorage.Get(ctx, accountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.String(http.StatusNotFound, "No banner image found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to retrieve image")
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, img.Metadata.MimeType, img.Data)
}
