package services

import (
	"errors"
	"fmt"
	"strings"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"

	"gorm.io/gorm"
)

func CreatePosVoid(void models.PosVoidDto) (models.PosVoid, error) {

	var posVoid models.PosVoid

	posVoid.BillNo = void.BillNo
	posVoid.VoidUser = void.VoidUser
	posVoid.VoidReason = void.VoidReason
	posVoid.VoidDate = *utils.TimeNowAsia()

	if config.DB_POS == nil {
		return posVoid, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&posVoid).Error; err != nil {
		return posVoid, fmt.Errorf("failed to create batch sub transaction: %w", err)

	}

	return posVoid, nil
}

func FindExistVoidByBillNo(billNo string) (*models.PosVoid, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var posVoid models.PosVoid
	lowerbillNo := strings.ToLower(billNo)

	err := config.DB_POS.
		Where("LOWER(bill_no) = ?", lowerbillNo).
		First(&posVoid).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// ไม่พบข้อมูล ไม่ถือว่า error
		return nil, nil
	} else if err != nil {
		// error อื่นๆ เช่น connection error
		return nil, err
	}

	return &posVoid, nil
}

func PosVoidSearchList(params models.SearchTaxInvoiceReport) (utils.SearchResult, error) {
	var result utils.SearchResult
	var transactions []models.PosTransactionVoidData

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	query := config.DB_POS.Table("pos_transaction pt").
		Select(`pt.id ,
				pt.bill_no,
				pt.bill_date,
				pt.bill_location,
				pt.cashier,
				pt.Member_tel,
				coalesce(pt.product_price,0) as product_price,
				pv.void_date,
				pv.void_reason,
				pv.void_user`).
		Joins("LEFT JOIN pos_void pv on pt.bill_no = pv.bill_no").
		Where("pt.bill_status = 'Void' and pt.is_delete = false")

	if params.Location != "" {
		query = query.Where("(pt.bill_location = ?)", params.Location)
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
	if err := query.Order("pt.bill_date ASC").Limit(params.Skip).Offset(offset).Find(&transactions).Error; err != nil {
		return result, fmt.Errorf("data query failed: %w", err)
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     transactions,
	}
	return result, nil
}
