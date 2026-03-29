package hub

import (
	"suscord/internal/domain/entity"
	"sync"

	"github.com/gorilla/websocket"
)

type HubClient struct {
	user       entity.User
	conn       *websocket.Conn
	chatRooms  map[uint]bool
	callRoomID uint
	mutex      sync.Mutex
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
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		return c.conn.WriteJSON(message)
	}
	return nil
}

func (c *HubClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	return err
}
