package models

import (
	"time"

	"github.com/google/uuid"
)

type PosTransactionDto struct {
	BillNo          string                 `json:"bill_no"`
	BillDate        time.Time              `json:"bill_date"`
	PosID           string                 `json:"pos_id"`
	BillLocation    string                 `json:"bill_location"`
	Cashier         string                 `json:"cashier"`
	CardNo          string                 `json:"card_no"`
	MemberTel       string                 `json:"member_tel"`
	ProductPrice    int                    `json:"product_price"`
	ECoin           int                    `json:"e_coin"`
	FreePoint       int                    `json:"free_point"`
	EBonus          int                    `json:"e_bonus"`
	BillPaymentId   uuid.UUID              `json:"bill_payment_id"`
	PosType         string                 `json:"pos_type"`
	BillStatus      string                 `json:"bill_status"`
	BonusStatus     string                 `json:"bonus_status"`
	BankDetail      string                 `json:"bank_detail"`
	SubTransactions []PosSubTransactionDto `json:"sub_transactions"`
}

type PosTransactionData struct {
	ID           uuid.UUID `json:"id"`
	BillNo       string    `json:"bill_no"`
	BillDate     time.Time `json:"bill_date"`
	PosID        string    `json:"pos_id"`
	PosCode      string    `json:"pos_code"`
	BillLocation string    `json:"bill_location"`
	Cashier      string    `json:"cashier"`
	CardNo       string    `json:"card_no"`
	MemberTel    string    `json:"member_tel"`
	ProductPrice int       `json:"price"`
	ECoin        int       `json:"e_coin"`
	EBonus       int       `json:"e_bonus"`
	BillStatus   string    `json:"bill_status"`
	BranchName   string    `json:"branch_name"`
	BillPayment  string    `json:"bill_payment"`
}

type PosTransactionVoidData struct {
	ID           uuid.UUID `json:"id"`
	BillNo       string    `json:"bill_no"`
	BillDate     time.Time `json:"bill_date"`
	BillLocation string    `json:"bill_location"`
	MemberTel    string    `json:"member_tel"`
	Cashier      string    `json:"cashier"`
	ProductPrice int       `json:"price"`
	VoidDate     time.Time `json:"void_date"`
	VoidReason   string    `json:"void_reason"`
	VoidUser     string    `json:"void_user"`
}

type TaxInvoiceSummaryReport struct {
	BillDate     time.Time `json:"bill_date"`     // วันที่ออกบิล (เฉพาะวัน)
	BillLocation string    `json:"bill_location"` // สาขาหรือจุดขาย
	MinBillNo    string    `json:"min_bill_no"`   // หมายเลขบิลเริ่มต้นของวันนั้น
	MaxBillNo    string    `json:"max_bill_no"`   // หมายเลขบิลสุดท้ายของวันนั้น
	Total        float64   `json:"total"`         // ยอดรวมราคาสินค้าทั้งหมดของวันนั้น
	Tax          float64   `json:"tax"`
	BeforeVat    float64   `json:"before_vat"` // ยอดรวมก่อน VAT

}

type TaxInvoiceResult struct {
	Data           []TaxInvoiceSummaryReport `json:"data"`
	TotalPrice     int64                     `json:"total_price"`
	TotalTax       float64                   `json:"total_tax"`
	TotalBeforeVat float64                   `json:"total_before_vat"`
}

type SpendingReport struct {
	BillNo          string    `json:"bill_no"`
	BillDate        time.Time `json:"bill_date"`
	PosID           string    `json:"pos_id"`
	PosCode         string    `json:"pos_code"`
	BillLocation    string    `json:"bill_location"`
	Cashier         string    `json:"cashier"`
	CardNo          string    `json:"card_no"`
	MemberTel       string    `json:"member_tel"`
	ProductPrice    int       `json:"price"`
	ECoin           int       `json:"e_coin"`
	EBonus          int       `json:"e_bonus"`
	PaymentTypeName string    `json:"payment_type_name"`
	BankDetail      string    `json:"bank_detail"`
	BranchName      string    `json:"branch_name"`
}

type ExportSpendingReport struct {
	BillNo          string    `json:"bill_no"`
	BillDate        time.Time `json:"bill_date"`
	PosID           string    `json:"pos_id"`
	PosCode         string    `json:"pos_code"`
	BillLocation    string    `json:"bill_location"`
	Cashier         string    `json:"cashier"`
	CardNo          string    `json:"card_no"`
	MemberTel       string    `json:"member_tel"`
	ProductPrice    int       `json:"price"`
	ECoin           int       `json:"e_coin"`
	EBonus          int       `json:"e_bonus"`
	PaymentTypeName string    `json:"payment_type_name"`
	BankDetail      string    `json:"bank_detail"`
	BranchName      string    `json:"branch_name"`
	MemberID        *int      `json:"member_id"`
}

type SpendingReportData struct {
	Data       []SpendingReport `json:"data"`
	TotalPrice int64            `json:"total_price"`
}

type PosReport struct {
	ProductList     []PosProductList    `json:"product_list"`
	BillNormalCount int                 `json:"bill_normal_count"`
	BillVoidCount   int                 `json:"bill_void_count"`
	Payments        []PaymentSummary    `json:"payments"`
	BankDetails     []BankDetailSummary `json:"bank_details"`
	Cashiers        []CashierSummary    `json:"cashiers"`
}

type PosProductList struct {
	ProductId    int     `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Qty          int     `json:"qty"`
	Price        int     `json:"price"`
	ECoin        float64 `json:"e_coin"`
	ProductPrice int     `json:"product_price"`
	Total        int     `json:"total"`
	EBonus       int     `json:"e_bonus"`
	Date         string  `json:"date"`
	Location     string  `json:"location"`
}

type PaymentSummary struct {
	PaymentName string  `json:"payment_name"`
	TotalAmount float64 `json:"total_amount"`
}

type BankDetailSummary struct {
	BankName    string  `json:"payment_name"`
	TotalAmount float64 `json:"total_amount"`
}

type CashierSummary struct {
	Cashier     string           `json:"cashier"`
	TotalAmount float64          `json:"total_amount"`
	Payments    []PaymentSummary `json:"payments"`
}
type PosTransactionReport struct {
	ID              uuid.UUID  `json:"id"`
	BillNo          string     `json:"bill_no"`
	BillDate        time.Time  `json:"bill_date"`
	PosID           string     `json:"pos_id"`
	BillLocation    string     `json:"bill_location"`
	Cashier         string     `json:"cashier"`
	CardNo          string     `json:"card_no"`
	MemberTel       string     `json:"member_tel"`
	ProductPrice    int        `json:"product_price"`
	ECoin           int        `json:"e_coin"`
	FreePoint       int        `json:"free_point"`
	EBonus          int        `json:"e_bonus"`
	BillPaymentId   uuid.UUID  `json:"bill_payment_id"`
	BillPaymentName string     `json:"bill_payment_name"`
	PosType         string     `json:"pos_type"`
	BillStatus      string     `json:"bill_status"`
	BonusStatus     string     `json:"bonus_status"`
	BankDetail      string     `json:"bank_detail"`
	CreateBy        int        `json:"create_by"`
	CreateDate      time.Time  `json:"create_date"`
	UpdateBy        *int       `json:"update_by"`
	UpdateDate      *time.Time `json:"update_date"`
	IsDelete        bool       `json:"is_delete"`
	DeleteBy        *int       `json:"delete_by"`
	DeleteDate      *time.Time `json:"delete_date"`
}

type PosTransaction struct {
	ID            uuid.UUID  `gorm:"column:id;type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	BillNo        string     `gorm:"column:bill_no;type:varchar(50)" json:"bill_no"`
	BillDate      time.Time  `gorm:"column:bill_date;default:now()" json:"bill_date"`
	PosID         string     `gorm:"column:pos_id" json:"pos_id"`
	BillLocation  string     `gorm:"column:bill_location;type:varchar(50)" json:"bill_location"`
	Cashier       string     `gorm:"column:cashier;type:varchar(100)" json:"cashier"`
	CardNo        string     `gorm:"column:card_no;type:varchar(50)" json:"card_no"`
	MemberTel     string     `gorm:"column:member_tel;type:varchar(50)" json:"member_tel"`
	ProductPrice  int        `gorm:"column:product_price;default:0" json:"product_price"`
	ECoin         int        `gorm:"column:e_coin;default:0" json:"e_coin"`
	FreePoint     int        `gorm:"column:free_point;default:0" json:"free_point"`
	EBonus        int        `gorm:"column:e_bonus;default:0" json:"e_bonus"`
	BillPaymentId uuid.UUID  `gorm:"column:bill_payment_id;type:uuid" json:"bill_payment_id"`
	PosType       string     `gorm:"column:pos_type;type:varchar(50)" json:"pos_type"`
	BillStatus    string     `gorm:"column:bill_status;type:varchar(20)" json:"bill_status"`
	BonusStatus   string     `gorm:"column:bonus_status;type:varchar(10)" json:"bonus_status"`
	BankDetail    string     `gorm:"column:bank_detail;type:varchar(50)" json:"bank_detail"`
	CreateBy      int        `gorm:"column:create_by" json:"create_by"`
	CreateDate    time.Time  `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateBy      *int       `gorm:"column:update_by" json:"update_by"`
	UpdateDate    *time.Time `gorm:"column:update_date" json:"update_date"`
	IsDelete      bool       `gorm:"column:is_delete;default:false" json:"is_delete"`
	DeleteBy      *int       `gorm:"column:delete_by" json:"delete_by"`
	DeleteDate    *time.Time `gorm:"column:delete_date" json:"delete_date"`
}

func (PosTransaction) TableName() string {
	return "pos_transaction"
}

type SearchPosTransactionParams struct {
	BillNo    string
	MemberTel string
	StartDate string
	EndDate   string
	CardNo    string
	POSID     string
	Page      int
	Skip      int
}

type SearchPosSaleReport struct {
	StartDate string
	EndDate   string
	Location  string
}

type SearchTaxInvoiceReport struct {
	StartDate string
	EndDate   string
	Location  string
	Page      int
	Skip      int
}

type SearchExportTaxInvoiceReport struct {
	StartDate    string
	EndDate      string
	Location     string
	ExportFormat string // "pdf" or "excel"
}

type SearchStampHouseReport struct {
	StartDate string
	EndDate   string
	Location  string
	Machine   string
	MemberTel string
	Page      int
	Skip      int
}

type SearchExportStampHouseReport struct {
	StartDate string
	EndDate   string
	Location  string
	Machine   string
	MemberTel string
}

type SearchSpendingParams struct {
	MemberTel string
	StartDate string
	EndDate   string
	Page      int
	Skip      int
}

type ExportSpendingParams struct {
	MemberTel string
	StartDate string
	EndDate   string
}
type GroupPosDailySalesSummaryReport struct {
	Date         string  `json:"date"`
	ProductPrice int     `json:"product_price"`
	ECoin        float64 `json:"e_coin"`
	EBonus       float64 `json:"e_bonus"`
	PosId        string  `json:"pos_id"`
	PosCode      string  `json:"pos_code"`
	BillPayment  string  `json:"bill_payment"`
}
type PosDailySalesSummaryReport struct {
	Data        []GroupPosDailySalesSummaryReport `json:"data"`
	TotalPrice  int64                             `json:"total_price"`
	TotalECoin  int64                             `json:"total_e_coin"`
	TotalEBonus int64                             `json:"total_e_bonus"`
}

type GroupPosMonthlySalesSummaryReport struct {
	Date         string  `json:"date"`
	ProductPrice int     `json:"product_price"`
	ECoin        float64 `json:"e_coin"`
	EBonus       int     `json:"e_bonus"`
	BillPayment  string  `json:"bill_payment"`
}
type PosMonthlySalesSummaryReport struct {
	Data        []GroupPosMonthlySalesSummaryReport `json:"data"`
	TotalPrice  int64                               `json:"total_price"`
	TotalECoin  int64                               `json:"total_e_coin"`
	TotalEBonus int64                               `json:"total_e_bonus"`
}

type SearchPosSalePaymentReport struct {
	StartDate     string
	EndDate       string
	Location      string
	PaymentTypeId string
	Page          int
	Skip          int
}

type ExportPosSalePaymentReport struct {
	StartDate     string
	EndDate       string
	Location      string
	PaymentTypeId string
}
