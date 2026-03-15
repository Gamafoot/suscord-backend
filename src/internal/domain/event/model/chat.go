package model

import (
	"suscord/internal/domain/entity"
	"suscord/pkg/urlpath"
)

type Chat struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func NewChat(chat entity.Chat, mediaURL string) Chat {
	return Chat{
		ID:        chat.ID,
		Name:      chat.Name,
		AvatarURL: urlpath.GetMediaURL(mediaURL, chat.AvatarPath),
	}
}
