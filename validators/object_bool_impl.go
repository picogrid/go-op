package validators

import (
	"fmt"
	"reflect"

	goop "github.com/picogrid/go-op"
)

// Core object schema struct (unexported)
type objectSchema struct {
	schema        map[string]interface{}
	strictMode    bool
	partialMode   bool
	minProperties int
	maxProperties int
	customFunc    func(map[string]interface{}) error
	required      bool
	optional      bool
	defaultValue  map[string]interface{}
	customError   map[string]string
	example       interface{}
	examples      map[string]ExampleObject
	externalValue string
}

// Core bool schema struct (unexported)
type boolSchema struct {
	customFunc    func(bool) error
	required      bool
	optional      bool
	defaultValue  *bool
	customError   map[string]string
	example       interface{}
	examples      map[string]ExampleObject
	externalValue string
}

// State wrapper types for objects
type requiredObjectSchema struct {
	*objectSchema
}

type optionalObjectSchema struct {
	*objectSchema
}

// State wrapper types for booleans
type requiredBoolSchema struct {
	*boolSchema
}

type optionalBoolSchema struct {
	*boolSchema
}

// ObjectBuilder implementation (initial state)
func (o *objectSchema) Strict() ObjectBuilder {
	o.strictMode = true
	return o
}

func (o *objectSchema) Partial() ObjectBuilder {
	o.partialMode = true
	return o
}

func (o *objectSchema) MinProperties(count int) ObjectBuilder {
	o.minProperties = count
	return o
}

func (o *objectSchema) MaxProperties(count int) ObjectBuilder {
	o.maxProperties = count
	return o
}

func (o *objectSchema) Custom(fn func(map[string]interface{}) error) ObjectBuilder {
	o.customFunc = fn
	return o
}

func (o *objectSchema) Required() RequiredObjectBuilder {
	o.required = true
	o.optional = false
	return &requiredObjectSchema{o}
}

func (o *objectSchema) Optional() OptionalObjectBuilder {
	o.optional = true
	o.required = false
	return &optionalObjectSchema{o}
}

func (o *objectSchema) WithMessage(validationType, message string) ObjectBuilder {
	if o.customError == nil {
		o.customError = make(map[string]string)
	}
	o.customError[validationType] = message
	return o
}

// RequiredObjectBuilder implementation
func (r *requiredObjectSchema) Strict() RequiredObjectBuilder {
	r.strictMode = true
	return r
}

func (r *requiredObjectSchema) Partial() RequiredObjectBuilder {
	r.partialMode = true
	return r
}

func (r *requiredObjectSchema) MinProperties(count int) RequiredObjectBuilder {
	r.minProperties = count
	return r
}

func (r *requiredObjectSchema) MaxProperties(count int) RequiredObjectBuilder {
	r.maxProperties = count
	return r
}

func (r *requiredObjectSchema) Custom(fn func(map[string]interface{}) error) RequiredObjectBuilder {
	r.customFunc = fn
	return r
}

func (r *requiredObjectSchema) WithMessage(validationType, message string) RequiredObjectBuilder {
	if r.customError == nil {
		r.customError = make(map[string]string)
	}
	r.customError[validationType] = message
	return r
}

func (r *requiredObjectSchema) WithRequiredMessage(message string) RequiredObjectBuilder {
	return r.WithMessage(errorKeys.Required, message)
}

func (r *requiredObjectSchema) Validate(data interface{}) error {
	return r.validate(data)
}

// OptionalObjectBuilder implementation
func (o *optionalObjectSchema) Strict() OptionalObjectBuilder {
	o.strictMode = true
	return o
}

func (o *optionalObjectSchema) Partial() OptionalObjectBuilder {
	o.partialMode = true
	return o
}

func (o *optionalObjectSchema) MinProperties(count int) OptionalObjectBuilder {
	o.minProperties = count
	return o
}

func (o *optionalObjectSchema) MaxProperties(count int) OptionalObjectBuilder {
	o.maxProperties = count
	return o
}

func (o *optionalObjectSchema) Custom(fn func(map[string]interface{}) error) OptionalObjectBuilder {
	o.customFunc = fn
	return o
}

func (o *optionalObjectSchema) Default(value map[string]interface{}) OptionalObjectBuilder {
	o.defaultValue = value
	return o
}

func (o *optionalObjectSchema) WithMessage(validationType, message string) OptionalObjectBuilder {
	if o.customError == nil {
		o.customError = make(map[string]string)
	}
	o.customError[validationType] = message
	return o
}

func (o *optionalObjectSchema) Validate(data interface{}) error {
	return o.validate(data)
}

// Object validation logic
func (o *objectSchema) validate(data interface{}) error {
	// Handle nil values
	if data == nil {
		if o.required {
			return goop.NewValidationError("", nil, o.getErrorMessage(errorKeys.Required, "field is required"))
		}
		if o.defaultValue != nil {
			return o.validate(o.defaultValue)
		}
		if o.optional {
			return nil
		}
		return goop.NewValidationError("", nil, o.getErrorMessage(errorKeys.Required, "field is required"))
	}

	// Type check - convert to map[string]interface{}
	var obj map[string]interface{}

	// Use reflection to handle different map types
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Map {
		return goop.NewValidationError(fmt.Sprintf("%v", data), data,
			o.getErrorMessage(errorKeys.Type, "invalid type, expected object"))
	}

	// Convert to map[string]interface{}
	obj = make(map[string]interface{})
	for _, key := range val.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		obj[keyStr] = val.MapIndex(key).Interface()
	}

	// Properties count validation
	propCount := len(obj)
	if o.minProperties > 0 && propCount < o.minProperties {
		return goop.NewValidationError(fmt.Sprintf("%v", obj), obj,
			o.getErrorMessage(errorKeys.MinProperties,
				fmt.Sprintf("object has too few properties, minimum is %d but got %d", o.minProperties, propCount)))
	}

	if o.maxProperties > 0 && propCount > o.maxProperties {
		return goop.NewValidationError(fmt.Sprintf("%v", obj), obj,
			o.getErrorMessage(errorKeys.MaxProperties,
				fmt.Sprintf("object has too many properties, maximum is %d but got %d", o.maxProperties, propCount)))
	}

	// Strict mode: check for unknown keys
	if o.strictMode {
		for key := range obj {
			if _, exists := o.schema[key]; !exists {
				return goop.NewValidationError(key, obj[key],
					o.getErrorMessage(errorKeys.UnknownKey,
						fmt.Sprintf("unknown key: %s", key)))
			}
		}
	}

	// Validate each field in the schema
	var details []goop.ValidationError
	for fieldName, fieldSchema := range o.schema {
		value, exists := obj[fieldName]

		// Handle missing fields
		if !exists {
			if !o.partialMode {
				// Check if field is required by trying to validate nil
				if err := o.validateField(fieldSchema, nil); err != nil {
					details = append(details, *goop.NewValidationError(fieldName, nil,
						fmt.Sprintf("missing required field: %s", fieldName)))
				}
			}
			continue
		}

		// Validate field
		if err := o.validateField(fieldSchema, value); err != nil {
			if validationErr, ok := err.(*goop.ValidationError); ok {
				validationErr.Field = fieldName
				details = append(details, *validationErr)
			} else {
				details = append(details, *goop.NewValidationError(fieldName, value, err.Error()))
			}
		}
	}

	if len(details) > 0 {
		return goop.NewNestedValidationError("", obj, "object validation failed", details)
	}

	// Custom validation
	if o.customFunc != nil {
		if err := o.customFunc(obj); err != nil {
			return err
		}
	}

	return nil
}

func (o *objectSchema) validateField(fieldSchema, value interface{}) error {
	// First, try the standard Validate method (for finalized schemas)
	if validator, ok := fieldSchema.(interface{ Validate(interface{}) error }); ok {
		return validator.Validate(value)
	}

	// Handle unfinalized schemas by type - automatically treat them as required
	switch schema := fieldSchema.(type) {
	case *stringSchema:
		// Create a required string validator from the unfinalized schema
		requiredSchema := &requiredStringSchema{schema}
		requiredSchema.required = true
		requiredSchema.optional = false
		return requiredSchema.Validate(value)

	case *numberSchema:
		// Create a required number validator from the unfinalized schema
		requiredSchema := &requiredNumberSchema{schema}
		requiredSchema.required = true
		requiredSchema.optional = false
		return requiredSchema.Validate(value)

	case *objectSchema:
		// Create a required object validator from the unfinalized schema
		requiredSchema := &requiredObjectSchema{schema}
		requiredSchema.required = true
		requiredSchema.optional = false
		return requiredSchema.Validate(value)

	case *boolSchema:
		// Create a required bool validator from the unfinalized schema
		requiredSchema := &requiredBoolSchema{schema}
		requiredSchema.required = true
		requiredSchema.optional = false
		return requiredSchema.Validate(value)

	case *arraySchema:
		// Create a required array validator from the unfinalized schema
		requiredSchema := &requiredArraySchema{schema}
		requiredSchema.required = true
		requiredSchema.optional = false
		return requiredSchema.Validate(value)
	}

	// Try reflection as a fallback for other types
	val := reflect.ValueOf(fieldSchema)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Look for a Validate method using reflection
	validateMethod := val.MethodByName("Validate")
	if validateMethod.IsValid() {
		// Call the Validate method
		results := validateMethod.Call([]reflect.Value{reflect.ValueOf(value)})
		if len(results) > 0 {
			if err, ok := results[0].Interface().(error); ok {
				return err
			}
		}
		return nil
	}

	// If we can't find a way to validate, that's an error in the schema definition
	return fmt.Errorf("field schema does not implement validation interface: %T", fieldSchema)
}

func (o *objectSchema) getErrorMessage(validationType, defaultMessage string) string {
	if o.customError != nil {
		if msg, exists := o.customError[validationType]; exists {
			return msg
		}
	}
	return defaultMessage
}

// BoolBuilder implementation (initial state)
func (b *boolSchema) Custom(fn func(bool) error) BoolBuilder {
	b.customFunc = fn
	return b
}

func (b *boolSchema) Required() RequiredBoolBuilder {
	b.required = true
	b.optional = false
	return &requiredBoolSchema{b}
}

func (b *boolSchema) Optional() OptionalBoolBuilder {
	b.optional = true
	b.required = false
	return &optionalBoolSchema{b}
}

func (b *boolSchema) WithMessage(validationType, message string) BoolBuilder {
	if b.customError == nil {
		b.customError = make(map[string]string)
	}
	b.customError[validationType] = message
	return b
}

// RequiredBoolBuilder implementation
func (r *requiredBoolSchema) Custom(fn func(bool) error) RequiredBoolBuilder {
	r.customFunc = fn
	return r
}

func (r *requiredBoolSchema) WithMessage(validationType, message string) RequiredBoolBuilder {
	if r.customError == nil {
		r.customError = make(map[string]string)
	}
	r.customError[validationType] = message
	return r
}

func (r *requiredBoolSchema) WithRequiredMessage(message string) RequiredBoolBuilder {
	return r.WithMessage(errorKeys.Required, message)
}

func (r *requiredBoolSchema) Validate(data interface{}) error {
	return r.validate(data)
}

// OptionalBoolBuilder implementation
func (o *optionalBoolSchema) Custom(fn func(bool) error) OptionalBoolBuilder {
	o.customFunc = fn
	return o
}

func (o *optionalBoolSchema) Default(value bool) OptionalBoolBuilder {
	o.defaultValue = &value
	return o
}

func (o *optionalBoolSchema) WithMessage(validationType, message string) OptionalBoolBuilder {
	if o.customError == nil {
		o.customError = make(map[string]string)
	}
	o.customError[validationType] = message
	return o
}

func (o *optionalBoolSchema) Validate(data interface{}) error {
	return o.validate(data)
}

// Bool validation logic
func (b *boolSchema) validate(data interface{}) error {
	// Handle nil values
	if data == nil {
		if b.required {
			return goop.NewValidationError("", nil, b.getErrorMessage(errorKeys.Required, "field is required"))
		}
		if b.defaultValue != nil {
			return b.validate(*b.defaultValue)
		}
		if b.optional {
			return nil
		}
		return goop.NewValidationError("", nil, b.getErrorMessage(errorKeys.Required, "field is required"))
	}

	// Type check
	boolVal, ok := data.(bool)
	if !ok {
		return goop.NewValidationError(fmt.Sprintf("%v", data), data,
			b.getErrorMessage(errorKeys.Type, "invalid type, expected boolean"))
	}

	// Custom validation
	if b.customFunc != nil {
		if err := b.customFunc(boolVal); err != nil {
			return err
		}
	}

	return nil
}

// Example methods for ObjectBuilder
func (o *objectSchema) Example(value interface{}) ObjectBuilder {
	o.example = value
	return o
}

func (o *objectSchema) Examples(examples map[string]ExampleObject) ObjectBuilder {
	o.examples = examples
	return o
}

func (o *objectSchema) ExampleFromFile(path string) ObjectBuilder {
	o.externalValue = path
	return o
}

// Example methods for RequiredObjectBuilder
func (r *requiredObjectSchema) Example(value interface{}) RequiredObjectBuilder {
	r.example = value
	return r
}

func (r *requiredObjectSchema) Examples(examples map[string]ExampleObject) RequiredObjectBuilder {
	r.examples = examples
	return r
}

func (r *requiredObjectSchema) ExampleFromFile(path string) RequiredObjectBuilder {
	r.externalValue = path
	return r
}

// Example methods for OptionalObjectBuilder
func (o *optionalObjectSchema) Example(value interface{}) OptionalObjectBuilder {
	o.example = value
	return o
}

func (o *optionalObjectSchema) Examples(examples map[string]ExampleObject) OptionalObjectBuilder {
	o.examples = examples
	return o
}

func (o *optionalObjectSchema) ExampleFromFile(path string) OptionalObjectBuilder {
	o.externalValue = path
	return o
}

// Example methods for BoolBuilder
func (b *boolSchema) Example(value interface{}) BoolBuilder {
	b.example = value
	return b
}

func (b *boolSchema) Examples(examples map[string]ExampleObject) BoolBuilder {
	b.examples = examples
	return b
}

func (b *boolSchema) ExampleFromFile(path string) BoolBuilder {
	b.externalValue = path
	return b
}

// Example methods for RequiredBoolBuilder
func (r *requiredBoolSchema) Example(value interface{}) RequiredBoolBuilder {
	r.example = value
	return r
}

func (r *requiredBoolSchema) Examples(examples map[string]ExampleObject) RequiredBoolBuilder {
	r.examples = examples
	return r
}

func (r *requiredBoolSchema) ExampleFromFile(path string) RequiredBoolBuilder {
	r.externalValue = path
	return r
}

// Example methods for OptionalBoolBuilder
func (o *optionalBoolSchema) Example(value interface{}) OptionalBoolBuilder {
	o.example = value
	return o
}

func (o *optionalBoolSchema) Examples(examples map[string]ExampleObject) OptionalBoolBuilder {
	o.examples = examples
	return o
}

func (o *optionalBoolSchema) ExampleFromFile(path string) OptionalBoolBuilder {
	o.externalValue = path
	return o
}

func (b *boolSchema) getErrorMessage(validationType, defaultMessage string) string {
	if b.customError != nil {
		if msg, exists := b.customError[validationType]; exists {
			return msg
		}
	}
	return defaultMessage
}
