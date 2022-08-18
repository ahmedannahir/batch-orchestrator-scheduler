package entities

import (
	"gestion-batches/entities/ExecutionStatus"
	"time"
)

type Execution struct {
	ID            uint `gorm:"primaryKey"`
	Status        ExecutionStatus.ExecutionStatus
	ExitCode      string     `gorm:"column:exitCode"`
	StartTime     *time.Time `gorm:"column:startTime"`
	EndTime       *time.Time `gorm:"column:endTime"`
	LogFileUrl    *string    `gorm:"column:logFileUrl"`
	ErrLogFileUrl *string    `gorm:"column:errLogFileUrl"`
	BatchID       *uint      `gorm:"column:batchId;not null"`
}

func (Execution) TableName() string {
	return "execution"
}
