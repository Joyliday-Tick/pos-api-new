package services

import (
	"fmt"
	"strings"

	// "strings"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
	"new-pos-api/utils"
	// "new-pos-api/utils"
	// "gorm.io/gorm"
)

func CreateCardDeposit(input models.CardDepositDto, userId int) (models.CardDeposit, error) {
	var deposit models.CardDeposit
	fmt.Println("Creating card deposit with input:", input)

	deposit.FromChannel = input.FromChannel
	deposit.CardNo = input.CardNo
	deposit.MemberTel = input.MemberTel
	deposit.Amount = input.Amount
	deposit.Coin = input.Coin
	deposit.Bonus = input.Bonus
	deposit.PosId = input.PosId
	deposit.PosMenuId = input.PosMenuId
	deposit.BonusExpireDate = input.BonusExpireDate
	deposit.BillNo = input.BillNo
	deposit.CreateDate = *utils.TimeNowAsia()
	deposit.BonusExpireDate = input.BonusExpireDate
	deposit.CardExpireDate = input.CardExpireDate
	deposit.DiscountCash = input.DiscountCash
	deposit.BalanceDiscountCash = input.BalanceDiscountCash
	deposit.DiscountCashExpire = input.DiscountCashExpire

	if input.BalanceCoin > 0 {
		deposit.BalanceCoin = input.BalanceCoin
	}
	if input.BalanceBonus > 0 {
		deposit.BalanceBonus = input.BalanceBonus
	}

	if config.DB_POS == nil {
		return deposit, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&deposit).Error; err != nil {
		return deposit, fmt.Errorf("failed to create card deposit: %w", err)
	}

	return deposit, nil
}

// func UpdateCardDeposit(id string, input models.CardRegisterDto, userId int) (models.CardEntity, error) {
// 	var existing models.CardEntity

// 	if config.DB_POS == nil {
// 		return existing, fmt.Errorf("database pos connection is nil")
// 	}

// 	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
// 		return existing, fmt.Errorf("card entity not found: %w", err)
// 	}

// 	// เตรียมข้อมูลที่ต้องการอัปเดต
// 	updateData := map[string]interface{}{
// 		"is_active":   input.IsActive,
// 		"update_by":   userId,
// 		"update_date": utils.TimeNowAsia(),
// 	}

// 	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
// 	if err := config.DB_POS.Model(&models.CardEntity{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
// 		return existing, fmt.Errorf("failed to update play type: %w", err)
// 	}

// 	// ดึงข้อมูลใหม่หลังอัปเดต
// 	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
// 		return existing, fmt.Errorf("failed to retrieve updated play type: %w", err)
// 	}

// 	return existing, nil
// }

// func FindCardDepositByMemberTel(tel string) ([]models.CardEntity, error) {
// 	if config.DB_POS == nil {
// 		return nil, fmt.Errorf("database connection is nil")
// 	}

// 	var entities []models.CardEntity
// 	lowerTel := strings.ToLower(tel)

// 	err := config.DB_POS.
// 		Where("LOWER(member_tel) = ?", lowerTel).
// 		Find(&entities).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	// ไม่ต้อง return error ถ้าไม่เจอ — แค่คืน slice ว่าง
// 	return entities, nil
// }

func FindCardDepositByCardNo(cardNo string) ([]models.CardDeposit, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var cardDeposit []models.CardDeposit
	// lowerTel := strings.ToLower(memberTel)
	lowerCard := strings.ToLower(cardNo)

	result := config.DB_POS.
		Where("LOWER(card_no) = ? and is_active = true  and (balance_bonus > 0 or balance_coin > 0)", lowerCard).
		Order("create_date ASC").
		Find(&cardDeposit)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return cardDeposit, nil
}

func FindCardDepositByIds(ids []string) ([]models.CardDeposit, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var cardDeposits []models.CardDeposit

	result := config.DB_POS.
		Where("id IN ?", ids).
		Order("create_date ASC").
		Find(&cardDeposits)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return cardDeposits, nil
}

// update balance
func UpdateCardDepositBalance(input models.CardDepositBalanceDto) (models.CardDepositBalanceDto, error) {
	if config.DB_POS == nil {
		return input, fmt.Errorf("database pos connection is nil")
	}
	updateDate := utils.TimeNowAsia()
	result := config.DB_POS.Exec("UPDATE card_deposit SET balance_coin = balance_coin  - ?, balance_bonus = balance_bonus - ?,	 balance_discount_cash = balance_discount_cash - ?, update_date = ? WHERE id = ?", input.BalanceCoin, input.BalanceBonus, &input.BalanceDiscountCash, updateDate, &input.ID)
	if result.Error != nil {
		return input, fmt.Errorf("failed to update card deposit balance: %w", result.Error)
	}
	// ปิด card play เมื่อยอดเหลือ 0
	// เดิมเงื่อนไขกลับด้าน (if err != nil) บล็อกนี้จึงทำงานเฉพาะตอน query ล้มเหลว
	// ซึ่งตอนนั้น deposit เป็น nil เสมอ -> deposit[0] panic
	// และตอน query สำเร็จก็ข้ามไปเลย ทำให้ card play ที่ยอดเหลือ 0 ไม่เคยถูกปิด
	//
	// ขั้นตอนนี้เป็นงานพ่วง ไม่ใช่ส่วนของการหักยอด (ยอดถูกหักสำเร็จไปแล้วด้านบน)
	// ถ้าล้มเหลวจึงแค่ log ไม่ return error ไม่งั้นผู้เรียกจะ retry แล้วหักซ้ำ
	var depositID []string
	depositID = append(depositID, input.ID.String())
	deposit, err := FindCardDepositByIds(depositID)
	if err != nil {
		fmt.Printf("card deposit %s: อ่านยอดคงเหลือหลังหักไม่สำเร็จ: %v\n", input.ID, err)
	} else if len(deposit) > 0 && deposit[0].BalanceCoin == 0 {
		if err := UpdateCardPlayIsDelete(input.ID); err != nil {
			fmt.Printf("card deposit %s: ปิด card play ไม่สำเร็จ: %v\n", input.ID, err)
		}
	}
	return input, nil
}

// get bonus expire total , date

func GetCardDepositBonusExpireTotalDate(cardNo string) (*models.CardDepositBonusExpireDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database pos connection is nil")
	}
	var bonusExpire models.CardDepositBonusExpireDto
	result := config.DB_POS.Raw("select sum(cd.balance_bonus) as bonus_expire_total,min(cd.bonus_expire_date::date) as bonus_expire_date from card_deposit cd where cd.card_no = ? and  cd.bonus_expire_date::date = (select min(cd2.bonus_expire_date::date) from card_deposit cd2 where cd2.card_no = ? and cd2.is_active = true and cd2.bonus_expire_date is not null) and cd.is_active = true;", cardNo, cardNo).Scan(&bonusExpire)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get card deposit bonus expire total and date: %w", result.Error)
	}
	fmt.Println("bonus expire total:", bonusExpire)
	return &bonusExpire, nil
}

// get deposit by member tel
func GetDepositByMemberTel(memberTel string) (*[]models.CardDeposit, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database pos connection is nil")
	}
	var deposit []models.CardDeposit
	config.DB_POS.Find(&deposit, "member_tel = ? and is_active = true ", memberTel)
	return &deposit, nil
}