package services

import (
	"encoding/json"
	"fmt"
	"strings"

	// "strings"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"

	"gorm.io/gorm"
	// "gorm.io/gorm"
)

func CreatePosMenu(input models.PosMenuDto, userId int) (models.PosMenu, error) {
	var menu models.PosMenu

	menu.Code = input.Code
	menu.Description = input.Description
	menu.Price = input.Price
	menu.ECoin = input.ECoin
	menu.EBonus = input.EBonus
	menu.BranchGroupID = input.BranchGroupID
	menu.MachineGroupID = input.MachineGroupID
	menu.IsActive = input.IsActive
	menu.LimitTime = input.LimitTime
	menu.CardTypeID = input.CardTypeID
	menu.DiscountCash = input.DiscountCash
	menu.DiscountCashExpireDate = input.DiscountCashExpireDate
	menu.DiscountCashLimitDays = input.DiscountCashlimitDays
	menu.BonusExpireLimitDays = input.BonusExpireLimitDays
	menu.CardExpireLimitMinutes = input.CardExpireLimitMinutes

	if input.BranchList != nil && len(*input.BranchList) > 0 {
		branchListJSON, err := json.Marshal(input.BranchList)
		if err != nil {
			return menu, fmt.Errorf("failed to marshal branch list: %w", err)
		}
		raw := json.RawMessage(branchListJSON) // แปลง []byte → RawMessage
		menu.BranchList = raw                  // เอา address ไปเก็บ
	} else {
		menu.BranchList = nil // กำหนดเป็น NULL
	}

	menu.BonusExpireDate = input.BonusExpireDate
	menu.StartDate = input.StartDate
	menu.EndDate = input.EndDate
	menu.CardExpireDate = input.CardExpireDate
	menu.GroupMenuID = input.GroupMenuID
	menu.CreateBy = &userId
	menu.CreateDate = utils.TimeNowAsia()

	if config.DB_POS == nil {
		return menu, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&menu).Error; err != nil {
		return menu, fmt.Errorf("failed to create pos menu: %w", err)
	}

	return menu, nil
}

func FindExistPosMenuName(name string) (*models.PosMenu, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var posMenu models.PosMenu
	lowerName := strings.ToLower(name)

	err := config.DB_POS.
		Where("LOWER(description) = ?", lowerName).
		First(&posMenu).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &posMenu, nil
}

func FindExistPosMenuCode(code string) (*models.PosMenu, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var posMenu models.PosMenu
	lowerCode := strings.ToLower(code)

	err := config.DB_POS.
		Where("LOWER(code) = ?", lowerCode).
		First(&posMenu).Error

	// คืน nil ถ้าไม่พบหรือเกิด error ใดๆ
	if err != nil {
		return nil, err
	}

	return &posMenu, nil
}

func FindPosMenuById(id int) (*models.PosMenu, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var posMenu models.PosMenu

	if err := config.DB_POS.Where("id = ?", id).First(&posMenu).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // ถ้าไม่พบข้อมูล
		}
		return nil, err // ถ้ามีข้อผิดพลาดอื่น
	}

	return &posMenu, nil
}

func FindPosMenuByGroupMenuId(groupMenuId string) ([]models.PosMenu, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var posMenu []models.PosMenu

	result := config.DB_POS.Where("group_menu_id = ? and is_active = true and is_delete = false", groupMenuId).Find(&posMenu)
	if result.Error != nil {
		return nil, result.Error
	}

	// หาก Query ไม่พบอะไร จะได้ array ว่างๆ ไม่ใช่นิล
	return posMenu, nil
}

func FindPosMenuByLocation(location string) (*[]models.PosMenuDataDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
	SELECT 
		pm.id AS menu_id,
		pm.code,
		pm.description AS menu_name,
		COALESCE(pm.price, 0) AS price,
		COALESCE(pm.e_coin, 0) AS e_coin,
		COALESCE(pm.e_bonus, 0) AS e_bonus,
		COALESCE(pm.token, 0) AS token,
		pm.branch_group_id,
		pm.machine_group_id,
		pm.is_active,
		pm.create_date,
		pm.bonus_expire_date,
		COALESCE(pm.limit_time, 0) AS limit_time,
		pm.card_type_id,
		ct.name as card_type_name,
		pm.branch_list,
		pm.start_date,
		pm.end_date,
		pm.card_expire_date,
		pm.group_menu_id,
		mgm.name AS group_menu_name,
		mgm.enable_member AS enable_member,
		mgm.display_name AS display_name,
		mgm.receipt_stub AS receipt_stub,
		pbg.group_name AS group_branch_name,
		pbsg.location_code,
		pm.discount_cash,
		pm.discount_cash_expire_date,
		pm.discount_cash_limit_days,
		pm.bonus_expire_limit_days,
		pm.card_expire_limit_minutes
	FROM pos_menu pm
	INNER JOIN pos_branch_group pbg ON pm.branch_group_id = pbg.id
	INNER JOIN pos_branch_sub_group pbsg ON pbg.id = pbsg.group_id
	LEFT JOIN master_group_menu mgm ON pm.group_menu_id = mgm.id
	LEFT  JOIN card_type ct on pm.card_type_id = ct.id
	WHERE LOWER(pbsg.location_code) = ?
	  AND pm.is_active = true
	  AND pm.is_delete = false
	`

	var result []models.PosMenuDataDto
	tx := config.DB_POS.Raw(query, strings.ToLower(location)).Scan(&result)
	if tx.Error != nil {
		return nil, fmt.Errorf("query error: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return nil, nil // ไม่พบข้อมูล
	}

	fmt.Println(len(result), "rows found for location:", location)

	return &result, nil
}

func FindExistPosMenuNameAndCodeNotCurrent(name string, code string, currentId int) (*models.PosMenu, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var posMenu models.PosMenu
	lowerName := strings.ToLower(name)
	lowerCode := strings.ToLower(code)

	result := config.DB_POS.
		Where("LOWER(description) = ? AND LOWER(code) = ? AND id != ?", lowerName, lowerCode, currentId).
		Find(&posMenu)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &posMenu, nil
}

func FindExistPosMenuCodeNotCurrent(code string, currentId int) (*models.PosMenu, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var posMenu models.PosMenu
	lowerCode := strings.ToLower(code)

	result := config.DB_POS.
		Where("LOWER(code) = ? AND id != ?", lowerCode, currentId).
		Find(&posMenu)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &posMenu, nil
}

func UpdatePosMenu(id int, input models.PosMenuDto, userId int) (models.PosMenu, error) {
	var existing models.PosMenu

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record ที่มีอยู่ก่อน
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("pos menu not found: %w", err)
	}

	// เตรียม map updates
	updates := map[string]interface{}{
		"code":                      input.Code,
		"description":               input.Description,
		"price":                     input.Price,
		"e_coin":                    input.ECoin,
		"e_bonus":                   input.EBonus,
		"branch_group_id":           input.BranchGroupID,
		"machine_group_id":          input.MachineGroupID,
		"limit_time":                input.LimitTime,
		"card_type_id":              input.CardTypeID,
		"group_menu_id":             input.GroupMenuID,
		"is_active":                 input.IsActive,
		"update_by":                 userId,
		"update_date":               utils.TimeNowAsia(),
		"bonus_expire_date":         nil,
		"start_date":                nil,
		"end_date":                  nil,
		"card_expire_date":          nil,
		"branch_list":               nil,
		"discount_cash":             input.DiscountCash,
		"discount_cash_expire_date": nil,
		"discount_cash_limit_days":  input.DiscountCashlimitDays,
		"bonus_expire_limit_days":   input.BonusExpireLimitDays,
		"card_expire_limit_minutes": input.CardExpireLimitMinutes,
	}

	if input.StartDate != nil {
		updates["start_date"] = input.StartDate
	}
	if input.EndDate != nil {
		updates["end_date"] = input.EndDate
	}
	if input.CardExpireDate != nil {
		updates["card_expire_date"] = input.CardExpireDate
	}
	if input.BonusExpireDate != nil {
		updates["bonus_expire_date"] = input.BonusExpireDate
	}
	if input.DiscountCashExpireDate != nil {
		updates["discount_cash_expire_date"] = input.DiscountCashExpireDate
	}

	// branchList
	if input.BranchList != nil && len(*input.BranchList) > 0 {
		branchListJSON, err := json.Marshal(input.BranchList)
		if err != nil {
			return existing, fmt.Errorf("failed to marshal branch list: %w", err)
		}
		updates["branch_list"] = branchListJSON
	}

	// ทำการ update
	if err := config.DB_POS.Model(&existing).Updates(updates).Error; err != nil {
		return existing, fmt.Errorf("failed to update pos menu: %w", err)
	}

	// ดึงอีกครั้งเพื่อตอบกลับ
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated pos menu: %w", err)
	}

	return existing, nil
}

func PosMenuSearchList(params models.SearchPosMenuParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var posMenues []models.PosMenuDataDto

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("pos_menu as pm").
		Select(`pm.id AS menu_id,
		        pm.code,
		        pm.description AS menu_name,
		        COALESCE(pm.price, 0) AS price,
		        COALESCE(pm.e_coin, 0) AS e_coin,
		        COALESCE(pm.e_bonus, 0) AS e_bonus,
		        COALESCE(pm.token, 0) AS token,
		        pm.branch_group_id,
		        pm.machine_group_id,
		        pm.is_active,
		        pm.create_date,
		        pm.bonus_expire_date,
		        COALESCE(pm.limit_time, 0) AS limit_time,
		        pm.card_type_id,
				ct.name as card_type_name,
		        pm.branch_list,
		        pm.start_date,
		        pm.end_date,
		        pm.card_expire_date,
		        pm.group_menu_id,
		        mgm.name AS group_menu_name,
				mgm.enable_member AS enable_member,
				mgm.display_name AS display_name,
				mgm.receipt_stub AS receipt_stub,
		        pbg.group_name AS group_branch_name,
				pm.discount_cash,
				pm.discount_cash_expire_date,
				pm.discount_cash_limit_days,
				pm.bonus_expire_limit_days,
				pm.card_expire_limit_minutes`).
		Joins("INNER JOIN pos_branch_group pbg ON pm.branch_group_id = pbg.id").
		Joins("LEFT  JOIN master_group_menu mgm ON pm.group_menu_id = mgm.id").
		Joins("LEFT  JOIN card_type ct on pm.card_type_id = ct.id").
		Where("pm.is_delete = false")

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("(pm.description ILIKE ? OR pm.code ILIKE ?)", search, search)
	}
	if params.GroupMenuID != "" {
		query = query.Where("pm.group_menu_id = ?", params.GroupMenuID)
	}
	if params.BranchGroupID != nil {
		query = query.Where("pm.branch_group_id = ?", params.BranchGroupID)
	}
	if params.MachineGroupID != nil {
		query = query.Where("pm.machine_group_id = ?", params.MachineGroupID)
	}
	if params.IsActive != nil {
		query = query.Where("pm.is_active = ?", *params.IsActive)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("pm.description ASC").Limit(params.Skip).Offset(offset).Find(&posMenues).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     posMenues,
	}
	return result, nil
}

func DeletePosMenuById(id int, userId int) (models.PosMenu, error) {
	var existing models.PosMenu

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("pos menu not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.PosMenu{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to delete pos menu: %w", err)
	}

	//ดึงข้อมูลใหม่ด้วย Where
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve pos menu: %w", err)
	}

	return existing, nil
}

func PosMenuSaleList(params models.SearchPosMenuSaleParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var posMenues []models.PosMenuDataDto

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("pos_menu AS pm").
		Select(`
		pm.id AS menu_id,
		pm.code,
		pm.description AS menu_name,
		COALESCE(pm.price, 0) AS price,
		COALESCE(pm.e_coin, 0) AS e_coin,
		COALESCE(pm.e_bonus, 0) AS e_bonus,
		COALESCE(pm.token, 0) AS token,
		pm.branch_group_id,
		pm.machine_group_id,
		pm.is_active,
		pm.create_date,
		pm.bonus_expire_date,
		COALESCE(pm.limit_time, 0) AS limit_time,
		pm.card_type_id,
		ct.name AS card_type_name,
		pm.branch_list,
		pm.start_date,
		pm.end_date,
		pm.card_expire_date,
		pm.group_menu_id,
		mgm.name AS group_menu_name,
		pbg.group_name AS group_branch_name,
		mgm.enable_member AS enable_member,
		mgm.display_name AS display_name,
		mgm.receipt_stub AS receipt_stub,
		pbsg.location_code,
		pm.discount_cash,
		pm.discount_cash_expire_date,
		pm.discount_cash_limit_days,
		pm.bonus_expire_limit_days,
		pm.card_expire_limit_minutes
	`).
		Joins("INNER JOIN pos_branch_group pbg ON pm.branch_group_id = pbg.id").
		Joins("INNER JOIN pos_branch_sub_group pbsg ON pbg.id = pbsg.group_id").
		Joins("LEFT JOIN master_group_menu mgm ON pm.group_menu_id = mgm.id").
		Joins("LEFT JOIN card_type ct ON pm.card_type_id = ct.id").
		Where(`
		pm.is_delete = false 
		AND pm.is_active = true 
		AND (
			(pm.start_date IS NOT NULL AND pm.end_date IS NOT NULL AND NOW() BETWEEN pm.start_date AND pm.end_date)
			OR (pm.start_date IS NULL AND pm.end_date IS NULL)
			OR (pm.start_date IS NULL AND pm.end_date IS NOT NULL AND NOW() <= pm.end_date)
			OR (pm.start_date IS NOT NULL AND pm.end_date IS NULL AND NOW() >= pm.start_date)
		)
	`)

	if params.GroupMenuID != "" {
		query = query.Where("pm.group_menu_id = ?", params.GroupMenuID)
	}

	if strings.ToLower(params.Location) != "" {
		query = query.Where("LOWER(pbsg.location_code) = ?", strings.ToLower(params.Location))
	}

	if strings.TrimSpace(params.Search) != "" {
		search := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("pm.description ILIKE ?", search)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("pm.price DESC").Limit(params.Skip).Offset(offset).Find(&posMenues).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	groupMenu, err := FindGroupMenuById(params.GroupMenuID)
	if err != nil {
		return result, fmt.Errorf("failed to find group menu: %w", err)
	}
	if groupMenu != nil && strings.ToLower(groupMenu.Name) == "package" {
		for i := range posMenues {

			machineGroup := posMenues[i].MachineGroupID

			var sum models.MachineSubGroupSum

			err := config.DB_POS.
				Table("machine_sub_group msg").
				Select(`
				COALESCE(SUM(msg.e_coin),0) as e_coin,
				COALESCE(SUM(msg.e_bonus),0) as e_bonus,
				COALESCE(SUM(msg.play_time),0) as play_time
			`).
				Where("msg.group_id = ?", machineGroup).
				Scan(&sum).Error

			if err != nil {
				return result, fmt.Errorf("failed to fetch sub machines for machine group %d: %v", machineGroup, err)
			}

			posMenues[i].ECoin = sum.ECoin
			posMenues[i].EBonus = sum.EBonus
			posMenues[i].LimitTime = &sum.PlayTime
		}
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     posMenues,
	}
	return result, nil
}

func FindPosMenuSaleAll() (*models.PosMenu, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var posMenu models.PosMenu

	err := config.DB_POS.
		Where("group_menu_id IS NULL AND is_active = true AND is_delete = false").
		First(&posMenu).Error

	if err != nil {
		return nil, err
	}

	return &posMenu, nil
}
