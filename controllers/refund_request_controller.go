package controllers

import (
	"fmt"
	"net/http"
	"sync"

	// "new-pos-api/middlewares"
	// "new-pos-api/models"

	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"

	// "strconv"

	"github.com/gin-gonic/gin"
	// "github.com/google/uuid"
)

// GetRefundRequestByTel godoc
// @Summary Get refund requests by telephone number
// @Tags Refund Request
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param tel path string true "Telephone number"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/refund-request/{tel} [get]
func GetRefundRequestByTel(c *gin.Context) {
	tel := c.Param("tel")

	fmt.Println("Telephone:", tel)

	findRefundRequest, err := services.RefundRequestByTel(tel)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve refund request: %v", err))
		return
	}

	utils.Success(c, "Refund request retrieved successfully", findRefundRequest)
}

// @Summary Create a new refund request
// @Tags Refund Request
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.RefundRequestDto true "Refund Request Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/refund-request [post]
func CreateRefundRequest(c *gin.Context) {
	var req models.RefundRequestDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	// Step 1: สร้าง Refund Request
	refundRequest, err := services.CreateRefundRequest(req)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to create refund request: %v", err))
		return
	}

	// Step 2: เตรียมข้อมูล parallel tasks
	bonusExpire := utils.GetBonusExpire()
	deposit := models.CardDepositDto{
		FromChannel:     "POS Refund",
		CardNo:          req.CardNo,
		MemberTel:       req.RefundTel,
		Amount:          0,
		Coin:            req.RefundEcoin,
		Bonus:           req.RefundEbonus,
		BalanceCoin:     req.RefundEcoin,
		BalanceBonus:    req.RefundEbonus,
		PosId:           "",
		PosMenuId:       0,
		BonusExpireDate: &bonusExpire,
		CardExpireDate:  nil,
	}

	cardPlay := models.NewCardPlayDto{
		CardNo:        req.CardNo,
		PlayBranch:    false,
		PlayMachine:   false,
		CardDepositId: nil, // จะถูกอัปเดตหลังจากสร้าง CardDeposit
	}

	// Step 3: Run parallel tasks
	var wg sync.WaitGroup
	errChan := make(chan error, 2) // เก็บ error ได้สูงสุด 2 ตัว

	wg.Add(2)

	// Task 1: update meter record
	go func() {
		defer wg.Done()
		if err := services.UpdateMeterRecordJubuJibi(req.RefundTel); err != nil {
			errChan <- fmt.Errorf("update meter record failed: %w", err)
		}
	}()

	// Task 2: create card deposit
	go func() {
		defer wg.Done()

		resultDeposit, err := services.CreateCardDeposit(deposit, userId)
		if err != nil {
			errChan <- fmt.Errorf("create card deposit failed: %w", err)
			return
		}

		// create card play after creating card deposit
		cardPlay.CardDepositId = &resultDeposit.ID

		if _, err := services.CreateCardPlay(cardPlay, userId); err != nil {
			errChan <- fmt.Errorf("create card play failed: %w", err)
			return
		}
	}()

	wg.Wait()
	close(errChan)

	// ตรวจสอบว่ามี error จาก goroutines หรือไม่
	for e := range errChan {
		if e != nil {
			utils.Error(c, http.StatusInternalServerError, e.Error())
			return
		}
	}

	utils.Success(c, "Refund request created successfully", refundRequest)
}
