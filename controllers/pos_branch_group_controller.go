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

// GetBranchGroupAll godoc
// @Summary Get branch group all active
// @Tags Branch Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch-group [get]
func GetBranchGroupList(c *gin.Context) {

	findAll, err := services.GetActiveBranchGroup()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve branch group: %v", err))
		return
	}

	utils.Success(c, "Branch group retrieved successfully", findAll)
}

// @Summary Create a new branch group
// @Tags Branch Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.PosBranchGroupDto true "Branch Group Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch-group [post]
func CreateBranchGroup(c *gin.Context) {
	var req models.PosBranchGroupDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	existBranchGroup, err := services.FindExistBranchGroupName(strings.TrimSpace(req.GroupName))
	if err == nil && existBranchGroup.ID != 0 {
		utils.Error(c, http.StatusBadRequest, "Branch group with this name already exists")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	branchGroup, err := services.CreateBranchGroup(req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create branch group: %v", err))
		return
	}

	branchs := req.BranchList
	var subBranchs []models.PosBranchSubGroup

	for _, branch := range branchs {
		fmt.Println("Branch loop:", branch)

		subBranch := models.PosBranchSubGroup{
			LocationCode: branch,
			GroupID:      &branchGroup.ID,
		}

		subBranchs = append(subBranchs, subBranch)
	}

	if len(subBranchs) > 0 {
		_, err := services.CreateBatchBranchSubGroup(subBranchs, userId)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create branch sub group: %v", err))
			return
		}
	}

	utils.Success(c, "Branch group created successfully", branchGroup)
}

// UpdateBranchGroup godoc
// @Summary Update branch group by ID
// @Tags Branch Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.PosBranchGroupDto true "Branch group Data"
// @Param   id  path  int  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch-group/{id} [put]
func UpdateBranchGroup(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	var req models.PosBranchGroupDto
	fmt.Printf("%+v\n", req)

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	findBranchGroup, err := services.FindBranchGroupById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findBranchGroup == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Branch group id: %s not found.", id))
		return
	}

	existBranchGroup, err := services.FindExistBranchGroupNameNotCurrent(req.GroupName, findBranchGroup.ID)
	if err == nil && existBranchGroup != nil && existBranchGroup.ID != 0 {
		utils.Error(c, http.StatusBadRequest, "Branch group with this name already exists other record")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateBranchGroup, err := services.UpdateBranchGroup(findBranchGroup.ID, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update branch group: %v", err))
		return
	}
	//Delete existing sub-branch groups
	_, err = services.DeleteBranchSubGroupByGroupId(findBranchGroup.ID, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete existing sub-branch groups: %v", err))
		return
	}
	// Update sub-branch groups
	branchs := req.BranchList
	var subBranchs []models.PosBranchSubGroup
	for _, branch := range branchs {
		fmt.Println("Branch loop:", branch)
		subBranch := models.PosBranchSubGroup{
			LocationCode: branch,
			GroupID:      &updateBranchGroup.ID,
		}
		subBranchs = append(subBranchs, subBranch)
	}
	if len(subBranchs) > 0 {
		_, err := services.CreateBatchBranchSubGroup(subBranchs, userId)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create branch sub group: %v", err))
			return
		}
	}

	utils.Success(c, "Branch group updated successfully", updateBranchGroup)
}

// GetBranchgroupById godoc
// @Summary Get branch group by id
// @Tags Branch Group
// @Security BasicAuth
// @Security BearerAuth
// @Param   id  path  int  true  "Id"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch-group/{id} [get]
func GetBranchGroupById(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findBranchGroup, err := services.FindBranchGroupById(dInt)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve branch group: %v", err))
		return
	}

	utils.Success(c, "Branch group retrieved successfully", findBranchGroup)
}

// SearchBranchGroupList godoc
// @Summary Get branch group list with paggination
// @Tags Branch Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search text"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch-group/list [get]
func SearchBranchGroupList(c *gin.Context) {

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
	result, err := services.BranchGroupSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve branch group: %v", err))
		return
	}

	utils.Success(c, "Branch group retrieved successfully", result)
}

// DeleteBranchgroupById godoc
// @Summary Delete branch group by ID
// @Tags Branch Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/branch-group/{id} [delete]
func DeleteBranchGroupById(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findBranchGroup, err := services.FindBranchGroupById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findBranchGroup == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Branch group  id: %s not found.", id))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	deleteBranchGroup, err := services.DeleteBranchGroupById(dInt, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to delete branch group: %v", err))
		return
	}

	utils.Success(c, "Branch group deleted successfully", deleteBranchGroup)
}
