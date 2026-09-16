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
			fmt.Println("Bearer token", token)
			claims, err := utils.VerifyJwt(token)
			if err != nil {
				utils.Error(c, http.StatusUnauthorized, "Invalid Bearer token")
				c.Abort()
				return
			} else {
				fmt.Println("JWT Claims:", claims)
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

			userPass := strings.SplitN(string(decodedBytes), ":", 2)
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
			claims := jwt.MapClaims{
				"userId": int(user.ID),
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
