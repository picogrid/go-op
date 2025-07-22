package generator

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	goop "github.com/picogrid/go-op"
	"github.com/picogrid/go-op/operations"
)

// Generator handles OpenAPI specification generation from Go source code
type Generator struct {
	config     *Config
	fileSet    *token.FileSet
	operations []OperationDefinition
	schemas    map[string]*SchemaDefinition
	spec       *operations.OpenAPISpec
	stats      GenerationStats
}

// OperationDefinition represents a discovered operation in source code
type OperationDefinition struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Params      *SchemaDefinition
	Query       *SchemaDefinition
	Body        *SchemaDefinition
	Response    *SchemaDefinition          // Deprecated: use Responses instead
	Responses   map[int]ResponseDefinition // Multiple responses with status codes
	SourceFile  string
	LineNumber  int
}

// ResponseDefinition represents a response with schema and description
type ResponseDefinition struct {
	Schema      *SchemaDefinition
	Description string
}

// SchemaDefinition represents a discovered schema definition
type SchemaDefinition struct {
	Type          string
	Properties    map[string]*SchemaDefinition
	Items         *SchemaDefinition
	Required      []string
	MinLength     *int
	MaxLength     *int
	Minimum       *float64
	Maximum       *float64
	Pattern       string
	Format        string
	Default       interface{}
	Description   string
	Example       interface{}
	Examples      map[string]ExampleObject
	ExternalValue string

	// OpenAPI 3.1 / JSON Schema 2020-12 fields
	Const            interface{} `json:"const,omitempty" yaml:"const,omitempty"`
	MultipleOf       *float64    `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	ExclusiveMinimum *float64    `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64    `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	UniqueItems      *bool       `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MinProperties    *int        `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	MaxProperties    *int        `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`

	// Schema composition fields for OpenAPI 3.1
	OneOf []*SchemaDefinition
	AllOf []*SchemaDefinition
	AnyOf []*SchemaDefinition
	Not   *SchemaDefinition
}

// ExampleObject represents an OpenAPI example object
type ExampleObject struct {
	Summary       string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string      `json:"description,omitempty" yaml:"description,omitempty"`
	Value         interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// New creates a new OpenAPI generator
func New(config *Config) *Generator {
	return &Generator{
		config:     config,
		fileSet:    token.NewFileSet(),
		operations: make([]OperationDefinition, 0),
		schemas:    make(map[string]*SchemaDefinition),
		stats:      GenerationStats{},
	}
}

// ScanOperations scans the input directory for go-op operations
func (g *Generator) ScanOperations() error {
	if g.config.Verbose {
		fmt.Printf("[VERBOSE] Scanning directory: %s\n", g.config.InputDir)
	}

	return filepath.Walk(g.config.InputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files for now
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip vendor directories
		if strings.Contains(path, "/vendor/") {
			return nil
		}

		if g.config.Verbose {
			fmt.Printf("[VERBOSE] Scanning file: %s\n", path)
		}

		return g.scanFile(path)
	})
}

// scanFile scans a single Go file for operations
func (g *Generator) scanFile(filename string) error {
	g.stats.FileCount++

	// Clean and validate the filename to prevent path traversal attacks
	filename = filepath.Clean(filename)
	if !filepath.IsAbs(filename) {
		return fmt.Errorf("filename must be an absolute path")
	}

	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Parse the Go source file
	file, err := parser.ParseFile(g.fileSet, filename, src, parser.ParseComments)
	if err != nil {
		if g.config.Verbose {
			fmt.Printf("[VERBOSE] Warning: failed to parse %s: %v\n", filename, err)
		}
		return nil // Skip files that can't be parsed
	}

	// Use sophisticated AST analyzer to extract operations
	analyzer := NewASTAnalyzer(g.fileSet, g.config.Verbose)
	operations := analyzer.ExtractOperations(file, filename)

	// Add discovered operations to the generator
	for _, op := range operations {
		g.operations = append(g.operations, op)
		g.stats.OperationCount++
		if g.config.Verbose {
			fmt.Printf("[VERBOSE] Found operation: %s %s\n", op.Method, op.Path)
		}
	}

	return nil
}

// GenerateSpec generates the OpenAPI specification from discovered operations
func (g *Generator) GenerateSpec() error {
	if g.config.Verbose {
		fmt.Printf("[VERBOSE] Generating OpenAPI spec with %d operations\n", len(g.operations))
	}

	// Create the base OpenAPI spec
	g.spec = &operations.OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: operations.OpenAPIInfo{
			Title:       g.getTitle(),
			Version:     g.config.Version,
			Description: g.config.Description,
		},
		Paths: make(map[string]map[string]operations.OpenAPIOperation),
	}

	// Add servers if specified
	if len(g.config.Servers) > 0 {
		g.spec.Servers = make([]operations.OpenAPIServer, len(g.config.Servers))
		for i, server := range g.config.Servers {
			g.spec.Servers[i] = operations.OpenAPIServer{
				URL: server,
			}
		}
	}

	// Convert operations to OpenAPI format
	for _, op := range g.operations {
		g.addOperationToSpec(op)
	}

	g.stats.PathCount = len(g.spec.Paths)

	return nil
}

// getTitle determines the API title
func (g *Generator) getTitle() string {
	if g.config.Title != "" {
		return g.config.Title
	}

	// Try to auto-detect from directory name
	dirName := filepath.Base(g.config.InputDir)
	if dirName != "." && dirName != "/" {
		// Replace strings.Title with manual title case conversion
		parts := strings.Split(strings.ReplaceAll(dirName, "-", " "), " ")
		for i, part := range parts {
			if len(part) > 0 {
				parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			}
		}
		return strings.Join(parts, " ") + " API"
	}

	return "Generated API"
}

// addOperationToSpec adds an operation to the OpenAPI spec
func (g *Generator) addOperationToSpec(op OperationDefinition) {
	// Initialize path if it doesn't exist
	if g.spec.Paths[op.Path] == nil {
		g.spec.Paths[op.Path] = make(map[string]operations.OpenAPIOperation)
	}

	// Create OpenAPI operation
	openAPIOp := operations.OpenAPIOperation{
		Summary:     op.Summary,
		Description: op.Description,
		Tags:        op.Tags,
		Parameters:  []operations.OpenAPIParameter{},
		Responses:   make(map[string]operations.OpenAPIResponse),
	}

	// Add parameters from path params
	if op.Params != nil {
		g.addParametersFromSchema(op.Params, "path", &openAPIOp)
	}

	// Add parameters from query params
	if op.Query != nil {
		g.addParametersFromSchema(op.Query, "query", &openAPIOp)
	}

	// Add request body if specified
	if op.Body != nil {
		openAPIOp.RequestBody = g.convertSchemaToRequestBody(op.Body)
	}

	// Add responses - prefer multiple responses if available
	switch {
	case len(op.Responses) > 0:
		// Use new multiple responses
		for code, respDef := range op.Responses {
			codeStr := fmt.Sprintf("%d", code)
			response := operations.OpenAPIResponse{
				Description: respDef.Description,
			}

			if respDef.Schema != nil {
				response.Content = map[string]operations.OpenAPIMediaType{
					"application/json": {
						Schema: g.convertSchemaToOpenAPI(respDef.Schema),
					},
				}
			}

			openAPIOp.Responses[codeStr] = response
		}
	case op.Response != nil:
		// Fallback to legacy single response
		openAPIOp.Responses["200"] = operations.OpenAPIResponse{
			Description: "Successful response",
			Content: map[string]operations.OpenAPIMediaType{
				"application/json": {
					Schema: g.convertSchemaToOpenAPI(op.Response),
				},
			},
		}
	default:
		// Add default success response
		openAPIOp.Responses["200"] = operations.OpenAPIResponse{
			Description: "Successful response",
		}
	}

	// Add the operation to the spec
	g.spec.Paths[op.Path][strings.ToLower(op.Method)] = openAPIOp
}

// addParametersFromSchema adds parameters to an operation from a schema
func (g *Generator) addParametersFromSchema(schema *SchemaDefinition, paramType string, openAPIOp *operations.OpenAPIOperation) {
	if schema.Type == "object" && schema.Properties != nil {
		for name, propSchema := range schema.Properties {
			param := operations.OpenAPIParameter{
				Name:     name,
				In:       paramType,
				Required: paramType == "path" || g.isPropertyRequired(name, schema.Required),
				Schema:   g.convertSchemaToOpenAPI(propSchema),
			}
			openAPIOp.Parameters = append(openAPIOp.Parameters, param)
		}
	}
}

// convertSchemaToRequestBody converts a schema to a request body
func (g *Generator) convertSchemaToRequestBody(schema *SchemaDefinition) *operations.OpenAPIRequestBody {
	return &operations.OpenAPIRequestBody{
		Required: true,
		Content: map[string]operations.OpenAPIMediaType{
			"application/json": {
				Schema: g.convertSchemaToOpenAPI(schema),
			},
		},
	}
}

// convertSchemaToOpenAPI converts internal schema to go-op OpenAPI schema
func (g *Generator) convertSchemaToOpenAPI(schema *SchemaDefinition) *goop.OpenAPISchema {
	openAPISchema := &goop.OpenAPISchema{
		Type:        schema.Type,
		Description: schema.Description,
		Format:      schema.Format,
		Pattern:     schema.Pattern,
		Default:     schema.Default,
		Example:     schema.Example,
	}

	// Add constraints
	if schema.MinLength != nil {
		openAPISchema.MinLength = schema.MinLength
	}
	if schema.MaxLength != nil {
		openAPISchema.MaxLength = schema.MaxLength
	}
	if schema.Minimum != nil {
		openAPISchema.Minimum = schema.Minimum
	}
	if schema.Maximum != nil {
		openAPISchema.Maximum = schema.Maximum
	}

	// Add OpenAPI 3.1 / JSON Schema 2020-12 constraints
	if schema.Const != nil {
		openAPISchema.Const = schema.Const
	}
	if schema.MultipleOf != nil {
		openAPISchema.MultipleOf = schema.MultipleOf
	}
	if schema.ExclusiveMinimum != nil {
		openAPISchema.ExclusiveMinimum = schema.ExclusiveMinimum
	}
	if schema.ExclusiveMaximum != nil {
		openAPISchema.ExclusiveMaximum = schema.ExclusiveMaximum
	}
	if schema.UniqueItems != nil {
		openAPISchema.UniqueItems = schema.UniqueItems
	}
	if schema.MinProperties != nil {
		openAPISchema.MinProperties = schema.MinProperties
	}
	if schema.MaxProperties != nil {
		openAPISchema.MaxProperties = schema.MaxProperties
	}

	// Handle object properties
	if schema.Type == "object" && schema.Properties != nil {
		openAPISchema.Properties = make(map[string]*goop.OpenAPISchema)
		for name, propSchema := range schema.Properties {
			openAPISchema.Properties[name] = g.convertSchemaToOpenAPI(propSchema)
		}
		if len(schema.Required) > 0 {
			openAPISchema.Required = schema.Required
		}
	}

	// Handle array items
	if schema.Type == "array" && schema.Items != nil {
		openAPISchema.Items = g.convertSchemaToOpenAPI(schema.Items)
	}

	// Handle schema composition (OpenAPI 3.1)
	if len(schema.OneOf) > 0 {
		openAPISchema.OneOf = make([]*goop.OpenAPISchema, len(schema.OneOf))
		for i, childSchema := range schema.OneOf {
			openAPISchema.OneOf[i] = g.convertSchemaToOpenAPI(childSchema)
		}
	}
	if len(schema.AllOf) > 0 {
		openAPISchema.AllOf = make([]*goop.OpenAPISchema, len(schema.AllOf))
		for i, childSchema := range schema.AllOf {
			openAPISchema.AllOf[i] = g.convertSchemaToOpenAPI(childSchema)
		}
	}
	if len(schema.AnyOf) > 0 {
		openAPISchema.AnyOf = make([]*goop.OpenAPISchema, len(schema.AnyOf))
		for i, childSchema := range schema.AnyOf {
			openAPISchema.AnyOf[i] = g.convertSchemaToOpenAPI(childSchema)
		}
	}
	if schema.Not != nil {
		openAPISchema.Not = g.convertSchemaToOpenAPI(schema.Not)
	}

	return openAPISchema
}

// isPropertyRequired checks if a property is in the required list
func (g *Generator) isPropertyRequired(propName string, required []string) bool {
	for _, req := range required {
		if req == propName {
			return true
		}
	}
	return false
}

// WriteSpec writes the OpenAPI specification to the output file
func (g *Generator) WriteSpec() error {
	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(g.config.OutputFile)
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the spec in the specified format
	switch strings.ToLower(g.config.Format) {
	case "json":
		return g.writeJSON()
	case "yaml", "yml":
		return g.writeYAML()
	default:
		return fmt.Errorf("unsupported format: %s (supported: yaml, json)", g.config.Format)
	}
}

// writeJSON writes the spec as JSON
func (g *Generator) writeJSON() error {
	file, err := os.Create(g.config.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(g.spec)
}

// writeYAML writes the spec as YAML
func (g *Generator) writeYAML() error {
	file, err := os.Create(g.config.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(g.spec)
}

// GetStats returns generation statistics
func (g *Generator) GetStats() GenerationStats {
	return g.stats
}
