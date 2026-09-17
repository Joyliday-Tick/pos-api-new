package middlewares

import (
	"encoding/base64"
	"fmt"
	"net/http"

	// "os"
	"strings"

	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("AuthMiddleware")
		path := c.Request.URL.Path
		// ✅ เงื่อนไขไม่ให้ตรวจสอบ Swagger
		if strings.HasPrefix(path, "/api/swagger") || path == "/api/auth/login" || strings.HasPrefix(path, "/logoes/") {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		// fmt.Println("authHeader", authHeader)

		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "Authorization header missing")
			c.Abort()
			return
		}

		// Bearer token
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := utils.VerifyJwt(token)
			if err != nil {
				utils.Error(c, http.StatusUnauthorized, "Invalid Bearer token")
				c.Abort()
				return
			} else {
				// สามารถใช้ claims ที่ได้จาก JWT token ต่อไปได้ที่นี่
				c.Set("userClaims", claims)
			}
			c.Next()
			return
		}

		// Basic auth
		if strings.HasPrefix(authHeader, "Basic ") {
			encoded := strings.TrimPrefix(authHeader, "Basic ")
			decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				utils.Error(c, http.StatusUnauthorized, "Invalid Basic auth encoding")
				c.Abort()
				return
			}

			// SplitN คืน slice ยาว 1 ถ้าไม่มี ":" -> userPass[1] panic
			// ส่ง "Basic YQ==" (base64 ของ "a") ก็ทำให้ panic ได้โดยไม่ต้อง login
			userPass := strings.SplitN(string(decodedBytes), ":", 2)
			if len(userPass) != 2 {
				utils.Error(c, http.StatusUnauthorized, "Invalid Basic auth encoding")
				c.Abort()
				return
			}

			auth, err := services.Authenticate(userPass[0], userPass[1])
			if err != nil {
				utils.Error(c, http.StatusUnauthorized, "Invalid username or password")
				c.Abort()
				return
			}
			if !auth.Auth {
				utils.Error(c, http.StatusUnauthorized, "Invalid username or password")
				c.Abort()
				return
			}

			user := auth.Data
			// ต้องใส่ roleId ตรงนี้ด้วย ไม่งั้นการยืนยันตัวตนแบบ Basic จะไม่มี role
			// แล้ว RequireRole จะปฏิเสธทุกคน (หรือแย่กว่านั้นถ้าเผลอเขียนให้ปล่อยผ่าน
			// ก็จะกลายเป็นช่องข้ามการตรวจสิทธิ์ทั้งหมดด้วย header Basic บรรทัดเดียว)
			// Basic auth รับได้ทุก endpoint จึงต้องได้ role เท่ากับตอน login ปกติ
			roleId := 0
			if user.UserRoleId != nil {
				roleId = *user.UserRoleId
			}
			claims := jwt.MapClaims{
				"userId": int(user.ID),
				"roleId": roleId,
				"custId": 0,
				"iat":    nil,
				"exp":    nil,
			}
			c.Set("userClaims", claims)

			c.Next()
			return
		}

		// Unsupporte
		utils.Error(c, http.StatusUnauthorized, "Unsupported authorization type")
		c.Abort()
	}
}

// บทบาทในตาราง user_role (ตรวจกับฐานข้อมูลจริงเมื่อ 2026-09-17)
const (
	RoleAdministrator = 1
	RoleRMBackoffice  = 2
	RoleManager       = 3
	RoleAssistManager = 4
	RoleEmployee      = 5
	RoleFullTime      = 6
	RolePartTime      = 7
	RoleDisabled      = 8
)

var roleNames = map[int]string{
	RoleAdministrator: "Administrator",
	RoleRMBackoffice:  "RM/Backoffice",
	RoleManager:       "Manager",
	RoleAssistManager: "Assist Manager",
	RoleEmployee:      "Employee",
	RoleFullTime:      "Full-time",
	RolePartTime:      "Part-time",
	RoleDisabled:      "Disabled",
}

// GetRoleIdFromClaims อ่าน roleId จาก token
//
// คืน error เมื่ออ่านไม่ได้ ผู้เรียกต้องถือว่า "ไม่มีสิทธิ์" ห้ามถือว่าผ่าน
// ค่า 0 แปลว่าผู้ใช้ไม่มี user_role_id (คอลัมน์เป็น nullable) ซึ่งก็ต้องไม่ผ่านเช่นกัน
func GetRoleIdFromClaims(c *gin.Context) (int, error) {
	claims, exists := c.Get("userClaims")
	if !exists {
		return 0, fmt.Errorf("userClaims not found in context")
	}
	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims type")
	}
	switch v := mapClaims["roleId"].(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("roleId not found or invalid type")
	}
}

// RequireRole ปฏิเสธคำขอที่ roleId ไม่อยู่ในรายการที่อนุญาต
//
// ต้องวางต่อจาก AuthMiddleware() เสมอ เพราะอ่าน claims ที่ middleware นั้นตั้งไว้
//
// ⚠️ ข้อจำกัดที่ต้องรู้ก่อนใช้:
// หน้าจอ POS ไม่ได้เรียก API นี้ด้วย token ของพนักงานที่ล็อกอิน แต่เรียกผ่าน
// service account ชื่อ pos_frontend ซึ่งมี user_role_id = 1 (Administrator)
// การตรวจตรงนี้จึง **ไม่ได้จำกัดสิ่งที่หน้าจอ POS ทำได้เลย**
//
// สิ่งที่มันปิดคือช่องที่รายงานไว้จริง ๆ: พนักงานเอา username/password ของตัวเอง
// ไปยิง API ตรงเพื่อข้ามการขออนุมัติจากหัวหน้าที่กั้นไว้แค่ในหน้าจอ
// กรณีนั้น token จะมี roleId ของพนักงานคนนั้นเอง และจะถูกปฏิเสธที่นี่
//
// ถ้าต้องการให้การตรวจสิทธิ์มีผลกับหน้าจอ POS ด้วย ต้องเลิกใช้ service account
// ร่วมกัน แล้วส่งตัวตนของพนักงานจริงขึ้นมา ซึ่งเป็นงานคนละก้อน
func RequireRole(allowed ...int) gin.HandlerFunc {
	allowedSet := make(map[int]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		roleId, err := GetRoleIdFromClaims(c)
		if err != nil {
			// อ่าน role ไม่ได้ = ไม่ผ่าน ไม่ใช่ปล่อยผ่าน
			utils.Error(c, http.StatusForbidden, "ไม่สามารถตรวจสอบสิทธิ์ของผู้ใช้ได้")
			c.Abort()
			return
		}

		if _, ok := allowedSet[roleId]; !ok {
			name := roleNames[roleId]
			if name == "" {
				name = fmt.Sprintf("roleId %d", roleId)
			}
			utils.Error(c, http.StatusForbidden,
				fmt.Sprintf("สิทธิ์ %s ไม่สามารถใช้งานรายการนี้ได้ กรุณาให้ผู้มีสิทธิ์เป็นผู้ทำรายการ", name))
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetUserIdFromClaims(c *gin.Context) (int, error) {
	claims, exists := c.Get("userClaims")
	if !exists {
		return 0, fmt.Errorf("userClaims not found in context")
	}
	mapClaims, ok := claims.(jwt.MapClaims)
	// fmt.Println("mapClaims", mapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims type")
	}
	switch v := mapClaims["userId"].(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("userId not found or invalid type")
	}
}
