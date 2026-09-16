package models

import (
	"time"
)

type LogUpdateCardEntity struct {
	UpdateID    int       `json:"update_id" gorm:"primaryKey"`
	CardNo      string    `json:"card_no"`
	OldMobile   string    `json:"old_mobile"`
	NewMobile   string    `json:"new_mobile"`
	Location    string    `json:"location"`
	UpdatedDate time.Time `json:"updated_date" gorm:"default:CURRENT_TIMESTAMP"`
}

type LogUpdateCardEntityDto struct {
	CardNo      string    `json:"card_no"`
	OldMobile   string    `json:"old_mobile"`
	NewMobile   string    `json:"new_mobile"`
	Location    string    `json:"location"`
	UpdatedDate time.Time `json:"updated_date"`
}

func (LogUpdateCardEntity) TableName() string {
	return "log_update_card_entity"
}
