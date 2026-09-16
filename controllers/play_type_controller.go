package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create a new card type
// @Tags Play Type
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.PlayTypeDto true "Play Type Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/play-type [post]
func CreatePlayType(c *gin.Context) {
	var req models.PlayTypeDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// ตรวจสอบว่าชื่อซ้ำหรือไม่
	existPlayType, err := services.FindExistPlayTypeName(req.Name)
	if err == nil && existPlayType.ID != uuid.Nil {
		utils.Error(c, http.StatusBadRequest, "Play type with this name already exists")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	playType, err := services.CreatePlayType(req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create play type")
		return
	}

	utils.Success(c, "Play type created successfully", playType)
}

// UpdateCardType godoc
// @Summary Update play type by ID
// @Tags Play Type
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.PlayTypeDto true "Play Type Data"
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/play-type/{id} [put]
func UpdatePlayType(c *gin.Context) {
	id := c.Param("id")

	var req models.PlayTypeDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	findPlayType, err := services.FindPlayTypeById(id)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findPlayType == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Play type id: %s not found.", id))
		return
	}

	existPlayType, err := services.FindExistPlayTypeNameNotCurrent(req.Name, findPlayType.ID.String())
	if err == nil && existPlayType != nil && existPlayType.ID != uuid.Nil {
		utils.Error(c, http.StatusBadRequest, "Play type with this name already exists other record")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updatePlayType, err := services.UpdatePlayType(id, req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to update play type")
		return
	}

	utils.Success(c, "Play type updated successfully", updatePlayType)
}

// GetCardTypeById godoc
// @Summary Get play type by id
// @Tags Play Type
// @Security BasicAuth
// @Security BearerAuth
// @Param   id  path  string  true  "Id"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/play-type/{id} [get]
func GetPlayTypeById(c *gin.Context) {
	id := c.Param("id")

	findPlayType, err := services.FindPlayTypeById(id)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve play type")
		return
	}

	utils.Success(c, "Play type retrieved successfully", findPlayType)
}

// GetCardTypeAll godoc
// @Summary Get play type all active
// @Tags Play Type
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/play-type [get]
func GetPlayTypeList(c *gin.Context) {

	findAll, err := services.GetActivePlayType()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve play type")
		return
	}

	utils.Success(c, "Play type retrieved successfully", findAll)
}

// SearchCardTypeList godoc
// @Summary Get play type list with paggination
// @Tags Play Type
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search text"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/play-type/list [get]
func SearchPlayTypeList(c *gin.Context) {

	search := c.Query("search")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	skip, err := strconv.Atoi(c.DefaultQuery("skip", "10"))
	if err != nil || skip < 1 {
		skip = 10
	}

	params := models.SearchParams{
		Search: search,
		Page:   page,
		Skip:   skip,
	}
	result, err := services.PlayTypeSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve card type")
		return
	}

	utils.Success(c, "Play type retrieved successfully", result)
}

// DeleteCardTypeById godoc
// @Summary Delete play type by ID
// @Tags Play Type
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/play-type/{id} [delete]
func DeletePlayTypeById(c *gin.Context) {
	id := c.Param("id")

	findPlayType, err := services.FindPlayTypeById(id)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findPlayType == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Play type id: %s not found.", id))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	deletePlayType, err := services.DeletePlayTypeById(id, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to delete play type")
		return
	}

	utils.Success(c, "Play type deleted successfully", deletePlayType)
}
