package models

import (
	"time"

	"github.com/google/uuid"
)

type RedeemHead struct {
	RhID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"rh_id"`
	RhBillNo      string    `gorm:"type:varchar(30)" json:"rh_bill_no"`
	RhDate        time.Time `gorm:"type:timestamp;default:now()" json:"rh_date"`
	RhPosID       string    `gorm:"type:varchar(20)" json:"rh_pos_id"`
	RhLocation    string    `gorm:"type:varchar(5)" json:"rh_location"`
	RhUser        string    `gorm:"type:varchar(45)" json:"rh_user"`
	RhMemberTel   string    `gorm:"type:varchar(15)" json:"rh_member_tel"`
	RhRedeemPrice int       `gorm:"type:int" json:"rh_redeem_price"`
	RhBillStatus  string    `gorm:"type:varchar(5);default:'Y'" json:"rh_bill_status"`
	RhType        string    `gorm:"type:varchar(50)" json:"rh_type"`
}

func (RedeemHead) TableName() string {
	return "redeem_head"
}

type RedeemSub struct {
	RsID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"rs_id"`
	RsBillNo    string    `gorm:"type:varchar(30)" json:"rs_bill_no"`
	RsCode      string    `gorm:"type:varchar(20)" json:"rs_code"`
	RsName      string    `gorm:"type:varchar(150)" json:"rs_name"`
	RsQty       int       `gorm:"type:int" json:"rs_qty"`
	RsPrice     int       `gorm:"type:int" json:"rs_price"`
	RsJoylicoin int       `gorm:"type:int" json:"rs_joylicoin"`
}

func (RedeemSub) TableName() string {
	return "redeem_sub"
}
