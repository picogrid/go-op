package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/picogrid/go-op/internal/combiner"
	"github.com/spf13/cobra"
)

var combineCmd = &cobra.Command{
	Use:   "combine",
	Short: "Combine multiple OpenAPI specifications into one",
	Long: `Combine multiple OpenAPI 3.1 specifications from different microservices
into a single unified specification file. This is useful for creating a
comprehensive API gateway documentation or for service mesh configurations.

Examples:
  # Combine specs from files
  go-op combine -o combined.yaml user-service.yaml order-service.yaml notification-service.yaml

  # Combine specs with custom title and base path
  go-op combine -o api-gateway.yaml -t "E-commerce API Gateway" -b "/api/v1" *.yaml

  # Combine specs with service prefix mapping
  go-op combine -o combined.yaml -c services.yaml

  # Generate JSON output instead of YAML
  go-op combine -o combined.json -f json user-*.yaml order-*.yaml`,
	RunE: runCombine,
}

var (
	combineOutput  string
	combineTitle   string
	combineVersion string
	combineBaseURL string
	combineFormat  string
	combineConfig  string
	combineVerbose bool
	servicePrefix  []string
	includeTags    []string
	excludeTags    []string
	mergeSchemas   bool
	validateOutput bool
)

func init() {
	rootCmd.AddCommand(combineCmd)

	// Output flags
	combineCmd.Flags().StringVarP(&combineOutput, "output", "o", "combined-api.yaml", "output file path")
	combineCmd.Flags().StringVarP(&combineFormat, "format", "f", "yaml", "output format (yaml or json)")

	// API metadata flags
	combineCmd.Flags().StringVarP(&combineTitle, "title", "t", "Combined API", "API title for the combined specification")
	combineCmd.Flags().StringVarP(&combineVersion, "version", "V", "1.0.0", "API version for the combined specification")
	combineCmd.Flags().StringVarP(&combineBaseURL, "base-url", "b", "", "base URL to prepend to all paths (e.g., '/api/v1')")

	// Advanced combination flags
	combineCmd.Flags().StringVarP(&combineConfig, "config", "c", "", "services configuration file (YAML)")
	combineCmd.Flags().StringSliceVarP(&servicePrefix, "prefix", "p", []string{}, "service prefix mapping (format: 'service:/prefix')")
	combineCmd.Flags().StringSliceVar(&includeTags, "include-tags", []string{}, "only include operations with these tags")
	combineCmd.Flags().StringSliceVar(&excludeTags, "exclude-tags", []string{}, "exclude operations with these tags")
	combineCmd.Flags().BoolVar(&mergeSchemas, "merge-schemas", true, "merge duplicate schemas in components section")
	combineCmd.Flags().BoolVar(&validateOutput, "validate", true, "validate the combined OpenAPI specification")

	// Global flags
	combineCmd.Flags().BoolVarP(&combineVerbose, "verbose", "v", false, "enable verbose output")
}

func runCombine(cmd *cobra.Command, args []string) error {
	if combineVerbose {
		verbose = true
	}

	verbosePrint("Starting OpenAPI specification combination...")
	verbosePrint("Output file: %s", combineOutput)
	verbosePrint("Format: %s", combineFormat)

	// Validate arguments
	if len(args) == 0 && combineConfig == "" {
		return fmt.Errorf("no input files specified. Provide OpenAPI spec files as arguments or use --config flag")
	}

	// Resolve absolute output path
	absOutputFile, err := filepath.Abs(combineOutput)
	if err != nil {
		return fmt.Errorf("failed to resolve output file path: %w", err)
	}

	verbosePrint("Resolved output file: %s", absOutputFile)

	// Create combiner configuration
	config := &combiner.Config{
		OutputFile:     absOutputFile,
		Format:         combineFormat,
		Title:          combineTitle,
		Version:        combineVersion,
		BaseURL:        combineBaseURL,
		ConfigFile:     combineConfig,
		ServicePrefix:  parseServicePrefixes(servicePrefix),
		IncludeTags:    includeTags,
		ExcludeTags:    excludeTags,
		MergeSchemas:   mergeSchemas,
		ValidateOutput: validateOutput,
		Verbose:        combineVerbose,
	}

	// Create and run the combiner
	c := combiner.New(config)

	if combineConfig != "" {
		verbosePrint("Loading services from configuration file: %s", combineConfig)
		if err := c.LoadFromConfig(); err != nil {
			return fmt.Errorf("failed to load services configuration: %w", err)
		}
	}

	// Add input files from command line arguments
	for _, file := range args {
		verbosePrint("Adding input file: %s", file)
		absFile, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("failed to resolve input file %s: %w", file, err)
		}
		c.AddInputFile(absFile)
	}

	verbosePrint("Loading and parsing OpenAPI specifications...")
	if err := c.LoadSpecs(); err != nil {
		return fmt.Errorf("failed to load specifications: %w", err)
	}

	verbosePrint("Combining specifications...")
	if err := c.CombineSpecs(); err != nil {
		return fmt.Errorf("failed to combine specifications: %w", err)
	}

	if validateOutput {
		verbosePrint("Validating combined specification...")
		if err := c.ValidateOutput(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	verbosePrint("Writing combined specification...")
	if err := c.WriteOutput(); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("✅ Combined OpenAPI specification generated successfully: %s\n", absOutputFile)

	if combineVerbose {
		stats := c.GetStats()
		fmt.Printf("📊 Combination statistics:\n")
		fmt.Printf("   Input files: %d\n", stats.InputFiles)
		fmt.Printf("   Total operations: %d\n", stats.TotalOperations)
		fmt.Printf("   Total paths: %d\n", stats.TotalPaths)
		fmt.Printf("   Merged schemas: %d\n", stats.MergedSchemas)
		fmt.Printf("   Services combined: %d\n", stats.ServicesCombined)
	}

	return nil
}

// parseServicePrefixes parses service prefix mappings from the format "service:/prefix"
func parseServicePrefixes(prefixes []string) map[string]string {
	result := make(map[string]string)
	for _, prefix := range prefixes {
		parts := strings.SplitN(prefix, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}
