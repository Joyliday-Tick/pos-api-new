package controllers

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// @Summary Create a new refund confirm
// @Tags Refund Confirm
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.RefundComfirmDto true "Refund Confirm Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/refund-confirm [post]
func CreateRefundConfirm(c *gin.Context) {
	var req models.RefundComfirmDto
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	refundConfirm, err := services.CreateRefundConfirm(req)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to create refund confirm: %v", err))
		return
	}

	cardDeposits, err := services.FindCardDepositByCardNo(req.CardNo)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to find card deposit with card no: %s : %v", req.CardNo, err))
		return
	}

	// ใช้ WaitGroup เพื่อรอให้ทั้งสองส่วนเสร็จพร้อมกัน
	var (
		allErrors []string
		updateErr error
		mu        sync.Mutex
		wg        sync.WaitGroup
	)

	wg.Add(2)

	// -------------------------------------------------
	// 🧩 Parallel 1: Update Refund Request Status
	// -------------------------------------------------
	go func() {
		defer wg.Done()
		if err := services.UpdateRefundRequestById(req.ReqID); err != nil {
			mu.Lock()
			updateErr = fmt.Errorf("failed to update refund request: %v", err)
			mu.Unlock()
		}
	}()

	// -------------------------------------------------
	// 🧩 Parallel 2: Handle Card Withdraw & Update Balance
	// -------------------------------------------------
	go func() {
		defer wg.Done()

		deducts, updateCardDeposits := applyCardWithdrawRefund(req, cardDeposits)

		// สร้าง withdraw
		for _, wt := range deducts {
			if _, err := services.CreateCardWithdraw(wt, userId); err != nil {
				mu.Lock()
				allErrors = append(allErrors, fmt.Sprintf("Failed to create withdraw for CardDepositID %v: %v", wt.CardDepositId, err))
				mu.Unlock()
			}
		}

		// update balance
		for _, cd := range updateCardDeposits {
			if _, err := services.UpdateCardDepositBalance(cd); err != nil {
				mu.Lock()
				allErrors = append(allErrors, fmt.Sprintf("Failed to update balance for CardDepositID %v: %v", cd.ID, err))
				mu.Unlock()
			}
		}
	}()

	// -------------------------------------------------
	// ✅ รอให้ทั้งสองส่วนเสร็จ
	// -------------------------------------------------
	wg.Wait()

	// ตรวจ error
	if updateErr != nil {
		utils.Error(c, http.StatusInternalServerError, updateErr.Error())
		return
	}

	if len(allErrors) > 0 {
		utils.Error(c, http.StatusBadRequest, strings.Join(allErrors, "; "))
		return
	}

	utils.Success(c, "Refund confirm created successfully", refundConfirm)
}

func applyCardWithdrawRefund(req models.RefundComfirmDto, deposits []models.CardDeposit) ([]models.CardWithdrawDto, []models.CardDepositBalanceDto) {
	var deducts []models.CardWithdrawDto
	var updateCardDeposits []models.CardDepositBalanceDto

	var found *models.CardDeposit

	for i, cd := range deposits {
		if cd.BalanceCoin == req.RefundEcoin && cd.BalanceBonus == req.RefundEbonus {
			found = &deposits[i]
			break
		}
	}

	fmt.Println("Found exact match:", found)

	if found != nil {
		updateCardDeposits = append(updateCardDeposits, models.CardDepositBalanceDto{
			ID:           &found.ID,
			BalanceCoin:  req.RefundEcoin,
			BalanceBonus: req.RefundEbonus,
		})
		deducts = append(deducts, models.CardWithdrawDto{
			FromChannel:   "POS Refund",
			CardNo:        req.CardNo,
			MemberTel:     req.RefMemberTel,
			AmountEcoin:   req.RefundEcoin,
			AmountEbonus:  req.RefundEbonus,
			CardDepositId: &found.ID,
		})
	} else {
		fmt.Println("ไม่พบข้อมูลที่ตรงเงื่อนไข")
		e_coin := req.RefundEcoin
		e_bonus := req.RefundEbonus

		if e_bonus > 0 {
			sort.Slice(deposits, func(i, j int) bool {

				if deposits[i].BonusExpireDate == nil {
					return false
				}
				if deposits[j].BonusExpireDate == nil {
					return true
				}
				// เรียงจากวันที่เก่ากว่า -> ใหม่กว่า (ASC)
				return deposits[i].BonusExpireDate.Before(*deposits[j].BonusExpireDate)
			})
		}

		for _, cd := range deposits {
			if e_coin == 0 && e_bonus == 0 {
				break
			}
			useCoin := min(e_coin, cd.BalanceCoin)
			useBonus := min(e_bonus, cd.BalanceBonus)

			e_coin -= useCoin
			e_bonus -= useBonus
			updateCardDeposits = append(updateCardDeposits, models.CardDepositBalanceDto{
				ID:           &cd.ID,
				BalanceCoin:  useCoin,
				BalanceBonus: useBonus,
			})

			deducts = append(deducts, models.CardWithdrawDto{
				FromChannel:   "POS Refund",
				CardNo:        req.CardNo,
				MemberTel:     req.RefMemberTel,
				AmountEcoin:   useCoin,
				AmountEbonus:  useBonus,
				CardDepositId: &cd.ID,
			})
		}

	}

	return deducts, updateCardDeposits
}
