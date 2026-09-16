package models

import (
	"time"
)

type MachineGroupDto struct {
	GroupName        string                `json:"group_name"`
	IsActive         bool                  `json:"is_active"`
	MachineGroupType *string               `json:"machine_group_type"`
	MachineList      []MachineSubGroupData `json:"machine_list_json"`
}

type MachineGroupList struct {
	ID               int        `json:"id"`
	GroupName        string     `json:"group_name"`
	IsActive         bool       `json:"is_active"`
	CreateBy         int        `json:"create_by"`
	CreateDate       time.Time  `json:"create_date"`
	UpdateBy         *int       `json:"update_by"`
	UpdateDate       *time.Time `json:"update_date"`
	IsDelete         bool       `json:"is_delete"`
	DeleteBy         *int       `json:"delete_by"`
	DeleteDate       *time.Time `json:"delete_date"`
	MachineGroupType string     `json:"machine_group_type"`
	SubMachineId     []string   `json:"sub_machine_id"`
}

type MachineGroupListV1 struct {
	ID               int               `json:"id"`
	GroupName        string            `json:"group_name"`
	IsActive         bool              `json:"is_active"`
	CreateBy         int               `json:"create_by"`
	CreateDate       time.Time         `json:"create_date"`
	UpdateBy         *int              `json:"update_by"`
	UpdateDate       *time.Time        `json:"update_date"`
	IsDelete         bool              `json:"is_delete"`
	DeleteBy         *int              `json:"delete_by"`
	DeleteDate       *time.Time        `json:"delete_date"`
	MachineGroupType string            `json:"machine_group_type"`
	SubMachine       []MachineSubGroup `json:"sub_machine"`
}

type MachineSubGroupData struct {
	MachineID   *int    `gorm:"column:machine_id" json:"machine_id"`
	GroupID     *int    `gorm:"column:group_id" json:"group_id"`
	PlayTime    *int    `gorm:"column:play_time;default:0" json:"play_time"`
	EBonus      *int    `gorm:"column:e_bonus;default:0" json:"e_bonus"`
	ECoin       *int    `gorm:"column:e_coin;default:0" json:"e_coin"`
	MachineName *string `gorm:"column:machine_name;size:50" json:"machine_name"`
	Category    *string `gorm:"column:category;size:30" json:"category"`
}

type MachineGroup struct {
	ID               int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupName        string     `gorm:"column:group_name" json:"group_name"`
	IsActive         bool       `gorm:"column:is_active;default:true" json:"is_active"`
	CreateBy         int        `gorm:"column:create_by" json:"create_by"`
	CreateDate       time.Time  `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateBy         *int       `gorm:"column:update_by" json:"update_by"`
	UpdateDate       *time.Time `gorm:"column:update_date" json:"update_date"`
	IsDelete         bool       `gorm:"column:is_delete;default:false" json:"is_delete"`
	DeleteBy         *int       `gorm:"column:delete_by" json:"delete_by"`
	DeleteDate       *time.Time `gorm:"column:delete_date" json:"delete_date"`
	MachineGroupType string     `gorm:"column:machine_group_type" json:"machine_group_type"`
}

func (MachineGroup) TableName() string {
	return "machine_group"
}
