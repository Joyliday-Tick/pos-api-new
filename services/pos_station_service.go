package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	// "strings"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
	// "gorm.io/gorm"
	// "new-pos-api/utils"
	// "gorm.io/gorm"
)

func GetActiveStation() ([]models.PosStation, error) {
	var stations []models.PosStation

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// ดึง branch ที่ Active เท่านั้น
	if err := config.DB_POS.Where("is_active = ?", true).Find(&stations).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active station: %w", err)
	}

	return stations, nil
}

func CreateStation(input models.PosStationDto, userId int) (models.PosStation, error) {
	var station models.PosStation

	station.PosCode = input.PosCode
	station.PosDescription = input.PosDescription
	station.PosLocation = input.PosLocation
	station.BillHeaderLine1 = input.BillHeaderLine1
	station.BillHeaderLine2 = input.BillHeaderLine2
	station.BillHeaderLine3 = input.BillHeaderLine3
	station.BillHeaderLine4 = input.BillHeaderLine4
	station.BillHeaderLine5 = input.BillHeaderLine5
	station.IsActive = input.IsActive
	station.CreateBy = userId
	station.CreateDate = utils.TimeNowAsia()

	if input.LogoId != "" {
		logoId := uuid.MustParse(input.LogoId)
		station.LogoId = &logoId
	} else {
		station.LogoId = nil
	}

	if len(input.MacAddress) > 0 {
		macAddressJSON, err := json.Marshal(input.MacAddress)
		if err != nil {
			return station, fmt.Errorf("failed to marshal mac address: %w", err)
		}
		station.MacAddress = macAddressJSON
	} else {
		station.MacAddress = nil
	}
	if config.DB_POS == nil {
		return station, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&station).Error; err != nil {
		return station, fmt.Errorf("failed to create station: %w", err)
	}

	return station, nil
}

func UpdateStation(id int, input models.PosStationDto, userId int) (models.PosStation, error) {
	var existing models.PosStation

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("station not found: %w", err)
	}

	var macAddress any = nil
	if len(input.MacAddress) > 0 {
		macAddressJSON, err := json.Marshal(input.MacAddress)
		if err != nil {
			return existing, fmt.Errorf("failed to marshal list menu: %w", err)
		}
		macAddress = macAddressJSON
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"pos_code":          input.PosCode,
		"pos_description":   input.PosDescription,
		"pos_location":      input.PosLocation,
		"mac_address":       macAddress,
		"bill_header_line1": input.BillHeaderLine1,
		"bill_header_line2": input.BillHeaderLine2,
		"bill_header_line3": input.BillHeaderLine3,
		"bill_header_line4": input.BillHeaderLine4,
		"bill_header_line5": input.BillHeaderLine5,
		"logo_id":           input.LogoId,
		"is_active":         input.IsActive,
		"update_by":         userId,
		"update_date":       utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.PosStation{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update station: %w", err)
	}

	// ดึงข้อมูลใหม่โดยใช้ Where ด้วย id
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated station: %w", err)
	}

	return existing, nil
}
func FindExistStationCode(code string) (*models.PosStation, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var station models.PosStation
	lowerCode := strings.ToLower(code)

	err := config.DB_POS.
		Where("LOWER(pos_code) = ?", lowerCode).
		First(&station).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &station, nil
}

func FindStationById(id int) (*models.PosStationData, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var station models.PosStationData

	err := config.DB_POS.
		Table("pos_station AS ps").
		Select("ps.*, ml.name AS logo_name, ml.logo as logo_url").
		Joins("LEFT JOIN master_logo ml ON ps.logo_id = ml.id").
		Where("ps.id = ?", id).
		Scan(&station).Error

	if err != nil {
		return nil, err
	}

	return &station, nil
}

func FindStationByMacAddress(macAddress string) (*models.PosStationData, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	address := strings.ToLower(macAddress)
	var station models.PosStationData

	err := config.DB_POS.
		Table("pos_station AS ps").
		Select("ps.*, ml.name AS logo_name, ml.logo as logo_url").
		Joins("LEFT JOIN master_logo ml ON ps.logo_id = ml.id").
		Where("? = ANY (SELECT jsonb_array_elements_text(ps.mac_address))", address).
		Where("ps.is_delete = false").
		Where("ps.is_active = true").
		Limit(1).
		Scan(&station).Error

	if err != nil {
		return nil, err
	}

	if station.ID == 0 {
		return nil, nil // ไม่เจอข้อมูล
	}

	return &station, nil
}

// func FindStationByMacAddressList(macAddresses []string) (*models.PosStationData, error) {
// 	if config.DB_POS == nil {
// 		return nil, fmt.Errorf("database connection is nil")
// 	}

// 	if len(macAddresses) == 0 {
// 		return nil, fmt.Errorf("macAddresses list is required")
// 	}

// 	for _, mac := range macAddresses {
// 		if strings.TrimSpace(mac) == "" {
// 			continue
// 		}

// 		address := strings.ToLower(mac)

// 		var station models.PosStationData
// 		err := config.DB_POS.
// 			Table("pos_station AS ps").
// 			Select("ps.*, ml.name AS logo_name, ml.logo AS logo_url").
// 			Joins("LEFT JOIN master_logo ml ON ps.logo_id = ml.id").
// 			Where("? = ANY (SELECT jsonb_array_elements_text(ps.mac_address))", address).
// 			Scan(&station).Error

// 		if err != nil {
// 			return nil, err
// 		}

// 		// ถ้าพบ station มีค่า id หรือข้อมูลใดๆ แสดงว่ามีข้อมูล
// 		if station.ID != 0 {
// 			return &station, nil
// 		}
// 	}

// 	// return nil, fmt.Errorf("no matching MAC address found")
// 	return nil, nil // ถ้าไม่พบ station ใดๆ จะคืน nil
// }

func FindStationByMacAddressList(macAddresses []string) (*models.PosStationData, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if len(macAddresses) == 0 {
		return nil, fmt.Errorf("macAddresses list is required")
	}

	// Ensure all MACs are lowercase for consistent matching
	for i, mac := range macAddresses {
		macAddresses[i] = strings.ToLower(mac)
	}

	var station models.PosStationData
	err := config.DB_POS.Raw(`
		SELECT ps.*, ml.name AS logo_name, ml.logo AS logo_url
		FROM pos_station ps
		LEFT JOIN master_logo ml ON ps.logo_id = ml.id,
		LATERAL jsonb_array_elements_text(ps.mac_address) AS mac(mac)
		WHERE LOWER(mac.mac) IN ?
		AND ps.is_delete = false
		AND ps.is_active = true
		LIMIT 1
	`, macAddresses).Scan(&station).Error

	if err != nil {
		return nil, err
	}
	if station.ID != 0 {
		return &station, nil
	}
	return nil, nil
}

func FindStationByMacAddressListNotCurrent(macAddresses []string, id int) (*models.PosStationData, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if len(macAddresses) == 0 {
		return nil, fmt.Errorf("macAddresses list is required")
	}

	for i, mac := range macAddresses {
		macAddresses[i] = strings.ToLower(mac)
	}

	var station models.PosStationData
	err := config.DB_POS.Raw(`
		SELECT ps.*, ml.name AS logo_name, ml.logo AS logo_url
		FROM pos_station ps
		LEFT JOIN master_logo ml ON ps.logo_id = ml.id,
		LATERAL jsonb_array_elements_text(ps.mac_address) AS mac(mac)
		WHERE LOWER(mac.mac) IN ?
		AND ps.id != ?
		AND ps.is_delete = false
		AND ps.is_active = true
		LIMIT 1
	`, macAddresses, id).Scan(&station).Error

	if err != nil {
		return nil, err
	}

	if station.ID != 0 {
		return &station, nil
	}
	return nil, nil
}

func FindStationByLocation(location string) (*[]models.PosStation, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var stations []models.PosStation
	lowerLocation := strings.ToLower(location)

	err := config.DB_POS.
		Where("LOWER(pos_location) = ?", lowerLocation).
		Find(&stations).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &stations, nil
}

func FindExistStationCodeNotCurrent(code string, currentId int) (*models.PosStation, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var station models.PosStation
	lowerCode := strings.ToLower(code)

	result := config.DB_POS.
		Where("LOWER(pos_code) = ? AND id != ?", lowerCode, currentId).
		Find(&station)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &station, nil
}

func StationSearchList(params models.SearchStationParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var stations []models.PosStationData

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	search := strings.ToLower(params.Search)

	query := config.DB_POS.
		Table("pos_station AS ps").
		Select("ps.*, ml.name AS logo_name, ml.logo as logo_url").
		Joins("LEFT JOIN master_logo ml ON ps.logo_id = ml.id").
		Where("ps.is_delete = false")

	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where(
			`LOWER(ps.pos_code) LIKE ? OR LOWER(ps.pos_description) LIKE ?`,
			searchLike, searchLike,
		)
	}
	if strings.ToLower(params.Location) != "" {
		query = query.Where("LOWER(ps.pos_location) = ?", strings.ToLower(params.Location))
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("ps.create_date DESC, ps.pos_location").Limit(params.Skip).Offset(offset).Find(&stations).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     stations,
	}
	return result, nil
}

func DeleteStationById(id int, userId int) (models.PosStation, error) {
	var existing models.PosStation

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("station not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.PosStation{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to delete station: %w", err)
	}

	//ดึงข้อมูลใหม่ด้วย Where
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve station: %w", err)
	}

	return existing, nil
}

func ResetMacAddressById(stationId int, userId int) (models.PosStation, error) {
	var existing models.PosStation

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", stationId).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("station not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"mac_address": nil,
		"update_by":   userId,
		"update_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.PosStation{}).Where("id = ?", stationId).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to delete station: %w", err)
	}

	//ดึงข้อมูลใหม่ด้วย Where
	if err := config.DB_POS.Where("id = ?", stationId).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve station: %w", err)
	}

	return existing, nil
}
