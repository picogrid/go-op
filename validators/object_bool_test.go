package validators

import (
	"testing"

	"github.com/picogrid/go-op"
)

// TestObjectValidation tests object schema validation
func TestObjectValidation(t *testing.T) {
	t.Run("Basic object validation", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Required(),
			"age":  Number().Min(0).Required(),
		}).Required()

		// Valid object
		validData := map[string]interface{}{
			"name": "John",
			"age":  25,
		}

		if err := schema.Validate(validData); err != nil {
			t.Errorf("Expected valid object to pass, got: %v", err)
		}

		// Invalid object - missing required field
		invalidData := map[string]interface{}{
			"name": "John",
		}

		if err := schema.Validate(invalidData); err == nil {
			t.Error("Expected invalid object (missing age) to fail")
		}

		// Invalid object - wrong type
		wrongTypeData := map[string]interface{}{
			"name": "John",
			"age":  "twenty-five",
		}

		if err := schema.Validate(wrongTypeData); err == nil {
			t.Error("Expected invalid object (wrong age type) to fail")
		}
	})

	t.Run("Object with nested validation", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"user": Object(map[string]interface{}{
				"name":  String().Min(2).Required(),
				"email": String().Email().Required(),
			}).Required(),
			"preferences": Object(map[string]interface{}{
				"theme":    String().Required(),
				"language": String().Optional(),
			}).Optional(),
		}).Required()

		// Valid nested object
		validData := map[string]interface{}{
			"user": map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			"preferences": map[string]interface{}{
				"theme": "dark",
			},
		}

		if err := schema.Validate(validData); err != nil {
			t.Errorf("Expected valid nested object to pass, got: %v", err)
		}

		// Invalid nested object - invalid email
		invalidData := map[string]interface{}{
			"user": map[string]interface{}{
				"name":  "Jo",
				"email": "invalid-email",
			},
		}

		if err := schema.Validate(invalidData); err == nil {
			t.Error("Expected invalid nested object to fail")
		}
	})

	t.Run("Object strict mode", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Required(),
			"age":  Number().Required(),
		}).Strict().Required()

		// Valid object with exact fields
		validData := map[string]interface{}{
			"name": "John",
			"age":  25,
		}

		if err := schema.Validate(validData); err != nil {
			t.Errorf("Expected valid strict object to pass, got: %v", err)
		}

		// Invalid object with extra fields
		invalidData := map[string]interface{}{
			"name":    "John",
			"age":     25,
			"country": "USA", // Extra field
		}

		if err := schema.Validate(invalidData); err == nil {
			t.Error("Expected strict object with extra fields to fail")
		}
	})

	t.Run("Object partial mode", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Required(),
			"age":  Number().Required(),
			"city": String().Required(),
		}).Partial().Required()

		// Valid partial object - only some fields
		partialData := map[string]interface{}{
			"name": "John",
			"age":  25,
		}

		if err := schema.Validate(partialData); err != nil {
			t.Errorf("Expected partial object to pass, got: %v", err)
		}

		// Still validate provided fields
		invalidPartialData := map[string]interface{}{
			"name": "", // Invalid
			"age":  25,
		}

		if err := schema.Validate(invalidPartialData); err == nil {
			t.Error("Expected partial object with invalid field to fail")
		}
	})

	t.Run("Object with custom validation", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"password":        String().Min(8).Required(),
			"confirmPassword": String().Required(),
		}).Custom(func(data map[string]interface{}) error {
			password, ok1 := data["password"].(string)
			confirm, ok2 := data["confirmPassword"].(string)
			if ok1 && ok2 && password != confirm {
				return goop.NewValidationError("confirmPassword", data, "Passwords must match")
			}
			return nil
		}).Required()

		// Valid object with matching passwords
		validData := map[string]interface{}{
			"password":        "password123",
			"confirmPassword": "password123",
		}

		if err := schema.Validate(validData); err != nil {
			t.Errorf("Expected valid object with matching passwords to pass, got: %v", err)
		}

		// Invalid object with non-matching passwords
		invalidData := map[string]interface{}{
			"password":        "password123",
			"confirmPassword": "differentpass",
		}

		if err := schema.Validate(invalidData); err == nil {
			t.Error("Expected object with non-matching passwords to fail")
		}
	})

	t.Run("Optional object validation", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Required(),
		}).Optional()

		// Valid case - nil should pass for optional
		if err := schema.Validate(nil); err != nil {
			t.Errorf("Expected nil to pass for optional object, got: %v", err)
		}

		// Valid case - valid object should pass
		validData := map[string]interface{}{
			"name": "John",
		}

		if err := schema.Validate(validData); err != nil {
			t.Errorf("Expected valid object to pass, got: %v", err)
		}

		// Invalid case - invalid object should fail
		invalidData := map[string]interface{}{
			"name": "",
		}

		if err := schema.Validate(invalidData); err == nil {
			t.Error("Expected invalid object to fail even when optional")
		}
	})

	t.Run("Object with default value", func(t *testing.T) {
		defaultObj := map[string]interface{}{
			"theme": "light",
			"lang":  "en",
		}

		schema := Object(map[string]interface{}{
			"theme": String().Required(),
			"lang":  String().Required(),
		}).Optional().Default(defaultObj)

		// Test default value application
		if err := schema.Validate(nil); err != nil {
			t.Errorf("Expected nil to pass with default, got: %v", err)
		}

		// Test with provided value
		providedData := map[string]interface{}{
			"theme": "dark",
			"lang":  "es",
		}

		if err := schema.Validate(providedData); err != nil {
			t.Errorf("Expected provided object to pass, got: %v", err)
		}
	})
}

// TestObjectCustomMessages tests custom error messages for objects
func TestObjectCustomMessages(t *testing.T) {
	t.Run("Object with custom error messages", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Required(),
		}).Required().
			WithMessage("required", "This object field is required").
			WithRequiredMessage("Object is required")

		// Test required message
		if err := schema.Validate(nil); err != nil {
			if !contains(err.Error(), "Object is required") {
				t.Errorf("Expected custom required message, got: %v", err)
			}
		} else {
			t.Error("Expected nil object to fail")
		}

		// Test strict mode message
		strictSchema := Object(map[string]interface{}{
			"name": String().Required(),
		}).Strict().Required()

		invalidData := map[string]interface{}{
			"name":  "John",
			"extra": "field",
		}

		if err := strictSchema.Validate(invalidData); err == nil {
			t.Error("Expected strict validation to fail")
		}
	})
}

// TestBoolValidation tests boolean schema validation
func TestBoolValidation(t *testing.T) {
	t.Run("Basic boolean validation", func(t *testing.T) {
		schema := Bool().Required()

		// Valid boolean values
		validValues := []interface{}{true, false}
		for _, value := range validValues {
			if err := schema.Validate(value); err != nil {
				t.Errorf("Expected %v to be valid boolean, got: %v", value, err)
			}
		}

		// Invalid boolean values
		invalidValues := []interface{}{"true", 1, 0, nil}
		for _, value := range invalidValues {
			if err := schema.Validate(value); err == nil {
				t.Errorf("Expected %v to be invalid boolean", value)
			}
		}
	})

	t.Run("Optional boolean validation", func(t *testing.T) {
		schema := Bool().Optional()

		// Nil should pass for optional
		if err := schema.Validate(nil); err != nil {
			t.Errorf("Expected nil to pass for optional boolean, got: %v", err)
		}

		// Valid boolean should pass
		if err := schema.Validate(true); err != nil {
			t.Errorf("Expected true to pass for optional boolean, got: %v", err)
		}

		// Invalid value should still fail
		if err := schema.Validate("true"); err == nil {
			t.Error("Expected string 'true' to fail even for optional boolean")
		}
	})

	t.Run("Boolean with default value", func(t *testing.T) {
		schema := Bool().Optional().Default(false)

		// Test default value application
		if err := schema.Validate(nil); err != nil {
			t.Errorf("Expected nil to pass with default, got: %v", err)
		}

		// Test with provided value
		if err := schema.Validate(true); err != nil {
			t.Errorf("Expected provided boolean to pass, got: %v", err)
		}
	})

	t.Run("Boolean with custom validation", func(t *testing.T) {
		schema := Bool().Custom(func(value bool) error {
			if !value {
				return goop.NewValidationError("agreement", value, "Must agree to terms")
			}
			return nil
		}).Required()

		// Valid case - true
		if err := schema.Validate(true); err != nil {
			t.Errorf("Expected true to pass custom validation, got: %v", err)
		}

		// Invalid case - false
		if err := schema.Validate(false); err == nil {
			t.Error("Expected false to fail custom validation")
		} else {
			if !contains(err.Error(), "Must agree to terms") {
				t.Errorf("Expected custom error message, got: %v", err)
			}
		}
	})
}

// TestBoolCustomMessages tests custom error messages for booleans
func TestBoolCustomMessages(t *testing.T) {
	t.Run("Boolean with custom error messages", func(t *testing.T) {
		schema := Bool().Required().
			WithMessage("required", "Boolean value is required").
			WithRequiredMessage("This boolean field is required")

		// Test required message (WithRequiredMessage takes precedence)
		if err := schema.Validate(nil); err != nil {
			if !contains(err.Error(), "This boolean field is required") {
				t.Errorf("Expected custom required message, got: %v", err)
			}
		} else {
			t.Error("Expected nil boolean to fail")
		}

		// Test type validation message
		if err := schema.Validate("not a boolean"); err != nil {
			// Should contain some indication it's not a boolean
			if err.Error() == "" {
				t.Error("Expected non-empty error message for type mismatch")
			}
		} else {
			t.Error("Expected non-boolean to fail")
		}
	})
}

// TestObjectValidationInfo and TestBoolValidationInfo are disabled because
// GetValidationInfo methods are not available on builder interfaces

// TestObjectTypesAndInterfaces tests object type validation
func TestObjectTypesAndInterfaces(t *testing.T) {
	t.Run("Object accepts map[string]interface{}", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Required(),
		}).Required()

		data := map[string]interface{}{
			"name": "John",
		}

		if err := schema.Validate(data); err != nil {
			t.Errorf("Expected map[string]interface{} to pass, got: %v", err)
		}
	})

	t.Run("Object rejects non-map types", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Required(),
		}).Required()

		invalidTypes := []interface{}{
			"string",
			123,
			[]string{"array"},
			true,
		}

		for _, invalidType := range invalidTypes {
			if err := schema.Validate(invalidType); err == nil {
				t.Errorf("Expected %T to be rejected by object schema", invalidType)
			}
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
