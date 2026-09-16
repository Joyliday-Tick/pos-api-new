package model_v2

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SMCLoginRequest is the payload sent to /v1/login
type SMCLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SMCLoginResponse is the response from /v1/login
type SMCLoginResponse struct {
	Token string `json:"token"`
}

type SMCPrizeDto struct {
	AssetId      string `json:"asset_id"`
	DiscountCash int    `json:"discount_cash"`
	Ebonus       int    `json:"ebonus"`
	ECoin        int    `json:"ecoin"`
	ItemId       int    `json:"item_id"`
	MemberTel    string `json:"member_tel"`
	CardNo       string `json:"card_no"`
}
type SmcTk struct {
	Token     string     `json:"token"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	IsActive  bool       `json:"is_active,omitempty"`
}

func (SmcTk) TableName() string {
	return "smc_tk"
}

type SMCPrizeResponse struct {
	Data    SMCPrizeData `json:"data"`
	Success bool         `json:"success"`
}
type SMCPrizeData struct {
	MemberTel              string       `json:"member_tel"`
	AssetID                string       `json:"asset_id"`
	DateTime               time.Time    `json:"date_time"`
	CalculateBalanceEcoin  int          `json:"calculate_balance_ecoin"`
	CalculateBalanceEBonus int          `json:"calculate_balance_ebonus"`
	DateTotalEBonus        int          `json:"date_total_ebonus"`
	DateTotalEcoin         int          `json:"date_total_ecoin"`
	GetPrize               bool         `json:"get_prize"`
	DiscountAmount         float64      `json:"discount"`
	DiscountMessage        string       `json:"discount_message"`
	SpecialDiscount        *DiscountDto `json:"special_discount,omitempty"`
	DiscountCode           *string      `json:"discount_code,omitempty"`
	DiscountCash           *int         `json:"discount_cash,omitempty"`
}
type DiscountDto struct {
	DiscountID     primitive.ObjectID `json:"discount_id"`
	DiscountCodeID primitive.ObjectID `json:"discount_code_id"`
	DiscountCode   string             `json:"discount_code"`
	DiscountValue  int                `json:"discount_value"`
	MemberTel      string             `json:"member_tel"`
	UsageDate      time.Time          `json:"usage_date"`
	Message        string             `json:"message"`
}
