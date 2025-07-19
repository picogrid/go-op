package operations

import (
	goop "github.com/picogrid/go-op"
)

// Handler represents a type-safe operation handler function
// Context provides access to the request context and other data
// P, Q, B represent Params, Query, and Body types
// R represents the Response type
type Handler[P, Q, B, R any] = goop.Handler[P, Q, B, R]

// HTTPHandler represents a generic HTTP handler function
// This is framework-agnostic and can be adapted to any HTTP framework
type HTTPHandler = goop.HTTPHandler

// CompiledOperation represents a fully compiled operation with all metadata
// This structure contains everything needed for zero-reflection runtime execution
type CompiledOperation = goop.CompiledOperation

// OperationInfo contains metadata about an operation for build-time analysis
// Used by generators to extract information without runtime reflection
type OperationInfo = goop.OperationInfo

// Generator interface for processing operations at build time
// Implementations can generate OpenAPI specs, gRPC definitions, etc.
type Generator = goop.Generator

// HTTPMethod constants for type safety
const (
	GET     = goop.GET
	POST    = goop.POST
	PUT     = goop.PUT
	PATCH   = goop.PATCH
	DELETE  = goop.DELETE
	HEAD    = goop.HEAD
	OPTIONS = goop.OPTIONS
)
