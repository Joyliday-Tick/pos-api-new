package services

import (
	"fmt"
	"strings"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
)

func CreateBatchSubTransaction(subs []models.PosSubTransaction) ([]models.PosSubTransaction, error) {

	if config.DB_POS == nil {
		return subs, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&subs).Error; err != nil {
		return subs, fmt.Errorf("failed to create batch sub transaction: %w", err)

	}

	return subs, nil
}

func FindSubTransactionByBillNo(billNo string) ([]models.PosSubTransactionData, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var subTransactions []models.PosSubTransactionData

	err := config.DB_POS.Table("pos_sub_transaction AS pst").
		Select(`pm.description AS menu_name,
				pst.product_id ,
				pst.qty,
				pst.price,
				pst.e_bonus,
				pst.e_coin,
				pst.bill_no,
				pst.card_deposit_id`).
		Joins("LEFT JOIN pos_menu AS pm ON pst.product_id = pm.id").
		Where("LOWER(pst.bill_no) = ?", strings.ToLower(billNo)).
		Scan(&subTransactions).Error

	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	if len(subTransactions) == 0 {
		return nil, nil
	}

	fmt.Println(len(subTransactions), "rows found for bill no:", billNo)

	return subTransactions, nil
}
