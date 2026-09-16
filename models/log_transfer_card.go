package models

import (
	"time"

	"github.com/google/uuid"
)

type LogTransferCardDto struct {
	OldCardNo      *string `json:"old_card_no"`
	MemberTel      *string `json:"member_tel"`
	BalanceEcoin   int     `json:"-"`
	BalanceEbonus  int     `json:"-"`
	TransferCardNo *string `json:"transfer_card_no"`
	CreatedBy      *string `json:"created_by"`
	Location       *string `json:"location"`
}

type LogTransferCard struct {
	ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	TransferDate   time.Time `gorm:"column:transfer_date;default:now()" json:"transfer_date"`
	OldCardNo      *string   `gorm:"column:old_card_no;size:15" json:"old_card_no"`
	MemberTel      *string   `gorm:"column:member_tel;size:20" json:"member_tel"`
	BalanceEcoin   int       `gorm:"column:balance_ecoin;default:0" json:"balance_ecoin"`
	BalanceEbonus  int       `gorm:"column:balance_ebonus;default:0" json:"balance_ebonus"`
	TransferCardNo *string   `gorm:"column:transfer_card_no;size:15" json:"transfer_card_no"`
	CreatedBy      *string   `gorm:"column:created_by;size:45" json:"created_by"`
	Location       *string   `gorm:"column:location;size:10" json:"location"`
}

func (LogTransferCard) TableName() string {
	return "log_transfer_card"
}
