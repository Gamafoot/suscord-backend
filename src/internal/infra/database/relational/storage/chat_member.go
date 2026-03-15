package storage

import (
	"context"
	"suscord/internal/domain/entity"
	domainErrors "suscord/internal/domain/errors"
	"suscord/internal/infra/database/relational/model"

	pkgerr "github.com/pkg/errors"
	"gorm.io/gorm"
)

type chatMemberStorage struct {
	db *gorm.DB
}

func NewChatMemberStorage(db *gorm.DB) *chatMemberStorage {
	return &chatMemberStorage{db: db}
}

func (s *chatMemberStorage) GetChatMembers(ctx context.Context, chatID uint) ([]entity.User, error) {
	users := make([]*model.User, 0)

	sql := `
	SELECT *
	FROM users
	WHERE id IN (
		SELECT user_id 
		FROM chat_members 
		WHERE chat_id = ?
	);`

	if err := s.db.WithContext(ctx).Raw(sql, chatID).Scan(&users).Error; err != nil {
		if pkgerr.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrRecordNotFound
		}
		return nil, pkgerr.WithStack(err)
	}

	domainUsers := make([]entity.User, len(users))
	for i, user := range users {
		domainUsers[i] = userModelToEntity(*user)
	}

	return domainUsers, nil
}

func (s *chatMemberStorage) GetNonMembers(ctx context.Context, chatID uint) ([]entity.User, error) {
	users := make([]model.User, 0)

	sql := `
	SELECT *
	FROM users
	WHERE id NOT IN (
		SELECT user_id 
		FROM chat_members 
		WHERE chat_id = ?
	);`

	if err := s.db.WithContext(ctx).Raw(sql, chatID).Scan(&users).Error; err != nil {
		if pkgerr.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrRecordNotFound
		}
		return nil, pkgerr.WithStack(err)
	}

	domainUsers := make([]entity.User, len(users))
	for i, user := range users {
		domainUsers[i] = userModelToEntity(user)
	}

	return domainUsers, nil
}

func (s *chatMemberStorage) AddUserToChat(ctx context.Context, userID, chatID uint) error {
	member := &entity.ChatMember{
		ChatID: chatID,
		UserID: userID,
	}

	if err := s.db.WithContext(ctx).Create(member).Error; err != nil {
		return pkgerr.WithStack(err)
	}
	return nil
}

func (s *chatMemberStorage) GetPrivateChatID(ctx context.Context, userID, friendID uint) (uint, error) {
	var chatID uint

	sql := `
		SELECT c.id
		FROM chats c
		INNER JOIN chat_members cm1 ON c.id = cm1.chat_id
		INNER JOIN chat_members cm2 ON c.id = cm2.chat_id
		WHERE c.type = 'private'
		AND cm1.user_id = ?
		AND cm2.user_id = ?
	`

	if err := s.db.WithContext(ctx).Raw(sql, userID, friendID).Scan(&chatID).Error; err != nil {
		return 0, pkgerr.WithStack(err)
	}

	if chatID == 0 {
		return 0, domainErrors.ErrRecordNotFound
	}

	return chatID, nil
}

func (s *chatMemberStorage) IsMemberOfChat(ctx context.Context, userID, chatID uint) (bool, error) {
	if err := s.db.WithContext(ctx).First(&entity.ChatMember{}, "chat_id = ? AND user_id = ?", chatID, userID).Error; err != nil {
		if pkgerr.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, pkgerr.WithStack(err)
	}
	return true, nil
}

func (s *chatMemberStorage) Delete(ctx context.Context, userID, chatID uint) error {
	if err := s.db.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chatID, userID).Delete(&entity.ChatMember{}).Error; err != nil {
		return pkgerr.WithStack(err)
	}
	return nil
}

func chatMemberEntityToModel(chatMember entity.ChatMember) model.ChatMember {
	return model.ChatMember{
		ID:     chatMember.ID,
		ChatID: chatMember.ChatID,
		UserID: chatMember.UserID,
	}
}
