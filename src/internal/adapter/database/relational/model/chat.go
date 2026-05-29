package model

type Chat struct {
	ID         uint
	Name       string `gorm:"type:varchar(50)"`
	AvatarPath string `gorm:"type:varchar(255)"`
	Type       string `gorm:"type:varchar(20)"`

	Messages []*Message    `gorm:"constraint:OnDelete:CASCADE"`
	Members  []*ChatMember `gorm:"constraint:OnDelete:CASCADE"`
}

func (c Chat) TableName() string {
	return "chats"
}
