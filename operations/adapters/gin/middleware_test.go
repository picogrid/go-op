package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWithMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		middleware     []GinHandler
		shouldAbort    bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "no middleware",
			middleware:     []GinHandler{},
			shouldAbort:    false,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"handler executed"}`,
		},
		{
			name: "middleware that continues",
			middleware: []GinHandler{
				func(c *gin.Context) {
					c.Set("middleware1", "executed")
					c.Next()
				},
				func(c *gin.Context) {
					c.Set("middleware2", "executed")
					c.Next()
				},
			},
			shouldAbort:    false,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"handler executed","middleware1":"executed","middleware2":"executed"}`,
		},
		{
			name: "middleware that aborts",
			middleware: []GinHandler{
				func(c *gin.Context) {
					c.Set("middleware1", "executed")
					c.Next()
				},
				func(c *gin.Context) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					c.Abort()
				},
			},
			shouldAbort:    true,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler that uses context values set by middleware
			testHandler := func(c *gin.Context) {
				response := gin.H{"message": "handler executed"}

				// Add any values set by middleware to the response
				if val, exists := c.Get("middleware1"); exists {
					response["middleware1"] = val
				}
				if val, exists := c.Get("middleware2"); exists {
					response["middleware2"] = val
				}

				c.JSON(http.StatusOK, response)
			}

			// Create router and setup
			engine := gin.New()
			router := NewGinRouter(engine)

			// Apply WithMiddleware
			finalHandler := router.WithMiddleware(testHandler, tt.middleware...)

			// Register the handler
			engine.GET("/test", finalHandler)

			// Create test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			engine.ServeHTTP(w, req)

			// Assert results
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestWithMiddleware_OrderMatters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test that middleware executes in the correct order
	var executionOrder []string

	middleware1 := func(c *gin.Context) {
		executionOrder = append(executionOrder, "middleware1")
		c.Next()
	}

	middleware2 := func(c *gin.Context) {
		executionOrder = append(executionOrder, "middleware2")
		c.Next()
	}

	testHandler := func(c *gin.Context) {
		executionOrder = append(executionOrder, "handler")
		c.JSON(http.StatusOK, gin.H{"execution_order": executionOrder})
	}

	// Create router and setup
	engine := gin.New()
	router := NewGinRouter(engine)

	// Apply WithMiddleware with specific order
	finalHandler := router.WithMiddleware(testHandler, middleware1, middleware2)
	engine.GET("/test", finalHandler)

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	// Assert execution order
	assert.Equal(t, http.StatusOK, w.Code)
	expected := []string{"middleware1", "middleware2", "handler"}
	assert.Equal(t, expected, executionOrder)
}

func TestMiddlewareErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test middleware that returns different error responses
	tests := []struct {
		name           string
		middleware     GinHandler
		expectedStatus int
		expectedError  string
	}{
		{
			name: "unauthorized middleware",
			middleware: func(c *gin.Context) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "unauthorized",
					"message": "Authentication is required to access this resource",
					"code":    401,
				})
				c.Abort()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name: "forbidden middleware",
			middleware: func(c *gin.Context) {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "forbidden",
					"message": "You do not have permission to access this resource",
					"code":    403,
				})
				c.Abort()
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
		},
		{
			name: "rate limit middleware",
			middleware: func(c *gin.Context) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "too_many_requests",
					"message": "Too many requests sent in a given amount of time",
					"code":    429,
					"details": "Rate limit exceeded. Please try again in 60 seconds",
				})
				c.Abort()
			},
			expectedStatus: http.StatusTooManyRequests,
			expectedError:  "too_many_requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handler that should not be reached
			testHandler := func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "handler should not execute"})
			}

			// Create router and setup
			engine := gin.New()
			router := NewGinRouter(engine)

			// Apply middleware that should abort
			finalHandler := router.WithMiddleware(testHandler, tt.middleware)
			engine.GET("/test", finalHandler)

			// Create test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			engine.ServeHTTP(w, req)

			// Assert error response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
		})
	}
}

func TestAuthenticationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock authentication middleware that validates tokens
	authMiddleware := func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication is required to access this resource",
				"code":    401,
				"details": "Missing Authorization header",
			})
			c.Abort()
			return
		}
		if token != "Bearer valid-token" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication is required to access this resource",
				"code":    401,
				"details": "Invalid or expired authentication token",
			})
			c.Abort()
			return
		}
		// Set user context for authorized requests
		c.Set("user_id", "usr_123")
		c.Next()
	}

	// Protected handler
	protectedHandler := func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{
			"message": "Protected resource accessed",
			"user_id": userID,
		})
	}

	// Create router
	engine := gin.New()
	router := NewGinRouter(engine)
	finalHandler := router.WithMiddleware(protectedHandler, authMiddleware)
	engine.GET("/protected", finalHandler)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "no token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Missing Authorization header",
		},
		{
			name:           "invalid token",
			token:          "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid or expired authentication token",
		},
		{
			name:           "valid token",
			token:          "Bearer valid-token",
			expectedStatus: http.StatusOK,
			expectedBody:   "Protected resource accessed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}
