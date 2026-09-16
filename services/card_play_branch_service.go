package services

import (
	"fmt"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
)

func CreateBatchCardPlayBranch(entities []models.CardPlayBranch, userId int) ([]models.CardPlayBranch, error) {

	if config.DB_POS == nil {
		return entities, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entities).Error; err != nil {
		return entities, fmt.Errorf("failed to create batch card play branch: %w", err)

	}

	return entities, nil
}
