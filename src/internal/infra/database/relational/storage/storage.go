package storage

import (
	dstorage "suscord/internal/domain/storage"

	"gorm.io/gorm"
)

type storage struct {
	user       *userStorage
	chat       *chatStorage
	chatMember *chatMemberStorage
	message    *messageStorage
	attachment *attachmentStorage
	session    *sessionStorage
}

func NewGormStorage(db *gorm.DB) dstorage.Storage {
	return &storage{
		user:       NewUserStorage(db),
		chat:       NewChatStorage(db),
		chatMember: NewChatMemberStorage(db),
		message:    NewMessageStorage(db),
		attachment: NewAttachmentStorage(db),
		session:    NewSessionStorage(db),
	}
}

func (s *storage) User() dstorage.UserStorage {
	return s.user
}

func (s *storage) Chat() dstorage.ChatStorage {
	return s.chat
}

func (s *storage) ChatMember() dstorage.ChatMemberStorage {
	return s.chatMember
}

func (s *storage) Message() dstorage.MessageStorage {
	return s.message
}

func (s *storage) Attachment() dstorage.AttachmentStorage {
	return s.attachment
}

func (s *storage) Session() dstorage.SessionStorage {
	return s.session
}
