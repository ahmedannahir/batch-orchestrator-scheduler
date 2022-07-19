package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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

	yesterdayTime := time.Now().Add(-24 * time.Hour)
	sqlQuery := fmt.Sprintf("SELECT * from transactions WHERE DATE(operation_date) > \"%s\"", yesterdayTime)

	transactions, errQ := db.Query(sqlQuery)
	defer transactions.Close()

	if errQ != nil {
		log.Fatal(errQ)
	}

	var number_transactions int
	db.QueryRow(fmt.Sprintf("SELECT count(*) from transactions WHERE DATE(operation_date) > \"%s\"", yesterdayTime)).Scan(&number_transactions)

	number_processed := 0
	for transactions.Next() {
		var transaction Transaction

		err := transactions.Scan(
			&transaction.OperationCode,
			&transaction.ClientID,
			&transaction.Amount,
			&transaction.OperationType,
			&transaction.OperationDate,
		)

		if err != nil {
			log.Fatal(err)
			break
		}

		var op string
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

	log.Println("::", (number_processed/number_transactions)*100, "%", "Completed in", elapsed)
}
