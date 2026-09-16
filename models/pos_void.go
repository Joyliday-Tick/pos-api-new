package models

import (
	"time"

	"github.com/google/uuid"
)

type PosVoidDto struct {
	BillNo     string `json:"bill_no"`
	VoidUser   string `json:"void_user"`
	VoidReason string `json:"void_reason"`
	CardNo     string `json:"card_no"`
	DeductCard bool   `json:"deduct_card"`
}

type PosVoid struct {
	VoidID     uuid.UUID `gorm:"column:void_id;type:uuid;default:uuid_generate_v4();primaryKey" json:"void_id"`
	BillNo     string    `gorm:"column:bill_no;type:varchar(50)" json:"bill_no"`
	VoidDate   time.Time `gorm:"column:void_date;default:now()" json:"void_date"`
	VoidUser   string    `gorm:"column:void_user;type:varchar(100)" json:"void_user"`
	VoidReason string    `gorm:"column:void_reason;type:varchar(100)" json:"void_reason"`
}

func (PosVoid) TableName() string {
	return "pos_void"
}
