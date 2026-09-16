package services

import (
	"bytes"
	"fmt"
	"net/http"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func MeterRecordSearchList(params models.SearchMeterRecordParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var meterRecords []models.MeterRecordData
	var dateString string

	if config.DB_JREADER == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	setting, err := GetBonusSetting()
	if err != nil {
		return result, fmt.Errorf("failed to retrieve bonus setting: %w", err)
	}
	if setting != nil {
		dateString = setting.SetDate.Format("2006-01-02")
		fmt.Println("Formatted date:", dateString)

	}

	query := config.DB_JREADER.Table("meter_record mr").
		Select(`TO_CHAR(mr.rc_date, 'YYYY-MM-DD') AS date,
			ma.machine_asset,
			md.machine_name,
			MAX(mr.rc_member) AS member_tel,
			MAX(mr.rc_card_id) AS card_no,
			MAX(mr.rc_asset_location) AS location,
			SUM(mr.rc_ecoin) AS rc_ecoin,
			SUM(mr.rc_bonus) AS rc_bonus,
			SUM(mr.rn_bonus) AS rn_bonus,
			SUM(mr.rn_ecoin) AS rn_ecoin`).
		Joins("LEFT JOIN machine_asset ma ON ma.asset_id = mr.rc_asset_id").
		Joins("LEFT JOIN machine_data md ON md.mc_id = ma.machine_id").
		Joins("LEFT JOIN machine_category mc ON mc.cat_id = md.category_id").
		Where("mc.cat_id = 1 AND mr.bonus_status = 'Y'")

	if params.MemberTel != "" {
		query = query.Where("mr.rc_member = ?", strings.TrimSpace(params.MemberTel))
	}
	if dateString != "" {
		query = query.Where("mr.rc_date::date >= ?", dateString)
	}

	query = query.Group("TO_CHAR(mr.rc_date, 'YYYY-MM-DD'), ma.machine_asset, md.machine_name")

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Run query
	if err := query.Order("TO_CHAR(mr.rc_date, 'YYYY-MM-DD') DESC").Find(&meterRecords).Error; err != nil {
		return result, fmt.Errorf("query execution failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       1,
		TotalCount: int(totalCount),
		Result:     meterRecords,
	}
	return result, nil
}

func UpdateMeterRecordClaim(memberTel string) error {
	if config.DB_JREADER == nil {
		return fmt.Errorf("database connection is nil")
	}

	// ดึงวันที่จากการตั้งค่า
	setting, err := GetBonusSetting()
	if err != nil {
		return fmt.Errorf("failed to retrieve bonus setting: %w", err)
	}

	if setting == nil {
		return fmt.Errorf("bonus setting not found")
	}

	dateString := setting.SetDate.Format("2006-01-02") // yyyy-mm-dd

	// ทำการอัปเดตข้อมูล stamp_point
	result := config.DB_JREADER.Exec(`
		UPDATE meter_record 
		SET bonus_status = 'N'
		WHERE rc_member = ? AND bonus_status = 'Y' AND rc_date::date >= ?`,
		memberTel, dateString)

	if result.Error != nil {
		return fmt.Errorf("failed to update meter record claim: %w", result.Error)
	}

	return nil
}
func SumMeterRecordJubuJibi(params models.SearchPrizeCounterParams) (models.MeterRecordSumEcoin, error) {
	// var result utils.SearchResult
	var meterRecords models.MeterRecordSumEcoin
	var date = utils.TimeNowAsia()
	var dateString = date.Format("2006-01-02") // yyyy-mm-dd

	if config.DB_JREADER == nil {
		return meterRecords, fmt.Errorf("database connection is nil")
	}

	query := config.DB_JREADER.Table("meter_record mr").
		Select(`SUM(mr.rc_ecoin) as used_ecoin,
            SUM(mr.rc_bonus) as used_bonus`).
		Joins("LEFT JOIN machine_asset ma ON ma.asset_id = mr.rc_asset_id").
		Joins("LEFT JOIN machine_data md ON ma.machine_id = md.mc_id").
		Joins("LEFT JOIN machine_category mc ON mc.cat_id = md.category_id").
		Where("mc.cat_id = 2 AND mr.prize_status = 'Y'")

	if params.MemberTel != "" {
		query = query.Where("mr.rc_member = ?", strings.TrimSpace(params.MemberTel))
	}
	if params.Location != "" {
		query = query.Where("LOWER(mr.rc_asset_location) = ?", strings.ToLower(params.Location))
	}
	if dateString != "" {
		query = query.Where("mr.rc_date::date = ?", dateString)
	}

	// query = query.Group("TO_CHAR(mr.rc_date, 'YYYY-MM-DD'), ma.machine_asset, md.machine_name")

	// Run query
	if err := query.Find(&meterRecords).Error; err != nil {
		return meterRecords, fmt.Errorf("query execution failed: %w", err)
	}

	// result = utils.SearchResult{
	// 	Page:       1,
	// 	TotalCount: 1,
	// 	Result:     meterRecords,
	// }
	return meterRecords, nil
}

func UpdateMeterRecordJubuJibi(memberTel string) error {
	if config.DB_JREADER == nil {
		return fmt.Errorf("database connection is nil")
	}

	var date = utils.TimeNowAsia()
	var dateString = date.Format("2006-01-02") // yyyy-mm-dd

	// ทำการอัปเดตข้อมูล stamp_point
	result := config.DB_JREADER.Exec(`
		UPDATE meter_record 
		SET prize_status = 'N'
		WHERE rc_member = ? AND prize_status = 'Y' AND rc_date::date = ?`,
		memberTel, dateString)

	if result.Error != nil {
		return fmt.Errorf("failed to update meter record jubu jibi: %w", result.Error)
	}

	return nil
}

func MeterRecordJubuJibiList(params models.SearchPrizeCounterParams) ([]models.MeterRecordJubuJibi, error) {
	var meterRecords []models.MeterRecordJubuJibi
	date := utils.TimeNowAsia()
	dateString := date.Format("2006-01-02") // yyyy-mm-dd

	if config.DB_JREADER == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := config.DB_JREADER.Table("meter_record mr").
		Select(`TO_CHAR(mr.rc_date, 'YYYY-MM-DD') AS date,
			ma.machine_asset,
			md.machine_name,
			mc.category as category,
			ma.machine_no,
			mr.rc_asset_head as asset_head,
			mr.rc_member AS member_tel,
			mr.rc_card_id AS card_no,
			mr.rc_asset_location AS location,
			mr.rc_ecoin AS rc_ecoin,
			mr.rc_bonus AS rc_bonus,
			mr.rn_bonus AS rn_bonus,
			mr.rn_ecoin AS rn_ecoin`).
		Joins("LEFT JOIN machine_asset ma ON ma.asset_id = mr.rc_asset_id").
		Joins("LEFT JOIN machine_data md ON ma.machine_id = md.mc_id").
		Joins("LEFT JOIN machine_category mc ON mc.cat_id = md.category_id").
		Where("mc.cat_id = 2 AND mr.prize_status = 'Y'")

	if params.MemberTel != "" {
		query = query.Where("mr.rc_member = ?", strings.TrimSpace(params.MemberTel))
	}
	if params.Location != "" {
		query = query.Where("LOWER(mr.rc_asset_location) = ?", strings.ToLower(params.Location))
	}
	// เงื่อนไขวันที่
	if dateString != "" {
		query = query.Where("mr.rc_date::date = ?", dateString)
	}

	// Execute
	if err := query.Find(&meterRecords).Error; err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return meterRecords, nil
}

func CheckCardList(params models.SearchCheckCard) ([]models.MeterRecordCheckCard, error) {
	var meterRecords []models.MeterRecordCheckCard

	if config.DB_JREADER == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := config.DB_JREADER.Table("meter_record mr").
		Select(`mr.rc_date AS date,
    			ma.machine_asset AS machine_code,
    			md.machine_name,
    			(ma.machine_no || '-' || mr.rc_asset_head) AS machine_no,
    			mr.rc_asset_location AS location,
    			mr.rc_member AS member_tel,
				mr.rc_card_id as crad_no,
    			mr.rc_ecoin AS rc_ecoin,
    			mr.rc_bonus AS rc_bonus,
    			mr.rn_bonus AS rn_bonus,
    			mr.rn_ecoin AS rn_ecoin,
    			mr.rc_time,
    			mr.rn_time,
				mr.card_type
				`).
		Joins("LEFT JOIN machine_asset ma ON ma.asset_id = mr.rc_asset_id").
		Joins("LEFT JOIN machine_data md ON ma.machine_id = md.mc_id")

	if params.CardNo != "" {
		query = query.Where("LOWER(mr.rc_card_id) = ?", strings.ToLower(params.CardNo))
	}

	offset := 0
	skip := 200

	if err := query.Order("mr.rc_date DESC").Limit(skip).Offset(offset).Find(&meterRecords).Error; err != nil {
		return meterRecords, fmt.Errorf("data query failed: %w", err)
	}

	return meterRecords, nil
}

func ExportCheckCardToExcel(c *gin.Context, data []models.MeterRecordCheckCard) error {
	f := excelize.NewFile()

	// สร้าง style
	boldCenterStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	numberFormat := "#,##0"
	numberStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &numberFormat,
	})

	maxRowsPerSheet := 1000000 // sheet ละ 1,000,000 row
	sheetIndex := 1

	for start := 0; start < len(data); start += maxRowsPerSheet {
		end := start + maxRowsPerSheet
		if end > len(data) {
			end = len(data)
		}

		sheetName := fmt.Sprintf("Sheet%d", sheetIndex)
		if sheetIndex == 1 {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			f.NewSheet(sheetName)
		}
		sheetIndex++

		sw, err := f.NewStreamWriter(sheetName)
		if err != nil {
			return err
		}

		// เขียน header
		headers := []interface{}{"Date", "Assetcode", "Machine name", "No.", "Location", "Member", "ecoin", "ebonus", "ecoin เหลือ", "ebonus เหลือ", "Time เหลือ", "Card Type"}
		if err := sw.SetRow("A1", headers, excelize.RowOpts{StyleID: boldCenterStyle}); err != nil {
			return err
		}

		// เขียน row ข้อมูล
		for i, row := range data[start:end] {
			rowNum := i + 2
			rowCells := []interface{}{
				row.Date.Format("2006-01-02 15:04"),
				row.MachineCode,
				row.MachineName,
				row.MachineNo,
				row.Location,
				utils.MaskPhone(row.MemberTel),
				row.RCEcoin,
				row.RCBonus,
				row.RNEcoin,
				row.RNBonus,
				row.RNTime,
				row.CardType,
			}

			cell, _ := excelize.CoordinatesToCellName(1, rowNum)
			if err := sw.SetRow(cell, rowCells); err != nil {
				return err
			}

			f.SetCellStyle(sheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("G%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("H%d", rowNum), fmt.Sprintf("H%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("I%d", rowNum), fmt.Sprintf("I%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("J%d", rowNum), fmt.Sprintf("J%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("K%d", rowNum), fmt.Sprintf("K%d", rowNum), numberStyle)
		}

		if err := sw.Flush(); err != nil {
			return err
		}
	}

	// เขียนลง buffer และส่ง response
	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		return err
	}

	c.Header("Content-Disposition", "attachment; filename=check_card.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())

	return nil
}

func MeterRecordAllSpend(params models.SearchMeterParams) ([]models.MeterRecordAllSpend, error) {
	var meterRecords []models.MeterRecordAllSpend
	date := utils.TimeNowAsia()
	dateString := date.Format("2006-01-02") // yyyy-mm-dd

	if config.DB_JREADER == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := config.DB_JREADER.Table("meter_record mr").
		Select(`mr.rc_id as id,
			mr.rc_date AS date,
			ma.machine_asset,
			md.machine_name,
			mc.category as category,
			ma.machine_no,
			mr.rc_asset_head as asset_head,
			mr.rc_member AS member_tel,
			mr.rc_card_id AS card_no,
			mr.rc_asset_location AS location,
			mr.rc_ecoin AS rc_ecoin,
			mr.rc_bonus AS rc_bonus,
			mr.rn_bonus AS rn_bonus,
			mr.rn_ecoin AS rn_ecoin`).
		Joins("LEFT JOIN machine_asset ma ON ma.asset_id = mr.rc_asset_id").
		Joins("LEFT JOIN machine_data md ON ma.machine_id = md.mc_id").
		Joins("LEFT JOIN machine_category mc ON mc.cat_id = md.category_id").
		Where("mc.cat_id = 2 AND mr.prize_status = 'Y'")

	if params.MemberTel != "" {
		query = query.Where("mr.rc_member = ?", strings.TrimSpace(params.MemberTel))
	}

	// เงื่อนไขวันที่
	if dateString != "" {
		query = query.Where("mr.rc_date::date = ?", dateString)
	}

	// Execute
	if err := query.Find(&meterRecords).Error; err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return meterRecords, nil
}
func UpdatePrizeStatusByIDs(mtIDs []string) error {
	if config.DB_JREADER == nil {
		return fmt.Errorf("database connection is nil")
	}

	if len(mtIDs) == 0 {
		return fmt.Errorf("no meter record IDs provided")
	}

	// แปลง []string เป็นรูปแบบ ('id1','id2','id3')
	placeholders := make([]string, len(mtIDs))
	args := make([]interface{}, len(mtIDs))
	for i, id := range mtIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		UPDATE meter_record 
		SET prize_status = 'N'
		WHERE rc_id IN (%s)
	`, strings.Join(placeholders, ","))

	// รัน query
	tx := config.DB_JREADER.Exec(query, args...)
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("no rows were updated")
	}

	return nil
}

func MeterRecordClaimDetail(params models.SearchClaimDetail) ([]models.ClaimDetail, error) {
	var meterRecords []models.ClaimDetail

	if config.DB_JREADER == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := config.DB_JREADER.Table("meter_record mr").
		Select(`ma.machine_asset as asset_code ,
				md.machine_name as machine_name ,
				ma.machine_no ,
				mr.rc_asset_location as location,		
				sum(mr.rc_ecoin) as e_coin,
				sum(mr.rc_bonus) as e_bonus`).
		Joins("LEFT JOIN machine_asset ma on mr.rc_asset_id = ma.asset_id").
		Joins("LEFT JOIN machine_data md on ma.machine_id = md.mc_id").
		Where("mr.bonus_status = 'N'")

	if params.MemberTel != "" {
		query = query.Where("mr.rc_member = ?", strings.TrimSpace(params.MemberTel))
	}
	// if params.Location != "" {
	// 	query = query.Where("LOWER(mr.rc_asset_location) = ?", strings.ToLower(params.Location))
	// }
	// เงื่อนไขวันที่
	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("mr.rc_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	query = query.
		Group("ma.machine_asset,md.machine_name, ma.machine_no, mr.rc_asset_location").
		Order("ma.machine_asset ASC")

	// Execute
	if err := query.Find(&meterRecords).Error; err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return meterRecords, nil
}
