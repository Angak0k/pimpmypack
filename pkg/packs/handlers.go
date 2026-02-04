package packs

import (
	"errors"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/gin-gonic/gin"
)

// Get all packs
// @Summary [ADMIN] Get all packs
// @Description Get all packs - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Success 200 {object} Packs
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/packs [get]
func GetPacks(c *gin.Context) {
	packs, err := returnPacks(c.Request.Context())

	if err != nil {
		helper.LogAndSanitize(err, "get packs: return packs failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if len(packs) != 0 {
		c.IndentedJSON(http.StatusOK, packs)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No packs founded"})
	}
}

// Get pack by ID
// @Summary [ADMIN] Get pack by ID
// @Description Get pack by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} Pack
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 404 {object} apitypes.ErrorResponse "Pack not found"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /admin/packs/{id} [get]
func GetPackByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	pack, err := FindPackByID(c.Request.Context(), id)

	if err != nil {
		if errors.Is(err, ErrPackNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
		helper.LogAndSanitize(err, "get pack by ID: find pack failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if pack == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, *pack)
}

// Create a new pack
// @Summary [ADMIN] Create a new pack
// @Description Create a new pack - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param pack body Pack true "Pack"
// @Success 201 {object} Pack
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/packs [post]
func PostPack(c *gin.Context) {
	var newPack Pack

	if err := c.BindJSON(&newPack); err != nil {
		helper.LogAndSanitize(err, "post pack: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	err := insertPack(c.Request.Context(), &newPack)
	if err != nil {
		helper.LogAndSanitize(err, "post pack: insert pack failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusCreated, newPack)
}

// Update a pack by ID
// @Summary [ADMIN] Update a pack by ID
// @Description Update a pack by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param id path int true "Pack ID"
// @Param pack body Pack true "Pack"
// @Success 200 {object} Pack
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/packs/{id} [put]
func PutPackByID(c *gin.Context) {
	var updatedPack Pack
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := c.BindJSON(&updatedPack); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body format"})
		return
	}

	err = updatePackByID(c.Request.Context(), id, &updatedPack)
	if err != nil {
		helper.LogAndSanitize(err, "put pack by ID: update pack failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, updatedPack)
}

// Delete a pack by ID
// @Summary [ADMIN] Delete a pack by ID
// @Description Delete a pack by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} apitypes.OkResponse
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/packs/{id} [delete]

func DeletePackByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = deletePackByID(c.Request.Context(), id)
	if err != nil {
		helper.LogAndSanitize(err, "delete pack by ID: delete pack failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Pack deleted"})
}
