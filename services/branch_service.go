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

func GetActiveBranch() ([]models.Branch, error) {
	var branchs []models.Branch

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// // ดึง branch ที่ Active เท่านั้น
	// if err := config.DB_POS.Where("is_active = ?", true).Find(&branchs).Order("branch_code").Error; err != nil {
	// 	return nil, fmt.Errorf("failed to fetch active branch: %w", err)
	// }

	if err := config.DB_POS.
		Where("is_active = ?", true).
		Order("branch_code").
		Find(&branchs).Error; err != nil {

		return nil, fmt.Errorf("failed to fetch active branch: %w", err)
	}

	return branchs, nil
}

func CreateBranch(input models.BranchDto, userId int) (models.Branch, error) {
	var branch models.Branch

	branch.BranchCode = input.BranchCode
	branch.BranchName = input.BranchName
	branch.IsActive = input.IsActive
	branch.CreateBy = userId
	branch.CreateDate = utils.TimeNowAsia()

	if config.DB_POS == nil {
		return branch, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&branch).Error; err != nil {
		return branch, fmt.Errorf("failed to create branch: %w", err)
	}

	return branch, nil
}

func UpdateBranch(code string, input models.BranchDto, userId int) (models.Branch, error) {
	var existing models.Branch

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("branch_code = ?", code).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("branch not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"branch_code": input.BranchCode,
		"branch_name": input.BranchName,
		"b_main":      input.BMain,
		"is_active":   input.IsActive,
		"update_by":   userId,
		"update_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.Branch{}).Where("branch_code = ?", code).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update branch: %w", err)
	}

	// ดึงข้อมูลใหม่โดยใช้ Where ด้วย id
	if err := config.DB_POS.Where("code = ?", code).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated branch: %w", err)
	}

	return existing, nil
}
func FindExistBranchName(name string) (*models.Branch, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var branch models.Branch
	lowerName := strings.ToLower(name)

	err := config.DB_POS.
		Where("LOWER(branch_name) = ?", lowerName).
		First(&branch).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &branch, nil
}

func FindExistBranchCode(code string) (*models.Branch, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var branch models.Branch
	lowerCode := strings.ToLower(code)

	err := config.DB_POS.
		Where("LOWER(branch_code) = ?", lowerCode).
		First(&branch).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &branch, nil
}

func FindExistBranchNameNotCurrent(name string, currentCode string) (*models.Branch, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var branch models.Branch
	lowerName := strings.ToLower(name)

	result := config.DB_POS.
		Where("LOWER(branch_name) = ? AND branch_code != ?", lowerName, currentCode).
		Find(&branch)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &branch, nil
}

func FindBranchByMain(branchMainCode string) (*[]models.Branch, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var branches []models.Branch
	lowerMain := strings.ToLower(branchMainCode)
	// ใช้ GORM เพื่อค้นหาด้วย Where แทนการใช้ Raw Query
	if err := config.DB_POS.Where("LOWER(b_main) = ?", lowerMain).Find(&branches).Error; err != nil {
		// ถ้าค้นหาไม่เจอ หรือเกิดข้อผิดพลาด
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &branches, nil
}

func BranchSearchList(params models.SearchParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var branches []models.Branch

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	search := strings.ToLower(params.Search)
	// var args []interface{}
	query := config.DB_POS.Model(&models.Branch{}).Where("is_delete = false")

	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where(
			`LOWER(branch_name) LIKE ? OR LOWER(branch_code) LIKE ?`,
			searchLike, searchLike,
		)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("branch_code").Limit(params.Skip).Offset(offset).Find(&branches).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     branches,
	}
	return result, nil
}

func DeleteBranchByCode(code string, userId int) (models.Branch, error) {
	var existing models.Branch

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("branch_code = ?", code).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("branch not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.Branch{}).Where("branch_code = ?", code).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to delete branch: %w", err)
	}

	//ดึงข้อมูลใหม่ด้วย Where
	if err := config.DB_POS.Where("branch_code = ?", code).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve branch: %w", err)
	}

	return existing, nil
}
