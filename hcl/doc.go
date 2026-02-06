// Package hcl provides parsing, sorting, and formatting for HashiCorp Configuration Language (HCL) files.
//
// This package handles Terraform (.tf) and Terragrunt (.hcl) files, providing functionality to:
//   - Parse HCL files and validate their structure
//   - Sort blocks by type (terraform, provider, variable, etc.) and labels
//   - Sort attributes alphabetically within blocks (with for_each always first)
//   - Format files using terraform fmt standards
//
// The main entry points are:
//   - ParseHCLFile: Parse and validate an HCL file
//   - SortHCLFile: Sort blocks and attributes in an HCL file
//   - FormatHCLFile: Apply terraform fmt formatting
//   - SortAndFormatHCLFile: Combined sort and format operation
package hcl
