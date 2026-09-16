package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create a new card play
// @Tags Card Play
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CardPlayDto true "Card Play Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card-play [post]
func CreateCardPlay(c *gin.Context) {
	var req models.NewCardPlayDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	cardPlay, err := services.CreateCardPlay(req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create card play: %v", err))
		return
	}

	fmt.Println("cardPlay", cardPlay)

	branchs := []string{"FAM", "ICS"}
	var playBranchs []models.CardPlayBranch

	for _, branch := range branchs {
		fmt.Println("Branch loop:", branch)

		playBranch := models.CardPlayBranch{
			CardPlayId: cardPlay.ID,
			BranchCode: branch,
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

	utils.Success(c, "Card play created successfully", cardPlay)
}

// @Summary Verify a card play
// @Tags Card Play
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CardPlayVerifyDto true "Card Play Verify Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 400 {object} utils.StandardErrorResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card-play-verify [post]
func VerifyCardPlay(c *gin.Context) {
	var req models.CardPlayVerifyDto
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	// Use optimized service function to fetch all data in parallel
	CardPlayID, cardInfo, cardPlayType, err := services.VerifyCardPlayOptimized(req, userId)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to verify card play: %v", err))
		return
	}
	// success stuct
	//  cardInfo:= models.CardDetailDto{}
	type Success struct {
		IsPlayed      bool       `json:"is_played"`
		CardNo        string     `json:"card_no"`
		CardPlayId    *uuid.UUID `json:"card_play_id"`
		CardType      *string    `json:"card_type"`
		ShowBalance   string     `json:"show_balance"`
		MemberTel     *string    `json:"member_tel"`
		Ecoin         *int       `json:"e_coin"`
		Ebonus        *int       `json:"e_bonus"`
		EbonusExp     *int       `json:"e_bonus_exp"`
		EbonusExpDate *time.Time `json:"e_bonus_exp_date"`
		EbonusExpDays *string    `json:"e_bonus_exp_days"`
		Etimes        *int       `json:"e_times"`
		BalanceEtimes *int       `json:"balance_etimes"`
		CardExpDate   *time.Time `json:"card_exp_date"`
		CardExpDays   *string    `json:"card_exp_days"`
	}
	// Calculate balance times
	//TODO : check card play type nil
	var balanceTime int = 0
	var playTime int = 0
	if cardPlayType != nil {
		balanceTime = *cardPlayType.PlayTime - *cardPlayType.UsedTime
		playTime = *cardPlayType.PlayTime
	}

	var ebonusExp int = 0
	bonusExpireTotalDate, _ := services.GetCardDepositBonusExpireTotalDate(req.CardNo)
	if bonusExpireTotalDate != nil {
		ebonusExp = bonusExpireTotalDate.BonusExpireTotal
	}

	SuccessResp := Success{
		IsPlayed:      CardPlayID != nil,
		CardPlayId:    CardPlayID,
		CardNo:        req.CardNo,
		CardType:      utils.StringToPtr(cardInfo.CardType),
		ShowBalance:   cardInfo.ShowBalance,
		MemberTel:     utils.StringToPtr(cardInfo.MemberTel),
		Ecoin:         utils.IntToPtr(cardInfo.ECoin),
		Ebonus:        utils.IntToPtr(cardInfo.EBonus),
		EbonusExp:     utils.IntToPtr(ebonusExp),
		EbonusExpDate: cardInfo.EBonusExpDate,
		EbonusExpDays: utils.StringToPtr(cardInfo.EBonusExpDays),
		Etimes:        utils.IntToPtr(playTime),
		BalanceEtimes: utils.IntToPtr(balanceTime),
		CardExpDate:   cardInfo.CardExpDate,
		CardExpDays:   utils.StringToPtr(cardInfo.CardExpDays),
	}

	utils.Success(c, "Card play verified successfully", SuccessResp)
}

// @Summary Deduct card play balance
// @Tags Card Play
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CardPlayDeductDto true "Card Play Deduct Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 400 {object} utils.StandardErrorResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/card-play-deduct [post]
func DeductCardPlay(c *gin.Context) {
	var req models.CardPlayDeductDto
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	//TODO : func deduct
	withdraws, err := services.DeductCardPlay(req, userId)
	fmt.Println("withdraws:", withdraws)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to deduct card play: %v", err))
		return
	}

	// success stuct
	type Success struct {
		CardNo    string     `json:"card_no"`
		DepositID *uuid.UUID `json:"deposit_id"`
	}
	DepositIDs := withdraws[0].Id // Example UUID, replace with actual logic
	SuccessResp := Success{
		CardNo:    req.CardNo,
		DepositID: DepositIDs,
	}

	utils.Success(c, "Card play balance deducted successfully", SuccessResp)
}
