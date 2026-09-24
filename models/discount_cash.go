package models

import (
	"time"

	"github.com/google/uuid"
)

type DiscountCashDto struct {
	MemberTel                string     `json:"member_tel"`
	BalanceDiscountCash      float64    `json:"balance_discount_cash"`
	BalanceDiscountNearExpire float64    `json:"balance_discount_near_expire"`
	DiscountCashNearExpireDate   *time.Time `json:"discount_cash_near_expire_date"`
	Discounts *[]DiscountsDto `json:"discounts"`
}

type DiscountsDto struct {
	DepositId  *uuid.UUID `json:"deposit_id"`
	BalanceDiscountCash      float64    `json:"balance_discount_cash"`
	DiscountCashExpireDate *time.Time `json:"discount_cash_expire_date"`
}

type DiscountCashDtoDeduct struct {
	DepositId *uuid.UUID `json:"deposit_id"`
	DiscountCashAmount    float64 `json:"discount_cash_amount"`
}