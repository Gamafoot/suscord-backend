package hub

import (
	gDTO "suscord/internal/transport/dto"
	"suscord/internal/transport/ws/hub/dto"
)

func (h *hub) handleClientMessage(client *HubClient, message *dto.ClientMessage) error {
	switch message.Event {
	// case onCallInvite:
	// 	h.joinCallRoom(message.RoomID, client)
	// 	h.broadcastToChatRoomExcept(message.RoomID, client.user.ID, message)

	case onCallJoin:
		h.joinCallRoom(message.ChatID, client)
		h.broadcastToChatRoomExcept(message.ChatID, client.user.ID, dto.ResponseMessage{
			Event: onCallJoin,
			Data: map[string]any{
				"chat_id": message.ChatID,
				"user":    gDTO.NewUser(client.user, h.cfg.Media.Url),
			},
		})
		// if len(h.callRooms[message.RoomID]) == 1 {
		// 	h.broadcastToChatRoomExcept(message.RoomID, client.user.ID, message)
		// }

	case onCallLeave:
		if client.callRoomID == 0 {
			return ErrNotInCall
		}

		h.leaveCallRoom(message.ChatID, client)
		h.broadcastToChatRoom(message.ChatID, dto.ResponseMessage{
			Event: onCallLeave,
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
