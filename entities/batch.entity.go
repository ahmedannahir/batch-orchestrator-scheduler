package entities

type Batch struct {
	ID              uint `gorm:"primaryKey"`
	Timing          string
	Name            string
	Description     string
	Url             string
	Independant     bool
	PrevBatchInput  bool  `gorm:"column:prevBatchInput"`
	PreviousBatchID *uint `gorm:"unique;column:previousBatchId"`
	PreviousBatch   *Batch
	Executions      []Execution `gorm:"foreignKey:BatchID"`
}

func (Batch) TableName() string {
	return "batch"
}
