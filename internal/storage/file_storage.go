package storage

import (
	"io/fs"
	"os"
	"path/filepath"
)

type FileStorage struct {
	Dir string
}

func NewFileStorage(dir string) *FileStorage {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err) // 对于简单脚手架，panic处理错误
	}
	return &FileStorage{Dir: dir}
}

// ListFiles 使用函数式风格，纯函数返回文件列表
func (s *FileStorage) ListFiles() ([]string, error) {
	var files []string
	err := filepath.WalkDir(s.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			relPath, _ := filepath.Rel(s.Dir, path)
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
}