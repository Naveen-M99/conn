package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"conn/internal/models"
	"conn/internal/repository"
	"conn/internal/service"
)

const version = "1.0.0"

func main() {
	// 1. Define command-line flags
	timeoutSec := flag.Int("t", 3, "Connection timeout in seconds")
	flag.IntVar(timeoutSec, "timeout", 3, "Connection timeout in seconds")
	concurrency := flag.Int("c", 50, "Number of concurrent workers")
	showVersion := flag.Bool("v", false, "Print tool version")
	flag.BoolVar(showVersion, "version", false, "Print tool version")

	// Custom usage message for --help or -h
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "conn - Multi-threaded Network Connectivity Prober\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  conn [OPTIONS] <ip_list_file> <port>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  conn ips.txt 22\n")
		fmt.Fprintf(os.Stderr, "  conn -t 5 -c 100 ips.txt 443\n")
	}

	flag.Parse()

	// 2. Handle version flag
	if *showVersion {
		fmt.Printf("conn version %s\n", version)
		os.Exit(0)
	}

	// 3. Validate positional arguments
	args := flag.Args()
	if len(args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	filePath := args[0]
	var port int
	if _, err := fmt.Sscanf(args[1], "%d", &port); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid port number '%s'\n", args[1])
		os.Exit(1)
	}

	cfg := models.ScanConfig{
		FilePath:    filePath,
		Port:        port,
		Timeout:     time.Duration(*timeoutSec) * time.Second,
		Concurrency: *concurrency,
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// 4. Trap OS interrupts (Ctrl+C) for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 5. Initialize the file output writer (repository layer)
	repo, err := repository.NewFileWriter()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to flush all buffers to disk: %v\n", err)
		}
	}()

	// 6. Initialize and start the scanner service (business logic layer)
	scanner := service.NewScannerService(repo)

	fmt.Printf("Starting scan on target port %d using %d concurrent workers (Timeout: %s)...\n",
		cfg.Port, cfg.Concurrency, cfg.Timeout)

	if err := scanner.Run(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Scan terminated with error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nScan finished successfully. Output files saved to current directory.")
}
