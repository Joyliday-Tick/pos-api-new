package middlewares

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func APILogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// ทำงานก่อนถึง handler
		path := c.Request.URL.Path
		method := c.Request.Method
		log.Printf("[API] %s %s", method, path)

		c.Next() // ไปยัง handler ต่อ

		// หลัง handler
		status := c.Writer.Status()
		duration := time.Since(start)
		log.Printf("[API] Done - %d (%v)", status, duration)
	}
}
