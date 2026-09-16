package services

import (
	"fmt"

	"new-pos-api/config"
	"new-pos-api/models"

	"new-pos-api/utils"
)

func CreateCardWithdraw(input models.CardWithdrawDto, userId int) (models.CardWithdraw, error) {
	var entity models.CardWithdraw

	// Map fields จาก DTO ไปยัง Entity
	entity.FromChannel = input.FromChannel
	entity.CardNo = input.CardNo
	entity.MemberTel = input.MemberTel
	entity.AmountEcoin = input.AmountEcoin
	entity.AmountEbonus = input.AmountEbonus
	entity.CardDepositId = input.CardDepositId
	entity.DateCreated = *utils.TimeNowAsia()
	entity.AmountDiscountCash = input.AmountDiscountCash

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create card withdraw: %w", err)
	}

	return entity, nil
}
