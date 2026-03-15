package dto

import (
	"suscord/internal/domain/entity"
	"suscord/pkg/urlpath"
)

type Chat struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
}

func NewChat(chat entity.Chat, mediaURL string) Chat {
	return Chat{
		ID:        chat.ID,
		Type:      chat.Type,
		Name:      chat.Name,
		AvatarUrl: urlpath.GetMediaURL(mediaURL, chat.AvatarPath),
	}
}

func NewChats(chats []entity.Chat, mediaURL string) []Chat {
	result := make([]Chat, len(chats))
	for i, chat := range chats {
		result[i] = NewChat(chat, mediaURL)
	}
	return result
}
