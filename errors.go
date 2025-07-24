package goop

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ValidationError struct {
	ErrorType string            `json:"errorType"`
	Message   string            `json:"message"`
	Field     string            `json:"field"`
	Value     interface{}       `json:"value"`
	Details   []ValidationError `json:"details,omitempty"`
}

func NewValidationError(field string, value interface{}, message string) *ValidationError {
	// Sanitize value to avoid showing cryptic pointer addresses
	sanitizedValue := sanitizeValueForError(value)
	return &ValidationError{
		ErrorType: "Validation Error",
		Message:   message,
		Field:     field,
		Value:     sanitizedValue,
	}
}

func NewNestedValidationError(field string, value interface{}, message string, details []ValidationError) *ValidationError {
	// Sanitize value to avoid showing cryptic pointer addresses
	sanitizedValue := sanitizeValueForError(value)
	return &ValidationError{
		ErrorType: "Nested Validation Error",
		Message:   message,
		Field:     field,
		Value:     sanitizedValue,
		Details:   details,
	}
}

// sanitizeValueForError creates a clean representation of values for error messages
func sanitizeValueForError(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		// For pointers, show the dereferenced value
		return sanitizeValueForError(val.Elem().Interface())
	case reflect.Map:
		// Keep small maps as-is for test compatibility, sanitize large ones
		if val.Len() <= 5 {
			return value
		}
		return fmt.Sprintf("<%s with %d items>", val.Type().String(), val.Len())
	case reflect.Slice, reflect.Array:
		// Keep small collections as-is for test compatibility, sanitize large ones
		if val.Len() <= 5 {
			return value
		}
		return fmt.Sprintf("<%s with %d items>", val.Type().String(), val.Len())
	case reflect.Struct:
		// For structs, show type name to avoid clutter in error messages
		return fmt.Sprintf("<%s>", val.Type().String())
	default:
		// For primitives, return as-is
		return value
	}
}

func (v *ValidationError) Error() string {
	if len(v.Details) > 0 {
		return v.formatNestedError()
	}
	// Always use the "Field: X, Error: Y" format for backward compatibility
	// even when field is empty string
	return fmt.Sprintf("Field: %s, Error: %s", v.Field, v.Message)
}

func (v *ValidationError) formatNestedError() string {
	var result string
	for _, detail := range v.Details {
		result += v.formatErrorWithNesting(&detail, "")
	}
	// Remove trailing newline to match expected format
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}
	return result
}

func (v *ValidationError) formatErrorWithNesting(err *ValidationError, prefix string) string {
	var result string

	// If this error has sub-details, show them instead of the generic message
	if len(err.Details) > 0 {
		for _, subDetail := range err.Details {
			var fieldPath string
			if prefix != "" {
				fieldPath = prefix + "." + err.Field
			} else {
				fieldPath = err.Field
			}

			if len(subDetail.Details) > 0 {
				// Recursively format nested errors
				result += v.formatErrorWithNesting(&subDetail, fieldPath)
			} else {
				// Leaf error - show with field path for struct validation
				finalPath := fieldPath
				if subDetail.Field != "" {
					finalPath = fieldPath + "." + subDetail.Field
				}
				result += fmt.Sprintf("%s: %s\n", finalPath, subDetail.Message)
			}
		}
	} else {
		// No sub-details, show this error directly with appropriate format
		fieldPath := err.Field
		if prefix != "" {
			fieldPath = prefix + "." + err.Field
		}
		if fieldPath != "" {
			// Use old format for compatibility with existing tests
			result += fmt.Sprintf("Field: %s, Error: %s\n", fieldPath, err.Message)
		} else {
			result += fmt.Sprintf("%s\n", err.Message)
		}
	}

	return result
}

func (v *ValidationError) ErrorJSON() string {
	if len(v.Details) > 0 {
		return v.flattenNestedErrors()
	}

	errJSON, _ := json.Marshal(map[string]string{
		"field":   v.Field,
		"message": v.Message,
	})
	return string(errJSON)
}

func (v *ValidationError) flattenNestedErrors() string {
	var flatErrors []map[string]string

	v.collectErrors(&flatErrors)

	flatErrorJSON, _ := json.Marshal(flatErrors)
	return string(flatErrorJSON)
}

func (v *ValidationError) collectErrors(flatErrors *[]map[string]string) {
	if v.Field != "" && v.Message != "" {
		*flatErrors = append(*flatErrors, map[string]string{
			"field":   v.Field,
			"message": v.Message,
		})
	}

	for _, detail := range v.Details {
		detail.collectErrors(flatErrors)
	}
}
