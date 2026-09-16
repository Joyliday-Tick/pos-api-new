package models

import (
	"time"

	"github.com/google/uuid"
)

type PosSubTransactionDto struct {
	BillNo        string `json:"bill_no"`
	ProductID     int    `json:"product_id"`
	Qty           int    `json:"qty"`
	Price         int    `json:"price"`
	ECoin         int    `json:"e_coin"`
	EBonus        int    `json:"e_bonus"`
	CardDepositId string `json:"card_deposit_id"`
}

type PosSubTransactionData struct {
	BillNo        string `json:"bill_no"`
	MenuName      string `json:"menu_name"`
	ProductID     int    `json:"product_id"`
	Qty           int    `json:"qty"`
	Price         int    `json:"price"`
	ECoin         int    `json:"e_coin"`
	EBonus        int    `json:"e_bonus"`
	CardDepositId string `json:"card_deposit_id"`
}

type PosSubTransaction struct {
	ID            uuid.UUID  `gorm:"column:id;type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	BillNo        string     `gorm:"column:bill_no;type:varchar(50)" json:"bill_no"`
	ProductID     int        `gorm:"column:product_id" json:"product_id"`
	Qty           int        `gorm:"column:qty" json:"qty"`
	Price         int        `gorm:"column:price;default:0" json:"price"`
	ECoin         int        `gorm:"column:e_coin;default:0" json:"e_coin"`
	EBonus        int        `gorm:"column:e_bonus;default:0" json:"e_bonus"`
	CreateBy      int        `gorm:"column:create_by" json:"create_by"`
	CreateDate    time.Time  `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateBy      *int       `gorm:"column:update_by" json:"update_by"`
	UpdateDate    *time.Time `gorm:"column:update_date" json:"update_date"`
	CardDepositId uuid.UUID  `json:"card_deposit_id"`
}

func (PosSubTransaction) TableName() string {
	return "pos_sub_transaction"
}
