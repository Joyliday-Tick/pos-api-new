package services

import (
	"fmt"

	// "log"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"

	"new-pos-api/utils"

	"github.com/google/uuid"
)

func CreateLogLockCard(input models.LockCardDto) (models.LogLockCard, error) {
	var entity models.LogLockCard

	entity.CardNo = input.CardNo
	entity.MemberTel = input.MemberTel
	entity.BalanceEcoin = input.BalanceEcoin
	entity.BalanceEbonus = input.BalanceEbonus
	entity.BalanceDiscountCash = input.BalanceDiscountCash
	// entity.IsLock = input.IsLock
	entity.LockBy = input.LockBy
	entity.LockDate = input.LockDate
	entity.LockLocation = input.LockLocation
	entity.LockReason = input.LockReason
	refId := uuid.MustParse(input.RefCardEntityId)
	entity.RefCardEntityId = &refId

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create log lock card: %w", err)
	}

	return entity, nil
}

func UnLockCardEntity(input models.UnLockCardDto) (models.LogLockCard, error) {
	var existing models.LogLockCard

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", input.LockId).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("log lock card not found: %w", err)
	}

	updateData := map[string]interface{}{
		"is_lock":         false,
		"unlock_by":       input.UnlockBy,
		"unlock_date":     utils.TimeNowAsia(),
		"unlock_location": input.UnlockLocation,
		"unlock_reason":   input.UnlockReason,
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.LogLockCard{}).Where("id = ?", input.LockId).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update log lock card: %w", err)
	}

	// ดึงข้อมูลใหม่หลังอัปเดต
	if err := config.DB_POS.Where("id = ?", input.LockId).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated log lock card: %w", err)
	}

	return existing, nil
}
