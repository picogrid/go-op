package operations

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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
	URL         string                           `json:"url" yaml:"url"`
	Description string                           `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]OpenAPIServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// OpenAPIServerVariable represents a server variable in OpenAPI spec
type OpenAPIServerVariable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

// OpenAPISpec represents the complete OpenAPI 3.1 specification
type OpenAPISpec struct {
	OpenAPI           string                                 `json:"openapi" yaml:"openapi"`
	Info              OpenAPIInfo                            `json:"info" yaml:"info"`
	Servers           []OpenAPIServer                        `json:"servers,omitempty" yaml:"servers,omitempty"`
	Security          []goop.SecurityRequirement             `json:"security,omitempty" yaml:"security,omitempty"`
	Paths             map[string]map[string]OpenAPIOperation `json:"paths" yaml:"paths"`
	Components        *OpenAPIComponents                     `json:"components,omitempty" yaml:"components,omitempty"`
	Tags              []OpenAPITag                           `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs      *OpenAPIExternalDocs                   `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Webhooks          map[string]OpenAPIWebhook              `json:"webhooks,omitempty" yaml:"webhooks,omitempty"`
	JsonSchemaDialect string                                 `json:"jsonSchemaDialect,omitempty" yaml:"jsonSchemaDialect,omitempty"`
}

// OpenAPITag represents a tag in OpenAPI spec
type OpenAPITag struct {
	Name         string               `json:"name" yaml:"name"`
	Description  string               `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *OpenAPIExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// OpenAPIWebhook represents a webhook in OpenAPI spec
type OpenAPIWebhook struct {
	Operations map[string]OpenAPIOperation `json:"-" yaml:"-"`
}

// OpenAPIInfo represents the info section of OpenAPI spec
type OpenAPIInfo struct {
	Title          string          `json:"title" yaml:"title"`
	Version        string          `json:"version" yaml:"version"`
	Description    string          `json:"description,omitempty" yaml:"description,omitempty"`
	Summary        string          `json:"summary,omitempty" yaml:"summary,omitempty"`
	TermsOfService string          `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *OpenAPIContact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *OpenAPILicense `json:"license,omitempty" yaml:"license,omitempty"`
}

// OpenAPIContact represents contact information for the API
type OpenAPIContact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// OpenAPILicense represents license information for the API
type OpenAPILicense struct {
	Name       string `json:"name" yaml:"name"`
	Identifier string `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	URL        string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Validate validates the OpenAPI license object according to OpenAPI 3.1 rules
func (l *OpenAPILicense) Validate() error {
	if l.Name == "" {
		return fmt.Errorf("license name is required")
	}

	// identifier and url are mutually exclusive
	if l.Identifier != "" && l.URL != "" {
		return fmt.Errorf("license identifier and url are mutually exclusive")
	}

	return nil
}

// OpenAPIOperation represents a single operation in OpenAPI spec
type OpenAPIOperation struct {
	Summary      string                     `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Tags         []string                   `json:"tags,omitempty" yaml:"tags,omitempty"`
	Parameters   []OpenAPIParameter         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  *OpenAPIRequestBody        `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses    map[string]OpenAPIResponse `json:"responses" yaml:"responses"`
	Callbacks    map[string]OpenAPICallback `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	Security     []goop.SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
	Servers      []OpenAPIServer            `json:"servers,omitempty" yaml:"servers,omitempty"`
	OperationId  string                     `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Deprecated   *bool                      `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	ExternalDocs *OpenAPIExternalDocs       `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// OpenAPIExternalDocs represents external documentation for the API
type OpenAPIExternalDocs struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url" yaml:"url"`
}

// OpenAPIParameter represents a parameter in OpenAPI spec
type OpenAPIParameter struct {
	Name            string                      `json:"name" yaml:"name"`
	In              string                      `json:"in" yaml:"in"` // "path", "query", "header", "cookie"
	Description     string                      `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                        `json:"required" yaml:"required"`
	Schema          *goop.OpenAPISchema         `json:"schema,omitempty" yaml:"schema,omitempty"`
	Deprecated      *bool                       `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Style           string                      `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         *bool                       `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowEmptyValue *bool                       `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	AllowReserved   *bool                       `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Example         interface{}                 `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]OpenAPIExample   `json:"examples,omitempty" yaml:"examples,omitempty"`
	Content         map[string]OpenAPIMediaType `json:"content,omitempty" yaml:"content,omitempty"`
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
	Headers     map[string]OpenAPIHeader    `json:"headers,omitempty" yaml:"headers,omitempty"`
	Links       map[string]OpenAPILink      `json:"links,omitempty" yaml:"links,omitempty"`
}

// OpenAPILink represents a link in OpenAPI spec
type OpenAPILink struct {
	OperationRef string                 `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
	OperationId  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  interface{}            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Server       *OpenAPIServer         `json:"server,omitempty" yaml:"server,omitempty"`
}

// OpenAPIMediaType represents a media type in OpenAPI spec
type OpenAPIMediaType struct {
	Schema   *goop.OpenAPISchema        `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example  interface{}                `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]OpenAPIExample  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding map[string]OpenAPIEncoding `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

// Validate validates the OpenAPI media type object according to OpenAPI 3.1 rules
func (m *OpenAPIMediaType) Validate() error {
	// example and examples are mutually exclusive
	if m.Example != nil && len(m.Examples) > 0 {
		return fmt.Errorf("example and examples are mutually exclusive")
	}
	return nil
}

// OpenAPIComponents represents the components section of OpenAPI spec
type OpenAPIComponents struct {
	Schemas         map[string]*goop.OpenAPISchema       `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	SecuritySchemes map[string]goop.SecuritySchemeObject `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Responses       map[string]OpenAPIResponse           `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters      map[string]OpenAPIParameter          `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Examples        map[string]OpenAPIExample            `json:"examples,omitempty" yaml:"examples,omitempty"`
	RequestBodies   map[string]OpenAPIRequestBody        `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Headers         map[string]OpenAPIHeader             `json:"headers,omitempty" yaml:"headers,omitempty"`
	Links           map[string]OpenAPILink               `json:"links,omitempty" yaml:"links,omitempty"`
	Callbacks       map[string]OpenAPICallback           `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	PathItems       map[string]OpenAPIPathItem           `json:"pathItems,omitempty" yaml:"pathItems,omitempty"`
}

// OpenAPIExample represents an example in OpenAPI spec
type OpenAPIExample struct {
	Summary       string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string      `json:"description,omitempty" yaml:"description,omitempty"`
	Value         interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// OpenAPIEncoding represents encoding information for OpenAPI spec
type OpenAPIEncoding struct {
	ContentType   string                   `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]OpenAPIHeader `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string                   `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       *bool                    `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved *bool                    `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
}

// OpenAPIHeader represents a header in OpenAPI spec
type OpenAPIHeader struct {
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Required    *bool                     `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated  *bool                     `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Schema      *goop.OpenAPISchema       `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example     interface{}               `json:"example,omitempty" yaml:"example,omitempty"`
	Examples    map[string]OpenAPIExample `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// OpenAPICallback represents a callback in OpenAPI spec
type OpenAPICallback map[string]OpenAPIPathItem

// OpenAPIPathItem represents a path item in OpenAPI spec
type OpenAPIPathItem struct {
	Ref         string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string             `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *OpenAPIOperation  `json:"get,omitempty" yaml:"get,omitempty"`
	Put         *OpenAPIOperation  `json:"put,omitempty" yaml:"put,omitempty"`
	Post        *OpenAPIOperation  `json:"post,omitempty" yaml:"post,omitempty"`
	Delete      *OpenAPIOperation  `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options     *OpenAPIOperation  `json:"options,omitempty" yaml:"options,omitempty"`
	Head        *OpenAPIOperation  `json:"head,omitempty" yaml:"head,omitempty"`
	Patch       *OpenAPIOperation  `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace       *OpenAPIOperation  `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers     []OpenAPIServer    `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters  []OpenAPIParameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
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
				Responses:       make(map[string]OpenAPIResponse),
				Parameters:      make(map[string]OpenAPIParameter),
				Examples:        make(map[string]OpenAPIExample),
				RequestBodies:   make(map[string]OpenAPIRequestBody),
				Headers:         make(map[string]OpenAPIHeader),
				Links:           make(map[string]OpenAPILink),
				Callbacks:       make(map[string]OpenAPICallback),
				PathItems:       make(map[string]OpenAPIPathItem),
			},
		},
	}
}

// SetDescription sets the API description
func (g *OpenAPIGenerator) SetDescription(description string) {
	g.Description = description
	g.Spec.Info.Description = description
}

// SetSummary sets the API summary
func (g *OpenAPIGenerator) SetSummary(summary string) {
	g.Spec.Info.Summary = summary
}

// SetTermsOfService sets the API terms of service
func (g *OpenAPIGenerator) SetTermsOfService(termsOfService string) {
	g.Spec.Info.TermsOfService = termsOfService
}

// SetContact sets the API contact information
func (g *OpenAPIGenerator) SetContact(contact *OpenAPIContact) {
	g.Spec.Info.Contact = contact
}

// SetLicense sets the API license information
func (g *OpenAPIGenerator) SetLicense(license *OpenAPILicense) error {
	if license != nil {
		if err := license.Validate(); err != nil {
			return fmt.Errorf("invalid license: %v", err)
		}
	}
	g.Spec.Info.License = license
	return nil
}

// AddServer adds a server to the OpenAPI specification
func (g *OpenAPIGenerator) AddServer(server OpenAPIServer) {
	g.Servers = append(g.Servers, server)
	g.Spec.Servers = append(g.Spec.Servers, server)
}

// AddTag adds a tag to the OpenAPI specification
func (g *OpenAPIGenerator) AddTag(tag OpenAPITag) {
	g.Spec.Tags = append(g.Spec.Tags, tag)
}

// SetExternalDocs sets the external documentation for the API
func (g *OpenAPIGenerator) SetExternalDocs(externalDocs *OpenAPIExternalDocs) {
	g.Spec.ExternalDocs = externalDocs
}

// AddWebhook adds a webhook to the OpenAPI specification
func (g *OpenAPIGenerator) AddWebhook(name string, webhook OpenAPIWebhook) {
	if g.Spec.Webhooks == nil {
		g.Spec.Webhooks = make(map[string]OpenAPIWebhook)
	}
	g.Spec.Webhooks[name] = webhook
}

// SetJsonSchemaDialect sets the JSON Schema dialect for the API
func (g *OpenAPIGenerator) SetJsonSchemaDialect(dialect string) {
	g.Spec.JsonSchemaDialect = dialect
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
		mediaType := OpenAPIMediaType{
			Schema: info.Operation.BodySpec,
		}

		// Add example from schema if available
		if info.Operation.BodySpec.Example != nil {
			mediaType.Example = info.Operation.BodySpec.Example
		}

		operation.RequestBody = &OpenAPIRequestBody{
			Required: info.BodyInfo != nil && info.BodyInfo.Required,
			Content: map[string]OpenAPIMediaType{
				"application/json": mediaType,
			},
		}
	}

	// Add responses - use multiple responses if defined, otherwise use legacy single response
	if len(info.Operation.Responses) > 0 {
		// Use new multiple responses system
		for code, responseDef := range info.Operation.Responses {
			codeStr := fmt.Sprintf("%d", code)

			response := OpenAPIResponse{
				Description: responseDef.Description,
			}

			// Add schema if present
			if responseDef.Schema != nil {
				if enhanced, ok := responseDef.Schema.(goop.EnhancedSchema); ok {
					mediaType := OpenAPIMediaType{
						Schema: enhanced.ToOpenAPISchema(),
					}

					// Add example from schema if available
					if enhanced.ToOpenAPISchema().Example != nil {
						mediaType.Example = enhanced.ToOpenAPISchema().Example
					}

					response.Content = map[string]OpenAPIMediaType{
						"application/json": mediaType,
					}
				}
			}

			operation.Responses[codeStr] = response
		}
	} else {
		// Fallback to legacy single response for backward compatibility
		successCode := fmt.Sprintf("%d", info.Operation.SuccessCode)
		response := OpenAPIResponse{
			Description: "Successful response",
		}

		if info.Operation.ResponseSpec != nil {
			mediaType := OpenAPIMediaType{
				Schema: info.Operation.ResponseSpec,
			}

			// Add example from schema if available
			if info.Operation.ResponseSpec.Example != nil {
				mediaType.Example = info.Operation.ResponseSpec.Example
			}

			response.Content = map[string]OpenAPIMediaType{
				"application/json": mediaType,
			}
		}

		operation.Responses[successCode] = response

		// Add default error responses only if no custom responses are defined
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

// ValidateComponentKey validates that a component key follows OpenAPI 3.1 rules
func ValidateComponentKey(key string) error {
	// Component keys must match the regex: ^[a-zA-Z0-9\.\-_]+$
	matched, err := regexp.MatchString(`^[a-zA-Z0-9\.\-_]+$`, key)
	if err != nil {
		return fmt.Errorf("regex error: %v", err)
	}
	if !matched {
		return fmt.Errorf("component key '%s' must match pattern ^[a-zA-Z0-9\\.\\-_]+$", key)
	}
	return nil
}

// GetSpec returns the complete OpenAPI specification
func (g *OpenAPIGenerator) GetSpec() *OpenAPISpec {
	return g.Spec
}
