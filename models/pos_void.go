package models

import (
	"time"

	"github.com/google/uuid"
)

type PosVoidDto struct {
	BillNo     string `json:"bill_no"`
	VoidUser   string `json:"void_user"`
	VoidReason string `json:"void_reason"`
	// ForceVoid = ผู้จัดการอนุมัติทับกรณียอดในบัตรไม่พอให้คืน
	//
	// ปกติถ้าลูกค้าเติมเงินแล้วเล่นไปหมด การยกเลิกบิลจะถูกปฏิเสธด้วย 409
	// เพราะคืนยอดไม่ครบ ธงนี้บอกว่ามีผู้มีสิทธิ์รับทราบและยืนยันให้ทำต่อ
	// โดยหักเท่าที่มีจริง ไม่ทำให้ยอดติดลบ และบันทึกส่วนต่างลง void_reason
	ForceVoid  bool   `json:"force_void"`
	CardNo     string `json:"card_no"`
	DeductCard bool   `json:"deduct_card"`
}

type PosVoid struct {
	VoidID uuid.UUID `gorm:"column:void_id;type:uuid;default:uuid_generate_v4();primaryKey" json:"void_id"`
	// varchar(100) ตามคอลัมน์จริง (ตรวจ 2026-09-24) — ไม่ใช่ 50 เหมือน pos_transaction.bill_no
	BillNo   string    `gorm:"column:bill_no;type:varchar(100)" json:"bill_no"`
	VoidDate time.Time `gorm:"column:void_date;default:now()" json:"void_date"`
	VoidUser string    `gorm:"column:void_user;type:varchar(100)" json:"void_user"`
	// คอลัมน์จริงในฐานข้อมูลเป็น varchar(500) ไม่ใช่ 100 — ไม่กระทบตอน insert
	// เพราะ GORM ใช้ type tag เฉพาะตอน migrate แต่ถ้าวันหนึ่งมีใครรัน AutoMigrate
	// tag ที่ผิดจะหดคอลัมน์และตัดข้อความที่ยาวกว่า 100 ทิ้ง ซึ่งเส้นทาง ForceVoid
	// เขียนส่วนต่างยอดลงช่องนี้ จึงยาวเกิน 100 ได้จริง
	VoidReason string `gorm:"column:void_reason;type:varchar(500)" json:"void_reason"`
}

func (PosVoid) TableName() string {
	return "pos_void"
}
