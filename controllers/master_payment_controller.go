package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// GetMasterPaymentAll godoc
// @Summary Get master payment all active
// @Tags Master Payment
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/master-payment [get]
func GetMasterPaymentList(c *gin.Context) {

	findAll, err := services.GetActiveMasterPayment()
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve master payment: %v", err))
		return
	}

	utils.Success(c, "Master payment retrieved successfully", findAll)
}
