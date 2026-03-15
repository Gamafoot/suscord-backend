package model

type Attachment struct {
	ID        uint
	MessageID uint
	FilePath  string `gorm:"varchar(255)"`
	FileSize  int64
	MimeType  string `gorm:"varchar(100)"`

	Message Message `gorm:"constraint:OnDelete:CASCADE"`
}

func (m *Attachment) TableName() string {
	return "message_attachments"
}
