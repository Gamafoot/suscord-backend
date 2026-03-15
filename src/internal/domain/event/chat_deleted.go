package event

const OnChatDeleted = "chat.deleted"

type ChatDeleted struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func NewChatDeleted(id uint, name string) ChatDeleted {
	return ChatDeleted{ID: id, Name: name}
}

func (e ChatDeleted) EventName() string {
	return OnChatDeleted
}
