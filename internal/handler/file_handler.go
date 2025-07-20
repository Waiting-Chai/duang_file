package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"duang_file/internal/storage"
)

type FileHandler struct {
	storage *storage.FileStorage
}

func NewFileHandler(storage *storage.FileStorage) *FileHandler {
	return &FileHandler{storage: storage}
}

func (h *FileHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "get form err: %s", err.Error())
		return
	}
	filename := filepath.Base(file.Filename)
	filePath := filepath.Join(h.storage.Dir, filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
		return
	}
	c.String(http.StatusOK, "File %s uploaded successfully", filename)
}

func (h *FileHandler) Download(c *gin.Context) {
	filename := c.Param("filename")
	filePath := filepath.Join(h.storage.Dir, filename)
	if _, err := os.Stat(filePath); err != nil {
		c.String(http.StatusNotFound, "File not found")
		return
	}
	c.File(filePath)
}

func (h *FileHandler) ListFiles(c *gin.Context) {
	files, err := h.storage.ListFiles()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error listing files: %s", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *FileHandler) Health(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}