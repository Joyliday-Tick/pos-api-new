package model_v2

import "time"

type VerifyCardRequest struct {
	CardNo string `json:"card_no" validate:"required"`
	AssetId string `json:"asset_id" validate:"required"`
	BranchId string `json:"branch_id" validate:"required"`
	MachineId int `json:"machine_id" validate:"required"`
	ItemId int `json:"item_id" validate:"required"`
	GamePrice int `json:"game_price" validate:"required"`
	DeductCondition string `json:"deduct_condition" validate:"required"`
	HeadId int `json:"head_id" validate:"required"`
	DiscountCash *int `json:"discount_cash"`
}

type VerifyCardResponse struct {
	Status string `json:"status" validate:"required"`
	Message string `json:"message" validate:"required"`
	Data VerifyCardResponseData `json:"data" validate:"required"`
}

type VerifyCardResponseData struct {
	CardType string `json:"card_type" validate:"required"`
	CardNo string `json:"card_no" validate:"required"`
	CardId string `json:"card_id" validate:"required"`
	MemberTel string `json:"member_tel" validate:"required"`
	BalanceEcoin int `json:"balance_ecoin" validate:"required"`
	BalanceEbonus int `json:"balance_ebonus" validate:"required"`
	BalanceEtimes int `json:"balance_etimes" validate:"required"`
	TimeRemaining *time.Time `json:"time_remaining" validate:"required"`	
	HeadId int `json:"head_id" validate:"required"`
}

type VerifyCardSmcResponse struct {
	Status string `json:"status" validate:"required"`
	Message string `json:"message" validate:"required"`
	Data VerifyCardResponseData `json:"data" validate:"required"`
	SmcData SMCPrizeData `json:"smc_data" validate:"required"`
}