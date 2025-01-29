// files.go

package repomap

import (
	"os"
	"path/filepath"
)

type FileSystem interface {
	GetFiles(dir string) ([]string, error)
	ReadFile(path string) (string, error)
}

type SimpleFileSystem struct{}

func (fs *SimpleFileSystem) GetFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (fs *SimpleFileSystem) ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
