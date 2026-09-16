package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	// "new-pos-api/middlewares"
	// "new-pos-api/models"
	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"

	// "strconv"

	"github.com/gin-gonic/gin"
	// "github.com/google/uuid"
)

// GetUserRoleAll godoc
// @Summary Get user role all active
// @Tags User Role
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-role [get]
func GetUserRoleList(c *gin.Context) {

	findAll, err := services.GetActiveUserRole()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve user role: %v", err))
		return
	}

	utils.Success(c, "User role retrieved successfully", findAll)
}

// @Summary Create a new user role
// @Tags User Role
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.UserRoleDto true "User role Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-role [post]
func CreateUserRole(c *gin.Context) {
	var req models.UserRoleDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	existUserRole, err := services.FindExistRoleName(strings.TrimSpace(req.RoleDescription))
	if err == nil && existUserRole.ID != 0 {
		utils.Error(c, http.StatusBadRequest, "User role with this name already exists")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	userRole, err := services.CreateUserRole(req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create user role: %v", err))
		return
	}

	utils.Success(c, "User role created successfully", userRole)
}

// UpdateUserrole godoc
// @Summary Update user role by ID
// @Tags User Role
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.UserRoleDto true "User role Data"
// @Param   id  path  int  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-role/{id} [put]
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	var req models.UserRoleDto
	fmt.Printf("%+v\n", req)

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	findUserRole, err := services.FindUserRoleById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findUserRole == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("User role id: %s not found.", id))
		return
	}

	existUserRole, err := services.FindExistRoleNameNotCurrent(req.RoleDescription, findUserRole.ID)
	if err == nil && existUserRole != nil && existUserRole.ID != 0 {
		utils.Error(c, http.StatusBadRequest, "User role with this name already exists other record")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateUserRole, err := services.UpdateUserRole(findUserRole.ID, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update user role: %v", err))
		return
	}

	utils.Success(c, "User role updated successfully", updateUserRole)
}

// GetUserRoleById godoc
// @Summary Get user role by id
// @Tags User Role
// @Security BasicAuth
// @Security BearerAuth
// @Param   id  path  int  true  "Id"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-role/{id} [get]
func GetUserRoleById(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findUserRole, err := services.FindUserRoleById(dInt)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve user role: %v", err))
		return
	}

	utils.Success(c, "User role retrieved successfully", findUserRole)
}

// SearchUserRoleList godoc
// @Summary Get user role list with paggination
// @Tags User Role
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search text"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-role/list [get]
func SearchUserRoleList(c *gin.Context) {

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
	result, err := services.UserRoleSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve user role: %v", err))
		return
	}

	utils.Success(c, "User role retrieved successfully", result)
}

// DeleteUserRoleById godoc
// @Summary Delete user role by ID
// @Tags User Role
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-role/{id} [delete]
func DeleteUserRoleById(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findUserRole, err := services.FindUserRoleById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findUserRole == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("User role  id: %s not found.", id))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	deleteUserRole, err := services.DeleteUserRoleById(dInt, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to delete user role: %v", err))
		return
	}

	utils.Success(c, "User role deleted successfully", deleteUserRole)
}
