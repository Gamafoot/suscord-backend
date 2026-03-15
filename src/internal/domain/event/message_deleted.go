package event

const OnMessageDeleted = "chat.message.deleted"

type MessageDeleted struct {
	ChatID       uint
	MessageID    uint
	ExceptUserID uint
}

func NewMessageDeleted(chatID, messageID, exceptUserID uint) MessageDeleted {
	return MessageDeleted{
		ChatID:       chatID,
		MessageID:    messageID,
		ExceptUserID: exceptUserID,
	}
}

func (e MessageDeleted) EventName() string {
	return OnMessageDeleted
}
