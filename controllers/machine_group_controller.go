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

// GetMachineGroupAll godoc
// @Summary Get machine group all active
// @Tags Machine Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/machine-group [get]
func GetMachineGroupList(c *gin.Context) {

	findAll, err := services.GetActiveMachineGroup()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve machine group: %v", err))
		return
	}

	utils.Success(c, "Machine group retrieved successfully", findAll)
}

// @Summary Create a new bramachinench group
// @Tags Machine Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.MachineGroupDto true "Machine Group Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/machine-group [post]
func CreateMachineGroup(c *gin.Context) {
	var req models.MachineGroupDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	existMachineGroup, err := services.FindExistMachineGroupName(strings.TrimSpace(req.GroupName))
	if err == nil && existMachineGroup.ID != 0 {
		utils.Error(c, http.StatusBadRequest, "Machine group with this name already exists")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	machineGroup, err := services.CreateMachineGroup(req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create machine group: %v", err))
		return
	}

	machines := req.MachineList
	var subMachines []models.MachineSubGroup

	for _, machine := range machines {
		fmt.Println("Machine loop:", machine)

		subMachine := models.MachineSubGroup{
			MachineID:   machine.MachineID,
			GroupID:     &machineGroup.ID,
			PlayTime:    machine.PlayTime,
			EBonus:      machine.EBonus,
			ECoin:       machine.ECoin,
			MachineName: machine.MachineName,
			Category:    machine.Category,
		}

		subMachines = append(subMachines, subMachine)
	}

	if len(subMachines) > 0 {
		_, err := services.CreateBatchMachineSubGroup(subMachines, userId)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create machine sub group: %v", err))
			return
		}
	}

	utils.Success(c, "Machine group created successfully", machineGroup)
}

// UpdateMachineGroup godoc
// @Summary Update machine group by ID
// @Tags Machine Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.MachineGroupDto true "Machine group Data"
// @Param   id  path  int  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/machine-group/{id} [put]
func UpdateMachineGroup(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	var req models.MachineGroupDto
	fmt.Printf("%+v\n", req)

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	findMachineGroup, err := services.FindMachineGroupById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findMachineGroup == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Machine group id: %s not found.", id))
		return
	}

	existMachineGroup, err := services.FindExistMachineGroupNameNotCurrent(req.GroupName, findMachineGroup.ID)
	if err == nil && existMachineGroup != nil && existMachineGroup.ID != 0 {
		utils.Error(c, http.StatusBadRequest, "Machine group with this name already exists other record")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateMachineGroup, err := services.UpdateMachineGroup(findMachineGroup.ID, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update machine group: %v", err))
		return
	}
	//Delete existing sub-branch groups
	_, err = services.DeleteMachineSubGroupByGroupId(findMachineGroup.ID, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete existing sub-machine groups: %v", err))
		return
	}
	// Update sub-branch groups
	machines := req.MachineList
	var subMachines []models.MachineSubGroup
	for _, machine := range machines {
		fmt.Println("Machine loop:", machine)
		subMachine := models.MachineSubGroup{
			MachineID:   machine.MachineID,
			GroupID:     &updateMachineGroup.ID,
			PlayTime:    machine.PlayTime,
			EBonus:      machine.EBonus,
			ECoin:       machine.ECoin,
			MachineName: machine.MachineName,
			Category:    machine.Category,
		}
		subMachines = append(subMachines, subMachine)
	}
	if len(subMachines) > 0 {
		_, err := services.CreateBatchMachineSubGroup(subMachines, userId)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create machine sub group: %v", err))
			return
		}
	}

	utils.Success(c, "Machine group updated successfully", updateMachineGroup)
}

// GetMachineGroupById godoc
// @Summary Get machine group by id
// @Tags Machine Group
// @Security BasicAuth
// @Security BearerAuth
// @Param   id  path  int  true  "Id"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/machine-group/{id} [get]
func GetMachineGroupById(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findMachineGroup, err := services.FindMachineGroupById(dInt)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve machine group: %v", err))
		return
	}

	utils.Success(c, "Machine group retrieved successfully", findMachineGroup)
}

// SearchMachineGroupList godoc
// @Summary Get machine group list with paggination
// @Tags Machine Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search text"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/machine-group/list [get]
func SearchMachineGroupList(c *gin.Context) {

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
	result, err := services.MachineGroupSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve machine group: %v", err))
		return
	}

	utils.Success(c, "Machine group retrieved successfully", result)
}

// DeleteMachineGroupById godoc
// @Summary Delete machine group by ID
// @Tags Machine Group
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/machine-group/{id} [delete]
func DeleteMachineGroupById(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findMachineGroup, err := services.FindMachineGroupById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findMachineGroup == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Machine group  id: %s not found.", id))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	deleteMachineGroup, err := services.DeleteMachineGroupById(dInt, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to delete machine group: %v", err))
		return
	}

	utils.Success(c, "Machine group deleted successfully", deleteMachineGroup)
}

// GetMachineGroupByIdV1 godoc
// @Summary Get machine group by id v1
// @Tags Machine Group
// @Security BasicAuth
// @Security BearerAuth
// @Param   id  path  int  true  "Id"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/machine-group/v1/{id} [get]
func GetMachineGroupByIdV1(c *gin.Context) {
	id := c.Param("id")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findMachineGroup, err := services.FindMachineGroupByIdV1(dInt)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve machine group: %v", err))
		return
	}

	utils.Success(c, "Machine group retrieved successfully", findMachineGroup)
}

// GetMachineSubGroupByGroupId godoc
// @Summary Get machine sub group by group id
// @Tags Machine Group
// @Security BasicAuth
// @Security BearerAuth
// @Param   groupId  path  int  true  "groupId"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/machine-group/sub/{groupId} [get]
func GetMachineSubGroupByGroupId(c *gin.Context) {
	id := c.Param("groupId")
	dInt, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Invalid id")
		return
	}

	findMachineSubGroup, err := services.FindMachineSubGroupByGroupId(dInt)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve machine sub group: %v", err))
		return
	}

	utils.Success(c, "Machine sub group retrieved successfully", findMachineSubGroup)
}
