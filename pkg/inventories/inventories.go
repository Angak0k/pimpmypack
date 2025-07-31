package inventories

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/dataset"
	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
)

// ErrNoItemFound is returned when no item is found for a given request.
var ErrNoItemFound = errors.New("no item found")

// GetInventories gets all inventories
// @Summary [ADMIN] Get all inventories
// @Description Retrieves a list of all inventories -  for admin use only
// @Security Bearer
// @Tags Internal
// @Produce json
// @Success 200 {object} dataset.Inventory "List of Inventories"
// @Failure 404 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
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

func returnInventories(ctx context.Context) (*dataset.Inventories, error) {
	var inventories dataset.Inventories

	rows, err := database.DB().QueryContext(ctx,
		`SELECT id, 
			user_id, 
			item_name, 
			category, 
			description, 
			weight, 
			url, 
			price, 
			currency, 
			created_at, 
			updated_at 
		FROM inventory;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoItemFound
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var inventory dataset.Inventory
		err := rows.Scan(
			&inventory.ID,
			&inventory.UserID,
			&inventory.ItemName,
			&inventory.Category,
			&inventory.Description,
			&inventory.Weight,
			&inventory.URL,
			&inventory.Price,
			&inventory.Currency,
			&inventory.CreatedAt,
			&inventory.UpdatedAt)
		if err != nil {
			return nil, err
		}
		inventories = append(inventories, inventory)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &inventories, nil
}

// GetMyInventories gets all inventories of the user
// @Summary Get all inventories of the user
// @Description Retrieves a list of all inventories of the user
// @Security Bearer
// @Tags Inventories
// @Produce json
// @Success 200 {object} dataset.Inventories
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 404 {object} dataset.ErrorResponse "No Inventory Found"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
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

func returnInventoriesByUserID(ctx context.Context, userID uint) (*dataset.Inventories, error) {
	var inventories dataset.Inventories

	rows, err := database.DB().QueryContext(ctx,
		`SELECT id, 
			user_id, 
			item_name, 
			category, 
			description, 
			weight, 
			url, 
			price, 
			currency, 
			created_at, 
			updated_at 
		FROM inventory WHERE user_id = $1 ORDER BY category;`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var inventory dataset.Inventory
		err := rows.Scan(
			&inventory.ID,
			&inventory.UserID,
			&inventory.ItemName,
			&inventory.Category,
			&inventory.Description,
			&inventory.Weight,
			&inventory.URL,
			&inventory.Price,
			&inventory.Currency,
			&inventory.CreatedAt,
			&inventory.UpdatedAt)
		if err != nil {
			return nil, err
		}
		inventories = append(inventories, inventory)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &inventories, nil
}

// GetInventoryByID gets an inventory by ID
// @Summary [ADMIN] Get an inventory by ID
// @Description Retrieves an inventory by ID -  for admin use only
// @Security Bearer
// @Tags Internal
// @Produce json
// @Param id path int true "Inventory ID"
// @Success 200 {object} dataset.Inventory
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 404 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
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
// @Success 200 {object} dataset.Inventory "Inventory"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 403 {object} dataset.ErrorResponse "This item does not belong to you"
// @Failure 404 {object} dataset.ErrorResponse "Inventory not found"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
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

func findInventoryByID(ctx context.Context, id uint) (*dataset.Inventory, error) {
	var inventory dataset.Inventory

	row := database.DB().QueryRowContext(ctx,
		`SELECT id, 
			user_id, 
			item_name, 
			category, 
			description, 
			weight, 
			url, 
			price, 
			currency, 
			created_at, 
			updated_at 
		FROM inventory WHERE id = $1;`,
		id)
	err := row.Scan(
		&inventory.ID,
		&inventory.UserID,
		&inventory.ItemName,
		&inventory.Category,
		&inventory.Description,
		&inventory.Weight,
		&inventory.URL,
		&inventory.Price,
		&inventory.Currency,
		&inventory.CreatedAt,
		&inventory.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoItemFound
		}
		return nil, err
	}

	return &inventory, nil
}

// PostInventory creates an inventory
// @Summary [ADMIN] Create an inventory
// @Description Creates an inventory -  for admin use only
// @Security Bearer
// @Tags Internal
// @Accept json
// @Produce json
// @Param inventory body dataset.Inventory true "Inventory"
// @Success 201 {object} dataset.Inventory
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/inventories [post]
func PostInventory(c *gin.Context) {
	var newInventory dataset.Inventory

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
// @Param inventory body dataset.Inventory true "Inventory"
// @Success 201 {object} dataset.Inventory "Inventory Updated"
// @Failure 400 {object} dataset.ErrorResponse "Invalid payload"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /v1/myinventory [post]
func PostMyInventory(c *gin.Context) {
	var newInventory dataset.Inventory

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

func InsertInventory(ctx context.Context, i *dataset.Inventory) error {
	if i == nil {
		return errors.New("payload is empty")
	}

	i.CreatedAt = time.Now().Truncate(time.Second)
	i.UpdatedAt = time.Now().Truncate(time.Second)

	//nolint:execinquery
	err := database.DB().QueryRowContext(ctx,
		`INSERT INTO inventory 
		(user_id, item_name, category, description, weight, url, price, currency, created_at, updated_at) 
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) 
		RETURNING id;`,
		i.UserID,
		i.ItemName,
		i.Category,
		i.Description,
		i.Weight,
		i.URL,
		i.Price,
		i.Currency,
		i.CreatedAt,
		i.UpdatedAt).Scan(&i.ID)

	if err != nil {
		return err
	}

	return nil
}

// PutInventoryByID updates an inventory by ID
// @Summary [ADMIN] Update an inventory by ID
// @Description Updates an inventory by ID -  for admin use only
// @Security Bearer
// @Tags Internal
// @Accept json
// @Produce json
// @Param id path int true "Inventory ID"
// @Param inventory body dataset.Inventory true "Inventory"
// @Success 200 {object} dataset.Inventory
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/inventories/{id} [put]
func PutInventoryByID(c *gin.Context) {
	var updatedInventory dataset.Inventory

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
// @Param inventory body dataset.Inventory true "Inventory"
// @Success 200 {object} dataset.Inventory "Inventory Updated"
// @Failure 400 {object} map[string]interface{} "Invalid ID format"
// @Failure 400 {object} map[string]interface{} "Invalid payload"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "This item does not belong to you"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /v1/myinventory/{id} [put]
func PutMyInventoryByID(c *gin.Context) {
	var updatedInventory dataset.Inventory

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

func updateInventoryByID(ctx context.Context, id uint, i *dataset.Inventory) error {
	if i == nil {
		return errors.New("payload is empty")
	}

	i.UpdatedAt = time.Now().Truncate(time.Second)
	statement, err := database.DB().PrepareContext(ctx,
		`UPDATE inventory 
		SET user_id=$1, 
			item_name=$2, 
			category=$3, 
			description=$4, 
			weight=$5, 
			url=$6, 
			price=$7, 
			currency=$8, 
			updated_at=$9 
		WHERE id=$10;`)
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.ExecContext(ctx,
		i.UserID,
		i.ItemName,
		i.Category,
		i.Description,
		i.Weight,
		i.URL,
		i.Price,
		i.Currency,
		i.UpdatedAt,
		id)
	if err != nil {
		return err
	}

	return nil
}

// DeleteInventoryByID deletes an inventory by ID
// @Summary [ADMIN] Delete an inventory by ID
// @Description Deletes an inventory by ID -  for admin use only
// @Security Bearer
// @Tags Internal
// @Produce json
// @Param id path int true "Inventory ID"
// @Success 200 {object} dataset.OkResponse "Inventory deleted"
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
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

func deleteInventoryByID(ctx context.Context, id uint) error {
	statement, err := database.DB().PrepareContext(ctx, "DELETE FROM inventory WHERE id=$1;")
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.ExecContext(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func checkInventoryOwnership(ctx context.Context, id uint, userID uint) (bool, error) {
	var rows int

	row := database.DB().QueryRowContext(ctx,
		"SELECT COUNT(id) FROM inventory WHERE id = $1 AND user_id = $2;", id, userID)
	err := row.Scan(&rows)
	if err != nil {
		return false, err
	}

	if rows == 0 {
		return false, nil
	}

	return true, nil
}
