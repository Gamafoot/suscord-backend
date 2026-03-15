package hub

func (h *hub) joinCallRoom(roomID uint, client *HubClient) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if len(h.callRooms[roomID]) == 0 {
		h.callRooms[roomID] = make(map[uint]bool)
	}

	client.callRoomID = roomID

	h.callRooms[roomID][client.user.ID] = true
}

func (h *hub) leaveCallRoom(roomID uint, client *HubClient) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, exists := h.callRooms[roomID]; exists {
		delete(h.callRooms[roomID], client.user.ID)
	}

	client.callRoomID = 0

	if len(h.callRooms[roomID]) == 0 {
		delete(h.callRooms, roomID)
	}
}
