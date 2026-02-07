package hcl

import (
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

// BlockType represents the type of a Terraform block.
// Block types are used to determine sorting order in HCL files.
type BlockType string

// Terraform block type constants.
// These define the canonical block types and their sorting order.
const (
	BlockTypeTerraform BlockType = "terraform" // Terraform configuration block
	BlockTypeProvider  BlockType = "provider"  // Provider configuration
	BlockTypeVariable  BlockType = "variable"  // Input variable
	BlockTypeOutput    BlockType = "output"    // Output value
	BlockTypeResource  BlockType = "resource"  // Managed resource
	BlockTypeData      BlockType = "data"      // Data source
	BlockTypeModule    BlockType = "module"    // Module invocation
	BlockTypeLocals    BlockType = "locals"    // Local values
	BlockTypeBackend   BlockType = "backend"   // Backend configuration
	BlockTypeOther     BlockType = "other"     // Unknown block type
)

// Block represents a parsed HCL block with its type, labels, and content.
// It is used internally for sorting blocks by type and labels.
type Block struct {
	Type   BlockType       // Block type (terraform, provider, resource, etc.)
	Labels []string        // Block labels (e.g., ["aws", "instance"] for a resource)
	Block  *hclwrite.Block // The actual HCL block
}

// blockTypeOrder defines the canonical order in which block types should appear.
// Lower numbers appear first. This follows Terraform best practices where
// configuration blocks (terraform, provider) come before declarations (variable)
// which come before implementations (resource, module).
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

// getBlockType determines the type of a block based on its name.
// The check is case-insensitive. Returns BlockTypeOther for unknown block types.
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

// copyBlockClean creates a clean copy of a block without token baggage.
// This prevents excessive blank lines from removed comments being carried over.
// The function recursively copies attributes and nested blocks.
func copyBlockClean(src *hclwrite.Block) *hclwrite.Block {
	// Create new block with same type and labels
	newBlock := hclwrite.NewBlock(src.Type(), src.Labels())

	// Get source body
	srcBody := src.Body()
	newBody := newBlock.Body()

	// Copy attributes only (no comment tokens)
	attributes := srcBody.Attributes()

	// If for_each exists, write it first
	if attr, ok := attributes["for_each"]; ok {
		newBody.SetAttributeRaw("for_each", attr.Expr().BuildTokens(nil))
	}

	// Get sorted attribute names (excluding for_each)
	var attrNames []string
	for name := range attributes {
		if name != "for_each" {
			attrNames = append(attrNames, name)
		}
	}
	sort.Strings(attrNames)

	// Copy attributes in sorted order
	for _, name := range attrNames {
		newBody.SetAttributeRaw(name, attributes[name].Expr().BuildTokens(nil))
	}

	// Recursively copy nested blocks
	nestedBlocks := srcBody.Blocks()
	if len(nestedBlocks) > 0 {
		// Sort nested blocks by type and labels
		sort.SliceStable(nestedBlocks, func(i, j int) bool {
			typeI := getBlockType(nestedBlocks[i].Type())
			typeJ := getBlockType(nestedBlocks[j].Type())
			typeOrderI := blockTypeOrder[typeI]
			typeOrderJ := blockTypeOrder[typeJ]

			if typeOrderI != typeOrderJ {
				return typeOrderI < typeOrderJ
			}

			return compareLabels(nestedBlocks[i].Labels(), nestedBlocks[j].Labels())
		})

		// Copy nested blocks cleanly
		for _, nestedBlock := range nestedBlocks {
			newBody.AppendBlock(copyBlockClean(nestedBlock))
		}
	}

	return newBlock
}

// SortHCLFile sorts all blocks and attributes in an HCL file.
//
// It sorts blocks by type according to Terraform conventions
// (terraform, provider, variable, locals, data, resource, module, output),
// then alphabetically by labels within each type.
// Attributes within blocks are sorted alphabetically, with for_each always first.
//
// Returns a new hclwrite.File with sorted content.
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
		// Create a clean copy of the block to avoid token baggage
		cleanBlock := copyBlockClean(block.Block)

		// Add the clean block to the new file
		body.AppendBlock(cleanBlock)

		// Add a newline after each block except the last one
		if i < len(blocks)-1 {
			body.AppendNewline()
		}
	}

	return newFile
}

// parseBlocks extracts all top-level blocks from an HCL body.
// Backend blocks at the top level are skipped as they should only
// appear nested within terraform blocks.
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

// sortBlocks sorts blocks by type (using blockTypeOrder) and then
// alphabetically by labels within each type. Uses stable sort to
// preserve relative order when keys are equal.
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

// compareLabels compares two label slices lexicographically.
// Returns true if labels1 should sort before labels2.
// If all common labels are equal, shorter slices sort first.
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

// sortBlockAttributes sorts attributes within a block alphabetically,
// with special handling for for_each which is always placed first.
// Also recursively sorts nested blocks by type and labels.
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

// SortBlocksByType sorts blocks by their type according to Terraform conventions.
// The sorting order is: terraform, provider, variable, locals, data, resource, module, output.
// The input slice is modified in place and also returned.
func SortBlocksByType(blocks []Block) []Block {
	sort.SliceStable(blocks, func(i, j int) bool {
		return blockTypeOrder[blocks[i].Type] < blockTypeOrder[blocks[j].Type]
	})
	return blocks
}

// SortBlocksByLabels sorts blocks with the same type by their labels alphabetically.
// Labels are compared lexicographically. The input slice is modified in place and also returned.
func SortBlocksByLabels(blocks []Block) []Block {
	sort.SliceStable(blocks, func(i, j int) bool {
		return compareLabels(blocks[i].Labels, blocks[j].Labels)
	})
	return blocks
}

// SortAttributes returns a sorted list of attribute names from a map of attributes.
// The names are sorted alphabetically. This is a convenience function for sorting attributes.
func SortAttributes(attributes map[string]*hclwrite.Attribute) []string {
	names := make([]string, 0, len(attributes))
	for name := range attributes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// SortAndFormatHCLFile sorts all blocks and attributes in an HCL file and returns the formatted string.
// This is the main entry point that combines sorting and formatting in one operation.
// It first sorts the file using SortHCLFile, then formats it using FormatHCLFile.
// Returns the formatted content as a string, or an HCLError with KindSorting if sorting or formatting fails.
func SortAndFormatHCLFile(file *hclwrite.File) (string, error) {
	sorted := SortHCLFile(file)
	formatted, err := FormatHCLFile(sorted)
	if err != nil {
		return formatted, &HCLError{
			Op:   "SortAndFormatHCLFile",
			Kind: KindSorting,
			Err:  err,
		}
	}
	return formatted, nil
}
