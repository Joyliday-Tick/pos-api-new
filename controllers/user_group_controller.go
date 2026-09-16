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

// GetUserGroupAll godoc
// @Summary Get user group all active
// @Tags User Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-group [get]
func GetUserGroupList(c *gin.Context) {

	findAll, err := services.GetActiveUserGroup()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve user group: %v", err))
		return
	}

	utils.Success(c, "User group retrieved successfully", findAll)
}

// @Summary Create a new user group
// @Tags User Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.UserGroupDto true "User Group Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-group [post]
func CreateUserGroup(c *gin.Context) {
	var req models.UserGroupDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	existUserGroup, err := services.FindExistUserGroupName(strings.TrimSpace(req.GroupName))
	if err == nil && existUserGroup.ID != 0 {
		utils.Error(c, http.StatusBadRequest, "User group with this name already exists")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	userGroup, err := services.CreateUserGroup(req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create user group: %v", err))
		return
	}

	branchs := req.BranchList
	var subBranchs []models.UsersSubGroup

	for _, branch := range branchs {
		fmt.Println("Branch loop:", branch)

		subBranch := models.UsersSubGroup{
			LocationCode: branch,
			GroupID:      &userGroup.ID,
		}

		subBranchs = append(subBranchs, subBranch)
	}

	if len(subBranchs) > 0 {
		_, err := services.CreateBatchUserSubGroup(subBranchs, userId)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create user sub group: %v", err))
			return
		}
	}

	utils.Success(c, "User group created successfully", userGroup)
}

// UpdateUserGroup godoc
// @Summary Update user group by ID
// @Tags User Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.UserGroupDto true "User group Data"
// @Param   id  path  int  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-group/{id} [put]
func UpdateUserGroup(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	var req models.UserGroupDto
	fmt.Printf("%+v\n", req)

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	findUserGroup, err := services.FindUserGroupById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findUserGroup == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("User group id: %s not found.", id))
		return
	}

	existUserGroup, err := services.FindExistUserGroupNameNotCurrent(req.GroupName, findUserGroup.ID)
	if err == nil && existUserGroup != nil && existUserGroup.ID != 0 {
		utils.Error(c, http.StatusBadRequest, "Branch group with this name already exists other record")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateUserGroup, err := services.UpdateUserGroup(findUserGroup.ID, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update branch group: %v", err))
		return
	}
	//Delete existing sub-branch groups
	_, err = services.DeleteUserSubGroupByGroupId(findUserGroup.ID, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete existing sub-user groups: %v", err))
		return
	}
	// Update sub-branch groups
	branchs := req.BranchList
	var subBranchs []models.UsersSubGroup
	for _, branch := range branchs {
		fmt.Println("Branch loop:", branch)
		subBranch := models.UsersSubGroup{
			LocationCode: branch,
			GroupID:      &updateUserGroup.ID,
		}
		subBranchs = append(subBranchs, subBranch)
	}
	if len(subBranchs) > 0 {
		_, err := services.CreateBatchUserSubGroup(subBranchs, userId)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create user sub group: %v", err))
			return
		}
	}

	utils.Success(c, "User group updated successfully", updateUserGroup)
}

// GetUserGroupById godoc
// @Summary Get user group by id
// @Tags User Group
// @Security BasicAuth
// @Security BearerAuth
// @Param   id  path  int  true  "Id"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-group/{id} [get]
func GetUserGroupById(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findUserGroup, err := services.FindUserGroupByIdWithLocation(dInt)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve user group: %v", err))
		return
	}

	utils.Success(c, "User group retrieved successfully", findUserGroup)
}

// SearchUserGroupList godoc
// @Summary Get user group list with paggination
// @Tags User Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search text"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-group/list [get]
func SearchUserGroupList(c *gin.Context) {

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
	result, err := services.UserGroupSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve user group: %v", err))
		return
	}

	utils.Success(c, "Branch group retrieved successfully", result)
}

// DeleteUserGroupById godoc
// @Summary Delete user group by ID
// @Tags User Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user-group/{id} [delete]
func DeleteUserGroupById(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findUserGroup, err := services.FindUserGroupById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findUserGroup == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("User group  id: %s not found.", id))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	deleteUserGroup, err := services.DeleteUserGroupById(dInt, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to delete user group: %v", err))
		return
	}

	utils.Success(c, "User group deleted successfully", deleteUserGroup)
}
