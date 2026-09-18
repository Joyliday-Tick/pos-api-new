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

func CreateMember(input models.MemberDto, userId int) (models.Member, error) {
	var entity models.Member

	// Map fields จาก DTO ไปยัง member
	entity.Tel = input.Tel
	entity.MName = input.MName
	entity.SName = input.SName
	entity.TotalPoint = input.TotalPoint
	entity.Bonus = input.Bonus
	entity.Ecoin = input.Ecoin
	entity.Finwow = input.Finwow
	entity.Estamp = input.Estamp
	entity.MSkill1 = input.MSkill1
	entity.MSkill2 = input.MSkill2
	entity.MSkill3 = input.MSkill3
	entity.MSkill4 = input.MSkill4
	entity.MSkill5 = input.MSkill5
	entity.JubuJibi = input.JubuJibi
	entity.IsActive = input.IsActive
	entity.CreateDate = *utils.TimeNowAsia()

	if config.DB_POS == nil {
		return entity, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to create member: %w", err)
	}

	return entity, nil
}

func UpdateMember(id int, input models.MemberDto, userId int) (models.Member, error) {
	var existing models.Member

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("member not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"tel":         input.Tel,
		"m_name":      input.MName,
		"s_name":      input.SName,
		"total_point": input.TotalPoint,
		"bonus":       input.Bonus,
		"ecoin":       input.Ecoin,
		"finwow":      input.Finwow,
		"estamp":      input.Estamp,
		"mskill1":     input.MSkill1,
		"mskill2":     input.MSkill2,
		"mskill3":     input.MSkill3,
		"mskill4":     input.MSkill4,
		"mskill5":     input.MSkill5,
		"jubu_jibi":   input.JubuJibi,
		"is_active":   input.IsActive,
		"update_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.Member{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update member: %w", err)
	}

	// ดึงข้อมูลใหม่โดยใช้ Where ด้วย id
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated member: %w", err)
	}

	return existing, nil
}

func FindExistMemberTel(tel string) (*models.Member, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var member models.Member

	err := config.DB_POS.
		Where("tel = ?", tel).
		First(&member).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ไม่พบข้อมูล => คืน nil, nil
			return nil, nil
		}
		// กรณีอื่น ๆ คืน error ตามปกติ
		return nil, err
	}

	return &member, nil
}

func FindMemberWithLatestTierByTel(tel string) (*models.MemberWithTier, error) {
	if config.DB_POS == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var result models.MemberWithTier

	sql := `
		SELECT 
			m.*,
			cust_tier.tier
		FROM "member" m
		LEFT JOIN (
			SELECT DISTINCT ON (tel) tel, tier
			FROM customer_tier
			ORDER BY tel, create_date DESC
		) cust_tier ON m.tel = cust_tier.tel
		WHERE m.tel = ?
	`

	err := config.DB_POS.Raw(sql, tel).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	// ไม่พบ member
	if result.ID == 0 {
		return nil, nil
	}

	return &result, nil
}

func UpdateClaimJoylicoin(jc int, memberTel string, finalPoint int) (models.Member, error) {
	var member models.Member

	if config.DB_POS == nil {
		return member, fmt.Errorf("database connection is nil")
	}

	updateDate := utils.TimeNowAsia()

	result := config.DB_POS.Exec(`
		UPDATE member 
		SET 
			bonus = bonus + ?, 
			total_point = total_point + ?, 
			ecoin = 0, 
			update_date = ? 
		WHERE tel = ?`, jc, finalPoint, updateDate, memberTel)

	if result.Error != nil {
		return member, fmt.Errorf("failed to update member joylicoin: %w", result.Error)
	}

	// ดึงข้อมูลสมาชิกล่าสุดกลับคืนมา (optional)
	err := config.DB_POS.
		Where("tel = ?", memberTel).
		First(&member).Error

	if err != nil {
		return member, fmt.Errorf("failed to retrieve updated member: %w", err)
	}

	return member, nil
}

func UpdateSkill(tel string, input models.SkillDto, userId int) (models.Member, error) {
	var existing models.Member

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	trimmedTel := strings.TrimSpace(tel)
	if err := config.DB_POS.Where("tel = ?", trimmedTel).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("member not found: %w", err)
	}

	updateDate := utils.TimeNowAsia()

	result := config.DB_POS.Exec(`
		UPDATE member 
		SET 
			mskill1 = mskill1 + ?, 
			mskill2 = mskill2 + ?, 
			mskill3 = mskill3 + ?, 
			mskill4 = mskill4 + ?, 
			mskill5 = mskill5 + ?, 
			update_date = ? 
		WHERE tel = ?`,
		input.MSkill1, input.MSkill2, input.MSkill3, input.MSkill4, input.MSkill5, updateDate, trimmedTel)

	if result.Error != nil {
		return existing, fmt.Errorf("failed to update member skill: %w", result.Error)
	}

	// ดึงข้อมูลใหม่โดยใช้ tel
	if err := config.DB_POS.Where("tel = ?", trimmedTel).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated member: %w", err)
	}

	return existing, nil
}

func UpdatePoint(point int, memberTel string) (models.Member, error) {
	var member models.Member

	if config.DB_POS == nil {
		return member, fmt.Errorf("database connection is nil")
	}

	updateDate := utils.TimeNowAsia()

	result := config.DB_POS.Exec(`
		UPDATE member 
		SET 
			total_point = CASE 
                   WHEN total_point + ? < 0 THEN 0 
                   ELSE total_point + ? 
                END,
			update_date = ? 
		WHERE tel = ?`, point, point, updateDate, memberTel)

	if result.Error != nil {
		return member, fmt.Errorf("failed to update member joylicoin: %w", result.Error)
	}

	// ดึงข้อมูลสมาชิกล่าสุดกลับคืนมา (optional)
	err := config.DB_POS.
		Where("tel = ?", memberTel).
		First(&member).Error

	if err != nil {
		return member, fmt.Errorf("failed to retrieve updated member: %w", err)
	}

	return member, nil
}

func UpdateDepositJubuJibi(jubuJibi int, joylicoin int, topupDeduct int, memberTel string) (models.Member, error) {
	var member models.Member

	if config.DB_POS == nil {
		return member, fmt.Errorf("database connection is nil")
	}

	updateDate := utils.TimeNowAsia()

	if topupDeduct != 0 {
		// Bonus ต้องลบออก (เหมือน C#)
		result := config.DB_POS.Exec(`
			UPDATE member 
			SET 
				jubu_jibi = jubu_jibi + ?, 
				bonus = CASE 
                   WHEN bonus - ? < 0 THEN 0 
                   ELSE bonus - ? 
                END,
				update_date = ? 
			WHERE tel = ?`, jubuJibi, topupDeduct, topupDeduct, updateDate, memberTel)
		if result.Error != nil {
			return member, fmt.Errorf("failed to update member joylicoin: %w", result.Error)
		}
	} else {
		if joylicoin > 0 {
			// กรณี topup > 0
			result := config.DB_POS.Exec(`
				UPDATE member 
				SET 
					jubu_jibi = jubu_jibi + ?, 
					bonus = bonus + ?,
					update_date = ? 
				WHERE tel = ?`, jubuJibi, joylicoin, updateDate, memberTel)
			if result.Error != nil {
				return member, fmt.Errorf("failed to update member joylicoin: %w", result.Error)
			}
		} else {
			// กรณีอื่น ใช้ topupDeduct แต่เป็นบวก
			result := config.DB_POS.Exec(`
				UPDATE member 
				SET 
					jubu_jibi = jubu_jibi + ?, 
					bonus = CASE 
                   		WHEN bonus + ? < 0 THEN 0 
                   		ELSE bonus + ? 
                	END,
					update_date = ? 
				WHERE tel = ?`, jubuJibi, topupDeduct, topupDeduct, updateDate, memberTel)
			if result.Error != nil {
				return member, fmt.Errorf("failed to update member jubu jibi: %w", result.Error)
			}
		}
	}

	// ดึงข้อมูลสมาชิกล่าสุดกลับคืนมา
	err := config.DB_POS.
		Where("tel = ?", memberTel).
		First(&member).Error

	if err != nil {
		return member, fmt.Errorf("failed to retrieve updated member: %w", err)
	}

	return member, nil
}

func UpdateRedeemJubuJibi(jubuJibi int, memberTel string) (models.Member, error) {
	var member models.Member

	if config.DB_POS == nil {
		return member, fmt.Errorf("database connection is nil")
	}

	updateDate := utils.TimeNowAsia()

	// Bonus ต้องลบออก (เหมือน C#)
	result := config.DB_POS.Exec(`
			UPDATE member 
			SET 
				jubu_jibi = jubu_jibi - ?, 
				update_date = ? 
			WHERE tel = ?`, jubuJibi, updateDate, memberTel)
	if result.Error != nil {
		return member, fmt.Errorf("failed to update member jubu jibi : %w", result.Error)
	}

	// ดึงข้อมูลสมาชิกล่าสุดกลับคืนมา
	err := config.DB_POS.
		Where("tel = ?", memberTel).
		First(&member).Error

	if err != nil {
		return member, fmt.Errorf("failed to retrieve updated member: %w", err)
	}

	return member, nil
}

func UpdateAdjustPoint(input models.AdjustPointDto) (models.Member, error) {
	var member models.Member

	if config.DB_POS == nil {
		return member, fmt.Errorf("database connection is nil")
	}

	updateDate := utils.TimeNowAsia()

	// adj_estamp ขยับสองคอลัมน์ด้วยค่าเดียวกัน (ecoin และ estamp) แต่ controller
	// ตรวจเพดานให้แค่ ecoin — ตรวจข้อมูล 2026-09-18: สมาชิก 8198 คน มีแค่ 16 คน
	// ที่ ecoin = estamp ส่วนใหญ่ estamp = 0 ขณะที่ ecoin มีค่าจริง การหัก E-Stamp
	// จากสมาชิกทั่วไปจึงทำให้ estamp ติดลบ
	//
	// ปัดพื้นที่ 0 แทนการปฏิเสธรายการ เพราะเพดานจริงของธุรกรรมนี้คือ ecoin
	// (ที่ controller ตรวจแล้ว) ถ้ามาปฏิเสธเพราะ estamp ไม่พอจะบล็อกการหัก E-Stamp
	// ของสมาชิกเกือบทุกคนซึ่งวันนี้ทำได้ปกติ — ใช้รูปแบบเดียวกับที่ bonus
	// ถูกปัดพื้นอยู่แล้วใน UpdateDepositJubuJibi
	//
	// ยังเหลือคำถามที่ต้องให้เจ้าของระบบตอบ: ecoin กับ estamp ตั้งใจให้หมายถึง
	// อะไรกันแน่ ถ้าเป็นคนละอย่างก็ควรแยกช่องปรับออกจากกัน ไม่ใช่ใช้ค่าเดียวขยับทั้งคู่
	result := config.DB_POS.Exec(`
		UPDATE member
		SET
			bonus       = bonus + ?,
			total_point = total_point + ?, 
			ecoin       = ecoin + ?, 
			finwow      = finwow + ?, 
			mskill1     = mskill1 + ?,
			mskill2     = mskill2 + ?,
			mskill3     = mskill3 + ?,
			mskill4     = mskill4 + ?,
			mskill5     = mskill5 + ?,
			jubu_jibi   = jubu_jibi + ?,
			estamp      = CASE WHEN estamp + ? < 0 THEN 0 ELSE estamp + ? END,
			update_date = ?
		WHERE tel = ?`,
		input.AdjJoylicoin, // bonus
		input.AdjPoint,     // total_point
		input.AdjEstamp,    // ecoin
		input.AdjFinwow,    // finwow
		input.AdjMskill1,   // mskill1
		input.AdjMskill2,   // mskill2
		input.AdjMskill3,   // mskill3
		input.AdjMskill4,   // mskill4
		input.AdjMskill5,   // mskill5
		input.AdjJubuJibi,  // jubu_jibi
		input.AdjEstamp,    // estamp (เงื่อนไข CASE)
		input.AdjEstamp,    // estamp (ค่าที่เขียนจริง)
		updateDate,
		input.MemberTel,
	)

	if result.Error != nil {
		return member, fmt.Errorf("failed to update member score: %w", result.Error)
	}

	// ดึงข้อมูลสมาชิกล่าสุดกลับคืนมา
	err := config.DB_POS.
		Where("tel = ?", input.MemberTel).
		First(&member).Error
	if err != nil {
		return member, fmt.Errorf("failed to retrieve updated member: %w", err)
	}

	return member, nil
}

func UpdateJoylicoin(tel string, joylicoin int, userId int) (models.Member, error) {
	var existing models.Member

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	trimmedTel := strings.TrimSpace(tel)
	if err := config.DB_POS.Where("tel = ?", trimmedTel).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("member not found: %w", err)
	}

	updateDate := utils.TimeNowAsia()

	result := config.DB_POS.Exec(`
		UPDATE member 
		SET 
			bonus = bonus + ?,
			update_date = ? 
		WHERE tel = ?`,
		joylicoin, updateDate, trimmedTel)

	if result.Error != nil {
		return existing, fmt.Errorf("failed to update member joylicoin: %w", result.Error)
	}

	// ดึงข้อมูลใหม่โดยใช้ tel
	if err := config.DB_POS.Where("tel = ?", trimmedTel).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated member: %w", err)
	}

	return existing, nil
}

func UpdateMemberName(id int, input models.MemberNameDto, userId int) (models.Member, error) {
	var existing models.Member

	if config.DB_POS == nil {
		return existing, fmt.Errorf("database pos connection is nil")
	}

	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("member not found: %w", err)
	}

	// เตรียมข้อมูลที่ต้องการอัปเดต
	updateData := map[string]interface{}{
		"tel":         input.Tel,
		"m_name":      input.MName,
		"s_name":      input.SName,
		"update_date": utils.TimeNowAsia(),
	}

	// ใช้ .Model().Where().Updates() เพื่ออัปเดตเฉพาะฟิลด์
	if err := config.DB_POS.Model(&models.Member{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return existing, fmt.Errorf("failed to update member: %w", err)
	}

	// ดึงข้อมูลใหม่โดยใช้ Where ด้วย id
	if err := config.DB_POS.Where("id = ?", id).First(&existing).Error; err != nil {
		return existing, fmt.Errorf("failed to retrieve updated member: %w", err)
	}

	return existing, nil
}

func FindMemberWithLatestTierByTelAndUpdate(tel string) (*models.Member, error) {

	// get member จาก E-STAMP
	eStampMember, err := FindEStampMemberByTel(tel)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%+v\n", eStampMember)

	if eStampMember != nil {
		//update score ไป POS
		err = UpdatePosMemberScore(tel, eStampMember)
		if err != nil {
			return nil, err
		}
	}

	//get member จาก POS
	return FindExistMemberTel(tel)
}
func UpdatePosMemberScore(tel string, eStampMember *models.EStampMember) error {

	if config.DB_POS == nil {
		return fmt.Errorf("POS database connection is nil")
	}
	updateData := map[string]interface{}{
		"total_point": eStampMember.TotalPoint,
		"bonus":       eStampMember.Bonus,
		"ecoin":       eStampMember.Ecoin,
		"finwow":      eStampMember.Finwow,
		"estamp":      eStampMember.Estamp,
		"mskill1":     eStampMember.MSkill1,
		"mskill2":     eStampMember.MSkill2,
		"mskill3":     eStampMember.MSkill3,
		"mskill4":     eStampMember.MSkill4,
		"mskill5":     eStampMember.MSkill5,
		"jubu_jibi":   eStampMember.JubuJibi,
		"update_date": utils.TimeNowAsia(),
	}

	result := config.DB_POS.
		Table("member").
		Where("tel = ?", tel).
		Updates(updateData)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no member found with tel: %s", tel)
	}

	return nil
}
func FindEStampMemberByTel(tel string) (*models.EStampMember, error) {

	if config.DB_ESTAMP == nil {
		return nil, fmt.Errorf("E-Stamp database connection is nil")
	}

	var member models.EStampMember

	err := config.DB_ESTAMP.
		Table("Member").
		Where("Tel = ?", tel).
		First(&member).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &member, nil
}
