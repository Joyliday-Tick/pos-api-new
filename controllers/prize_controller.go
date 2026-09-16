package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// SearchPrizeCounter godoc
// @Summary Get Prize Counter Record Jubu Jibi
// @Tags Prize Counter
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param memberTel query string false "Search member tel"
// @Param location query string false "Search location"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/prize-counter/jubu-jibi [get]
func SearchPrizeCounterJubuJibi(c *gin.Context) {
	memberTel := c.Query("memberTel")
	location := c.Query("location")

	params := models.SearchPrizeCounterParams{
		MemberTel: memberTel,
		Location:  location,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.PrizeCounterJubuJibiList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve prize counter record list: %v", err))
		return
	}

	utils.Success(c, "Prize counter record list retrieved successfully", result)
}
