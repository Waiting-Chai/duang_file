package main

import (
	"duang_file/comm"
	"duang_file/handler"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化WebSocket管理器
	wsManager := comm.NewWebSocketManager()
	go wsManager.Start()

	// 初始化TransferHandler
	transferHandler := handler.NewTransferHandler(wsManager)

	// 设置Gin引擎
	r := gin.Default()

	// 设置WebSocket路由
	r.GET("/ws", func(c *gin.Context) {
		wsManager.HandleConnections(c.Writer, c.Request, transferHandler)
	})

	// 启动HTTP服务器
	log.Printf("Server started at :%d", 8081)
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
