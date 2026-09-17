package services

import (
	"errors"
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

// ErrInsufficientDepositBalance คืนเมื่อ UPDATE ไม่โดนแถวไหนเลย แปลว่ายอดคงเหลือ
// ไม่พอให้หัก หรือแถวนั้นหายไป/ถูกปิดไปแล้ว ผู้เรียกต้องถือว่า "ยังไม่ได้หัก"
// และห้ามถือว่าสำเร็จ
var ErrInsufficientDepositBalance = errors.New("ยอดคงเหลือในกระเป๋าไม่พอสำหรับการหักครั้งนี้")

// ErrNegativeDeduction กันการส่งจำนวนติดลบเข้ามาหัก ซึ่ง SQL จะกลายเป็นการบวกเพิ่ม
// เท่ากับเสกยอดขึ้นมาจากอากาศ
var ErrNegativeDeduction = errors.New("จำนวนที่ขอหักติดลบ")

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
	if input.ID == nil {
		return input, fmt.Errorf("card deposit id is nil")
	}

	// BalanceDiscountCash เป็น *float32 และ 7 ใน 8 จุดที่สร้าง CardDepositBalanceDto
	// ไม่ได้ตั้งค่านี้ (void, claim-prize, refund) ค่าที่ bind จึงเป็น NULL
	// คอลัมน์ balance_discount_cash เป็น nullable ฐานข้อมูลจึงไม่ error
	// แต่ balance_discount_cash - NULL = NULL ยอดส่วนลดหายไปเงียบ ๆ และเมื่อเป็น NULL
	// แล้วทุกการคำนวณต่อจากนั้นก็เป็น NULL ตลอด ส่วน SUM() ในรายงานก็ข้ามแถวนั้นไป
	// ตรวจ UAT เมื่อ 2026-09-17 พบเกิดไปแล้ว 9 แถวจาก 2678
	// nil ต้องแปลว่า "ไม่มีส่วนลดให้หัก" = 0 ไม่ใช่ NULL
	var discountCash float32
	if input.BalanceDiscountCash != nil {
		discountCash = *input.BalanceDiscountCash
	}

	// จำนวนติดลบจะกลายเป็นการบวกเพิ่ม (x - (-5) = x + 5) และ guard ด้านล่างก็ผ่านเสมอ
	// เพราะ balance >= จำนวนติดลบ เป็นจริงตลอด ต้องกันตั้งแต่ต้นทาง
	if input.BalanceCoin < 0 || input.BalanceBonus < 0 || discountCash < 0 {
		return input, fmt.Errorf("%w: coin=%d bonus=%d discount_cash=%v",
			ErrNegativeDeduction, input.BalanceCoin, input.BalanceBonus, discountCash)
	}

	updateDate := utils.TimeNowAsia()

	// เงื่อนไข "ยอดพอไหม" ต้องอยู่ใน WHERE ไม่ใช่เช็คใน Go แล้วค่อยเขียน
	// เดิมอ่านยอดที่ card_play_service.go:778-797 แล้วมา UPDATE ที่นี่ คนละ connection
	// ไม่มี transaction ไม่มี FOR UPDATE — กดสองครั้งห่างกัน 50ms ทั้งคู่อ่านเห็นยอดพอ
	// แล้วต่างคนต่างหัก ยอดติดลบ ลูกค้าได้เล่นสองรอบจากยอดรอบเดียว
	// ตรวจ UAT เมื่อ 2026-09-17 พบ balance_coin ติดลบ 5 แถว balance_bonus ติดลบ 1 แถว
	// ย้ายเงื่อนไขเข้า WHERE แล้วเช็ค RowsAffected ทำให้ atomic ในคำสั่งเดียว
	// โดยไม่ต้องล็อกแถว และแก้ทั้ง double-spend กับ void ซ้ำพร้อมกัน
	//
	// COALESCE เพราะทั้งสามคอลัมน์เป็น nullable ถ้าเจอแถวที่เป็น NULL อยู่ก่อนแล้ว
	// NULL >= ? จะได้ NULL ซึ่งไม่ใช่ true — แถวนั้นจะหักไม่ได้ตลอดกาล
	result := config.DB_POS.Exec(`
		UPDATE card_deposit
		SET balance_coin          = COALESCE(balance_coin, 0) - ?,
		    balance_bonus         = COALESCE(balance_bonus, 0) - ?,
		    balance_discount_cash = COALESCE(balance_discount_cash, 0) - ?,
		    update_date           = ?
		WHERE id = ?
		  AND COALESCE(balance_coin, 0)          >= ?
		  AND COALESCE(balance_bonus, 0)         >= ?
		  AND COALESCE(balance_discount_cash, 0) >= ?`,
		input.BalanceCoin, input.BalanceBonus, discountCash, updateDate, input.ID,
		input.BalanceCoin, input.BalanceBonus, discountCash)
	if result.Error != nil {
		return input, fmt.Errorf("failed to update card deposit balance: %w", result.Error)
	}

	// RowsAffected == 0 แปลว่ายอดไม่พอ หรือไม่มีแถวนั้น — ไม่ได้หักอะไรเลย
	// ต้องคืน error ไม่ใช่ปล่อยผ่านเหมือนเดิม ไม่งั้นเครื่องจะปลดล็อกให้เล่นฟรี
	if result.RowsAffected == 0 {
		return input, fmt.Errorf("%w (card_deposit id=%s ขอหัก coin=%d bonus=%d discount_cash=%v)",
			ErrInsufficientDepositBalance, input.ID, input.BalanceCoin, input.BalanceBonus, discountCash)
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