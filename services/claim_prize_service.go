package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	// "strings"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	// "new-pos-api/utils"
	// "gorm.io/gorm"
)

func CreateClaimPrize(input models.ClaimPrizeDto) (models.ClaimPrize, error) {
	var claimPrize models.ClaimPrize

	claimPrize.CPDate = *utils.TimeNowAsia()
	claimPrize.MemberTel = input.MemberTel
	claimPrize.CPSpend = input.SpendAmount
	claimPrize.SpendECoin = input.SpendECoin
	claimPrize.SpendEBonus = input.SpendEBonus
	claimPrize.Discount = input.Discount
	claimPrize.ProductPrice = input.ProductPrice
	claimPrize.OptionID = input.OptionID
	claimPrize.CardNo = input.CardNo
	claimPrize.CPUser = input.ClaimUser
	claimPrize.CPLocation = input.Location
	claimPrize.ClaimAmount = input.ClaimAmount

	if config.DB_POS == nil {
		return claimPrize, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&claimPrize).Error; err != nil {
		return claimPrize, fmt.Errorf("failed to create claim prize: %w", err)
	}

	return claimPrize, nil
}

func ClaimJoylicoinSyncHistory(input models.ClaimPrizeDto, id string) (string, error) {
	var historys []models.ScoreHistory

	// ✅ Get customer
	status, customer, err := GetCustomerByMobileNo(strings.TrimSpace(input.MemberTel))
	if err != nil || status != 200 || customer == nil {
		return "", fmt.Errorf("customer not found with tel: %s", input.MemberTel)
	}

	// ✅ Get score types
	status, scoreTypes, err := GetScoreType()
	if err != nil || status != 200 || len(scoreTypes) == 0 {
		return "", fmt.Errorf("score type not found")
	}

	// ✅ Get branch
	status, branch, err := GetBranchByCode(strings.TrimSpace(input.Location))
	if err != nil || status != 200 || branch == nil {
		return "", fmt.Errorf("branch not found with code: %s", input.Location)
	}

	// ✅ Map score types
	scoreTypeMap := make(map[string]int)
	for _, st := range scoreTypes {
		scoreTypeMap[st.Name] = st.ID
	}

	coinID, ok := scoreTypeMap["Coin"]
	if !ok {
		return "", fmt.Errorf("coin score type not found")
	}

	// ✅ Description
	description := "NEW-POS ได้รับ Joylicoin จากตู้ขาย"
	if input.Joylicoin < 0 {
		description = "NEW-POS หัก Joylicoin จากตู้ขาย"
	}

	// ✅ Append data ให้ historys
	historys = append(historys, models.ScoreHistory{
		CustomerID:       customer.ID,
		ScoreTypeID:      coinID,
		BranchID:         branch.ID,
		Amount:           int(input.Joylicoin),
		TransactionDate:  input.CPDate,
		CreateBy:         input.ClaimUser,
		MobileNo:         input.MemberTel,
		RefTransaction:   StringPtr("claim_prize"),
		RefTransactionID: &id,
		PosID:            nil,
		Description:      &description,
	})

	// ✅ Parallel sync
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []string
	var firstErr error

	for _, his := range historys {
		wg.Add(1)
		h := his

		go func() {
			defer wg.Done()

			status, body, err := SyncHistory(h)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("failed to create score history: %s", err.Error())
				}
				return
			}

			if status == 200 && body != nil {

				var prettyJSON bytes.Buffer

				if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
					results = append(results,
						fmt.Sprintf("Score history synced successfully:\n%s", prettyJSON.String()))
				} else {
					results = append(results,
						fmt.Sprintf("Score history synced successfully: %s", string(body)))
				}

			} else {
				if firstErr == nil {
					firstErr = fmt.Errorf("failed to sync score history with status %d", status)
				}
			}

		}()
	}

	wg.Wait()

	if firstErr != nil {
		return "", firstErr
	}

	return strings.Join(results, "\n"), nil
}

func ClaimPrizeSearchList(params models.SearchClaimPrizeParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var claimPrizes []models.ClaimPrizeList
	var response models.ClaimPrizeData
	var actions []models.CliamPrizeByAction

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	// Helper: apply filters
	applyFilters := func(db *gorm.DB) *gorm.DB {
		if params.Location != "" {
			db = db.Where("cp.cp_location ILIKE ?", "%"+params.Location+"%")
		}
		if params.MemberTel != "" {
			db = db.Where("cp.member_tel ILIKE ?", "%"+params.MemberTel+"%")
		}
		if params.StartDate != "" && params.EndDate != "" {
			db = db.Where("cp.cp_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
		}
		return db
	}

	// Helper: NullInt64 -> int64
	nullToInt := func(n sql.NullInt64) int64 {
		if n.Valid {
			return n.Int64
		}
		return 0
	}

	query := applyFilters(
		config.DB_POS.Table("claim_prize cp").
			Select(`cp.cp_date as date,
					cp.member_tel,
					coalesce(cp.cp_spend, 0) as spend_amount,
					coalesce(cp.spend_ecoin, 0) as spend_e_coin,
					coalesce(cp.spend_ebonus, 0) as spend_e_bonus,
					coalesce(cp.discount, 0) as discount,
					coalesce(cp.product_price, 0) as product_price,
					coalesce(cp.card_no, '') as card_no,
					cp.cp_user as cashier,
					cp.cp_location as location,
					coalesce(cp.claim_amount, 0) as claim_amount,
					ma.action_name`).
			Joins("LEFT JOIN master_action ma ON cp.option_id = ma.id"),
	)

	sumQuery := applyFilters(
		config.DB_POS.Table("claim_prize cp").
			Select(`sum(coalesce(cp.cp_spend, 0)) as total_spending,
					sum(coalesce(cp.spend_ecoin, 0)) as total_ecoin,
					sum(coalesce(cp.spend_ebonus, 0)) as total_ebonus,
					sum(coalesce(cp.product_price, 0)) as total_product_price`),
	)

	actionQuery := applyFilters(
		config.DB_POS.Table("claim_prize cp").
			Select("ma.id as action_id, ma.action_name as action_name, sum(coalesce(cp.claim_amount, 0)) as total_claim").
			Joins("LEFT JOIN master_action ma ON cp.option_id = ma.id").
			Group("ma.id, ma.action_name"),
	)

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Sum
	var totalSpending, totalEcoin, totalEBonus, totalProductPrice sql.NullInt64
	if err := sumQuery.Row().Scan(&totalSpending, &totalEcoin, &totalEBonus, &totalProductPrice); err != nil {
		return result, fmt.Errorf("sum query failed: %w", err)
	}

	// Group by action
	if err := actionQuery.Scan(&actions).Error; err != nil {
		return result, fmt.Errorf("group by action query failed: %w", err)
	}

	// Paginate
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("cp.cp_date DESC").Limit(params.Skip).Offset(offset).Find(&claimPrizes).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	response.ClaimPrizeList = claimPrizes
	response.ClaimPrizeTotal.TotalSpending = nullToInt(totalSpending)
	response.ClaimPrizeTotal.TotalECoin = nullToInt(totalEcoin)
	response.ClaimPrizeTotal.TotalEBonus = nullToInt(totalEBonus)
	response.ClaimPrizeTotal.TotalProductPrice = nullToInt(totalProductPrice)
	response.ClaimPrizeTotal.Actions = actions

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     response,
	}
	return result, nil
}

func ExportClaimPrizeSearch(params models.ExportClaimPrizeParams) ([]models.ExportClaimPrizeList, error) {
	var claimPrizes []models.ExportClaimPrizeList

	if config.DB_POS == nil {
		return claimPrizes, fmt.Errorf("database connection is nil")
	}

	// Helper: apply filters
	applyFilters := func(db *gorm.DB) *gorm.DB {
		if params.Location != "" {
			db = db.Where("cp.cp_location ILIKE ?", "%"+params.Location+"%")
		}
		if params.MemberTel != "" {
			db = db.Where("cp.member_tel ILIKE ?", "%"+params.MemberTel+"%")
		}
		if params.StartDate != "" && params.EndDate != "" {
			db = db.Where("cp.cp_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
		}
		return db
	}

	query := applyFilters(
		config.DB_POS.Table("claim_prize cp").
			Select(`cp.cp_date as date,
					cp.member_tel,
					coalesce(cp.cp_spend, 0) as spend_amount,
					coalesce(cp.spend_ecoin, 0) as spend_e_coin,
					coalesce(cp.spend_ebonus, 0) as spend_e_bonus,
					coalesce(cp.discount, 0) as discount,
					coalesce(cp.product_price, 0) as product_price,
					coalesce(cp.card_no, '') as card_no,
					cp.cp_user as cashier,
					cp.cp_location as location,
					coalesce(cp.claim_amount, 0) as claim_amount,
					ma.action_name,
					m.id as member_id`).
			Joins("LEFT JOIN master_action ma ON cp.option_id = ma.id").
			Joins("LEFT JOIN member m on cp.member_tel = m.tel"),
	)

	if err := query.Order("cp.cp_date DESC").Find(&claimPrizes).Error; err != nil {
		return claimPrizes, fmt.Errorf("data query failed: %w", err)
	}

	return claimPrizes, nil
}

func ExportClaimPrizeSearchToExcel(c *gin.Context, data []models.ExportClaimPrizeList) error {
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
	memberID := 0
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
		headers := []interface{}{"วันที่ทำรายการ", "Card", "Member ID", "ยอดใช้จ่าย", "ecoin", "ebonus", "มูลค่าสินค้า", "Option", "ยอดเคลม", "ผู้ทำรายการ"}
		if err := sw.SetRow("A1", headers, excelize.RowOpts{StyleID: boldCenterStyle}); err != nil {
			return err
		}

		// เขียน row ข้อมูล
		for i, row := range data[start:end] {
			rowNum := i + 2
			if row.MemberID != nil {
				memberID = *row.MemberID
			}
			if row.ClaimAmount < 0 {
				row.ClaimAmount = row.ClaimAmount * -1
			}
			rowCells := []interface{}{
				row.Date,
				strPtr(row.CardNo),
				memberID,
				row.SpendAmount,
				row.SpendECoin,
				row.SpendEBonus,
				row.ProductPrice,
				row.ActionName,
				row.ClaimAmount,
				row.Cashier,
			}

			cell, _ := excelize.CoordinatesToCellName(1, rowNum)
			if err := sw.SetRow(cell, rowCells); err != nil {
				return err
			}

			// ตั้ง style ให้วันที่และคะแนน
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("B%d", rowNum), dateTimeStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("D%d", rowNum), fmt.Sprintf("D%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("E%d", rowNum), fmt.Sprintf("E%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("F%d", rowNum), fmt.Sprintf("F%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("G%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("I%d", rowNum), fmt.Sprintf("I%d", rowNum), numberStyle)
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

	c.Header("Content-Disposition", "attachment; filename=claim_prize_report.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())

	return nil
}

func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
