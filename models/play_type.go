package models

import (
	"time"

	"github.com/google/uuid"
)

type PlayTypeDto struct {
	Name          string `json:"name"`
	IsActive      bool   `json:"is_active"`
	PromotionTime *int   `json:"promotion_time"`
}

type PlayType struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name          string     `json:"name"`
	PromotionTime *int       `json:"promotion_time" gorm:"default:0"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	CreateBy      int        `json:"create_by"`
	CreateDate    time.Time  `json:"create_date" gorm:"default:CURRENT_TIMESTAMP"`
	UpdateBy      *int       `json:"update_by"`
	UpdateDate    *time.Time `json:"update_date"`
	IsDelete      bool       `json:"is_delete" gorm:"default:false"`
	DeleteBy      *int       `json:"delete_by"`
	DeleteDate    *time.Time `json:"delete_date"`
}

func (PlayType) TableName() string {
	return "play_type"
}

type PlayTypeSearchParams struct {
	Search string
	Page   int
	Skip   int
}
