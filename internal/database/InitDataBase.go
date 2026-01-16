package database

import (
	"auction/internal/model"
	"os"

	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {

	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatalln("读取.env文件失败❌")
	// }
	// // 从环境变量获取MySQL连接配置
	// dbHost := GetEnv("DB_HOST", "")
	// dbPort := GetEnv("DB_PORT", "")
	// dbUser := GetEnv("DB_USERNAME", "")
	// dbPassword := GetEnv("DB_PASSWORD", "")
	// dbName := GetEnv("DB_NAME", "")

	// dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)

	// gorm connection
	gormDb, err := gorm.Open(mysql.Open("root:st123456@tcp(127.0.0.1:3306)/auction_db_v1?charset=utf8mb4&parseTime=True&loc=Local"))
	if err != nil {
		panic(err)
	}
	log.Println("connect db success")

	// &model.Auction{}
	gormDb.AutoMigrate(&model.Auction{}, model.Bid{})

	log.Println("init db tables end")
	return gormDb
}

func GetEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); len(value) != 0 {
		return value
	}
	return defaultValue
}
