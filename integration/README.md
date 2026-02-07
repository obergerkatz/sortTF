# Integration Tests

This directory contains end-to-end integration tests for sortTF that run against the actual compiled binary.

## Overview

These integration tests differ from unit tests in that they:

1. **Run the actual binary** - Tests execute the compiled `sorttf` binary as a subprocess, not Go functions
2. **Test the entire workflow** - From CLI argument parsing through to file system modifications
3. **Use realistic scenarios** - Tests include complex, real-world Terraform configurations
4. **Verify file system state** - Tests check that files are actually modified correctly on disk
5. **Test error handling** - Comprehensive coverage of error scenarios and edge cases

## Running Integration Tests

```bash
# Run all integration tests
go test ./integration/...

# Run with verbose output
go test -v ./integration/...

# Run a specific test
go test -v ./integration/... -run TestIntegration_ComplexRealWorld

# Run tests with coverage
go test -cover ./integration/...
```

## Test Categories

### Basic Operations

- **TestIntegration_Help** - Verifies help command displays usage information
- **TestIntegration_SingleFile** - Tests sorting a single unsorted file
- **TestIntegration_AlreadySorted** - Verifies already-sorted files are handled correctly
- **TestIntegration_EmptyDirectory** - Tests behavior with empty directories

### CLI Modes

- **TestIntegration_DryRun** - Validates dry-run mode doesn't modify files
- **TestIntegration_Validate** - Tests validation mode for CI/CD pipelines
- **TestIntegration_VerboseMode** - Verifies verbose output includes processing details
- **TestIntegration_Recursive** - Tests recursive directory processing

### File Processing

- **TestIntegration_MixedFileTypes** - Tests directory with .tf, .hcl, and non-Terraform files
- **TestIntegration_ComplexRealWorld** - Realistic complex Terraform configuration with:
  - terraform blocks with backend configuration
  - provider configuration
  - multiple variables
  - locals
  - data sources
  - multiple resources
  - module invocations
  - multiple outputs

### Error Handling

- **TestIntegration_InvalidSyntax** - Tests handling of files with syntax errors
- **TestIntegration_NonExistentPath** - Tests error handling for missing files/directories

## Test Structure

Each test follows this pattern:

1. **Setup** - Create temporary directory and test files
2. **Execute** - Run the sorttf binary with specific arguments
3. **Verify** - Check exit code, stdout/stderr output, and file system state
4. **Cleanup** - Automatic cleanup via `t.TempDir()`

## Binary Building

The test suite automatically builds the binary before running tests (in `TestMain`):

```go
func TestMain(m *testing.M) {
    // Build binary
    cmd := exec.Command("go", "build", "-o", "sorttf-test", "..")
    cmd.Run()

    // Run tests
    code := m.Run()

    // Cleanup
    os.Remove("sorttf-test")
    os.Exit(code)
}
```

## CI/CD Usage

These integration tests are designed to run in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run integration tests
  run: go test -v ./integration/...
```

The tests verify the actual behavior users will experience, making them ideal for:

- Pre-release validation
- Regression testing
- Cross-platform compatibility testing
- Performance benchmarking

## Adding New Tests

When adding new integration tests:

1. Follow the naming convention: `TestIntegration_<Feature>`
2. Use `t.TempDir()` for temporary files
3. Use the `runSortTF()` helper to execute the binary
4. Verify both output and file system state
5. Test error cases and edge conditions
6. Document what the test verifies

Example:

```go
func TestIntegration_NewFeature(t *testing.T) {
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.tf")

    // Create test file
    content := `...`
    os.WriteFile(testFile, []byte(content), 0644)

    // Run sorttf
    stdout, stderr, exitCode := runSortTF(t, testFile)

    // Verify results
    if exitCode != 0 {
        t.Errorf("expected exit code 0, got %d", exitCode)
    }
    // ... more assertions
}
```

## Coverage

Integration tests provide coverage for:

- ✅ CLI argument parsing and flag handling
- ✅ File discovery and traversal
- ✅ HCL parsing and validation
- ✅ Block sorting logic
- ✅ Attribute sorting within blocks
- ✅ File writing and atomic updates
- ✅ Error messages and exit codes
- ✅ Dry-run mode
- ✅ Validation mode (for CI/CD)
- ✅ Recursive processing
- ✅ Mixed file type handling
- ✅ Complex real-world configurations

## Comparison with Unit Tests

| Aspect | Unit Tests | Integration Tests |
|--------|-----------|-------------------|
| Scope | Individual functions/packages | Entire application |
| Execution | Direct function calls | Binary subprocess |
| Speed | Fast (~0.01s per test) | Slower (~0.03s per test) |
| Setup | Minimal | Full binary build |
| Isolation | High (mocked dependencies) | Low (real file system) |
| Purpose | Code correctness | User experience |

Both types of tests are valuable:

- **Unit tests** catch bugs early during development
- **Integration tests** verify the system works end-to-end for users
