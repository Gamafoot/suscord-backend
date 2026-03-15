package model

type ChatMember struct {
	ID     uint
	ChatID uint
	UserID uint

	Chat Chat `gorm:"foreignKey:ChatID"`
	User User `gorm:"foreignKey:UserID"`
}

func (c ChatMember) TableName() string {
	return "chat_members"
}
