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
	voidMode := voidModeNormal
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

		// ---------------------------------------------------------------------
		// ตัดสินว่าจะยกเลิกแบบแพ็กเกจ/เวลาเล่น หรือแบบคืนยอดเหรียญปกติ
		// โดยดู "เมนูที่บิลนี้ขาย" ไม่ใช่ "ชนิดบัตร"
		//
		// เดิมใช้ cardType ตรง ๆ ทั้งที่บัตร Package/Time play เติมเงินเหรียญ
		// ธรรมดาได้ บิลเติมเงิน (pos_menu STD0001 "เติมเงิน/แลกเหรียญ" ที่
		// machine_group_id = 0) บนบัตร Package จึงถูกบังคับไปทางแพ็กเกจ
		// แล้วไปหา card_play_machine ที่ผูกกับ machine group ของเมนูนั้น
		// ซึ่งไม่มี -> ตอบ 400 "ไม่พบรายละเอียดแพ็กเกจ" และยกเลิกไม่ได้เลย
		//
		// ตรวจ UAT 2026-09-18: บิลที่ยังไม่ยกเลิกและติดกับดักนี้มี 31 ใบ
		// (บัตร Package 15, Time play 16) ส่วนบิลที่ขายแพ็กเกจจริง 311 ใบ
		// ยังหา card_play_machine เจอเหมือนเดิม จึงไม่กระทบ
		//
		// คิดครั้งเดียวที่นี่แล้วใช้ทั้งตอน precheck ข้างล่างและตอนขั้นที่ 2
		// ถ้าแก้แค่ precheck ขั้นที่ 2 จะยังเข้าสาขา Package แล้วเรียก ClearCard
		// ล้างบัตรทั้งใบด้วย packageEcoin = 0 ซึ่งแย่กว่าเดิม
		// ---------------------------------------------------------------------
		soldPackage, err := billSoldPackageMenu(existSub)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		if soldPackage {
			switch cardType {
			case "Time play":
				voidMode = voidModeTimePlay
			case "Package":
				voidMode = voidModePackage
			}
		}

		if req.DeductCard {
			if voidMode == voidModeTimePlay {
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
			if voidMode == voidModePackage {
				packageEcoin, packcageEbonus, packagePlayTime, err = handleCardPackageVoidProcess(req, existSub)
				if err != nil {
					utils.Error(c, http.StatusBadRequest, err.Error())
					return
				}
			}
		} else {
			req.VoidReason = "U-" + req.VoidReason
		}
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
	// ⚠️ การจองที่สร้างในขั้น 1 จะไม่ถูกถอนคืนไม่ว่าขั้นไหนจะพัง
	//
	// รอบแรกเคยเขียนให้ถอนคืนเมื่อขั้น 2 ล้มเหลว โดยคิดว่า "ล้มเหลว = ยังไม่ได้เขียนอะไร"
	// ซึ่งผิด — handleCardVoidProcess ลง card_withdraw ครบทุกแถวก่อน (บรรทัด 350-354)
	// แล้วค่อยหักยอดทีละก้อน (356-360) ความล้มเหลวกลางทางจึงแปลว่าเขียนไปแล้วเสมอ
	// การถอนการจองคืนตอนนั้นเท่ากับเปิดทางให้กดยกเลิกซ้ำแล้วคืนยอดซ้ำ
	// ซึ่งคือบั๊กที่ 409 ตั้งใจปิดพอดี
	//
	// จึงเลือกให้ fail closed: บิลที่ยกเลิกไม่สำเร็จจะถูกล็อกไว้ด้วย 409
	// ต้องให้คนมาตรวจว่าคืนยอดไปถึงไหนแล้วก่อนตัดสินใจ ไม่ใช่ปล่อยให้ retry มั่ว
	// แลกกับกรณีที่พังตั้งแต่ยังไม่เขียนอะไร ซึ่งก็จะถูกล็อกไปด้วย — ยอมรับได้
	// เพราะ "ต้องเรียกคนมาดู" ปลอดภัยกว่า "หักเงินลูกค้าซ้ำ"
	// ---------------------------------------------------------------------

	// ---------------------------------------------------------------------
	// ตรวจว่าคืนยอดได้ครบไหม ก่อนจองบิล
	//
	// ลูกค้าเติมเงินแล้วเล่นไปหมด ยอดในบัตรจึงไม่พอให้กลับรายการ
	// ของเดิมปล่อยให้หักเท่าที่มีแล้วรายงานว่าสำเร็จ ส่วนต่างหายไปเงียบ ๆ
	// (และก่อนมี guard ของ UpdateCardDepositBalance ยอดจะติดลบไปเลย
	//  ซึ่งเป็นที่มาของแถวติดลบ 5 แถวที่เจอใน UAT)
	//
	// ตรวจ UAT 2026-09-18: บิลอายุ 8-30 วันที่คืนไม่ครบมี 145 จาก 581 ใบ (25%)
	// จึงไม่ใช่เคสหายาก ต้องมีทางออกที่ชัดเจน ไม่ใช่ปฏิเสธแล้วจบ
	//
	// ทางออกที่เจ้าของระบบเลือก: ให้ผู้มีสิทธิ์อนุมัติทับได้ โดย
	//   - หักเท่าที่มีจริง ไม่ทำให้ยอดติดลบ
	//   - บันทึกว่าใครอนุมัติและขาดเท่าไรลง void_reason
	// ---------------------------------------------------------------------
	if req.CardNo != "" && req.DeductCard && voidMode == voidModeNormal {
		_, _, shortCoin, shortBonus, planErr := buildVoidPlan(req, existTransaction, existSub)
		if planErr != nil {
			utils.Error(c, http.StatusInternalServerError, planErr.Error())
			return
		}

		if shortCoin > 0 || shortBonus > 0 {
			if !req.ForceVoid {
				utils.Error(c, http.StatusConflict, fmt.Sprintf(
					"ยอดในบัตร %s ไม่พอให้กลับรายการบิล %s (ขาด e_coin %d, e_bonus %d) "+
						"ลูกค้าใช้ยอดนี้ไปแล้ว ต้องให้ผู้จัดการอนุมัติทับจึงจะยกเลิกได้",
					req.CardNo, req.BillNo, shortCoin, shortBonus))
				return
			}

			// ธงอนุมัติทับต้องมาพร้อมสิทธิ์จริง ไม่ใช่แค่ส่ง json มา
			// route นี้ถูก RequireRole กั้นอยู่แล้ว แต่เช็คซ้ำที่นี่เพราะ
			// หน้าจอ POS เรียกผ่าน service account ที่เป็น Administrator
			// การเช็คตรงนี้จึงกันได้เฉพาะคนที่ยิง API ตรงด้วยบัญชีตัวเอง
			roleId, roleErr := middlewares.GetRoleIdFromClaims(c)
			if roleErr != nil || !canApproveForceVoid(roleId) {
				utils.Error(c, http.StatusForbidden,
					"การอนุมัติทับต้องใช้สิทธิ์ผู้จัดการขึ้นไป")
				return
			}

			req.VoidReason = annotateForceVoidReason(req.VoidReason, req.VoidUser, shortCoin, shortBonus)

			fmt.Printf("[VOID] bill_no=%s card_no=%s: อนุมัติทับโดย %s (roleId=%d) "+
				"ยอดไม่พอ ขาด coin %d bonus %d — หักเท่าที่มีจริง\n",
				req.BillNo, req.CardNo, req.VoidUser, roleId, shortCoin, shortBonus)
		}
	}

	// ขั้น 1: จองบิล
	posVoid, err := services.CreatePosVoid(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to create pos void: %w", err).Error())
		return
	}

	// ขั้น 2: คืนยอดในบัตร
	if err := func() error {
		if req.CardNo != "" {
			// สาขาเดียวกับ precheck ข้างบน — ใช้ voidMode ไม่ใช่ cardType
			// (ดูเหตุผลที่คอมเมนต์ยาวตรงที่คำนวณ voidMode)
			switch voidMode {
			case voidModeTimePlay:
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
			case voidModePackage:
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
				if req.DeductCard {
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
		// คืนยอดไปแล้วบางส่วนแน่นอน เพราะ handleCardVoidProcess ลง card_withdraw
		// ครบทุกแถวก่อนจะเริ่มหักยอด การจองถูกคงไว้ บิลนี้จะโดน 409 ถ้ามีคนกดซ้ำ
		fmt.Printf("[VOID] bill_no=%s card_no=%s: คืนยอดไม่สำเร็จกลางทาง: %v "+
			"— บิลถูกล็อกไว้ด้วยแถว pos_void (void_id=%s) แล้ว ต้องตรวจว่าคืนยอดไปถึงไหน "+
			"ก่อนตัดสินใจ ถ้าจะให้ยกเลิกใหม่ได้ต้องลบแถวนั้นด้วยมือหลังตรวจเสร็จ\n",
			req.BillNo, req.CardNo, err, posVoid.VoidID)
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf(
			"%v (บิล %s: ยอดถูกคืนไปบางส่วนแล้ว ห้ามกดยกเลิกซ้ำ แจ้งผู้ดูแลระบบให้ตรวจสอบ)",
			err, req.BillNo))
		return
	}

	// ขั้น 3: mark บิลเป็น VOID
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

// buildVoidPlan อ่านข้อมูลแล้วคำนวณว่าจะคืนยอดจาก deposit ไหนเท่าไร
//
// เป็นการอ่านล้วน ไม่เขียนอะไร จึงเรียกล่วงหน้าเพื่อดูส่วนต่างก่อนตัดสินใจได้
// shortCoin/shortBonus = ส่วนที่หาที่คืนไม่ได้ เพราะยอดในบัตรถูกใช้ไปแล้ว
func buildVoidPlan(req models.PosVoidDto,
	existTransaction *models.PosTransaction,
	existSub []models.PosSubTransactionData) ([]models.CardWithdrawDto, []models.CardDepositBalanceDto, int, int, error) {

	cardDepositIds := make([]string, len(existSub))
	for i, sub := range existSub {
		cardDepositIds[i] = sub.CardDepositId
	}

	cardDepositWithSub, err := services.FindCardDepositByIds(cardDepositIds)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("failed to find card deposit by ids: %w", err)
	}

	cardDeposits, err := services.FindCardDepositByCardNo(req.CardNo)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("failed to find card deposit: %w", err)
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

	deducts, updates, shortCoin, shortBonus := applyCardWithdrawLogic(
		req, existTransaction, existSub, cardDepositWithSub, cardDepositNotSub,
		existTransaction.ECoin, existTransaction.EBonus)

	return deducts, updates, shortCoin, shortBonus, nil
}

func handleCardVoidProcess(req models.PosVoidDto,
	existTransaction *models.PosTransaction,
	existSub []models.PosSubTransactionData,
	userId int,
	cardMemberTel string,
	cardType string) error {

	// ส่วนต่างถูกตรวจและตัดสินไปแล้วที่ controller ก่อนจองบิล
	// (ไม่พอ + ไม่มีอนุมัติทับ = ถูกปฏิเสธตั้งแต่ตอนนั้น ยังไม่เขียนอะไร)
	deducts, updateCardDeposits, _, _, err := buildVoidPlan(req, existTransaction, existSub)
	if err != nil {
		return err
	}

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

	// ---------------------------------------------------------------------
	// ปิดบัตรให้เรียบร้อยถ้ายกเลิกบิลแล้วบัตรว่างเปล่า
	//
	// ⚠️ ต้องเช็คสองชั้น เดิมเช็คแค่ CheckCardInfo ซึ่งไม่ตรงกับยอดดิบ
	//
	// เหตุการณ์จริง 2026-09-18: ยกเลิกบิล JYN-9-2603300003 (100 coin) บนบัตร
	// F443416C ที่มียอดดิบเหลือ 7630 coin / 1130 bonus แต่ CheckCardInfo
	// รายงาน 0/0 (กรองวันหมดอายุ — บัตรใบนั้นเคยถูก clear ไปเมื่อ พ.ค. 2026)
	// เงื่อนไขชั้นเดียวจึงเป็นจริง แล้ว ClearCard ปิด deposit ทั้ง 72 แถว
	// ลบ card_play 40 แถว และบันทึก clear_card ว่ายอด 0/0 ทั้งที่ของจริง 7630
	// การยกเลิกบิลใบเดียวลากไปล้างมูลค่าทั้งใบ
	//
	// FindCardDepositByCardNo กรอง is_active AND (balance_coin > 0 OR
	// balance_bonus > 0) อยู่แล้ว ถ้าคืน 0 แถวก็คือยอดดิบเป็นศูนย์จริง
	// ---------------------------------------------------------------------
	checkCard, err := services.CheckCardInfo(req.CardNo)
	if err != nil {
		return fmt.Errorf("failed to check card: %w", err)
	}
	if checkCard.ECoin == 0 && checkCard.EBonus == 0 {
		remaining, err := services.FindCardDepositByCardNo(req.CardNo)
		if err != nil {
			return fmt.Errorf("failed to verify remaining balance before clearing card: %w", err)
		}

		if len(remaining) > 0 {
			// ยอดดิบยังมีอยู่ ห้ามล้าง — ดังไว้ให้เห็นว่าทั้งสองแหล่งไม่ตรงกัน
			var rawCoin, rawBonus int
			for _, d := range remaining {
				rawCoin += d.BalanceCoin
				rawBonus += d.BalanceBonus
			}
			fmt.Printf("[VOID] bill_no=%s card_no=%s: ไม่ล้างบัตร เพราะ CheckCardInfo รายงาน 0/0 "+
				"แต่ยอดดิบยังเหลือ %d coin / %d bonus ใน %d deposit — คืนยอดของบิลนี้เรียบร้อยแล้ว\n",
				req.BillNo, req.CardNo, rawCoin, rawBonus, len(remaining))
			return nil
		}

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

// applyCardWithdrawLogic คืน (รายการ withdraw, รายการหักยอด, ส่วนต่าง coin, ส่วนต่าง bonus)
//
// สองค่าท้ายคือยอดที่หาที่คืนไม่ได้ เพราะลูกค้าใช้ไปแล้ว
// เดิมฟังก์ชันนี้กลืนส่วนต่างไว้เงียบ ๆ (min() แล้วจบ) ผู้เรียกจึงไม่รู้ว่าคืนไม่ครบ
// และการยกเลิกบิลก็รายงานว่าสำเร็จทั้งที่กลับรายการได้แค่บางส่วน
func applyCardWithdrawLogic(req models.PosVoidDto, existTransaction *models.PosTransaction, existSub []models.PosSubTransactionData, withSub, notSub []models.CardDeposit, e_coin, e_bonus int) ([]models.CardWithdrawDto, []models.CardDepositBalanceDto, int, int) {
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

	// เหลือเท่าไรคือหาที่คืนไม่ได้
	if e_coin < 0 {
		e_coin = 0
	}
	if e_bonus < 0 {
		e_bonus = 0
	}

	return deducts, updateCardDeposits, e_coin, e_bonus
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

// ความยาวจริงของ pos_void.void_reason ในฐานข้อมูล
//
// โมเดล Go เขียน gorm type varchar(100) แต่คอลัมน์จริงเป็น varchar(500)
// (ตรวจ information_schema 2026-09-18) ใช้ค่าจริงเป็นเกณฑ์ตัด
// นับเป็นตัวอักษรไม่ใช่ไบต์ เพราะ Postgres varchar นับ character
const voidReasonMaxLen = 500

// canApproveForceVoid — ใครอนุมัติทับได้ ตรงกับ RequireRole ที่กั้น /pos-void อยู่
func canApproveForceVoid(roleId int) bool {
	switch roleId {
	case middlewares.RoleAdministrator,
		middlewares.RoleRMBackoffice,
		middlewares.RoleManager,
		middlewares.RoleAssistManager:
		return true
	}
	return false
}

// annotateForceVoidReason ต่อข้อความอนุมัติทับเข้ากับเหตุผลเดิม
//
// เก็บลง void_reason ไปก่อนเพราะยังไม่เพิ่มคอลัมน์ — ฐานข้อมูลนี้ใช้ร่วมกับ
// ระบบเดิมที่ยังทำงานอยู่ การเพิ่มคอลัมน์ต้องรอตกลงกันก่อน
// ถ้ายาวเกิน จะตัดเหตุผลเดิมทิ้ง ไม่ตัดข้อความอนุมัติ เพราะส่วนนั้นคือหลักฐาน
func annotateForceVoidReason(reason, approver string, shortCoin, shortBonus int) string {
	note := fmt.Sprintf(" [อนุมัติทับโดย %s ยอดไม่พอ ขาด coin %d bonus %d]",
		approver, shortCoin, shortBonus)

	noteRunes := []rune(note)
	if len(noteRunes) >= voidReasonMaxLen {
		return string(noteRunes[:voidReasonMaxLen])
	}

	room := voidReasonMaxLen - len(noteRunes)
	reasonRunes := []rune(reason)
	if len(reasonRunes) > room {
		reasonRunes = reasonRunes[:room]
	}

	return string(reasonRunes) + note
}

// โหมดการยกเลิก — ตัดสินจากเมนูที่บิลขาย ไม่ใช่ชนิดบัตร
const (
	voidModeNormal   = "normal"
	voidModePackage  = "package"
	voidModeTimePlay = "timeplay"
)

// billSoldPackageMenu บอกว่าบิลนี้ขายเมนูที่ผูกกับ machine group (คือแพ็กเกจ) หรือไม่
//
// machine_group_id เป็น 0 หรือ NULL แปลว่าเป็นเมนูเติมเงิน/แลกเหรียญธรรมดา
// ซึ่งขายบนบัตรชนิดไหนก็ได้ รวมถึงบัตร Package และ Time play
func billSoldPackageMenu(existSub []models.PosSubTransactionData) (bool, error) {
	for _, sub := range existSub {
		menu, err := services.FindPosMenuById(sub.ProductID)
		if err != nil {
			return false, fmt.Errorf("failed to find pos menu %d: %w", sub.ProductID, err)
		}
		if menu == nil {
			continue
		}
		if menu.MachineGroupID != nil && *menu.MachineGroupID > 0 {
			return true, nil
		}
	}
	return false, nil
}

func handleCardPackageVoidProcess(req models.PosVoidDto, existSub []models.PosSubTransactionData) (int, int, int, error) {
	packageEcoin, packageEbonus, packagePlayTime := 0, 0, 0

	// FindSubTransactionByBillNo คืน nil เมื่อไม่พบแถว existSub[0] จึง panic ได้
	// ตรวจ UAT 2026-09-18: มีบิลที่ยังไม่ยกเลิกและไม่มี pos_sub_transaction เลย 3 ใบ
	// (1 ใบอยู่บนบัตร Package) ซึ่งจะ panic แทนที่จะได้ error ที่อ่านรู้เรื่อง
	if len(existSub) == 0 {
		return 0, 0, 0, fmt.Errorf("บิล %s ไม่มีรายการสินค้า จึงตรวจแพ็กเกจไม่ได้", req.BillNo)
	}

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
