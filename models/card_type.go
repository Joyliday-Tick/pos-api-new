package models

import (
	"time"

	"github.com/google/uuid"
)

type CardTypeDto struct {
	Name        string `json:"name"`
	IsActive    bool   `json:"is_active"`
	ShowBalance string `json:"show_balance"`
	// CreateBy int `json:"create_by"`

}

type CardType struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string     `json:"name"`
	IsActive    bool       `json:"is_active" gorm:"default:true"` // default เป็น true (หรือ false)
	CreateBy    int        `json:"create_by"`
	CreateDate  time.Time  `json:"create_date" gorm:"default:CURRENT_TIMESTAMP"` // default now()
	UpdateBy    *int       `json:"update_by"`                                    // nullable
	IsDelete    bool       `json:"is_delete" gorm:"default:false"`
	UpdateDate  *time.Time `json:"update_date"` // nullable
	DeleteBy    *int       `json:"delete_by"`   // nullable
	DeleteDate  *time.Time `json:"delete_date"` // nullable
	ShowBalance string     `json:"show_balance"`
}

func (CardType) TableName() string {
	return "card_type"
}

type SearchParams struct {
	Search string
	Page   int
	Skip   int
}
