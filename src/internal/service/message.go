package service

import (
	"context"
	"mime"
	"path/filepath"
	"suscord/internal/config"
	"suscord/internal/domain/entity"
	domainErrors "suscord/internal/domain/errors"
	"suscord/internal/domain/event"
	"suscord/internal/domain/eventbus"
	"suscord/internal/domain/storage"
	file "suscord/internal/infra/file_manager"

	"go.uber.org/zap"
)

type MessageService interface {
	GetChatMessages(ctx context.Context, input entity.GetMessagesInput) ([]*entity.Message, error)
	Create(ctx context.Context, userID, chatID uint, input entity.CreateMessageInput) (*entity.Message, error)
	Update(ctx context.Context, userID, messageID uint, input entity.UpdateMessageInput) (*entity.Message, error)
	Delete(ctx context.Context, userID, messageID uint) error
}

type messageService struct {
	cfg         *config.Config
	storage     storage.Storage
	eventbus    eventbus.EventBus
	fileManager file.FileManager
	logger      *zap.SugaredLogger
}

func NewMessageService(
	cfg *config.Config,
	storage storage.Storage,
	eventbus eventbus.EventBus,
	fileManager file.FileManager,
	logger *zap.SugaredLogger,
) *messageService {
	return &messageService{
		cfg:         cfg,
		storage:     storage,
		eventbus:    eventbus,
		fileManager: fileManager,
		logger:      logger,
	}
}

func (s *messageService) GetChatMessages(ctx context.Context, input entity.GetMessagesInput) ([]*entity.Message, error) {
	log := s.logger.With(
		"user_id", input.UserID,
		"chat_id", input.ChatID,
		"last_message_id", input.LastMessageID,
		"limit", input.Limit,
	)

	ok, err := s.storage.ChatMember().IsMemberOfChat(ctx, input.UserID, input.ChatID)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, domainErrors.ErrUserIsNotMemberOfChat
	}

	messages, err := s.storage.Message().FindMessages(ctx, input.ChatID, input.LastMessageID, input.Limit)
	if err != nil {
		return nil, err
	}

	log.Infow(
		"get chat messages",
		"count", len(messages),
	)

	return messages, nil
}

func (s *messageService) Create(
	ctx context.Context,
	userID uint,
	chatID uint,
	input entity.CreateMessageInput,
) (*entity.Message, error) {
	log := s.logger.With(
		"user_id", userID,
		"chat_id", chatID,
		"attachments", len(input.Files),
	)

	data := entity.CreateMessageData{
		Type:    input.Type,
		Content: input.Content,
	}

	messageID, err := s.storage.Message().Create(ctx, userID, chatID, data)
	if err != nil {
		return nil, err
	}

	if len(input.Files) > 0 {
		log.Debugw("uploading attachments")

		err := s.createAttachments(ctx, messageID, input.Files)
		if err != nil {
			return nil, err
		}
	}

	log = log.With(
		"message_id", messageID,
	)

	message, err := s.storage.Message().GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	log.Infow("message created")

	payload := event.NewMessageCreated(message, s.cfg.Media.Url)
	s.eventbus.Publish(payload.EventName(), payload)

	return message, nil
}

func (s *messageService) Update(ctx context.Context, userID, messageID uint, data entity.UpdateMessageInput) (*entity.Message, error) {
	log := s.logger.With(
		"user_id", userID,
		"message_id", messageID,
	)

	ok, err := s.storage.Message().IsAuthor(ctx, userID, messageID)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, domainErrors.ErrUserIsNotMemberOfChat
	}

	err = s.storage.Message().Update(ctx, messageID, data)
	if err != nil {
		return nil, err
	}

	message, err := s.storage.Message().GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	log.Infow(
		"message updated",
		"chat_id", message.ChatID,
	)

	payload := event.NewMessageUpdated(message, s.cfg.Media.Url)
	s.eventbus.Publish(payload.EventName(), payload)

	return message, nil
}

func (s *messageService) Delete(ctx context.Context, userID, messageID uint) error {
	log := s.logger.With(
		"user_id", userID,
		"message_id", messageID,
	)

	ok, err := s.storage.Message().IsAuthor(ctx, userID, messageID)
	if err != nil {
		return err
	}

	if !ok {
		return domainErrors.ErrUserIsNotMemberOfChat
	}

	message, err := s.storage.Message().GetByID(ctx, messageID)
	if err != nil {
		return err
	}

	err = s.storage.Message().Delete(ctx, messageID)
	if err != nil {
		return err
	}

	log.Infow(
		"message deleted",
		"chat_id", message.ChatID,
	)

	payload := event.NewMessageDeleted(message.ChatID, message.ID, userID)
	s.eventbus.Publish(payload.EventName(), payload)

	return nil
}

func (s *messageService) createAttachments(ctx context.Context, messageID uint, files []*entity.File) error {
	log := s.logger.With("message_id", messageID)

	for _, file := range files {
		mimetype := mime.TypeByExtension(filepath.Ext(file.Name))

		filepath, err := s.fileManager.Upload(file, "messages")
		if err != nil {
			return err
		}

		log.Debugw(
			"attachment uploaded",
			"path", filepath,
			"mime", mimetype,
		)

		input := entity.CreateAttachmentInput{
			FilePath: filepath,
			FileSize: file.Size,
			MimeType: mimetype,
		}

		_, err = s.storage.Attachment().Create(ctx, messageID, input)
		if err != nil {
			return err
		}
	}

	return nil
}
