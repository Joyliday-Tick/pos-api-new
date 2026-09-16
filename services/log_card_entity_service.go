package services
import (
	"fmt"
	"new-pos-api/config"
	"new-pos-api/models"
)

func CreateLogUpdateCardEntity(input models.LogUpdateCardEntityDto) (models.LogUpdateCardEntity, error) {
	var entity models.LogUpdateCardEntity

	// Map fields from DTO to Entity
	entity.CardNo = input.CardNo
	entity.OldMobile = input.OldMobile
	entity.NewMobile = input.NewMobile
	entity.Location = input.Location
	entity.UpdatedDate = input.UpdatedDate

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create log update card entity: %w", err)
	}

	return entity, nil
}
