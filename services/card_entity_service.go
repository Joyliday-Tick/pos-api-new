package services

import (
	"fmt"
	"sync"
	"time"

	// "log"
	"strings"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"

	"new-pos-api/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateCardEntity(input models.CardRegisterDto, userId int) (models.CardEntity, error) {
	var entity models.CardEntity

	entity.CardNo = input.CardNo
	entity.CardTypeId = uuid.MustParse(input.CardTypeId)
	// entity.CardTypeId = input.CardTypeId
	entity.IsActive = input.IsActive
	entity.MemberTel = input.MemberTel
	entity.CreateBy = userId
	entity.CreateDate = *utils.TimeNowAsia()

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create card entity: %w", err)
	}

	return entity, nil
}

func UpdateCardEntity(id string, input models.CardRegisterDto, userId int) (models.CardEntity, error) {
	var existing models.CardEntity

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("card entity not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   input.IsActive,
		"update_by":   userId,
		"update_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.CardEntity{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update card entity: %w", err)
	}

	// ดึงข้อมูลใหม่หลังอัปเดต
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated card entity: %w", err)
	}

	return existing, nil
}

func FindCardActiveByMemberTel(tel string) ([]models.CardEntity, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var entities []models.CardEntity
	lowerTel := strings.ToLower(tel)

	err := config.DB_POS.
		Where("LOWER(member_tel) = ? and is_active = true and is_delete = false", lowerTel).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	// ไม่ต้อง return error ถ้าไม่เจอ — แค่คืน slice ว่าง
	return entities, nil
}

func FindActiveCardNo(cardNo string) (*models.CardEntity, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var cardEntity models.CardEntity
	// lowerTel := strings.ToLower(memberTel)
	lowerCard := strings.ToLower(cardNo)

	result := config.DB_POS.
		Where("LOWER(card_no) = ? and is_active = true and is_delete = false", lowerCard).
		Find(&cardEntity)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &cardEntity, nil
}

func FindActiveMemberTelAndCardNo(memberTel string, cardNo string) (*models.CardEntity, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var cardEntity models.CardEntity
	lowerTel := strings.ToLower(memberTel)
	lowerCard := strings.ToLower(cardNo)

	result := config.DB_POS.
		Where("LOWER(member_tel) = ? AND LOWER(card_no) = ? and is_active = true", lowerTel, lowerCard).
		Find(&cardEntity)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &cardEntity, nil
}

func FindLockCard(cardNo string) (*models.CardEntity, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var cardEntity models.CardEntity
	// lowerTel := strings.ToLower(memberTel)
	lowerCard := strings.ToLower(cardNo)

	result := config.DB_POS.
		Where("LOWER(card_no) = ? and is_active = false and is_delete = false", lowerCard).
		Find(&cardEntity)

	if result.Error != nil {
		return nil, result.Error
	}

	// ถ้าไม่พบข้อมูล
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &cardEntity, nil
}

func CheckCardInfo(cardNo string) (*models.CardSummaryDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
	SELECT 
    ce.card_no,
    ct.name AS card_type,
    ct.show_balance,
    ce.member_tel,
    COALESCE(cd_summary.e_coin, 0) AS e_coin,
    COALESCE(cd_summary.e_bonus, 0) AS e_bonus,
	COALESCE(cd_summary.discount_cash, 0) AS discount_cash,
	cust_tier.tier as customer_tier
FROM card_entity ce
LEFT JOIN card_type ct ON ce.card_type_id = ct.id
LEFT JOIN LATERAL (
    SELECT
        SUM(balance_coin) AS e_coin,
        SUM(
            CASE
                WHEN bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')
                THEN balance_bonus
                ELSE 0
            END
        ) AS e_bonus,
        SUM(
            CASE
                WHEN discount_cash_expire >= (NOW() AT TIME ZONE 'Asia/Bangkok')
                THEN balance_discount_cash
                ELSE 0
            END
        ) AS discount_cash
    FROM card_deposit cd
    WHERE cd.is_active = TRUE
AND (
    (
        (ce.member_tel = '0000000000'
         OR ce.member_tel = ''
         OR ce.member_tel IS NULL)
        AND cd.card_no = ce.card_no
    )
    OR
    (
        ce.member_tel <> '0000000000'
        AND ce.member_tel <> ''
        AND ce.member_tel IS NOT NULL
        AND cd.member_tel = ce.member_tel
        AND cd.card_no = ce.card_no
    )
)
) cd_summary ON TRUE
LEFT JOIN (
SELECT DISTINCT ON (tel) tel, tier
		FROM customer_tier
		ORDER BY tel, create_date DESC
) cust_tier on ce.member_tel = cust_tier.tel
WHERE LOWER(ce.card_no) = ?
  AND ce.is_active = TRUE
  AND ce.is_delete = FALSE
	`

	var result models.CardSummaryDto
	tx := config.DB_POS.Raw(query, strings.ToLower(cardNo)).Scan(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil // ไม่พบข้อมูล
	}

	return &result, nil
}

func FindCardEntityById(id string) (*models.CardEntity, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var entity models.CardEntity

	// ใช้ GORM เพื่อค้นหาด้วย Where แทนการใช้ Raw Query
	if err := config.DB_POS.Where("id = ?", id).First(&entity).Error; err != nil {
		// ถ้าค้นหาไม่เจอ หรือเกิดข้อผิดพลาด
		if err == gorm.ErrRecordNotFound {
			return nil, nil // ถ้าไม่พบข้อมูล
		}
		return nil, err // ถ้ามีข้อผิดพลาดอื่น
	}

	return &entity, nil
}

func FindMemberTelActive() ([]models.PlayType, error) {
	var playTypes []models.PlayType

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// ใช้ GORM แบบง่ายและปลอดภัย
	if err := config.DB_POS.Where("is_active = ?", true).Find(&playTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active play types: %w", err)
	}

	return playTypes, nil
}

func DeleteCardEntityById(id string, userId int) (models.PlayType, error) {
	var existing models.PlayType

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	// หา record เดิมก่อนเพื่อ validate หรือใช้คืนใน response
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("play type not found: %w", err)
	}
	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.PlayType{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update play type: %w", err)
	}

	// ดึงข้อมูลใหม่หลังอัปเดต
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated play type: %w", err)
	}

	return existing, nil
}

func DeleteCardEntityByCardNo(cardNo string, userId int) ([]models.CardEntity, error) {
	var existing []models.CardEntity

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	lowerName := strings.ToLower(cardNo)

	// ดึงข้อมูล card ตาม card_no
	if err := config.DB_POS.Where("LOWER(card_no) = ?", lowerName).Find(&existing).Error; err != nil {
		return existing, fmt.Errorf("card not found: %w", err)
	}

	if len(existing) == 0 {
		return existing, nil
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   false,
		"is_delete":   true,
		"delete_by":   userId,
		"delete_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.CardEntity{}).
		Where("LOWER(card_no) = ?", lowerName).
		Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update card: %w", err)
	}

	// ดึงข้อมูลใหม่หลังอัปเดต
	if err := config.DB_POS.Where("LOWER(card_no) = ?", lowerName).Find(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated card: %w", err)
	}

	return existing, nil
}

func CheckCardByTel(tel string) (*[]models.CardListData, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
	SELECT 
  ce.card_no,
  ct.name AS card_type,
  ce.member_tel AS mobile_no,
  COALESCE(cd_summary.e_coin, 0) AS e_coin,
  COALESCE(cd_summary.e_bonus, 0) AS e_bonus,
  COALESCE(cd_summary.topup_amount, 0) AS topup_amount
FROM card_entity ce
LEFT JOIN card_type ct ON ce.card_type_id = ct.id
LEFT JOIN (
  SELECT 
    card_no, 
    SUM(balance_coin) AS e_coin, 
    SUM(
      CASE 
        WHEN bonus_expire_date IS NOT NULL 
             AND bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')
        THEN balance_bonus
        ELSE 0
      END
    ) AS e_bonus,
    MAX(amount) AS topup_amount
  FROM card_deposit
  WHERE is_active = true
  GROUP BY card_no
) cd_summary ON ce.card_no = cd_summary.card_no
WHERE ce.member_tel = ?
  AND ce.is_active = true
  AND ce.is_delete = false
order by ce.create_date desc
`

	var result []models.CardListData
	tx := config.DB_POS.Raw(query, tel).Scan(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil // ไม่พบข้อมูล
	}
	fmt.Printf("%+v\n", result)

	return &result, nil
}

func LockCardEntity(id string, is_active bool, userId int) (models.CardEntity, error) {
	var existing models.CardEntity

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("card entity not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"is_active":   is_active,
		"update_by":   userId,
		"update_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.CardEntity{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update card entity: %w", err)
	}

	// ดึงข้อมูลใหม่หลังอัปเดต
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated card entity: %w", err)
	}

	return existing, nil
}

// find card entity by card no
func FindCardEntityByCardNo(cardNo string) (*models.CardEntity, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	var cardEntity models.CardEntity
	lowerCard := strings.ToLower(cardNo)
	result := config.DB_POS.
		Where("LOWER(card_no) = ? and is_active = true and is_delete = false", lowerCard).
		Find(&cardEntity)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &cardEntity, nil
}

// update member tel to card entity
func UpdateMemberTelToCardEntity(cardEntityId uuid.UUID, entityMobile string, cardNo string, memberTel string) (*uuid.UUID, error) {
	now := utils.TimeNowAsia()
	// defMobile := "0000000000"
	if strings.ToLower(cardNo) == "card id" {
		return nil, nil
	}
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	if strings.TrimSpace(memberTel) != strings.TrimSpace(entityMobile) {
		config.DB_POS.Exec("UPDATE card_entity SET member_tel = ?, update_date = ? WHERE id = ?", memberTel, now, cardEntityId)
		return &cardEntityId, nil
	}
	return nil, nil
}

func PermanentlyDeleteCard(cardNo string) error {
	if config.DB_POS == nil {
		return fmt.Errorf("database connection is nil")
	}
	// get POS transaction by card no
	var posTransactions []models.PosTransaction
	config.DB_POS.Where("card_no = ?", cardNo).Find(&posTransactions)
	if len(posTransactions) > 0 {
		for _, posTransaction := range posTransactions {
			// delete pos sub transaction
			config.DB_POS.Where("bill_no = ?", posTransaction.BillNo).Delete(&models.PosSubTransaction{})
		}
		// delete pos transaction
		config.DB_POS.Where("card_no = ?", cardNo).Delete(&models.PosTransaction{})
	}
	// delete card play
	cardPlayDelete := make(chan bool)
	go PermanentlyDeleteCardPlay(cardNo, cardPlayDelete)
	// delete card deposit
	cardDepositDelete := make(chan bool)
	go PermanentlyDeleteCardDeposit(cardNo, cardDepositDelete)
	// delete card withdraw
	cardWithdrawDelete := make(chan bool)
	go PermanentlyDeleteCardWithdraw(cardNo, cardWithdrawDelete)

	// delete card entity
	cardEntityDelete := make(chan bool)
	go PermanentlyDeleteCardEntity(cardNo, cardEntityDelete)
	if <-cardEntityDelete && <-cardPlayDelete && <-cardDepositDelete && <-cardWithdrawDelete {
		return nil
	} else {
		return fmt.Errorf("failed to delete card")
	}
}

// Permanent Delete card Play
func PermanentlyDeleteCardPlay(cardNo string, cardPlayDelete chan<- bool) {
	if config.DB_POS == nil {
		cardPlayDelete <- false
	}
	// get card play by card no
	var cardPlays models.CardPlay
	config.DB_POS.Where("card_no = ?", cardNo).Delete(&cardPlays)
	cardPlayDelete <- true
	defer close(cardPlayDelete)
}

// Permanent Delete card Deposit
func PermanentlyDeleteCardDeposit(cardNo string, cardDepositDelete chan<- bool) {
	if config.DB_POS == nil {
		cardDepositDelete <- false
	}
	// get card deposit by card no
	var cardDeposits models.CardDeposit
	config.DB_POS.Where("card_no = ?", cardNo).Delete(&cardDeposits)
	cardDepositDelete <- true
	defer close(cardDepositDelete)
}

// Permanent Delete card withdraw
func PermanentlyDeleteCardWithdraw(cardNo string, cardWithdrawDelete chan<- bool) {
	if config.DB_POS == nil {
		cardWithdrawDelete <- false
	}
	// get card withdraw by card no
	var cardWithdraws models.CardWithdraw
	config.DB_POS.Where("card_no = ?", cardNo).Delete(&cardWithdraws)
	cardWithdrawDelete <- true
	defer close(cardWithdrawDelete)
}

// Permanent Delete card entity
func PermanentlyDeleteCardEntity(cardNo string, cardEntityDelete chan<- bool) {
	if config.DB_POS == nil {
		cardEntityDelete <- false
	}
	// get card entity by card no
	var cardEntities models.CardEntity
	config.DB_POS.Where("card_no = ?", cardNo).Delete(&cardEntities)
	cardEntityDelete <- true
	defer close(cardEntityDelete)
}

func CheckCardInfoWithExpire(cardNo string) (*models.CardDetailDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var (
		result        models.CardDetailDto
		balance       *models.CardSummaryDto
		cardExpire    *models.CardExpireDto
		balanceETimes *int
		eBonusExpire  *models.BonusExpireDto
		isMachine     *bool
		err           error
	)

	var wg sync.WaitGroup
	errChan := make(chan error, 5)

	// หา balance coin and bonus
	wg.Add(1)
	go func() {
		defer wg.Done()
		balance, err = FindBalanceCoinAndBonus(cardNo)
		if err != nil {
			errChan <- err
		}
	}()

	// หา วันหมดอายุบัตร
	wg.Add(1)
	go func() {
		defer wg.Done()
		cardExpire, err = FindCardExpire(cardNo)
		if err != nil {
			errChan <- err
		}
	}()

	// หา ยอดคงเหลือครั้ง
	wg.Add(1)
	go func() {
		defer wg.Done()
		balanceETimes, err = FindBalnceETimes(cardNo)
		if err != nil {
			errChan <- err
		}
	}()

	// หา วันหมดอายุ E-Bonus
	wg.Add(1)
	go func() {
		defer wg.Done()
		eBonusExpire, err = FindEBonusExpire(cardNo)
		if err != nil {
			errChan <- err
		}
	}()

	// หา is machine
	wg.Add(1)
	go func() {
		defer wg.Done()
		isMachine, err = FindCardMachine(cardNo)
		if err != nil {
			errChan <- err
		}

	}()

	wg.Wait()
	close(errChan)

	// ตรวจ error จาก goroutine
	for e := range errChan {
		if e != nil {
			return nil, e
		}
	}

	if balance != nil {
		result.CardNo = balance.CardNo
		result.CardType = balance.CardType
		result.ShowBalance = balance.ShowBalance
		result.MemberTel = balance.MemberTel
		result.ECoin = balance.ECoin
		result.EBonus = balance.EBonus
		result.DiscountCash = balance.DiscountCash
		result.CustomerTier = balance.CustomerTier
	}

	if cardExpire != nil {
		result.CardExpDate = cardExpire.CardExpDate
		result.CardExpDays = cardExpire.CardExpDays
	}

	if balanceETimes != nil {
		result.ETimes = *balanceETimes
	}

	if eBonusExpire != nil {
		result.EBonusExpDate = &eBonusExpire.EBonusExpDate
		result.EBonusExpDays = eBonusExpire.EBonusExpDays
	}
	result.IsMachine = isMachine != nil && *isMachine

	return &result, nil
}

func CheckCardInfoV2(cardNo string) (*models.CardDetailDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
	SELECT 
		ce.card_no,
		ct.name AS card_type,
		ct.show_balance,
		ce.member_tel,
		COALESCE(cd_summary.e_coin, 0) AS e_coin,
		COALESCE(cd_summary.e_bonus, 0) AS e_bonus,
		cd_summary.e_bonus_exp_date,
		COALESCE(cd_summary.discount_cash, 0) AS discount_cash,
		cd_summary.card_exp_date,
		COALESCE(cp_summary.e_times, 0) AS e_times,
		COALESCE(cp_summary.is_machine, false) AS is_machine,
		cust_tier.tier as customer_tier
	FROM card_entity ce
	LEFT JOIN card_type ct ON ce.card_type_id = ct.id
	LEFT JOIN (
		SELECT 
			card_no,
			SUM(balance_coin) AS e_coin,
			SUM(CASE WHEN (bonus_expire_date IS NOT NULL AND bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')) THEN balance_bonus ELSE 0 END) AS e_bonus,
			MIN(CASE WHEN (balance_bonus > 0 AND bonus_expire_date IS NOT NULL AND bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')) THEN bonus_expire_date END) AS e_bonus_exp_date,
			SUM(CASE WHEN (discount_cash_expire IS NOT NULL AND discount_cash_expire >= (NOW() AT TIME ZONE 'Asia/Bangkok')) THEN balance_discount_cash ELSE 0 END) AS discount_cash,
			MAX(card_expire_date) AS card_exp_date
		FROM card_deposit
		WHERE is_active = TRUE
		GROUP BY card_no
	) cd_summary ON ce.card_no = cd_summary.card_no
	LEFT JOIN (
		SELECT 
			cp.card_no,
			SUM(COALESCE(cpt.play_time, 0) - COALESCE(cpt.used_time, 0)) AS e_times,
			BOOL_OR(cp.play_machine) AS is_machine
		FROM card_play cp
		LEFT JOIN card_play_type cpt ON cp.id = cpt.card_play_id
		WHERE cp.is_delete = false
		GROUP BY cp.card_no
	) cp_summary ON ce.card_no = cp_summary.card_no
	LEFT JOIN (
		SELECT DISTINCT ON (tel) tel, tier
		FROM customer_tier
		ORDER BY tel, create_date DESC
	) cust_tier ON ce.member_tel = cust_tier.tel
	WHERE LOWER(ce.card_no) = LOWER(?)
	  AND ce.is_active = TRUE
	  AND ce.is_delete = FALSE
	`

	var rawResult struct {
		CardNo        string
		CardType      string
		ShowBalance   string
		MemberTel     string
		ECoin         int
		EBonus        int
		EBonusExpDate *time.Time
		DiscountCash  float64
		CardExpDate   *time.Time
		ETimes        int
		IsMachine     bool
		CustomerTier  string
	}

	tx := config.DB_POS.Raw(query, cardNo).Scan(&rawResult)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil // ไม่พบข้อมูล
	}

	result := models.CardDetailDto{
		CardNo:       rawResult.CardNo,
		CardType:     rawResult.CardType,
		ShowBalance:  rawResult.ShowBalance,
		MemberTel:    rawResult.MemberTel,
		ECoin:        rawResult.ECoin,
		EBonus:       rawResult.EBonus,
		ETimes:       rawResult.ETimes,
		DiscountCash: rawResult.DiscountCash,
		IsMachine:    rawResult.IsMachine,
		CustomerTier: rawResult.CustomerTier,
	}

	if rawResult.EBonusExpDate != nil {
		result.EBonusExpDate = rawResult.EBonusExpDate
		result.EBonusExpDays = utils.CalculateDate(*rawResult.EBonusExpDate)
	}

	if rawResult.CardExpDate != nil {
		result.CardExpDate = rawResult.CardExpDate
		result.CardExpDays = utils.CalculateDate(*rawResult.CardExpDate)
	}

	return &result, nil
}

// TODO: verify package
func VerifyCardFullV2(cardNo string, gamePrice int, deductCondition string) (*models.CardDetailDto, string, error) {
	if config.DB_POS == nil {
		return nil, "", fmt.Errorf("database connection is nil")
	}

	query := `
	WITH target_card AS (
		SELECT id, card_no, member_tel, card_type_id, is_active, is_delete
		FROM card_entity
		WHERE LOWER(card_no) = LOWER(?)
		  AND is_active = true
		  AND is_delete = false
	)
	SELECT 
		tc.card_no,
		tc.is_active,
		tc.is_delete,
		ct.name AS card_type,
		ct.show_balance,
		tc.member_tel,
		COALESCE(cd.e_coin, 0) AS e_coin,
		COALESCE(cd.e_bonus, 0) AS e_bonus,
		cd.e_bonus_exp_date,
		COALESCE(cd.discount_cash, 0) AS discount_cash,
		cd.card_exp_date,
		COALESCE(cp.e_times, 0) AS e_times,
		COALESCE(cp.is_machine, false) AS is_machine,
		cust.tier as customer_tier
	FROM target_card tc
	LEFT JOIN card_type ct ON tc.card_type_id = ct.id
	LEFT JOIN LATERAL (
		SELECT 
			SUM(CASE WHEN card_no = tc.card_no THEN balance_coin ELSE 0 END) AS e_coin,
			SUM(
				CASE 
					WHEN card_no = tc.card_no
					AND bonus_expire_date IS NOT NULL 
					AND bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')
					THEN balance_bonus 
					ELSE 0 
				END
			) AS e_bonus,
			MIN(
				CASE 
					WHEN card_no = tc.card_no
					AND balance_bonus > 0 
					AND bonus_expire_date IS NOT NULL 
					AND bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')
					THEN bonus_expire_date 
				END
			) AS e_bonus_exp_date,
			SUM(CASE WHEN (discount_cash_expire IS NOT NULL AND discount_cash_expire >= (NOW() AT TIME ZONE 'Asia/Bangkok')) THEN balance_discount_cash ELSE 0 END) AS discount_cash,
			MAX(CASE WHEN card_no = tc.card_no THEN card_expire_date END) AS card_exp_date
		FROM card_deposit
		WHERE (
			(tc.member_tel = '0000000000' OR tc.member_tel = '' OR tc.member_tel IS NULL) AND card_no = tc.card_no
			OR
			(tc.member_tel != '0000000000' AND tc.member_tel != '' AND tc.member_tel IS NOT NULL AND member_tel = tc.member_tel)
		)
		AND is_active = TRUE
	) cd ON TRUE
	LEFT JOIN LATERAL (
		SELECT 
			SUM(COALESCE(cpt.play_time, 0) - COALESCE(cpt.used_time, 0)) AS e_times,
			BOOL_OR(cp.play_machine) AS is_machine
		FROM card_play cp
		LEFT JOIN card_play_type cpt ON cp.id = cpt.card_play_id
		WHERE cp.card_no = tc.card_no AND cp.is_delete = false
	) cp ON TRUE
	LEFT JOIN LATERAL (
		SELECT tier
		FROM customer_tier
		WHERE tel = tc.member_tel
		ORDER BY create_date DESC
		LIMIT 1
	) cust ON TRUE
	`

	var rawResult struct {
		CardNo        string
		IsActive      bool
		IsDelete      bool
		CardType      string
		ShowBalance   string
		MemberTel     string
		ECoin         int
		EBonus        int
		EBonusExpDate *time.Time
		DiscountCash  float64
		CardExpDate   *time.Time
		ETimes        int
		IsMachine     bool
		CustomerTier  string
	}

	tx := config.DB_POS.Raw(query, cardNo).Scan(&rawResult)
	if tx.Error != nil {
		fmt.Printf("VerifyCardFullV2 DB Error for card %s: %v\n", cardNo, tx.Error)
		return nil, "", tx.Error
	}

	fmt.Printf("VerifyCardFullV2 Query Result - CardNo: %s, RowsAffected: %d, IsDelete: %v, IsActive: %v\n",
		cardNo, tx.RowsAffected, rawResult.IsDelete, rawResult.IsActive)

	if tx.RowsAffected == 0 || rawResult.IsDelete {
		fmt.Printf("VerifyCardFullV2 Card not found or deleted: RowsAffected=%d, IsDelete=%v\n", tx.RowsAffected, rawResult.IsDelete)
		return nil, "not_found", nil
	}

	if !rawResult.IsActive {
		fmt.Printf("VerifyCardFullV2 Card locked: %s\n", cardNo)
		return nil, "locked", nil
	}

	if rawResult.CardType == "Normal" {
		switch deductCondition {
		case "ecoin_first", "ebonus_first":
			if (rawResult.ECoin + rawResult.EBonus) < gamePrice {
				return nil, "not_enough", nil
			}
		case "ecoin_only":
			if rawResult.ECoin < gamePrice {
				return nil, "not_enough", nil
			}
		case "ebonus_only":
			if rawResult.EBonus < gamePrice {
				return nil, "not_enough", nil
			}
		}
	}

	result := models.CardDetailDto{
		CardNo:       rawResult.CardNo,
		CardType:     rawResult.CardType,
		ShowBalance:  rawResult.ShowBalance,
		MemberTel:    rawResult.MemberTel,
		ECoin:        rawResult.ECoin,
		EBonus:       rawResult.EBonus,
		ETimes:       rawResult.ETimes,
		DiscountCash: rawResult.DiscountCash,
		IsMachine:    rawResult.IsMachine,
		CustomerTier: rawResult.CustomerTier,
	}

	if rawResult.EBonusExpDate != nil {
		result.EBonusExpDate = rawResult.EBonusExpDate
		result.EBonusExpDays = utils.CalculateDate(*rawResult.EBonusExpDate)
	}

	if rawResult.CardExpDate != nil {
		result.CardExpDate = rawResult.CardExpDate
		result.CardExpDays = utils.CalculateDate(*rawResult.CardExpDate)
	}

	return &result, "valid", nil
}

func FindBalanceCoinAndBonus(cardNo string) (*models.CardSummaryDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	// 	query := `
	// SELECT
	//     ce.card_no,
	//     ct.name AS card_type,
	//     ct.show_balance,
	//     ce.member_tel,
	//     COALESCE(cd_summary.e_coin, 0) AS e_coin,
	//     COALESCE(cd_summary.e_bonus, 0) AS e_bonus,
	// 	COALESCE(cd_summary.discount_cash, 0) AS discount_cash,
	// 	cust_tier.tier as customer_tier
	// FROM card_entity ce
	// LEFT JOIN card_type ct ON ce.card_type_id = ct.id
	// LEFT JOIN (
	//     SELECT
	//         card_no,
	//         SUM(balance_coin) AS e_coin,
	//         SUM(
	//             CASE
	//                 WHEN bonus_expire_date IS NOT NULL
	//                      AND bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')
	//                 THEN balance_bonus
	//                 ELSE 0
	//             END
	//         ) AS e_bonus,
	// 		 SUM(
	//             CASE
	//                 WHEN discount_cash_expire IS NOT NULL
	//                      AND discount_cash_expire >= (NOW() AT TIME ZONE 'Asia/Bangkok')
	//                 THEN balance_discount_cash
	//                 ELSE 0
	//             END
	//         ) AS discount_cash
	//     FROM card_deposit
	//     WHERE is_active = TRUE
	//     GROUP BY card_no
	// ) cd_summary ON ce.card_no = cd_summary.card_no
	// LEFT JOIN (
	// SELECT DISTINCT ON (tel) tel, tier
	// 		FROM customer_tier
	// 		ORDER BY tel, create_date DESC
	// ) cust_tier on ce.member_tel = cust_tier.tel
	// WHERE LOWER(ce.card_no) = ?
	//   AND ce.is_active = TRUE
	//   AND ce.is_delete = FALSE
	// `

	query := `
SELECT 
    ce.card_no,
    ct.name AS card_type,
    ct.show_balance,
    ce.member_tel,
    COALESCE(cd_summary.e_coin, 0) AS e_coin,
    COALESCE(cd_summary.e_bonus, 0) AS e_bonus,
	COALESCE(cd_summary.discount_cash, 0) AS discount_cash,
	cust_tier.tier as customer_tier
FROM card_entity ce
LEFT JOIN card_type ct ON ce.card_type_id = ct.id
LEFT JOIN LATERAL (
    SELECT
        SUM(balance_coin) AS e_coin,
        SUM(
            CASE
                WHEN bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')
                THEN balance_bonus
                ELSE 0
            END
        ) AS e_bonus,
        SUM(
            CASE
                WHEN discount_cash_expire >= (NOW() AT TIME ZONE 'Asia/Bangkok')
                THEN balance_discount_cash
                ELSE 0
            END
        ) AS discount_cash
    FROM card_deposit cd
    WHERE cd.is_active = TRUE
AND (
    (
        (ce.member_tel = '0000000000'
         OR ce.member_tel = ''
         OR ce.member_tel IS NULL)
        AND cd.card_no = ce.card_no
    )
    OR
    (
        ce.member_tel <> '0000000000'
        AND ce.member_tel <> ''
        AND ce.member_tel IS NOT NULL
        AND cd.member_tel = ce.member_tel
        AND cd.card_no = ce.card_no
    )
)
) cd_summary ON TRUE
LEFT JOIN (
SELECT DISTINCT ON (tel) tel, tier
		FROM customer_tier
		ORDER BY tel, create_date DESC
) cust_tier on ce.member_tel = cust_tier.tel
WHERE LOWER(ce.card_no) = ?
  AND ce.is_active = TRUE
  AND ce.is_delete = FALSE
`

	var result models.CardSummaryDto
	tx := config.DB_POS.Raw(query, strings.ToLower(cardNo)).Scan(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil // ไม่พบข้อมูล
	}
	fmt.Printf("%+v\n", result)

	return &result, nil
}

func FindEBonusExpire(cardNo string) (*models.BonusExpireDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var cardDeposit models.CardDeposit
	var result models.BonusExpireDto

	query := `
	SELECT *
	FROM card_deposit cd
	WHERE LOWER(cd.card_no) = ?
	  AND cd.is_active = true 
	  AND cd.balance_bonus != 0
	ORDER BY cd.bonus_expire_date ASC
	LIMIT 1
	`

	tx := config.DB_POS.Raw(query, strings.ToLower(cardNo)).Scan(&cardDeposit)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 || cardDeposit.BonusExpireDate == nil {
		return nil, nil // ไม่พบข้อมูล หรือไม่มีวันหมดอายุ
	}

	result.EBonusExpDate = *cardDeposit.BonusExpireDate
	result.EBonusExpDays = utils.CalculateDate(*cardDeposit.BonusExpireDate)

	return &result, nil
}

func FindBalnceETimes(cardNo string) (*int, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var result models.EtimesDto

	query := `
	SELECT 
		COALESCE(SUM(cpt.play_time), 0) AS play_time,
		COALESCE(SUM(cpt.used_time), 0) AS used_time
	FROM card_play cp
	LEFT JOIN card_play_type cpt ON cp.id = cpt.card_play_id
	WHERE LOWER(cp.card_no) = ?
	  AND cp.is_delete = false
	`

	tx := config.DB_POS.Raw(query, strings.ToLower(cardNo)).Scan(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}

	balanceTimes := result.PlayTime - result.UsedTime
	return &balanceTimes, nil
}

func FindCardExpire(cardNo string) (*models.CardExpireDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var result models.CardExpireDto

	query := `
	SELECT *
	FROM card_deposit cd
	WHERE LOWER(cd.card_no) = ?
	  AND cd.is_active = true 
	  AND cd.card_expire_date is not null
	ORDER BY cd.card_expire_date DESC
	LIMIT 1
	`

	tx := config.DB_POS.Raw(query, strings.ToLower(cardNo)).Scan(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	fmt.Printf("%+v\n", result)
	if tx.RowsAffected == 0 || !result.CardExpireDate.Valid {
		return nil, nil // ไม่พบข้อมูล หรือไม่มีวันหมดอายุ
	}
	fmt.Printf("Card Expire Date: %+v\n", result.CardExpireDate.Time)

	result.CardExpDate = &result.CardExpireDate.Time
	result.CardExpDays = utils.CalculateDate(*result.CardExpDate)

	return &result, nil
}

// Update Card Type to Card Entity
func UpdateCardTypeToCardEntity(cardEntityId uuid.UUID, cardTypeId uuid.UUID) (*uuid.UUID, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
	UPDATE card_entity
	SET card_type_id = ?
	WHERE id = ? 
	`

	tx := config.DB_POS.Exec(query, cardTypeId, cardEntityId)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &cardEntityId, nil
}

func CheckCardMemberList(params models.SearchLockCardParams) (utils.SearchResult, error) {
	var result utils.SearchResult

	if config.DB_POS == nil {
		return result, fmt.Errorf("database connection is nil")
	}

	// Base query
	baseQuery := `
	SELECT 
		ce.id AS card_entity_id,
		ce.member_tel AS mobile_no,
		ce.card_no,
		ct.name AS card_type,
		COALESCE(cd_summary.e_coin, 0) AS balance_e_coin,
		COALESCE(cd_summary.e_bonus, 0) AS balance_e_bonus,
		llc.id AS lock_card_id,
		COALESCE(llc.is_lock, false) AS is_lock,
		llc.lock_by,
		llc.lock_date,
		llc.lock_location,
		llc.lock_reason,
		llc.unlock_by,
		llc.unlock_date,
		llc.unlock_location,
		llc.unlock_reason,
		llc.ref_card_entity_id,
		COALESCE(ce.is_active, false) AS ref_card_entity_active
	FROM card_entity ce
	LEFT JOIN card_type ct ON ce.card_type_id = ct.id
	LEFT JOIN (
		SELECT 
			card_no, 
			SUM(balance_coin) AS e_coin, 
			SUM(
				CASE 
					WHEN bonus_expire_date IS NOT NULL 
						AND bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')
					THEN balance_bonus
					ELSE 0
				END
			) AS e_bonus
		FROM card_deposit
		WHERE is_active = true
		GROUP BY card_no
	) cd_summary ON ce.card_no = cd_summary.card_no
	LEFT JOIN (
    SELECT DISTINCT ON (card_no, member_tel)
        id,
        card_no,
        member_tel,
        is_lock,
        lock_by,
        lock_date,
        lock_location,
        lock_reason,
        unlock_by,
        unlock_date,
        unlock_location,
        unlock_reason,
        ref_card_entity_id
    FROM log_lock_card
    WHERE is_lock = true
    ORDER BY card_no, member_tel, lock_date DESC
) llc ON ce.card_no = llc.card_no AND ce.member_tel = llc.member_tel
	WHERE ce.is_delete = false
	`

	// Dynamic filter
	var args []interface{}
	if params.MemberTel != "" {
		baseQuery += " AND ce.member_tel ILIKE ?"
		args = append(args, "%"+params.MemberTel+"%")
	}
	if params.CardNo != "" {
		baseQuery += " AND ce.card_no ILIKE ?"
		args = append(args, "%"+params.CardNo+"%")
	}
	if params.IsLocked != nil {
		baseQuery += " AND llc.is_lock = ?"
		args = append(args, *params.IsLocked)
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS total", baseQuery)
	var totalCount int64
	if err := config.DB_POS.Raw(countQuery, args...).Scan(&totalCount).Error; err != nil {
		return result, fmt.Errorf("count query failed: %w", err)
	}

	offset := (params.Page - 1) * params.Skip

	// Add order + pagination
	finalQuery := baseQuery + " ORDER BY ce.create_date DESC LIMIT ? OFFSET ?"
	args = append(args, params.Skip, offset)

	// Run main query
	var cardList []models.CardMemberListData
	if err := config.DB_POS.Raw(finalQuery, args...).Scan(&cardList).Error; err != nil {
		return result, err
	}

	result = utils.SearchResult{
		Page:       params.Page,
		TotalCount: int(totalCount),
		Result:     cardList,
	}

	return result, nil
}

func CheckCardRefundByTel(tel string) (*[]models.CardListData, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
	WITH recent_cards AS (
    SELECT DISTINCT ON (cd.card_no)
        cd.card_no,
        ct.name AS card_type,
        cd.member_tel AS mobile_no,
        COALESCE(cd.update_date, cd.create_date) AS last_used_date
    FROM card_deposit cd
    INNER JOIN card_entity ce ON cd.card_no = ce.card_no
    INNER JOIN card_type ct ON ce.card_type_id = ct.id
    WHERE ce.member_tel = ? AND ct.name = 'Normal'
    ORDER BY cd.card_no, COALESCE(cd.update_date, cd.create_date) DESC
    LIMIT 5
)
SELECT 
    rc.card_no,
    rc.card_type,
    rc.mobile_no,
    COALESCE(cd_summary.e_coin, 0) AS e_coin,
    COALESCE(cd_summary.e_bonus, 0) AS e_bonus,
    COALESCE(cd_summary.topup_amount, 0) AS topup_amount,
    rc.last_used_date
FROM recent_cards rc
LEFT JOIN (
    SELECT 
        card_no,
        SUM(balance_coin) AS e_coin,
        SUM(
            CASE 
                WHEN bonus_expire_date IS NOT NULL AND bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')
                THEN balance_bonus
                ELSE 0
            END
        ) AS e_bonus,
        MAX(amount) AS topup_amount
    FROM card_deposit
    WHERE is_active = true
    GROUP BY card_no
) cd_summary ON rc.card_no = cd_summary.card_no
ORDER BY rc.last_used_date DESC NULLS LAST;
	`

	var result []models.CardListData
	tx := config.DB_POS.Raw(query, tel).Scan(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil // ไม่พบข้อมูล
	}

	fmt.Printf("%+v\n", result)
	return &result, nil
}

func FindCardMachine(cardNo string) (*bool, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var isMachine bool
	var cardPlay *models.CardPlay

	query := `
	SELECT *
	FROM card_play cp
	WHERE LOWER(cp.card_no) = ?
	  AND cp.play_machine = true
	LIMIT 1
	`

	tx := config.DB_POS.Raw(query, strings.ToLower(cardNo)).Scan(&cardPlay)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 || cardPlay == nil {
		isMachine = false
		return &isMachine, nil
	}
	isMachine = true

	return &isMachine, nil
}

func CheckCardPackageInfo(cardNo string) (*models.CardPackageSummaryDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var (
		result     models.CardPackageSummaryDto
		info       *models.CardPackageDto
		details    []models.CardPackageDetailDto
		cardExpire *models.CardExpireDto
		menuName   string
		// eBonusExpire *models.BonusExpireDto
		err error
	)

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	wg.Add(1)
	go func() {
		defer wg.Done()
		info, err = FindCardPackageInfo(cardNo)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		details, menuName, err = FindCardPackageDetail(cardNo)
		if err != nil {
			errChan <- err
		}
	}()

	// หา วันหมดอายุบัตร
	wg.Add(1)
	go func() {
		defer wg.Done()
		cardExpire, err = FindCardExpire(cardNo)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	// ตรวจ error จาก goroutine
	for e := range errChan {
		if e != nil {
			return nil, e
		}
	}

	if info != nil {
		result = models.CardPackageSummaryDto{
			CardNo:       info.CardNo,
			CardType:     info.CardType,
			MemberTel:    info.MemberTel,
			CustomerTier: info.CustomerTier,
			PosMenuName:  menuName,
		}
	}

	result.CardPackages = details

	if cardExpire != nil {
		result.CardExpDate = cardExpire.CardExpDate
		result.CardExpDays = cardExpire.CardExpDays
	}

	// if eBonusExpire != nil {
	// 	result.EBonusExpDate = &eBonusExpire.EBonusExpDate
	// 	result.EBonusExpDays = eBonusExpire.EBonusExpDays
	// }

	return &result, nil
}

func FindCardPackageInfo(cardNo string) (*models.CardPackageDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	query := `
SELECT 
    ce.card_no,
    ct.name AS card_type,
    ce.member_tel,
	cust_tier.tier as customer_tier
FROM card_entity ce
LEFT JOIN card_type ct ON ce.card_type_id = ct.id
LEFT JOIN (
SELECT DISTINCT ON (tel) tel, tier
		FROM customer_tier
		ORDER BY tel, create_date DESC
) cust_tier on ce.member_tel = cust_tier.tel
WHERE LOWER(ce.card_no) = ?
  AND ce.is_active = TRUE
  AND ce.is_delete = FALSE
`

	var result models.CardPackageDto
	tx := config.DB_POS.Raw(query, strings.ToLower(cardNo)).Scan(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil // ไม่พบข้อมูล
	}
	fmt.Printf("%+v\n", result)

	return &result, nil
}

func FindCardPackageDetail(cardNo string) ([]models.CardPackageDetailDto, string, error) {
	if config.DB_POS == nil {
		return nil, "", fmt.Errorf("database connection is nil")
	}

	var carDeposit models.CardDeposit
	var posMenu models.PosMenu
	queryCard := `
SELECT *
FROM card_deposit cd
WHERE LOWER(cd.card_no) = ?
AND cd.is_active = true
and coalesce(cd.pos_menu_id ,0) <> 0
order by cd.create_date desc
LIMIT 1`
	tx := config.DB_POS.Raw(queryCard, strings.ToLower(cardNo)).Scan(&carDeposit)
	if tx.Error != nil {
		return nil, "", tx.Error
	}

	if tx.RowsAffected == 0 {
		return nil, "", nil // ไม่พบข้อมูล
	}

	queryPosMenu := `
select * from pos_menu pm where id = ?
LIMIT 1`
	tx2 := config.DB_POS.Raw(queryPosMenu, carDeposit.PosMenuId).Scan(&posMenu)
	if tx2.Error != nil {
		return nil, "", tx2.Error
	}

	query := `
SELECT 
    cpm.machine_id,
    msg.machine_name,
    COALESCE(msg.e_coin, 0) AS e_coin,
    COALESCE(msg.e_bonus, 0) AS e_bonus,
    COALESCE(msg.play_time, 0) AS play_time,
    COALESCE(cpm.ecoin, 0) AS remain_e_coin,
    COALESCE(cpm.ebonus, 0) AS remain_e_bonus,
    CASE 
        WHEN cpm.play_time = -1 THEN 0 
        ELSE COALESCE(cpm.play_time, 0) 
    END AS remain_play_time,
    COALESCE(msg.e_coin, 0) - COALESCE(cpm.ecoin, 0) as used_e_coin,
    COALESCE(msg.e_bonus, 0) - COALESCE(cpm.ebonus, 0) as used_e_bonus,
    
    COALESCE(msg.play_time, 0) - 
    CASE 
        WHEN cpm.play_time = -1 THEN 0 
        ELSE COALESCE(cpm.play_time, 0) 
    END AS used_play_time 
FROM card_play cp
LEFT JOIN card_play_machine cpm 
    ON cp.id = cpm.card_play_id
INNER JOIN (
    select msg2.e_bonus, 
    msg2.e_coin, 
    msg2.play_time,
    msg2.machine_id,
    msg2.machine_name 
    from pos_menu pm 
    left join machine_group mg on pm.machine_group_id = mg.id
    left join machine_sub_group msg2 on mg.id = msg2.group_id
    where pm.id = ?
) msg 
    ON msg.machine_id = cpm.machine_id
WHERE LOWER(cp.card_no) = ?
AND cp.is_delete = false 
order by cpm.machine_id
`

	var result []models.CardPackageDetailDto
	tx1 := config.DB_POS.Raw(query, carDeposit.PosMenuId, strings.ToLower(cardNo)).Scan(&result)
	if tx1.Error != nil {
		return nil, "", tx1.Error
	}

	return result, posMenu.Description, nil
}

func FindActiveCardNoWithCardType(cardNo string) (*models.CardEntityType, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var cardEntity models.CardEntityType
	lowerCard := strings.ToLower(cardNo)

	result := config.DB_POS.
		Table("card_entity ce").
		Select("ct.name as card_type_name").
		Joins("LEFT JOIN card_type ct ON ct.id = ce.card_type_id").
		Where("LOWER(ce.card_no) = ? AND ce.is_active = true AND ce.is_delete = false", lowerCard).
		Scan(&cardEntity)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &cardEntity, nil
}

func CheckCardPackageInfoWithExpire(cardNo string) (*models.CardDetailDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var (
		result       models.CardDetailDto
		info         *models.CardPackageDto
		balance      *models.CardPackageAccDto
		cardExpire   *models.CardExpireDto
		eBonusExpire *models.BonusExpireDto
		err          error
	)

	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	// info
	wg.Add(1)
	go func() {
		defer wg.Done()
		info, err = FindCardPackageInfo(cardNo)
		if err != nil {
			errChan <- err
		}
	}()

	// หา balance coin and bonus
	wg.Add(1)
	go func() {
		defer wg.Done()
		balance, err = FindCardPackageAcc(cardNo)
		if err != nil {
			errChan <- err
		}
	}()

	// หา วันหมดอายุบัตร
	wg.Add(1)
	go func() {
		defer wg.Done()
		cardExpire, err = FindCardExpire(cardNo)
		if err != nil {
			errChan <- err
		}
	}()

	// หา วันหมดอายุ E-Bonus
	wg.Add(1)
	go func() {
		defer wg.Done()
		eBonusExpire, err = FindEBonusExpire(cardNo)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	// ตรวจ error จาก goroutine
	for e := range errChan {
		if e != nil {
			return nil, e
		}
	}

	if balance != nil {
		result.CardNo = info.CardNo
		result.CardType = info.CardType
		result.ShowBalance = ""
		result.MemberTel = info.MemberTel
		result.ECoin = balance.ECoin
		result.EBonus = balance.EBonus
		result.DiscountCash = 0
		result.CustomerTier = info.CustomerTier
		result.ETimes = balance.PlayTime
	}

	if cardExpire != nil {
		result.CardExpDate = cardExpire.CardExpDate
		result.CardExpDays = cardExpire.CardExpDays
	}

	if eBonusExpire != nil {
		result.EBonusExpDate = &eBonusExpire.EBonusExpDate
		result.EBonusExpDays = eBonusExpire.EBonusExpDays
	}
	result.IsMachine = true

	return &result, nil
}

func FindCardPackageAcc(cardNo string) (*models.CardPackageAccDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	query := `
SELECT 
    SUM(COALESCE(cpm.ecoin, 0)) AS e_coin,
    SUM(COALESCE(cpm.ebonus, 0)) AS e_bonus,
    SUM(
        CASE 
            WHEN cpm.play_time = -1 THEN 0 
            ELSE COALESCE(cpm.play_time, 0)
        END
    ) AS play_time
FROM card_play cp
LEFT JOIN card_play_machine cpm 
    ON cp.id = cpm.card_play_id
WHERE LOWER(cp.card_no) = ?
AND cp.is_delete = false
`

	var result models.CardPackageAccDto
	tx := config.DB_POS.Raw(query, strings.ToLower(cardNo)).Scan(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &result, nil
}

func FindPlayTimesByCardNo(cardNo string) (*models.EtimesDto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var result models.EtimesDto

	query := `
	SELECT 
		COALESCE(SUM(cpt.play_time), 0) AS play_time,
		COALESCE(SUM(cpt.used_time), 0) AS used_time
	FROM card_play cp
	LEFT JOIN card_play_type cpt ON cp.id = cpt.card_play_id
	WHERE LOWER(cp.card_no) = ?
	  AND cp.is_delete = false
	`

	tx := config.DB_POS.Raw(query, strings.ToLower(cardNo)).Scan(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &result, nil
}

func FindCardPackageDetailByCardNoAndPosMenu(cardNo string, posMenuId int) ([]models.CardPackageDetailV1Dto, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
SELECT 
    cpm.machine_id,
    msg.machine_name,
    msg.category,
    COALESCE(msg.e_coin, 0) AS e_coin,
    COALESCE(msg.e_bonus, 0) AS e_bonus,
    COALESCE(msg.play_time, 0) AS play_time,
    COALESCE(cpm.ecoin, 0) AS remain_e_coin,
    COALESCE(cpm.ebonus, 0) AS remain_e_bonus,
    CASE 
        WHEN cpm.play_time = -1 THEN 0 
        ELSE COALESCE(cpm.play_time, 0) 
    END AS remain_play_time
FROM card_play cp
LEFT JOIN card_play_machine cpm 
    ON cp.id = cpm.card_play_id
INNER JOIN (
    select msg2.e_bonus, 
    msg2.e_coin, 
    msg2.play_time,
    msg2.machine_id,
    msg2.machine_name,
    msg2.category
    from pos_menu pm 
    left join machine_group mg on pm.machine_group_id = mg.id
    left join machine_sub_group msg2 on mg.id = msg2.group_id
    where pm.id = ?
) msg 
    ON msg.machine_id = cpm.machine_id
WHERE LOWER(cp.card_no) = ?
AND cp.is_delete = false 
order by cpm.machine_id
`

	var result []models.CardPackageDetailV1Dto
	tx1 := config.DB_POS.Raw(query, posMenuId, strings.ToLower(cardNo)).Scan(&result)
	if tx1.Error != nil {
		return nil, tx1.Error
	}

	return result, nil
}
