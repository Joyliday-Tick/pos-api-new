package controllers

import (
	"fmt"
	"new-pos-api/services"
	"new-pos-api/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetCustomer godoc
// @Summary Get Customer crm by mobile number
// @Description Get Customer crm by mobile number
// @Tags CRM
// @Security BasicAuth
// @Security BearerAuth
// @Param   tel  path  string  true  "Telephone Number"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/crm/customer/{tel} [get]
func GetCustomerByTel(c *gin.Context) {
	tel := strings.TrimSpace(c.Param("tel"))
	fmt.Print("tel: ", tel)
	status, customer, err := services.GetCustomerByMobileNo(tel)
	if err != nil {
		utils.Error(c, status, fmt.Sprintf("Failed to get customer by tel: %s", err.Error()))
		return
	}

	utils.Success(c, "Customer retrieved successfully", customer)
}

// GetBranchCrm godoc
// @Summary Get Branch crm by code
// @Description Get Branch crm by code
// @Tags CRM
// @Security BasicAuth
// @Security BearerAuth
// @Param   branchCode  path  string  true  "Branch Code"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/crm/branch/{branchCode} [get]
func GetBranchCrmByCode(c *gin.Context) {
	branchCode := strings.TrimSpace(c.Param("branchCode"))
	fmt.Print("branchCode: ", branchCode)
	status, branch, err := services.GetBranchByCode(branchCode)
	if err != nil {
		utils.Error(c, status, fmt.Sprintf("Failed to get branch by code: %s", err.Error()))
		return
	}

	utils.Success(c, "Branch crm retrieved successfully", branch)
}

// GetScoreTypeCrm godoc
// @Summary Get score type crm
// @Description Get score type crm
// @Tags CRM
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/crm/score-type [get]
func GetScoreType(c *gin.Context) {
	status, scoreType, err := services.GetScoreType()
	if err != nil {
		utils.Error(c, status, fmt.Sprintf("Failed to get score type: %s", err.Error()))
		return
	}

	utils.Success(c, "Score type crm retrieved successfully", scoreType)
}
