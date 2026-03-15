package model

type Session struct {
	UUID   string `gorm:"primaryKey;type:varchar(255)"`
	UserID uint
	User   User
}

func (s *Session) TableName() string {
	return "sessions"
}
