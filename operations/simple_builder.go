package operations

import (
	goop "github.com/picogrid/go-op"
)

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
	responseSchema goop.Schema
	headerSchema   goop.Schema
	security       goop.SecurityRequirements
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

// WithResponse sets the response schema
func (s *SimpleOperationBuilder) WithResponse(schema goop.Schema) *SimpleOperationBuilder {
	s.config.responseSchema = schema
	return s
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

// Handler compiles the operation with the provided handler
func (s *SimpleOperationBuilder) Handler(handler HTTPHandler) CompiledOperation {
	return s.config.compile(handler)
}
