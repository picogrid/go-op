package validators

import (
	"fmt"
	"reflect"

	goop "github.com/picogrid/go-op"
)

// Core array schema struct (unexported)
// This contains all the validation configuration and is wrapped by state-specific types
type arraySchema struct {
	elementSchema interface{}
	minItems      int
	maxItems      int
	contains      interface{}
	uniqueItems   bool
	customFunc    func([]interface{}) error
	required      bool
	optional      bool
	defaultValue  []interface{}
	customError   map[string]string
	example       interface{}
	examples      map[string]ExampleObject
	externalValue string
}

// State wrapper types for compile-time safety
type requiredArraySchema struct {
	*arraySchema
}

type optionalArraySchema struct {
	*arraySchema
}

// ArrayBuilder implementation (initial state)
// These methods return ArrayBuilder to allow continued configuration

func (a *arraySchema) MinItems(count int) ArrayBuilder {
	a.minItems = count
	return a
}

func (a *arraySchema) MaxItems(count int) ArrayBuilder {
	a.maxItems = count
	return a
}

func (a *arraySchema) Contains(value interface{}) ArrayBuilder {
	a.contains = value
	return a
}

func (a *arraySchema) UniqueItems() ArrayBuilder {
	a.uniqueItems = true
	return a
}

func (a *arraySchema) Custom(fn func([]interface{}) error) ArrayBuilder {
	a.customFunc = fn
	return a
}

// State transition methods - these change the return type to enforce compile-time safety
func (a *arraySchema) Required() RequiredArrayBuilder {
	a.required = true
	a.optional = false
	return &requiredArraySchema{a}
}

func (a *arraySchema) Optional() OptionalArrayBuilder {
	a.optional = true
	a.required = false
	return &optionalArraySchema{a}
}

// Error message methods for ArrayBuilder
func (a *arraySchema) WithMessage(validationType, message string) ArrayBuilder {
	if a.customError == nil {
		a.customError = make(map[string]string)
	}
	a.customError[validationType] = message
	return a
}

func (a *arraySchema) WithMinItemsMessage(message string) ArrayBuilder {
	return a.WithMessage(errorKeys.MinItems, message)
}

func (a *arraySchema) WithMaxItemsMessage(message string) ArrayBuilder {
	return a.WithMessage(errorKeys.MaxItems, message)
}

func (a *arraySchema) WithContainsMessage(message string) ArrayBuilder {
	return a.WithMessage(errorKeys.Contains, message)
}

// RequiredArrayBuilder implementation
// These methods return RequiredArrayBuilder to maintain the required state

func (r *requiredArraySchema) MinItems(count int) RequiredArrayBuilder {
	r.minItems = count
	return r
}

func (r *requiredArraySchema) MaxItems(count int) RequiredArrayBuilder {
	r.maxItems = count
	return r
}

func (r *requiredArraySchema) Contains(value interface{}) RequiredArrayBuilder {
	r.contains = value
	return r
}

func (r *requiredArraySchema) UniqueItems() RequiredArrayBuilder {
	r.uniqueItems = true
	return r
}

func (r *requiredArraySchema) Custom(fn func([]interface{}) error) RequiredArrayBuilder {
	r.customFunc = fn
	return r
}

// Error message methods for RequiredArrayBuilder
func (r *requiredArraySchema) WithMessage(validationType, message string) RequiredArrayBuilder {
	if r.customError == nil {
		r.customError = make(map[string]string)
	}
	r.customError[validationType] = message
	return r
}

func (r *requiredArraySchema) WithMinItemsMessage(message string) RequiredArrayBuilder {
	return r.WithMessage(errorKeys.MinItems, message)
}

func (r *requiredArraySchema) WithMaxItemsMessage(message string) RequiredArrayBuilder {
	return r.WithMessage(errorKeys.MaxItems, message)
}

func (r *requiredArraySchema) WithContainsMessage(message string) RequiredArrayBuilder {
	return r.WithMessage(errorKeys.Contains, message)
}

func (r *requiredArraySchema) WithRequiredMessage(message string) RequiredArrayBuilder {
	return r.WithMessage(errorKeys.Required, message)
}

// OptionalArrayBuilder implementation
// These methods return OptionalArrayBuilder to maintain the optional state

func (o *optionalArraySchema) MinItems(count int) OptionalArrayBuilder {
	o.minItems = count
	return o
}

func (o *optionalArraySchema) MaxItems(count int) OptionalArrayBuilder {
	o.maxItems = count
	return o
}

func (o *optionalArraySchema) Contains(value interface{}) OptionalArrayBuilder {
	o.contains = value
	return o
}

func (o *optionalArraySchema) UniqueItems() OptionalArrayBuilder {
	o.uniqueItems = true
	return o
}

func (o *optionalArraySchema) Custom(fn func([]interface{}) error) OptionalArrayBuilder {
	o.customFunc = fn
	return o
}

// Default is only available on optional builders - this is the key DX improvement!
func (o *optionalArraySchema) Default(value []interface{}) OptionalArrayBuilder {
	o.defaultValue = value
	return o
}

// Error message methods for OptionalArrayBuilder
func (o *optionalArraySchema) WithMessage(validationType, message string) OptionalArrayBuilder {
	if o.customError == nil {
		o.customError = make(map[string]string)
	}
	o.customError[validationType] = message
	return o
}

func (o *optionalArraySchema) WithMinItemsMessage(message string) OptionalArrayBuilder {
	return o.WithMessage(errorKeys.MinItems, message)
}

func (o *optionalArraySchema) WithMaxItemsMessage(message string) OptionalArrayBuilder {
	return o.WithMessage(errorKeys.MaxItems, message)
}

func (o *optionalArraySchema) WithContainsMessage(message string) OptionalArrayBuilder {
	return o.WithMessage(errorKeys.Contains, message)
}

// Validation methods - these are the final methods in the builder chain
func (r *requiredArraySchema) Validate(data interface{}) error {
	return r.validate(data)
}

func (o *optionalArraySchema) Validate(data interface{}) error {
	return o.validate(data)
}

// Core validation logic (shared between required and optional)
func (a *arraySchema) validate(data interface{}) error {
	// Handle nil values
	if data == nil {
		if a.required {
			return goop.NewValidationError("", nil, a.getErrorMessage(errorKeys.Required, "field is required"))
		}
		if a.defaultValue != nil {
			return a.validate(a.defaultValue)
		}
		if a.optional {
			return nil
		}
		return goop.NewValidationError("", nil, a.getErrorMessage(errorKeys.Required, "field is required"))
	}

	// Type check - convert to []interface{} if possible
	var arr []interface{}

	// Use reflection to handle different slice types
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return goop.NewValidationError(fmt.Sprintf("%v", data), data,
			a.getErrorMessage(errorKeys.Type, "invalid type, expected array"))
	}

	// Convert to []interface{}
	length := val.Len()
	arr = make([]interface{}, length)
	for i := 0; i < length; i++ {
		arr[i] = val.Index(i).Interface()
	}

	// Length validations
	if a.minItems > 0 && len(arr) < a.minItems {
		return goop.NewValidationError(fmt.Sprintf("%v", arr), arr,
			a.getErrorMessage(errorKeys.MinItems,
				fmt.Sprintf("array has too few items, minimum is %d", a.minItems)))
	}

	if a.maxItems > 0 && len(arr) > a.maxItems {
		return goop.NewValidationError(fmt.Sprintf("%v", arr), arr,
			a.getErrorMessage(errorKeys.MaxItems,
				fmt.Sprintf("array has too many items, maximum is %d", a.maxItems)))
	}

	// Element validation
	if a.elementSchema != nil {
		var details []goop.ValidationError
		for i, item := range arr {
			if err := a.validateElement(item); err != nil {
				if validationErr, ok := err.(*goop.ValidationError); ok {
					// Add index information to the error
					indexedErr := *validationErr
					indexedErr.Field = fmt.Sprintf("[%d]", i)
					details = append(details, indexedErr)
				} else {
					details = append(details, *goop.NewValidationError(fmt.Sprintf("[%d]", i), item, err.Error()))
				}
			}
		}
		if len(details) > 0 {
			return goop.NewNestedValidationError("", arr, "array contains invalid items", details)
		}
	}

	// Contains validation
	if a.contains != nil {
		found := false
		for _, item := range arr {
			if reflect.DeepEqual(item, a.contains) {
				found = true
				break
			}
		}
		if !found {
			return goop.NewValidationError(fmt.Sprintf("%v", arr), arr,
				a.getErrorMessage(errorKeys.Contains,
					fmt.Sprintf("array must contain value: %v", a.contains)))
		}
	}

	// Unique items validation
	if a.uniqueItems {
		seen := make(map[string]bool)
		for i, item := range arr {
			// Create a string representation for comparison
			key := fmt.Sprintf("%T:%v", item, item)
			if seen[key] {
				return goop.NewValidationError(fmt.Sprintf("%v", arr), arr,
					a.getErrorMessage(errorKeys.UniqueItems,
						fmt.Sprintf("array contains duplicate item at index %d: %v", i, item)))
			}
			seen[key] = true
		}
	}

	// Custom validation
	if a.customFunc != nil {
		if err := a.customFunc(arr); err != nil {
			return err
		}
	}

	return nil
}

// validateElement validates a single array element against the element schema
func (a *arraySchema) validateElement(item interface{}) error {
	// First, try the standard Validate method (for finalized schemas)
	if validator, ok := a.elementSchema.(interface{ Validate(interface{}) error }); ok {
		return validator.Validate(item)
	}

	// Handle unfinalized schemas by type - automatically treat them as required
	// IMPORTANT: Create COPIES to avoid data races in concurrent usage
	switch schema := a.elementSchema.(type) {
	case *stringSchema:
		// Create a COPY of the string schema to avoid race conditions
		schemaCopy := *schema // This creates a copy of the struct
		requiredSchema := &requiredStringSchema{&schemaCopy}
		requiredSchema.required = true
		requiredSchema.optional = false
		return requiredSchema.Validate(item)

	case *numberSchema:
		// Create a COPY of the number schema to avoid race conditions
		schemaCopy := *schema // This creates a copy of the struct
		requiredSchema := &requiredNumberSchema{&schemaCopy}
		requiredSchema.required = true
		requiredSchema.optional = false
		return requiredSchema.Validate(item)

	case *objectSchema:
		// Create a COPY of the object schema to avoid race conditions
		schemaCopy := *schema // This creates a copy of the struct
		requiredSchema := &requiredObjectSchema{&schemaCopy}
		requiredSchema.required = true
		requiredSchema.optional = false
		return requiredSchema.Validate(item)

	case *boolSchema:
		// Create a COPY of the bool schema to avoid race conditions
		schemaCopy := *schema // This creates a copy of the struct
		requiredSchema := &requiredBoolSchema{&schemaCopy}
		requiredSchema.required = true
		requiredSchema.optional = false
		return requiredSchema.Validate(item)

	case *arraySchema:
		// Create a COPY of the array schema to avoid race conditions
		schemaCopy := *schema // This creates a copy of the struct
		requiredSchema := &requiredArraySchema{&schemaCopy}
		requiredSchema.required = true
		requiredSchema.optional = false
		return requiredSchema.Validate(item)
	}

	// Try reflection as a fallback for other types
	val := reflect.ValueOf(a.elementSchema)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Look for a Validate method using reflection
	validateMethod := val.MethodByName("Validate")
	if validateMethod.IsValid() {
		// Call the Validate method
		results := validateMethod.Call([]reflect.Value{reflect.ValueOf(item)})
		if len(results) > 0 {
			if err, ok := results[0].Interface().(error); ok {
				return err
			}
		}
		return nil
	}

	// If we can't find a way to validate, that's an error in the schema definition
	return fmt.Errorf("element schema does not implement validation interface: %T", a.elementSchema)
}

// Example methods for ArrayBuilder
func (a *arraySchema) Example(value interface{}) ArrayBuilder {
	a.example = value
	return a
}

func (a *arraySchema) Examples(examples map[string]ExampleObject) ArrayBuilder {
	a.examples = examples
	return a
}

func (a *arraySchema) ExampleFromFile(path string) ArrayBuilder {
	a.externalValue = path
	return a
}

// Example methods for RequiredArrayBuilder
func (r *requiredArraySchema) Example(value interface{}) RequiredArrayBuilder {
	r.example = value
	return r
}

func (r *requiredArraySchema) Examples(examples map[string]ExampleObject) RequiredArrayBuilder {
	r.examples = examples
	return r
}

func (r *requiredArraySchema) ExampleFromFile(path string) RequiredArrayBuilder {
	r.externalValue = path
	return r
}

// Example methods for OptionalArrayBuilder
func (o *optionalArraySchema) Example(value interface{}) OptionalArrayBuilder {
	o.example = value
	return o
}

func (o *optionalArraySchema) Examples(examples map[string]ExampleObject) OptionalArrayBuilder {
	o.examples = examples
	return o
}

func (o *optionalArraySchema) ExampleFromFile(path string) OptionalArrayBuilder {
	o.externalValue = path
	return o
}

// Helper methods (unexported)
func (a *arraySchema) getErrorMessage(validationType, defaultMessage string) string {
	if a.customError != nil {
		if msg, exists := a.customError[validationType]; exists {
			return msg
		}
	}
	return defaultMessage
}
