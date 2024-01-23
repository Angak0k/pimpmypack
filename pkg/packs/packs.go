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

// Get all packs
// @Summary [ADMIN] Get all packs
// @Description Get all packs - for admin use only
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Success 200 {object} dataset.Packs
// @Failure 500 {object} map[string]interface{} "error"
// @Router /packs [get]
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

	rows, err := database.Db().Query("SELECT id, user_id, pack_name, pack_description, created_at, updated_at FROM pack;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	for rows.Next() {
		var pack dataset.Pack
		err := rows.Scan(&pack.ID, &pack.User_id, &pack.Pack_name, &pack.Pack_description, &pack.Created_at, &pack.Updated_at)
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
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} dataset.Pack
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 404 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /packs/{id} [get]
func GetPackByID(c *gin.Context) {

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	pack, err := findPackById(id)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if pack == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, *pack)

}

// Get pack by ID
// @Summary Get pack by ID
// @Description Get pack by ID
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} dataset.Pack
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 404 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /mypack/{id} [get]
func GetMyPackByID(c *gin.Context) {

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

	myPack, err := checkPackOwnership(id, user_id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		pack, err := findPackById(id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if pack != nil {
			c.IndentedJSON(http.StatusOK, *pack)
		} else {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func findPackById(id uint) (*dataset.Pack, error) {
	var pack dataset.Pack

	row := database.Db().QueryRow("SELECT id, user_id, pack_name, pack_description, created_at, updated_at FROM pack WHERE id = $1;", id)
	err := row.Scan(&pack.ID, &pack.User_id, &pack.Pack_name, &pack.Pack_description, &pack.Created_at, &pack.Updated_at)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &pack, nil
}

// Create a new pack
// @Summary [ADMIN] Create a new pack
// @Description Create a new pack - for admin use only
// @Security Bearer
// @Tags Packs
// @Accept  json
// @Produce  json
// @Param pack body dataset.Pack true "Pack"
// @Success 201 {object} dataset.Pack
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /packs [post]
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
// @Success 201 {object} dataset.Pack
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /mypack [post]
func PostMyPack(c *gin.Context) {
	var newPack dataset.Pack

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.BindJSON(&newPack); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newPack.User_id = user_id

	err = insertPack(&newPack)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newPack)

}

func insertPack(p *dataset.Pack) error {
	if p == nil {
		return errors.New("payload is empty")
	}
	p.Created_at = time.Now().Truncate(time.Second)
	p.Updated_at = time.Now().Truncate(time.Second)

	err := database.Db().QueryRow("INSERT INTO pack (user_id, pack_name, pack_description, created_at, updated_at) VALUES ($1,$2,$3,$4,$5) RETURNING id;", p.User_id, p.Pack_name, p.Pack_description, p.Created_at, p.Updated_at).Scan(&p.ID)
	if err != nil {
		return err
	}

	return nil

}

// Update a pack by ID
// @Summary [ADMIN] Update a pack by ID
// @Description Update a pack by ID - for admin use only
// @Security Bearer
// @Tags Packs
// @Accept  json
// @Produce  json
// @Param id path int true "Pack ID"
// @Param pack body dataset.Pack true "Pack"
// @Success 200 {object} dataset.Pack
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /packs/{id} [put]
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

	err = updatePackById(id, &updatedPack)
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
// @Success 200 {object} dataset.Pack
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /mypack/{id} [put]
func PutMyPackByID(c *gin.Context) {
	var updatedPack dataset.Pack

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := c.BindJSON(&updatedPack); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Json body"})
		return
	}

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	myPack, err := checkPackOwnership(id, user_id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		updatedPack.User_id = user_id
		err = updatePackById(id, &updatedPack)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, updatedPack)
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func updatePackById(id uint, p *dataset.Pack) error {
	if p == nil {
		return errors.New("payload is empty")
	}
	p.Updated_at = time.Now().Truncate(time.Second)

	statement, err := database.Db().Prepare("UPDATE pack SET user_id=$1, pack_name=$2, pack_description=$3, updated_at=$4 WHERE id=$5;")
	if err != nil {
		return err
	}
	_, err = statement.Exec(p.User_id, p.Pack_name, p.Pack_description, p.Updated_at, id)
	if err != nil {
		return err
	}

	return nil

}

// Delete a pack by ID
// @Summary [ADMIN] Delete a pack by ID
// @Description Delete a pack by ID - for admin use only
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack ID"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /packs/{id} [delete]

func DeletePackByID(c *gin.Context) {

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = deletePackById(id)
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
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /mypack/{id} [delete]
func DeleteMyPackByID(c *gin.Context) {

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	myPack, err := checkPackOwnership(id, user_id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		err := deletePackById(id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Pack deleted"})
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func deletePackById(id uint) error {
	statement, err := database.Db().Prepare("DELETE FROM pack WHERE id=$1;")
	if err != nil {
		return err
	}
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
// @Tags Packs
// @Produce  json
// @Success 200 {object} dataset.PackContents
// @Failure 500 {object} map[string]interface{} "error"
// @Router /packcontents [get]
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

	rows, err := database.Db().Query("SELECT id, pack_id, item_id, quantity, worn, consumable, created_at, updated_at FROM pack_content;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var packContent dataset.PackContent
		err := rows.Scan(&packContent.ID, &packContent.Pack_id, &packContent.Item_id, &packContent.Quantity, &packContent.Worn, &packContent.Consumable, &packContent.Created_at, &packContent.Updated_at)
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
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack Content ID"
// @Success 200 {object} dataset.PackContent
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 404 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /packcontents/{id} [get]
func GetPackContentByID(c *gin.Context) {

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	packcontent, err := findPackContentById(id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if packcontent == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack Item not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, *packcontent)

}

func findPackContentById(id uint) (*dataset.PackContent, error) {
	var packcontent dataset.PackContent

	row := database.Db().QueryRow("SELECT id, pack_id, item_id, quantity, worn, consumable, created_at, updated_at FROM pack_content WHERE id = $1;", id)
	err := row.Scan(&packcontent.ID, &packcontent.Pack_id, &packcontent.Item_id, &packcontent.Quantity, &packcontent.Worn, &packcontent.Consumable, &packcontent.Created_at, &packcontent.Updated_at)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &packcontent, nil
}

// Create a new pack content
// @Summary [ADMIN] Create a new pack content
// @Description Create a new pack content - for admin use only
// @Security Bearer
// @Tags Packs
// @Accept  json
// @Produce  json
// @Param packcontent body dataset.PackContent true "Pack Content"
// @Success 201 {object} dataset.PackContent
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /packcontents [post]
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
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /mypackcontent [post]
func PostMyPackContent(c *gin.Context) {
	var newPackContent dataset.PackContent

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.BindJSON(&newPackContent); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body format"})
		return
	}

	myPack, err := checkPackOwnership(id, user_id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		newPackContent.Pack_id = id
		err := insertPackContent(&newPackContent)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusCreated, newPackContent)
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func insertPackContent(pc *dataset.PackContent) error {
	if pc == nil {
		return errors.New("payload is empty")
	}
	pc.Created_at = time.Now().Truncate(time.Second)
	pc.Updated_at = time.Now().Truncate(time.Second)

	err := database.Db().QueryRow("INSERT INTO pack_content (pack_id, item_id, quantity, worn, consumable, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id;", pc.Pack_id, pc.Item_id, pc.Quantity, pc.Worn, pc.Consumable, pc.Created_at, pc.Updated_at).Scan(&pc.ID)

	if err != nil {
		return err
	}
	return nil
}

// Update a pack content by ID
// @Summary [ADMIN] Update a pack content by ID
// @Description Update a pack content by ID - for admin use only
// @Security Bearer
// @Tags Packs
// @Accept  json
// @Produce  json
// @Param id path int true "Pack Content ID"
// @Param packcontent body dataset.PackContent true "Pack Content"
// @Success 200 {object} dataset.PackContent
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /packcontents/{id} [put]
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

func PutMyPackContentByID(c *gin.Context) {

	var updatedPackContent dataset.PackContent

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	item_id, err := helper.StringToUint(c.Param("item_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.BindJSON(&updatedPackContent); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body format"})
		return
	}

	myPack, err := checkPackOwnership(id, user_id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		updatedPackContent.Pack_id = id
		err := updatePackContentByID(item_id, &updatedPackContent)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, updatedPackContent)
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func updatePackContentByID(id uint, pc *dataset.PackContent) error {
	if pc == nil {
		return errors.New("payload is empty")
	}
	pc.Updated_at = time.Now().Truncate(time.Second)

	statement, err := database.Db().Prepare("UPDATE pack_content SET pack_id=$1, item_id=$2, quantity=$3, worn=$4, consumable=$5, updated_at=$6 WHERE id=$7;")
	if err != nil {
		return err
	}
	_, err = statement.Exec(pc.Pack_id, pc.Item_id, pc.Quantity, pc.Worn, pc.Consumable, pc.Updated_at, id)
	if err != nil {
		return err
	}
	return nil
}

// Delete a pack content by ID
// @Summary [ADMIN] Delete a pack content by ID
// @Description Delete a pack content by ID - for admin use only
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack Content ID"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /packcontents/{id} [delete]
func DeletePackContentByID(c *gin.Context) {

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = deletePackContentById(id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Pack Item deleted"})

}

// Delete a pack content by ID
// @Summary Delete a pack content by ID
// @Description Delete a pack content by ID
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Param id path int true "Pack Content ID"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /mypackcontent/{id} [delete]
func DeleteMyPackContentByID(c *gin.Context) {

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	item_id, err := helper.StringToUint(c.Param("item_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	myPack, err := checkPackOwnership(id, user_id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		err := deletePackContentById(item_id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Pack Item deleted"})
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func deletePackContentById(id uint) error {
	statement, err := database.Db().Prepare("DELETE FROM pack_content WHERE id=$1;")
	if err != nil {
		return err
	}
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
// @Tags Packs
// @Produce  json
// @Success 200 {object} dataset.PackContents
// @Failure 500 {object} map[string]interface{} "error"
// @Router /packs/:id/packcontents [get]
func GetPackContentsByPackID(c *gin.Context) {

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	packContents, err := returnPackContentsByPackID(id)
	if err != nil {
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
// @Success 200 {object} dataset.PackContent
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 404 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /mypackcontent/{id} [get]
func GetMyPackContentsByPackID(c *gin.Context) {
	var packContents *dataset.PackContentWithItems

	id, err := helper.StringToUint(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	myPack, err := checkPackOwnership(id, user_id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if myPack {
		packContents, err = returnPackContentsByPackID(id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if packContents == nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
			return
		}
		c.IndentedJSON(http.StatusOK, packContents)
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "This pack does not belong to you"})
		return
	}
}

func returnPackContentsByPackID(id uint) (*dataset.PackContentWithItems, error) {
	var packWithItems dataset.PackContentWithItems

	rows, err := database.Db().Query("SELECT pc.id AS pack_content_id, pc.pack_id as pack_id, i.id AS inventory_id, i.item_name, i.category, i.description AS item_description, i.weight, i.weight_unit, i.url AS item_url, i.price, i.currency, pc.quantity, pc.worn, pc.consumable FROM pack_content pc JOIN inventory i ON pc.item_id = i.id WHERE pc.pack_id = $1;", id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item dataset.PackContentWithItem
		err := rows.Scan(&item.Pack_content_id, &item.Pack_id, &item.Inventory_id, &item.Item_name, &item.Category, &item.Item_description, &item.Weight, &item.Weight_unit, &item.Item_url, &item.Price, &item.Currency, &item.Quantity, &item.Worn, &item.Consumable)
		if err != nil {
			return nil, err
		}
		packWithItems = append(packWithItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(packWithItems) == 0 {
		return nil, nil
	}
	return &packWithItems, nil
}

// Get all packs
// @Summary [ADMIN] Get all packs
// @Description Get all packs - for admin use only
// @Security Bearer
// @Tags Packs
// @Produce  json
// @Success 200 {object} []dataset.Pack
// @Failure 500 {object} map[string]interface{} "error"
// @Router /mypacks [get]
func GetMyPacks(c *gin.Context) {
	user_id, err := security.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	packs, err := findPacksByUserId(user_id)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if packs == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No pack found"})
		return
	}

	c.IndentedJSON(http.StatusOK, *packs)

}

func findPacksByUserId(id uint) (*dataset.Packs, error) {
	var packs dataset.Packs

	rows, err := database.Db().Query("SELECT id, user_id, pack_name, pack_description, created_at, updated_at FROM pack WHERE user_id = $1;", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pack dataset.Pack
		err := rows.Scan(&pack.ID, &pack.User_id, &pack.Pack_name, &pack.Pack_description, &pack.Created_at, &pack.Updated_at)
		if err != nil {
			return nil, err
		}
		packs = append(packs, pack)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(packs) == 0 {
		return nil, nil
	}

	return &packs, nil
}

func checkPackOwnership(id uint, user_id uint) (bool, error) {
	var rows int

	row := database.Db().QueryRow("SELECT COUNT(id) FROM pack WHERE id = $1 AND user_id = $2;", id, user_id)
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

// Import from lighterpack
// @Summary Import from lighterpack csv pack file
// @Description Import from lighterpack csv pack file
// @Security Bearer
// @Tags Packs
// @Accept  multipart/form-data
// @Produce  json
// @Param file formData file true "CSV file"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]interface{} "error"
// @Router /mypack/import [post]
func ImportFromLighterPack(c *gin.Context) {
	var lighterPack dataset.LighterPack

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	record, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the CSV header"})
		return
	}
	if len(record) < 10 {
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

		lighterPackItem.Item_name = record[0]
		lighterPackItem.Category = record[1]
		lighterPackItem.Desc = record[2]
		lighterPackItem.Unit = record[5]
		lighterPackItem.Url = record[6]

		qty, err := strconv.Atoi(record[3])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format - failed to convert quantity to number"})
			return
		} else {
			lighterPackItem.Qty = qty
		}

		weight, err := strconv.Atoi(record[4])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format - failed to convert weight to number"})
			return
		} else {
			lighterPackItem.Weight = weight
		}

		price, err := strconv.Atoi(record[7])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format - failed to convert price to number"})
			return
		} else {
			lighterPackItem.Price = price
		}

		if record[8] == "worn" {
			lighterPackItem.Worn = true
		}

		if record[9] == "consumable" {
			lighterPackItem.Consumable = true
		}

		lighterPack = append(lighterPack, lighterPackItem)
	}

	// Perform your database insertion
	err = insertLighterPack(&lighterPack, user_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "CSV data imported successfully"})
}

func insertLighterPack(lp *dataset.LighterPack, user_id uint) error {
	if lp == nil {
		return errors.New("payload is empty")
	}

	// Create new pack
	var newPack dataset.Pack
	newPack.User_id = user_id
	newPack.Pack_name = "LighterPack Import"
	newPack.Pack_description = "LighterPack Import"
	err := insertPack(&newPack)
	if err != nil {
		return err
	}

	// Insert content in new pack with insertPackContent
	for _, item := range *lp {
		var i dataset.Inventory
		i.User_id = user_id
		i.Item_name = item.Item_name
		i.Category = item.Category
		i.Description = item.Desc
		i.Weight = item.Weight
		i.Weight_unit = helper.ConvertWeightUnit(item.Unit)
		i.Url = item.Url
		i.Price = item.Price
		i.Currency = "USD"
		err := inventories.InsertInventory(&i)
		if err != nil {
			return err
		}
		var pc dataset.PackContent
		pc.Pack_id = newPack.ID
		pc.Item_id = i.ID
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
