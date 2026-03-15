package model

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
