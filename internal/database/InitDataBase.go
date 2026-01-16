package database

import (
	"auction/internal/model"
	"fmt"
	"os"

	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	// 从环境变量获取MySQL连接配置
	dbHost := GetEnv("DB_HOST", "127.0.0.1")
	dbPort := GetEnv("DB_PORT", "3306")
	dbUser := GetEnv("DB_USERNAME", "root")
	dbPassword := GetEnv("DB_PASSWORD", "password")
	dbName := GetEnv("DB_NAME", "auction_db")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)

	// gorm connection
	gormDb, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(err)
	}
	log.Println("connect db success")

	// 自动迁移数据库表
	gormDb.AutoMigrate(&model.Auction{}, model.Bid{}, model.ProcessedLog{})

	log.Println("init db tables end")
	return gormDb
}

func GetEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); len(value) != 0 {
		return value
	}
	return defaultValue
}
