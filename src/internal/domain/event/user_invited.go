package event

const OnUserInvited = "chat.user.invited"

type UserInvited struct {
	ChatID uint   `json:"chat_id"`
	UserID uint   `json:"user_id"`
	Code   string `json:"code"`
}

func NewUserInvited(chatID, userID uint, code string) UserInvited {
	return UserInvited{
		ChatID: chatID,
		UserID: userID,
		Code:   code,
	}
}

func (e UserInvited) EventName() string {
	return OnUserInvited
}
