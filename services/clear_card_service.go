package services

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	// "log"
	"strings"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"

	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func ClearCard(input models.ClearCardDto, userId int) (models.ClearCard, error) {
	var entity models.ClearCard

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	// Transaction
	err := config.DB_POS.Transaction(func(tx *gorm.DB) error {
		// 1. create clear card
		entity = models.ClearCard{
			CardNo:        input.CardNo,
			MemberTel:     input.MemberTel,
			BalanceEbonus: input.BalanceEbonus,
			BalanceEcoin:  input.BalanceEcoin,
			CreatedBy:     input.CreatedBy,
			Location:      input.Location,
			ClearDate:     *utils.TimeNowAsia(),
			TimePlay:      input.TimePlay,
			CardType:      input.CardType,
		}

		if err := tx.Create(&entity).Error; err != nil {
			return fmt.Errorf("failed to create clear card: %w", err)
		}

		// 2. ทำ sequential ทีละอัน (ห้าม parallel บน tx เดียวกัน)
		if err := DefaultEntityByCardNoTx(tx, input.CardNo, userId, input.MemberTel, input.Location); err != nil {
			return err
		}

		if err := InActiveCardDepositTx(tx, input.CardNo, userId); err != nil {
			return err
		}
		fmt.Println("BEFORE DeleteCardPlayTx")
		if err := DeleteCardPlayTx(tx, input.CardNo, userId); err != nil {
			return err
		}

		fmt.Println("AFTER DeleteCardPlayTx")

		return nil // commit transaction
	})

	fmt.Println("TRANSACTION FINISHED")
	fmt.Println("TRANSACTION ERROR:", err)

	if err != nil {
		return entity, err
	}

	var cards []models.CardPlay

	err = config.DB_POS.
		Where("LOWER(card_no) = ?", strings.ToLower(input.CardNo)).
		Find(&cards).Error

	if err != nil {
		return entity, err
	}

	for _, card := range cards {
		fmt.Printf(
			"AFTER COMMIT card_no=%s is_delete=%v delete_by=%v\n",
			card.CardNo,
			card.IsDelete,
			card.DeleteBy,
		)
	}

	return entity, nil
}

func InActiveCardDepositTx(tx *gorm.DB, cardNo string, userId int) error {
	const defaultTel = "0000000000"
	now := utils.TimeNowAsia()

	// 1. Inactive CardDeposit
	updateData := map[string]interface{}{
		"is_active":   false,
		"update_date": now,
	}

	if err := tx.Model(&models.CardDeposit{}).
		Where("LOWER(card_no) = ?", strings.ToLower(cardNo)).
		Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to inactive card deposit: %w", err)
	}

	// 2. สร้างรายการใหม่หลังจาก inactive
	bonusExpire := utils.GetBonusExpire()
	newDeposit := models.CardDeposit{
		FromChannel:     "POS",
		CardNo:          cardNo,
		MemberTel:       defaultTel,
		Amount:          0,
		Coin:            0,
		Bonus:           0,
		BalanceCoin:     0,
		BalanceBonus:    0,
		PosId:           "",
		PosMenuId:       0,
		BonusExpireDate: &bonusExpire,
	}

	if err := tx.Create(&newDeposit).Error; err != nil {
		return fmt.Errorf("failed to create new card deposit: %w", err)
	}

	return nil
}

func DeleteCardPlayTx(tx *gorm.DB, cardNo string, userId int) error {
	fmt.Println("🔥 ENTER DeleteCardPlayTx:", cardNo)
	now := utils.TimeNowAsia()
	updateData := map[string]interface{}{
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": now,
	}
	result := tx.Model(&models.CardPlay{}).
		Where("LOWER(card_no) = ?", strings.ToLower(cardNo)).
		Updates(updateData)

	fmt.Println("🔥 DeleteCardPlayTx RowsAffected:", result.RowsAffected)
	fmt.Println("🔥 DeleteCardPlayTx Error:", result.Error)

	if result.Error != nil {
		return fmt.Errorf("failed to delete card play: %w", result.Error)
	}
	return nil
}

func DefaultEntityByCardNoTx(tx *gorm.DB, cardNo string, userId int, oldMemberTel, location string) error {
	// กำหนดค่า default
	const defaultTel = "0000000000"
	now := utils.TimeNowAsia()

	// อัปเดต member_tel เป็นค่า default
	updateData := map[string]interface{}{
		"member_tel":  defaultTel,
		"update_by":   userId,
		"update_date": now,
		"is_active":   true,
	}

	if err := tx.Model(&models.CardEntity{}).
		Where("LOWER(card_no) = ?", strings.ToLower(cardNo)).
		Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update card entity: %w", err)
	}

	// บันทึก log ถ้าเบอร์เดิมไม่ใช่ default
	if oldMemberTel != defaultTel {
		logUpdate := models.LogUpdateCardEntity{
			CardNo:      cardNo,
			OldMobile:   oldMemberTel,
			NewMobile:   defaultTel,
			Location:    location,
			UpdatedDate: time.Now(),
		}
		if err := tx.Create(&logUpdate).Error; err != nil {
			return fmt.Errorf("failed to log update card entity: %w", err)
		}
	}

	return nil
}

func PosClearCardSearchList(params models.SearchTaxInvoiceReport) (utils.SearchResult, error) {
	var result utils.SearchResult
	var transactions []models.ClearCard
	var response models.ClearCardData
	var TotalECoin, TotalEBonus, TotalTimePlay sql.NullInt64

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("clear_card cc").
		Select(`*`)

	sumQuery := config.DB_POS.Table("clear_card cc")

	if params.Location != "" {
		query = query.Where("(cc.location = ?)", params.Location)
		sumQuery = sumQuery.Where("(cc.location = ?)", params.Location)
	}
	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("cc.clear_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
		sumQuery = sumQuery.Where("cc.clear_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	if err := sumQuery.Select(`
		COALESCE(SUM(cc.balance_ecoin),0) AS total_ecoin,
		COALESCE(SUM(cc.balance_ebonus),0) AS total_ebonus,
		COALESCE(SUM(cc.time_play),0) AS total_time_play
	`).Row().Scan(&TotalECoin, &TotalEBonus, &TotalTimePlay); err != nil {
		return result, fmt.Errorf("sum query failed: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("cc.clear_date ASC").Limit(params.Skip).Offset(offset).Find(&transactions).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	response.ClearCardList = transactions
	response.TotalECoin = TotalECoin.Int64
	response.TotalEBonus = TotalEBonus.Int64
	response.TotalTimePlay = TotalTimePlay.Int64

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     response,
	}
	return result, nil
}

func ExportPosClearCardSearchList(params models.ExportClearCardReport) ([]models.ClearCardExport, error) {
	var transactions []models.ClearCardExport

	if config.DB_POS == nil {
		return transactions, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("clear_card cc").
		Select(`cc.*, m.id as member_id`).
		Joins("left join member m on cc.member_tel = m.tel")

	if params.Location != "" {
		query = query.Where("(cc.location = ?)", params.Location)

	}
	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("cc.clear_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)

	}

	if err := query.Order("cc.clear_date ASC").Find(&transactions).Error; err != nil {
		return transactions, fmt.Errorf("data query failed: %w", err)
	}

	return transactions, nil
}

func ExportClearCardToExcel(c *gin.Context, data []models.ClearCardExport) error {
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
		headers := []interface{}{"Date", "Card ID", "Member ID", "ecoin", "ebonus", "time play", "สาขา", "user"}
		if err := sw.SetRow("A1", headers, excelize.RowOpts{StyleID: boldCenterStyle}); err != nil {
			return err
		}

		// เขียน row ข้อมูล
		for i, row := range data[start:end] {
			if row.MemberID != nil {
				memberID = *row.MemberID
			}
			rowNum := i + 2
			rowCells := []interface{}{
				row.ClearDate,
				row.CardNo,
				memberID,
				row.BalanceEcoin,
				row.BalanceEbonus,
				row.TimePlay,
				row.Location,
				row.CreatedBy,
			}

			cell, _ := excelize.CoordinatesToCellName(1, rowNum)
			if err := sw.SetRow(cell, rowCells); err != nil {
				return err
			}

			// ตั้ง style ให้วันที่และคะแนน
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("B%d", rowNum), dateTimeStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("D%d", rowNum), fmt.Sprintf("E%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("E%d", rowNum), fmt.Sprintf("E%d", rowNum), numberStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("F%d", rowNum), fmt.Sprintf("F%d", rowNum), numberStyle)
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

	c.Header("Content-Disposition", "attachment; filename=clear_card_report.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())

	return nil
}
