package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// GetDepositCronByTel godoc
// @Summary Get deposit head cron by tel
// @Tags Deposit Head
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param tel path string true "Telephone number"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/deposit-cron/{tel} [get]
func GetDepositCronByTel(c *gin.Context) {
	tel := c.Param("tel")

	// เดิม fmt.Println("Telephone:", tel) — พิมพ์เบอร์ลูกค้าเต็มหมายเลขลง log
	// ทุกครั้งที่มีการเรียก โดยไม่ได้ใช้ทำอะไร ลบทิ้ง

	findDeposit, err := services.GetDepositHeadByTel(tel)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve deposit cron: %v", err))
		return
	}

	utils.Success(c, "Deposit cron retrieved successfully", findDeposit)
}
