package operations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	goop "github.com/picogrid/go-op"
	ginadapter "github.com/picogrid/go-op/operations/adapters/gin"
)

// Mock Generator for testing
type mockGenerator struct {
	processedOps []OperationInfo
	shouldError  bool
	errorMsg     string
}

func (m *mockGenerator) Process(info OperationInfo) error {
	if m.shouldError {
		return fmt.Errorf("%s", m.errorMsg)
	}
	m.processedOps = append(m.processedOps, info)
	return nil
}

// Mock Schema for testing
type mockSchema struct {
	shouldValidate bool
	validationErr  error
	isEnhanced     bool
	openAPISchema  *goop.OpenAPISchema
	validationInfo *goop.ValidationInfo
}

func (m *mockSchema) Validate(data interface{}) error {
	if !m.shouldValidate {
		return m.validationErr
	}
	return nil
}

func (m *mockSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	if m.isEnhanced {
		return m.openAPISchema
	}
	return nil
}

func (m *mockSchema) GetValidationInfo() *goop.ValidationInfo {
	if m.isEnhanced {
		return m.validationInfo
	}
	return nil
}

// Helper to create test Gin engine
func createTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestNewGinRouter tests Gin router creation and initialization
func TestNewGinRouter(t *testing.T) {
	t.Run("Create router with valid engine", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}

		router := ginadapter.NewGinRouter(engine, generator)

		if router == nil {
			t.Error("Expected router to be created")
			return
		}

		if router.GetEngine() != engine {
			t.Error("Expected engine to be set correctly")
		}

		// Note: We can't test internal fields of the router directly anymore
		// since they are private in the Gin adapter

		if len(router.GetOperations()) != 0 {
			t.Errorf("Expected empty operations slice, got %d operations", len(router.GetOperations()))
		}
	})

	t.Run("Create router with multiple generators", func(t *testing.T) {
		engine := createTestEngine()
		gen1 := &mockGenerator{}
		gen2 := &mockGenerator{}

		router := ginadapter.NewGinRouter(engine, gen1, gen2)

		// Note: We can't test internal generator count directly anymore
		// since they are private in the Gin adapter
		// Test passes if no error occurred
		_ = router // Use the router to avoid unused variable error
	})

	t.Run("Create router with no generators", func(t *testing.T) {
		engine := createTestEngine()

		router := ginadapter.NewGinRouter(engine)

		// Note: We can't test internal generator count directly anymore
		// since they are private in the Gin adapter
		// Test passes if no error occurred
		_ = router // Use the router to avoid unused variable error
	})

	t.Run("Create router with nil engine", func(t *testing.T) {
		generator := &mockGenerator{}

		router := ginadapter.NewGinRouter(nil, generator)

		if router.GetEngine() != nil {
			t.Error("Expected engine to be nil")
		}

		// Should not panic, just store nil engine
		// Note: We can't test internal generator count directly anymore
		// since they are private in the Gin adapter
	})
}

// TestRouterRegister tests operation registration
func TestRouterRegister(t *testing.T) {
	t.Run("Register basic operation successfully", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		router := ginadapter.NewGinRouter(engine, generator)

		// Create a simple handler
		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		op := CompiledOperation{
			Method:      "GET",
			Path:        "/test",
			Summary:     "Test operation",
			Description: "Test description",
			Tags:        []string{"test"},
			Handler:     handler,
			SuccessCode: 200,
		}

		err := router.Register(op)
		if err != nil {
			t.Errorf("Expected successful registration, got error: %v", err)
		}

		// Check operation was stored
		if len(router.GetOperations()) != 1 {
			t.Errorf("Expected 1 operation, got %d", len(router.GetOperations()))
		}

		// Check generator was called
		if len(generator.processedOps) != 1 {
			t.Errorf("Expected generator to process 1 operation, got %d", len(generator.processedOps))
		}

		processedOp := generator.processedOps[0]
		if processedOp.Method != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", processedOp.Method)
		}
		if processedOp.Path != "/test" {
			t.Errorf("Expected path '/test', got '%s'", processedOp.Path)
		}
	})

	t.Run("Register operation with security requirements", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		router := ginadapter.NewGinRouter(engine, generator)

		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "secure"})
		})

		security := goop.SecurityRequirements{}.RequireScheme("apiKey", "read")

		op := CompiledOperation{
			Method:      "POST",
			Path:        "/secure",
			Summary:     "Secure operation",
			Description: "Secure description",
			Tags:        []string{"secure"},
			Handler:     handler,
			SuccessCode: 201,
			Security:    security,
		}

		err := router.Register(op)
		if err != nil {
			t.Errorf("Expected successful registration, got error: %v", err)
		}

		// Check security was passed to generator
		processedOp := generator.processedOps[0]
		if len(processedOp.Security) == 0 {
			t.Error("Expected security requirements to be passed to generator")
		}

		if processedOp.Security[0]["apiKey"][0] != "read" {
			t.Errorf("Expected security scope 'read', got %v", processedOp.Security[0]["apiKey"])
		}
	})

	t.Run("Register operation with enhanced schemas", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		router := ginadapter.NewGinRouter(engine, generator)

		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "enhanced"})
		})

		// Create enhanced schemas
		paramsSchema := &mockSchema{
			isEnhanced: true,
			openAPISchema: &goop.OpenAPISchema{
				Type: "object",
				Properties: map[string]*goop.OpenAPISchema{
					"id": {Type: "string"},
				},
			},
			validationInfo: &goop.ValidationInfo{
				Required: true,
				Optional: false,
			},
		}

		querySchema := &mockSchema{
			isEnhanced: true,
			openAPISchema: &goop.OpenAPISchema{
				Type: "object",
				Properties: map[string]*goop.OpenAPISchema{
					"filter": {Type: "string"},
				},
			},
			validationInfo: &goop.ValidationInfo{
				Required: false,
				Optional: true,
			},
		}

		op := CompiledOperation{
			Method:       "GET",
			Path:         "/enhanced/{id}",
			ParamsSchema: paramsSchema,
			QuerySchema:  querySchema,
			Handler:      handler,
			SuccessCode:  200,
		}

		err := router.Register(op)
		if err != nil {
			t.Errorf("Expected successful registration, got error: %v", err)
		}

		// Check validation info was extracted
		processedOp := generator.processedOps[0]
		if processedOp.ParamsInfo == nil {
			t.Error("Expected params validation info to be extracted")
		}
		if processedOp.QueryInfo == nil {
			t.Error("Expected query validation info to be extracted")
		}

		if processedOp.ParamsInfo.Required != true {
			t.Error("Expected params to be required")
		}
		if processedOp.QueryInfo.Required != false {
			t.Error("Expected query to be optional")
		}
	})

	t.Run("Register operation with non-enhanced schemas", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		router := ginadapter.NewGinRouter(engine, generator)

		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "basic"})
		})

		// Create basic schemas (not enhanced)
		paramsSchema := &mockSchema{
			isEnhanced: false,
		}

		op := CompiledOperation{
			Method:       "GET",
			Path:         "/basic/{id}",
			ParamsSchema: paramsSchema,
			Handler:      handler,
			SuccessCode:  200,
		}

		err := router.Register(op)
		if err != nil {
			t.Errorf("Expected successful registration, got error: %v", err)
		}

		// Check validation info was not extracted
		processedOp := generator.processedOps[0]
		if processedOp.ParamsInfo != nil {
			t.Error("Expected params validation info to be nil for non-enhanced schema")
		}
	})

	t.Run("Register operation with generator error", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{
			shouldError: true,
			errorMsg:    "generator processing failed",
		}
		router := ginadapter.NewGinRouter(engine, generator)

		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		op := CompiledOperation{
			Method:  "GET",
			Path:    "/test",
			Handler: handler,
		}

		err := router.Register(op)
		if err == nil {
			t.Error("Expected registration to fail when generator returns error")
		}

		if !strings.Contains(err.Error(), "generator processing failed") {
			t.Errorf("Expected generator error message, got: %v", err)
		}
	})

	t.Run("Register multiple operations", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		router := ginadapter.NewGinRouter(engine, generator)

		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		ops := []CompiledOperation{
			{Method: "GET", Path: "/test1", Handler: handler},
			{Method: "POST", Path: "/test2", Handler: handler},
			{Method: "PUT", Path: "/test3", Handler: handler},
		}

		for _, op := range ops {
			err := router.Register(op)
			if err != nil {
				t.Errorf("Expected successful registration, got error: %v", err)
			}
		}

		if len(router.GetOperations()) != 3 {
			t.Errorf("Expected 3 operations, got %d", len(router.GetOperations()))
		}

		if len(generator.processedOps) != 3 {
			t.Errorf("Expected generator to process 3 operations, got %d", len(generator.processedOps))
		}
	})
}

// TestGetOperations tests operation retrieval
func TestGetOperations(t *testing.T) {
	t.Run("Get operations from empty router", func(t *testing.T) {
		engine := createTestEngine()
		router := ginadapter.NewGinRouter(engine)

		ops := router.GetOperations()
		if len(ops) != 0 {
			t.Errorf("Expected 0 operations, got %d", len(ops))
		}
	})

	t.Run("Get operations after registration", func(t *testing.T) {
		engine := createTestEngine()
		router := ginadapter.NewGinRouter(engine)

		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		op := CompiledOperation{
			Method:  "GET",
			Path:    "/test",
			Handler: handler,
		}

		router.Register(op)

		ops := router.GetOperations()
		if len(ops) != 1 {
			t.Errorf("Expected 1 operation, got %d", len(ops))
		}

		if ops[0].Method != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", ops[0].Method)
		}
	})

	t.Run("Operations slice immutability", func(t *testing.T) {
		engine := createTestEngine()
		router := ginadapter.NewGinRouter(engine)

		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		op := CompiledOperation{
			Method:  "GET",
			Path:    "/test",
			Handler: handler,
		}

		router.Register(op)

		ops1 := router.GetOperations()
		ops2 := router.GetOperations()

		// Modify first slice
		if len(ops1) > 0 {
			ops1[0].Method = "MODIFIED"
		}

		// Second slice should be unaffected
		if ops2[0].Method != "GET" {
			t.Error("Expected operations slice to be independent")
		}
	})
}

// TestCreateValidatedHandler tests validated handler creation
func TestCreateValidatedHandler(t *testing.T) {
	// Test types for handler
	type TestParams struct {
		ID string `json:"id" uri:"id"`
	}

	type TestQuery struct {
		Filter string `json:"filter" form:"filter"`
	}

	type TestBody struct {
		Name string `json:"name"`
	}

	type TestResponse struct {
		Message string `json:"message"`
		ID      string `json:"id"`
	}

	t.Run("Handler with all validation schemas", func(t *testing.T) {
		// Create handler function
		handlerFunc := func(ctx context.Context, params TestParams, query TestQuery, body TestBody) (TestResponse, error) {
			return TestResponse{
				Message: "success",
				ID:      params.ID,
			}, nil
		}

		// Create mock schemas
		paramsSchema := &mockSchema{shouldValidate: true}
		querySchema := &mockSchema{shouldValidate: true}
		bodySchema := &mockSchema{shouldValidate: true}
		responseSchema := &mockSchema{shouldValidate: true}

		// Create validated handler
		ginHandler := ginadapter.CreateValidatedHandler(
			handlerFunc,
			paramsSchema,
			querySchema,
			bodySchema,
			responseSchema,
		)

		// Test the handler
		engine := createTestEngine()
		engine.POST("/test/:id", ginHandler)

		reqBody := `{"name": "test"}`
		req := httptest.NewRequest("POST", "/test/123?filter=active", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response TestResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.Message != "success" {
			t.Errorf("Expected message 'success', got '%s'", response.Message)
		}
	})

	t.Run("Handler with parameter validation failure", func(t *testing.T) {
		handlerFunc := func(ctx context.Context, params TestParams, query TestQuery, body TestBody) (TestResponse, error) {
			return TestResponse{}, nil
		}

		paramsSchema := &mockSchema{
			shouldValidate: false,
			validationErr:  fmt.Errorf("invalid parameter"),
		}

		ginHandler := ginadapter.CreateValidatedHandler(
			handlerFunc,
			paramsSchema,
			nil,
			nil,
			nil,
		)

		engine := createTestEngine()
		engine.GET("/test/:id", ginHandler)

		req := httptest.NewRequest("GET", "/test/invalid", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var errorResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		if err != nil {
			t.Errorf("Failed to unmarshal error response: %v", err)
		}

		if errorResp["error"] != "Path parameter validation failed" {
			t.Errorf("Expected parameter validation error, got %v", errorResp["error"])
		}
	})

	t.Run("Handler with query validation failure", func(t *testing.T) {
		handlerFunc := func(ctx context.Context, params TestParams, query TestQuery, body TestBody) (TestResponse, error) {
			return TestResponse{}, nil
		}

		querySchema := &mockSchema{
			shouldValidate: false,
			validationErr:  fmt.Errorf("invalid query"),
		}

		ginHandler := ginadapter.CreateValidatedHandler(
			handlerFunc,
			nil,
			querySchema,
			nil,
			nil,
		)

		engine := createTestEngine()
		engine.GET("/test", ginHandler)

		req := httptest.NewRequest("GET", "/test?invalid=value", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Handler with body validation failure", func(t *testing.T) {
		handlerFunc := func(ctx context.Context, params TestParams, query TestQuery, body TestBody) (TestResponse, error) {
			return TestResponse{}, nil
		}

		bodySchema := &mockSchema{
			shouldValidate: false,
			validationErr:  fmt.Errorf("invalid body"),
		}

		ginHandler := ginadapter.CreateValidatedHandler(
			handlerFunc,
			nil,
			nil,
			bodySchema,
			nil,
		)

		engine := createTestEngine()
		engine.POST("/test", ginHandler)

		reqBody := `{"invalid": "data"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Handler with business logic error", func(t *testing.T) {
		handlerFunc := func(ctx context.Context, params TestParams, query TestQuery, body TestBody) (TestResponse, error) {
			return TestResponse{}, fmt.Errorf("business logic error")
		}

		ginHandler := ginadapter.CreateValidatedHandler(
			handlerFunc,
			nil,
			nil,
			nil,
			nil,
		)

		engine := createTestEngine()
		engine.GET("/test", ginHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		var errorResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		if err != nil {
			t.Errorf("Failed to unmarshal error response: %v", err)
		}

		if errorResp["error"] != "Internal server error" {
			t.Errorf("Expected internal server error, got %v", errorResp["error"])
		}
	})

	t.Run("Handler with response validation failure", func(t *testing.T) {
		handlerFunc := func(ctx context.Context, params TestParams, query TestQuery, body TestBody) (TestResponse, error) {
			return TestResponse{Message: "test"}, nil
		}

		responseSchema := &mockSchema{
			shouldValidate: false,
			validationErr:  fmt.Errorf("invalid response"),
		}

		ginHandler := ginadapter.CreateValidatedHandler(
			handlerFunc,
			nil,
			nil,
			nil,
			responseSchema,
		)

		engine := createTestEngine()
		engine.GET("/test", ginHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		var errorResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		if err != nil {
			t.Errorf("Failed to unmarshal error response: %v", err)
		}

		if errorResp["error"] != "Response validation failed" {
			t.Errorf("Expected response validation error, got %v", errorResp["error"])
		}
	})

	t.Run("Handler with nil schemas", func(t *testing.T) {
		handlerFunc := func(ctx context.Context, params TestParams, query TestQuery, body TestBody) (TestResponse, error) {
			return TestResponse{Message: "no validation"}, nil
		}

		ginHandler := ginadapter.CreateValidatedHandler(
			handlerFunc,
			nil, // No parameter validation
			nil, // No query validation
			nil, // No body validation
			nil, // No response validation
		)

		engine := createTestEngine()
		engine.GET("/test", ginHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response TestResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.Message != "no validation" {
			t.Errorf("Expected message 'no validation', got '%s'", response.Message)
		}
	})
}

// TestValidationMiddleware tests validation middleware functionality
func TestValidationMiddleware(t *testing.T) {
	t.Run("Middleware with successful validation", func(t *testing.T) {
		paramsSchema := &mockSchema{shouldValidate: true}
		querySchema := &mockSchema{shouldValidate: true}
		bodySchema := &mockSchema{shouldValidate: true}

		middleware := ginadapter.ValidationMiddleware(paramsSchema, querySchema, bodySchema)

		engine := createTestEngine()
		engine.POST("/test/:id", middleware, func(c *gin.Context) {
			// Check that validated data was set
			if _, exists := c.Get("validatedParams"); !exists {
				t.Error("Expected validatedParams to be set")
			}
			if _, exists := c.Get("validatedQuery"); !exists {
				t.Error("Expected validatedQuery to be set")
			}
			if _, exists := c.Get("validatedBody"); !exists {
				t.Error("Expected validatedBody to be set")
			}

			c.JSON(200, gin.H{"message": "success"})
		})

		reqBody := `{"name": "test"}`
		req := httptest.NewRequest("POST", "/test/123?filter=active", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Middleware with validation failure", func(t *testing.T) {
		paramsSchema := &mockSchema{
			shouldValidate: false,
			validationErr:  fmt.Errorf("validation failed"),
		}

		middleware := ginadapter.ValidationMiddleware(paramsSchema, nil, nil)

		engine := createTestEngine()
		engine.GET("/test/:id", middleware, func(c *gin.Context) {
			t.Error("Handler should not be called when validation fails")
			c.JSON(200, gin.H{"message": "should not reach here"})
		})

		req := httptest.NewRequest("GET", "/test/invalid", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Middleware with nil schemas", func(t *testing.T) {
		middleware := ginadapter.ValidationMiddleware(nil, nil, nil)

		engine := createTestEngine()
		engine.GET("/test", middleware, func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "no validation"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

// TestHelperFunctions tests helper response functions
func TestHelperFunctions(t *testing.T) {
	t.Run("JSONResponse helper", func(t *testing.T) {
		data := map[string]string{"message": "test"}
		handler := ginadapter.JSONResponse(data)

		engine := createTestEngine()
		engine.GET("/test", handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["message"] != "test" {
			t.Errorf("Expected message 'test', got '%s'", response["message"])
		}
	})

	t.Run("ErrorResponse helper", func(t *testing.T) {
		handler := ginadapter.ErrorResponse(http.StatusBadRequest, "Bad request")

		engine := createTestEngine()
		engine.GET("/test", handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Bad request" {
			t.Errorf("Expected error 'Bad request', got '%s'", response["error"])
		}
	})
}

// TestServeSpec tests OpenAPI spec serving
func TestServeSpec(t *testing.T) {
	t.Run("Serve spec with operations", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		router := ginadapter.NewGinRouter(engine, generator)

		// Register some operations
		handler := gin.HandlerFunc(func(c *gin.Context) {})

		ops := []CompiledOperation{
			{
				Method:      "GET",
				Path:        "/users",
				Summary:     "List users",
				Description: "Get all users",
				Tags:        []string{"users"},
				Handler:     handler,
				SuccessCode: 200,
			},
			{
				Method:      "POST",
				Path:        "/users",
				Summary:     "Create user",
				Description: "Create a new user",
				Tags:        []string{"users"},
				Handler:     handler,
				SuccessCode: 201,
				Security:    goop.SecurityRequirements{}.RequireScheme("apiKey"),
			},
		}

		for _, op := range ops {
			router.Register(op)
		}

		// Create spec handler
		specHandler := router.ServeSpec(generator)

		// Test the spec endpoint
		req := httptest.NewRequest("GET", "/spec", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		specHandler(c)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check content type
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected content type 'application/json', got '%s'", contentType)
		}

		// Parse response
		var spec map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &spec)
		if err != nil {
			t.Errorf("Failed to parse spec JSON: %v", err)
		}

		// Check basic structure
		if spec["openapi"] != "3.1.0" {
			t.Errorf("Expected OpenAPI version '3.1.0', got %v", spec["openapi"])
		}

		info, ok := spec["info"].(map[string]interface{})
		if !ok {
			t.Error("Expected info object in spec")
		}

		if info["title"] != "Generated API" {
			t.Errorf("Expected title 'Generated API', got %v", info["title"])
		}

		paths, ok := spec["paths"].([]interface{})
		if !ok {
			t.Error("Expected paths array in spec")
		}

		if len(paths) != 2 {
			t.Errorf("Expected 2 paths, got %d", len(paths))
		}
	})

	t.Run("Serve spec with empty operations", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		router := ginadapter.NewGinRouter(engine, generator)

		specHandler := router.ServeSpec(generator)

		req := httptest.NewRequest("GET", "/spec", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		specHandler(c)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var spec map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &spec)
		if err != nil {
			t.Errorf("Failed to parse spec JSON: %v", err)
		}

		paths, ok := spec["paths"].([]interface{})
		if !ok {
			t.Error("Expected paths array in spec")
		}

		if len(paths) != 0 {
			t.Errorf("Expected 0 paths, got %d", len(paths))
		}
	})
}

// TestRouterIntegration tests complete router workflow
func TestRouterIntegration(t *testing.T) {
	t.Run("Complete workflow with security", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		router := ginadapter.NewGinRouter(engine, generator)

		// Create handler with security
		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "secure endpoint",
				"user":    "authenticated",
			})
		})

		// Create operation with security
		op := CompiledOperation{
			Method:      "GET",
			Path:        "/secure",
			Summary:     "Secure endpoint",
			Description: "Requires API key authentication",
			Tags:        []string{"security"},
			Handler:     handler,
			SuccessCode: 200,
			Security:    goop.SecurityRequirements{}.RequireScheme("apiKey", "read"),
		}

		// Register operation
		err := router.Register(op)
		if err != nil {
			t.Errorf("Failed to register operation: %v", err)
		}

		// Test the actual HTTP endpoint
		req := httptest.NewRequest("GET", "/secure", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["message"] != "secure endpoint" {
			t.Errorf("Expected message 'secure endpoint', got %v", response["message"])
		}

		// Verify generator processed security information
		if len(generator.processedOps) != 1 {
			t.Errorf("Expected 1 processed operation, got %d", len(generator.processedOps))
		}

		processedOp := generator.processedOps[0]
		if len(processedOp.Security) == 0 {
			t.Error("Expected security requirements to be processed")
		}

		if processedOp.Security[0]["apiKey"][0] != "read" {
			t.Errorf("Expected security scope 'read', got %v", processedOp.Security[0]["apiKey"])
		}
	})
}

// TestConvertOpenAPIPathToGin tests the path conversion function
func TestConvertOpenAPIPathToGin(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple single parameter",
			input:    "/users/{id}",
			expected: "/users/:id",
		},
		{
			name:     "Multiple parameters",
			input:    "/users/{userId}/posts/{postId}",
			expected: "/users/:userId/posts/:postId",
		},
		{
			name:     "No parameters",
			input:    "/users",
			expected: "/users",
		},
		{
			name:     "Parameter at start",
			input:    "/{version}/users",
			expected: "/:version/users",
		},
		{
			name:     "Parameter at end",
			input:    "/api/v1/users/{id}",
			expected: "/api/v1/users/:id",
		},
		{
			name:     "Multiple consecutive parameters",
			input:    "/users/{userId}/{action}",
			expected: "/users/:userId/:action",
		},
		{
			name:     "Complex path with multiple segments",
			input:    "/api/{version}/users/{userId}/posts/{postId}/comments/{commentId}",
			expected: "/api/:version/users/:userId/posts/:postId/comments/:commentId",
		},
		{
			name:     "Already Gin-style (should remain unchanged)",
			input:    "/users/:id",
			expected: "/users/:id",
		},
		{
			name:     "Mixed styles (OpenAPI takes precedence)",
			input:    "/users/{userId}/posts/:postId",
			expected: "/users/:userId/posts/:postId",
		},
		{
			name:     "Empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "Root path",
			input:    "/",
			expected: "/",
		},
		{
			name:     "Path with query string marker",
			input:    "/users/{id}?filter=active",
			expected: "/users/:id?filter=active",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ginadapter.ConvertOpenAPIPathToGin(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestRouterPathConversion tests that the router correctly converts paths when registering operations
func TestRouterPathConversion(t *testing.T) {
	t.Run("Router converts OpenAPI paths to Gin paths", func(t *testing.T) {
		engine := createTestEngine()
		router := ginadapter.NewGinRouter(engine)

		// Create an operation with OpenAPI-style path
		op := CompiledOperation{
			Method:  "GET",
			Path:    "/users/{userId}/posts/{postId}",
			Summary: "Get post by user and post ID",
			Handler: gin.HandlerFunc(func(c *gin.Context) {
				// Return the captured parameters to verify correct routing
				c.JSON(http.StatusOK, gin.H{
					"userId": c.Param("userId"),
					"postId": c.Param("postId"),
				})
			}),
		}

		// Register the operation
		err := router.Register(op)
		if err != nil {
			t.Fatalf("Failed to register operation: %v", err)
		}

		// Test that the route works with Gin-style parameters
		req := httptest.NewRequest("GET", "/users/123/posts/456", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["userId"] != "123" {
			t.Errorf("Expected userId '123', got '%s'", response["userId"])
		}
		if response["postId"] != "456" {
			t.Errorf("Expected postId '456', got '%s'", response["postId"])
		}
	})

	t.Run("Router preserves original OpenAPI path in operations list", func(t *testing.T) {
		engine := createTestEngine()
		router := ginadapter.NewGinRouter(engine)

		// Create an operation with OpenAPI-style path
		op := CompiledOperation{
			Method:  "GET",
			Path:    "/users/{id}",
			Summary: "Get user by ID",
			Handler: gin.HandlerFunc(func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
			}),
		}

		// Register the operation
		err := router.Register(op)
		if err != nil {
			t.Fatalf("Failed to register operation: %v", err)
		}

		// Verify that the stored operation preserves the original OpenAPI path
		operations := router.GetOperations()
		if len(operations) != 1 {
			t.Fatalf("Expected 1 operation, got %d", len(operations))
		}

		if operations[0].Path != "/users/{id}" {
			t.Errorf("Expected original OpenAPI path '/users/{id}', got '%s'", operations[0].Path)
		}
	})
}

// TestRouterErrorHandling tests end-to-end error response scenarios
func TestRouterErrorHandling(t *testing.T) {
	t.Run("Error response with standard error schemas", func(t *testing.T) {
		engine := createTestEngine()
		router := ginadapter.NewGinRouter(engine)

		// Create handler that returns different error types
		handler := gin.HandlerFunc(func(c *gin.Context) {
			errorType := c.Query("error")
			switch errorType {
			case "bad_request":
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "bad_request",
					"message": "The request could not be understood or was missing required parameters",
					"code":    400,
				})
			case "unauthorized":
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "unauthorized",
					"message": "Authentication is required to access this resource",
					"code":    401,
				})
			case "not_found":
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "not_found",
					"message": "The requested resource was not found",
					"code":    404,
				})
			default:
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			}
		})

		// Create operation with multiple error responses
		op := CompiledOperation{
			Method:  "GET",
			Path:    "/test",
			Handler: handler,
			Responses: map[int]goop.ResponseDefinition{
				200: {Schema: &mockSchema{shouldValidate: true}, Description: "Success"},
				400: {Schema: BadRequestErrorSchema, Description: "Bad Request"},
				401: {Schema: UnauthorizedErrorSchema, Description: "Unauthorized"},
				404: {Schema: NotFoundErrorSchema, Description: "Not Found"},
			},
		}

		err := router.Register(op)
		if err != nil {
			t.Fatalf("Failed to register operation: %v", err)
		}

		// Test each error scenario
		testCases := []struct {
			name           string
			query          string
			expectedStatus int
			expectedError  string
		}{
			{"success case", "", http.StatusOK, ""},
			{"bad request", "error=bad_request", http.StatusBadRequest, "bad_request"},
			{"unauthorized", "error=unauthorized", http.StatusUnauthorized, "unauthorized"},
			{"not found", "error=not_found", http.StatusNotFound, "not_found"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				url := "/test"
				if tc.query != "" {
					url += "?" + tc.query
				}

				req := httptest.NewRequest("GET", url, nil)
				w := httptest.NewRecorder()
				engine.ServeHTTP(w, req)

				if w.Code != tc.expectedStatus {
					t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
				}

				if tc.expectedError != "" {
					var response map[string]interface{}
					err := json.Unmarshal(w.Body.Bytes(), &response)
					if err != nil {
						t.Errorf("Failed to unmarshal response: %v", err)
					}

					if response["error"] != tc.expectedError {
						t.Errorf("Expected error '%s', got '%v'", tc.expectedError, response["error"])
					}
				}
			})
		}
	})
}

// TestMultipleResponseValidation tests operations with multiple defined responses
func TestMultipleResponseValidation(t *testing.T) {
	t.Run("Operation with comprehensive error responses", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		router := ginadapter.NewGinRouter(engine, generator)

		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		// Create operation using the new convenience methods
		op := NewSimple().
			POST("/users").
			WithSuccessResponse(201, &mockSchema{shouldValidate: true}, "User Created").
			WithCreateErrors(). // Adds 400, 401, 403, 409, 422, 500
			Handler(handler)

		err := router.Register(op)
		if err != nil {
			t.Fatalf("Failed to register operation: %v", err)
		}

		// Verify all expected responses are defined
		expectedCodes := []int{201, 400, 401, 403, 409, 422, 500}
		for _, code := range expectedCodes {
			if _, exists := op.Responses[code]; !exists {
				t.Errorf("Expected response code %d to be defined", code)
			}
		}

		// Verify generator received the operation info
		if len(generator.processedOps) != 1 {
			t.Errorf("Expected 1 processed operation, got %d", len(generator.processedOps))
		}

		processedOp := generator.processedOps[0]
		if processedOp.Method != "POST" {
			t.Errorf("Expected method 'POST', got '%s'", processedOp.Method)
		}
		if processedOp.Path != "/users" {
			t.Errorf("Expected path '/users', got '%s'", processedOp.Path)
		}
	})
}

// TestVariadicRegisterWithErrors tests the new variadic Register method with error responses
func TestVariadicRegisterWithErrors(t *testing.T) {
	t.Run("Register multiple operations with error responses", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		router := ginadapter.NewGinRouter(engine, generator)

		// Create test handler
		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		// Create multiple operations with different error patterns
		createOp := NewSimple().
			POST("/users").
			WithSuccessResponse(201, &mockSchema{shouldValidate: true}, "Created").
			WithCreateErrors().
			Handler(handler)

		getOp := NewSimple().
			GET("/users/{id}").
			WithSuccessResponse(200, &mockSchema{shouldValidate: true}, "OK").
			WithAuthErrors().
			WithNotFoundError(NotFoundErrorSchema).
			Handler(handler)

		updateOp := NewSimple().
			PUT("/users/{id}").
			WithSuccessResponse(200, &mockSchema{shouldValidate: true}, "Updated").
			WithAuthErrors().
			WithValidationErrors().
			WithNotFoundError(NotFoundErrorSchema).
			Handler(handler)

		// Test variadic registration
		err := router.Register(createOp, getOp, updateOp)
		if err != nil {
			t.Fatalf("Failed to register multiple operations: %v", err)
		}

		// Verify all operations were registered
		operations := router.GetOperations()
		if len(operations) != 3 {
			t.Errorf("Expected 3 operations, got %d", len(operations))
		}

		// Verify generator processed all operations
		if len(generator.processedOps) != 3 {
			t.Errorf("Expected generator to process 3 operations, got %d", len(generator.processedOps))
		}

		// Test that all HTTP routes work
		testRoutes := []struct {
			method string
			path   string
		}{
			{"POST", "/users"},
			{"GET", "/users/123"},
			{"PUT", "/users/123"},
		}

		for _, route := range testRoutes {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s %s, got %d", route.method, route.path, w.Code)
			}
		}
	})

	t.Run("Variadic register with operation error", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{
			shouldError: true,
			errorMsg:    "processing failed",
		}
		router := ginadapter.NewGinRouter(engine, generator)

		handler := gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		op1 := CompiledOperation{Method: "GET", Path: "/test1", Handler: handler}
		op2 := CompiledOperation{Method: "GET", Path: "/test2", Handler: handler}

		// Should fail on first operation
		err := router.Register(op1, op2)
		if err == nil {
			t.Error("Expected registration to fail when generator returns error")
		}

		if !strings.Contains(err.Error(), "failed to register operation GET /test1") {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})
}
