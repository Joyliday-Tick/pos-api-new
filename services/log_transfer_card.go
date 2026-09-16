package services

import (
	"fmt"
	"new-pos-api/config"
	"new-pos-api/models"
)

func CreateLogTransferCard(input models.LogTransferCardDto) (models.LogTransferCard, error) {
	var entity models.LogTransferCard

	// Map fields from DTO to Entity
	entity.OldCardNo = input.OldCardNo
	entity.MemberTel = input.MemberTel
	entity.BalanceEcoin = input.BalanceEcoin
	entity.BalanceEbonus = input.BalanceEbonus
	entity.TransferCardNo = input.TransferCardNo
	entity.CreatedBy = input.CreatedBy
	entity.Location = input.Location

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create log transfer card: %w", err)
	}

	return entity, nil
}
