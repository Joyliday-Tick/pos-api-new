package models

import (
	"time"

	"github.com/google/uuid"
)

type MasterLogo struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name       string     `json:"name"`
	Logo       string     `json:"logo"`
	IsActive   bool       `json:"is_active" gorm:"default:true"` // default เป็น true (หรือ false)
	CreateBy   int        `json:"create_by"`
	CreateDate time.Time  `json:"create_date" gorm:"default:CURRENT_TIMESTAMP"` // default now()
	UpdateBy   *int       `json:"update_by"`                                    // nullable
	IsDelete   bool       `json:"is_delete" gorm:"default:false"`
	UpdateDate *time.Time `json:"update_date"` // nullable
	DeleteBy   *int       `json:"delete_by"`   // nullable
	DeleteDate *time.Time `json:"delete_date"` // nullable
}

func (MasterLogo) TableName() string {
	return "master_logo"
}
