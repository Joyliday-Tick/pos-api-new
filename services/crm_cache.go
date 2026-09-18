package services

import (
	"strings"
	"sync"
	"time"

	"new-pos-api/models"
)

// แคชข้อมูลอ้างอิงของ CRM ที่แทบไม่เปลี่ยน
//
// ทำไมต้องมี: เส้นทางเติมเงินที่มี free point เรียก CRM ห้าครั้งเรียงกัน
// วัดจริงเมื่อ 2026-09-18 ด้วย container ที่ต่อ CRM ได้ปกติ — ใช้เวลา 4.65 วินาที
// ต่อการเติมเงินหนึ่งรายการ แคชเชียร์ต้องยืนรอต่อหน้าลูกค้าทุกครั้ง
// และถ้า CRM ช้า แต่ละครั้ง timeout 10 วินาที (ดู crm_service.go) รวมได้ถึง 50 วินาที
//
// สองตัวนี้เป็นข้อมูลตั้งต้น ไม่ใช่ข้อมูลรายการ:
//
//	ScoreType  ประเภทคะแนน — ค่าคงที่ของระบบ
//	Branch     รหัสสาขา — เปลี่ยนตอนเปิด/ปิดสาขาเท่านั้น
//
// ถูกเรียกจากห้าที่ (topup, adjust-point, claim-prize, return-bonus และ
// crm_controller) การแคชในตัว service จึงได้ผลทุกเส้นทางโดยไม่ต้องแก้ผู้เรียก
//
// ⚠️ แคชเฉพาะผลที่สำเร็จ ความล้มเหลวต้องไปถาม CRM ใหม่เสมอ
// ไม่งั้น CRM ล่มชั่วคราวจะกลายเป็นล่มค้างไปทั้ง TTL
const crmRefDataTTL = 30 * time.Minute

var (
	scoreTypeMu     sync.RWMutex
	scoreTypeCache  []models.ScoreType
	scoreTypeSetAt  time.Time
	scoreTypeStatus int

	branchMu    sync.RWMutex
	branchCache = map[string]branchCacheEntry{}
)

type branchCacheEntry struct {
	branch *models.BranchCrm
	status int
	setAt  time.Time
}

// GetScoreType คืนประเภทคะแนนจากแคช ถ้าหมดอายุจึงไปถาม CRM
func GetScoreType() (int, []models.ScoreType, error) {
	scoreTypeMu.RLock()
	if len(scoreTypeCache) > 0 && time.Since(scoreTypeSetAt) < crmRefDataTTL {
		status, cached := scoreTypeStatus, scoreTypeCache
		scoreTypeMu.RUnlock()
		return status, cached, nil
	}
	scoreTypeMu.RUnlock()

	status, types, err := fetchScoreTypesFromCRM()
	if err != nil || status != 200 || len(types) == 0 {
		return status, types, err
	}

	scoreTypeMu.Lock()
	scoreTypeCache, scoreTypeStatus, scoreTypeSetAt = types, status, time.Now()
	scoreTypeMu.Unlock()

	return status, types, nil
}

// GetBranchByCode คืนข้อมูลสาขาจากแคช ถ้าหมดอายุจึงไปถาม CRM
func GetBranchByCode(code string) (int, *models.BranchCrm, error) {
	key := strings.ToLower(strings.TrimSpace(code))

	branchMu.RLock()
	entry, ok := branchCache[key]
	branchMu.RUnlock()
	if ok && entry.branch != nil && time.Since(entry.setAt) < crmRefDataTTL {
		return entry.status, entry.branch, nil
	}

	status, branch, err := fetchBranchByCodeFromCRM(code)
	if err != nil || status != 200 || branch == nil {
		return status, branch, err
	}

	branchMu.Lock()
	branchCache[key] = branchCacheEntry{branch: branch, status: status, setAt: time.Now()}
	branchMu.Unlock()

	return status, branch, nil
}
