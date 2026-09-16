package models

import (
	"time"
)

type BranchDto struct {
	BranchCode string  `json:"branch_code"`
	BranchName string  `json:"branch_name"`
	BMain      *string `json:"b_main"`
	IsActive   bool    `json:"is_active"`
}

type Branch struct {
	BranchCode   string     `gorm:"column:branch_code;primaryKey" json:"branch_code"`
	BranchName   string     `gorm:"column:branch_name" json:"branch_name"`
	BMain        *string    `gorm:"column:b_main" json:"b_main"`
	BArea        int        `gorm:"column:b_area" json:"b_area"`
	PrizeZone    int        `gorm:"column:prize_zone" json:"prize_zone"`
	ActionZone   int        `gorm:"column:action_zone" json:"action_zone"`
	KidsZone     int        `gorm:"column:kids_zone" json:"kids_zone"`
	PlayportZone int        `gorm:"column:playport_zone" json:"playport_zone"`
	IsActive     bool       `gorm:"column:is_active;default:true" json:"is_active"`
	CreateBy     int        `gorm:"column:create_by" json:"create_by"`
	CreateDate   *time.Time `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateBy     *int       `gorm:"column:update_by" json:"update_by"`
	UpdateDate   *time.Time `gorm:"column:update_date" json:"update_date"`
	IsDelete     bool       `gorm:"column:is_delete;default:false" json:"is_delete"`
	DeleteBy     *int       `gorm:"column:delete_by" json:"delete_by"`
	DeleteDate   *time.Time `gorm:"column:delete_date" json:"delete_date"`
}

func (Branch) TableName() string {
	return "branch"
}
