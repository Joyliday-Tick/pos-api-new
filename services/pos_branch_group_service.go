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

func GetActiveBranchGroup() ([]models.PosBranchGroupList, error) {
	var branchGroups []models.PosBranchGroup

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// ดึง branch ที่ Active เท่านั้น
	if err := config.DB_POS.Where("is_active = ?", true).Find(&branchGroups).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active branch group: %w", err)
	}

	// เตรียม slice สำหรับ PosBranchGroupList
	var result []models.PosBranchGroupList

	for i := range branchGroups {
		branch := branchGroups[i]
		branchGroupId := branch.ID

		// หา sub branch โดย Pluck ค่า location_code เท่านั้น
		var subBranches []string
		if err := config.DB_POS.Model(&models.PosBranchSubGroup{}).
			Where("group_id = ?", branchGroupId).
			Pluck("location_code", &subBranches).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch sub branches for branch group %d: %w", branchGroupId, err)
		}

		// แมปจาก PosBranch ไปเป็น PosBranchGroupList
		result = append(result, models.PosBranchGroupList{
			ID:               branch.ID,
			GroupName:        branch.GroupName,
			IsActive:         branch.IsActive,
			CreateBy:         branch.CreateBy,
			CreateDate:       branch.CreateDate,
			UpdateBy:         branch.UpdateBy,
			UpdateDate:       branch.UpdateDate,
			IsDelete:         branch.IsDelete,
			DeleteBy:         branch.DeleteBy,
			DeleteDate:       branch.DeleteDate,
			SubLocationCodes: subBranches,
		})
	}

	return result, nil
}

func CreateBranchGroup(input models.PosBranchGroupDto, userId int) (models.PosBranchGroup, error) {
	var entity models.PosBranchGroup

	entity.GroupName = input.GroupName
	entity.IsActive = input.IsActive
	entity.CreateBy = userId
	entity.CreateDate = *utils.TimeNowAsia()

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create branch group: %w", err)
	}

	return entity, nil
}

func UpdateBranchGroup(id int, input models.PosBranchGroupDto, userId int) (models.PosBranchGroup, error) {
	var existing models.PosBranchGroup

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("branch group not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"group_name":  input.GroupName,
		"is_active":   input.IsActive,
		"update_by":   userId,
		"update_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.PosBranchGroup{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update branch group: %w", err)
	}

	// ดึงข้อมูลใหม่โดยใช้ Where ด้วย id
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve branch group: %w", err)
	}

	return existing, nil
}

func FindExistBranchGroupName(name string) (*models.PosBranchGroup, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var branchGroup models.PosBranchGroup
	lowerName := strings.ToLower(name)

	err := config.DB_POS.
		Where("LOWER(group_name) = ?", lowerName).
		First(&branchGroup).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &branchGroup, nil
}

func FindExistBranchGroupNameNotCurrent(name string, currentId int) (*models.PosBranchGroup, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var branchGroup models.PosBranchGroup
	lowerName := strings.ToLower(name)

	result := config.DB_POS.
		Where("LOWER(group_name) = ? AND id != ?", lowerName, currentId).
		Find(&branchGroup)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &branchGroup, nil
}

func FindBranchGroupById(id int) (*models.PosBranchGroupList, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var branchGroup models.PosBranchGroup

	if err := config.DB_POS.Where("id = ?", id).First(&branchGroup).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	var subBranches []string
	if err := config.DB_POS.Model(&models.PosBranchSubGroup{}).
		Where("group_id = ?", id).
		Pluck("location_code", &subBranches).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sub branches for branch group %d: %w", id, err)
	}

	result := models.PosBranchGroupList{
		ID:               branchGroup.ID,
		GroupName:        branchGroup.GroupName,
		IsActive:         branchGroup.IsActive,
		CreateBy:         branchGroup.CreateBy,
		CreateDate:       branchGroup.CreateDate,
		UpdateBy:         branchGroup.UpdateBy,
		UpdateDate:       branchGroup.UpdateDate,
		IsDelete:         branchGroup.IsDelete,
		DeleteBy:         branchGroup.DeleteBy,
		DeleteDate:       branchGroup.DeleteDate,
		SubLocationCodes: subBranches,
	}

	return &result, nil
}

func BranchGroupSearchList(params models.SearchParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var branchgroups []models.PosBranchGroup

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	search := strings.ToLower(params.Search)
	// var args []interface{}
	query := config.DB_POS.Model(&models.PosBranchGroup{}).Where("is_delete = false")

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
	if err := query.Order("group_name").Limit(params.Skip).Offset(offset).Find(&branchgroups).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     branchgroups,
	}
	return result, nil
}

func DeleteBranchGroupById(id int, userId int) (models.PosBranchGroup, error) {
	var existing models.PosBranchGroup

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("branch group not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.PosBranchGroup{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to delete branch group: %w", err)
	}

	//ดึงข้อมูลใหม่ด้วย Where
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve branch group: %w", err)
	}

	return existing, nil
}

func CreateBatchBranchSubGroup(subGroups []models.PosBranchSubGroup, userId int) ([]models.PosBranchSubGroup, error) {

	if config.DB_POS == nil {
		return subGroups, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&subGroups).Error; err != nil {
		return subGroups, fmt.Errorf("failed to create batch  branch sub group: %w", err)

	}

	return subGroups, nil
}

func DeleteBranchSubGroupByGroupId(groupId int, userId int) ([]models.PosBranchSubGroup, error) {
	var existing []models.PosBranchSubGroup

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("group_id = ?", groupId).Find(&existing).Error; err != nil {
		return nil, fmt.Errorf("branch sub group not found: %w", err)
	}

	if len(existing) == 0 {
		return nil, nil
	}

	// delete
	if err := config.DB_POS.Where("group_id = ?", groupId).Delete(&models.PosBranchSubGroup{}).Error; err != nil {
		return nil, fmt.Errorf("failed to delete branch sub groups: %w", err)
	}

	return existing, nil
}
