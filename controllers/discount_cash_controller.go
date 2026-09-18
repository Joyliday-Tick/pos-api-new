package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/middlewares"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// @Summry get discount cash by memberTel
// @Tags Discount Cash
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Parameter member_tel
// @Success 200 {object} models.DiscountCashDto
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/discountcash/:member_tel [get]
func GetDiscountCashByMemberTel(c *gin.Context) {
	userId, _ := middlewares.GetUserIdFromClaims(c)
	memberTel := c.Param("member_tel")
	// คงไว้เพราะเป็นร่องรอยว่าใครดูข้อมูลของใคร แต่ปิดบังเบอร์
	// (userId ยังระบุคนดูได้ครบ ส่วนเบอร์เต็มอยู่ในฐานข้อมูลอยู่แล้ว)
	fmt.Printf("User %d is requesting discount cash for memberTel: %s\n", userId, utils.MaskTel(memberTel))
	discountCash, err := services.GetDiscountCashByMemberTel(memberTel)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.Success(c, "success", discountCash)
}
