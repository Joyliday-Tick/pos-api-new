package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"

	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func CreateReturnBonus(input models.ReturnBonusDto) (models.ReturnBonus, error) {
	var entity models.ReturnBonus

	// Map fields จาก DTO ไปยัง Entity
	entity.BonusNo = input.BonusNo
	entity.BonusDate = *utils.TimeNowAsia()
	entity.MemberTel = input.MemberTel
	entity.Branch = input.Branch
	entity.SetBonus = input.SetBonus
	entity.SetPointRate = input.SetPointRate
	entity.CSpend = input.CSpend
	entity.PointStamp = input.PointStamp
	entity.BonusReturn = input.BonusReturn
	entity.Employee = input.Employee
	entity.LastPoint = input.LastPoint
	entity.LastJoylicoin = input.LastJoylicoin
	entity.Extra = input.Extra
	entity.EstampBaht = input.EstampBaht
	entity.LastEstamp = input.LastEstamp

	docNo, err := GetReturnBonusDocumentNo(input.PosId)
	if err != nil {
		return entity, fmt.Errorf("failed to generate document no: %w", err)
	}
	fmt.Println("Document No:", docNo)
	entity.BonusNo = docNo

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create return bonus: %w", err)
	}

	return entity, nil
}

func GetReturnBonusDocumentNo(posID string) (string, error) {
	now := time.Now()
	yy := now.Format("06") // Last 2 digits of year
	mm := now.Format("01") // 2-digit month

	if config.DB_POS == nil {
		return "", fmt.Errorf("database POS connection is nil")
	}

	query := `
		SELECT COUNT(*) AS count
		FROM return_bonus
		WHERE DATE_TRUNC('month', bonus_date) = DATE_TRUNC('month', CURRENT_DATE)
	`

	var count int
	err := config.DB_POS.Raw(query).Scan(&count).Error
	if err != nil {
		return "", fmt.Errorf("query count failed: %w", err)
	}

	running := fmt.Sprintf("%05d", count+1)
	return fmt.Sprintf("JC%s-%s%s%s", posID, yy, mm, running), nil
}

func ReturnBonusSyncHistory(input models.ReturnBonusDto, id string) (string, error) {
	var historys []models.ScoreHistory

	status, customer, err := GetCustomerByMobileNo(strings.TrimSpace(input.MemberTel))
	if err != nil || status != 200 || customer == nil {
		return "", fmt.Errorf("customer not found with tel: %s", input.MemberTel)
	}

	status, scoreTypes, err := GetScoreType()
	if err != nil || status != 200 || len(scoreTypes) == 0 {
		return "", fmt.Errorf("score type not found")
	}

	status, branch, err := GetBranchByCode(strings.TrimSpace(input.Branch))
	if err != nil || status != 200 || branch == nil {
		return "", fmt.Errorf("branch not found with code: %s", input.Branch)
	}

	// Map score types
	scoreTypeMap := make(map[string]int)
	for _, st := range scoreTypes {
		scoreTypeMap[st.Name] = st.ID
	}

	pointID, ok := scoreTypeMap["Point"]
	if !ok {
		return "", fmt.Errorf("point score type not found")
	}

	coinID, ok := scoreTypeMap["Coin"]
	if !ok {
		return "", fmt.Errorf("coin score type not found")
	}

	stampID, ok := scoreTypeMap["E-Stamp"]
	if !ok {
		return "", fmt.Errorf("e-stamp score type not found")
	}

	// Create history entries
	if input.PointStamp != nil && *input.PointStamp > 0 {
		historys = append(historys, models.ScoreHistory{
			CustomerID:       customer.ID,
			ScoreTypeID:      pointID,
			BranchID:         branch.ID,
			Amount:           int(*input.PointStamp),
			TransactionDate:  input.BonusDate,
			Description:      StringPtr("NEW-POS Claim"),
			CreateBy:         input.Employee,
			MobileNo:         input.MemberTel,
			RefTransaction:   StringPtr("return_bonus"),
			RefTransactionID: &id,
			PosID:            StringPtr(input.PosId),
		})
	}

	if input.BonusReturn != nil && *input.BonusReturn > 0 {
		historys = append(historys, models.ScoreHistory{
			CustomerID:       customer.ID,
			ScoreTypeID:      coinID,
			BranchID:         branch.ID,
			Amount:           int(*input.BonusReturn),
			TransactionDate:  input.BonusDate,
			Description:      StringPtr("NEW-POS Claim"),
			CreateBy:         input.Employee,
			MobileNo:         input.MemberTel,
			RefTransaction:   StringPtr("return_bonus"),
			RefTransactionID: &id,
			PosID:            StringPtr(input.PosId),
		})
	}

	if input.LastEstamp != 0 {
		historys = append(historys, models.ScoreHistory{
			CustomerID:       customer.ID,
			ScoreTypeID:      stampID,
			BranchID:         branch.ID,
			Amount:           int(input.LastEstamp) * (-1),
			TransactionDate:  input.BonusDate,
			Description:      StringPtr("NEW-POS Claim"),
			CreateBy:         input.Employee,
			MobileNo:         input.MemberTel,
			RefTransaction:   StringPtr("return_bonus"),
			RefTransactionID: &id,
			PosID:            StringPtr(input.PosId),
		})
	}

	if len(historys) == 0 {
		// return "", fmt.Errorf("no score history to sync")
		return "", nil
	}

	// Use goroutines to parallel sync
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []string
	var firstErr error

	for _, his := range historys {
		wg.Add(1)
		h := his // ป้องกัน closure capture issue
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
				// แปลง JSON []byte เป็น JSON ที่สวยงาม
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
					results = append(results, fmt.Sprintf("Score history synced successfully:\n%s", prettyJSON.String()))
				} else {
					// ถ้าแปลงไม่สำเร็จ fallback ใช้ string(body) แทน
					results = append(results, fmt.Sprintf("Score history synced successfully: %s", string(body)))
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

func StringPtr(s string) *string {
	return &s
}

func ReturnBonusSearchList(params models.SearchClaimReport) (utils.SearchResult, error) {
	var result utils.SearchResult
	var transactions []models.ReturnBonusData
	var response models.ClaimData

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	baseQuery := config.DB_POS.Table("return_bonus rb").
		Select(`
			rb.id,
			rb.bonus_no,
			rb.bonus_date,
			rb.member_tel,
			rb.branch,
			rb.c_spend,
			rb.point_stamp,
			rb.bonus_return,
			rb.employee,
			rb.last_point,
			rb.last_joylicoin,
			rb.last_estamp,
			rb.extra,
			rb.estamp_baht
		`)

	// Filters
	if params.StartDate != "" && params.EndDate != "" {
		baseQuery = baseQuery.Where("rb.bonus_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}
	if len(params.Location) > 0 {
		baseQuery = baseQuery.Where("rb.branch IN ?", params.Location)
	}

	if params.MemberTel != "" {
		baseQuery = baseQuery.Where("rb.member_tel LIKE ?", "%"+params.MemberTel+"%")
	}

	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	var totalJoylicoin, totalSpending sql.NullInt64

	sumQuery := config.DB_POS.Table("return_bonus rb")

	// apply filters again
	if params.StartDate != "" && params.EndDate != "" {
		sumQuery = sumQuery.Where("rb.bonus_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}
	if len(params.Location) > 0 {
		sumQuery = sumQuery.Where("rb.branch IN ?", params.Location)
	}

	if err := sumQuery.Select(`
		COALESCE(SUM(rb.bonus_return),0) AS total_joylicoin,
		COALESCE(SUM(rb.c_spend),0) AS total_spending
	`).Row().Scan(&totalJoylicoin, &totalSpending); err != nil {
		return result, fmt.Errorf("sum query failed: %w", err)
	}

	// ---------- 3) PAGINATION ----------
	offset := (params.Page - 1) * params.Skip

	if err := baseQuery.
		Order("rb.bonus_date ASC").
		Limit(params.Skip).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	response.ClaimList = transactions
	cal := (float64(totalJoylicoin.Int64) / float64(totalSpending.Int64)) * 100
	calRounded := math.Round(cal*100) / 100
	response.Percent = calRounded
	response.TotalJoylicoin = int64(totalJoylicoin.Int64)
	response.TotalSpending = int64(totalSpending.Int64)

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     response,
	}

	return result, nil
}

func ExportReturnBonusSearchList(params models.ExportClaimReport) ([]models.ReturnBonusDataExport, error) {

	var transactions []models.ReturnBonusDataExport

	if config.DB_POS == nil {
		return transactions, fmt.Errorf("database connection is nil")
	}

	baseQuery := config.DB_POS.Table("return_bonus rb").
		Select(`
			rb.id,
			rb.bonus_no,
			rb.bonus_date,
			rb.member_tel,
			rb.branch,
			rb.c_spend,
			rb.point_stamp,
			rb.bonus_return,
			rb.employee,
			rb.last_point,
			rb.last_joylicoin,
			rb.last_estamp,
			rb.extra,
			rb.estamp_baht,
			m.id as member_id
		`).Joins("left join member m on rb.member_tel = m.tel")

	// Filters
	if params.StartDate != "" && params.EndDate != "" {
		baseQuery = baseQuery.Where("rb.bonus_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}
	if len(params.Location) > 0 {
		baseQuery = baseQuery.Where("rb.branch IN ?", params.Location)
	}
	if params.MemberTel != "" {
		baseQuery = baseQuery.Where("rb.member_tel LIKE ?", "%"+params.MemberTel+"%")
	}

	if err := baseQuery.
		Order("rb.bonus_date ASC").
		Find(&transactions).Error; err != nil {
		return transactions, fmt.Errorf("data query failed: %w", err)
	}

	return transactions, nil
}

func ExportClaimToExcel(c *gin.Context, data []models.ReturnBonusDataExport) error {
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
		headers := []interface{}{"เลขที่", "วันที่", "MemberId", "สาขา", "ยอดใช้จ่าย", "estamp", "estamp บาท", "extra", "JC"}
		if err := sw.SetRow("A1", headers, excelize.RowOpts{StyleID: boldCenterStyle}); err != nil {
			return err
		}

		// เขียน row ข้อมูล
		for i, row := range data[start:end] {
			rowNum := i + 2
			rowCells := []interface{}{
				row.BonusNo,
				row.BonusDate,
				row.MemberId,
				row.Branch,
				ptrInt64(row.CSpend),
				ptrInt64(row.PointStamp),
				row.EstampBaht,
				row.Extra,
				ptrInt64(row.BonusReturn),
			}

			cell, _ := excelize.CoordinatesToCellName(1, rowNum)
			if err := sw.SetRow(cell, rowCells); err != nil {
				return err
			}

			// ตั้ง style ให้วันที่และคะแนน
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", rowNum), fmt.Sprintf("B%d", rowNum), dateTimeStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("E%d", rowNum), fmt.Sprintf("E%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("F%d", rowNum), fmt.Sprintf("F%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("G%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("H%d", rowNum), fmt.Sprintf("H%d", rowNum), numberStyle)
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

	c.Header("Content-Disposition", "attachment; filename=claim_report.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())

	return nil
}
func ptrInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func ClaimJoylicoinTx(
	req models.ReturnBonusDto,
) (*models.ReturnBonus, *models.Member, error) {

	setting, err := GetBonusSetting()
	if err != nil {
		return nil, nil, err
	}

	if setting == nil {
		return nil, nil, fmt.Errorf("bonus setting not found")
	}

	dateString := setting.SetDate.Format("2006-01-02")

	var returnBonus *models.ReturnBonus
	var member *models.Member
	dateNow := utils.TimeNowAsia()
	req.BonusDate = *dateNow

	err = config.DB_POS.Transaction(func(tx *gorm.DB) error {

		// 1️⃣ create return bonus
		rb, err := CreateReturnBonusTx(tx, req)
		if err != nil {
			return err
		}
		returnBonus = rb

		// 2️⃣ update member
		m, err := UpdateClaimJoylicoinTx(
			tx,
			int(*req.BonusReturn),
			req.MemberTel,
			int(*req.PointStamp),
		)
		if err != nil {
			return err
		}
		member = m

		// 3️⃣ update stamp_point (POS DB)
		if err := UpdateStampPointClaimTx(tx, req.MemberTel, dateString, req.BonusDate); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// ⚠️ meter_record อยู่ DB_JREADER → ทำแยก
	if err := config.DB_JREADER.
		Exec(`
		UPDATE meter_record mr
		SET
			bonus_status = 'N',
			update_claim_date = ?
		FROM machine_asset ma
		JOIN machine_data md
			ON ma.machine_id = md.mc_id
		WHERE mr.rc_member = ?
		  AND mr.bonus_status = 'Y'
		  AND mr.rc_date::date >= ?
		  AND md.category_id = 1
		  AND mr.rc_asset_id = ma.asset_id
	`,
			req.BonusDate,
			req.MemberTel,
			dateString,
		).Error; err != nil {

		return nil, nil, fmt.Errorf(
			"failed to update meter record: %w",
			err,
		)
	}

	return returnBonus, member, nil
}

func CreateReturnBonusTx(tx *gorm.DB, input models.ReturnBonusDto) (*models.ReturnBonus, error) {

	var entity models.ReturnBonus

	entity.BonusDate = input.BonusDate
	entity.MemberTel = input.MemberTel
	entity.Branch = input.Branch
	entity.SetBonus = input.SetBonus
	entity.SetPointRate = input.SetPointRate
	entity.CSpend = input.CSpend
	entity.PointStamp = input.PointStamp
	entity.BonusReturn = input.BonusReturn
	entity.Employee = input.Employee
	entity.LastPoint = input.LastPoint
	entity.LastJoylicoin = input.LastJoylicoin
	entity.Extra = input.Extra
	entity.EstampBaht = input.EstampBaht
	entity.LastEstamp = input.LastEstamp

	docNo, err := GetReturnBonusDocumentNo(input.PosId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate document no: %w", err)
	}

	entity.BonusNo = docNo

	if err := tx.Create(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create return bonus: %w", err)
	}

	return &entity, nil
}

func UpdateClaimJoylicoinTx(
	tx *gorm.DB,
	jc int,
	memberTel string,
	finalPoint int,
) (*models.Member, error) {

	updateDate := utils.TimeNowAsia()

	result := tx.Exec(`
		UPDATE member 
		SET 
			bonus = bonus + ?, 
			total_point = total_point + ?, 
			ecoin = 0, 
			update_date = ? 
		WHERE tel = ?`,
		jc, finalPoint, updateDate, memberTel,
	)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to update member joylicoin: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("member not found")
	}

	var member models.Member
	if err := tx.Where("tel = ?", memberTel).First(&member).Error; err != nil {
		return nil, err
	}

	return &member, nil
}

func UpdateStampPointClaimTx(
	tx *gorm.DB,
	memberTel string,
	setDate string,
	bonusDate time.Time,
) error {
	result := tx.Exec(`
		UPDATE stamp_point 
		SET st_status = 'N',
		update_claim_date = ?
		WHERE st_tel = ?
		AND st_status = 'Y'
		AND st_date::date >= ?`,
		bonusDate, memberTel, setDate,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update stamp point claim: %w", result.Error)
	}

	return nil
}

func SyncAfterClaim(
	req models.ReturnBonusDto,
	member *models.Member,
	returnBonus *models.ReturnBonus,
) error {

	// 1️⃣ Sync Score
	scores := models.ScoreMember{
		MobileNo:   member.Tel,
		Bonus:      member.Bonus,
		TotalPoint: member.TotalPoint,
		ECoin:      member.Ecoin,
		JubuJibi:   member.JubuJibi,
		FinWow:     member.Finwow,
		EStamp:     member.Estamp,
		MSkill1:    member.MSkill1,
		MSkill2:    member.MSkill2,
		MSkill3:    member.MSkill3,
		MSkill4:    member.MSkill4,
		MSkill5:    member.MSkill5,
	}

	statusCode, _, err := SyncScoreMember(scores)
	if err != nil {
		return fmt.Errorf("SyncScoreMember error: %w", err)
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("SyncScoreMember failed with status: %d", statusCode)
	}

	// 2️⃣ Sync History
	req.BonusDate = returnBonus.BonusDate

	_, err = ReturnBonusSyncHistory(req, returnBonus.ID.String())
	if err != nil {
		return fmt.Errorf("ReturnBonusSyncHistory error: %w", err)
	}

	return nil
}

func SyncScoreAfterClaim(member *models.Member) error {

	scores := models.ScoreMember{
		MobileNo:   member.Tel,
		Bonus:      member.Bonus,
		TotalPoint: member.TotalPoint,
		ECoin:      member.Ecoin,
		JubuJibi:   member.JubuJibi,
		FinWow:     member.Finwow,
		EStamp:     member.Estamp,
		MSkill1:    member.MSkill1,
		MSkill2:    member.MSkill2,
		MSkill3:    member.MSkill3,
		MSkill4:    member.MSkill4,
		MSkill5:    member.MSkill5,
	}

	statusCode, _, err := SyncScoreMember(scores)
	if err != nil {
		return fmt.Errorf("SyncScoreMember error: %w", err)
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("SyncScoreMember failed status: %d", statusCode)
	}

	return nil
}

func SyncHistoryAfterClaim(
	req models.ReturnBonusDto,
	returnBonus *models.ReturnBonus,
) {

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in SyncHistoryAfterClaim:", r)
		}
	}()

	req.BonusDate = returnBonus.BonusDate

	_, err := ReturnBonusSyncHistory(req, returnBonus.ID.String())
	if err != nil {
		log.Println("ReturnBonusSyncHistory failed:", err)
	}
}
