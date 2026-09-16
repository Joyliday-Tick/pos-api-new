package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// // ClaimJoylicoin godoc
// // @Summary Claim Joylicoin
// // @Tags Claim
// // @Security BasicAuth
// // @Security BearerAuth
// // @Accept json
// // @Produce json
// // @Param body body models.ReturnBonusDto true "Claim Data"
// // @Success 200 {object} utils.StandardSuccessResponse
// // @Failure 500 {object} utils.StandardErrorResponse
// // @Router /api/claim [post]
// func ClaimJoylicoin(c *gin.Context) {
// 	var req models.ReturnBonusDto

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		utils.Error(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
// 		return
// 	}

// 	if req.BonusReturn == nil || req.PointStamp == nil {
// 		utils.Error(c, http.StatusBadRequest, "bonus_return or point_stamp is missing")
// 		return
// 	}

// 	returnBonus, err := services.CreateReturnBonus(req)
// 	if err != nil {
// 		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to create return bonus: %w", err).Error())
// 		return
// 	}

// 	g := new(errgroup.Group)

// 	// เก็บค่าที่จะ return จาก goroutines
// 	var syncResult string

// 	// update member
// 	g.Go(func() error {
// 		member, err := services.UpdateClaimJoylicoin(int(*req.BonusReturn), req.MemberTel, int(*req.PointStamp))
// 		if err != nil {
// 			return fmt.Errorf("failed to update member claim: %w", err)
// 		}
// 		fmt.Printf("Updated Member: %+v\n", member)
// 		scores := models.ScoreMember{
// 			MobileNo:   member.Tel,
// 			Bonus:      member.Bonus,
// 			TotalPoint: member.TotalPoint,
// 			ECoin:      member.Ecoin,
// 			JubuJibi:   member.JubuJibi,
// 			FinWow:     member.Finwow,
// 			EStamp:     member.Estamp,
// 			MSkill1:    member.MSkill1,
// 			MSkill2:    member.MSkill2,
// 			MSkill3:    member.MSkill3,
// 			MSkill4:    member.MSkill4,
// 			MSkill5:    member.MSkill5,
// 		}
// 		statusCode, _, err := services.SyncScoreMember(scores)
// 		if err != nil {
// 			return fmt.Errorf("failed to update member crm: %w", err)
// 		}
// 		if statusCode != http.StatusOK {
// 			return fmt.Errorf("failed to update member crm non-200 status: %d", statusCode)
// 		}

// 		return nil
// 	})

// 	// update meter record
// 	g.Go(func() error {
// 		if err := services.UpdateMeterRecordClaim(req.MemberTel); err != nil {
// 			return fmt.Errorf("failed to update meter record claim: %w", err)
// 		}
// 		return nil
// 	})

// 	// update stamp point
// 	g.Go(func() error {
// 		if err := services.UpdateStampPointClaim(req.MemberTel); err != nil {
// 			return fmt.Errorf("failed to update stamp point claim: %w", err)
// 		}
// 		return nil
// 	})

// 	// sync history with CRM
// 	g.Go(func() error {
// 		req.BonusDate = returnBonus.BonusDate
// 		res, err := services.ReturnBonusSyncHistory(req, returnBonus.ID.String())
// 		if err != nil {
// 			return fmt.Errorf("failed to sync history to CRM: %w", err)
// 		}
// 		syncResult = res
// 		return nil
// 	})

// 	// รอให้ทั้งหมดเสร็จ
// 	if err := g.Wait(); err != nil {
// 		utils.Error(c, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	fmt.Printf("Sync History Result: %s", syncResult)

// 	utils.Success(c, "Return bonus created successfully", returnBonus)
// }

// ClaimJoylicoin godoc
// @Summary Claim Joylicoin
// @Tags Claim
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.ReturnBonusDto true "Claim Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/claim [post]
func NewClaimJoylicoin(c *gin.Context) {
	var req models.ReturnBonusDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
		return
	}

	if req.BonusReturn == nil || req.PointStamp == nil {
		utils.Error(c, http.StatusBadRequest, "bonus_return or point_stamp is missing")
		return
	}

	// ====== TRANSACTION PART ======
	result, member, err := services.ClaimJoylicoinTx(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to claim joylicoin: %w", err).Error())
		return
	}

	// ====== WAIT FOR SCORE SYNC ======
	if err := services.SyncScoreAfterClaim(member); err != nil {
		utils.Error(c, http.StatusInternalServerError,
			fmt.Sprintf("sync score failed: %v", err))
		return
	}

	// ====== FIRE & FORGET HISTORY ======
	go services.SyncHistoryAfterClaim(req, result)

	utils.Success(c, "Return bonus created successfully", result)
}
