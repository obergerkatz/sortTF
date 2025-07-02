package sortingutil

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"sorttf/utils/formattingutil"
)

// SortingError represents an error during sorting or formatting
// It wraps the operation, file path, and the underlying error
// Path is optional (may be empty for in-memory operations)
type SortingError struct {
	Op   string
	Path string
	Err  error
}

func (e *SortingError) Error() string {
	if e.Err != nil {
		if e.Path != "" {
			return fmt.Sprintf("sortingutil %s %s: %v", e.Op, e.Path, e.Err)
		}
		return fmt.Sprintf("sortingutil %s: %v", e.Op, e.Err)
	}
	if e.Path != "" {
		return fmt.Sprintf("sortingutil %s %s", e.Op, e.Path)
	}
	return fmt.Sprintf("sortingutil %s", e.Op)
}

func (e *SortingError) Unwrap() error {
	return e.Err
}

// BlockType represents the type of a Terraform block
type BlockType string

const (
	BlockTypeTerraform BlockType = "terraform"
	BlockTypeProvider  BlockType = "provider"
	BlockTypeVariable  BlockType = "variable"
	BlockTypeOutput    BlockType = "output"
	BlockTypeResource  BlockType = "resource"
	BlockTypeData      BlockType = "data"
	BlockTypeModule    BlockType = "module"
	BlockTypeLocals    BlockType = "locals"
	BlockTypeBackend   BlockType = "backend"
	BlockTypeOther     BlockType = "other"
)

// Block represents a parsed HCL block with its type, labels, and content
type Block struct {
	Type   BlockType
	Labels []string
	Block  *hclwrite.Block
}

// blockTypeOrder defines the order in which block types should appear
var blockTypeOrder = map[BlockType]int{
	BlockTypeTerraform: 1,
	BlockTypeProvider:  2,
	BlockTypeVariable:  3,
	BlockTypeLocals:    4,
	BlockTypeData:      5,
	BlockTypeResource:  6,
	BlockTypeModule:    7,
	BlockTypeOutput:    8,
	BlockTypeBackend:   9,
	BlockTypeOther:     10,
}

// getBlockType determines the type of a block based on its name
func getBlockType(name string) BlockType {
	switch strings.ToLower(name) {
	case "terraform":
		return BlockTypeTerraform
	case "provider":
		return BlockTypeProvider
	case "variable":
		return BlockTypeVariable
	case "output":
		return BlockTypeOutput
	case "resource":
		return BlockTypeResource
	case "data":
		return BlockTypeData
	case "module":
		return BlockTypeModule
	case "locals":
		return BlockTypeLocals
	case "backend":
		return BlockTypeBackend
	default:
		return BlockTypeOther
	}
}

// SortHCLFile sorts all blocks and attributes in an HCL file
func SortHCLFile(file *hclwrite.File) *hclwrite.File {
	if file == nil {
		return hclwrite.NewEmptyFile()
	}

	// Parse blocks from the file
	blocks := parseBlocks(file.Body())

	// Sort blocks
	sortBlocks(blocks)

	// Create a new file with sorted content
	newFile := hclwrite.NewEmptyFile()
	body := newFile.Body()

	// Add sorted blocks to the new file
	for i, block := range blocks {
		// Sort attributes within the block
		sortBlockAttributes(block.Block)

		// Add the block to the new file
		body.AppendBlock(block.Block)

		// Add a newline after each block except the last one
		if i < len(blocks)-1 {
			body.AppendNewline()
		}
	}

	return newFile
}

// parseBlocks extracts all blocks from an HCL body
func parseBlocks(body *hclwrite.Body) []Block {
	var blocks []Block

	for _, block := range body.Blocks() {
		blockType := getBlockType(block.Type())

		// Handle backend blocks specially - they should be treated as nested blocks within terraform
		if blockType == BlockTypeBackend {
			// Skip backend blocks at the top level - they should be nested within terraform blocks
			continue
		}

		blocks = append(blocks, Block{
			Type:   blockType,
			Labels: block.Labels(),
			Block:  block,
		})
	}

	return blocks
}

// sortBlocks sorts blocks by type and then by labels
func sortBlocks(blocks []Block) {
	sort.SliceStable(blocks, func(i, j int) bool {
		// First, sort by block type order
		typeOrderI := blockTypeOrder[blocks[i].Type]
		typeOrderJ := blockTypeOrder[blocks[j].Type]

		if typeOrderI != typeOrderJ {
			return typeOrderI < typeOrderJ
		}

		// If same type, sort by labels
		return compareLabels(blocks[i].Labels, blocks[j].Labels)
	})
}

// compareLabels compares two label slices lexicographically
func compareLabels(labels1, labels2 []string) bool {
	minLen := len(labels1)
	if len(labels2) < minLen {
		minLen = len(labels2)
	}

	for i := 0; i < minLen; i++ {
		if labels1[i] != labels2[i] {
			return labels1[i] < labels2[i]
		}
	}

	// If all labels up to minLen are equal, shorter list comes first
	return len(labels1) < len(labels2)
}

// sortBlockAttributes sorts attributes within a block alphabetically
func sortBlockAttributes(block *hclwrite.Block) {
	if block == nil {
		return
	}

	body := block.Body()
	attributes := body.Attributes()

	// Get all attribute names
	var attrNames []string
	for name := range attributes {
		attrNames = append(attrNames, name)
	}

	// Remove all attributes
	for _, name := range attrNames {
		body.RemoveAttribute(name)
	}

	// If for_each exists, write it first
	if _, ok := attributes["for_each"]; ok {
		body.SetAttributeRaw("for_each", attributes["for_each"].Expr().BuildTokens(nil))
	}

	// Sort the rest alphabetically, skipping for_each
	var rest []string
	for _, name := range attrNames {
		if name != "for_each" {
			rest = append(rest, name)
		}
	}
	sort.Strings(rest)
	for _, name := range rest {
		body.SetAttributeRaw(name, attributes[name].Expr().BuildTokens(nil))
	}

	// Get all nested blocks and sort them
	nestedBlocks := body.Blocks()
	if len(nestedBlocks) > 0 {
		// Sort nested blocks by type and then by labels
		sort.SliceStable(nestedBlocks, func(i, j int) bool {
			// First sort by block type
			typeI := getBlockType(nestedBlocks[i].Type())
			typeJ := getBlockType(nestedBlocks[j].Type())
			typeOrderI := blockTypeOrder[typeI]
			typeOrderJ := blockTypeOrder[typeJ]

			if typeOrderI != typeOrderJ {
				return typeOrderI < typeOrderJ
			}

			// If same type, sort by labels
			return compareLabels(nestedBlocks[i].Labels(), nestedBlocks[j].Labels())
		})

		// Remove all nested blocks
		for _, nestedBlock := range nestedBlocks {
			body.RemoveBlock(nestedBlock)
		}

		// Re-add nested blocks in sorted order and sort their attributes
		for _, nestedBlock := range nestedBlocks {
			sortBlockAttributes(nestedBlock)
			body.AppendBlock(nestedBlock)
		}
	}
}

// SortBlocksByType sorts blocks by their type according to Terraform conventions
func SortBlocksByType(blocks []Block) []Block {
	sort.SliceStable(blocks, func(i, j int) bool {
		return blockTypeOrder[blocks[i].Type] < blockTypeOrder[blocks[j].Type]
	})
	return blocks
}

// SortBlocksByLabels sorts blocks with the same type by their labels
func SortBlocksByLabels(blocks []Block) []Block {
	sort.SliceStable(blocks, func(i, j int) bool {
		return compareLabels(blocks[i].Labels, blocks[j].Labels)
	})
	return blocks
}

// SortAttributes sorts attributes alphabetically by name
func SortAttributes(attributes map[string]*hclwrite.Attribute) []string {
	var names []string
	for name := range attributes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// SortAndFormatHCLFile sorts all blocks and attributes in an HCL file and returns the formatted string
func SortAndFormatHCLFile(file *hclwrite.File) (string, error) {
	sorted := SortHCLFile(file)
	formatted, err := formattingutil.FormatHCLFile(sorted)
	if err != nil {
		return formatted, &SortingError{
			Op:  "SortAndFormatHCLFile",
			Err: err,
		}
	}
	return formatted, nil
}

// Error helper functions
// IsSortingError checks if an error is a SortingError
func IsSortingError(err error) bool {
	_, ok := err.(*SortingError)
	return ok
}

// GetSortingErrorOp extracts the operation from a SortingError
func GetSortingErrorOp(err error) string {
	if sortingErr, ok := err.(*SortingError); ok {
		return sortingErr.Op
	}
	return ""
}

// GetSortingErrorPath extracts the path from a SortingError
func GetSortingErrorPath(err error) string {
	if sortingErr, ok := err.(*SortingError); ok {
		return sortingErr.Path
	}
	return ""
}
