package comm

import "duang_file/model"

// WebSocketMessage 是通用的 WebSocket 消息结构
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// ClientListPayload 是客户端列表消息的载荷
type ClientListPayload struct {
	Clients []model.ClientInfo `json:"clients"`
}