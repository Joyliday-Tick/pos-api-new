package models

import (
	"time"

	"github.com/google/uuid"
)

type MasterPaymentDto struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

type MasterPayment struct {
	ID         uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name       string     `gorm:"type:varchar(100);not null" json:"name"`
	IsActive   bool       `gorm:"default:true;not null" json:"is_active"`
	CreateBy   int        `gorm:"column:create_by" json:"create_by"`
	CreateDate time.Time  `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateBy   *int       `gorm:"column:update_by" json:"update_by"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"update_date"`
	IsDelete   bool       `gorm:"default:false;not null" json:"is_delete"`
	DeleteBy   *int       `gorm:"column:delete_by" json:"delete_by"`
	DeleteDate *time.Time `gorm:"column:delete_date" json:"delete_date"`
}

func (MasterPayment) TableName() string {
	return "master_payment"
}
