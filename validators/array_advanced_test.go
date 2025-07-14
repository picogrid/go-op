package validators

import (
	"testing"

	goop "github.com/picogrid/go-op"
)

// TestArrayContains tests array contains validation
func TestArrayContains(t *testing.T) {
	t.Run("Array contains specific value", func(t *testing.T) {
		schema := Array(String()).Contains("required").Required()

		// Valid array with required value
		validArray := []interface{}{"apple", "required", "banana"}
		if err := schema.Validate(validArray); err != nil {
			t.Errorf("Expected array containing 'required' to pass, got: %v", err)
		}

		// Invalid array without required value
		invalidArray := []interface{}{"apple", "banana"}
		if err := schema.Validate(invalidArray); err == nil {
			t.Error("Expected array without 'required' to fail")
		}

		// Empty array should fail
		if err := schema.Validate([]interface{}{}); err == nil {
			t.Error("Expected empty array to fail contains validation")
		}
	})

	t.Run("Array contains with different types", func(t *testing.T) {
		schema := Array(Number()).Contains(42).Required()

		// Valid array with required number
		validArray := []interface{}{1, 42, 3}
		if err := schema.Validate(validArray); err != nil {
			t.Errorf("Expected array containing 42 to pass, got: %v", err)
		}

		// Invalid array without required number
		invalidArray := []interface{}{1, 2, 3}
		if err := schema.Validate(invalidArray); err == nil {
			t.Error("Expected array without 42 to fail")
		}
	})

	t.Run("Array contains with complex objects", func(t *testing.T) {
		requiredItem := map[string]interface{}{"id": 1, "name": "required"}
		schema := Array(Object(map[string]interface{}{
			"id":   Number().Required(),
			"name": String().Required(),
		})).Contains(requiredItem).Required()

		// Valid array with required object
		validArray := []interface{}{
			map[string]interface{}{"id": 1, "name": "required"},
			map[string]interface{}{"id": 2, "name": "other"},
		}
		if err := schema.Validate(validArray); err != nil {
			t.Errorf("Expected array containing required object to pass, got: %v", err)
		}

		// Invalid array without required object
		invalidArray := []interface{}{
			map[string]interface{}{"id": 2, "name": "other"},
		}
		if err := schema.Validate(invalidArray); err == nil {
			t.Error("Expected array without required object to fail")
		}
	})
}

// TestArrayCustomValidation tests custom array validation
func TestArrayCustomValidation(t *testing.T) {
	t.Run("Array custom validation function", func(t *testing.T) {
		schema := Array(String()).Custom(func(arr []interface{}) error {
			// Custom rule: array must have at least one string that starts with "prefix_"
			for _, item := range arr {
				if str, ok := item.(string); ok {
					if len(str) >= 7 && str[:7] == "prefix_" {
						return nil
					}
				}
			}
			return goop.NewValidationError("array", arr, "Array must contain at least one item starting with 'prefix_'")
		}).Required()

		// Valid array with prefixed string
		validArray := []interface{}{"hello", "prefix_test", "world"}
		if err := schema.Validate(validArray); err != nil {
			t.Errorf("Expected array with prefixed string to pass, got: %v", err)
		}

		// Invalid array without prefixed string
		invalidArray := []interface{}{"hello", "world"}
		if err := schema.Validate(invalidArray); err == nil {
			t.Error("Expected array without prefixed string to fail")
		} else if !contains(err.Error(), "prefix_") {
			t.Errorf("Expected custom error message, got: %v", err)
		}
	})

	t.Run("Array custom validation with complex logic", func(t *testing.T) {
		schema := Array(Number()).Custom(func(arr []interface{}) error {
			// Custom rule: sum of all numbers must be greater than 100
			sum := 0.0
			for _, item := range arr {
				if num, ok := item.(float64); ok {
					sum += num
				} else if num, ok := item.(int); ok {
					sum += float64(num)
				}
			}
			if sum <= 100 {
				return goop.NewValidationError("array", arr, "Sum of array elements must be greater than 100")
			}
			return nil
		}).Required()

		// Valid array with sum > 100
		validArray := []interface{}{50, 60}
		if err := schema.Validate(validArray); err != nil {
			t.Errorf("Expected array with sum > 100 to pass, got: %v", err)
		}

		// Invalid array with sum <= 100
		invalidArray := []interface{}{30, 40}
		if err := schema.Validate(invalidArray); err == nil {
			t.Error("Expected array with sum <= 100 to fail")
		}
	})
}

// TestArrayCustomMessages tests custom error messages for arrays
func TestArrayCustomMessages(t *testing.T) {
	t.Run("Array with all custom messages", func(t *testing.T) {
		schema := Array(String()).
			MinItems(2).
			MaxItems(5).
			Contains("required").
			Required().
			WithMessage("general", "Array validation failed").
			WithMinItemsMessage("Array needs at least 2 items").
			WithMaxItemsMessage("Array can have maximum 5 items").
			WithContainsMessage("Array must contain 'required' value").
			WithRequiredMessage("Array field is required")

		// Test required message
		if err := schema.Validate(nil); err != nil {
			if !contains(err.Error(), "Array field is required") {
				t.Errorf("Expected custom required message, got: %v", err)
			}
		} else {
			t.Error("Expected nil array to fail")
		}

		// Test min items message
		shortArray := []interface{}{"one"}
		if err := schema.Validate(shortArray); err != nil {
			if !contains(err.Error(), "at least 2 items") {
				t.Errorf("Expected custom min items message, got: %v", err)
			}
		} else {
			t.Error("Expected short array to fail")
		}

		// Test max items message
		longArray := []interface{}{"1", "2", "3", "4", "5", "6"}
		if err := schema.Validate(longArray); err != nil {
			if !contains(err.Error(), "maximum 5 items") {
				t.Errorf("Expected custom max items message, got: %v", err)
			}
		} else {
			t.Error("Expected long array to fail")
		}

		// Test contains message
		noRequiredArray := []interface{}{"one", "two"}
		if err := schema.Validate(noRequiredArray); err != nil {
			if !contains(err.Error(), "must contain 'required'") {
				t.Errorf("Expected custom contains message, got: %v", err)
			}
		} else {
			t.Error("Expected array without 'required' to fail")
		}
	})
}

// TestArrayBuilderMethods tests array builder method chaining
func TestArrayBuilderMethods(t *testing.T) {
	t.Run("Required array builder methods", func(t *testing.T) {
		// Test all builder methods on required array
		schema := Array(String()).
			Required().
			MinItems(1).
			MaxItems(10).
			Contains("test").
			Custom(func(arr []interface{}) error {
				return nil // Always pass
			}).
			WithMessage("general", "Custom message").
			WithMinItemsMessage("Min items message").
			WithMaxItemsMessage("Max items message").
			WithContainsMessage("Contains message").
			WithRequiredMessage("Required message")

		// Valid array should pass
		validArray := []interface{}{"test", "other"}
		if err := schema.Validate(validArray); err != nil {
			t.Errorf("Expected valid array to pass all validations, got: %v", err)
		}
	})

	t.Run("Optional array builder methods", func(t *testing.T) {
		// Test all builder methods on optional array
		// Use a default array that satisfies the Contains("test") requirement
		defaultArray := []interface{}{"test", "default"}
		schema := Array(String()).
			MinItems(1).
			MaxItems(10).
			Contains("test").
			Custom(func(arr []interface{}) error {
				// Allow any array that contains "test" or is the default
				for _, item := range arr {
					if item == "test" {
						return nil
					}
				}
				return goop.NewValidationError("array", arr, "Must contain test")
			}).
			Optional().
			Default(defaultArray).
			WithMessage("general", "Custom message").
			WithMinItemsMessage("Min items message").
			WithMaxItemsMessage("Max items message").
			WithContainsMessage("Contains message")

		// Nil should pass with default
		if err := schema.Validate(nil); err != nil {
			t.Errorf("Expected nil to pass with default, got: %v", err)
		}

		// Valid array should pass
		validArray := []interface{}{"test", "other"}
		if err := schema.Validate(validArray); err != nil {
			t.Errorf("Expected valid array to pass, got: %v", err)
		}
	})

	t.Run("Plain array builder methods", func(t *testing.T) {
		// Test builder methods on plain array (not required/optional)
		schema := Array(Number()).
			MinItems(2).
			MaxItems(5).
			Contains(42).
			Custom(func(arr []interface{}) error {
				return nil // Always pass
			}).
			WithMessage("general", "Custom message").
			WithMinItemsMessage("Min items message").
			WithMaxItemsMessage("Max items message").
			WithContainsMessage("Contains message").
			Required()

		// Valid array should pass
		validArray := []interface{}{1, 42, 3}
		if err := schema.Validate(validArray); err != nil {
			t.Errorf("Expected valid array to pass, got: %v", err)
		}

		// Invalid array should fail
		invalidArray := []interface{}{1}
		if err := schema.Validate(invalidArray); err == nil {
			t.Error("Expected invalid array to fail")
		}
	})
}

// TestArrayEdgeCases tests edge cases for array validation
func TestArrayEdgeCases(t *testing.T) {
	t.Run("Empty arrays", func(t *testing.T) {
		schema := Array(String()).MinItems(0).Required()

		// Empty array should pass with MinItems(0)
		if err := schema.Validate([]interface{}{}); err != nil {
			t.Errorf("Expected empty array to pass with MinItems(0), got: %v", err)
		}

		// Empty array should fail with MinItems(1)
		schemaMinOne := Array(String()).MinItems(1).Required()
		if err := schemaMinOne.Validate([]interface{}{}); err == nil {
			t.Error("Expected empty array to fail with MinItems(1)")
		}
	})

	t.Run("Large arrays", func(t *testing.T) {
		schema := Array(Number()).MaxItems(1000).Required()

		// Create large but valid array
		largeArray := make([]interface{}, 500)
		for i := range largeArray {
			largeArray[i] = i
		}

		if err := schema.Validate(largeArray); err != nil {
			t.Errorf("Expected large valid array to pass, got: %v", err)
		}

		// Create too large array
		tooLargeArray := make([]interface{}, 1001)
		for i := range tooLargeArray {
			tooLargeArray[i] = i
		}

		if err := schema.Validate(tooLargeArray); err == nil {
			t.Error("Expected too large array to fail")
		}
	})

	t.Run("String arrays with various strings", func(t *testing.T) {
		// Array schema that accepts strings
		schema := Array(String()).Required()

		stringArray := []interface{}{
			"hello",
			"world",
			"test",
		}

		if err := schema.Validate(stringArray); err != nil {
			t.Errorf("Expected string array to pass, got: %v", err)
		}

		// Mixed types should fail
		mixedArray := []interface{}{
			"string",
			42, // This should cause failure
		}

		if err := schema.Validate(mixedArray); err == nil {
			t.Error("Expected mixed type array to fail for string schema")
		}
	})

	t.Run("Nil vs empty array", func(t *testing.T) {
		requiredSchema := Array(String()).Required()
		optionalSchema := Array(String()).Optional()

		// Nil should fail for required
		if err := requiredSchema.Validate(nil); err == nil {
			t.Error("Expected nil to fail for required array")
		}

		// Nil should pass for optional
		if err := optionalSchema.Validate(nil); err != nil {
			t.Errorf("Expected nil to pass for optional array, got: %v", err)
		}

		// Empty array should pass for both (if no MinItems constraint)
		emptyArray := []interface{}{}
		if err := requiredSchema.Validate(emptyArray); err != nil {
			t.Errorf("Expected empty array to pass for required array, got: %v", err)
		}
		if err := optionalSchema.Validate(emptyArray); err != nil {
			t.Errorf("Expected empty array to pass for optional array, got: %v", err)
		}
	})

	t.Run("Arrays with complex nested validation", func(t *testing.T) {
		// Array of objects with nested arrays
		schema := Array(Object(map[string]interface{}{
			"name": String().Required(),
			"tags": Array(String()).MinItems(1).Required(),
		})).MinItems(1).Required()

		validData := []interface{}{
			map[string]interface{}{
				"name": "Item 1",
				"tags": []interface{}{"tag1", "tag2"},
			},
			map[string]interface{}{
				"name": "Item 2",
				"tags": []interface{}{"tag3"},
			},
		}

		if err := schema.Validate(validData); err != nil {
			t.Errorf("Expected complex nested array to pass, got: %v", err)
		}

		// Invalid - nested array doesn't meet MinItems
		invalidData := []interface{}{
			map[string]interface{}{
				"name": "Item 1",
				"tags": []interface{}{}, // Empty tags array
			},
		}

		if err := schema.Validate(invalidData); err == nil {
			t.Error("Expected invalid nested array to fail")
		}
	})
}

// TestArrayTypeValidation tests type validation for array elements
func TestArrayTypeValidation(t *testing.T) {
	t.Run("String array type validation", func(t *testing.T) {
		schema := Array(String()).Required()

		// Valid string array
		validArray := []interface{}{"hello", "world"}
		if err := schema.Validate(validArray); err != nil {
			t.Errorf("Expected string array to pass, got: %v", err)
		}

		// Invalid - contains non-string
		invalidArray := []interface{}{"hello", 123}
		if err := schema.Validate(invalidArray); err == nil {
			t.Error("Expected array with non-string to fail")
		}
	})

	t.Run("Number array type validation", func(t *testing.T) {
		schema := Array(Number()).Required()

		// Valid number array
		validArray := []interface{}{1, 2.5, 3}
		if err := schema.Validate(validArray); err != nil {
			t.Errorf("Expected number array to pass, got: %v", err)
		}

		// Invalid - contains non-number
		invalidArray := []interface{}{1, "not a number"}
		if err := schema.Validate(invalidArray); err == nil {
			t.Error("Expected array with non-number to fail")
		}
	})

	t.Run("Object array type validation", func(t *testing.T) {
		schema := Array(Object(map[string]interface{}{
			"id": Number().Required(),
		})).Required()

		// Valid object array
		validArray := []interface{}{
			map[string]interface{}{"id": 1},
			map[string]interface{}{"id": 2},
		}
		if err := schema.Validate(validArray); err != nil {
			t.Errorf("Expected object array to pass, got: %v", err)
		}

		// Invalid - contains non-object
		invalidArray := []interface{}{
			map[string]interface{}{"id": 1},
			"not an object",
		}
		if err := schema.Validate(invalidArray); err == nil {
			t.Error("Expected array with non-object to fail")
		}
	})
}
