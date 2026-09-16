package services

import (
	"fmt"

	// "strings"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"

	"os"
	"time"
	// "new-pos-api/utils"
	// "gorm.io/gorm"
)

func CreateRedeemTransaction(input models.RedeemJubuJibi) (models.RedeemHead, error) {
	var transaction models.RedeemHead

	// Validate: ต้องมี SubTransactions อย่างน้อย 1 รายการ
	if len(input.SubItems) == 0 {
		return transaction, fmt.Errorf("sub redeem transaction is required")
	}

	// Generate Bill No
	BillNo := GenerateRedeemBillNo(input.PosID)
	// Map input DTO ไปยัง model
	transaction.RhBillNo = BillNo
	transaction.RhPosID = input.PosID
	transaction.RhLocation = input.Location
	transaction.RhUser = input.Cashier
	transaction.RhMemberTel = input.MemberTel
	transaction.RhRedeemPrice = input.RedeemPrice
	transaction.RhType = input.Type
	transaction.RhDate = *utils.TimeNowAsia()

	// ตรวจสอบ DB connection
	if config.DB_POS == nil {
		return transaction, fmt.Errorf("database POS connection is nil")
	}

	// สร้าง transaction หลัก
	if err := config.DB_POS.Create(&transaction).Error; err != nil {
		return transaction, fmt.Errorf("failed to create redeem head transaction: %w", err)
	}

	// เตรียมและบันทึก SubTransactions
	var subs []models.RedeemSub
	for _, sub := range input.SubItems {

		subTransaction := models.RedeemSub{
			RsBillNo:    transaction.RhBillNo,
			RsCode:      sub.Sku,
			RsName:      sub.ProductName,
			RsQty:       sub.Qty,
			RsPrice:     sub.Price,
			RsJoylicoin: sub.Joylicoin,
		}
		subs = append(subs, subTransaction)
	}

	// เรียกฟังก์ชันบันทึก batch
	if _, err := CreateBatchSubRedeem(subs); err != nil {
		return transaction, fmt.Errorf("failed to create sub deposit: %w", err)
	}

	return transaction, nil
}

func CreateBatchSubRedeem(subs []models.RedeemSub) ([]models.RedeemSub, error) {

	if config.DB_POS == nil {
		return subs, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&subs).Error; err != nil {
		return subs, fmt.Errorf("failed to create batch sub redeem: %w", err)

	}

	return subs, nil
}

func GenerateRedeemBillNo(PosID string) string {
	var totalRecord int64
	currentDate := time.Now()
	YYYY := currentDate.Year()
	YY := YYYY % 100
	MM := int(currentDate.Month())

	// นับจำนวนเรคคอร์ดของเดือนนี้
	config.DB_POS.Model(&models.RedeemHead{}).
		Where("EXTRACT(YEAR FROM rh_date) = ? AND EXTRACT(MONTH FROM rh_date) = ?", YYYY, MM).
		Count(&totalRecord)

	totalRecord += 1
	padTotal := fmt.Sprintf("%04d", totalRecord)

	prefix := os.Getenv("REDEEM_PREFIX_NORMAL")
	if prefix == "" {
		prefix = "FWR" // default เผื่อ env ไม่มีค่า
	}

	BillNo := fmt.Sprintf("%s-%s-%02d%02d%s", prefix, PosID, YY, MM, padTotal)
	fmt.Println("totalRecord:", totalRecord)

	return BillNo
}
