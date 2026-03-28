package hub

import (
	gDTO "suscord/internal/transport/dto"
	"suscord/internal/transport/ws/hub/dto"
)

const (
	onError = "error"

	onCallInvite = "call.invite"
	onCallJoin   = "call.join"
	onCallLeave  = "call.leave"
	onRunDemo    = "call.demo.run"
	onStopDemo   = "call.demo.stop"
)

func (h *hub) handleClientMessage(client *HubClient, message *dto.ClientMessage) error {
	switch message.Event {
	case onCallJoin:
		h.joinCallRoom(message.ChatID, client)
		h.broadcastToChatRoomExcept(message.ChatID, client.user.ID, dto.ResponseMessage{
			Event: onCallJoin,
			Data: map[string]any{
				"chat_id": message.ChatID,
				"user":    gDTO.NewUser(client.user, h.cfg.Media.Url),
			},
		})

	case onCallLeave:
		if client.callRoomID == 0 {
			return ErrNotInCall
		}

		h.onLeaveCallRoom(message.ChatID, client)

	case onRunDemo, onStopDemo:
		if client.callRoomID == 0 {
			return ErrNotInCall
		}

		h.broadcastToChatRoomExcept(message.ChatID, client.user.ID, dto.ResponseMessage{
			Event: message.Event,
			Data:  map[string]uint{"chat_id": message.ChatID, "user_id": client.user.ID},
		})

	default:
		return client.SendMessage(&dto.ResponseMessage{
			Event: onError,
			Data: map[string]any{
				"message": "unknown message type",
				"data":    message,
			},
		})
	}

	return nil
}

func (h *hub) onLeaveCallRoom(chatID uint, client *HubClient) {
	h.leaveCallRoom(chatID, client)
	h.broadcastToChatRoom(chatID, dto.ResponseMessage{
		Event: onCallLeave,
		Data:  map[string]uint{"chat_id": chatID, "user_id": client.user.ID},
	})
}
