package trails

import (
	"errors"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/gin-gonic/gin"
)

// Get all trails
// @Summary [ADMIN] Get all trails
// @Description Get all trails - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Success 200 {object} Trails
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/trails [get]
func GetTrails(c *gin.Context) {
	trails, err := returnTrails(c.Request.Context())
	if err != nil {
		helper.LogAndSanitize(err, "get trails: return trails failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if len(trails) != 0 {
		c.IndentedJSON(http.StatusOK, trails)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No trails found"})
	}
}

// Get trail by ID
// @Summary [ADMIN] Get trail by ID
// @Description Get trail by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param id path int true "Trail ID"
// @Success 200 {object} Trail
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 404 {object} apitypes.ErrorResponse "Trail not found"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /admin/trails/{id} [get]
func GetTrailByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	trail, err := findTrailByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTrailNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Trail not found"})
			return
		}
		helper.LogAndSanitize(err, "get trail by ID: find trail failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, *trail)
}

// Create a new trail
// @Summary [ADMIN] Create a new trail
// @Description Create a new trail - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param trail body TrailCreateRequest true "Trail"
// @Success 201 {object} Trail
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 409 {object} apitypes.ErrorResponse "Trail name already exists"
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/trails [post]
func PostTrail(c *gin.Context) {
	var input TrailCreateRequest

	if err := c.BindJSON(&input); err != nil {
		helper.LogAndSanitize(err, "post trail: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	newTrail := Trail{
		Name:       input.Name,
		Country:    input.Country,
		Continent:  input.Continent,
		DistanceKm: input.DistanceKm,
		URL:        input.URL,
	}

	err := insertTrail(c.Request.Context(), &newTrail)
	if err != nil {
		if errors.Is(err, ErrTrailNameExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Trail name already exists"})
			return
		}
		helper.LogAndSanitize(err, "post trail: insert trail failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusCreated, newTrail)
}

// Update a trail by ID
// @Summary [ADMIN] Update a trail by ID
// @Description Update a trail by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param id path int true "Trail ID"
// @Param trail body TrailUpdateRequest true "Trail"
// @Success 200 {object} Trail
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 404 {object} apitypes.ErrorResponse "Trail not found"
// @Failure 409 {object} apitypes.ErrorResponse "Trail name already exists"
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/trails/{id} [put]
func PutTrailByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var input TrailUpdateRequest
	if err := c.BindJSON(&input); err != nil {
		helper.LogAndSanitize(err, "put trail: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	existingTrail, err := findTrailByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTrailNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Trail not found"})
			return
		}
		helper.LogAndSanitize(err, "put trail by ID: find trail failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	updatedTrail := *existingTrail
	updatedTrail.Name = input.Name
	updatedTrail.Country = input.Country
	updatedTrail.Continent = input.Continent
	updatedTrail.DistanceKm = input.DistanceKm
	updatedTrail.URL = input.URL

	err = updateTrailByID(c.Request.Context(), id, &updatedTrail)
	if err != nil {
		if errors.Is(err, ErrTrailNameExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Trail name already exists"})
			return
		}
		helper.LogAndSanitize(err, "put trail by ID: update trail failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, updatedTrail)
}

// Delete a trail by ID
// @Summary [ADMIN] Delete a trail by ID
// @Description Delete a trail by ID - for admin use only. Fails if trail is in use by packs.
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param id path int true "Trail ID"
// @Success 200 {object} apitypes.OkResponse "Trail deleted"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 404 {object} apitypes.ErrorResponse "Trail not found"
// @Failure 409 {object} apitypes.ErrorResponse "Trail is in use"
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/trails/{id} [delete]
func DeleteTrailByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = deleteTrailByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTrailNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Trail not found"})
			return
		}
		if errors.Is(err, ErrTrailInUse) {
			c.IndentedJSON(http.StatusConflict, gin.H{"error": "Trail is in use by one or more packs"})
			return
		}
		helper.LogAndSanitize(err, "delete trail by ID: delete trail failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Trail deleted"})
}

// Bulk create trails
// @Summary [ADMIN] Bulk create trails
// @Description Bulk create trails - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param trails body TrailBulkCreateRequest true "Trails"
// @Success 201 {object} Trails
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 409 {object} apitypes.ErrorResponse "Duplicate trail name"
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/trails/bulk [post]
func PostTrailsBulk(c *gin.Context) {
	var input TrailBulkCreateRequest

	if err := c.BindJSON(&input); err != nil {
		helper.LogAndSanitize(err, "post trails bulk: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	trailsToCreate := make([]Trail, len(input.Trails))
	for i, req := range input.Trails {
		trailsToCreate[i] = Trail{
			Name:       req.Name,
			Country:    req.Country,
			Continent:  req.Continent,
			DistanceKm: req.DistanceKm,
			URL:        req.URL,
		}
	}

	created, err := insertTrailsBulk(c.Request.Context(), trailsToCreate)
	if err != nil {
		if errors.Is(err, ErrTrailNameExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "One or more trail names already exist"})
			return
		}
		helper.LogAndSanitize(err, "post trails bulk: insert trails failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusCreated, created)
}

// Bulk delete trails
// @Summary [ADMIN] Bulk delete trails
// @Description Bulk delete trails - for admin use only. Fails if any trail is in use by packs.
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param ids body TrailBulkDeleteRequest true "Trail IDs"
// @Success 200 {object} apitypes.OkResponse "Trails deleted"
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 409 {object} apitypes.ErrorResponse "Trail in use"
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/trails/bulk [delete]
func DeleteTrailsBulk(c *gin.Context) {
	var input TrailBulkDeleteRequest

	if err := c.BindJSON(&input); err != nil {
		helper.LogAndSanitize(err, "delete trails bulk: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	err := deleteTrailsBulk(c.Request.Context(), input.IDs)
	if err != nil {
		if errors.Is(err, ErrTrailInUse) {
			c.JSON(http.StatusConflict, gin.H{"error": "One or more trails are in use by packs"})
			return
		}
		if errors.Is(err, ErrTrailNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "One or more trail IDs not found"})
			return
		}
		helper.LogAndSanitize(err, "delete trails bulk: delete trails failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Trails deleted"})
}
