package models

import (
	"time"

	"github.com/google/uuid"
)

type GroupMenuDto struct {
	Name         string `json:"name"`
	IsActive     bool   `json:"is_active"`
	EnableMember bool   `json:"enable_member"`
	DisplayName  string `json:"display_name"`
	ReceiptStub  bool   `json:"receipt_stub"`
	Sorting      int    `json:"sorting"`
}

type GroupMenu struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name         string     `json:"name"`
	DisplayName  string     `json:"display_name"`
	IsActive     bool       `json:"is_active"`
	EnableMember bool       `json:"enable_member"`
	ReceiptStub  bool       `json:"receipt_stub"`
	Sorting      int        `json:"sorting"`
	CreateBy     int        `json:"create_by"`
	CreateDate   time.Time  `json:"create_date" gorm:"default:CURRENT_TIMESTAMP"` // default now()
	UpdateBy     *int       `json:"update_by"`                                    // nullable
	IsDelete     bool       `json:"is_delete" gorm:"default:false"`
	UpdateDate   *time.Time `json:"update_date"` // nullable
	DeleteBy     *int       `json:"delete_by"`   // nullable
	DeleteDate   *time.Time `json:"delete_date"` // nullable
}

func (GroupMenu) TableName() string {
	return "master_group_menu"
}
