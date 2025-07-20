# Duang File - 局域网文件传输工具

## 概述
这是一个简单的Go后端项目，使用Gin框架实现局域网内文件上传、下载和列表功能，并支持WebSocket实时通信。项目设计轻量，适合学习Go web开发。

## 安装
1. 确保安装Go (推荐1.18+)。
2. 克隆仓库或在当前目录初始化。
3. 运行 `go mod tidy` 安装依赖。

## 运行
- 执行 `go run main.go` 启动服务器于端口8080。
- 访问 http://localhost:8080/health 检查服务器状态。

## API 端点

### 文件操作
- POST /upload: 上传文件 (form-data, key: file)
- GET /download/:filename: 下载文件
- GET /files: 获取文件列表 (JSON)
- GET /health: 健康检查

### WebSocket
- GET /ws: WebSocket连接端点
- GET /ws/status: 获取当前WebSocket连接状态
- POST /ws/broadcast: 广播消息给所有连接的客户端
  ```json
  {
    "message": "要广播的消息内容"
  }
  ```
- POST /ws/send/:client_id: 发送消息给特定客户端
  ```json
  {
    "message": "要发送的消息内容"
  }
  ```

## 测试页面
访问 http://localhost:8080/download/websocket_test.html 可以打开WebSocket测试页面，用于测试WebSocket连接和消息收发功能。

## 学习点
- Gin路由和中间件
- 文件处理在Go中
- WebSocket实时通信
- 并发和goroutine
- 错误处理最佳实践

## 项目结构
```
├── comm/
│   └── connection.go      # WebSocket连接管理
├── internal/
│   ├── handler/
│   │   ├── file_handler.go # 文件处理器
│   │   └── ws_handler.go   # WebSocket处理器
│   └── storage/
│       └── file_storage.go # 文件存储
├── main.go                # 主程序入口
└── uploads/              # 上传文件存储目录
    └── websocket_test.html # WebSocket测试页面
```

更多细节见代码注释。