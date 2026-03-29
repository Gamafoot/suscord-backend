package hub

import (
	"errors"
	"fmt"
	"suscord/internal/config"
	"suscord/internal/domain/entity"
	"suscord/internal/domain/eventbus"
	"suscord/internal/domain/storage"
	gDTO "suscord/internal/transport/dto"
	"suscord/internal/transport/ws/hub/dto"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Clients map[uint]*HubClient

type ChatRooms map[uint]map[uint]bool
type CallRooms map[uint]map[uint]bool

type Hub interface {
	Register(client *HubClient, chats []entity.Chat)
	Unregister(client *HubClient)
	ReceiveMessageHandler(client *HubClient)
	GetCurrentCallMembers(clientID uint) ([]entity.User, error)
}

type hub struct {
	cfg   *config.Config
	mutex sync.RWMutex

	chatRooms ChatRooms
	callRooms CallRooms
	clients   Clients

	storage storage.Storage
	logger  *zap.SugaredLogger
}

type callParticipant struct {
	chatID uint
	user   entity.User
}

func NewHub(
	config *config.Config,
	storage storage.Storage,
	eventbus eventbus.EventBus,
	logger *zap.SugaredLogger,
) *hub {
	hub := &hub{
		cfg:   config,
		mutex: sync.RWMutex{},

		chatRooms: make(ChatRooms),
		callRooms: make(CallRooms),
		clients:   make(Clients),

		storage: storage,
		logger:  logger,
	}
	hub.registerEvents(eventbus)
	return hub
}

func (h *hub) ReceiveMessageHandler(client *HubClient) {
	for {
		message := new(dto.ClientMessage)
		err := client.conn.ReadJSON(message)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				h.logger.Infow("ws closed by client", "user_id", client.user.ID, "error", err)
			} else {
				h.logger.Errorw("ws ReadJSON error", "user_id", client.user.ID, "error", err)
			}
			h.Unregister(client)
			return
		}

		err = h.handleClientMessage(client, message)
		if err != nil {
			h.logger.Errorw("ws handleClientMessage error", "error", err)
		}
	}
}

func (h *hub) Register(client *HubClient, chats []entity.Chat) {
	h.mutex.Lock()
	h.clients[client.user.ID] = client
	h.mutex.Unlock()
	h.joinToUserChatRooms(client, chats)

	for _, participant := range h.snapshotCallParticipants(client) {
		client.SendMessage(dto.ResponseMessage{
			Event: onCallJoin,
			Data: map[string]any{
				"chat_id": participant.chatID,
				"user":    gDTO.NewUser(participant.user, h.cfg.Media.Url),
			},
		})
	}
}

func (h *hub) GetCurrentCallMembers(clientID uint) ([]entity.User, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	client, ok := h.clients[clientID]
	if !ok {
		return nil, fmt.Errorf("client with id = %d not found in websocket connections", clientID)
	}
	if client.callRoomID == 0 {
		return nil, errors.New("you are not in a call")
	}

	result := make([]entity.User, 0)

	if clients, ok := h.callRooms[client.callRoomID]; ok {
		for cID := range clients {
			if member, ok := h.clients[cID]; ok {
				result = append(result, member.user)
			}
		}
	}

	return result, nil
}

func (h *hub) Unregister(client *HubClient) {
	defer client.Close()

	pendingRoomIDs := make([]uint, 0)

	h.mutex.Lock()
	if _, exists := h.clients[client.user.ID]; exists {
		for roomID := range client.chatRooms {
			pendingRoomIDs = append(pendingRoomIDs, roomID)
			if room, ok := h.chatRooms[roomID]; ok {
				delete(room, client.user.ID)

				if len(room) == 0 {
					delete(h.chatRooms, roomID)
				}
			}
		}

		for roomID, room := range h.callRooms {
			if _, ok := room[client.user.ID]; ok {
				delete(room, client.user.ID)

				if len(room) == 0 {
					delete(h.callRooms, roomID)
				}
			}
		}

		delete(h.clients, client.user.ID)
	}
	h.mutex.Unlock()

	if client.callRoomID != 0 {
		h.onLeaveCallRoom(client.callRoomID, client)
	}

	for _, roomID := range pendingRoomIDs {
		h.broadcastToChatRoom(roomID, dto.ResponseMessage{
			Event: onCallLeave,
			Data: map[string]uint{
				"user_id": client.user.ID,
				"chat_id": roomID,
			},
		})
	}
}

func (h *hub) getClient(userID uint) (*HubClient, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	client, exists := h.clients[userID]
	return client, exists
}

func (h *hub) snapshotCallParticipants(client *HubClient) []callParticipant {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	participants := make([]callParticipant, 0)

	for chatRoomID := range client.chatRooms {
		room, exists := h.callRooms[chatRoomID]
		if !exists {
			continue
		}

		for participantID := range room {
			if participantID == client.user.ID {
				continue
			}

			participant, ok := h.clients[participantID]
			if !ok {
				continue
			}

			participants = append(participants, callParticipant{
				chatID: chatRoomID,
				user:   participant.user,
			})
		}
	}

	return participants
}
