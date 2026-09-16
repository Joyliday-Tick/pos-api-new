package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// GetBonusSetting godoc
// @Summary Get bonus setting
// @Tags Bonus Setting
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/bonus-setting [get]
func GetBonusSetting(c *gin.Context) {

	find, err := services.GetBonusSetting()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve bonus setting: %v", err))
		return
	}

	utils.Success(c, "Bonus setting retrieved successfully", find)
}
