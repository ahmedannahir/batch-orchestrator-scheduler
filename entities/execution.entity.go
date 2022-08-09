package entities

import "time"

type ExecutionStatus string

const (
	IDLE      ExecutionStatus = "0"
	RUNNING   ExecutionStatus = "1"
	COMPLETED ExecutionStatus = "2"
	FAILED    ExecutionStatus = "3"
)

type Execution struct {
	ID            uint `gorm:"primaryKey"`
	Status        ExecutionStatus
	Active        bool
	ExitCode      string     `gorm:"column:exitCode"`
	StartTime     time.Time  `gorm:"column:startTime"`
	EndTime       *time.Time `gorm:"column:endTime"`
	LogFileUrl    string     `gorm:"column:logFileUrl"`
	ErrLogFileUrl *string    `gorm:"column:errLogFileUrl"`
	BatchID       *uint      `gorm:"column:batchId"`
}

func (Execution) TableName() string {
	return "execution"
}
