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

func CreateGroupMenu(input models.GroupMenuDto, userId int) (models.GroupMenu, error) {
	var entity models.GroupMenu

	// Map fields จาก DTO ไปยัง Entity
	entity.Name = input.Name
	entity.IsActive = input.IsActive
	entity.EnableMember = input.EnableMember
	entity.DisplayName = input.DisplayName
	entity.ReceiptStub = input.ReceiptStub
	entity.CreateBy = userId
	entity.Sorting = input.Sorting
	entity.CreateDate = *utils.TimeNowAsia()

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Select("*").Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create group menu: %w", err)
	}

	return entity, nil
}

func UpdateGroupMenu(id string, input models.GroupMenuDto, userId int) (models.GroupMenu, error) {
	var existing models.GroupMenu

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("group menu not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"name":          input.Name,
		"display_name":  input.DisplayName,
		"is_active":     input.IsActive,
		"enable_member": input.EnableMember,
		"receipt_stub":  input.ReceiptStub,
		"sorting":       input.Sorting,
		"update_by":     userId,
		"update_date":   utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.GroupMenu{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update group menu: %w", err)
	}

	// ดึงข้อมูลใหม่โดยใช้ Where ด้วย id
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated group menu: %w", err)
	}

	return existing, nil
}
func FindExistGroupMenuName(name string) (*models.GroupMenu, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var groupMenu models.GroupMenu
	lowerName := strings.ToLower(name)

	err := config.DB_POS.
		Where("LOWER(name) = ?", lowerName).
		First(&groupMenu).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &groupMenu, nil
}

func FindExistGroupMenuNameNotCurrent(name string, currentId string) (*models.GroupMenu, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var groupMenu models.GroupMenu
	lowerName := strings.ToLower(name)

	result := config.DB_POS.
		Where("LOWER(name) = ? AND id != ?", lowerName, currentId).
		Find(&groupMenu)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &groupMenu, nil
}

func FindGroupMenuById(id string) (*models.GroupMenu, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var groupMenu models.GroupMenu

	// ใช้ GORM เพื่อค้นหาด้วย Where แทนการใช้ Raw Query
	if err := config.DB_POS.Where("id = ?", id).First(&groupMenu).Error; err != nil {
		// ถ้าค้นหาไม่เจอ หรือเกิดข้อผิดพลาด
		if err == gorm.ErrRecordNotFound {
			return nil, nil // ถ้าไม่พบข้อมูล
		}
		return nil, err // ถ้ามีข้อผิดพลาดอื่น
	}

	return &groupMenu, nil
}

func GetActiveGroupMenu() ([]models.GroupMenu, error) {
	var groupMenus []models.GroupMenu

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// ใช้ GORM แบบง่ายและปลอดภัย
	if err := config.DB_POS.Where("is_active = ?", true).Order("sorting").Find(&groupMenus).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active group menu: %w", err)
	}

	return groupMenus, nil
}

func GroupMenuSearchList(params models.SearchParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var groupMenus []models.GroupMenu

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	search := strings.ToLower(params.Search)
	query := config.DB_POS.Model(&models.GroupMenu{}).Where("is_delete = false")

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
	if err := query.Order("name").Limit(params.Skip).Offset(offset).Find(&groupMenus).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     groupMenus,
	}
	return result, nil
}

func DeleteGroupMenuById(id string, userId int) (models.GroupMenu, error) {
	var existing models.GroupMenu

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {

		return existing, fmt.Errorf("group menu not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.GroupMenu{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update group menu: %w", err)
	}

	//ดึงข้อมูลใหม่ด้วย Where
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated group menu: %w", err)
	}

	return existing, nil
}
