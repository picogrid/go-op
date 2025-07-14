package validators

import (
	"testing"
)

func TestObjectValidator_Strict(t *testing.T) {
	schema := Object(map[string]interface{}{
		"name": String().Required(),
	}).Strict().Required()

	// Test with valid data
	validData := map[string]interface{}{"name": "test"}
	if err := schema.Validate(validData); err != nil {
		t.Errorf("Expected no error for valid data in strict mode, but got %v", err)
	}

	// Test with an unknown key
	invalidData := map[string]interface{}{"name": "test", "unknown": "key"}
	if err := schema.Validate(invalidData); err == nil {
		t.Errorf("Expected an error for unknown key in strict mode, but got nil")
	}
}

func TestObjectValidator_Partial(t *testing.T) {
	schema := Object(map[string]interface{}{
		"name":     String().Required(),
		"age":      Number().Required(),
		"optional": String().Optional(),
	}).Partial().Required()

	// Test with a subset of keys
	partialData := map[string]interface{}{"name": "test"}
	if err := schema.Validate(partialData); err != nil {
		t.Errorf("Expected no error for partial data, but got %v", err)
	}
}

func TestObjectValidator_Default(t *testing.T) {
	defaultValue := map[string]interface{}{"name": "default"}
	schema := Object(map[string]interface{}{
		"name": String().Required(),
	}).Optional().Default(defaultValue)

	if err := schema.Validate(nil); err != nil {
		t.Errorf("Expected no error for nil with default, but got %v", err)
	}
}

func TestBoolValidator_Default(t *testing.T) {
	schema := Bool().Optional().Default(true)
	if err := schema.Validate(nil); err != nil {
		t.Errorf("Expected no error for nil with default, but got %v", err)
	}
}
