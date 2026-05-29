package storage

import (
	"context"
	"suscord/internal/domain/entity"
)

type MessageStorage interface {
	FindMessages(ctx context.Context, chatID, lastMessageID uint, limit int) ([]*entity.Message, error)
	GetByID(ctx context.Context, messageID uint) (*entity.Message, error)
	Create(ctx context.Context, userID, chatID uint, payload entity.CreateMessageData) (uint, error)
	Update(ctx context.Context, messageID uint, input entity.UpdateMessageInput) error
	Delete(ctx context.Context, messageID uint) error
	Exists(ctx context.Context, messageID uint) (bool, error)
	IsAuthor(ctx context.Context, userID, messageID uint) (bool, error)
}
