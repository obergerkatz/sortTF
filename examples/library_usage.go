// +build ignore

// This example demonstrates using sortTF as a library in your own Go programs.
package main

import (
	"errors"
	"fmt"
	"log"

	"sorttf/lib"
)

func main() {
	// Example 1: Sort a single file
	fmt.Println("=== Example 1: Sort a single file ===")
	err := lib.SortFile("main.tf", lib.Options{})
	if err != nil {
		if errors.Is(err, lib.ErrNoChanges) {
			fmt.Println("File is already sorted")
		} else {
			log.Fatalf("Error sorting file: %v", err)
		}
	} else {
		fmt.Println("File sorted successfully")
	}

	// Example 2: Validate files without modifying them
	fmt.Println("\n=== Example 2: Validate mode (CI/CD) ===")
	err = lib.SortFile("main.tf", lib.Options{Validate: true})
	if err != nil {
		if errors.Is(err, lib.ErrNeedsSorting) {
			fmt.Println("❌ File needs sorting - CI check failed")
			// Exit with error code in CI
		} else if errors.Is(err, lib.ErrNoChanges) {
			fmt.Println("✅ File is properly sorted")
		} else {
			log.Fatalf("Error validating: %v", err)
		}
	}

	// Example 3: Get sorted content for preview/diff
	fmt.Println("\n=== Example 3: Get sorted content ===")
	content, changed, err := lib.GetSortedContent("main.tf")
	if err != nil {
		log.Fatalf("Error getting sorted content: %v", err)
	}
	if changed {
		fmt.Println("File would be changed. New content:")
		fmt.Println(content)
	} else {
		fmt.Println("File is already sorted")
	}

	// Example 4: Sort multiple files
	fmt.Println("\n=== Example 4: Sort multiple files ===")
	files := []string{"main.tf", "variables.tf", "outputs.tf"}
	results := lib.SortFiles(files, lib.Options{})
	for path, err := range results {
		if err != nil && !errors.Is(err, lib.ErrNoChanges) {
			fmt.Printf("❌ %s: %v\n", path, err)
		} else if errors.Is(err, lib.ErrNoChanges) {
			fmt.Printf("✅ %s: already sorted\n", path)
		} else {
			fmt.Printf("✅ %s: sorted successfully\n", path)
		}
	}

	// Example 5: Sort entire directory
	fmt.Println("\n=== Example 5: Sort directory ===")
	results, err = lib.SortDirectory("./terraform", true, lib.Options{})
	if err != nil {
		log.Fatalf("Error sorting directory: %v", err)
	}
	fmt.Printf("Processed %d files\n", len(results))
	errorCount := 0
	for _, err := range results {
		if err != nil && !errors.Is(err, lib.ErrNoChanges) {
			errorCount++
		}
	}
	if errorCount > 0 {
		fmt.Printf("Encountered %d errors\n", errorCount)
	}

	// Example 6: Integration with pre-commit hook
	fmt.Println("\n=== Example 6: Pre-commit hook usage ===")
	stagedFiles := []string{"main.tf", "variables.tf"} // From git diff --cached --name-only
	needsSorting := false
	for _, file := range stagedFiles {
		err := lib.SortFile(file, lib.Options{})
		if err != nil && !errors.Is(err, lib.ErrNoChanges) {
			log.Printf("Error sorting %s: %v", file, err)
			needsSorting = true
		}
	}
	if needsSorting {
		fmt.Println("Some files were sorted - please stage the changes")
	} else {
		fmt.Println("All staged files are properly sorted")
	}
}
