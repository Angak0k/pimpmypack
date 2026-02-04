package packs

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
)

// ErrPackNotFound is returned when a pack is not found
var ErrPackNotFound = errors.New("pack not found")

// ErrPackContentNotFound is returned when no item are found in a given pack
var ErrPackContentNotFound = errors.New("pack content not found")

// Share a pack by ID
// @Summary Share a pack by ID
// @Description Generate a sharing code for a pack to make it publicly accessible (idempotent)
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} map[string]string "Pack shared successfully"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} apitypes.ErrorResponse "This pack does not belong to you"
// @Failure 404 {object} apitypes.ErrorResponse "Pack not found"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id}/share [post]
func ShareMyPack(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "share my pack: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	// Check if pack exists
	_, err = FindPackByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrPackNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
		helper.LogAndSanitize(err, "share my pack: find pack failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	// Share the pack (idempotent)
	sharingCode, err := sharePackByID(c.Request.Context(), id, userID)
	if err != nil {
		if err.Error() == "pack does not belong to user" {
			c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
			return
		}
		helper.LogAndSanitize(err, "share my pack: share pack failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"message":      "Pack shared successfully",
		"sharing_code": sharingCode,
	})
}

// Unshare a pack by ID
// @Summary Unshare a pack by ID
// @Description Remove the sharing code from a pack to make it private (idempotent)
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} apitypes.OkResponse "Pack unshared successfully"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} apitypes.ErrorResponse "This pack does not belong to you"
// @Failure 404 {object} apitypes.ErrorResponse "Pack not found"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id}/share [delete]
func UnshareMyPack(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "unshare my pack: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	// Check if pack exists
	_, err = FindPackByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrPackNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
		helper.LogAndSanitize(err, "unshare my pack: find pack failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	// Unshare the pack (idempotent)
	err = unsharePackByID(c.Request.Context(), id, userID)
	if err != nil {
		if err.Error() == "pack does not belong to user" {
			c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
			return
		}
		helper.LogAndSanitize(err, "unshare my pack: unshare pack failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Pack unshared successfully"})
}

// Import from lighterpack
// @Summary Import from lighterpack csv pack file
// @Description Import from lighterpack csv pack file
// @Security Bearer
// @Tags Packs
// @Accept  multipart/form-data
// @Produce  json
// @Param file formData file true "CSV file"
// @Success 200 {object} ImportLighterPackResponse "CSV data imported successfully with pack ID"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid CSV format"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/importfromlighterpack [post]
func ImportFromLighterPack(c *gin.Context) {
	var lighterPack LighterPack

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "import from lighterpack: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		helper.LogAndSanitize(err, "import from lighterpack: form file failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}
	defer file.Close()

	// Parse the CSV file
	reader := csv.NewReader(file)
	reader.Comma = ','

	// Read and discard the first line (header) after checking it
	firstRecord, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the CSV header"})
		return
	}
	if len(firstRecord) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format - wrong number of columns"})
		return
	}

	// Iterate through CSV records and process them
	for {
		var lighterPackItem LighterPackItem
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			helper.LogAndSanitize(err, "import from lighterpack: read CSV record failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}

		// Assuming the CSV columns order is: Item Name,Category,desc,qty,weight,unit,url,price,worn,consumable
		if len(record) < 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format - wrong number of columns"})
			return
		}

		lighterPackItem, err = readLineFromCSV(context.Background(), record)
		if err != nil {
			helper.LogAndSanitize(err, "import from lighterpack: read line from CSV failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}

		lighterPack = append(lighterPack, lighterPackItem)
	}

	// Perform database insertion
	packID, err := insertLighterPack(c.Request.Context(), &lighterPack, userID)
	if err != nil {
		helper.LogAndSanitize(err, "import from lighterpack: insert lighterpack failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "CSV data imported successfully",
		"pack_id": packID,
	})
}

// SharedList gets pack metadata and contents for a shared pack
// @Summary Get shared pack with metadata
// @Description Retrieves pack metadata and contents using a sharing code
// @Tags Public
// @Accept json
// @Produce json
// @Param sharing_code path string true "Pack sharing code"
// @Success 200 {object} SharedPackResponse "Shared pack with metadata and contents"
// @Failure 404 {object} apitypes.ErrorResponse "Pack not found or not shared"
// @Failure 500 {object} apitypes.ErrorResponse "Internal server error"
// @Router /sharedlist/{sharing_code} [get]
func SharedList(c *gin.Context) {
	sharingCode := c.Param("sharing_code")

	// Get shared pack with metadata and contents
	sharedPack, err := returnSharedPack(c.Request.Context(), sharingCode)
	if err != nil {
		if errors.Is(err, ErrPackNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
		helper.LogAndSanitize(err, "shared list: return shared pack failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, sharedPack)
}

