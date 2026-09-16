package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PosMenuDto struct {
	Code                   string     `json:"code" validate:"required"`
	Description            string     `json:"description" validate:"required"`
	Price                  int        `json:"price" validate:"required"`
	ECoin                  int        `json:"e_coin"`
	EBonus                 int        `json:"e_bonus"`
	BranchGroupID          *int       `json:"branch_group_id" validate:"required"`
	MachineGroupID         *int       `json:"machine_group_id"`
	IsActive               bool       `json:"is_active" validate:"required"`
	BonusExpireDate        *time.Time `json:"bonus_expire_date"`
	LimitTime              *int       `json:"limit_time"`
	CardTypeID             *uuid.UUID `json:"card_type_id" validate:"required"`
	BranchList             *[]string  `json:"branch_list"`
	StartDate              *time.Time `json:"start_date"`
	EndDate                *time.Time `json:"end_date"`
	CardExpireDate         *time.Time `json:"card_expire_date"`
	GroupMenuID            *uuid.UUID `json:"group_menu_id" validate:"required"`
	DiscountCash           int        `json:"discount_cash"`
	DiscountCashExpireDate *time.Time `json:"discount_cash_expire_date"`
	DiscountCashlimitDays  *int       `json:"discount_cash_limit_days"`
	BonusExpireLimitDays   *int       `json:"bonus_expire_limit_days"`
	CardExpireLimitMinutes *int       `json:"card_expire_limit_minutes"`
}

type PosMenuDataDto struct {
	MenuID                 int             `json:"menu_id"`
	Code                   string          `json:"code"`
	MenuName               string          `json:"menu_name"`
	Price                  int             `json:"price"`
	ECoin                  int             `json:"e_coin"`
	EBonus                 int             `json:"e_bonus"`
	Token                  *int            `json:"token"`
	BranchGroupID          *int            `json:"branch_group_id"`
	MachineGroupID         *int            `json:"machine_group_id"`
	IsActive               bool            `json:"is_active"`
	BonusExpireDate        *time.Time      `json:"bonus_expire_date"`
	LimitTime              *int            `json:"limit_time"`
	CardTypeID             *uuid.UUID      `json:"card_type_id"`
	CardTypeName           string          `json:"card_type_name"`
	BranchList             json.RawMessage `json:"branch_list"`
	StartDate              *time.Time      `json:"start_date"`
	EndDate                *time.Time      `json:"end_date"`
	CardExpireDate         *time.Time      `json:"card_expire_date"`
	GroupMenuID            *uuid.UUID      `json:"group_menu_id"`
	GroupMenuName          string          `json:"group_menu_name"`
	CreateDate             *time.Time      `json:"create_date"`
	GroupBranchName        string          `json:"group_branch_name"`
	LocationCode           string          `json:"location_code"`
	EnableMember           bool            `json:"enable_member"`
	ReceiptStub            bool            `json:"receipt_stub"`
	DisplayName            string          `json:"menu_display_name"`
	DiscountCash           int             `json:"discount_cash"`
	DiscountCashExpireDate *time.Time      `json:"discount_cash_expire_date"`
	DiscountCashLimitDays  *int            `json:"discount_cash_limit_days"`
	BonusExpireLimitDays   *int            `json:"bonus_expire_limit_days"`
	CardExpireLimitMinutes *int            `json:"card_expire_limit_minutes"`
}

type PosMenu struct {
	ID                     int             `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code                   string          `gorm:"column:code" json:"code"`
	Description            string          `gorm:"column:description" json:"description"`
	Price                  int             `gorm:"column:price" json:"price"`
	ECoin                  int             `gorm:"column:e_coin" json:"e_coin"`
	EBonus                 int             `gorm:"column:e_bonus" json:"e_bonus"`
	Token                  *int            `gorm:"column:token" json:"token"`
	BranchGroupID          *int            `gorm:"column:branch_group_id" json:"branch_group_id"`
	MachineGroupID         *int            `gorm:"column:machine_group_id" json:"machine_group_id"`
	IsActive               bool            `gorm:"column:is_active;default:true" json:"is_active"`
	BonusExpireDate        *time.Time      `gorm:"column:bonus_expire_date" json:"bonus_expire_date"`
	LimitTime              *int            `gorm:"column:limit_time;default:0" json:"limit_time"`
	CardTypeID             *uuid.UUID      `gorm:"column:card_type_id" json:"card_type_id"`
	BranchList             json.RawMessage `gorm:"column:branch_list" json:"branch_list"`
	StartDate              *time.Time      `gorm:"column:start_date" json:"start_date"`
	EndDate                *time.Time      `gorm:"column:end_date" json:"end_date"`
	CardExpireDate         *time.Time      `gorm:"column:card_expire_date" json:"card_expire_date"`
	GroupMenuID            *uuid.UUID      `gorm:"column:group_menu_id" json:"group_menu_id"`
	CreateBy               *int            `gorm:"column:create_by" json:"create_by"`
	CreateDate             *time.Time      `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateBy               *int            `gorm:"column:update_by" json:"update_by"`
	UpdateDate             *time.Time      `gorm:"column:update_date" json:"update_date"`
	IsDelete               bool            `gorm:"column:is_delete;default:false" json:"is_delete"`
	DeleteBy               *int            `gorm:"column:delete_by" json:"delete_by"`
	DeleteDate             *time.Time      `gorm:"column:delete_date" json:"delete_date"`
	DiscountCash           int             `gorm:"column:discount_cash" json:"discount_cash"`
	DiscountCashExpireDate *time.Time      `gorm:"column:discount_cash_expire_date" json:"discount_cash_expire_date"`
	DiscountCashLimitDays  *int            `gorm:"column:discount_cash_limit_days" json:"discount_cash_limit_days"`
	BonusExpireLimitDays   *int            `gorm:"column:bonus_expire_limit_days" json:"bonus_expire_limit_days"`
	CardExpireLimitMinutes *int            `gorm:"column:card_expire_limit_minutes" json:"card_expire_limit_minutes"`
}

func (PosMenu) TableName() string {
	return "pos_menu"
}

type SearchPosMenuParams struct {
	Search         string
	Location       string
	GroupMenuID    string
	BranchGroupID  *int
	MachineGroupID *int
	CardTypeID     string
	IsActive       *bool
	Page           int
	Skip           int
}

type SearchPosMenuSaleParams struct {
	Location    string
	GroupMenuID string
	Search      string
	Page        int
	Skip        int
}
