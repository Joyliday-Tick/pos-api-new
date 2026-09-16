package models

import (
	"time"

	"github.com/google/uuid"
)

type RefundRequestDto struct {
	RefundDate   time.Time `json:"-"`
	RefundTel    string    `json:"refund_tel"`
	RefundAmount int       `json:"refund_amount"`
	RefundMeter  string    `json:"refund_meter"`
	RefundEcoin  int       `json:"refund_ecoin"`
	RefundEbonus int       `json:"refund_ebonus"`
	CardNo       string    `json:"card_no"`
}

type RefundRequest struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	RefundDate   time.Time `gorm:"type:timestamp;not null" json:"refund_date"`
	RefundTel    string    `gorm:"type:varchar(10);not null" json:"refund_tel"`
	RefundAmount int       `gorm:"type:int;not null" json:"refund_amount"`
	RefundMeter  string    `gorm:"type:varchar(1500)" json:"refund_meter,omitempty"`
	RefundPz     string    `gorm:"type:varchar(500)" json:"refund_pz,omitempty"`
	RefundSts    string    `gorm:"type:varchar(10);default:Y;not null" json:"refund_sts"`
	RefundEcoin  int       `gorm:"type:int;default:0" json:"refund_ecoin"`
	RefundEbonus int       `gorm:"type:int;default:0" json:"refund_ebonus"`
	CardNo       string    `gorm:"type:varchar(20)" json:"card_no,omitempty"`
}

// TableName overrides the table name used by GORM
func (RefundRequest) TableName() string {
	return "refund_request"
}

type RefundRequestData struct {
	ID              uuid.UUID `json:"id"`
	RefundDate      time.Time `json:"refund_date"`
	RefundTel       string    `json:"refund_tel"`
	RefundAmount    int       `json:"refund_amount"`
	RefundEcoin     int       `json:"refund_ecoin"`
	RefundEbonus    int       `json:"refund_ebonus"`
	CardNo          string    `json:"card_no"`
	RefundCashCount int       `json:"refund_cash_count"`
	Firstname       string    `json:"firstname"`
	Lastname        string    `json:"lastname"`
}
