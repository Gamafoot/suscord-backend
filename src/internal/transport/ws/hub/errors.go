package hub

import "errors"

var (
	ErrNotInCall         = errors.New("you are not in a call")
	ErrAlreadyInCall     = errors.New("you are already in this call")
	ErrInvalidCallRoomID = errors.New("chat_id does not match the current call")
)
