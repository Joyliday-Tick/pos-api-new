package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// GetStampMachineAll godoc
// @Summary Get stamp machine all active
// @Tags Stamp Machine
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/stamp-machine [get]
func GetStampMachineList(c *gin.Context) {

	findAll, err := services.GetActiveStampMachine()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve stamp machine: %v", err))
		return
	}

	utils.Success(c, "Stamp machine retrieved successfully", findAll)
}
