package models

import "github.com/google/uuid"

type MachineSubGroup struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	MachineID   *int      `gorm:"column:machine_id" json:"machine_id"`
	GroupID     *int      `gorm:"column:group_id" json:"group_id"`
	PlayTime    *int      `gorm:"column:play_time;default:0" json:"play_time"`
	EBonus      *int      `gorm:"column:e_bonus;default:0" json:"e_bonus"`
	ECoin       *int      `gorm:"column:e_coin;default:0" json:"e_coin"`
	MachineName *string   `gorm:"column:machine_name;size:50" json:"machine_name"`
	Category    *string   `gorm:"column:category;size:30" json:"category"`
}

func (MachineSubGroup) TableName() string {
	return "machine_sub_group"
}

type MachineSubGroupSum struct {
	ECoin    int `json:"e_coin"`
	EBonus   int `json:"e_bonus"`
	PlayTime int `json:"play_time"`
}
