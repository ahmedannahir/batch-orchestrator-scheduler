package database

import (
	"fmt"
	"gestion-batches/entities"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Init() *gorm.DB {
	var db *gorm.DB

	err1 := godotenv.Load()
	if err1 != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_URL"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	var err2 error
	db, err2 = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err2 != nil {
		log.Fatal(err2)
	}

	db.AutoMigrate(&entities.Config{}, &entities.Batch{}, &entities.Execution{})

	return db
}
