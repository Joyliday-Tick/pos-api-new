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

func CreateCardType(input models.CardTypeDto, userId int) (models.CardType, error) {
	var entity models.CardType

	// Map fields จาก DTO ไปยัง Entity
	entity.Name = input.Name
	entity.IsActive = input.IsActive
	entity.ShowBalance = input.ShowBalance
	entity.CreateBy = userId
	entity.CreateDate = *utils.TimeNowAsia()

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create card type: %w", err)
	}

	return entity, nil
}

func UpdateCardType(id string, input models.CardTypeDto, userId int) (models.CardType, error) {
	var existing models.CardType

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("card type not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"name":         input.Name,
		"is_active":    input.IsActive,
		"show_balance": input.ShowBalance,
		"update_by":    userId,
		"update_date":  utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.CardType{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update card type: %w", err)
	}

	// ดึงข้อมูลใหม่โดยใช้ Where ด้วย id
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated card type: %w", err)
	}

	return existing, nil
}
func FindExistCardTypeName(name string) (*models.CardType, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var cardType models.CardType
	lowerName := strings.ToLower(name)

	err := config.DB_POS.
		Where("LOWER(name) = ?", lowerName).
		First(&cardType).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &cardType, nil
}

func FindExistCardTypeNameNotCurrent(name string, currentId string) (*models.CardType, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var cardType models.CardType
	lowerName := strings.ToLower(name)

	result := config.DB_POS.
		Where("LOWER(name) = ? AND id != ?", lowerName, currentId).
		Find(&cardType)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &cardType, nil
}

func FindCardTypeById(id string) (*models.CardType, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var cardType models.CardType

	// ใช้ GORM เพื่อค้นหาด้วย Where แทนการใช้ Raw Query
	if err := config.DB_POS.Where("id = ?", id).First(&cardType).Error; err != nil {
		// ถ้าค้นหาไม่เจอ หรือเกิดข้อผิดพลาด
		if err == gorm.ErrRecordNotFound {
			return nil, nil // ถ้าไม่พบข้อมูล
		}
		return nil, err // ถ้ามีข้อผิดพลาดอื่น
	}

	return &cardType, nil
}

func GetActiveCardType() ([]models.CardType, error) {
	var cardTypes []models.CardType

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// ใช้ GORM แบบง่ายและปลอดภัย
	if err := config.DB_POS.Where("is_active = ?", true).Find(&cardTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active card types: %w", err)
	}

	return cardTypes, nil
}

func CardTypeSearchList(params models.SearchParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var cardTypes []models.CardType

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	search := strings.ToLower(params.Search)
	// var args []interface{}
	query := config.DB_POS.Model(&models.CardType{}).Where("is_delete = false")

	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where(
			`LOWER(name) LIKE ?`,
			searchLike,
		)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("name").Limit(params.Skip).Offset(offset).Find(&cardTypes).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     cardTypes,
	}
	return result, nil
}

func DeleteCardTypeById(id string, userId int) (models.CardType, error) {
	var existing models.CardType

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("card type not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.CardType{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to delete card type: %w", err)
	}

	//ดึงข้อมูลใหม่ด้วย Where
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve card type: %w", err)
	}

	return existing, nil
}
