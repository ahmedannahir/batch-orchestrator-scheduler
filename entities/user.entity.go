package entities

type User struct {
	ID          uint `gorm:"primaryKey"`
	Email       string
	PhoneNumber string `gorm:"column:phoneNumber"`
	Password    string
	Profile     Profile `gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "user"
}
