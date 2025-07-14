package operations

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	goop "github.com/picogrid/go-op"
)

// OpenAPIGenerator generates OpenAPI 3.1 specifications from operations
// This generator runs at build time to create API documentation
type OpenAPIGenerator struct {
	Title           string
	Version         string
	Description     string
	Servers         []OpenAPIServer
	SecuritySchemes map[string]goop.SecurityScheme
	GlobalSecurity  goop.SecurityRequirements
	Spec            *OpenAPISpec
}

// OpenAPIServer represents a server in the OpenAPI spec
type OpenAPIServer struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// OpenAPISpec represents the complete OpenAPI 3.1 specification
type OpenAPISpec struct {
	OpenAPI    string                                 `json:"openapi" yaml:"openapi"`
	Info       OpenAPIInfo                            `json:"info" yaml:"info"`
	Servers    []OpenAPIServer                        `json:"servers,omitempty" yaml:"servers,omitempty"`
	Security   []goop.SecurityRequirement             `json:"security,omitempty" yaml:"security,omitempty"`
	Paths      map[string]map[string]OpenAPIOperation `json:"paths" yaml:"paths"`
	Components *OpenAPIComponents                     `json:"components,omitempty" yaml:"components,omitempty"`
}

// OpenAPIInfo represents the info section of OpenAPI spec
type OpenAPIInfo struct {
	Title       string `json:"title" yaml:"title"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// OpenAPIOperation represents a single operation in OpenAPI spec
type OpenAPIOperation struct {
	Summary     string                     `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string                   `json:"tags,omitempty" yaml:"tags,omitempty"`
	Parameters  []OpenAPIParameter         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody        `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses" yaml:"responses"`
	Security    []goop.SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
}

// OpenAPIParameter represents a parameter in OpenAPI spec
type OpenAPIParameter struct {
	Name        string              `json:"name" yaml:"name"`
	In          string              `json:"in" yaml:"in"` // "path", "query", "header", "cookie"
	Description string              `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool                `json:"required" yaml:"required"`
	Schema      *goop.OpenAPISchema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// OpenAPIRequestBody represents a request body in OpenAPI spec
type OpenAPIRequestBody struct {
	Description string                      `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool                        `json:"required,omitempty" yaml:"required,omitempty"`
	Content     map[string]OpenAPIMediaType `json:"content" yaml:"content"`
}

// OpenAPIResponse represents a response in OpenAPI spec
type OpenAPIResponse struct {
	Description string                      `json:"description" yaml:"description"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// OpenAPIMediaType represents a media type in OpenAPI spec
type OpenAPIMediaType struct {
	Schema *goop.OpenAPISchema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// OpenAPIComponents represents the components section of OpenAPI spec
type OpenAPIComponents struct {
	Schemas         map[string]*goop.OpenAPISchema       `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	SecuritySchemes map[string]goop.SecuritySchemeObject `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
}

// NewOpenAPIGenerator creates a new OpenAPI generator
func NewOpenAPIGenerator(title, version string) *OpenAPIGenerator {
	return &OpenAPIGenerator{
		Title:           title,
		Version:         version,
		SecuritySchemes: make(map[string]goop.SecurityScheme),
		Spec: &OpenAPISpec{
			OpenAPI: "3.1.0",
			Info: OpenAPIInfo{
				Title:   title,
				Version: version,
			},
			Paths: make(map[string]map[string]OpenAPIOperation),
			Components: &OpenAPIComponents{
				Schemas:         make(map[string]*goop.OpenAPISchema),
				SecuritySchemes: make(map[string]goop.SecuritySchemeObject),
			},
		},
	}
}

// AddSecurityScheme adds a security scheme to the OpenAPI specification
func (g *OpenAPIGenerator) AddSecurityScheme(name string, scheme goop.SecurityScheme) error {
	// Validate the security scheme name
	if err := goop.ValidateSecuritySchemeName(name); err != nil {
		return fmt.Errorf("invalid security scheme name: %v", err)
	}

	// Validate the security scheme
	if err := scheme.Validate(); err != nil {
		return fmt.Errorf("invalid security scheme '%s': %v", name, err)
	}

	// Add to both the generator and the OpenAPI spec
	g.SecuritySchemes[name] = scheme
	g.Spec.Components.SecuritySchemes[name] = scheme.ToOpenAPI()

	return nil
}

// SetGlobalSecurity sets the global security requirements for the API
func (g *OpenAPIGenerator) SetGlobalSecurity(requirements goop.SecurityRequirements) {
	g.GlobalSecurity = requirements
	g.Spec.Security = []goop.SecurityRequirement(requirements)
}

// GetSecurityScheme retrieves a security scheme by name
func (g *OpenAPIGenerator) GetSecurityScheme(name string) (goop.SecurityScheme, bool) {
	scheme, exists := g.SecuritySchemes[name]
	return scheme, exists
}

// ListSecuritySchemes returns all registered security scheme names
func (g *OpenAPIGenerator) ListSecuritySchemes() []string {
	names := make([]string, 0, len(g.SecuritySchemes))
	for name := range g.SecuritySchemes {
		names = append(names, name)
	}
	return names
}

// Process processes an operation and adds it to the OpenAPI specification
func (g *OpenAPIGenerator) Process(info OperationInfo) error {
	// Create path if it doesn't exist
	if g.Spec.Paths[info.Path] == nil {
		g.Spec.Paths[info.Path] = make(map[string]OpenAPIOperation)
	}

	// Create the operation
	operation := OpenAPIOperation{
		Summary:     info.Summary,
		Description: info.Description,
		Tags:        info.Tags,
		Parameters:  []OpenAPIParameter{},
		Responses:   make(map[string]OpenAPIResponse),
		Security:    []goop.SecurityRequirement(info.Operation.Security),
	}

	// Add path parameters
	if info.Operation.ParamsSpec != nil {
		params := g.extractPathParameters(info.Path, info.Operation.ParamsSpec)
		operation.Parameters = append(operation.Parameters, params...)
	}

	// Add query parameters
	if info.Operation.QuerySpec != nil {
		queryParams := g.extractQueryParameters(info.Operation.QuerySpec)
		operation.Parameters = append(operation.Parameters, queryParams...)
	}

	// Add header parameters
	if info.Operation.HeaderSpec != nil {
		headerParams := g.extractHeaderParameters(info.Operation.HeaderSpec)
		operation.Parameters = append(operation.Parameters, headerParams...)
	}

	// Add request body
	if info.Operation.BodySpec != nil {
		operation.RequestBody = &OpenAPIRequestBody{
			Required: info.BodyInfo != nil && info.BodyInfo.Required,
			Content: map[string]OpenAPIMediaType{
				"application/json": {
					Schema: info.Operation.BodySpec,
				},
			},
		}
	}

	// Add response
	successCode := fmt.Sprintf("%d", info.Operation.SuccessCode)
	response := OpenAPIResponse{
		Description: "Successful response",
	}

	if info.Operation.ResponseSpec != nil {
		response.Content = map[string]OpenAPIMediaType{
			"application/json": {
				Schema: info.Operation.ResponseSpec,
			},
		}
	}

	operation.Responses[successCode] = response

	// Add common error responses
	operation.Responses["400"] = OpenAPIResponse{
		Description: "Bad Request",
		Content: map[string]OpenAPIMediaType{
			"application/json": {
				Schema: &goop.OpenAPISchema{
					Type: "object",
					Properties: map[string]*goop.OpenAPISchema{
						"error":   {Type: "string"},
						"details": {Type: "string"},
					},
					Required: []string{"error"},
				},
			},
		},
	}

	operation.Responses["500"] = OpenAPIResponse{
		Description: "Internal Server Error",
		Content: map[string]OpenAPIMediaType{
			"application/json": {
				Schema: &goop.OpenAPISchema{
					Type: "object",
					Properties: map[string]*goop.OpenAPISchema{
						"error":   {Type: "string"},
						"details": {Type: "string"},
					},
					Required: []string{"error"},
				},
			},
		},
	}

	// Store the operation
	g.Spec.Paths[info.Path][strings.ToLower(info.Method)] = operation

	return nil
}

// extractPathParameters extracts path parameters from the schema and path
func (g *OpenAPIGenerator) extractPathParameters(path string, schema *goop.OpenAPISchema) []OpenAPIParameter {
	var parameters []OpenAPIParameter

	if schema.Type == "object" && schema.Properties != nil {
		for paramName, paramSchema := range schema.Properties {
			// Check if this parameter is in the path
			if strings.Contains(path, "{"+paramName+"}") {
				parameter := OpenAPIParameter{
					Name:     paramName,
					In:       "path",
					Required: true, // Path parameters are always required
					Schema:   paramSchema,
				}
				parameters = append(parameters, parameter)
			}
		}
	}

	return parameters
}

// extractQueryParameters extracts query parameters from the schema
func (g *OpenAPIGenerator) extractQueryParameters(schema *goop.OpenAPISchema) []OpenAPIParameter {
	var parameters []OpenAPIParameter

	if schema.Type == "object" && schema.Properties != nil {
		for paramName, paramSchema := range schema.Properties {
			required := false
			for _, reqField := range schema.Required {
				if reqField == paramName {
					required = true
					break
				}
			}

			parameter := OpenAPIParameter{
				Name:     paramName,
				In:       "query",
				Required: required,
				Schema:   paramSchema,
			}
			parameters = append(parameters, parameter)
		}
	}

	return parameters
}

// extractHeaderParameters extracts header parameters from the schema
func (g *OpenAPIGenerator) extractHeaderParameters(schema *goop.OpenAPISchema) []OpenAPIParameter {
	var parameters []OpenAPIParameter

	if schema.Type == "object" && schema.Properties != nil {
		for paramName, paramSchema := range schema.Properties {
			required := false
			for _, reqField := range schema.Required {
				if reqField == paramName {
					required = true
					break
				}
			}

			parameter := OpenAPIParameter{
				Name:     paramName,
				In:       "header",
				Required: required,
				Schema:   paramSchema,
			}
			parameters = append(parameters, parameter)
		}
	}

	return parameters
}

// WriteToFile writes the OpenAPI specification to a file
func (g *OpenAPIGenerator) WriteToFile(filename string) error {
	// Clean and validate the filename to prevent path traversal attacks
	filename = filepath.Clean(filename)
	if !filepath.IsAbs(filename) {
		return fmt.Errorf("filename must be an absolute path")
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(g.Spec); err != nil {
		return fmt.Errorf("failed to encode OpenAPI spec: %w", err)
	}

	return nil
}

// WriteToWriter writes the OpenAPI specification to a writer
func (g *OpenAPIGenerator) WriteToWriter(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(g.Spec); err != nil {
		return fmt.Errorf("failed to encode OpenAPI spec: %w", err)
	}
	return nil
}

// GetSpec returns the complete OpenAPI specification
func (g *OpenAPIGenerator) GetSpec() *OpenAPISpec {
	return g.Spec
}
