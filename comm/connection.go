package comm

import (
	"duang_file/model"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 从连接写入消息的超时时间
	writeWait = 10 * time.Second

	// 读取下一次 pong 消息的超时时间
	pongWait = 60 * time.Second

	// 向连接发送 ping 消息的周期
	pingPeriod = (pongWait * 9) / 10

	// 允许从连接读取的最大消息大小
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应配置具体的源
	},
}

// WebSocketClient 表示一个 WebSocket 客户端连接
type WebSocketClient struct {
	manager    *WebSocketManager
	conn       *websocket.Conn
	send       chan []byte
	ID         string
	DeviceInfo model.DeviceInfo
}

// readPump 从 WebSocket 连接中读取消息
func (c *WebSocketClient) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

// writePump 将消息从 send 通道写入 WebSocket 连接
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// manager 关闭了 send 通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// WebSocketManager 管理所有 WebSocket 客户端连接
type WebSocketManager struct {
	clients    map[string]*WebSocketClient
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mutex      sync.RWMutex
}

// NewWebSocketManager 创建一个新的 WebSocketManager 实例
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[string]*WebSocketClient),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
	}
}

// Start 启动 WebSocket 管理器
func (m *WebSocketManager) Start() {
	log.Println("WebSocket 管理器已启动")
	for {
		select {
		case client := <-m.register:
			m.registerClient(client)
		case client := <-m.unregister:
			m.unregisterClient(client)
		}
	}
}

// registerClient 注册一个新的客户端
func (m *WebSocketManager) registerClient(client *WebSocketClient) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.clients[client.ID] = client
	log.Printf("客户端 %s 已连接", client.ID)
	go m.BroadcastClientList() // 连接成功后广播客户端列表
}

// unregisterClient 注销一个客户端
func (m *WebSocketManager) unregisterClient(client *WebSocketClient) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.clients[client.ID]; ok {
		delete(m.clients, client.ID)
		close(client.send)
		log.Printf("客户端 %s 已断开连接", client.ID)
		go m.BroadcastClientList() // 断开连接后广播客户端列表
	}
}

// BroadcastClientList 广播当前所有连接的客户端列表
func (m *WebSocketManager) BroadcastClientList() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	clientsInfo := make([]model.ClientInfo, 0, len(m.clients))
	for _, client := range m.clients {
		clientsInfo = append(clientsInfo, model.ClientInfo{ID: client.ID, DeviceInfo: client.DeviceInfo})
	}

	payload := ClientListPayload{Clients: clientsInfo}
	message := WebSocketMessage{Type: "client_list", Payload: payload}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("序列化客户端列表失败: %v", err)
		return
	}

	log.Println("正在广播客户端列表...")
	for id, client := range m.clients {
		select {
		case client.send <- jsonMessage:
		default:
			log.Printf("客户端 %s 的发送通道已满，无法广播", id)
		}
	}
}

// Send 发送消息给指定的客户端
func (m *WebSocketManager) Send(clientID string, message []byte) error {
	m.mutex.RLock()
	client, ok := m.clients[clientID]
	m.mutex.RUnlock()

	if !ok {
		return errors.New("客户端 " + clientID + " 不存在")
	}

	select {
	case client.send <- message:
		return nil
	default:
		// 如果通道已满，可以考虑关闭连接或返回错误
		return errors.New("客户端 " + clientID + " 的发送通道已满")
	}
}

// HandleConnections 处理新的 WebSocket 连接请求
func (m *WebSocketManager) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 连接升级失败: %v", err)
		return
	}

	deviceId := r.URL.Query().Get("deviceId")
	if deviceId == "" {
		log.Println("deviceId 为空，连接关闭")
		conn.Close()
		return
	}
	log.Printf("接收到来自 deviceId: %s 的新连接", deviceId)

	deviceInfo := model.DeviceInfo{
		IP:        r.RemoteAddr,
		Device:    r.URL.Query().Get("device"),
		OS:        r.URL.Query().Get("os"),
		Browser:   r.URL.Query().Get("browser"),
		Timestamp: time.Now().Unix(),
	}

	client := &WebSocketClient{
		manager:    m,
		conn:       conn,
		send:       make(chan []byte, 256),
		ID:         deviceId,
		DeviceInfo: deviceInfo,
	}

	m.register <- client

	go client.writePump()
	go client.readPump()
}
