# Integration Tests Instructions

**Scope**: Applies only to `integration/` directory.

This file **extends** the repository-level `.claude/CLAUDE.md`. Read that first for global standards. This file contains only what's **specific** to integration tests.

---

## Purpose

The `integration/` directory contains **end-to-end integration tests** for sortTF:

- Test the compiled binary (not just library code)
- Verify CLI behavior and exit codes
- Test real file I/O and directory operations
- Validate error messages and output formatting
- Ensure cross-platform compatibility

**Philosophy**:

- Test user workflows end-to-end
- Use real files and processes
- Verify actual binary behavior
- Complement unit tests with realistic scenarios

---

## Package Structure

```text
integration/
├── integration_test.go   # Integration tests
└── CLAUDE.md             # This file
```

---

## Integration Test Strategy

### What Integration Tests Cover

**✅ Integration tests verify**:

- CLI flags and argument parsing
- File sorting end-to-end
- Directory traversal
- Dry-run mode
- Validate mode
- Error handling and messages
- Exit codes
- Multi-file operations
- Recursive directory processing

**❌ Integration tests don't verify**:

- Internal package logic (covered by unit tests)
- HCL parsing details (covered by hcl package tests)
- API functions directly (covered by api package tests)

### Integration vs Unit Tests

**Unit tests** (`*_test.go` in each package):

- Test individual functions
- Fast, isolated
- Mock dependencies
- 90%+ coverage

**Integration tests** (`integration/integration_test.go`):

- Test full workflows
- Slower, realistic
- Real file I/O
- Verify user experience

---

## Test Structure

### Building Test Binary

```go
func buildTestBinary(t *testing.T) string {
 t.Helper()

 // Build in temporary directory
 tmpDir := t.TempDir()
 binaryPath := filepath.Join(tmpDir, "sorttf")

 // Build command
 cmd := exec.Command("go", "build", "-o", binaryPath, "../cmd/sorttf")
 output, err := cmd.CombinedOutput()
 if err != nil {
  t.Fatalf("failed to build test binary: %v\n%s", err, output)
 }

 return binaryPath
}
```

### Running Test Binary

```go
func runSortTF(t *testing.T, binary string, args ...string) (stdout, stderr string, exitCode int) {
 t.Helper()

 cmd := exec.Command(binary, args...)

 var outBuf, errBuf bytes.Buffer
 cmd.Stdout = &outBuf
 cmd.Stderr = &errBuf

 err := cmd.Run()

 stdout = outBuf.String()
 stderr = errBuf.String()

 // Get exit code
 if err != nil {
  if exitErr, ok := err.(*exec.ExitError); ok {
   exitCode = exitErr.ExitCode()
  } else {
   t.Fatalf("unexpected error running command: %v", err)
  }
 } else {
  exitCode = 0
 }

 return stdout, stderr, exitCode
}
```

---

## Test Examples

### Test Help Flag

```go
func TestIntegration_Help(t *testing.T) {
 binary := buildTestBinary(t)

 stdout, stderr, exitCode := runSortTF(t, binary, "--help")

 if exitCode != 0 {
  t.Errorf("expected exit code 0, got %d", exitCode)
 }

 if !strings.Contains(stdout, "Usage:") {
  t.Errorf("expected help output, got:\n%s", stdout)
 }

 if stderr != "" {
  t.Errorf("expected no stderr, got:\n%s", stderr)
 }
}
```

### Test Single File Sorting

```go
func TestIntegration_SortFile(t *testing.T) {
 binary := buildTestBinary(t)
 tmpDir := t.TempDir()

 // Create test file with unsorted content
 testFile := filepath.Join(tmpDir, "test.tf")
 unsortedContent := `resource "aws_instance" "web" {
  ami = "ami-123"
}
provider "aws" {
  region = "us-west-2"
}
`
 err := os.WriteFile(testFile, []byte(unsortedContent), 0644)
 if err != nil {
  t.Fatal(err)
 }

 // Run sortTF
 stdout, stderr, exitCode := runSortTF(t, binary, testFile)

 // Verify exit code
 if exitCode != 0 {
  t.Errorf("expected exit code 0, got %d\nstdout: %s\nstderr: %s",
   exitCode, stdout, stderr)
 }

 // Verify file was modified
 sortedContent, err := os.ReadFile(testFile)
 if err != nil {
  t.Fatal(err)
 }

 // Verify provider comes before resource
 sortedStr := string(sortedContent)
 providerIdx := strings.Index(sortedStr, "provider")
 resourceIdx := strings.Index(sortedStr, "resource")

 if providerIdx == -1 || resourceIdx == -1 {
  t.Fatalf("expected both provider and resource blocks in output:\n%s", sortedStr)
 }

 if providerIdx > resourceIdx {
  t.Errorf("provider should come before resource:\n%s", sortedStr)
 }
}
```

### Test Dry Run Mode

```go
func TestIntegration_DryRun(t *testing.T) {
 binary := buildTestBinary(t)
 tmpDir := t.TempDir()

 // Create test file
 testFile := filepath.Join(tmpDir, "test.tf")
 originalContent := `resource "aws_instance" "web" {}
provider "aws" {}`

 err := os.WriteFile(testFile, []byte(originalContent), 0644)
 if err != nil {
  t.Fatal(err)
 }

 // Run with --dry-run
 stdout, stderr, exitCode := runSortTF(t, binary, "--dry-run", testFile)

 // Verify exit code
 if exitCode != 0 {
  t.Errorf("expected exit code 0, got %d", exitCode)
 }

 // Verify file was NOT modified
 currentContent, err := os.ReadFile(testFile)
 if err != nil {
  t.Fatal(err)
 }

 if string(currentContent) != originalContent {
  t.Error("file should not be modified in dry-run mode")
 }

 // Verify output shows sorted content
 if !strings.Contains(stdout, "provider") {
  t.Error("stdout should show sorted content")
 }
}
```

### Test Validate Mode

```go
func TestIntegration_Validate_NeedsSorting(t *testing.T) {
 binary := buildTestBinary(t)
 tmpDir := t.TempDir()

 // Create unsorted file
 testFile := filepath.Join(tmpDir, "test.tf")
 unsortedContent := `resource "aws_instance" "web" {}
provider "aws" {}`

 err := os.WriteFile(testFile, []byte(unsortedContent), 0644)
 if err != nil {
  t.Fatal(err)
 }

 // Run with --validate
 stdout, stderr, exitCode := runSortTF(t, binary, "--validate", testFile)

 // Should exit with code 1 (needs sorting)
 if exitCode != 1 {
  t.Errorf("expected exit code 1, got %d", exitCode)
 }

 // Should indicate file needs sorting
 output := stdout + stderr
 if !strings.Contains(output, "needs sorting") && !strings.Contains(output, "Needs sorting") {
  t.Errorf("expected 'needs sorting' message, got:\n%s", output)
 }
}

func TestIntegration_Validate_AlreadySorted(t *testing.T) {
 binary := buildTestBinary(t)
 tmpDir := t.TempDir()

 // Create already sorted file
 testFile := filepath.Join(tmpDir, "test.tf")
 sortedContent := `provider "aws" {}
resource "aws_instance" "web" {}
`

 err := os.WriteFile(testFile, []byte(sortedContent), 0644)
 if err != nil {
  t.Fatal(err)
 }

 // Run with --validate
 stdout, stderr, exitCode := runSortTF(t, binary, "--validate", testFile)

 // Should exit with code 0 (already sorted)
 if exitCode != 0 {
  t.Errorf("expected exit code 0, got %d\nstdout: %s\nstderr: %s",
   exitCode, stdout, stderr)
 }
}
```

### Test Recursive Directory

```go
func TestIntegration_RecursiveDirectory(t *testing.T) {
 binary := buildTestBinary(t)
 tmpDir := t.TempDir()

 // Create directory structure
 //   tmpDir/
 //     main.tf
 //     subdir/
 //       variables.tf

 mainFile := filepath.Join(tmpDir, "main.tf")
 os.WriteFile(mainFile, []byte(`resource "aws_instance" "web" {}`), 0644)

 subDir := filepath.Join(tmpDir, "subdir")
 os.Mkdir(subDir, 0755)
 varFile := filepath.Join(subDir, "variables.tf")
 os.WriteFile(varFile, []byte(`variable "region" {}`), 0644)

 // Run with --recursive
 stdout, stderr, exitCode := runSortTF(t, binary, "--recursive", tmpDir)

 if exitCode != 0 {
  t.Errorf("expected exit code 0, got %d\nstdout: %s\nstderr: %s",
   exitCode, stdout, stderr)
 }

 // Verify output mentions both files
 output := stdout + stderr
 if !strings.Contains(output, "main.tf") {
  t.Error("output should mention main.tf")
 }
 if !strings.Contains(output, "variables.tf") {
  t.Error("output should mention variables.tf")
 }
}
```

### Test Error Cases

```go
func TestIntegration_FileNotFound(t *testing.T) {
 binary := buildTestBinary(t)

 stdout, stderr, exitCode := runSortTF(t, binary, "nonexistent.tf")

 // Should fail with exit code 1
 if exitCode != 1 {
  t.Errorf("expected exit code 1, got %d", exitCode)
 }

 // Should have error message
 if stderr == "" && stdout == "" {
  t.Error("expected error message")
 }
}

func TestIntegration_InvalidHCL(t *testing.T) {
 binary := buildTestBinary(t)
 tmpDir := t.TempDir()

 // Create file with invalid HCL
 testFile := filepath.Join(tmpDir, "invalid.tf")
 invalidContent := `resource "aws_instance" "web" {
  ami = "unclosed-string
}
`
 err := os.WriteFile(testFile, []byte(invalidContent), 0644)
 if err != nil {
  t.Fatal(err)
 }

 // Run sortTF
 stdout, stderr, exitCode := runSortTF(t, binary, testFile)

 // Should fail with exit code 1
 if exitCode != 1 {
  t.Errorf("expected exit code 1, got %d", exitCode)
 }

 // Should have error message about parsing
 output := stdout + stderr
 if !strings.Contains(strings.ToLower(output), "error") {
  t.Errorf("expected error message, got:\n%s", output)
 }
}
```

---

## Running Integration Tests

### Run All Tests

```bash
go test ./integration
```

### Run Specific Test

```bash
go test ./integration -run TestIntegration_Help
```

### Run with Verbose Output

```bash
go test -v ./integration
```

### Run with Race Detector

```bash
go test -race ./integration
```

---

## Best Practices

### 1. Build Binary Once

```go
func TestIntegration(t *testing.T) {
 // Build once for all subtests
 binary := buildTestBinary(t)

 t.Run("help", func(t *testing.T) {
  runSortTF(t, binary, "--help")
 })

 t.Run("version", func(t *testing.T) {
  runSortTF(t, binary, "--version")
 })
}
```

### 2. Use Temporary Directories

```go
func TestIntegration_Example(t *testing.T) {
 tmpDir := t.TempDir()  // Automatically cleaned up
 testFile := filepath.Join(tmpDir, "test.tf")
 // ... use tmpDir
}
```

### 3. Check Exit Codes

Always verify exit codes:

```go
if exitCode != 0 {
 t.Errorf("expected exit code 0, got %d", exitCode)
}
```

### 4. Test Stdout and Stderr Separately

```go
// Success output -> stdout
if !strings.Contains(stdout, "Sorted:") {
 t.Error("expected success message in stdout")
}

// Error output -> stderr
if !strings.Contains(stderr, "Error:") {
 t.Error("expected error message in stderr")
}
```

### 5. Verify File Content Changes

```go
// Read file before
originalContent, _ := os.ReadFile(testFile)

// Run sortTF
runSortTF(t, binary, testFile)

// Read file after
newContent, _ := os.ReadFile(testFile)

// Verify changes
if bytes.Equal(originalContent, newContent) {
 t.Error("file content should have changed")
}
```

---

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run integration tests
  run: go test -v ./integration
```

### Test on Multiple Platforms

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
runs-on: ${{ matrix.os }}
steps:
  - name: Run integration tests
    run: go test ./integration
```

---

## Acceptance Checklist (Integration Tests)

Before considering integration tests complete:

- [ ] Tests build and run real binary
- [ ] Tests use temporary directories (cleaned up automatically)
- [ ] Exit codes verified for all scenarios
- [ ] Stdout and stderr checked appropriately
- [ ] Success cases tested (files sorted correctly)
- [ ] Error cases tested (invalid input, file not found)
- [ ] CLI flags tested (--help, --version, --dry-run, --validate, --recursive)
- [ ] Multi-file and directory operations tested
- [ ] Cross-platform compatibility verified (if applicable)
- [ ] Tests run quickly (use minimal fixtures)
- [ ] Tests are deterministic and reproducible
