package combiner

// Config holds the configuration for OpenAPI specification combination
type Config struct {
	// Input/Output settings
	OutputFile string // Output file path
	Format     string // Output format: "yaml" or "json"

	// Combined API metadata
	Title   string // API title for the combined specification
	Version string // API version for the combined specification
	BaseURL string // Base URL to prepend to all paths

	// Service configuration
	ConfigFile    string            // Services configuration file path
	ServicePrefix map[string]string // Service name to path prefix mapping

	// Filtering options
	IncludeTags []string // Only include operations with these tags
	ExcludeTags []string // Exclude operations with these tags

	// Combination settings
	MergeSchemas   bool // Merge duplicate schemas in components section
	ValidateOutput bool // Validate the combined OpenAPI specification

	// Generation settings
	Verbose bool // Enable verbose output
}

// ServicesConfig represents the structure of services.yaml configuration file
type ServicesConfig struct {
	// Global settings
	Title       string `yaml:"title,omitempty"`
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
	BaseURL     string `yaml:"base_url,omitempty"`

	// Services list
	Services []ServiceConfig `yaml:"services"`

	// Global combination settings
	Settings CombinationSettings `yaml:"settings,omitempty"`
}

// ServiceConfig represents configuration for a single service
type ServiceConfig struct {
	Name         string   `yaml:"name"`
	SpecFile     string   `yaml:"spec_file"`
	PathPrefix   string   `yaml:"path_prefix,omitempty"`
	Tags         []string `yaml:"tags,omitempty"`
	Enabled      bool     `yaml:"enabled,omitempty"`
	Description  string   `yaml:"description,omitempty"`
	HealthCheck  string   `yaml:"health_check,omitempty"`
	Version      string   `yaml:"version,omitempty"`
}

// CombinationSettings holds global settings for spec combination
type CombinationSettings struct {
	MergeSchemas     bool     `yaml:"merge_schemas,omitempty"`
	ValidateOutput   bool     `yaml:"validate_output,omitempty"`
	IncludeTags      []string `yaml:"include_tags,omitempty"`
	ExcludeTags      []string `yaml:"exclude_tags,omitempty"`
	ConflictStrategy string   `yaml:"conflict_strategy,omitempty"` // "override", "merge", "error"
}

// CombinationStats holds statistics about the combination process
type CombinationStats struct {
	InputFiles       int // Number of input specification files
	ServicesCombined int // Number of services successfully combined
	TotalOperations  int // Total number of operations in combined spec
	TotalPaths       int // Total number of paths in combined spec
	MergedSchemas    int // Number of schemas merged in components section
	Conflicts        int // Number of path/operation conflicts resolved
}

// Default values for configuration
const (
	DefaultTitle          = "Combined API"
	DefaultVersion        = "1.0.0"
	DefaultFormat         = "yaml"
	DefaultOutputFile     = "combined-api.yaml"
	DefaultMergeSchemas   = true
	DefaultValidateOutput = true
)