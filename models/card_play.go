package models

import (
	"time"

	"github.com/google/uuid"
)

type CardPlayDto struct {
	CardNo      string `json:"card_no"`
	PlayBranch  bool   `json:"play_branch"`
	PlayMachine bool   `json:"play_machine"`
}

// create a new CardPlay DTO from CardPlay
type NewCardPlayDto struct {
	CardNo        string     `json:"card_no"`
	PlayBranch    bool       `json:"play_branch"`
	PlayMachine   bool       `json:"play_machine"`
	CardDepositId *uuid.UUID `json:"card_deposit_id"`
}

type CardPlay struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CardNo        string     `json:"card_no"`
	PlayBranch    bool       `json:"play_branch"`
	PlayMachine   bool       `json:"play_machine"`
	CreateBy      int        `json:"create_by"`
	CreateDate    time.Time  `json:"create_date" gorm:"default:CURRENT_TIMESTAMP"`
	UpdateBy      *int       `json:"update_by"`
	UpdateDate    *time.Time `json:"update_date"`
	IsDelete      bool       `json:"is_delete" gorm:"default:false"`
	DeleteBy      *int       `json:"delete_by"`
	DeleteDate    *time.Time `json:"delete_date"`
	CardDepositId *uuid.UUID `json:"card_deposit_id" gorm:"type:uuid;"`
}

func (CardPlay) TableName() string {
	return "card_play"
}

// card play verify DTO
type CardPlayVerifyDto struct {
	CardNo      string `json:"card_no"`
	PlayBranch  string `json:"play_branch"`
	PlayMachine string `json:"play_machine"`
	UseCoin     *int   `json:"use_coin"`
	UseBonus    *int   `json:"use_bonus"`
}

// card play deduct DTO
type CardPlayDeductDto struct {
	CardNo     string     `json:"card_no"`
	CardPlayId *uuid.UUID `json:"card_play_id"`
	ECoin      int        `json:"e_coin"`
	EBonus     int        `json:"e_bonus"`
	FromChannel *string     `json:"from_channel"`
	DiscountCashAmount *float32 `json:"discount_cash_amount"`
}
type CardPlayVerifyResponse struct {
	CardPlayId    *uuid.UUID     `json:"card_play_id"`
	CardNo        string         `json:"card_no"`
	PlayBranch    bool           `json:"play_branch"`
	PlayMachine   bool           `json:"play_machine"`
	BranchList    *[]BranchList  `json:"branch_list"`
	MachineList   *[]MachineList `json:"machine_list"`
	Period        *Period        `json:"period"`
	LimitUsage    *int           `json:"limit_usage"`
	UsedTime      *int           `json:"usage_time"`
	ConditionID   int
	CardDepositId *uuid.UUID `json:"card_deposit_id"`
}
type BranchList struct {
	BranchCode string `json:"branch_code"`
}
type MachineList struct {
	Machine int `json:"machine"`
}
type Period struct {
	StartDate time.Time
	EndDate   time.Time
}
type WithdrawUsed struct {
	CardDepositId *uuid.UUID
	RemoveCoin    int64
}
