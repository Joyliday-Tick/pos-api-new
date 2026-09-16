package services

import (
	"fmt"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"
)

func CreateRefundConfirm(input models.RefundComfirmDto) (models.RefundComfirm, error) {
	var entity models.RefundComfirm

	entity.RefconDate = utils.TimeNowAsia()
	entity.RefconUser = input.RefconUser
	entity.RefconLocation = input.RefconLocation
	entity.ReqID = input.ReqID
	entity.RefundBy = input.RefundBy
	entity.MemberTel = input.RefMemberTel

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create refund request: %w", err)
	}

	return entity, nil
}
