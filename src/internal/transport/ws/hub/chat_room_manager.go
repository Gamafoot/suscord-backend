package hub

import (
	"suscord/internal/domain/entity"
	"suscord/internal/domain/event"
	"suscord/internal/transport/ws/hub/dto"
)

func (h *hub) joinChatRoom(chatID uint, client *HubClient) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, exists := h.chatRooms[chatID]; !exists {
		h.chatRooms[chatID] = make(map[uint]bool)
	}

	h.chatRooms[chatID][client.user.ID] = true
	client.chatRooms[chatID] = true
}

func (h *hub) leaveChatRoom(roomID, clientID uint) {
	h.mutex.Lock()

	if room, exists := h.chatRooms[roomID]; exists {
		delete(room, clientID)
		if len(room) == 0 {
			delete(h.chatRooms, roomID)
		}
	}

	if client, exists := h.clients[clientID]; exists {
		delete(client.chatRooms, roomID)
	}

	h.mutex.Unlock()

	h.broadcastToChatRoomExcept(roomID, clientID, &dto.ResponseMessage{
		Event: event.OnUserLeft,
		Data:  map[string]uint{"user_id": clientID},
	})
}

func (h *hub) deleteChatRoom(roomID uint) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.chatRooms[roomID] = make(map[uint]bool)

	for clientID := range h.clients {
		delete(h.clients[clientID].chatRooms, roomID)
	}
}

func (h *hub) joinToUserChatRooms(client *HubClient, chats []entity.Chat) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, chat := range chats {
		if _, exists := h.chatRooms[chat.ID]; !exists {
			h.chatRooms[chat.ID] = make(map[uint]bool)
		}

		h.chatRooms[chat.ID][client.user.ID] = true
		client.chatRooms[chat.ID] = true
	}
}

func (h *hub) broadcastToChatRoom(roomID uint, message interface{}) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if room, exists := h.chatRooms[roomID]; exists {
		for userID := range room {
			if client, exists := h.clients[userID]; exists {
				client.SendMessage(message)
			}
		}
	}
}

func (h *hub) broadcastToChatRoomExcept(roomID, exceptUserID uint, message any) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if room, exists := h.chatRooms[roomID]; exists {
		for userID := range room {
			if userID != exceptUserID {
				if client, exists := h.clients[userID]; exists {
					client.SendMessage(message)
				}
			}
		}
	}
}

func (h *hub) broadcastToCallRoom(roomID uint, message any) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if room, exists := h.callRooms[roomID]; exists {
		for userID := range room {
			if client, exists := h.clients[userID]; exists {
				client.SendMessage(message)
			}
		}
	}
}

func (h *hub) broadcastToCallRoomExcept(roomID, exceptUserID uint, message any) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if room, exists := h.callRooms[roomID]; exists {
		for userID := range room {
			if userID != exceptUserID {
				if client, exists := h.clients[userID]; exists {
					client.SendMessage(message)
				}
			}
		}
	}
}

func (h *hub) broadcastToAll(message interface{}) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for userID := range h.clients {
		h.clients[userID].SendMessage(message)
	}
}

func (h *hub) broadcastToAllExcept(userID uint, message interface{}) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for uID := range h.clients {
		if uID == userID {
			continue
		}
		h.clients[uID].SendMessage(message)
	}
}
