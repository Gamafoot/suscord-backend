package hub

import (
	"suscord/internal/domain/entity"

	"github.com/gorilla/websocket"
)

type HubClient struct {
	user       entity.User
	conn       *websocket.Conn
	chatRooms  map[uint]bool
	callRoomID uint
}

func NewHubClient(conn *websocket.Conn, user entity.User) *HubClient {
	return &HubClient{
		user:       user,
		conn:       conn,
		chatRooms:  make(map[uint]bool),
		callRoomID: 0,
	}
}

func (c *HubClient) SendMessage(message any) error {
	if c.conn != nil {
		return c.conn.WriteJSON(message)
	}
	return nil
}
