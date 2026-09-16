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

// @Summary Create a group menu
// @Tags Group Menu
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.GroupMenuDto true "Group Menu Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/group-menu [post]
func CreateGroupMenu(c *gin.Context) {
	var req models.GroupMenuDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// ตรวจสอบว่าชื่อซ้ำหรือไม่
	existGroupMenu, err := services.FindExistGroupMenuName(req.Name)
	if err == nil && existGroupMenu.ID != uuid.Nil {
		utils.Error(c, http.StatusBadRequest, "Group menu with this name already exists")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	groupMenu, err := services.CreateGroupMenu(req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to create card type: %v", err))
		return
	}

	utils.Success(c, "Group menu created successfully", groupMenu)
}

// UpdateGroupMenu godoc
// @Summary Update group menu by ID
// @Tags Group Menu
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.GroupMenuDto true "Group Menu Data"
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/group-menu/{id} [put]
func UpdateGroupMenu(c *gin.Context) {
	id := c.Param("id")

	var req models.GroupMenuDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	findGroupMenu, err := services.FindGroupMenuById(id)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve group menu: %v", err))
		return
	}
	if findGroupMenu == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Group menu id: %s not found.", id))
		return
	}

	existGroupMenu, err := services.FindExistGroupMenuNameNotCurrent(req.Name, findGroupMenu.ID.String())
	if err == nil && existGroupMenu != nil && existGroupMenu.ID != uuid.Nil {
		utils.Error(c, http.StatusBadRequest, "Group menu with this name already exists other record")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateGroupMenu, err := services.UpdateGroupMenu(id, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update group menu: %v", err))
		return
	}

	utils.Success(c, "Group menu updated successfully", updateGroupMenu)
}

// GetGroupMenuById godoc
// @Summary Get group menu by id
// @Tags Group Menu
// @Security BasicAuth
// @Security BearerAuth
// @Param   id  path  string  true  "Id"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/group-menu/{id} [get]
func GetGroupMenuById(c *gin.Context) {
	id := c.Param("id")

	findGroupMenu, err := services.FindGroupMenuById(id)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve group menu: %v", err))
		return
	}

	utils.Success(c, "Group menu retrieved successfully", findGroupMenu)
}

// GetGroupMenuAll godoc
// @Summary Get group menu all active
// @Tags Group Menu
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/group-menu [get]
func GetGroupMenuList(c *gin.Context) {

	findAll, err := services.GetActiveGroupMenu()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve group menu: %v", err))
		return
	}

	utils.Success(c, "Group menu retrieved successfully", findAll)
}

// SearchGroupMenuList godoc
// @Summary Get group menu list with paggination
// @Tags Group Menu
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search text"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/group-menu/list [get]
func SearchGroupMenuList(c *gin.Context) {

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
	result, err := services.GroupMenuSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve group menu: %v", err))
		return
	}

	utils.Success(c, "Card type retrieved successfully", result)
}

// DeleteGroupMenuById godoc
// @Summary Delete group menu by ID
// @Tags Group Menu
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/group-menu/{id} [delete]
func DeleteGroupMenuById(c *gin.Context) {
	id := c.Param("id")

	findGroupMenu, err := services.FindGroupMenuById(id)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve group menu: %v", err))
		return
	}
	if findGroupMenu == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Group menu id: %s not found.", id))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	deleteGroupMenu, err := services.DeleteGroupMenuById(id, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to delete group menu: %v", err))
		return
	}

	utils.Success(c, "Group menu deleted successfully", deleteGroupMenu)
}
