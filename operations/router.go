package operations

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/picogrid/go-op"
)

// Router provides zero-reflection operation registration and handler creation
// This is the core component that enables high-performance API operations
type Router struct {
	engine     *gin.Engine
	generators []Generator
	operations []CompiledOperation
}

// NewRouter creates a new router with the specified Gin engine and generators
func NewRouter(engine *gin.Engine, generators ...Generator) *Router {
	return &Router{
		engine:     engine,
		generators: generators,
		operations: make([]CompiledOperation, 0),
	}
}

// Register registers a compiled operation with the router
// This method performs zero reflection and maximum performance registration
func (r *Router) Register(op CompiledOperation) error {
	// Store the operation for generator processing
	r.operations = append(r.operations, op)

	// Register the handler with Gin - zero reflection, maximum performance
	r.engine.Handle(op.Method, op.Path, op.Handler)

	// Process with all generators (build-time analysis)
	info := OperationInfo{
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
func (r *Router) GetOperations() []CompiledOperation {
	// Return a copy to prevent external modification
	operations := make([]CompiledOperation, len(r.operations))
	copy(operations, r.operations)
	return operations
}

// CreateValidatedHandler creates a high-performance handler with automatic validation
// This function generates optimized validation code without reflection
func CreateValidatedHandler[P, Q, B, R any](
	handler Handler[P, Q, B, R],
	paramsSchema goop.Schema,
	querySchema goop.Schema,
	bodySchema goop.Schema,
	responseSchema goop.Schema,
) GinHandler {
	return func(c *gin.Context) {
		var params P
		var query Q
		var body B

		// Validate and bind parameters with zero allocation paths
		if paramsSchema != nil {
			if err := c.ShouldBindUri(&params); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid path parameters",
					"details": err.Error(),
				})
				return
			}
			if err := paramsSchema.Validate(params); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Path parameter validation failed",
					"details": err.Error(),
				})
				return
			}
		}

		// Validate and bind query parameters
		if querySchema != nil {
			if err := c.ShouldBindQuery(&query); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid query parameters",
					"details": err.Error(),
				})
				return
			}
			if err := querySchema.Validate(query); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Query parameter validation failed",
					"details": err.Error(),
				})
				return
			}
		}

		// Validate and bind request body
		if bodySchema != nil {
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid request body",
					"details": err.Error(),
				})
				return
			}
			if err := bodySchema.Validate(body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Request body validation failed",
					"details": err.Error(),
				})
				return
			}
		}

		// Call the business logic handler
		result, err := handler(c.Request.Context(), params, query, body)
		if err != nil {
			// Handle business logic errors
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"details": err.Error(),
			})
			return
		}

		// Validate response if schema is provided
		if responseSchema != nil {
			if err := responseSchema.Validate(result); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Response validation failed",
					"details": err.Error(),
				})
				return
			}
		}

		// Return successful response
		c.JSON(http.StatusOK, result)
	}
}

// ValidationMiddleware creates middleware for automatic request validation
// This provides an alternative approach for adding validation to existing handlers
func ValidationMiddleware(
	paramsSchema goop.Schema,
	querySchema goop.Schema,
	bodySchema goop.Schema,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate path parameters
		if paramsSchema != nil {
			var params interface{}
			if err := c.ShouldBindUri(&params); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid path parameters",
					"details": err.Error(),
				})
				c.Abort()
				return
			}
			if err := paramsSchema.Validate(params); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Path parameter validation failed",
					"details": err.Error(),
				})
				c.Abort()
				return
			}
			c.Set("validatedParams", params)
		}

		// Validate query parameters
		if querySchema != nil {
			var query interface{}
			if err := c.ShouldBindQuery(&query); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid query parameters",
					"details": err.Error(),
				})
				c.Abort()
				return
			}
			if err := querySchema.Validate(query); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Query parameter validation failed",
					"details": err.Error(),
				})
				c.Abort()
				return
			}
			c.Set("validatedQuery", query)
		}

		// Validate request body
		if bodySchema != nil {
			var body interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid request body",
					"details": err.Error(),
				})
				c.Abort()
				return
			}
			if err := bodySchema.Validate(body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Request body validation failed",
					"details": err.Error(),
				})
				c.Abort()
				return
			}
			c.Set("validatedBody", body)
		}

		// Continue to next handler
		c.Next()
	}
}

// Helper function to create a simple JSON response handler
func JSONResponse(data interface{}) GinHandler {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, data)
	}
}

// Helper function to create an error response handler
func ErrorResponse(statusCode int, message string) GinHandler {
	return func(c *gin.Context) {
		c.JSON(statusCode, gin.H{
			"error": message,
		})
	}
}

// ServeSpec serves the OpenAPI specification as JSON
// This is useful for development and documentation purposes
func (r *Router) ServeSpec(generator Generator) gin.HandlerFunc {
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
			if op.Security != nil && len(op.Security) > 0 {
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