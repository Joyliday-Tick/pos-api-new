package models

import (
	"time"

	"github.com/google/uuid"
)

type LockCardDto struct {
	CardNo              string `json:"card_no"`
	MemberTel           string `json:"member_tel"`
	BalanceEcoin        int32  `json:"balance_ecoin"`
	BalanceEbonus       int32  `json:"balance_ebonus"`
	BalanceDiscountCash int32  `json:"balance_discount_cash"`
	// IsLock              bool       `json:"is_lock"`
	LockBy          string    `json:"lock_by"`
	LockDate        time.Time `json:"-"`
	LockLocation    string    `json:"lock_location"`
	LockReason      string    `json:"lock_reason"`
	RefCardEntityId string    `json:"ref_card_entity_id"`
}

type UnLockCardDto struct {
	LockId string `json:"lock_id"`
	// IsLock          bool       `json:"is_lock"`
	UnlockBy        string     `json:"unlock_by"`
	UnlockDate      *time.Time `json:"-"`
	UnlockLocation  string     `json:"unlock_location"`
	UnlockReason    string     `json:"unlock_reason"`
	RefCardEntityId string     `json:"ref_card_entity_id"`
}

type LogLockCard struct {
	ID                  uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	CardNo              string     `gorm:"type:varchar(15)" json:"card_no"`
	MemberTel           string     `gorm:"type:varchar(20)" json:"member_tel"`
	BalanceEcoin        int32      `gorm:"type:int4;default:0" json:"balance_ecoin"`
	BalanceEbonus       int32      `gorm:"type:int4;default:0" json:"balance_ebonus"`
	BalanceDiscountCash int32      `gorm:"type:int4;default:0" json:"balance_discount_cash"`
	IsLock              bool       `gorm:"type:boolean;default:true;not null" json:"is_lock"`
	LockBy              string     `gorm:"type:varchar(45)" json:"lock_by"`
	LockDate            time.Time  `gorm:"type:timestamp;default:now();not null" json:"lock_date"`
	LockLocation        string     `gorm:"type:varchar(20)" json:"lock_location"`
	LockReason          string     `gorm:"type:varchar(500)" json:"lock_reason"`
	UnlockBy            string     `gorm:"type:varchar(45)" json:"unlock_by"`
	UnlockDate          *time.Time `gorm:"type:timestamp" json:"unlock_date,omitempty"`
	UnlockLocation      string     `gorm:"type:varchar(20)" json:"unlock_location"`
	UnlockReason        string     `gorm:"type:varchar(500)" json:"unlock_reason"`
	RefCardEntityId     *uuid.UUID `gorm:"type:uuid" json:"ref_card_entity_id"`
}

func (LogLockCard) TableName() string {
	return "log_lock_card"
}

type SearchLockCardParams struct {
	MemberTel string
	CardNo    string
	IsLocked  *bool
	Page      int
	Skip      int
}

type CardMemberListData struct {
	CardEntityId  uuid.UUID `json:"card_entity_id"`
	MobileNo      string    `json:"mobile_no"`
	CardNo        string    `json:"card_no"`
	CardType      string    `json:"card_type"`
	BalanceECoin  int       `json:"balance_e_coin"`
	BalanceEBonus int       `json:"balance_e_bonus"`
	// DiscountCash    int        `json:"balance_discount_cash"`
	LockCardId          *uuid.UUID `json:"lock_card_id,omitempty"`
	IsLock              bool       `json:"is_lock"`
	LockBy              string     `json:"lock_by"`
	LockDate            *time.Time `json:"lock_date"`
	LockLocation        string     `json:"lock_location"`
	LockReason          string     `json:"lock_reason"`
	UnlockBy            string     `json:"unlock_by"`
	UnlockDate          *time.Time `json:"unlock_date"`
	UnlockLocation      string     `json:"unlock_location"`
	UnlockReason        string     `json:"unlock_reason"`
	RefCardEntityId     *uuid.UUID `json:"ref_card_entity_id"`
	RefCardEntityActive bool       `json:"ref_card_entity_active"`
}
