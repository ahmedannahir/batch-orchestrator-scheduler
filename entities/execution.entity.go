package entities

import "time"

type ExecutionStatus string

const (
	RUNNING   ExecutionStatus = "RUNNING"
	COMPLETED ExecutionStatus = "COMPLETED"
	FAILED    ExecutionStatus = "FAILED"
)

type Execution struct {
	ID         uint `gorm:"primaryKey"`
	Status     ExecutionStatus
	ExitCode   string
	StartTime  time.Time
	EndTime    *time.Time //default value for time.Time is 0000... instead of null, *time.Time works as intended
	LogFileUrl string
	BatchID    uint
}
