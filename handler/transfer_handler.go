package handler

import (
	"duang_file/comm"
	"encoding/json"
	"log"
)

// TransferHandler 只处理信令转发，不处理文件传输状态

type TransferHandler struct {
	WsManager *comm.WebSocketManager
}

// NewTransferHandler 创建一个新的 TransferHandler
func NewTransferHandler(wsManager *comm.WebSocketManager) *TransferHandler {
	return &TransferHandler{
		WsManager: wsManager,
	}
}

// HandleMessage 根据消息类型路由到不同的处理函数
func (h *TransferHandler) HandleMessage(client *comm.WebSocketClient, message comm.WebSocketMessage) {
	log.Printf("接收到客户端 %s 的消息，类型: %s", client.ID, message.Type)

	switch message.Type {
	case "webrtc_signal", "file_transfer_request", "file_transfer_response":
		h.forwardMessage(client, message)
	case "get_client_list":
		// 收到获取客户端列表的请求，立即广播客户端列表
		log.Printf("收到获取客户端列表请求，立即广播")
		h.WsManager.BroadcastClientList()
	default:
		// 对于非信令消息，可以选择记录日志或直接忽略
		log.Printf("忽略非信令消息: %s", message.Type)
	}
}

// forwardMessage 负责将消息转发给指定的接收方
func (h *TransferHandler) forwardMessage(client *comm.WebSocketClient, message comm.WebSocketMessage) {
	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Printf("webrtc_signal payload 不是有效的map: %v", message.Payload)
		return
	}

	// 从 payload 中提取 toId
	targetID, ok := payloadMap["toId"].(string)
	if !ok {
		log.Printf("在payload中未找到接收方ID 'toId'")
		return
	}

	// 在转发的消息中添加 fromId
	payloadMap["fromId"] = client.ID

	// 重新构造消息以确保 fromId 被包含
	forwardMessage := comm.WebSocketMessage{
		Type:    message.Type,
		Payload: payloadMap,
	}

	jsonMessage, err := json.Marshal(forwardMessage)
	if err != nil {
		log.Printf("序列化转发消息失败: %v", err)
		return
	}

	// 将消息发送给目标客户端
	if err := h.WsManager.Send(targetID, jsonMessage); err != nil {
		log.Printf("转发消息给 %s 失败: %v", targetID, err)
	}
}

// mapToStruct 是一个辅助函数，用于将 map[string]interface{} 转换为结构体
func mapToStruct(data interface{}, result interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, result)
}
