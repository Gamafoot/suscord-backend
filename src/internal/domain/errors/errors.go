package errors

import (
	"errors"
)

var (
	// Validation
	ErrIsNotDigit         = errors.New("value is not digit")
	ErrIsNotPositiveDigit = errors.New("value must be positive digit")

	// General
	ErrBadRequest   = errors.New("invalid request data")
	ErrForbidden    = errors.New("user have no permissions for the resource")
	ErrNoContextVar = errors.New("no context var")

	// Database
	ErrRecordNotFound = errors.New("record no found")
	ErrIsNotOwner     = errors.New("user is not owner")

	// Cache
	ErrKeyNotFound = errors.New("key not found")

	// Auth
	ErrLoginIsExists          = errors.New("login is exists")
	ErrInvalidLoginOrPassword = errors.New("invalid login or password")

	// Chat
	ErrChatFull              = errors.New("the chat is full")
	ErrJoinChat              = errors.New("fail to join chat")
	ErrUserIsNotMemberOfChat = errors.New("user is not member of chat")
	ErrChatIsNotGroup        = errors.New("chat is not group")

	// Media
	ErrInvalidFileExtention = errors.New("invalid file extiontion")
	ErrInvalidFile          = errors.New("invalid file")
	ErrIsNotImage           = errors.New("file must be image")

	// File
	ErrFileNotFound = errors.New("file no found")

	// Eventbus
	ErrInvalidPayload = errors.New("invalid payload")
)
