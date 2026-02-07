# Test Fixtures

This directory contains test fixtures organized by category to ensure sortTF handles a wide variety of Terraform configurations correctly.

## Directory Structure

```text
testdata/fixtures/
├── control/         # Control flow constructs (for_each, count, conditionals)
├── edge_cases/      # Edge cases and unusual patterns
├── large/           # Large files for performance testing
├── real_world/      # Real-world complex scenarios
├── realistic/       # Realistic infrastructure examples
├── structure/       # Block structure and nesting
├── syntax/          # Syntax edge cases (comments, heredocs, whitespace)
└── types/           # Terraform type system edge cases
```

## Categories

### `control/` - Control Flow

Tests for Terraform control flow features:

- **conditionals.tf**: Conditional expressions and ternary operators
- **count.tf**: Count meta-argument usage
- **for_each.tf**: For_each loops and iteration
- **functions.tf**: Built-in function usage
- **interpolation.tf**: String interpolation and templates

### `edge_cases/` - Edge Cases

Unusual patterns and boundary conditions:

- **deep_nesting.tf**: Deeply nested blocks (10+ levels) to test recursion limits
- **lifecycle_and_meta.tf**: All lifecycle options, meta-arguments (moved, import, check)
- **long_values.tf**: Very long strings, complex expressions, large JSON blobs
- **many_similar_names.tf**: Many resources with similar names (tests alphabetical sorting)
- **special_characters.tf**: Special characters, Unicode, escaped strings, regex patterns

### `large/` - Large Files

Performance testing with large configurations:

- **aws_multi_region.tf**: 500+ line AWS multi-region setup
- **kubernetes_deployment.tf**: 400+ line Kubernetes deployment
- **performance_test.tf**: 2000+ line file with 200+ resources (stress test)

### `real_world/` - Real World Scenarios

Complex real-world patterns:

- **module_with_complex_variables.tf**: Complex variable types (objects, lists, maps, validation)
- **multiple_providers.tf**: Multiple cloud providers with aliases
- **remote_state_configuration.tf**: Remote state backends and data sources

### `realistic/` - Realistic Examples

Smaller realistic infrastructure:

- **aws_infrastructure.tf**: Basic AWS infrastructure (VPC, EC2, S3)
- **terragrunt.hcl**: Terragrunt configuration file

### `structure/` - Block Structure

Tests for block organization:

- **all_block_types.tf**: All Terraform block types in one file
- **dynamic_blocks.tf**: Dynamic block generation
- **nested_blocks.tf**: Nested block structures
- **repeated_blocks.tf**: Multiple blocks of same type

### `syntax/` - Syntax Edge Cases

HCL syntax variations:

- **comments_*.tf**: Various comment styles (hash, double-slash, multiline, mixed)
- **empty.tf**: Empty file
- **heredoc_*.tf**: Heredoc strings (standard, indented)
- **whitespace.tf**: Whitespace-only file

### `types/` - Type System

Terraform type system edge cases:

- **booleans.tf**: Boolean values and expressions
- **empty_collections.tf**: Empty lists, maps, sets
- **nested_collections.tf**: Complex nested data structures
- **nulls.tf**: Null values and nullable types
- **numbers.tf**: Number types and arithmetic

## File Sizes

| Category | Files | Largest File | Total Lines |
|----------|-------|--------------|-------------|
| control | 5 | functions.tf (16) | ~70 |
| edge_cases | 5 | long_values.tf (115) | ~400 |
| large | 3 | performance_test.tf (2093) | ~3000 |
| real_world | 3 | multiple_providers.tf (370) | ~850 |
| realistic | 2 | aws_infrastructure.tf (111) | ~135 |
| structure | 4 | all_block_types.tf (31) | ~73 |
| syntax | 8 | comments_multiline.tf (13) | ~50 |
| types | 5 | nested_collections.tf (16) | ~67 |
| **Total** | **35** | **performance_test.tf** | **~4645** |

## Usage in Tests

These fixtures are used by:

1. **Unit Tests** (`hcl/` package) - Test sorting logic on specific patterns
2. **Integration Tests** (`integration/` package) - End-to-end CLI testing
3. **Performance Tests** - Benchmark sorting on large files
4. **CI/CD** - Automated testing across all fixtures

## Adding New Fixtures

When adding new fixtures:

1. **Choose the right category** based on what you're testing
2. **Add documentation** in this README
3. **Keep files focused** - test one concept per file when possible
4. **Name files descriptively** - use snake_case, describe what's tested
5. **Add comments** explaining edge cases being tested
6. **Update the table** above with new file info

## Testing Best Practices

- **Edge cases first**: Start with edge cases, then realistic examples
- **Performance**: Test with large files to catch performance regressions
- **Real world**: Include realistic configurations users actually write
- **Syntax coverage**: Cover all HCL syntax variations
- **Type coverage**: Test all Terraform types and combinations

## Notes

- Files are tested both sorted and unsorted
- sortTF should handle all these files without errors
- Large files (1000+ lines) are primarily for performance testing
- Comments in test files are for documentation; **sortTF removes comments during processing**
