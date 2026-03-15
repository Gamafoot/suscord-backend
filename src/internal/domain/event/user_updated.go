package event

import "suscord/internal/domain/entity"

const OnUserUpdated = "user.updated"

type UserUpdated struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	AvatarPath string `json:"avatar_path"`
}

func NewUserUpdated(user entity.User) UserUpdated {
	return UserUpdated{
		ID:         user.ID,
		Username:   user.Username,
		AvatarPath: user.AvatarPath,
	}
}

func (e UserUpdated) EventName() string {
	return OnUserUpdated
}
