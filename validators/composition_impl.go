package validators

import (
	"fmt"
	"reflect"

	goop "github.com/picogrid/go-op"
)

// compositionSchema implements schema composition (OneOf, AllOf, AnyOf, Not)
type compositionSchema struct {
	compositionType CompositionType
	schemas         []interface{}
	defaultValue    interface{}
	hasDefault      bool
	description     string
}

// OneOf creates a schema that validates against exactly one of the provided schemas
func OneOf(schemas ...interface{}) CompositionBuilder {
	return &compositionSchema{
		compositionType: CompositionTypeOneOf,
		schemas:         schemas,
	}
}

// AllOf creates a schema that validates against all of the provided schemas
func AllOf(schemas ...interface{}) CompositionBuilder {
	return &compositionSchema{
		compositionType: CompositionTypeAllOf,
		schemas:         schemas,
	}
}

// AnyOf creates a schema that validates against one or more of the provided schemas
func AnyOf(schemas ...interface{}) CompositionBuilder {
	return &compositionSchema{
		compositionType: CompositionTypeAnyOf,
		schemas:         schemas,
	}
}

// Not creates a schema that validates against the inverse of the provided schema
func Not(schema interface{}) CompositionBuilder {
	return &compositionSchema{
		compositionType: CompositionTypeNot,
		schemas:         []interface{}{schema},
	}
}

// Required makes the composition schema required
func (c *compositionSchema) Required() RequiredCompositionBuilder {
	return &requiredCompositionSchema{compositionSchema: *c}
}

// Optional makes the composition schema optional
func (c *compositionSchema) Optional() OptionalCompositionBuilder {
	return &optionalCompositionSchema{compositionSchema: *c}
}

// requiredCompositionSchema represents a required composition schema
type requiredCompositionSchema struct {
	compositionSchema
}

// optionalCompositionSchema represents an optional composition schema
type optionalCompositionSchema struct {
	compositionSchema
}

// Default sets a default value for the optional composition schema
func (o *optionalCompositionSchema) Default(value interface{}) OptionalCompositionBuilder {
	o.defaultValue = value
	o.hasDefault = true
	return o
}

// Validate validates the data against the composition schema
func (c *compositionSchema) Validate(data interface{}) error {
	switch c.compositionType {
	case CompositionTypeOneOf:
		return c.validateOneOf(data)
	case CompositionTypeAllOf:
		return c.validateAllOf(data)
	case CompositionTypeAnyOf:
		return c.validateAnyOf(data)
	case CompositionTypeNot:
		return c.validateNot(data)
	default:
		return goop.NewValidationError("composition", data, fmt.Sprintf("unknown composition type: %s", c.compositionType))
	}
}

// validateOneOf ensures exactly one schema matches
func (c *compositionSchema) validateOneOf(data interface{}) error {
	var matchCount int

	for i, schema := range c.schemas {
		if validator, ok := schema.(goop.Schema); ok {
			if err := validator.Validate(data); err == nil {
				matchCount++
			}
		} else {
			return goop.NewValidationError("oneOf", data, fmt.Sprintf("schema at index %d does not implement Schema interface", i))
		}
	}

	if matchCount == 0 {
		return goop.NewValidationError("oneOf", data, "data does not match any schema")
	}

	if matchCount > 1 {
		return goop.NewValidationError("oneOf", data, fmt.Sprintf("data matches %d schemas, expected exactly 1", matchCount))
	}

	return nil
}

// validateAllOf ensures all schemas match
func (c *compositionSchema) validateAllOf(data interface{}) error {
	var errors []error

	for i, schema := range c.schemas {
		if validator, ok := schema.(goop.Schema); ok {
			if err := validator.Validate(data); err != nil {
				errors = append(errors, err)
			}
		} else {
			return goop.NewValidationError("allOf", data, fmt.Sprintf("schema at index %d does not implement Schema interface", i))
		}
	}

	if len(errors) > 0 {
		return goop.NewValidationError("allOf", data, fmt.Sprintf("data does not match all schemas (%d failures)", len(errors)))
	}

	return nil
}

// validateAnyOf ensures at least one schema matches
func (c *compositionSchema) validateAnyOf(data interface{}) error {
	for i, schema := range c.schemas {
		if validator, ok := schema.(goop.Schema); ok {
			if err := validator.Validate(data); err == nil {
				return nil // At least one schema matches
			}
		} else {
			return goop.NewValidationError("anyOf", data, fmt.Sprintf("schema at index %d does not implement Schema interface", i))
		}
	}

	return goop.NewValidationError("anyOf", data, "data does not match any schema")
}

// validateNot ensures the schema does not match
func (c *compositionSchema) validateNot(data interface{}) error {
	if len(c.schemas) != 1 {
		return goop.NewValidationError("not", data, "not schema must have exactly one schema")
	}

	if validator, ok := c.schemas[0].(goop.Schema); ok {
		if err := validator.Validate(data); err == nil {
			return goop.NewValidationError("not", data, "data matches the not schema (should not match)")
		}
		return nil // Schema doesn't match, which is what we want for "not"
	}

	return goop.NewValidationError("not", data, "schema does not implement Schema interface")
}

// ToOpenAPISchema generates OpenAPI schema for composition
func (c *compositionSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	schema := &goop.OpenAPISchema{
		Description: c.description,
	}

	// Convert child schemas to OpenAPI schemas
	var childSchemas []*goop.OpenAPISchema
	for _, childSchema := range c.schemas {
		if enhancedSchema, ok := childSchema.(goop.EnhancedSchema); ok {
			childSchemas = append(childSchemas, enhancedSchema.ToOpenAPISchema())
		}
	}

	// Set the appropriate composition field
	switch c.compositionType {
	case CompositionTypeOneOf:
		schema.OneOf = childSchemas
	case CompositionTypeAllOf:
		schema.AllOf = childSchemas
	case CompositionTypeAnyOf:
		schema.AnyOf = childSchemas
	case CompositionTypeNot:
		if len(childSchemas) > 0 {
			schema.Not = childSchemas[0]
		}
	}

	// Add default value if present
	if c.hasDefault {
		schema.Default = c.defaultValue
	}

	return schema
}

// Implement the Required and Optional interfaces for composition schemas
func (r *requiredCompositionSchema) Required() RequiredCompositionBuilder {
	return r
}

func (r *requiredCompositionSchema) Optional() OptionalCompositionBuilder {
	return &optionalCompositionSchema{compositionSchema: r.compositionSchema}
}

func (o *optionalCompositionSchema) Required() RequiredCompositionBuilder {
	return &requiredCompositionSchema{compositionSchema: o.compositionSchema}
}

func (o *optionalCompositionSchema) Optional() OptionalCompositionBuilder {
	return o
}

// Helper method to check if data is nil or empty for optional schemas
func (c *compositionSchema) isNilOrEmpty(data interface{}) bool {
	if data == nil {
		return true
	}

	value := reflect.ValueOf(data)
	switch value.Kind() {
	case reflect.String:
		return value.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return value.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return value.IsNil()
	default:
		return false
	}
}

// Override Validate for optional schemas to handle nil/empty values
func (o *optionalCompositionSchema) Validate(data interface{}) error {
	// For optional schemas, nil/empty values are valid if there's a default
	if o.isNilOrEmpty(data) {
		if o.hasDefault {
			return o.compositionSchema.Validate(o.defaultValue)
		}
		return nil // Optional field with no default, nil is valid
	}

	return o.compositionSchema.Validate(data)
}

// GetValidationInfo returns metadata about the composition validation configuration
func (c *compositionSchema) GetValidationInfo() *goop.ValidationInfo {
	constraints := make(map[string]interface{})
	constraints["compositionType"] = string(c.compositionType)
	constraints["schemaCount"] = len(c.schemas)

	return &goop.ValidationInfo{
		Required:     false, // This will be set by Required/Optional wrappers
		Optional:     true,
		HasDefault:   c.hasDefault,
		DefaultValue: c.defaultValue,
		Constraints:  constraints,
	}
}

func (r *requiredCompositionSchema) GetValidationInfo() *goop.ValidationInfo {
	info := r.compositionSchema.GetValidationInfo()
	info.Required = true
	info.Optional = false
	return info
}

func (o *optionalCompositionSchema) GetValidationInfo() *goop.ValidationInfo {
	info := o.compositionSchema.GetValidationInfo()
	info.Required = false
	info.Optional = true
	return info
}
