package models

import (
	"errors"
	"time"
)

type ScanStatus string

const (
	StatusSuccess          ScanStatus = "SUCCESS"
	StatusTimeout          ScanStatus = "TIMEOUT"
	StatusRejected         ScanStatus = "REJECTED"
	StatusPermissionDenied ScanStatus = "PERMISSION_DENIED"
	StatusUnknown          ScanStatus = "UNKNOWN"
)

type ScanTarget struct {
	IP   string
	Port int
}

type ScanResult struct {
	Target ScanTarget
	Status ScanStatus
	Err    error
}

type ScanConfig struct {
	FilePath    string
	Port        int
	Timeout     time.Duration
	Concurrency int
}

func (c *ScanConfig) Validate() error {
	if c.FilePath == "" {
		return errors.New("file path cannot be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}
	if c.Concurrency < 1 {
		c.Concurrency = 1
	} else if c.Concurrency > 1000 {
		c.Concurrency = 1000
	}
	return nil
}
