// utils/function.go
package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

// "fmt"
// "time"

func IsAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

var macAddressRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`)

func IsValidMacAddress(mac string) bool {
	return macAddressRegex.MatchString(mac)
}

func ToStringSlice(raw json.RawMessage) []string {
	var macs []string
	_ = json.Unmarshal(raw, &macs)
	return macs
}

func GetBonusExpire() time.Time {
	// เวลา ณ ปัจจุบัน
	now := time.Now()

	// อ่านค่าปีจาก ENV (string → int)
	yearsStr := os.Getenv("BONUS_EXPIRE_YEARS")
	years, err := strconv.Atoi(yearsStr)
	if err != nil {
		years = 0
	}

	// +years ปี
	future := now.AddDate(years, 0, 0)

	// set เวลาเป็น 00:00:00
	futureMidnight := time.Date(
		future.Year(),
		future.Month(),
		future.Day(),
		0, 0, 0, 0,
		future.Location(),
	)

	fmt.Println("Now:          ", now.Format("2006-01-02 15:04:05"))
	fmt.Println("Future (00:00)", futureMidnight.Format("2006-01-02 15:04:05"))

	return futureMidnight
}

func ToRawMessage(v interface{}) (*json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	rm := json.RawMessage(b)
	return &rm, nil
}

func MaskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return phone[:len(phone)-4] + "XXXX"
}
