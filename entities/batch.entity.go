package entities

import "gestion-batches/entities/BatchStatus"

type Batch struct {
	ID              uint `gorm:"primaryKey"`
	Timing          string
	Name            string
	Description     string
	Url             string
	Independant     bool
	Status          BatchStatus.BatchStatus
	PrevBatchInput  bool  `gorm:"column:prevBatchInput"`
	PreviousBatchID *uint `gorm:"unique;column:previousBatchId"`
	PreviousBatch   *Batch
	Executions      []Execution `gorm:"foreignKey:BatchID"`
}

func (Batch) TableName() string {
	return "batch"
}
