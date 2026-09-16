package models

import (
	"time"

	"github.com/google/uuid"
)


type CardWithdraw struct {
	ID *uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	FromChannel string `json:"from_channel"`
	CardNo string `json:"card_no"`
	MemberTel string `json:"member_tel"`
	AmountEcoin int `json:"amount_ecoin" default:"0"`
	AmountEbonus int `json:"amount_ebonus" default:"0"`
	CardDepositId *uuid.UUID `json:"card_deposit_id"`
	DateCreated time.Time `json:"date_created" gorm:"default:CURRENT_TIMESTAMP"`
	AmountDiscountCash float32 `json:"amount_discount_cash" default:"0"`

}

 type CardWithdrawDto struct {
	FromChannel string `json:"from_channel"`
	CardNo string `json:"card_no"`
	MemberTel string `json:"member_tel"`
	AmountEcoin int `json:"amount_ecoin"`
	AmountEbonus int `json:"amount_ebonus"`
	CardDepositId *uuid.UUID `json:"card_deposit_id"`
	AmountDiscountCash float32 `json:"amount_discount_cash"`
}

func (CardWithdraw) TableName() string {
	return "card_withdraw"
}
