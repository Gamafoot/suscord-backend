package storage

import (
	"context"
	"suscord/internal/domain/entity"
	derr "suscord/internal/domain/errors"
	"suscord/internal/infra/database/relational/model"

	pkgerr "github.com/pkg/errors"
	"gorm.io/gorm"
)

type messageStorage struct {
	db *gorm.DB
}

func NewMessageStorage(db *gorm.DB) *messageStorage {
	return &messageStorage{db: db}
}

func (s *messageStorage) FindMessages(
	ctx context.Context,
	chatID uint,
	lastMessageID uint,
	limit int,
) ([]*entity.Message, error) {
	messages := make([]*model.Message, 0)

	db := s.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Joins("User").
		Preload("Attachments")

	if lastMessageID != 0 {
		db = db.Where("id < ?", lastMessageID)
	}

	if err := db.Find(&messages).Error; err != nil {
		return nil, pkgerr.WithStack(err)
	}

	result := make([]*entity.Message, len(messages))
	for i, message := range messages {
		result[i] = messageModelToDomain(message)
	}

	return result, nil
}

func (s *messageStorage) GetByID(ctx context.Context, messageID uint) (*entity.Message, error) {
	message := new(model.Message)

	err := s.db.WithContext(ctx).
		Joins("User").
		Preload("Attachments").
		First(message, "messages.id = ?", messageID).Error

	if err != nil {
		if pkgerr.Is(err, gorm.ErrRecordNotFound) {
			return nil, derr.ErrRecordNotFound
		}
		return nil, pkgerr.WithStack(err)
	}

	return messageModelToDomain(message), nil
}

func (s *messageStorage) Exists(ctx context.Context, messageID uint) (bool, error) {
	var count int64

	err := s.db.WithContext(ctx).Find(&model.Message{}, "id = ?", messageID).Count(&count).Error
	if err != nil {
		return false, pkgerr.WithStack(err)
	}

	return count > 0, nil
}

func (s *messageStorage) Create(ctx context.Context, userID, chatID uint, payload entity.CreateMessageData) (uint, error) {
	message := &model.Message{
		UserID:  userID,
		ChatID:  chatID,
		Type:    payload.Type,
		Content: payload.Content,
	}
	if err := s.db.WithContext(ctx).Create(message).Error; err != nil {
		return 0, pkgerr.WithStack(err)
	}
	return message.ID, nil
}

func (s *messageStorage) Update(ctx context.Context, messageID uint, input entity.UpdateMessageInput) error {
	err := s.db.WithContext(ctx).Model(&model.Message{ID: messageID}).Update("content", input.Content).Error
	if err != nil {
		return pkgerr.WithStack(err)
	}
	return nil
}

func (s *messageStorage) Delete(ctx context.Context, messageID uint) error {
	if err := s.db.WithContext(ctx).Delete(&model.Message{ID: messageID}).Error; err != nil {
		return pkgerr.WithStack(err)
	}
	return nil
}

func (s *messageStorage) IsAuthor(ctx context.Context, userID, messageID uint) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Find(&model.Message{}, "user_id = ? AND id = ?", userID, messageID).
		Count(&count).Error

	if err != nil {
		return false, pkgerr.WithStack(err)
	}
	return count > 0, nil
}

func messageModelToDomain(message *model.Message) *entity.Message {
	attachments := make([]entity.Attachment, len(message.Attachments))
	for i, attachment := range attachments {
		attachments[i] = entity.Attachment{
			ID:        attachment.ID,
			MessageID: attachment.MessageID,
			FilePath:  attachment.FilePath,
			FileSize:  attachment.FileSize,
			MimeType:  attachment.MimeType,
		}
	}

	return &entity.Message{
		ID:     message.ID,
		ChatID: message.ChatID,
		User: entity.User{
			ID:         message.User.ID,
			Username:   message.User.Username,
			AvatarPath: message.User.AvatarPath,
		},
		Type:        message.Type,
		Content:     message.Content,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
		Attachments: attachments,
	}
}
