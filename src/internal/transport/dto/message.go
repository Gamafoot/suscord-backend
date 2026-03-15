package dto

import (
	"suscord/internal/domain/entity"
	"time"
)

type Message struct {
	ID          uint         `json:"id"`
	ChatID      uint         `json:"chat_id"`
	User        User         `json:"user_id"`
	Type        string       `json:"type"`
	Content     string       `json:"content"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Attachments []Attachment `json:"attachments"`
}

func NewMessage(message *entity.Message, mediaURL string) Message {
	attachments := make([]Attachment, len(message.Attachments))
	for i, attachment := range message.Attachments {
		attachments[i] = NewAttachmentResponse(attachment, mediaURL)
	}

	return Message{
		ID:          message.ID,
		ChatID:      message.ChatID,
		User:        NewUser(message.User, mediaURL),
		Type:        message.Type,
		Content:     message.Content,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
		Attachments: attachments,
	}
}

func NewMessages(messages []*entity.Message, mediaURL string) []Message {
	result := make([]Message, len(messages))
	for i, message := range messages {
		result[i] = NewMessage(message, mediaURL)
	}
	return result
}
