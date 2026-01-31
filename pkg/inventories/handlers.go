package inventories

import (
	"errors"
	"net/http"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
)

// GetInventories gets all inventories
// @Summary [ADMIN] Get all inventories
// @Description Retrieves a list of all inventories - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce json
// @Success 200 {object} inventories.Inventories "List of Inventories"
// @Failure 404 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/inventories [get]
func GetInventories(c *gin.Context) {
	inventories, err := returnInventories(c.Request.Context())

	if err != nil {
		if errors.Is(err, ErrNoItemFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No Inventory Found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(*inventories) != 0 {
		c.IndentedJSON(http.StatusOK, *inventories)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No inventories empty"})
	}
}

// GetMyInventory gets all inventories of the user
// @Summary Get all inventories of the user
// @Description Retrieves a list of all inventories of the user
// @Security Bearer
// @Tags Inventories
// @Produce json
// @Success 200 {object} inventories.Inventories
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 404 {object} apitypes.ErrorResponse "No Inventory Found"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/myinventory [get]
func GetMyInventory(c *gin.Context) {
	userID, err := security.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	inventories, err := returnInventoriesByUserID(c.Request.Context(), userID)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(*inventories) != 0 {
		c.IndentedJSON(http.StatusOK, *inventories)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No inventories empty"})
	}
}

// GetInventoryByID gets an inventory by ID
// @Summary [ADMIN] Get an inventory by ID
// @Description Retrieves an inventory by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce json
// @Param id path int true "Inventory ID"
// @Success 200 {object} inventories.Inventory
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 404 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/inventories/{id} [get]
func GetInventoryByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	inventory, err := findInventoryByID(c.Request.Context(), id)

	if err != nil {
		if errors.Is(err, ErrNoItemFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if inventory == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, inventory)
}

// GetMyInventoryByID gets an inventory by ID
// @Summary Get an inventory by ID
// @Description Retrieves an inventory by ID
// @Security Bearer
// @Tags Inventories
// @Produce json
// @Param id path int true "Inventory ID"
// @Success 200 {object} inventories.Inventory "Inventory"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} apitypes.ErrorResponse "This item does not belong to you"
// @Failure 404 {object} apitypes.ErrorResponse "Inventory not found"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/myinventory/{id} [get]
func GetMyInventoryByID(c *gin.Context) {
	userID, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	myInventory, err := checkInventoryOwnership(c.Request.Context(), id, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myInventory {
		inventory, err := findInventoryByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, ErrNoItemFound) {
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
				return
			}
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, inventory)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This item does not belong to you"})
	}
}

// PostInventory creates an inventory
// @Summary [ADMIN] Create an inventory
// @Description Creates an inventory - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept json
// @Produce json
// @Param inventory body inventories.Inventory true "Inventory"
// @Success 201 {object} inventories.Inventory
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/inventories [post]
func PostInventory(c *gin.Context) {
	var newInventory Inventory

	if err := c.BindJSON(&newInventory); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	err := InsertInventory(c.Request.Context(), &newInventory)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newInventory)
}

// PostMyInventory creates an inventory
// @Summary Create an inventory
// @Description Creates an inventory
// @Security Bearer
// @Tags Inventories
// @Accept json
// @Produce json
// @Param inventory body inventories.Inventory true "Inventory"
// @Success 201 {object} inventories.Inventory "Inventory Updated"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid payload"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/myinventory [post]
func PostMyInventory(c *gin.Context) {
	var newInventory Inventory

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	newInventory.UserID = userID

	err = c.BindJSON(&newInventory)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = InsertInventory(c.Request.Context(), &newInventory)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newInventory)
}

// PutInventoryByID updates an inventory by ID
// @Summary [ADMIN] Update an inventory by ID
// @Description Updates an inventory by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept json
// @Produce json
// @Param id path int true "Inventory ID"
// @Param inventory body inventories.Inventory true "Inventory"
// @Success 200 {object} inventories.Inventory
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/inventories/{id} [put]
func PutInventoryByID(c *gin.Context) {
	var updatedInventory Inventory

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := c.BindJSON(&updatedInventory); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedInventory.ID = id
	updatedInventory.UpdatedAt = time.Now().Truncate(time.Second)

	err = updateInventoryByID(c.Request.Context(), id, &updatedInventory)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, updatedInventory)
}

// PutMyInventoryByID updates an inventory by ID
// @Summary Update an inventory by ID
// @Description Updates an inventory by ID
// @Security Bearer
// @Tags Inventories
// @Accept json
// @Produce json
// @Param id path int true "Inventory ID"
// @Param inventory body inventories.Inventory true "Inventory"
// @Success 200 {object} inventories.Inventory "Inventory Updated"
// @Failure 400 {object} map[string]interface{} "Invalid ID format"
// @Failure 400 {object} map[string]interface{} "Invalid payload"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "This item does not belong to you"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /v1/myinventory/{id} [put]
func PutMyInventoryByID(c *gin.Context) {
	var updatedInventory Inventory

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	myInventory, err := checkInventoryOwnership(c.Request.Context(), id, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myInventory {
		updatedInventory.UserID = userID
		if err := c.BindJSON(&updatedInventory); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updatedInventory.UpdatedAt = time.Now().Truncate(time.Second)
		err = updateInventoryByID(c.Request.Context(), id, &updatedInventory)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, updatedInventory)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This item does not belong to you"})
	}
}

// DeleteInventoryByID deletes an inventory by ID
// @Summary [ADMIN] Delete an inventory by ID
// @Description Deletes an inventory by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce json
// @Param id path int true "Inventory ID"
// @Success 200 {object} apitypes.OkResponse "Inventory deleted"
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/inventories/{id} [delete]
func DeleteInventoryByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = deleteInventoryByID(c.Request.Context(), id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Inventory deleted"})
}

// DeleteMyInventoryByID deletes an inventory by ID
// @Summary Delete an inventory by ID
// @Description Deletes an inventory by ID
// @Security Bearer
// @Tags Inventories
// @Produce json
// @Param id path int true "Inventory ID"
// @Success 200 {object} apitypes.OkResponse "Inventory deleted"
// @Failure 400 {object} apitypes.ErrorResponse "Invalid ID format"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} apitypes.ErrorResponse "This item does not belong to you"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/myinventory/{id} [delete]
func DeleteMyInventoryByID(c *gin.Context) {
	userID, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	myInventory, err := checkInventoryOwnership(c.Request.Context(), id, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myInventory {
		err = deleteInventoryByID(c.Request.Context(), id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Inventory deleted"})
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This item does not belong to you"})
	}
}
