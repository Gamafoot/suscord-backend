package service

import (
	"context"
	"suscord/internal/config"
	"suscord/internal/domain/cache"
	"suscord/internal/domain/entity"
	derr "suscord/internal/domain/errors"
	"suscord/internal/domain/event"
	"suscord/internal/domain/eventbus"
	"suscord/internal/domain/storage"
	"time"

	"github.com/google/uuid"
	pkgerr "github.com/pkg/errors"
	"go.uber.org/zap"
)

type ChatMemberService interface {
	GetNonMembers(ctx context.Context, chatID, memberID uint) ([]entity.User, error)
	GetNotChatMembers(ctx context.Context, chatID, memberID uint) ([]entity.User, error)
	IsMemberOfChat(ctx context.Context, userID, chatID uint) (bool, error)
	SendInvite(ctx context.Context, ownerID, chatID, userID uint) error
	AcceptInvite(ctx context.Context, userID uint, code string) error
	LeaveFromChat(ctx context.Context, chatID, userID uint) error
}

type chatMemberService struct {
	cfg      *config.Config
	storage  storage.Storage
	cache    cache.Cache
	eventbus eventbus.EventBus
	logger   *zap.SugaredLogger
}

func NewChatMemberService(
	cfg *config.Config,
	storage storage.Storage,
	cache cache.Cache,
	eventbus eventbus.EventBus,
	logger *zap.SugaredLogger,
) *chatMemberService {
	return &chatMemberService{
		cfg:      cfg,
		storage:  storage,
		cache:    cache,
		eventbus: eventbus,
		logger:   logger,
	}
}

func (s *chatMemberService) IsMemberOfChat(ctx context.Context, userID, chatID uint) (bool, error) {
	return s.storage.ChatMember().IsMemberOfChat(ctx, userID, chatID)
}

func (s *chatMemberService) GetNonMembers(ctx context.Context, chatID, memberID uint) ([]entity.User, error) {
	log := s.logger.With(
		"chat_id", chatID,
		"member_id", memberID,
	)

	ok, err := s.storage.ChatMember().IsMemberOfChat(ctx, memberID, chatID)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, derr.ErrUserIsNotMemberOfChat
	}

	users, err := s.storage.ChatMember().GetChatMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}

	log.Infow(
		"get chat members",
		"count", len(users),
	)

	return users, nil
}

func (s *chatMemberService) GetNotChatMembers(ctx context.Context, chatID, memberID uint) ([]entity.User, error) {
	log := s.logger.With(
		"chat_id", chatID,
		"member_id", memberID,
	)

	ok, err := s.storage.ChatMember().IsMemberOfChat(ctx, memberID, chatID)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, derr.ErrUserIsNotMemberOfChat
	}

	users, err := s.storage.ChatMember().GetNonMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}

	log.Infow(
		"get chat non-members",
		"count", len(users),
	)

	return users, nil
}

func (s *chatMemberService) SendInvite(ctx context.Context, ownerID, chatID, userID uint) error {
	log := s.logger.With(
		"owner_id", ownerID,
		"chat_id", chatID,
		"user_id", userID,
	)

	ok, err := s.storage.ChatMember().IsMemberOfChat(ctx, ownerID, chatID)
	if err != nil {
		return err
	}

	if !ok {
		return derr.ErrUserIsNotMemberOfChat
	}

	chat, err := s.storage.Chat().GetByID(ctx, chatID)
	if err != nil {
		return err
	}

	if chat.Type != "group" {
		return derr.ErrChatIsNotGroup
	}

	uuid := uuid.New()
	code := uuid.String()

	err = s.cache.Set(ctx, code, chatID, 10*time.Second)
	if err != nil {
		return err
	}

	log.Infow("invite created")

	payload := event.NewUserInvited(chatID, userID, code)
	s.eventbus.Publish(payload.EventName(), payload)

	return nil
}

func (s *chatMemberService) AcceptInvite(ctx context.Context, userID uint, code string) error {
	log := s.logger.With(
		"user_id", userID,
	)

	value, err := s.cache.Get(ctx, code)
	if err != nil {
		return err
	}

	chatID, ok := value.(uint)
	if !ok {
		return pkgerr.Errorf("invalid value in chatID: %v", chatID)
	}

	log = log.With(
		"chat_id", chatID,
	)

	err = s.storage.ChatMember().AddUserToChat(ctx, userID, chatID)
	if err != nil {
		return err
	}

	user, err := s.storage.User().GetByID(ctx, userID)
	if err != nil {
		return err
	}

	log.Infow("invite accepted")

	payload := event.NewUserJoinedGroupChat(chatID, user, s.cfg.Media.Url)
	s.eventbus.Publish(payload.EventName(), payload)

	return nil
}

func (s *chatMemberService) LeaveFromChat(ctx context.Context, chatID, userID uint) error {
	log := s.logger.With(
		"chat_id", chatID,
		"user_id", userID,
	)

	ok, err := s.storage.ChatMember().IsMemberOfChat(ctx, userID, chatID)
	if err != nil {
		return err
	}

	if !ok {
		return derr.ErrUserIsNotMemberOfChat
	}

	err = s.storage.ChatMember().Delete(ctx, userID, chatID)
	if err != nil {
		return err
	}

	log.Infow("left chat")

	payload := event.NewUserLeft(chatID, userID)
	s.eventbus.Publish(payload.EventName(), payload)

	return nil
}
