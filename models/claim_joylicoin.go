package models

import (
	"time"

	"github.com/google/uuid"
)

type MeterRecordData struct {
	ID           int    `json:"id"`
	Date         string `json:"date"`
	AssetID      int    `json:"asset_id"`
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

type StampHouseData struct {
	ID         uuid.UUID `json:"id"`
	MemberTel  string    `json:"member_tel"`
	StampCount int       `json:"stamp_count"`
	EStamp     int       `json:"e_stamp"`
	Status     string    `json:"status"`
	Date       time.Time `json:"date"`
	Machine    string    `json:"machine"`
	Location   string    `json:"location"`
	Sync       string    `json:"sync"`
	LastEstamp int       `json:"last_estamp"`
}

type SearchMeterRecordParams struct {
	MemberTel string
}

type SearchCheckCard struct {
	CardNo string
}
