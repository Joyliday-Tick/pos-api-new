package models

import (
	"time"

	"github.com/google/uuid"
)

type CustomerTierImportDTO struct {
	Tel  string
	Tier string
}

type CustomerTier struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Tel        string    `gorm:"type:varchar(10);not null;index" json:"tel"`
	Tier       string    `gorm:"type:varchar(50);not null" json:"tier"`
	CreateDate time.Time `gorm:"type:timestamp;default:now()" json:"create_date"`
}

func (CustomerTier) TableName() string {
	return "customer_tier"
}
