package services

import (
	"fmt"
	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"

	"github.com/google/uuid"
)

func CreateCardPlayType(input models.CardPlayTypeDto, userId int) (models.CardPlayType, error) {
	var entity models.CardPlayType

	// Map fields from DTO to Entity
	entity.CardPlayId = input.CardPlayId
	entity.CreateBy = userId
	entity.CreateDate = *utils.TimeNowAsia()

	if input.StartDate != nil {
		entity.StartDate = input.StartDate
	}
	if input.EndDate != nil {
		entity.EndDate = input.EndDate
	}
	if *input.PlayTime != 0 {
		entity.PlayTime = input.PlayTime
	}

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create card play type: %w", err)
	}

	return entity, nil
}

func GetCardPlayTypeByCardPlayId(cardPlayId uuid.UUID) (models.CardPlayType, error) {
	var entity models.CardPlayType
	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}
	if err := config.DB_POS.Where("is_delete = false and card_play_id = ?", cardPlayId).First(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to get card play type: %w", err)
	}
	return entity, nil
}

func UpdateCardPlayTypeUsed(cardPlayId *uuid.UUID, cardNo string) (*models.CardPlayType, error) {
	if config.DB_POS == nil {
	return nil, fmt.Errorf("database pos connection is nil")
	}
	// update used_time = used_time + 1 where card_play_id = cardPlayId and used_time < play_time
	result := config.DB_POS.Exec("update card_play_type set used_time = used_time + 1 where card_play_id = ? and used_time < play_time", cardPlayId)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to update card play type used: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("no rows affected, either card play type not found or used_time >= play_time")
	}
	var entity models.CardPlayType
	if err := config.DB_POS.Where("is_delete = false and card_play_id = ?", cardPlayId).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated card play type: %w", err)
	}
	return &entity, nil
}

