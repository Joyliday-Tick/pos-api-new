package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetUsers godoc
// @Summary Get all users
// @Description Get all users with Bearer token
// @Tags User
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/users [get]
func GetUsers(c *gin.Context) {
	claims, _ := c.Get("userClaims")
	fmt.Println(claims)
	users, err := services.GetAllUsers()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}
	// c.JSON(http.StatusOK, users)
	utils.Success(c, "Users retrieved successfully", users)
}

// GetUserByusername godoc
// @Summary Get users by username
// @Description Get users by username with Bearer token
// @Tags User
// @Security BasicAuth
// @Security BearerAuth
// @Param   username  path  string  true  "Username"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/users/{username} [get]
func GetUsersByUsername(c *gin.Context) {
	username := c.Param("username")
	users, err := services.FindUserDbByUsername(username)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	utils.Success(c, "Users retrieved successfully", users)
}

// @Summary Create a new user
// @Tags User
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.UsersCreateDto true "User Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user [post]
func CreateUser(c *gin.Context) {
	var req models.UsersCreateDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// ทั้งสองคอลัมน์เป็น nullable และ DTO ไม่ได้บังคับ ถ้าเรียกมาโดยไม่ส่งสองค่านี้
	// จะได้บัญชีที่ "สร้างสำเร็จ" แต่ล็อกอินไม่ได้ เพราะ AuthenticateWithBranch
	// ต้องใช้ทั้ง role และ group เพื่อประกอบ token — ปิดทางไว้ตั้งแต่ตอนสร้าง
	if req.UserRoleId == nil {
		utils.Error(c, http.StatusBadRequest, "ต้องระบุ user_role_id ไม่งั้นบัญชีนี้จะล็อกอินไม่ได้")
		return
	}
	if req.UserGroupId == nil {
		utils.Error(c, http.StatusBadRequest, "ต้องระบุ user_group_id ไม่งั้นบัญชีนี้จะล็อกอินไม่ได้")
		return
	}

	existUser, err := services.FindUserDbByUsername(strings.TrimSpace(req.Username))
	if err == nil && existUser != nil && existUser.ID != 0 {
		utils.Error(c, http.StatusBadRequest, "User with this username already exists")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	user, err := services.CreateUser(req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create user: %v", err))
		return
	}

	utils.Success(c, "User created successfully", user)
}

// UpdateUser godoc
// @Summary Update user by ID
// @Tags User
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.UsersUpdateDto true "User Data"
// @Param   id  path  int  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user/{id} [put]
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// เดิม return เฉย ๆ โดยไม่เขียน response — gin จึงตอบ 200 ตัวเปล่า
		// ผู้เรียกเห็นเป็น "แก้ไขสำเร็จ" ทั้งที่ไม่ได้แตะฐานข้อมูลเลย
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("id ไม่ถูกต้อง: %s", id))
		return
	}

	var req models.UsersUpdateDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// updateData เขียนทับด้วย map เสมอ ถ้าสองค่านี้เป็น nil จะกลายเป็น
	// UPDATE ... SET user_role_id = NULL ทำให้บัญชีที่เคยใช้ได้ล็อกอินไม่ได้อีก
	if req.UserRoleId == nil {
		utils.Error(c, http.StatusBadRequest, "ต้องระบุ user_role_id ไม่งั้นบัญชีนี้จะล็อกอินไม่ได้")
		return
	}
	if req.UserGroupId == nil {
		utils.Error(c, http.StatusBadRequest, "ต้องระบุ user_group_id ไม่งั้นบัญชีนี้จะล็อกอินไม่ได้")
		return
	}

	findUser, err := services.FindUserById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findUser == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("User id: %s not found.", id))
		return
	}

	existUser, err := services.FindExistUserNameNotCurrent(req.Username, findUser.ID)
	if err == nil && existUser != nil && existUser.ID != 0 {
		utils.Error(c, http.StatusBadRequest, "User with this username already exists other record")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateUser, err := services.UpdateUser(findUser.ID, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update user: %v", err))
		return
	}

	utils.Success(c, "User updated successfully", updateUser)
}

// GetUserById godoc
// @Summary Get user by id
// @Tags User
// @Security BasicAuth
// @Security BearerAuth
// @Param   id  path  int  true  "Id"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user/{id} [get]
func GetUserById(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findUser, err := services.FindUserById(dInt)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve user: %v", err))
		return
	}

	utils.Success(c, "User retrieved successfully", findUser)
}

// SearchUserList godoc
// @Summary Get user list with paggination
// @Tags User
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search text"
// @Param username query string false "Username text"
// @Param userRoleId query string false "UserRoleId"
// @Param userGroupId query string false "UserGroupId"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user/list [get]
func SearchUserList(c *gin.Context) {

	search := c.Query("search")
	username := c.Query("username")
	userRoleId := c.Query("userRoleId")
	var userRoleIdInt *int
	if userRoleId != "" {
		id, err := strconv.Atoi(userRoleId)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, "Invalid userRoleId")
			return
		}
		userRoleIdInt = &id
	}
	if userRoleIdInt != nil && *userRoleIdInt < 0 {
		utils.Error(c, http.StatusBadRequest, "Invalid userRoleId")
		return
	}
	userGroupId := c.Query("userGroupId")
	var userGroupIdInt *int
	if userGroupId != "" {
		id, err := strconv.Atoi(userGroupId)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, "Invalid userGroupId")
			return
		}
		userGroupIdInt = &id
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	skip, err := strconv.Atoi(c.DefaultQuery("skip", "10"))
	if err != nil || skip < 1 {
		skip = 10
	}

	params := models.SearchUserParams{
		Search:      search,
		UserRoleId:  userRoleIdInt,
		UserGroupId: userGroupIdInt,
		Username:    username,
		Page:        page,
		Skip:        skip,
	}
	result, err := services.UserSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve user: %v", err))
		return
	}

	utils.Success(c, "User retrieved successfully", result)
}

// DeleteUserById godoc
// @Summary Delete user by ID
// @Tags User
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/user/{id} [delete]
func DeleteUserById(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findUser, err := services.FindUserById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findUser == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("User id: %s not found.", id))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	deleteUser, err := services.DeleteUserById(dInt, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to delete user: %v", err))
		return
	}

	utils.Success(c, "User deleted successfully", deleteUser)
}
