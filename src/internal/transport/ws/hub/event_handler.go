package hub

import (
	"context"
	derr "suscord/internal/domain/errors"
	"suscord/internal/domain/event"
	"suscord/internal/domain/eventbus"
	"suscord/internal/transport/ws/hub/dto"

	pkgerr "github.com/pkg/errors"
)

func (h *hub) registerEvents(bus eventbus.EventBus) {
	bus.Subscribe(event.OnMessageCreated, h.onMessageCreated)
	bus.Subscribe(event.OnMessageUpdated, h.onMessageUpdated)
	bus.Subscribe(event.OnMessageDeleted, h.onMessageDeleted)
	bus.Subscribe(event.OnGroupChatUpdated, h.onGroupChatUpdated)
	bus.Subscribe(event.OnChatDeleted, h.onDeleteChat)
	bus.Subscribe(event.OnUserInvited, h.onUserInvited)
	bus.Subscribe(event.OnUserJoinedGroupChat, h.onUserJoinedGroupChat)
	bus.Subscribe(event.OnUserJoinedPrivateChat, h.onUserJoinedPrivateChat)
	bus.Subscribe(event.OnUserLeft, h.onUserLeft)
	bus.Subscribe(event.OnUserUpdated, h.onUserUpdated)
}

func (h *hub) onMessageCreated(ctx context.Context, payload any) error {
	log := h.logger.With("payload", payload)

	data, ok := payload.(event.MessageCreated)
	if !ok {
		return pkgerr.WithStack(derr.ErrInvalidPayload)
	}

	h.broadcastToChatRoom(data.ChatID, dto.ResponseMessage{
		Event:  data.EventName(),
		ChatID: data.ChatID,
		Data:   data,
	})

	log.Info("broadcasted " + data.EventName() + " event")

	return nil
}

func (h *hub) onMessageUpdated(ctx context.Context, payload any) error {
	data, ok := payload.(event.MessageUpdated)
	if !ok {
		return pkgerr.WithStack(derr.ErrInvalidPayload)
	}

	h.broadcastToChatRoomExcept(data.ChatID, data.User.ID, dto.ResponseMessage{
		Event:  data.EventName(),
		ChatID: data.ChatID,
		Data:   data,
	})

	return nil
}

func (h *hub) onMessageDeleted(ctx context.Context, payload any) error {
	data, ok := payload.(event.MessageDeleted)
	if !ok {
		return pkgerr.WithStack(derr.ErrInvalidPayload)
	}

	h.broadcastToChatRoomExcept(data.ChatID, data.ExceptUserID, dto.ResponseMessage{
		Event:  data.EventName(),
		ChatID: data.ChatID,
		Data:   data,
	})

	return nil
}

func (h *hub) onUserInvited(ctx context.Context, payload any) error {
	data, ok := payload.(event.UserInvited)
	if !ok {
		return pkgerr.WithStack(derr.ErrInvalidPayload)
	}

	if client, exists := h.getClient(data.UserID); exists {
		client.SendMessage(dto.ResponseMessage{
			Event: data.EventName(),
			Data: map[string]string{
				"code": data.Code,
			},
		})
	}

	return nil
}

func (h *hub) onGroupChatUpdated(ctx context.Context, payload any) error {
	data, ok := payload.(event.GroupChatUpdated)
	if !ok {
		return pkgerr.WithStack(derr.ErrInvalidPayload)
	}

	h.broadcastToChatRoom(data.Chat.ID, dto.ResponseMessage{
		Event: data.EventName(),
		Data:  data,
	})

	return nil
}

func (h *hub) onUserJoinedGroupChat(ctx context.Context, payload any) error {
	data, ok := payload.(event.UserJoinedGroupChat)
	if !ok {
		return pkgerr.WithStack(derr.ErrInvalidPayload)
	}

	if client, exists := h.getClient(data.User.ID); exists {
		h.joinChatRoom(data.ChatID, client)
		h.broadcastToChatRoomExcept(data.ChatID, data.User.ID, dto.ResponseMessage{
			Event:  data.EventName(),
			ChatID: data.ChatID,
			Data:   data,
		})
	}

	return nil
}

func (h *hub) onUserJoinedPrivateChat(ctx context.Context, payload any) error {
	data, ok := payload.(event.UserJoinedPrivateChat)
	if !ok {
		return pkgerr.WithStack(derr.ErrInvalidPayload)
	}

	if client, exists := h.getClient(data.User.ID); exists {
		h.joinChatRoom(data.ChatID, client)
		client.SendMessage(&dto.ResponseMessage{
			Event: data.EventName(),
			Data:  data,
		})
	}

	return nil
}

func (h *hub) onUserLeft(ctx context.Context, payload any) error {
	data, ok := payload.(event.UserLeft)
	if !ok {
		return pkgerr.WithStack(derr.ErrInvalidPayload)
	}

	h.leaveChatRoom(data.ChatID, data.UserID)

	h.broadcastToChatRoomExcept(data.ChatID, data.UserID, dto.ResponseMessage{
		Event:  data.EventName(),
		ChatID: data.ChatID,
		Data:   data,
	})

	return nil
}

func (h *hub) onUserUpdated(ctx context.Context, payload any) error {
	data, ok := payload.(event.UserUpdated)
	if !ok {
		return pkgerr.WithStack(derr.ErrInvalidPayload)
	}

	h.broadcastToAllExcept(data.ID, dto.ResponseMessage{
		Event: data.EventName(),
		Data:  data,
	})

	return nil
}

func (h *hub) onDeleteChat(ctx context.Context, payload any) error {
	data, ok := payload.(event.ChatDeleted)
	if !ok {
		return pkgerr.WithStack(derr.ErrInvalidPayload)
	}

	h.broadcastToChatRoom(data.ID, dto.ResponseMessage{
		Event: data.EventName(),
		Data:  data,
	})
	h.deleteChatRoom(data.ID)

	return nil
}
