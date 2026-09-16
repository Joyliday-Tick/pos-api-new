package services

import (
	"fmt"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
)

func GetActiveLogo() ([]models.MasterLogo, error) {
	var logoes []models.MasterLogo

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if err := config.DB_POS.Where("is_active = ?", true).Find(&logoes).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active logo: %w", err)
	}

	return logoes, nil
}
