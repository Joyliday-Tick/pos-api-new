package v2

import (
	"fmt"
	"new-pos-api/config"
	model_v2 "new-pos-api/models/v2"
)

func CreateMeterRecord(dto model_v2.MeterRecordDto) error {
	meterRecord := model_v2.MeterRecord{
		RcDate:          dto.RcDate,
		RcAssetID:       &dto.RcAssetID,
		RcAssetHead:     &dto.RcAssetHead,
		RcAssetLocation: &dto.RcAssetLocation,
		RcMember:        &dto.RcMember,
		RcCardID:        &dto.RcCardID,
		RcEcoin:         &dto.RcEcoin,
		RcBonus:         &dto.RcBonus,
		RnEcoin:         &dto.RnEcoin,
		RnBonus:         &dto.RnBonus,
		CardType:        &dto.CardType,
		RcTime:          &dto.RcTime,
		RnTime:          &dto.RnTime,
	}

	if err := config.DB_JREADER.Create(&meterRecord).Error; err != nil {
		return err
	}

	fmt.Printf("Created MeterRecord: %+v\n", meterRecord)

	return nil
}
