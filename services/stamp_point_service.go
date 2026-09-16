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
	"gorm.io/gorm"
)

func StampPointSearchList(params models.SearchMeterRecordParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var stampHouses []models.StampHouseData
	var dateString string

	if config.DB_POS == nil {
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

	query := config.DB_POS.Table("stamp_point sp").
		Select(`sp.id ,
				sp.st_tel as member_tel,
				sp.stamp_count,
				sp.stamp_point as e_stamp,
				sp.st_status as status,
				sp.st_date as date, 
				sp.st_id as machine,
				sp.st_branch as location,
				sp.st_sync as sync,
				sp.last_estamp`).
		Where("sp.st_status = 'Y'")

	if params.MemberTel != "" {
		query = query.Where("sp.st_tel = ?", strings.TrimSpace(params.MemberTel))
	}
	if dateString != "" {
		query = query.Where("sp.st_date::date >= ?", dateString)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Run query
	if err := query.Order("sp.st_date DESC").Find(&stampHouses).Error; err != nil {
		return result, fmt.Errorf("query execution failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       1,
		TotalCount: int(totalCount),
		Result:     stampHouses,
	}
	return result, nil
}

func UpdateStampPointClaim(memberTel string) error {
	if config.DB_POS == nil {
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
	result := config.DB_POS.Exec(`
		UPDATE stamp_point 
		SET st_status = 'N'
		WHERE st_tel = ? AND st_status = 'Y' AND st_date::date >= ?`,
		memberTel, dateString)

	if result.Error != nil {
		return fmt.Errorf("failed to update stamp point claim: %w", result.Error)
	}

	return nil
}

func StampHouseListReport(params models.SearchStampHouseReport) (utils.SearchResult, error) {
	var result utils.SearchResult
	var transactions []models.StampHouseDataList
	var stampHouseReport models.StampHouseReport

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("stamp_point sp").
		Select(`sp.st_tel as member_tel,
				sp.stamp_count ,
				sp.stamp_point as e_stamp,
				sp.st_date as date,
				sp.st_id as machine_code,
				sm.name as machine_name,
				sp.st_branch as location,
				m.m_name as firstname,
				m.s_name as lastname,
				m.id as member_id`).
		Joins("LEFT JOIN stamp_machine sm on sp.st_id = sm.code").
		Joins("LEFT JOIN member m on sp.st_tel = m.tel")

	if params.MemberTel != "" {
		tel := "%" + params.MemberTel + "%"
		query = query.Where("(sp.st_tel ILIKE ?)", tel)
	}
	if params.Location != "" {
		query = query.Where("(sp.st_branch = ?)", params.Location)
	}
	if params.Machine != "" {
		query = query.Where("(sp.st_id = ?)", params.Machine)
	}
	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("sp.st_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	var totalStamp int64
	if err := query.Session(&gorm.Session{}).
		Select("COALESCE(SUM(sp.stamp_point), 0)").Scan(&totalStamp).Error; err != nil {
		return result, fmt.Errorf("failed to calculate total stamp points: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("sp.st_date ASC").Limit(params.Skip).Offset(offset).Find(&transactions).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}
	stampHouseReport.TotalEStamp = totalStamp
	stampHouseReport.Data = transactions

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     stampHouseReport,
	}
	return result, nil
}
func ExportStampHouseListReport(params models.SearchExportStampHouseReport) ([]models.StampHouseDataList, error) {
	var transactions []models.StampHouseDataList

	if config.DB_POS == nil {
		return transactions, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("stamp_point sp").
		Select(`sp.st_tel as member_tel,
				sp.stamp_count ,
				sp.stamp_point as e_stamp,
				sp.st_date as date,
				sp.st_id as machine_code,
				sm.name as machine_name,
				sp.st_branch as location,
				m.m_name as firstname,
				m.s_name as lastname,
				m.id as member_id`).
		Joins("LEFT JOIN stamp_machine sm on sp.st_id = sm.code").
		Joins("LEFT JOIN member m on sp.st_tel = m.tel")

	if params.MemberTel != "" {
		tel := "%" + params.MemberTel + "%"
		query = query.Where("(sp.st_tel ILIKE ?)", tel)
	}
	if params.Location != "" {
		query = query.Where("(sp.st_branch = ?)", params.Location)
	}
	if params.Machine != "" {
		query = query.Where("(sp.st_id = ?)", params.Machine)
	}
	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("sp.st_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	if err := query.Order("sp.st_date ASC").Find(&transactions).Error; err != nil {
		return transactions, fmt.Errorf("data query failed: %w", err)
	}

	return transactions, nil
}

func ExportStampHouseToExcel(c *gin.Context, data []models.StampHouseDataList, params models.SearchExportStampHouseReport) error {
	f := excelize.NewFile()

	// สร้าง style
	boldCenterStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	dateFormat := "dd/mm/yyyy HH:mm"
	numberFormat := "#,##0"

	dateTimeStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &dateFormat,
	})
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
		headers := []interface{}{"วันที่", "Member Id", "ชื่อ", "นามสกุล", "คะแนน", "รหัสเครื่อง", "ชื่อเครื่อง", "สาขา"}
		if err := sw.SetRow("A1", headers, excelize.RowOpts{StyleID: boldCenterStyle}); err != nil {
			return err
		}

		// เขียน row ข้อมูล
		for i, row := range data[start:end] {
			rowNum := i + 2
			rowCells := []interface{}{
				row.Date,
				row.MemberID,
				row.Firstname,
				row.Lastname,
				row.EStamp,
				row.MachineCode,
				row.MachineName,
				row.Location,
			}

			cell, _ := excelize.CoordinatesToCellName(1, rowNum)
			if err := sw.SetRow(cell, rowCells); err != nil {
				return err
			}

			// ตั้ง style ให้วันที่และคะแนน
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("A%d", rowNum), dateTimeStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("E%d", rowNum), fmt.Sprintf("E%d", rowNum), numberStyle)
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

	c.Header("Content-Disposition", "attachment; filename=stamp_house_report.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())

	return nil
}
