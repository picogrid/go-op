package generator

// Config holds the configuration for OpenAPI generation
type Config struct {
	// Input/Output settings
	InputDir   string   // Directory to scan for Go files
	OutputFile string   // Output file path
	Format     string   // Output format: "yaml" or "json"

	// OpenAPI metadata
	Title       string   // API title
	Version     string   // API version
	Description string   // API description
	Servers     []string // Server URLs

	// Generation settings
	Verbose bool // Enable verbose output
}

// GenerationStats holds statistics about the generation process
type GenerationStats struct {
	OperationCount int // Number of operations found
	SchemaCount    int // Number of schemas generated
	PathCount      int // Number of paths in the spec
	FileCount      int // Number of Go files scanned
}