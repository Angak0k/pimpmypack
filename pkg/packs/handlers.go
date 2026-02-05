package packs

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/security"
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

// Get My packs
// @Summary Get My Packs
// @Description Get my packs
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Success 200 {object} Packs "Packs"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 404 {object} apitypes.ErrorResponse "No pack found"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/mypacks [get]
func GetMyPacks(c *gin.Context) {
	userID, err := security.ExtractTokenID(c)

	if err != nil {
		helper.LogAndSanitize(err, "get my packs: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	packs, err := findPacksByUserID(c.Request.Context(), userID)

	if err != nil {
		helper.LogAndSanitize(err, "get my packs: find packs failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if packs == nil || len(*packs) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No pack found"})
		return
	}

	c.IndentedJSON(http.StatusOK, *packs)
}

// Get My pack by ID
// @Summary Get My pack by ID
// @Description Get pack by ID
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} Pack
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} apitypes.ErrorResponse "This pack does not belong to you"
// @Failure 404 {object} apitypes.ErrorResponse "Pack not found"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id} [get]
func GetMyPackByID(c *gin.Context) {
	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "get my pack by ID: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	myPack, err := CheckPackOwnership(c.Request.Context(), id, userID)
	if err != nil {
		helper.LogAndSanitize(err, "get my pack by ID: check ownership failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if myPack {
		pack, err := FindPackByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, ErrPackNotFound) {
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
				return
			}
			helper.LogAndSanitize(err, "get my pack by ID: find pack failed")
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}

		c.IndentedJSON(http.StatusOK, *pack)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

// Create a new pack
// @Summary Create a new pack
// @Description Create a new pack
// @Security Bearer
// @Tags Packs
// @Accept  json
// @Produce  json
// @Param pack body Pack true "Pack"
// @Success 201 {object} Pack "Pack created"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid Body format"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/mypack [post]
func PostMyPack(c *gin.Context) {
	var newPack Pack

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "post my pack: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	if err := c.BindJSON(&newPack); err != nil {
		helper.LogAndSanitize(err, "post my pack: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	newPack.UserID = userID

	err = insertPack(c.Request.Context(), &newPack)
	if err != nil {
		helper.LogAndSanitize(err, "post my pack: insert pack failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusCreated, newPack)
}

// Update a pack by ID
// @Summary Update a pack by ID
// @Description Update a pack by ID
// @Security Bearer
// @Tags Packs
// @Accept  json
// @Produce  json
// @Param id path int true "Pack ID"
// @Param pack body Pack true "Pack"
// @Success 200 {object} Pack "Pack updated"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid Payload"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} apitypes.ErrorResponse "This pack does not belong to you"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id} [put]
func PutMyPackByID(c *gin.Context) {
	var updatedPack Pack

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := c.BindJSON(&updatedPack); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Payload"})
		return
	}

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "put my pack by ID: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	myPack, err := CheckPackOwnership(c.Request.Context(), id, userID)
	if err != nil {
		helper.LogAndSanitize(err, "put my pack by ID: check ownership failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if myPack {
		updatedPack.UserID = userID
		err = updatePackByID(c.Request.Context(), id, &updatedPack)
		if err != nil {
			helper.LogAndSanitize(err, "put my pack by ID: update pack failed")
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}
		c.IndentedJSON(http.StatusOK, updatedPack)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

// Delete a pack by ID
// @Summary Delete a pack by ID
// @Description Delete a pack by ID
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} apitypes.OkResponse "Pack deleted"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} apitypes.ErrorResponse "This pack does not belong to you"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id} [delete]
func DeleteMyPackByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "delete my pack by ID: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	myPack, err := CheckPackOwnership(c.Request.Context(), id, userID)
	if err != nil {
		helper.LogAndSanitize(err, "delete my pack by ID: check ownership failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if myPack {
		err := deletePackByID(c.Request.Context(), id)
		if err != nil {
			helper.LogAndSanitize(err, "delete my pack by ID: delete pack failed")
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Pack deleted"})
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

// Get all pack contents
// @Summary [ADMIN] Get all pack contents
// @Description Get all pack contents - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Success 200 {object} PackContents
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/packcontents [get]
func GetPackContents(c *gin.Context) {
	packContents, err := returnPackContents(c.Request.Context())
	if err != nil {
		helper.LogAndSanitize(err, "get pack contents: return pack contents failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, packContents)
}

// Get pack content by ID
// @Summary [ADMIN] Get pack content by ID
// @Description Get pack content by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param id path int true "Pack Content ID"
// @Success 200 {object} PackContent
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 404 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/packcontents/{id} [get]
func GetPackContentByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	packcontent, err := findPackContentByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrPackContentNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack Item not found"})
			return
		}
		helper.LogAndSanitize(err, "get pack content by ID: find pack content failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if packcontent == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack Item not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, *packcontent)
}

// Create a new pack content
// @Summary [ADMIN] Create a new pack content
// @Description Create a new pack content - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param packcontent body PackContent true "Pack Content"
// @Success 201 {object} PackContent
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/packcontents [post]
func PostPackContent(c *gin.Context) {
	var newPackContent PackContent

	if err := c.BindJSON(&newPackContent); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body format"})
		return
	}

	err := insertPackContent(c.Request.Context(), &newPackContent)
	if err != nil {
		helper.LogAndSanitize(err, "post pack content: insert pack content failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusCreated, newPackContent)
}

// Create a new pack content
// @Summary Create a new pack content
// @Description Create a new pack content
// @Security Bearer
// @Tags Packs
// @Accept  json
// @Produce  json
// @Param packcontent body PackContent true "Pack Content"
// @Success 201 {object} PackContent
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 401 {object} apitypes.ErrorResponse
// @Failure 403 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /v1/mypack/:id/packcontent [post]
func PostMyPackContent(c *gin.Context) {
	var requestData PackContentRequest
	var newPackContent PackContent

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "post my pack content: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	// Bind the JSON data to our request struct
	if err := c.BindJSON(&requestData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body format"})
		return
	}

	// Map the request data to the PackContent struct
	newPackContent.PackID = id
	newPackContent.ItemID = requestData.InventoryID
	newPackContent.Quantity = requestData.Quantity
	newPackContent.Worn = requestData.Worn
	newPackContent.Consumable = requestData.Consumable

	myPack, err := CheckPackOwnership(c.Request.Context(), id, userID)
	if err != nil {
		helper.LogAndSanitize(err, "post my pack content: check ownership failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if myPack {
		err := insertPackContent(c.Request.Context(), &newPackContent)
		if err != nil {
			helper.LogAndSanitize(err, "post my pack content: insert pack content failed")
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}
		c.IndentedJSON(http.StatusCreated, newPackContent)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

// Update a pack content by ID
// @Summary [ADMIN] Update a pack content by ID
// @Description Update a pack content by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param id path int true "Pack Content ID"
// @Param packcontent body PackContent true "Pack Content"
// @Success 200 {object} PackContent
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/packcontents/{id} [put]
func PutPackContentByID(c *gin.Context) {
	var updatedPackContent PackContent

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := c.BindJSON(&updatedPackContent); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body format"})
		return
	}

	err = updatePackContentByID(c.Request.Context(), id, &updatedPackContent)
	if err != nil {
		helper.LogAndSanitize(err, "put pack content by ID: update pack content failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, updatedPackContent)
}

// Update My pack content ID by Pack ID
// @Summary Update My pack content ID by Pack ID
// @Description Update My pack content ID by Pack ID
// @Security Bearer
// @Tags Packs
// @Accept  json
// @Produce  json
// @Param id path int true "Pack ID"
// @Param item_id path int true "Item ID"
// @Param packcontent body PackContent true "Pack Content"
// @Success 200 {object} PackContent "Pack Content updated"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid Body format"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} apitypes.ErrorResponse "This pack does not belong to you"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id}/packcontent/{item_id} [put]
func PutMyPackContentByID(c *gin.Context) {
	var updatedPackContent PackContent

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	itemID, err := helper.StringToUint(c.Param("item_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "put my pack content by ID: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	if err := c.BindJSON(&updatedPackContent); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body format"})
		return
	}

	myPack, err := CheckPackOwnership(c.Request.Context(), id, userID)
	if err != nil {
		helper.LogAndSanitize(err, "put my pack content by ID: check ownership failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if myPack {
		updatedPackContent.PackID = id
		err := updatePackContentByID(c.Request.Context(), itemID, &updatedPackContent)
		if err != nil {
			helper.LogAndSanitize(err, "put my pack content by ID: update pack content failed")
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}
		c.IndentedJSON(http.StatusOK, updatedPackContent)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

// Delete a pack content by ID
// @Summary [ADMIN] Delete a pack content by ID
// @Description Delete a pack content by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param id path int true "Pack Content ID"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/packcontents/{id} [delete]
func DeletePackContentByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = deletePackContentByID(c.Request.Context(), id)
	if err != nil {
		helper.LogAndSanitize(err, "delete pack content by ID: delete pack content failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Pack Item deleted"})
}

// Delete a pack content ID by Pack ID
// @Summary Delete a pack content by ID
// @Description Delete a pack content by ID
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack ID"
// @Param item_id path int true "Item ID"
// @Success 200 {object} apitypes.OkResponse "Pack Item deleted"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} apitypes.ErrorResponse "This pack does not belong to you"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id}/packcontent/{item_id} [delete]
func DeleteMyPackContentByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Pack ID format"})
		return
	}

	itemID, err := helper.StringToUint(c.Param("item_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Item ID format"})
		return
	}

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "delete my pack content by ID: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	myPack, err := CheckPackOwnership(c.Request.Context(), id, userID)
	if err != nil {
		helper.LogAndSanitize(err, "delete my pack content by ID: check ownership failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if myPack {
		err := deletePackContentByID(c.Request.Context(), itemID)
		if err != nil {
			helper.LogAndSanitize(err, "delete my pack content by ID: delete pack content failed")
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Pack Item deleted"})
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

// Get all pack contents
// @Summary [ADMIN] Get all pack contents
// @Description Get all pack contents - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Success 200 {object} PackContents
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 404 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/packs/:id/packcontents [get]
func GetPackContentsByPackID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	packContents, err := returnPackContentsByPackID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrPackNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
		helper.LogAndSanitize(err, "get pack contents by pack ID: return pack contents failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}
	if packContents == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, packContents)
}

// Get pack content by ID
// @Summary Get pack content by ID
// @Description Get pack content by ID
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack Content ID"
// @Success 200 {object} PackContent "Pack Item"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} apitypes.ErrorResponse "This pack does not belong to you"
// @Failure 404 {object} apitypes.ErrorResponse "Pack not found"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id}/packcontents [get]
func GetMyPackContentsByPackID(c *gin.Context) {
	var packContents *PackContentWithItems

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "get my pack contents by pack ID: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	myPack, err := CheckPackOwnership(c.Request.Context(), id, userID)
	if err != nil {
		helper.LogAndSanitize(err, "get my pack contents by pack ID: check ownership failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if myPack {
		packContents, err = returnPackContentsByPackID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, ErrPackNotFound) {
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
				return
			}
			helper.LogAndSanitize(err, "get my pack contents by pack ID: return pack contents failed")
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}

		c.IndentedJSON(http.StatusOK, packContents)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

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

		lighterPackItem, err = readLineFromCSV(record)
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
