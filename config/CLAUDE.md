# Config Package Instructions

**Scope**: Applies only to `config/` directory.

This file **extends** the repository-level `.claude/CLAUDE.md`. Read that first for global standards. This file contains only what's **specific** to the config package.

---

## Purpose

The `config` package defines **configuration structures and validation** for sortTF:

- Options struct for sorting operations
- Configuration validation
- Default values
- Type definitions for public API

**Philosophy**:

- Simple, flat configuration
- Explicit over implicit
- Validate early, fail fast
- Backwards-compatible extensions

---

## Package Structure

```text
config/
├── config.go          # Configuration types and validation
├── config_test.go     # Configuration tests
└── CLAUDE.md          # This file
```

---

## Configuration Types

### Options Struct

```go
// Options configures sorting behavior.
type Options struct {
 // DryRun previews changes without modifying files.
 DryRun bool

 // Validate checks if files need sorting without modifying them.
 // Returns ErrNeedsSorting if file needs sorting.
 // Useful for CI/CD validation.
 Validate bool
}
```

**Design Principles**:

- **Flat structure**: No nested config objects
- **Boolean flags**: Simple enable/disable options
- **Zero value is safe**: Default behavior is normal sorting
- **Documented fields**: Each field has Godoc comment

---

## Configuration Validation

### Validate Function

```go
// Validate checks if Options configuration is valid.
//
// Returns error if configuration is invalid or conflicting.
func (o Options) Validate() error {
 // Check for conflicting options
 if o.DryRun && o.Validate {
  return fmt.Errorf("cannot use both dry-run and validate modes")
 }

 // Add more validation as needed

 return nil
}
```

**Validation Rules**:

- Conflicting flags (e.g., DryRun + Validate)
- Invalid combinations
- Future: validate config file settings

---

## Usage Patterns

### Default Configuration

```go
// Default options (zero value)
opts := config.Options{}
// DryRun: false, Validate: false

// Normal sorting mode
err := api.SortFile("main.tf", opts)
```

### Dry Run Mode

```go
opts := config.Options{
 DryRun: true,
}

// Preview changes without writing
content, changed, err := api.GetSortedContent(path)
```

### Validate Mode (CI/CD)

```go
opts := config.Options{
 Validate: true,
}

// Check if file needs sorting
err := api.SortFile("main.tf", opts)
if errors.Is(err, api.ErrNeedsSorting) {
 fmt.Println("File needs sorting")
 os.Exit(1)
}
```

---

## Testing Configuration

### Test Coverage

- Test default values
- Test validation logic
- Test conflicting options
- Test zero-value safety

### Example Tests

```go
func TestOptions_Validate(t *testing.T) {
 tests := []struct {
  name    string
  opts    Options
  wantErr bool
 }{
  {
   name:    "default options valid",
   opts:    Options{},
   wantErr: false,
  },
  {
   name:    "dry-run only valid",
   opts:    Options{DryRun: true},
   wantErr: false,
  },
  {
   name:    "validate only valid",
   opts:    Options{Validate: true},
   wantErr: false,
  },
  {
   name:    "both dry-run and validate invalid",
   opts:    Options{DryRun: true, Validate: true},
   wantErr: true,
  },
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   err := tt.opts.Validate()
   if (err != nil) != tt.wantErr {
    t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
   }
  })
 }
}

func TestOptions_ZeroValue(t *testing.T) {
 // Zero value should be safe to use
 var opts Options

 if opts.DryRun {
  t.Error("zero value should have DryRun = false")
 }
 if opts.Validate {
  t.Error("zero value should have Validate = false")
 }

 if err := opts.Validate(); err != nil {
  t.Errorf("zero value should validate: %v", err)
 }
}
```

---

## Future Extensions

### Configuration Files (Future)

If adding config file support:

```go
// Config represents a configuration file.
type Config struct {
 // Sorting options
 Sort SortOptions

 // Formatting options
 Format FormatOptions

 // File filtering
 Exclude []string
}

type SortOptions struct {
 // Custom block order
 BlockOrder map[string]int

 // Attribute sorting rules
 AttributeRules []AttributeRule
}

type FormatOptions struct {
 // Indentation (spaces)
 Indent int

 // Maximum line length
 MaxLineLength int
}
```

**Config file format** (YAML or TOML):

```yaml
# .sorttf.yml
sort:
  block_order:
    terraform: 0
    moved: 1
    provider: 2

format:
  indent: 2
  max_line_length: 120

exclude:
  - ".terraform/**"
  - "**/.terragrunt-cache/**"
```

### Loading Configuration

```go
// LoadConfig loads configuration from file.
func LoadConfig(path string) (*Config, error) {
 data, err := os.ReadFile(path)
 if err != nil {
  return nil, fmt.Errorf("failed to read config: %w", err)
 }

 var cfg Config
 if err := yaml.Unmarshal(data, &cfg); err != nil {
  return nil, fmt.Errorf("failed to parse config: %w", err)
 }

 if err := cfg.Validate(); err != nil {
  return nil, fmt.Errorf("invalid config: %w", err)
 }

 return &cfg, nil
}
```

**Not implemented yet** - this is for future expansion.

---

## API Stability

### Backwards Compatibility

When adding new options:

**✅ Safe additions**:

```go
// Adding new optional field with zero-value default
type Options struct {
 DryRun   bool
 Validate bool
 Verbose  bool  // NEW: zero value (false) is safe
}
```

**❌ Breaking changes**:

```go
// Renaming fields
type Options struct {
 DryRun   bool
 Check    bool  // BREAKING: renamed from Validate
}

// Changing field types
type Options struct {
 DryRun   string  // BREAKING: was bool
 Validate bool
}

// Removing fields
type Options struct {
 // DryRun removed  // BREAKING
 Validate bool
}
```

---

## Dependencies

**External**:

- None (only standard library)

**Internal**:

- None (config is a low-level package with no internal dependencies)

---

## Common Patterns

### Using Options in API

```go
// API function accepts Options
func SortFile(path string, opts Options) error {
 // Validate configuration
 if err := opts.Validate(); err != nil {
  return fmt.Errorf("invalid options: %w", err)
 }

 // Use configuration
 if opts.DryRun {
  // Preview mode
 } else if opts.Validate {
  // Validation mode
 } else {
  // Normal mode
 }
}
```

### Passing Options Through Layers

```go
// CLI builds Options from flags
opts := config.Options{
 DryRun:   flagDryRun,
 Validate: flagValidate,
}

// Pass to API
err := api.SortFile(path, opts)
```

---

## Documentation

### Godoc Requirements

All exported types and fields must have documentation:

```go
// Options configures sorting behavior.
//
// The zero value is safe to use and represents normal sorting mode
// (no dry-run, no validation).
type Options struct {
 // DryRun previews changes without modifying files.
 //
 // When true, files are parsed and sorted but not written.
 // Use GetSortedContent to retrieve the sorted content.
 DryRun bool

 // Validate checks if files need sorting without modifying them.
 //
 // When true, returns ErrNeedsSorting if file needs sorting.
 // Returns ErrNoChanges if file is already sorted.
 // Returns other errors for parsing or I/O failures.
 //
 // Useful for CI/CD pipelines to enforce sorted files.
 Validate bool
}
```

---

## Acceptance Checklist (Config Package)

Before considering config changes complete:

- [ ] All exported types have Godoc
- [ ] All exported fields have Godoc comments
- [ ] Zero value is safe and well-defined
- [ ] Validation function checks for conflicts
- [ ] No breaking changes to existing Options fields
- [ ] Tests cover all validation rules
- [ ] Tests verify zero-value behavior
- [ ] Tests check conflicting options
- [ ] Test coverage 100% (`go test -cover ./config`)
- [ ] No dependencies on other internal packages
- [ ] Documentation updated if API changed
