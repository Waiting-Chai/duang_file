package main

import (
	"duang_file/comm"
	"duang_file/internal/handler"
	"duang_file/internal/storage"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化WebSocket管理器
	wsManager := comm.NewWebSocketManager()
	go wsManager.Start()

	// 初始化处理器
	fileStorage := storage.NewFileStorage("./uploads")
	fileHandler := handler.NewFileHandler(fileStorage)

	// 设置Gin引擎
	r := gin.Default()

	// 设置文件上传、下载、列表和健康检查的路由
	r.POST("/upload", fileHandler.Upload)
	r.GET("/download/:filename", fileHandler.Download)
	r.GET("/files", fileHandler.ListFiles)
	r.GET("/health", fileHandler.Health)

	// 设置WebSocket路由
	r.GET("/ws", func(c *gin.Context) {
		wsManager.HandleConnections(c.Writer, c.Request)
	})

	// 启动HTTP服务器
	log.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
