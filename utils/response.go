// utils/response.go
package utils

import (
	"new-pos-api/models"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(200, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"success": false,
		"message": message,
		"data":    nil,
	})
}

// ErrorWithData เหมือน Error แต่แนบข้อมูลให้ฝั่งหน้าจออ่านได้
//
// ใช้เมื่อหน้าจอต้องแยกแยะสาเหตุ ไม่ใช่แค่แสดงข้อความ เช่น 409 ของ /pos-void
// มีสองความหมาย (บิลถูกยกเลิกไปแล้ว / ยอดไม่พอต้องขออนุมัติ) ซึ่งต้องทำต่อ
// คนละอย่าง การให้หน้าจอเดาจากข้อความไทยเปราะเกินไป
//
// ซองยังเหมือนเดิมทุกฟิลด์ ผู้เรียกที่อ่านแค่ message จึงไม่กระทบ
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, gin.H{
		"success": false,
		"message": message,
		"data":    data,
	})
}

type StandardSuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data"`
}

type StandardErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Message string      `json:"message" example:"Failed"`
	Data    interface{} `json:"data"` // null
}
type UserListResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    []models.Users `json:"data"`
}

type SearchResult struct {
	Page       int         `json:"page"`
	TotalCount int         `json:"totalCount"`
	Result     interface{} `json:"result"`
}
