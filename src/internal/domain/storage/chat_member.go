package storage

import (
	"context"
	"suscord/internal/domain/entity"
)

type ChatMemberStorage interface {
	GetChatMembers(ctx context.Context, chatID uint) ([]entity.User, error)
	GetNonMembers(ctx context.Context, chatID uint) ([]entity.User, error)
	GetPrivateChatID(ctx context.Context, userID, friendID uint) (uint, error)
	AddUserToChat(ctx context.Context, userID, chatID uint) error
	IsMemberOfChat(ctx context.Context, userID, chatID uint) (bool, error)
	Delete(ctx context.Context, userID, chatID uint) error
}
