# sortTF

A powerful command-line tool for sorting and formatting Terraform (.tf) and Terragrunt (.hcl) files to ensure consistency and readability across your infrastructure code.

## 🚀 Features at a Glance

- **Smart Block Sorting**: Orders Terraform blocks for readability and best practices.
- **Attribute Sorting**: Alphabetizes attributes, with special handling for `for_each`.
- **Nested Block Support**: Handles deeply nested and complex HCL structures.
- **Formatting**: Applies `terraform fmt` standards.
- **Multiple Modes**: Dry-run, validation, and recursive directory processing.
- **Comprehensive Error Handling**: Detailed, colorized error messages.
- **Cross-Platform**: Works on macOS, Linux, and Windows.
- **Tested & Modular**: Well-tested, DRY, and maintainable codebase.

## ⚡ Quickstart

```bash
# Install (from source)
git clone https://github.com/obergerkatz/sortTF.git
cd sortTF
go build -o sorttf
sudo mv sorttf /usr/local/bin/

# Or using Go install
go install github.com/OBerger96/sortTF@latest

# Basic usage
sorttf .
```

## 🛠️ Prerequisites

- **Go 1.24.4+** (for building from source)
- No external dependencies required (Terraform not needed)

## 📖 Usage

### Basic Usage

```bash
sorttf .                # Sort and format files in current directory
sorttf main.tf          # Sort and format a specific file
sorttf --recursive .    # Recursively process subdirectories
sorttf --dry-run .      # Show what would change without writing
sorttf --validate .     # Validate files without making changes
sorttf --verbose .      # Verbose output
```

### Command Line Options

| Flag         | Description                                                        |
|--------------|--------------------------------------------------------------------|
| `--recursive`| Scan directories recursively                                       |
| `--dry-run`  | Show what would be changed without writing (shows a unified diff)  |
| `--verbose`  | Print detailed logs about which files were parsed, sorted, and formatted |
| `--validate` | Exit with a non-zero code if any files are not sorted/formatted    |

### Examples

```bash
sorttf main.tf
sorttf --recursive --verbose .
sorttf --validate --recursive .
sorttf --dry-run --recursive .
```

## 🔧 How It Works

### Block Sorting Order

sortTF sorts Terraform blocks in the following order:

1. **terraform**
2. **provider**
3. **variable**
4. **locals**
5. **data**
6. **resource**
7. **module**
8. **output**

### Attribute Sorting

- `for_each` is always placed first (if present)
- Other attributes are sorted alphabetically
- Nested blocks are sorted by type and then by labels

#### Example Transformation

**Before:**
```hcl
resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami           = "ami-123456"
  tags = { Name = "web-server" }
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
  tags = { Name = "web-server" }
}
```

## 🧪 Testing

sortTF has comprehensive test coverage with both unit and integration tests:

```bash
# Run all tests (unit + integration)
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage report
go test -cover ./...

# Run only integration tests
go test -v ./integration/...

# Run only unit tests (exclude integration)
go test ./... -short
```

### Test Suite

- **73+ test functions** covering all major functionality
- **Unit tests** for individual packages (CLI, config, errors, files, lib)
- **Integration tests** that execute the actual binary end-to-end
- **Coverage**: 75-100% across packages (CLI: 82%, config: 100%, lib: 75%)

The integration tests verify:
- Complete CLI workflows from argument parsing to file output
- Real-world Terraform configurations
- Error handling and edge cases
- CI/CD validation mode
- Concurrent file processing
- Attribute and block sorting accuracy

See [integration/README.md](integration/README.md) for details on the integration test suite.

## 🏗️ Development

### Project Structure

```
sortTF/
├── main.go              # Application entry point
├── cli/                 # Command-line interface and execution
├── config/              # Configuration and flag parsing
├── hcl/                 # HCL parsing, sorting, formatting, and errors
├── lib/                 # Reusable library API for programmatic use
├── integration/         # End-to-end integration tests
├── internal/
│   ├── errors/         # Unified error handling
│   └── files/          # File traversal and validation
└── examples/            # Library usage examples
```

### Building

```bash
# Build for current platform
go build -o sorttf

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -o sorttf-linux
GOOS=darwin GOARCH=amd64 go build -o sorttf-darwin
GOOS=windows GOARCH=amd64 go build -o sorttf-windows.exe
```

### Code Quality

```bash
# Run linter
go vet ./...

# Format code
go fmt ./...

# Run tests with coverage
go test -cover ./...
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go coding standards
- Add tests for new features
- Update documentation as needed
- Ensure all tests pass before submitting PR

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🐛 Known Issues

- Some edge cases with deeply nested blocks may need manual review
- Comments are preserved but may be repositioned during sorting

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/obergerkatz/sortTF/issues)

## 🙏 Acknowledgments

- Built with [HashiCorp HCL](https://github.com/hashicorp/hcl)
- Inspired by `terraform fmt` and similar formatting tools
- Thanks to the Go community for excellent tooling and libraries 
