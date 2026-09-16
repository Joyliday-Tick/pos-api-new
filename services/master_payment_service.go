package services

import (
	"fmt"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
)

func GetActiveMasterPayment() ([]models.MasterPayment, error) {
	var payments []models.MasterPayment

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// ใช้ GORM แบบง่ายและปลอดภัย
	if err := config.DB_POS.Where("is_active = ?", true).Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active master payment: %w", err)
	}

	return payments, nil
}
