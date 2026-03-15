package storage

import (
	"context"
	"suscord/internal/domain/entity"
)

type UserStorage interface {
	GetByID(ctx context.Context, id uint) (entity.User, error)
	GetByUsername(ctx context.Context, username string) (entity.UnsafeUser, error)
	SearchUsers(ctx context.Context, exceptUserID uint, username string) ([]entity.User, error)
	Create(ctx context.Context, user entity.UnsafeUser) error
	Update(ctx context.Context, userID uint, data map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
}
