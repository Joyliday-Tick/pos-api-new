package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// JubujibiDeposit godoc
// @Summary JubuJibi Deposit
// @Tags JubuJibi
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.DepositJubuJibi true "Jubu jibi deposit Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/jubu-jibi/deposit [post]
func DepositJubuJibi(c *gin.Context) {
	var req models.DepositJubuJibi

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
		return
	}
	if req.MemberTel == "" || req.Location == "" || len(req.SubItems) == 0 {
		utils.Error(c, http.StatusBadRequest, "member_tel, location and sub_jubu_jibi are required")
		return
	}

	req.Type = "jubu_jibi"

	depositHead, err := services.CreateDepositTransaction(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to create deposit: %w", err).Error())
		return
	}

	//*update member*//

	member, err := services.UpdateDepositJubuJibi(req.Deposit, req.Joylicoin, req.TopupDeduct, req.MemberTel)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("failed to update member deposit jubu jibi: %v", err))
		return
	}

	//*update meter record*//
	if err = services.UpdateMeterRecordJubuJibi(req.MemberTel); err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to update meter record jubu jibi: "+err.Error())
		return
	}

	//*update prize*//
	if err = services.UpdatePrizeCounterJubuJibi(req.MemberTel); err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to update prize counter jubu jibi: "+err.Error())
		return
	}

	//update member crm//
	fmt.Printf("Updated Member: %+v\n", member)
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
		// return fmt.Errorf("failed to update member crm: %w", err)
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to update member crm: %w", err).Error())
		return
	}
	if status != http.StatusOK {
		// return fmt.Errorf("failed to update member crm non-200 status: %d", statusCode)
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to update member crm non-200 status: %d", status).Error())
		return
	}

	utils.Success(c, "Deposit created successfully", depositHead)
}

// JubujibiRedeem godoc
// @Summary JubuJibi Redeem
// @Tags JubuJibi
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.RedeemJubuJibi true "Jubu jibi redeem Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/jubu-jibi/redeem [post]
func RedeemJubuJibi(c *gin.Context) {
	var req models.RedeemJubuJibi

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
		return
	}
	if req.MemberTel == "" || req.Location == "" || len(req.SubItems) == 0 {
		utils.Error(c, http.StatusBadRequest, "member_tel, location and sub_jubu_jibi are required")
		return
	}

	req.Type = "jubu_jibi"

	redeemHead, err := services.CreateRedeemTransaction(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to create deposit: %w", err).Error())
		return
	}

	//*update member*//

	member, err := services.UpdateRedeemJubuJibi(req.RedeemPrice, req.MemberTel)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("failed to update member redeem jubu jibi: %v", err))
		return
	}

	//update member crm//
	fmt.Printf("Updated Member: %+v\n", member)
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
		// return fmt.Errorf("failed to update member crm: %w", err)
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to update member crm: %w", err).Error())
		return
	}
	if status != http.StatusOK {
		// return fmt.Errorf("failed to update member crm non-200 status: %d", statusCode)
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to update member crm non-200 status: %d", status).Error())
		return
	}

	utils.Success(c, "Redeem created successfully", redeemHead)
}
