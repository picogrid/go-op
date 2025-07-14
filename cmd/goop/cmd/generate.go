package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/picogrid/go-op/internal/generator"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate OpenAPI specification from Go source code",
	Long: `Generate OpenAPI 3.1 specification from Go source code using go-op operations.

This command scans your Go source code for go-op operation definitions and generates
a complete OpenAPI specification file. It uses static analysis (go/ast) to extract
schema information without requiring runtime execution.

Examples:
  # Generate spec from current directory
  go-op generate

  # Generate from specific input directory
  go-op generate -i ./api -o ./openapi.yaml

  # Generate with custom title and version
  go-op generate -t "My API" -V "2.0.0"

  # Generate with verbose output
  go-op generate -v -i ./api`,
	RunE: runGenerate,
}

var (
	inputDir    string
	outputFile  string
	title       string
	version     string
	description string
	servers     []string
	format      string
)

func init() {
	rootCmd.AddCommand(generateCmd)

	// Input/Output flags
	generateCmd.Flags().StringVarP(&inputDir, "input", "i", ".", "input directory to scan for Go files")
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "openapi.yaml", "output file path")
	generateCmd.Flags().StringVarP(&format, "format", "f", "yaml", "output format (yaml or json)")

	// OpenAPI metadata flags
	generateCmd.Flags().StringVarP(&title, "title", "t", "", "API title (auto-detected if not specified)")
	generateCmd.Flags().StringVarP(&version, "version", "V", "1.0.0", "API version")
	generateCmd.Flags().StringVarP(&description, "description", "d", "", "API description")
	generateCmd.Flags().StringSliceVarP(&servers, "server", "s", []string{}, "server URLs (can be specified multiple times)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	verbosePrint("Starting OpenAPI generation...")
	verbosePrint("Input directory: %s", inputDir)
	verbosePrint("Output file: %s", outputFile)

	// Resolve absolute paths
	absInputDir, err := filepath.Abs(inputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve input directory: %w", err)
	}

	absOutputFile, err := filepath.Abs(outputFile)
	if err != nil {
		return fmt.Errorf("failed to resolve output file: %w", err)
	}

	verbosePrint("Resolved input directory: %s", absInputDir)
	verbosePrint("Resolved output file: %s", absOutputFile)

	// Create generator configuration
	config := &generator.Config{
		InputDir:    absInputDir,
		OutputFile:  absOutputFile,
		Format:      format,
		Title:       title,
		Version:     version,
		Description: description,
		Servers:     servers,
		Verbose:     verbose,
	}

	// Create and run the generator
	gen := generator.New(config)

	verbosePrint("Scanning for go-op operations...")
	if err := gen.ScanOperations(); err != nil {
		return fmt.Errorf("failed to scan operations: %w", err)
	}

	verbosePrint("Generating OpenAPI specification...")
	if err := gen.GenerateSpec(); err != nil {
		return fmt.Errorf("failed to generate specification: %w", err)
	}

	verbosePrint("Writing specification to file...")
	if err := gen.WriteSpec(); err != nil {
		return fmt.Errorf("failed to write specification: %w", err)
	}

	fmt.Printf("✅ OpenAPI specification generated successfully: %s\n", absOutputFile)

	if verbose {
		stats := gen.GetStats()
		fmt.Printf("📊 Generation statistics:\n")
		fmt.Printf("   Operations: %d\n", stats.OperationCount)
		fmt.Printf("   Schemas: %d\n", stats.SchemaCount)
		fmt.Printf("   Paths: %d\n", stats.PathCount)
	}

	return nil
}
