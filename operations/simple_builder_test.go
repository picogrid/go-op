package operations

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/picogrid/go-op"
)

// TestNewSimple tests simple builder creation
func TestNewSimple(t *testing.T) {
	t.Run("Create new simple builder", func(t *testing.T) {
		builder := NewSimple()
		
		if builder == nil {
			t.Error("Expected builder to be created")
		}
		
		if builder.config == nil {
			t.Error("Expected config to be initialized")
		}
		
		// Check default values
		if len(builder.config.tags) != 0 {
			t.Errorf("Expected empty tags slice, got %d tags", len(builder.config.tags))
		}
		
		if builder.config.successCode != 200 {
			t.Errorf("Expected default success code 200, got %d", builder.config.successCode)
		}
		
		if builder.config.method != "" {
			t.Errorf("Expected empty method, got '%s'", builder.config.method)
		}
		
		if builder.config.path != "" {
			t.Errorf("Expected empty path, got '%s'", builder.config.path)
		}
	})
}

// TestSimpleBuilderMethods tests HTTP method setting
func TestSimpleBuilderMethods(t *testing.T) {
	t.Run("Method sets method and path", func(t *testing.T) {
		builder := NewSimple().Method("POST", "/users")
		
		if builder.config.method != "POST" {
			t.Errorf("Expected method 'POST', got '%s'", builder.config.method)
		}
		
		if builder.config.path != "/users" {
			t.Errorf("Expected path '/users', got '%s'", builder.config.path)
		}
		
		// Check fluent interface returns the same builder instance
		builder2 := builder.Summary("test")
		if builder != builder2 {
			t.Error("Expected fluent interface to return same builder instance")
		}
	})

	t.Run("GET sets method to GET", func(t *testing.T) {
		builder := NewSimple().GET("/users")
		
		if builder.config.method != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", builder.config.method)
		}
		
		if builder.config.path != "/users" {
			t.Errorf("Expected path '/users', got '%s'", builder.config.path)
		}
	})

	t.Run("POST sets method to POST", func(t *testing.T) {
		builder := NewSimple().POST("/users")
		
		if builder.config.method != "POST" {
			t.Errorf("Expected method 'POST', got '%s'", builder.config.method)
		}
		
		if builder.config.path != "/users" {
			t.Errorf("Expected path '/users', got '%s'", builder.config.path)
		}
	})

	t.Run("PUT sets method to PUT", func(t *testing.T) {
		builder := NewSimple().PUT("/users/1")
		
		if builder.config.method != "PUT" {
			t.Errorf("Expected method 'PUT', got '%s'", builder.config.method)
		}
		
		if builder.config.path != "/users/1" {
			t.Errorf("Expected path '/users/1', got '%s'", builder.config.path)
		}
	})

	t.Run("PATCH sets method to PATCH", func(t *testing.T) {
		builder := NewSimple().PATCH("/users/1")
		
		if builder.config.method != "PATCH" {
			t.Errorf("Expected method 'PATCH', got '%s'", builder.config.method)
		}
		
		if builder.config.path != "/users/1" {
			t.Errorf("Expected path '/users/1', got '%s'", builder.config.path)
		}
	})

	t.Run("DELETE sets method to DELETE", func(t *testing.T) {
		builder := NewSimple().DELETE("/users/1")
		
		if builder.config.method != "DELETE" {
			t.Errorf("Expected method 'DELETE', got '%s'", builder.config.method)
		}
		
		if builder.config.path != "/users/1" {
			t.Errorf("Expected path '/users/1', got '%s'", builder.config.path)
		}
	})
}

// TestSimpleBuilderMetadata tests metadata setting methods
func TestSimpleBuilderMetadata(t *testing.T) {
	t.Run("Summary sets operation summary", func(t *testing.T) {
		builder := NewSimple().Summary("List all users")
		
		if builder.config.summary != "List all users" {
			t.Errorf("Expected summary 'List all users', got '%s'", builder.config.summary)
		}
	})

	t.Run("Description sets operation description", func(t *testing.T) {
		builder := NewSimple().Description("Returns a list of all users in the system")
		
		if builder.config.description != "Returns a list of all users in the system" {
			t.Errorf("Expected description to be set correctly")
		}
	})

	t.Run("Tags adds tags to operation", func(t *testing.T) {
		builder := NewSimple().Tags("users", "admin")
		
		if len(builder.config.tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(builder.config.tags))
		}
		
		if builder.config.tags[0] != "users" || builder.config.tags[1] != "admin" {
			t.Errorf("Expected tags [users, admin], got %v", builder.config.tags)
		}
	})

	t.Run("Tags can be called multiple times", func(t *testing.T) {
		builder := NewSimple().Tags("users").Tags("admin", "api")
		
		if len(builder.config.tags) != 3 {
			t.Errorf("Expected 3 tags, got %d", len(builder.config.tags))
		}
		
		expectedTags := []string{"users", "admin", "api"}
		for i, tag := range expectedTags {
			if builder.config.tags[i] != tag {
				t.Errorf("Expected tag '%s' at index %d, got '%s'", tag, i, builder.config.tags[i])
			}
		}
	})

	t.Run("SuccessCode sets success HTTP status code", func(t *testing.T) {
		builder := NewSimple().SuccessCode(201)
		
		if builder.config.successCode != 201 {
			t.Errorf("Expected success code 201, got %d", builder.config.successCode)
		}
	})
}

// TestSimpleBuilderSchemas tests schema setting methods
func TestSimpleBuilderSchemas(t *testing.T) {
	mockSchema := &mockSchema{shouldValidate: true}

	t.Run("WithParams sets params schema", func(t *testing.T) {
		builder := NewSimple().WithParams(mockSchema)
		
		if builder.config.paramsSchema != mockSchema {
			t.Error("Expected params schema to be set")
		}
	})

	t.Run("WithQuery sets query schema", func(t *testing.T) {
		builder := NewSimple().WithQuery(mockSchema)
		
		if builder.config.querySchema != mockSchema {
			t.Error("Expected query schema to be set")
		}
	})

	t.Run("WithBody sets body schema", func(t *testing.T) {
		builder := NewSimple().WithBody(mockSchema)
		
		if builder.config.bodySchema != mockSchema {
			t.Error("Expected body schema to be set")
		}
	})

	t.Run("WithResponse sets response schema", func(t *testing.T) {
		builder := NewSimple().WithResponse(mockSchema)
		
		if builder.config.responseSchema != mockSchema {
			t.Error("Expected response schema to be set")
		}
	})

	t.Run("WithHeaders sets header schema", func(t *testing.T) {
		builder := NewSimple().WithHeaders(mockSchema)
		
		if builder.config.headerSchema != mockSchema {
			t.Error("Expected header schema to be set")
		}
	})
}

// TestSimpleBuilderSecurity tests security configuration methods
func TestSimpleBuilderSecurity(t *testing.T) {
	t.Run("WithSecurity sets security requirements", func(t *testing.T) {
		security := goop.SecurityRequirements{}.RequireScheme("apiKey", "read")
		builder := NewSimple().WithSecurity(security)
		
		if len(builder.config.security) != 1 {
			t.Errorf("Expected 1 security requirement, got %d", len(builder.config.security))
		}
		
		if builder.config.security[0]["apiKey"][0] != "read" {
			t.Errorf("Expected security scope 'read', got %v", builder.config.security[0]["apiKey"])
		}
	})

	t.Run("RequireAuth adds security requirement", func(t *testing.T) {
		builder := NewSimple().RequireAuth("apiKey", "read", "write")
		
		if len(builder.config.security) != 1 {
			t.Errorf("Expected 1 security requirement, got %d", len(builder.config.security))
		}
		
		if len(builder.config.security[0]["apiKey"]) != 2 {
			t.Errorf("Expected 2 scopes, got %d", len(builder.config.security[0]["apiKey"]))
		}
		
		if builder.config.security[0]["apiKey"][0] != "read" || builder.config.security[0]["apiKey"][1] != "write" {
			t.Errorf("Expected scopes [read, write], got %v", builder.config.security[0]["apiKey"])
		}
	})

	t.Run("RequireAuth with no existing security", func(t *testing.T) {
		builder := NewSimple()
		if builder.config.security != nil {
			t.Error("Expected nil security initially")
		}
		
		builder.RequireAuth("oauth2", "scope1")
		
		if builder.config.security == nil {
			t.Error("Expected security to be initialized")
		}
		
		if len(builder.config.security) != 1 {
			t.Errorf("Expected 1 security requirement, got %d", len(builder.config.security))
		}
	})

	t.Run("RequireAnyOf adds OR logic security schemes", func(t *testing.T) {
		builder := NewSimple().RequireAnyOf("apiKey", "oauth2", "bearer")
		
		if len(builder.config.security) != 3 {
			t.Errorf("Expected 3 security requirements (OR logic), got %d", len(builder.config.security))
		}
		
		// Check each requirement has the expected scheme
		schemes := []string{"apiKey", "oauth2", "bearer"}
		for i, scheme := range schemes {
			if _, exists := builder.config.security[i][scheme]; !exists {
				t.Errorf("Expected requirement %d to have scheme '%s'", i, scheme)
			}
		}
	})

	t.Run("RequireAPIKey is convenience for API key auth", func(t *testing.T) {
		builder := NewSimple().RequireAPIKey("myApiKey")
		
		if len(builder.config.security) != 1 {
			t.Errorf("Expected 1 security requirement, got %d", len(builder.config.security))
		}
		
		if _, exists := builder.config.security[0]["myApiKey"]; !exists {
			t.Error("Expected API key security requirement")
		}
	})

	t.Run("RequireBearer is convenience for Bearer auth", func(t *testing.T) {
		builder := NewSimple().RequireBearer("bearerToken")
		
		if len(builder.config.security) != 1 {
			t.Errorf("Expected 1 security requirement, got %d", len(builder.config.security))
		}
		
		if _, exists := builder.config.security[0]["bearerToken"]; !exists {
			t.Error("Expected Bearer token security requirement")
		}
	})

	t.Run("RequireOAuth2 is convenience for OAuth2 with scopes", func(t *testing.T) {
		builder := NewSimple().RequireOAuth2("oauth2", "read", "write")
		
		if len(builder.config.security) != 1 {
			t.Errorf("Expected 1 security requirement, got %d", len(builder.config.security))
		}
		
		if len(builder.config.security[0]["oauth2"]) != 2 {
			t.Errorf("Expected 2 OAuth2 scopes, got %d", len(builder.config.security[0]["oauth2"]))
		}
	})

	t.Run("NoAuth removes all authentication", func(t *testing.T) {
		builder := NewSimple().RequireAuth("apiKey").NoAuth()
		
		// NoAuth() should create a SecurityRequirements with an empty SecurityRequirement
		if len(builder.config.security) != 1 {
			t.Errorf("Expected 1 security requirement (empty), got %d", len(builder.config.security))
		}
		
		if len(builder.config.security[0]) != 0 {
			t.Errorf("Expected empty security requirement, got %d schemes", len(builder.config.security[0]))
		}
	})
}

// TestOperationConfigCompile tests operation compilation
func TestOperationConfigCompile(t *testing.T) {
	t.Run("Compile with basic configuration", func(t *testing.T) {
		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		}
		
		builder := NewSimple().
			GET("/test").
			Summary("Test operation").
			Description("Test description").
			Tags("test").
			SuccessCode(200)
		
		op := builder.Handler(handler)
		
		if op.Method != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", op.Method)
		}
		
		if op.Path != "/test" {
			t.Errorf("Expected path '/test', got '%s'", op.Path)
		}
		
		if op.Summary != "Test operation" {
			t.Errorf("Expected summary 'Test operation', got '%s'", op.Summary)
		}
		
		if op.Description != "Test description" {
			t.Errorf("Expected description 'Test description', got '%s'", op.Description)
		}
		
		if len(op.Tags) != 1 || op.Tags[0] != "test" {
			t.Errorf("Expected tags ['test'], got %v", op.Tags)
		}
		
		if op.SuccessCode != 200 {
			t.Errorf("Expected success code 200, got %d", op.SuccessCode)
		}
		
		if op.Handler == nil {
			t.Error("Expected handler to be set")
		}
	})

	t.Run("Compile with enhanced schemas", func(t *testing.T) {
		handler := func(c *gin.Context) {}
		
		enhancedSchema := &mockSchema{
			isEnhanced: true,
			openAPISchema: &goop.OpenAPISchema{
				Type: "object",
				Properties: map[string]*goop.OpenAPISchema{
					"id": {Type: "string"},
				},
			},
		}
		
		builder := NewSimple().
			GET("/test/{id}").
			WithParams(enhancedSchema).
			WithQuery(enhancedSchema).
			WithBody(enhancedSchema).
			WithResponse(enhancedSchema).
			WithHeaders(enhancedSchema)
		
		op := builder.Handler(handler)
		
		// Check schemas are set
		if op.ParamsSchema != enhancedSchema {
			t.Error("Expected params schema to be set")
		}
		
		if op.QuerySchema != enhancedSchema {
			t.Error("Expected query schema to be set")
		}
		
		if op.BodySchema != enhancedSchema {
			t.Error("Expected body schema to be set")
		}
		
		if op.ResponseSchema != enhancedSchema {
			t.Error("Expected response schema to be set")
		}
		
		if op.HeaderSchema != enhancedSchema {
			t.Error("Expected header schema to be set")
		}
		
		// Check OpenAPI specs are generated for enhanced schemas
		if op.ParamsSpec == nil {
			t.Error("Expected params spec to be generated")
		}
		
		if op.QuerySpec == nil {
			t.Error("Expected query spec to be generated")
		}
		
		if op.BodySpec == nil {
			t.Error("Expected body spec to be generated")
		}
		
		if op.ResponseSpec == nil {
			t.Error("Expected response spec to be generated")
		}
		
		if op.HeaderSpec == nil {
			t.Error("Expected header spec to be generated")
		}
		
		// Check spec content
		if op.ParamsSpec.Type != "object" {
			t.Errorf("Expected params spec type 'object', got '%s'", op.ParamsSpec.Type)
		}
	})

	t.Run("Compile with non-enhanced schemas", func(t *testing.T) {
		handler := func(c *gin.Context) {}
		
		basicSchema := &mockSchema{
			isEnhanced: false,
		}
		
		builder := NewSimple().
			GET("/test").
			WithParams(basicSchema).
			WithQuery(basicSchema).
			WithBody(basicSchema).
			WithResponse(basicSchema).
			WithHeaders(basicSchema)
		
		op := builder.Handler(handler)
		
		// Check schemas are set
		if op.ParamsSchema != basicSchema {
			t.Error("Expected params schema to be set")
		}
		
		// Check OpenAPI specs are NOT generated for non-enhanced schemas
		if op.ParamsSpec != nil {
			t.Error("Expected params spec to be nil for non-enhanced schema")
		}
		
		if op.QuerySpec != nil {
			t.Error("Expected query spec to be nil for non-enhanced schema")
		}
		
		if op.BodySpec != nil {
			t.Error("Expected body spec to be nil for non-enhanced schema")
		}
		
		if op.ResponseSpec != nil {
			t.Error("Expected response spec to be nil for non-enhanced schema")
		}
		
		if op.HeaderSpec != nil {
			t.Error("Expected header spec to be nil for non-enhanced schema")
		}
	})

	t.Run("Compile with security requirements", func(t *testing.T) {
		handler := func(c *gin.Context) {}
		
		builder := NewSimple().
			GET("/secure").
			RequireAuth("apiKey", "read").
			RequireOAuth2("oauth2", "write")
		
		op := builder.Handler(handler)
		
		if len(op.Security) != 2 {
			t.Errorf("Expected 2 security requirements, got %d", len(op.Security))
		}
		
		// Check first requirement (apiKey)
		if _, exists := op.Security[0]["apiKey"]; !exists {
			t.Error("Expected first security requirement to have apiKey")
		}
		
		// Check second requirement (oauth2)
		if _, exists := op.Security[1]["oauth2"]; !exists {
			t.Error("Expected second security requirement to have oauth2")
		}
	})
}

// TestSimpleBuilderChaining tests method chaining
func TestSimpleBuilderChaining(t *testing.T) {
	t.Run("All methods support fluent interface", func(t *testing.T) {
		handler := func(c *gin.Context) {}
		mockSchema := &mockSchema{shouldValidate: true}
		
		// Test that all methods can be chained together
		op := NewSimple().
			GET("/users/{id}").
			Summary("Get user").
			Description("Retrieves a user by ID").
			Tags("users", "public").
			SuccessCode(200).
			WithParams(mockSchema).
			WithQuery(mockSchema).
			WithBody(mockSchema).
			WithResponse(mockSchema).
			WithHeaders(mockSchema).
			RequireAuth("apiKey", "read").
			RequireOAuth2("oauth2", "write").
			Handler(handler)
		
		// Verify the final operation has all the expected properties
		if op.Method != "GET" {
			t.Error("Expected method to be set through chaining")
		}
		
		if op.Path != "/users/{id}" {
			t.Error("Expected path to be set through chaining")
		}
		
		if op.Summary != "Get user" {
			t.Error("Expected summary to be set through chaining")
		}
		
		if len(op.Tags) != 2 {
			t.Error("Expected tags to be set through chaining")
		}
		
		if op.ParamsSchema == nil {
			t.Error("Expected params schema to be set through chaining")
		}
		
		if len(op.Security) != 2 {
			t.Error("Expected security requirements to be set through chaining")
		}
	})
}