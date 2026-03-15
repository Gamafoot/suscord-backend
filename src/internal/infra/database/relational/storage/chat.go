package storage

import (
	"context"
	"strings"
	"suscord/internal/domain/entity"
	"suscord/internal/infra/database/relational/model"

	pkgerr "github.com/pkg/errors"
	"gorm.io/gorm"
)

type chatStorage struct {
	db *gorm.DB
}

func NewChatStorage(db *gorm.DB) *chatStorage {
	return &chatStorage{db: db}
}

func (s *chatStorage) GetByID(ctx context.Context, chatID uint) (entity.Chat, error) {
	chat := new(model.Chat)
	if err := s.db.WithContext(ctx).First(&chat, "id = ?", chatID).Error; err != nil {
		return entity.Chat{}, pkgerr.WithStack(err)
	}
	return chatModelToDomain(*chat), nil
}

func (s *chatStorage) GetUserChat(ctx context.Context, chatID, userID uint) (entity.Chat, error) {
	chat := new(model.Chat)

	err := s.db.WithContext(ctx).
		Raw(userChatsQuery(`
			AND c.id = ?
		`), userID, userID, userID, chatID).
		Scan(chat).
		Error
	if err != nil {
		return entity.Chat{}, pkgerr.WithStack(err)
	}

	return chatModelToDomain(*chat), nil
}

func (s *chatStorage) GetUserChats(ctx context.Context, userID uint) ([]entity.Chat, error) {
	chats := make([]model.Chat, 0)

	err := s.db.WithContext(ctx).
		Raw(userChatsQuery(""), userID, userID, userID).
		Scan(&chats).
		Error
	if err != nil {
		return nil, pkgerr.WithStack(err)
	}

	chatDomains := make([]entity.Chat, len(chats))

	for i, chat := range chats {
		chatDomains[i] = chatModelToDomain(chat)
	}

	return chatDomains, nil
}

func (s *chatStorage) SearchUserChats(ctx context.Context, userID uint, searchPattern string) ([]entity.Chat, error) {
	chats := make([]model.Chat, 0)

	err := s.db.WithContext(ctx).
		Raw(userChatsQuery(`
			AND (
				CASE
					WHEN c.type = 'private' THEN COALESCE((
						SELECT u.username
						FROM users u
						INNER JOIN chat_members cm2 ON cm2.user_id = u.id
						WHERE cm2.chat_id = c.id
						  AND cm2.user_id != ?
						LIMIT 1
					), '')
					ELSE COALESCE(c.name, '')
				END
			) LIKE LOWER(?)
			ESCAPE '\'
		`), userID, userID, userID, userID, "%"+escapeLike(searchPattern)+"%").
		Scan(&chats).Error
	if err != nil {
		return nil, pkgerr.WithStack(err)
	}

	chatDomains := make([]entity.Chat, len(chats))

	for i, chat := range chats {
		chatDomains[i] = chatModelToDomain(chat)
	}

	return chatDomains, nil
}

func (s *chatStorage) Create(ctx context.Context, input entity.CreateChatInput) (uint, error) {
	chat := &model.Chat{
		Type:       input.Type,
		Name:       input.Name,
		AvatarPath: input.AvatarPath,
	}

	if err := s.db.WithContext(ctx).Create(chat).Error; err != nil {
		return 0, pkgerr.WithStack(err)
	}

	return chat.ID, nil
}

func (s *chatStorage) Update(ctx context.Context, chatID uint, data map[string]any) error {
	err := s.db.WithContext(ctx).Model(&model.Chat{}).Where("id = ?", chatID).Updates(data).Error
	if err != nil {
		return pkgerr.WithStack(err)
	}

	return nil
}

func (s *chatStorage) Delete(ctx context.Context, chatID uint) error {
	if err := s.db.WithContext(ctx).Delete(&entity.Chat{ID: chatID}).Error; err != nil {
		return pkgerr.WithStack(err)
	}
	return nil
}

func chatModelToDomain(chat model.Chat) entity.Chat {
	return entity.Chat{
		ID:         chat.ID,
		Name:       chat.Name,
		AvatarPath: chat.AvatarPath,
		Type:       chat.Type,
	}
}

func userChatsQuery(extraWhere string) string {
	return `
	SELECT
		c.id,
		CASE
			WHEN c.type = 'private' THEN (
				SELECT u.username
				FROM users u
				INNER JOIN chat_members cm2 ON cm2.user_id = u.id
				WHERE cm2.chat_id = c.id
				  AND cm2.user_id != ?
				LIMIT 1
			)
			ELSE c.name
		END AS name,
		CASE
			WHEN c.type = 'private' THEN (
				SELECT u.avatar_path
				FROM users u
				INNER JOIN chat_members cm2 ON cm2.user_id = u.id
				WHERE cm2.chat_id = c.id
				  AND cm2.user_id != ?
				LIMIT 1
			)
			ELSE c.avatar_path
		END AS avatar_path,
		c.type
	FROM chats c
	INNER JOIN chat_members cm ON cm.chat_id = c.id
	WHERE cm.user_id = ?
	` + extraWhere + `
	ORDER BY (
		SELECT m.created_at
		FROM messages m
		WHERE m.chat_id = c.id
		  AND m.user_id = cm.user_id
		ORDER BY m.created_at DESC
		LIMIT 1
	) DESC,
	c.id DESC
	`
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	value = strings.ReplaceAll(value, `_`, `\_`)
	return value
}
