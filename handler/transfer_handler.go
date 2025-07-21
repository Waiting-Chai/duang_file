package handler

import (
	"duang_file/comm"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileState 存储文件上传的状态
type FileState struct {
	FileID      string
	FileName    string
	FileSize    int64
	TotalChunks int
	Chunks      map[int][]byte
	TargetID    string
	FromID      string
	StartTime   time.Time
	mu          sync.Mutex
}

// TransferHandler 处理文件传输相关的 WebSocket 消息
type TransferHandler struct {
	WsManager *comm.WebSocketManager
	files     map[string]*FileState
	filesMu   sync.RWMutex
}

// NewTransferHandler 创建一个新的 TransferHandler
func NewTransferHandler(wsManager *comm.WebSocketManager) *TransferHandler {
	return &TransferHandler{
		WsManager: wsManager,
		files:     make(map[string]*FileState),
	}
}

// HandleMessage 根据消息类型路由到不同的处理函数
func (h *TransferHandler) HandleMessage(client *comm.WebSocketClient, message comm.WebSocketMessage) {
	log.Printf("接收到客户端 %s 的消息，类型: %s", client.ID, message.Type)

	switch message.Type {
	case "get_client_list":
		return
	case "start_upload":
		h.handleStartUpload(client, message.Payload)
	case "upload_chunk":
		h.handleUploadChunk(client, message.Payload)
	case "accept_transfer":
		h.handleAcceptTransfer(client, message.Payload)
	case "reject_transfer":
		h.handleRejectTransfer(client, message.Payload)
	default:
		log.Printf("未知的消息类型: %s", message.Type)
	}
}

func (h *TransferHandler) handleStartUpload(client *comm.WebSocketClient, payload interface{}) {
	var p comm.StartUploadPayload
	if err := mapToStruct(payload, &p); err != nil {
		log.Printf("解析 start_upload payload 失败: %v", err)
		return
	}

	h.filesMu.Lock()
	h.files[p.FileID] = &FileState{
		FileID:      p.FileID,
		FileName:    p.FileName,
		FileSize:    p.FileSize,
		TotalChunks: p.TotalChunks,
		Chunks:      make(map[int][]byte),
		TargetID:    p.TargetID,
		FromID:      client.ID,
		StartTime:   time.Now(),
	}
	h.filesMu.Unlock()

	log.Printf("文件上传初始化: %s (%d bytes, %d chunks) from %s to %s", p.FileName, p.FileSize, p.TotalChunks, client.ID, p.TargetID)

	// 发送准备接收文件块的消息
	progressPayload := comm.TransferProgressPayload{
		FileID:   p.FileID,
		Progress: 0,
		Speed:    "0 kb/s",
		Status:   "ready_to_receive_chunks",
	}
	progressMessage := comm.WebSocketMessage{
		Type:    "transfer_progress",
		Payload: progressPayload,
	}
	jsonProgress, err := json.Marshal(progressMessage)
	if err != nil {
		log.Printf("序列化 transfer_progress 消息失败: %v", err)
		return
	}

	// 发送给发送方
	if err := h.WsManager.Send(client.ID, jsonProgress); err != nil {
		log.Printf("发送 ready_to_receive_chunks 消息失败: %v", err)
	}
}

func (h *TransferHandler) handleUploadChunk(client *comm.WebSocketClient, payload interface{}) {
	var p comm.UploadChunkPayload
	if err := mapToStruct(payload, &p); err != nil {
		log.Printf("解析 upload_chunk payload 失败: %v", err)
		return
	}

	h.filesMu.RLock()
	fileState, ok := h.files[p.FileID]
	h.filesMu.RUnlock()

	if !ok {
		log.Printf("未找到文件状态: %s", p.FileID)
		return
	}

	fileState.mu.Lock()
	fileState.Chunks[p.ChunkID] = p.Data
	fileState.mu.Unlock()

	// 计算进度和速度
	progress := (len(fileState.Chunks) * 100) / fileState.TotalChunks
	elapsed := time.Since(fileState.StartTime).Seconds()
	bytesSent := int64(len(fileState.Chunks) * len(p.Data)) // 这是一个估算值
	speed := float64(bytesSent) / elapsed / 1024            // kb/s

	log.Printf("接收到文件块: %s, chunk %d/%d, progress: %d%%, speed: %.2f kb/s", fileState.FileName, len(fileState.Chunks), fileState.TotalChunks, progress, speed)

	// 广播进度
	progressPayload := comm.TransferProgressPayload{
		FileID:   p.FileID,
		Progress: progress,
		Speed:    fmt.Sprintf("%.2f kb/s", speed),
	}
	progressMessage := comm.WebSocketMessage{
		Type:    "transfer_progress",
		Payload: progressPayload,
	}
	jsonProgress, _ := json.Marshal(progressMessage)
	// 发送给发送方和接收方
	h.WsManager.Send(fileState.FromID, jsonProgress)
	h.WsManager.Send(fileState.TargetID, jsonProgress)

	if len(fileState.Chunks) == fileState.TotalChunks {
		go h.assembleAndSaveFile(fileState, client.ID)
	}
}

func (h *TransferHandler) assembleAndSaveFile(fileState *FileState, fromID string) {
	filePath := filepath.Join("uploads", fileState.FileName)
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		log.Printf("创建上传目录失败: %v", err)
		return
	}

	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("创建文件失败: %v", err)
		return
	}
	defer out.Close()

	for i := 0; i < fileState.TotalChunks; i++ {
		if _, err := out.Write(fileState.Chunks[i]); err != nil {
			log.Printf("写入文件块失败: %v", err)
			return
		}
	}

	log.Printf("文件 %s 已成功保存到 %s", fileState.FileName, filePath)

	// 清理内存
	h.filesMu.Lock()
	delete(h.files, fileState.FileID)
	h.filesMu.Unlock()

	// 通知目标客户端
	h.notifyTargetClient(fileState, fromID)
}

func (h *TransferHandler) handleAcceptTransfer(client *comm.WebSocketClient, payload interface{}) {
	var p comm.FileTransferResponsePayload
	if err := mapToStruct(payload, &p); err != nil {
		log.Printf("解析 accept_transfer payload 失败: %v", err)
		return
	}

	h.filesMu.RLock()
	fileState, ok := h.files[p.FileID]
	h.filesMu.RUnlock()

	if !ok {
		log.Printf("未找到文件状态: %s", p.FileID)
		return
	}

	// 通知发送方，对方已接受
	acceptMessage := comm.WebSocketMessage{
		Type: "transfer_accepted",
	}
	jsonAccept, _ := json.Marshal(acceptMessage)
	h.WsManager.Send(fileState.FromID, jsonAccept)

	// 开始将文件块发送给接收方
	go h.sendFileChunks(client, fileState)
}

func (h *TransferHandler) sendFileChunks(client *comm.WebSocketClient, fileState *FileState) {
	filePath := filepath.Join("uploads", fileState.FileName)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("打开文件失败: %v", err)
		return
	}
	defer file.Close()

	buffer := make([]byte, 1024*64) // 64KB buffer
	chunkID := 0
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("读取文件块失败: %v", err)
			return
		}

		chunkPayload := comm.UploadChunkPayload{
			FileID:  fileState.FileID,
			ChunkID: chunkID,
			Data:    buffer[:bytesRead],
		}

		message := comm.WebSocketMessage{
			Type:    "upload_chunk",
			Payload: chunkPayload,
		}

		jsonMessage, _ := json.Marshal(message)
		if err := client.SendBinary(jsonMessage); err != nil {
			log.Printf("发送文件块失败: %v", err)
			return
		}
		chunkID++
	}

	log.Printf("文件 %s 已成功发送给 %s", fileState.FileName, client.ID)

	// 清理文件
	h.filesMu.Lock()
	delete(h.files, fileState.FileID)
	h.filesMu.Unlock()
	os.Remove(filePath)
}

func (h *TransferHandler) handleRejectTransfer(client *comm.WebSocketClient, payload interface{}) {
	var p comm.FileTransferResponsePayload
	if err := mapToStruct(payload, &p); err != nil {
		log.Printf("解析 reject_transfer payload 失败: %v", err)
		return
	}

	// 通知发送方对方已拒绝
	h.filesMu.RLock()
	fileState, ok := h.files[p.FileID]
	h.filesMu.RUnlock()

	if !ok {
		log.Printf("未找到文件状态: %s", p.FileID)
		return
	}

	// 通知发送方对方已拒绝
	errorPayload := comm.ErrorPayload{
		Message: "对方拒绝了文件传输",
	}
	message := comm.WebSocketMessage{
		Type:    "transfer_rejected",
		Payload: errorPayload,
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("序列化 transfer_rejected 消息失败: %v", err)
		return
	}

	if err := h.WsManager.Send(fileState.FromID, jsonMessage); err != nil {
		log.Printf("发送 transfer_rejected 消息失败: %v", err)
	}

	// 清理文件状态
	h.filesMu.Lock()
	delete(h.files, p.FileID)
	h.filesMu.Unlock()
}

func (h *TransferHandler) notifyTargetClient(fileState *FileState, fromID string) {
	payload := comm.FileTransferRequestPayload{
		FileID:   fileState.FileID,
		FileName: fileState.FileName,
		FileSize: fileState.FileSize,
		FromID:   fromID,
	}

	message := comm.WebSocketMessage{
		Type:    "file_transfer_request",
		Payload: payload,
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("序列化文件传输请求失败: %v", err)
		return
	}

	if err := h.WsManager.Send(fileState.TargetID, jsonMessage); err != nil {
		log.Printf("发送文件传输请求失败: %v", err)
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
