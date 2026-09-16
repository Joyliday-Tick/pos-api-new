package main

import (
	"fmt"
	"net/http"
	"new-pos-api/config"
	docs "new-pos-api/docs" // Import docs (ต้อง generate ก่อนใช้งาน)
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
)

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
	app, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("pos-production"),
		newrelic.ConfigLicense(os.Getenv("NEWRELIC_LICENSE")),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)

	sqlDB1, err := config.DB.DB()
	if err != nil {
		fmt.Println("Failed to get main database instance:", err)
	} else {

		sqlDB1.SetMaxOpenConns(50)
		sqlDB1.SetMaxIdleConns(10)
		sqlDB1.SetConnMaxLifetime(5 * time.Minute)
		sqlDB1.SetConnMaxIdleTime(5 * time.Minute)

		defer sqlDB1.Close()
	}

	sqlDB2, err := config.DB_POS.DB()
	if err != nil {
		fmt.Println("Failed to get POS database instance:", err)
	} else {
		sqlDB2.SetMaxOpenConns(50)
		sqlDB2.SetMaxIdleConns(10)
		sqlDB2.SetConnMaxLifetime(5 * time.Minute)
		sqlDB2.SetConnMaxIdleTime(5 * time.Minute)

		defer sqlDB2.Close()
	}

	sqlDB3, err := config.DB_ESTAMP.DB()
	if err != nil {
		fmt.Println("Failed to get E-Stamp database instance:", err)
	} else {

		sqlDB3.SetMaxOpenConns(100)
		sqlDB3.SetMaxIdleConns(25)
		sqlDB3.SetConnMaxLifetime(5 * time.Minute)
		defer sqlDB3.Close()
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
