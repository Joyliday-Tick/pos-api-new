package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"
	"strconv"

	// "strconv"

	"github.com/gin-gonic/gin"
)

// @Summary Create a new pos menu
// @Tags POS Menu
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.PosMenuDto true "POS Menu"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-menu [post]
func CreatePosMenu(c *gin.Context) {
	var req models.PosMenuDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	fmt.Printf("%+v\n", req)

	// existPosMenuName, err := services.FindExistPosMenuName(req.Description)
	// if err == nil && existPosMenuName.ID != 0 {
	// 	utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Pos menu with this name %s already exists", req.Description))
	// 	return
	// }

	existPosMenuCode, err := services.FindExistPosMenuCode(req.Code)
	if err == nil && existPosMenuCode.ID != 0 {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Pos menu with this code %s already exists", req.Code))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	posMenu, err := services.CreatePosMenu(req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create pos menu")
		return
	}

	utils.Success(c, "Pos menu created successfully", posMenu)
}

// UpdatePosMenu godoc
// @Summary Update pos menu by ID
// @Tags POS Menu
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.PosMenuDto true "Pos menu Data"
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-menu/{id} [put]
func UpdatePosMenu(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}
	var req models.PosMenuDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	findPosMenu, err := services.FindPosMenuById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findPosMenu == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Pos menu id: %s not found.", id))
		return
	}

	existPosMenu, err := services.FindExistPosMenuCodeNotCurrent(req.Code, findPosMenu.ID)
	if err == nil && existPosMenu != nil && existPosMenu.ID != 0 {
		msg := fmt.Sprintf("Pos menu with this name is %s or code is %s already exists other record", req.Description, req.Code)
		utils.Error(c, http.StatusBadRequest, msg)
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updatePosMenu, err := services.UpdatePosMenu(dInt, req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update pos menu: %v", err))
		return
	}

	utils.Success(c, "Pos menu updated successfully", updatePosMenu)
}

// GetPosMenuById godoc
// @Summary Get pos menu by id
// @Tags POS Menu
// @Security BasicAuth
// @Security BearerAuth
// @Param   id  path  string  true  "Id"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-menu/{id} [get]
func GetPosMenuById(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findPosMenu, err := services.FindPosMenuById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve menu: %v", err))
		return
	}

	utils.Success(c, "Pos menu retrieved successfully", findPosMenu)
}

// GetPosMenuByGroupMenuId godoc
// @Summary Get pos menu by group menu id and active
// @Tags POS Menu
// @Security BasicAuth
// @Security BearerAuth
// @Param   groupMenuId  path  string  true  "GroupMenuId"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-menu/group/{groupMenuId} [get]
func GetPosMenuByGroupMenuId(c *gin.Context) {
	groupMenuId := c.Param("groupMenuId")
	fmt.Println("groupMenuId", groupMenuId)

	if groupMenuId == "" {
		utils.Error(c, http.StatusBadRequest, "groupMenuId is required.")
		return
	}

	findPosMenu, err := services.FindPosMenuByGroupMenuId(groupMenuId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve menu: %v", err))
		return
	}

	utils.Success(c, "Pos menu retrieved successfully", findPosMenu)
}

// GetPosMenuByLocation godoc
// @Summary Get pos menu by location and active
// @Tags POS Menu
// @Security BasicAuth
// @Security BearerAuth
// @Param   location  path  string  true  "Location"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-menu/location/{location} [get]
func GetPosMenuByLocation(c *gin.Context) {
	location := c.Param("location")
	fmt.Println("location", location)

	if location == "" {
		utils.Error(c, http.StatusBadRequest, "location is required.")
		return
	}

	findPosMenu, err := services.FindPosMenuByLocation(location)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve menu: %v", err))
		return
	}

	utils.Success(c, "Pos menu retrieved successfully", findPosMenu)
}

// SearchPosMenuList godoc
// @Summary Get pos menu list with paggination
// @Tags POS Menu
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search text name or code"
// @Param groupMenuId query string false "Search group menu"
// @Param branchGroupId query int false "Search branch group"
// @Param machineGroupId query int false "Search machine group"
// @Param cardTypeId query string false "Search card type"
// @Param isActive query bool false "Search active status (true/false)"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-menu/list [get]
func SearchPosMenuList(c *gin.Context) {

	search := c.Query("search")
	// location := c.Query("location")
	groupMenuId := c.Query("groupMenuId")
	branchGroupIdStr := c.Query("branchGroupId")
	machineGroupIdStr := c.Query("machineGroupId")
	cardTypeId := c.Query("cardTypeId")
	isActiveStr := c.Query("isActive")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	skip, err := strconv.Atoi(c.DefaultQuery("skip", "10"))
	if err != nil || skip < 1 {
		skip = 10
	}

	var branchGroupId *int
	if branchGroupIdStr != "" {
		id, err := strconv.Atoi(branchGroupIdStr)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Invalid branchGroupId: %v", err))
			return
		}
		branchGroupId = &id
	}

	var machineGroupId *int
	if machineGroupIdStr != "" {
		id, err := strconv.Atoi(machineGroupIdStr)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Invalid branchGroupId: %v", err))
			return
		}
		machineGroupId = &id
	}

	var isActive *bool
	if isActiveStr != "" {
		active, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Invalid isActive value: %v", err))
			return
		}
		isActive = &active
	}

	params := models.SearchPosMenuParams{
		Search: search,
		// Location:       location,
		GroupMenuID:    groupMenuId,
		BranchGroupID:  branchGroupId,
		MachineGroupID: machineGroupId,
		CardTypeID:     cardTypeId,
		IsActive:       isActive,
		Page:           page,
		Skip:           skip,
	}
	result, err := services.PosMenuSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve pos menu list")
		return
	}

	utils.Success(c, "Pos menu list retrieved successfully", result)
}

// DeletePosMenuById godoc
// @Summary Delete pos menu by ID
// @Tags POS Menu
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-menu/{id} [delete]
func DeletePosMenuById(c *gin.Context) {
	idStr := c.Param("id")
	dInt, err := strconv.Atoi(idStr)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findPosMenu, err := services.FindPosMenuById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findPosMenu == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Pos menu id: %s not found.", idStr))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	deletePosMenu, err := services.DeletePosMenuById(dInt, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to delete pos menu")
		return
	}

	utils.Success(c, "Pos menu deleted successfully", deletePosMenu)
}

// SalePosMenuList godoc
// @Summary Get pos menu list with paggination by loaction and group menu
// @Tags POS Menu
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search text location code"
// @Param groupMenuId query string false "Search group menu"
// @Param search query string false "Search menu name"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-menu/sale/list [get]
func SalePosMenuList(c *gin.Context) {

	location := c.Query("location")
	groupMenuId := c.Query("groupMenuId")
	search := c.Query("search")
	if location == "" {
		utils.Error(c, http.StatusBadRequest, "location is required.")
		return
	}

	if groupMenuId == "" {
		utils.Error(c, http.StatusBadRequest, "group menu id is required.")
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	skip, err := strconv.Atoi(c.DefaultQuery("skip", "10"))
	if err != nil || skip < 1 {
		skip = 10
	}

	params := models.SearchPosMenuSaleParams{
		Location:    location,
		GroupMenuID: groupMenuId,
		Search:      search,
		Page:        page,
		Skip:        skip,
	}
	result, err := services.PosMenuSaleList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve pos menu sale list")
		return
	}

	utils.Success(c, "Pos menu sale list retrieved successfully", result)
}

// GetPosMenuSaleAll godoc
// @Summary Get pos menu sale all
// @Tags POS Menu
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-menu/sale-all [get]
func GetPosMenuSaleAll(c *gin.Context) {

	findPosMenu, err := services.FindPosMenuSaleAll()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve menu: %v", err))
		return
	}

	utils.Success(c, "Pos menu retrieved successfully", findPosMenu)
}
