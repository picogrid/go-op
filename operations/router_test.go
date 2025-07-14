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
	"github.com/picogrid/go-op"
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

// TestNewRouter tests router creation and initialization
func TestNewRouter(t *testing.T) {
	t.Run("Create router with valid engine", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		
		router := NewRouter(engine, generator)
		
		if router == nil {
			t.Error("Expected router to be created")
		}
		
		if router.engine != engine {
			t.Error("Expected engine to be set correctly")
		}
		
		if len(router.generators) != 1 {
			t.Errorf("Expected 1 generator, got %d", len(router.generators))
		}
		
		if len(router.operations) != 0 {
			t.Errorf("Expected empty operations slice, got %d operations", len(router.operations))
		}
	})

	t.Run("Create router with multiple generators", func(t *testing.T) {
		engine := createTestEngine()
		gen1 := &mockGenerator{}
		gen2 := &mockGenerator{}
		
		router := NewRouter(engine, gen1, gen2)
		
		if len(router.generators) != 2 {
			t.Errorf("Expected 2 generators, got %d", len(router.generators))
		}
	})

	t.Run("Create router with no generators", func(t *testing.T) {
		engine := createTestEngine()
		
		router := NewRouter(engine)
		
		if len(router.generators) != 0 {
			t.Errorf("Expected 0 generators, got %d", len(router.generators))
		}
	})

	t.Run("Create router with nil engine", func(t *testing.T) {
		generator := &mockGenerator{}
		
		router := NewRouter(nil, generator)
		
		if router.engine != nil {
			t.Error("Expected engine to be nil")
		}
		
		// Should not panic, just store nil engine
		if len(router.generators) != 1 {
			t.Errorf("Expected 1 generator, got %d", len(router.generators))
		}
	})
}

// TestRouterRegister tests operation registration
func TestRouterRegister(t *testing.T) {
	t.Run("Register basic operation successfully", func(t *testing.T) {
		engine := createTestEngine()
		generator := &mockGenerator{}
		router := NewRouter(engine, generator)

		// Create a simple handler
		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		}

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
		if len(router.operations) != 1 {
			t.Errorf("Expected 1 operation, got %d", len(router.operations))
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
		router := NewRouter(engine, generator)

		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "secure"})
		}

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
		router := NewRouter(engine, generator)

		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "enhanced"})
		}

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
		router := NewRouter(engine, generator)

		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "basic"})
		}

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
		router := NewRouter(engine, generator)

		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		}

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
		router := NewRouter(engine, generator)

		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		}

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

		if len(router.operations) != 3 {
			t.Errorf("Expected 3 operations, got %d", len(router.operations))
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
		router := NewRouter(engine)

		ops := router.GetOperations()
		if len(ops) != 0 {
			t.Errorf("Expected 0 operations, got %d", len(ops))
		}
	})

	t.Run("Get operations after registration", func(t *testing.T) {
		engine := createTestEngine()
		router := NewRouter(engine)

		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		}

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
		router := NewRouter(engine)

		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		}

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
		ginHandler := CreateValidatedHandler(
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

		ginHandler := CreateValidatedHandler(
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

		ginHandler := CreateValidatedHandler(
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

		ginHandler := CreateValidatedHandler(
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

		ginHandler := CreateValidatedHandler(
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

		ginHandler := CreateValidatedHandler(
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

		ginHandler := CreateValidatedHandler(
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

		middleware := ValidationMiddleware(paramsSchema, querySchema, bodySchema)

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

		middleware := ValidationMiddleware(paramsSchema, nil, nil)

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
		middleware := ValidationMiddleware(nil, nil, nil)

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
		handler := JSONResponse(data)

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
		handler := ErrorResponse(http.StatusBadRequest, "Bad request")

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
		router := NewRouter(engine, generator)

		// Register some operations
		handler := func(c *gin.Context) {}
		
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
		router := NewRouter(engine, generator)

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
		router := NewRouter(engine, generator)

		// Create handler with security
		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "secure endpoint",
				"user":    "authenticated",
			})
		}

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