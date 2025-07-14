package validators

import (
	"errors"
	"math"
	"regexp"
	"sync"
	"testing"
)

// TestNumberSchemaMapInitialization tests the critical bug fix for number schema
// Previously, the customError map was not initialized, causing panic on WithMessage calls
func TestNumberSchemaMapInitialization(t *testing.T) {
	t.Run("WithMessage does not panic", func(t *testing.T) {
		// This should not panic - the map should be initialized in Number()
		schema := Number().WithMessage("test", "custom error message")
		if schema == nil {
			t.Error("Expected schema to be created successfully")
		}
	})

	t.Run("All WithMessage methods work", func(t *testing.T) {
		schema := Number().
			WithMinMessage("custom min error").
			WithMaxMessage("custom max error").
			WithIntegerMessage("custom integer error").
			WithPositiveMessage("custom positive error").
			WithNegativeMessage("custom negative error")

		if schema == nil {
			t.Error("Expected schema to be created successfully")
		}

		// Test that we can transition to required state and continue
		requiredSchema := schema.Required()
		if requiredSchema == nil {
			t.Error("Expected required schema transition to work")
		}

		// Test validation with custom messages
		err := requiredSchema.Validate("not a number")
		if err == nil {
			t.Error("Expected validation to fail for non-number input")
		}
	})

	t.Run("Custom error messages are preserved", func(t *testing.T) {
		schema := Number().
			Min(10).
			WithMinMessage("Value must be at least 10").
			Required()

		err := schema.Validate(5.0)
		if err == nil {
			t.Error("Expected validation to fail for value below minimum")
		}

		// The error should contain our custom message in the expected format
		expectedError := "Field: 5, Error: Value must be at least 10"
		if err != nil && err.Error() != expectedError {
			t.Errorf("Expected custom error message '%s', got: '%v'", expectedError, err.Error())
		}
	})
}

// TestRegexErrorHandling tests regex compilation error handling in string validation
func TestRegexErrorHandling(t *testing.T) {
	t.Run("Invalid regex pattern handling", func(t *testing.T) {
		// Test invalid regex patterns that should be handled gracefully
		invalidPatterns := []string{
			"[",      // Unclosed bracket
			"(",      // Unclosed parenthesis
			"*",      // Invalid quantifier
			"(?P<>)", // Invalid named group
			"\\",     // Trailing backslash
		}

		for _, pattern := range invalidPatterns {
			t.Run("pattern_"+pattern, func(t *testing.T) {
				// This should not panic, even with invalid regex
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Regex pattern '%s' caused panic: %v", pattern, r)
					}
				}()

				schema := String().Pattern(pattern)
				if schema == nil {
					t.Error("Expected schema to be created even with invalid regex")
				}

				// Validation might fail, but should not panic
				requiredSchema := schema.Required()
				err := requiredSchema.Validate("test")

				// We expect an error due to invalid regex, but not a panic
				if err == nil {
					t.Logf("Warning: Invalid regex pattern '%s' did not produce validation error", pattern)
				}
			})
		}
	})

	t.Run("Valid regex patterns work correctly", func(t *testing.T) {
		validPatterns := map[string]string{
			"^[a-z]+$":             "lowercase",
			"\\d{3}-\\d{3}-\\d{4}": "123-456-7890",
			"^[A-Z][a-z]*$":        "Title",
		}

		for pattern, testValue := range validPatterns {
			t.Run("pattern_"+pattern, func(t *testing.T) {
				schema := String().Pattern(pattern).Required()

				// Test valid input
				err := schema.Validate(testValue)
				if err != nil {
					t.Errorf("Valid pattern '%s' failed for value '%s': %v", pattern, testValue, err)
				}

				// Test invalid input
				err = schema.Validate("invalid-123-XYZ")
				if err == nil {
					t.Errorf("Invalid input should have failed for pattern '%s'", pattern)
				}
			})
		}
	})
}

// TestArrayValidationRaceConditions tests thread safety in array validation
func TestArrayValidationRaceConditions(t *testing.T) {
	t.Run("Concurrent array validation is thread-safe", func(t *testing.T) {
		// Create an array validator with a string element schema
		elementSchema := String().Min(3).Required()
		arraySchema := Array(elementSchema).Required()

		// Test data
		validArray := []interface{}{"test1", "test2", "test3"}
		invalidArray := []interface{}{"ab", "test2", "cd"} // Two elements too short

		var wg sync.WaitGroup
		numGoroutines := 50
		errors := make([]error, numGoroutines*2)

		// Run concurrent validations
		for i := 0; i < numGoroutines; i++ {
			wg.Add(2)

			go func(index int) {
				defer wg.Done()
				errors[index*2] = arraySchema.Validate(validArray)
			}(i)

			go func(index int) {
				defer wg.Done()
				errors[index*2+1] = arraySchema.Validate(invalidArray)
			}(i)
		}

		wg.Wait()

		// Check results
		for i := 0; i < numGoroutines; i++ {
			// Valid array should pass
			if errors[i*2] != nil {
				t.Errorf("Valid array failed validation in goroutine %d: %v", i, errors[i*2])
			}

			// Invalid array should fail
			if errors[i*2+1] == nil {
				t.Errorf("Invalid array passed validation in goroutine %d", i)
			}
		}
	})

	t.Run("Schema copying prevents race conditions", func(t *testing.T) {
		// This test ensures that schema copying in validateElement works correctly
		baseSchema := String().Min(5).Required()
		arraySchema := Array(baseSchema).Required()

		testArrays := [][]interface{}{
			{"hello", "world", "testing"},        // All >= 5 chars: should pass
			{"short", "verylongstring", "hi"},    // "hi" < 5 chars: should fail
			{"test1", "test2", "test3", "test4"}, // All >= 5 chars: should pass
		}

		var wg sync.WaitGroup
		results := make([]error, len(testArrays))

		for i, testArray := range testArrays {
			wg.Add(1)
			go func(index int, array []interface{}) {
				defer wg.Done()
				results[index] = arraySchema.Validate(array)
			}(i, testArray)
		}

		wg.Wait()

		// Verify results are consistent
		expectedResults := []bool{true, false, true} // Based on min length 5
		for i, err := range results {
			hasError := err != nil
			expectError := !expectedResults[i]

			if hasError != expectError {
				t.Errorf("Array %d: expected error=%v, got error=%v (%v)",
					i, expectError, hasError, err)
			}
		}
	})
}

// TestEdgeCaseValues tests validation with extreme or edge case values
func TestEdgeCaseValues(t *testing.T) {
	t.Run("Number validation with special float values", func(t *testing.T) {
		schema := Number().Required()

		testCases := []struct {
			name     string
			value    interface{}
			hasError bool
		}{
			{"positive infinity", math.Inf(1), false},
			{"negative infinity", math.Inf(-1), false},
			{"NaN", math.NaN(), false}, // NaN is still a float64
			{"very large number", 1.7976931348623157e+308, false},
			{"very small number", 4.9406564584124654e-324, false},
			{"zero", 0.0, false},
			{"negative zero", math.Copysign(0, -1), false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := schema.Validate(tc.value)
				hasError := err != nil

				if hasError != tc.hasError {
					t.Errorf("Value %v: expected error=%v, got error=%v (%v)",
						tc.value, tc.hasError, hasError, err)
				}
			})
		}
	})

	t.Run("Number constraints with special values", func(t *testing.T) {
		t.Run("Min/Max with infinity", func(t *testing.T) {
			schema := Number().Min(math.Inf(-1)).Max(math.Inf(1)).Required()

			err := schema.Validate(1000000.0)
			if err != nil {
				t.Errorf("Regular number should pass with infinite bounds: %v", err)
			}
		})

		t.Run("Integer validation with special values", func(t *testing.T) {
			schema := Number().Integer().Required()

			testCases := []struct {
				name     string
				value    float64
				hasError bool
			}{
				{"integer as float", 42.0, false},
				{"non-integer", 42.5, true},
				{"very large integer", 9007199254740992.0, false}, // 2^53
				{"infinity", math.Inf(1), false},                  // Infinity might be accepted as valid float64
				{"NaN", math.NaN(), true},                         // NaN is not an integer
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					err := schema.Validate(tc.value)
					hasError := err != nil

					if hasError != tc.hasError {
						t.Errorf("Value %v: expected error=%v, got error=%v (%v)",
							tc.value, tc.hasError, hasError, err)
					}
				})
			}
		})
	})

	t.Run("String validation with extreme Unicode", func(t *testing.T) {
		schema := String().Min(1).Max(10).Required()

		testCases := []struct {
			name     string
			value    string
			hasError bool
		}{
			{"emoji", "🚀", false},
			{"multi-byte unicode", "你好", false},
			{"combining characters", "e\u0301", false},     // e with acute accent
			{"zero width characters", "test\u200B", false}, // Zero-width space
			{"surrogate pairs", "𝓗𝓮𝓵𝓵𝓸", true},             // These are longer than 10 characters
			{"long emoji sequence", "👨‍👩‍👧‍👦", true},       // This is longer than 10 characters
			{"empty string", "", true},                     // Below minimum
			{"very long unicode", "🚀🚀🚀🚀🚀🚀🚀🚀🚀🚀🚀", true},     // Above maximum
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := schema.Validate(tc.value)
				hasError := err != nil

				if hasError != tc.hasError {
					t.Errorf("Value '%s': expected error=%v, got error=%v (%v)",
						tc.value, tc.hasError, hasError, err)
				}
			})
		}
	})
}

// TestCustomValidationFunctions tests custom validation functions and error handling
func TestCustomValidationFunctions(t *testing.T) {
	t.Run("String custom validation", func(t *testing.T) {
		// Custom validator that rejects strings containing "bad"
		customValidator := func(s string) error {
			if regexp.MustCompile("bad").MatchString(s) {
				return errors.New("string contains forbidden word")
			}
			return nil
		}

		schema := String().Custom(customValidator).Required()

		// Test valid string
		err := schema.Validate("good string")
		if err != nil {
			t.Errorf("Valid string should pass custom validation: %v", err)
		}

		// Test invalid string
		err = schema.Validate("bad string")
		if err == nil {
			t.Error("Invalid string should fail custom validation")
		}
		if err != nil && err.Error() != "string contains forbidden word" {
			t.Errorf("Expected custom error message, got: %v", err.Error())
		}
	})

	t.Run("Number custom validation", func(t *testing.T) {
		// Custom validator that only allows prime numbers
		isPrime := func(n float64) error {
			if n != float64(int(n)) || n < 2 {
				return errors.New("must be a prime number")
			}

			num := int(n)
			for i := 2; i*i <= num; i++ {
				if num%i == 0 {
					return errors.New("must be a prime number")
				}
			}
			return nil
		}

		schema := Number().Custom(isPrime).Required()

		// Test prime numbers
		primes := []float64{2, 3, 5, 7, 11, 13}
		for _, prime := range primes {
			err := schema.Validate(prime)
			if err != nil {
				t.Errorf("Prime number %v should pass validation: %v", prime, err)
			}
		}

		// Test non-prime numbers
		nonPrimes := []float64{1, 4, 6, 8, 9, 10}
		for _, nonPrime := range nonPrimes {
			err := schema.Validate(nonPrime)
			if err == nil {
				t.Errorf("Non-prime number %v should fail validation", nonPrime)
			}
		}
	})
}

// TestAdvancedErrorMessageCustomization tests custom error message functionality
func TestAdvancedErrorMessageCustomization(t *testing.T) {
	t.Run("All error types can be customized", func(t *testing.T) {
		// Test different error types with separate schemas to avoid conflict
		testCases := []struct {
			name          string
			schema        func() RequiredNumberBuilder
			value         interface{}
			expectedError string
		}{
			{
				"below minimum",
				func() RequiredNumberBuilder {
					return Number().Min(10).WithMinMessage("Custom min error").Required()
				},
				5.0,
				"Field: 5, Error: Custom min error",
			},
			{
				"above maximum",
				func() RequiredNumberBuilder {
					return Number().Max(100).WithMaxMessage("Custom max error").Required()
				},
				150.0,
				"Field: 150, Error: Custom max error",
			},
			{
				"not integer",
				func() RequiredNumberBuilder {
					return Number().Integer().WithIntegerMessage("Custom integer error").Required()
				},
				50.5,
				"Field: 50.5, Error: Custom integer error",
			},
			{
				"negative number",
				func() RequiredNumberBuilder {
					return Number().Positive().WithPositiveMessage("Custom positive error").Required()
				},
				-5.0,
				"Field: -5, Error: Custom positive error",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				schema := tc.schema()
				err := schema.Validate(tc.value)
				if err == nil {
					t.Errorf("Expected validation to fail for %v", tc.value)
					return
				}

				if err.Error() != tc.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tc.expectedError, err.Error())
				}
			})
		}
	})

	t.Run("Default error messages work when no custom message", func(t *testing.T) {
		schema := Number().Min(10).Required()

		err := schema.Validate(5.0)
		if err == nil {
			t.Error("Expected validation to fail")
			return
		}

		// Should contain some indication of minimum value constraint
		errorStr := err.Error()
		if errorStr == "" {
			t.Error("Error message should not be empty")
		}
	})
}
