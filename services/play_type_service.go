package services

import (
	"fmt"
	"strings"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"

	"new-pos-api/utils"

	"gorm.io/gorm"
)

func CreatePlayType(input models.PlayTypeDto, userId int) (models.PlayType, error) {
	var entity models.PlayType

	// Map fields จาก DTO ไปยัง Entity
	entity.Name = input.Name
	entity.PromotionTime = input.PromotionTime
	entity.IsActive = input.IsActive
	entity.CreateBy = userId
	entity.CreateDate = *utils.TimeNowAsia()

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create play type: %w", err)
	}

	return entity, nil
}

func UpdatePlayType(id string, input models.PlayTypeDto, userId int) (models.PlayType, error) {
	var existing models.PlayType

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("play type not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"name":           input.Name,
		"promotion_time": input.PromotionTime,
		"is_active":      input.IsActive,
		"update_by":      userId,
		"update_date":    utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.PlayType{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update play type: %w", err)
	}

	// ดึงข้อมูลใหม่หลังอัปเดต
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated play type: %w", err)
	}

	return existing, nil
}

func FindExistPlayTypeName(name string) (*models.PlayType, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var playType models.PlayType
	lowerName := strings.ToLower(name)

	err := config.DB_POS.
		Where("LOWER(name) = ?", lowerName).
		First(&playType).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &playType, nil
}

func FindExistPlayTypeNameNotCurrent(name string, currentId string) (*models.PlayType, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var playType models.PlayType
	lowerName := strings.ToLower(name)

	result := config.DB_POS.
		Where("LOWER(name) = ? AND id != ?", lowerName, currentId).
		Find(&playType)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &playType, nil
}

func FindPlayTypeById(id string) (*models.PlayType, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var playType models.PlayType

	// ใช้ GORM เพื่อค้นหาด้วย Where แทนการใช้ Raw Query
	if err := config.DB_POS.Where("id = ?", id).First(&playType).Error; err != nil {
		// ถ้าค้นหาไม่เจอ หรือเกิดข้อผิดพลาด
		if err == gorm.ErrRecordNotFound {
			return nil, nil // ถ้าไม่พบข้อมูล
		}
		return nil, err // ถ้ามีข้อผิดพลาดอื่น
	}

	return &playType, nil
}

func GetActivePlayType() ([]models.PlayType, error) {
	var playTypes []models.PlayType

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// ใช้ GORM แบบง่ายและปลอดภัย
	if err := config.DB_POS.Where("is_active = ?", true).Find(&playTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active play types: %w", err)
	}

	return playTypes, nil
}

func PlayTypeSearchList(params models.SearchParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var playTypes []models.PlayType

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	search := strings.ToLower(params.Search)
	// var args []interface{}
	query := config.DB_POS.Model(&models.PlayType{}).Where("is_delete = false")

	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where(
			`LOWER(name) LIKE ?`,
			searchLike,
		)
	} else {

	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("name").Limit(params.Skip).Offset(offset).Find(&playTypes).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     playTypes,
	}
	return result, nil
}

func DeletePlayTypeById(id string, userId int) (models.PlayType, error) {
	var existing models.PlayType

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("play type not found: %w", err)
	}
	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.PlayType{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to delete play type: %w", err)
	}

	// ดึงข้อมูลใหม่หลังอัปเดต
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve play type: %w", err)
	}

	return existing, nil
}
