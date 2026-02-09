# Internal Package Instructions

**Scope**: Applies only to `internal/` directory and its subdirectories.

This file **extends** the repository-level `.claude/CLAUDE.md`. Read that first for global standards. This file contains only what's **specific** to the internal package.

---

## Purpose

The `internal/` directory contains **private utility packages** for sortTF:

- `internal/errors/`: Error handling utilities and custom error types
- `internal/files/`: File system operations and directory walking

**Philosophy**:

- Shared utilities used across multiple packages
- Not importable by external projects (Go internal package convention)
- Well-tested and documented (used internally, but still important)
- Keep focused and minimal (avoid dumping ground)

---

## Package Structure

```text
internal/
├── errors/
│   ├── errors.go        # Error utilities
│   ├── errors_test.go   # Error tests
│   └── CLAUDE.md        # (this would be separate)
├── files/
│   ├── files.go         # File operations
│   ├── files_test.go    # File tests
│   └── CLAUDE.md        # (this would be separate)
└── CLAUDE.md            # This file
```

---

## Internal Package Convention

### Go's Internal Package Rule

Packages in `internal/` directories are **only importable** by code in the same module.

**Can import**:

```go
// In github.com/obergerkatz/sortTF/api
import "github.com/obergerkatz/sortTF/internal/files"  // ✅ OK
```

**Cannot import**:

```go
// In external project
import "github.com/obergerkatz/sortTF/internal/files"  // ❌ Compile error
```

### Why Use Internal?

- **Hide implementation details**: Users shouldn't depend on these
- **Freedom to refactor**: Can change without breaking external users
- **Clear API boundary**: Only packages outside `internal/` are public API

---

## Internal/Errors Package

### Purpose

Custom error types and error handling utilities.

### Key Types

```go
// MultiError aggregates multiple errors
type MultiError struct {
 Errors []error
}

func (e *MultiError) Error() string {
 // Format multiple errors
}

func (e *MultiError) Unwrap() []error {
 return e.Errors
}

// ParseError represents HCL parsing errors
type ParseError struct {
 Path string
 Line int
 Err  error
}

func (e *ParseError) Error() string {
 return fmt.Sprintf("parse error in %s at line %d: %v", e.Path, e.Line, e.Err)
}

func (e *ParseError) Unwrap() error {
 return e.Err
}
```

### Utilities

```go
// WrapPath wraps error with file path context
func WrapPath(err error, path string) error {
 if err == nil {
  return nil
 }
 return fmt.Errorf("%s: %w", path, err)
}

// IsParseError checks if error is a ParseError
func IsParseError(err error) bool {
 var parseErr *ParseError
 return errors.As(err, &parseErr)
}
```

### Testing

```go
func TestMultiError(t *testing.T) {
 err := &MultiError{
  Errors: []error{
   errors.New("error 1"),
   errors.New("error 2"),
  },
 }

 if err.Error() == "" {
  t.Error("MultiError.Error() should not be empty")
 }

 unwrapped := err.Unwrap()
 if len(unwrapped) != 2 {
  t.Errorf("expected 2 errors, got %d", len(unwrapped))
 }
}
```

---

## Internal/Files Package

### Purpose

File system operations and directory walking.

### Key Functions

```go
// Walk traverses directory tree and finds .tf and .hcl files
func Walk(root string, recursive bool) ([]string, error)

// ShouldProcess checks if file should be processed
func ShouldProcess(path string) bool

// IsExcluded checks if path should be excluded
func IsExcluded(path string) bool

// GetFileExtension returns file extension
func GetFileExtension(path string) string
```

### Filtering Rules

**Included files**:

- `.tf` files (Terraform)
- `.hcl` files (Terragrunt)

**Excluded directories**:

- `.terraform/` (Terraform cache)
- `.terragrunt-cache/` (Terragrunt cache)
- `.git/` (Git repository)
- Hidden directories (starting with `.`)

### Implementation

```go
// Walk finds all Terraform/Terragrunt files in directory tree
func Walk(root string, recursive bool) ([]string, error) {
 var files []string

 walkFn := func(path string, info os.FileInfo, err error) error {
  if err != nil {
   return err
  }

  // Skip excluded directories
  if info.IsDir() && IsExcluded(path) {
   return filepath.SkipDir
  }

  // Only process files
  if info.IsDir() {
   // If not recursive, skip subdirectories
   if !recursive && path != root {
    return filepath.SkipDir
   }
   return nil
  }

  // Check if file should be processed
  if ShouldProcess(path) {
   files = append(files, path)
  }

  return nil
 }

 err := filepath.Walk(root, walkFn)
 if err != nil {
  return nil, fmt.Errorf("failed to walk directory: %w", err)
 }

 return files, nil
}

func ShouldProcess(path string) bool {
 ext := filepath.Ext(path)
 return ext == ".tf" || ext == ".hcl"
}

func IsExcluded(path string) bool {
 base := filepath.Base(path)

 // Exclude hidden directories
 if strings.HasPrefix(base, ".") {
  return true
 }

 // Exclude specific directories
 excluded := []string{
  ".terraform",
  ".terragrunt-cache",
  "node_modules",
 }

 for _, dir := range excluded {
  if base == dir {
   return true
  }
 }

 return false
}
```

### Testing

```go
func TestWalk(t *testing.T) {
 // Create temporary directory structure
 tmpDir := t.TempDir()

 // Create test files
 createFile(t, tmpDir, "main.tf")
 createFile(t, tmpDir, "variables.tf")
 createFile(t, tmpDir, "README.md")  // Should be excluded

 // Create excluded directory
 terraformDir := filepath.Join(tmpDir, ".terraform")
 os.Mkdir(terraformDir, 0755)
 createFile(t, terraformDir, "cache.tf")  // Should be excluded

 // Walk directory
 files, err := Walk(tmpDir, true)
 if err != nil {
  t.Fatal(err)
 }

 // Verify results
 if len(files) != 2 {
  t.Errorf("expected 2 files, got %d", len(files))
 }

 // Verify no .terraform files included
 for _, file := range files {
  if strings.Contains(file, ".terraform") {
   t.Errorf("should not include .terraform files: %s", file)
  }
 }
}

func TestShouldProcess(t *testing.T) {
 tests := []struct {
  path string
  want bool
 }{
  {"main.tf", true},
  {"terragrunt.hcl", true},
  {"README.md", false},
  {"script.sh", false},
  {"data.json", false},
 }

 for _, tt := range tests {
  t.Run(tt.path, func(t *testing.T) {
   got := ShouldProcess(tt.path)
   if got != tt.want {
    t.Errorf("ShouldProcess(%q) = %v, want %v", tt.path, got, tt.want)
   }
  })
 }
}
```

---

## Best Practices for Internal Packages

### 1. Keep Focused

Each internal package should have a single, clear purpose:

**✅ Good**:

- `internal/errors/`: Error utilities only
- `internal/files/`: File operations only

**❌ Bad**:

- `internal/utils/`: Dumping ground for everything

### 2. High Test Coverage

Internal packages are critical infrastructure:

- Aim for **100% coverage**
- Test all edge cases
- Test error paths

### 3. Document Like Public API

Even though internal, document well:

```go
// Walk traverses the directory tree rooted at root and returns a list of
// Terraform (.tf) and Terragrunt (.hcl) file paths.
//
// If recursive is false, only files in the root directory are returned.
// If recursive is true, all subdirectories are traversed.
//
// Excluded directories:
//   - .terraform/ (Terraform cache)
//   - .terragrunt-cache/ (Terragrunt cache)
//   - .git/ (Git repository)
//   - Hidden directories (starting with .)
//
// Returns an error if the root directory does not exist or cannot be accessed.
func Walk(root string, recursive bool) ([]string, error)
```

### 4. No External Dependencies

Internal packages should depend on:

- Standard library only
- Other internal packages (sparingly)

**Avoid**:

- External dependencies
- Importing from `api` or `cli` packages (creates circular dependency)

---

## Adding New Internal Packages

### When to Add

Add internal package when:

- Utility is used by 2+ public packages
- Code doesn't belong in any public package
- Want to hide implementation details

### Steps

1. Create directory: `internal/newpackage/`
2. Create source file: `internal/newpackage/newpackage.go`
3. Add package doc comment
4. Create test file: `internal/newpackage/newpackage_test.go`
5. Document in this CLAUDE.md file
6. Import where needed

### Example

```bash
mkdir -p internal/cache
touch internal/cache/cache.go
touch internal/cache/cache_test.go
```

```go
// Package cache provides caching utilities for sortTF.
//
// This package is internal and not part of the public API.
package cache

// Cache represents an in-memory cache
type Cache struct {
 // ...
}
```

---

## Dependencies

**External**:

- None (standard library only)

**Internal**:

- `internal/errors` may be imported by `internal/files`
- Avoid circular dependencies

---

## Acceptance Checklist (Internal Packages)

Before considering internal package changes complete:

- [ ] Package has single, clear purpose
- [ ] Package doc comment explains purpose
- [ ] All exported functions/types have Godoc
- [ ] Tests cover all functionality
- [ ] Test coverage 100% (aim for this)
- [ ] No external dependencies (standard library only)
- [ ] No imports from public packages (api, cli)
- [ ] Error handling follows repository standards
- [ ] Examples provided for complex functions
- [ ] Integration with public packages tested
