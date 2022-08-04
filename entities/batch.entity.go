package entities

type Batch struct {
	ID              uint `gorm:"primaryKey"`
	Name            string
	Description     string
	Url             string
	ConfigID        *uint
	PreviousBatchID *uint `gorm:"unique"`
	PreviousBatch   *Batch
	Executions      []Execution `gorm:"foreignKey:BatchID"`
}
