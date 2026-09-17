package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// @Summary Create a new pos void
// @Tags POS Void
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.PosVoidDto true "Pos Void Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-void [post]
func CreatePosVoid(c *gin.Context) {
	var req models.PosVoidDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
		return
	}

	existTransaction, err := services.FindExistPosTransactionBillNo(req.BillNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to find pos transaction: %w", err).Error())
		return
	}
	if existTransaction == nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("billNo %s not found", req.BillNo))
		return
	}

	// ---------------------------------------------------------------------
	// เช็คว่าบิลนี้ถูก void ไปแล้วหรือยัง — ต้องอยู่ตรงนี้ ก่อนทุกอย่าง
	//
	// เดิมเช็คนี้อยู่ข้างใน goroutine ตัวที่ 1 เท่านั้น และผลของมันถูกใช้แค่
	// ตัดสินใจว่าจะสร้างแถว pos_void ไหม ส่วน goroutine 2/3/4 ไม่เคยเห็นค่านั้นเลย
	// กด void ซ้ำจึงได้: ไม่สร้าง pos_void ซ้ำ (ดูเหมือนปลอดภัย) แต่ goroutine 3
	// ยังรัน balance_coin = balance_coin - 500 บนบัตรที่เหลือ 0 อยู่แล้ว -> ติดลบ
	// และ goroutine 4 ยังหักคะแนนสมาชิกซ้ำอีกรอบ
	//
	// เกิดง่ายมากเพราะ goroutine 4 คุยกับ CRM ข้างนอก ถ้า CRM timeout g.Wait()
	// จะคืน error -> POS ขึ้น 500 "ล้มเหลว" ทั้งที่ยอดถูกหักไปเรียบร้อยแล้ว
	// แคชเชียร์เห็นล้มเหลวก็กดซ้ำ กดสามครั้งยอดติดลบสองเท่า
	// ---------------------------------------------------------------------
	existVoid, err := services.FindExistVoidByBillNo(req.BillNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to find pos void: %w", err).Error())
		return
	}
	if existVoid != nil {
		// 409 ไม่ใช่ 500 — นี่ไม่ใช่ความผิดพลาดของระบบ แต่เป็นบิลที่ยกเลิกไปแล้ว
		// ต้องบอกให้ชัดเพื่อไม่ให้แคชเชียร์กดซ้ำอีก
		utils.Error(c, http.StatusConflict, fmt.Sprintf(
			"บิล %s ถูกยกเลิกไปแล้วเมื่อ %s โดย %s — ไม่ต้องยกเลิกซ้ำ "+
				"ถ้ายอดในบัตรยังดูไม่ถูกต้อง ให้ตรวจสอบก่อน อย่ากดยกเลิกอีก",
			req.BillNo,
			existVoid.VoidDate.Format("2006-01-02 15:04:05"),
			existVoid.VoidUser))
		return
	}

	existSub, err := services.FindSubTransactionByBillNo(req.BillNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to find pos sub transaction: %w", err).Error())
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Errorf("unauthorized: %w", err).Error())
		return
	}

	cardMemberTel := ""
	cardType := ""
	var cardTimePlay models.EtimesDto
	packageEcoin, packcageEbonus, packagePlayTime := 0, 0, 0
	if req.CardNo != "" {
		registeredCard, err := services.FindActiveCardNo(req.CardNo)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "failed to query card: "+err.Error())
			return
		}
		if registeredCard == nil {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("card number %s is not registered", req.CardNo))
			return
		}
		cardMemberTel = registeredCard.MemberTel

		checkCardType, err := services.FindActiveCardNoWithCardType(req.CardNo)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card type: %v", err))
			return
		}

		cardType = checkCardType.CardTypeName
		// fmt.Println("Card type:", cardType)
		if req.DeductCard {
			fmt.Println("Deduct card")
			if cardType == "Time play" {
				cardPlayTime, err := handleCardPlayTimeVoidProcess(req)
				// fmt.Println("Time play", cardPlayTime)
				if err != nil {
					utils.Error(c, http.StatusBadRequest, err.Error())
					return
				}

				if cardPlayTime != nil {
					cardTimePlay = *cardPlayTime
				}
			}
			if cardType == "Package" {
				packageEcoin, packcageEbonus, packagePlayTime, err = handleCardPackageVoidProcess(req, existSub)
				// fmt.Println("Package", packageEcoin, packcageEbonus, packagePlayTime)
				if err != nil {
					utils.Error(c, http.StatusBadRequest, err.Error())
					return
				}
			}
		} else {
			req.VoidReason = "U-" + req.VoidReason
		}

		fmt.Println("Not deduct card")
	} else {
		req.VoidReason = "U-" + req.VoidReason
	}

	// ---------------------------------------------------------------------
	// เดิมสี่ขั้นตอนนี้รันขนานกันด้วย errgroup.Group (ไม่ใช่ WithContext)
	// ซึ่งแปลว่า goroutine ตัวหนึ่งพังไม่ได้ยกเลิกตัวอื่น ทุกตัววิ่งจนจบและ commit หมด
	// ผลคือ: ขั้นหักยอด deposit พังกลางคัน (คืนได้ 1 จาก 3 ก้อน) แต่บิลยังถูก mark
	// เป็น VOID และคะแนน CRM ก็ถูกหักไปแล้ว บัญชีบอกว่ากลับรายการครบ
	// ส่วนลูกค้ายังถือ coin ค้างอยู่
	//
	// เปลี่ยนเป็นรันตามลำดับ ขั้นไหนพัง ขั้นถัดไปไม่ทำงาน
	// ความขนานตรงนี้ไม่ได้ช่วยอะไรเลย (คนละไม่กี่ query) แต่แลกมาด้วยความถูกต้อง
	//
	// ลำดับสำคัญ:
	//   1. จองบิลด้วย pos_void ก่อน   กันยิงซ้ำ/ยิงพร้อมกัน
	//   2. คืนยอดในบัตร               ส่วนที่เป็นเงิน
	//   3. mark บิลเป็น VOID
	//   4. คะแนน CRM + ลบประวัติ      ระบบข้างนอก rollback ไม่ได้ จึงไว้ท้ายสุด
	//
	// ถ้าขั้น 2 พัง จะถอนการจองขั้น 1 คืน เพื่อให้ยกเลิกใหม่ได้สะอาด
	// ถ้าขั้น 3 หรือ 4 พัง จะไม่ถอน เพราะยอดถูกคืนไปแล้ว การยกเลิกซ้ำจะหักซ้ำ
	// ---------------------------------------------------------------------

	// ขั้น 1: จองบิล
	posVoid, err := services.CreatePosVoid(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to create pos void: %w", err).Error())
		return
	}

	// ถอนการจองคืนเมื่อขั้นคืนยอดล้มเหลว
	releaseClaim := func(reason string) {
		if delErr := services.DeletePosVoidByID(posVoid.VoidID); delErr != nil {
			fmt.Printf("[VOID] bill_no=%s: %s และถอนการจอง pos_void ไม่สำเร็จด้วย: %v "+
				"— ต้องลบแถว void_id=%s ด้วยมือ ไม่งั้นบิลนี้จะยกเลิกใหม่ไม่ได้\n",
				req.BillNo, reason, delErr, posVoid.VoidID)
		}
	}

	// ขั้น 2: คืนยอดในบัตร
	if err := func() error {
		if req.CardNo != "" {
			switch cardType {
			case "Time play":
				fmt.Println("time play")
				var playTime int
				if cardTimePlay.PlayTime != 0 {
					playTime = cardTimePlay.PlayTime
				} else {
					playTime = 0
				}
				clearCard := models.ClearCardDto{
					CardNo:        req.CardNo,
					MemberTel:     cardMemberTel,
					BalanceEbonus: 0,
					BalanceEcoin:  0,
					CreatedBy:     req.VoidUser,
					Location:      existTransaction.BillLocation,
					TimePlay:      playTime,
					CardType:      cardType,
				}
				if _, err := services.ClearCard(clearCard, userId); err != nil {
					return fmt.Errorf("failed to clear card: %w", err)
				}
			case "Package":
				fmt.Println("package")
				fmt.Println("Package", packageEcoin, packcageEbonus, packagePlayTime)
				clearCard := models.ClearCardDto{
					CardNo:        req.CardNo,
					MemberTel:     cardMemberTel,
					BalanceEbonus: packcageEbonus,
					BalanceEcoin:  packageEcoin,
					CreatedBy:     req.VoidUser,
					Location:      existTransaction.BillLocation,
					TimePlay:      packagePlayTime,
					CardType:      cardType,
				}
				if _, err := services.ClearCard(clearCard, userId); err != nil {
					return fmt.Errorf("failed to clear card: %w", err)
				}

			default:
				fmt.Println("other")
				if req.DeductCard {
					fmt.Println("other deduct card")
					if err := handleCardVoidProcess(req, existTransaction, existSub, userId, cardMemberTel, cardType); err != nil {
						return fmt.Errorf("failed to void card process: %w", err)
					}
				}
			}

		} else {
			if err := handleNotCardVoidProcess(req, existTransaction, existSub, userId); err != nil {
				return fmt.Errorf("failed to void no card process: %w", err)
			}
		}
		return nil
	}(); err != nil {
		// ยอดยังไม่ถูกคืน (หรือคืนไปบางส่วน) ถอนการจองเพื่อให้ลองใหม่ได้
		releaseClaim(fmt.Sprintf("คืนยอดไม่สำเร็จ: %v", err))
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf(
			"%v (บิล %s: ยอดอาจถูกคืนไปบางส่วน ตรวจยอดในบัตรก่อนยกเลิกซ้ำ)", err, req.BillNo))
		return
	}

	// ขั้น 3: mark บิลเป็น VOID
	// ถึงตรงนี้ยอดถูกคืนแล้ว ห้ามถอนการจองอีก ไม่งั้นยกเลิกซ้ำจะคืนยอดซ้ำ
	if _, err := services.UpdatePosTransactionStatus(req.BillNo, services.TRANSACTION_TYPE_VOID, userId); err != nil {
		fmt.Printf("[VOID] bill_no=%s: คืนยอดในบัตรสำเร็จแล้ว แต่ mark บิลเป็น VOID ไม่สำเร็จ: %v "+
			"— ห้ามยกเลิกซ้ำ ต้องแก้สถานะบิลด้วยมือ\n", req.BillNo, err)
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf(
			"failed to update transaction status: %v (บิล %s: ยอดถูกคืนเข้าบัตรแล้ว "+
				"ห้ามกดยกเลิกซ้ำ แจ้งผู้ดูแลระบบให้แก้สถานะบิล)", err, req.BillNo))
		return
	}

	// ขั้น 4: คะแนน CRM + ลบประวัติ — ระบบข้างนอก ไว้ท้ายสุดเพราะ rollback ไม่ได้
	if err := func() error {
		if existTransaction.FreePoint > 0 {
			deductPoint := -existTransaction.FreePoint
			member, err := services.UpdatePoint(deductPoint, existTransaction.MemberTel)
			if err != nil {
				return fmt.Errorf("failed to update member claim: %w", err)
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

			status, _, err := services.SyncScoreMember(scores)
			if err != nil {
				return fmt.Errorf("failed to update member crm: %w", err)
			}
			if status != http.StatusOK {
				return fmt.Errorf("failed to update member crm non-200 status: %d", status)
			}

			// delete history
			statusCode, _, err := services.DeleteHistoryByRefTransaction("pos_transaction", existTransaction.BillNo)
			if err != nil {
				return fmt.Errorf("failed to delete history: %w", err)
			}
			if statusCode != http.StatusOK {
				return fmt.Errorf("delete history returned non-200 status")
			}
		}
		return nil
	}(); err != nil {
		// จุดนี้คือสาเหตุหลักที่ทำให้เกิดการกดยกเลิกซ้ำมาตลอด: CRM timeout
		// แล้วทั้งคำขอตอบ 500 ทั้งที่ยอดในบัตรถูกคืนเรียบร้อยแล้ว
		// ตอนนี้บิลถูกจองไว้แล้ว การกดซ้ำจะโดน 409 ปฏิเสธ ไม่หักซ้ำอีก
		fmt.Printf("[VOID] bill_no=%s: ยกเลิกบิลในระบบ POS สำเร็จครบแล้ว "+
			"แต่ซิงก์คะแนนกับ CRM ไม่สำเร็จ: %v — ต้องปรับคะแนนสมาชิก %s ด้วยมือ\n",
			req.BillNo, err, existTransaction.MemberTel)
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf(
			"%v (บิล %s: ยกเลิกในระบบ POS เรียบร้อยแล้ว ยอดในบัตรถูกคืนถูกต้อง "+
				"เหลือแค่คะแนนสมาชิกที่ยังไม่ซิงก์ ห้ามกดยกเลิกซ้ำ)", err, req.BillNo))
		return
	}

	utils.Success(c, "POS void created successfully", posVoid)

}

func handleCardVoidProcess(req models.PosVoidDto,
	existTransaction *models.PosTransaction,
	existSub []models.PosSubTransactionData,
	userId int,
	cardMemberTel string,
	cardType string) error {

	cardDepositIds := make([]string, len(existSub))
	for i, sub := range existSub {
		cardDepositIds[i] = sub.CardDepositId
	}

	cardDepositWithSub, err := services.FindCardDepositByIds(cardDepositIds)
	if err != nil {
		return fmt.Errorf("failed to find card deposit by ids: %w", err)
	}

	cardDeposits, err := services.FindCardDepositByCardNo(req.CardNo)
	if err != nil {
		return fmt.Errorf("failed to find card deposit: %w", err)
	}

	idMap := map[string]bool{}
	for _, id := range cardDepositIds {
		idMap[id] = true
	}

	var cardDepositNotSub []models.CardDeposit
	for _, cd := range cardDeposits {
		if !idMap[cd.ID.String()] {
			cardDepositNotSub = append(cardDepositNotSub, cd)
		}
	}

	e_coin := existTransaction.ECoin
	e_bonus := existTransaction.EBonus

	deducts, updateCardDeposits := applyCardWithdrawLogic(req, existTransaction, existSub, cardDepositWithSub, cardDepositNotSub, e_coin, e_bonus)

	for _, wt := range deducts {
		if _, err := services.CreateCardWithdraw(wt, userId); err != nil {
			return fmt.Errorf("failed to create withdraw: %w", err)
		}
	}

	for _, cd := range updateCardDeposits {
		if _, err := services.UpdateCardDepositBalance(cd); err != nil {
			return fmt.Errorf("failed to update card deposit balance: %w", err)
		}
	}

	checkCard, err := services.CheckCardInfo(req.CardNo)
	if err != nil {
		return fmt.Errorf("failed to check card: %w", err)
	}
	if checkCard.ECoin == 0 && checkCard.EBonus == 0 {
		clearCard := models.ClearCardDto{
			CardNo:        req.CardNo,
			MemberTel:     cardMemberTel,
			BalanceEbonus: checkCard.EBonus,
			BalanceEcoin:  checkCard.ECoin,
			CreatedBy:     req.VoidUser,
			Location:      existTransaction.BillLocation,
			TimePlay:      0,
			CardType:      cardType,
		}

		if _, err := services.ClearCard(clearCard, userId); err != nil {
			return fmt.Errorf("failed to clear card: %w", err)
		}
	}
	return nil
}

func handleNotCardVoidProcess(req models.PosVoidDto, existTransaction *models.PosTransaction, existSub []models.PosSubTransactionData, userId int) error {
	cardDepositIds := make([]string, len(existSub))
	for i, sub := range existSub {
		cardDepositIds[i] = sub.CardDepositId
	}

	cardDepositWithSub, err := services.FindCardDepositByIds(cardDepositIds)
	if err != nil {
		return fmt.Errorf("failed to find card deposit by ids: %w", err)
	}

	e_coin := existTransaction.ECoin
	e_bonus := existTransaction.EBonus

	deducts, updateCardDeposits := applyCardWithdrawNotCard(req, existTransaction, existSub, cardDepositWithSub, e_coin, e_bonus)

	for _, wt := range deducts {
		if _, err := services.CreateCardWithdraw(wt, userId); err != nil {
			return fmt.Errorf("failed to create withdraw: %w", err)
		}
	}

	for _, cd := range updateCardDeposits {
		if _, err := services.UpdateCardDepositBalance(cd); err != nil {
			return fmt.Errorf("failed to update card deposit balance: %w", err)
		}
	}

	return nil
}
func applyCardWithdrawNotCard(req models.PosVoidDto, existTransaction *models.PosTransaction, existSub []models.PosSubTransactionData, withSub []models.CardDeposit, e_coin, e_bonus int) ([]models.CardWithdrawDto, []models.CardDepositBalanceDto) {
	var deducts []models.CardWithdrawDto
	var updateCardDeposits []models.CardDepositBalanceDto

	for _, sub := range existSub {
		for _, cd := range withSub {
			if sub.CardDepositId == cd.ID.String() {
				useCoin := min(sub.ECoin, cd.BalanceCoin)
				useBonus := min(sub.EBonus, cd.BalanceBonus)

				e_coin -= useCoin
				e_bonus -= useBonus

				updateCardDeposits = append(updateCardDeposits, models.CardDepositBalanceDto{
					ID:           &cd.ID,
					BalanceCoin:  useCoin,
					BalanceBonus: useBonus,
				})

				deducts = append(deducts, models.CardWithdrawDto{
					FromChannel:   "POS",
					CardNo:        req.CardNo,
					MemberTel:     existTransaction.MemberTel,
					AmountEcoin:   useCoin,
					AmountEbonus:  useBonus,
					CardDepositId: &cd.ID,
				})
				break
			}
		}
	}

	return deducts, updateCardDeposits
}
func applyCardWithdrawLogic(req models.PosVoidDto, existTransaction *models.PosTransaction, existSub []models.PosSubTransactionData, withSub, notSub []models.CardDeposit, e_coin, e_bonus int) ([]models.CardWithdrawDto, []models.CardDepositBalanceDto) {
	var deducts []models.CardWithdrawDto
	var updateCardDeposits []models.CardDepositBalanceDto

	for _, sub := range existSub {
		for _, cd := range withSub {
			if sub.CardDepositId == cd.ID.String() {
				useCoin := min(sub.ECoin, cd.BalanceCoin)
				useBonus := min(sub.EBonus, cd.BalanceBonus)

				e_coin -= useCoin
				e_bonus -= useBonus

				updateCardDeposits = append(updateCardDeposits, models.CardDepositBalanceDto{
					ID:           &cd.ID,
					BalanceCoin:  useCoin,
					BalanceBonus: useBonus,
				})

				deducts = append(deducts, models.CardWithdrawDto{
					FromChannel:   "POS",
					CardNo:        req.CardNo,
					MemberTel:     existTransaction.MemberTel,
					AmountEcoin:   useCoin,
					AmountEbonus:  useBonus,
					CardDepositId: &cd.ID,
				})
				break
			}
		}
	}

	if e_coin > 0 || e_bonus > 0 {
		for _, cd := range notSub {
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
				FromChannel:   "POS",
				CardNo:        req.CardNo,
				MemberTel:     existTransaction.MemberTel,
				AmountEcoin:   useCoin,
				AmountEbonus:  useBonus,
				CardDepositId: &cd.ID,
			})
		}
	}

	return deducts, updateCardDeposits
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func handleCardPlayTimeVoidProcess(req models.PosVoidDto) (*models.EtimesDto, error) {

	cardPlayTimes, err := services.FindPlayTimesByCardNo(req.CardNo)
	if err != nil {
		return nil, fmt.Errorf("failed to find card deposit: %w", err)
	}

	if cardPlayTimes == nil {
		return nil, fmt.Errorf("ไม่พบรายละเอียดการใช้งานสำหรับบัตร %s", req.CardNo)
	}

	if cardPlayTimes.UsedTime > 0 {
		return nil, fmt.Errorf("บัตรหมายเลข %s ได้ถูกใช้งานแล้ว ไม่สามารถยกเลิกได้", req.CardNo)
	}

	return cardPlayTimes, nil
}

func handleCardPackageVoidProcess(req models.PosVoidDto, existSub []models.PosSubTransactionData) (int, int, int, error) {
	packageEcoin, packageEbonus, packagePlayTime := 0, 0, 0

	posMenuID := existSub[0].ProductID

	cardPackage, err := services.FindCardPackageDetailByCardNoAndPosMenu(req.CardNo, posMenuID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to find card package: %w", err)
	}

	if len(cardPackage) == 0 {
		return 0, 0, 0, fmt.Errorf("ไม่พบรายละเอียดแพ็กเกจสำหรับบัตร %s", req.CardNo)
	}

	for _, cp := range cardPackage {
		if cp.Category != "Playport" {

			if cp.EBonus != cp.RemainEBonus {
				fmt.Println("EBonus", cp.EBonus, cp.RemainEBonus)
				return 0, 0, 0, fmt.Errorf("บัตรหมายเลข %s ได้ถูกใช้งานแล้ว ไม่สามารถยกเลิกได้", req.CardNo)
			}
			if cp.ECoin != cp.RemainECoin {
				fmt.Println("ECoin", cp.ECoin, cp.RemainECoin)
				return 0, 0, 0, fmt.Errorf("บัตรหมายเลข %s ได้ถูกใช้งานแล้ว ไม่สามารถยกเลิกได้", req.CardNo)
			}

			if cp.PlayTime != cp.RemainPlayTime {
				fmt.Println("PlayTime", cp.PlayTime, cp.RemainPlayTime)
				return 0, 0, 0, fmt.Errorf("บัตรหมายเลข %s ได้ถูกใช้งานแล้ว ไม่สามารถยกเลิกได้", req.CardNo)
			}
		}

		packageEcoin += cp.ECoin
		packageEbonus += cp.EBonus
		packagePlayTime += cp.PlayTime
	}

	return packageEcoin, packageEbonus, packagePlayTime, nil
}
