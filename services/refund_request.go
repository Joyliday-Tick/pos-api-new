package services

import (
	"fmt"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"
)

func CreateRefundRequest(input models.RefundRequestDto) (models.RefundRequest, error) {
	var entity models.RefundRequest

	entity.RefundDate = *utils.TimeNowAsia()
	entity.RefundTel = input.RefundTel
	entity.RefundAmount = input.RefundAmount
	entity.RefundMeter = input.RefundMeter
	entity.RefundEcoin = input.RefundEcoin
	entity.RefundEbonus = input.RefundEbonus
	entity.CardNo = input.CardNo

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create refund request: %w", err)
	}

	return entity, nil
}

func RefundRequestByTel(tel string) ([]models.RefundRequestData, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
	SELECT 
		rr.id,
		rr.refund_date,
		rr.refund_tel,
		rr.refund_amount,
		rr.refund_ebonus, 
		rr.refund_ecoin, 
		rr.card_no,
		rrc.tel_count AS refund_cash_count,
		m.m_name as firstname,
		m.s_name as lastname
	FROM refund_request rr
	LEFT JOIN (
		SELECT 
			COUNT(id) AS tel_count, 
			member_tel  
		FROM refund_confirm 
		WHERE member_tel = ?
		GROUP BY member_tel
	) rrc ON rrc.member_tel = rr.refund_tel 
	LEFT JOIN member m on rr.refund_tel = m.tel 
	WHERE rr.refund_tel = ?
	  AND rr.refund_sts = 'Y';
	`

	var result []models.RefundRequestData
	tx := config.DB_POS.Raw(query, tel, tel).Scan(&result) // ส่งค่า tel 2 ตัว

	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil // ไม่พบข้อมูล
	}

	fmt.Printf("%+v\n", result)
	return result, nil
}

// FindRefundRequestStatusByID คืนค่า refund_sts ของคำขอคืนเงิน
//
// refund_sts = 'Y' คือยังรอคืนเงิน (หน้าจอดึงเฉพาะสถานะนี้มาแสดง)
// refund_sts = 'N' คือยืนยันไปแล้ว
//
// คืน "" เมื่อไม่พบคำขอ ผู้เรียกต้องเช็คก่อนทำรายการ ไม่งั้นยืนยันซ้ำได้
// แล้วหักยอดในบัตรซ้ำ เพราะไม่มีที่ไหนกันไว้เลย
func FindRefundRequestStatusByID(reqId string) (string, error) {
	if config.DB_POS == nil {
		return "", fmt.Errorf("database connection is nil")
	}

	var status string
	result := config.DB_POS.
		Raw("SELECT COALESCE(refund_sts, '') FROM refund_request WHERE id = ?", reqId).
		Scan(&status)

	if result.Error != nil {
		return "", fmt.Errorf("failed to read refund request: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return "", nil
	}

	return status, nil
}

func UpdateRefundRequestById(reqId string) error {
	if config.DB_POS == nil {
		return fmt.Errorf("database connection is nil")
	}

	// ขั้นแรก: ตรวจสอบว่ามี record นี้จริงไหม
	var exists bool
	check := config.DB_POS.
		Raw("SELECT EXISTS(SELECT 1 FROM refund_request WHERE id = ?) AS found", reqId).
		Scan(&exists)

	if check.Error != nil {
		return fmt.Errorf("failed to check refund request: %w", check.Error)
	}

	if !exists {
		return fmt.Errorf("refund request not found with id: %s", reqId)
	}

	// ขั้นสอง: update ถ้ามี record จริง
	result := config.DB_POS.Exec(`
		UPDATE refund_request
		SET refund_sts = 'N'
		WHERE id = ?
	`, reqId)

	if result.Error != nil {
		return fmt.Errorf("failed to update refund request: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows updated for id: %s", reqId)
	}

	return nil
}
