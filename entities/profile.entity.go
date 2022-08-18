package entities

type Profile struct {
	ID      uint `gorm:"primaryKey"`
	Name    string
	Surname string
	UserID  *uint   `gorm:"column:userId;not null"`
	Batches []Batch `gorm:"foreignKey:ProfileID"`
}

func (Profile) TableName() string {
	return "profile"
}
