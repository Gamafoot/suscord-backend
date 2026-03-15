package event

import (
	"suscord/internal/domain/entity"
	"suscord/internal/domain/event/model"
)

const OnMessageUpdated = "chat.message.updated"

type MessageUpdated struct {
	*model.Message
}

func NewMessageUpdated(message *entity.Message, mediaURL string) MessageUpdated {
	return MessageUpdated{
		Message: model.NewMessage(message, mediaURL),
	}
}

func (e MessageUpdated) EventName() string {
	return OnMessageUpdated
}
