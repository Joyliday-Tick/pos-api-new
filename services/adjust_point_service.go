package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/phpdave11/gofpdf"
	"github.com/xuri/excelize/v2"
)

func CreateAdjPoint(input models.AdjustPointDto) (models.AdjustPoint, error) {
	var entity models.AdjustPoint

	entity.Username = input.Username
	entity.MemberTel = input.MemberTel
	entity.Reason = input.Reason
	entity.AdjBranch = input.AdjBranch
	entity.AdjPoint = input.AdjPoint
	entity.AdjJoylicoin = input.AdjJoylicoin
	entity.AdjFinwow = input.AdjFinwow
	entity.AdjEstamp = input.AdjEstamp
	entity.AdjMskill1 = input.AdjMskill1
	entity.AdjMskill2 = input.AdjMskill2
	entity.AdjMskill3 = input.AdjMskill3
	entity.AdjMskill4 = input.AdjMskill4
	entity.AdjMskill5 = input.AdjMskill5
	entity.AdjJubuJibi = input.AdjJubuJibi
	entity.AdjDate = input.AdjDate

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create adjust point: %w", err)
	}

	return entity, nil
}

func AdjustPointSyncHistory(input models.AdjustPointDto, id string) (string, error) {
	var historys []models.ScoreHistory

	status, customer, err := GetCustomerByMobileNo(strings.TrimSpace(input.MemberTel))
	if err != nil || status != 200 || customer == nil {
		return "", fmt.Errorf("customer not found with tel: %s", input.MemberTel)
	}

	status, scoreTypes, err := GetScoreType()
	if err != nil || status != 200 || len(scoreTypes) == 0 {
		return "", fmt.Errorf("score type not found")
	}

	status, branch, err := GetBranchByCode(strings.TrimSpace(input.AdjBranch))
	if err != nil || status != 200 || branch == nil {
		return "", fmt.Errorf("branch not found with code: %s", input.AdjBranch)
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

	jubuID, ok := scoreTypeMap["Z_Jubu_Jibi"]
	if !ok {
		return "", fmt.Errorf("jubu_jibi score type not found")
	}

	// Create history entries
	if input.AdjPoint != 0 {
		historys = append(historys, models.ScoreHistory{
			CustomerID:       customer.ID,
			ScoreTypeID:      pointID,
			BranchID:         branch.ID,
			Amount:           int(input.AdjPoint),
			TransactionDate:  input.AdjDate,
			Description:      StringPtr("NEW-POS Adjust Point"),
			CreateBy:         input.Username,
			MobileNo:         input.MemberTel,
			RefTransaction:   StringPtr("adjust_point"),
			RefTransactionID: &id,
			PosID:            nil,
		})
	}

	if input.AdjJoylicoin != 0 {
		historys = append(historys, models.ScoreHistory{
			CustomerID:       customer.ID,
			ScoreTypeID:      coinID,
			BranchID:         branch.ID,
			Amount:           int(input.AdjJoylicoin),
			TransactionDate:  input.AdjDate,
			Description:      StringPtr("NEW-POS Adjust Point"),
			CreateBy:         input.Username,
			MobileNo:         input.MemberTel,
			RefTransaction:   StringPtr("adjust_point"),
			RefTransactionID: &id,
			PosID:            nil,
		})
	}

	if input.AdjEstamp != 0 {
		historys = append(historys, models.ScoreHistory{
			CustomerID:       customer.ID,
			ScoreTypeID:      stampID,
			BranchID:         branch.ID,
			Amount:           int(input.AdjEstamp),
			TransactionDate:  input.AdjDate,
			Description:      StringPtr("NEW-POS Adjust Point"),
			CreateBy:         input.Username,
			MobileNo:         input.MemberTel,
			RefTransaction:   StringPtr("adjust_point"),
			RefTransactionID: &id,
			PosID:            nil,
		})
	}

	if input.AdjJubuJibi != 0 {
		historys = append(historys, models.ScoreHistory{
			CustomerID:       customer.ID,
			ScoreTypeID:      jubuID,
			BranchID:         branch.ID,
			Amount:           int(input.AdjJubuJibi),
			TransactionDate:  input.AdjDate,
			Description:      StringPtr("NEW-POS Adjust Point"),
			CreateBy:         input.Username,
			MobileNo:         input.MemberTel,
			RefTransaction:   StringPtr("adjust_point"),
			RefTransactionID: &id,
			PosID:            nil,
		})
	}

	if len(historys) == 0 {
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

func AdjustPointReport(params models.SearchAdjustPointReport) (utils.SearchResult, error) {
	var result utils.SearchResult
	var adjPoints []models.AdjustPointData

	query := config.DB_POS.Table("adjust_point ap").
		Select(`
        ap.*, m.m_name as firstname, m.s_name as lastname
    `).Joins("left join member m on ap.member_tel = m.tel")

	if params.Location != "" {
		query = query.Where("ap.adj_branch ILIKE ?", "%"+params.Location+"%")
	}

	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("ap.adj_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("ap.adj_date ASC").Limit(params.Skip).Offset(offset).Find(&adjPoints).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     adjPoints,
	}
	return result, nil
}

func ExportAdjustPointReport(params models.ExportAdjustPointReport) ([]models.AdjustPointDataExport, error) {
	var adjPoints []models.AdjustPointDataExport

	query := config.DB_POS.Table("adjust_point ap").
		Select(`
        ap.*, m.m_name as firstname, m.s_name as lastname, m.id as member_id
    `).Joins("left join member m on ap.member_tel = m.tel")

	if params.Location != "" {
		query = query.Where("ap.adj_branch ILIKE ?", "%"+params.Location+"%")
	}

	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("ap.adj_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	if err := query.Order("ap.adj_date ASC").Find(&adjPoints).Error; err != nil {
		return adjPoints, fmt.Errorf("data query failed: %w", err)
	}

	return adjPoints, nil
}

func AdjExportToPDF(c *gin.Context, data []models.AdjustPointDataExport, params models.ExportAdjustPointReport) error {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(12, 15, 12)
	pdf.SetAutoPageBreak(true, 12)

	// โหลดฟอนต์ไทย
	fontName := "THSarabunNew"
	fontPath := "./fonts/THSarabunNew/THSarabunNew.ttf"
	if _, err := os.Stat(fontPath); err != nil {
		log.Fatalf("ไม่พบฟอนต์ %s: %v", fontPath, err)
	}
	pdf.AddUTF8Font(fontName, "", fontPath)
	pdf.AddUTF8Font(fontName, "B", fontPath)

	// column widths
	type widths struct {
		date, memberId, firstname, lastname, stamp, joylicoin, point,
		finwow, power, agility, reaction_time, balance, speed,
		jubu_jibi, reason, branch, user float64
	}
	colW := widths{
		date:          24,
		memberId:      22,
		firstname:     20,
		lastname:      22,
		stamp:         12,
		joylicoin:     12,
		point:         12,
		finwow:        12,
		power:         12,
		agility:       12,
		reaction_time: 12,
		balance:       12,
		speed:         12,
		jubu_jibi:     14,
		reason:        30,
		branch:        18,
		user:          20,
	}

	bodyFontSize := 10.0
	lineH := 5.5
	pdf.SetFont(fontName, "", bodyFontSize)

	cellPadX := 1.5

	// Header
	addHeader := func() {
		pdf.SetFont(fontName, "B", 28)
		pdf.CellFormat(0, 12, "Joyliday", "", 1, "C", false, 0, "")
		pdf.Ln(1)

		pdf.SetFont(fontName, "B", 16)
		pdf.CellFormat(0, 10, "รายงานการปรับปรุงคะแนน", "", 1, "C", false, 0, "")
		pdf.Ln(1)

		startDate := strings.ReplaceAll(params.StartDate, "-", "/")
		endDate := strings.ReplaceAll(params.EndDate, "-", "/")
		pdf.SetFont(fontName, "B", 12)
		pdf.CellFormat(0, 8, fmt.Sprintf("ตั้งแต่วันที่ : %s    ถึงวันที่ : %s", startDate, endDate), "", 1, "C", false, 0, "")
		pdf.Ln(2)

		pdf.SetLineWidth(0.2)
		pdf.Line(12, pdf.GetY(), 297-12, pdf.GetY())
		pdf.Ln(3)

		pdf.SetFont(fontName, "B", 11)
		pdf.SetFillColor(245, 245, 245)
		headers := []struct {
			w   float64
			txt string
		}{
			{colW.date, "วันที่"},
			{colW.memberId, "memberId"},
			{colW.firstname, "ชื่อ"},
			{colW.lastname, "นามสกุล"},
			{colW.stamp, "e-stamp"},
			{colW.joylicoin, "joylicoin"},
			{colW.point, "point"},
			{colW.finwow, "finwow"},
			{colW.power, "power"},
			{colW.agility, "agility"},
			{colW.reaction_time, "reaction"},
			{colW.balance, "balance"},
			{colW.speed, "speed"},
			{colW.jubu_jibi, "jubu_jibi"},
			{colW.reason, "หมายเหตุ"},
			{colW.branch, "สาขา"},
			{colW.user, "user"},
		}
		for i, h := range headers {
			endLine := 0
			if i == len(headers)-1 {
				endLine = 1
			}
			pdf.CellFormat(h.w, lineH, h.txt, "1", endLine, "C", true, 0, "")
		}
		pdf.SetFont(fontName, "", bodyFontSize)
	}

	// Footer
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont(fontName, "", 9)
		pdf.CellFormat(0, 10, fmt.Sprintf("หน้า %d", pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	// draw row
	drawRow := func(r models.AdjustPointDataExport) {
		cols := []struct {
			w     float64
			txt   string
			align string
			wrap  bool
		}{
			{colW.date, r.AdjDate.Format("02/01/2006 15:04"), "L", false},
			{colW.memberId, fmt.Sprintf("%d", r.MemberId), "L", false},
			{colW.firstname, r.Firstname, "L", false},
			{colW.lastname, r.Lastname, "L", false},
			{colW.stamp, fmt.Sprintf("%d", r.AdjEstamp), "R", false},
			{colW.joylicoin, fmt.Sprintf("%d", r.AdjJoylicoin), "R", false},
			{colW.point, fmt.Sprintf("%d", r.AdjPoint), "R", false},
			{colW.finwow, fmt.Sprintf("%d", r.AdjFinwow), "R", false},
			{colW.power, fmt.Sprintf("%d", r.AdjMskill1), "R", false},
			{colW.agility, fmt.Sprintf("%d", r.AdjMskill2), "R", false},
			{colW.reaction_time, fmt.Sprintf("%d", r.AdjMskill3), "R", false},
			{colW.balance, fmt.Sprintf("%d", r.AdjMskill4), "R", false},
			{colW.speed, fmt.Sprintf("%d", r.AdjMskill5), "R", false},
			{colW.jubu_jibi, fmt.Sprintf("%d", r.AdjJubuJibi), "R", false},
			{colW.reason, r.Reason, "L", true}, // wrap
			{colW.branch, r.AdjBranch, "L", false},
			{colW.user, r.Username, "L", false},
		}

		// คำนวณความสูงจริง
		maxH := lineH
		for _, c := range cols {
			if c.wrap {
				usableW := c.w - 2*cellPadX
				lines := pdf.SplitLines([]byte(c.txt), usableW)
				h := float64(len(lines)) * lineH
				if h > maxH {
					maxH = h
				}
			}
		}

		// ตรวจหน้าใหม่
		_, pageH := pdf.GetPageSize()
		bottomMargin := 12.0
		y := pdf.GetY()
		if y+maxH > pageH-bottomMargin {
			pdf.AddPage()
			addHeader()
			y = pdf.GetY()
		}

		leftMargin, _, _, _ := pdf.GetMargins()

		for _, c := range cols {
			x := pdf.GetX()
			pdf.Rect(x, y, c.w, maxH, "D") // วาดกรอบ
			usableW := c.w - 2*cellPadX
			if usableW < 1 {
				usableW = c.w
			}

			pdf.SetXY(x+cellPadX, y) // ชิดบนสำหรับทุก column

			if c.wrap {
				// column wrap → MultiCell อัตโนมัติ (wrap text)
				pdf.MultiCell(usableW, lineH, c.txt, "", c.align, false)
			} else {
				// column ปกติ → MultiCell ชิดบน
				pdf.MultiCell(usableW, lineH, c.txt, "", c.align, false)
			}

			pdf.SetXY(x+c.w, y) // ไป column ถัดไป
		}

		pdf.SetXY(leftMargin, y+maxH)
	}

	// สร้างหน้าแรก + header
	pdf.AddPage()
	addHeader()

	for _, r := range data {

		drawRow(r)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="adjust_point_report.pdf"`)
	c.Header("Content-Transfer-Encoding", "binary")
	return pdf.Output(c.Writer)
}

func AdjExportToExcel(c *gin.Context, data []models.AdjustPointDataExport, params models.ExportAdjustPointReport) error {
	f := excelize.NewFile()
	sheet := "Summary"

	// เปลี่ยนชื่อ Sheet1 เป็น Summary ถ้ามี
	index, err := f.GetSheetIndex("Sheet1")
	if err == nil && index != -1 {
		f.SetSheetName("Sheet1", sheet)
	}

	// ตั้งให้ Summary เป็น active sheet
	if idx, err := f.GetSheetIndex(sheet); err == nil {
		f.SetActiveSheet(idx)
	}

	boldCenterStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	// สร้าง style ธรรมดาเน้นกึ่งกลางสำหรับ header ใหญ่
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 14,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	logoStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 30,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	// rightAlignSmallStyle, err := f.NewStyle(&excelize.Style{
	// 	Font: &excelize.Font{
	// 		Bold: false,
	// 		Size: 10,
	// 	},
	// 	Alignment: &excelize.Alignment{
	// 		Horizontal: "right",
	// 		Vertical:   "center",
	// 	},
	// })
	// if err != nil {
	// 	return err
	// }

	startDate := strings.ReplaceAll(params.StartDate, "-", "/")
	endDate := strings.ReplaceAll(params.EndDate, "-", "/")
	// dataNow := utils.TimeNowAsia()

	// เพิ่มหัวตาราง 3 บรรทัด
	f.MergeCell(sheet, "A1", "Q1") // รวม 15 คอลัมน์ (A-O)
	f.SetCellValue(sheet, "A1", "Joyliday")
	f.SetCellStyle(sheet, "A1", "A1", logoStyle)

	f.MergeCell(sheet, "A2", "Q2") // รวม 15 คอลัมน์ (A-O)
	f.SetCellValue(sheet, "A2", "รายงานการปรับปรุงคะแนน")
	f.SetCellStyle(sheet, "A2", "A2", headerStyle)

	f.MergeCell(sheet, "A3", "Q3")
	f.SetCellValue(sheet, "A3", fmt.Sprintf("ตั้งแต่วันที่ : %s ถึงวันที่ : %s", startDate, endDate))
	f.SetCellStyle(sheet, "A3", "A3", boldCenterStyle)

	// f.MergeCell(sheet, "A3", "G3")
	// printDate := dataNow.Format("02/01/2006 15:04:05")
	// f.SetCellValue(sheet, "A3", "พิมพ์วันที่ : "+printDate)
	// f.SetCellStyle(sheet, "A3", "A3", rightAlignSmallStyle)

	// ตารางหัวตารางจริงเริ่มที่แถว 5 (เว้นวรรค 1 บรรทัด)
	headers := []string{"วันที่", "MemberId", "ชื่อ", "นามสกุล", "e-stamp", "joylicoin", "point", "finwow", "power", "agility", "reaction_time", "balance", "speed", "jubu_jibi", "หมายเหตุ", "สาขา", "user"}
	for col, header := range headers {
		cell := fmt.Sprintf("%s5", string(rune('A'+col))) // A5, B5, ...
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, boldCenterStyle)
	}

	// Style วันที่ และตัวเลข เหมือนเดิม
	// dateFormat := "dd/mm/yyyy HH:mm"
	format := "dd/mm/yyyy HH:mm"
	dateStyle, err := f.NewStyle(&excelize.Style{
		CustomNumFmt: &format,
	})
	if err != nil {
		return err
	}

	// numberFormat := "#,##0.00"
	// numberStyle, err := f.NewStyle(&excelize.Style{
	// 	CustomNumFmt: &numberFormat,
	// })
	// if err != nil {
	// 	return err
	// }

	// เริ่มเขียนข้อมูลจริงจากแถว 6
	for i, row := range data {
		rowNum := i + 6

		dateCell := fmt.Sprintf("A%d", rowNum)
		f.SetCellValue(sheet, dateCell, row.AdjDate)
		f.SetCellStyle(sheet, dateCell, dateCell, dateStyle)

		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), row.MemberId)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), row.Firstname)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowNum), row.Lastname)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowNum), row.AdjEstamp)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowNum), row.AdjJoylicoin)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", rowNum), row.AdjPoint)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", rowNum), row.AdjFinwow)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", rowNum), row.AdjMskill1)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", rowNum), row.AdjMskill2)
		f.SetCellValue(sheet, fmt.Sprintf("K%d", rowNum), row.AdjMskill3)
		f.SetCellValue(sheet, fmt.Sprintf("L%d", rowNum), row.AdjMskill4)
		f.SetCellValue(sheet, fmt.Sprintf("M%d", rowNum), row.AdjMskill5)
		f.SetCellValue(sheet, fmt.Sprintf("N%d", rowNum), row.AdjJubuJibi)
		f.SetCellValue(sheet, fmt.Sprintf("O%d", rowNum), row.Reason)
		f.SetCellValue(sheet, fmt.Sprintf("P%d", rowNum), row.AdjBranch)
		f.SetCellValue(sheet, fmt.Sprintf("Q%d", rowNum), row.Username)
	}

	// เขียนลง buffer และส่ง response
	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		return err
	}

	c.Header("Content-Disposition", "attachment; filename=adjust_point_report.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())
	return nil
}
