package models

import (
	"encoding/json"
	"time"
)

type UserRoleDto struct {
	RoleDescription string   `json:"role_description"`
	ListMenu        []string `json:"list_menu"`
	IsActive        bool     `json:"is_active"`
}

type UserRole struct {
	ID              int             `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoleDescription string          `gorm:"column:role_description" json:"role_description"`
	ListMenu        json.RawMessage `gorm:"column:list_menu" json:"list_menu"`
	IsActive        bool            `gorm:"column:is_active" json:"is_active"`
	CreateBy        int             `gorm:"column:create_by" json:"create_by"`
	CreateDate      *time.Time      `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateBy        *int            `gorm:"column:update_by" json:"update_by"`
	UpdateDate      *time.Time      `gorm:"column:update_date" json:"update_date"`
	IsDelete        bool            `gorm:"column:is_delete;default:false" json:"is_delete"`
	DeleteBy        *int            `gorm:"column:delete_by" json:"delete_by"`
	DeleteDate      *time.Time      `gorm:"column:delete_date" json:"delete_date"`
}

func (UserRole) TableName() string {
	return "user_role"
}
