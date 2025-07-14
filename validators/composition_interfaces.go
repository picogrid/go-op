package validators

import goop "github.com/picogrid/go-op"

// CompositionBuilder provides the base interface for schema composition validators
type CompositionBuilder interface {
	Required() RequiredCompositionBuilder
	Optional() OptionalCompositionBuilder
}

// RequiredCompositionBuilder represents a required composition schema
type RequiredCompositionBuilder interface {
	CompositionBuilder
	goop.EnhancedSchema
	goop.Schema
}

// OptionalCompositionBuilder represents an optional composition schema
type OptionalCompositionBuilder interface {
	CompositionBuilder
	goop.EnhancedSchema
	goop.Schema
	Default(value interface{}) OptionalCompositionBuilder
}

// CompositionType represents the type of schema composition
type CompositionType string

const (
	CompositionTypeOneOf CompositionType = "oneOf"
	CompositionTypeAllOf CompositionType = "allOf"
	CompositionTypeAnyOf CompositionType = "anyOf"
	CompositionTypeNot   CompositionType = "not"
)
