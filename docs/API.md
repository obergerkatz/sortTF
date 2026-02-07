# Library API Documentation

sortTF can be imported and used as a library in your own Go programs. This allows you to integrate Terraform file sorting into tools like pre-commit hooks, CI/CD pipelines, or custom build systems.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Use Cases](#use-cases)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Installation

```bash
go get github.com/obergerkatz/sortTF/api
```

## Quick Start

```go
package main

import (
    "errors"
    "log"

    "github.com/obergerkatz/sortTF/api"
)

func main() {
    // Sort a single file
    err := api.SortFile("main.tf", api.Options{})
    if err != nil && !errors.Is(err, api.ErrNoChanges) {
        log.Fatal(err)
    }
}
```

## API Reference

### Functions

#### SortFile

```go
func SortFile(path string, opts Options) error
```

Sorts and formats a single Terraform or Terragrunt file.

**Parameters:**

- `path`: Path to the `.tf` or `.hcl` file
- `opts`: Sorting options (see [Options](#options))

**Returns:**

- `nil`: File was successfully sorted and written
- `ErrNoChanges`: File is already sorted (not considered an error)
- `ErrNeedsSorting`: File needs sorting (only in Validate mode)
- `error`: Parsing, validation, or I/O error

**Example:**

```go
err := api.SortFile("main.tf", api.Options{})
if err != nil {
    if errors.Is(err, api.ErrNoChanges) {
        fmt.Println("File is already sorted")
    } else {
        log.Fatal(err)
    }
}
```

#### GetSortedContent

```go
func GetSortedContent(path string) (content string, changed bool, err error)
```

Reads a file and returns its sorted content without modifying the file. Useful for previewing changes or computing diffs.

**Parameters:**

- `path`: Path to the file

**Returns:**

- `content`: The sorted file content
- `changed`: Whether the content would be different from the original
- `err`: Any error that occurred

**Example:**

```go
content, changed, err := api.GetSortedContent("main.tf")
if err != nil {
    log.Fatal(err)
}
if changed {
    fmt.Println("File would be changed:")
    fmt.Println(content)
}
```

#### SortFiles

```go
func SortFiles(paths []string, opts Options) map[string]error
```

Sorts multiple files and returns results for each. Continues processing even if individual files fail.

**Parameters:**

- `paths`: Slice of file paths to sort
- `opts`: Sorting options

**Returns:**

- Map of file path to error (or nil if successful)

**Example:**

```go
files := []string{"main.tf", "variables.tf", "outputs.tf"}
results := api.SortFiles(files, api.Options{})

for path, err := range results {
    if err != nil && !errors.Is(err, api.ErrNoChanges) {
        fmt.Printf("❌ %s: %v\n", path, err)
    } else {
        fmt.Printf("✓ %s\n", path)
    }
}
```

#### SortDirectory

```go
func SortDirectory(dir string, recursive bool, opts Options) (map[string]error, error)
```

Sorts all Terraform/Terragrunt files in a directory.

**Parameters:**

- `dir`: Directory path
- `recursive`: Whether to process subdirectories
- `opts`: Sorting options

**Returns:**

- Map of file path to error for each file processed
- Error if directory traversal fails

**Example:**

```go
results, err := api.SortDirectory("./terraform", true, api.Options{})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Processed %d files\n", len(results))
for path, err := range results {
    if err != nil && !errors.Is(err, api.ErrNoChanges) {
        fmt.Printf("Error in %s: %v\n", path, err)
    }
}
```

### Types

#### Options

```go
type Options struct {
    DryRun   bool  // Don't modify files, just check what would change
    Validate bool  // Return ErrNeedsSorting if changes are needed
}
```

Configuration options for sorting operations.

**Fields:**

- `DryRun`: If true, files are not modified. Useful for previewing changes.
- `Validate`: If true, returns `ErrNeedsSorting` if file needs sorting instead of modifying it. Useful for CI/CD validation.

**Examples:**

```go
// Normal sorting
api.SortFile("main.tf", api.Options{})

// Dry run (preview only)
api.SortFile("main.tf", api.Options{DryRun: true})

// Validate mode (CI/CD)
err := api.SortFile("main.tf", api.Options{Validate: true})
if errors.Is(err, api.ErrNeedsSorting) {
    fmt.Println("File needs sorting")
}
```

### Sentinel Errors

#### ErrNoChanges

```go
var ErrNoChanges = errors.New("file is already sorted")
```

Returned when a file is already properly sorted. This is not considered an error condition - you should check for it with `errors.Is()` and typically ignore it.

**Example:**

```go
err := api.SortFile("main.tf", api.Options{})
if errors.Is(err, api.ErrNoChanges) {
    // File was already sorted - this is fine
    return nil
}
```

#### ErrNeedsSorting

```go
var ErrNeedsSorting = errors.New("file needs sorting")
```

Returned in Validate mode when a file needs sorting. Used for CI/CD validation.

**Example:**

```go
err := api.SortFile("main.tf", api.Options{Validate: true})
if errors.Is(err, api.ErrNeedsSorting) {
    fmt.Println("File is not sorted - CI check failed")
    os.Exit(1)
}
```

## Examples

### Example 1: Sort a Single File

```go
package main

import (
    "errors"
    "fmt"
    "log"

    "github.com/obergerkatz/sortTF/api"
)

func main() {
    err := api.SortFile("main.tf", api.Options{})
    if err != nil {
        if errors.Is(err, api.ErrNoChanges) {
            fmt.Println("✓ File is already sorted")
        } else {
            log.Fatalf("Error: %v", err)
        }
    } else {
        fmt.Println("✓ File sorted successfully")
    }
}
```

### Example 2: Validate Files (CI/CD)

```go
package main

import (
    "errors"
    "fmt"
    "os"

    "github.com/obergerkatz/sortTF/api"
)

func main() {
    results, err := api.SortDirectory(".", true, api.Options{Validate: true})
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    needsSorting := false
    for path, err := range results {
        if errors.Is(err, api.ErrNeedsSorting) {
            fmt.Printf("❌ %s needs sorting\n", path)
            needsSorting = true
        } else if err != nil {
            fmt.Printf("❌ %s: %v\n", path, err)
            needsSorting = true
        }
    }

    if needsSorting {
        fmt.Println("\nRun 'sorttf --recursive .' to fix")
        os.Exit(1)
    }

    fmt.Println("✓ All files are properly sorted")
}
```

### Example 3: Preview Changes

```go
package main

import (
    "fmt"
    "log"

    "github.com/obergerkatz/sortTF/api"
    "github.com/pmezard/go-difflib/difflib"
)

func main() {
    path := "main.tf"

    // Read original content
    original, _ := os.ReadFile(path)

    // Get sorted content
    sorted, changed, err := api.GetSortedContent(path)
    if err != nil {
        log.Fatal(err)
    }

    if !changed {
        fmt.Println("File is already sorted")
        return
    }

    // Generate diff
    diff := difflib.UnifiedDiff{
        A:        difflib.SplitLines(string(original)),
        B:        difflib.SplitLines(sorted),
        FromFile: path + " (original)",
        ToFile:   path + " (sorted)",
        Context:  3,
    }

    text, _ := difflib.GetUnifiedDiffString(diff)
    fmt.Println(text)
}
```

### Example 4: Batch Processing

```go
package main

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "github.com/obergerkatz/sortTF/api"
)

func main() {
    // Find all Terraform files
    var files []string
    filepath.Walk("./terraform", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && (filepath.Ext(path) == ".tf" || filepath.Ext(path) == ".hcl") {
            files = append(files, path)
        }
        return nil
    })

    // Sort all files
    results := api.SortFiles(files, api.Options{})

    // Report results
    sorted := 0
    skipped := 0
    failed := 0

    for path, err := range results {
        if err != nil {
            if errors.Is(err, api.ErrNoChanges) {
                skipped++
            } else {
                fmt.Printf("❌ %s: %v\n", path, err)
                failed++
            }
        } else {
            sorted++
        }
    }

    fmt.Printf("\n✓ Sorted: %d, ⚠ Skipped: %d, ❌ Failed: %d\n", sorted, skipped, failed)
}
```

## Use Cases

### Pre-commit Hook

```go
package main

import (
    "errors"
    "fmt"
    "os"
    "os/exec"
    "strings"

    "github.com/obergerkatz/sortTF/api"
)

func main() {
    // Get staged files
    cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM")
    output, err := cmd.Output()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error getting staged files: %v\n", err)
        os.Exit(1)
    }

    files := strings.Split(strings.TrimSpace(string(output)), "\n")

    // Filter Terraform files
    var tfFiles []string
    for _, file := range files {
        if strings.HasSuffix(file, ".tf") || strings.HasSuffix(file, ".hcl") {
            tfFiles = append(tfFiles, file)
        }
    }

    if len(tfFiles) == 0 {
        os.Exit(0)
    }

    // Sort files
    results := api.SortFiles(tfFiles, api.Options{})

    needsRestage := false
    for path, err := range results {
        if err != nil && !errors.Is(err, api.ErrNoChanges) {
            fmt.Fprintf(os.Stderr, "Error sorting %s: %v\n", path, err)
            os.Exit(1)
        }
        if err == nil {
            // File was modified, need to re-stage
            needsRestage = true
        }
    }

    if needsRestage {
        fmt.Println("Terraform files were sorted. Please re-stage and commit again.")
        // Re-stage files
        for _, file := range tfFiles {
            exec.Command("git", "add", file).Run()
        }
    }
}
```

### LSP/Editor Integration

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/obergerkatz/sortTF/api"
)

// LSP text document formatting
type FormattingRequest struct {
    TextDocument struct {
        URI string `json:"uri"`
    } `json:"textDocument"`
}

type FormattingResponse struct {
    NewText string `json:"newText"`
}

func handleFormatRequest(req FormattingRequest) (FormattingResponse, error) {
    // Convert URI to file path (remove file:// prefix)
    path := req.TextDocument.URI[7:]

    content, changed, err := api.GetSortedContent(path)
    if err != nil {
        return FormattingResponse{}, err
    }

    if !changed {
        // Return empty response if no changes needed
        return FormattingResponse{}, nil
    }

    return FormattingResponse{NewText: content}, nil
}
```

### Terraform Module Validator

```go
package main

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "github.com/obergerkatz/sortTF/api"
)

func validateModule(modulePath string) error {
    // Check if module files are properly sorted
    results, err := api.SortDirectory(modulePath, false, api.Options{Validate: true})
    if err != nil {
        return fmt.Errorf("failed to validate module: %w", err)
    }

    var validationErrors []string
    for path, err := range results {
        if errors.Is(err, api.ErrNeedsSorting) {
            rel, _ := filepath.Rel(modulePath, path)
            validationErrors = append(validationErrors, rel)
        } else if err != nil {
            return fmt.Errorf("error in %s: %w", path, err)
        }
    }

    if len(validationErrors) > 0 {
        return fmt.Errorf("files need sorting: %v", validationErrors)
    }

    return nil
}

func main() {
    if err := validateModule("./modules/vpc"); err != nil {
        fmt.Fprintf(os.Stderr, "Module validation failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("✓ Module is properly formatted")
}
```

## Error Handling

### Checking for Specific Errors

Always use `errors.Is()` to check for sentinel errors:

```go
err := api.SortFile("main.tf", api.Options{})
if err != nil {
    if errors.Is(err, api.ErrNoChanges) {
        // File already sorted - not an error
        return nil
    }
    if errors.Is(err, api.ErrNeedsSorting) {
        // File needs sorting (validate mode)
        return fmt.Errorf("file not sorted: %w", err)
    }
    // Some other error (parse error, I/O error, etc.)
    return fmt.Errorf("failed to sort: %w", err)
}
```

### Wrapping Errors

Wrap errors for better context:

```go
err := api.SortFile(path, api.Options{})
if err != nil && !errors.Is(err, api.ErrNoChanges) {
    return fmt.Errorf("failed to sort %s: %w", path, err)
}
```

### Handling Multiple Files

When processing multiple files, collect all errors:

```go
results := api.SortFiles(files, api.Options{})

var errs []error
for path, err := range results {
    if err != nil && !errors.Is(err, api.ErrNoChanges) {
        errs = append(errs, fmt.Errorf("%s: %w", path, err))
    }
}

if len(errs) > 0 {
    return fmt.Errorf("failed to sort %d files: %v", len(errs), errs)
}
```

## Best Practices

### 1. Always Check ErrNoChanges

Don't treat "already sorted" as an error:

```go
err := api.SortFile(path, opts)
if err != nil && !errors.Is(err, api.ErrNoChanges) {
    return err
}
```

### 2. Use Validate Mode in CI

Don't modify files in CI, only validate:

```go
results, _ := api.SortDirectory(".", true, api.Options{Validate: true})
for _, err := range results {
    if errors.Is(err, api.ErrNeedsSorting) {
        os.Exit(1)
    }
}
```

### 3. Preview Before Modifying

Use `GetSortedContent()` to preview changes:

```go
content, changed, _ := api.GetSortedContent(path)
if changed {
    // Show diff to user
    // Ask for confirmation
    // Then sort with SortFile()
}
```

### 4. Handle Errors Appropriately

Different contexts need different error handling:

```go
// In pre-commit hooks: fail fast
if err := api.SortFile(path, opts); err != nil {
    log.Fatal(err)
}

// In CI: collect all errors
results := api.SortFiles(files, opts)
for path, err := range results {
    if err != nil {
        fmt.Printf("Error: %s\n", path)
    }
}

// In servers: return errors to caller
if err := api.SortFile(path, opts); err != nil {
    return fmt.Errorf("sort failed: %w", err)
}
```

### 5. Use Concurrent Processing for Multiple Files

`SortFiles()` processes files concurrently for performance:

```go
// Efficient - concurrent processing
results := api.SortFiles(manyFiles, opts)

// Inefficient - sequential
for _, file := range manyFiles {
    api.SortFile(file, opts)
}
```

### 6. Don't Sort Generated Files

Check if file is generated before sorting:

```go
func shouldSort(path string) bool {
    content, _ := os.ReadFile(path)
    header := string(content[:min(200, len(content))])

    // Skip generated files
    if strings.Contains(header, "Code generated by") {
        return false
    }

    return true
}
```

## Next Steps

- Read the [Usage Guide](USAGE.md) for CLI usage patterns
- Check [Contributing Guide](CONTRIBUTING.md) to contribute improvements
