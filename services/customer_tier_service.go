package services

import (
	"fmt"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"

	"gorm.io/gorm"
)

func CreateBatchCustomerTierTx(items []models.CustomerTier, batchSize int) error {

	if len(items) == 0 {
		return nil
	}

	tx := config.DB_POS.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	for i := 0; i < len(items); i += batchSize {

		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]

		if err := tx.Create(&batch).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("batch %d failed: %w", i/batchSize, err)
		}
	}

	return tx.Commit().Error
}

func ImportCustomerTier(
	dtos []models.CustomerTierImportDTO) error {

	if config.DB_POS == nil {
		return nil
	}
	now := utils.TimeNowAsia()

	db := config.DB_POS
	// 1. fileMap
	fileMap := map[string]string{}
	for _, d := range dtos {
		fileMap[d.Tel] = d.Tier
	}

	// 2. load latest tier
	type latest struct {
		Tel  string
		Tier string
	}

	var rows []latest
	if err := db.Raw(`
		SELECT DISTINCT ON (tel) tel, tier
		FROM customer_tier
		ORDER BY tel, create_date DESC
	`).Scan(&rows).Error; err != nil {
		return err
	}

	dbMap := map[string]string{}
	for _, r := range rows {
		dbMap[r.Tel] = r.Tier
	}

	// 3. transaction
	return db.Transaction(func(tx *gorm.DB) error {

		var inserts []models.CustomerTier

		// downgrade
		for tel, oldTier := range dbMap {
			if _, exists := fileMap[tel]; exists {
				continue
			}

			newTier := downgradeTier(oldTier)
			if newTier == oldTier {
				continue
			}

			inserts = append(inserts, models.CustomerTier{
				Tel:        tel,
				Tier:       newTier,
				CreateDate: *now,
			})
		}

		// new / upgrade
		for tel, newTier := range fileMap {
			oldTier, exists := dbMap[tel]

			if !exists || oldTier != newTier {
				inserts = append(inserts, models.CustomerTier{
					Tel:        tel,
					Tier:       newTier,
					CreateDate: *now,
				})
			}
		}

		if len(inserts) > 0 {
			return tx.CreateInBatches(inserts, 500).Error
		}

		return nil
	})
}

func downgradeTier(tier string) string {
	switch tier {
	case "PLATINUM":
		return "GOLD"
	case "GOLD":
		return "WELCOME"
	default:
		return tier
	}
}

func FindExistCustomerTier(tel string) ([]models.CustomerTier, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var customerTiers []models.CustomerTier

	if err := config.DB_POS.
		Where("tel = ?", tel).
		Order("create_date DESC").
		Find(&customerTiers).Error; err != nil {
		return nil, err
	}

	return customerTiers, nil
}
