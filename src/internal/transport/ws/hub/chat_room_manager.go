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
	for _, client := range h.snapshotChatRoomClients(roomID, 0) {
		client.SendMessage(message)
	}
}

func (h *hub) broadcastToChatRoomExcept(roomID, exceptUserID uint, message any) {
	for _, client := range h.snapshotChatRoomClients(roomID, exceptUserID) {
		client.SendMessage(message)
	}
}

func (h *hub) broadcastToCallRoom(roomID uint, message any) {
	for _, client := range h.snapshotCallRoomClients(roomID, 0) {
		client.SendMessage(message)
	}
}

func (h *hub) broadcastToCallRoomExcept(roomID, exceptUserID uint, message any) {
	for _, client := range h.snapshotCallRoomClients(roomID, exceptUserID) {
		client.SendMessage(message)
	}
}

func (h *hub) broadcastToAll(message interface{}) {
	for _, client := range h.snapshotAllClients(0) {
		client.SendMessage(message)
	}
}

func (h *hub) broadcastToAllExcept(userID uint, message interface{}) {
	for _, client := range h.snapshotAllClients(userID) {
		client.SendMessage(message)
	}
}

func (h *hub) snapshotChatRoomClients(roomID, exceptUserID uint) []*HubClient {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients := make([]*HubClient, 0)

	room, exists := h.chatRooms[roomID]
	if !exists {
		return clients
	}

	for userID := range room {
		if userID == exceptUserID && exceptUserID != 0 {
			continue
		}

		client, exists := h.clients[userID]
		if !exists {
			continue
		}

		clients = append(clients, client)
	}

	return clients
}

func (h *hub) snapshotCallRoomClients(roomID, exceptUserID uint) []*HubClient {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients := make([]*HubClient, 0)

	room, exists := h.callRooms[roomID]
	if !exists {
		return clients
	}

	for userID := range room {
		if userID == exceptUserID && exceptUserID != 0 {
			continue
		}

		client, exists := h.clients[userID]
		if !exists {
			continue
		}

		clients = append(clients, client)
	}

	return clients
}

func (h *hub) snapshotAllClients(exceptUserID uint) []*HubClient {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients := make([]*HubClient, 0, len(h.clients))

	for userID, client := range h.clients {
		if userID == exceptUserID && exceptUserID != 0 {
			continue
		}

		clients = append(clients, client)
	}

	return clients
}
