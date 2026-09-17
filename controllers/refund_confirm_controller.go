package controllers

import (
	"fmt"
	"net/http"
	"sort"

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

	// ---------------------------------------------------------------------
	// กันยืนยันคืนเงินซ้ำ — ต้องอยู่ก่อนทุกอย่าง
	//
	// เดิมไม่มีที่ไหนเช็คเลยว่าคำขอนี้ยืนยันไปแล้วหรือยัง UpdateRefundRequestById
	// ตั้ง refund_sts = 'N' ตอนจบ แต่ไม่มีใครอ่านค่านั้นก่อนทำรายการ
	// ยิงซ้ำจึงหักยอดในบัตรซ้ำได้ทันที รูปแบบเดียวกับ void ซ้ำ (C2)
	// ---------------------------------------------------------------------
	refundStatus, err := services.FindRefundRequestStatusByID(req.ReqID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to read refund request: %v", err))
		return
	}
	if refundStatus == "" {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("ไม่พบคำขอคืนเงิน id: %s", req.ReqID))
		return
	}
	if refundStatus != "Y" {
		utils.Error(c, http.StatusConflict, fmt.Sprintf(
			"คำขอคืนเงิน %s ถูกยืนยันไปแล้ว — ไม่ต้องยืนยันซ้ำ "+
				"ถ้ายอดในบัตรยังดูไม่ถูกต้อง ให้ตรวจสอบก่อน อย่ากดยืนยันอีก", req.ReqID))
		return
	}

	cardDeposits, err := services.FindCardDepositByCardNo(req.CardNo)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to find card deposit with card no: %s : %v", req.CardNo, err))
		return
	}

	// คำนวณรายการหักให้เสร็จก่อนแตะข้อมูล ยอดไม่พอจะได้ไม่ต้องเขียนอะไรเลย
	deducts, updateCardDeposits, err := applyCardWithdrawRefund(req, cardDeposits)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// จองคำขอด้วยการปิดสถานะก่อนลงมือ กันยิงซ้ำ/ยิงพร้อมกัน
	// ถ้าขั้นหักยอดพังหลังจากนี้ คำขอจะถูกล็อกไว้และการยืนยันรอบหน้าจะโดน 409
	// ตั้งใจให้ fail closed แบบเดียวกับ void — ต้องให้คนมาตรวจ ไม่ใช่ปล่อยให้ retry
	if err := services.UpdateRefundRequestById(req.ReqID); err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update refund request: %v", err))
		return
	}

	refundConfirm, err := services.CreateRefundConfirm(req)
	if err != nil {
		fmt.Printf("[REFUND] req_id=%s card_no=%s: ปิดสถานะคำขอแล้วแต่สร้าง refund_confirm ไม่สำเร็จ: %v "+
			"— ยังไม่ได้หักยอดในบัตร ต้องเปิดสถานะคำขอกลับด้วยมือถ้าจะให้ทำใหม่\n",
			req.ReqID, req.CardNo, err)
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to create refund confirm: %v", err))
		return
	}

	// ---------------------------------------------------------------------
	// หักยอดในบัตร — ทำทีละคู่และเรียงกัน
	//
	// เดิมส่วนนี้เป็น goroutine ขนานสองตัว: ตัวหนึ่งปิดสถานะคำขอ อีกตัวหักยอด
	// ทั้งคู่วิ่งจนจบไม่ว่าอีกฝั่งจะพังหรือไม่ และข้างในตัวหักยอดก็ยังวนสร้าง
	// withdraw ครบทุกแถวก่อน แล้วค่อยวนหักยอด โดยเก็บ error ใส่ slice แล้ววนต่อ
	//
	// สภาพที่เกิดได้เมื่อ guard ของ UpdateCardDepositBalance คืน error:
	// แถว refund_confirm ถูกเขียน คำขอถูกปิด ร่องรอย withdraw ครบ
	// แต่ coin ยังอยู่บนบัตร — เงินสดออกจากลิ้นชักแต่บัตรไม่ถูกหัก
	//
	// ตอนนี้: หักยอดก่อน (ตัวเงินจริง) แล้วค่อยลง withdraw เจอ error หยุดทันที
	// รูปแบบเดียวกับที่แก้ไปแล้วใน pos_void และ claim_prize
	// ---------------------------------------------------------------------
	for i := range updateCardDeposits {
		if _, err := services.UpdateCardDepositBalance(updateCardDeposits[i]); err != nil {
			fmt.Printf("[REFUND] req_id=%s card_no=%s: หักยอดไม่สำเร็จที่รายการ %d จาก %d: %v "+
				"— คำขอถูกปิดไปแล้ว การยืนยันรอบหน้าจะโดน 409 ต้องตรวจว่าหักไปถึงไหน\n",
				req.ReqID, req.CardNo, i+1, len(updateCardDeposits), err)
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf(
				"หักยอดในบัตรไม่สำเร็จ (รายการที่ %d จาก %d): %v "+
					"(คำขอ %s: ยอดถูกหักไปบางส่วนแล้ว ห้ามยืนยันซ้ำ แจ้งผู้ดูแลระบบ)",
				i+1, len(updateCardDeposits), err, req.ReqID))
			return
		}

		if _, err := services.CreateCardWithdraw(deducts[i], userId); err != nil {
			// ยอดถูกหักแล้วแต่ลงร่องรอยไม่ได้ ตัวเงินถูกต้องแต่ audit trail ขาด
			fmt.Printf("[REFUND] req_id=%s card_no=%s deposit=%v: หักยอดสำเร็จแล้วแต่สร้าง "+
				"card_withdraw ไม่สำเร็จ: %v — ยอดในบัตรถูกต้องแล้ว แต่ไม่มีร่องรอยการหัก\n",
				req.ReqID, req.CardNo, updateCardDeposits[i].ID, err)
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf(
				"บันทึกร่องรอยการหักยอดไม่สำเร็จ (รายการที่ %d จาก %d): %v "+
					"(คำขอ %s: ห้ามยืนยันซ้ำ)", i+1, len(updateCardDeposits), err, req.ReqID))
			return
		}
	}

	utils.Success(c, "Refund confirm created successfully", refundConfirm)
}

// applyCardWithdrawRefund คืน (รายการ withdraw, รายการหักยอด, error)
//
// deducts[i] กับ updateCardDeposits[i] เป็นคู่กันเสมอ เพราะถูก append
// ในรอบลูปเดียวกัน ผู้เรียกต้องจับคู่ตาม index และรันเรียงกัน
//
// error จะไม่เป็น nil เมื่อยอดในบัตรรวมกันไม่พอกับที่ขอคืน
// เดิมวนหักเท่าที่มีแล้วจบ ไม่บอกใครว่าหักไม่ครบ ผู้เรียกก็ไม่ได้เช็ค
// ผลคือคืนเงินสดเต็มจำนวนออกจากลิ้นชัก แต่หักออกจากบัตรได้ไม่ครบ
func applyCardWithdrawRefund(req models.RefundComfirmDto, deposits []models.CardDeposit) ([]models.CardWithdrawDto, []models.CardDepositBalanceDto, error) {
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

		// เหลือค้างแปลว่ายอดในบัตรไม่พอกับที่ขอคืน ต้องไม่ให้ทำรายการต่อ
		if e_coin > 0 || e_bonus > 0 {
			return nil, nil, fmt.Errorf(
				"ยอดในบัตร %s ไม่พอสำหรับการคืนเงินครั้งนี้ (ขาด e_coin %d, e_bonus %d)",
				req.CardNo, e_coin, e_bonus)
		}
	}

	return deducts, updateCardDeposits, nil
}
