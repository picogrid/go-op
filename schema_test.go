package goop

import (
	"sync"
	"testing"
	"time"
)

// MockSchema implements Schema interface for testing
type MockSchema struct {
	ValidateFunc func(data interface{}) error
	CallCount    int
	mu           sync.Mutex
}

func (m *MockSchema) Validate(data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CallCount++
	if m.ValidateFunc != nil {
		return m.ValidateFunc(data)
	}
	return nil
}

func (m *MockSchema) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.CallCount
}

// TestValidateSchema tests the ValidateSchema helper function
func TestValidateSchema(t *testing.T) {
	t.Run("ValidateSchema with valid data", func(t *testing.T) {
		schema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				if data == "valid" {
					return nil
				}
				return NewValidationError("data", data, "Invalid data")
			},
		}

		err := ValidateSchema(schema, "valid")
		if err != nil {
			t.Errorf("Expected no error for valid data, got %v", err)
		}
		if schema.GetCallCount() != 1 {
			t.Errorf("Expected schema to be called once, got %d calls", schema.GetCallCount())
		}
	})

	t.Run("ValidateSchema with invalid data", func(t *testing.T) {
		schema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				return NewValidationError("data", data, "Data is invalid")
			},
		}

		err := ValidateSchema(schema, "invalid")
		if err == nil {
			t.Error("Expected error for invalid data")
		}

		validationErr, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		}
		if validationErr.Message != "Data is invalid" {
			t.Errorf("Expected message 'Data is invalid', got '%s'", validationErr.Message)
		}
	})

	t.Run("ValidateSchema with nil schema", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when calling ValidateSchema with nil schema")
			}
		}()
		ValidateSchema(nil, "data")
	})
}

// TestValidator tests the Validator goroutine function
func TestValidator(t *testing.T) {
	t.Run("Validator with valid data", func(t *testing.T) {
		schema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				return nil
			},
		}

		results := make(chan Result, 1)
		var wg sync.WaitGroup
		wg.Add(1)

		go Validator(schema, "valid", results, &wg)
		wg.Wait()
		close(results)

		result := <-results
		if !result.IsValid {
			t.Error("Expected result to be valid")
		}
		if result.Error != nil {
			t.Errorf("Expected no error, got %v", result.Error)
		}
	})

	t.Run("Validator with invalid data", func(t *testing.T) {
		expectedError := NewValidationError("field", "invalid", "Invalid data")
		schema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				return expectedError
			},
		}

		results := make(chan Result, 1)
		var wg sync.WaitGroup
		wg.Add(1)

		go Validator(schema, "invalid", results, &wg)
		wg.Wait()
		close(results)

		result := <-results
		if result.IsValid {
			t.Error("Expected result to be invalid")
		}
		if result.Error != expectedError {
			t.Errorf("Expected error %v, got %v", expectedError, result.Error)
		}
	})

	t.Run("Validator WaitGroup coordination", func(t *testing.T) {
		schema := &MockSchema{}
		results := make(chan Result, 3)
		var wg sync.WaitGroup

		// Start multiple validators
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go Validator(schema, i, results, &wg)
		}

		// Wait for all to complete
		done := make(chan bool)
		go func() {
			wg.Wait()
			done <- true
		}()

		// Should complete within reasonable time
		select {
		case <-done:
			// Good
		case <-time.After(1 * time.Second):
			t.Error("Validators did not complete within expected time")
		}

		close(results)

		// Count results
		resultCount := 0
		for range results {
			resultCount++
		}
		if resultCount != 3 {
			t.Errorf("Expected 3 results, got %d", resultCount)
		}
	})
}

// TestValidateConcurrently tests concurrent validation
func TestValidateConcurrently(t *testing.T) {
	t.Run("Concurrent validation with all valid data", func(t *testing.T) {
		schema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				// Simulate some work
				time.Sleep(10 * time.Millisecond)
				return nil
			},
		}

		dataList := []interface{}{"data1", "data2", "data3", "data4", "data5"}
		results := ValidateConcurrently(schema, dataList, 3)

		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}

		for i, result := range results {
			if !result.IsValid {
				t.Errorf("Expected result %d to be valid", i)
			}
			if result.Error != nil {
				t.Errorf("Expected no error for result %d, got %v", i, result.Error)
			}
		}

		// Verify schema was called for each data item
		if schema.GetCallCount() != 5 {
			t.Errorf("Expected schema to be called 5 times, got %d", schema.GetCallCount())
		}
	})

	t.Run("Concurrent validation with some invalid data", func(t *testing.T) {
		schema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				if data == "invalid" {
					return NewValidationError("data", data, "Invalid data")
				}
				return nil
			},
		}

		dataList := []interface{}{"valid1", "invalid", "valid2", "invalid", "valid3"}
		results := ValidateConcurrently(schema, dataList, 2)

		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}

		validCount := 0
		invalidCount := 0
		for _, result := range results {
			if result.IsValid {
				validCount++
				if result.Error != nil {
					t.Error("Valid result should not have error")
				}
			} else {
				invalidCount++
				if result.Error == nil {
					t.Error("Invalid result should have error")
				}
			}
		}

		if validCount != 3 {
			t.Errorf("Expected 3 valid results, got %d", validCount)
		}
		if invalidCount != 2 {
			t.Errorf("Expected 2 invalid results, got %d", invalidCount)
		}
	})

	t.Run("Concurrent validation with different worker counts", func(t *testing.T) {
		schema := &MockSchema{}
		dataList := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		// Test with different worker counts
		workerCounts := []int{1, 2, 5, 10, 20}
		for _, workerCount := range workerCounts {
			schema.CallCount = 0 // Reset call count
			results := ValidateConcurrently(schema, dataList, workerCount)

			if len(results) != 10 {
				t.Errorf("Worker count %d: Expected 10 results, got %d", workerCount, len(results))
			}

			if schema.GetCallCount() != 10 {
				t.Errorf("Worker count %d: Expected 10 schema calls, got %d", workerCount, schema.GetCallCount())
			}
		}
	})

	t.Run("Concurrent validation performance test", func(t *testing.T) {
		schema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				// Simulate some CPU work
				time.Sleep(1 * time.Millisecond)
				return nil
			},
		}

		dataList := make([]interface{}, 50)
		for i := range dataList {
			dataList[i] = i
		}

		// Test sequential vs concurrent performance
		start := time.Now()
		results := ValidateConcurrently(schema, dataList, 10)
		concurrentDuration := time.Since(start)

		if len(results) != 50 {
			t.Errorf("Expected 50 results, got %d", len(results))
		}

		// Should complete much faster than sequential (50 * 1ms = 50ms)
		// With 10 workers, should be closer to 5-10ms
		if concurrentDuration > 100*time.Millisecond {
			t.Errorf("Concurrent validation took too long: %v", concurrentDuration)
		}
	})

	t.Run("Concurrent validation with empty data list", func(t *testing.T) {
		schema := &MockSchema{}
		dataList := []interface{}{}

		results := ValidateConcurrently(schema, dataList, 5)

		if len(results) != 0 {
			t.Errorf("Expected 0 results for empty data list, got %d", len(results))
		}
		if schema.GetCallCount() != 0 {
			t.Errorf("Expected 0 schema calls for empty data list, got %d", schema.GetCallCount())
		}
	})

	t.Run("Concurrent validation with single item", func(t *testing.T) {
		schema := &MockSchema{}
		dataList := []interface{}{"single"}

		results := ValidateConcurrently(schema, dataList, 5)

		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if !results[0].IsValid {
			t.Error("Expected single result to be valid")
		}
	})

	t.Run("Concurrent validation with many workers", func(t *testing.T) {
		schema := &MockSchema{}
		dataList := []interface{}{"data1", "data2"}

		// More workers than data should still work
		results := ValidateConcurrently(schema, dataList, 10)

		if len(results) != 2 {
			t.Errorf("Expected 2 results with many workers, got %d", len(results))
		}
	})
}

// TestResult tests the Result struct
func TestResult(t *testing.T) {
	t.Run("Result with valid data", func(t *testing.T) {
		result := Result{IsValid: true, Error: nil}

		if !result.IsValid {
			t.Error("Expected result to be valid")
		}
		if result.Error != nil {
			t.Errorf("Expected no error, got %v", result.Error)
		}
	})

	t.Run("Result with invalid data", func(t *testing.T) {
		err := NewValidationError("field", "value", "error")
		result := Result{IsValid: false, Error: err}

		if result.IsValid {
			t.Error("Expected result to be invalid")
		}
		if result.Error != err {
			t.Errorf("Expected error %v, got %v", err, result.Error)
		}
	})
}

// TestSchemaIntegration tests integration scenarios
func TestSchemaIntegration(t *testing.T) {
	t.Run("Integration with real validation scenarios", func(t *testing.T) {
		// Create a schema that validates user data
		userSchema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				user, ok := data.(map[string]interface{})
				if !ok {
					return NewValidationError("user", data, "User must be an object")
				}

				var errors []ValidationError

				// Validate name
				name, hasName := user["name"]
				if !hasName {
					errors = append(errors, *NewValidationError("name", nil, "Name is required"))
				} else if nameStr, ok := name.(string); !ok || len(nameStr) < 2 {
					errors = append(errors, *NewValidationError("name", name, "Name must be at least 2 characters"))
				}

				// Validate age
				age, hasAge := user["age"]
				if !hasAge {
					errors = append(errors, *NewValidationError("age", nil, "Age is required"))
				} else if ageInt, ok := age.(int); !ok || ageInt < 0 || ageInt > 120 {
					errors = append(errors, *NewValidationError("age", age, "Age must be between 0 and 120"))
				}

				if len(errors) > 0 {
					return NewNestedValidationError("user", data, "User validation failed", errors)
				}
				return nil
			},
		}

		// Test valid user
		validUser := map[string]interface{}{
			"name": "John Doe",
			"age":  30,
		}

		err := ValidateSchema(userSchema, validUser)
		if err != nil {
			t.Errorf("Expected valid user to pass validation, got %v", err)
		}

		// Test invalid user
		invalidUser := map[string]interface{}{
			"name": "J",
			"age":  -5,
		}

		err = ValidateSchema(userSchema, invalidUser)
		if err == nil {
			t.Error("Expected invalid user to fail validation")
		}

		// Check that it's a nested validation error
		if nestedErr, ok := err.(*ValidationError); ok {
			if nestedErr.ErrorType != "Nested Validation Error" {
				t.Errorf("Expected nested validation error, got %s", nestedErr.ErrorType)
			}
			if len(nestedErr.Details) != 2 {
				t.Errorf("Expected 2 validation errors, got %d", len(nestedErr.Details))
			}
		} else {
			t.Errorf("Expected ValidationError, got %T", err)
		}
	})

	t.Run("Schema composition", func(t *testing.T) {
		// Create schemas that can be composed
		nameSchema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				if name, ok := data.(string); !ok || len(name) < 2 {
					return NewValidationError("name", data, "Invalid name")
				}
				return nil
			},
		}

		ageSchema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				if age, ok := data.(int); !ok || age < 0 {
					return NewValidationError("age", data, "Invalid age")
				}
				return nil
			},
		}

		// Composite validation using multiple schemas
		validateUser := func(user map[string]interface{}) error {
			if name, hasName := user["name"]; hasName {
				if err := ValidateSchema(nameSchema, name); err != nil {
					return err
				}
			}
			if age, hasAge := user["age"]; hasAge {
				if err := ValidateSchema(ageSchema, age); err != nil {
					return err
				}
			}
			return nil
		}

		// Test composite validation
		validUser := map[string]interface{}{"name": "John", "age": 25}
		if err := validateUser(validUser); err != nil {
			t.Errorf("Expected valid user to pass composite validation, got %v", err)
		}

		invalidUser := map[string]interface{}{"name": "J", "age": 25}
		if err := validateUser(invalidUser); err == nil {
			t.Error("Expected invalid user to fail composite validation")
		}
	})

	t.Run("Concurrent validation stress test", func(t *testing.T) {
		// Create a schema that does real work
		complexSchema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				// Simulate complex validation logic
				time.Sleep(1 * time.Millisecond)

				if data == nil {
					return NewValidationError("data", data, "Data cannot be nil")
				}

				if str, ok := data.(string); ok && len(str) == 0 {
					return NewValidationError("data", data, "String cannot be empty")
				}

				return nil
			},
		}

		// Generate large dataset
		dataList := make([]interface{}, 1000)
		for i := range dataList {
			if i%10 == 0 {
				dataList[i] = nil // 10% invalid data (every 10th: 0, 10, 20, 30, ...)
			} else if i%25 == 1 {
				dataList[i] = "" // 4% empty strings (1, 26, 51, 76, ...)
			} else {
				dataList[i] = "valid data"
			}
		}

		start := time.Now()
		results := ValidateConcurrently(complexSchema, dataList, 50)
		duration := time.Since(start)

		t.Logf("Processed %d items in %v", len(dataList), duration)

		if len(results) != len(dataList) {
			t.Errorf("Expected %d results, got %d", len(dataList), len(results))
		}

		validCount := 0
		invalidCount := 0
		for _, result := range results {
			if result.IsValid {
				validCount++
			} else {
				invalidCount++
			}
		}

		expectedValid := 860   // 86% should be valid (1000 - 100 nil - 40 empty = 860)
		expectedInvalid := 140 // 14% should be invalid (100 nil + 40 empty = 140)

		if validCount != expectedValid {
			t.Errorf("Expected %d valid results, got %d", expectedValid, validCount)
		}
		if invalidCount != expectedInvalid {
			t.Errorf("Expected %d invalid results, got %d", expectedInvalid, invalidCount)
		}
	})
}
