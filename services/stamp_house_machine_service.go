package services

import (
	"fmt"
	"new-pos-api/config"
	"new-pos-api/models"
)

func GetActiveStampMachine() ([]models.StampMachine, error) {
	var stamp_machines []models.StampMachine

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// ดึง stamp machine ที่ Active เท่านั้น
	if err := config.DB_POS.Where("is_active = ?", true).Find(&stamp_machines).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active stamp machine: %w", err)
	}

	return stamp_machines, nil
}
