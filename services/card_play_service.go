package services

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	// "log"
	"new-pos-api/config"
	"new-pos-api/models"
	model_v2 "new-pos-api/models/v2"
	"new-pos-api/utils"
	"time"

	"github.com/google/uuid"
)

func CreateCardPlay(input models.NewCardPlayDto, userId int) (models.CardPlay, error) {
	fmt.Println("Creating card play with input:", input)
	var entity models.CardPlay

	entity.CardNo = input.CardNo
	entity.PlayBranch = input.PlayBranch
	entity.PlayMachine = input.PlayMachine
	entity.CreateBy = userId
	entity.CardDepositId = input.CardDepositId
	entity.CreateDate = *utils.TimeNowAsia()

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create card entity: %w", err)
	}

	return entity, nil
}

// VerifyCardPlay checks if the card play is valid and returns a boolean value
func VerifyCardPlay(input models.CardPlayVerifyDto, userId int) (*uuid.UUID, error) {
	var entity models.CardPlay
	isValid := true

	if config.DB_POS == nil {
		return nil, fmt.Errorf("database pos connection is nil")
	}

	//TODO: card_deposit SET IS_ACTIVE = TRUE

	deposit := models.CardDeposit{}
	if err := config.DB_POS.Where("card_no = ? AND is_active = true AND pos_menu_id >= 0", input.CardNo).First(&deposit).Error; err != nil {
		return nil, fmt.Errorf("failed to verify card deposit: %w", err)
	}
	fmt.Printf("Deposit found: %+v\n", deposit)
	if err := config.DB_POS.Where("is_delete = false and card_no = ?", input.CardNo).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to verify card play: %w", err)
	}

	// verify card entity
	_, err := VerifyCardEntity(input.CardNo)
	if err != nil {
		fmt.Println("failed to verify card entity: %w", err)
		isValid = false
		return nil, fmt.Errorf("failed to verify card entity: %w", err)
	}

	if !isValid {
		return nil, fmt.Errorf("1 . card entity is not valid")
	}

	// verify cards deposit

	// verify cards play
	validCardPlay, _ := ValiddateCardPlay(input.CardNo)
	if validCardPlay == nil {
		isValid = false
	}
	if !isValid {
		return nil, fmt.Errorf("2. card play is not valid")
	}
	// input
	fmt.Println("input :", input)
	// prepare sum e_coin and e_bonus
	coin := make(chan int64)
	go SumCoin(validCardPlay, coin)
	AmtCoin := <-coin // Assuming this will block until the sum is received
	fmt.Println("Sum Coin:", AmtCoin)

	bonus := make(chan int64)
	go SumBonus(validCardPlay, bonus)
	AmtBonus := <-bonus // Assuming this will block until the sum is received
	fmt.Println("Sum Bonus:", AmtBonus)

	for _, cardPlay := range validCardPlay {
		depositID := cardPlay.CardDepositId
		isValid = false
		//fmt.Printf("Card Play ID: %s Deposit ID: %v period %v - %v limit %d  brach %t machine %t order %d\n", cardPlay.CardPlayId, cardPlay.CardDepositId, cardPlay.Period.StartDate, cardPlay.Period.EndDate, *cardPlay.LimitUsage, cardPlay.PlayBranch, cardPlay.PlayMachine, cardPlay.ConditionID)
		// if cardPlay.BranchList != nil {
		// 	for _, branch := range *cardPlay.BranchList {
		// 		fmt.Printf("Branch Code: %s\n", branch.BranchCode)
		// 	}
		// }

		// condition 1.1 card_play_type period
		fmt.Println("Condition 1.1")
		fmt.Println("Period Start Date:", cardPlay.Period.StartDate, cardPlay.Period.StartDate.IsZero(), "Period End Date:", cardPlay.Period.EndDate.IsZero())
		if !cardPlay.Period.StartDate.IsZero() && !cardPlay.Period.EndDate.IsZero() {
			var now = time.Now()
			inTime := now.Before(cardPlay.Period.EndDate) && now.After(cardPlay.Period.StartDate)
			if !inTime {
				//TODO disable card play
				_ = UpdateCardPlayIsDelete(depositID)
				// TODO disable card play type
				_ = UpdateCardPlayTypeIsDelete(cardPlay.CardPlayId)
				fmt.Println("c1 not in time")
				continue
			}
			fmt.Printf("c1 in time:%v card_play_id:%s\n", inTime, cardPlay.CardPlayId)
		}

		// condition 1.2 card_play_type limit usage
		fmt.Println("Condition 1.2")
		if cardPlay.LimitUsage != nil && *cardPlay.LimitUsage > 0 {
			if cardPlay.UsedTime != nil && *cardPlay.UsedTime >= *cardPlay.LimitUsage {
				//TODO disable card play
				_ = UpdateCardPlayIsDelete(depositID)
				// TODO disable card play type
				_ = UpdateCardPlayTypeIsDelete(cardPlay.CardPlayId)
				fmt.Println("c2 limit usage reached")
				continue
			}
		}
		// condition 1.3 card_play_branch
		fmt.Println("Condition 1.3 card_play_branch")
		if cardPlay.PlayBranch && len(*cardPlay.BranchList) > 0 {
			// check if the branch exists in the branch list
			found := false
			for _, branch := range *cardPlay.BranchList {
				if branch.BranchCode == input.PlayBranch {
					found = true
					break
				}
			}
			if !found {
				fmt.Println("c3 play branch not found")
				continue
			}

		}
		// condition 1.4 card_play_machine
		fmt.Println("Condition 1.4 card_play_machine")
		if cardPlay.PlayMachine && len(*cardPlay.MachineList) > 0 {
			found := false
			machineList := *cardPlay.MachineList
			machineId, _ := strconv.Atoi(input.PlayMachine)
			for _, machine := range machineList {
				if machine.Machine == machineId {
					fmt.Println("Found play machine:", machine.Machine)
					found = true
					break
				}
			}
			if !found {
				fmt.Println("c4 play machine not found")
				continue
			}
		}

		// condition 1.5 card_deposit balance coin
		fmt.Println("Condition 1.5 card_deposit balance coin")
		userCoin := input.UseCoin
		userBonus := input.UseBonus
		if userCoin != nil && *userCoin > 0 {
			if AmtCoin < int64(*userCoin) {
				return nil, fmt.Errorf("not enough coin balance")
			}
		}

		// condition 1.6 card_deposit balance bonus
		fmt.Println("Condition 1.6 card_deposit balance bonus")

		if userBonus != nil && *userBonus > 0 {
			if AmtBonus < int64(*userBonus) {
				return nil, fmt.Errorf("not enough bonus balance")
			}
		}
		// fmt.Println("card play available :", cardPlay.CardPlayId)
		return cardPlay.CardPlayId, nil
	}

	// verify card play type

	// varify card play branch

	// verify card play machine

	return nil, fmt.Errorf("card play is not valid")
}

// DeductCardPlay deducts the amount from the card play
func DeductCardPlay(input models.CardPlayDeductDto, userId int) ([]Deposit, error) {
	var Withdraws []Deposit
	cardPlayId := input.CardPlayId
	cardNo := input.CardNo
	// create withdraw
	fromChannel := "POS"
	if input.FromChannel != nil {
		fromChannel = *input.FromChannel
	}
	// get member_tel
	memberTel := make(chan string)
	go GetMemberTel(input.CardNo, memberTel)
	MemberTel := <-memberTel

	if cardPlayId == nil || cardNo == "" {
		return nil, fmt.Errorf("card play ID or card number is empty")
	}

	// deduct card play type used
	//
	// งานเดียวของขั้นตอนนี้คือเพิ่มตัวนับ used_time ใน card_play_type
	//
	// เดิมมีการ append เข้า Withdraws ต่อท้ายตรงนี้ด้วย โดยใส่ cpt.CardPlayId
	// ซึ่งเป็น card_play.id ลงช่องที่ปลายทางเอาไปใช้เป็น card_deposit.id
	// (UpdateCardDepositBalance ทำ UPDATE card_deposit WHERE id = ?)
	// และใส่จำนวนเป็น 0 ทั้ง coin และ bonus จึงไม่เคยหักอะไรได้เลยตั้งแต่ต้น
	//
	// ตอนที่ UpdateCardDepositBalance ยังไม่เช็ค RowsAffected แถวนี้ไม่มีพิษภัย
	// UPDATE ไม่โดนแถวไหนแล้วเงียบผ่านไป ส่วน CreateCardWithdraw ก็ลงแถวได้
	// เพราะ card_withdraw.card_deposit_id ไม่มี foreign key
	// (ตรวจ UAT 2026-09-17: มีแถวขยะแบบนี้สะสมอยู่ 221 แถว และไม่มี card_deposit
	//  สักแถวที่ id ตรงกับ card_play.id เลย — 0 จาก 2678)
	//
	// พอเพิ่ม guard RowsAffected == 0 -> error เข้าไปใน aff52a5 แถวนี้กลายเป็น
	// สมาชิกตัวแรกของ Withdraws ที่คืน error เสมอ ลูปจึง return ตั้งแต่รอบแรก
	// และรายการที่ต้องหักจริงไม่เคยถูกแตะ ผู้เรียกที่ v2/card_entity_controller.go
	// รันแบบ fire-and-forget หลังตอบ success ให้เครื่องไปแล้ว
	// ผลคือบัตร Package และ Time-play ทุกใบเล่นฟรีโดยไม่มีใครรู้
	//
	// ห้ามใส่กลับมา ถ้าต้องการบันทึกร่องรอยการใช้สิทธิ์ ต้องบันทึกที่ฝั่ง card_play
	// ไม่ใช่ยัดผ่านเส้นทางหักยอดของ card_deposit
	_, err := GetCardPlayTypeByCardPlayId(*cardPlayId)
	if err == nil {
		// เดิมกลืน error ของขั้นนี้เงียบ ๆ (if err == nil โดยไม่มี else)
		// ไม่เปลี่ยนให้ล้มทั้งรายการ เพราะตัวนับไม่ใช่ตัวเงิน แต่ต้องเห็นว่ามันไม่ขยับ
		if _, err := UpdateCardPlayTypeUsed(cardPlayId, cardNo); err != nil {
			fmt.Printf("card play %s: อัปเดต used_time ของ card_play_type ไม่สำเร็จ: %v\n",
				cardPlayId, err)
		}
	}

	// func DeductEcoin
	if input.ECoin > 0 {
		ecoin, err := DeductEcoin(input)
		if err != nil {
			return nil, fmt.Errorf("failed to deduct card play: %w", err)
		}

		// if ecoinSlice, ok := ecoin.([]struct {
		// 	DepositId  *uuid.UUID
		// 	RemoveCoin int64
		// }); ok {
		// 	fmt.Println("ecoinSlice ok :", ecoinSlice, ok)
		// 	fmt.Println("ecoinSlice:", ecoinSlice)
		// 	for _, data := range ecoinSlice {
		// 		Withdraws = append(Withdraws, Deposit{
		// 			Id:           data.DepositId,
		// 			BalanceCoin:  data.RemoveCoin,
		// 			BalanceBonus: 0,
		// 		})
		// 	}
		// }

		// for _, data := range ecoin {
		// 	Withdraws = append(Withdraws, Deposit{
		// 		Id:           data.DepositId,
		// 		BalanceCoin:  data.RemoveCoin,
		// 		BalanceBonus: 0,
		// 	})
		// }

		for _, data := range ecoin {
			Withdraws = append(Withdraws, Deposit{
				Id:                  data.DepositId,
				BalanceCoin:         data.RemoveCoin,
				BalanceBonus:        0,
				BalanceDiscountCash: 0,
			})
		}
		fmt.Println("withdraws 1:", Withdraws)
	}
	// func DeductBonus
	if input.EBonus > 0 {
		deposits, err := GetCardDepositBonusByCardNo(input.CardNo) //TODO: update ebonus for none member
		if err != nil {
			return nil, fmt.Errorf("failed to deduct card play: %w", err)
		}
		ebonus, err := DeductBonus(deposits, input)
		if err != nil {
			return nil, fmt.Errorf("failed to deduct card play: %w", err)
		}
		// uuid := uuid.MustParse("87fbd462-58a1-4752-8807-566b79955c1c")
		for _, data := range ebonus {
			idx := slices.IndexFunc(Withdraws, func(d Deposit) bool {
				return *d.Id == *data.CardDepositId
			})
			if idx != -1 {
				fmt.Println("found bonus: ", data.CardDepositId)
				Withdraws[idx].BalanceBonus += data.RemoveCoin
				continue
			} else {
				Withdraws = append(Withdraws, Deposit{
					Id:                  data.CardDepositId,
					BalanceCoin:         0,
					BalanceBonus:        data.RemoveCoin,
					BalanceDiscountCash: 0,
				})
			}
		}
	}
	//TODO deduct discount cash
	if input.DiscountCashAmount != nil && *input.DiscountCashAmount > 0 {
		removeDiscountCash, err := DeductDiscountCash(input)
		if err != nil {
			return nil, fmt.Errorf("failed to deduct discount cash: %w", err)
		}
		for _, discount := range *removeDiscountCash {
			idx := slices.IndexFunc(Withdraws, func(d Deposit) bool {
				return *d.Id == *discount.DepositId
			})
			if idx != -1 {
				fmt.Println("found discount cash: ", discount.DepositId)
				Withdraws[idx].BalanceDiscountCash += discount.DiscountCashAmount
				continue
			} else {
				Withdraws = append(Withdraws, Deposit{
					Id:                  discount.DepositId,
					BalanceCoin:         0,
					BalanceBonus:        0,
					BalanceDiscountCash: discount.DiscountCashAmount,
				})
			}
		}
	}

	for _, WithDraw := range Withdraws {
		_, err := CreateCardWithdraw(models.CardWithdrawDto{
			CardDepositId:      WithDraw.Id,
			FromChannel:        fromChannel,
			CardNo:             cardNo,
			MemberTel:          MemberTel,
			AmountEcoin:        int(WithDraw.BalanceCoin),
			AmountEbonus:       int(WithDraw.BalanceBonus),
			AmountDiscountCash: WithDraw.BalanceDiscountCash,
		}, userId)
		// เดิมการหักยอดอยู่ใน if err == nil โดยไม่มี else
		// ถ้า CreateCardWithdraw ล้มเหลว err จะถูกทิ้ง วนต่อ แล้ว return nil = สำเร็จ
		// ผลคือเครื่องปลดล็อกให้เล่นโดยที่ยอดในบัตรไม่ถูกหัก
		if err != nil {
			return nil, fmt.Errorf("failed to create card withdraw: %w", err)
		}

		_, err = UpdateCardDepositBalance(models.CardDepositBalanceDto{
			ID:                  WithDraw.Id,
			BalanceCoin:         int(WithDraw.BalanceCoin),
			BalanceBonus:        int(WithDraw.BalanceBonus),
			BalanceDiscountCash: &WithDraw.BalanceDiscountCash,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update card deposit balance: %w", err)
		}
	}
	fmt.Println("withdraws 2:", Withdraws)
	return Withdraws, nil
}

/*
* verify card entity
* parameter: card_no
* return: card entity struct or error
 */
func VerifyCardEntity(cardNo string) (*models.CardEntity, error) {
	cardEntity, err := FindCardEntityByCardNo(cardNo)
	if err != nil {
		return nil, fmt.Errorf("failed to verify card entity: %w", err)
	}
	return cardEntity, nil
}

func ValiddateCardPlay(CardNo string) ([]models.CardPlayVerifyResponse, error) {
	var entities []models.CardPlay
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database pos connection is nil")
	}
	result := config.DB_POS.Select("id,card_no, play_branch, play_machine, card_deposit_id").
		Where("card_no = ? and is_delete = false", CardNo).Order("play_branch DESC, play_machine DESC").Find(&entities)
	if result.Error != nil {
		fmt.Println("failed to find card play: %w", result.Error)
		return nil, fmt.Errorf("failed to find card play: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("card play not found for card_no: %s", CardNo)
	}
	var cardPlays []models.CardPlayVerifyResponse
	for _, cardPlay := range entities {
		branchList, _ := GetBranchListByCardNo(&cardPlay.ID)
		if branchList == nil {
			branchList = []models.BranchList{} // Initialize to empty slice if nil
		}

		machineList, _ := GetMachineListByCardPlayId(&cardPlay.ID)
		if machineList == nil {
			machineList = []models.MachineList{} // Initialize to empty slice if nil
		}
		CardplayType := make(chan any)
		go GetPeriod(&cardPlay.ID, CardplayType)
		CardplayTypeData := <-CardplayType // Assuming this will block until the period is received
		p := CardplayTypeData.(struct {
			StartDate  *time.Time
			EndDate    *time.Time
			LimitUsage int
			UsedTime   int
			IsDelete   bool
		})
		var startDate, endDate time.Time
		today := time.Now()
		startDate = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 1, 0, today.Location())
		endDate = time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, today.Location())
		if p.StartDate != nil {
			startDate = *p.StartDate
		}
		if p.EndDate != nil {
			endDate = *p.EndDate
		}
		periodData := models.Period{
			StartDate: startDate,
			EndDate:   endDate,
		}

		cardPlayID := cardPlay.ID // Convert UUID to stringcardPlay.ID
		orderID, _ := GetConditionID(p.LimitUsage, periodData.StartDate.IsZero(), cardPlay.PlayMachine, cardPlay.PlayBranch)
		cardPlays = append(cardPlays, models.CardPlayVerifyResponse{
			CardPlayId:    &cardPlayID,
			CardNo:        cardPlay.CardNo,
			PlayBranch:    cardPlay.PlayBranch,
			PlayMachine:   cardPlay.PlayMachine,
			BranchList:    &branchList,
			MachineList:   &machineList,
			Period:        &periodData,
			LimitUsage:    &p.LimitUsage,
			UsedTime:      &p.UsedTime,
			ConditionID:   *orderID, // Dereference pointer to assign int value
			CardDepositId: cardPlay.CardDepositId,
		})
	}
	cardPlays = SortCardPlayByConditionID(cardPlays)
	return cardPlays, nil
}

// sorting ASC by ConditionID
func SortCardPlayByConditionID(cardPlays []models.CardPlayVerifyResponse) []models.CardPlayVerifyResponse {
	sort.Slice(cardPlays, func(i, j int) bool {
		return cardPlays[i].ConditionID < cardPlays[j].ConditionID
	})
	return cardPlays
}

// get branch list by cardPlayId
func GetBranchListByCardNo(cardPlayId *uuid.UUID) ([]models.BranchList, error) {
	var branchList []models.BranchList
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database pos connection is nil")
	}

	result := config.DB_POS.Table("card_play_branch").
		Select("branch_code").
		Where("card_play_id = ?", cardPlayId).
		Find(&branchList)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get branch list: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("no branch found")
	}

	return branchList, nil
}

func GetPeriod(cardPlayId *uuid.UUID, CardplayType chan<- any) {
	// Assuming you will implement a function to get period
	// This is a placeholder function
	// You can use the cardPlayId to query the period from the database
	// and send the result to the period channel
	var p struct {
		StartDate  *time.Time
		EndDate    *time.Time
		LimitUsage int
		UsedTime   int
		IsDelete   bool
	}
	result := config.DB_POS.Table("card_play_type").
		Select("start_date, end_date, play_time as limit_usage,used_time,is_delete").
		Where("card_play_id = ?", cardPlayId).
		First(&p)
	if result.Error != nil {
		// fmt.Println("Error fetching period for card play ID:", cardPlayId, "Error:", result.Error)
		p.StartDate = nil
		p.EndDate = nil
		p.LimitUsage = 0  // Set default value for LimitUsage
		p.UsedTime = 0    // Set default value for UsageTime
		CardplayType <- p // Send nil to indicate an error

	}
	if result.RowsAffected == 0 {
		fmt.Println("No period found for card play ID:", cardPlayId)
		p.StartDate = nil
		p.EndDate = nil
		p.LimitUsage = 0  // Set default value for LimitUsage
		p.UsedTime = 0    // Set default value for UsageTime
		CardplayType <- p // Send nil to indicate an error
	}
	if result.RowsAffected > 0 { // Example limit usage
		if p.IsDelete {
			yesterday := time.Now().AddDate(0, 0, -1)
			startTime := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
			endTime := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
			p.StartDate = &startTime
			p.EndDate = &endTime
			p.LimitUsage = 0  // Set default value for LimitUsage
			p.UsedTime = 0    // Set default value for UsageTime
			CardplayType <- p // Send nil to indicate an error
		} else {
			fmt.Println("Period found for card play ID:", cardPlayId)
			CardplayType <- p
		}
	}
	close(CardplayType)
}

// ValiddateCardPlayOptimized performs card play validation with optimized database queries
func ValiddateCardPlayOptimized(CardNo string) ([]models.CardPlayVerifyResponse, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database pos connection is nil")
	}

	// Single optimized query with joins to get all required data
	var results []struct {
		// Card Play fields
		ID            uuid.UUID  `gorm:"column:id"`
		CardNo        string     `gorm:"column:card_no"`
		PlayBranch    bool       `gorm:"column:play_branch"`
		PlayMachine   bool       `gorm:"column:play_machine"`
		CardDepositId *uuid.UUID `gorm:"column:card_deposit_id"`

		// Card Play Type fields
		StartDate  *time.Time `gorm:"column:start_date"`
		EndDate    *time.Time `gorm:"column:end_date"`
		LimitUsage int        `gorm:"column:play_time"`
		UsedTime   int        `gorm:"column:used_time"`
	}

	// Optimized query with LEFT JOIN to get card play and card play type data in one query
	query := `
		SELECT 
			cp.id, cp.card_no, cp.play_branch, cp.play_machine, cp.card_deposit_id,
			cpt.start_date, cpt.end_date, cpt.play_time, cpt.used_time
		FROM card_play cp
		LEFT JOIN card_play_type cpt ON cp.id = cpt.card_play_id AND cpt.is_delete = false
		WHERE cp.card_no = ? AND cp.is_delete = false
		ORDER BY cp.play_branch DESC, cp.play_machine DESC
	`

	if err := config.DB_POS.Raw(query, CardNo).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to find card play: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("card play not found for card_no: %s", CardNo)
	}

	// Get all branch lists in a single query
	branchMap, err := getBranchListMapOptimized(results)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch lists: %w", err)
	}

	// Build response array
	cardPlays := make([]models.CardPlayVerifyResponse, 0, len(results))
	for _, result := range results {
		// Get branch list for this card play
		branchList := branchMap[result.ID]
		if branchList == nil {
			branchList = []models.BranchList{}
		}

		// Build period data
		var startDate, endDate time.Time
		if result.StartDate != nil {
			startDate = *result.StartDate
		}
		if result.EndDate != nil {
			endDate = *result.EndDate
		}
		periodData := models.Period{
			StartDate: startDate,
			EndDate:   endDate,
		}

		// Get condition ID
		orderID, _ := GetConditionID(result.LimitUsage, periodData.StartDate.IsZero(), result.PlayMachine, result.PlayBranch)

		cardPlays = append(cardPlays, models.CardPlayVerifyResponse{
			CardPlayId:    &result.ID,
			CardNo:        result.CardNo,
			PlayBranch:    result.PlayBranch,
			PlayMachine:   result.PlayMachine,
			BranchList:    &branchList,
			MachineList:   nil,
			Period:        &periodData,
			LimitUsage:    &result.LimitUsage,
			UsedTime:      &result.UsedTime,
			ConditionID:   *orderID,
			CardDepositId: result.CardDepositId,
		})
	}

	return SortCardPlayByConditionID(cardPlays), nil
}

// getBranchListMapOptimized fetches all branch lists in a single query
func getBranchListMapOptimized(cardPlays []struct {
	ID            uuid.UUID  `gorm:"column:id"`
	CardNo        string     `gorm:"column:card_no"`
	PlayBranch    bool       `gorm:"column:play_branch"`
	PlayMachine   bool       `gorm:"column:play_machine"`
	CardDepositId *uuid.UUID `gorm:"column:card_deposit_id"`
	StartDate     *time.Time `gorm:"column:start_date"`
	EndDate       *time.Time `gorm:"column:end_date"`
	LimitUsage    int        `gorm:"column:play_time"`
	UsedTime      int        `gorm:"column:used_time"`
}) (map[uuid.UUID][]models.BranchList, error) {
	if len(cardPlays) == 0 {
		return make(map[uuid.UUID][]models.BranchList), nil
	}

	// Collect all card play IDs
	cardPlayIDs := make([]uuid.UUID, len(cardPlays))
	for i, cp := range cardPlays {
		cardPlayIDs[i] = cp.ID
	}

	// Single query to get all branch lists
	var branchResults []struct {
		CardPlayID uuid.UUID `gorm:"column:card_play_id"`
		BranchCode string    `gorm:"column:branch_code"`
	}

	if err := config.DB_POS.Table("card_play_branch").
		Select("card_play_id, branch_code").
		Where("card_play_id IN ?", cardPlayIDs).
		Find(&branchResults).Error; err != nil {
		return nil, err
	}

	// Build map of card play ID to branch lists
	branchMap := make(map[uuid.UUID][]models.BranchList)
	for _, br := range branchResults {
		branchMap[br.CardPlayID] = append(branchMap[br.CardPlayID], models.BranchList{
			BranchCode: br.BranchCode,
		})
	}

	return branchMap, nil
}

func GetConditionID(play_time int, period, machine, branch bool) (*int, error) {
	var id int
	if !period && play_time > 0 && branch && machine {
		id = 1
	} else if !period && play_time > 0 && branch && !machine {
		id = 2
	} else if !period && play_time > 0 && !branch && !machine {
		id = 3
	} else if period && play_time == 0 && branch && machine {
		id = 4
	} else if period && play_time == 0 && branch && !machine {
		id = 5
	} else if period && play_time == 0 && !branch && machine {
		id = 6
	} else {
		id = 7 // Default case if none of the conditions match
	}
	// fmt.Printf("GetConditionID called with play_time: %d, period: %t, machine: %t, branch: %t result: %d\n", play_time, period, machine, branch, id)

	return &id, nil
}

func SumCoin(cardPlays []models.CardPlayVerifyResponse, coin chan<- int64) {
	var sum int64 = 0
	var CardDepositIds []*uuid.UUID
	for _, cardPlay := range cardPlays {
		if cardPlay.CardDepositId != nil {
			CardDepositIds = append(CardDepositIds, cardPlay.CardDepositId)
		}
	}
	if len(CardDepositIds) > 0 {
		var cardDeposits []models.CardDeposit
		if config.DB_POS == nil {
			return
		}
		//TODO: card_deposit SET IS_ACTIVE = TRUE
		result := config.DB_POS.Table("card_deposit").
			Select("balance_coin AS balance_coin").
			Where("id IN (?) AND is_active = true", CardDepositIds).
			Find(&cardDeposits)
		if result.Error != nil {
			fmt.Println("Error fetching card deposits:", result.Error)
			coin <- sum // Send the sum (which is 0) to the channel
			return
		}
		if result.RowsAffected > 0 {
			for _, cardDeposit := range cardDeposits {
				sum += int64(cardDeposit.BalanceCoin)
			}
		}

	}
	coin <- sum
	close(coin)

}

func SumBonus(cardPlays []models.CardPlayVerifyResponse, bonus chan<- int64) {
	var sum int64 = 0
	var cardEntity *models.CardEntity
	var cardDeposit models.CardDeposit
	cardNo := cardPlays[0].CardNo
	cardEntity, err := FindActiveCardNo(cardNo)
	if err != nil {
		sum = 0
	}
	mobile := cardEntity.MemberTel
	if mobile == "0000000000" || mobile == "" {
		sum = 0
	} else {
		if config.DB_POS == nil {
			sum = 0
		}
		//TODO: card_deposit SET IS_ACTIVE = TRUE
		result := config.DB_POS.Table("card_deposit").
			Select("sum(balance_bonus) AS balance_bonus").
			Where("card_no = ? AND is_active = true", cardNo).
			Find(&cardDeposit)
		if result.Error != nil {
			sum = 0
		}
		if result.RowsAffected > 0 {
			sum = int64(cardDeposit.BalanceBonus)
		}
	}
	bonus <- sum
	defer close(bonus)
}

type Deposit struct {
	Id                  *uuid.UUID `json:"id"`
	BalanceCoin         int64      `json:"balance_coin"`
	BalanceBonus        int64      `json:"balance_bonus"`
	BalanceDiscountCash float32    `json:"balance_discount_cash"`
}

// find card deposit by card play id
func DeductEcoin(input models.CardPlayDeductDto) ([]struct {
	DepositId  *uuid.UUID
	RemoveCoin int64
}, error) {
	var RetCardPlay struct {
		CardDepositId *uuid.UUID `json:"card_deposit_id"`
		CardNo        string     `json:"card_no"`
	}
	var Withdraws []struct {
		DepositId  *uuid.UUID
		RemoveCoin int64
	}

	resultCardPlay := config.DB_POS.Table("card_play").
		Select("card_deposit_id,card_no").
		Where("id = ? and is_delete = false", input.CardPlayId).
		Scan(&RetCardPlay)
	if resultCardPlay.Error != nil {
		return nil, fmt.Errorf("failed to find card play: %w", resultCardPlay.Error)
	}
	if resultCardPlay.RowsAffected == 0 {
		return nil, fmt.Errorf("card play not found for card_play_id: %s", input.CardPlayId)
	}
	if resultCardPlay.RowsAffected > 0 {
		var deposits []Deposit
		//TODO: card_deposit SET IS_ACTIVE = TRUE
		resultDeposit := config.DB_POS.Table("card_deposit").
			Select("id, balance_coin, balance_bonus").
			Where("id = ? AND is_active = true", RetCardPlay.CardDepositId).
			First(&deposits)
		if resultDeposit.Error != nil {
			return nil, fmt.Errorf("failed to find card deposit: %w", resultDeposit.Error)

		}
		deposit := deposits[0]
		fmt.Printf("id : %v, balance_coin: %d, balance_bonus: %d\n", deposit.Id, deposit.BalanceCoin, deposit.BalanceBonus)
		balanceCoin := deposit.BalanceCoin
		if balanceCoin >= int64(input.ECoin) {
			Withdraws = append(Withdraws, struct {
				DepositId  *uuid.UUID
				RemoveCoin int64
			}{
				DepositId:  deposit.Id,
				RemoveCoin: int64(input.ECoin),
			})
			return Withdraws, nil
		} else {
			deposits, err := GetCardDepositByCardNo(RetCardPlay.CardNo)
			if err != nil {
				return nil, fmt.Errorf("failed to get card deposit by card_no: %w", err)
			}
			if deposits != nil {
				fmt.Printf("get deposit by card_no: %v\n", deposits)
				// var tmpBalanceCoin int64 = int64(input.ECoin)

				typedDeposits, ok := deposits.([]struct {
					Id           *uuid.UUID `json:"id"`
					BalanceCoin  int64      `json:"balance_coin"`
					BalanceBonus int64      `json:"balance_bonus"`
				})
				if !ok {
					return nil, fmt.Errorf("failed to type assert deposits")
				}

				tmpBalanceCoin := int64(input.ECoin)
				for _, deposit := range typedDeposits {
					fmt.Println("balance_coin", deposit.BalanceCoin)
					if tmpBalanceCoin > 0 {
						if deposit.BalanceCoin > tmpBalanceCoin {
							Withdraws = append(Withdraws, struct {
								DepositId  *uuid.UUID
								RemoveCoin int64
							}{
								DepositId:  deposit.Id,
								RemoveCoin: tmpBalanceCoin,
							})
							tmpBalanceCoin = 0
							break

						} else {
							Withdraws = append(Withdraws, struct {
								DepositId  *uuid.UUID
								RemoveCoin int64
							}{
								DepositId:  deposit.Id,
								RemoveCoin: deposit.BalanceCoin,
							})
							tmpBalanceCoin -= deposit.BalanceCoin
						}
					}
				}
				return Withdraws, nil
			}
		}

	}
	return nil, fmt.Errorf("deduct coin must be greater than 0")
}

// get card_deposit by card_no
func GetCardDepositByCardNo(cardNo string) (any, error) {

	var deposits []struct {
		Id           *uuid.UUID `json:"id"`
		BalanceCoin  int64      `json:"balance_coin"`
		BalanceBonus int64      `json:"balance_bonus"`
	}
	//TODO: card_deposit SET IS_ACTIVE = TRUE
	result := config.DB_POS.Raw("SELECT d.id, d.balance_coin, d.balance_bonus FROM card_play p join card_deposit d on p.card_deposit_id = d.id WHERE p.card_no = ? AND p.is_delete = false AND d.is_active = true", cardNo).Find(&deposits)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find card deposit by card_no: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		sort.Slice(deposits, func(i, j int) bool {
			return deposits[i].BalanceCoin > deposits[j].BalanceCoin
		})
		return deposits, nil
	}

	return nil, nil
}

func DeductBonus(deposits []Deposit, input models.CardPlayDeductDto) ([]models.WithdrawUsed, error) {
	useBonus := int64(input.EBonus)
	var Withdraws []models.WithdrawUsed

	for _, deposit := range deposits {
		if useBonus > 0 {
			if deposit.BalanceBonus >= useBonus {
				Withdraws = append(Withdraws, models.WithdrawUsed{
					CardDepositId: deposit.Id,
					RemoveCoin:    useBonus,
				})
				useBonus = 0
				break
			} else {
				Withdraws = append(Withdraws, models.WithdrawUsed{
					CardDepositId: deposit.Id,
					RemoveCoin:    deposit.BalanceBonus,
				})
				useBonus -= deposit.BalanceBonus
			}
		}
	}
	if useBonus > 0 {
		return nil, fmt.Errorf("not enough bonus")
	}

	return Withdraws, nil

}

func GetCardDepositBonusByCardNo(cardNo string) ([]Deposit, error) {
	var cardEntity *models.CardEntity
	var deposits []Deposit
	cardEntity, err := FindActiveCardNo(cardNo)
	if err != nil {
		return nil, fmt.Errorf("failed to find card entity: %w", err)
	}
	if cardEntity.MemberTel == "0000000000" || cardEntity.MemberTel == "" {
		//TODO: ebonus for none member
		result := config.DB_POS.Table("card_deposit").
			Select("id, balance_coin, balance_bonus").
			Where("card_no = ? and balance_bonus > 0 and is_active = true and bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')", cardEntity.CardNo).
			Order("bonus_expire_date ASC").
			Find(&deposits)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to find card deposit: %w", result.Error)
		}
		return deposits, nil
	}

	//TODO: card_deposit SET IS_ACTIVE = TRUE
	result := config.DB_POS.Table("card_deposit").
		Select("id, balance_coin, balance_bonus").
		Where("card_no = ? and balance_bonus > 0 and is_active = true and bonus_expire_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')", cardEntity.CardNo).
		Order("bonus_expire_date ASC").
		Find(&deposits)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find card deposit: %w", result.Error)
	}
	// fmt.Printf("Member tel : %v \r\n",cardEntity.MemberTel)
	// fmt.Printf("deposits : %v \r\n",deposits)
	return deposits, nil
}

// get mobile_tel
func GetMemberTel(cardNo string, mobileTel chan<- string) {
	var cardEntity *models.CardEntity
	cardEntity, err := FindActiveCardNo(cardNo)
	if err != nil {
		mobileTel <- "0000000000"
	} else {
		mobileTel <- cardEntity.MemberTel
	}
	close(mobileTel)
}

// update card play is_delete by card deposit id
func UpdateCardPlayIsDelete(cardDepositId *uuid.UUID) error {
	if config.DB_POS == nil {
		return fmt.Errorf("database pos connection is nil")
	}

	var cardPlay models.CardPlay
	return config.DB_POS.Model(&cardPlay).
		Where("card_deposit_id = ?", cardDepositId).
		Updates(map[string]interface{}{
			"is_delete":   true,
			"delete_date": utils.TimeNowAsia(),
		}).Error
}

// update card play type is_delete by card play id
func UpdateCardPlayTypeIsDelete(cardPlayId *uuid.UUID) error {
	if config.DB_POS == nil {
		return fmt.Errorf("database pos connection is nil")
	}

	var cardPlayType models.CardPlayType
	return config.DB_POS.Model(&cardPlayType).
		Where("card_play_id = ?", cardPlayId).
		Updates(map[string]interface{}{
			"is_delete":   true,
			"delete_date": utils.TimeNowAsia(),
		}).Error
}

// VerifyCardPlayOptimized performs card play verification and fetches all required data in parallel
func VerifyCardPlayOptimized(input models.CardPlayVerifyDto, userId int) (*uuid.UUID, *models.CardDetailDto, *models.CardPlayType, error) {
	type result struct {
		cardPlayID   *uuid.UUID
		cardInfo     *models.CardDetailDto
		cardPlayType *models.CardPlayType
		err          error
	}

	// Channel to collect results
	resultChan := make(chan result, 3)

	// Goroutine 1: Verify card play
	go func() {
		cardPlayID, err := VerifyCardPlay(input, userId)
		if err != nil {
			resultChan <- result{err: fmt.Errorf("verify card play: %w", err)}
			return
		}
		resultChan <- result{cardPlayID: cardPlayID}
	}()

	// Goroutine 2: Get card info
	go func() {
		cardInfo, err := CheckCardInfoWithExpire(input.CardNo)
		if err != nil {
			resultChan <- result{err: fmt.Errorf("check card info: %w", err)}
			return
		}
		resultChan <- result{cardInfo: cardInfo}
	}()

	// Wait for card play verification first to get the ID
	var cardPlayID *uuid.UUID
	var cardInfo *models.CardDetailDto
	// var cardPlayType *models.CardPlayType
	var errors []error

	for i := 0; i < 2; i++ {
		res := <-resultChan
		if res.err != nil {
			errors = append(errors, res.err)
			continue
		}
		if res.cardPlayID != nil {
			cardPlayID = res.cardPlayID
		}
		if res.cardInfo != nil {
			cardInfo = res.cardInfo
		}
	}

	if len(errors) > 0 {
		return nil, nil, nil, errors[0]
	}

	if cardPlayID == nil {
		return nil, nil, nil, fmt.Errorf("card play verification failed")
	}

	// // Goroutine 3: Get card play type (depends on cardPlayID)
	// go func() {
	// 	cardPlayType, err := GetCardPlayTypeByCardPlayId(*cardPlayID)
	// 	if err != nil {
	// 		// resultChan <- result{err: fmt.Errorf("get card play type: %w", err)}
	// 		resultChan <- result{cardPlayType: nil}
	// 		return
	// 	}
	// 	resultChan <- result{cardPlayType: &cardPlayType}
	// }()

	// // Wait for card play type result
	// res := <-resultChan
	// if res.err != nil {
	// 	return nil, nil, nil, res.err
	// }
	// cardPlayType = res.cardPlayType
	// fmt.Println("cardPlayType 3", cardPlayType)

	return cardPlayID, cardInfo, nil, nil
}

// TODO deduct discount cash
func DeductDiscountCash(input models.CardPlayDeductDto) (*[]models.DiscountCashDtoDeduct, error) {

	var cardEntity *models.CardEntity
	var removeDiscountCash []models.DiscountCashDtoDeduct
	cardEntity, err := FindActiveCardNo(input.CardNo)
	if err != nil {
		return nil, fmt.Errorf("failed to find card entity: %w", err)
	}
	if cardEntity.MemberTel == "0000000000" || cardEntity.MemberTel == "" {
		return nil, fmt.Errorf("not found member tel")
	}
	discountCash, err := GetDiscountCashByMemberTel(cardEntity.MemberTel)
	if err != nil {
		return nil, fmt.Errorf("failed to get discount cash: %w", err)
	}
	if discountCash.BalanceDiscountCash < *input.DiscountCashAmount {
		return nil, fmt.Errorf("not enough discount cash")
	}
	discounts := discountCash.Discounts
	tmpDiscountAmount := *input.DiscountCashAmount
	for _, discount := range *discounts {
		if tmpDiscountAmount > 0 {
			if discount.BalanceDiscountCash > tmpDiscountAmount {
				removeDiscountCash = append(removeDiscountCash, models.DiscountCashDtoDeduct{
					DepositId:          discount.DepositId,
					DiscountCashAmount: tmpDiscountAmount,
				})
				tmpDiscountAmount -= tmpDiscountAmount
			} else {
				tmpDiscountAmount -= discount.BalanceDiscountCash
				removeDiscountCash = append(removeDiscountCash, models.DiscountCashDtoDeduct{
					DepositId:          discount.DepositId,
					DiscountCashAmount: discount.BalanceDiscountCash,
				})
				discount.BalanceDiscountCash = 0
			}
		}
	}
	fmt.Println("removeDiscountCash", removeDiscountCash)
	return &removeDiscountCash, nil
}

// VerifyCardPlayV2 performs optimized card play verification for V2 API
// Accepts pre-fetched card details to avoid redundant queries
func VerifyCardPlayV2(cardNo string, cardDetails *models.CardDetailDto, input models.CardPlayVerifyDto, deductCondition string) (cardPlayId *uuid.UUID, balanceEcoin *int, balanceEbonus *int, balanceETimes *int, err error, machineDeduct *model_v2.CardPlayMachineDeductDto) {
	if config.DB_POS == nil {
		return nil, nil, nil, nil, fmt.Errorf("database pos connection is nil"), nil
	}

	// Validate balance requirements upfront (SKIPPED)
	// if input.UseCoin != nil && *input.UseCoin > 0 {
	// 	if cardDetails.ECoin < *input.UseCoin {
	// 		return nil, fmt.Errorf("not enough coin balance")
	// 	}
	// }
	// if input.UseBonus != nil && *input.UseBonus > 0 {
	// 	if cardDetails.EBonus < *input.UseBonus {
	// 		return nil, fmt.Errorf("not enough bonus balance")
	// 	}
	// }

	// Parse input parameters
	var playBranchFilter string
	var playMachineFilter string
	if input.PlayBranch != "" {
		playBranchFilter = input.PlayBranch
	}
	if input.PlayMachine != "" && input.PlayMachine != "0" {
		playMachineFilter = input.PlayMachine
	}

	// Single optimized query with all validation logic in SQL
	query := `
	WITH valid_plays AS (
		SELECT 
			cp.id as card_play_id,
			cp.play_branch,
			cp.play_machine,
			cp.card_deposit_id,
			cpt.start_date,
			cpt.end_date,
			cpt.play_time as limit_usage,
			cpt.used_time,
			CASE 
				WHEN cp.play_branch = true THEN (
					SELECT COUNT(*) 
					FROM card_play_branch cpb 
					WHERE cpb.card_play_id = cp.id 
					AND cpb.branch_code = ?
				)
				ELSE 1
			END as branch_match,
			CASE 
				WHEN cp.play_machine = true THEN (
					SELECT COUNT(*) 
					FROM card_play_machine cpm 
					WHERE cpm.card_play_id = cp.id 
					AND cpm.machine_id = ?
				)
				ELSE 1
			END as machine_match,
			CASE 
				WHEN cpt.start_date IS NOT NULL AND cpt.end_date IS NOT NULL THEN
					CASE 
						WHEN (NOW() AT TIME ZONE 'Asia/Bangkok') BETWEEN cpt.start_date AND cpt.end_date THEN 1
						ELSE 0
					END
				ELSE 1
			END as period_valid,
			CASE 
				WHEN cpt.play_time IS NULL OR cpt.play_time = 0 THEN 1
				WHEN cpt.play_time > 0 AND cpt.used_time >= cpt.play_time THEN 0
				ELSE 1
			END as usage_valid,
			CASE
				WHEN cpt.play_time > 0 THEN
					CASE
						WHEN cp.play_branch = true AND cp.play_machine = true THEN 1
						WHEN cp.play_branch = true AND cp.play_machine = false THEN 2
						ELSE 3
					END
				WHEN cpt.play_time = 0 AND (cpt.start_date IS NOT NULL AND cpt.start_date != '0001-01-01') THEN
					CASE
						WHEN cp.play_branch = true AND cp.play_machine = true THEN 4
						WHEN cp.play_branch = true AND cp.play_machine = false THEN 5
						WHEN cp.play_branch = false AND cp.play_machine = true THEN 6
						ELSE 7
					END
				ELSE 7
			END as condition_id
		FROM card_play cp
		LEFT JOIN card_play_type cpt ON cp.id = cpt.card_play_id AND cpt.is_delete = false
		WHERE cp.card_no = ?
		  AND cp.is_delete = false
		  AND EXISTS (
			  SELECT 1 FROM card_deposit cd 
			  WHERE cd.card_no = cp.card_no 
			  AND cd.is_active = true 
			  AND cd.pos_menu_id >= 0
		  )
	)
	SELECT card_play_id, condition_id , period_valid, usage_valid , start_date, end_date, limit_usage,used_time
	FROM valid_plays
	WHERE branch_match > 0
	  AND machine_match > 0
	  AND period_valid = 1
	  AND usage_valid = 1
	ORDER BY condition_id ASC
	LIMIT 1
	`

	var result struct {
		CardPlayID  string     `gorm:"column:card_play_id"`
		ConditionID int        `gorm:"column:condition_id"`
		PeriodValid int        `gorm:"column:period_valid"`
		UsageValid  int        `gorm:"column:usage_valid"`
		StartDate   *time.Time `gorm:"column:start_date"`
		EndDate     *time.Time `gorm:"column:end_date"`
		LimitUsage  int        `gorm:"column:limit_usage"`
		UsedTime    int        `gorm:"column:used_time"`
	}

	err = config.DB_POS.Raw(query, playBranchFilter, playMachineFilter, cardNo).Scan(&result).Error
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to verify card play: %w", err), nil
	}

	// fmt.Printf("VerifyCardPlayV2 result: CardPlayID=%s, ConditionID=%d, PeriodValid=%d, UsageValid=%d\n", result.CardPlayID, result.ConditionID, result.PeriodValid, result.UsageValid)
	//  fmt.Printf("result %v",result)

	// Check if we got a result
	if result.CardPlayID == "" {
		return nil, nil, nil, nil, fmt.Errorf("card play is not valid"), nil
	}

	// Parse the UUID string
	cardPlayID, err := uuid.Parse(result.CardPlayID)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to parse card play ID: %w", err), nil
	}
	// deduct usage time if applicable
	var remainTimes int
	if result.StartDate != nil && result.EndDate != nil && result.LimitUsage > 0 {
		// TODO: update used time
		remainTimes = result.LimitUsage - result.UsedTime - 1
		fmt.Printf("remainTimes %d", remainTimes)
	}
	// --- Machine-level override check (Package card type only) ---
	// "Package" cards always deduct from the machine's own budget (card_play_machine),
	// never from the global card balance. The machine row columns determine which
	// deduction mode applies: play_time takes priority, then ecoin/ebonus.
	if playMachineFilter != "" && strings.ToLower(cardDetails.CardType) == "package" {
		var machineRow struct {
			ECoin    *int32 `gorm:"column:ecoin"`
			EBonus   *int32 `gorm:"column:ebonus"`
			PlayTime *int32 `gorm:"column:play_time"`
		}
		machineQuery := `
			SELECT cpm.ecoin, cpm.ebonus, cpm.play_time
			FROM card_play_machine cpm
			WHERE cpm.card_play_id = ?
			  AND cpm.machine_id = ?
			  AND cpm.is_delete = false
			LIMIT 1
		`
		if err = config.DB_POS.Raw(machineQuery, cardPlayID, playMachineFilter).Scan(&machineRow).Error; err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to query card play machine: %w", err), nil
		}

		machineHasEcoin := machineRow.ECoin != nil && *machineRow.ECoin > 0
		machineHasEbonus := machineRow.EBonus != nil && *machineRow.EBonus > 0
		machineHasPlayTime := machineRow.PlayTime != nil && *machineRow.PlayTime > 0

		// helper: convert int → *int32 for DeductCardPlayMachine
		int32Ptr := func(v int) *int32 { i := int32(v); return &i }

		var machineIdInt32 int32
		if parsed, parseErr := strconv.ParseInt(playMachineFilter, 10, 32); parseErr == nil {
			machineIdInt32 = int32(parsed)
		}

		var mEcoin, mEbonus, mPlayTime int
		if machineRow.ECoin != nil {
			mEcoin = int(*machineRow.ECoin)
		}
		if machineRow.EBonus != nil {
			mEbonus = int(*machineRow.EBonus)
		}
		if machineRow.PlayTime != nil {
			mPlayTime = int(*machineRow.PlayTime)
		}
		beforeEcoin := mEcoin
		beforeEbonus := mEbonus
		beforePlayTime := mPlayTime

		// play_time takes priority — deduct_condition is ignored in this mode
		if machineHasPlayTime {
			mPlayTime -= 1
			if mPlayTime < 0 {
				return nil, nil, nil, nil, fmt.Errorf("machine: no play time remaining"), nil
			}
			if err = DeductCardPlayMachine(cardPlayID, machineIdInt32, nil, nil, int32Ptr(mPlayTime)); err != nil {
				return nil, nil, nil, nil, err, nil
			}
			// Machine deduction complete — do NOT deduct from global card balance

			return &cardPlayID, nil, nil, nil, nil, &model_v2.CardPlayMachineDeductDto{
				BeforeEcoin:  &beforeEcoin,
				BeforeEbonus: &beforeEbonus,
				AfterEcoin:   &mEcoin,
				AfterEbonus:  &mEbonus,
				BeforeETimes: &beforePlayTime,
				AfterETimes:  &mPlayTime,
			}
		}

		// play_time not active — deduct from machine ecoin/ebonus per deduct_condition
		if machineHasEcoin || machineHasEbonus {
			deductPrice := 0
			if input.UseCoin != nil {
				deductPrice = *input.UseCoin
			}
			switch deductCondition {
			case "ecoin_first":
				mEcoin -= deductPrice
				if mEcoin < 0 {
					mEbonus += mEcoin
					mEcoin = 0
					if mEbonus < 0 {
						return nil, nil, nil, nil, fmt.Errorf("machine: not enough ecoin and ebonus"), nil
					}
				}
				if err = DeductCardPlayMachine(cardPlayID, machineIdInt32, int32Ptr(mEcoin), int32Ptr(mEbonus), nil); err != nil {
					return nil, nil, nil, nil, err, nil
				}
				return &cardPlayID, nil, nil, nil, nil, &model_v2.CardPlayMachineDeductDto{
					BeforeEcoin:  &beforeEcoin,
					BeforeEbonus: &beforeEbonus,
					AfterEcoin:   &mEcoin,
					AfterEbonus:  &mEbonus,
					BeforeETimes: &beforePlayTime,
					AfterETimes:  &mPlayTime,
				}
			case "ecoin_only":
				mEcoin -= deductPrice
				if mEcoin < 0 {
					return nil, nil, nil, nil, fmt.Errorf("machine: not enough ecoin"), nil
				}
				if err = DeductCardPlayMachine(cardPlayID, machineIdInt32, int32Ptr(mEcoin), nil, nil); err != nil {
					return nil, nil, nil, nil, err, nil
				}
				return &cardPlayID, nil, nil, nil, nil, &model_v2.CardPlayMachineDeductDto{
					BeforeEcoin:  &beforeEcoin,
					BeforeEbonus: &beforeEbonus,
					AfterEcoin:   &mEcoin,
					AfterEbonus:  &mEbonus,
					BeforeETimes: &beforePlayTime,
					AfterETimes:  &mPlayTime,
				}
			case "ebonus_first":
				mEbonus -= deductPrice
				if mEbonus < 0 {
					mEcoin += mEbonus
					mEbonus = 0
					if mEcoin < 0 {
						return nil, nil, nil, nil, fmt.Errorf("machine: not enough ecoin and ebonus"), nil
					}
				}
				if err = DeductCardPlayMachine(cardPlayID, machineIdInt32, int32Ptr(mEcoin), int32Ptr(mEbonus), nil); err != nil {
					return nil, nil, nil, nil, err, nil
				}
				return &cardPlayID, nil, nil, nil, nil, &model_v2.CardPlayMachineDeductDto{
					BeforeEcoin:  &beforeEcoin,
					BeforeEbonus: &beforeEbonus,
					AfterEcoin:   &mEcoin,
					AfterEbonus:  &mEbonus,
					BeforeETimes: &beforePlayTime,
					AfterETimes:  &mPlayTime,
				}
			case "ebonus_only":
				mEbonus -= deductPrice
				if mEbonus < 0 {
					return nil, nil, nil, nil, fmt.Errorf("machine: not enough ebonus"), nil
				}
				if err = DeductCardPlayMachine(cardPlayID, machineIdInt32, nil, int32Ptr(mEbonus), nil); err != nil {
					return nil, nil, nil, nil, err, nil
				}
				return &cardPlayID, nil, nil, nil, nil, &model_v2.CardPlayMachineDeductDto{
					BeforeEcoin:  &beforeEcoin,
					BeforeEbonus: &beforeEbonus,
					AfterEcoin:   &mEcoin,
					AfterEbonus:  &mEbonus,
					BeforeETimes: &beforePlayTime,
					AfterETimes:  &mPlayTime,
				}
			default:
				return nil, nil, nil, nil, fmt.Errorf("invalid deduct condition: %s", deductCondition), nil
			}
			//TODO: add meter record here for ecoin/ebonus deduction
		}

		// Package card reached here but machine has nothing configured — block the play
		return nil, nil, nil, nil, fmt.Errorf("machine: not enough ecoin/ebonus or no play time"), nil
	}
	// --- End machine-level override check ---

	// fmt.Printf("card type %s \\r\\n",cardDetails.CardType)
	// validate ecoin balance , ebonus balance
	if result.ConditionID >= 1 && result.ConditionID <= 7 && strings.ToLower(cardDetails.CardType) == "normal" {
		deductPrice := input.UseCoin
		balanceEcoin := cardDetails.ECoin
		balanceEbonus := cardDetails.EBonus
		switch deductCondition {
		case "ecoin_first":
			balanceEcoin -= *deductPrice
			if balanceEcoin < 0 {
				balanceEbonus += balanceEcoin
				balanceEcoin = 0
				if balanceEbonus < 0 {
					return nil, nil, nil, nil, fmt.Errorf("not enough ecoin and ebonus balance"), nil
				}
			}
			//TODO: update balance ecoin , ebonus
			return &cardPlayID, &balanceEcoin, &balanceEbonus, nil, nil, nil
		case "ecoin_only":
			balanceEcoin -= *deductPrice
			if balanceEcoin < 0 {
				return nil, nil, nil, nil, fmt.Errorf("not enough ecoin balance"), nil
			}
			//TODO: update balance ecoin
			return &cardPlayID, &balanceEcoin, nil, nil, nil, nil
		case "ebonus_first":
			balanceEbonus -= *deductPrice
			if balanceEbonus < 0 {
				balanceEcoin += balanceEbonus
				balanceEbonus = 0
				if balanceEcoin < 0 {
					return nil, nil, nil, nil, fmt.Errorf("not enough ecoin and ebonus balance"), nil
				}
			}
			//TODO: update balance ecoin , ebonus
			return &cardPlayID, &balanceEcoin, &balanceEbonus, nil, nil, nil
		case "ebonus_only":
			balanceEbonus -= *deductPrice
			if balanceEbonus < 0 {
				return nil, nil, nil, nil, fmt.Errorf("not enough ebonus balance"), nil
			}
			//TODO: update balance ebonus
			return &cardPlayID, nil, &balanceEbonus, nil, nil, nil
		default:
			return nil, nil, nil, nil, fmt.Errorf("invalid deduct condition: %s", deductCondition), nil
		}
	}

	return &cardPlayID, nil, nil, &remainTimes, nil, nil
}

func DeleteCardPlayByCardNo(cardNo string) error {
	if config.DB_POS == nil {
		return fmt.Errorf("database pos connection is nil")
	}

	var cardPlay models.CardPlay
	return config.DB_POS.Model(&cardPlay).
		Where("LOWER(card_no) = ?", strings.ToLower(cardNo)).
		Updates(map[string]interface{}{
			"is_delete":   true,
			"delete_date": utils.TimeNowAsia(),
		}).Error
}
