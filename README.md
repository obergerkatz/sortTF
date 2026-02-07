# sortTF

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/obergerkatz/sortTF)](https://goreportcard.com/report/github.com/obergerkatz/sortTF)
[![Coverage](https://img.shields.io/badge/coverage-95%25-brightgreen.svg)](https://github.com/obergerkatz/sortTF)

A command-line tool and Go library for sorting and formatting Terraform (.tf) and Terragrunt (.hcl) files to ensure consistency and readability across your infrastructure code.

## Features

- **Smart Block Sorting**: Orders Terraform blocks according to best practices
- **Attribute Sorting**: Alphabetizes attributes with special handling for `for_each`
- **Nested Block Support**: Handles deeply nested and complex HCL structures
- **Formatting**: Applies `terraform fmt` standards automatically
- **Multiple Modes**: Dry-run, validation, and recursive directory processing
- **Library API**: Use as a Go library in your own tools
- **Cross-Platform**: Works on macOS, Linux, and Windows
- **Well-Tested**: 95% test coverage with 155+ unit tests and 29 integration tests

## Quick Start

### Installation

```bash
# Using go install (recommended)
go install github.com/obergerkatz/sortTF/cmd/sorttf@latest

# Or download pre-built binaries from releases
# https://github.com/obergerkatz/sortTF/releases
```

See [Installation Guide](docs/INSTALLATION.md) for more options.

### Basic Usage

```bash
# Sort files in current directory
sorttf .

# Sort specific file
sorttf main.tf

# Recursively sort all files
sorttf --recursive .

# Preview changes without modifying files
sorttf --dry-run .

# Validate files (useful in CI/CD)
sorttf --validate .
```

See [Usage Guide](docs/USAGE.md) for detailed examples and options.

## How It Works

sortTF reorders Terraform blocks in a standardized sequence:

1. `terraform` → 2. `provider` → 3. `variable` → 4. `locals` → 5. `data` → 6. `resource` → 7. `module` → 8. `output`

Within blocks, attributes are sorted alphabetically with `for_each` always placed first.

### Example

**Before:**

```hcl
resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami           = "ami-123456"
}
provider "aws" { region = "us-west-2" }
variable "environment" { type = string }
```

**After:**

```hcl
provider "aws" { region = "us-west-2" }
variable "environment" { type = string }
resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
}
```

## Documentation

- **[Installation Guide](docs/INSTALLATION.md)** - Installation instructions for all platforms
- **[Usage Guide](docs/USAGE.md)** - Comprehensive CLI usage and examples
- **[Library API](docs/API.md)** - Using sortTF as a Go library
- **[Contributing](docs/CONTRIBUTING.md)** - Contribution guidelines and development setup
- **[Development](docs/DEVELOPMENT.md)** - Building, testing, and code structure
- **[Architecture](docs/ARCHITECTURE.md)** - Technical design and implementation details
- **[Releasing](docs/RELEASING.md)** - How to create and publish releases

## Using as a Library

Import sortTF in your Go programs:

```go
import "github.com/obergerkatz/sortTF/api"

// Sort a single file
err := api.SortFile("main.tf", api.Options{})

// Sort multiple files
results := api.SortFiles(paths, api.Options{DryRun: true})

// Sort entire directory
results, err := api.SortDirectory("./terraform", true, api.Options{})
```

See [Library API Documentation](docs/API.md) for more details.

## Contributing

Contributions are welcome! Please read the [Contributing Guide](docs/CONTRIBUTING.md) for details on:

- Setting up your development environment
- Running tests and linters
- Submitting pull requests
- Code style and conventions

## Project Status

sortTF is actively maintained and used in production environments. The project follows semantic versioning and aims for backward compatibility.

- **Latest Release**: See [GitHub Releases](https://github.com/obergerkatz/sortTF/releases)
- **Changelog**: See [GitHub Releases](https://github.com/obergerkatz/sortTF/releases) for version history
- **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/obergerkatz/sortTF/issues)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [HashiCorp HCL](https://github.com/hashicorp/hcl)
- Inspired by `terraform fmt` and similar formatting tools
- Thanks to the Go community for excellent tooling and libraries

---

Made with ❤️ for the Terraform community
