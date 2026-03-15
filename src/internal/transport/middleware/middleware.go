package middleware

import (
	"suscord/internal/config"
	"suscord/internal/domain/storage"
)

type Middleware struct {
	config  *config.Config
	storage storage.Storage
}

func NewMiddleware(cfg *config.Config, storage storage.Storage) *Middleware {
	return &Middleware{
		config:  cfg,
		storage: storage,
	}
}
