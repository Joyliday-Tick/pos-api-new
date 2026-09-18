// utils/response.go
package utils

import (
	"fmt"
	"strings"
	"time"
)

func TimeFormattedNow() string {
	now := time.Now()
	formatted := now.Format("2006-01-02 15:04:05")
	fmt.Println(formatted)
	return formatted
}

func TimeNowPtr() *time.Time {
	now := time.Now()
	fmt.Println(&now)
	return &now
}

// TimeNowAsia คืนเวลาปัจจุบันตามเขตเวลาไทย
//
// ใช้กับทุกคอลัมน์วันที่ในระบบนี้ เพราะคอลัมน์เป็น timestamp without time zone
// ค่าที่เก็บจึงเป็น "เวลาไทยแบบไม่มีโซน" ตรง ๆ
// เวลาจะเทียบหรือจัดกลุ่มตามวัน/เดือน ต้องสร้างช่วงเวลาจากฟังก์ชันนี้ด้วย
// ไม่ใช่ time.Now() ซึ่งใน container เป็น UTC (ดู GenerateBillNo)
//
// เดิมพิมพ์ 3 บรรทัดทุกครั้งที่ถูกเรียก และมันถูกเรียกในทุกเส้นทางที่ขยับเงิน
// เสียงรบกวนนั้นกลบ log ที่ใช้ตามปัญหาจริง จึงเอาออก
//
// ⚠️ panic ถ้าโหลด tzdata ไม่ได้ — ตอนนี้ปลอดภัยเพราะ Dockerfile ใช้
// golang:1.24 ซึ่งมี tzdata ติดมา ถ้าย้ายไป base image เล็กลง (alpine, scratch)
// ต้อง import _ "time/tzdata" หรือติดตั้ง tzdata ไม่งั้นระบบจะตายทันทีที่บูต
func TimeNowAsia() *time.Time {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}
	now := time.Now().In(loc)
	return &now
}

type DateOnly struct {
	time.Time
}

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", d.Time.Format("2006-01-02"))), nil
}

// func CalculateDate(date time.Time) string {
// 	fmt.Println("Input Date:", date)
// 	// เวลา ณ ปัจจุบัน
// 	loc, err := time.LoadLocation("Asia/Bangkok")
// 	if err != nil {
// 		panic(err)
// 	}
// 	now := time.Now().In(loc)

// 	// คำนวณระยะเวลา
// 	duration := date.Sub(now)

// 	fmt.Println("Duration:", duration)

// 	if duration <= 0 {
// 		return "expired"
// 	}

// 	// แยกวัน ชั่วโมง นาที
// 	days := int(duration.Hours()) / 24
// 	hours := int(duration.Hours()) % 24
// 	minutes := int(duration.Minutes()) % 60

// 	return fmt.Sprintf("%ddays %dhour %dmin", days, hours, minutes)
// }

func CalculateDate(date time.Time) string {
	fmt.Println("Input Date (raw):", date)

	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}

	// ✅ บังคับให้ Input Date ถูกตีความว่าเป็นเวลาไทย
	// โดยไม่ขยับเวลา (ไม่บวก 7 ชม.)
	dateInThai := time.Date(
		date.Year(), date.Month(), date.Day(),
		date.Hour(), date.Minute(), date.Second(), date.Nanosecond(),
		loc,
	)

	now := time.Now().In(loc)
	duration := dateInThai.Sub(now)

	fmt.Println("Date in Thai:", dateInThai)
	fmt.Println("Now in Thai:", now)
	fmt.Println("Duration:", duration)

	if duration <= 0 {
		return "expired"
	}

	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	return fmt.Sprintf("%ddays %dhour %dmin", days, hours, minutes)
}

// AddDaysBangkok เพิ่มวันจากวันที่ปัจจุบันใน timezone Asia/Bangkok
func AddDaysBangkok(days int) *time.Time {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		loc = time.UTC // fallback เผื่อ timezone หาย
	}

	t := time.Now().In(loc).AddDate(0, 0, days)
	truncated := t.Truncate(time.Minute)
	return &truncated
}

func AddMinutesBangkok(minutes int) *time.Time {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		loc = time.UTC
	}
	t := time.Now().In(loc).Add(time.Duration(minutes) * time.Minute)

	truncated := t.Truncate(time.Minute)
	return &truncated
}
