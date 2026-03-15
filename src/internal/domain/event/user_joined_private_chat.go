package event

import (
	"suscord/internal/domain/entity"
	"suscord/internal/domain/event/model"
)

const OnUserJoinedPrivateChat = "chat.private.user.joined"

type UserJoinedPrivateChat struct {
	ChatID uint       `json:"chat_id"`
	User   model.User `json:"user"`
}

func NewUserJoinedPrivateChat(chatID uint, user entity.User, mediaURL string) UserJoinedPrivateChat {
	return UserJoinedPrivateChat{
		ChatID: chatID,
		User:   model.NewUser(user, mediaURL),
	}
}

func (e UserJoinedPrivateChat) EventName() string {
	return OnUserJoinedPrivateChat
}
