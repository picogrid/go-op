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
