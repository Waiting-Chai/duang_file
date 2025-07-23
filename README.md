# Duang File - 局域网文件传输工具后端

## 概述
这是一个Go后端项目，使用Gin框架实现局域网内文件传输、WebSocket信令转发和广播功能。支持WebRTC文件传输、客户端列表广播和消息广播。

## 安装
1. 确保安装Go (推荐1.20+)。
2. 克隆仓库。
3. 运行 `go mod tidy` 安装依赖。

## 运行
- 执行 `go run main.go` 启动服务器于端口8081。
- 访问 http://localhost:8080/health 检查服务器状态。

## API 端点

### 文件操作
- POST /upload: 上传文件 (form-data, key: file)
- GET /download/:filename: 下载文件
- GET /files: 获取文件列表 (JSON)
- GET /health: 健康检查

### WebSocket
- GET /ws: WebSocket连接端点，支持设备注册和消息转发。
消息类型包括：
  - webrtc_signal: WebRTC信令转发
  - file_transfer_request/response: 文件传输请求/响应
  - get_client_list: 获取客户端列表
  - broadcastMessage: 广播消息

## 部署
### Docker部署
1. 构建镜像：`docker build -t duang-backend .`
2. 运行容器：`docker run -d -p 8080:8080 --name duang-backend duang-backend`

或使用 deploy.sh 脚本：`./deploy.sh`

## 项目结构
```
├── comm/
│   ├── connection.go  # WebSocket连接管理和广播
│   └── message.go     # 消息结构体
├── handler/
│   └── transfer_handler.go  # 信令和广播处理
├── main.go            # 主程序入口
└── model/
    └── device.go      # 设备模型
```

更多细节见代码注释。