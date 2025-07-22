package gin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	goop "github.com/picogrid/go-op"
)

// ConvertOpenAPIPathToGin converts OpenAPI-style path parameters to Gin-style
// Example: /users/{id} -> /users/:id
func ConvertOpenAPIPathToGin(path string) string {
	// Find all occurrences of {parameter} and replace with :parameter
	result := path
	for {
		start := strings.Index(result, "{")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		// Extract parameter name and replace {param} with :param
		paramName := result[start+1 : end]
		result = result[:start] + ":" + paramName + result[end+1:]
	}
	return result
}

// Register registers one or more compiled operations with the Gin router
// This method performs zero reflection and maximum performance registration
func (r *GinRouter) Register(ops ...goop.CompiledOperation) error {
	for _, op := range ops {
		if err := r.registerSingle(op); err != nil {
			return fmt.Errorf("failed to register operation %s %s: %w", op.Method, op.Path, err)
		}
	}
	return nil
}

// registerSingle registers a single compiled operation with the Gin router
func (r *GinRouter) registerSingle(op goop.CompiledOperation) error {
	// Store the operation for generator processing
	r.operations = append(r.operations, op)

	// Convert OpenAPI path format to Gin format for routing
	// This keeps the framework-agnostic operation definition while adapting to Gin's requirements
	ginPath := ConvertOpenAPIPathToGin(op.Path)

	// Register the handler with Gin - zero reflection, maximum performance
	var ginHandler GinHandler
	if handler, ok := op.Handler.(GinHandler); ok {
		ginHandler = handler
	} else {
		// If it's not a GinHandler, we can't register it
		return fmt.Errorf("handler must be a gin.HandlerFunc for Gin router, got %T", op.Handler)
	}
	r.engine.Handle(op.Method, ginPath, ginHandler)

	// Process with all generators (build-time analysis)
	info := goop.OperationInfo{
		Method:      op.Method,
		Path:        op.Path,
		Summary:     op.Summary,
		Description: op.Description,
		Tags:        op.Tags,
		Security:    op.Security,
		Operation:   &op,
	}

	// Extract validation info if schemas are present and enhanced
	if op.ParamsSchema != nil {
		if enhanced, ok := op.ParamsSchema.(goop.EnhancedSchema); ok {
			info.ParamsInfo = enhanced.GetValidationInfo()
		}
	}
	if op.QuerySchema != nil {
		if enhanced, ok := op.QuerySchema.(goop.EnhancedSchema); ok {
			info.QueryInfo = enhanced.GetValidationInfo()
		}
	}
	if op.BodySchema != nil {
		if enhanced, ok := op.BodySchema.(goop.EnhancedSchema); ok {
			info.BodyInfo = enhanced.GetValidationInfo()
		}
	}
	if op.ResponseSchema != nil {
		if enhanced, ok := op.ResponseSchema.(goop.EnhancedSchema); ok {
			info.ResponseInfo = enhanced.GetValidationInfo()
		}
	}
	if op.HeaderSchema != nil {
		if enhanced, ok := op.HeaderSchema.(goop.EnhancedSchema); ok {
			info.HeaderInfo = enhanced.GetValidationInfo()
		}
	}

	// Process with all generators
	for _, generator := range r.generators {
		if err := generator.Process(info); err != nil {
			return fmt.Errorf("generator processing failed: %w", err)
		}
	}

	return nil
}

// GetOperations returns all registered operations
// Useful for build-time analysis and spec generation
func (r *GinRouter) GetOperations() []goop.CompiledOperation {
	// Return a copy to prevent external modification
	ops := make([]goop.CompiledOperation, len(r.operations))
	copy(ops, r.operations)
	return ops
}

// WithMiddleware chains middleware with a handler for operation-specific middleware application
// Usage: Handler(router.WithMiddleware(handlerFunc, middleware1, middleware2))
func (r *GinRouter) WithMiddleware(handler GinHandler, middleware ...GinHandler) GinHandler {
	return func(c *gin.Context) {
		// Apply each middleware in order
		for _, mw := range middleware {
			mw(c)
			if c.IsAborted() {
				return // Stop if middleware aborted the request
			}
		}
		// Call the actual handler if middleware didn't abort
		handler(c)
	}
}

// ServeSpec serves the OpenAPI specification as JSON
// This is useful for development and documentation purposes
func (r *GinRouter) ServeSpec(generator goop.Generator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would be implemented by specific generators
		// For now, return basic operation info
		specs := make([]map[string]interface{}, 0, len(r.operations))
		for _, op := range r.operations {
			spec := map[string]interface{}{
				"method":      op.Method,
				"path":        op.Path,
				"summary":     op.Summary,
				"description": op.Description,
				"tags":        op.Tags,
			}
			if op.ParamsSpec != nil {
				spec["parameters"] = op.ParamsSpec
			}
			if op.BodySpec != nil {
				spec["requestBody"] = op.BodySpec
			}
			if op.ResponseSpec != nil {
				spec["responses"] = map[string]interface{}{
					fmt.Sprintf("%d", op.SuccessCode): op.ResponseSpec,
				}
			}
			if len(op.Security) > 0 {
				spec["security"] = op.Security
			}
			if op.HeaderSpec != nil {
				spec["headerParameters"] = op.HeaderSpec
			}
			specs = append(specs, spec)
		}

		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, map[string]interface{}{
			"openapi": "3.1.0",
			"info": map[string]interface{}{
				"title":   "Generated API",
				"version": "1.0.0",
			},
			"paths": specs,
		})
	}
}
