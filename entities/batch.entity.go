package entities

type Batch struct {
	ID              uint `gorm:"primaryKey"`
	Name            string
	Description     string
	Url             string
	ConfigID        *uint `gorm:"column:configId"`
	PreviousBatchID *uint `gorm:"unique;column:previousBatchId"`
	PreviousBatch   *Batch
	Executions      []Execution `gorm:"foreignKey:BatchID"`
}

func (Batch) TableName() string {
	return "batch"
}
