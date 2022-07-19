package database

import (
	"batch/main/database/seeders"
	"batch/main/models"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	logFile, err := os.Create("logs/gorm_" + time.Now().Format("2006-01-02_15-04-05") + ".log")
	if err != nil {
		panic(err)
	}

	multiOutput := io.MultiWriter(os.Stdout, logFile)
	fileLogger := logger.New(
		log.New(multiOutput, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)

	DbCon, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: fileLogger,
	})
	if err != nil {
		log.Fatal(err)
	}

	DbCon.Migrator().DropTable(&models.Client{}, &models.Transaction{})
	DbCon.AutoMigrate(&models.Client{}, &models.Transaction{})

	seeders.BenchmarkToDb(DbCon)
	seeders.BenchmarkToFile(DbCon, "jobs/batches/transactions.json")
}
