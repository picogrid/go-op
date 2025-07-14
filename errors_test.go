package goop

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestNewValidationError tests the creation of validation errors
func TestNewValidationError(t *testing.T) {
	t.Run("Basic validation error creation", func(t *testing.T) {
		field := "username"
		value := "ab"
		message := "Username must be at least 3 characters"

		err := NewValidationError(field, value, message)

		if err.ErrorType != "Validation Error" {
			t.Errorf("Expected ErrorType 'Validation Error', got '%s'", err.ErrorType)
		}
		if err.Field != field {
			t.Errorf("Expected Field '%s', got '%s'", field, err.Field)
		}
		if err.Value != value {
			t.Errorf("Expected Value '%v', got '%v'", value, err.Value)
		}
		if err.Message != message {
			t.Errorf("Expected Message '%s', got '%s'", message, err.Message)
		}
		if len(err.Details) != 0 {
			t.Errorf("Expected no details, got %d details", len(err.Details))
		}
	})

	t.Run("Validation error with nil value", func(t *testing.T) {
		err := NewValidationError("field", nil, "field is required")

		if err.Value != nil {
			t.Errorf("Expected Value to be nil, got '%v'", err.Value)
		}
	})

	t.Run("Validation error with complex value", func(t *testing.T) {
		complexValue := map[string]interface{}{
			"nested": "value",
			"number": 42,
		}

		err := NewValidationError("data", complexValue, "invalid data structure")

		// For maps, we need to check content rather than direct comparison
		if errMap, ok := err.Value.(map[string]interface{}); !ok {
			t.Errorf("Expected Value to be map[string]interface{}, got %T", err.Value)
		} else if errMap["nested"] != "value" || errMap["number"] != 42 {
			t.Errorf("Expected Value to contain correct map content, got %v", errMap)
		}
	})
}

// TestNewNestedValidationError tests nested validation error creation
func TestNewNestedValidationError(t *testing.T) {
	t.Run("Nested validation error creation", func(t *testing.T) {
		// Create child errors
		childErr1 := *NewValidationError("name", "", "Name is required")
		childErr2 := *NewValidationError("age", -1, "Age must be positive")
		details := []ValidationError{childErr1, childErr2}

		// Create parent error
		parentErr := NewNestedValidationError("user", map[string]interface{}{}, "User validation failed", details)

		if parentErr.ErrorType != "Nested Validation Error" {
			t.Errorf("Expected ErrorType 'Nested Validation Error', got '%s'", parentErr.ErrorType)
		}
		if parentErr.Field != "user" {
			t.Errorf("Expected Field 'user', got '%s'", parentErr.Field)
		}
		if len(parentErr.Details) != 2 {
			t.Errorf("Expected 2 details, got %d", len(parentErr.Details))
		}
		if parentErr.Details[0].Field != "name" {
			t.Errorf("Expected first detail field 'name', got '%s'", parentErr.Details[0].Field)
		}
		if parentErr.Details[1].Field != "age" {
			t.Errorf("Expected second detail field 'age', got '%s'", parentErr.Details[1].Field)
		}
	})

	t.Run("Nested validation error with empty details", func(t *testing.T) {
		err := NewNestedValidationError("field", "value", "message", []ValidationError{})

		if len(err.Details) != 0 {
			t.Errorf("Expected no details, got %d details", len(err.Details))
		}
	})
}

// TestValidationErrorError tests the Error() method formatting
func TestValidationErrorError(t *testing.T) {
	t.Run("Simple validation error formatting", func(t *testing.T) {
		err := NewValidationError("email", "invalid-email", "Email format is invalid")
		errorString := err.Error()

		expected := "Field: email, Error: Email format is invalid"
		if errorString != expected {
			t.Errorf("Expected error string '%s', got '%s'", expected, errorString)
		}
	})

	t.Run("Validation error with empty field", func(t *testing.T) {
		err := NewValidationError("", "value", "Some error")
		errorString := err.Error()

		expected := "Field: , Error: Some error"
		if errorString != expected {
			t.Errorf("Expected error string '%s', got '%s'", expected, errorString)
		}
	})

	t.Run("Nested validation error formatting", func(t *testing.T) {
		childErr1 := *NewValidationError("name", "", "Name is required")
		childErr2 := *NewValidationError("age", -1, "Age must be positive")
		details := []ValidationError{childErr1, childErr2}

		parentErr := NewNestedValidationError("user", map[string]interface{}{}, "User validation failed", details)
		errorString := parentErr.Error()

		// Should contain both child errors
		if !strings.Contains(errorString, "Field: name, Error: Name is required") {
			t.Errorf("Expected error string to contain name error, got '%s'", errorString)
		}
		if !strings.Contains(errorString, "Field: age, Error: Age must be positive") {
			t.Errorf("Expected error string to contain age error, got '%s'", errorString)
		}
	})

	t.Run("Nested validation error with empty details", func(t *testing.T) {
		parentErr := NewNestedValidationError("field", "value", "message", []ValidationError{})
		errorString := parentErr.Error()

		// Should format as simple error when no details
		expected := "Field: field, Error: message"
		if errorString != expected {
			t.Errorf("Expected error string '%s', got '%s'", expected, errorString)
		}
	})
}

// TestValidationErrorErrorJSON tests JSON error formatting
func TestValidationErrorErrorJSON(t *testing.T) {
	t.Run("Simple validation error JSON", func(t *testing.T) {
		err := NewValidationError("email", "test@", "Invalid email format")
		jsonString := err.ErrorJSON()

		var result map[string]string
		if jsonErr := json.Unmarshal([]byte(jsonString), &result); jsonErr != nil {
			t.Fatalf("Failed to parse JSON: %v", jsonErr)
		}

		if result["field"] != "email" {
			t.Errorf("Expected field 'email', got '%s'", result["field"])
		}
		if result["message"] != "Invalid email format" {
			t.Errorf("Expected message 'Invalid email format', got '%s'", result["message"])
		}
	})

	t.Run("Nested validation error JSON", func(t *testing.T) {
		childErr1 := *NewValidationError("name", "", "Name is required")
		childErr2 := *NewValidationError("age", -1, "Age must be positive")
		childErr3 := *NewValidationError("email", "invalid", "Email format is invalid")
		details := []ValidationError{childErr1, childErr2, childErr3}

		parentErr := NewNestedValidationError("user", map[string]interface{}{}, "User validation failed", details)
		jsonString := parentErr.ErrorJSON()

		var result []map[string]string
		if jsonErr := json.Unmarshal([]byte(jsonString), &result); jsonErr != nil {
			t.Fatalf("Failed to parse JSON array: %v", jsonErr)
		}

		if len(result) != 4 {
			t.Errorf("Expected 4 errors in JSON (parent + 3 children), got %d", len(result))
		}

		// Check each error is present
		fields := make(map[string]string)
		for _, err := range result {
			fields[err["field"]] = err["message"]
		}

		if fields["name"] != "Name is required" {
			t.Errorf("Expected name error 'Name is required', got '%s'", fields["name"])
		}
		if fields["age"] != "Age must be positive" {
			t.Errorf("Expected age error 'Age must be positive', got '%s'", fields["age"])
		}
		if fields["email"] != "Email format is invalid" {
			t.Errorf("Expected email error 'Email format is invalid', got '%s'", fields["email"])
		}
	})

	t.Run("Empty field and message handling", func(t *testing.T) {
		err := NewValidationError("", "", "")
		jsonString := err.ErrorJSON()

		var result map[string]string
		if jsonErr := json.Unmarshal([]byte(jsonString), &result); jsonErr != nil {
			t.Fatalf("Failed to parse JSON: %v", jsonErr)
		}

		if result["field"] != "" {
			t.Errorf("Expected empty field, got '%s'", result["field"])
		}
		if result["message"] != "" {
			t.Errorf("Expected empty message, got '%s'", result["message"])
		}
	})
}

// TestCollectErrors tests the error collection mechanism
func TestCollectErrors(t *testing.T) {
	t.Run("Collect errors from simple validation error", func(t *testing.T) {
		err := NewValidationError("field", "value", "error message")
		var flatErrors []map[string]string
		err.collectErrors(&flatErrors)

		if len(flatErrors) != 1 {
			t.Errorf("Expected 1 collected error, got %d", len(flatErrors))
		}
		if flatErrors[0]["field"] != "field" {
			t.Errorf("Expected field 'field', got '%s'", flatErrors[0]["field"])
		}
		if flatErrors[0]["message"] != "error message" {
			t.Errorf("Expected message 'error message', got '%s'", flatErrors[0]["message"])
		}
	})

	t.Run("Collect errors with empty field and message", func(t *testing.T) {
		err := NewValidationError("", "", "")
		var flatErrors []map[string]string
		err.collectErrors(&flatErrors)

		// Should not collect errors with empty field and message
		if len(flatErrors) != 0 {
			t.Errorf("Expected 0 collected errors for empty field/message, got %d", len(flatErrors))
		}
	})

	t.Run("Collect errors from nested validation error", func(t *testing.T) {
		// Create a complex nested structure
		grandChildErr := *NewValidationError("email", "invalid", "Invalid email")
		childErr1 := *NewValidationError("name", "", "Name required")
		childErr2 := *NewNestedValidationError("contact", nil, "Contact invalid", []ValidationError{grandChildErr})

		parentErr := NewNestedValidationError("user", nil, "User invalid", []ValidationError{childErr1, childErr2})

		var flatErrors []map[string]string
		parentErr.collectErrors(&flatErrors)

		// Should collect: user, name, contact, email = 4 errors
		if len(flatErrors) != 4 {
			t.Errorf("Expected 4 collected errors, got %d", len(flatErrors))
		}

		// Check all errors are present
		fields := make(map[string]string)
		for _, err := range flatErrors {
			fields[err["field"]] = err["message"]
		}

		if fields["user"] != "User invalid" {
			t.Errorf("Expected user error, got '%s'", fields["user"])
		}
		if fields["name"] != "Name required" {
			t.Errorf("Expected name error, got '%s'", fields["name"])
		}
		if fields["contact"] != "Contact invalid" {
			t.Errorf("Expected contact error, got '%s'", fields["contact"])
		}
		if fields["email"] != "Invalid email" {
			t.Errorf("Expected email error, got '%s'", fields["email"])
		}
	})
}

// TestValidationErrorEdgeCases tests edge cases and error conditions
func TestValidationErrorEdgeCases(t *testing.T) {
	t.Run("Very long field names and messages", func(t *testing.T) {
		longField := strings.Repeat("field", 1000)
		longMessage := strings.Repeat("message", 1000)
		longValue := strings.Repeat("value", 1000)

		err := NewValidationError(longField, longValue, longMessage)

		if err.Field != longField {
			t.Error("Long field name not preserved")
		}
		if err.Message != longMessage {
			t.Error("Long message not preserved")
		}
		if err.Value != longValue {
			t.Error("Long value not preserved")
		}

		// Test JSON serialization with long strings
		jsonString := err.ErrorJSON()
		if jsonString == "" {
			t.Error("JSON serialization failed for long strings")
		}
	})

	t.Run("Special characters in fields and messages", func(t *testing.T) {
		specialField := "field\nwith\tspecial\rcharacters\""
		specialMessage := "message with \"quotes\" and \nnewlines"

		err := NewValidationError(specialField, nil, specialMessage)

		jsonString := err.ErrorJSON()
		var result map[string]string
		if jsonErr := json.Unmarshal([]byte(jsonString), &result); jsonErr != nil {
			t.Errorf("JSON serialization failed with special characters: %v", jsonErr)
		}

		if result["field"] != specialField {
			t.Error("Special characters in field not preserved in JSON")
		}
		if result["message"] != specialMessage {
			t.Error("Special characters in message not preserved in JSON")
		}
	})

	t.Run("Deeply nested validation errors", func(t *testing.T) {
		// Create a deeply nested structure (5 levels)
		level5 := *NewValidationError("level5", "value", "Level 5 error")
		level4 := *NewNestedValidationError("level4", nil, "Level 4 error", []ValidationError{level5})
		level3 := *NewNestedValidationError("level3", nil, "Level 3 error", []ValidationError{level4})
		level2 := *NewNestedValidationError("level2", nil, "Level 2 error", []ValidationError{level3})
		level1 := NewNestedValidationError("level1", nil, "Level 1 error", []ValidationError{level2})

		// Test error formatting doesn't cause stack overflow
		errorString := level1.Error()
		if errorString == "" {
			t.Error("Deep nesting caused error formatting to fail")
		}

		// Test JSON formatting doesn't cause stack overflow
		jsonString := level1.ErrorJSON()
		if jsonString == "" {
			t.Error("Deep nesting caused JSON formatting to fail")
		}

		// Verify all 5 levels are collected
		var flatErrors []map[string]string
		level1.collectErrors(&flatErrors)
		if len(flatErrors) != 5 {
			t.Errorf("Expected 5 errors from deep nesting, got %d", len(flatErrors))
		}
	})

	t.Run("Multiple errors at same level", func(t *testing.T) {
		// Create many errors at the same level
		var details []ValidationError
		for i := 0; i < 100; i++ {
			details = append(details, *NewValidationError(
				strings.Repeat("field", i+1),
				i,
				strings.Repeat("error", i+1),
			))
		}

		parentErr := NewNestedValidationError("parent", nil, "Many errors", details)

		// Test performance and correctness with many errors
		jsonString := parentErr.ErrorJSON()
		var result []map[string]string
		if jsonErr := json.Unmarshal([]byte(jsonString), &result); jsonErr != nil {
			t.Errorf("JSON serialization failed with many errors: %v", jsonErr)
		}

		// Should have parent + 100 child errors = 101 total
		if len(result) != 101 {
			t.Errorf("Expected 101 errors, got %d", len(result))
		}
	})
}

// TestValidationErrorIntegration tests integration with error interface
func TestValidationErrorIntegration(t *testing.T) {
	t.Run("Validation error implements error interface", func(t *testing.T) {
		var err error = NewValidationError("field", "value", "message")

		errorString := err.Error()
		expected := "Field: field, Error: message"
		if errorString != expected {
			t.Errorf("Expected error string '%s', got '%s'", expected, errorString)
		}
	})

	t.Run("Validation error in error handling patterns", func(t *testing.T) {
		validateData := func(data interface{}) error {
			if data == nil {
				return NewValidationError("data", data, "Data cannot be nil")
			}
			return nil
		}

		// Test successful validation
		if err := validateData("valid"); err != nil {
			t.Errorf("Expected no error for valid data, got %v", err)
		}

		// Test failed validation
		err := validateData(nil)
		if err == nil {
			t.Error("Expected error for nil data")
		}

		// Test error type assertion
		if validationErr, ok := err.(*ValidationError); ok {
			if validationErr.Field != "data" {
				t.Errorf("Expected field 'data', got '%s'", validationErr.Field)
			}
		} else {
			t.Error("Expected ValidationError type")
		}
	})
}
