package packs

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/dataset"
	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/inventories"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
)

// ErrPackNotFound is returned when a pack is not found
var ErrPackNotFound = errors.New("pack not found")

// ErrPackContentNotFound is returned when no item are found in a given pack
var ErrPackContentNotFound = errors.New("pack content not found")

// Get all packs
// @Summary [ADMIN] Get all packs
// @Description Get all packs - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Success 200 {object} dataset.Packs
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/packs [get]
func GetPacks(c *gin.Context) {
	packs, err := returnPacks()

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(packs) != 0 {
		c.IndentedJSON(http.StatusOK, packs)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No packs founded"})
	}
}

func returnPacks() (dataset.Packs, error) {
	var packs dataset.Packs

	rows, err := database.DB().Query(
		`SELECT p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code, p.created_at, p.updated_at,
		COALESCE(SUM(pc.quantity), 0) as items_count,
		COALESCE(SUM(i.weight * pc.quantity), 0) as total_weight
		FROM pack p
		LEFT JOIN pack_content pc ON p.id = pc.pack_id
		LEFT JOIN inventory i ON pc.item_id = i.id
		GROUP BY p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code, p.created_at, p.updated_at
		ORDER BY p.id;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pack dataset.Pack
		err := rows.Scan(
			&pack.ID,
			&pack.UserID,
			&pack.PackName,
			&pack.PackDescription,
			&pack.SharingCode,
			&pack.CreatedAt,
			&pack.UpdatedAt,
			&pack.PackItemsCount,
			&pack.PackWeight,
		)
		if err != nil {
			return nil, err
		}
		packs = append(packs, pack)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return packs, nil
}

// Get pack by ID
// @Summary [ADMIN] Get pack by ID
// @Description Get pack by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} dataset.Pack
// @Failure 400 {object} dataset.ErrorResponse "Invalid ID format"
// @Failure 404 {object} dataset.ErrorResponse "Pack not found"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /admin/packs/{id} [get]
func GetPackByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	pack, err := findPackByID(id)

	if err != nil {
		if errors.Is(err, ErrPackNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if pack == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, *pack)
}

// Get My pack by ID
// @Summary Get My pack by ID
// @Description Get pack by ID
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} dataset.Pack
// @Failure 400 {object} dataset.ErrorResponse "Invalid ID format"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 403 {object} dataset.ErrorResponse "This pack does not belong to you"
// @Failure 404 {object} dataset.ErrorResponse "Pack not found"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id} [get]
func GetMyPackByID(c *gin.Context) {
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

	myPack, err := checkPackOwnership(id, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		pack, err := findPackByID(id)
		if err != nil {
			if errors.Is(err, ErrPackNotFound) {
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
				return
			}
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, *pack)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func findPackByID(id uint) (*dataset.Pack, error) {
	var pack dataset.Pack

	row := database.DB().QueryRow(
		`SELECT p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code, p.created_at, p.updated_at,
		COALESCE(SUM(pc.quantity), 0) as items_count,
		COALESCE(SUM(i.weight * pc.quantity), 0) as total_weight
		FROM pack p
		LEFT JOIN pack_content pc ON p.id = pc.pack_id
		LEFT JOIN inventory i ON pc.item_id = i.id
		WHERE p.id = $1
		GROUP BY p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code, p.created_at, p.updated_at;`,
		id)
	err := row.Scan(
		&pack.ID,
		&pack.UserID,
		&pack.PackName,
		&pack.PackDescription,
		&pack.SharingCode,
		&pack.CreatedAt,
		&pack.UpdatedAt,
		&pack.PackItemsCount,
		&pack.PackWeight)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPackNotFound
		}
		return nil, err
	}

	return &pack, nil
}

// Create a new pack
// @Summary [ADMIN] Create a new pack
// @Description Create a new pack - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param pack body dataset.Pack true "Pack"
// @Success 201 {object} dataset.Pack
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/packs [post]
func PostPack(c *gin.Context) {
	var newPack dataset.Pack

	if err := c.BindJSON(&newPack); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := insertPack(&newPack)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newPack)
}

// Create a new pack
// @Summary Create a new pack
// @Description Create a new pack
// @Security Bearer
// @Tags Packs
// @Accept  json
// @Produce  json
// @Param pack body dataset.Pack true "Pack"
// @Success 201 {object} dataset.Pack "Pack created"
// @Failure 400 {object} dataset.ErrorResponse "Invalid Body format"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /v1/mypack [post]
func PostMyPack(c *gin.Context) {
	var newPack dataset.Pack

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := c.BindJSON(&newPack); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newPack.UserID = userID

	err = insertPack(&newPack)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newPack)
}

func insertPack(p *dataset.Pack) error {
	var err error
	if p == nil {
		return errors.New("payload is empty")
	}
	p.CreatedAt = time.Now().Truncate(time.Second)
	p.UpdatedAt = time.Now().Truncate(time.Second)
	p.SharingCode, err = helper.GenerateRandomCode(30)
	if err != nil {
		return errors.New("failed to generate a sharing code")
	}

	//nolint:execinquery
	err = database.DB().QueryRow(
		`INSERT INTO pack (user_id, pack_name, pack_description, sharing_code, created_at, updated_at) 
		VALUES ($1,$2,$3,$4,$5,$6) 
		RETURNING id;`,
		p.UserID,
		p.PackName,
		p.PackDescription,
		p.SharingCode,
		p.CreatedAt,
		p.UpdatedAt).Scan(&p.ID)
	if err != nil {
		return err
	}

	return nil
}

// Update a pack by ID
// @Summary [ADMIN] Update a pack by ID
// @Description Update a pack by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param id path int true "Pack ID"
// @Param pack body dataset.Pack true "Pack"
// @Success 200 {object} dataset.Pack
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/packs/{id} [put]
func PutPackByID(c *gin.Context) {
	var updatedPack dataset.Pack
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := c.BindJSON(&updatedPack); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body format"})
		return
	}

	err = updatePackByID(id, &updatedPack)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, updatedPack)
}

// Update a pack by ID
// @Summary Update a pack by ID
// @Description Update a pack by ID
// @Security Bearer
// @Tags Packs
// @Accept  json
// @Produce  json
// @Param id path int true "Pack ID"
// @Param pack body dataset.Pack true "Pack"
// @Success 200 {object} dataset.Pack "Pack updated"
// @Failure 400 {object} dataset.ErrorResponse "Invalid ID format"
// @Failure 400 {object} dataset.ErrorResponse "Invalid Payload"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 403 {object} dataset.ErrorResponse "This pack does not belong to you"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id} [put]
func PutMyPackByID(c *gin.Context) {
	var updatedPack dataset.Pack

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	myPack, err := checkPackOwnership(id, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		updatedPack.UserID = userID
		err = updatePackByID(id, &updatedPack)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, updatedPack)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func updatePackByID(id uint, p *dataset.Pack) error {
	if p == nil {
		return errors.New("payload is empty")
	}
	p.UpdatedAt = time.Now().Truncate(time.Second)

	statement, err := database.DB().Prepare(
		`UPDATE pack SET user_id=$1, pack_name=$2, pack_description=$3, updated_at=$4 
		WHERE id=$5;`)
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(p.UserID, p.PackName, p.PackDescription, p.UpdatedAt, id)
	if err != nil {
		return err
	}

	return nil
}

// Delete a pack by ID
// @Summary [ADMIN] Delete a pack by ID
// @Description Delete a pack by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} dataset.OkResponse
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/packs/{id} [delete]

func DeletePackByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = deletePackByID(id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Pack deleted"})
}

// Delete a pack by ID
// @Summary Delete a pack by ID
// @Description Delete a pack by ID
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} dataset.OkResponse "Pack deleted"
// @Failure 400 {object} dataset.ErrorResponse "Invalid ID format"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 403 {object} dataset.ErrorResponse "This pack does not belong to you"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id} [delete]
func DeleteMyPackByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	myPack, err := checkPackOwnership(id, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		err := deletePackByID(id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Pack deleted"})
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func deletePackByID(id uint) error {
	statement, err := database.DB().Prepare("DELETE FROM pack WHERE id=$1;")
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(id)
	if err != nil {
		return err
	}

	return nil
}

// Get all pack contents
// @Summary [ADMIN] Get all pack contents
// @Description Get all pack contents - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Success 200 {object} dataset.PackContents
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/packcontents [get]
func GetPackContents(c *gin.Context) {
	packContents, err := returnPackContents()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, packContents)
}

func returnPackContents() (*dataset.PackContents, error) {
	var packContents dataset.PackContents

	rows, err := database.DB().Query(
		`SELECT id, pack_id, item_id, quantity, worn, consumable, created_at, updated_at 
		FROM pack_content;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var packContent dataset.PackContent
		err := rows.Scan(
			&packContent.ID,
			&packContent.PackID,
			&packContent.ItemID,
			&packContent.Quantity,
			&packContent.Worn,
			&packContent.Consumable,
			&packContent.CreatedAt,
			&packContent.UpdatedAt)
		if err != nil {
			return nil, err
		}
		packContents = append(packContents, packContent)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &packContents, nil
}

// Get pack content by ID
// @Summary [ADMIN] Get pack content by ID
// @Description Get pack content by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param id path int true "Pack Content ID"
// @Success 200 {object} dataset.PackContent
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 404 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/packcontents/{id} [get]
func GetPackContentByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	packcontent, err := findPackContentByID(id)
	if err != nil {
		if errors.Is(err, ErrPackContentNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack Item not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if packcontent == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack Item not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, *packcontent)
}

func findPackContentByID(id uint) (*dataset.PackContent, error) {
	var packcontent dataset.PackContent

	row := database.DB().QueryRow(
		`SELECT id, pack_id, item_id, quantity, worn, consumable, created_at, updated_at 
		FROM pack_content 
		WHERE id = $1;`,
		id)
	err := row.Scan(
		&packcontent.ID,
		&packcontent.PackID,
		&packcontent.ItemID,
		&packcontent.Quantity,
		&packcontent.Worn,
		&packcontent.Consumable,
		&packcontent.CreatedAt,
		&packcontent.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPackContentNotFound
		}
		return nil, err
	}

	return &packcontent, nil
}

// Create a new pack content
// @Summary [ADMIN] Create a new pack content
// @Description Create a new pack content - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param packcontent body dataset.PackContent true "Pack Content"
// @Success 201 {object} dataset.PackContent
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/packcontents [post]
func PostPackContent(c *gin.Context) {
	var newPackContent dataset.PackContent

	if err := c.BindJSON(&newPackContent); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body format"})
		return
	}

	err := insertPackContent(&newPackContent)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
// @Param packcontent body dataset.PackContent true "Pack Content"
// @Success 201 {object} dataset.PackContent
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 401 {object} dataset.ErrorResponse
// @Failure 403 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /mypack/:id/packcontent [post]
func PostMyPackContent(c *gin.Context) {
	var requestData dataset.PackContentRequest
	var newPackContent dataset.PackContent

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
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

	myPack, err := checkPackOwnership(id, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		err := insertPackContent(&newPackContent)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusCreated, newPackContent)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func insertPackContent(pc *dataset.PackContent) error {
	if pc == nil {
		return errors.New("payload is empty")
	}
	pc.CreatedAt = time.Now().Truncate(time.Second)
	pc.UpdatedAt = time.Now().Truncate(time.Second)

	//nolint:execinquery
	err := database.DB().QueryRow(`
		INSERT INTO pack_content 
		(pack_id, item_id, quantity, worn, consumable, created_at, updated_at) 
		VALUES ($1,$2,$3,$4,$5,$6,$7) 
		RETURNING id;`,
		pc.PackID,
		pc.ItemID,
		pc.Quantity,
		pc.Worn,
		pc.Consumable,
		pc.CreatedAt,
		pc.UpdatedAt).Scan(&pc.ID)

	if err != nil {
		return err
	}
	return nil
}

// Update a pack content by ID
// @Summary [ADMIN] Update a pack content by ID
// @Description Update a pack content by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param id path int true "Pack Content ID"
// @Param packcontent body dataset.PackContent true "Pack Content"
// @Success 200 {object} dataset.PackContent
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/packcontents/{id} [put]
func PutPackContentByID(c *gin.Context) {
	var updatedPackContent dataset.PackContent

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := c.BindJSON(&updatedPackContent); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body format"})
		return
	}

	err = updatePackContentByID(id, &updatedPackContent)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
// @Param packcontent body dataset.PackContent true "Pack Content"
// @Success 200 {object} dataset.PackContent "Pack Content updated"
// @Failure 400 {object} dataset.ErrorResponse "Invalid ID format"
// @Failure 400 {object} dataset.ErrorResponse "Invalid Body format"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 403 {object} dataset.ErrorResponse "This pack does not belong to you"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id}/packcontent/{item_id} [put]
func PutMyPackContentByID(c *gin.Context) {
	var updatedPackContent dataset.PackContent

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := c.BindJSON(&updatedPackContent); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body format"})
		return
	}

	myPack, err := checkPackOwnership(id, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		updatedPackContent.PackID = id
		err := updatePackContentByID(itemID, &updatedPackContent)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, updatedPackContent)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func updatePackContentByID(id uint, pc *dataset.PackContent) error {
	if pc == nil {
		return errors.New("payload is empty")
	}
	pc.UpdatedAt = time.Now().Truncate(time.Second)

	statement, err := database.DB().Prepare(`
		UPDATE pack_content 
		SET pack_id=$1, item_id=$2, quantity=$3, worn=$4, consumable=$5, updated_at=$6 
		WHERE id=$7;`)
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(pc.PackID, pc.ItemID, pc.Quantity, pc.Worn, pc.Consumable, pc.UpdatedAt, id)
	if err != nil {
		return err
	}
	return nil
}

// Delete a pack content by ID
// @Summary [ADMIN] Delete a pack content by ID
// @Description Delete a pack content by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param id path int true "Pack Content ID"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/packcontents/{id} [delete]
func DeletePackContentByID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = deletePackContentByID(id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
// @Success 200 {object} dataset.OkResponse "Pack Item deleted"
// @Failure 400 {object} dataset.ErrorResponse "Invalid ID format"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 403 {object} dataset.ErrorResponse "This pack does not belong to you"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	myPack, err := checkPackOwnership(id, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		err := deletePackContentByID(itemID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Pack Item deleted"})
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func deletePackContentByID(id uint) error {
	statement, err := database.DB().Prepare("DELETE FROM pack_content WHERE id=$1;")
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

// Get all pack contents
// @Summary [ADMIN] Get all pack contents
// @Description Get all pack contents - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Success 200 {object} dataset.PackContents
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 404 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/packs/:id/packcontents [get]
func GetPackContentsByPackID(c *gin.Context) {
	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	packContents, err := returnPackContentsByPackID(id)
	if err != nil {
		if errors.Is(err, ErrPackNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
// @Success 200 {object} dataset.PackContent "Pack Item"
// @Failure 400 {object} dataset.ErrorResponse "Invalid ID format"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 403 {object} dataset.ErrorResponse "This pack does not belong to you"
// @Failure 404 {object} dataset.ErrorResponse "Pack not found"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /v1/mypack/{id}/packcontents [get]
func GetMyPackContentsByPackID(c *gin.Context) {
	var packContents *dataset.PackContentWithItems

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	myPack, err := checkPackOwnership(id, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		packContents, err = returnPackContentsByPackID(id)
		if err != nil {
			if errors.Is(err, ErrPackNotFound) {
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
				return
			}
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, packContents)
	} else {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func returnPackContentsByPackID(id uint) (*dataset.PackContentWithItems, error) {
	// First, check if the pack exists in the database
	packExists, err := checkPackExists(id)
	if err != nil {
		return nil, err
	}
	if !packExists {
		return nil, ErrPackNotFound
	}

	// Pack exists, continue with fetching its contents
	var packWithItems dataset.PackContentWithItems

	rows, err := database.DB().Query(
		`SELECT pc.id AS pack_content_id, 
			pc.pack_id as pack_id, 
			i.id AS inventory_id, 
			i.item_name, 
			i.category,
			i.description AS item_description, 
			i.weight, 
			i.url AS item_url, 
			i.price, 
			i.currency, 
			pc.quantity, 
			pc.worn, 
			pc.consumable 
			FROM pack_content pc JOIN inventory i ON pc.item_id = i.id 
			WHERE pc.pack_id = $1;`,
		id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item dataset.PackContentWithItem
		err := rows.Scan(
			&item.PackContentID,
			&item.PackID,
			&item.InventoryID,
			&item.ItemName,
			&item.Category,
			&item.ItemDescription,
			&item.Weight,
			&item.ItemURL,
			&item.Price,
			&item.Currency,
			&item.Quantity,
			&item.Worn,
			&item.Consumable)
		if err != nil {
			return nil, err
		}
		packWithItems = append(packWithItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &packWithItems, nil
}

// Helper function to check if a pack exists
func checkPackExists(id uint) (bool, error) {
	var count int
	err := database.DB().QueryRow("SELECT COUNT(*) FROM pack WHERE id = $1", id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Get My packs
// @Summary Get My Packs
// @Description Get my packs
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Success 200 {object} dataset.Packs "Packs"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 404 {object} dataset.ErrorResponse "No pack found"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /v1/mypacks [get]
func GetMyPacks(c *gin.Context) {
	userID, err := security.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	packs, err := findPacksByUserID(userID)

	if err != nil {
		if errors.Is(err, ErrPackContentNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No pack found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if packs == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No pack found"})
		return
	}

	c.IndentedJSON(http.StatusOK, *packs)
}

func findPacksByUserID(id uint) (*dataset.Packs, error) {
	var packs dataset.Packs

	rows, err := database.DB().Query(`
		SELECT p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code, p.created_at, p.updated_at,
		COALESCE(SUM(pc.quantity), 0) as items_count,
		COALESCE(SUM(i.weight * pc.quantity), 0) as total_weight
		FROM pack p
		LEFT JOIN pack_content pc ON p.id = pc.pack_id
		LEFT JOIN inventory i ON pc.item_id = i.id
		WHERE p.user_id = $1
		GROUP BY p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code, p.created_at, p.updated_at;`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pack dataset.Pack
		err := rows.Scan(
			&pack.ID,
			&pack.UserID,
			&pack.PackName,
			&pack.PackDescription,
			&pack.SharingCode,
			&pack.CreatedAt,
			&pack.UpdatedAt,
			&pack.PackItemsCount,
			&pack.PackWeight)
		if err != nil {
			return nil, err
		}
		packs = append(packs, pack)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(packs) == 0 {
		return nil, ErrPackContentNotFound
	}

	return &packs, nil
}

func checkPackOwnership(id uint, userID uint) (bool, error) {
	var rows int

	row := database.DB().QueryRow("SELECT COUNT(id) FROM pack WHERE id = $1 AND user_id = $2;", id, userID)
	err := row.Scan(&rows)
	if err != nil {
		return false, err
	}

	if rows == 0 {
		return false, nil
	}

	return true, nil
}

// Import from lighterpack
// @Summary Import from lighterpack csv pack file
// @Description Import from lighterpack csv pack file
// @Security Bearer
// @Tags Packs
// @Accept  multipart/form-data
// @Produce  json
// @Param file formData file true "CSV file"
// @Success 200 {object} dataset.OkResponse "CSV data imported successfully"
// @Failure 400 {object} dataset.ErrorResponse "Invalid CSV format"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /v1/importfromlighterpack [post]
func ImportFromLighterPack(c *gin.Context) {
	var lighterPack dataset.LighterPack

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		var lighterPackItem dataset.LighterPackItem
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Assuming the CSV columns order is: Item Name,Category,desc,qty,weight,unit,url,price,worn,consumable
		if len(record) < 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format - wrong number of columns"})
			return
		}

		lighterPackItem, err = readLineFromCSV(record)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		lighterPack = append(lighterPack, lighterPackItem)
	}

	// Perform database insertion
	err = insertLighterPack(&lighterPack, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "CSV data imported successfully"})
}

// Take a record from csv.Newreder and return a LighterPackItem
func readLineFromCSV(record []string) (dataset.LighterPackItem, error) {
	var lighterPackItem dataset.LighterPackItem

	lighterPackItem.ItemName = record[0]
	lighterPackItem.Category = record[1]
	lighterPackItem.Desc = record[2]
	lighterPackItem.Unit = record[5]
	lighterPackItem.URL = record[6]

	qty, err := strconv.Atoi(record[3])
	if err != nil {
		return lighterPackItem, errors.New("invalid CSV format - failed to convert qty to number")
	}

	lighterPackItem.Qty = qty

	weight, err := strconv.Atoi(record[4])
	if err != nil {
		return lighterPackItem, errors.New("invalid CSV format - failed to convert weight to number")
	}

	lighterPackItem.Weight = weight

	price, err := strconv.Atoi(record[7])
	if err != nil {
		return lighterPackItem, errors.New("invalid CSV format - failed to convert price to number")
	}

	lighterPackItem.Price = price

	if record[8] == "worn" {
		lighterPackItem.Worn = true
	}

	if record[9] == "consumable" {
		lighterPackItem.Consumable = true
	}

	return lighterPackItem, nil
}

func insertLighterPack(lp *dataset.LighterPack, userID uint) error {
	if lp == nil {
		return errors.New("payload is empty")
	}

	// Create new pack
	var newPack dataset.Pack
	newPack.UserID = userID
	newPack.PackName = "LighterPack Import"
	newPack.PackDescription = "LighterPack Import"
	err := insertPack(&newPack)
	if err != nil {
		return err
	}

	// Insert content in new pack with insertPackContent
	for _, item := range *lp {
		var i dataset.Inventory
		i.UserID = userID
		i.ItemName = item.ItemName
		i.Category = item.Category
		i.Description = item.Desc
		i.Weight = item.Weight
		i.URL = item.URL
		i.Price = item.Price
		i.Currency = "USD"
		err := inventories.InsertInventory(&i)
		if err != nil {
			return err
		}
		var pc dataset.PackContent
		pc.PackID = newPack.ID
		pc.ItemID = i.ID
		pc.Quantity = item.Qty
		pc.Worn = item.Worn
		pc.Consumable = item.Consumable
		err = insertPackContent(&pc)
		if err != nil {
			return err
		}
	}
	return nil
}

// Get pack content for a given sharing code
// @Summary Get pack content for a given sharing code
// @Description Get pack content for a given sharing code
// @Tags Public
// @Produce  json
// @Param sharing_code path string true "Sharing Code"
// @Success 200 {object} dataset.PackContents "Pack Contents"
// @Failure 404 {object} dataset.ErrorResponse "Pack not found"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /public/packs/{sharing_code} [get]
func SharedList(c *gin.Context) {
	sharingCode := c.Param("sharing_code")

	packID, err := findPackIDBySharingCode(sharingCode)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if packID == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
		return
	}

	packContents, err := returnPackContentsByPackID(packID)
	if err != nil {
		if errors.Is(err, ErrPackContentNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if packContents == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, packContents)
}

func findPackIDBySharingCode(sharingCode string) (uint, error) {
	var packID uint
	row := database.DB().QueryRow("SELECT id FROM pack WHERE sharing_code = $1;", sharingCode)
	err := row.Scan(&packID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return packID, nil
}
