package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB_ESTAMP *gorm.DB

func ConnectDatabaseEStamp() {

	// โหลด .env (ควรเรียกครั้งเดียวใน main จะดีกว่า)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found (production mode)")
	}

	dsn := os.Getenv("E_STAMP_DSN")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database E-Stamp:", err)
	}

	DB_ESTAMP = db

	fmt.Println("E-Stamp Database connected successfully")
}
