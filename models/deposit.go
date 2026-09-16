package models

import (
	"time"

	"github.com/google/uuid"
)

type DepositHead struct {
	DhID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"dh_id"`
	DhBillNo           string    `gorm:"type:varchar(30)" json:"dh_bill_no"`
	DhDate             time.Time `gorm:"type:timestamp;default:now()" json:"dh_date"`
	DhPosID            string    `gorm:"type:int" json:"dh_pos_id"`
	DhLocation         string    `gorm:"type:varchar(5)" json:"dh_location"`
	DhUser             string    `gorm:"type:varchar(45)" json:"dh_user"`
	DhMemberTel        string    `gorm:"type:varchar(15)" json:"dh_member_tel"`
	DhSpending         int       `gorm:"type:int" json:"dh_spending"`
	DhDeposit          int       `gorm:"type:int" json:"dh_deposit"`
	DhDiscount         float64   `gorm:"type:numeric(5,2)" json:"dh_discount"`
	DhJoylicoin        int       `gorm:"type:int" json:"dh_joylicoin"`
	DhBillStatus       string    `gorm:"type:varchar(5);default:'Y'" json:"dh_bill_status"`
	DhType             string    `gorm:"type:varchar(50)" json:"dh_type"`
	DhDepositJoylicoin int       `gorm:"type:int;default:0" json:"dh_deposit_joylicoin"`
}

func (DepositHead) TableName() string {
	return "deposit_head"
}

type DepositSub struct {
	DsID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"ds_id"`
	DsBillNo    string    `gorm:"type:varchar(30)" json:"ds_bill_no"`
	DsCode      string    `gorm:"type:varchar(20)" json:"ds_code"`
	DsName      string    `gorm:"type:varchar(150)" json:"ds_name"`
	DsQty       int       `gorm:"type:int" json:"ds_qty"`
	DsPrice     int       `gorm:"type:int" json:"ds_price"`
	DsJoylicoin int       `gorm:"type:int;default:0" json:"ds_joylicoin"`
}

func (DepositSub) TableName() string {
	return "deposit_sub"
}

type DepositHeadCron struct {
	DhID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"dh_id"`
	DhBillNo           string    `gorm:"type:varchar(30)" json:"dh_bill_no"`
	DhDate             time.Time `gorm:"type:timestamp;default:now()" json:"dh_date"`
	DhPosID            string    `gorm:"type:int" json:"dh_pos_id"`
	DhLocation         string    `gorm:"type:varchar(5)" json:"dh_location"`
	DhUser             string    `gorm:"type:varchar(45)" json:"dh_user"`
	DhMemberTel        string    `gorm:"type:varchar(15)" json:"dh_member_tel"`
	DhSpending         int       `gorm:"type:int" json:"dh_spending"`
	DhDeposit          int       `gorm:"type:int" json:"dh_deposit"`
	DhDiscount         float64   `gorm:"type:numeric(5,2)" json:"dh_discount"`
	DhJoylicoin        int       `gorm:"type:int" json:"dh_joylicoin"`
	DhBillStatus       string    `gorm:"type:varchar(5);default:'Y'" json:"dh_bill_status"`
	DhType             string    `gorm:"type:varchar(50)" json:"dh_type"`
	DhDepositJoylicoin int       `gorm:"type:int;default:0" json:"dh_deposit_joylicoin"`
}

func (DepositHeadCron) TableName() string {
	return "deposit_head_cron"
}

type DepositSubCron struct {
	DsID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"ds_id"`
	DsBillNo    string    `gorm:"type:varchar(30)" json:"ds_bill_no"`
	DsCode      string    `gorm:"type:varchar(20)" json:"ds_code"`
	DsName      string    `gorm:"type:varchar(150)" json:"ds_name"`
	DsQty       int       `gorm:"type:int" json:"ds_qty"`
	DsPrice     int       `gorm:"type:int" json:"ds_price"`
	DsJoylicoin int       `gorm:"type:int;default:0" json:"ds_joylicoin"`
}

func (DepositSubCron) TableName() string {
	return "deposit_sub_cron"
}
