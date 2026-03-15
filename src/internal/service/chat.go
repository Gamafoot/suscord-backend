package service

import (
	"context"
	"suscord/internal/config"
	"suscord/internal/domain/entity"
	derr "suscord/internal/domain/errors"
	"suscord/internal/domain/event"
	"suscord/internal/domain/eventbus"
	"suscord/internal/domain/storage"
	file "suscord/internal/infra/file_manager"

	pkgerr "github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type ChatService interface {
	GetUserChats(ctx context.Context, userID uint, searchPattern string) ([]entity.Chat, error)
	GetOrCreatePrivateChat(ctx context.Context, userID, friendID uint) (entity.Chat, error)
	CreateGroupChat(ctx context.Context, userID uint, input entity.CreateGroupChatInput) (entity.Chat, error)
	UpdateGroupChat(ctx context.Context, userID, chatID uint, input entity.UpdateChatInput) (entity.Chat, error)
	DeletePrivateChat(ctx context.Context, userID, chatID uint) error
}

type chatService struct {
	cfg         *config.Config
	storage     storage.Storage
	eventbus    eventbus.EventBus
	fileManager file.FileManager
	logger      *zap.SugaredLogger
}

func NewChatService(
	cfg *config.Config,
	storage storage.Storage,
	eventbus eventbus.EventBus,
	fileManager file.FileManager,
	logger *zap.SugaredLogger,
) *chatService {
	return &chatService{
		cfg:         cfg,
		storage:     storage,
		eventbus:    eventbus,
		fileManager: fileManager,
		logger:      logger,
	}
}

func (s *chatService) GetUserChats(ctx context.Context, userID uint, searchPattern string) ([]entity.Chat, error) {
	log := s.logger.With(
		"user_id", userID,
		"search_pattern", searchPattern,
	)

	if searchPattern != "" {
		chats, err := s.storage.Chat().SearchUserChats(ctx, userID, searchPattern)
		if err != nil {
			return nil, err
		}

		log.Info("found chats with search pattern")

		return chats, nil
	}

	chats, err := s.storage.Chat().GetUserChats(ctx, userID)
	if err != nil {
		return nil, err
	}

	log.Info("get list chats")

	return chats, nil
}

func (s *chatService) GetOrCreatePrivateChat(ctx context.Context, userID, friendID uint) (entity.Chat, error) {
	createChat := false
	empty := entity.Chat{}

	log := s.logger.With(
		"user_id", userID,
		"friend_id", friendID,
	)

	chatID, err := s.storage.ChatMember().GetPrivateChatID(ctx, userID, friendID)
	if err != nil {
		if pkgerr.Is(err, derr.ErrRecordNotFound) {
			chatID, err = s.createPrivateChat(ctx, userID, friendID)
			if err != nil {
				return empty, err
			}
			createChat = true
		} else {
			return empty, err
		}
	}

	chat, err := s.storage.Chat().GetUserChat(ctx, chatID, userID)
	if err != nil {
		return empty, err
	}

	if createChat {
		user, err := s.storage.User().GetByID(ctx, userID)
		if err != nil {
			return empty, err
		}

		payload := event.NewUserJoinedPrivateChat(chatID, user, s.cfg.Static.URL)

		s.eventbus.Publish(
			payload.EventName(),
			payload,
		)

		user, err = s.storage.User().GetByID(ctx, friendID)
		if err != nil {
			return empty, err
		}

		log.Infow(
			"private chat created",
			"id", chatID,
		)

		payload = event.NewUserJoinedPrivateChat(chatID, user, s.cfg.Static.URL)
		s.eventbus.Publish(payload.EventName(), payload)
	}

	log.Infow(
		"get private chat",
		"id", chatID,
	)

	return chat, nil
}

func (s *chatService) createPrivateChat(ctx context.Context, userID, friendID uint) (uint, error) {
	chatID, err := s.storage.Chat().Create(ctx, entity.CreateChatInput{Type: "private"})
	if err != nil {
		return 0, err
	}

	err = s.storage.ChatMember().AddUserToChat(ctx, userID, chatID)
	if err != nil {
		return 0, err
	}

	err = s.storage.ChatMember().AddUserToChat(ctx, friendID, chatID)
	if err != nil {
		return 0, err
	}

	return chatID, nil
}

func (s *chatService) CreateGroupChat(
	ctx context.Context,
	userID uint,
	input entity.CreateGroupChatInput,
) (entity.Chat, error) {
	empty := entity.Chat{}

	log := s.logger.With(
		"user_id", userID,
	)

	var (
		filepath string
		err      error
	)

	if input.File != nil {
		log.Debugw(
			"uploading avatar for new chat",
			"filename", input.File.Name,
			"size", input.File.Size,
		)

		filepath, err = s.fileManager.Upload(input.File, "chats/avatars")
		if err != nil {
			return empty, pkgerr.WithStack(err)
		}
	}

	chatID, err := s.storage.Chat().Create(ctx, entity.CreateChatInput{
		Type:       "group",
		Name:       input.Name,
		AvatarPath: filepath,
	})
	if err != nil {
		if errDel := s.fileManager.Delete(filepath); errDel != nil {
			log.Errorw(
				"failed to rollback avatar upload",
				"path", filepath,
				"err", errDel,
			)
			err = multierr.Append(err, errDel)
		}
		return empty, err
	}

	err = s.storage.ChatMember().AddUserToChat(ctx, userID, chatID)
	if err != nil {
		return empty, err
	}

	chat, err := s.storage.Chat().GetByID(ctx, chatID)
	if err != nil {
		return empty, err
	}

	log.Infow(
		"group chat was created",
		"chat.id", chat.ID,
		"chat.name", chat.Name,
		"chat.avatar_path", chat.AvatarPath,
		"chat.type", chat.Type,
	)

	return chat, nil
}

func (s *chatService) UpdateGroupChat(
	ctx context.Context,
	userID uint,
	chatID uint,
	input entity.UpdateChatInput,
) (entity.Chat, error) {
	empty := entity.Chat{}

	log := s.logger.With(
		"user_id", userID,
		"chat_id", chatID,
	)

	var (
		filepath string
		err      error
	)

	if input.File != nil {
		log.Debugw(
			"uploading new avatar",
			"filename", input.File.Name,
			"size", input.File.Size,
		)

		filepath, err = s.fileManager.Upload(input.File, "chats/avatars")
		if err != nil {
			return empty, pkgerr.WithStack(err)
		}
	}

	data := make(map[string]any)

	if input.Name != nil {
		data["name"] = input.Name
	}
	if len(filepath) > 0 {
		data["avatar_path"] = filepath
	}

	err = s.storage.Chat().Update(ctx, chatID, data)
	if err != nil {
		if errDel := s.fileManager.Delete(filepath); errDel != nil {
			log.Errorw(
				"failed to rollback avatar upload",
				"path", filepath,
				"err", errDel,
			)
			err = multierr.Append(err, errDel)
		}
		return empty, err
	}

	chat, err := s.storage.Chat().GetByID(ctx, chatID)
	if err != nil {
		return empty, err
	}

	log.Infow(
		"chat was updated",
		"chat.id", chat.ID,
		"chat.name", chat.Name,
		"chat.avatar_path", chat.AvatarPath,
		"chat.type", chat.Type,
	)

	payload := event.NewGroupChatUpdated(chat, s.cfg.Media.Url)
	s.eventbus.Publish(payload.EventName(), payload)

	return chat, nil
}

func (s *chatService) DeletePrivateChat(ctx context.Context, userID, chatID uint) error {
	ok, err := s.storage.ChatMember().IsMemberOfChat(ctx, userID, chatID)
	if err != nil {
		return err
	}

	if !ok {
		return derr.ErrForbidden
	}

	log := s.logger.With(
		"user_id", userID,
		"chat_id", chatID,
	)

	chat, err := s.storage.Chat().GetByID(ctx, chatID)
	if err != nil {
		return err
	}

	if chat.Type != "private" {
		return derr.ErrForbidden
	}

	if err = s.storage.Chat().Delete(ctx, chatID); err != nil {
		return err
	}

	if err = s.fileManager.Delete(chat.AvatarPath); err != nil {
		log.Errorw("fail to delete chat avatar", "err", err)
	}

	payload := event.NewChatDeleted(chatID, chat.Name)
	s.eventbus.Publish(payload.EventName(), payload)

	return nil
}
