package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bitfield/script"
	_ "github.com/go-sql-driver/mysql"
)

type Transaction struct {
	OperationCode uint
	ClientID      uint
	Amount        float64
	OperationType string
	OperationDate time.Time
}

func main() {
	start := time.Now()

	var transactions []Transaction

	transactionsFromJson, errIo := script.File("jobs/batches/transactions.json").Bytes()
	if errIo != nil {
		log.Fatal(errIo)
	}

	json.Unmarshal(transactionsFromJson, &transactions)

	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_URL"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	))

	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	var op string

	var number_transactions int
	number_transactions = len(transactions)

	number_processed := 0
	for _, transaction := range transactions {
		if transaction.OperationType == "IN" {
			op = "+"
		} else {
			op = "-"
		}
		updateQuery := fmt.Sprintf("UPDATE clients SET balance = balance %s %f WHERE code = %d", op, transaction.Amount, transaction.ClientID)

		db.Query(updateQuery)
		number_processed++
	}

	elapsed := time.Since(start)

	log.Println("::", (number_processed/number_transactions)*100, "%", "Completed in", elapsed.String())
}
