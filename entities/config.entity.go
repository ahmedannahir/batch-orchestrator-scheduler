package entities

type Config struct {
	ID      uint `gorm:"primaryKey"`
	Name    string
	Url     string
	Batches []Batch `gorm:"foreignKey:ConfigID"`
}
