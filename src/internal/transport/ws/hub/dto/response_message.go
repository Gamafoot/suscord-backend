package dto

type ResponseMessage struct {
	Event  string `json:"event"`
	Data   any    `json:"data"`
	ChatID uint   `json:"-"`
}
