package event

import (
	"suscord/internal/domain/entity"
	"suscord/internal/domain/event/model"
)

const OnUserJoinedGroupChat = "chat.group.joined"

type UserJoinedGroupChat struct {
	ChatID uint       `json:"chat_id"`
	User   model.User `json:"user"`
}

func NewUserJoinedGroupChat(chatID uint, user entity.User, mediaURL string) UserJoinedGroupChat {
	return UserJoinedGroupChat{
		ChatID: chatID,
		User:   model.NewUser(user, mediaURL),
	}
}

func (e UserJoinedGroupChat) EventName() string {
	return OnUserJoinedGroupChat
}
