package services

import (
	"fmt"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"
	"time"
)

func GetDiscountCashByMemberTel(memberTel string) (*models.DiscountCashDto, error){
	var deposit []models.CardDeposit
	var discountCash models.DiscountCashDto
	var discounts []models.DiscountsDto
	var balanceDiscountCash float32
	var balanceDiscountNearExpire float32
	var discountNearExpireDate *time.Time
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database pos connection is nil")
	}
	currentData := utils.TimeNowAsia()
	config.DB_POS.Find(&deposit, "member_tel = ? and is_active = true and discount_cash > 0 and discount_cash_expire > ? order by discount_cash_expire asc", memberTel, currentData)
	for _, discount := range deposit {
		balanceDiscountCash += discount.BalanceDiscountCash
		discounts = append(discounts, models.DiscountsDto{
			DepositId:       &discount.ID,
			BalanceDiscountCash: discount.BalanceDiscountCash,
			DiscountCashExpireDate: discount.DiscountCashExpire,
		})
		// first loop will be nearest expire
		if discountNearExpireDate == nil {
			balanceDiscountNearExpire = discount.BalanceDiscountCash
			discountNearExpireDate = discount.DiscountCashExpire
		}
	}
	discountCash = models.DiscountCashDto{
		BalanceDiscountCash: balanceDiscountCash,
		Discounts: &discounts,
		BalanceDiscountNearExpire: balanceDiscountNearExpire,
		DiscountCashNearExpireDate: discountNearExpireDate,
	}
	
	

	
	return  &discountCash, nil
}