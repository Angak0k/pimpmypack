package inventories

import (
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

// GetInventories gets all inventories
// @Summary [ADMIN] Get all inventories
// @Description Retrieves a list of all inventories -  for admin use only
// @Security Bearer
// @Tags Internal
// @Produce json
// @Success 200 {object} dataset.Inventory "List of Inventories"
// @Failure 404 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /inventories [get]
func GetInventories(c *gin.Context) {

	inventories, err := returnInventories()

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

func returnInventories() (*dataset.Inventories, error) {
	var inventories dataset.Inventories

	rows, err := database.Db().Query("SELECT id, user_id, item_name, category, description, weight, weight_unit, url, price, currency, created_at, updated_at FROM inventory;")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var inventory dataset.Inventory
		err := rows.Scan(&inventory.ID, &inventory.User_id, &inventory.Item_name, &inventory.Category, &inventory.Description, &inventory.Weight, &inventory.Weight_unit, &inventory.Url, &inventory.Price, &inventory.Currency, &inventory.Created_at, &inventory.Updated_at)
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
// @Router /myinventory [get]
func GetMyInventory(c *gin.Context) {

	user_id, err := security.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	inventories, err := returnInventoriesByUserID(user_id)

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

func returnInventoriesByUserID(user_id uint) (*dataset.Inventories, error) {
	var inventories dataset.Inventories

	rows, err := database.Db().Query("SELECT id, user_id, item_name, category, description, weight, weight_unit, url, price, currency, created_at, updated_at FROM inventory WHERE user_id = $1 ORDER BY id;", user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var inventory dataset.Inventory
		err := rows.Scan(&inventory.ID, &inventory.User_id, &inventory.Item_name, &inventory.Category, &inventory.Description, &inventory.Weight, &inventory.Weight_unit, &inventory.Url, &inventory.Price, &inventory.Currency, &inventory.Created_at, &inventory.Updated_at)
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
// @Router /inventories/{id} [get]
func GetInventoryByID(c *gin.Context) {

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	inventory, err := findInventoryById(id)

	if err != nil {
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
// @Router /myinventory/{id} [get]
func GetMyInventoryByID(c *gin.Context) {

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	myInventory, err := checkInventoryOwnership(id, user_id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myInventory {
		inventory, err := findInventoryById(id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.IndentedJSON(http.StatusOK, inventory)
		}
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This item does not belong to you"})
	}
}

func findInventoryById(id uint) (*dataset.Inventory, error) {
	var inventory dataset.Inventory

	row := database.Db().QueryRow("SELECT id, user_id, item_name, category, description, weight, weight_unit, url, price, currency, created_at, updated_at FROM inventory WHERE id = $1;", id)
	err := row.Scan(&inventory.ID, &inventory.User_id, &inventory.Item_name, &inventory.Category, &inventory.Description, &inventory.Weight, &inventory.Weight_unit, &inventory.Url, &inventory.Price, &inventory.Currency, &inventory.Created_at, &inventory.Updated_at)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
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
// @Router /inventories [post]
func PostInventory(c *gin.Context) {
	var newInventory dataset.Inventory

	if err := c.BindJSON(&newInventory); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	err := InsertInventory(&newInventory)
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
// @Router /myinventory [post]
func PostMyInventory(c *gin.Context) {
	var newInventory dataset.Inventory

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	newInventory.User_id = user_id

	err = c.BindJSON(&newInventory)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = InsertInventory(&newInventory)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newInventory)

}

func InsertInventory(i *dataset.Inventory) error {

	if i == nil {
		return errors.New("payload is empty")
	}

	i.Created_at = time.Now().Truncate(time.Second)
	i.Updated_at = time.Now().Truncate(time.Second)

	err := database.Db().QueryRow("INSERT INTO inventory (user_id, item_name, category, description, weight, weight_unit, url, price, currency, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id;", i.User_id, i.Item_name, i.Category, i.Description, i.Weight, i.Weight_unit, i.Url, i.Price, i.Currency, i.Created_at, i.Updated_at).Scan(&i.ID)

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
// @Router /inventories/{id} [put]
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

	err = updateInventoryById(id, &updatedInventory)
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
// @Router /myinventory/{id} [put]
func PutMyInventoryByID(c *gin.Context) {
	var updatedInventory dataset.Inventory

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	myInventory, err := checkInventoryOwnership(id, user_id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myInventory {
		updatedInventory.User_id = user_id
		if err := c.BindJSON(&updatedInventory); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = updateInventoryById(id, &updatedInventory)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.IndentedJSON(http.StatusOK, updatedInventory)
		}
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This item does not belong to you"})
	}
}

func updateInventoryById(id uint, i *dataset.Inventory) error {

	if i == nil {
		return errors.New("payload is empty")
	}

	i.Updated_at = time.Now().Truncate(time.Second)
	statement, err := database.Db().Prepare("UPDATE inventory SET user_id=$1, item_name=$2, category=$3, description=$4, weight=$5, weight_unit=$6, url=$7, price=$8, currency=$9, updated_at=$10 WHERE id=$11;")
	if err != nil {
		return err
	}
	_, err = statement.Exec(i.User_id, i.Item_name, i.Category, i.Description, i.Weight, i.Weight_unit, i.Url, i.Price, i.Currency, i.Updated_at, id)
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
// @Router /inventories/{id} [delete]
func DeleteInventoryByID(c *gin.Context) {

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = deleteInventoryById(id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Inventory deleted"})

}

func DeleteMyInventoryByID(c *gin.Context) {

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	myInventory, err := checkInventoryOwnership(id, user_id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myInventory {
		err = deleteInventoryById(id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Inventory deleted"})
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This item does not belong to you"})
	}
}

func deleteInventoryById(id uint) error {
	statement, err := database.Db().Prepare("DELETE FROM inventory WHERE id=$1;")
	if err != nil {
		return err
	}
	_, err = statement.Exec(id)
	if err != nil {
		return err
	}

	return nil
}

func checkInventoryOwnership(id uint, user_id uint) (bool, error) {
	var rows int

	row := database.Db().QueryRow("SELECT COUNT(id) FROM inventory WHERE id = $1 AND user_id = $2;", id, user_id)
	err := row.Scan(&rows)
	if err != nil {
		return false, err
	}

	if rows == 0 {
		return false, nil
	} else {
		return true, nil
	}
}
