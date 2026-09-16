package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// SearchStampPointList godoc
// @Summary Get Stamp point List with mobile number and start date
// @Tags Stamp House
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param memberTel query string false "Search member tel"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/stamp-house/list [get]
func SearchStampPointList(c *gin.Context) {
	memberTel := c.Query("memberTel")

	params := models.SearchMeterRecordParams{
		MemberTel: memberTel,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.StampPointSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve stamp house list: %v", err))
		return
	}

	utils.Success(c, "Stamp house list retrieved successfully", result)
}
