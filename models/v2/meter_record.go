package model_v2

import (
	"time"

	"github.com/google/uuid"
)

type MeterRecord struct {
	RcID            uuid.UUID  `json:"rc_id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	RcDate          *time.Time `json:"rc_date"`
	RcAssetID       *int       `json:"rc_asset_id"`
	RcAssetHead     *int       `json:"rc_asset_head"`
	RcAssetLocation *string    `json:"rc_asset_location"`
	RcMember        *string    `json:"rc_member"`
	RcCardID        *string    `json:"rc_card_id"`
	RcEcoin         *int       `json:"rc_ecoin"`
	RcBonus         *int       `json:"rc_bonus"`
	RnEcoin         *int       `json:"rn_ecoin"`
	RnBonus         *int       `json:"rn_bonus"`
	PrizeStatus     *string    `json:"prize_status" gorm:"default:Y"`
	BonusStatus     *string    `json:"bonus_status" gorm:"default:Y"`
	OldID           *int       `json:"old_id"`
	CardType        *string    `json:"card_type"`
	RcTime          *int       `json:"rc_time" gorm:"default:0"`
	RnTime          *int       `json:"rn_time" gorm:"default:0"`
}

func (MeterRecord) TableName() string {
	return "meter_record"
}

type MeterRecordDto struct {
	RcDate          *time.Time `json:"rc_date"`
	RcAssetID       int        `json:"rc_asset_id"`
	RcAssetHead     int        `json:"rc_asset_head"`
	RcAssetLocation string     `json:"rc_asset_location"`
	RcMember        string     `json:"rc_member"`
	RcCardID        string     `json:"rc_card_id"`
	RcEcoin         int        `json:"rc_ecoin"`
	RcBonus         int        `json:"rc_bonus"`
	RnEcoin         int        `json:"rn_ecoin"`
	RnBonus         int        `json:"rn_bonus"`
	CardType        string     `json:"card_type"`
	RcTime          int        `json:"rc_time"`
	RnTime          int        `json:"rn_time"`
	OldID		   int        `json:"old_id"`
}

type CardPlayMachineDeductDto struct {
	BeforeEcoin   *int	 `json:"before_ecoin"`
	BeforeEbonus   *int     `json:"before_ebonus"`
	BeforeETimes   *int     `json:"before_etimes"`
	AfterEcoin   *int	 `json:"after_ecoin"`
	AfterEbonus   *int     `json:"after_ebonus"`
	AfterETimes   *int     `json:"after_etimes"`
}