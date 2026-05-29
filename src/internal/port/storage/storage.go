package storage

type Storage interface {
	User() UserStorage
	Chat() ChatStorage
	ChatMember() ChatMemberStorage
	Message() MessageStorage
	Attachment() AttachmentStorage
	Session() SessionStorage
}
