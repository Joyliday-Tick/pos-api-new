package models

import (
	"time"

	"github.com/google/uuid"
)

type CardPlayType struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CardPlayId uuid.UUID  `json:"card_play_id"`
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	CreateBy   int        `json:"create_by"`
	CreateDate time.Time  `json:"create_date" gorm:"default:CURRENT_TIMESTAMP"`
	UpdateBy   *int       `json:"update_by"`
	UpdateDate *time.Time `json:"update_date"`
	IsDelete   bool       `json:"is_delete" gorm:"default:false"`
	DeleteBy   *int       `json:"delete_by"`
	DeleteDate *time.Time `json:"delete_date"`
	PlayTime   *int       `json:"play_time" gorm:"default:0"`
	UsedTime   *int       `json:"used_time" gorm:"default:0"`
}
type CardPlayTypeDto struct {
	CardPlayId uuid.UUID  `json:"card_play_id"`
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	PlayTime   *int       `json:"play_time" gorm:"default:0"`
}

func (CardPlayType) TableName() string {
	return "card_play_type"
}
