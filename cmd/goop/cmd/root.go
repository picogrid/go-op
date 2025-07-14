package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goop",
	Short: "Go Operations & Parsing - OpenAPI generation tool for go-op framework",
	Long: `go-op is a CLI tool for generating OpenAPI specifications from Go code using the go-op framework.

It provides build-time generation of OpenAPI specs, microservice spec combination,
and validation tools for maintaining high-quality API documentation.

Key features:
- Generate OpenAPI specs from Go source code
- Combine multiple microservice specs
- Validate and diff OpenAPI specifications
- Support for OpenAPI 3.1`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default is go-op.yaml)")
}

var (
	verbose    bool
	configFile string
)

// Helper function to print verbose output
func verbosePrint(format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] "+format+"\n", args...)
	}
}
