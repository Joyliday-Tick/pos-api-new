package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// GetLogoAll godoc
// @Summary Get logo all active
// @Tags Logo
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/logo [get]
func GetLogoList(c *gin.Context) {

	findAll, err := services.GetActiveLogo()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve logo: %v", err))
		return
	}

	utils.Success(c, "Logo retrieved successfully", findAll)
}
