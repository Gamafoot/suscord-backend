package dto

import (
	"suscord/internal/domain/entity"
	"suscord/pkg/urlpath"
)

type User struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
}

func NewUser(user entity.User, mediaURL string) User {
	return User{
		ID:        user.ID,
		Username:  user.Username,
		AvatarUrl: urlpath.GetMediaURL(mediaURL, user.AvatarPath),
	}
}

func NewUsers(users []entity.User, mediaURL string) []User {
	result := make([]User, len(users))
	for i, user := range users {
		result[i] = NewUser(user, mediaURL)
	}
	return result
}

type Me struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
}

func NewMe(user entity.User, mediaURL string) Me {
	return Me{
		ID:        user.ID,
		Username:  user.Username,
		AvatarUrl: user.AvatarPath,
	}
}
