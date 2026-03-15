package storage

import (
	"context"
	"suscord/internal/domain/entity"
)

type SessionStorage interface {
	GetByUUID(ctx context.Context, uuid string) (*entity.Session, error)
	GetByUserID(ctx context.Context, userID uint) (*entity.Session, error)
	Create(ctx context.Context, userID uint) (string, error)
	Update(ctx context.Context, userID uint) (string, error)
	Delete(ctx context.Context, uuid string) error
}
