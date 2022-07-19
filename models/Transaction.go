package models

import "time"

type OperationType string

const (
	IN  OperationType = "IN"
	OUT OperationType = "OUT"
)

type Transaction struct {
	OperationCode uint `gorm:"primaryKey"`
	ClientID      uint
	Amount        float64
	OperationType OperationType
	OperationDate time.Time
}
