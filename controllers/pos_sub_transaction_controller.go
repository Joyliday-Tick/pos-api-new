package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// GetPosSubByBillNo godoc
// @Summary Get pos sub transaction by bill no
// @Tags POS Sub Transaction
// @Security BasicAuth
// @Security BearerAuth
// @Param   billNo  path  string  true  "Bill No"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-sub-transaction/{billNo} [get]
func GetPosSubTransactionByBillNo(c *gin.Context) {
	billNo := c.Param("billNo")
	fmt.Println("billNo", billNo)

	if billNo == "" {
		utils.Error(c, http.StatusBadRequest, "bill no is required.")
		return
	}

	findPosSub, err := services.FindSubTransactionByBillNo(billNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve sub transaction: %v", err))
		return
	}

	utils.Success(c, "Pos sub transaction retrieved successfully", findPosSub)
}
