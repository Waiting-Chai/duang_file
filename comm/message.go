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

// StartUploadPayload 客户端通知服务端开始上传文件
type StartUploadPayload struct {
	FileID      string   `json:"fileId"`
	FileName    string   `json:"fileName"`
	FileSize    int64    `json:"fileSize"`
	TotalChunks int      `json:"totalChunks"`
	TargetIDs   []string `json:"targetIds"` // 支持多接收方
	ChunkSize   int      `json:"chunkSize"` // 分片大小
}

// FileTransferRequestPayload 服务端通知接收方有文件待接收
type FileTransferRequestPayload struct {
	FileID      string `json:"fileId"`
	FileName    string `json:"fileName"`
	FileSize    int64  `json:"fileSize"`
	FromID      string `json:"fromId"`
	TotalChunks int    `json:"totalChunks"`
	ChunkSize   int    `json:"chunkSize"`
}

// FileTransferResponsePayload 接收方响应服务端的文件传输请求
type FileTransferResponsePayload struct {
	FileID string `json:"fileId"`
	Accept bool   `json:"accept"`
}

// TransferControlPayload 传输控制（暂停/恢复/取消）
type TransferControlPayload struct {
	FileID string `json:"fileId"`
	Action string `json:"action"` // "pause", "resume", "cancel"
}

// TransferProgressPayload 传输进度
type TransferProgressPayload struct {
	FileID           string `json:"fileId"`
	Progress         int    `json:"progress"`         // 0-100
	Speed            string `json:"speed"`            // e.g. "20kb/s"
	Status           string `json:"status"`           // e.g. "ready_to_receive_chunks", "sending", "receiving", "paused", "completed", "error"
	BytesTransferred int64  `json:"bytesTransferred"` // 已传输的字节数
	TotalBytes       int64  `json:"totalBytes"`       // 总字节数
}

// ErrorPayload 错误信息
type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code"` // 错误代码，方便前端处理
}
