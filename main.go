package main

import (
	"fmt"
	"log"
	"net/http"
	"new-pos-api/config"
	docs "new-pos-api/docs" // Import docs (ต้อง generate ก่อนใช้งาน)
	"new-pos-api/utils"
	"os"
	"time"

	"new-pos-api/routes"

	"new-pos-api/middlewares"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/cors"

	nrgin "github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gorm.io/gorm"
)

// tunePool ตั้งขนาด connection pool ของฐานหนึ่ง ๆ แล้วคืนฟังก์ชันปิดการเชื่อมต่อ
//
// เหตุผลที่ต้องรวมไว้ที่เดียว: ของเดิมเขียนแยกเป็นบล็อกซ้ำ ๆ สามชุด
// แล้ว "ลืม" ตั้งของ DB_JREADER ไปทั้งตัว ทำให้ pool นั้นใช้ค่า default ของ Go
// ซึ่งคือ MaxOpenConns ไม่จำกัด
func tunePool(name string, db *gorm.DB, maxOpen, maxIdle int) func() {
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("ตั้งค่า connection pool ของฐาน %s ไม่สำเร็จ: %v\n", name, err)
		return nil
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	return func() { sqlDB.Close() }
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, authenticated user!")
}

//// @host 139.59.223.142

//// @host localhost:8080

//// @host pos-api.apices.info

// @title New POS APIls
// @version 1.0
// @description This is a POS API with Gin and Swagger
//// @host localhost:8080
// @BasePath /

// @securityDefinitions.basic BasicAuth
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	config.ConnectDatabase()
	config.ConnectDatabasePos()
	config.ConnectDatabaseJRFeader()
	config.ConnectDatabaseEStamp()

	// ต้องอยู่ "หลัง" ConnectDatabase เพราะตัวนั้นเป็นที่เรียก godotenv.Load()
	// ล้มตั้งแต่ตอนบูตถ้ากุญแจไม่พร้อม ดีกว่าขึ้นมาให้บริการด้วยกุญแจว่าง
	// ซึ่งจะทำงานปกติทุกอย่างแต่ใครก็ปลอม token admin ได้
	if err := utils.InitJwtSecret(); err != nil {
		log.Fatalf("เริ่มระบบไม่ได้: %v", err)
	}
	app, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("pos-production"),
		newrelic.ConfigLicense(os.Getenv("NEWRELIC_LICENSE")),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)

	// ขนาด pool มาจากการวัดโควตาจริงของแต่ละฐานเมื่อ 2026-09-16 ไม่ใช่ค่าที่เดา
	//
	//   ฐาน                        เพดาน   ใช้อยู่ตอนวัด   ตั้งไว้เดิม
	//   joyliday      (DB)          100      13          50
	//   defaultdb     (DB_POS)       50      15          50   <-- กินโควตาทั้งฐานพอดี
	//   jreader_iot   (DB_JREADER)  100      12          ไม่ได้ตั้ง = ไม่จำกัด
	//   Estamp MySQL  (DB_ESTAMP)   วัดไม่ได้   -          100
	//
	// ต้องตั้งให้แต่ละ instance กินน้อย เพราะมีหลาย instance ชี้ฐานเดียวกัน
	// (UAT droplet เดิม + UAT2 + production ที่ HPA ขยายได้ถึง 2 replica)
	// และต้องเหลือ slot ให้คนที่เปิด DBeaver ด้วย ซึ่งเป็นต้นเหตุที่ POS ล่ม
	// เมื่อ 2026-09-15 (ตอนนั้น pos-api ใช้แค่ 2 connection จึงไม่ใช่ผู้ร้าย
	// แต่ค่าเดิมข้างบนจะทำให้รอบหน้า pos-api เป็นผู้ร้ายเสียเอง)
	//
	// ตัวที่ deploy อยู่จริงใช้แค่ 3 connection ค่า 10 จึงเหลือเฟือ
	pools := []struct {
		name    string
		db      *gorm.DB
		maxOpen int
		maxIdle int
	}{
		{"joyliday", config.DB, 15, 3},
		{"pos", config.DB_POS, 10, 2},
		{"jreader", config.DB_JREADER, 10, 2},
		{"e-stamp", config.DB_ESTAMP, 10, 2},
	}
	for _, p := range pools {
		if closer := tunePool(p.name, p.db, p.maxOpen, p.maxIdle); closer != nil {
			defer closer()
		}
	}

	host := os.Getenv("API_HOST")
	if host == "" {
		// ถ้าไม่กำหนด env var ให้ใช้ค่า default (เช่นสำหรับ local dev)
		host = "localhost:8080"
	}

	// กำหนด Host ให้กับ SwaggerInfo
	docs.SwaggerInfo.Host = host

	r := gin.Default()

	r.Use(gin.Logger())   // Log requests
	r.Use(gin.Recovery()) // Handle panic gracefully
	r.Use(cors.Default())
	r.Use(nrgin.Middleware(app))
	r.Use(middlewares.AuthMiddleware())
	r.Static("/logoes", "./logoes")
	// เพิ่ม Swagger Routes
	r.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.DocExpansion("none")))
	routes.SetupRouter(r)
	r.Run(":8080")
}
