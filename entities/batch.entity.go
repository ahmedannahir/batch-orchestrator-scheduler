package entities

import "gestion-batches/entities/BatchStatus"

type Batch struct {
	ID              uint `gorm:"primaryKey"`
	Active          bool
	Timing          string
	Name            string
	Description     string
	Url             string
	Args            *string
	Independant     bool
	Status          BatchStatus.BatchStatus
	PreviousBatchID *uint `gorm:"unique;column:previousBatchId"`
	PreviousBatch   *Batch
	ProfileID       *uint       `gorm:"column:profileId;not null"`
	Executions      []Execution `gorm:"foreignKey:BatchID"`
}

func (Batch) TableName() string {
	return "batch"
}
