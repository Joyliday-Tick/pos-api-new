package services

import (
	"errors"
	"fmt"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"
	"strings"

	"gorm.io/gorm"
)

func GetAllUsers() ([]models.UserLite, error) {
	var users []models.UserLite

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	err := config.DB_POS.
		Model(&models.Users{}).
		Select("id", "name", "s_name"). // ✅ ถูกแล้ว
		Scan(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	return users, nil
}

func FindUserDbByUsername(username string) (*models.Users, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	var user models.Users
	result := config.DB_POS.
		Where("LOWER(username) = ? AND is_delete = false AND is_active = true", strings.ToLower(username)).
		First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func FindUserById(id int) (*models.Users, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	var user models.Users
	result := config.DB_POS.
		Where("id = ? AND is_delete = false", id).
		First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func CreateUser(input models.UsersCreateDto, userId int) (models.Users, error) {
	var entity models.Users

	entity.Username = input.Username
	entity.Password = input.Password
	entity.Name = input.Name
	entity.SName = input.SName
	entity.Status = input.Status
	entity.UserRoleId = input.UserRoleId
	entity.UserGroupId = input.UserGroupId
	entity.IsActive = input.IsActive
	entity.CreateBy = userId
	entity.CreateDate = utils.TimeNowAsia()

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create user: %w", err)
	}

	return entity, nil
}

func UpdateUser(id int, input models.UsersUpdateDto, userId int) (models.Users, error) {
	var existing models.Users

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("user not found: %w", err)
	}

	// pass := existing.Password
	// if input.NewPassword != "" {
	// 	if input.NewPassword == existing.Password {
	// 		return existing, fmt.Errorf("new password cannot be the same as the old password")
	// 	}
	// 	if input.Password != existing.Password {
	// 		return existing, fmt.Errorf("current password is incorrect")
	// 	}
	// 	pass = input.NewPassword
	// }

	updateData := map[string]interface{}{
		"username":      input.Username,
		"password":      input.Password,
		"name":          input.Name,
		"s_name":        input.SName,
		"status":        input.Status,
		"user_role_id":  input.UserRoleId,
		"user_group_id": input.UserGroupId,
		"is_active":     input.IsActive,
		"update_by":     userId,
		"update_date":   utils.TimeNowAsia(),
	}

	if err := config.DB_POS.Model(&models.Users{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update user: %w", err)
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated user: %w", err)
	}

	return existing, nil
}

func FindExistUserNameNotCurrent(name string, currentId int) (*models.Users, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var users models.Users
	lowerName := strings.ToLower(name)

	result := config.DB_POS.
		Where("LOWER(username) = ? AND id != ? and is_delete = false", lowerName, currentId).
		Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &users, nil
}

func UserSearchList(params models.SearchUserParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var userList []models.UsersData

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	username := strings.ToLower(params.Username)
	search := strings.ToLower(params.Search)

	query := config.DB_POS.
		Table("users AS u").
		Select("u.*, ur.role_description AS user_role_name, ug.group_name AS user_group_name").
		Joins("LEFT JOIN user_role ur ON u.user_role_id = ur.id").
		Joins("LEFT JOIN user_group ug ON u.user_group_id = ug.id").
		Where("u.is_delete = false")

	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where(
			`LOWER(u.name) LIKE ? OR LOWER(u.s_name) LIKE ?`,
			searchLike, searchLike,
		)
	}
	if username != "" {
		query = query.Where("LOWER(u.username) LIKE ?", "%"+username+"%")
	}

	if params.UserRoleId != nil {
		query = query.Where("u.user_role_id = ?", params.UserRoleId)
	}

	if params.UserGroupId != nil {
		query = query.Where("u.user_group_id = ?", params.UserGroupId)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("u.create_date DESC").Limit(params.Skip).Offset(offset).Find(&userList).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     userList,
	}
	return result, nil
}

func DeleteUserById(id int, userId int) (models.Users, error) {
	var existing models.Users

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("user not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.Users{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to delete user: %w", err)
	}

	//ดึงข้อมูลใหม่ด้วย Where
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve user: %w", err)
	}

	return existing, nil
}
