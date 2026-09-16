package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"
	"os"
	"strconv"
	"strings"
	"sync"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// @Summary Register a new card
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CardRegisterDto true "Card Entity Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/register [post]
func RegisterCardEntity(c *gin.Context) {
	var req models.CardRegisterDto

	// Validate Input
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.CardNo == "" {
		utils.Error(c, http.StatusBadRequest, "CardNo is required")
		return
	}
	// Validate MemberTel
	if req.MemberTel == "" {
		utils.Error(c, http.StatusBadRequest, "MemberTel is required")
		return
	}
	if len(req.MemberTel) != 10 || !utils.IsAllDigits(req.MemberTel) {
		utils.Error(c, http.StatusBadRequest, "MemberTel must be 10 digits")
		return
	}

	if req.CardTypeId == "" {
		utils.Error(c, http.StatusBadRequest, "CardTypeId is required")
		return
	}

	// Get userId from token claims
	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	// Check if card is locked
	lockCard, err := services.FindLockCard(req.CardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find locked card: %v", err))
		return
	}

	if lockCard != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is locked.", req.CardNo))
		return
	}

	// Check if card is already registered
	registeredCard, err := services.FindActiveCardNo(req.CardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find active card: %v", err))
		return
	}

	if registeredCard != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is already registered.", req.CardNo))
		return
	}

	// Get card type
	cardType, err := services.FindCardTypeById(req.CardTypeId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find card type: %v", err))
		return
	}

	// Create Card Entity
	cardEntity, err := services.CreateCardEntity(req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to register card: %v", err))
		return
	}

	bonusExpire := utils.GetBonusExpire()

	// Create Card Deposit
	deposit := models.CardDepositDto{
		FromChannel:     req.FromChannel,
		CardNo:          cardEntity.CardNo,
		MemberTel:       cardEntity.MemberTel,
		Amount:          0, // no amount at registration
		Coin:            req.ECoin,
		Bonus:           req.EBonus,
		BalanceCoin:     req.ECoin,
		BalanceBonus:    req.EBonus,
		PosId:           "", // Assuming no POS ID at registration
		PosMenuId:       0,
		BonusExpireDate: &bonusExpire,
		CardExpireDate:  nil,
	}

	createDeposit, err := services.CreateCardDeposit(deposit, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create card deposit: %v", err))
		return
	}

	if strings.ToLower(cardType.Name) == "jubujibi" {

		// 8. Create Card Play
		cardPlayDto := models.NewCardPlayDto{
			CardNo:        cardEntity.CardNo,
			PlayBranch:    true,
			PlayMachine:   false,
			CardDepositId: &createDeposit.ID, // Assuming no card deposit ID at registration
		}

		cardPlay, err := services.CreateCardPlay(cardPlayDto, userId)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create card play: %v", err))
			return
		}
		// fmt.Println("cardPlay", cardPlay)

		// Create Card Play Branches
		branchJubujibi := os.Getenv("JUBU_JIBI_BRANCH") // "MBP-J,THK,FIS-J"
		branchList := strings.Split(branchJubujibi, ",")
		// branchs := []string{"FAM", "ICS"}
		var playBranchs []models.CardPlayBranch

		for _, branch := range branchList {

			playBranch := models.CardPlayBranch{
				CardPlayId: cardPlay.ID,
				BranchCode: branch,
				CreateDate: *utils.TimeNowAsia(),
			}

			playBranchs = append(playBranchs, playBranch)
		}

		if len(playBranchs) > 0 {
			_, err := services.CreateBatchCardPlayBranch(playBranchs, userId)
			if err != nil {
				utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create card play branch: %v", err))
				return
			}
		}
	} else {
		if req.EBonus > 0 || req.ECoin > 0 {
			cardPlayDto := models.NewCardPlayDto{
				CardNo:        cardEntity.CardNo,
				PlayBranch:    false,
				PlayMachine:   false,
				CardDepositId: &createDeposit.ID,
			}

			_, err := services.CreateCardPlay(cardPlayDto, userId)
			if err != nil {
				utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create card play: %v", err))
				return
			}
		}
	}

	// Return success
	utils.Success(c, "Card register successfully", cardEntity)
}

// CheckCardByCardNo godoc
// @Summary Get card info by card number
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Param   cardNo  path  string  true  "cardNo"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/check/{cardNo} [get]
func CheckCardByCardNo(c *gin.Context) {
	cardNo := c.Param("cardNo")

	lockCard, err := services.FindLockCard(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find locked card: %v", err))
		return
	}

	if lockCard != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is locked.", cardNo))
		return
	}

	registeredCard, err := services.FindActiveCardNo(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find active card: %v", err))
		return
	}

	if registeredCard == nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is not register.", cardNo))
		return
	}

	checkCard, err := services.CheckCardInfo(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card: %v", err))
		return
	}

	utils.Success(c, "Check card successfully", checkCard)
}

// @Summary Topup card via POS
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.PosTopupDto true "Topup Card Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/topup-pos [post]
func TopupCardPOS(c *gin.Context) {

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	var req models.PosTopupDto

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
		// utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if card is locked
	lockCard, err := services.FindLockCard(req.CardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find locked card: %v", err))
		return
	}
	if lockCard != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is locked.", req.CardNo))
		return
	}
	// Check if card is already registered
	registeredCard, err := services.FindActiveCardNo(req.CardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find active card: %v", err))
		return
	}
	if registeredCard == nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is not register.", req.CardNo))
		return

	} else {
		// TODO : update member tel
		cardEntityId, _ := services.UpdateMemberTelToCardEntity(registeredCard.ID, registeredCard.MemberTel, req.CardNo, req.MemberTel)
		// insert Log update mobile
		if cardEntityId != nil {
			services.CreateLogUpdateCardEntity(models.LogUpdateCardEntityDto{
				CardNo:      req.CardNo,
				OldMobile:   registeredCard.MemberTel,
				NewMobile:   req.MemberTel,
				Location:    req.BillLocation,
				UpdatedDate: *utils.TimeNowAsia(),
			})
		}

		// TODO : UPDATE CARD TYPE
		if req.CardTypeId != "" && req.CardTypeId != registeredCard.CardTypeId.String() {
			cardTypeId := uuid.MustParse(req.CardTypeId)
			_, err := services.UpdateCardTypeToCardEntity(registeredCard.ID, cardTypeId)
			if err != nil {
				utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update card type: %v", err))
				return
			}
		}

		//rank products
		if len(req.Products) == 0 {
			utils.Error(c, http.StatusBadRequest, "Products are required for topup")
			return
		}

		// get Bill No
		BillNo := ""
		if req.BillNo != nil {
			BillNo = *req.BillNo
		}
		bonusExpire := utils.GetBonusExpire()
		var (
			products        []models.CardDepositDto
			posMenuCache    = make(map[int]*models.PosMenu)
			branchListCache = make(map[int][]string)
		)
		for _, product := range req.Products {
			// fmt.Printf("Product: %+v %d\n", product.ProductId, product.Quantity)
			// get posmenu by id
			// verify if posmenu is active and not deleted
			posMenu, err := services.FindPosMenuById(product.ProductId)
			if err != nil {
				utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find pos menu: %v", err))
				return
			}
			if posMenu == nil {
				utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Product with ID %d not found", product.ProductId))
				return
			}
			if !posMenu.IsActive {
				utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Product with ID %d is not active", product.ProductId))
				return
			}
			if posMenu.IsDelete {
				utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Product with ID %d is deleted", product.ProductId))
				return
			}
			posMenuCache[posMenu.ID] = posMenu
			if len(posMenu.BranchList) > 0 {
				var branchList []string
				_ = json.Unmarshal([]byte(posMenu.BranchList), &branchList)
				branchListCache[posMenu.ID] = branchList
			}

			limitTime := posMenu.LimitTime
			if limitTime == nil {
				limitTime = new(int)
				*limitTime = 0 // Default to 0 if LimitTime is nil
			}
			// Calculate balance coin by POS Menu
			balanceCoin := 0
			ogCoin := 0
			var ogPrice float64
			ogQty := 0
			var amtPrice float64
			if posMenu.ID == 1 {
				balanceCoin = product.Quantity * 1
				ogCoin = product.Quantity
				ogPrice = float64(product.Quantity)
				ogQty = 1
				amtPrice = float64(product.Quantity)
			} else {
				balanceCoin = posMenu.ECoin * product.Quantity
				ogCoin = posMenu.ECoin
				ogPrice = float64(posMenu.Price)
				ogQty = product.Quantity
				amtPrice = float64(posMenu.Price * product.Quantity)
			}

			products = append(products, models.CardDepositDto{
				FromChannel:  req.FromChannel,
				CardNo:       req.CardNo,
				MemberTel:    req.MemberTel,
				Amount:       amtPrice,
				Coin:         balanceCoin,
				Bonus:        posMenu.EBonus * product.Quantity,
				BalanceCoin:  balanceCoin,
				BalanceBonus: posMenu.EBonus * product.Quantity,
				TotalTimes:   *limitTime * product.Quantity, // Assuming total times is the limit time multiplied by quantity
				PosId:        req.PosId,
				PosMenuId:    posMenu.ID,
				BonusExpireDate: func() *time.Time {
					if posMenu.BonusExpireLimitDays != nil && *posMenu.BonusExpireLimitDays > 0 {
						convertBonusExpire := utils.AddDaysBangkok(*posMenu.BonusExpireLimitDays)
						return convertBonusExpire
					} else {
						if posMenu.BonusExpireDate != nil {
							utc := posMenu.BonusExpireDate.UTC()
							return &utc
						}
					}

					return &bonusExpire
				}(),
				BillNo:         &BillNo,
				OriginPrice:    &ogPrice,
				OriginQuantity: &ogQty,
				OriginCoin:     &ogCoin,
				OriginBonus:    &posMenu.EBonus,
				CardExpireDate: func() *time.Time {
					if posMenu.CardExpireLimitMinutes != nil && *posMenu.CardExpireLimitMinutes > 0 {
						convertCardExpire := utils.AddMinutesBangkok(*posMenu.CardExpireLimitMinutes)
						return convertCardExpire
					} else {
						if posMenu.CardExpireDate != nil {
							utc := posMenu.CardExpireDate.UTC()
							return &utc
						}
					}

					return nil
				}(),
				//TODO : discount cash
				DiscountCash:        float32(posMenu.DiscountCash),
				BalanceDiscountCash: float32(posMenu.DiscountCash),
				DiscountCashExpire: func() *time.Time {
					if posMenu.DiscountCashLimitDays != nil && *posMenu.DiscountCashLimitDays > 0 {
						convertDiscountCashExpire := utils.AddDaysBangkok(*posMenu.DiscountCashLimitDays)
						return convertDiscountCashExpire
					} else {
						if posMenu.DiscountCashExpireDate != nil {
							utc := posMenu.DiscountCashExpireDate.UTC()
							return &utc
						}
					}
					return nil
				}(),
			})
		}
		var AmountPrice float64
		var AmountCoin int
		var AmountBonus int
		var PosTransactionSub []models.PosSubTransactionDto
		transactionNo := ""
		billStatus := services.TRANSACTION_TYPE_NORMAL
		BonusStatus := "Y"
		if BillNo != "" {
			transactionNo = BillNo
		} else {
			transactionNo = services.GenerateBillNo(req.PosId)
		}
		var aggregateMu sync.Mutex
		var firstErr error
		var errOnce sync.Once
		setError := func(err error) {
			errOnce.Do(func() {
				firstErr = err
			})
		}
		var wg sync.WaitGroup
		sem := make(chan struct{}, 5)
		for _, product := range products {
			item := product
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				menuDetail, ok := posMenuCache[item.PosMenuId]
				if !ok {
					setError(fmt.Errorf("pos menu not cached: %d", item.PosMenuId))
					return
				}
				branchList := branchListCache[item.PosMenuId]
				deposit, err := services.CreateCardDeposit(models.CardDepositDto{
					FromChannel:         item.FromChannel,
					CardNo:              item.CardNo,
					MemberTel:           item.MemberTel,
					Amount:              item.Amount,
					Coin:                item.Coin,
					Bonus:               item.Bonus,
					BalanceCoin:         item.BalanceCoin,
					BalanceBonus:        item.BalanceBonus,
					TotalTimes:          item.TotalTimes,
					PosId:               item.PosId,
					PosMenuId:           item.PosMenuId,
					BonusExpireDate:     item.BonusExpireDate,
					BillNo:              &transactionNo,
					CardExpireDate:      item.CardExpireDate,
					DiscountCash:        item.DiscountCash,
					BalanceDiscountCash: item.BalanceDiscountCash,
					DiscountCashExpire:  item.DiscountCashExpire,
				}, userId)
				if err != nil {
					setError(fmt.Errorf("failed to create card deposit: %w", err))
					return
				}
				playBranch := len(menuDetail.BranchList) > 0
				//TODO : Play Machine
				playMachine := false
				if machineGroupId := menuDetail.MachineGroupID; machineGroupId != nil && *machineGroupId > 0 {
					playMachine = true
				}
				depositID := deposit.ID
				cardPlay, err := services.CreateCardPlay(models.NewCardPlayDto{
					CardNo:        item.CardNo,
					PlayBranch:    playBranch,
					PlayMachine:   playMachine,
					CardDepositId: &depositID,
				}, userId)
				if err != nil {
					setError(fmt.Errorf("failed to create card play: %w", err))
					return
				}
				cardPlayId := cardPlay.ID
				if playBranch && len(branchList) > 0 {
					var playBranchs []models.CardPlayBranch
					for _, branch := range branchList {
						playBranch := models.CardPlayBranch{
							CardPlayId: cardPlayId,
							BranchCode: branch,
						}
						playBranchs = append(playBranchs, playBranch)
					}
					if _, err = services.CreateBatchCardPlayBranch(playBranchs, userId); err != nil {
						setError(fmt.Errorf("failed to create card play branch: %w", err))
						return
					}
				}

				//TODO: Package condition
				//TODO : Play Machine
				if playMachine && menuDetail.MachineGroupID != nil && *menuDetail.MachineGroupID > 0 {
					// get machin id by machine group id
					// machineGroup, err := services.FindMachineGroupById(*menuDetail.MachineGroupID)
					machineSubGroups, err := services.FindMachineSubGroupByGroupId(*menuDetail.MachineGroupID)
					if err != nil {
						setError(fmt.Errorf("failed to find machine group: %w", err))
						return
					}
					if len(machineSubGroups) > 0 {
						//TODO: create card play machine
						for _, machineSubGroup := range machineSubGroups {
							if machineSubGroup.MachineID == nil {
								continue
							}

							var pt, ec, eb int32
							if machineSubGroup.PlayTime != nil && *machineSubGroup.PlayTime > 0 {
								pt = int32(*machineSubGroup.PlayTime)
							} else {
								pt = -1
							}
							if machineSubGroup.ECoin != nil {
								ec = int32(*machineSubGroup.ECoin)
							} else {
								ec = 0
							}
							if machineSubGroup.EBonus != nil {
								eb = int32(*machineSubGroup.EBonus)
							} else {
								eb = 0
							}

							if _, err := services.CreateCardPlayMachine(cardPlayId, int32(*machineSubGroup.MachineID), int32(userId), &pt, &ec, &eb); err != nil {
								setError(fmt.Errorf("failed to create card play machine: %w", err))
								return
							}
						}
					}
				}

				//TODO : add card play type
				//TODO : Package condition
				if _, err = CreateCardPlayType(*menuDetail, userId, cardPlayId, *item.OriginQuantity, item.CardExpireDate); err != nil {
					setError(fmt.Errorf("failed to create card play type: %w", err))
					return
				}
				aggregateMu.Lock()
				AmountPrice += item.Amount
				AmountCoin += item.Coin
				AmountBonus += item.Bonus
				PosTransactionSub = append(PosTransactionSub, models.PosSubTransactionDto{
					BillNo:        transactionNo,
					ProductID:     item.PosMenuId,
					Qty:           *item.OriginQuantity,
					Price:         int(*item.OriginPrice),
					ECoin:         *item.OriginCoin,
					EBonus:        *item.OriginBonus,
					CardDepositId: depositID.String(),
				})
				aggregateMu.Unlock()
			}()
		}
		wg.Wait()
		if firstErr != nil {
			utils.Error(c, http.StatusInternalServerError, firstErr.Error())
			return
		}

		// TODO : add POSTransaction
		billDate := *utils.TimeNowAsia()
		billPaymentId := uuid.MustParse(req.BillPaymenytId)
		transaction := models.PosTransactionDto{
			BillNo:          transactionNo,
			PosID:           req.PosId,
			BillDate:        billDate,
			BillLocation:    req.BillLocation,
			Cashier:         req.Cashier,
			CardNo:          req.CardNo,
			MemberTel:       req.MemberTel,
			ProductPrice:    int(AmountPrice),
			ECoin:           AmountCoin,
			FreePoint:       *req.FreePoint,
			EBonus:          AmountBonus,
			BillPaymentId:   billPaymentId,
			PosType:         req.PosType,
			BillStatus:      billStatus,
			BonusStatus:     BonusStatus,
			BankDetail:      *req.BankDetail,
			SubTransactions: PosTransactionSub,
		}

		_, err = services.CreatePosTransaction(transaction, userId)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create pos transaction: %v", err))
			return
		}

		// TODO : add Free point
		if req.FreePoint != nil && *req.FreePoint > 0 && req.MemberTel != "0000000000" {

			// update CRM history
			status, customer, err := services.GetCustomerByMobileNo(strings.TrimSpace(req.MemberTel))
			if err != nil || status != 200 || customer == nil {
				utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get customer by mobile no: %v", err))
				return
			}
			status, scoreTypes, err := services.GetScoreType()
			if err != nil || status != 200 || len(scoreTypes) == 0 {
				utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get score type: %v", err))
				return
			}
			scoreTypeMap := make(map[string]int)
			for _, st := range scoreTypes {
				scoreTypeMap[st.Name] = st.ID
			}
			pointID, ok := scoreTypeMap["Point"]
			if !ok {
				utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get point score type: %v", err))
				return
			}

			status, branch, err := services.GetBranchByCode(strings.TrimSpace(req.BillLocation))
			if err != nil || status != 200 || branch == nil {
				utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get branch by code: %v", err))
				return
			}

			description := fmt.Sprintf("NEW-POS >> %s", transactionNo)
			// TODO : add CRM history
			historys := models.ScoreHistory{
				CustomerID:       customer.ID,
				ScoreTypeID:      pointID,
				BranchID:         branch.ID,
				Amount:           int(*req.FreePoint),
				TransactionDate:  billDate,
				CreateBy:         req.Cashier,
				MobileNo:         req.MemberTel,
				RefTransaction:   services.StringPtr("pos_transaction"),
				RefTransactionID: &transactionNo,
				PosID:            services.StringPtr(req.PosId),
				Description:      &description,
			}

			services.SyncHistory(historys)
			// update POS member point
			member, err := services.UpdatePoint(*req.FreePoint, req.MemberTel)
			if err != nil {
				utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update point: %v", err))
				return
			}
			// update CRM member point
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
			statusCode, _, err := services.SyncScoreMember(scores)
			if err != nil {
				utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update member crm: %v", err))
				return
			}
			if statusCode != http.StatusOK {
				utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update member crm non-200 status: %d", statusCode))
				return
			}

		}

		var retStruct struct {
			BillNo string `json:"bill_no"`
		}
		retStruct.BillNo = transactionNo
		utils.Success(c, "Topup successfully", retStruct)
		return
	}
}

// GetCardEntityByMembertel godoc
// @Summary Get card by member tel
// @Description Retrieve card entity by member telephone number and is active status
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Param   memberTel  path  string  true  "memberTel"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/{memberTel} [get]
func GetCardEntityByMembertel(c *gin.Context) {
	memberTel := c.Param("memberTel")

	findMemberTel, err := services.FindCardActiveByMemberTel(memberTel)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve card entity: %v", err))
		return
	}

	utils.Success(c, "Card entity retrieved successfully", findMemberTel)
}

// GetCardActiveByTelAndCardNo godoc
// @Summary Get card by member tel and card no
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Param   memberTel  path  string  true  "memberTel"
// @Param   cardNo  path  string  true  "cardNo"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/{memberTel}/{cardNo} [get]
func GetCardActiveByTelAndCardNo(c *gin.Context) {
	memberTel := c.Param("memberTel")
	cardNo := c.Param("cardNo")

	findActive, err := services.FindActiveMemberTelAndCardNo(memberTel, cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve card entity: %v", err))
		return
	}

	utils.Success(c, "Card entity retrieved successfully", findActive)
}

func CreateCardPlayType(posMenu models.PosMenu, userId int, cardPlayId uuid.UUID, qty int, cardExpire *time.Time) (*models.CardPlayType, error) {
	if posMenu.StartDate == nil && posMenu.EndDate == nil && (posMenu.LimitTime == nil || *posMenu.LimitTime == 0) {
		return nil, nil
	}
	var startDate *time.Time
	var endDate *time.Time
	var playTime *int

	if cardExpire != nil {
		startDate = utils.TimeNowAsia()
	} else {
		startDate = nil
	}
	if cardExpire != nil {
		endDate = cardExpire
	} else {
		endDate = nil
	}
	if posMenu.LimitTime != nil {
		lt := *posMenu.LimitTime
		result := lt * qty
		playTime = &result
	} else {
		defaultPT := 0
		playTime = &defaultPT
	}
	cardPlayTypeData := models.CardPlayTypeDto{
		CardPlayId: cardPlayId,
		StartDate:  startDate,
		EndDate:    endDate,
		PlayTime:   playTime,
	}
	cardPlayType, err := services.CreateCardPlayType(cardPlayTypeData, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to create card play type: %w", err)
	}

	return &cardPlayType, nil

}

// CheckCardByTel godoc
// @Summary Get card list by tel
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Param   tel  path  string  true  "tel"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/tel/{tel} [get]
func CheckCardByTel(c *gin.Context) {
	tel := c.Param("tel")
	if len(tel) != 10 || !utils.IsAllDigits(tel) {
		utils.Error(c, http.StatusBadRequest, "MemberTel must be 10 digits")
		return
	}

	findCards, err := services.CheckCardByTel(tel)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find card list: %v", err))
		return
	}

	utils.Success(c, "Check card by tel successfully", findCards)
}

// DeleteCardEntityPermanent godoc
// @Summary Delete card entity permanently
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Param   cardNo  path  string  true  "cardNo"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/{cardNo} [delete]
func DeleteCardEntityPermanent(c *gin.Context) {
	cardNo := c.Param("cardNo")
	error := services.PermanentlyDeleteCard(cardNo)
	if error != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete card entity permanently: %v", error))
		return
	}
	utils.Success(c, "Delete card entity permanently successfully", "")
}

// ClearCardEntity godoc
// @Summary Clear card entity by card number
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.ClearCardDto true "Clear card Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/clear-card [post]
func ClearCard(c *gin.Context) {
	var req models.ClearCardDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	registeredCard, err := services.FindActiveCardNo(req.CardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find active card: %v", err))
		return
	}

	if registeredCard == nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is not register.", req.CardNo))
		return
	}

	clearCard, err := services.ClearCard(req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create clear card: %v", err))
		return
	}

	utils.Success(c, "Clear card successfully", clearCard)
}

// LockCardEntity godoc
// @Summary Lock card entity
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.LockCardDto true "lock card Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/lock-card [post]
func LockCard(c *gin.Context) {
	var req models.LockCardDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	findCard, err := services.FindCardEntityById(req.RefCardEntityId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find card: %v", err))
		return
	}

	if findCard == nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s and id: %s not found.", req.CardNo, req.RefCardEntityId))
		return
	}

	var (
		lockCardResult    interface{}
		createLockCardRes interface{}
		lockCardErr       error
		createLockCardErr error
		wg                sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		lockCardResult, lockCardErr = services.LockCardEntity(req.RefCardEntityId, false, userId)
	}()

	go func() {
		defer wg.Done()
		req.LockDate = *utils.TimeNowAsia()
		createLockCardRes, createLockCardErr = services.CreateLogLockCard(req)
	}()

	wg.Wait()

	if lockCardErr != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to lock card: %v", lockCardErr))
		return
	}
	if createLockCardErr != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create lock card: %v", createLockCardErr))
		return
	}

	utils.Success(c, "Locked card successfully", gin.H{
		"lockCard":      lockCardResult,
		"createLogCard": createLockCardRes,
	})
}

// CheckCardDetailByCardNo godoc
// @Summary Get card detail by card number
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Param   cardNo  path  string  true  "cardNo"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/check/info/{cardNo} [get]
func CheckCardDetailByCardNo(c *gin.Context) {
	cardNo := c.Param("cardNo")

	lockCard, err := services.FindLockCard(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find locked card: %v", err))
		return
	}

	if lockCard != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is locked.", cardNo))
		return
	}

	registeredCard, err := services.FindActiveCardNo(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find active card: %v", err))
		return
	}

	if registeredCard == nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is not register.", cardNo))
		return
	}

	checkCardType, err := services.FindActiveCardNoWithCardType(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card type: %v", err))
		return
	}

	if checkCardType == nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is not register.", cardNo))
		return
	}

	var checkCard *models.CardDetailDto

	if checkCardType.CardTypeName == "Package" {
		checkCard, err = services.CheckCardPackageInfoWithExpire(cardNo)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card: %v", err))
			return
		}

	} else {

		checkCard, err = services.CheckCardInfoWithExpire(cardNo)
		// checkCard, err := services.CheckCardInfoV2(cardNo)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card: %v", err))
			return
		}
	}

	//*check Package & Time play*//
	if checkCard.CardType == "Package" || checkCard.CardType == "Time play" {

		fmt.Printf(
			"CHECK CLEAR CARD: cardNo=%s type=%s ecoin=%v ebonus=%v etimes=%v\n",
			cardNo,
			checkCard.CardType,
			checkCard.ECoin,
			checkCard.EBonus,
			checkCard.ETimes,
		)

		if checkCard.ECoin == 0 &&
			checkCard.EBonus == 0 &&
			checkCard.ETimes == 0 {

			fmt.Printf("DELETE CARD PLAY: %s\n", cardNo)

			err = services.DeleteCardPlayByCardNo(cardNo)
			if err != nil {
				utils.Error(
					c,
					http.StatusInternalServerError,
					fmt.Sprintf("Failed to clear card play: %v", err),
				)
				return
			}
		}
	}

	utils.Success(c, "Check card successfully", checkCard)

}

// UnLockCardEntity godoc
// @Summary Unlock card entity
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.UnLockCardDto true "unlock card Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/unlock-card [post]
func UnLockCard(c *gin.Context) {
	var req models.UnLockCardDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	findCard, err := services.FindCardEntityById(req.RefCardEntityId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find card: %v", err))
		return
	}

	if findCard == nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Id: %s not found.", req.RefCardEntityId))
		return
	}

	var (
		unlockCardResult  interface{}
		updateLockCardRes interface{}
		unlockCardErr     error
		updateLockCardErr error
		wg                sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		unlockCardResult, unlockCardErr = services.LockCardEntity(req.RefCardEntityId, true, userId)
	}()

	go func() {
		defer wg.Done()
		req.UnlockDate = utils.TimeNowAsia()
		updateLockCardRes, updateLockCardErr = services.UnLockCardEntity(req)
	}()

	wg.Wait()

	if unlockCardErr != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to unlock card: %v", unlockCardErr))
		return
	}
	if updateLockCardErr != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update lock card: %v", updateLockCardErr))
		return
	}

	utils.Success(c, "Unlocked card successfully", gin.H{
		"lockCard":      unlockCardResult,
		"createLogCard": updateLockCardRes,
	})
}

// CheckCardDetailByTel godoc
// @Summary Get card detail list by tel
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Param tel query string false "Search by tel"
// @Param cardNo query string false "Search by card no"
// @Param isLock query bool false "Search lock status (true/false)"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/member [get]
func CheckCardMemberByTel(c *gin.Context) {
	tel := c.Query("tel")
	cardNo := c.Query("cardNo")
	isLockedStr := c.Query("isLock")
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	skip, err := strconv.Atoi(c.DefaultQuery("skip", "10"))
	if err != nil || skip < 1 {
		skip = 10
	}

	var isLocked *bool
	if isLockedStr != "" {
		locked, err := strconv.ParseBool(isLockedStr)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Invalid isLocked value: %v", err))
			return
		}
		isLocked = &locked
	}
	params := models.SearchLockCardParams{
		MemberTel: tel,
		CardNo:    cardNo,
		IsLocked:  isLocked,
		Page:      page,
		Skip:      skip,
	}
	findCards, err := services.CheckCardMemberList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find card list: %v", err))
		return
	}

	utils.Success(c, "Check card member by tel successfully", findCards)
}

// CheckCardRefundByTel godoc
// @Summary Get card list refund by tel
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Param   tel  path  string  true  "tel"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/refund/{tel} [get]
func CheckCardRefundByTel(c *gin.Context) {
	tel := c.Param("tel")
	if len(tel) != 10 || !utils.IsAllDigits(tel) {
		utils.Error(c, http.StatusBadRequest, "MemberTel must be 10 digits")
		return
	}

	findCards, err := services.CheckCardRefundByTel(tel)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find card list: %v", err))
		return
	}

	utils.Success(c, "Check card by tel successfully", findCards)
}

// CheckCardPackageByCardNo godoc
// @Summary Get card package detail by card number
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Param   cardNo  path  string  true  "cardNo"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/check/package/{cardNo} [get]
func CheckCardPackageByCardNo(c *gin.Context) {
	cardNo := c.Param("cardNo")

	lockCard, err := services.FindLockCard(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find locked card: %v", err))
		return
	}

	if lockCard != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is locked.", cardNo))
		return
	}

	registeredCard, err := services.FindActiveCardNo(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find active card: %v", err))
		return
	}

	if registeredCard == nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is not register.", cardNo))
		return
	}

	cardType, err := services.FindCardTypeById(registeredCard.CardTypeId.String())
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get card type: %v", err))
		return
	}

	if cardType != nil && cardType.Name != "Package" {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is not a package card.", cardNo))
		return
	}

	checkCard, err := services.CheckCardPackageInfo(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card: %v", err))
		return
	}

	utils.Success(c, "Check card successfully", checkCard)
}

// CheckCardTypeByCardNo godoc
// @Summary Get card type by card number
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Param   cardNo  path  string  true  "cardNo"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/check/card-type/{cardNo} [get]
func CheckCardTypeByCardNo(c *gin.Context) {
	cardNo := c.Param("cardNo")

	lockCard, err := services.FindLockCard(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find locked card: %v", err))
		return
	}

	if lockCard != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is locked.", cardNo))
		return
	}

	registeredCard, err := services.FindActiveCardNo(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find active card: %v", err))
		return
	}

	if registeredCard == nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is not register.", cardNo))
		return
	}

	checkCardType, err := services.FindActiveCardNoWithCardType(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card type: %v", err))
		return
	}

	utils.Success(c, "Check card successfully", checkCardType)
}

// CheckCardDetailByCardNo godoc
// @Summary Get card package info by card number
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Param   cardNo  path  string  true  "cardNo"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/check/package-info/{cardNo} [get]
func CheckCardPackageInfoByCardNo(c *gin.Context) {
	cardNo := c.Param("cardNo")

	lockCard, err := services.FindLockCard(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find locked card: %v", err))
		return
	}

	if lockCard != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is locked.", cardNo))
		return
	}

	registeredCard, err := services.FindActiveCardNo(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find active card: %v", err))
		return
	}

	if registeredCard == nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is not registered.", cardNo))
		return
	}

	checkCard, err := services.CheckCardPackageInfoWithExpire(cardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card: %v", err))
		return
	}

	utils.Success(c, "Check card successfully", checkCard)
}

// TransferCard godoc
// @Summary Transfer card
// @Tags Card
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.LogTransferCardDto true "transfer card Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card/transfer [post]
func TransferCard(c *gin.Context) {
	var req models.LogTransferCardDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
		return
	}

	if req.OldCardNo == nil || req.TransferCardNo == nil || req.MemberTel == nil {
		utils.Error(c, http.StatusBadRequest, "missing required fields")
		return
	}

	// Get userId from token claims
	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	checkCardType, err := services.FindActiveCardNoWithCardType(*req.OldCardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card type: %v", err))
		return
	}

	checkCardType1, err := services.FindActiveCardNoWithCardType(*req.TransferCardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card type: %v", err))
		return
	}

	if checkCardType.CardTypeName != "Normal" || checkCardType1.CardTypeName != "Normal" {
		utils.Error(c, http.StatusBadRequest, "สามารถโอนย้ายบัตรได้เฉพาะบัตรประเภท Normal เท่านั้น")
		return
	}

	checkOldCard, err := services.CheckCardInfoWithExpire(*req.OldCardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card no %s: %v", *req.OldCardNo, err))
		return
	}

	if checkOldCard.CardExpDays == "expired" {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("ไม่สามารถโอนย้ายบัตรได้ เนื่องจากบัตรหมายเลข %s หมดอายุแล้ว", *req.OldCardNo))
		return
	}

	checkNewCard, err := services.CheckCardInfoWithExpire(*req.TransferCardNo)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to check card no %s: %v", *req.TransferCardNo, err))
		return
	}

	if checkNewCard.CardExpDays == "expired" {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("ไม่สามารถโอนย้ายบัตรได้ เนื่องจากบัตรหมายเลข %s หมดอายุแล้ว", *req.TransferCardNo))
		return
	}

	// Create Card Deposit
	deposit := models.CardDepositDto{
		FromChannel:     "Transfer Card",
		CardNo:          *req.TransferCardNo,
		MemberTel:       *req.MemberTel,
		Amount:          0,
		Coin:            checkOldCard.ECoin,
		Bonus:           checkOldCard.EBonus,
		BalanceCoin:     checkOldCard.ECoin,
		BalanceBonus:    checkOldCard.EBonus,
		PosId:           "",
		PosMenuId:       0,
		BonusExpireDate: checkOldCard.EBonusExpDate,
		CardExpireDate:  nil,
	}

	_, err = services.CreateCardDeposit(deposit, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create card deposit: %v", err))
		return
	}

	// clear old card
	ClearCardDto := models.ClearCardDto{
		CardNo:        *req.OldCardNo,
		MemberTel:     *req.MemberTel,
		BalanceEcoin:  checkOldCard.ECoin,
		BalanceEbonus: checkOldCard.EBonus,
		CreatedBy:     *req.CreatedBy,
		Location:      *req.Location,
		TimePlay:      checkOldCard.ETimes,
		CardType:      checkCardType.CardTypeName,
	}

	clearCard, err := services.ClearCard(ClearCardDto, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create clear card: %v", err))
		return
	}

	req.BalanceEbonus = checkOldCard.EBonus
	req.BalanceEcoin = checkOldCard.ECoin

	logTransferCard, err := services.CreateLogTransferCard(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create log transfer card: %v", err))
		return
	}

	utils.Success(c, "Transfer card successfully", gin.H{
		"card_deposit":      deposit,
		"clear_card":        clearCard,
		"log_transfer_card": logTransferCard,
	})
}
