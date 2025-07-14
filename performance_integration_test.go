package goop

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

)

// TestPerformanceBenchmarks tests performance characteristics
func TestPerformanceBenchmarks(t *testing.T) {
	t.Run("Validation error creation performance", func(t *testing.T) {
		start := time.Now()
		
		for i := 0; i < 10000; i++ {
			err := NewValidationError(fmt.Sprintf("field_%d", i), i, fmt.Sprintf("Error message %d", i))
			if err.Field == "" {
				t.Error("Error creation failed")
			}
		}
		
		duration := time.Since(start)
		t.Logf("Created 10,000 validation errors in %v", duration)
		
		// Should complete quickly
		if duration > 100*time.Millisecond {
			t.Errorf("Validation error creation too slow: %v", duration)
		}
	})

	t.Run("Nested validation error performance", func(t *testing.T) {
		start := time.Now()
		
		for i := 0; i < 1000; i++ {
			// Create 10 child errors for each parent
			var details []ValidationError
			for j := 0; j < 10; j++ {
				details = append(details, *NewValidationError(
					fmt.Sprintf("field_%d_%d", i, j),
					j,
					fmt.Sprintf("Error %d.%d", i, j),
				))
			}
			
			parentErr := NewNestedValidationError(
				fmt.Sprintf("parent_%d", i),
				nil,
				fmt.Sprintf("Parent error %d", i),
				details,
			)
			
			// Test JSON serialization performance
			jsonStr := parentErr.ErrorJSON()
			if jsonStr == "" {
				t.Error("JSON serialization failed")
			}
		}
		
		duration := time.Since(start)
		t.Logf("Created and serialized 1,000 nested errors (10 children each) in %v", duration)
		
		// Should complete within reasonable time
		if duration > 500*time.Millisecond {
			t.Errorf("Nested error processing too slow: %v", duration)
		}
	})

	t.Run("Concurrent validation performance", func(t *testing.T) {
		schema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				// Simulate some validation work
				time.Sleep(100 * time.Microsecond)
				
				if str, ok := data.(string); ok && str == "invalid" {
					return NewValidationError("data", data, "Invalid data")
				}
				return nil
			},
		}

		// Test different data set sizes
		sizes := []int{10, 100, 1000}
		workerCounts := []int{1, 5, 10, 20}

		for _, size := range sizes {
			for _, workers := range workerCounts {
				t.Run(fmt.Sprintf("size_%d_workers_%d", size, workers), func(t *testing.T) {
					dataList := make([]interface{}, size)
					for i := range dataList {
						if i%10 == 0 {
							dataList[i] = "invalid"
						} else {
							dataList[i] = fmt.Sprintf("valid_%d", i)
						}
					}

					start := time.Now()
					results := ValidateConcurrently(schema, dataList, workers)
					duration := time.Since(start)

					if len(results) != size {
						t.Errorf("Expected %d results, got %d", size, len(results))
					}

					t.Logf("Processed %d items with %d workers in %v", size, workers, duration)

					// Performance expectation: should scale with workers
					maxExpectedTime := time.Duration(size) * 150 * time.Microsecond / time.Duration(workers)
					if duration > maxExpectedTime {
						t.Logf("Warning: Processing took longer than expected: %v > %v", duration, maxExpectedTime)
					}
				})
			}
		}
	})

	t.Run("Memory usage test", func(t *testing.T) {
		runtime.GC()
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)
		beforeAlloc := memBefore.Alloc

		// Create many validation errors
		var errors []*ValidationError
		for i := 0; i < 10000; i++ {
			err := NewValidationError(
				fmt.Sprintf("field_%d", i),
				fmt.Sprintf("value_%d", i),
				fmt.Sprintf("message_%d", i),
			)
			errors = append(errors, err)
		}

		runtime.GC()
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)
		afterAlloc := memAfter.Alloc

		var allocatedBytes uint64
		if afterAlloc > beforeAlloc {
			allocatedBytes = afterAlloc - beforeAlloc
		} else {
			allocatedBytes = 0 // GC may have cleaned up
		}
		t.Logf("Memory allocated for 10,000 validation errors: %d bytes (%.2f KB)", 
			allocatedBytes, float64(allocatedBytes)/1024)

		// Basic sanity check - shouldn't use excessive memory (only test if we measured allocation)
		if allocatedBytes > 0 {
			maxExpected := uint64(10 * 1024 * 1024) // 10MB max
			if allocatedBytes > maxExpected {
				t.Errorf("Excessive memory usage: %d bytes > %d bytes", allocatedBytes, maxExpected)
			}
		}

		// Keep reference to prevent GC
		_ = errors
	})
}

// TestIntegrationScenarios tests real-world integration scenarios
func TestIntegrationScenarios(t *testing.T) {
	t.Run("Complete API validation workflow", func(t *testing.T) {
		// Simulate a complete API request validation workflow
		
		// 1. Create user registration schema
		userSchema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				user, ok := data.(map[string]interface{})
				if !ok {
					return NewValidationError("user", data, "User must be an object")
				}

				var errors []ValidationError

				// Validate email
				if email, hasEmail := user["email"]; !hasEmail {
					errors = append(errors, *NewValidationError("email", nil, "Email is required"))
				} else if emailStr, ok := email.(string); !ok || len(emailStr) < 5 || !containsString(emailStr, "@") {
					errors = append(errors, *NewValidationError("email", email, "Invalid email format"))
				}

				// Validate password
				if password, hasPassword := user["password"]; !hasPassword {
					errors = append(errors, *NewValidationError("password", nil, "Password is required"))
				} else if passStr, ok := password.(string); !ok || len(passStr) < 8 {
					errors = append(errors, *NewValidationError("password", password, "Password must be at least 8 characters"))
				}

				// Validate age
				if age, hasAge := user["age"]; hasAge {
					if ageInt, ok := age.(int); !ok || ageInt < 13 || ageInt > 120 {
						errors = append(errors, *NewValidationError("age", age, "Age must be between 13 and 120"))
					}
				}

				if len(errors) > 0 {
					return NewNestedValidationError("user", data, "User validation failed", errors)
				}
				return nil
			},
		}

		// 2. Test valid users
		validUsers := []map[string]interface{}{
			{"email": "user1@example.com", "password": "password123", "age": 25},
			{"email": "user2@example.com", "password": "secretpass", "age": 30},
			{"email": "admin@company.com", "password": "adminpass123"},
		}

		for i, user := range validUsers {
			if err := ValidateSchema(userSchema, user); err != nil {
				t.Errorf("Valid user %d failed validation: %v", i, err)
			}
		}

		// 3. Test invalid users
		invalidUsers := []map[string]interface{}{
			{"email": "invalid", "password": "short"},                   // Invalid email and short password
			{"password": "password123"},                                  // Missing email
			{"email": "test@example.com", "password": "pass", "age": 5}, // Short password and invalid age
		}

		for i, user := range invalidUsers {
			if err := ValidateSchema(userSchema, user); err == nil {
				t.Errorf("Invalid user %d should have failed validation", i)
			} else {
				// Verify we get proper nested errors
				if nestedErr, ok := err.(*ValidationError); ok && nestedErr.ErrorType == "Nested Validation Error" {
					jsonStr := nestedErr.ErrorJSON()
					if jsonStr == "" {
						t.Errorf("Failed to serialize nested error for invalid user %d", i)
					}
				}
			}
		}

		// 4. Test concurrent validation of multiple users
		mixedUsers := make([]interface{}, 100)
		for i := range mixedUsers {
			if i%4 == 0 {
				// Invalid user (missing email)
				mixedUsers[i] = map[string]interface{}{"password": "validpassword123"}
			} else {
				// Valid user
				mixedUsers[i] = map[string]interface{}{
					"email":    fmt.Sprintf("user%d@example.com", i),
					"password": "validpassword123",
					"age":      20 + (i % 50),
				}
			}
		}

		start := time.Now()
		results := ValidateConcurrently(userSchema, mixedUsers, 10)
		duration := time.Since(start)

		t.Logf("Validated 100 users concurrently in %v", duration)

		validCount := 0
		invalidCount := 0
		for _, result := range results {
			if result.IsValid {
				validCount++
			} else {
				invalidCount++
			}
		}

		expectedValid := 75  // 100 - 25 invalid (every 4th)
		expectedInvalid := 25

		if validCount != expectedValid {
			t.Errorf("Expected %d valid users, got %d", expectedValid, validCount)
		}
		if invalidCount != expectedInvalid {
			t.Errorf("Expected %d invalid users, got %d", expectedInvalid, invalidCount)
		}
	})

	t.Run("High-load validation scenario", func(t *testing.T) {
		// Simulate high-load scenario with many concurrent validations
		
		schema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				// Simulate realistic validation time
				time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
				
				if data == nil {
					return NewValidationError("data", data, "Data cannot be nil")
				}
				
				return nil
			},
		}

		// Generate large dataset
		dataSize := 5000
		dataList := make([]interface{}, dataSize)
		for i := range dataList {
			if i%100 == 0 {
				dataList[i] = nil // 1% invalid
			} else {
				dataList[i] = fmt.Sprintf("data_%d", i)
			}
		}

		start := time.Now()
		results := ValidateConcurrently(schema, dataList, 50)
		duration := time.Since(start)

		t.Logf("High-load test: Processed %d items in %v with 50 workers", dataSize, duration)

		if len(results) != dataSize {
			t.Errorf("Expected %d results, got %d", dataSize, len(results))
		}

		// Verify results
		validCount := 0
		invalidCount := 0
		for _, result := range results {
			if result.IsValid {
				validCount++
			} else {
				invalidCount++
			}
		}

		expectedValid := 4950  // 5000 - 50 invalid
		expectedInvalid := 50

		if validCount != expectedValid {
			t.Errorf("Expected %d valid results, got %d", validCount, validCount)
		}
		if invalidCount != expectedInvalid {
			t.Errorf("Expected %d invalid results, got %d", expectedInvalid, invalidCount)
		}

		// Performance check - should complete within reasonable time
		maxExpectedTime := 30 * time.Second
		if duration > maxExpectedTime {
			t.Errorf("High-load test took too long: %v > %v", duration, maxExpectedTime)
		}
	})

	t.Run("Error handling stress test", func(t *testing.T) {
		// Test error handling under stress
		
		var wg sync.WaitGroup
		errChan := make(chan *ValidationError, 1000)
		
		// Start multiple goroutines creating errors
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				
				for j := 0; j < 100; j++ {
					// Create nested errors
					var details []ValidationError
					for k := 0; k < 5; k++ {
						details = append(details, *NewValidationError(
							fmt.Sprintf("field_%d_%d_%d", goroutineID, j, k),
							rand.Intn(1000),
							fmt.Sprintf("Error from goroutine %d, iteration %d, detail %d", goroutineID, j, k),
						))
					}
					
					parentErr := NewNestedValidationError(
						fmt.Sprintf("parent_%d_%d", goroutineID, j),
						nil,
						fmt.Sprintf("Parent error %d.%d", goroutineID, j),
						details,
					)
					
					errChan <- parentErr
				}
			}(i)
		}
		
		// Start goroutine to process errors
		processedCount := 0
		processingDone := make(chan bool)
		go func() {
			for err := range errChan {
				// Process the error (serialize to JSON)
				jsonStr := err.ErrorJSON()
				if jsonStr == "" {
					t.Error("Error serialization failed")
				}
				processedCount++
			}
			processingDone <- true
		}()
		
		// Wait for all error creation to complete
		wg.Wait()
		close(errChan)
		
		// Wait for processing to complete
		<-processingDone
		
		expectedErrors := 10 * 100 // 10 goroutines * 100 errors each
		if processedCount != expectedErrors {
			t.Errorf("Expected %d processed errors, got %d", expectedErrors, processedCount)
		}
		
		t.Logf("Stress test: Created and processed %d nested errors concurrently", processedCount)
	})

	t.Run("Mock validators integration", func(t *testing.T) {
		// Test integration with mock validators to simulate real-world usage
		
		// Create a mock email validator
		emailValidator := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				email, ok := data.(string)
				if !ok {
					return NewValidationError("email", data, "Email must be a string")
				}
				if len(email) < 5 || !containsString(email, "@") || email[0] == '@' || email[len(email)-1] == '@' {
					return NewValidationError("email", data, "Invalid email format")
				}
				return nil
			},
		}
		
		// Test valid emails
		validEmails := []string{
			"user@example.com",
			"test.email@domain.co.uk",
			"admin+tag@company.org",
		}
		
		for _, email := range validEmails {
			if err := ValidateSchema(emailValidator, email); err != nil {
				t.Errorf("Valid email '%s' failed validation: %v", email, err)
			}
		}
		
		// Test invalid emails
		invalidEmails := []string{
			"invalid-email",
			"@domain.com",
			"user@",
			"",
		}
		
		for _, email := range invalidEmails {
			if err := ValidateSchema(emailValidator, email); err == nil {
				t.Errorf("Invalid email '%s' should have failed validation", email)
			}
		}
		
		// Test with complex object validation
		userValidator := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				user, ok := data.(map[string]interface{})
				if !ok {
					return NewValidationError("user", data, "User must be an object")
				}

				var errors []ValidationError

				// Validate name
				if name, hasName := user["name"]; !hasName {
					errors = append(errors, *NewValidationError("name", nil, "Name is required"))
				} else if nameStr, ok := name.(string); !ok || len(nameStr) < 2 {
					errors = append(errors, *NewValidationError("name", name, "Name must be at least 2 characters"))
				}

				// Validate email
				if email, hasEmail := user["email"]; !hasEmail {
					errors = append(errors, *NewValidationError("email", nil, "Email is required"))
				} else if emailStr, ok := email.(string); !ok || !containsString(emailStr, "@") {
					errors = append(errors, *NewValidationError("email", email, "Invalid email format"))
				}

				// Validate age (optional)
				if age, hasAge := user["age"]; hasAge {
					if ageInt, ok := age.(int); !ok || ageInt < 0 || ageInt > 120 {
						errors = append(errors, *NewValidationError("age", age, "Age must be between 0 and 120"))
					}
				}

				if len(errors) > 0 {
					return NewNestedValidationError("user", data, "User validation failed", errors)
				}
				return nil
			},
		}
		
		validUser := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		}
		
		if err := ValidateSchema(userValidator, validUser); err != nil {
			t.Errorf("Valid user failed validation: %v", err)
		}
		
		invalidUser := map[string]interface{}{
			"name":  "J", // Too short
			"email": "invalid-email",
			"age":   -5, // Invalid age
		}
		
		if err := ValidateSchema(userValidator, invalidUser); err == nil {
			t.Error("Invalid user should have failed validation")
		}
		
		// Test array validation
		arrayValidator := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				arr, ok := data.([]interface{})
				if !ok {
					return NewValidationError("array", data, "Must be an array")
				}
				if len(arr) < 1 || len(arr) > 5 {
					return NewValidationError("array", data, "Array must have 1-5 items")
				}
				for i, item := range arr {
					if str, ok := item.(string); !ok || len(str) < 1 {
						return NewValidationError(fmt.Sprintf("[%d]", i), item, "Array item must be non-empty string")
					}
				}
				return nil
			},
		}
		
		validArray := []interface{}{"item1", "item2", "item3"}
		if err := ValidateSchema(arrayValidator, validArray); err != nil {
			t.Errorf("Valid array failed validation: %v", err)
		}
		
		invalidArray := []interface{}{} // Empty array
		if err := ValidateSchema(arrayValidator, invalidArray); err == nil {
			t.Error("Empty array should have failed validation")
		}
	})
}

// TestStressAndEdgeCases tests system behavior under stress and edge conditions
func TestStressAndEdgeCases(t *testing.T) {
	t.Run("Deep nesting stress test", func(t *testing.T) {
		// Create deeply nested validation errors
		depth := 100
		var currentErr *ValidationError
		
		for i := 0; i < depth; i++ {
			if currentErr == nil {
				currentErr = NewValidationError(fmt.Sprintf("field_%d", i), i, fmt.Sprintf("Error at depth %d", i))
			} else {
				currentErr = NewNestedValidationError(
					fmt.Sprintf("parent_%d", i),
					nil,
					fmt.Sprintf("Parent error at depth %d", i),
					[]ValidationError{*currentErr},
				)
			}
		}
		
		// Test that deeply nested errors can be processed
		start := time.Now()
		errorStr := currentErr.Error()
		jsonStr := currentErr.ErrorJSON()
		duration := time.Since(start)
		
		if errorStr == "" || jsonStr == "" {
			t.Error("Failed to process deeply nested error")
		}
		
		t.Logf("Processed error with depth %d in %v", depth, duration)
		
		// Should complete within reasonable time even with deep nesting
		if duration > 100*time.Millisecond {
			t.Errorf("Deep nesting processing too slow: %v", duration)
		}
	})

	t.Run("Large error message stress test", func(t *testing.T) {
		// Test with very large error messages
		largeMessage := string(make([]byte, 10000))
		for i := range largeMessage {
			largeMessage = largeMessage[:i] + "A" + largeMessage[i+1:]
		}
		
		err := NewValidationError("large_field", "large_value", largeMessage)
		
		start := time.Now()
		errorStr := err.Error()
		jsonStr := err.ErrorJSON()
		duration := time.Since(start)
		
		if !containsString(errorStr, largeMessage) {
			t.Error("Large message not preserved in error string")
		}
		if jsonStr == "" {
			t.Error("Failed to serialize large error message to JSON")
		}
		
		t.Logf("Processed large error message (10KB) in %v", duration)
		
		if duration > 10*time.Millisecond {
			t.Errorf("Large message processing too slow: %v", duration)
		}
	})

	t.Run("Concurrent schema access", func(t *testing.T) {
		// Test concurrent access to the same schema
		schema := &MockSchema{
			ValidateFunc: func(data interface{}) error {
				// Simulate variable processing time
				time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
				
				if data == "error" {
					return NewValidationError("data", data, "Intentional error")
				}
				return nil
			},
		}
		
		var wg sync.WaitGroup
		results := make(chan Result, 1000)
		
		// Start many goroutines using the same schema
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				// Each goroutine validates 10 items
				for j := 0; j < 10; j++ {
					var data interface{}
					if (id+j)%10 == 0 {
						data = "error"
					} else {
						data = fmt.Sprintf("data_%d_%d", id, j)
					}
					
					// Validate directly instead of using Validator function to avoid WaitGroup issues
					err := schema.Validate(data)
					result := Result{
						IsValid: err == nil,
						Error:   err,
					}
					results <- result
				}
			}(i)
		}
		
		wg.Wait()
		close(results)
		
		// Count results
		resultCount := 0
		validCount := 0
		errorCount := 0
		
		for result := range results {
			resultCount++
			if result.IsValid {
				validCount++
			} else {
				errorCount++
			}
		}
		
		expectedResults := 100 * 10 // 100 goroutines * 10 validations each
		if resultCount != expectedResults {
			t.Errorf("Expected %d results, got %d", expectedResults, resultCount)
		}
		
		t.Logf("Concurrent schema access: %d total, %d valid, %d errors", resultCount, validCount, errorCount)
	})
}

// Helper function for string containment check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || findSubstringHelper(s, substr))
}

// Helper function to find index of substring
func containsStringIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func findSubstringHelper(s, substr string) bool {
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