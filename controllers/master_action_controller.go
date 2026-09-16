package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// GetMasterActionAll godoc
// @Summary Get master action all active
// @Tags Master Action
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/master-action [get]
func GetMasterActionList(c *gin.Context) {

	findAll, err := services.GetActiveMasterAction()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve master action: %v", err))
		return
	}

	utils.Success(c, "Master action retrieved successfully", findAll)
}
