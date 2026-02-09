# CMD Package Instructions

**Scope**: Applies only to `cmd/` directory.

This file **extends** the repository-level `.claude/CLAUDE.md`. Read that first for global standards. This file contains only what's **specific** to the cmd package.

---

## Purpose

The `cmd/` directory contains **command-line entry points** for sortTF:

- `cmd/sorttf/main.go`: Main CLI entry point
- Future commands can be added here (e.g., `cmd/sorttf-server/`, `cmd/sorttf-lsp/`)

**Philosophy**:

- Minimal logic in main.go
- Delegate to `cli` package for all functionality
- Handle only process-level concerns (exit codes, signals)
- Never implement business logic here

---

## Package Structure

```text
cmd/
└── sorttf/
    └── main.go          # CLI entry point
```

---

## Main Entry Point

### main.go Structure

```go
package main

import (
 "os"

 "github.com/obergerkatz/sortTF/cli"
)

func main() {
 // Pass command-line arguments to CLI
 // Exit with the code returned by CLI
 os.Exit(cli.RunCLI(os.Args[1:]))
}
```

**That's it.** No other logic should be in `main.go`.

---

## Design Principles

### 1. Minimal Entry Point

**Why**: Keep `main()` simple and untestable code minimal.

- All logic in `cli` package (testable)
- `main()` just wires things together
- No flag parsing, no output, no business logic

### 2. Pass Args Explicitly

```go
// GOOD: Pass args explicitly
os.Exit(cli.RunCLI(os.Args[1:]))

// BAD: CLI reads os.Args internally
os.Exit(cli.RunCLI())
```

**Why**: Makes CLI testable without process-level mocking.

### 3. Exit Code Handling

```go
// CLI returns exit code
exitCode := cli.RunCLI(os.Args[1:])

// Main calls os.Exit with that code
os.Exit(exitCode)
```

**Why**: Allows CLI to be tested without actually exiting process.

---

## Building and Installing

### Build Commands

```bash
# Build for current platform
go build -o bin/sorttf ./cmd/sorttf

# Build with version info
go build -ldflags "-X main.version=1.0.2" -o bin/sorttf ./cmd/sorttf

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o bin/sorttf-linux ./cmd/sorttf
GOOS=darwin GOARCH=amd64 go build -o bin/sorttf-darwin ./cmd/sorttf
GOOS=windows GOARCH=amd64 go build -o bin/sorttf.exe ./cmd/sorttf

# Install to $GOPATH/bin
go install ./cmd/sorttf
```

### Build Flags

**Optimization**:

```bash
# Smaller binary
go build -ldflags="-s -w" -o bin/sorttf ./cmd/sorttf
```

**Version embedding**:

```bash
# Embed version at build time
go build -ldflags "-X main.version=$(git describe --tags)" ./cmd/sorttf
```

---

## Testing

### Why main.go Has No Tests

`main()` is not directly testable because it calls `os.Exit()`.

**Instead**: Test `cli.RunCLI()` extensively in `cli/cli_test.go`.

### Integration Tests

The `integration/` package tests the **compiled binary**:

```go
func TestCLI_Integration(t *testing.T) {
 // Build binary
 binary := buildTestBinary(t)

 // Run binary as subprocess
 cmd := exec.Command(binary, "--help")
 output, err := cmd.CombinedOutput()

 // Verify output and exit code
 if err != nil {
  t.Errorf("unexpected error: %v", err)
 }
 if !strings.Contains(string(output), "Usage:") {
  t.Errorf("expected help output, got: %s", output)
 }
}
```

---

## Version Management

### Embedding Version Information

**Option 1**: Version constant in CLI package

```go
// In cli/cli.go
const Version = "1.0.2"

func printVersion() {
 fmt.Printf("sortTF v%s\n", Version)
}
```

**Option 2**: Build-time injection

```go
// In cmd/sorttf/main.go
package main

import (
 "os"

 "github.com/obergerkatz/sortTF/cli"
)

var version = "dev"  // Overridden at build time

func main() {
 cli.Version = version  // Set version in CLI package
 os.Exit(cli.RunCLI(os.Args[1:]))
}
```

Build with:

```bash
go build -ldflags "-X main.version=1.0.2" ./cmd/sorttf
```

---

## Signal Handling (Future Enhancement)

If graceful shutdown needed:

```go
package main

import (
 "context"
 "os"
 "os/signal"
 "syscall"

 "github.com/obergerkatz/sortTF/cli"
)

func main() {
 // Create context that cancels on interrupt
 ctx, cancel := context.WithCancel(context.Background())
 defer cancel()

 // Handle interrupt signals
 sigChan := make(chan os.Signal, 1)
 signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

 go func() {
  <-sigChan
  cancel()  // Cancel context on signal
 }()

 // Pass context to CLI
 os.Exit(cli.RunCLIWithContext(ctx, os.Args[1:]))
}
```

**Note**: Currently not needed for sortTF (operations are fast).

---

## Multi-Command Setup (Future)

If adding multiple commands:

```text
cmd/
├── sorttf/        # Main CLI
│   └── main.go
├── sorttf-server/ # Future: Server mode
│   └── main.go
└── sorttf-lsp/    # Future: LSP server
    └── main.go
```

Each has its own `main.go` with minimal logic:

```go
// cmd/sorttf-server/main.go
package main

import (
 "os"

 "github.com/obergerkatz/sortTF/server"
)

func main() {
 os.Exit(server.RunServer(os.Args[1:]))
}
```

---

## Cross-Platform Considerations

### Windows Support

```go
// Works on all platforms
os.Exit(cli.RunCLI(os.Args[1:]))
```

**No special handling needed** for basic CLI.

### Platform-Specific Builds

Use build tags if needed:

```go
// +build !windows

package main

// Unix-specific code
```

```go
// +build windows

package main

// Windows-specific code
```

**Currently not needed** for sortTF.

---

## CI/CD Integration

### GitHub Actions Build

```yaml
- name: Build binaries
  run: |
    GOOS=linux GOARCH=amd64 go build -o bin/sorttf-linux-amd64 ./cmd/sorttf
    GOOS=darwin GOARCH=amd64 go build -o bin/sorttf-darwin-amd64 ./cmd/sorttf
    GOOS=windows GOARCH=amd64 go build -o bin/sorttf-windows-amd64.exe ./cmd/sorttf
```

### Release Binaries

See `.github/workflows/release.yml` for automated release builds.

---

## Dependencies

**Internal**:

- `github.com/obergerkatz/sortTF/cli` - CLI logic

**External**:

- None (only standard library `os`)

---

## Common Mistakes to Avoid

### ❌ Adding Logic to main()

```go
// BAD: Business logic in main
func main() {
 if len(os.Args) < 2 {
  fmt.Println("Error: no arguments")
  os.Exit(1)
 }
 // ... more logic
}
```

```go
// GOOD: Delegate to CLI
func main() {
 os.Exit(cli.RunCLI(os.Args[1:]))
}
```

### ❌ Direct Flag Parsing

```go
// BAD: Parse flags in main
func main() {
 flag.Parse()
 // ...
}
```

```go
// GOOD: CLI handles flags
func main() {
 os.Exit(cli.RunCLI(os.Args[1:]))
}
```

### ❌ Output in main()

```go
// BAD: Print from main
func main() {
 fmt.Println("Starting sortTF...")
 os.Exit(cli.RunCLI(os.Args[1:]))
}
```

```go
// GOOD: No output in main
func main() {
 os.Exit(cli.RunCLI(os.Args[1:]))
}
```

---

## Acceptance Checklist (CMD Package)

Before considering cmd changes complete:

- [ ] `main()` function is minimal (≤5 lines)
- [ ] No business logic in main.go
- [ ] All logic delegated to `cli` package
- [ ] Args passed explicitly to CLI
- [ ] Exit code properly returned from CLI
- [ ] No direct output (stdout/stderr) in main()
- [ ] No flag parsing in main()
- [ ] Builds successfully on all target platforms
- [ ] Integration tests pass (`go test ./integration`)
- [ ] Binary is executable and functional
