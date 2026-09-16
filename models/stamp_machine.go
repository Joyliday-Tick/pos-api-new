package models

import (
	"time"
)

type StampMachineDto struct {
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Location *string `json:"location"`
	IsActive bool    `json:"is_active"`
}

type StampMachine struct {
	Code       string     `gorm:"column:code;primaryKey" json:"code"`
	Name       string     `gorm:"column:name" json:"name"`
	Location   *string    `gorm:"column:location" json:"location"`
	IsActive   bool       `gorm:"column:is_active;default:true" json:"is_active"`
	CreateBy   *int       `gorm:"column:create_by" json:"create_by,omitempty"`
	CreateDate time.Time  `gorm:"column:create_date;autoCreateTime" json:"create_date"`
	UpdateBy   *int       `gorm:"column:update_by" json:"update_by,omitempty"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"update_date,omitempty"`
	IsDelete   bool       `gorm:"column:is_delete;default:false" json:"is_delete"`
	DeleteBy   *int       `gorm:"column:delete_by" json:"delete_by,omitempty"`
	DeleteDate *time.Time `gorm:"column:delete_date" json:"delete_date,omitempty"`
}

func (StampMachine) TableName() string {
	return "stamp_machine"
}

type StampHouseDataList struct {
	MemberTel   string    `json:"member_tel"`
	StampCount  int       `json:"stamp_count"`
	EStamp      int       `json:"e_stamp"`
	Date        time.Time `json:"date"`
	MachineCode string    `json:"machine_code"`
	MachineName string    `json:"machine_name"`
	Location    string    `json:"location"`
	Firstname   string    `json:"firstname"`
	Lastname    string    `json:"lastname"`
	MemberID    int       `json:"member_id"`
}
type StampHouseReport struct {
	Data        []StampHouseDataList `json:"data"`
	TotalEStamp int64                `json:"total_estamp"`
}
