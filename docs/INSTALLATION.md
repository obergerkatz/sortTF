# Installation Guide

This guide covers all the ways to install sortTF on your system.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Using go install (Recommended)](#using-go-install-recommended)
- [Download Pre-built Binaries](#download-pre-built-binaries)
- [Building from Source](#building-from-source)
- [Docker](#docker)
- [Verification](#verification)
- [Upgrading](#upgrading)
- [Uninstallation](#uninstallation)

## Prerequisites

- **For pre-built binaries**: No prerequisites
- **For `go install` or building from source**: Go 1.22+ (Go 1.23+ recommended)
- **For Docker**: Docker installed on your system

## Using go install (Recommended)

The easiest way to install sortTF if you have Go installed:

```bash
go install github.com/obergerkatz/sortTF/cmd/sorttf@latest
```

This installs the latest version to `$GOPATH/bin` (usually `~/go/bin`).

### Install Specific Version

```bash
# Install specific tagged version
go install github.com/obergerkatz/sortTF/cmd/sorttf@v1.0.0

# Install from specific commit
go install github.com/obergerkatz/sortTF/cmd/sorttf@abc1234
```

### Verify Installation

```bash
sorttf --help
```

If you get "command not found", add `$GOPATH/bin` to your PATH:

```bash
# Add to ~/.bashrc, ~/.zshrc, or equivalent
export PATH=$PATH:$(go env GOPATH)/bin
```

## Download Pre-built Binaries

Pre-built binaries are available for Linux, macOS, and Windows from the [releases page](https://github.com/obergerkatz/sortTF/releases).

### Linux

```bash
# AMD64
wget https://github.com/obergerkatz/sortTF/releases/download/v1.0.0/sorttf-linux-amd64
chmod +x sorttf-linux-amd64
sudo mv sorttf-linux-amd64 /usr/local/bin/sorttf

# ARM64
wget https://github.com/obergerkatz/sortTF/releases/download/v1.0.0/sorttf-linux-arm64
chmod +x sorttf-linux-arm64
sudo mv sorttf-linux-arm64 /usr/local/bin/sorttf
```

### macOS

```bash
# Intel
wget https://github.com/obergerkatz/sortTF/releases/download/v1.0.0/sorttf-darwin-amd64
chmod +x sorttf-darwin-amd64
sudo mv sorttf-darwin-amd64 /usr/local/bin/sorttf

# Apple Silicon (M1/M2/M3)
wget https://github.com/obergerkatz/sortTF/releases/download/v1.0.0/sorttf-darwin-arm64
chmod +x sorttf-darwin-arm64
sudo mv sorttf-darwin-arm64 /usr/local/bin/sorttf
```

**Note for macOS**: You may need to allow the binary in System Preferences > Security & Privacy if you get a security warning.

### Windows

1. Download `sorttf-windows-amd64.exe` from the [releases page](https://github.com/obergerkatz/sortTF/releases)
2. Rename to `sorttf.exe`
3. Move to a directory in your PATH (e.g., `C:\Windows\System32` or add a custom directory to PATH)

**PowerShell:**
```powershell
# Download (replace v1.0.0 with desired version)
Invoke-WebRequest -Uri "https://github.com/obergerkatz/sortTF/releases/download/v1.0.0/sorttf-windows-amd64.exe" -OutFile "sorttf.exe"

# Move to a directory in PATH
Move-Item sorttf.exe C:\Windows\System32\
```

## Building from Source

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/obergerkatz/sortTF.git
cd sortTF

# Build for current platform
go build -o sorttf ./cmd/sorttf

# Install to $GOPATH/bin
go install ./cmd/sorttf

# Or move to system path
sudo mv sorttf /usr/local/bin/
```

### Build for Specific Platform

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o sorttf-linux ./cmd/sorttf

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o sorttf-macos ./cmd/sorttf

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o sorttf.exe ./cmd/sorttf
```

### Build with Version Information

```bash
# Build with version, commit, and build date embedded
VERSION="v1.0.0"
COMMIT=$(git rev-parse HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

go build \
  -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
  -o sorttf \
  ./cmd/sorttf
```

## Docker

### Pull from GitHub Container Registry

```bash
# Pull latest
docker pull ghcr.io/obergerkatz/sorttf:latest

# Pull specific version
docker pull ghcr.io/obergerkatz/sorttf:v1.0.0
```

### Build Docker Image

```bash
# Clone repository
git clone https://github.com/obergerkatz/sortTF.git
cd sortTF

# Build image
docker build -t sorttf:local .
```

### Run with Docker

```bash
# Sort files in current directory
docker run --rm -v $(pwd):/workspace sorttf:latest .

# With options
docker run --rm -v $(pwd):/workspace sorttf:latest --recursive --dry-run .
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'
services:
  sorttf:
    image: ghcr.io/obergerkatz/sorttf:latest
    volumes:
      - ./terraform:/workspace
    command: ["--recursive", "."]
```

Run:
```bash
docker-compose run sorttf
```

## Verification

### Verify Binary Integrity

Download the checksums file:

```bash
wget https://github.com/obergerkatz/sortTF/releases/download/v1.0.0/checksums.txt
```

Verify (Linux/macOS):
```bash
sha256sum -c checksums.txt
```

Verify (macOS alternative):
```bash
shasum -a 256 -c checksums.txt
```

### Test Installation

```bash
# Check version
sorttf --version

# Run help
sorttf --help

# Test on a file
echo 'resource "test" "example" { name = "test" }' > test.tf
sorttf test.tf
cat test.tf
rm test.tf
```

## Upgrading

### Using go install

```bash
go install github.com/obergerkatz/sortTF/cmd/sorttf@latest
```

### Using Pre-built Binaries

Download and replace the binary following the same steps as installation.

### Check Current Version

```bash
sorttf --version
```

## Uninstallation

### Installed via go install

```bash
rm $(which sorttf)
# Or
rm $GOPATH/bin/sorttf
```

### Installed to /usr/local/bin

```bash
sudo rm /usr/local/bin/sorttf
```

### Installed via package manager

Follow the package manager's uninstallation process.

### Docker

```bash
docker rmi sorttf:latest
# Or
docker rmi ghcr.io/obergerkatz/sorttf:latest
```

## Troubleshooting

### "command not found" after installation

- **go install**: Add `$GOPATH/bin` to your PATH
- **Manual install**: Ensure `/usr/local/bin` is in your PATH
- **Windows**: Add the directory containing `sorttf.exe` to your PATH

### Permission denied (Linux/macOS)

```bash
chmod +x sorttf
```

### macOS security warning

Go to System Preferences > Security & Privacy > General and click "Allow" when you see the warning about sortTF.

Alternatively, remove the quarantine attribute:
```bash
xattr -d com.apple.quarantine sorttf
```

### "cannot execute binary file" (Linux)

You may have downloaded the wrong architecture. Check your system:

```bash
uname -m
# x86_64 = amd64
# aarch64 = arm64
```

Download the correct binary for your architecture.

## Next Steps

- Read the [Usage Guide](USAGE.md) to learn how to use sortTF
- See [Contributing Guide](CONTRIBUTING.md) if you want to contribute
