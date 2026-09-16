package controllers

import (
	"fmt"
	"net/http"
	"strconv"

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

// GetBranchAll godoc
// @Summary Get branch all active
// @Tags Branch
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch [get]
func GetBranchList(c *gin.Context) {

	findAll, err := services.GetActiveBranch()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve branch: %v", err))
		return
	}

	utils.Success(c, "Branch retrieved successfully", findAll)
}

// @Summary Create a new branch
// @Tags Branch
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.BranchDto true "Branch Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch [post]
func CreateBranch(c *gin.Context) {
	var req models.BranchDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	existBranchName, err := services.FindExistBranchName(req.BranchName)
	if err == nil && existBranchName.BranchCode != "" {
		utils.Error(c, http.StatusBadRequest, "Branch with this name already exists")
		return
	}

	existBranchCode, err := services.FindExistBranchCode(req.BranchCode)
	if err == nil && existBranchCode.BranchCode != "" {
		utils.Error(c, http.StatusBadRequest, "Branch with this code already exists")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	branch, err := services.CreateBranch(req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to create branch: %v", err))
		return
	}

	utils.Success(c, "Branch created successfully", branch)
}

// UpdateBranch godoc
// @Summary Update branch by code
// @Tags Branch
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.BranchDto true "Branch Data"
// @Param   code  path  string  true  "Code"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch/{code} [put]
func UpdateBranch(c *gin.Context) {
	code := c.Param("code")

	var req models.BranchDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	findBranch, err := services.FindExistBranchCode(code)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findBranch == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Branch code: %s not found.", code))
		return
	}

	existBranch, err := services.FindExistBranchNameNotCurrent(req.BranchCode, findBranch.BranchCode)
	if err == nil && existBranch != nil && existBranch.BranchCode != "" {
		utils.Error(c, http.StatusBadRequest, "Branch with this name already exists other record")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateBranch, err := services.UpdateBranch(code, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update branch: %v", err))
		return
	}

	utils.Success(c, "Branch updated successfully", updateBranch)
}

// GetBranchByCode godoc
// @Summary Get branch by code
// @Tags Branch
// @Security BasicAuth
// @Security BearerAuth
// @Param   code  path  string  true  "Code"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch/{code} [get]
func GetBranchByCode(c *gin.Context) {
	code := c.Param("code")

	findBranch, err := services.FindExistBranchCode(code)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve branch: %v", err))
		return
	}

	utils.Success(c, "Branch retrieved successfully", findBranch)
}

// SearchBranchList godoc
// @Summary Get branch list with paggination
// @Tags Branch
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search text"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch/list [get]
func SearchBranchList(c *gin.Context) {

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
	result, err := services.BranchSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve branch: %v", err))
		return
	}

	utils.Success(c, "Branch retrieved successfully", result)
}

// DeleteBranchByCode godoc
// @Summary Delete branch by code
// @Tags Branch
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param   code  path  string  true  "Code"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch/{code} [delete]
func DeleteBranchByCode(c *gin.Context) {
	code := c.Param("code")

	findBranch, err := services.FindExistBranchCode(code)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findBranch == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Branch code: %s not found.", code))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	deleteBranch, err := services.DeleteBranchByCode(code, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to delete branch: %v", err))
		return
	}

	utils.Success(c, "Branch deleted successfully", deleteBranch)
}
