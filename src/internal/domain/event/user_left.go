package event

const OnUserLeft = "chat.user.leave"

type UserLeft struct {
	ChatID uint `json:"chat_id"`
	UserID uint `json:"user_id"`
}

func NewUserLeft(chatID, userID uint) UserLeft {
	return UserLeft{
		ChatID: chatID,
		UserID: userID,
	}
}

func (e UserLeft) EventName() string {
	return OnUserLeft
}
