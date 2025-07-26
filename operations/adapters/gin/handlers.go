package gin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	goop "github.com/picogrid/go-op"
)

// structToMap converts a struct to map[string]interface{} for validation.
// This is necessary because ForStruct validators expect map data, not struct types.
func structToMap(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	// Marshal struct to JSON
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON to map
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return m, nil
}

// CreateValidatedHandler creates a high-performance Gin handler with automatic validation
// This function generates optimized validation code without reflection
func CreateValidatedHandler[P, Q, B, R any](
	handler goop.Handler[P, Q, B, R],
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
					"error":   "Invalid path parameters",
					"details": err.Error(),
				})
				return
			}

			// Convert struct to map for validation
			paramsMap, err := structToMap(params)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Failed to process path parameters",
					"details": err.Error(),
				})
				return
			}

			if err := paramsSchema.Validate(paramsMap); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Path parameter validation failed",
					"details": err.Error(),
				})
				return
			}
		}

		// Validate and bind query parameters
		if querySchema != nil {
			if err := c.ShouldBindQuery(&query); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid query parameters",
					"details": err.Error(),
				})
				return
			}

			// Convert struct to map for validation
			queryMap, err := structToMap(query)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Failed to process query parameters",
					"details": err.Error(),
				})
				return
			}

			if err := querySchema.Validate(queryMap); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Query parameter validation failed",
					"details": err.Error(),
				})
				return
			}
		}

		// Validate and bind request body
		if bodySchema != nil {
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid request body",
					"details": err.Error(),
				})
				return
			}

			// Convert struct to map for validation
			// ForStruct validators expect map[string]interface{}, not struct types
			bodyMap, err := structToMap(body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Failed to process request body",
					"details": err.Error(),
				})
				return
			}

			if err := bodySchema.Validate(bodyMap); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Request body validation failed",
					"details": err.Error(),
				})
				return
			}
		}

		// Transfer all Gin context values to standard context
		// We intentionally use string keys here to preserve Gin's context keys
		ctx := c.Request.Context()
		for key, value := range c.Keys {
			ctx = context.WithValue(ctx, key, value) //nolint:staticcheck // SA1029: Gin uses string keys, we must preserve them
		}

		// Call the business logic handler
		result, err := handler(ctx, params, query, body)
		if err != nil {
			// Handle business logic errors
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"details": err.Error(),
			})
			return
		}

		// Validate response if schema is provided
		if responseSchema != nil {
			// Convert struct to map for validation
			resultMap, err := structToMap(result)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to process response",
					"details": err.Error(),
				})
				return
			}

			if err := responseSchema.Validate(resultMap); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Response validation failed",
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
					"error":   "Invalid path parameters",
					"details": err.Error(),
				})
				c.Abort()
				return
			}
			if err := paramsSchema.Validate(params); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Path parameter validation failed",
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
					"error":   "Invalid query parameters",
					"details": err.Error(),
				})
				c.Abort()
				return
			}
			if err := querySchema.Validate(query); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Query parameter validation failed",
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
					"error":   "Invalid request body",
					"details": err.Error(),
				})
				c.Abort()
				return
			}
			if err := bodySchema.Validate(body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Request body validation failed",
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
