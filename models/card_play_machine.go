package models

import (
	"time"

	"github.com/google/uuid"
)

type CardPlayMachine struct {
	ID         uuid.UUID  `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CardPlayID uuid.UUID     `gorm:"column:card_play_id" json:"card_play_id"`
	MachineID  *int32     `gorm:"column:machine_id" json:"machine_id"`
	PlayTime   *int32     `gorm:"column:play_time;default:-1" json:"play_time"`
	ECoin      *int32     `gorm:"column:ecoin;default:0" json:"ecoin"`
	EBonus     *int32     `gorm:"column:ebonus;default:0" json:"ebonus"`
	CreateBy   *int32     `gorm:"column:create_by" json:"create_by"`
	CreateDate *time.Time `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateBy   *int32     `gorm:"column:update_by" json:"update_by"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"update_date"`
	IsDelete   bool       `gorm:"column:is_delete" json:"is_delete"`
	DeleteBy   *int32     `gorm:"column:delete_by" json:"delete_by"`
	DeleteDate *time.Time `gorm:"column:delete_date" json:"delete_date"`
}

type CardPlayMachineDto struct {
	ID         uuid.UUID  `json:"id"`
	CardPlayID uuid.UUID     `json:"card_play_id"`
	MachineID  *int32     `json:"machine_id"`
	PlayTime   *int32     `json:"play_time"`
	ECoin      *int32     `json:"ecoin"`
	EBonus     *int32     `json:"ebonus"`
	CreateBy   *int32     `json:"create_by"`
	CreateDate *time.Time `json:"create_date"`
	UpdateBy   *int32     `json:"update_by"`
	UpdateDate *time.Time `json:"update_date"`
	IsDelete   bool       `json:"is_delete"`
	DeleteBy   *int32     `json:"delete_by"`
	DeleteDate *time.Time `json:"delete_date"`
}

func (CardPlayMachine) TableName() string {
	return "card_play_machine"
}
