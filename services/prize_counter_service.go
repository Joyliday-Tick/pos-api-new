package services

import (
	"fmt"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"
	"strings"
)

func PrizeCounterJubuJibiList(params models.SearchPrizeCounterParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var prizeRecords []models.PrizeCounterData
	var date = utils.TimeNowAsia()
	var dateString = date.Format("2006-01-02") // yyyy-mm-dd

	if config.DB_JREADER == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	query := config.DB_JREADER.Table("prizecounter_record pr").
		Select(`TO_CHAR(pr.pz_date, 'YYYY-MM-DD') as date, 
    ma.asset_id ,
    ma.machine_asset,
    md.machine_name,
    ma.machine_no,
    mc.mc_head,
    ei.esi_code as sku, 
    ei.esi_description as product_name, 
    ei.esi_price as price, 
    ei.esi_coin as joylicoin,
    pr.pz_status as status,
    pr.pz_id as id,
    ei.esi_id as product_id,
    pr.pz_member as member_tel,
    pr.pz_asset_location as location`).
		Joins("LEFT JOIN machine_config mc ON mc.masset_id = pr.pz_asset_id AND pr.pz_asset_head = mc.mc_head").
		Joins("LEFT JOIN machine_asset ma ON ma.asset_id = mc.masset_id").
		Joins("LEFT JOIN easywin_item ei ON pr.pz_itemid = ei.esi_id").
		Joins("LEFT JOIN machine_data md ON md.mc_id = ma.machine_id").
		Where("pr.pz_status IN ('Y','T')")

	if params.MemberTel != "" {
		query = query.Where("pr.pz_member = ?", strings.TrimSpace(params.MemberTel))
	}
	if params.Location != "" {
		query = query.Where("LOWER(pr.pz_asset_location) = ?", strings.ToLower(params.Location))
	}
	if dateString != "" {
		query = query.Where("pr.pz_date::date = ?", dateString)
	}
	fmt.Println("Query:", query)
	// query = query.Group("TO_CHAR(mr.rc_date, 'YYYY-MM-DD'), ma.machine_asset, md.machine_name")

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Run query
	if err := query.Order("pr.pz_id").Find(&prizeRecords).Error; err != nil {
		return result, fmt.Errorf("query execution failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       1,
		TotalCount: int(totalCount),
		Result:     prizeRecords,
	}
	return result, nil
}

func UpdatePrizeCounterJubuJibi(memberTel string) error {
	if config.DB_JREADER == nil {
		return fmt.Errorf("database connection is nil")
	}

	var date = utils.TimeNowAsia()
	var dateString = date.Format("2006-01-02") // yyyy-mm-dd

	// ทำการอัปเดตข้อมูล stamp_point
	result := config.DB_JREADER.Exec(`
		UPDATE prizecounter_record 
		SET pz_status = 'N'
		WHERE pz_member = ? AND pz_status IN ('Y','T') AND pz_date::date = ?`,
		memberTel, dateString)

	if result.Error != nil {
		return fmt.Errorf("failed to prize counter jubu jibi: %w", result.Error)
	}

	return nil
}
