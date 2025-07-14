package operations

import (
	"context"

	"github.com/gin-gonic/gin"

	goop "github.com/picogrid/go-op"
)

// Handler represents a type-safe operation handler function
// Context provides access to the request context and other data
// P, Q, B represent Params, Query, and Body types
// R represents the Response type
type Handler[P, Q, B, R any] func(ctx context.Context, params P, query Q, body B) (R, error)

// GinHandler represents a raw Gin handler function for maximum performance
// This is what gets registered with the Gin router - no reflection needed
type GinHandler = gin.HandlerFunc

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
	ParamsSpec   *goop.OpenAPISchema
	QuerySpec    *goop.OpenAPISchema
	BodySpec     *goop.OpenAPISchema
	ResponseSpec *goop.OpenAPISchema
	HeaderSpec   *goop.OpenAPISchema

	// Validation schemas for runtime validation
	ParamsSchema   goop.Schema
	QuerySchema    goop.Schema
	BodySchema     goop.Schema
	ResponseSchema goop.Schema
	HeaderSchema   goop.Schema

	// Security requirements for this operation
	Security goop.SecurityRequirements

	// Raw handler function - no reflection, maximum performance
	Handler GinHandler

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
	Security     goop.SecurityRequirements
	Operation    *CompiledOperation
	ParamsInfo   *goop.ValidationInfo
	QueryInfo    *goop.ValidationInfo
	BodyInfo     *goop.ValidationInfo
	ResponseInfo *goop.ValidationInfo
	HeaderInfo   *goop.ValidationInfo
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
