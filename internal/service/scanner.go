package service

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"conn/internal/models"
	"conn/internal/repository"
)

type ScannerService struct {
	repo repository.OutputRepository
}

func NewScannerService(repo repository.OutputRepository) *ScannerService {
	return &ScannerService{repo: repo}
}

func (s *ScannerService) Run(ctx context.Context, cfg models.ScanConfig) error {
	targets, err := s.loadTargets(cfg.FilePath, cfg.Port)
	if err != nil {
		return err
	}

	if len(targets) == 0 {
		return fmt.Errorf("no valid IP addresses found in %s", cfg.FilePath)
	}

	jobs := make(chan models.ScanTarget, len(targets))
	results := make(chan models.ScanResult, cfg.Concurrency*2)

	var workerWg sync.WaitGroup

	for i := 0; i < cfg.Concurrency; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for target := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					results <- s.probe(ctx, target, cfg.Timeout)
				}
			}
		}()
	}

	go func() {
		for _, target := range targets {
			jobs <- target
		}
		close(jobs)
	}()

	go func() {
		workerWg.Wait()
		close(results)
	}()

	for res := range results {
		s.logResult(res)
		if err := s.repo.Write(res); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing result for %s: %v\n", res.Target.IP, err)
		}
	}

	return nil
}

func (s *ScannerService) probe(ctx context.Context, target models.ScanTarget, timeout time.Duration) models.ScanResult {
	addr := net.JoinHostPort(target.IP, fmt.Sprintf("%d", target.Port))

	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err == nil {
		_ = conn.Close()
		return models.ScanResult{Target: target, Status: models.StatusSuccess}
	}

	// 1. Check for OS-level or Firewall Permission Denied errors
	if isPermissionDenied(err) {
		return models.ScanResult{Target: target, Status: models.StatusPermissionDenied, Err: err}
	}

	// 2. Timeout check
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return models.ScanResult{Target: target, Status: models.StatusTimeout, Err: err}
	}

	// 3. Port rejected/reset check
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "refused") || strings.Contains(errStr, "reset") {
		return models.ScanResult{Target: target, Status: models.StatusRejected, Err: err}
	}

	// 4. Default to Unknown
	return models.ScanResult{Target: target, Status: models.StatusUnknown, Err: err}
}

// isPermissionDenied inspects OS syscalls and error strings across Linux and Windows
func isPermissionDenied(err error) bool {
	if err == nil {
		return false
	}

	// Check underlying syscall error (Linux/Unix EACCES, EPERM)
	var syscallErr syscall.Errno
	if errors.As(err, &syscallErr) {
		if errors.Is(syscallErr, syscall.EACCES) || errors.Is(syscallErr, syscall.EPERM) {
			return true
		}
	}

	// String match fallback for Windows firewall/socket block messages and ICMP prohibit text
	errLower := strings.ToLower(err.Error())
	if strings.Contains(errLower, "permission denied") ||
		strings.Contains(errLower, "operation not permitted") ||
		strings.Contains(errLower, "forbidden by its access permissions") || // Windows Winsock error 10013 (WSAEACCES)
		strings.Contains(errLower, "administratively prohibited") { // ICMP Type 3 Code 10/13
		return true
	}

	return false
}

func (s *ScannerService) loadTargets(filePath string, port int) ([]models.ScanTarget, error) {
	cleanPath := filepath.Clean(filePath)
	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open target file: %w", err)
	}
	defer file.Close()

	var targets []models.ScanTarget
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			targets = append(targets, models.ScanTarget{
				IP:   fields[0],
				Port: port,
			})
		}
	}

	return targets, scanner.Err()
}

func (s *ScannerService) logResult(res models.ScanResult) {
	switch res.Status {
	case models.StatusSuccess:
		fmt.Printf("\033[32m[OK]\033[0m         %s\n", res.Target.IP)
	case models.StatusRejected:
		fmt.Printf("\033[31m[REJECTED]\033[0m   %s\n", res.Target.IP)
	case models.StatusTimeout:
		fmt.Printf("\033[33m[TIMEOUT]\033[0m    %s\n", res.Target.IP)
	case models.StatusPermissionDenied:
		fmt.Printf("\033[36m[BLOCKED/DENIED]\033[0m %s (%v)\n", res.Target.IP, res.Err)
	default:
		fmt.Printf("\033[35m[UNKNOWN]\033[0m    %s (%v)\n", res.Target.IP, res.Err)
	}
}
