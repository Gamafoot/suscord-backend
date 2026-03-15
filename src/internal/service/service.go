package service

import (
	"suscord/internal/config"
	"suscord/internal/domain/eventbus"
	"suscord/internal/domain/storage"
	file "suscord/internal/infra/file_manager"

	"suscord/internal/domain/cache"

	"go.uber.org/zap"
)

type Service interface {
	User() UserService
	Auth() AuthService
	Chat() ChatService
	ChatMember() ChatMemberService
	Message() MessageService
}

type service struct {
	user       *userService
	auth       *authService
	chat       *chatService
	chatMember *chatMemberService
	message    *messageService
}

type ServiceConfig struct {
	AppConfig   *config.Config
	Storage     storage.Storage
	Cache       cache.Cache
	Eventbus    eventbus.EventBus
	FileManager file.FileManager
	Logger      *zap.SugaredLogger
}

func NewService(cfg ServiceConfig) *service {
	return &service{
		user:       NewUserService(cfg.Storage, cfg.FileManager, cfg.Eventbus, cfg.Logger),
		auth:       NewAuthService(cfg.AppConfig, cfg.Storage, cfg.Logger),
		chat:       NewChatService(cfg.AppConfig, cfg.Storage, cfg.Eventbus, cfg.FileManager, cfg.Logger),
		chatMember: NewChatMemberService(cfg.AppConfig, cfg.Storage, cfg.Cache, cfg.Eventbus, cfg.Logger),
		message:    NewMessageService(cfg.AppConfig, cfg.Storage, cfg.Eventbus, cfg.FileManager, cfg.Logger),
	}
}

func (s *service) User() UserService {
	return s.user
}

func (s *service) Auth() AuthService {
	return s.auth
}

func (s *service) Chat() ChatService {
	return s.chat
}

func (s *service) ChatMember() ChatMemberService {
	return s.chatMember
}

func (s *service) Message() MessageService {
	return s.message
}
