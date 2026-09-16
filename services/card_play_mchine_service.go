package services
import (
	"fmt"
	"new-pos-api/config"
	"new-pos-api/models"
	"time"

	"github.com/google/uuid"
)
func CreateCardPlayMachine(cardPlayId uuid.UUID, machineId int32, userId int32, playTime *int32, eCoin *int32, eBonus *int32) (*models.CardPlayMachine, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	cardPlayMachine := &models.CardPlayMachine{
		ID:         uuid.New(),
		CardPlayID: cardPlayId,
		MachineID:  &machineId,
		CreateBy:   &userId,
		PlayTime:   playTime,
		ECoin:      eCoin,
		EBonus:     eBonus,
		CreateDate: ptrTime(time.Now()),
		IsDelete:   false,
	}

	if err := config.DB_POS.Create(cardPlayMachine).Error; err != nil {
		return nil, fmt.Errorf("failed to create card play machine: %w", err)
	}

	return cardPlayMachine, nil
}

func GetMachineListByCardPlayId(cardPlayId *uuid.UUID) ([]models.MachineList, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var machineList []models.MachineList
	if err := config.DB_POS.Model(&models.CardPlayMachine{}).
		Select("machine_id AS machine").
		Where("card_play_id = ? AND is_delete = false", cardPlayId).
		Scan(&machineList).Error; err != nil {
		return nil, fmt.Errorf("failed to get machine list: %w", err)
	}

	return machineList, nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

// DeductCardPlayMachine updates ecoin, ebonus, play_time on the card_play_machine row
// matched by card_play_id + machine_id. Only fields with non-nil pointers are updated.
// newEcoin, newEbonus, newPlayTime are the already-calculated remaining values.
func DeductCardPlayMachine(cardPlayId uuid.UUID, machineId int32, newEcoin *int32, newEbonus *int32, newPlayTime *int32) error {
	if config.DB_POS == nil {
		return fmt.Errorf("database connection is nil")
	}

	updates := map[string]interface{}{
		"update_date": ptrTime(time.Now()),
	}
	if newEcoin != nil {
		updates["ecoin"] = *newEcoin
	}
	if newEbonus != nil {
		updates["ebonus"] = *newEbonus
	}
	if newPlayTime != nil {
		updates["play_time"] = *newPlayTime
	}

	if len(updates) == 1 { // only update_date — nothing to do
		return nil
	}

	result := config.DB_POS.Model(&CardPlayMachineModel{}).
		Where("card_play_id = ? AND machine_id = ? AND is_delete = false", cardPlayId, machineId).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to deduct card play machine: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("card play machine not found for card_play_id=%s machine_id=%d", cardPlayId, machineId)
	}

	return nil
}

// CardPlayMachineModel is an alias used by GORM to resolve the table name inside this package.
type CardPlayMachineModel = models.CardPlayMachine