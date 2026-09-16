package services

import (
	"fmt"
	"strings"

	// "strings"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"

	"gorm.io/gorm"
	// "new-pos-api/utils"
	// "gorm.io/gorm"
)

func GetActiveMachineGroup() ([]models.MachineGroupList, error) {
	var machineGroups []models.MachineGroup

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// ดึง machine ที่ Active เท่านั้น
	if err := config.DB_POS.Where("is_active = ?", true).Find(&machineGroups).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active machine group: %w", err)
	}

	var result []models.MachineGroupList

	for i := range machineGroups {
		group := machineGroups[i]
		machineGroupId := group.ID

		var subMachines []string
		if err := config.DB_POS.Model(&models.MachineSubGroup{}).
			Where("group_id = ?", machineGroupId).
			Pluck("machine_id", &subMachines).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch sub machines for machine group %d: %w", machineGroupId, err)
		}

		// แมปจาก PosBranch ไปเป็น PosBranchGroupList
		result = append(result, models.MachineGroupList{
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
			MachineGroupType: group.MachineGroupType,
			SubMachineId:     subMachines,
		})
	}

	return result, nil
}

func CreateMachineGroup(input models.MachineGroupDto, userId int) (models.MachineGroup, error) {
	var entity models.MachineGroup

	entity.GroupName = input.GroupName
	entity.IsActive = input.IsActive
	entity.CreateBy = userId
	entity.CreateDate = *utils.TimeNowAsia()
	entity.MachineGroupType = *input.MachineGroupType

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create branch group: %w", err)
	}

	return entity, nil
}

func UpdateMachineGroup(id int, input models.MachineGroupDto, userId int) (models.MachineGroup, error) {
	var existing models.MachineGroup

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("machine group not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"group_name":         input.GroupName,
		"is_active":          input.IsActive,
		"update_by":          userId,
		"update_date":        utils.TimeNowAsia(),
		"machine_group_type": input.MachineGroupType,
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.MachineGroup{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update machine group: %w", err)
	}

	// ดึงข้อมูลใหม่โดยใช้ Where ด้วย id
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve machine group: %w", err)
	}

	return existing, nil
}

func FindExistMachineGroupName(name string) (*models.MachineGroup, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var machineGroup models.MachineGroup
	lowerName := strings.ToLower(name)

	err := config.DB_POS.
		Where("LOWER(group_name) = ?", lowerName).
		First(&machineGroup).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &machineGroup, nil
}

func FindExistMachineGroupNameNotCurrent(name string, currentId int) (*models.MachineGroup, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var machineGroup models.MachineGroup
	lowerName := strings.ToLower(name)

	result := config.DB_POS.
		Where("LOWER(group_name) = ? AND id != ?", lowerName, currentId).
		Find(&machineGroup)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &machineGroup, nil
}

func FindMachineGroupById(id int) (*models.MachineGroupList, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var machineGroup models.MachineGroup

	if err := config.DB_POS.Where("id = ?", id).First(&machineGroup).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	var subMachines []string
	if err := config.DB_POS.Model(&models.MachineSubGroup{}).
		Where("group_id = ?", id).
		Pluck("machine_id", &subMachines).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sub machines for machine group %d: %w", id, err)
	}

	result := models.MachineGroupList{
		ID:               machineGroup.ID,
		GroupName:        machineGroup.GroupName,
		IsActive:         machineGroup.IsActive,
		CreateBy:         machineGroup.CreateBy,
		CreateDate:       machineGroup.CreateDate,
		UpdateBy:         machineGroup.UpdateBy,
		UpdateDate:       machineGroup.UpdateDate,
		IsDelete:         machineGroup.IsDelete,
		DeleteBy:         machineGroup.DeleteBy,
		DeleteDate:       machineGroup.DeleteDate,
		MachineGroupType: machineGroup.MachineGroupType,
		SubMachineId:     subMachines,
	}

	return &result, nil
}

func MachineGroupSearchList(params models.SearchParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var machineGroups []models.MachineGroup

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	search := strings.ToLower(params.Search)
	// var args []interface{}
	query := config.DB_POS.Model(&models.MachineGroup{}).Where("is_delete = false")

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
	if err := query.Order("group_name").Limit(params.Skip).Offset(offset).Find(&machineGroups).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     machineGroups,
	}
	return result, nil
}

func DeleteMachineGroupById(id int, userId int) (models.MachineGroup, error) {
	var existing models.MachineGroup

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("machine group not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.MachineGroup{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to delete machine group: %w", err)
	}

	//ดึงข้อมูลใหม่ด้วย Where
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve machine group: %w", err)
	}

	return existing, nil
}

func CreateBatchMachineSubGroup(subGroups []models.MachineSubGroup, userId int) ([]models.MachineSubGroup, error) {

	if config.DB_POS == nil {
		return subGroups, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&subGroups).Error; err != nil {
		return subGroups, fmt.Errorf("failed to create batch  machine sub group: %w", err)

	}

	return subGroups, nil
}

func DeleteMachineSubGroupByGroupId(groupId int, userId int) ([]models.MachineSubGroup, error) {
	var existing []models.MachineSubGroup

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("group_id = ?", groupId).Find(&existing).Error; err != nil {
		return nil, fmt.Errorf("machine sub group not found: %w", err)
	}

	if len(existing) == 0 {
		return nil, nil
	}

	// delete
	if err := config.DB_POS.Where("group_id = ?", groupId).Delete(&models.MachineSubGroup{}).Error; err != nil {
		return nil, fmt.Errorf("failed to delete machine sub groups: %w", err)
	}

	return existing, nil
}

func FindMachineGroupByIdV1(id int) (*models.MachineGroupListV1, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var machineGroup models.MachineGroup

	if err := config.DB_POS.Where("id = ?", id).First(&machineGroup).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	var subMachinesList []models.MachineSubGroup
	if err := config.DB_POS.Where("group_id = ?", id).Find(&subMachinesList).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sub machines for machine group %d: %w", id, err)
	}

	result := models.MachineGroupListV1{
		ID:               machineGroup.ID,
		GroupName:        machineGroup.GroupName,
		IsActive:         machineGroup.IsActive,
		CreateBy:         machineGroup.CreateBy,
		CreateDate:       machineGroup.CreateDate,
		UpdateBy:         machineGroup.UpdateBy,
		UpdateDate:       machineGroup.UpdateDate,
		IsDelete:         machineGroup.IsDelete,
		DeleteBy:         machineGroup.DeleteBy,
		DeleteDate:       machineGroup.DeleteDate,
		MachineGroupType: machineGroup.MachineGroupType,
		SubMachine:       subMachinesList,
	}

	return &result, nil
}

func FindMachineSubGroupByGroupId(id int) ([]models.MachineSubGroup, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var subMachinesList []models.MachineSubGroup
	if err := config.DB_POS.Where("group_id = ?", id).Find(&subMachinesList).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sub machines for machine group %d: %w", id, err)
	}

	return subMachinesList, nil
}
