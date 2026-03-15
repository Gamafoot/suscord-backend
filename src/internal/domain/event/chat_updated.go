package event

import (
	"suscord/internal/domain/entity"
	"suscord/internal/domain/event/model"
)

const OnGroupChatUpdated = "chat.group.updated"

type GroupChatUpdated struct {
	model.Chat
}

func NewGroupChatUpdated(chat entity.Chat, mediaURL string) GroupChatUpdated {
	return GroupChatUpdated{
		Chat: model.NewChat(chat, mediaURL),
	}
}

func (e GroupChatUpdated) EventName() string {
	return OnGroupChatUpdated
}
