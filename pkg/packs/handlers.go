package packs

import (
	"errors"
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
		if errors.Is(err, ErrPackContentNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No pack found"})
			return
		}
		helper.LogAndSanitize(err, "get my packs: find packs failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if packs == nil {
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
