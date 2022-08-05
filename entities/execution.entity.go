package entities

import "time"

type ExecutionStatus string

const (
	IDLE      ExecutionStatus = "IDLE"
	RUNNING   ExecutionStatus = "RUNNING"
	COMPLETED ExecutionStatus = "COMPLETED"
	FAILED    ExecutionStatus = "FAILED"
)

type Execution struct {
	ID            uint `gorm:"primaryKey"`
	Status        ExecutionStatus
	ExitCode      string
	StartTime     time.Time
	EndTime       *time.Time
	LogFileUrl    string
	ErrLogFileUrl *string
	BatchID       *uint
}
