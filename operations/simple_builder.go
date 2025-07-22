package operations

import (
	goop "github.com/picogrid/go-op"
)

// ResponseDefinition represents a response with its schema, description, and optional headers
type ResponseDefinition struct {
	Schema      goop.Schema
	Description string
	Headers     map[string]goop.Schema
}

// Core operation configuration struct
// This contains all the operation metadata and schemas
type operationConfig struct {
	method         string
	path           string
	summary        string
	description    string
	tags           []string
	successCode    int
	paramsSchema   goop.Schema
	querySchema    goop.Schema
	bodySchema     goop.Schema
	responseSchema goop.Schema // Keep for backward compatibility
	headerSchema   goop.Schema
	security       goop.SecurityRequirements
	responses      map[int]ResponseDefinition // New: Multiple responses support
}

// Helper method to compile the final operation
func (config *operationConfig) compile(handler HTTPHandler) CompiledOperation {
	op := CompiledOperation{
		Method:      config.method,
		Path:        config.path,
		Summary:     config.summary,
		Description: config.description,
		Tags:        config.tags,
		SuccessCode: config.successCode,
		Handler:     handler,
		Security:    config.security,
		Responses:   make(map[int]goop.ResponseDefinition),
	}

	// Copy all defined responses
	for code, response := range config.responses {
		op.Responses[code] = goop.ResponseDefinition{
			Schema:      response.Schema,
			Description: response.Description,
			Headers:     response.Headers,
		}
	}

	// Generate OpenAPI specs if schemas are provided
	if config.paramsSchema != nil {
		op.ParamsSchema = config.paramsSchema
		if enhanced, ok := config.paramsSchema.(goop.EnhancedSchema); ok {
			op.ParamsSpec = enhanced.ToOpenAPISchema()
		}
	}
	if config.querySchema != nil {
		op.QuerySchema = config.querySchema
		if enhanced, ok := config.querySchema.(goop.EnhancedSchema); ok {
			op.QuerySpec = enhanced.ToOpenAPISchema()
		}
	}
	if config.bodySchema != nil {
		op.BodySchema = config.bodySchema
		if enhanced, ok := config.bodySchema.(goop.EnhancedSchema); ok {
			op.BodySpec = enhanced.ToOpenAPISchema()
		}
	}
	if config.responseSchema != nil {
		op.ResponseSchema = config.responseSchema
		if enhanced, ok := config.responseSchema.(goop.EnhancedSchema); ok {
			op.ResponseSpec = enhanced.ToOpenAPISchema()
		}
	}
	if config.headerSchema != nil {
		op.HeaderSchema = config.headerSchema
		if enhanced, ok := config.headerSchema.(goop.EnhancedSchema); ok {
			op.HeaderSpec = enhanced.ToOpenAPISchema()
		}
	}

	return op
}

// SimpleOperationBuilder provides a simplified builder for the MVP
// This avoids the complexity of the full builder pattern for now
type SimpleOperationBuilder struct {
	config *operationConfig
}

// NewSimple creates a new simple operation builder
func NewSimple() *SimpleOperationBuilder {
	return &SimpleOperationBuilder{
		config: &operationConfig{
			tags:        []string{},
			successCode: 200,
			responses:   make(map[int]ResponseDefinition),
		},
	}
}

// Method sets the HTTP method and path
func (s *SimpleOperationBuilder) Method(method, path string) *SimpleOperationBuilder {
	s.config.method = method
	s.config.path = path
	return s
}

// GET sets the HTTP method to GET
func (s *SimpleOperationBuilder) GET(path string) *SimpleOperationBuilder {
	return s.Method("GET", path)
}

// POST sets the HTTP method to POST
func (s *SimpleOperationBuilder) POST(path string) *SimpleOperationBuilder {
	return s.Method("POST", path)
}

// PUT sets the HTTP method to PUT
func (s *SimpleOperationBuilder) PUT(path string) *SimpleOperationBuilder {
	return s.Method("PUT", path)
}

// PATCH sets the HTTP method to PATCH
func (s *SimpleOperationBuilder) PATCH(path string) *SimpleOperationBuilder {
	return s.Method("PATCH", path)
}

// DELETE sets the HTTP method to DELETE
func (s *SimpleOperationBuilder) DELETE(path string) *SimpleOperationBuilder {
	return s.Method("DELETE", path)
}

// Summary sets the operation summary
func (s *SimpleOperationBuilder) Summary(summary string) *SimpleOperationBuilder {
	s.config.summary = summary
	return s
}

// Description sets the operation description
func (s *SimpleOperationBuilder) Description(description string) *SimpleOperationBuilder {
	s.config.description = description
	return s
}

// Tags adds tags to the operation
func (s *SimpleOperationBuilder) Tags(tags ...string) *SimpleOperationBuilder {
	s.config.tags = append(s.config.tags, tags...)
	return s
}

// SuccessCode sets the success HTTP status code
func (s *SimpleOperationBuilder) SuccessCode(code int) *SimpleOperationBuilder {
	s.config.successCode = code
	return s
}

// WithParams sets the parameters schema
func (s *SimpleOperationBuilder) WithParams(schema goop.Schema) *SimpleOperationBuilder {
	s.config.paramsSchema = schema
	return s
}

// WithQuery sets the query parameters schema
func (s *SimpleOperationBuilder) WithQuery(schema goop.Schema) *SimpleOperationBuilder {
	s.config.querySchema = schema
	return s
}

// WithBody sets the request body schema
func (s *SimpleOperationBuilder) WithBody(schema goop.Schema) *SimpleOperationBuilder {
	s.config.bodySchema = schema
	return s
}

// WithResponse sets the response schema (backward compatibility - maps to 200 response)
func (s *SimpleOperationBuilder) WithResponse(schema goop.Schema) *SimpleOperationBuilder {
	s.config.responseSchema = schema
	// Also add as 200 response for new system
	s.config.responses[200] = ResponseDefinition{
		Schema:      schema,
		Description: "Successful response",
	}
	return s
}

// WithResponseCode sets a response schema for a specific HTTP status code
func (s *SimpleOperationBuilder) WithResponseCode(code int, schema goop.Schema, description string) *SimpleOperationBuilder {
	s.config.responses[code] = ResponseDefinition{
		Schema:      schema,
		Description: description,
	}
	return s
}

// WithSuccessResponse sets a success response (2xx range)
func (s *SimpleOperationBuilder) WithSuccessResponse(code int, schema goop.Schema, description string) *SimpleOperationBuilder {
	if code < 200 || code >= 300 {
		panic("Success response codes must be in the 2xx range")
	}
	return s.WithResponseCode(code, schema, description)
}

// WithErrorResponse sets an error response (4xx or 5xx range)
func (s *SimpleOperationBuilder) WithErrorResponse(code int, schema goop.Schema, description string) *SimpleOperationBuilder {
	if code < 400 {
		panic("Error response codes must be in the 4xx or 5xx range")
	}
	return s.WithResponseCode(code, schema, description)
}

// Convenience methods for common success responses
func (s *SimpleOperationBuilder) WithCreatedResponse(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithSuccessResponse(201, schema, "Resource created successfully")
}

func (s *SimpleOperationBuilder) WithAcceptedResponse(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithSuccessResponse(202, schema, "Request accepted for processing")
}

func (s *SimpleOperationBuilder) WithNoContentResponse() *SimpleOperationBuilder {
	return s.WithResponseCode(204, nil, "No content")
}

// Convenience methods for common error responses
func (s *SimpleOperationBuilder) WithBadRequestError(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithErrorResponse(400, schema, "Bad Request")
}

func (s *SimpleOperationBuilder) WithUnauthorizedError(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithErrorResponse(401, schema, "Unauthorized")
}

func (s *SimpleOperationBuilder) WithForbiddenError(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithErrorResponse(403, schema, "Forbidden")
}

func (s *SimpleOperationBuilder) WithNotFoundError(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithErrorResponse(404, schema, "Not Found")
}

func (s *SimpleOperationBuilder) WithConflictError(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithErrorResponse(409, schema, "Conflict")
}

func (s *SimpleOperationBuilder) WithUnprocessableEntityError(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithErrorResponse(422, schema, "Unprocessable Entity")
}

func (s *SimpleOperationBuilder) WithTooManyRequestsError(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithErrorResponse(429, schema, "Too Many Requests")
}

func (s *SimpleOperationBuilder) WithServerError(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithErrorResponse(500, schema, "Internal Server Error")
}

func (s *SimpleOperationBuilder) WithBadGatewayError(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithErrorResponse(502, schema, "Bad Gateway")
}

func (s *SimpleOperationBuilder) WithServiceUnavailableError(schema goop.Schema) *SimpleOperationBuilder {
	return s.WithErrorResponse(503, schema, "Service Unavailable")
}

// WithHeaders sets the header parameters schema
func (s *SimpleOperationBuilder) WithHeaders(schema goop.Schema) *SimpleOperationBuilder {
	s.config.headerSchema = schema
	return s
}

// WithSecurity sets the security requirements for this operation
func (s *SimpleOperationBuilder) WithSecurity(requirements goop.SecurityRequirements) *SimpleOperationBuilder {
	s.config.security = requirements
	return s
}

// RequireAuth adds a security requirement for a specific scheme with optional scopes
func (s *SimpleOperationBuilder) RequireAuth(schemeName string, scopes ...string) *SimpleOperationBuilder {
	if s.config.security == nil {
		s.config.security = goop.SecurityRequirements{}
	}
	s.config.security = s.config.security.RequireScheme(schemeName, scopes...)
	return s
}

// RequireAnyOf adds multiple security schemes where any one can satisfy authentication (OR logic)
func (s *SimpleOperationBuilder) RequireAnyOf(schemes ...string) *SimpleOperationBuilder {
	if s.config.security == nil {
		s.config.security = goop.SecurityRequirements{}
	}

	requirements := make([]goop.SecurityRequirement, len(schemes))
	for i, scheme := range schemes {
		requirements[i] = goop.SecurityRequirement{scheme: []string{}}
	}

	s.config.security = s.config.security.RequireAny(requirements...)
	return s
}

// RequireAPIKey is a convenience method for API key authentication
func (s *SimpleOperationBuilder) RequireAPIKey(schemeName string) *SimpleOperationBuilder {
	return s.RequireAuth(schemeName)
}

// RequireBearer is a convenience method for Bearer token authentication
func (s *SimpleOperationBuilder) RequireBearer(schemeName string) *SimpleOperationBuilder {
	return s.RequireAuth(schemeName)
}

// RequireOAuth2 is a convenience method for OAuth2 authentication with specific scopes
func (s *SimpleOperationBuilder) RequireOAuth2(schemeName string, scopes ...string) *SimpleOperationBuilder {
	return s.RequireAuth(schemeName, scopes...)
}

// NoAuth removes all authentication requirements (public endpoint)
func (s *SimpleOperationBuilder) NoAuth() *SimpleOperationBuilder {
	s.config.security = goop.NoAuth()
	return s
}

// WithCommonErrors adds standard error responses (400, 401, 403, 500) to the operation
func (s *SimpleOperationBuilder) WithCommonErrors() *SimpleOperationBuilder {
	return s.
		WithBadRequestError(BadRequestErrorSchema).
		WithUnauthorizedError(UnauthorizedErrorSchema).
		WithForbiddenError(ForbiddenErrorSchema).
		WithServerError(InternalServerErrorSchema)
}

// WithAuthErrors adds authentication-related error responses (401, 403) to the operation
func (s *SimpleOperationBuilder) WithAuthErrors() *SimpleOperationBuilder {
	return s.
		WithUnauthorizedError(UnauthorizedErrorSchema).
		WithForbiddenError(ForbiddenErrorSchema)
}

// WithValidationErrors adds validation-related error responses (400, 422) to the operation
func (s *SimpleOperationBuilder) WithValidationErrors() *SimpleOperationBuilder {
	return s.
		WithBadRequestError(ValidationErrorSchema).
		WithUnprocessableEntityError(UnprocessableEntityErrorSchema)
}

// WithCRUDErrors adds common CRUD operation error responses (400, 401, 403, 404, 500)
func (s *SimpleOperationBuilder) WithCRUDErrors() *SimpleOperationBuilder {
	return s.
		WithBadRequestError(ValidationErrorSchema).
		WithUnauthorizedError(UnauthorizedErrorSchema).
		WithForbiddenError(ForbiddenErrorSchema).
		WithNotFoundError(NotFoundErrorSchema).
		WithServerError(InternalServerErrorSchema)
}

// WithCreateErrors adds error responses typical for create operations (400, 401, 403, 409, 422, 500)
func (s *SimpleOperationBuilder) WithCreateErrors() *SimpleOperationBuilder {
	return s.
		WithBadRequestError(ValidationErrorSchema).
		WithUnauthorizedError(UnauthorizedErrorSchema).
		WithForbiddenError(ForbiddenErrorSchema).
		WithConflictError(ConflictErrorSchema).
		WithUnprocessableEntityError(UnprocessableEntityErrorSchema).
		WithServerError(InternalServerErrorSchema)
}

// WithStandardErrorsByCode allows adding multiple standard error responses by status codes
func (s *SimpleOperationBuilder) WithStandardErrorsByCode(codes ...int) *SimpleOperationBuilder {
	for _, code := range codes {
		s.WithErrorResponse(code, GetStandardErrorSchema(code), getStandardErrorDescription(code))
	}
	return s
}

// getStandardErrorDescription returns standard descriptions for HTTP status codes
func getStandardErrorDescription(code int) string {
	switch code {
	case 400:
		return "Bad Request - The request could not be understood"
	case 401:
		return "Unauthorized - Authentication is required"
	case 403:
		return "Forbidden - Insufficient permissions"
	case 404:
		return "Not Found - The requested resource was not found"
	case 409:
		return "Conflict - The request conflicts with current state"
	case 422:
		return "Unprocessable Entity - Request contains semantic errors"
	case 429:
		return "Too Many Requests - Rate limit exceeded"
	case 500:
		return "Internal Server Error - An unexpected error occurred"
	case 502:
		return "Bad Gateway - Upstream service unavailable"
	case 503:
		return "Service Unavailable - Service temporarily unavailable"
	default:
		return "Error"
	}
}

// Handler compiles the operation with the provided handler
func (s *SimpleOperationBuilder) Handler(handler HTTPHandler) CompiledOperation {
	return s.config.compile(handler)
}
