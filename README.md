# sortTF

A powerful command-line tool for sorting and formatting Terraform (.tf) and Terragrunt (.hcl) files to ensure consistency and readability across your infrastructure code.

## 🚀 Features

- **Smart Block Sorting**: Automatically sorts Terraform blocks in the correct order (terraform → provider → variable → locals → data → resource → module → output)
- **Attribute Sorting**: Sorts attributes within blocks alphabetically, with special handling for `for_each`
- **Nested Block Support**: Handles nested blocks within resources and providers
- **Formatting**: Applies `terraform fmt` formatting standards
- **Multiple Modes**: Dry-run, validation, and recursive directory processing
- **Comprehensive Error Handling**: Detailed error messages with file paths and context
- **Cross-Platform**: Works on macOS, Linux, and Windows

## 📦 Installation

### From Source
```bash
git clone https://github.com/yourusername/sortTF.git
cd sortTF
go build -o sorttf
# Move to your PATH
sudo mv sorttf /usr/local/bin/
```

### Using Go Install
```bash
go install github.com/OBerger96/sortTF@latest
```

## 🛠️ Prerequisites

- **Go 1.24.4+** (for building from source)
- **Terraform** (for formatting functionality)

## 📖 Usage

### Basic Usage

```bash
# Sort and format files in current directory
sorttf .

# Sort and format a specific file
sorttf main.tf

# Recursively process subdirectories
sorttf --recursive .

# Show what would change without writing
sorttf --dry-run .

# Validate files without making changes
sorttf --validate .

# Verbose output
sorttf --verbose .
```

### Command Line Options

| Flag | Description |
|------|-------------|
| `--recursive` | Scan directories recursively |
| `--dry-run` | Show what would be changed without writing (shows a unified diff) |
| `--verbose` | Print detailed logs about which files were parsed, sorted, and formatted |
| `--validate` | Exit with a non-zero code if any files are not sorted/formatted |

### Examples

#### Sort a Single File
```bash
sorttf main.tf
```

#### Process an Entire Project
```bash
sorttf --recursive --verbose .
```

#### Validate CI/CD Pipeline
```bash
sorttf --validate --recursive .
```

#### Preview Changes
```bash
sorttf --dry-run --recursive .
```

## 🔧 How It Works

### Block Sorting Order

sortTF sorts Terraform blocks in the following order:

1. **terraform** - Terraform configuration blocks
2. **provider** - Provider configurations
3. **variable** - Input variables
4. **locals** - Local values
5. **data** - Data sources
6. **resource** - Resource definitions
7. **module** - Module calls
8. **output** - Output values

### Attribute Sorting

Within each block, attributes are sorted alphabetically with special handling:

- `for_each` is always placed first (if present)
- Other attributes are sorted alphabetically
- Nested blocks are sorted by type and then by labels

### Example Transformation

**Before:**
```hcl
resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami           = "ami-123456"
  
  tags = {
    Name = "web-server"
  }
}

provider "aws" {
  region = "us-west-2"
}

variable "environment" {
  type = string
}
```

**After:**
```hcl
provider "aws" {
  region = "us-west-2"
}

variable "environment" {
  type = string
}

resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t3.micro"

  tags = {
    Name = "web-server"
  }
}
```

## 🧪 Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./utils/sortingutil

# Run integration tests
go test ./utils/sortingutil -run TestSortHCLFileWithRealFiles
```

## 🏗️ Development

### Project Structure

```
sortTF/
├── main.go                 # Application entry point
├── utils/
│   ├── cliutil/           # Command-line interface
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
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/sortTF/discussions)

## 🙏 Acknowledgments

- Built with [HashiCorp HCL](https://github.com/hashicorp/hcl)
- Inspired by `terraform fmt` and similar formatting tools
- Thanks to the Go community for excellent tooling and libraries 