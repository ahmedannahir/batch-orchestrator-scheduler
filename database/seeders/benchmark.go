package seeders

import (
	"batch/main/models"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"
)

var clientsNumber = 100
var transactionPerClient = 50

var clients []models.Client
var transactions []models.Transaction

func BenchmarkToDb(db *gorm.DB) {
	for i := 1; i <= clientsNumber; i++ {
		clients = append(clients, models.Client{
			Name:      "Client" + strconv.Itoa(i),
			Balance:   500 * float64(i),
			Mail:      "client" + strconv.Itoa(i) + "@mail.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	operationsTypes := []models.OperationType{models.IN, models.OUT}

	for k := 1; k <= clientsNumber; k++ {
		for j := 1; j <= transactionPerClient; j++ {
			transactions = append(transactions, models.Transaction{
				ClientID:      uint(k),
				Amount:        90 * float64(rand.Intn(3)+1),
				OperationType: operationsTypes[rand.Intn(2)],
				OperationDate: time.Now(),
			})
		}
	}

	db.Create(&clients)
	db.Create(&transactions)
}

func BenchmarkToFile(dbCon *gorm.DB, filePath string) {
	var transactions []models.Transaction
	dbCon.Where("operation_date > ?", time.Now().AddDate(0, 0, -1)).Find(&transactions)

	bytes_transactions, errJ := json.Marshal(transactions)
	if errJ != nil {
		log.Fatal(errJ)
	}

	file, errF := os.Create(filePath)
	if errF != nil {
		log.Fatal(errF)
	}
	defer file.Close()

	_, errW := file.Write(bytes_transactions)
	if errW != nil {
		log.Fatal(errW)
	}
}
