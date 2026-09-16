package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"
	"strings"

	"gorm.io/gorm"
)

func FindUserRoleById(roleId int) (*models.UserRole, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	var role models.UserRole
	result := config.DB_POS.
		Where("id = ? AND is_delete = false", roleId).
		First(&role)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &role, nil
}

func GetActiveUserRole() ([]models.UserRole, error) {
	var userRoles []models.UserRole

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// ใช้ GORM แบบง่ายและปลอดภัย
	if err := config.DB_POS.Where("is_active = ?", true).Find(&userRoles).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active user roles: %w", err)
	}

	return userRoles, nil
}
func CreateUserRole(input models.UserRoleDto, userId int) (models.UserRole, error) {
	var entity models.UserRole

	entity.RoleDescription = input.RoleDescription
	entity.IsActive = input.IsActive
	entity.CreateBy = userId
	entity.CreateDate = utils.TimeNowAsia()

	if len(input.ListMenu) > 0 {
		listMenuJSON, err := json.Marshal(input.ListMenu)
		if err != nil {
			return entity, fmt.Errorf("failed to marshal list menu: %w", err)
		}
		entity.ListMenu = listMenuJSON
	} else {
		entity.ListMenu = nil
	}

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create user role: %w", err)
	}

	return entity, nil
}

func UpdateUserRole(id int, input models.UserRoleDto, userId int) (models.UserRole, error) {
	var existing models.UserRole

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("user role not found: %w", err)
	}

	// listMenu := json.RawMessage("null")
	var listMenu any = nil
	if len(input.ListMenu) > 0 {
		listMenuJSON, err := json.Marshal(input.ListMenu)
		if err != nil {
			return existing, fmt.Errorf("failed to marshal list menu: %w", err)
		}
		listMenu = listMenuJSON
	}

	updateData := map[string]interface{}{
		"role_description": input.RoleDescription,
		"list_menu":        listMenu,
		"is_active":        input.IsActive,
		"update_by":        userId,
		"update_date":      utils.TimeNowAsia(),
	}

	if err := config.DB_POS.Model(&models.UserRole{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update user role: %w", err)
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated user role: %w", err)
	}

	return existing, nil
}

func FindExistRoleName(name string) (*models.UserRole, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var userRole models.UserRole
	lowerName := strings.ToLower(name)

	err := config.DB_POS.
		Where("LOWER(role_description) = ?", lowerName).
		First(&userRole).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &userRole, nil
}

func FindExistRoleNameNotCurrent(name string, currentId int) (*models.UserRole, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var userRole models.UserRole
	lowerName := strings.ToLower(name)

	result := config.DB_POS.
		Where("LOWER(role_description) = ? AND id != ?", lowerName, currentId).
		Find(&userRole)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &userRole, nil
}

func UserRoleSearchList(params models.SearchParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var userRole []models.UserRole

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	search := strings.ToLower(params.Search)
	// var args []interface{}
	query := config.DB_POS.Model(&models.UserRole{}).Where("is_delete = false")

	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where(
			`LOWER(role_description) LIKE ?`,
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
	if err := query.Order("role_description").Limit(params.Skip).Offset(offset).Find(&userRole).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     userRole,
	}
	return result, nil
}

func DeleteUserRoleById(id int, userId int) (models.UserRole, error) {
	var existing models.UserRole

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("user role not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.UserRole{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to delete user role: %w", err)
	}

	//ดึงข้อมูลใหม่ด้วย Where
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve user role: %w", err)
	}

	return existing, nil
}
