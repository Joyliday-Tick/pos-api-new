package models

import (
	"time"

	"github.com/google/uuid"
)

type PrizeCounterData struct {
	ID           uuid.UUID `json:"id"`
	Date         string    `json:"date"`
	AssetID      int       `json:"asset_id"`
	MachineAsset string    `json:"machine_asset"`
	MachineName  string    `json:"machine_name"`
	MachineNo    string    `json:"machine_no"`
	McHead       int       `json:"mc_head"`
	Sku          string    `json:"sku"`
	ProductName  string    `json:"product_name"`
	ProductID    int       `json:"product_id"`
	Price        int       `json:"price"`
	Joylicoin    int       `json:"joylicoin"`
	MemberTel    string    `json:"member_tel"`
	Status       string    `json:"status"`
	Location     string    `json:"location"`
}

type MeterRecordSumEcoin struct {
	RCEcoin int `json:"used_ecoin" gorm:"column:used_ecoin"`
	RCBonus int `json:"used_bonus" gorm:"column:used_bonus"`
}

type SearchPrizeCounterParams struct {
	MemberTel string `json:"member_tel"`
	Location  string `json:"location"`
}

type SearchMeterParams struct {
	MemberTel string `json:"member_tel"`
}

type MeterIds struct {
	MeterIds []string `json:"meter_ids"`
}

type DepositJubuJibi struct {
	PosID            string               `json:"pos_id"`
	Location         string               `json:"location"`
	Cashier          string               `json:"cashier"`
	MemberTel        string               `json:"member_tel"`
	Spending         int                  `json:"spending"`
	Deposit          int                  `json:"deposit"`
	Discount         float64              `json:"discount"`
	Joylicoin        int                  `json:"joylicoin"`
	DepositJoylicoin int                  `json:"deposit_joylicoin"`
	TopupDeduct      int                  `json:"topup_deduct"`
	Type             string               `json:"-"`
	SubItems         []DepositSubJubuJibi `json:"sub_items"`
}

type DepositSubJubuJibi struct {
	Sku         string `json:"sku"`
	ProductName string `json:"product_name"`
	Qty         int    `json:"qty"`
	Price       int    `json:"price"`
	Joylicoin   int    `json:"joylicoin"`
}

type MeterRecordJubuJibi struct {
	ID           int    `json:"id"`
	Date         string `json:"date"`
	MachineNo    int    `json:"machine_no"`
	MachineAsset string `json:"machine_asset"`
	MachineName  string `json:"machine_name"`
	Category     string `json:"category"`
	AssetHead    int    `json:"asset_head"`
	Location     string `json:"location"`
	MemberTel    string `json:"member_tel"`
	CardNo       string `json:"card_no"`
	RCEcoin      int    `json:"used_ecoin"`
	RCBonus      int    `json:"used_bonus"`
	RNBonus      int    `json:"balance_bonus"`
	RNEcoin      int    `json:"balance_ecoin"`
	BonusStatus  string `json:"bonus_status"`
}

type MeterRecordAllSpend struct {
	ID           string    `json:"id"`
	Date         time.Time `json:"date"`
	MachineNo    int       `json:"machine_no"`
	MachineAsset string    `json:"machine_asset"`
	MachineName  string    `json:"machine_name"`
	Category     string    `json:"category"`
	AssetHead    int       `json:"asset_head"`
	Location     string    `json:"location"`
	MemberTel    string    `json:"member_tel"`
	CardNo       string    `json:"card_no"`
	RCEcoin      int       `json:"used_ecoin"`
	RCBonus      int       `json:"used_bonus"`
	RNBonus      int       `json:"balance_bonus"`
	RNEcoin      int       `json:"balance_ecoin"`
	BonusStatus  string    `json:"bonus_status"`
}

type RedeemJubuJibi struct {
	PosID       string              `json:"pos_id"`
	Location    string              `json:"location"`
	Cashier     string              `json:"cashier"`
	MemberTel   string              `json:"member_tel"`
	RedeemPrice int                 `json:"redeem_price"`
	Type        string              `json:"-"`
	SubItems    []RedeemSubJubuJibi `json:"sub_items"`
}

type RedeemSubJubuJibi struct {
	Sku         string `json:"sku"`
	ProductName string `json:"product_name"`
	Qty         int    `json:"qty"`
	Price       int    `json:"price"`
	Joylicoin   int    `json:"joylicoin"`
}

type MeterRecordCheckCard struct {
	Date        time.Time `json:"date"`
	MachineNo   string    `json:"machine_no"`
	MachineCode string    `json:"machine_code"`
	MachineName string    `json:"machine_name"`
	Location    string    `json:"location"`
	MemberTel   string    `json:"member_tel"`
	CardNo      string    `json:"card_no"`
	RCEcoin     int       `json:"used_ecoin"`
	RCBonus     int       `json:"used_bonus"`
	RNBonus     int       `json:"balance_bonus"`
	RNEcoin     int       `json:"balance_ecoin"`
	RCTime      int       `json:"used_time"`
	RNTime      int       `json:"balance_time"`
	CardType    string    `json:"card_type"`
}
