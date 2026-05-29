package storage

import (
	"context"
	"suscord/internal/domain/entity"
)

type AttachmentStorage interface {
	GetByID(ctx context.Context, attachmentID uint) (entity.Attachment, error)
	GetByMessageID(ctx context.Context, messageID uint) ([]entity.Attachment, error)
	Create(ctx context.Context, messageID uint, input entity.CreateAttachmentInput) (uint, error)
	IsOwner(ctx context.Context, userID, attachmentID uint) (bool, error)
}
