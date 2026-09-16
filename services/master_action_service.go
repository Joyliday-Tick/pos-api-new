package services

import (
	"fmt"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
)

func GetActiveMasterAction() ([]models.MasterAction, error) {
	var actions []models.MasterAction

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if err := config.DB_POS.Where("is_active = ?", true).Find(&actions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active master action: %w", err)
	}

	return actions, nil
}
