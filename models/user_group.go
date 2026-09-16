package models

import "time"

type UserGroupDto struct {
	GroupName  string   `json:"group_name"`
	IsActive   bool     `json:"is_active"`
	BranchList []string `json:"branch_list"`
}
type UserGroupList struct {
	ID               int        `json:"id"`
	GroupName        string     `json:"group_name"`
	IsActive         bool       `json:"is_active"`
	CreateBy         int        `json:"create_by"`
	CreateDate       *time.Time `json:"create_date"`
	UpdateBy         *int       `json:"update_by"`
	UpdateDate       *time.Time `json:"update_date"`
	IsDelete         bool       `json:"is_delete"`
	DeleteBy         *int       `json:"delete_by"`
	DeleteDate       *time.Time `json:"delete_date"`
	SubLocationCodes []string   `json:"sub_location_codes"`
}

type UserGroup struct {
	ID         int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupName  string     `gorm:"column:group_name" json:"group_name"`
	IsActive   bool       `gorm:"column:is_active" json:"is_active"`
	CreateBy   int        `gorm:"column:create_by" json:"create_by"`
	CreateDate *time.Time `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateBy   *int       `gorm:"column:update_by" json:"update_by"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"update_date"`
	IsDelete   bool       `gorm:"column:is_delete;default:false" json:"is_delete"`
	DeleteBy   *int       `gorm:"column:delete_by" json:"delete_by"`
	DeleteDate *time.Time `gorm:"column:delete_date" json:"delete_date"`
}

func (UserGroup) TableName() string {
	return "user_group"
}
