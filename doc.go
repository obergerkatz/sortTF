// Package sorttf is a command-line tool for sorting and formatting Terraform and Terragrunt files.
//
// sortTF sorts Terraform (.tf) and Terragrunt (.hcl) files in a consistent, deterministic way
// to improve readability and reduce diff noise in version control.
//
// # Features
//
//   - Sort blocks by type (terraform, provider, variable, locals, data, resource, module, output)
//   - Sort blocks of the same type alphabetically by label
//   - Sort attributes alphabetically within blocks (with for_each always first)
//   - Apply canonical HCL formatting (compatible with terraform fmt)
//   - Process individual files or entire directories (with optional recursion)
//   - Parallel processing for fast performance on large repositories
//   - Dry-run mode to preview changes
//   - Validate mode for CI/CD pipelines
//   - No external dependencies (terraform command not required)
//
// # Usage
//
// Basic usage:
//
//	sorttf [flags] [path]
//
// Sort files in the current directory:
//
//	sorttf
//
// Sort a specific file:
//
//	sorttf main.tf
//
// Sort all files recursively:
//
//	sorttf --recursive ./terraform
//
// Preview changes without modifying files:
//
//	sorttf --dry-run main.tf
//
// Validate files are sorted (useful in CI/CD):
//
//	sorttf --validate .
//
// # Flags
//
//	-dry-run      Show what would be changed without writing (shows a unified diff)
//	-recursive    Process directories recursively
//	-validate     Check if files are sorted without modifying (exits 1 if changes needed)
//	-verbose      Enable verbose output showing file processing details
//	-help         Display usage information
//
// # Package Organization
//
// The project is organized into the following packages:
//
//   - api: Public API for sorting files programmatically
//   - cli: Command-line interface and main execution logic
//   - config: Configuration parsing and flag handling
//   - hcl: HCL parsing, sorting, and formatting
//   - internal/errors: Unified error handling
//   - internal/files: File traversal and validation utilities
//
// # Exit Codes
//
//   - 0: Success (files processed or already sorted)
//   - 1: Error occurred during processing
//   - 2: Invalid command-line arguments
//
// # Example
//
// Sort a Terraform configuration:
//
//	$ sorttf main.tf
//	✅ Updated: main.tf
//	✅ Processed 1 files
//
// Validate files in CI/CD:
//
//	$ sorttf --validate .
//	⚠️  Needs update: main.tf
//	❌ Encountered 1 errors
//	$ echo $?
//	1
//
// # Library Usage
//
// Import the api package to use sortTF programmatically:
//
//	import "github.com/obergerkatz/sortTF/api"
//
//	err := api.SortFile("main.tf", api.Options{})
package sorttf
