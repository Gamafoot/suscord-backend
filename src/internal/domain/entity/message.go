package entity

import (
	"time"
)

type Message struct {
	ID          uint
	ChatID      uint
	User        User
	Type        string
	Content     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Attachments []Attachment
}

type GetMessagesInput struct {
	ChatID        uint
	UserID        uint
	LastMessageID uint
	Limit         int
}

type CreateMessageInput struct {
	Type    string
	Content string
	Files   []*File
}

type CreateMessageData struct {
	Type    string
	Content string
}

type UpdateMessageInput struct {
	Content string
}
