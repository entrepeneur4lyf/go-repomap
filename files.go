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
	dir = filepath.Clean(dir)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Normalize path separators for cross-platform consistency
			path = filepath.ToSlash(path)
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (fs *SimpleFileSystem) ReadFile(path string) (string, error) {
	// Clean and normalize the path
	path = filepath.Clean(filepath.FromSlash(path))
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Ensure consistent line endings
	return string(content), nil
}
