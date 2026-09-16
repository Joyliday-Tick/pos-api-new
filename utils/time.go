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

func TimeNowAsia() *time.Time {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}
	now := time.Now().In(loc)
	fmt.Println(now)
	fmt.Println("DateTime (Bangkok):", now.Format("2006-01-02 15:04:05"))
	fmt.Println("Timezone:", now.Location())
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
