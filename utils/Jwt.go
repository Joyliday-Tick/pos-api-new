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

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

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

	signedToken, err := token.SignedString(jwtSecret)
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
		return jwtSecret, nil
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
