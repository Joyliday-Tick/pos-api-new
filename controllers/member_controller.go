package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// @Summary Create a new member
// @Tags Member
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.MemberDto true "Member Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/member [post]
func CreateMember(c *gin.Context) {
	var req models.MemberDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Tel == "" {
		utils.Error(c, http.StatusBadRequest, "Tel is required")
		return
	}
	if len(req.Tel) != 10 || !utils.IsAllDigits(req.Tel) {
		utils.Error(c, http.StatusBadRequest, "Tel must be 10 digits")
		return
	}

	// ตรวจสอบว่าชื่อซ้ำหรือไม่
	existMember, err := services.FindExistMemberTel(req.Tel)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find exist member: %v", err))
		return
	}

	if existMember != nil && existMember.ID != 0 {
		utils.Error(c, http.StatusBadRequest, "Member with this tel already exists")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	member, err := services.CreateMember(req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to create member: %v", err))
		return
	}

	utils.Success(c, "Member created successfully", member)
}

// UpdateMember godoc
// @Summary Update member by tel
// @Tags Member
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.MemberDto true "Member Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/member [put]
func UpdateMember(c *gin.Context) {

	var req models.MemberDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Tel == "" {
		utils.Error(c, http.StatusBadRequest, "Tel is required")
		return
	}
	if len(req.Tel) != 10 || !utils.IsAllDigits(req.Tel) {
		utils.Error(c, http.StatusBadRequest, "MemberTel must be 10 digits")
		return
	}

	findMember, err := services.FindExistMemberTel(strings.TrimSpace(req.Tel))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findMember == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Member tel: %s not found.", req.Tel))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateMember, err := services.UpdateMember(findMember.ID, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update member: %v", err))
		return
	}

	utils.Success(c, "Member updated successfully", updateMember)
}

// GetMemberByTel godoc
// @Summary Get member by tel
// @Tags Member
// @Security BasicAuth
// @Security BearerAuth
// @Param   tel  path  string  true  "Tel"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/member/{tel} [get]
func GetMemberByTel(c *gin.Context) {
	tel := c.Param("tel")

	// findMember, err := services.FindExistMemberTel(tel)
	findMember, err := services.FindMemberWithLatestTierByTel(tel)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve member: %v", err))
		return
	}

	utils.Success(c, "Member retrieved successfully", findMember)
}

// UpdateMemberSkill godoc
// @Summary Update member skill by tel
// @Tags Member
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.SkillDto true "Skill Data"
// @Param   tel  path  string  true  "Tel"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/member/{tel} [post]
func UpdateSkill(c *gin.Context) {

	var req models.SkillDto

	tel := c.Param("tel")

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if tel == "" {
		utils.Error(c, http.StatusBadRequest, "Tel is required")
		return
	}
	if len(tel) != 10 || !utils.IsAllDigits(tel) {
		utils.Error(c, http.StatusBadRequest, "MemberTel must be 10 digits")
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateMember, err := services.UpdateSkill(tel, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update member skill: %v", err))
		return
	}

	utils.Success(c, "Member skill updated successfully", updateMember)
}

// DeductJoylicoin godoc
// @Summary Deduct joylicoin by tel
// @Tags Member
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.JoylicoinDto true "Joylicoin Data"
// @Param   tel  path  string  true  "Tel"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/member/deduct-joylicoin/{tel} [post]
func UpdateJoylicoin(c *gin.Context) {

	var req models.JoylicoinDto

	tel := c.Param("tel")

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if tel == "" {
		utils.Error(c, http.StatusBadRequest, "Tel is required")
		return
	}
	if len(tel) != 10 || !utils.IsAllDigits(tel) {
		utils.Error(c, http.StatusBadRequest, "MemberTel must be 10 digits")
		return
	}

	findMember, err := services.FindExistMemberTel(strings.TrimSpace(tel))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findMember == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("ไม่พบสมาชิกเบอร์: %s", tel))
		return
	}

	bonus := findMember.Bonus
	reqValue := req.Joylicoin

	// แปลงให้เป็นค่าลบเสมอ (เพื่อใช้หัก)
	deductJoylicoin := reqValue
	if reqValue > 0 {
		deductJoylicoin = reqValue * -1
	}

	// ตรวจสอบว่า Joylicoin พอให้หักไหม
	if (deductJoylicoin * -1) > bonus {
		utils.Error(c, http.StatusBadRequest,
			fmt.Sprintf("สมาชิกเบอร์: %s มี Joylicoin ไม่เพียงพอ", tel))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateMember, err := services.UpdateJoylicoin(tel, deductJoylicoin, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update member joylicoin: %v", err))
		return
	}

	utils.Success(c, "Member joylicoin updated successfully", updateMember)
}

// UpdateMemberName godoc
// @Summary Update member name by tel
// @Tags Member
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.MemberNameDto true "Member Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/member/name [post]
func UpdateMemberName(c *gin.Context) {

	var req models.MemberNameDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Tel == "" {
		utils.Error(c, http.StatusBadRequest, "Tel is required")
		return
	}
	if len(req.Tel) != 10 || !utils.IsAllDigits(req.Tel) {
		utils.Error(c, http.StatusBadRequest, "MemberTel must be 10 digits")
		return
	}

	findMember, err := services.FindExistMemberTel(strings.TrimSpace(req.Tel))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findMember == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Member tel: %s not found.", req.Tel))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	updateMember, err := services.UpdateMemberName(findMember.ID, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update member: %v", err))
		return
	}

	utils.Success(c, "Member updated successfully", updateMember)
}

// DeductPoint godoc
// @Summary Deduct point by tel
// @Tags Member
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.TotalPointDto true "Total point data"
// @Param   tel  path  string  true  "Tel"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/member/deduct-totalpoint/{tel} [post]
func UpdateTotalPoint(c *gin.Context) {

	var req models.TotalPointDto

	tel := c.Param("tel")

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if tel == "" {
		utils.Error(c, http.StatusBadRequest, "Tel is required")
		return
	}
	if len(tel) != 10 || !utils.IsAllDigits(tel) {
		utils.Error(c, http.StatusBadRequest, "MemberTel must be 10 digits")
		return
	}

	findMember, err := services.FindExistMemberTel(strings.TrimSpace(tel))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findMember == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("ไม่พบสมาชิกเบอร์: %s", tel))
		return
	}

	totalPoint := findMember.TotalPoint
	reqValue := req.TotalPoint

	// แปลงให้เป็นค่าลบเสมอ (เพื่อใช้หัก)
	deductTotalPoint := reqValue
	if reqValue > 0 {
		deductTotalPoint = reqValue * -1
	}

	// ตรวจสอบว่า Joylicoin พอให้หักไหม
	if (deductTotalPoint * -1) > totalPoint {
		utils.Error(c, http.StatusBadRequest,
			fmt.Sprintf("สมาชิกเบอร์: %s มี Total point ไม่เพียงพอ", tel))
		return
	}

	updateMember, err := services.UpdatePoint(deductTotalPoint, tel)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update member point: %v", err))
		return
	}

	utils.Success(c, "Member point updated successfully", updateMember)
}

// GetMemberByTelAndUpdate godoc
// @Summary Get member by tel
// @Tags Member
// @Security BasicAuth
// @Security BearerAuth
// @Param   tel  path  string  true  "Tel"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/member/search-sync/{tel} [get]
func GetMemberByTelAndUpdate(c *gin.Context) {
	tel := c.Param("tel")

	// findMember, err := services.FindExistMemberTel(tel)
	findMember, err := services.FindMemberWithLatestTierByTelAndUpdate(tel)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve member: %v", err))
		return
	}

	utils.Success(c, "Member retrieved successfully", findMember)
}
