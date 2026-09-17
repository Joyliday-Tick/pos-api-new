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

// @Summary Create a new claim prize
// @Tags Claim Prize
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.ClaimPrizeDto true "Claim prize Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/claim-prize [post]
func CreateClaimPrize(c *gin.Context) {
	var req models.ClaimPrizeDto
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	if *req.OptionID == 0 {
		req.OptionID = nil
	}

	claim, err := services.CreateClaimPrize(req)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to create claim prize: %v", err))
		return
	}

	req.CPDate = claim.CPDate

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := services.UpdateMeterRecordJubuJibi(req.MemberTel); err != nil {
			errChan <- fmt.Errorf("update meter failed: %v", err)
		}
	}()

	//parallel update prize counter
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := services.UpdatePrizeCounterJubuJibi(req.MemberTel); err != nil {
			errChan <- fmt.Errorf("update prize failed: %v", err)
		}
	}()

	// parallel handle option
	wg.Add(1)
	go func() {
		defer wg.Done()
		if req.OptionID != nil && *req.OptionID != 0 {
			if err := handleClaimOption(*req.OptionID, req, userId, claim.CPID.String()); err != nil {
				errChan <- err
			}
		}
	}()

	//รอทุกงานเสร็จ
	wg.Wait()
	close(errChan)

	//ตรวจ error ทั้งหมด
	var allErrors []string
	for e := range errChan {
		allErrors = append(allErrors, e.Error())
	}

	if len(allErrors) > 0 {
		utils.Error(c, http.StatusInternalServerError, strings.Join(allErrors, " | "))
		return
	}

	//สำเร็จ
	utils.Success(c, "Claim prize created successfully", claim)
}

func handleClaimOption(option int, req models.ClaimPrizeDto, userId int, id string) error {

	switch option {

	case 1:
		member, err := services.UpdateJoylicoin(req.MemberTel, req.Joylicoin, userId)
		if err != nil {
			return fmt.Errorf("failed to update joylicoin: %v", err)
		}

		scores := models.ScoreMember{
			MobileNo:   member.Tel,
			Bonus:      member.Bonus,
			TotalPoint: member.TotalPoint,
			ECoin:      member.Ecoin,
			JubuJibi:   member.JubuJibi,
			FinWow:     member.Finwow,
			EStamp:     member.Estamp,
			MSkill1:    member.MSkill1,
			MSkill2:    member.MSkill2,
			MSkill3:    member.MSkill3,
			MSkill4:    member.MSkill4,
			MSkill5:    member.MSkill5,
		}
		if err := parallelSync(scores, req, id); err != nil {
			return err
		}
		return nil

	case 3:
		bonusExpire := utils.GetBonusExpire()
		deposit := models.CardDepositDto{
			FromChannel:     "POS Claim Prize",
			CardNo:          *req.CardNo,
			MemberTel:       req.MemberTel,
			Amount:          0,
			Coin:            req.ECoin,
			Bonus:           req.EBonus,
			BalanceCoin:     req.ECoin,
			BalanceBonus:    req.EBonus,
			BonusExpireDate: &bonusExpire,
		}

		resultDeposit, err := services.CreateCardDeposit(deposit, userId)
		if err != nil {
			return fmt.Errorf("failed to create card deposit: %v", err)
		}

		cardPlay := models.NewCardPlayDto{
			CardNo:        *req.CardNo,
			PlayBranch:    false,
			PlayMachine:   false,
			CardDepositId: &resultDeposit.ID, // จะถูกอัปเดตหลังจากสร้าง CardDeposit
		}
		_, err = services.CreateCardPlay(cardPlay, userId)
		if err != nil {
			return fmt.Errorf("failed to create card play: %v", err)
		}

		return nil

	case 4:
		deduct := req.Joylicoin
		if req.Joylicoin > 0 {
			deduct = -req.Joylicoin
		}

		member, err := services.UpdateJoylicoin(req.MemberTel, deduct, userId)
		if err != nil {
			return fmt.Errorf("failed to deduct joylicoin: %v", err)
		}

		scores := models.ScoreMember{
			MobileNo:   member.Tel,
			Bonus:      member.Bonus,
			TotalPoint: member.TotalPoint,
			ECoin:      member.Ecoin,
			JubuJibi:   member.JubuJibi,
			FinWow:     member.Finwow,
			EStamp:     member.Estamp,
			MSkill1:    member.MSkill1,
			MSkill2:    member.MSkill2,
			MSkill3:    member.MSkill3,
			MSkill4:    member.MSkill4,
			MSkill5:    member.MSkill5,
		}

		req.Joylicoin = deduct
		if err := parallelSync(scores, req, id); err != nil {
			return err
		}
		return nil

	case 5:
		cardDeposits, err := services.FindCardDepositByCardNo(*req.CardNo)
		if err != nil {
			return fmt.Errorf("failed to fetch card deposit: %v", err)
		}

		deducts, updates, err := applyCardWithdrawClaim(req, cardDeposits)
		if err != nil {
			return err
		}

		// เดิมยิง withdraw ทั้งชุดใน goroutine ชุดหนึ่ง และหักยอดทั้งชุดในอีกชุดหนึ่ง
		// ขนานกันโดยไม่รู้จักกันเลย ทั้งที่ deducts[i] กับ updates[i] คือรายการเดียวกัน
		// ถ้าหักยอดล้มเหลวแต่ withdraw สำเร็จ จะได้ร่องรอยว่า "หักไปแล้ว" ทั้งที่ยอดไม่ลด
		// และ error ทั้งหมดถูกเก็บไว้รายงานตอนท้าย ไม่มีใครหยุดใคร ทุกตัว commit หมด
		//
		// ตอนนี้ทำทีละคู่และเรียงกัน: หักยอดก่อน (เป็นตัวเงินจริง) แล้วค่อยลง withdraw
		// เจอ error เมื่อไหร่หยุดทันที ไม่เดินต่อ
		for i := range updates {
			if _, err := services.UpdateCardDepositBalance(updates[i]); err != nil {
				return fmt.Errorf("หักยอดในบัตรไม่สำเร็จ (รายการที่ %d จาก %d): %w",
					i+1, len(updates), err)
			}

			if _, err := services.CreateCardWithdraw(deducts[i], userId); err != nil {
				// ยอดถูกหักไปแล้วแต่ลงร่องรอยไม่ได้ ตัวเงินถูกต้องแต่ audit trail ขาด
				// ต้องดังไว้ ไม่ใช่กลืนหาย
				fmt.Printf("[CLAIM] card_no=%s deposit=%v: หักยอดสำเร็จแล้วแต่สร้าง card_withdraw "+
					"ไม่สำเร็จ: %v — ยอดในบัตรถูกต้องแล้ว แต่ไม่มีร่องรอยการหัก ต้องบันทึกย้อนหลัง\n",
					*req.CardNo, updates[i].ID, err)
				return fmt.Errorf("บันทึกร่องรอยการหักยอดไม่สำเร็จ (รายการที่ %d จาก %d): %w",
					i+1, len(updates), err)
			}
		}

		return nil

	default:
		return nil
	}
}

// applyCardWithdrawClaim คืน (รายการ withdraw, รายการหักยอด, error)
//
// deducts[i] กับ updateCardDeposits[i] เป็นคู่กันเสมอ เพราะถูก append
// ในรอบลูปเดียวกัน ผู้เรียกต้องจับคู่ตาม index และรันเรียงกัน
//
// error จะไม่เป็น nil เมื่อยอดในบัตรรวมกันแล้วไม่พอกับที่ขอแลก
// เดิมฟังก์ชันนี้วนหักเท่าที่มีแล้วจบ ไม่บอกใครว่าหักไม่ครบ ผู้เรียกก็ไม่ได้เช็ค
// ผลคือแลกของรางวัลที่ราคาสูงกว่ายอดในบัตรได้ โดยหักไปเท่าที่มี
func applyCardWithdrawClaim(req models.ClaimPrizeDto, deposits []models.CardDeposit) ([]models.CardWithdrawDto, []models.CardDepositBalanceDto, error) {
	var deducts []models.CardWithdrawDto
	var updateCardDeposits []models.CardDepositBalanceDto

	var found *models.CardDeposit

	for i, cd := range deposits {
		if cd.BalanceCoin == req.ECoin && cd.BalanceBonus == req.EBonus {
			found = &deposits[i]
			break
		}
	}

	fmt.Println("Found exact match:", found)

	if found != nil {
		updateCardDeposits = append(updateCardDeposits, models.CardDepositBalanceDto{
			ID:           &found.ID,
			BalanceCoin:  req.ECoin,
			BalanceBonus: req.EBonus,
		})
		deducts = append(deducts, models.CardWithdrawDto{
			FromChannel:   "POS Claim Prize",
			CardNo:        *req.CardNo,
			MemberTel:     req.MemberTel,
			AmountEcoin:   req.ECoin,
			AmountEbonus:  req.EBonus,
			CardDepositId: &found.ID,
		})
	} else {
		fmt.Println("ไม่พบข้อมูลที่ตรงเงื่อนไข")
		e_coin := req.ECoin
		e_bonus := req.EBonus

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
				FromChannel:   "POS Claim Prize",
				CardNo:        *req.CardNo,
				MemberTel:     req.MemberTel,
				AmountEcoin:   useCoin,
				AmountEbonus:  useBonus,
				CardDepositId: &cd.ID,
			})
		}

		// เหลือค้างแปลว่ายอดในบัตรไม่พอ ต้องไม่ให้ทำรายการต่อ
		if e_coin > 0 || e_bonus > 0 {
			return nil, nil, fmt.Errorf(
				"ยอดในบัตร %s ไม่พอสำหรับการแลกครั้งนี้ (ขาด e_coin %d, e_bonus %d)",
				*req.CardNo, e_coin, e_bonus)
		}
	}

	return deducts, updateCardDeposits, nil
}
func parallelSync(scores models.ScoreMember, req models.ClaimPrizeDto, id string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// ✅ Task 1: SyncScoreMember
	wg.Add(1)
	go func() {
		defer wg.Done()
		statusCode, _, err := services.SyncScoreMember(scores)
		if err != nil {
			errChan <- fmt.Errorf("failed to update member crm: %w", err)
			return
		}
		if statusCode != http.StatusOK {
			errChan <- fmt.Errorf("failed to update member crm non-200 status: %d", statusCode)
		}
	}()

	// ✅ Task 2: Sync History
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := services.ClaimJoylicoinSyncHistory(req, id)
		if err != nil {
			errChan <- fmt.Errorf("failed to sync history to CRM: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		var errs []string
		for e := range errChan {
			errs = append(errs, e.Error())
		}
		return fmt.Errorf("%s", strings.Join(errs, " | "))
	}

	return nil
}
