package v2

import (
	"fmt"
	"net/http"
	"new-pos-api/models"
	model_v2 "new-pos-api/models/v2"
	"new-pos-api/services"
	service_v2 "new-pos-api/services/v2"
	"new-pos-api/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// var bkkLoc, _ = time.LoadLocation("Asia/Bangkok")
// var nowInBKK = time.Now().In(bkkLoc)

// VerifyCard godoc
// @Summary Verify card information V2
// @Tags V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body model_v2.VerifyCardRequest true "Verify Card Request"
// @Success 200 {object} model_v2.VerifyCardResponse
// @Failure 401 {object} utils.StandardErrorResponse
// @Router /api/v2/verify-card [post]
func VerifyCard(c *gin.Context) {
	// Enforce Bearer Authentication Only
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		utils.Error(c, http.StatusUnauthorized, "Bearer authentication required")
		return
	}

	// Extract claims (already verified by AuthMiddleware, but we ensure it was via Bearer)
	_, exists := c.Get("userClaims")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User claims not found")
		return
	}
	// fmt.Println("V2 VerifyCard accessed by Bearer token:", claims)

	var req model_v2.VerifyCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// start process
	/*
		1. check card
		2. check card play
		3. verify card play (TODO later)
		4. Deduct card play by sequence (TODO later)
	*/

	//TODO: check card type
	cardNo := req.CardNo
	// Consolidated Check: Status (Lock/Active) + Data Retrieval (Optimized V2)
	checkCard, status, err := services.VerifyCardFullV2(cardNo, req.GamePrice, req.DeductCondition)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to verify card: %v", err))
		return
	}

	switch status {
	case "not_found":
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is not register.", cardNo))
		return
	case "locked":
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is locked.", cardNo))
		return
	case "not_enough":
		utils.Error(c, http.StatusBadRequest, "not enough ecoin or ebonus")
		return
	}

	fmt.Printf("checkCard: %v\n", checkCard)

	// Step 2: Check card play (using pre-fetched card details)
	// Map V2 request to CardPlayVerifyDto
	playVerifyDto := models.CardPlayVerifyDto{
		CardNo:      cardNo,
		PlayBranch:  req.BranchId,                // asset_id maps to branch/machine location
		PlayMachine: strconv.Itoa(req.MachineId), // asset_id maps to branch/machine location
		UseCoin:     &req.GamePrice,
		UseBonus:    nil,
	}
	// DeductCondition
	/*
		- ecoin_first
		- ebonus_first
		- ecoin_only
		- ebonus_only
	*/
	// Set deduction amounts based on condition
	// TODO: verify package
	cardPlayID, balanceEcoin, balanceEbonus, balanceETimes, err, deductMachine := services.VerifyCardPlayV2(cardNo, checkCard, playVerifyDto, req.DeductCondition)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Card play verification failed: %v", err))
		return
	}
	// fmt.Printf("cardPlayID: %v\n", cardPlayID)
	// fmt.Printf("Deduct machine : %v\n", deductMachine)

	// Calculate amounts to deduct based on VerifyV2 result
	// Note: VerifyCardPlayV2 is a preview. We must calculate the difference to perform the actual deduction.
	deductEcoin := 0
	if balanceEcoin != nil {
		deductEcoin = checkCard.ECoin - *balanceEcoin
	}

	deductEbonus := 0
	if balanceEbonus != nil {
		deductEbonus = checkCard.EBonus - *balanceEbonus
	}

	// Map results for response and meter record
	ecoin := checkCard.ECoin
	ebonus := checkCard.EBonus
	etimes := checkCard.ETimes

	if balanceEcoin != nil {
		ecoin = *balanceEcoin
	}
	if balanceEbonus != nil {
		ebonus = *balanceEbonus
	}
	if balanceETimes != nil {
		etimes = *balanceETimes
	}

	// Prepare Deduction Input
	// Use 0 as systemUserId if actual user ID is not extracted from claims yet
	systemUserId := 0
	// if claims != nil {
	// 	// attempt simple type assertion if your claims structure supports it, otherwise 0
	// 	// systemUserId = ...
	// }

	channelName := "POS_V2"
	deductInput := models.CardPlayDeductDto{
		CardPlayId:  cardPlayID,
		CardNo:      cardNo,
		ECoin:       deductEcoin,
		EBonus:      deductEbonus,
		FromChannel: &channelName,
	}

	// Prepare Meter Record DTO
	// loc, _ := time.LoadLocation("Asia/Bangkok")
	// now := time.Now().In(loc)
	now := utils.TimeNowAsia()
	deductEtime := 0
	if strings.ToLower(checkCard.CardType) == "time play" {
		deductEtime = 1
	}
	assetID, convErr := strconv.Atoi(req.AssetId)
	if convErr != nil {
		assetID = 0 // default to 0 if conversion fails, but ideally should handle this error properly
	}
	meterDto := model_v2.MeterRecordDto{
		RcDate:          now,
		RcAssetID:       assetID,
		RcAssetHead:     req.HeadId,
		RcAssetLocation: req.BranchId,
		RcMember:        checkCard.MemberTel,
		RcCardID:        checkCard.CardNo,
		RcEcoin:         deductEcoin,
		RcBonus:         deductEbonus,
		RnEcoin:         ecoin,  // remaining ecoin after deduction
		RnBonus:         ebonus, // remaining ebonus after deduction
		CardType:        checkCard.CardType,
		RcTime:          deductEtime,
		RnTime:          etimes,
		OldID:           0,
	}

	//TODO: Overide return card type package
	if deductMachine != nil {
		beforeEtime := 1
		afterEtime := max(*deductMachine.AfterETimes, 0)
		meterDto.RcEcoin = *deductMachine.BeforeEcoin - *deductMachine.AfterEcoin
		meterDto.RcBonus = *deductMachine.BeforeEbonus - *deductMachine.AfterEbonus
		meterDto.RnEcoin = *deductMachine.AfterEcoin
		meterDto.RnBonus = *deductMachine.AfterEbonus
		meterDto.RcTime = beforeEtime
		meterDto.RnTime = afterEtime
		if err := service_v2.CreateMeterRecord(meterDto); err != nil {
			fmt.Printf("ERROR: Failed to create meter record for Card %s: %v\n", meterDto.RcCardID, err)
		}
		responseData := model_v2.VerifyCardResponseData{
			CardType:      checkCard.CardType,
			CardNo:        checkCard.CardNo,
			CardId:        cardPlayID.String(),
			MemberTel:     checkCard.MemberTel,
			BalanceEcoin:  *deductMachine.AfterEcoin,
			BalanceEbonus: *deductMachine.AfterEbonus,
			BalanceEtimes: afterEtime,
			TimeRemaining: checkCard.CardExpDate,
			HeadId:        req.HeadId,
		}

		response := model_v2.VerifyCardResponse{
			Status:  "success",
			Message: "Card verified successfully",
			Data:    responseData,
		}

		c.JSON(http.StatusOK, response)

		return
	}

	// Asynchronous Deduction Step (Performance Optimized)
	// This returns response immediately and handles deduction in background
	go func(input models.CardPlayDeductDto, userId int, mDto model_v2.MeterRecordDto) {
		_, err := services.DeductCardPlay(input, userId)
		if err != nil {
			// In production, log this critical error to a monitoring system (e.g. Sentry, ELK)
			// because the client thinks the transaction succeeded.
			fmt.Printf("CRITICAL: Async Deduction failed for Card %s: %v\n", input.CardNo, err)
		} else {
			// Insert meter record
			if err := service_v2.CreateMeterRecord(mDto); err != nil {
				fmt.Printf("ERROR: Failed to create meter record for Card %s: %v\n", mDto.RcCardID, err)
			}
			fmt.Printf("Async Deduction success for Card %s\n", input.CardNo)
		}
	}(deductInput, systemUserId, meterDto)

	responseData := model_v2.VerifyCardResponseData{
		CardType:      checkCard.CardType,
		CardNo:        checkCard.CardNo,
		CardId:        cardPlayID.String(),
		MemberTel:     checkCard.MemberTel,
		BalanceEcoin:  ecoin,
		BalanceEbonus: ebonus,
		BalanceEtimes: etimes,
		TimeRemaining: checkCard.CardExpDate,
		HeadId:        req.HeadId,
	}

	response := model_v2.VerifyCardResponse{
		Status:  "success",
		Message: "Card verified successfully",
		Data:    responseData,
	}

	c.JSON(http.StatusOK, response)
}

// VerifyCardSMC godoc
// @Summary Verify card information and SMC
// @Tags V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body model_v2.VerifyCardRequest true "Verify Card Request"
// @Success 200 {object} model_v2.VerifyCardSmcResponse
// @Failure 401 {object} utils.StandardErrorResponse
// @Router /api/v2/verify-card-smc [post]
func VerifyCardSMC(c *gin.Context) {
	// Enforce Bearer Authentication Only
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		utils.Error(c, http.StatusUnauthorized, "Bearer authentication required")
		return
	}

	// Extract claims (already verified by AuthMiddleware, but we ensure it was via Bearer)
	_, exists := c.Get("userClaims")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User claims not found")
		return
	}
	// fmt.Println("V2 VerifyCard accessed by Bearer token:", claims)

	var req model_v2.VerifyCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// start process
	/*
		1. check card
		2. check card play
		3. verify card play (TODO later)
		4. Deduct card play by sequence (TODO later)
	*/

	//TODO: check card type
	cardNo := req.CardNo
	// Consolidated Check: Status (Lock/Active) + Data Retrieval (Optimized V2)
	checkCard, status, err := services.VerifyCardFullV2(cardNo, req.GamePrice, req.DeductCondition)
	// verify member only
	if checkCard.MemberTel == "0000000000" || checkCard.MemberTel == "" {
		utils.Error(c, http.StatusBadRequest, "Failed to verify card: member card only")
		return
	}

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to verify card: %v", err))
		return
	}

	switch status {
	case "not_found":
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is not register.", cardNo))
		return
	case "locked":
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("CardNo: %s is locked.", cardNo))
		return
	case "not_enough":
		utils.Error(c, http.StatusBadRequest, "not enough ecoin or ebonus")
		return
	}

	// fmt.Printf("checkCard: %v\n", checkCard)

	// Step 2: Check card play (using pre-fetched card details)
	// Map V2 request to CardPlayVerifyDto
	playVerifyDto := models.CardPlayVerifyDto{
		CardNo:      cardNo,
		PlayBranch:  req.BranchId,                // asset_id maps to branch/machine location
		PlayMachine: strconv.Itoa(req.MachineId), // asset_id maps to branch/machine location
		UseCoin:     &req.GamePrice,
		UseBonus:    nil,
	}
	// DeductCondition
	/*
		- ecoin_first
		- ebonus_first
		- ecoin_only
		- ebonus_only
	*/
	// Set deduction amounts based on condition

	cardPlayID, balanceEcoin, balanceEbonus, balanceETimes, err, deductMachine := services.VerifyCardPlayV2(cardNo, checkCard, playVerifyDto, req.DeductCondition)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Card play verification failed: %v", err))
		return
	}
	fmt.Printf("cardPlayID: %v\n", cardPlayID)
	fmt.Printf("deductMachine: %v\n", deductMachine)

	// fmt.Printf("SMC Prize Response: %+v\n", *smc)

	// Calculate amounts to deduct based on VerifyV2 result
	// Note: VerifyCardPlayV2 is a preview. We must calculate the difference to perform the actual deduction.
	deductEcoin := 0
	if balanceEcoin != nil {
		deductEcoin = checkCard.ECoin - *balanceEcoin
	}

	deductEbonus := 0
	if balanceEbonus != nil {
		deductEbonus = checkCard.EBonus - *balanceEbonus
	}

	// Map results for response and meter record
	ecoin := checkCard.ECoin
	ebonus := checkCard.EBonus
	etimes := checkCard.ETimes

	if balanceEcoin != nil {
		ecoin = *balanceEcoin
	}
	if balanceEbonus != nil {
		ebonus = *balanceEbonus
	}
	if balanceETimes != nil {
		etimes = *balanceETimes
	}

	// Step 3: SMC Prize Sync (IoT)
	discountValue := 0
	if req.DiscountCash != nil {
		discountValue = *req.DiscountCash
	}

	reqSMC := model_v2.SMCPrizeDto{
		AssetId:      req.AssetId,
		DiscountCash: discountValue,
		Ebonus:       deductEbonus,
		ECoin:        deductEcoin,
		ItemId:       req.ItemId,
		MemberTel:    checkCard.MemberTel,
		CardNo:       cardNo,
	}

	//TODO: Package for SMC override request
	if deductMachine != nil {
		useEbonus := *deductMachine.BeforeEbonus - *deductMachine.AfterEbonus
		useEcoin := *deductMachine.BeforeEcoin - *deductMachine.AfterEcoin
		fmt.Printf("beforeEbonus=%d, beforeEcoin=%d\n", *deductMachine.BeforeEbonus, *deductMachine.BeforeEcoin)
		fmt.Printf("Overriding SMC request with actual deduction amounts: useEbonus=%d, useEcoin=%d\n", useEbonus, useEcoin)
		reqSMC.Ebonus = useEbonus
		reqSMC.ECoin = useEcoin
	}

	// fmt.Printf("request %+v \n",reqSMC)
	smc, err := service_v2.SendPrizeSMC(reqSMC)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("SMC prize request failed: %v", err))
		return
	}
	// Prepare Deduction Input
	// Use 0 as systemUserId if actual user ID is not extracted from claims yet
	systemUserId := 0
	// if claims != nil {
	// 	// attempt simple type assertion if your claims structure supports it, otherwise 0
	// 	// systemUserId = ...
	// }

	channelName := "POS_V2_SMC"
	deductInput := models.CardPlayDeductDto{
		CardPlayId:  cardPlayID,
		CardNo:      cardNo,
		ECoin:       deductEcoin,
		EBonus:      deductEbonus,
		FromChannel: &channelName,
	}

	// Prepare Meter Record DTO
	// loc, _ := time.LoadLocation("Asia/Bangkok")
	// now := time.Now().In(loc)
	deductEtime := 0
	if strings.ToLower(checkCard.CardType) == "time play" {
		deductEtime = 1
	}
	assetID, convErr := strconv.Atoi(req.AssetId)
	if convErr != nil {
		assetID = 0 // default to 0 if conversion fails, but ideally should handle this error properly
	}
	now := utils.TimeNowAsia()
	meterDto := model_v2.MeterRecordDto{
		RcDate:          now,
		RcAssetID:       assetID,
		RcAssetHead:     req.HeadId,
		RcAssetLocation: req.BranchId,
		RcMember:        checkCard.MemberTel,
		RcCardID:        checkCard.CardNo,
		RcEcoin:         deductEcoin,
		RcBonus:         deductEbonus,
		RnEcoin:         ecoin,  // remaining ecoin after deduction
		RnBonus:         ebonus, // remaining ebonus after deduction
		CardType:        checkCard.CardType,
		RcTime:          deductEtime,
		RnTime:          etimes,
		OldID:           0,
	}

	fmt.Printf("Prepared MeterRecordDto: %+v\n", meterDto)

	//TODO: Overide return card type package
	if deductMachine != nil {
		fmt.Printf("if deductMachine: %+v\n", deductMachine)
		beforeEtime := 1
		afterEtime := max(*deductMachine.AfterETimes, 0)
		meterDto.RcEcoin = *deductMachine.BeforeEcoin - *deductMachine.AfterEcoin
		meterDto.RcBonus = *deductMachine.BeforeEbonus - *deductMachine.AfterEbonus
		meterDto.RnEcoin = *deductMachine.AfterEcoin
		meterDto.RnBonus = *deductMachine.AfterEbonus
		meterDto.RcTime = beforeEtime
		meterDto.RnTime = afterEtime
		if err := service_v2.CreateMeterRecord(meterDto); err != nil {
			fmt.Printf("ERROR: Failed to create meter record for Card %s: %v\n", meterDto.RcCardID, err)
		}
		responseData := model_v2.VerifyCardResponseData{
			CardType:      checkCard.CardType,
			CardNo:        checkCard.CardNo,
			CardId:        cardPlayID.String(),
			MemberTel:     checkCard.MemberTel,
			BalanceEcoin:  *deductMachine.AfterEcoin,
			BalanceEbonus: *deductMachine.AfterEbonus,
			BalanceEtimes: afterEtime,
			TimeRemaining: checkCard.CardExpDate,
			HeadId:        req.HeadId,
		}

		smcData := model_v2.SMCPrizeData{}
		if smc != nil {
			smcData = smc.Data
		}
		response := model_v2.VerifyCardSmcResponse{
			Status:  "success",
			Message: "Card SMC verified successfully",
			Data:    responseData,
			SmcData: smcData,
		}

		c.JSON(http.StatusOK, response)

		return
	}

	// Asynchronous Deduction Step (Performance Optimized)
	// This returns response immediately and handles deduction in background
	go func(input models.CardPlayDeductDto, userId int, mDto model_v2.MeterRecordDto) {
		result, err := services.DeductCardPlay(input, userId)
		fmt.Printf("result deductCardPlay: %+v\n", result)
		fmt.Printf("err deductCardPlay: %+v\n", err)
		if err != nil {
			// In production, log this critical error to a monitoring system (e.g. Sentry, ELK)
			// because the client thinks the transaction succeeded.
			fmt.Printf("CRITICAL: Async Deduction failed for Card %s: %v\n", input.CardNo, err)
		} else {
			fmt.Printf("insert meter record for Card %s\n", input.CardNo)
			// Insert meter record
			if err := service_v2.CreateMeterRecord(mDto); err != nil {
				fmt.Printf("ERROR: Failed to create meter record for Card %s: %v\n", mDto.RcCardID, err)
			}
			fmt.Printf("Async Deduction success for Card %s\n", input.CardNo)
		}
	}(deductInput, systemUserId, meterDto)

	responseData := model_v2.VerifyCardResponseData{
		CardType:      checkCard.CardType,
		CardNo:        checkCard.CardNo,
		CardId:        cardPlayID.String(),
		MemberTel:     checkCard.MemberTel,
		BalanceEcoin:  ecoin,
		BalanceEbonus: ebonus,
		BalanceEtimes: etimes,
		TimeRemaining: checkCard.CardExpDate,
		HeadId:        req.HeadId,
	}

	smcData := model_v2.SMCPrizeData{}
	if smc != nil {
		smcData = smc.Data
	}
	response := model_v2.VerifyCardSmcResponse{
		Status:  "success",
		Message: "Card SMC verified successfully",
		Data:    responseData,
		SmcData: smcData,
	}

	c.JSON(http.StatusOK, response)
}
