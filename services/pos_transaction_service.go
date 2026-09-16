package services

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"

	// "strings"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"

	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
	"github.com/phpdave11/gofpdf"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	// "new-pos-api/utils"
	// "gorm.io/gorm"
)

func CreatePosTransaction(input models.PosTransactionDto, userId int) (models.PosTransaction, error) {
	var transaction models.PosTransaction

	// Validate: ต้องมี SubTransactions อย่างน้อย 1 รายการ
	if len(input.SubTransactions) == 0 {
		return transaction, fmt.Errorf("sub transaction is required")
	}

	// Map input DTO ไปยัง model
	transaction.BillNo = input.BillNo
	transaction.BillDate = input.BillDate
	transaction.PosID = input.PosID
	transaction.BillLocation = input.BillLocation
	transaction.Cashier = input.Cashier
	transaction.CardNo = input.CardNo
	transaction.MemberTel = input.MemberTel
	transaction.ProductPrice = input.ProductPrice
	transaction.ECoin = input.ECoin
	transaction.FreePoint = input.FreePoint
	transaction.EBonus = input.EBonus
	transaction.BillPaymentId = input.BillPaymentId // Assuming this is UUID
	transaction.PosType = input.PosType
	transaction.BillStatus = input.BillStatus
	transaction.BonusStatus = input.BonusStatus
	transaction.BankDetail = input.BankDetail
	transaction.CreateBy = userId

	// ตรวจสอบ DB connection
	if config.DB_POS == nil {
		return transaction, fmt.Errorf("database POS connection is nil")
	}

	// สร้าง transaction หลัก
	if err := config.DB_POS.Create(&transaction).Error; err != nil {
		return transaction, fmt.Errorf("failed to create POS transaction: %w", err)
	}

	// เตรียมและบันทึก SubTransactions
	var subs []models.PosSubTransaction
	for _, sub := range input.SubTransactions {
		var cardDepositUUID uuid.UUID
		if sub.CardDepositId != "" {
			parsedUUID, err := uuid.Parse(sub.CardDepositId)
			if err != nil {
				return transaction, fmt.Errorf("invalid card_deposit_id: %w", err)
			}
			cardDepositUUID = parsedUUID
		}
		subTransaction := models.PosSubTransaction{
			BillNo:        transaction.BillNo,
			ProductID:     sub.ProductID,
			Qty:           sub.Qty,
			Price:         sub.Price,
			ECoin:         sub.ECoin,
			EBonus:        sub.EBonus,
			CardDepositId: cardDepositUUID,
			CreateBy:      userId,
			CreateDate:    *utils.TimeNowAsia(),
		}
		subs = append(subs, subTransaction)
	}

	// เรียกฟังก์ชันบันทึก batch
	if _, err := CreateBatchSubTransaction(subs); err != nil {
		return transaction, fmt.Errorf("failed to create sub transactions: %w", err)
	}

	return transaction, nil
}

func UpdatePosTransactionStatus(billNo string, status string, userId int) (models.PosTransaction, error) {
	var existing models.PosTransaction

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database POS connection is nil")
	}

	// ค้นหาข้อมูลที่มีอยู่
	err := config.DB_POS.Where("LOWER(bill_no) = ?", strings.ToLower(billNo)).
		First(&existing).Error
	if err != nil {
		return existing, fmt.Errorf("pos transaction not found: %w", err)
	}

	// อัปเดตเฉพาะฟิลด์ที่ต้องการ
	updateFields := map[string]interface{}{
		"bill_status": status,
		"update_by":   userId,
		"update_date": utils.TimeNowAsia(),
	}

	if err := config.DB_POS.Model(&models.PosTransaction{}).
		Where("id = ?", existing.ID).
		Updates(updateFields).Error; err != nil {
		return existing, fmt.Errorf("failed to update pos transaction: %w", err)
	}

	// โหลดข้อมูลล่าสุดหลังอัปเดต
	if err := config.DB_POS.Where("id = ?", existing.ID).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated pos transaction: %w", err)
	}

	return existing, nil
}

func PosTransactionSearchList(params models.SearchPosTransactionParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var transactions []models.PosTransactionData

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("pos_transaction pt").
		Select(`pt.id ,
				pt.bill_no,
				pt.bill_date,
				pt.member_tel,
				pt.card_no,
				coalesce(pt.product_price,0) as product_price,
				coalesce(pt.e_coin,0) as e_coin,
				coalesce(pt.e_bonus,0) as e_bonus,
				pt.bill_location,
				b.branch_name ,
				pt.bill_status ,
				ps.pos_code ,
				pt.pos_id,
				pt.cashier,
				mp.name as bill_payment`).
		Joins("LEFT  JOIN pos_station ps ON pt.pos_id = CAST(ps.id AS VARCHAR)").
		Joins("LEFT  JOIN branch b on pt.bill_location = b.branch_code").
		Joins("LEFT  JOIN master_payment mp on pt.bill_payment_id = mp.id").
		Where("pt.is_delete = false")

	if params.BillNo != "" {
		search := "%" + params.BillNo + "%"
		query = query.Where("(pt.bill_no ILIKE ?)", search)
	}
	if params.MemberTel != "" {
		tel := "%" + params.MemberTel + "%"
		query = query.Where("(pt.member_tel ILIKE ?)", tel)
	}
	if params.CardNo != "" {
		card := "%" + params.CardNo + "%"
		query = query.Where("(pt.card_no ILIKE ?)", card)
	}
	if params.POSID != "" {
		query = query.Where("(pt.pos_id = ?)", params.POSID)
	}
	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("pt.bill_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("pt.create_date DESC").Limit(params.Skip).Offset(offset).Find(&transactions).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     transactions,
	}
	return result, nil
}

func FindExistPosTransactionBillNo(billNo string) (*models.PosTransaction, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var pos models.PosTransaction
	lowerBillNo := strings.ToLower(billNo)

	err := config.DB_POS.
		Where("LOWER(bill_no) = ?", lowerBillNo).
		First(&pos).Error

	// ถ้าไม่พบ record → return nil, nil
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	// กรณี error อื่น → คืน error
	if err != nil {
		return nil, err
	}

	// พบข้อมูล → คืนค่าปกติ
	return &pos, nil
}

func PosSaleReport(params models.SearchPosSaleReport) (models.PosReport, error) {
	var posReport models.PosReport

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errQuery error

	var products []models.PosProductList
	var posTransaction []models.PosTransactionReport

	// ปิดเมื่อทำ Goroutine เสร็จ
	wg.Add(2)

	// ดึง ProductList
	go func() {
		defer wg.Done()
		query := config.DB_POS.Table("pos_sub_transaction pst").
			Select(`pst.product_id,
				pm.description as product_name,
				sum(pt.product_price) as product_price,
				sum(pst.qty) as qty,
				sum(pst.price) as price,
				sum(pst.qty*pst.price) as total,
				sum(pst.qty*pst.e_coin) as e_coin,
				sum(pst.qty*pst.e_bonus) as e_bonus,
				max(TO_CHAR(bill_date, 'YYYY-MM-DD')) as date,
				max(pt.bill_location) as location`).
			Joins("LEFT JOIN pos_menu pm on pst.product_id = pm.id").
			Joins("LEFT JOIN pos_transaction pt on pst.bill_no = pt.bill_no").
			Where("pt.bill_status = 'Normal' AND pt.pos_type != 'kiosk'")

		if params.Location != "" {
			query = query.Where("pt.bill_location = ?", params.Location)
		}
		if params.StartDate != "" && params.EndDate != "" {
			query = query.Where("pt.bill_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
		}

		query = query.Group("pst.product_id, pm.description")

		if err := query.Find(&products).Error; err != nil {
			mu.Lock()                                                  //ล็อก mutex ก่อนเขียนค่า errQuery เพื่อไม่ให้ goroutines อื่นมาเขียนพร้อมกัน
			errQuery = fmt.Errorf("query pos product failed: %w", err) //บันทึก error ที่เกิดขึ้นลงในตัวแปรกลาง errQuery
			mu.Unlock()                                                //ปลดล็อก mutex เพื่อให้ goroutine อื่นสามารถเข้ามาแก้ไขค่าตัวแปรร่วมได้
		}
	}()

	// ดึง Transaction รายการ
	go func() {
		defer wg.Done()
		query1 := config.DB_POS.Table("pos_transaction pt").
			Select(`pt.*, mp."name" as bill_payment_name`).
			Joins("LEFT JOIN master_payment mp on pt.bill_payment_id = mp.id").
			Where("pt.pos_type != 'kiosk'")

		if params.Location != "" {
			query1 = query1.Where("pt.bill_location = ?", params.Location)
		}
		if params.StartDate != "" && params.EndDate != "" {
			query1 = query1.Where("pt.bill_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
		}

		if err := query1.Find(&posTransaction).Error; err != nil {
			mu.Lock()
			errQuery = fmt.Errorf("query pos transaction failed: %w", err)
			mu.Unlock()
		}
	}()

	wg.Wait()
	if errQuery != nil {
		return posReport, errQuery
	}

	// Assign ProductList
	posReport.ProductList = products

	if len(posTransaction) > 0 {
		// ใช้ local map แล้ว assign ภายหลัง
		summaryCashierMap := make(map[string]float64)
		summaryCashierPaymentMap := make(map[string]map[string]float64) // cashier -> payment -> amount
		summaryPaymentMap := make(map[string]float64)
		summaryBankMap := make(map[string]float64)

		countNormal := 0
		countVoid := 0

		for _, pt := range posTransaction {
			if pt.BillStatus == "Void" {
				fmt.Println("void", pt.BillNo)
				countVoid++
				continue
			}
			if pt.BillStatus == "Normal" {
				countNormal++
				price := float64(pt.ProductPrice)

				summaryCashierMap[pt.Cashier] += price
				// รวมยอดตามประเภทการชำระเงินของแคชเชียร์คนนั้น
				if _, ok := summaryCashierPaymentMap[pt.Cashier]; !ok {
					summaryCashierPaymentMap[pt.Cashier] = make(map[string]float64)
				}
				summaryCashierPaymentMap[pt.Cashier][pt.BillPaymentName] += price

				// รวมยอดตามประเภทการชำระเงินทั้งหมด
				summaryPaymentMap[pt.BillPaymentName] += price
				if pt.BankDetail != "" {
					summaryBankMap[pt.BankDetail] += price
				}
			}
		}

		posReport.BillNormalCount = countNormal
		posReport.BillVoidCount = countVoid

		// สร้างสรุป
		posReport.Cashiers = toSortedCashierSummary(summaryCashierMap, summaryCashierPaymentMap)
		posReport.Payments = toSortedPaymentSummary(summaryPaymentMap)
		posReport.BankDetails = toSortedBankSummary(summaryBankMap)
	}

	return posReport, nil
}

func toSortedCashierSummary(
	summary map[string]float64,
	paymentMap map[string]map[string]float64,
) []models.CashierSummary {

	var summaries []models.CashierSummary
	for cashierName, total := range summary {
		payments := make([]models.PaymentSummary, 0)

		if pm, ok := paymentMap[cashierName]; ok {
			for paymentName, amount := range pm {
				payments = append(payments, models.PaymentSummary{
					PaymentName: paymentName,
					TotalAmount: amount,
				})
			}
			// เรียงตามยอดสูง -> ต่ำ
			sort.Slice(payments, func(i, j int) bool {
				return payments[i].TotalAmount > payments[j].TotalAmount
			})
		}

		summaries = append(summaries, models.CashierSummary{
			Cashier:     cashierName,
			TotalAmount: total,
			Payments:    payments,
		})
	}

	// เรียง cashier ตามยอดรวมสูง -> ต่ำ
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].TotalAmount > summaries[j].TotalAmount
	})

	return summaries
}

func toSortedPaymentSummary(summary map[string]float64) []models.PaymentSummary {
	var summaries []models.PaymentSummary
	for name, total := range summary {
		summaries = append(summaries, models.PaymentSummary{
			PaymentName: name,
			TotalAmount: total,
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].TotalAmount > summaries[j].TotalAmount
	})
	return summaries
}

func toSortedBankSummary(summary map[string]float64) []models.BankDetailSummary {
	var summaries []models.BankDetailSummary
	for name, total := range summary {
		summaries = append(summaries, models.BankDetailSummary{
			BankName:    name,
			TotalAmount: total,
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].TotalAmount > summaries[j].TotalAmount
	})
	return summaries
}

// // GenerateBillNo generate bill no
// func GenerateBillNo(PosID string) string {
// 	var BillNo string
// 	var totalRecord int64
// 	currentDate := time.Now()
// 	YYYY := currentDate.Year()
// 	YY := YYYY % 100
// 	MMM := currentDate.Month()
// 	MM := int(MMM)
// 	DD := currentDate.Day()
// 	MMStr := fmt.Sprintf("%02d", MM)
// 	DDStr := fmt.Sprintf("%02d", DD)

// 	// config.DB_POS.Model(&models.PosTransaction{}).Where("pos_id = ? AND Extract(year from bill_date) = ? AND Extract(month from bill_date) = ?", PosID, YYYY, MM).Count(&totalRecord)
// 	config.DB_POS.Model(&models.PosTransaction{}).
// 		Where(
// 			"pos_id = ? AND Extract(year from bill_date) = ? AND Extract(month from bill_date) = ? AND Extract(day from bill_date) = ?",
// 			PosID, YYYY, MM, DD,
// 		).
// 		Count(&totalRecord)
// 	totalRecord += 1
// 	padTotal := fmt.Sprintf("%04d", totalRecord)
// 	fmt.Println("totalRecord", totalRecord)

// 	prefix := os.Getenv("BILL_PREFIX_NORMAL")
// 	BillNo = prefix + "-" + PosID + "-" + strconv.Itoa(YY) + MMStr + DDStr + padTotal
// 	return BillNo
// }

func GenerateBillNo(PosID string) string {
	var totalRecord int64

	now := time.Now()
	year, month, day := now.Date()

	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	startOfNextDay := startOfDay.AddDate(0, 0, 1)

	config.DB_POS.Model(&models.PosTransaction{}).
		Where(
			"pos_id = ? AND bill_date >= ? AND bill_date < ?",
			PosID,
			startOfDay,
			startOfNextDay,
		).
		Count(&totalRecord)

	totalRecord++

	prefix := os.Getenv("BILL_PREFIX_NORMAL")

	// totalRecord นับเฉพาะบิลของ "วันนี้" (startOfDay..startOfNextDay)
	// เลขบิลจึงต้องมี day ด้วย ไม่งั้นพอขึ้นวันใหม่ตัวนับรีเซ็ตเป็น 1
	// แล้วได้เลขซ้ำกับบิลของวันก่อนหน้าในเดือนเดียวกัน
	BillNo := fmt.Sprintf(
		"%s-%s-%02d%02d%02d%04d",
		prefix,
		PosID,
		year%100,
		month,
		day,
		totalRecord,
	)

	return BillNo
}

// func TaxInvoiceReport(params models.SearchTaxInvoiceReport) (utils.SearchResult, error) {
// 	var result utils.SearchResult
// 	var taxInvoices []models.TaxInvoiceSummaryReport
// 	var taxInvoiceResult models.TaxInvoiceResult

// 	query := config.DB_POS.Table("pos_transaction pt").
// 		Select(`
//         DATE(pt.bill_date) AS bill_date,
//         pt.bill_location,
//         MIN(pt.bill_no) AS min_bill_no,
//         MAX(pt.bill_no) AS max_bill_no,
//         SUM(pt.product_price) AS total,
//         ROUND((CAST(SUM(pt.product_price) AS numeric) * 7) / 107, 2) AS tax,
//         ROUND(CAST(SUM(pt.product_price) AS numeric) - ((CAST(SUM(pt.product_price) AS numeric) * 7) / 107), 2) AS before_vat
//     `).
// 		Where("pt.bill_status = 'Normal' AND pt.pos_type != 'kiosk'")

// 	if params.Location != "" {
// 		query = query.Where("pt.bill_location ILIKE ?", "%"+params.Location+"%")
// 	}

// 	if params.StartDate != "" && params.EndDate != "" {
// 		query = query.Where("pt.bill_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)

// 	}

// 	query = query.Group("DATE(pt.bill_date), pt.bill_location")
// 	// Count
// 	var totalCount int64
// 	if err := query.Count(&totalCount).Error; err != nil {
// 		return result, fmt.Errorf("count query failed: %w", err)
// 	}
// 	var totalPrice int64
// 	if err := query.Session(&gorm.Session{}).
// 		Select("COALESCE(SUM(pt.product_price), 0)").Scan(&totalPrice).Error; err != nil {
// 		return result, fmt.Errorf("failed to calculate total price: %w", err)
// 	}

// 	var totalTax float64
// 	if err := query.Session(&gorm.Session{}).
// 		Select("COALESCE(ROUND((CAST(SUM(pt.product_price) AS numeric) * 7) / 107, 2), 0)").Scan(&totalTax).Error; err != nil {
// 		return result, fmt.Errorf("failed to calculate total tax: %w", err)
// 	}

// 	var totalBeforeVat float64
// 	if err := query.Session(&gorm.Session{}).
// 		Select("COALESCE(ROUND(CAST(SUM(pt.product_price) AS numeric) - ((CAST(SUM(pt.product_price) AS numeric) * 7) / 107), 2), 0)").Scan(&totalBeforeVat).Error; err != nil {
// 		return result, fmt.Errorf("failed to calculate total before VAT: %w", err)
// 	}

// 	// Pagination
// 	offset := (params.Page - 1) * params.Skip
// 	if err := query.Order("DATE(pt.bill_date), pt.bill_location").Limit(params.Skip).Offset(offset).Find(&taxInvoices).Error; err != nil {
// 		return result, fmt.Errorf("data query failed: %w", err)
// 	}

// 	taxInvoiceResult.Data = taxInvoices
// 	taxInvoiceResult.TotalPrice = totalPrice
// 	taxInvoiceResult.TotalTax = totalTax
// 	taxInvoiceResult.TotalBeforeVat = totalBeforeVat

// 	result = utils.SearchResult{
// 		Page:       params.Page,
// 		TotalCount: int(totalCount),
// 		Result:     taxInvoiceResult,
// 	}
// 	return result, nil
// }

func TaxInvoiceReport(params models.SearchTaxInvoiceReport) (utils.SearchResult, error) {
	var result utils.SearchResult
	var taxInvoices []models.TaxInvoiceSummaryReport
	var taxInvoiceResult models.TaxInvoiceResult

	// ---------------- Base Query ----------------
	query := config.DB_POS.Table("pos_transaction pt").
		Where("pt.bill_status = ? AND pt.pos_type != ?", "Normal", "kiosk")

	if params.Location != "" {
		query = query.Where("pt.bill_location ILIKE ?", "%"+params.Location+"%")
	}

	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where(
			"pt.bill_date::date BETWEEN ? AND ?",
			params.StartDate,
			params.EndDate,
		)
	}

	// ---------------- Summary Query (ไม่มี GROUP BY) ----------------
	summaryQuery := query.Session(&gorm.Session{})

	var totalPrice int64
	if err := summaryQuery.
		Select("COALESCE(SUM(pt.product_price), 0)").
		Scan(&totalPrice).Error; err != nil {
		return result, fmt.Errorf("failed to calculate total price: %w", err)
	}

	var totalTax float64
	if err := summaryQuery.
		Select(`
			COALESCE(
				ROUND((SUM(pt.product_price)::numeric * 7) / 107, 2),
				0
			)
		`).
		Scan(&totalTax).Error; err != nil {
		return result, fmt.Errorf("failed to calculate total tax: %w", err)
	}

	var totalBeforeVat float64
	if err := summaryQuery.
		Select(`
			COALESCE(
				ROUND(
					SUM(pt.product_price)::numeric -
					((SUM(pt.product_price)::numeric * 7) / 107),
					2
				),
				0
			)
		`).
		Scan(&totalBeforeVat).Error; err != nil {
		return result, fmt.Errorf("failed to calculate total before VAT: %w", err)
	}

	// ---------------- Data Query ----------------
	query = query.Select(`
		DATE(pt.bill_date) AS bill_date,
		pt.bill_location,
		MIN(pt.bill_no) AS min_bill_no,
		MAX(pt.bill_no) AS max_bill_no,
		SUM(pt.product_price) AS total,
		ROUND((SUM(pt.product_price)::numeric * 7) / 107, 2) AS tax,
		ROUND(
			SUM(pt.product_price)::numeric -
			((SUM(pt.product_price)::numeric * 7) / 107),
			2
		) AS before_vat
	`).Group("DATE(pt.bill_date), pt.bill_location")

	// ---------------- Count ----------------
	var totalCount int64
	countQuery := config.DB_POS.Table("(?) as t", query.Session(&gorm.Session{}))
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	// ---------------- Pagination ----------------
	offset := (params.Page - 1) * params.Skip
	if err := query.
		Order("DATE(pt.bill_date), pt.bill_location").
		Limit(params.Skip).
		Offset(offset).
		Find(&taxInvoices).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	// ---------------- Result ----------------
	taxInvoiceResult.Data = taxInvoices
	taxInvoiceResult.TotalPrice = totalPrice
	taxInvoiceResult.TotalTax = totalTax
	taxInvoiceResult.TotalBeforeVat = totalBeforeVat

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     taxInvoiceResult,
	}

	return result, nil
}

func ExportTaxInvoiceReport(params models.SearchExportTaxInvoiceReport) ([]models.TaxInvoiceSummaryReport, error) {
	var taxInvoices []models.TaxInvoiceSummaryReport

	query := config.DB_POS.Table("pos_transaction pt").
		Select(`
			DATE(pt.bill_date) AS bill_date,
			pt.bill_location,
			MIN(pt.bill_no) AS min_bill_no,
			MAX(pt.bill_no) AS max_bill_no,
			SUM(pt.product_price) AS total,
			ROUND((CAST(SUM(pt.product_price) AS numeric) * 7) / 107, 2) AS tax,
        ROUND(CAST(SUM(pt.product_price) AS numeric) - ((CAST(SUM(pt.product_price) AS numeric) * 7) / 107), 2) AS before_vat`).
		Where("pt.bill_status = 'Normal' AND pt.pos_type != 'kiosk'")

	if params.Location != "" {
		query = query.Where("pt.bill_location ILIKE ?", "%"+params.Location+"%")
	}

	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("pt.bill_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	query = query.Group("DATE(pt.bill_date), pt.bill_location").Order("DATE(pt.bill_date), pt.bill_location")

	// ✅ สำคัญ: รัน query และดึงข้อมูล
	if err := query.Find(&taxInvoices).Error; err != nil {
		return nil, err
	}

	return taxInvoices, nil
}

func ExportToPDF(c *gin.Context, data []models.TaxInvoiceSummaryReport, params models.SearchExportTaxInvoiceReport) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
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

	// ขนาดคอลัมน์
	type widths struct{ date, from, to, branch, value, vat, total float64 }
	colW := widths{22, 40, 40, 18, 25, 22, 25}
	lineH := 8.0

	// แปลงวันที่เป็น YYYY/MM/DD
	startDate := strings.ReplaceAll(params.StartDate, "-", "/")
	endDate := strings.ReplaceAll(params.EndDate, "-", "/")
	dataNow := utils.TimeNowAsia()

	// ฟังก์ชันวาดหัวรายงาน
	addHeader := func() {
		// ชื่อรายงาน
		pdf.SetFont(fontName, "B", 18)
		pdf.CellFormat(0, 12, "รายงานใบกำกับภาษีอย่างย่อแบบสรุป เรียงรายวัน", "", 1, "C", false, 0, "")
		pdf.Ln(1)

		// ช่วงวันที่
		pdf.SetFont(fontName, "B", 14)
		pdf.CellFormat(
			0, 8,
			fmt.Sprintf("ตั้งแต่วันที่ : %s    ถึงวันที่ : %s", startDate, endDate),
			"", 1, "C", false, 0, "",
		)

		pdf.Ln(1)

		// วันที่พิมพ์
		nowStr := "พิมพ์วันที่ : " + dataNow.Format("02/01/2006 15:04:05")
		pdf.SetFont(fontName, "", 12)
		pdf.CellFormat(0, 6, nowStr, "", 1, "R", false, 0, "")

		// เส้นคั่น
		pdf.Ln(1)
		pdf.SetLineWidth(0.2)
		pdf.Line(12, pdf.GetY(), 210-12, pdf.GetY())
		pdf.Ln(2)

		// หัวคอลัมน์ตาราง
		pdf.SetFont(fontName, "B", 12.5)
		pdf.SetFillColor(245, 245, 245)
		pdf.CellFormat(colW.date, lineH, "วันที่", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW.from, lineH, "ตั้งแต่เลขที่", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW.to, lineH, "ถึงเลขที่", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW.branch, lineH, "สาขา", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW.value, lineH, "มูลค่าสินค้า", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW.vat, lineH, "ภาษี", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW.total, lineH, "รวมมูลค่า", "1", 1, "C", true, 0, "")
		pdf.SetFont(fontName, "", 12.5)
	}

	// Footer: เลขหน้า
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont(fontName, "", 11)
		pdf.CellFormat(0, 10, fmt.Sprintf("หน้า %d", pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	// หน้าแรก
	pdf.AddPage()
	addHeader()

	// แถวข้อมูล
	for _, r := range data {
		if pdf.GetY() > 280 {
			pdf.AddPage()
			addHeader()
		}
		pdf.CellFormat(colW.date, lineH, r.BillDate.Format("02/01/2006"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colW.from, lineH, r.MinBillNo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colW.to, lineH, r.MaxBillNo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colW.branch, lineH, r.BillLocation, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colW.value, lineH, thMoney(r.Total), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colW.vat, lineH, thMoney(r.Tax), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colW.total, lineH, thMoney(r.BeforeVat), "1", 1, "R", false, 0, "")
	}

	// รวมยอดท้าย
	var sumValue, sumVat, sumTotal float64
	for _, r := range data {
		sumValue += r.Total
		sumVat += r.Tax
		sumTotal += r.BeforeVat
	}
	pdf.SetFont(fontName, "B", 12.5)
	pdf.CellFormat(colW.date+colW.from+colW.to+colW.branch, lineH, "รวม", "1", 0, "R", false, 0, "")
	pdf.CellFormat(colW.value, lineH, thMoney(sumValue), "1", 0, "R", false, 0, "")
	pdf.CellFormat(colW.vat, lineH, thMoney(sumVat), "1", 0, "R", false, 0, "")
	pdf.CellFormat(colW.total, lineH, thMoney(sumTotal), "1", 1, "R", false, 0, "")

	// ส่งออก PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="report.pdf"`)
	c.Header("Content-Transfer-Encoding", "binary")
	return pdf.Output(c.Writer)
}

func thMoney(n float64) string {
	// รูปแบบ 1,234,567.89
	s := fmt.Sprintf("%.2f", n)
	// ใส่ comma ด้วยวิธีง่าย ๆ
	intPart, frac := s[:len(s)-3], s[len(s)-3:]
	out := ""
	for i, c := range reverse(intPart) {
		if i != 0 && i%3 == 0 {
			out = "," + out
		}
		out = string(c) + out
	}
	return out + frac
}
func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func ExportTaxInvoiceToExcel(c *gin.Context, data []models.TaxInvoiceSummaryReport, params models.SearchExportTaxInvoiceReport) error {
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

	rightAlignSmallStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: false,
			Size: 10,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	startDate := strings.ReplaceAll(params.StartDate, "-", "/")
	endDate := strings.ReplaceAll(params.EndDate, "-", "/")
	dataNow := utils.TimeNowAsia()

	// เพิ่มหัวตาราง 3 บรรทัด
	f.MergeCell(sheet, "A1", "G1") // รวม 7 คอลัมน์ (A-G)
	f.SetCellValue(sheet, "A1", "รายงานใบกำกับภาษีอย่างย่อแบบสรุป เรียงรายวัน")
	f.SetCellStyle(sheet, "A1", "A1", headerStyle)

	f.MergeCell(sheet, "A2", "G2")
	f.SetCellValue(sheet, "A2", fmt.Sprintf("ตั้งแต่วันที่ : %s ถึงวันที่ : %s", startDate, endDate))
	f.SetCellStyle(sheet, "A2", "A2", boldCenterStyle)

	f.MergeCell(sheet, "A3", "G3")
	printDate := dataNow.Format("02/01/2006 15:04:05")
	f.SetCellValue(sheet, "A3", "พิมพ์วันที่ : "+printDate)
	f.SetCellStyle(sheet, "A3", "A3", rightAlignSmallStyle)

	// ตารางหัวตารางจริงเริ่มที่แถว 5 (เว้นวรรค 1 บรรทัด)
	headers := []string{"วันที่", "ตั้งแต่เลขที่", "ถึงเลขที่", "สาขา", "มูลค่าสินค้า", "ภาษี", "รวมมูลค่า"}
	for col, header := range headers {
		cell := fmt.Sprintf("%s5", string(rune('A'+col))) // A5, B5, ...
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, boldCenterStyle)
	}

	// Style วันที่ และตัวเลข เหมือนเดิม
	format := "dd/mm/yyyy"
	dateStyle, err := f.NewStyle(&excelize.Style{
		CustomNumFmt: &format,
	})
	if err != nil {
		return err
	}

	numberFormat := "#,##0.00"
	numberStyle, err := f.NewStyle(&excelize.Style{
		CustomNumFmt: &numberFormat,
	})
	if err != nil {
		return err
	}

	// เริ่มเขียนข้อมูลจริงจากแถว 6
	for i, row := range data {
		rowNum := i + 6

		dateCell := fmt.Sprintf("A%d", rowNum)
		f.SetCellValue(sheet, dateCell, row.BillDate)
		f.SetCellStyle(sheet, dateCell, dateCell, dateStyle)

		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), row.MinBillNo)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), row.MaxBillNo)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowNum), row.BillLocation)

		cellTotal := fmt.Sprintf("E%d", rowNum)
		f.SetCellValue(sheet, cellTotal, row.Total)
		f.SetCellStyle(sheet, cellTotal, cellTotal, numberStyle)

		cellTax := fmt.Sprintf("F%d", rowNum)
		f.SetCellValue(sheet, cellTax, row.Tax)
		f.SetCellStyle(sheet, cellTax, cellTax, numberStyle)

		cellBeforeVat := fmt.Sprintf("G%d", rowNum)
		f.SetCellValue(sheet, cellBeforeVat, row.BeforeVat)
		f.SetCellStyle(sheet, cellBeforeVat, cellBeforeVat, numberStyle)
	}

	// คำนวณผลรวม
	var sumTotal, sumTax, sumBeforeVat float64
	for _, row := range data {
		sumTotal += row.Total
		sumTax += row.Tax
		sumBeforeVat += row.BeforeVat
	}

	// แถวสำหรับผลรวม (แถวหลังข้อมูลสุดท้าย)
	totalRow := len(data) + 6

	// ใส่ข้อความ "รวมทั้งหมด" ในคอลัมน์ A (หรือจะให้อยู่ซ้ายสุดที่ D ก็ได้)
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "รวมทั้งหมด")
	// รวม 7 คอลัมน์ อาจ merge cell A-D ด้วยก็ได้ (ถ้าต้องการให้ข้อความยาวขึ้น)
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("D%d", totalRow))
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("D%d", totalRow), boldCenterStyle)

	// ใส่ผลรวมในแต่ละคอลัมน์
	f.SetCellValue(sheet, fmt.Sprintf("E%d", totalRow), sumTotal)
	f.SetCellStyle(sheet, fmt.Sprintf("E%d", totalRow), fmt.Sprintf("E%d", totalRow), numberStyle)

	f.SetCellValue(sheet, fmt.Sprintf("F%d", totalRow), sumTax)
	f.SetCellStyle(sheet, fmt.Sprintf("F%d", totalRow), fmt.Sprintf("F%d", totalRow), numberStyle)

	f.SetCellValue(sheet, fmt.Sprintf("G%d", totalRow), sumBeforeVat)
	f.SetCellStyle(sheet, fmt.Sprintf("G%d", totalRow), fmt.Sprintf("G%d", totalRow), numberStyle)

	// เขียนลง buffer และส่ง response
	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		return err
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=report.xlsx")
	c.Header("Content-Length", fmt.Sprintf("%d", buffer.Len()))
	c.Data(http.StatusOK, "application/octet-stream", buffer.Bytes())
	return nil
}

const (
	TRANSACTION_TYPE_NORMAL = "Normal"
	TRANSACTION_TYPE_VOID   = "Void"
)

func SpendingSearchList(params models.SearchSpendingParams) (utils.SearchResult, error) {
	var result utils.SearchResult
	var transactions []models.SpendingReport
	var spendingReport models.SpendingReportData

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("pos_transaction pt").
		Select(`pt.bill_no,
				pt.bill_date,
				pt.member_tel,
				pt.card_no,
				coalesce(pt.product_price,0) as product_price,
				coalesce(pt.e_coin,0) as e_coin,
				coalesce(pt.e_bonus,0) as e_bonus,
				pt.bill_location,
				b.branch_name ,
				pt.bill_status ,
				ps.pos_code ,
				pt.pos_id,
				pt.cashier,
				mp.name as payment_type_name,
				pt.bank_detail`).
		Joins("LEFT JOIN pos_station ps ON pt.pos_id = CAST(ps.id AS VARCHAR)").
		Joins("LEFT JOIN branch b on pt.bill_location = b.branch_code").
		Joins("LEFT JOIN master_payment mp on pt.bill_payment_id  = mp.id").
		Where("pt.bill_status = 'Normal'")

	if params.MemberTel != "" {
		tel := "%" + params.MemberTel + "%"
		query = query.Where("(pt.member_tel ILIKE ?)", tel)
	}

	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("pt.bill_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	var totalPrice int64
	if err := query.Session(&gorm.Session{}).
		Select("COALESCE(SUM(pt.product_price), 0)").Scan(&totalPrice).Error; err != nil {
		return result, fmt.Errorf("failed to calculate total price: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("pt.bill_date DESC").Limit(params.Skip).Offset(offset).Find(&transactions).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	spendingReport.TotalPrice = totalPrice
	spendingReport.Data = transactions

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     spendingReport,
	}
	return result, nil
}

func ExportSpendingList(params models.ExportSpendingParams) ([]models.ExportSpendingReport, error) {

	var transactions []models.ExportSpendingReport

	if config.DB_POS == nil {
		return transactions, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("pos_transaction pt").
		Select(`pt.bill_no,
				pt.bill_date,
				pt.member_tel,
				pt.card_no,
				coalesce(pt.product_price,0) as product_price,
				coalesce(pt.e_coin,0) as e_coin,
				coalesce(pt.e_bonus,0) as e_bonus,
				pt.bill_location,
				b.branch_name ,
				pt.bill_status ,
				ps.pos_code ,
				pt.pos_id,
				pt.cashier,
				mp.name as payment_type_name,
				pt.bank_detail,
				m.id as member_id`).
		Joins("LEFT JOIN pos_station ps ON pt.pos_id = CAST(ps.id AS VARCHAR)").
		Joins("LEFT JOIN branch b on pt.bill_location = b.branch_code").
		Joins("LEFT JOIN master_payment mp on pt.bill_payment_id  = mp.id").
		Joins("LEFT JOIN member m on pt.member_tel = m.tel").
		Where("pt.bill_status = 'Normal'")

	if params.MemberTel != "" {
		tel := "%" + params.MemberTel + "%"
		query = query.Where("(pt.member_tel ILIKE ?)", tel)
	}

	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("pt.bill_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	if err := query.Order("pt.bill_date DESC").Find(&transactions).Error; err != nil {
		return transactions, fmt.Errorf("data query failed: %w", err)
	}

	return transactions, nil
}

func ExportSpendingToExcel(c *gin.Context, data []models.ExportSpendingReport) error {
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
		headers := []interface{}{"เลขที่บิล", "วันที่บิล", "สาขา", "Cashier", "Card", "Member Id", "ยอดรวม", "ecoin", "ebonus", "Payment", "Bank", "POS"}
		if err := sw.SetRow("A1", headers, excelize.RowOpts{StyleID: boldCenterStyle}); err != nil {
			return err
		}

		// เขียน row ข้อมูล
		for i, row := range data[start:end] {
			memberID := 0

			if row.MemberID != nil {
				memberID = *row.MemberID
			}
			rowNum := i + 2
			rowCells := []interface{}{
				row.BillNo,
				row.BillDate,
				row.BillLocation,
				row.Cashier,
				row.CardNo,
				memberID,
				row.ProductPrice,
				row.ECoin,
				row.EBonus,
				row.PaymentTypeName,
				row.BankDetail,
				row.PosCode,
			}

			cell, _ := excelize.CoordinatesToCellName(1, rowNum)
			if err := sw.SetRow(cell, rowCells); err != nil {
				return err
			}

			// ตั้ง style ให้วันที่และคะแนน
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", rowNum), fmt.Sprintf("B%d", rowNum), dateTimeStyle)
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

	c.Header("Content-Disposition", "attachment; filename=spending_report.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())

	return nil
}

func GroupPosDailySalesSummaryReport(params models.SearchPosSaleReport) (models.PosDailySalesSummaryReport, error) {
	var transactions []models.GroupPosDailySalesSummaryReport
	var dailySalesReport models.PosDailySalesSummaryReport

	if config.DB_POS == nil {
		return dailySalesReport, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("pos_transaction pt").
		Select(`to_char(pt.bill_date, 'YYYY-MM-DD') as date,
				COALESCE(SUM(pt.product_price), 0) as product_price,
				COALESCE(SUM(pt.e_coin), 0) as e_coin,
				COALESCE(SUM(pt.e_bonus), 0) as e_bonus,
				pt.pos_id,
				ps.pos_code,
				mp.name as bill_payment`).
		Joins("LEFT JOIN pos_station ps ON pt.pos_id = CAST(ps.id AS VARCHAR)").
		Joins("LEFT JOIN master_payment mp on pt.bill_payment_id  = mp.id").
		Where("pt.bill_status = 'Normal' and pt.pos_type != 'kiosk'")

	if params.Location != "" {
		query = query.Where("pt.bill_location ILIKE ?", "%"+params.Location+"%")
	}

	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("pt.bill_date::date >= ? AND pt.bill_date::date <= ?", params.StartDate, params.EndDate)
	}

	type SummaryTotal struct {
		TotalPrice  int64
		TotalECoin  int64
		TotalEBonus int64
	}
	var total SummaryTotal

	if err := query.Session(&gorm.Session{}).
		Select(`
		COALESCE(SUM(pt.product_price),0) as total_price,
		COALESCE(SUM(pt.e_coin),0) as total_e_coin,
		COALESCE(SUM(pt.e_bonus),0) as total_e_bonus
	`).
		Scan(&total).Error; err != nil {
		return dailySalesReport, err
	}

	query = query.Group(`
	to_char(pt.bill_date, 'YYYY-MM-DD'),
	pt.pos_id,
	ps.pos_code,
	mp.name
`)

	if err := query.Order("date ASC").Find(&transactions).Error; err != nil {
		return dailySalesReport, fmt.Errorf("data query failed: %w", err)
	}

	dailySalesReport.Data = transactions
	dailySalesReport.TotalPrice = total.TotalPrice
	dailySalesReport.TotalECoin = total.TotalECoin
	dailySalesReport.TotalEBonus = total.TotalEBonus

	return dailySalesReport, nil
}

func GroupPosMonthlySalesSummaryReport(params models.SearchPosSaleReport) (models.PosMonthlySalesSummaryReport, error) {
	var transactions []models.GroupPosMonthlySalesSummaryReport
	var monthlySalesReport models.PosMonthlySalesSummaryReport

	if config.DB_POS == nil {
		return monthlySalesReport, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("pos_transaction pt").
		Select(`to_char(pt.bill_date, 'YYYY-MM') as date,
				COALESCE(SUM(pt.product_price), 0) as product_price,
				COALESCE(SUM(pt.e_coin), 0) as e_coin,
				COALESCE(SUM(pt.e_bonus), 0) as e_bonus,
				mp.name as bill_payment`).
		Joins("LEFT JOIN pos_station ps ON pt.pos_id = CAST(ps.id AS VARCHAR)").
		Joins("LEFT JOIN master_payment mp on pt.bill_payment_id  = mp.id").
		Where("pt.bill_status = 'Normal' and pt.pos_type != 'kiosk'")

	if params.Location != "" {
		query = query.Where("pt.bill_location ILIKE ?", "%"+params.Location+"%")
	}

	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("pt.bill_date::date >= ? AND pt.bill_date::date <= ?", params.StartDate, params.EndDate)
	}

	type SummaryTotal struct {
		TotalPrice  int64
		TotalECoin  int64
		TotalEBonus int64
	}
	var total SummaryTotal

	if err := query.Session(&gorm.Session{}).
		Select(`
		COALESCE(SUM(pt.product_price),0) as total_price,
		COALESCE(SUM(pt.e_coin),0) as total_e_coin,
		COALESCE(SUM(pt.e_bonus),0) as total_e_bonus
	`).
		Scan(&total).Error; err != nil {
		return monthlySalesReport, err
	}

	query = query.Group(`
	to_char(pt.bill_date, 'YYYY-MM'),
	mp.name
`)

	if err := query.Order("date ASC").Find(&transactions).Error; err != nil {
		return monthlySalesReport, fmt.Errorf("data query failed: %w", err)
	}

	monthlySalesReport.Data = transactions
	monthlySalesReport.TotalPrice = total.TotalPrice
	monthlySalesReport.TotalECoin = total.TotalECoin
	monthlySalesReport.TotalEBonus = total.TotalEBonus

	return monthlySalesReport, nil
}

func PosSalePaymentReport(params models.SearchPosSalePaymentReport) (utils.SearchResult, error) {
	var result utils.SearchResult
	var transactions []models.SpendingReport
	var paymentReport models.SpendingReportData

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("pos_transaction pt").
		Select(`pt.bill_no,
				pt.bill_date,
				pt.member_tel,
				pt.card_no,
				coalesce(pt.product_price,0) as product_price,
				coalesce(pt.e_coin,0) as e_coin,
				coalesce(pt.e_bonus,0) as e_bonus,
				pt.bill_location,
				b.branch_name ,
				pt.bill_status ,
				ps.pos_code ,
				pt.pos_id,
				pt.cashier,
				mp.name as payment_type_name,
				pt.bank_detail`).
		Joins("LEFT JOIN pos_station ps ON pt.pos_id = CAST(ps.id AS VARCHAR)").
		Joins("LEFT JOIN branch b on pt.bill_location = b.branch_code").
		Joins("LEFT JOIN master_payment mp on pt.bill_payment_id  = mp.id").
		Where("pt.bill_status = 'Normal' and pt.pos_type != 'kiosk'")
	if params.Location != "" {
		query = query.Where("pt.bill_location ILIKE ?", "%"+params.Location+"%")
	}

	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("pt.bill_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	if params.PaymentTypeId != "" {
		if _, err := uuid.Parse(params.PaymentTypeId); err != nil {
			return result, fmt.Errorf("invalid payment_type_id")
		}
		query = query.Where("pt.bill_payment_id = ?::uuid", params.PaymentTypeId)
	}
	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	var totalPrice int64
	if err := query.Session(&gorm.Session{}).
		Select("COALESCE(SUM(pt.product_price), 0)").Scan(&totalPrice).Error; err != nil {
		return result, fmt.Errorf("failed to calculate total price: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Skip
	if err := query.Order("pt.bill_date DESC").Limit(params.Skip).Offset(offset).Find(&transactions).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	paymentReport.TotalPrice = totalPrice
	paymentReport.Data = transactions

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     paymentReport,
	}
	return result, nil
}

func ExportPosSalePaymentReport(params models.ExportPosSalePaymentReport) ([]models.ExportSpendingReport, error) {
	var transactions []models.ExportSpendingReport

	if config.DB_POS == nil {
		return transactions, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("pos_transaction pt").
		Select(`pt.bill_no,
				pt.bill_date,
				pt.member_tel,
				m.id as member_id,
				pt.card_no,
				coalesce(pt.product_price,0) as product_price,
				coalesce(pt.e_coin,0) as e_coin,
				coalesce(pt.e_bonus,0) as e_bonus,
				pt.bill_location,
				b.branch_name ,
				pt.bill_status ,
				ps.pos_code ,
				pt.pos_id,
				pt.cashier,
				mp.name as payment_type_name,
				pt.bank_detail`).
		Joins("LEFT JOIN pos_station ps ON pt.pos_id = CAST(ps.id AS VARCHAR)").
		Joins("LEFT JOIN branch b on pt.bill_location = b.branch_code").
		Joins("LEFT JOIN master_payment mp on pt.bill_payment_id  = mp.id").
		Joins("left join member m on pt.member_tel = m.tel").
		Where("pt.bill_status = 'Normal' and pt.pos_type != 'kiosk'")
	if params.Location != "" {
		query = query.Where("pt.bill_location ILIKE ?", "%"+params.Location+"%")
	}

	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("pt.bill_date::date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}

	if params.PaymentTypeId != "" {
		if _, err := uuid.Parse(params.PaymentTypeId); err != nil {
			return transactions, fmt.Errorf("invalid payment_type_id")
		}
		query = query.Where("pt.bill_payment_id = ?::uuid", params.PaymentTypeId)
	}

	if err := query.Order("pt.bill_date DESC").Find(&transactions).Error; err != nil {
		return transactions, fmt.Errorf("data query failed: %w", err)
	}

	return transactions, nil
}

func ExportPosSalePaymentToExcel(c *gin.Context, data []models.ExportSpendingReport) error {
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
		headers := []interface{}{"เลขที่บิล", "วันที่บิล", "สาขา", "Cashier", "Card ID", "Member ID", "ยอดรวม", "ecoin", "ebonus", "Payment", "Bank", "POS"}
		if err := sw.SetRow("A1", headers, excelize.RowOpts{StyleID: boldCenterStyle}); err != nil {
			return err
		}
		memberID := 0

		// เขียน row ข้อมูล
		for i, row := range data[start:end] {
			if row.MemberID != nil {
				memberID = *row.MemberID
			}
			rowNum := i + 2
			rowCells := []interface{}{
				row.BillNo,
				row.BillDate,
				row.BranchName,
				row.Cashier,
				row.CardNo,
				memberID,
				row.ProductPrice,
				row.ECoin,
				row.EBonus,
				row.PaymentTypeName,
				row.BankDetail,
				row.PosCode,
			}

			cell, _ := excelize.CoordinatesToCellName(1, rowNum)
			if err := sw.SetRow(cell, rowCells); err != nil {
				return err
			}

			// ตั้ง style ให้วันที่และคะแนน
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", rowNum), fmt.Sprintf("B%d", rowNum), dateTimeStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("H%d", rowNum), numberStyle)
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

	c.Header("Content-Disposition", "attachment; filename=pos_sale_payment_report.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())

	return nil
}

func ExportPOSDailyToPDF(
	c *gin.Context,
	data models.PosDailySalesSummaryReport,
	params models.SearchPosSaleReport,
	branchName string,
) error {

	// =========================================================
	// DATA
	// =========================================================

	reportData := data.Data

	log.Printf(
		"ExportPOSDailyToPDF: reportData length: %d",
		len(reportData),
	)

	log.Printf(
		"ExportPOSDailyToPDF data: %+v",
		reportData,
	)

	// =========================================================
	// PDF
	// =========================================================

	pdf := fpdf.New("P", "mm", "A4", "")

	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 15)

	// =========================================================
	// FONT
	// =========================================================

	fontName := "THSarabunNew"
	fontPath := "./fonts/THSarabunNew/THSarabunNew.ttf"

	if _, err := os.Stat(fontPath); err != nil {
		log.Printf(
			"ไม่พบฟอนต์ %s: %v",
			fontPath,
			err,
		)

		return fmt.Errorf(
			"ไม่พบฟอนต์ %s: %w",
			fontPath,
			err,
		)
	}

	// Regular
	pdf.AddUTF8Font(
		fontName,
		"",
		fontPath,
	)

	// Bold
	pdf.AddUTF8Font(
		fontName,
		"B",
		fontPath,
	)

	// =========================================================
	// PAGE
	// =========================================================

	pdf.AddPage()

	pdf.SetFont(
		fontName,
		"",
		14,
	)

	// =========================================================
	// TITLE
	// =========================================================

	pdf.SetFont(
		fontName,
		"B",
		20,
	)

	pdf.CellFormat(
		190,
		10,
		"รายงานยอดขาย",
		"",
		1,
		"C",
		false,
		0,
		"",
	)

	pdf.Ln(3)

	// =========================================================
	// REPORT HEADER
	// =========================================================

	pdf.SetFont(
		fontName,
		"B",
		12,
	)

	// ตั้งแต่
	pdf.CellFormat(
		18,
		7,
		"ตั้งแต่ :",
		"",
		0,
		"L",
		false,
		0,
		"",
	)

	pdf.SetFont(
		fontName,
		"",
		12,
	)

	pdf.CellFormat(
		45,
		7,
		params.StartDate,
		"",
		0,
		"L",
		false,
		0,
		"",
	)

	// ถึง
	pdf.SetFont(
		fontName,
		"B",
		12,
	)

	pdf.CellFormat(
		10,
		7,
		"ถึง :",
		"",
		0,
		"L",
		false,
		0,
		"",
	)

	pdf.SetFont(
		fontName,
		"",
		12,
	)

	pdf.CellFormat(
		45,
		7,
		params.EndDate,
		"",
		1,
		"L",
		false,
		0,
		"",
	)

	// สาขา
	pdf.SetFont(
		fontName,
		"B",
		12,
	)

	pdf.CellFormat(
		18,
		7,
		"สาขา :",
		"",
		0,
		"L",
		false,
		0,
		"",
	)

	pdf.SetFont(
		fontName,
		"",
		12,
	)

	pdf.CellFormat(
		30,
		7,
		params.Location,
		"",
		0,
		"L",
		false,
		0,
		"",
	)

	pdf.CellFormat(
		70,
		7,
		branchName,
		"",
		1,
		"L",
		false,
		0,
		"",
	)

	pdf.Ln(5)

	// =========================================================
	// SORT DATA
	// =========================================================

	sort.Slice(
		reportData,
		func(i, j int) bool {

			// Date
			if reportData[i].Date != reportData[j].Date {
				return reportData[i].Date <
					reportData[j].Date
			}

			// POS
			if reportData[i].PosCode != reportData[j].PosCode {
				return reportData[i].PosCode <
					reportData[j].PosCode
			}

			// Payment
			return reportData[i].BillPayment <
				reportData[j].BillPayment
		},
	)

	// =========================================================
	// COLUMN WIDTH
	// =========================================================

	// รวม = 186 mm
	// A4 = 210
	// Margin ซ้าย/ขวา = 10 + 10
	// พื้นที่ใช้งาน = 190 mm

	colDate := 28.0
	colPOS := 25.0
	colPayment := 43.0
	colSale := 40.0
	colCredit := 30.0
	colToken := 20.0

	// =========================================================
	// TABLE HEADER
	// =========================================================

	drawHeader := func() {

		pdf.SetFont(
			fontName,
			"B",
			12,
		)

		pdf.SetFillColor(
			180,
			200,
			220,
		)

		// วันที่
		pdf.CellFormat(
			colDate,
			8,
			"วันที่",
			"1",
			0,
			"L",
			true,
			0,
			"",
		)

		// POS station
		pdf.CellFormat(
			colPOS,
			8,
			"POS station",
			"1",
			0,
			"L",
			true,
			0,
			"",
		)

		// Payment
		pdf.CellFormat(
			colPayment,
			8,
			"Payment",
			"1",
			0,
			"L",
			true,
			0,
			"",
		)

		// ยอดขาย
		pdf.CellFormat(
			colSale,
			8,
			"ยอดขาย",
			"1",
			0,
			"R",
			true,
			0,
			"",
		)

		// Credit
		pdf.CellFormat(
			colCredit,
			8,
			"ECoin",
			"1",
			0,
			"R",
			true,
			0,
			"",
		)

		// Token
		pdf.CellFormat(
			colToken,
			8,
			"Bonus",
			"1",
			1,
			"R",
			true,
			0,
			"",
		)
	}

	// วาด Header ครั้งแรก
	drawHeader()

	// =========================================================
	// GRAND TOTAL
	// =========================================================

	var grandSale float64
	var grandCredit float64
	var grandToken float64

	// =========================================================
	// GROUP BY DATE
	// =========================================================

	dateMap := make(
		map[string][]models.GroupPosDailySalesSummaryReport,
	)

	for _, item := range reportData {

		dateMap[item.Date] =
			append(
				dateMap[item.Date],
				item,
			)
	}

	// =========================================================
	// SORT DATE
	// =========================================================

	dates := make(
		[]string,
		0,
		len(dateMap),
	)

	for date := range dateMap {
		dates = append(
			dates,
			date,
		)
	}

	sort.Strings(dates)

	// =========================================================
	// LOOP DATE
	// =========================================================

	for _, date := range dates {

		dateData := dateMap[date]

		// =====================================================
		// GROUP BY POS
		// =====================================================

		posMap := make(
			map[string][]models.GroupPosDailySalesSummaryReport,
		)

		for _, item := range dateData {

			posMap[item.PosCode] =
				append(
					posMap[item.PosCode],
					item,
				)
		}

		// =====================================================
		// SORT POS
		// =====================================================

		posCodes := make(
			[]string,
			0,
			len(posMap),
		)

		for posCode := range posMap {

			posCodes = append(
				posCodes,
				posCode,
			)
		}

		sort.Strings(posCodes)

		// วันที่แสดงแค่ POS แรกของวัน
		firstPOS := true

		// =====================================================
		// LOOP POS
		// =====================================================

		for _, posCode := range posCodes {

			posData := posMap[posCode]

			// =================================================
			// POS TOTAL
			// =================================================

			var posSale float64
			var posCredit float64
			var posToken float64

			for _, item := range posData {

				posSale += float64(
					item.ProductPrice,
				)

				posCredit += item.ECoin

				// ตอนนี้ Token ในข้อมูลไม่มี field
				// จึงให้เป็น 0 ตามรูปตัวอย่าง
				//
				// ถ้ามี field Token ภายหลัง
				// สามารถเปลี่ยนเป็น:
				//
				posToken += float64(item.EBonus)
			}

			// =================================================
			// PAYMENT ROW
			// =================================================

			for i, item := range posData {

				dateText := ""

				if firstPOS && i == 0 {

					dateText =
						formatDate(date)
				}

				posText := ""

				if i == 0 {

					posText =
						item.PosCode
				}

				pdf.SetFont(
					fontName,
					"",
					12,
				)

				// ---------------------------------------------
				// Date
				// ---------------------------------------------

				pdf.CellFormat(
					colDate,
					7,
					dateText,
					"",
					0,
					"L",
					false,
					0,
					"",
				)

				// ---------------------------------------------
				// POS
				// ---------------------------------------------

				pdf.CellFormat(
					colPOS,
					7,
					posText,
					"",
					0,
					"L",
					false,
					0,
					"",
				)

				// ---------------------------------------------
				// Payment
				// ---------------------------------------------

				pdf.CellFormat(
					colPayment,
					7,
					item.BillPayment,
					"",
					0,
					"L",
					false,
					0,
					"",
				)

				// ---------------------------------------------
				// Sale
				// ---------------------------------------------

				pdf.CellFormat(
					colSale,
					7,
					money(
						float64(item.ProductPrice),
					),
					"",
					0,
					"R",
					false,
					0,
					"",
				)

				// ---------------------------------------------
				// Credit
				// ---------------------------------------------

				pdf.CellFormat(
					colCredit,
					7,
					moneyNoDecimal(
						item.ECoin,
					),
					"",
					0,
					"R",
					false,
					0,
					"",
				)

				// ---------------------------------------------
				// Token
				// ---------------------------------------------

				pdf.CellFormat(
					colToken,
					7,
					moneyNoDecimal(
						item.EBonus,
					),
					"",
					1,
					"R",
					false,
					0,
					"",
				)

				// ---------------------------------------------
				// PAGE CHECK
				// ---------------------------------------------

				if pdf.GetY() > 270 {

					pdf.AddPage()

					drawHeader()
				}
			}

			// =================================================
			// BEFORE VAT
			// =================================================

			beforeVAT :=
				posSale / 1.07

			vat :=
				posSale - beforeVAT

			pdf.SetFont(
				fontName,
				"B",
				12,
			)

			// ---------------------------------------------
			// Before Vat
			// ---------------------------------------------

			pdf.CellFormat(
				colDate+colPOS,
				7,
				"",
				"",
				0,
				"L",
				false,
				0,
				"",
			)

			pdf.CellFormat(
				colPayment,
				7,
				"Before Vat",
				"T",
				0,
				"L",
				false,
				0,
				"",
			)

			pdf.CellFormat(
				colSale,
				7,
				money(beforeVAT),
				"T",
				0,
				"R",
				false,
				0,
				"",
			)

			pdf.CellFormat(
				colCredit+colToken,
				7,
				"",
				"T",
				1,
				"L",
				false,
				0,
				"",
			)

			// ---------------------------------------------
			// VAT
			// ---------------------------------------------

			pdf.CellFormat(
				colDate+colPOS,
				7,
				"",
				"",
				0,
				"L",
				false,
				0,
				"",
			)

			pdf.CellFormat(
				colPayment,
				7,
				"Vat 7%",
				"",
				0,
				"L",
				false,
				0,
				"",
			)

			pdf.CellFormat(
				colSale,
				7,
				money(vat),
				"",
				0,
				"R",
				false,
				0,
				"",
			)

			pdf.CellFormat(
				colCredit+colToken,
				7,
				"",
				"",
				1,
				"L",
				false,
				0,
				"",
			)

			// ---------------------------------------------
			// TOTAL
			// ---------------------------------------------

			pdf.CellFormat(
				colDate+colPOS,
				7,
				"",
				"",
				0,
				"L",
				false,
				0,
				"",
			)

			pdf.CellFormat(
				colPayment,
				7,
				"Total :",
				"",
				0,
				"L",
				false,
				0,
				"",
			)

			pdf.CellFormat(
				colSale,
				7,
				money(posSale),
				"",
				0,
				"R",
				false,
				0,
				"",
			)

			pdf.CellFormat(
				colCredit+colToken,
				7,
				"",
				"",
				1,
				"L",
				false,
				0,
				"",
			)

			// =================================================
			// GRAND TOTAL
			// =================================================

			grandSale += posSale
			grandCredit += posCredit
			grandToken += posToken

			firstPOS = false

			// =================================================
			// LINE
			// =================================================

			pdf.Line(
				10,
				pdf.GetY()+2,
				196,
				pdf.GetY()+2,
			)

			pdf.Ln(3)
		}
	}

	// =========================================================
	// GRAND TOTAL
	// =========================================================
	//
	// จัด column ให้ตรงกับตารางด้านบนโดยใช้ตัวแปรความกว้าง
	// colDate, colPOS, colPayment, colSale, colCredit, colToken
	// เดียวกับที่ใช้วาด header/แถวข้อมูล (รวมกันได้ 186mm พอดี)
	// แทนที่จะใช้เลขคงที่แบบเดิม (65, 55, 43, 30) ซึ่งไม่ตรงคอลัมน์
	// =========================================================

	grandBeforeVAT :=
		grandSale / 1.07

	grandVAT :=
		grandSale - grandBeforeVAT

	pdf.SetFont(
		fontName,
		"B",
		14,
	)

	// ---------------------------------------------------------
	// BEFORE VAT  (Date+POS | Payment | Sale | Credit+Token)
	// ---------------------------------------------------------

	pdf.CellFormat(
		colDate+colPOS,
		7,
		"",
		"",
		0,
		"L",
		false,
		0,
		"",
	)

	pdf.SetTextColor(
		80,
		130,
		220,
	)

	pdf.CellFormat(
		colPayment,
		7,
		"Before Vat :",
		"",
		0,
		"L",
		false,
		0,
		"",
	)

	pdf.CellFormat(
		colSale,
		7,
		money(grandBeforeVAT),
		"",
		0,
		"R",
		false,
		0,
		"",
	)

	pdf.CellFormat(
		colCredit+colToken,
		7,
		"",
		"",
		1,
		"R",
		false,
		0,
		"",
	)

	// ---------------------------------------------------------
	// VAT
	// ---------------------------------------------------------

	pdf.CellFormat(
		colDate+colPOS,
		7,
		"",
		"",
		0,
		"L",
		false,
		0,
		"",
	)

	pdf.CellFormat(
		colPayment,
		7,
		"VAT 7% :",
		"",
		0,
		"L",
		false,
		0,
		"",
	)

	pdf.CellFormat(
		colSale,
		7,
		money(grandVAT),
		"",
		0,
		"R",
		false,
		0,
		"",
	)

	pdf.CellFormat(
		colCredit+colToken,
		7,
		"",
		"",
		1,
		"R",
		false,
		0,
		"",
	)

	// ---------------------------------------------------------
	// TOTAL / CREDIT / TOKEN
	// ---------------------------------------------------------

	pdf.CellFormat(
		colDate+colPOS,
		7,
		"",
		"",
		0,
		"L",
		false,
		0,
		"",
	)

	pdf.SetTextColor(
		0,
		0,
		120,
	)

	pdf.CellFormat(
		colPayment,
		7,
		"Total :",
		"",
		0,
		"L",
		false,
		0,
		"",
	)

	pdf.CellFormat(
		colSale,
		7,
		money(grandSale),
		"",
		0,
		"R",
		false,
		0,
		"",
	)

	pdf.SetTextColor(
		0,
		120,
		0,
	)

	pdf.CellFormat(
		colCredit,
		7,
		moneyNoDecimal(grandCredit),
		"",
		0,
		"R",
		false,
		0,
		"",
	)

	pdf.SetTextColor(
		180,
		0,
		0,
	)

	pdf.CellFormat(
		colToken,
		7,
		moneyNoDecimal(grandToken),
		"",
		1,
		"R",
		false,
		0,
		"",
	)

	// Reset color
	pdf.SetTextColor(
		0,
		0,
		0,
	)

	// =========================================================
	// BOTTOM LINE
	// =========================================================

	pdf.Line(
		10,
		pdf.GetY()+3,
		196,
		pdf.GetY()+3,
	)

	// =========================================================
	// RESPONSE HEADER
	// =========================================================

	c.Header(
		"Content-Type",
		"application/pdf",
	)

	c.Header(
		"Content-Disposition",
		`attachment; filename="report_pos_daily_summary.pdf"`,
	)

	c.Header(
		"Content-Transfer-Encoding",
		"binary",
	)

	// =========================================================
	// OUTPUT PDF
	// =========================================================

	err := pdf.Output(c.Writer)

	if err != nil {

		log.Printf(
			"PDF Output Error: %v",
			err,
		)

		return fmt.Errorf(
			"generate PDF failed: %w",
			err,
		)
	}

	return nil
}
func money(v float64) string {
	return formatNumber(v, 2)
}

func moneyNoDecimal(v float64) string {
	return formatNumber(v, 0)
}

func formatNumber(v float64, decimals int) string {

	// format ตัวเลขก่อน เช่น
	// 3150 -> "3150.00"
	// 6514.02 -> "6514.02"
	s := fmt.Sprintf("%.*f", decimals, v)

	negative := false

	if strings.HasPrefix(s, "-") {
		negative = true
		s = strings.TrimPrefix(s, "-")
	}

	parts := strings.Split(s, ".")

	intPart := parts[0]

	// ใส่ comma ทุก 3 หลัก
	for i := len(intPart) - 3; i > 0; i -= 3 {
		intPart = intPart[:i] + "," + intPart[i:]
	}

	result := intPart

	if decimals > 0 && len(parts) > 1 {
		result += "." + parts[1]
	}

	if negative {
		result = "-" + result
	}

	return result
}

func formatDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}

	return t.Format("02/01/2006")
}
