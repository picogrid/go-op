package goop

import (
	"encoding/json"
	"fmt"
)

// OpenAPISchema represents the structure of an OpenAPI 3.1 schema
// This is generated at build time, not runtime, for zero performance overhead
type OpenAPISchema struct {
	Type        string                    `json:"type,omitempty" yaml:"type,omitempty"`
	Format      string                    `json:"format,omitempty" yaml:"format,omitempty"`
	Properties  map[string]*OpenAPISchema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items       *OpenAPISchema            `json:"items,omitempty" yaml:"items,omitempty"`
	Required    []string                  `json:"required,omitempty" yaml:"required,omitempty"`
	MinLength   *int                      `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength   *int                      `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Minimum     *float64                  `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum     *float64                  `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	Pattern     string                    `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Enum        []interface{}             `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     interface{}               `json:"default,omitempty" yaml:"default,omitempty"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Example     interface{}               `json:"example,omitempty" yaml:"example,omitempty"`

	// OpenAPI 3.1 Fixed Fields - Numeric validation
	MultipleOf       *float64 `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`

	// OpenAPI 3.1 Fixed Fields - Array validation
	MaxItems    *int  `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems    *int  `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems *bool `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`

	// OpenAPI 3.1 Fixed Fields - Object validation
	MaxProperties        *int                 `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties        *int                 `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	AdditionalProperties *OpenAPISchemaOrBool `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`

	// OpenAPI 3.1 Fixed Fields - Schema composition
	AllOf []*OpenAPISchema `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf []*OpenAPISchema `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf []*OpenAPISchema `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Not   *OpenAPISchema   `json:"not,omitempty" yaml:"not,omitempty"`

	// OpenAPI 3.1 Fixed Fields - Metadata
	Title      string      `json:"title,omitempty" yaml:"title,omitempty"`
	Const      interface{} `json:"const,omitempty" yaml:"const,omitempty"`
	ReadOnly   *bool       `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly  *bool       `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	Deprecated *bool       `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

// OpenAPISchemaOrBool represents either a schema or a boolean value
// Used for additionalProperties field which can be either a schema or boolean
type OpenAPISchemaOrBool struct {
	Schema *OpenAPISchema `json:"-" yaml:"-"`
	Bool   *bool          `json:"-" yaml:"-"`
}

// MarshalJSON implements custom JSON marshaling for OpenAPISchemaOrBool
func (s OpenAPISchemaOrBool) MarshalJSON() ([]byte, error) {
	if s.Schema != nil {
		return json.Marshal(s.Schema)
	}
	if s.Bool != nil {
		return json.Marshal(*s.Bool)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements custom JSON unmarshaling for OpenAPISchemaOrBool
func (s *OpenAPISchemaOrBool) UnmarshalJSON(data []byte) error {
	var schema OpenAPISchema
	if err := json.Unmarshal(data, &schema); err == nil {
		s.Schema = &schema
		return nil
	}

	var boolVal bool
	if err := json.Unmarshal(data, &boolVal); err == nil {
		s.Bool = &boolVal
		return nil
	}

	return fmt.Errorf("additionalProperties must be either a schema or boolean")
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
