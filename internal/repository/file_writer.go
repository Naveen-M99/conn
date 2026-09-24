package repository

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"conn/internal/models"
)

type OutputRepository interface {
	Write(result models.ScanResult) error
	Close() error
}

type FileWriter struct {
	mu      sync.Mutex
	files   map[models.ScanStatus]*os.File
	writers map[models.ScanStatus]*bufio.Writer
}

func NewFileWriter() (*FileWriter, error) {
	statusFileMap := map[models.ScanStatus]string{
		models.StatusSuccess:          "success_ips.txt",
		models.StatusTimeout:          "timeout_ips.txt",
		models.StatusRejected:         "rejected_ips.txt",
		models.StatusPermissionDenied: "permission_denied_ips.txt",
		models.StatusUnknown:          "unknown_ips.txt",
	}

	files := make(map[models.ScanStatus]*os.File)
	writers := make(map[models.ScanStatus]*bufio.Writer)

	for status, filename := range statusFileMap {
		cleanPath := filepath.Clean(filename)

		// os.Create handles O_RDWR|O_CREATE|O_TRUNC properly on both Windows and Linux
		f, err := os.Create(cleanPath)
		if err != nil {
			for _, opened := range files {
				_ = opened.Close()
			}
			return nil, fmt.Errorf("failed to create %s: %w", filename, err)
		}
		files[status] = f
		writers[status] = bufio.NewWriterSize(f, 32*1024)
	}

	return &FileWriter{
		files:   files,
		writers: writers,
	}, nil
}

func (fw *FileWriter) Write(result models.ScanResult) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	writer, ok := fw.writers[result.Status]
	if !ok {
		return fmt.Errorf("unknown status: %s", result.Status)
	}

	_, err := writer.WriteString(result.Target.IP + "\r\n")
	return err
}

func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	var firstErr error
	for status, writer := range fw.writers {
		if err := writer.Flush(); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := fw.files[status].Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
