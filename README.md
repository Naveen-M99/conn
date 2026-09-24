# conn

A high-performance, concurrent TCP network connectivity prober built in Go.

`conn` reads a list of IP addresses or hostnames, tests TCP socket connectivity against a target port, and sorts the results into 5 distinct status files. It runs as a self-contained, zero-dependency binary across Linux and Windows (x86 and x64).

---

## Features

- **Zero Runtime Dependencies:** Pure compiled machine code; requires no Python, Node.js, Netcat, or runtime libraries.
- **Concurrent Engine:** Goroutine-backed worker pool enables scanning thousands of hosts in seconds.
- **Granular Classification:** Separates dropped packets, active resets, and OS/firewall blocks instead of returning a generic failure.
- **Memory & Resource Safe:** Buffered I/O reduces syscall overhead; bounds file descriptors to avoid socket exhaustion.
- **Graceful Interrupt Handling:** Captures `Ctrl+C` / `SIGINT` signals and cleanly flushes in-memory buffers to disk before exiting.

---

## Installation

### Linux (One-Line Installer)

Installs the native architecture binary (`amd64` or `386`) into `/usr/local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/Naveen-M99/conn/main/install.sh | bash
```

### Windows (One-Line Installer)
```powershell
irm https://raw.githubusercontent.com/Naveen-M99/conn/main/install.ps1 | iex
```