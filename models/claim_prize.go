package models

import (
	"time"

	"github.com/google/uuid"
)

type ClaimPrizeDto struct {
	CPDate       time.Time `json:"-"`
	MemberTel    string    `json:"member_tel"`
	SpendAmount  int       `json:"spend_amount"`
	SpendECoin   int       `json:"spend_ecoin"`
	SpendEBonus  int       `json:"spend_ebonus"`
	Discount     int       `json:"discount"`
	ProductPrice int       `json:"product_price"`
	OptionID     *int      `json:"option_id"`
	CardNo       *string   `json:"card_no"`
	ClaimUser    string    `json:"claim_user"`
	Joylicoin    int       `json:"joylicoin"`
	ECoin        int       `json:"e_coin"`
	EBonus       int       `json:"e_bonus"`
	Location     string    `json:"location"`
	ClaimAmount  int       `json:"claim_amount"`
}

type ClaimPrize struct {
	CPID         uuid.UUID `gorm:"column:cp_id;type:uuid;default:uuid_generate_v4();primaryKey" json:"cp_id"`
	CPDate       time.Time `gorm:"column:cp_date;type:timestamp;not null" json:"cp_date"`
	MemberTel    string    `gorm:"column:member_tel;type:varchar(20)" json:"member_tel"`
	CPSpend      int       `gorm:"column:cp_spend;default:0;not null" json:"cp_spend"`
	SpendECoin   int       `gorm:"column:spend_ecoin;default:0;not null" json:"spend_ecoin"`
	SpendEBonus  int       `gorm:"column:spend_ebonus;default:0;not null" json:"spend_ebonus"`
	Discount     int       `gorm:"column:discount;default:0;not null" json:"discount"`
	ProductPrice int       `gorm:"column:product_price;default:0;not null" json:"product_price"`
	OptionID     *int      `gorm:"column:option_id" json:"option_id"`
	CardNo       *string   `gorm:"column:card_no;type:varchar(20)" json:"card_no"`
	CPUser       string    `gorm:"column:cp_user;type:varchar(45);not null" json:"cp_user"`
	CPLocation   string    `gorm:"column:cp_location;type:varchar(10)" json:"cp_location"`
	ClaimAmount  int       `gorm:"column:claim_amount;default:0;not null" json:"claim_amount"`
}

func (ClaimPrize) TableName() string {
	return "claim_prize"
}

type SearchClaimPrizeParams struct {
	MemberTel string
	Location  string
	StartDate string
	EndDate   string
	Page      int
	Skip      int
}

type ExportClaimPrizeParams struct {
	MemberTel string
	Location  string
	StartDate string
	EndDate   string
}

type ClaimPrizeList struct {
	Date         time.Time `json:"date"`
	MemberTel    string    `json:"member_tel"`
	SpendAmount  int       `json:"spend_amount"`
	SpendECoin   int       `json:"spend_ecoin"`
	SpendEBonus  int       `json:"spend_ebonus"`
	Discount     int       `json:"discount"`
	ProductPrice int       `json:"product_price"`
	CardNo       *string   `json:"card_no"`
	Cashier      string    `json:"cashier"`
	Location     string    `json:"location"`
	ClaimAmount  int       `json:"claim_amount"`
	ActionName   string    `json:"action_name"`
}

type ExportClaimPrizeList struct {
	Date         time.Time `json:"date"`
	MemberTel    string    `json:"member_tel"`
	SpendAmount  int       `json:"spend_amount"`
	SpendECoin   int       `json:"spend_ecoin"`
	SpendEBonus  int       `json:"spend_ebonus"`
	Discount     int       `json:"discount"`
	ProductPrice int       `json:"product_price"`
	CardNo       *string   `json:"card_no"`
	Cashier      string    `json:"cashier"`
	Location     string    `json:"location"`
	ClaimAmount  int       `json:"claim_amount"`
	ActionName   string    `json:"action_name"`
	MemberID     *int      `json:"member_id"`
}

type ClaimPrizeData struct {
	ClaimPrizeList  []ClaimPrizeList `json:"claim_prize_list"`
	ClaimPrizeTotal CliamprizeTotal  `json:"claim_prize_total"`
}

type CliamprizeTotal struct {
	TotalSpending     int64                `json:"total_spending"`
	TotalECoin        int64                `json:"total_ecoin"`
	TotalEBonus       int64                `json:"total_ebonus"`
	TotalProductPrice int64                `json:"total_product_price"`
	Actions           []CliamPrizeByAction `json:"actions"`
}

type CliamPrizeByAction struct {
	ActionID   int    `json:"action_id"`
	ActionName string `json:"action_name"`
	TotalClaim int64  `json:"total_claim"`
}
