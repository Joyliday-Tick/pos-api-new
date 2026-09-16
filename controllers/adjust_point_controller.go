package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

// AdjustPoint godoc
// @Summary AdjustPoint
// @Tags AdjustPoint
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.AdjustPointDto true "Adjust Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/adjust-point [post]
func CreateAdjustPoint(c *gin.Context) {
	var req models.AdjustPointDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
		return
	}

	fmt.Printf("Received Adjust Point Request: %+v\n", req)

	findMember, err := services.FindExistMemberTel(strings.TrimSpace(req.MemberTel))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findMember == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Member tel: %s not found.", req.MemberTel))
		return
	}

	if req.AdjJoylicoin < 0 {
		deductAmount := -req.AdjJoylicoin // แปลงเป็นค่าบวก = จำนวนที่จะหัก
		if findMember.Bonus < deductAmount {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Joylicoin ไม่เพียงพอ ในระบบมี %v", findMember.Bonus))
			return
		}
	}
	if req.AdjPoint < 0 {
		deductAmount := -req.AdjPoint // แปลงเป็นค่าบวก = จำนวนที่จะหัก
		if findMember.TotalPoint < deductAmount {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Point ไม่เพียงพอ ในระบบมี %v", findMember.TotalPoint))
			return
		}
	}

	if req.AdjFinwow < 0 {
		deductAmount := -req.AdjFinwow // แปลงเป็นค่าบวก = จำนวนที่จะหัก
		if findMember.Finwow < deductAmount {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Finwow ไม่เพียงพอ ในระบบมี %v", findMember.Finwow))
			return
		}
	}

	if req.AdjEstamp < 0 {
		deductAmount := -req.AdjEstamp // แปลงเป็นค่าบวก = จำนวนที่จะหัก
		if findMember.Ecoin < deductAmount {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Estamp ไม่เพียงพอ ในระบบมี %v", findMember.Ecoin))
			return
		}
	}

	if req.AdjMskill1 < 0 {
		deductAmount := -req.AdjMskill1 // แปลงเป็นค่าบวก = จำนวนที่จะหัก
		if findMember.MSkill1 < deductAmount {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Power (mskill1) ไม่เพียงพอ ในระบบมี %v", findMember.MSkill1))
			return
		}
	}

	if req.AdjMskill2 < 0 {
		deductAmount := -req.AdjMskill2 // แปลงเป็นค่าบวก = จำนวนที่จะหัก
		if findMember.MSkill2 < deductAmount {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Agility (mskill2) ไม่เพียงพอ ในระบบมี %v", findMember.MSkill2))
			return
		}
	}
	if req.AdjMskill3 < 0 {
		deductAmount := -req.AdjMskill3 // แปลงเป็นค่าบวก = จำนวนที่จะหัก
		if findMember.MSkill3 < deductAmount {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Reaction Time (mskill3) ไม่เพียงพอ ในระบบมี %v", findMember.MSkill3))
			return
		}
	}
	if req.AdjMskill4 < 0 {
		deductAmount := -req.AdjMskill4 // แปลงเป็นค่าบวก = จำนวนที่จะหัก
		if findMember.MSkill4 < deductAmount {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Balance (mskill4) ไม่เพียงพอ ในระบบมี %v", findMember.MSkill4))
			return
		}
	}

	if req.AdjMskill5 < 0 {
		deductAmount := -req.AdjMskill5 // แปลงเป็นค่าบวก = จำนวนที่จะหัก
		if findMember.MSkill5 < deductAmount {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Speed (mskill5) ไม่เพียงพอ ในระบบมี %v", findMember.MSkill5))
			return
		}
	}

	if req.AdjJubuJibi < 0 {
		deductAmount := -req.AdjJubuJibi // แปลงเป็นค่าบวก = จำนวนที่จะหัก
		if findMember.JubuJibi < deductAmount {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("JubuJibi ไม่เพียงพอ ในระบบมี %v", findMember.JubuJibi))
			return
		}
	}

	req.AdjDate = *utils.TimeNowAsia()
	adjustPoint, err := services.CreateAdjPoint(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Errorf("failed to create adjust point: %w", err).Error())
		return
	}

	g := new(errgroup.Group)

	// เก็บค่าที่จะ return จาก goroutines
	var syncResult string

	// update member
	g.Go(func() error {
		member, err := services.UpdateAdjustPoint(req)
		if err != nil {
			return fmt.Errorf("failed to update member claim: %w", err)
		}
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
		statusCode, _, err := services.SyncScoreMember(scores)
		if err != nil {
			return fmt.Errorf("failed to update member crm: %w", err)
		}
		if statusCode != http.StatusOK {
			return fmt.Errorf("failed to update member crm non-200 status: %d", statusCode)
		}

		return nil
	})

	// sync history with CRM
	g.Go(func() error {
		res, err := services.AdjustPointSyncHistory(req, adjustPoint.ID.String())
		if err != nil {
			return fmt.Errorf("failed to sync history to CRM: %w", err)
		}
		syncResult = res
		return nil
	})

	// รอให้ทั้งหมดเสร็จ
	if err := g.Wait(); err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Printf("Sync History Result: %s", syncResult)

	utils.Success(c, "Adjust point created successfully", adjustPoint)
}
