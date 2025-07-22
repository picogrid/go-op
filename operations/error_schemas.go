package operations

import (
	goop "github.com/picogrid/go-op"
	"github.com/picogrid/go-op/validators"
)

// StandardErrorResponse represents a standard API error response
type StandardErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ValidationErrorResponse represents a validation error with field details
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Code    int               `json:"code,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Common error response schemas that can be reused across operations
var (
	// BadRequestErrorSchema represents a 400 Bad Request response
	BadRequestErrorSchema = validators.Object(map[string]interface{}{
		"error": validators.String().
			Example("bad_request").
			Required(),
		"message": validators.String().
			Example("The request could not be understood or was missing required parameters").
			Required(),
		"code": validators.Number().
			Example(400).
			Optional(),
		"details": validators.String().
			Example("Invalid JSON format in request body").
			Optional(),
	}).Example(map[string]interface{}{
		"error":   "bad_request",
		"message": "The request could not be understood or was missing required parameters",
		"code":    400,
		"details": "Invalid JSON format in request body",
	}).Required()

	// ValidationErrorSchema represents a 400 Bad Request with validation errors
	ValidationErrorSchema = validators.Object(map[string]interface{}{
		"error": validators.String().
			Example("validation_failed").
			Required(),
		"message": validators.String().
			Example("Request validation failed").
			Required(),
		"code": validators.Number().
			Example(400).
			Optional(),
		"fields": validators.Object(map[string]interface{}{
			"email": validators.String().Optional(),
			"age":   validators.String().Optional(),
		}).
			Example(map[string]interface{}{
				"email": "Invalid email format",
				"age":   "Age must be between 13 and 120",
			}).
			Optional(),
	}).Example(map[string]interface{}{
		"error":   "validation_failed",
		"message": "Request validation failed",
		"code":    400,
		"fields": map[string]interface{}{
			"email": "Invalid email format",
			"age":   "Age must be between 13 and 120",
		},
	}).Required()

	// UnauthorizedErrorSchema represents a 401 Unauthorized response
	UnauthorizedErrorSchema = validators.Object(map[string]interface{}{
		"error": validators.String().
			Example("unauthorized").
			Required(),
		"message": validators.String().
			Example("Authentication is required to access this resource").
			Required(),
		"code": validators.Number().
			Example(401).
			Optional(),
		"details": validators.String().
			Example("Invalid or expired authentication token").
			Optional(),
	}).Example(map[string]interface{}{
		"error":   "unauthorized",
		"message": "Authentication is required to access this resource",
		"code":    401,
		"details": "Invalid or expired authentication token",
	}).Required()

	// ForbiddenErrorSchema represents a 403 Forbidden response
	ForbiddenErrorSchema = validators.Object(map[string]interface{}{
		"error": validators.String().
			Example("forbidden").
			Required(),
		"message": validators.String().
			Example("You do not have permission to access this resource").
			Required(),
		"code": validators.Number().
			Example(403).
			Optional(),
		"details": validators.String().
			Example("Insufficient privileges for this operation").
			Optional(),
	}).Example(map[string]interface{}{
		"error":   "forbidden",
		"message": "You do not have permission to access this resource",
		"code":    403,
		"details": "Insufficient privileges for this operation",
	}).Required()

	// NotFoundErrorSchema represents a 404 Not Found response
	NotFoundErrorSchema = validators.Object(map[string]interface{}{
		"error": validators.String().
			Example("not_found").
			Required(),
		"message": validators.String().
			Example("The requested resource was not found").
			Required(),
		"code": validators.Number().
			Example(404).
			Optional(),
		"details": validators.String().
			Example("User with ID 'usr_123' does not exist").
			Optional(),
	}).Example(map[string]interface{}{
		"error":   "not_found",
		"message": "The requested resource was not found",
		"code":    404,
		"details": "User with ID 'usr_123' does not exist",
	}).Required()

	// ConflictErrorSchema represents a 409 Conflict response
	ConflictErrorSchema = validators.Object(map[string]interface{}{
		"error": validators.String().
			Example("conflict").
			Required(),
		"message": validators.String().
			Example("The request conflicts with the current state of the resource").
			Required(),
		"code": validators.Number().
			Example(409).
			Optional(),
		"details": validators.String().
			Example("A user with this email already exists").
			Optional(),
	}).Example(map[string]interface{}{
		"error":   "conflict",
		"message": "The request conflicts with the current state of the resource",
		"code":    409,
		"details": "A user with this email already exists",
	}).Required()

	// UnprocessableEntityErrorSchema represents a 422 Unprocessable Entity response
	UnprocessableEntityErrorSchema = validators.Object(map[string]interface{}{
		"error": validators.String().
			Example("unprocessable_entity").
			Required(),
		"message": validators.String().
			Example("The request was well-formed but contains semantic errors").
			Required(),
		"code": validators.Number().
			Example(422).
			Optional(),
		"details": validators.String().
			Example("Cannot create user: business rules violation").
			Optional(),
	}).Example(map[string]interface{}{
		"error":   "unprocessable_entity",
		"message": "The request was well-formed but contains semantic errors",
		"code":    422,
		"details": "Cannot create user: business rules violation",
	}).Required()

	// TooManyRequestsErrorSchema represents a 429 Too Many Requests response
	TooManyRequestsErrorSchema = validators.Object(map[string]interface{}{
		"error": validators.String().
			Example("too_many_requests").
			Required(),
		"message": validators.String().
			Example("Too many requests sent in a given amount of time").
			Required(),
		"code": validators.Number().
			Example(429).
			Optional(),
		"details": validators.String().
			Example("Rate limit exceeded. Please try again in 60 seconds").
			Optional(),
	}).Example(map[string]interface{}{
		"error":   "too_many_requests",
		"message": "Too many requests sent in a given amount of time",
		"code":    429,
		"details": "Rate limit exceeded. Please try again in 60 seconds",
	}).Required()

	// InternalServerErrorSchema represents a 500 Internal Server Error response
	InternalServerErrorSchema = validators.Object(map[string]interface{}{
		"error": validators.String().
			Example("internal_server_error").
			Required(),
		"message": validators.String().
			Example("An unexpected error occurred on the server").
			Required(),
		"code": validators.Number().
			Example(500).
			Optional(),
		"details": validators.String().
			Example("Database connection failed").
			Optional(),
	}).Example(map[string]interface{}{
		"error":   "internal_server_error",
		"message": "An unexpected error occurred on the server",
		"code":    500,
		"details": "Database connection failed",
	}).Required()

	// BadGatewayErrorSchema represents a 502 Bad Gateway response
	BadGatewayErrorSchema = validators.Object(map[string]interface{}{
		"error": validators.String().
			Example("bad_gateway").
			Required(),
		"message": validators.String().
			Example("Bad gateway - upstream service is unavailable").
			Required(),
		"code": validators.Number().
			Example(502).
			Optional(),
		"details": validators.String().
			Example("Unable to connect to authentication service").
			Optional(),
	}).Example(map[string]interface{}{
		"error":   "bad_gateway",
		"message": "Bad gateway - upstream service is unavailable",
		"code":    502,
		"details": "Unable to connect to authentication service",
	}).Required()

	// ServiceUnavailableErrorSchema represents a 503 Service Unavailable response
	ServiceUnavailableErrorSchema = validators.Object(map[string]interface{}{
		"error": validators.String().
			Example("service_unavailable").
			Required(),
		"message": validators.String().
			Example("The service is temporarily unavailable").
			Required(),
		"code": validators.Number().
			Example(503).
			Optional(),
		"details": validators.String().
			Example("Service is under maintenance. Please try again later").
			Optional(),
	}).Example(map[string]interface{}{
		"error":   "service_unavailable",
		"message": "The service is temporarily unavailable",
		"code":    503,
		"details": "Service is under maintenance. Please try again later",
	}).Required()
)

// GetStandardErrorSchema returns the appropriate standard error schema for a given HTTP status code
func GetStandardErrorSchema(statusCode int) goop.Schema {
	switch statusCode {
	case 400:
		return BadRequestErrorSchema
	case 401:
		return UnauthorizedErrorSchema
	case 403:
		return ForbiddenErrorSchema
	case 404:
		return NotFoundErrorSchema
	case 409:
		return ConflictErrorSchema
	case 422:
		return UnprocessableEntityErrorSchema
	case 429:
		return TooManyRequestsErrorSchema
	case 500:
		return InternalServerErrorSchema
	case 502:
		return BadGatewayErrorSchema
	case 503:
		return ServiceUnavailableErrorSchema
	default:
		// Return generic error schema for unknown status codes
		return BadRequestErrorSchema
	}
}
