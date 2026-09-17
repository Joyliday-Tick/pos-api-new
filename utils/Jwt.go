package utils

import (
	"errors"
	"fmt"
	"log"
	"new-pos-api/models"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ความยาวขั้นต่ำของกุญแจ HS256 — 32 ไบต์คือขนาดของ SHA-256 เอง
// สั้นกว่านี้เปิดช่องให้ brute force แบบ offline ถ้ามีใครดัก token ได้สักใบ
const minJwtSecretLen = 32

var jwtSecret []byte

// InitJwtSecret ต้องถูกเรียกจาก main() "หลัง" โหลด .env เสร็จแล้ว
//
// ของเดิมเขียนเป็น package-level var:
//
//	var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
//
// ซึ่ง Go สร้างค่าตอน package init คือ "ก่อน" main() เรียก godotenv.Load()
// ถ้า JWT_SECRET มาจากไฟล์ .env อย่างเดียว กุญแจจะเป็น []byte("") ทันที
// และจะไม่มีอาการผิดปกติใด ๆ ให้เห็นเลย เพราะทั้งการเซ็นและการตรวจ
// ใช้กุญแจว่างเหมือนกัน ระบบทำงานปกติทุกอย่าง แต่ใครก็ปลอม token admin ได้
//
// ย้ายมาอ่านตอน runtime และให้ล้มตั้งแต่ตอนบูตถ้าค่าไม่พร้อม
// ตั้งใจให้ "ตั้งค่าผิดแล้วดัง" แทน "ตั้งค่าผิดแล้วเงียบ"
func InitJwtSecret() error {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return errors.New(
			"ไม่ได้ตั้งค่า JWT_SECRET — ต้องส่งเป็น environment variable จริง " +
				"ไม่ใช่มีแค่ในไฟล์ .env เพราะโค้ดอ่านค่าก่อน godotenv.Load()",
		)
	}
	if len(secret) < minJwtSecretLen {
		return fmt.Errorf(
			"JWT_SECRET ยาว %d ตัวอักษร สั้นเกินไปสำหรับ HS256 (ต้องอย่างน้อย %d) "+
				"สร้างใหม่ด้วย: openssl rand -base64 48",
			len(secret), minJwtSecretLen,
		)
	}
	jwtSecret = []byte(secret)
	return nil
}

// กันกรณีลืมเรียก InitJwtSecret — ยอมให้ request พังดีกว่าเซ็นด้วยกุญแจว่าง
func requireJwtSecret() ([]byte, error) {
	if len(jwtSecret) == 0 {
		return nil, errors.New("JWT_SECRET ยังไม่ถูกตั้งค่า (ลืมเรียก utils.InitJwtSecret ใน main?)")
	}
	return jwtSecret, nil
}

func GetJWT(user models.AuthResultBranch, from string) (map[string]interface{}, error) {
	// Load timezone Bangkok
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return nil, err
	}
	expireDaysStr := os.Getenv("TOKEN_EXPIRE")
	fmt.Println("expireDaysStr", expireDaysStr)
	expireDays, err := strconv.Atoi(expireDaysStr)
	if err != nil {
		log.Fatalf("Invalid TOKEN_EXPIRE value: %v", err)
	}
	now := time.Now().In(loc)
	exp := time.Now().Add(time.Hour * 24 * time.Duration(expireDays))

	claims := jwt.MapClaims{
		"userId": int(user.Data.ID),
		"roleId": user.Data.UserRoleId,
		"iat":    now.Unix(),
		"exp":    exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret, err := requireJwtSecret()
	if err != nil {
		return nil, err
	}

	signedToken, err := token.SignedString(secret)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"userId":     int(user.Data.ID),
		"username":   user.Data.Username,
		"token":      signedToken,
		"roleId":     user.Data.UserRoleId,
		"roleName":   user.RoleName,
		"groupId":    user.Data.UserGroupId,
		"groupName":  user.GroupName,
		"branchList": user.BranchList,
	}

	return data, nil
}

// VerifyJwt รับ token และตรวจสอบความถูกต้องของมัน
func VerifyJwt(tokenStr string) (jwt.MapClaims, error) {
	// Parse token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// ตรวจสอบว่า algorithm ที่ใช้คือ "HS256"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return requireJwtSecret()
	})

	if err != nil {
		return nil, err
	}

	// ตรวจสอบว่า token เป็น valid หรือไม่
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("invalid token")
	}
}
