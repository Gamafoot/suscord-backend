package event

import (
	"suscord/internal/domain/entity"
	"suscord/internal/domain/event/model"
)

const OnMessageCreated = "chat.message.created"

type MessageCreated struct {
	*model.Message
}

func NewMessageCreated(message *entity.Message, mediaURL string) MessageCreated {
	return MessageCreated{
		Message: model.NewMessage(message, mediaURL),
	}
}

func (e MessageCreated) EventName() string {
	return OnMessageCreated
}
