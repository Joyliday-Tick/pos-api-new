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

func GetActiveUserGroup() ([]models.UserGroupList, error) {
	var userGroups []models.UserGroup

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// ดึง user group ที่ Active เท่านั้น
	if err := config.DB_POS.Where("is_active = ?", true).Find(&userGroups).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active user group: %w", err)
	}

	// เตรียม slice สำหรับ UserGroupList
	var result []models.UserGroupList

	for i := range userGroups {
		group := userGroups[i]
		userGroupId := group.ID

		// หา sub branch โดย Pluck ค่า location_code เท่านั้น
		var subBranches []string
		if err := config.DB_POS.Model(&models.UsersSubGroup{}).
			Where("group_id = ?", userGroupId).
			Pluck("location_code", &subBranches).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch sub branches for user group %d: %w", userGroupId, err)
		}

		// แมปจาก UsersSubGroup ไปเป็น UserGroupList
		result = append(result, models.UserGroupList{
			ID:               group.ID,
			GroupName:        group.GroupName,
			IsActive:         group.IsActive,
			CreateBy:         group.CreateBy,
			CreateDate:       group.CreateDate,
			UpdateBy:         group.UpdateBy,
			UpdateDate:       group.UpdateDate,
			IsDelete:         group.IsDelete,
			DeleteBy:         group.DeleteBy,
			DeleteDate:       group.DeleteDate,
			SubLocationCodes: subBranches,
		})
	}

	return result, nil
}

func FindUserGroupById(groupId int) (*models.UserGroup, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	var group models.UserGroup
	result := config.DB_POS.
		Where("id = ? AND is_delete = false AND is_active = true", groupId).
		First(&group)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &group, nil
}

func FindUserSubGroupByGroupId(groupId int) ([]string, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var locations []string
	err := config.DB_POS.
		Table("user_sub_group").
		Where("group_id = ?", groupId).
		Order("location_code ASC").
		Pluck("location_code", &locations).Error

	return locations, err
}

func CreateUserGroup(input models.UserGroupDto, userId int) (models.UserGroup, error) {
	var entity models.UserGroup

	entity.GroupName = input.GroupName
	entity.IsActive = input.IsActive
	entity.CreateBy = userId
	entity.CreateDate = utils.TimeNowAsia()

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create user group: %w", err)
	}

	return entity, nil
}

func UpdateUserGroup(id int, input models.UserGroupDto, userId int) (models.UserGroup, error) {
	var existing models.UserGroup

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("user group not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"group_name":  input.GroupName,
		"is_active":   input.IsActive,
		"update_by":   userId,
		"update_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.UserGroup{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update user group: %w", err)
	}

	// ดึงข้อมูลใหม่โดยใช้ Where ด้วย id
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve user group: %w", err)
	}

	return existing, nil
}

func FindExistUserGroupName(name string) (*models.UserGroup, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var userGroup models.UserGroup
	lowerName := strings.ToLower(name)

	err := config.DB_POS.
		Where("LOWER(group_name) = ?", lowerName).
		First(&userGroup).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &userGroup, nil
}

func FindExistUserGroupNameNotCurrent(name string, currentId int) (*models.UserGroup, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var userGroup models.UserGroup
	lowerName := strings.ToLower(name)

	result := config.DB_POS.
		Where("LOWER(group_name) = ? AND id != ?", lowerName, currentId).
		Find(&userGroup)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &userGroup, nil
}

func FindUserGroupByIdWithLocation(id int) (*models.UserGroupList, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var userGroup models.UserGroup

	if err := config.DB_POS.Where("id = ?", id).First(&userGroup).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	var subBranches []string
	if err := config.DB_POS.Model(&models.UsersSubGroup{}).
		Where("group_id = ?", id).
		Pluck("location_code", &subBranches).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sub branches for user group %d: %w", id, err)
	}

	result := models.UserGroupList{
		ID:               userGroup.ID,
		GroupName:        userGroup.GroupName,
		IsActive:         userGroup.IsActive,
		CreateBy:         userGroup.CreateBy,
		CreateDate:       userGroup.CreateDate,
		UpdateBy:         userGroup.UpdateBy,
		UpdateDate:       userGroup.UpdateDate,
		IsDelete:         userGroup.IsDelete,
		DeleteBy:         userGroup.DeleteBy,
		DeleteDate:       userGroup.DeleteDate,
		SubLocationCodes: subBranches,
	}

	return &result, nil
}

func DeleteUserGroupById(id int, userId int) (models.UserGroup, error) {
	var existing models.UserGroup

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("user group not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.UserGroup{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to delete user group: %w", err)
	}

	//ดึงข้อมูลใหม่ด้วย Where
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve user group: %w", err)
	}

	return existing, nil
}

func UserGroupSearchList(params models.SearchParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var userGroups []models.UserGroup

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	search := strings.ToLower(params.Search)
	// var args []interface{}
	query := config.DB_POS.Model(&models.UserGroup{}).Where("is_delete = false")

	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where(
			`LOWER(group_name) LIKE ?`,
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
	if err := query.Order("group_name").Limit(params.Skip).Offset(offset).Find(&userGroups).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     userGroups,
	}
	return result, nil
}

func CreateBatchUserSubGroup(subGroups []models.UsersSubGroup, userId int) ([]models.UsersSubGroup, error) {

	if config.DB_POS == nil {
		return subGroups, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&subGroups).Error; err != nil {
		return subGroups, fmt.Errorf("failed to create batch  branch sub group: %w", err)

	}

	return subGroups, nil
}

func DeleteUserSubGroupByGroupId(groupId int, userId int) ([]models.UsersSubGroup, error) {
	var existing []models.UsersSubGroup

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("group_id = ?", groupId).Find(&existing).Error; err != nil {
		return nil, fmt.Errorf("user sub group not found: %w", err)
	}

	if len(existing) == 0 {
		return nil, nil
	}

	// delete
	if err := config.DB_POS.Where("group_id = ?", groupId).Delete(&models.UsersSubGroup{}).Error; err != nil {
		return nil, fmt.Errorf("failed to delete user sub groups: %w", err)
	}

	return existing, nil
}
