package http

import (
	"fmt"
	"io"
	"os"
)

type FileAdapter struct{}

func NewFileAdapter() *FileAdapter {
	return &FileAdapter{}
}

func (a *FileAdapter) CreateFile(path string) (io.WriteCloser, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	return file, nil
}
