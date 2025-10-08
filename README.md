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
- **Terraform** (for formatting functionality)

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

```bash
go test ./...           # Run all tests
go test -v ./...        # Verbose output
go test -cover ./...    # Run with coverage
```

## 🏗️ Development

### Project Structure

```
sortTF/
├── main.go                 # Application entry point
├── utils/
│   ├── cliutil/           # Command-line interface
│   ├── argsutil/          # CLI argument parsing
│   ├── errorutil/         # Error helpers and types
│   ├── fileutil/          # File system operations
│   ├── formattingutil/    # HCL formatting
│   ├── parsingutil/       # HCL parsing and validation
│   └── sortingutil/       # Block and attribute sorting
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

- Requires Terraform to be installed for formatting functionality
- Some edge cases with deeply nested blocks may need manual review
- Comments are preserved but may be repositioned during sorting

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/sortTF/issues)

## 🙏 Acknowledgments

- Built with [HashiCorp HCL](https://github.com/hashicorp/hcl)
- Inspired by `terraform fmt` and similar formatting tools
- Thanks to the Go community for excellent tooling and libraries 
