package operations

import (
	"testing"

	goop "github.com/picogrid/go-op"
	"github.com/picogrid/go-op/validators"
)

func TestMultipleResponsesBuilder(t *testing.T) {
	// Test WithCreateErrors convenience method
	op := NewSimple().
		POST("/test").
		WithSuccessResponse(201, validators.String().Required(), "Created").
		WithCreateErrors().
		Handler(nil)

	// Should have multiple response codes defined
	if len(op.Responses) == 0 {
		t.Error("Expected multiple responses to be defined, but got none")
	}

	// Check specific status codes
	expectedCodes := []int{201, 400, 401, 403, 409, 422, 500}
	for _, code := range expectedCodes {
		if _, exists := op.Responses[code]; !exists {
			t.Errorf("Expected response code %d to be defined", code)
		}
	}

	t.Logf("Defined response codes: %v", getResponseCodes(op.Responses))
}

func TestMultipleResponsesIndividual(t *testing.T) {
	// Test individual response definitions
	op := NewSimple().
		GET("/test/{id}").
		WithSuccessResponse(200, validators.String().Required(), "Success").
		WithBadRequestError(BadRequestErrorSchema).
		WithNotFoundError(NotFoundErrorSchema).
		WithServerError(InternalServerErrorSchema).
		Handler(nil)

	expectedCodes := []int{200, 400, 404, 500}
	for _, code := range expectedCodes {
		if _, exists := op.Responses[code]; !exists {
			t.Errorf("Expected response code %d to be defined", code)
		}
	}

	// Verify descriptions are set correctly
	if resp, exists := op.Responses[200]; !exists || resp.Description != "Success" {
		t.Error("Expected 200 response with 'Success' description")
	}
	if resp, exists := op.Responses[404]; !exists || resp.Description != "Not Found" {
		t.Error("Expected 404 response with 'Not Found' description")
	}

	t.Logf("Defined response codes: %v", getResponseCodes(op.Responses))
}

func TestStandardErrorsByCode(t *testing.T) {
	// Test WithStandardErrorsByCode
	op := NewSimple().
		GET("/test").
		WithSuccessResponse(200, validators.String().Required(), "OK").
		WithStandardErrorsByCode(400, 401, 404, 500).
		Handler(nil)

	expectedCodes := []int{200, 400, 401, 404, 500}
	for _, code := range expectedCodes {
		if _, exists := op.Responses[code]; !exists {
			t.Errorf("Expected response code %d to be defined", code)
		}
	}

	t.Logf("Defined response codes: %v", getResponseCodes(op.Responses))
}

func TestErrorSchemaIntegration(t *testing.T) {
	// Test that error schemas are properly integrated
	op := NewSimple().
		POST("/users").
		WithSuccessResponse(201, validators.String().Required(), "Created").
		WithBadRequestError(BadRequestErrorSchema).
		WithUnauthorizedError(UnauthorizedErrorSchema).
		WithUnprocessableEntityError(ValidationErrorSchema).
		Handler(nil)

	// Test that schemas are properly set
	if resp, exists := op.Responses[400]; !exists {
		t.Error("Expected 400 response to be defined")
	} else if resp.Schema == nil {
		t.Error("Expected 400 response to have schema")
	} else {
		// Test schema validation with valid error data
		errorData := map[string]interface{}{
			"error":   "bad_request",
			"message": "Invalid request",
			"code":    400,
		}
		if err := resp.Schema.Validate(errorData); err != nil {
			t.Errorf("Valid error data should pass validation: %v", err)
		}
	}

	// Test unprocessable entity error schema specifically
	if resp, exists := op.Responses[422]; !exists {
		t.Error("Expected 422 response to be defined")
	} else if resp.Schema == nil {
		t.Error("Expected 422 response to have schema")
	} else {
		// Test with field validation errors
		validationData := map[string]interface{}{
			"error":   "validation_failed",
			"message": "Request validation failed",
			"fields": map[string]interface{}{
				"email": "Invalid email format",
			},
		}
		if err := resp.Schema.Validate(validationData); err != nil {
			t.Errorf("Valid validation error data should pass: %v", err)
		}
	}

	t.Logf("Error schema integration test passed")
}

func TestGetStandardErrorSchema(t *testing.T) {
	// Test GetStandardErrorSchema helper function
	tests := []struct {
		code     int
		expected goop.Schema
	}{
		{400, BadRequestErrorSchema},
		{401, UnauthorizedErrorSchema},
		{403, ForbiddenErrorSchema},
		{404, NotFoundErrorSchema},
		{409, ConflictErrorSchema},
		{422, UnprocessableEntityErrorSchema},
		{429, TooManyRequestsErrorSchema},
		{500, InternalServerErrorSchema},
		{502, BadGatewayErrorSchema},
		{503, ServiceUnavailableErrorSchema},
		{999, BadRequestErrorSchema}, // Default case
	}

	for _, test := range tests {
		result := GetStandardErrorSchema(test.code)
		if result != test.expected {
			t.Errorf("GetStandardErrorSchema(%d) returned unexpected schema", test.code)
		}
	}
}

func TestConvenienceErrorMethods(t *testing.T) {
	// Test all convenience error methods
	op := NewSimple().
		POST("/test").
		WithSuccessResponse(201, validators.String().Required(), "Created").
		WithAuthErrors().       // Adds 401, 403
		WithValidationErrors(). // Adds 400, 422
		WithCreateErrors().     // Adds 400, 401, 403, 409, 422, 500
		Handler(nil)

	// All these codes should be present
	expectedCodes := []int{201, 400, 401, 403, 409, 422, 500}
	for _, code := range expectedCodes {
		if _, exists := op.Responses[code]; !exists {
			t.Errorf("Expected response code %d to be defined after convenience methods", code)
		}
	}

	// Test that schemas are properly assigned
	if resp, exists := op.Responses[401]; exists && resp.Schema == nil {
		t.Error("Expected 401 response to have schema")
	}
	if resp, exists := op.Responses[422]; exists && resp.Schema == nil {
		t.Error("Expected 422 response to have schema")
	}

	t.Logf("Convenience methods test passed with codes: %v", getResponseCodes(op.Responses))
}

// Helper function to get response codes for logging
func getResponseCodes(responses map[int]goop.ResponseDefinition) []int {
	codes := make([]int, 0, len(responses))
	for code := range responses {
		codes = append(codes, code)
	}
	return codes
}
