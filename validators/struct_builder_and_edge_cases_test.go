package validators

import (
	"testing"

	goop "github.com/picogrid/go-op"
)

// Tests for struct builder Schema() method and validator edge cases
func TestStructBuilderSchema(t *testing.T) {
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	t.Run("Schema method returns same as Build", func(t *testing.T) {
		builder := ForStruct[User]().
			Field("name", String().Min(1).Required()).
			Field("email", Email()).
			Field("age", Number().Min(0).Required())

		// Test Schema() method
		schema1 := builder.Schema()
		schema2 := builder.Build()

		// Both should be valid schemas
		if schema1 == nil {
			t.Error("Schema() returned nil")
		}
		if schema2 == nil {
			t.Error("Build() returned nil")
		}

		// Test validation with both schemas
		validUser := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		}

		err1 := schema1.Validate(validUser)
		err2 := schema2.Validate(validUser)

		if err1 != nil {
			t.Errorf("Schema() validation failed: %v", err1)
		}
		if err2 != nil {
			t.Errorf("Build() validation failed: %v", err2)
		}
	})

	t.Run("Schema method with invalid data", func(t *testing.T) {
		builder := ForStruct[User]().
			Field("name", String().Min(5).Required()).
			Field("email", Email()).
			Field("age", Number().Min(18).Required())

		schema := builder.Schema()

		invalidUser := map[string]interface{}{
			"name":  "Jo",      // Too short
			"email": "invalid", // Invalid email
			"age":   15,        // Too young
		}

		err := schema.Validate(invalidUser)
		if err == nil {
			t.Error("Expected validation error for invalid user")
		}
	})
}

// TestArrayUniqueItemsOptional tests UniqueItems for OptionalArrayBuilder (0% coverage)
func TestArrayUniqueItemsOptional(t *testing.T) {
	t.Run("Optional array with unique items constraint", func(t *testing.T) {
		schema := Array(String()).UniqueItems().Optional()

		// Test valid unique items
		validData := []interface{}{"apple", "banana", "cherry"}
		if err := schema.Validate(validData); err != nil {
			t.Errorf("Expected valid unique items, got error: %v", err)
		}

		// Test duplicate items (should fail)
		duplicateData := []interface{}{"apple", "banana", "apple"}
		if err := schema.Validate(duplicateData); err == nil {
			t.Error("Expected error for duplicate items")
		}

		// Test empty array (should be valid)
		if err := schema.Validate([]interface{}{}); err != nil {
			t.Errorf("Expected valid empty array, got error: %v", err)
		}

		// Test nil (should be valid for optional)
		if err := schema.Validate(nil); err != nil {
			t.Errorf("Expected valid nil for optional array, got error: %v", err)
		}
	})

	t.Run("Unique items with different types", func(t *testing.T) {
		// Create a schema that accepts any type by using Object with no required properties
		anySchema := Object(map[string]interface{}{}).Optional()
		schema := Array(anySchema).UniqueItems().Optional()

		// Mixed types but unique
		mixedData := []interface{}{
			map[string]interface{}{"type": "string"},
			map[string]interface{}{"type": "number"},
			map[string]interface{}{"type": "boolean"},
		}
		if err := schema.Validate(mixedData); err != nil {
			t.Errorf("Expected valid mixed unique items, got error: %v", err)
		}

		// Same objects (should fail uniqueness)
		duplicateMixed := []interface{}{
			map[string]interface{}{"type": "string"},
			map[string]interface{}{"type": "string"},
		}
		if err := schema.Validate(duplicateMixed); err == nil {
			t.Error("Expected error for duplicate mixed items")
		}
	})
}

// TestArrayMinItemsPlain tests MinItems for PlainArrayBuilder (0% coverage)
func TestArrayMinItemsPlain(t *testing.T) {
	t.Run("Plain array with min items constraint", func(t *testing.T) {
		schema := Array(String()).MinItems(2).Required()

		// Test valid array with enough items
		validData := []interface{}{"apple", "banana", "cherry"}
		if err := schema.Validate(validData); err != nil {
			t.Errorf("Expected valid array with min items, got error: %v", err)
		}

		// Test array with exactly min items
		exactData := []interface{}{"apple", "banana"}
		if err := schema.Validate(exactData); err != nil {
			t.Errorf("Expected valid array with exact min items, got error: %v", err)
		}

		// Test array with too few items
		tooFewData := []interface{}{"apple"}
		if err := schema.Validate(tooFewData); err == nil {
			t.Error("Expected error for array with too few items")
		}

		// Test empty array
		if err := schema.Validate([]interface{}{}); err == nil {
			t.Error("Expected error for empty array")
		}
	})
}

// TestStringBuilderEdgeCases tests edge cases and error conditions
func TestStringBuilderEdgeCases(t *testing.T) {
	t.Run("String validation with extreme constraints", func(t *testing.T) {
		// Very restrictive string
		schema := String().Min(10).Max(15).Pattern("^[a-z]+$").Required()

		// Valid string
		if err := schema.Validate("abcdefghij"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// String too short
		if err := schema.Validate("abc"); err == nil {
			t.Error("Expected error for too short string")
		}

		// String too long
		if err := schema.Validate("abcdefghijklmnop"); err == nil {
			t.Error("Expected error for too long string")
		}

		// String with invalid pattern
		if err := schema.Validate("ABC123"); err == nil {
			t.Error("Expected error for invalid pattern")
		}
	})

	t.Run("Email validation edge cases", func(t *testing.T) {
		schema := Email()

		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"user+tag@example.org",
			"123@domain.com",
		}

		for _, email := range validEmails {
			if err := schema.Validate(email); err != nil {
				t.Errorf("Expected valid email '%s', got error: %v", email, err)
			}
		}

		invalidEmails := []string{
			"invalid",
			"@domain.com",
			"user@",
			"",
		}

		for _, email := range invalidEmails {
			if err := schema.Validate(email); err == nil {
				t.Errorf("Expected invalid email '%s' to fail validation", email)
			}
		}
	})

	t.Run("URL validation edge cases", func(t *testing.T) {
		schema := URL()

		validURLs := []string{
			"https://example.com",
			"http://localhost:8080",
			"https://sub.domain.com/path?query=value",
			"ftp://files.example.com/file.txt",
		}

		for _, url := range validURLs {
			if err := schema.Validate(url); err != nil {
				t.Errorf("Expected valid URL '%s', got error: %v", url, err)
			}
		}

		invalidURLs := []string{
			"not-a-url",
			"://missing-scheme",
			"http://",
			"",
		}

		for _, url := range invalidURLs {
			if err := schema.Validate(url); err == nil {
				t.Errorf("Expected invalid URL '%s' to fail validation", url)
			}
		}
	})
}

// TestNumberBuilderEdgeCases tests number validation edge cases
func TestNumberBuilderEdgeCases(t *testing.T) {
	t.Run("Number validation with multiple constraints", func(t *testing.T) {
		schema := Number().Min(10.5).Max(99.9).MultipleOf(0.5).Required()

		// Valid numbers
		validNumbers := []interface{}{11.0, 50.5, 99.0, 15.5}
		for _, num := range validNumbers {
			if err := schema.Validate(num); err != nil {
				t.Errorf("Expected valid number %v, got error: %v", num, err)
			}
		}

		// Invalid numbers
		invalidNumbers := []interface{}{10.0, 100.0, 11.3, 50.7}
		for _, num := range invalidNumbers {
			if err := schema.Validate(num); err == nil {
				t.Errorf("Expected invalid number %v to fail validation", num)
			}
		}
	})

	t.Run("Integer validation with bounds", func(t *testing.T) {
		schema := Number().Integer().Min(1).Max(100).Required()

		// Valid integers
		validInts := []interface{}{1, 50, 100, int64(75), int32(25)}
		for _, num := range validInts {
			if err := schema.Validate(num); err != nil {
				t.Errorf("Expected valid integer %v, got error: %v", num, err)
			}
		}

		// Invalid values
		invalidValues := []interface{}{0, 101, 50.5, "50", true}
		for _, val := range invalidValues {
			if err := schema.Validate(val); err == nil {
				t.Errorf("Expected invalid value %v to fail validation", val)
			}
		}
	})
}

// TestObjectValidationEdgeCases tests object validation edge cases
func TestObjectValidationEdgeCases(t *testing.T) {
	t.Run("Object with nested validation constraints", func(t *testing.T) {
		userSchema := Object(map[string]interface{}{
			"profile": Object(map[string]interface{}{
				"name":   String().Min(1).Required(),
				"age":    Number().Min(0).Max(150).Required(),
				"email":  Email(),
				"active": Bool().Required(),
			}).Required(),
			"metadata": Object(map[string]interface{}{
				"tags":    Array(String()).Optional(),
				"created": String().Required(),
			}).Optional(),
		}).Required()

		// Valid nested object
		validData := map[string]interface{}{
			"profile": map[string]interface{}{
				"name":   "John Doe",
				"age":    30,
				"email":  "john@example.com",
				"active": true,
			},
			"metadata": map[string]interface{}{
				"tags":    []interface{}{"user", "premium"},
				"created": "2023-01-01",
			},
		}

		if err := userSchema.Validate(validData); err != nil {
			t.Errorf("Expected valid nested object, got error: %v", err)
		}

		// Invalid nested object (missing required field)
		invalidData := map[string]interface{}{
			"profile": map[string]interface{}{
				"name":  "John Doe",
				"age":   30,
				"email": "john@example.com",
				// missing "active" field
			},
		}

		if err := userSchema.Validate(invalidData); err == nil {
			t.Error("Expected error for missing required nested field")
		}
	})

	t.Run("Object with min/max properties", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"name": String().Optional(),
			"age":  Number().Optional(),
			"city": String().Optional(),
		}).MinProperties(1).MaxProperties(2).Required()

		// Valid object with acceptable property count
		validData := map[string]interface{}{
			"name": "John",
			"age":  30,
		}

		if err := schema.Validate(validData); err != nil {
			t.Errorf("Expected valid object with property count, got error: %v", err)
		}

		// Too few properties
		tooFewData := map[string]interface{}{}
		if err := schema.Validate(tooFewData); err == nil {
			t.Error("Expected error for too few properties")
		}

		// Too many properties
		tooManyData := map[string]interface{}{
			"name": "John",
			"age":  30,
			"city": "NYC",
		}
		if err := schema.Validate(tooManyData); err == nil {
			t.Error("Expected error for too many properties")
		}
	})
}

// TestBoolValidationEdgeCases tests boolean validation edge cases
func TestBoolValidationEdgeCases(t *testing.T) {
	t.Run("Bool validation with type strictness", func(t *testing.T) {
		schema := Bool().Required()

		// Valid boolean values
		validBools := []interface{}{true, false}
		for _, val := range validBools {
			if err := schema.Validate(val); err != nil {
				t.Errorf("Expected valid boolean %v, got error: %v", val, err)
			}
		}

		// Invalid non-boolean values
		invalidValues := []interface{}{1, 0, "true", "false", nil, []bool{true}}
		for _, val := range invalidValues {
			if err := schema.Validate(val); err == nil {
				t.Errorf("Expected invalid value %v to fail boolean validation", val)
			}
		}
	})

	t.Run("Optional bool with default", func(t *testing.T) {
		schema := Bool().Optional().Default(true)

		// Test valid boolean
		if err := schema.Validate(false); err != nil {
			t.Errorf("Expected valid boolean false, got error: %v", err)
		}

		// Test nil (should use default behavior)
		if err := schema.Validate(nil); err != nil {
			t.Errorf("Expected valid nil for optional boolean, got error: %v", err)
		}
	})
}

// TestCustomValidationEdgeCases tests custom validation functionality
func TestCustomValidationEdgeCases(t *testing.T) {
	t.Run("Custom string validation with complex logic", func(t *testing.T) {
		// Custom validation that checks for palindromes
		isPalindrome := func(str string) error {
			runes := []rune(str)
			length := len(runes)
			for i := 0; i < length/2; i++ {
				if runes[i] != runes[length-1-i] {
					return goop.NewValidationError("palindrome", str, "string must be a palindrome")
				}
			}
			return nil
		}

		schema := String().Custom(isPalindrome).Required()

		// Valid palindromes
		validPalindromes := []string{"racecar", "level", "madam", "a"}
		for _, palindrome := range validPalindromes {
			if err := schema.Validate(palindrome); err != nil {
				t.Errorf("Expected valid palindrome '%s', got error: %v", palindrome, err)
			}
		}

		// Invalid non-palindromes
		invalidStrings := []string{"hello", "world", "test"}
		for _, str := range invalidStrings {
			if err := schema.Validate(str); err == nil {
				t.Errorf("Expected invalid non-palindrome '%s' to fail validation", str)
			}
		}
	})

	t.Run("Custom number validation with mathematical constraints", func(t *testing.T) {
		// Custom validation for prime numbers
		isPrime := func(value float64) error {
			// Check if it's an integer
			if value != float64(int(value)) {
				return goop.NewValidationError("prime", value, "number must be an integer for prime check")
			}

			num := int(value)
			if num < 2 {
				return goop.NewValidationError("prime", value, "number must be prime (>= 2)")
			}

			for i := 2; i*i <= num; i++ {
				if num%i == 0 {
					return goop.NewValidationError("prime", value, "number must be prime")
				}
			}
			return nil
		}

		schema := Number().Custom(isPrime).Required()

		// Valid prime numbers
		validPrimes := []interface{}{2, 3, 5, 7, 11, 13, 17, 19, 23}
		for _, prime := range validPrimes {
			if err := schema.Validate(prime); err != nil {
				t.Errorf("Expected valid prime %v, got error: %v", prime, err)
			}
		}

		// Invalid non-prime numbers
		invalidNumbers := []interface{}{0, 1, 4, 6, 8, 9, 10, 12}
		for _, num := range invalidNumbers {
			if err := schema.Validate(num); err == nil {
				t.Errorf("Expected invalid non-prime %v to fail validation", num)
			}
		}
	})
}

// TestValidationErrorHandling tests error handling edge cases
func TestValidationErrorHandling(t *testing.T) {
	t.Run("Multiple nested validation errors", func(t *testing.T) {
		schema := Object(map[string]interface{}{
			"users": Array(Object(map[string]interface{}{
				"name":  String().Min(2).Required(),
				"email": Email(),
				"age":   Number().Min(18).Required(),
			}).Required()).Required(),
		}).Required()

		// Data with multiple validation errors
		invalidData := map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{
					"name":  "A",             // Too short
					"email": "invalid-email", // Invalid format
					"age":   15,              // Too young
				},
				map[string]interface{}{
					"name":  "",              // Empty (too short)
					"email": "test@test.com", // Valid
					"age":   "not-a-number",  // Wrong type
				},
			},
		}

		err := schema.Validate(invalidData)
		if err == nil {
			t.Error("Expected validation errors for invalid nested data")
		}

		// Verify we get specific error details
		if validationErr, ok := err.(*goop.ValidationError); ok {
			t.Logf("Validation errors: %s", validationErr.Error())
		} else {
			t.Error("Expected ValidationError type")
		}
	})

	t.Run("Custom validation with edge case handling", func(t *testing.T) {
		// Custom validation that handles edge cases
		validLength := func(str string) error {
			if len(str) < 5 || len(str) > 10 {
				return goop.NewValidationError("length", str, "string must be between 5 and 10 characters")
			}
			return nil
		}

		schema := String().Custom(validLength).Required()

		// Valid strings
		if err := schema.Validate("hello"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}
		if err := schema.Validate("helloworld"); err != nil {
			t.Errorf("Expected valid string, got error: %v", err)
		}

		// Invalid strings
		if err := schema.Validate("hi"); err == nil {
			t.Error("Expected error for too short string")
		}
		if err := schema.Validate("this is too long"); err == nil {
			t.Error("Expected error for too long string")
		}
	})
}
