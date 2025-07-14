package validators

import (
	"encoding/json"
	"fmt"

	goop "github.com/picogrid/go-op"
)

// StructValidator creates a type-safe validator for struct type T.
// The schemaBuilder function receives a zero-value pointer of T and returns
// the validation schema as a map. This approach provides compile-time type safety
// without using runtime reflection.
//
// Example:
//
//	userSchema := StructValidator(func(u *User) map[string]interface{} {
//	    return map[string]interface{}{
//	        "email":    Email().Required(),
//	        "username": String().Min(3).Max(50).Required(),
//	    }
//	})
func StructValidator[T any](schemaBuilder func(*T) map[string]interface{}) goop.Schema {
	var zero T
	schemaMap := schemaBuilder(&zero)
	return Object(schemaMap).Required()
}

// ValidateStruct validates data against a schema and returns a typed result.
// This provides type-safe validation without reflection by using compile-time
// type assertions.
//
// Example:
//
//	user, err := ValidateStruct[User](userSchema, requestData)
//	if err != nil {
//	    // Handle validation error
//	}
//	// user is now *User type
func ValidateStruct[T any](schema goop.Schema, data interface{}) (*T, error) {
	// Convert struct to map for validation if needed
	var validateData interface{}
	switch v := data.(type) {
	case map[string]interface{}:
		// Already a map, use as-is
		validateData = v
	case *T:
		// Convert struct pointer to map via JSON (zero reflection approach)
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal struct: %w", err)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(jsonData, &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal struct: %w", err)
		}
		validateData = m
	case T:
		// Convert struct to map via JSON (zero reflection approach)
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal struct: %w", err)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(jsonData, &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
		}
		validateData = m
	default:
		// Try to convert via JSON for other types
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(jsonData, &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
		}
		validateData = m
	}

	// Validate the map data
	if err := schema.Validate(validateData); err != nil {
		return nil, err
	}

	// Convert back to struct type
	var result T
	jsonData, err := json.Marshal(validateData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal validated data: %w", err)
	}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to struct: %w", err)
	}

	return &result, nil
}

// StructSchemaBuilder provides a fluent interface for building struct validation schemas.
// This builder pattern offers better ergonomics while maintaining type safety.
type StructSchemaBuilder[T any] struct {
	fields      map[string]interface{}
	required    bool
	optional    bool
	strict      bool
	customError map[string]string
}

// ForStruct creates a new schema builder for struct type T.
//
// Example:
//
//	userSchema := ForStruct[User]().
//	    Field("email", Email().Required()).
//	    Field("username", String().Min(3).Max(50).Required()).
//	    Required()
func ForStruct[T any]() *StructSchemaBuilder[T] {
	return &StructSchemaBuilder[T]{
		fields:      make(map[string]interface{}),
		customError: make(map[string]string),
	}
}

// Field adds a field validator to the schema.
// The name should match the JSON tag of the struct field.
func (b *StructSchemaBuilder[T]) Field(name string, validator interface{}) *StructSchemaBuilder[T] {
	b.fields[name] = validator
	return b
}

// Fields adds multiple field validators at once.
func (b *StructSchemaBuilder[T]) Fields(fields map[string]interface{}) *StructSchemaBuilder[T] {
	for name, validator := range fields {
		b.fields[name] = validator
	}
	return b
}

// Required makes the entire struct required (cannot be nil).
func (b *StructSchemaBuilder[T]) Required() *StructSchemaBuilder[T] {
	b.required = true
	b.optional = false
	return b
}

// Optional makes the entire struct optional (can be nil).
func (b *StructSchemaBuilder[T]) Optional() *StructSchemaBuilder[T] {
	b.optional = true
	b.required = false
	return b
}

// Strict enables strict mode - unknown fields will cause validation errors.
func (b *StructSchemaBuilder[T]) Strict() *StructSchemaBuilder[T] {
	b.strict = true
	return b
}

// CustomError sets a custom error message for a specific validation error.
func (b *StructSchemaBuilder[T]) CustomError(key, message string) *StructSchemaBuilder[T] {
	b.customError[key] = message
	return b
}

// Build creates the final Schema from the builder configuration.
func (b *StructSchemaBuilder[T]) Build() goop.Schema {
	builder := Object(b.fields)

	// Apply modifiers
	if b.strict {
		builder = builder.Strict()
	}

	// Apply custom errors - WithMessage is the available method
	for key, msg := range b.customError {
		builder = builder.WithMessage(key, msg)
	}

	// Apply required/optional state
	if b.required {
		return builder.Required()
	} else if b.optional {
		return builder.Optional()
	}

	// Default to required if not specified
	return builder.Required()
}

// Schema is a convenience method that builds and returns the schema.
// It's equivalent to calling Build().
func (b *StructSchemaBuilder[T]) Schema() goop.Schema {
	return b.Build()
}

// TypedValidator creates a validator function that returns typed results.
// This is useful for creating reusable validation functions.
//
// Example:
//
//	validateUser := TypedValidator[User](userSchema)
//	user, err := validateUser(requestData)
func TypedValidator[T any](schema goop.Schema) func(interface{}) (*T, error) {
	return func(data interface{}) (*T, error) {
		return ValidateStruct[T](schema, data)
	}
}

// MustValidateStruct is like ValidateStruct but panics on validation error.
// This should only be used in contexts where validation errors are programming errors.
func MustValidateStruct[T any](schema goop.Schema, data interface{}) *T {
	result, err := ValidateStruct[T](schema, data)
	if err != nil {
		panic(fmt.Sprintf("validation failed: %v", err))
	}
	return result
}

// FieldMapping provides a type-safe way to map struct fields to validators.
// This is an experimental API that might help with future enhancements.
type FieldMapping[T any] struct {
	name      string
	validator interface{}
}

// MapField creates a field mapping with compile-time field name validation.
// Note: This is a placeholder for potential future enhancements where we might
// use code generation to validate field names at compile time.
func MapField[T any](name string, validator interface{}) FieldMapping[T] {
	return FieldMapping[T]{
		name:      name,
		validator: validator,
	}
}

// WithFields is an alternative way to build schemas using field mappings.
func WithFields[T any](mappings ...FieldMapping[T]) goop.Schema {
	fields := make(map[string]interface{}, len(mappings))
	for _, m := range mappings {
		fields[m.name] = m.validator
	}
	return Object(fields).Required()
}
