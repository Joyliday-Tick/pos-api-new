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
// @Tags Card Type
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CardTypeDto true "Card Type Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card-type [post]
func CreateCardType(c *gin.Context) {
	var req models.CardTypeDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// ตรวจสอบว่าชื่อซ้ำหรือไม่
	existCardType, err := services.FindExistCardTypeName(req.Name)
	if err == nil && existCardType.ID != uuid.Nil {
		utils.Error(c, http.StatusBadRequest, "Card type with this name already exists")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	cardType, err := services.CreateCardType(req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to create card type: %v", err))
		return
	}

	utils.Success(c, "Card type created successfully", cardType)
}

// UpdateCardType godoc
// @Summary Update card type by ID
// @Tags Card Type
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CardTypeDto true "Card Type Data"
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card-type/{id} [put]
func UpdateCardType(c *gin.Context) {
	id := c.Param("id")

	var req models.CardTypeDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	findCardType, err := services.FindCardTypeById(id)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findCardType == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Card type id: %s not found.", id))
		return
	}

	existCardType, err := services.FindExistCardTypeNameNotCurrent(req.Name, findCardType.ID.String())
	if err == nil && existCardType != nil && existCardType.ID != uuid.Nil {
		utils.Error(c, http.StatusBadRequest, "Card type with this name already exists other record")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateCardType, err := services.UpdateCardType(id, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update card type: %v", err))
		return
	}

	utils.Success(c, "Card type updated successfully", updateCardType)
}

// GetCardTypeById godoc
// @Summary Get card type by id
// @Tags Card Type
// @Security BasicAuth
// @Security BearerAuth
// @Param   id  path  string  true  "Id"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card-type/{id} [get]
func GetCardTypeById(c *gin.Context) {
	id := c.Param("id")

	findCardType, err := services.FindCardTypeById(id)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve card type: %v", err))
		return
	}

	utils.Success(c, "Card type retrieved successfully", findCardType)
}

// GetCardTypeAll godoc
// @Summary Get card type all active
// @Tags Card Type
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card-type [get]
func GetCardTypeList(c *gin.Context) {

	findAll, err := services.GetActiveCardType()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve card type: %v", err))
		return
	}

	utils.Success(c, "Card type retrieved successfully", findAll)
}

// SearchCardTypeList godoc
// @Summary Get card type list with paggination
// @Tags Card Type
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search text"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card-type/list [get]
func SearchCardTypeList(c *gin.Context) {

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
	result, err := services.CardTypeSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve card type: %v", err))
		return
	}

	utils.Success(c, "Card type retrieved successfully", result)
}

// DeleteCardTypeById godoc
// @Summary Delete card type by ID
// @Tags Card Type
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card-type/{id} [delete]
func DeleteCardTypeById(c *gin.Context) {
	id := c.Param("id")

	findCardType, err := services.FindCardTypeById(id)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findCardType == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Card type id: %s not found.", id))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	deleteCardType, err := services.DeleteCardTypeById(id, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to delete card type: %v", err))
		return
	}

	utils.Success(c, "Card type deleted successfully", deleteCardType)
}
