package services

import (
	"fmt"
	"strconv"

	// "strings"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"

	"os"
	"time"
	// "new-pos-api/utils"
	// "gorm.io/gorm"
)

func CreateDepositTransaction(input models.DepositJubuJibi) (models.DepositHead, error) {
	var transaction models.DepositHead

	// Validate: ต้องมี SubTransactions อย่างน้อย 1 รายการ
	if len(input.SubItems) == 0 {
		return transaction, fmt.Errorf("sub deposit transaction is required")
	}

	// Generate Bill No
	BillNo := GenerateDepositBillNo(input.PosID)
	// Map input DTO ไปยัง model
	transaction.DhBillNo = BillNo
	transaction.DhPosID = input.PosID
	transaction.DhLocation = input.Location
	transaction.DhUser = input.Cashier
	transaction.DhMemberTel = input.MemberTel
	transaction.DhSpending = input.Spending
	transaction.DhDeposit = input.Deposit
	transaction.DhDiscount = input.Discount
	transaction.DhJoylicoin = input.Joylicoin
	transaction.DhDepositJoylicoin = input.DepositJoylicoin
	transaction.DhType = input.Type
	transaction.DhDate = *utils.TimeNowAsia()

	// ตรวจสอบ DB connection
	if config.DB_POS == nil {
		return transaction, fmt.Errorf("database POS connection is nil")
	}

	// สร้าง transaction หลัก
	if err := config.DB_POS.Create(&transaction).Error; err != nil {
		return transaction, fmt.Errorf("failed to create deposit head transaction: %w", err)
	}

	// เตรียมและบันทึก SubTransactions
	var subs []models.DepositSub
	for _, sub := range input.SubItems {

		subTransaction := models.DepositSub{
			DsBillNo:    transaction.DhBillNo,
			DsCode:      sub.Sku,
			DsName:      sub.ProductName,
			DsQty:       sub.Qty,
			DsPrice:     sub.Price,
			DsJoylicoin: sub.Joylicoin,
		}
		subs = append(subs, subTransaction)
	}

	// เรียกฟังก์ชันบันทึก batch
	if _, err := CreateBatchSubDeposit(subs); err != nil {
		return transaction, fmt.Errorf("failed to create sub deposit: %w", err)
	}

	return transaction, nil
}

func CreateBatchSubDeposit(subs []models.DepositSub) ([]models.DepositSub, error) {

	if config.DB_POS == nil {
		return subs, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&subs).Error; err != nil {
		return subs, fmt.Errorf("failed to create batch sub deposit: %w", err)

	}

	return subs, nil
}

// GenerateBillNo generate bill no
func GenerateDepositBillNo(PosID string) string {
	var BillNo string
	var totalRecord int64
	currentDate := time.Now()
	YYYY := currentDate.Year()
	YY := YYYY % 100
	MMM := currentDate.Month()
	MM := int(MMM)
	MMStr := fmt.Sprintf("%02d", MM)

	config.DB_POS.Model(&models.DepositHead{}).Where("Extract(year from dh_date) = ? AND Extract(month from dh_date) = ?", YYYY, MM).Count(&totalRecord)
	totalRecord += 1
	padTotal := fmt.Sprintf("%04d", totalRecord)
	fmt.Println("totalRecord", totalRecord)

	prefix := os.Getenv("DEPOSIT_PREFIX_NORMAL")
	if prefix == "" {
		prefix = "FWD" // default เผื่อ env ไม่มีค่า
	}

	BillNo = prefix + "-" + PosID + "-" + strconv.Itoa(YY) + MMStr + padTotal
	return BillNo
}

func GetDepositHeadByTel(tel string) ([]models.DepositHeadCron, error) {
	var results []models.DepositHeadCron

	err := config.DB_POS.
		Where("dh_member_tel = ?", tel).
		Order("dh_date DESC").
		Limit(20).
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}
