package models

import (
	"time"
)

type BonusSetting struct {
	ID            int       `gorm:"primaryKey;column:id"`
	ReturnPersent float64   `gorm:"column:return_persent"`
	PricePoint    float64   `gorm:"column:price_point"`
	SetDate       time.Time `gorm:"column:setdate"`
	MinPrize      float64   `gorm:"column:min_prize;default:0"`
	MaxPrize      float64   `gorm:"column:max_prize;default:0.4"`
	MinJc         float64   `gorm:"column:min_jc;default:0"`
	MaxJc         float64   `gorm:"column:max_jc;default:0.4"`
	FreePointEvr  int       `gorm:"column:freepoint_evr;default:10"`
	FreePointX    int       `gorm:"column:freepoint_x;default:1"`
}

// TableName overrides the table name used by GORM
func (BonusSetting) TableName() string {
	return "bonus_setting"
}
