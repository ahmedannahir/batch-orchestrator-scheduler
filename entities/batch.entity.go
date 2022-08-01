package entities

type Batch struct {
	ID          uint `gorm:"primaryKey"`
	Name        string
	Description string
	Url         string
	ConfigID    uint
	Executions  []Execution `gorm:"foreignKey:BatchID"`
}
