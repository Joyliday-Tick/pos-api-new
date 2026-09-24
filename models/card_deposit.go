package models

import (
	"time"

	"github.com/google/uuid"
)

type CardDepositDto struct {
	FromChannel     string     `json:"from_channel"`
	CardNo          string     `json:"card_no"`
	MemberTel       string     `json:"member_tel"`
	Amount          float64    `json:"amount"`
	Coin            int        `json:"coin"`
	Bonus           int        `json:"bonus"`
	BalanceCoin     int        `json:"-"`
	BalanceBonus    int        `json:"-"`
	TotalTimes      int        `json:"-"`
	PosId           string     `json:"-"`
	PosMenuId       int        `json:"pos_menu_id"`
	BonusExpireDate *time.Time `json:"bonus_expire_date"`
	BillNo          *string    `json:"bill_no"`
	OriginPrice     *float64   `json:"origin_price"`
	OriginQuantity  *int       `json:"origin_quantity"`
	OriginCoin      *int       `json:"origin_coin"`
	OriginBonus     *int       `json:"origin_bonus"`
	CardExpireDate  *time.Time `json:"card_expire_date"`
	DiscountCash    float64        `json:"discount_cash"`
	BalanceDiscountCash float64        `json:"balance_discount_cash"`
	DiscountCashExpire *time.Time `json:"discount_cash_expire"`
}

type CardDeposit struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	FromChannel     string     `gorm:"column:from_channel" json:"from_channel"`
	CardNo          string     `gorm:"column:card_no;not null" json:"card_no"`
	MemberTel       string     `gorm:"column:member_tel;not null" json:"member_tel"`
	Amount          float64    `gorm:"type:numeric(10,2);not null" json:"amount"`
	CreateDate      time.Time  `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateDate      *time.Time `gorm:"column:update_date" json:"update_date"`
	Coin            int        `gorm:"column:coin;default:0" json:"coin"`
	Bonus           int        `gorm:"column:bonus;default:0" json:"bonus"`
	BalanceCoin     int        `gorm:"column:balance_coin;default:0" json:"balance_coin"`
	BalanceBonus    int        `gorm:"column:balance_bonus;default:0" json:"balance_bonus"`
	TotalTimes      int        `gorm:"column:total_times;default:0" json:"total_times"`
	PosId           string     `gorm:"column:pos_id" json:"pos_id"`
	PosMenuId       int        `gorm:"column:pos_menu_id;not null" json:"pos_menu_id"`
	BonusExpireDate *time.Time `gorm:"column:bonus_expire_date" json:"bonus_expire_date"`
	BillNo          *string    `gorm:"column:bill_no" json:"bill_no"`
	IsActive        bool       `gorm:"column:is_active;default:true" json:"is_active"`
	CardExpireDate  *time.Time `gorm:"column:card_expire_date" json:"card_expire_date"`
	DiscountCash    float64        `gorm:"column:discount_cash;default:0" json:"discount_cash"`
	BalanceDiscountCash float64        `gorm:"column:balance_discount_cash;default:0" json:"balance_discount_cash"`
	DiscountCashExpire *time.Time `gorm:"column:discount_cash_expire" json:"discount_cash_expire"`
}

type CardDepositBalanceDto struct {
	ID           *uuid.UUID `json:"id"`
	BalanceCoin  int        `json:"balance_coin"`
	BalanceBonus int        `json:"balance_bonus"`
	BalanceDiscountCash *float64        `json:"balance_discount_cash"`
}

type CardDepositBonusExpireDto struct {
	BonusExpireTotal int        `json:"bonus_expire_total"`
	BonusExpireDate  *time.Time `json:"bonus_expire_date"`
}

func (CardDeposit) TableName() string {
	return "card_deposit"
}
