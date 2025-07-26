package gin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	goop "github.com/picogrid/go-op"
)

// contextKey is the type for our context keys in tests
type contextKey string

func TestContextTransfer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("transfers single value from gin context", func(t *testing.T) {
		router := gin.New()

		// Middleware sets a value
		router.Use(func(c *gin.Context) {
			c.Set("userID", "123")
			c.Next()
		})

		// Handler expects value in context
		handler := func(ctx context.Context, _ struct{}, _ struct{}, _ struct{}) (map[string]string, error) {
			userID := ctx.Value("userID")
			if userID == nil {
				return nil, errors.New("userID not found in context")
			}
			return map[string]string{"userID": userID.(string)}, nil
		}

		// Create validated handler
		validatedHandler := CreateValidatedHandler(handler, nil, nil, nil, nil)
		router.GET("/test", validatedHandler)

		// Test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// Should succeed with context value available
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"userID":"123"`)
	})

	t.Run("transfers multiple values from gin context", func(t *testing.T) {
		router := gin.New()

		// Middleware sets multiple values
		router.Use(func(c *gin.Context) {
			c.Set("userID", "123")
			c.Set("orgID", "456")
			c.Set("role", "admin")
			c.Next()
		})

		// Handler expects all values in context
		handler := func(ctx context.Context, _ struct{}, _ struct{}, _ struct{}) (map[string]string, error) {
			userID := ctx.Value("userID")
			orgID := ctx.Value("orgID")
			role := ctx.Value("role")

			if userID == nil || orgID == nil || role == nil {
				return nil, errors.New("context values not found")
			}

			return map[string]string{
				"userID": userID.(string),
				"orgID":  orgID.(string),
				"role":   role.(string),
			}, nil
		}

		// Create validated handler
		validatedHandler := CreateValidatedHandler(handler, nil, nil, nil, nil)
		router.GET("/test", validatedHandler)

		// Test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// Should succeed with all context values available
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"userID":"123"`)
		assert.Contains(t, w.Body.String(), `"orgID":"456"`)
		assert.Contains(t, w.Body.String(), `"role":"admin"`)
	})

	t.Run("transfers complex value types", func(t *testing.T) {
		router := gin.New()

		type User struct {
			ID    string
			Email string
		}

		// Middleware sets complex value
		router.Use(func(c *gin.Context) {
			user := User{ID: "123", Email: "test@example.com"}
			c.Set("user", user)
			c.Next()
		})

		// Handler expects complex value in context
		handler := func(ctx context.Context, _ struct{}, _ struct{}, _ struct{}) (User, error) {
			userValue := ctx.Value("user")
			if userValue == nil {
				return User{}, errors.New("user not found in context")
			}
			user, ok := userValue.(User)
			if !ok {
				return User{}, errors.New("user value is not of expected type")
			}
			return user, nil
		}

		// Create validated handler
		validatedHandler := CreateValidatedHandler(handler, nil, nil, nil, nil)
		router.GET("/test", validatedHandler)

		// Test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// Should succeed with complex value available
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"ID":"123"`)
		assert.Contains(t, w.Body.String(), `"Email":"test@example.com"`)
	})

	t.Run("preserves existing request context values", func(t *testing.T) {
		router := gin.New()

		// Add initial context value to request
		router.Use(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), contextKey("requestID"), "req-123")
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		})

		// Middleware sets gin context value
		router.Use(func(c *gin.Context) {
			c.Set("userID", "456")
			c.Next()
		})

		// Handler expects both values
		handler := func(ctx context.Context, _ struct{}, _ struct{}, _ struct{}) (map[string]string, error) {
			requestID := ctx.Value(contextKey("requestID"))
			userID := ctx.Value("userID")

			if requestID == nil || userID == nil {
				return nil, errors.New("context values not found")
			}

			return map[string]string{
				"requestID": requestID.(string),
				"userID":    userID.(string),
			}, nil
		}

		// Create validated handler
		validatedHandler := CreateValidatedHandler(handler, nil, nil, nil, nil)
		router.GET("/test", validatedHandler)

		// Test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// Should have both values
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"requestID":"req-123"`)
		assert.Contains(t, w.Body.String(), `"userID":"456"`)
	})

	t.Run("works with empty gin context", func(t *testing.T) {
		router := gin.New()

		// Handler doesn't expect any context values
		handler := func(ctx context.Context, _ struct{}, _ struct{}, _ struct{}) (string, error) {
			return "success", nil
		}

		// Create validated handler
		validatedHandler := CreateValidatedHandler(handler, nil, nil, nil, nil)
		router.GET("/test", validatedHandler)

		// Test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// Should succeed
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})

	t.Run("context values available with validation", func(t *testing.T) {
		router := gin.New()

		// Middleware sets context
		router.Use(func(c *gin.Context) {
			c.Set("userID", "123")
			c.Next()
		})

		// Create body schema for validation
		bodySchema := struct{ goop.Schema }{
			Schema: mockSchema{
				validateFunc: func(data interface{}) error {
					// Schema validation passes
					return nil
				},
			},
		}.Schema

		type RequestBody struct {
			Name string `json:"name"`
		}

		// Handler expects context value
		handler := func(ctx context.Context, _ struct{}, _ struct{}, body RequestBody) (map[string]string, error) {
			userID := ctx.Value("userID")
			if userID == nil {
				return nil, errors.New("userID not found in context")
			}
			return map[string]string{
				"userID": userID.(string),
				"name":   body.Name,
			}, nil
		}

		// Create validated handler with body validation
		validatedHandler := CreateValidatedHandler(handler, nil, nil, bodySchema, nil)
		router.POST("/test", validatedHandler)

		// Test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Should succeed with context value available even with validation
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"userID":"123"`)
	})
}

// mockSchema for testing
type mockSchema struct {
	validateFunc func(data interface{}) error
}

func (m mockSchema) Validate(data interface{}) error {
	if m.validateFunc != nil {
		return m.validateFunc(data)
	}
	return nil
}
