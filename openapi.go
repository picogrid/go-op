package goop

// OpenAPISchema represents the structure of an OpenAPI 3.1 schema
// This is generated at build time, not runtime, for zero performance overhead
type OpenAPISchema struct {
	Type        string                     `json:"type,omitempty" yaml:"type,omitempty"`
	Format      string                     `json:"format,omitempty" yaml:"format,omitempty"`
	Properties  map[string]*OpenAPISchema  `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items       *OpenAPISchema             `json:"items,omitempty" yaml:"items,omitempty"`
	Required    []string                   `json:"required,omitempty" yaml:"required,omitempty"`
	MinLength   *int                       `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength   *int                       `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Minimum     *float64                   `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum     *float64                   `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	Pattern     string                     `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Enum        []interface{}              `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     interface{}                `json:"default,omitempty" yaml:"default,omitempty"`
	Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Example     interface{}                `json:"example,omitempty" yaml:"example,omitempty"`
}

// ValidationInfo contains metadata about validation rules
// Used by build-time generators to understand schema constraints
type ValidationInfo struct {
	Required     bool
	Optional     bool
	HasDefault   bool
	DefaultValue interface{}
	Constraints  map[string]interface{}
}

// OpenAPIGenerator interface for schemas that can generate OpenAPI specifications
// This is only called at build time, never at runtime
type OpenAPIGenerator interface {
	ToOpenAPISchema() *OpenAPISchema
	GetValidationInfo() *ValidationInfo
}

// Enhanced Schema interface that adds OpenAPI generation capabilities
// Maintains backward compatibility while adding build-time spec generation
type EnhancedSchema interface {
	Schema
	OpenAPIGenerator
}