package models

import (
	"time"

	"github.com/google/uuid"
)

type RefundComfirmDto struct {
	ReqID          string `json:"req_id"`
	RefconUser     string `json:"refcon_user"`
	RefconLocation string `json:"refcon_location"`
	RefundBy       string `json:"refund_by"`
	RefundEcoin    int    `json:"refcon_ecoin"`
	RefundEbonus   int    `json:"refcon_ebonus"`
	CardNo         string `json:"ref_card_no"`
	RefMemberTel   string `json:"ref_member_tel"`
}

type RefundComfirm struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ReqID          string     `gorm:"type:varchar(45)" json:"req_id"`
	RefconDate     *time.Time `gorm:"type:timestamp" json:"refcon_date"`
	RefconUser     string     `gorm:"type:varchar(30)" json:"refcon_user"`
	RefconLocation string     `gorm:"type:varchar(10)" json:"refcon_location"`
	RefconStatus   string     `gorm:"type:varchar(10);default:Y" json:"refcon_status"`
	RefundBy       string     `gorm:"type:varchar(30)" json:"refund_by"`
	MemberTel      string     `gorm:"type:varchar(10)" json:"member_tel"`
}

func (RefundComfirm) TableName() string {
	return "refund_confirm"
}
