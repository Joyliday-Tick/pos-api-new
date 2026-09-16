package services

import (
	"fmt"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
)

func GetBonusSetting() (*models.BonusSetting, error) {
	var bonusSetting models.BonusSetting

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// แก้ตรงนี้: เช็ก .Error จากผลลัพธ์ของ Find()
	if err := config.DB_POS.Limit(1).Find(&bonusSetting).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch bonus setting: %w", err)
	}

	return &bonusSetting, nil
}
