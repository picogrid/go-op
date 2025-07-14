package combiner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/picogrid/go-op/operations"
	"gopkg.in/yaml.v3"
)

// Combiner handles combining multiple OpenAPI specifications
type Combiner struct {
	config     *Config
	inputFiles []string
	specs      []*SpecWithMetadata
	combined   *operations.OpenAPISpec
	stats      CombinationStats
}

// SpecWithMetadata wraps an OpenAPI spec with additional metadata
type SpecWithMetadata struct {
	Spec        *operations.OpenAPISpec
	SourceFile  string
	ServiceName string
	Prefix      string
}

// New creates a new OpenAPI specification combiner
func New(config *Config) *Combiner {
	return &Combiner{
		config:     config,
		inputFiles: make([]string, 0),
		specs:      make([]*SpecWithMetadata, 0),
		stats:      CombinationStats{},
	}
}

// AddInputFile adds an input OpenAPI specification file
func (c *Combiner) AddInputFile(filename string) {
	c.inputFiles = append(c.inputFiles, filename)
}

// LoadFromConfig loads service configuration from a YAML file
func (c *Combiner) LoadFromConfig() error {
	if c.config.ConfigFile == "" {
		return nil
	}

	configData, err := ioutil.ReadFile(c.config.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var serviceConfig ServicesConfig
	if err := yaml.Unmarshal(configData, &serviceConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Add services from configuration
	for _, service := range serviceConfig.Services {
		if service.SpecFile != "" {
			absPath, err := filepath.Abs(service.SpecFile)
			if err != nil {
				return fmt.Errorf("failed to resolve spec file path %s: %w", service.SpecFile, err)
			}
			c.inputFiles = append(c.inputFiles, absPath)

			// Store service metadata
			if c.config.ServicePrefix == nil {
				c.config.ServicePrefix = make(map[string]string)
			}
			if service.PathPrefix != "" {
				c.config.ServicePrefix[service.Name] = service.PathPrefix
			}
		}
	}

	if c.config.Verbose {
		fmt.Printf("[VERBOSE] Loaded %d services from config file\n", len(serviceConfig.Services))
	}

	return nil
}

// LoadSpecs loads and parses all input OpenAPI specifications
func (c *Combiner) LoadSpecs() error {
	c.stats.InputFiles = len(c.inputFiles)

	for _, filename := range c.inputFiles {
		if c.config.Verbose {
			fmt.Printf("[VERBOSE] Loading specification: %s\n", filename)
		}

		spec, err := c.loadSingleSpec(filename)
		if err != nil {
			return fmt.Errorf("failed to load spec from %s: %w", filename, err)
		}

		serviceName := c.extractServiceName(filename)
		prefix := c.config.ServicePrefix[serviceName]

		specWithMeta := &SpecWithMetadata{
			Spec:        spec,
			SourceFile:  filename,
			ServiceName: serviceName,
			Prefix:      prefix,
		}

		c.specs = append(c.specs, specWithMeta)

		if c.config.Verbose {
			pathCount := len(spec.Paths)
			fmt.Printf("[VERBOSE] Loaded %d paths from %s (service: %s, prefix: %s)\n",
				pathCount, filepath.Base(filename), serviceName, prefix)
		}
	}

	return nil
}

// loadSingleSpec loads a single OpenAPI specification file
func (c *Combiner) loadSingleSpec(filename string) (*operations.OpenAPISpec, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var spec operations.OpenAPISpec

	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &spec); err != nil {
			if jsonErr := json.Unmarshal(data, &spec); jsonErr != nil {
				return nil, fmt.Errorf("failed to parse as YAML or JSON: YAML error: %v, JSON error: %v", err, jsonErr)
			}
		}
	}

	return &spec, nil
}

// extractServiceName extracts service name from filename
func (c *Combiner) extractServiceName(filename string) string {
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// Remove common suffixes
	name = strings.TrimSuffix(name, "-service")
	name = strings.TrimSuffix(name, ".service")
	name = strings.TrimSuffix(name, "-api")
	name = strings.TrimSuffix(name, ".api")

	return name
}

// CombineSpecs combines all loaded specifications into a single one
func (c *Combiner) CombineSpecs() error {
	if len(c.specs) == 0 {
		return fmt.Errorf("no specifications loaded")
	}

	if c.config.Verbose {
		fmt.Printf("[VERBOSE] Combining %d specifications\n", len(c.specs))
	}

	// Create the base combined specification
	c.combined = &operations.OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: operations.OpenAPIInfo{
			Title:   c.config.Title,
			Version: c.config.Version,
		},
		Paths: make(map[string]map[string]operations.OpenAPIOperation),
	}

	// Combine paths from all specs
	for _, specMeta := range c.specs {
		c.stats.ServicesCombined++

		if err := c.combineSpecPaths(specMeta); err != nil {
			return fmt.Errorf("failed to combine paths from %s: %w", specMeta.SourceFile, err)
		}
	}

	// Merge schemas if requested
	if c.config.MergeSchemas {
		if err := c.mergeSchemas(); err != nil {
			return fmt.Errorf("failed to merge schemas: %w", err)
		}
	}

	c.stats.TotalPaths = len(c.combined.Paths)

	// Count total operations
	for _, pathMethods := range c.combined.Paths {
		c.stats.TotalOperations += len(pathMethods)
	}

	if c.config.Verbose {
		fmt.Printf("[VERBOSE] Combined specification has %d paths and %d operations\n",
			c.stats.TotalPaths, c.stats.TotalOperations)
	}

	return nil
}

// combineSpecPaths combines paths from a single specification
func (c *Combiner) combineSpecPaths(specMeta *SpecWithMetadata) error {
	for path, methods := range specMeta.Spec.Paths {
		// Apply path transformations
		finalPath := c.transformPath(path, specMeta)

		// Filter operations by tags if specified
		filteredMethods := c.filterMethodsByTags(methods)

		if len(filteredMethods) == 0 {
			continue // Skip if all operations filtered out
		}

		// Initialize path in combined spec if it doesn't exist
		if c.combined.Paths[finalPath] == nil {
			c.combined.Paths[finalPath] = make(map[string]operations.OpenAPIOperation)
		}

		// Add each HTTP method
		for method, operation := range filteredMethods {
			// Add service tag to operation
			operation.Tags = c.addServiceTag(operation.Tags, specMeta.ServiceName)

			// Check for conflicts
			if existingOp, exists := c.combined.Paths[finalPath][method]; exists {
				if c.config.Verbose {
					fmt.Printf("[VERBOSE] Warning: Overriding %s %s (was from %s, now from %s)\n",
						method, finalPath, c.findOperationSource(existingOp), specMeta.ServiceName)
				}
			}

			c.combined.Paths[finalPath][method] = operation
		}
	}

	return nil
}

// transformPath applies path transformations (base URL, service prefix)
func (c *Combiner) transformPath(originalPath string, specMeta *SpecWithMetadata) string {
	path := originalPath

	// Add service prefix if specified
	if specMeta.Prefix != "" {
		path = specMeta.Prefix + path
	}

	// Add base URL if specified
	if c.config.BaseURL != "" {
		path = c.config.BaseURL + path
	}

	// Normalize path (remove double slashes)
	path = strings.ReplaceAll(path, "//", "/")

	return path
}

// filterMethodsByTags filters operations based on include/exclude tag rules
func (c *Combiner) filterMethodsByTags(methods map[string]operations.OpenAPIOperation) map[string]operations.OpenAPIOperation {
	if len(c.config.IncludeTags) == 0 && len(c.config.ExcludeTags) == 0 {
		return methods // No filtering needed
	}

	filtered := make(map[string]operations.OpenAPIOperation)

	for method, operation := range methods {
		include := true

		// Check include tags (if specified, operation must have at least one)
		if len(c.config.IncludeTags) > 0 {
			include = false
			for _, tag := range operation.Tags {
				for _, includeTag := range c.config.IncludeTags {
					if tag == includeTag {
						include = true
						break
					}
				}
				if include {
					break
				}
			}
		}

		// Check exclude tags (if operation has any exclude tag, exclude it)
		if include && len(c.config.ExcludeTags) > 0 {
			for _, tag := range operation.Tags {
				for _, excludeTag := range c.config.ExcludeTags {
					if tag == excludeTag {
						include = false
						break
					}
				}
				if !include {
					break
				}
			}
		}

		if include {
			filtered[method] = operation
		}
	}

	return filtered
}

// addServiceTag adds a service tag to operation tags
func (c *Combiner) addServiceTag(tags []string, serviceName string) []string {
	serviceTag := fmt.Sprintf("service:%s", serviceName)

	// Check if service tag already exists
	for _, tag := range tags {
		if tag == serviceTag {
			return tags
		}
	}

	// Add service tag at the beginning
	return append([]string{serviceTag}, tags...)
}

// findOperationSource attempts to find the source of an operation (for conflict warnings)
func (c *Combiner) findOperationSource(operation operations.OpenAPIOperation) string {
	for _, tag := range operation.Tags {
		if strings.HasPrefix(tag, "service:") {
			return strings.TrimPrefix(tag, "service:")
		}
	}
	return "unknown"
}

// mergeSchemas merges duplicate schemas in the components section
func (c *Combiner) mergeSchemas() error {
	// For now, we'll implement a basic version that collects unique schemas
	// A full implementation would need sophisticated schema comparison and merging

	if c.config.Verbose {
		fmt.Printf("[VERBOSE] Schema merging not yet implemented - will be added in future version\n")
	}

	return nil
}

// ValidateOutput validates the combined OpenAPI specification
func (c *Combiner) ValidateOutput() error {
	if c.combined == nil {
		return fmt.Errorf("no combined specification available")
	}

	// Basic validation checks
	if c.combined.OpenAPI == "" {
		return fmt.Errorf("OpenAPI version is required")
	}

	if c.combined.Info.Title == "" {
		return fmt.Errorf("API title is required")
	}

	if c.combined.Info.Version == "" {
		return fmt.Errorf("API version is required")
	}

	// Validate that we have at least one path
	if len(c.combined.Paths) == 0 {
		return fmt.Errorf("combined specification has no paths")
	}

	// Validate each path has at least one operation
	for path, methods := range c.combined.Paths {
		if len(methods) == 0 {
			return fmt.Errorf("path %s has no operations", path)
		}

		// Validate each operation has required responses
		for method, operation := range methods {
			if len(operation.Responses) == 0 {
				return fmt.Errorf("operation %s %s has no responses", method, path)
			}
		}
	}

	if c.config.Verbose {
		fmt.Printf("[VERBOSE] Combined specification validation passed\n")
	}

	return nil
}

// WriteOutput writes the combined specification to the output file
func (c *Combiner) WriteOutput() error {
	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(c.config.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the spec in the specified format
	switch strings.ToLower(c.config.Format) {
	case "json":
		return c.writeJSON()
	case "yaml", "yml":
		return c.writeYAML()
	default:
		return fmt.Errorf("unsupported format: %s (supported: yaml, json)", c.config.Format)
	}
}

// writeJSON writes the combined spec as JSON
func (c *Combiner) writeJSON() error {
	file, err := os.Create(c.config.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c.combined)
}

// writeYAML writes the combined spec as YAML
func (c *Combiner) writeYAML() error {
	file, err := os.Create(c.config.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(c.combined)
}

// GetStats returns combination statistics
func (c *Combiner) GetStats() CombinationStats {
	return c.stats
}
