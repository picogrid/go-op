package goop

// HTTPHandler represents a generic HTTP handler function
// This is framework-agnostic and can be adapted to any HTTP framework
type HTTPHandler interface{}

// CompiledOperation represents a fully compiled operation with all metadata
// This structure contains everything needed for zero-reflection runtime execution
type CompiledOperation struct {
	// HTTP metadata
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string

	// Schema specifications (pre-computed at build time)
	ParamsSpec   *OpenAPISchema
	QuerySpec    *OpenAPISchema
	BodySpec     *OpenAPISchema
	ResponseSpec *OpenAPISchema
	HeaderSpec   *OpenAPISchema

	// Validation schemas for runtime validation
	ParamsSchema   Schema
	QuerySchema    Schema
	BodySchema     Schema
	ResponseSchema Schema
	HeaderSchema   Schema

	// Security requirements for this operation
	Security SecurityRequirements

	// Raw handler function - no reflection, maximum performance
	// This is framework-specific and should be cast to the appropriate type
	Handler HTTPHandler

	// Success HTTP status code
	SuccessCode int
}

// OperationInfo contains metadata about an operation for build-time analysis
// Used by generators to extract information without runtime reflection
type OperationInfo struct {
	Method       string
	Path         string
	Summary      string
	Description  string
	Tags         []string
	Security     SecurityRequirements
	Operation    *CompiledOperation
	ParamsInfo   *ValidationInfo
	QueryInfo    *ValidationInfo
	BodyInfo     *ValidationInfo
	ResponseInfo *ValidationInfo
	HeaderInfo   *ValidationInfo
}

// Generator interface for processing operations at build time
// Implementations can generate OpenAPI specs, gRPC definitions, etc.
type Generator interface {
	Process(info OperationInfo) error
}

// HTTPMethod constants for type safety
const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	PATCH   = "PATCH"
	DELETE  = "DELETE"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
)
