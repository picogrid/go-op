package validators

import (
	"testing"

	goop "github.com/picogrid/go-op"
)

// Tests for all validator builder methods (Required/Optional builders for Array, Number, String, Object, Bool)
func TestOptionalArrayMethodCoverage(t *testing.T) {
	t.Run("OptionalArrayBuilder methods", func(t *testing.T) {
		// Test MinItems on OptionalArrayBuilder
		schema1 := Array(String()).MinItems(2).Optional()
		if err := schema1.Validate([]interface{}{"a", "b", "c"}); err != nil {
			t.Errorf("Expected valid array, got error: %v", err)
		}
		if err := schema1.Validate([]interface{}{"a"}); err == nil {
			t.Error("Expected error for too few items")
		}

		// Test MaxItems on OptionalArrayBuilder
		schema2 := Array(String()).MaxItems(2).Optional()
		if err := schema2.Validate([]interface{}{"a", "b"}); err != nil {
			t.Errorf("Expected valid array, got error: %v", err)
		}
		if err := schema2.Validate([]interface{}{"a", "b", "c"}); err == nil {
			t.Error("Expected error for too many items")
		}

		// Test Contains on OptionalArrayBuilder
		schema3 := Array(String()).Contains("test").Optional()
		if err := schema3.Validate([]interface{}{"test", "other"}); err != nil {
			t.Errorf("Expected valid array with required item, got error: %v", err)
		}
		if err := schema3.Validate([]interface{}{"other", "items"}); err == nil {
			t.Error("Expected error for missing required item")
		}

		// Test UniqueItems on OptionalArrayBuilder
		schema4 := Array(String()).UniqueItems().Optional()
		if err := schema4.Validate([]interface{}{"a", "b", "c"}); err != nil {
			t.Errorf("Expected valid unique array, got error: %v", err)
		}
		if err := schema4.Validate([]interface{}{"a", "b", "a"}); err == nil {
			t.Error("Expected error for duplicate items")
		}

		// Test Custom on OptionalArrayBuilder
		customValidator := func(arr []interface{}) error {
			if len(arr) > 0 && arr[0] == "forbidden" {
				return goop.NewValidationError("custom", arr, "first item cannot be 'forbidden'")
			}
			return nil
		}
		schema5 := Array(String()).Custom(customValidator).Optional()
		if err := schema5.Validate([]interface{}{"allowed", "items"}); err != nil {
			t.Errorf("Expected valid array, got error: %v", err)
		}
		if err := schema5.Validate([]interface{}{"forbidden", "items"}); err == nil {
			t.Error("Expected error for custom validation failure")
		}
	})

	t.Run("OptionalArrayBuilder example methods", func(t *testing.T) {
		// Test Examples method on OptionalArrayBuilder
		examples := map[string]ExampleObject{
			"simple": {Value: []interface{}{"a", "b"}, Summary: "Simple array"},
			"empty":  {Value: []interface{}{}, Summary: "Empty array"},
		}
		schema := Array(String()).Examples(examples).Optional()
		if err := schema.Validate([]interface{}{"test"}); err != nil {
			t.Errorf("Expected valid array, got error: %v", err)
		}

		// Test ExampleFromFile method on OptionalArrayBuilder
		schema2 := Array(String()).ExampleFromFile("test.json").Optional()
		if err := schema2.Validate([]interface{}{"test"}); err != nil {
			t.Errorf("Expected valid array, got error: %v", err)
		}
	})
}

// TestRequiredNumberMethodCoverage tests uncovered RequiredNumberBuilder methods
func TestRequiredNumberMethodCoverage(t *testing.T) {
	t.Run("RequiredNumberBuilder methods", func(t *testing.T) {
		// Test Min on RequiredNumberBuilder
		schema1 := Number().Min(10).Required()
		if err := schema1.Validate(15.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}
		if err := schema1.Validate(5.0); err == nil {
			t.Error("Expected error for number below minimum")
		}

		// Test Max on RequiredNumberBuilder
		schema2 := Number().Max(100).Required()
		if err := schema2.Validate(50.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}
		if err := schema2.Validate(150.0); err == nil {
			t.Error("Expected error for number above maximum")
		}

		// Test ExclusiveMin on RequiredNumberBuilder
		schema3 := Number().ExclusiveMin(0).Required()
		if err := schema3.Validate(1.0); err != nil {
			t.Errorf("Expected valid positive number, got error: %v", err)
		}
		if err := schema3.Validate(0.0); err == nil {
			t.Error("Expected error for number equal to exclusive minimum")
		}

		// Test ExclusiveMax on RequiredNumberBuilder
		schema4 := Number().ExclusiveMax(100).Required()
		if err := schema4.Validate(99.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}
		if err := schema4.Validate(100.0); err == nil {
			t.Error("Expected error for number equal to exclusive maximum")
		}

		// Test MultipleOf on RequiredNumberBuilder
		schema5 := Number().MultipleOf(5).Required()
		if err := schema5.Validate(25.0); err != nil {
			t.Errorf("Expected valid multiple, got error: %v", err)
		}
		if err := schema5.Validate(23.0); err == nil {
			t.Error("Expected error for non-multiple")
		}

		// Test Integer on RequiredNumberBuilder
		schema6 := Number().Integer().Required()
		if err := schema6.Validate(42.0); err != nil {
			t.Errorf("Expected valid integer, got error: %v", err)
		}
		if err := schema6.Validate(42.5); err == nil {
			t.Error("Expected error for non-integer")
		}

		// Test Positive on RequiredNumberBuilder
		schema7 := Number().Positive().Required()
		if err := schema7.Validate(5.0); err != nil {
			t.Errorf("Expected valid positive number, got error: %v", err)
		}
		if err := schema7.Validate(-5.0); err == nil {
			t.Error("Expected error for negative number")
		}

		// Test Negative on RequiredNumberBuilder
		schema8 := Number().Negative().Required()
		if err := schema8.Validate(-5.0); err != nil {
			t.Errorf("Expected valid negative number, got error: %v", err)
		}
		if err := schema8.Validate(5.0); err == nil {
			t.Error("Expected error for positive number")
		}

		// Test Custom on RequiredNumberBuilder
		customValidator := func(num float64) error {
			if num == 13.0 {
				return goop.NewValidationError("custom", num, "number cannot be 13")
			}
			return nil
		}
		schema9 := Number().Custom(customValidator).Required()
		if err := schema9.Validate(12.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}
		if err := schema9.Validate(13.0); err == nil {
			t.Error("Expected error for custom validation failure")
		}
	})

	t.Run("RequiredNumberBuilder message methods", func(t *testing.T) {
		// Test WithMinMessage on RequiredNumberBuilder
		schema := Number().Min(10).WithMinMessage("Must be at least 10").Required()
		err := schema.Validate(5.0)
		if err == nil {
			t.Error("Expected validation error")
		}

		// Test WithMaxMessage on RequiredNumberBuilder
		schema2 := Number().Max(100).WithMaxMessage("Must be at most 100").Required()
		err2 := schema2.Validate(150.0)
		if err2 == nil {
			t.Error("Expected validation error")
		}

		// Test WithIntegerMessage on RequiredNumberBuilder
		schema3 := Number().Integer().WithIntegerMessage("Must be a whole number").Required()
		err3 := schema3.Validate(42.5)
		if err3 == nil {
			t.Error("Expected validation error")
		}

		// Test WithPositiveMessage on RequiredNumberBuilder
		schema4 := Number().Positive().WithPositiveMessage("Must be positive").Required()
		err4 := schema4.Validate(-5.0)
		if err4 == nil {
			t.Error("Expected validation error")
		}

		// Test WithNegativeMessage on RequiredNumberBuilder
		schema5 := Number().Negative().WithNegativeMessage("Must be negative").Required()
		err5 := schema5.Validate(5.0)
		if err5 == nil {
			t.Error("Expected validation error")
		}
	})

	t.Run("RequiredNumberBuilder example methods", func(t *testing.T) {
		// Test ExampleFromFile on RequiredNumberBuilder
		schema := Number().ExampleFromFile("example.json").Required()
		if err := schema.Validate(42.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}

		// Test Example on RequiredNumberBuilder
		schema2 := Number().Example(42.0).Required()
		if err := schema2.Validate(42.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}

		// Test Examples on RequiredNumberBuilder
		examples := map[string]ExampleObject{
			"small": {Value: 1.0, Summary: "Small number"},
			"large": {Value: 1000.0, Summary: "Large number"},
		}
		schema3 := Number().Examples(examples).Required()
		if err := schema3.Validate(500.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}
	})
}

// TestOptionalNumberMethodCoverage tests uncovered OptionalNumberBuilder methods
func TestOptionalNumberMethodCoverage(t *testing.T) {
	t.Run("OptionalNumberBuilder methods", func(t *testing.T) {
		// Test Min on OptionalNumberBuilder
		schema1 := Number().Min(10).Optional()
		if err := schema1.Validate(15.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}
		if err := schema1.Validate(nil); err != nil {
			t.Errorf("Expected valid nil for optional, got error: %v", err)
		}

		// Test Max on OptionalNumberBuilder
		schema2 := Number().Max(100).Optional()
		if err := schema2.Validate(50.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}

		// Test ExclusiveMin on OptionalNumberBuilder
		schema3 := Number().ExclusiveMin(0).Optional()
		if err := schema3.Validate(1.0); err != nil {
			t.Errorf("Expected valid positive number, got error: %v", err)
		}

		// Test ExclusiveMax on OptionalNumberBuilder
		schema4 := Number().ExclusiveMax(100).Optional()
		if err := schema4.Validate(99.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}

		// Test MultipleOf on OptionalNumberBuilder
		schema5 := Number().MultipleOf(5).Optional()
		if err := schema5.Validate(25.0); err != nil {
			t.Errorf("Expected valid multiple, got error: %v", err)
		}

		// Test Integer on OptionalNumberBuilder
		schema6 := Number().Integer().Optional()
		if err := schema6.Validate(42.0); err != nil {
			t.Errorf("Expected valid integer, got error: %v", err)
		}

		// Test Positive on OptionalNumberBuilder
		schema7 := Number().Positive().Optional()
		if err := schema7.Validate(5.0); err != nil {
			t.Errorf("Expected valid positive number, got error: %v", err)
		}

		// Test Negative on OptionalNumberBuilder
		schema8 := Number().Negative().Optional()
		if err := schema8.Validate(-5.0); err != nil {
			t.Errorf("Expected valid negative number, got error: %v", err)
		}

		// Test Custom on OptionalNumberBuilder
		customValidator := func(num float64) error {
			if num == 13.0 {
				return goop.NewValidationError("custom", num, "number cannot be 13")
			}
			return nil
		}
		schema9 := Number().Custom(customValidator).Optional()
		if err := schema9.Validate(12.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}
	})

	t.Run("OptionalNumberBuilder message methods", func(t *testing.T) {
		// Test WithMessage on OptionalNumberBuilder
		schema := Number().Min(10).WithMessage("min", "Custom min message").Optional()
		if err := schema.Validate(15.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}

		// Test WithMinMessage on OptionalNumberBuilder
		schema2 := Number().Min(10).WithMinMessage("Must be at least 10").Optional()
		if err := schema2.Validate(15.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}

		// Test WithMaxMessage on OptionalNumberBuilder
		schema3 := Number().Max(100).WithMaxMessage("Must be at most 100").Optional()
		if err := schema3.Validate(50.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}

		// Test WithIntegerMessage on OptionalNumberBuilder
		schema4 := Number().Integer().WithIntegerMessage("Must be a whole number").Optional()
		if err := schema4.Validate(42.0); err != nil {
			t.Errorf("Expected valid integer, got error: %v", err)
		}

		// Test WithPositiveMessage on OptionalNumberBuilder
		schema5 := Number().Positive().WithPositiveMessage("Must be positive").Optional()
		if err := schema5.Validate(5.0); err != nil {
			t.Errorf("Expected valid positive number, got error: %v", err)
		}

		// Test WithNegativeMessage on OptionalNumberBuilder
		schema6 := Number().Negative().WithNegativeMessage("Must be negative").Optional()
		if err := schema6.Validate(-5.0); err != nil {
			t.Errorf("Expected valid negative number, got error: %v", err)
		}
	})

	t.Run("OptionalNumberBuilder example methods", func(t *testing.T) {
		// Test Example on OptionalNumberBuilder
		schema := Number().Example(42.0).Optional()
		if err := schema.Validate(42.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}

		// Test Examples on OptionalNumberBuilder
		examples := map[string]ExampleObject{
			"small": {Value: 1.0, Summary: "Small number"},
			"large": {Value: 1000.0, Summary: "Large number"},
		}
		schema2 := Number().Examples(examples).Optional()
		if err := schema2.Validate(500.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}

		// Test ExampleFromFile on OptionalNumberBuilder
		schema3 := Number().ExampleFromFile("example.json").Optional()
		if err := schema3.Validate(42.0); err != nil {
			t.Errorf("Expected valid number, got error: %v", err)
		}
	})
}

// TestRequiredStringMethodCoverage tests uncovered RequiredStringBuilder methods
func TestRequiredStringMethodCoverage(t *testing.T) {
	t.Run("RequiredStringBuilder methods", func(t *testing.T) {
		// Test Max on RequiredStringBuilder
		schema1 := String().Max(10).Required()
		if err := schema1.Validate("short"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}
		if err := schema1.Validate("this is too long"); err == nil {
			t.Error("Expected error for string too long")
		}

		// Test Pattern on RequiredStringBuilder
		schema2 := String().Pattern("^[A-Z]+$").Required()
		if err := schema2.Validate("UPPERCASE"); err != nil {
			t.Errorf("Expected valid uppercase string, got error: %v", err)
		}
		if err := schema2.Validate("lowercase"); err == nil {
			t.Error("Expected error for non-matching pattern")
		}

		// Test Email on RequiredStringBuilder
		schema3 := String().Email().Required()
		if err := schema3.Validate("test@example.com"); err != nil {
			t.Errorf("Expected valid email, got error: %v", err)
		}
		if err := schema3.Validate("invalid-email"); err == nil {
			t.Error("Expected error for invalid email")
		}

		// Test URL on RequiredStringBuilder
		schema4 := String().URL().Required()
		if err := schema4.Validate("https://example.com"); err != nil {
			t.Errorf("Expected valid URL, got error: %v", err)
		}
		if err := schema4.Validate("not-a-url"); err == nil {
			t.Error("Expected error for invalid URL")
		}

		// Test Const on RequiredStringBuilder
		schema5 := String().Const("fixed").Required()
		if err := schema5.Validate("fixed"); err != nil {
			t.Errorf("Expected valid constant string, got error: %v", err)
		}
		if err := schema5.Validate("different"); err == nil {
			t.Error("Expected error for non-matching constant")
		}

		// Test Custom on RequiredStringBuilder
		customValidator := func(str string) error {
			if str == "forbidden" {
				return goop.NewValidationError("custom", str, "string cannot be 'forbidden'")
			}
			return nil
		}
		schema6 := String().Custom(customValidator).Required()
		if err := schema6.Validate("allowed"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}
		if err := schema6.Validate("forbidden"); err == nil {
			t.Error("Expected error for custom validation failure")
		}
	})

	t.Run("RequiredStringBuilder message methods", func(t *testing.T) {
		// Test WithMinLengthMessage on RequiredStringBuilder
		schema := String().Min(5).WithMinLengthMessage("Must be at least 5 characters").Required()
		if err := schema.Validate("hello"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// Test WithMaxLengthMessage on RequiredStringBuilder
		schema2 := String().Max(10).WithMaxLengthMessage("Must be at most 10 characters").Required()
		if err := schema2.Validate("short"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// Test WithPatternMessage on RequiredStringBuilder
		schema3 := String().Pattern("^[A-Z]+$").WithPatternMessage("Must be uppercase only").Required()
		if err := schema3.Validate("VALID"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// Test WithURLMessage on RequiredStringBuilder
		schema4 := String().URL().WithURLMessage("Must be a valid URL").Required()
		if err := schema4.Validate("https://example.com"); err != nil {
			t.Errorf("Expected valid URL, got error: %v", err)
		}
	})

	t.Run("RequiredStringBuilder example methods", func(t *testing.T) {
		// Test Example on RequiredStringBuilder
		schema := String().Example("example").Required()
		if err := schema.Validate("test"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// Test Examples on RequiredStringBuilder
		examples := map[string]ExampleObject{
			"short": {Value: "hi", Summary: "Short greeting"},
			"long":  {Value: "hello world", Summary: "Long greeting"},
		}
		schema2 := String().Examples(examples).Required()
		if err := schema2.Validate("test"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// Test ExampleFromFile on RequiredStringBuilder
		schema3 := String().ExampleFromFile("example.txt").Required()
		if err := schema3.Validate("test"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}
	})
}

// TestOptionalStringMethodCoverage tests uncovered OptionalStringBuilder methods
func TestOptionalStringMethodCoverage(t *testing.T) {
	t.Run("OptionalStringBuilder methods", func(t *testing.T) {
		// Test Min on OptionalStringBuilder
		schema1 := String().Min(5).Optional()
		if err := schema1.Validate("hello"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}
		if err := schema1.Validate(nil); err != nil {
			t.Errorf("Expected valid nil for optional, got error: %v", err)
		}

		// Test Max on OptionalStringBuilder
		schema2 := String().Max(10).Optional()
		if err := schema2.Validate("short"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// Test Pattern on OptionalStringBuilder
		schema3 := String().Pattern("^[A-Z]+$").Optional()
		if err := schema3.Validate("UPPERCASE"); err != nil {
			t.Errorf("Expected valid uppercase string, got error: %v", err)
		}

		// Test Email on OptionalStringBuilder
		schema4 := String().Email().Optional()
		if err := schema4.Validate("test@example.com"); err != nil {
			t.Errorf("Expected valid email, got error: %v", err)
		}

		// Test URL on OptionalStringBuilder
		schema5 := String().URL().Optional()
		if err := schema5.Validate("https://example.com"); err != nil {
			t.Errorf("Expected valid URL, got error: %v", err)
		}

		// Test Const on OptionalStringBuilder
		schema6 := String().Const("fixed").Optional()
		if err := schema6.Validate("fixed"); err != nil {
			t.Errorf("Expected valid constant string, got error: %v", err)
		}

		// Test Custom on OptionalStringBuilder
		customValidator := func(str string) error {
			if str == "forbidden" {
				return goop.NewValidationError("custom", str, "string cannot be 'forbidden'")
			}
			return nil
		}
		schema7 := String().Custom(customValidator).Optional()
		if err := schema7.Validate("allowed"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}
	})

	t.Run("OptionalStringBuilder message methods", func(t *testing.T) {
		// Test WithMessage on OptionalStringBuilder
		schema := String().Min(5).WithMessage("min", "Custom min message").Optional()
		if err := schema.Validate("hello"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// Test WithMinLengthMessage on OptionalStringBuilder
		schema2 := String().Min(5).WithMinLengthMessage("Must be at least 5 characters").Optional()
		if err := schema2.Validate("hello"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// Test WithMaxLengthMessage on OptionalStringBuilder
		schema3 := String().Max(10).WithMaxLengthMessage("Must be at most 10 characters").Optional()
		if err := schema3.Validate("short"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// Test WithPatternMessage on OptionalStringBuilder
		schema4 := String().Pattern("^[A-Z]+$").WithPatternMessage("Must be uppercase only").Optional()
		if err := schema4.Validate("VALID"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// Test WithEmailMessage on OptionalStringBuilder
		schema5 := String().Email().WithEmailMessage("Must be a valid email").Optional()
		if err := schema5.Validate("test@example.com"); err != nil {
			t.Errorf("Expected valid email, got error: %v", err)
		}

		// Test WithURLMessage on OptionalStringBuilder
		schema6 := String().URL().WithURLMessage("Must be a valid URL").Optional()
		if err := schema6.Validate("https://example.com"); err != nil {
			t.Errorf("Expected valid URL, got error: %v", err)
		}
	})
}

// TestRequiredObjectMethodCoverage tests uncovered RequiredObjectBuilder methods
func TestRequiredObjectMethodCoverage(t *testing.T) {
	t.Run("RequiredObjectBuilder methods", func(t *testing.T) {
		// Test Strict on RequiredObjectBuilder
		schema1 := Object(map[string]interface{}{
			"name": String().Required(),
		}).Strict().Required()
		if err := schema1.Validate(map[string]interface{}{"name": "test"}); err != nil {
			t.Errorf("Expected valid strict object, got error: %v", err)
		}

		// Test Partial on RequiredObjectBuilder
		schema2 := Object(map[string]interface{}{
			"name": String().Required(),
			"age":  Number().Required(),
		}).Partial().Required()
		if err := schema2.Validate(map[string]interface{}{"name": "test"}); err != nil {
			t.Errorf("Expected valid partial object, got error: %v", err)
		}

		// Test MinProperties on RequiredObjectBuilder
		schema3 := Object(map[string]interface{}{
			"a": String().Optional(),
			"b": String().Optional(),
			"c": String().Optional(),
		}).MinProperties(2).Required()
		if err := schema3.Validate(map[string]interface{}{"a": "1", "b": "2"}); err != nil {
			t.Errorf("Expected valid object with min properties, got error: %v", err)
		}

		// Test MaxProperties on RequiredObjectBuilder
		schema4 := Object(map[string]interface{}{
			"a": String().Optional(),
			"b": String().Optional(),
		}).MaxProperties(1).Required()
		if err := schema4.Validate(map[string]interface{}{"a": "1"}); err != nil {
			t.Errorf("Expected valid object with max properties, got error: %v", err)
		}

		// Test Custom on RequiredObjectBuilder
		customValidator := func(obj map[string]interface{}) error {
			if name, ok := obj["name"].(string); ok && name == "forbidden" {
				return goop.NewValidationError("custom", obj, "name cannot be 'forbidden'")
			}
			return nil
		}
		schema5 := Object(map[string]interface{}{
			"name": String().Required(),
		}).Custom(customValidator).Required()
		if err := schema5.Validate(map[string]interface{}{"name": "allowed"}); err != nil {
			t.Errorf("Expected valid object, got error: %v", err)
		}
	})
}

// TestOptionalObjectMethodCoverage tests uncovered OptionalObjectBuilder methods
func TestOptionalObjectMethodCoverage(t *testing.T) {
	t.Run("OptionalObjectBuilder methods", func(t *testing.T) {
		// Test Strict on OptionalObjectBuilder
		schema1 := Object(map[string]interface{}{
			"name": String().Required(),
		}).Strict().Optional()
		if err := schema1.Validate(map[string]interface{}{"name": "test"}); err != nil {
			t.Errorf("Expected valid strict object, got error: %v", err)
		}
		if err := schema1.Validate(nil); err != nil {
			t.Errorf("Expected valid nil for optional, got error: %v", err)
		}

		// Test Partial on OptionalObjectBuilder
		schema2 := Object(map[string]interface{}{
			"name": String().Required(),
			"age":  Number().Required(),
		}).Partial().Optional()
		if err := schema2.Validate(map[string]interface{}{"name": "test"}); err != nil {
			t.Errorf("Expected valid partial object, got error: %v", err)
		}

		// Test MinProperties on OptionalObjectBuilder
		schema3 := Object(map[string]interface{}{
			"a": String().Optional(),
			"b": String().Optional(),
		}).MinProperties(1).Optional()
		if err := schema3.Validate(map[string]interface{}{"a": "1"}); err != nil {
			t.Errorf("Expected valid object with min properties, got error: %v", err)
		}

		// Test MaxProperties on OptionalObjectBuilder
		schema4 := Object(map[string]interface{}{
			"a": String().Optional(),
			"b": String().Optional(),
		}).MaxProperties(1).Optional()
		if err := schema4.Validate(map[string]interface{}{"a": "1"}); err != nil {
			t.Errorf("Expected valid object with max properties, got error: %v", err)
		}

		// Test Custom on OptionalObjectBuilder
		customValidator := func(obj map[string]interface{}) error {
			if name, ok := obj["name"].(string); ok && name == "forbidden" {
				return goop.NewValidationError("custom", obj, "name cannot be 'forbidden'")
			}
			return nil
		}
		schema5 := Object(map[string]interface{}{
			"name": String().Optional(),
		}).Custom(customValidator).Optional()
		if err := schema5.Validate(map[string]interface{}{"name": "allowed"}); err != nil {
			t.Errorf("Expected valid object, got error: %v", err)
		}

		// Test WithMessage on OptionalObjectBuilder
		schema6 := Object(map[string]interface{}{
			"name": String().Required(),
		}).WithMessage("required", "Object is required").Optional()
		if err := schema6.Validate(map[string]interface{}{"name": "test"}); err != nil {
			t.Errorf("Expected valid object, got error: %v", err)
		}
	})
}

// TestRequiredBoolMethodCoverage tests uncovered RequiredBoolBuilder methods
func TestRequiredBoolMethodCoverage(t *testing.T) {
	t.Run("RequiredBoolBuilder methods", func(t *testing.T) {
		// Test Custom on RequiredBoolBuilder
		customValidator := func(b bool) error {
			if !b {
				return goop.NewValidationError("custom", b, "boolean must be true")
			}
			return nil
		}
		schema := Bool().Custom(customValidator).Required()
		if err := schema.Validate(true); err != nil {
			t.Errorf("Expected valid boolean, got error: %v", err)
		}
		if err := schema.Validate(false); err == nil {
			t.Error("Expected error for custom validation failure")
		}

		// Test WithMessage on RequiredBoolBuilder
		schema2 := Bool().WithMessage("required", "Boolean is required").Required()
		if err := schema2.Validate(true); err != nil {
			t.Errorf("Expected valid boolean, got error: %v", err)
		}
	})
}

// TestOptionalBoolMethodCoverage tests uncovered OptionalBoolBuilder methods
func TestOptionalBoolMethodCoverage(t *testing.T) {
	t.Run("OptionalBoolBuilder methods", func(t *testing.T) {
		// Test Custom on OptionalBoolBuilder
		customValidator := func(b bool) error {
			if !b {
				return goop.NewValidationError("custom", b, "boolean must be true")
			}
			return nil
		}
		schema := Bool().Custom(customValidator).Optional()
		if err := schema.Validate(true); err != nil {
			t.Errorf("Expected valid boolean, got error: %v", err)
		}
		if err := schema.Validate(nil); err != nil {
			t.Errorf("Expected valid nil for optional, got error: %v", err)
		}

		// Test WithMessage on OptionalBoolBuilder
		schema2 := Bool().WithMessage("type", "Must be boolean").Optional()
		if err := schema2.Validate(true); err != nil {
			t.Errorf("Expected valid boolean, got error: %v", err)
		}
	})
}
