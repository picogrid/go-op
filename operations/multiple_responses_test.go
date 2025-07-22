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

// Helper function to get response codes for logging
func getResponseCodes(responses map[int]goop.ResponseDefinition) []int {
	codes := make([]int, 0, len(responses))
	for code := range responses {
		codes = append(codes, code)
	}
	return codes
}
