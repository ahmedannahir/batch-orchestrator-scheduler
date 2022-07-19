package database

import (
	"batch/main/database/seeders"
	"batch/main/models"
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DbCon *gorm.DB

func init() {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_URL"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	DbCon, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	DbCon.Migrator().DropTable(&models.Client{}, &models.Transaction{})
	DbCon.AutoMigrate(&models.Client{}, &models.Transaction{})

	seeders.BenchmarkToDb(DbCon)
	seeders.BenchmarkToFile(DbCon, "jobs/batches/transactions.json")
}
