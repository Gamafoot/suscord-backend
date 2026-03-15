package dto

import "encoding/json"

type ClientMessage struct {
	Event  string          `json:"event"`
	ChatID uint            `json:"chat_id,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}
