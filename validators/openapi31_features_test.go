package validators

import (
	"testing"

	goop "github.com/picogrid/go-op"
)

func TestOpenAPI31StringConst(t *testing.T) {
	t.Run("Required string const - valid", func(t *testing.T) {
		schema := String().Const("fixed-value").Required()
		err := schema.Validate("fixed-value")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Required string const - invalid", func(t *testing.T) {
		schema := String().Const("fixed-value").Required()
		err := schema.Validate("different-value")
		if err == nil {
			t.Error("Expected error for invalid const value")
		}
		if validationErr, ok := err.(*goop.ValidationError); ok {
			expected := "value must be exactly 'fixed-value'"
			if validationErr.Message != expected {
				t.Errorf("Expected error message '%s', got '%s'", expected, validationErr.Message)
			}
		}
	})

	t.Run("Optional string const with valid value", func(t *testing.T) {
		schema := String().Const("api-version").Optional()
		err := schema.Validate("api-version")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Optional string const with nil value", func(t *testing.T) {
		schema := String().Const("api-version").Optional()
		err := schema.Validate(nil)
		if err != nil {
			t.Errorf("Expected no error for nil optional value, got: %v", err)
		}
	})
}

func TestOpenAPI31NumberMultipleOf(t *testing.T) {
	t.Run("MultipleOf - valid values", func(t *testing.T) {
		schema := Number().MultipleOf(5.0).Required()

		testCases := []float64{0, 5, 10, 15, 25, 100}
		for _, value := range testCases {
			err := schema.Validate(value)
			if err != nil {
				t.Errorf("Expected no error for value %v, got: %v", value, err)
			}
		}
	})

	t.Run("MultipleOf - invalid values", func(t *testing.T) {
		schema := Number().MultipleOf(3.0).Required()

		testCases := []float64{1, 2, 4, 5, 7, 8}
		for _, value := range testCases {
			err := schema.Validate(value)
			if err == nil {
				t.Errorf("Expected error for value %v (not multiple of 3)", value)
			}
		}
	})

	t.Run("MultipleOf with decimals", func(t *testing.T) {
		schema := Number().MultipleOf(0.5).Required()

		validCases := []float64{0, 0.5, 1.0, 1.5, 2.0, 2.5}
		for _, value := range validCases {
			err := schema.Validate(value)
			if err != nil {
				t.Errorf("Expected no error for value %v, got: %v", value, err)
			}
		}

		invalidCases := []float64{0.1, 0.3, 0.7, 1.2, 1.7}
		for _, value := range invalidCases {
			err := schema.Validate(value)
			if err == nil {
				t.Errorf("Expected error for value %v (not multiple of 0.5)", value)
			}
		}
	})
}

func TestOpenAPI31NumberExclusiveBounds(t *testing.T) {
	t.Run("ExclusiveMin - valid values", func(t *testing.T) {
		schema := Number().ExclusiveMin(10.0).Required()

		validCases := []float64{10.1, 11, 15, 100}
		for _, value := range validCases {
			err := schema.Validate(value)
			if err != nil {
				t.Errorf("Expected no error for value %v, got: %v", value, err)
			}
		}
	})

	t.Run("ExclusiveMin - invalid values", func(t *testing.T) {
		schema := Number().ExclusiveMin(10.0).Required()

		invalidCases := []float64{10.0, 9.9, 5, 0, -5}
		for _, value := range invalidCases {
			err := schema.Validate(value)
			if err == nil {
				t.Errorf("Expected error for value %v (not greater than 10)", value)
			}
		}
	})

	t.Run("ExclusiveMax - valid values", func(t *testing.T) {
		schema := Number().ExclusiveMax(100.0).Required()

		validCases := []float64{99.9, 50, 10, 0, -10}
		for _, value := range validCases {
			err := schema.Validate(value)
			if err != nil {
				t.Errorf("Expected no error for value %v, got: %v", value, err)
			}
		}
	})

	t.Run("ExclusiveMax - invalid values", func(t *testing.T) {
		schema := Number().ExclusiveMax(100.0).Required()

		invalidCases := []float64{100.0, 100.1, 150, 1000}
		for _, value := range invalidCases {
			err := schema.Validate(value)
			if err == nil {
				t.Errorf("Expected error for value %v (not less than 100)", value)
			}
		}
	})

	t.Run("Combined exclusive bounds", func(t *testing.T) {
		schema := Number().ExclusiveMin(0.0).ExclusiveMax(10.0).Required()

		validCases := []float64{0.1, 1, 5, 9.9}
		for _, value := range validCases {
			err := schema.Validate(value)
			if err != nil {
				t.Errorf("Expected no error for value %v, got: %v", value, err)
			}
		}

		invalidCases := []float64{0.0, 10.0, -1, 11}
		for _, value := range invalidCases {
			err := schema.Validate(value)
			if err == nil {
				t.Errorf("Expected error for value %v (not in range (0,10))", value)
			}
		}
	})
}

func TestOpenAPI31ArrayUniqueItems(t *testing.T) {
	t.Run("UniqueItems - valid arrays", func(t *testing.T) {
		schema := Array(String().Required()).UniqueItems().Required()

		validCases := [][]interface{}{
			{},
			{"single"},
			{"a", "b", "c"},
			{"unique", "values", "only"},
		}

		for _, value := range validCases {
			err := schema.Validate(value)
			if err != nil {
				t.Errorf("Expected no error for value %v, got: %v", value, err)
			}
		}
	})

	t.Run("UniqueItems - invalid arrays with duplicates", func(t *testing.T) {
		schema := Array(String().Required()).UniqueItems().Required()

		invalidCases := [][]interface{}{
			{"a", "a"},
			{"unique", "values", "unique"},
			{"a", "b", "c", "b"},
		}

		for _, value := range invalidCases {
			err := schema.Validate(value)
			if err == nil {
				t.Errorf("Expected error for duplicate array %v", value)
			}
		}
	})

	t.Run("UniqueItems with number array", func(t *testing.T) {
		schema := Array(Number().Required()).UniqueItems().Required()

		// Should work with different values
		validCase := []interface{}{1.0, 2.0, 3.0}
		err := schema.Validate(validCase)
		if err != nil {
			t.Errorf("Expected no error for unique numbers, got: %v", err)
		}

		// Should fail with duplicates
		invalidCase := []interface{}{1.0, 2.0, 1.0}
		err = schema.Validate(invalidCase)
		if err == nil {
			t.Error("Expected error for duplicate numbers")
		}
	})
}

func TestOpenAPI31ObjectProperties(t *testing.T) {
	t.Run("MinProperties - valid objects", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Optional(),
			"age":  Number().Optional(),
		}).MinProperties(1).Required()

		validCases := []map[string]interface{}{
			{"name": "John"},
			{"name": "John", "age": 30},
		}

		for _, value := range validCases {
			err := schema.Validate(value)
			if err != nil {
				t.Errorf("Expected no error for value %v, got: %v", value, err)
			}
		}
	})

	t.Run("MinProperties - invalid objects", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Optional(),
			"age":  Number().Optional(),
		}).MinProperties(2).Required()

		invalidCases := []map[string]interface{}{
			{},
			{"name": "John"},
		}

		for _, value := range invalidCases {
			err := schema.Validate(value)
			if err == nil {
				t.Errorf("Expected error for object %v (less than 2 properties)", value)
			}
		}
	})

	t.Run("MaxProperties - valid objects", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Optional(),
			"age":  Number().Optional(),
		}).MaxProperties(2).Required()

		validCases := []map[string]interface{}{
			{},
			{"name": "John"},
			{"name": "John", "age": 30},
		}

		for _, value := range validCases {
			err := schema.Validate(value)
			if err != nil {
				t.Errorf("Expected no error for value %v, got: %v", value, err)
			}
		}
	})

	t.Run("MaxProperties - invalid objects", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Optional(),
			"age":  Number().Optional(),
		}).MaxProperties(1).Required()

		invalidCase := map[string]interface{}{
			"name": "John",
			"age":  30,
		}

		err := schema.Validate(invalidCase)
		if err == nil {
			t.Errorf("Expected error for object %v (more than 1 property)", invalidCase)
		}
	})

	t.Run("Combined MinProperties and MaxProperties", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name":  String().Optional(),
			"age":   Number().Optional(),
			"email": String().Optional(),
		}).MinProperties(1).MaxProperties(2).Required()

		validCases := []map[string]interface{}{
			{"name": "John"},
			{"name": "John", "age": 30},
		}

		for _, value := range validCases {
			err := schema.Validate(value)
			if err != nil {
				t.Errorf("Expected no error for value %v, got: %v", value, err)
			}
		}

		invalidCases := []map[string]interface{}{
			{}, // too few
			{"name": "John", "age": 30, "email": "john@example.com"}, // too many
		}

		for _, value := range invalidCases {
			err := schema.Validate(value)
			if err == nil {
				t.Errorf("Expected error for object %v (not in range 1-2 properties)", value)
			}
		}
	})
}

func TestOpenAPI31CombinedFeatures(t *testing.T) {
	t.Run("Complex schema with multiple OpenAPI 3.1 features", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"api_version": String().Const("v1").Required(),
			"count":       Number().MultipleOf(5.0).ExclusiveMin(0.0).ExclusiveMax(100.0).Required(),
			"tags":        Array(String().Required()).UniqueItems().MinItems(1).MaxItems(3).Required(),
			"metadata": Object(map[string]interface{}{
				"source": String().Optional(),
			}).MinProperties(1).MaxProperties(5).Optional(),
		}).MinProperties(3).MaxProperties(4).Required()

		validCase := map[string]interface{}{
			"api_version": "v1",
			"count":       25.0,                          // Multiple of 5, between 0 and 100
			"tags":        []interface{}{"tag1", "tag2"}, // Unique items, 1-3 items
			"metadata": map[string]interface{}{
				"source": "web",
			}, // 1-5 properties
		}

		err := schema.Validate(validCase)
		if err != nil {
			t.Errorf("Expected no error for complex valid case, got: %v", err)
		}

		// Test various invalid cases
		invalidCases := []struct {
			name string
			data map[string]interface{}
		}{
			{
				"Wrong const value",
				map[string]interface{}{
					"api_version": "v2", // Should be "v1"
					"count":       25.0,
					"tags":        []interface{}{"tag1", "tag2"},
				},
			},
			{
				"Count not multiple of 5",
				map[string]interface{}{
					"api_version": "v1",
					"count":       23.0, // Not multiple of 5
					"tags":        []interface{}{"tag1", "tag2"},
				},
			},
			{
				"Count at exclusive boundary",
				map[string]interface{}{
					"api_version": "v1",
					"count":       100.0, // Should be < 100
					"tags":        []interface{}{"tag1", "tag2"},
				},
			},
			{
				"Duplicate tags",
				map[string]interface{}{
					"api_version": "v1",
					"count":       25.0,
					"tags":        []interface{}{"tag1", "tag1"}, // Duplicates not allowed
				},
			},
		}

		for _, testCase := range invalidCases {
			t.Run(testCase.name, func(t *testing.T) {
				err := schema.Validate(testCase.data)
				if err == nil {
					t.Errorf("Expected error for %s", testCase.name)
				}
			})
		}
	})
}
