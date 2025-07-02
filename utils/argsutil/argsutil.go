// Package argsutil provides argument parsing and config helpers for CLI code.
package argsutil

import (
	"bytes"
	"flag"
	"fmt"
	"io"
)

type Config struct {
	Root      string
	Recursive bool
	DryRun    bool
	Verbose   bool
	Validate  bool
}

// parseFlags parses command line arguments and returns a Config
func ParseFlags(args []string, stderr io.Writer) (*Config, error) {
	fs := flag.NewFlagSet("sorttf", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Suppress default error output

	var config Config

	fs.BoolVar(&config.Recursive, "recursive", false, "Scan directories recursively")
	fs.BoolVar(&config.DryRun, "dry-run", false, "Show what would be changed without writing (shows a unified diff)")
	fs.BoolVar(&config.Verbose, "verbose", false, "Print detailed logs about which files were parsed, sorted, and formatted")
	fs.BoolVar(&config.Validate, "validate", false, "Exit with a non-zero code if any files are not sorted/formatted")

	// Custom usage function
	fs.Usage = func() {
		_, _ = fmt.Fprintf(stderr, "Usage: sorttf [flags] [path]\n")
		_, _ = fmt.Fprintf(stderr, "\nSort and format Terraform (.tf) and Terragrunt (.hcl) files for consistency and readability.\n")
		_, _ = fmt.Fprintf(stderr, "\nPath can be a file or directory. If no path is provided, the current directory is used.\n")
		_, _ = fmt.Fprintf(stderr, "\nFlags:\n")

		// Create a temporary buffer to capture flag output
		var flagOutput bytes.Buffer
		fs.SetOutput(&flagOutput)
		fs.PrintDefaults()
		fs.SetOutput(io.Discard) // Reset to discard

		_, _ = fmt.Fprintf(stderr, "%s", flagOutput.String())
		_, _ = fmt.Fprintf(stderr, "\nExamples:\n")
		_, _ = fmt.Fprintf(stderr, "  sorttf .                    # Sort and format files in current directory\n")
		_, _ = fmt.Fprintf(stderr, "  sorttf main.tf              # Sort and format a specific file\n")
		_, _ = fmt.Fprintf(stderr, "  sorttf --recursive .        # Recursively process subdirectories\n")
		_, _ = fmt.Fprintf(stderr, "  sorttf --validate .         # Check if files are properly sorted/formatted\n")
		_, _ = fmt.Fprintf(stderr, "  sorttf --dry-run .          # Show what would change, with a unified diff\n")
	}

	if err := fs.Parse(args); err != nil {
		// Don't treat help request as an error
		if err.Error() == "flag: help requested" {
			return nil, fmt.Errorf("help")
		}
		return nil, fmt.Errorf("parseFlags: %w", err)
	}

	// Get positional arguments
	positionalArgs := fs.Args()
	if len(positionalArgs) > 1 {
		return nil, fmt.Errorf("parseFlags: too many arguments provided")
	}

	// Set root directory
	if len(positionalArgs) == 0 {
		config.Root = "."
	} else {
		config.Root = positionalArgs[0]
	}

	return &config, nil
}

// ... code will be added in the next step ...
