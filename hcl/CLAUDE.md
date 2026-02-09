# HCL Package Instructions

**Scope**: Applies only to `hcl/` directory.

This file **extends** the repository-level `.claude/CLAUDE.md`. Read that first for global standards. This file contains only what's **specific** to the hcl package.

---

## Purpose

The `hcl` package is the **core sorting and formatting engine** for sortTF:

- Parse HCL files using HashiCorp's hclwrite library
- Sort Terraform blocks by type and name
- Sort attributes within blocks (alphabetically with special handling)
- Format output according to `terraform fmt` standards
- Detect changes between original and sorted content

**Philosophy**:

- Leverage official HashiCorp HCL library for correctness
- Implement deterministic sorting rules
- Preserve HCL semantics (only reorder, don't change meaning)
- Return errors, don't print (library code, not CLI)

---

## Package Structure

```text
hcl/
├── doc.go              # Package documentation
├── parser.go           # HCL parsing logic
├── sorter.go           # Sorting algorithms
├── formatter.go        # Output formatting
├── errors.go           # Custom error types
├── parser_test.go      # Parser tests
├── sorter_test.go      # Sorter tests
├── formatter_test.go   # Formatter tests
├── errors_test.go      # Error tests
└── CLAUDE.md           # This file
```

---

## Core Components

### 1. Parser (parser.go)

**Purpose**: Parse HCL content into an internal representation.

**Key Functions**:

```go
// Parse parses HCL content and returns a structured representation
func Parse(content []byte) (*File, error)

// File represents a parsed HCL file
type File struct {
 Body   *hclwrite.Body
 Blocks []*Block
}

// Block represents an HCL block (resource, provider, etc.)
type Block struct {
 Type       string   // "resource", "provider", etc.
 Labels     []string // Block labels
 HCLBlock   *hclwrite.Block
}
```

**Implementation Notes**:

- Uses `github.com/hashicorp/hcl/v2/hclwrite` for parsing
- Extracts top-level blocks only (nested blocks handled separately)
- Preserves block structure and attributes
- Returns parse errors with line numbers

### 2. Sorter (sorter.go)

**Purpose**: Sort blocks and attributes according to rules.

**Block Sorting Rules**:

```go
// Block order map (Terraform best practices)
var blockOrder = map[string]int{
 "terraform": 0,  // Terraform settings
 "provider":  1,  // Provider configurations
 "variable":  2,  // Input variables
 "locals":    3,  // Local values
 "data":      4,  // Data sources
 "resource":  5,  // Resources
 "module":    6,  // Module calls
 "output":    7,  // Output values
}

// Unknown block types default to 999 (sorted last)
```

**Key Functions**:

```go
// SortBlocks sorts blocks by type and labels
func SortBlocks(file *File)

// SortAttributes sorts attributes within a block
func SortAttributes(block *hclwrite.Block)
```

**Sorting Algorithm**:

1. **Primary sort**: By block type (terraform, provider, variable, etc.)
2. **Secondary sort**: Within same type, alphabetically by labels
3. **Tertiary sort**: Attributes within blocks (special attributes first, then alphabetical)

**Attribute Sorting**:

- Special attributes first: `for_each`, `count`, `depends_on`
- Regular attributes: Alphabetical order
- Nested blocks: Keep after attributes

### 3. Formatter (formatter.go)

**Purpose**: Generate formatted HCL output.

**Key Functions**:

```go
// Format formats a File into HCL bytes
func Format(file *File) ([]byte, error)

// CompareContent checks if two HCL contents are equivalent
func CompareContent(original, sorted []byte) (bool, error)
```

**Implementation Notes**:

- Uses `hclwrite.File.Bytes()` for output
- Applies `terraform fmt` style automatically
- **Important**: Comments are NOT preserved (HCL library limitation)
- Normalizes whitespace and formatting

### 4. Error Handling (errors.go)

**Custom Error Types**:

```go
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

// SortError represents sorting failures
type SortError struct {
 BlockType string
 Err       error
}
```

---

## Detailed Specifications

### Block Sorting

**Algorithm**:

```go
func SortBlocks(file *File) {
 sort.SliceStable(file.Blocks, func(i, j int) bool {
  // Compare by block type order
  orderI := getBlockOrder(file.Blocks[i].Type)
  orderJ := getBlockOrder(file.Blocks[j].Type)

  if orderI != orderJ {
   return orderI < orderJ
  }

  // Same type: compare by labels
  return compareLabels(file.Blocks[i].Labels, file.Blocks[j].Labels)
 })

 // Sort attributes within each block
 for _, block := range file.Blocks {
  SortAttributes(block.HCLBlock)
 }
}

func getBlockOrder(blockType string) int {
 if order, ok := blockOrder[blockType]; ok {
  return order
 }
 return 999  // Unknown types go last
}

func compareLabels(a, b []string) bool {
 // Lexicographic comparison of label slices
 for i := 0; i < len(a) && i < len(b); i++ {
  if a[i] != b[i] {
   return a[i] < b[i]
  }
 }
 return len(a) < len(b)
}
```

**Examples**:

```hcl
# Before sorting
resource "aws_instance" "web" { }
provider "aws" { }
variable "region" { }

# After sorting (provider -> variable -> resource)
provider "aws" { }
variable "region" { }
resource "aws_instance" "web" { }
```

**Multiple resources**:

```hcl
# Before
resource "aws_s3_bucket" "logs" { }
resource "aws_instance" "web" { }

# After (alphabetical by label)
resource "aws_instance" "web" { }
resource "aws_s3_bucket" "logs" { }
```

### Attribute Sorting

**Algorithm**:

```go
func SortAttributes(block *hclwrite.Block) {
 attrs := block.Body().Attributes()

 // Separate special and regular attributes
 var special []*hclwrite.Attribute
 var regular []*hclwrite.Attribute

 for name, attr := range attrs {
  if isSpecialAttribute(name) {
   special = append(special, attr)
  } else {
   regular = append(regular, attr)
  }
 }

 // Sort regular attributes alphabetically
 sort.Slice(regular, func(i, j int) bool {
  return getName(regular[i]) < getName(regular[j])
 })

 // Rebuild body with special attrs first, then regular
 newBody := hclwrite.NewBody()
 for _, attr := range special {
  newBody.SetAttributeRaw(getName(attr), attr.Expr().BuildTokens(nil))
 }
 for _, attr := range regular {
  newBody.SetAttributeRaw(getName(attr), attr.Expr().BuildTokens(nil))
 }

 // Replace block body
 block.SetBody(newBody)
}

func isSpecialAttribute(name string) bool {
 return name == "for_each" || name == "count" || name == "depends_on"
}
```

**Examples**:

```hcl
# Before
resource "aws_instance" "web" {
  instance_type = "t3.micro"
  for_each      = var.instances
  ami           = "ami-123"
  tags = {
    Name = "web"
  }
}

# After (for_each first, then alphabetical)
resource "aws_instance" "web" {
  for_each      = var.instances
  ami           = "ami-123"
  instance_type = "t3.micro"
  tags = {
    Name = "web"
  }
}
```

---

## Testing HCL Package

### Test Strategy

- **Unit tests**: Test individual functions (Parse, SortBlocks, SortAttributes)
- **Integration tests**: Test full parse -> sort -> format flow
- **Edge cases**: Empty files, single block, deeply nested blocks
- **Error cases**: Invalid HCL, parse errors

### Test Fixtures

Use `testdata/` for test files:

```text
hcl/
└── testdata/
    ├── valid_unsorted.tf
    ├── valid_sorted.tf
    ├── invalid_syntax.tf
    ├── empty.tf
    └── complex_nested.tf
```

### Table-Driven Tests

```go
func TestSortBlocks(t *testing.T) {
 tests := []struct {
  name  string
  input string
  want  []string // Expected block types in order
 }{
  {
   name:  "already sorted",
   input: `provider "aws" {} resource "aws_instance" "web" {}`,
   want:  []string{"provider", "resource"},
  },
  {
   name:  "needs sorting",
   input: `resource "aws_instance" "web" {} provider "aws" {}`,
   want:  []string{"provider", "resource"},
  },
  {
   name:  "all block types",
   input: `output "x" {} resource "a" "b" {} terraform {} variable "y" {}`,
   want:  []string{"terraform", "variable", "resource", "output"},
  },
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   file, err := Parse([]byte(tt.input))
   if err != nil {
    t.Fatal(err)
   }

   SortBlocks(file)

   got := make([]string, len(file.Blocks))
   for i, block := range file.Blocks {
    got[i] = block.Type
   }

   if !reflect.DeepEqual(got, tt.want) {
    t.Errorf("block order = %v, want %v", got, tt.want)
   }
  })
 }
}
```

### Error Testing

```go
func TestParse_InvalidHCL(t *testing.T) {
 invalidHCL := `resource "aws_instance" "web" {
  ami = "unclosed-string
 }`

 _, err := Parse([]byte(invalidHCL))
 if err == nil {
  t.Fatal("expected parse error, got nil")
 }

 var parseErr *ParseError
 if !errors.As(err, &parseErr) {
  t.Errorf("expected ParseError, got %T", err)
 }
}
```

---

## HCL Library Integration

### Using hashicorp/hcl

**Key imports**:

```go
import (
 "github.com/hashicorp/hcl/v2"
 "github.com/hashicorp/hcl/v2/hclwrite"
)
```

**Parsing**:

```go
file, diags := hclwrite.ParseConfig(content, filename, hcl.InitialPos)
if diags.HasErrors() {
 return nil, &ParseError{
  Path: filename,
  Line: getFirstDiagLine(diags),
  Err:  diags,
 }
}
```

**Accessing blocks**:

```go
for _, block := range file.Body().Blocks() {
 blockType := block.Type()
 labels := block.Labels()
 // Process block
}
```

**Accessing attributes**:

```go
attrs := block.Body().Attributes()
for name, attr := range attrs {
 // name: attribute name
 // attr: *hclwrite.Attribute
}
```

**Writing HCL**:

```go
output := file.Bytes()  // []byte
```

---

## Performance Considerations

### Parsing Performance

- HCL parsing is the main bottleneck
- Use HashiCorp's optimized parser (don't reimplement)
- Cache parsed results if processing same file multiple times

### Sorting Performance

- Sorting is O(n log n) for blocks and attributes
- Negligible compared to I/O and parsing time
- Stable sort preserves relative order of equal elements

### Memory Usage

- Entire file loaded into memory
- AST representation has overhead
- Suitable for typical Terraform files (<1MB)

### Benchmarks

```go
func BenchmarkParse(b *testing.B) {
 content := generateTestContent()
 b.ResetTimer()
 for i := 0; i < b.N; i++ {
  Parse(content)
 }
}

func BenchmarkSortBlocks(b *testing.B) {
 file := createTestFile()
 b.ResetTimer()
 for i := 0; i < b.N; i++ {
  SortBlocks(file)
 }
}
```

---

## Error Handling

### Parse Errors

```go
// Wrap HCL diagnostics in custom error type
if diags.HasErrors() {
 return &ParseError{
  Path: filename,
  Line: extractLine(diags),
  Err:  diags,
 }
}
```

### Sort Errors

```go
// Rare, but handle block processing errors
if err := processBlock(block); err != nil {
 return &SortError{
  BlockType: block.Type,
  Err:       err,
 }
}
```

### Format Errors

```go
// Format errors are rare (hclwrite is robust)
// But check for edge cases
if len(output) == 0 {
 return nil, fmt.Errorf("format produced empty output")
}
```

---

## Common Pitfalls

### Comment Loss

**Important**: HCL library does NOT preserve comments.

**Why**: `hclwrite` reconstructs files from AST, which doesn't include comments.

**Documented**: This is clearly documented in README and user docs.

**Not a bug**: This is a limitation of the underlying library.

### Block Label Handling

Handle blocks with multiple labels correctly:

```go
resource "aws_instance" "web" {}  // Labels: ["aws_instance", "web"]
provider "aws" {}                  // Labels: ["aws"] (single label)
terraform {}                       // Labels: [] (no labels)
```

### Nested Blocks

Nested blocks within blocks are preserved but not sorted:

```hcl
resource "aws_instance" "web" {
  ami = "ami-123"

  lifecycle {  # Nested block (not top-level)
    create_before_destroy = true
  }
}
```

**Design decision**: Only sort top-level blocks, preserve nested structure.

---

## Dependencies

**External**:

- `github.com/hashicorp/hcl/v2` - HCL parsing
- `github.com/hashicorp/hcl/v2/hclwrite` - HCL writing

**Internal**:

- None (hcl is a low-level package, depends only on external libs)

---

## Extending the HCL Package

### Adding New Block Types

To support new Terraform block types:

1. Add to `blockOrder` map in `sorter.go`
2. Add tests for new block type
3. Update documentation

**Example**:

```go
// Add "moved" block support (Terraform 1.5+)
var blockOrder = map[string]int{
 "terraform": 0,
 "moved":     1,  // Add here
 "provider":  2,  // Adjust numbers
 // ...
}
```

### Custom Sorting Rules

To implement custom sorting (e.g., per-project rules):

1. Accept sorting config as parameter
2. Use config in `SortBlocks()` and `SortAttributes()`
3. Maintain backwards compatibility with default rules

---

## Acceptance Checklist (HCL Package)

Before considering HCL changes complete:

- [ ] All exported functions have Godoc
- [ ] Parsing handles all valid HCL syntax
- [ ] Sorting follows documented rules
- [ ] Special attributes handled correctly (`for_each`, `count`)
- [ ] Unknown block types handled gracefully (sorted last)
- [ ] Errors include context (file path, line number)
- [ ] No direct output (return errors, don't print)
- [ ] Tests cover all block types in `blockOrder`
- [ ] Tests cover attribute sorting edge cases
- [ ] Tests verify error types with `errors.As()`
- [ ] Test coverage >90% (`go test -cover ./hcl`)
- [ ] Benchmarks updated if algorithm changed
- [ ] Comment loss documented (if not already)
- [ ] Integration tests pass with new changes
