package model

type User struct {
	ID         uint
	Username   string `gorm:"size:20"`
	Password   string `gorm:"type:text"`
	AvatarPath string `gorm:"size:255"`
}

func (u User) TableName() string {
	return "users"
}
