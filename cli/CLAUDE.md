# CLI Package Instructions

**Scope**: Applies only to `cli/` directory.

This file **extends** the repository-level `.claude/CLAUDE.md`. Read that first for global standards. This file contains only what's **specific** to the cli package.

---

## Purpose

The `cli` package implements **command-line interface logic** for sortTF:

- Flag parsing and validation
- Execution of sorting operations via API calls
- User-facing output formatting (success, errors, progress)
- Exit code management
- Colored terminal output

**Philosophy**:

- Thin wrapper around `api` package
- User-friendly output with color
- Clear error messages
- Proper exit codes for scripting

---

## Package Structure

```text
cli/
├── cli.go              # Main CLI logic
├── cli_test.go         # CLI tests
├── cli_bench_test.go   # Benchmarks
└── CLAUDE.md           # This file
```

---

## CLI Design

### Entry Point

```go
// RunCLI is the main entry point for CLI execution
// Returns exit code (0 = success, 1 = error)
func RunCLI(args []string) int
```

**Called from `cmd/sorttf/main.go`**:

```go
func main() {
 os.Exit(cli.RunCLI(os.Args[1:]))
}
```

### Responsibilities

1. **Parse flags** using Go's `flag` package
2. **Validate inputs** (paths exist, flags valid)
3. **Call API functions** from `api` package
4. **Format output** with colors and symbols
5. **Return exit code** (0 or 1)

---

## Flag Definitions

### Supported Flags

```go
var (
 flagDryRun    bool   // --dry-run, -n: Preview without writing
 flagValidate  bool   // --validate, -c: Check if sorted (CI/CD)
 flagRecursive bool   // --recursive, -r: Process subdirectories
 flagHelp      bool   // --help, -h: Show usage
 flagVersion   bool   // --version, -v: Show version
)
```

### Flag Parsing

```go
fs := flag.NewFlagSet("sorttf", flag.ContinueOnError)
fs.BoolVar(&flagDryRun, "dry-run", false, "preview changes without modifying files")
fs.BoolVar(&flagDryRun, "n", false, "short form of --dry-run")
fs.BoolVar(&flagValidate, "validate", false, "check if files need sorting")
fs.BoolVar(&flagValidate, "c", false, "short form of --validate")
fs.BoolVar(&flagRecursive, "recursive", false, "process directories recursively")
fs.BoolVar(&flagRecursive, "r", false, "short form of --recursive")
fs.BoolVar(&flagHelp, "help", false, "show help message")
fs.BoolVar(&flagHelp, "h", false, "short form of --help")
fs.BoolVar(&flagVersion, "version", false, "show version")
fs.BoolVar(&flagVersion, "v", false, "short form of --version")

err := fs.Parse(args)
```

---

## Output Formatting

### Color Scheme

Use `github.com/fatih/color` for colored output:

```go
var (
 colorGreen  = color.New(color.FgGreen).SprintFunc()
 colorRed    = color.New(color.FgRed).SprintFunc()
 colorYellow = color.New(color.FgYellow).SprintFunc()
 colorCyan   = color.New(color.FgCyan).SprintFunc()
)
```

### Output Patterns

**Success (green checkmark)**:

```go
fmt.Printf("%s Sorted: %s\n", colorGreen("✓"), path)
```

**Error (red X)**:

```go
fmt.Fprintf(os.Stderr, "%s Error: %s\n", colorRed("✗"), err.Error())
```

**Warning (yellow warning)**:

```go
fmt.Printf("%s Already sorted: %s\n", colorYellow("⚠"), path)
```

**Info (cyan info)**:

```go
fmt.Printf("%s Processing: %s\n", colorCyan("ℹ"), path)
```

### Progress Output

For multiple files, show progress:

```go
fmt.Printf("Processing %d files...\n", len(paths))

// Process files

successCount := 0
errorCount := 0
noChangeCount := 0

// After processing
fmt.Printf("\nResults:\n")
fmt.Printf("  %s Sorted: %d\n", colorGreen("✓"), successCount)
fmt.Printf("  %s Already sorted: %d\n", colorYellow("⚠"), noChangeCount)
if errorCount > 0 {
 fmt.Printf("  %s Errors: %d\n", colorRed("✗"), errorCount)
}
```

---

## Exit Codes

### Standard Exit Codes

```go
const (
 ExitSuccess = 0  // All operations succeeded
 ExitError   = 1  // One or more operations failed
)
```

### Exit Code Logic

**Success (0)**:

- All files sorted successfully
- OR all files already sorted
- OR dry-run completed successfully

**Error (1)**:

- One or more files failed to sort
- Invalid flags or arguments
- File not found
- Parse errors
- Validate mode: one or more files need sorting

### Examples

```go
// Success: file sorted
if err == nil {
 return ExitSuccess
}

// Success: file already sorted
if errors.Is(err, api.ErrNoChanges) {
 return ExitSuccess
}

// Error: validation failed
if errors.Is(err, api.ErrNeedsSorting) {
 return ExitError
}

// Error: actual failure
if err != nil {
 return ExitError
}
```

---

## CLI Execution Flow

### 1. Parse Arguments

```go
func RunCLI(args []string) int {
 // Parse flags
 fs := flag.NewFlagSet("sorttf", flag.ContinueOnError)
 // ... define flags

 if err := fs.Parse(args); err != nil {
  fmt.Fprintf(os.Stderr, "Error: %v\n", err)
  return ExitError
 }

 // Handle --help
 if flagHelp {
  printUsage()
  return ExitSuccess
 }

 // Handle --version
 if flagVersion {
  printVersion()
  return ExitSuccess
 }

 // Get positional arguments (paths)
 paths := fs.Args()
 if len(paths) == 0 {
  fmt.Fprintf(os.Stderr, "Error: no paths specified\n")
  printUsage()
  return ExitError
 }

 // Continue processing...
}
```

### 2. Validate Inputs

```go
// Check if paths exist
for _, path := range paths {
 if _, err := os.Stat(path); err != nil {
  fmt.Fprintf(os.Stderr, "Error: path not found: %s\n", path)
  return ExitError
 }
}
```

### 3. Build Options

```go
opts := api.Options{
 DryRun:   flagDryRun,
 Validate: flagValidate,
}
```

### 4. Process Files/Directories

```go
// Single file
if isFile(path) {
 err := api.SortFile(path, opts)
 return handleResult(path, err)
}

// Directory
if isDir(path) {
 results, err := api.SortDirectory(path, flagRecursive, opts)
 if err != nil {
  fmt.Fprintf(os.Stderr, "Error: %v\n", err)
  return ExitError
 }
 return handleResults(results)
}
```

### 5. Handle Results

```go
func handleResult(path string, err error) int {
 if err == nil {
  fmt.Printf("%s Sorted: %s\n", colorGreen("✓"), path)
  return ExitSuccess
 }

 if errors.Is(err, api.ErrNoChanges) {
  fmt.Printf("%s Already sorted: %s\n", colorYellow("⚠"), path)
  return ExitSuccess
 }

 if errors.Is(err, api.ErrNeedsSorting) {
  fmt.Fprintf(os.Stderr, "%s Needs sorting: %s\n", colorRed("✗"), path)
  return ExitError
 }

 fmt.Fprintf(os.Stderr, "%s Error: %s: %v\n", colorRed("✗"), path, err)
 return ExitError
}
```

---

## Testing CLI

### Test Strategy

CLI tests verify:

- Flag parsing
- Input validation
- Correct API calls
- Output formatting
- Exit codes

### Test Patterns

**Table-driven tests**:

```go
func TestRunCLI(t *testing.T) {
 tests := []struct {
  name     string
  args     []string
  wantExit int
  wantOut  string
 }{
  {
   name:     "help flag",
   args:     []string{"--help"},
   wantExit: 0,
   wantOut:  "Usage:",
  },
  {
   name:     "no arguments",
   args:     []string{},
   wantExit: 1,
   wantOut:  "no paths specified",
  },
  // ... more test cases
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   // Capture output
   oldStdout := os.Stdout
   r, w, _ := os.Pipe()
   os.Stdout = w

   // Run CLI
   exitCode := RunCLI(tt.args)

   // Restore stdout
   w.Close()
   os.Stdout = oldStdout

   // Check results
   var buf bytes.Buffer
   io.Copy(&buf, r)
   output := buf.String()

   if exitCode != tt.wantExit {
    t.Errorf("exit code = %d, want %d", exitCode, tt.wantExit)
   }
   if !strings.Contains(output, tt.wantOut) {
    t.Errorf("output = %q, want substring %q", output, tt.wantOut)
   }
  })
 }
}
```

### Integration with API

CLI tests should mock or use temporary files:

```go
func TestRunCLI_SortFile(t *testing.T) {
 // Create temp file
 tmpDir := t.TempDir()
 testFile := filepath.Join(tmpDir, "test.tf")

 // Write unsorted content
 content := `resource "aws_instance" "web" { ami = "ami-123" }
provider "aws" { region = "us-west-2" }`

 err := os.WriteFile(testFile, []byte(content), 0644)
 if err != nil {
  t.Fatal(err)
 }

 // Run CLI
 exitCode := RunCLI([]string{testFile})

 // Verify
 if exitCode != 0 {
  t.Errorf("expected exit code 0, got %d", exitCode)
 }

 // Check file was sorted
 sorted, err := os.ReadFile(testFile)
 if err != nil {
  t.Fatal(err)
 }

 // Verify provider comes before resource
 if !strings.Contains(string(sorted), "provider") {
  t.Error("file should contain provider block")
 }
}
```

### Benchmarks

Benchmark CLI operations:

```go
func BenchmarkRunCLI(b *testing.B) {
 tmpDir := b.TempDir()
 testFile := filepath.Join(tmpDir, "bench.tf")

 // Write test file
 content := generateLargeTFFile()  // Helper function
 os.WriteFile(testFile, []byte(content), 0644)

 b.ResetTimer()
 for i := 0; i < b.N; i++ {
  RunCLI([]string{testFile})
 }
}
```

---

## Error Handling

### User-Friendly Error Messages

Convert technical errors to user-friendly messages:

```go
// Parse error
var parseErr *hcl.ParseError
if errors.As(err, &parseErr) {
 fmt.Fprintf(os.Stderr, "%s Syntax error in %s at line %d\n",
  colorRed("✗"), path, parseErr.Line)
 return ExitError
}

// File not found
if os.IsNotExist(err) {
 fmt.Fprintf(os.Stderr, "%s File not found: %s\n",
  colorRed("✗"), path)
 return ExitError
}

// Permission denied
if os.IsPermission(err) {
 fmt.Fprintf(os.Stderr, "%s Permission denied: %s\n",
  colorRed("✗"), path)
 return ExitError
}

// Generic error
fmt.Fprintf(os.Stderr, "%s Error: %v\n", colorRed("✗"), err)
return ExitError
```

### Stderr vs Stdout

**Stdout**: Normal output, success messages
**Stderr**: Errors and warnings

```go
// Success -> stdout
fmt.Printf("%s Sorted: %s\n", colorGreen("✓"), path)

// Error -> stderr
fmt.Fprintf(os.Stderr, "%s Error: %v\n", colorRed("✗"), err)
```

---

## Help and Usage

### Usage Message

```go
func printUsage() {
 fmt.Println("Usage: sorttf [options] <path>...")
 fmt.Println()
 fmt.Println("Sort and format Terraform and Terragrunt files.")
 fmt.Println()
 fmt.Println("Options:")
 fmt.Println("  -r, --recursive     Process directories recursively")
 fmt.Println("  -n, --dry-run       Preview changes without modifying files")
 fmt.Println("  -c, --validate      Check if files need sorting (exit 1 if unsorted)")
 fmt.Println("  -h, --help          Show this help message")
 fmt.Println("  -v, --version       Show version information")
 fmt.Println()
 fmt.Println("Examples:")
 fmt.Println("  sorttf main.tf              # Sort a single file")
 fmt.Println("  sorttf .                    # Sort files in current directory")
 fmt.Println("  sorttf -r .                 # Sort files recursively")
 fmt.Println("  sorttf -n main.tf           # Preview changes")
 fmt.Println("  sorttf -c main.tf           # Check if file needs sorting")
}
```

### Version Information

```go
const version = "1.0.2"

func printVersion() {
 fmt.Printf("sortTF v%s\n", version)
 fmt.Println("https://github.com/obergerkatz/sortTF")
}
```

---

## Dependencies

**Internal**:

- `github.com/obergerkatz/sortTF/api` - Core API functions
- `github.com/obergerkatz/sortTF/config` - Options type

**External**:

- `github.com/fatih/color` - Colored terminal output
- Standard library: `flag`, `fmt`, `os`, `path/filepath`

---

## Common Patterns

### Handling Multiple Files

```go
results := api.SortFiles(paths, opts)

var exitCode int
for path, err := range results {
 code := handleResult(path, err)
 if code != ExitSuccess {
  exitCode = ExitError
 }
}

return exitCode
```

### Dry Run Mode

```go
if opts.DryRun {
 content, changed, err := api.GetSortedContent(path)
 if err != nil {
  return handleError(err)
 }

 if changed {
  fmt.Printf("%s Would sort: %s\n", colorCyan("ℹ"), path)
  fmt.Println(content)
 } else {
  fmt.Printf("%s Already sorted: %s\n", colorYellow("⚠"), path)
 }

 return ExitSuccess
}
```

### Validate Mode (CI/CD)

```go
if opts.Validate {
 err := api.SortFile(path, opts)

 if errors.Is(err, api.ErrNoChanges) {
  fmt.Printf("%s Valid: %s\n", colorGreen("✓"), path)
  return ExitSuccess
 }

 if errors.Is(err, api.ErrNeedsSorting) {
  fmt.Fprintf(os.Stderr, "%s Needs sorting: %s\n", colorRed("✗"), path)
  return ExitError
 }

 // Other error
 fmt.Fprintf(os.Stderr, "%s Error: %v\n", colorRed("✗"), err)
 return ExitError
}
```

---

## Performance Considerations

### Buffered Output

For large result sets, buffer output:

```go
var output strings.Builder

for path, err := range results {
 if err == nil {
  output.WriteString(fmt.Sprintf("%s %s\n", colorGreen("✓"), path))
 }
}

fmt.Print(output.String())
```

### Avoid Excessive API Calls

Process all paths in one batch:

```go
// GOOD: Single batch operation
results := api.SortFiles(allPaths, opts)

// BAD: Multiple individual calls
for _, path := range allPaths {
 api.SortFile(path, opts)  // Inefficient
}
```

---

## Acceptance Checklist (CLI Package)

Before considering CLI changes complete:

- [ ] All flags documented in help message
- [ ] Exit codes correct (0 = success, 1 = error)
- [ ] Errors written to stderr, success to stdout
- [ ] Colored output used appropriately
- [ ] User-friendly error messages (not technical stack traces)
- [ ] Tests cover all flag combinations
- [ ] Tests verify exit codes
- [ ] Tests check output formatting
- [ ] Benchmarks updated if performance-critical changes
- [ ] Help text updated if flags changed
- [ ] Integration tests pass (`go test ./integration`)
- [ ] CLI behavior documented in docs/USAGE.md
