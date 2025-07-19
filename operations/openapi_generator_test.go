package operations

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	goop "github.com/picogrid/go-op"
)

// TestNewOpenAPIGenerator tests OpenAPI generator creation
func TestNewOpenAPIGenerator(t *testing.T) {
	t.Run("Create new OpenAPI generator", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		if generator == nil {
			t.Error("Expected generator to be created")
			return
		}

		if generator.Title != "Test API" {
			t.Errorf("Expected title 'Test API', got '%s'", generator.Title)
		}

		if generator.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", generator.Version)
		}

		if generator.SecuritySchemes == nil {
			t.Error("Expected security schemes map to be initialized")
		}

		if generator.Spec == nil {
			t.Error("Expected spec to be initialized")
		}

		// Check spec initialization
		if generator.Spec.OpenAPI != "3.1.0" {
			t.Errorf("Expected OpenAPI version '3.1.0', got '%s'", generator.Spec.OpenAPI)
		}

		if generator.Spec.Info.Title != "Test API" {
			t.Errorf("Expected info title 'Test API', got '%s'", generator.Spec.Info.Title)
		}

		if generator.Spec.Info.Version != "1.0.0" {
			t.Errorf("Expected info version '1.0.0', got '%s'", generator.Spec.Info.Version)
		}

		if generator.Spec.Paths == nil {
			t.Error("Expected paths map to be initialized")
		}

		if generator.Spec.Components == nil {
			t.Error("Expected components to be initialized")
		}

		if generator.Spec.Components.Schemas == nil {
			t.Error("Expected schemas map to be initialized")
		}

		if generator.Spec.Components.SecuritySchemes == nil {
			t.Error("Expected security schemes map to be initialized")
		}
	})
}

// TestAddSecurityScheme tests security scheme management
func TestAddSecurityScheme(t *testing.T) {
	t.Run("Add valid API key security scheme", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		scheme := goop.NewAPIKeyHeader("X-API-Key", "API key authentication")
		err := generator.AddSecurityScheme("apiKey", scheme)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check scheme was added to generator
		if _, exists := generator.SecuritySchemes["apiKey"]; !exists {
			t.Error("Expected security scheme to be added to generator")
		}

		// Check scheme was added to OpenAPI spec
		if _, exists := generator.Spec.Components.SecuritySchemes["apiKey"]; !exists {
			t.Error("Expected security scheme to be added to OpenAPI spec")
		}

		// Check OpenAPI conversion
		openAPIScheme := generator.Spec.Components.SecuritySchemes["apiKey"]
		if openAPIScheme.Type != "apiKey" {
			t.Errorf("Expected type 'apiKey', got '%s'", openAPIScheme.Type)
		}

		if openAPIScheme.Name != "X-API-Key" {
			t.Errorf("Expected name 'X-API-Key', got '%s'", openAPIScheme.Name)
		}

		if openAPIScheme.In != "header" {
			t.Errorf("Expected in 'header', got '%s'", openAPIScheme.In)
		}
	})

	t.Run("Add valid OAuth2 security scheme", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		scopes := map[string]string{
			"read":  "Read access",
			"write": "Write access",
		}

		scheme := goop.NewOAuth2AuthorizationCode(
			"https://auth.example.com/oauth/authorize",
			"https://auth.example.com/oauth/token",
			"https://auth.example.com/oauth/refresh",
			scopes,
			"OAuth2 authentication",
		)

		err := generator.AddSecurityScheme("oauth2", scheme)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check OpenAPI conversion
		openAPIScheme := generator.Spec.Components.SecuritySchemes["oauth2"]
		if openAPIScheme.Type != "oauth2" {
			t.Errorf("Expected type 'oauth2', got '%s'", openAPIScheme.Type)
		}

		if openAPIScheme.Flows == nil {
			t.Error("Expected flows to be set")
		}

		if openAPIScheme.Flows.AuthorizationCode == nil {
			t.Error("Expected authorization code flow to be set")
		}

		if openAPIScheme.Flows.AuthorizationCode.AuthorizationURL != "https://auth.example.com/oauth/authorize" {
			t.Error("Expected authorization URL to be set correctly")
		}
	})

	t.Run("Add security scheme with invalid name", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		scheme := goop.NewAPIKeyHeader("X-API-Key", "API key authentication")
		err := generator.AddSecurityScheme("invalid name!", scheme)

		if err == nil {
			t.Error("Expected error for invalid security scheme name")
		}

		if !strings.Contains(err.Error(), "invalid security scheme name") {
			t.Errorf("Expected invalid name error, got: %v", err)
		}
	})

	t.Run("Add invalid security scheme", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		// Create invalid scheme (empty name)
		scheme := &goop.APIKeySecurityScheme{
			Name: "", // Invalid: empty name
			In:   goop.HeaderLocation,
		}

		err := generator.AddSecurityScheme("apiKey", scheme)

		if err == nil {
			t.Error("Expected error for invalid security scheme")
		}

		if !strings.Contains(err.Error(), "invalid security scheme") {
			t.Errorf("Expected invalid scheme error, got: %v", err)
		}
	})
}

// TestSecuritySchemeManagement tests security scheme retrieval and listing
func TestSecuritySchemeManagement(t *testing.T) {
	t.Run("Get existing security scheme", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		scheme := goop.NewAPIKeyHeader("X-API-Key", "API key authentication")
		generator.AddSecurityScheme("apiKey", scheme)

		retrievedScheme, exists := generator.GetSecurityScheme("apiKey")

		if !exists {
			t.Error("Expected security scheme to exist")
		}

		if retrievedScheme != scheme {
			t.Error("Expected to retrieve the same scheme instance")
		}
	})

	t.Run("Get non-existent security scheme", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		_, exists := generator.GetSecurityScheme("nonexistent")

		if exists {
			t.Error("Expected security scheme to not exist")
		}
	})

	t.Run("List security schemes", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		// Initially empty
		schemes := generator.ListSecuritySchemes()
		if len(schemes) != 0 {
			t.Errorf("Expected 0 schemes, got %d", len(schemes))
		}

		// Add some schemes
		generator.AddSecurityScheme("apiKey", goop.NewAPIKeyHeader("X-API-Key", ""))
		generator.AddSecurityScheme("bearer", goop.NewBearerAuth("JWT", ""))

		schemes = generator.ListSecuritySchemes()
		if len(schemes) != 2 {
			t.Errorf("Expected 2 schemes, got %d", len(schemes))
		}

		// Check schemes are in the list
		schemeSet := make(map[string]bool)
		for _, name := range schemes {
			schemeSet[name] = true
		}

		if !schemeSet["apiKey"] {
			t.Error("Expected 'apiKey' to be in schemes list")
		}

		if !schemeSet["bearer"] {
			t.Error("Expected 'bearer' to be in schemes list")
		}
	})
}

// TestSetGlobalSecurity tests global security configuration
func TestSetGlobalSecurity(t *testing.T) {
	t.Run("Set global security requirements", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		security := goop.SecurityRequirements{}.RequireScheme("apiKey", "read")
		generator.SetGlobalSecurity(security)

		if len(generator.GlobalSecurity) != 1 {
			t.Errorf("Expected 1 global security requirement, got %d", len(generator.GlobalSecurity))
		}

		if len(generator.Spec.Security) != 1 {
			t.Errorf("Expected 1 spec security requirement, got %d", len(generator.Spec.Security))
		}

		if generator.Spec.Security[0]["apiKey"][0] != "read" {
			t.Errorf("Expected security scope 'read', got %v", generator.Spec.Security[0]["apiKey"])
		}
	})
}

// TestProcessOperation tests operation processing
func TestProcessOperation(t *testing.T) {
	t.Run("Process basic operation", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		handler := func(c *gin.Context) {}
		op := CompiledOperation{
			Method:      "GET",
			Path:        "/users",
			Summary:     "List users",
			Description: "Get all users",
			Tags:        []string{"users"},
			Handler:     handler,
			SuccessCode: 200,
		}

		info := OperationInfo{
			Method:      op.Method,
			Path:        op.Path,
			Summary:     op.Summary,
			Description: op.Description,
			Tags:        op.Tags,
			Operation:   &op,
		}

		err := generator.Process(info)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check operation was added to spec
		if generator.Spec.Paths["/users"] == nil {
			t.Error("Expected path '/users' to be created")
		}

		operation := generator.Spec.Paths["/users"]["get"]
		if operation.Summary != "List users" {
			t.Errorf("Expected summary 'List users', got '%s'", operation.Summary)
		}

		if operation.Description != "Get all users" {
			t.Errorf("Expected description 'Get all users', got '%s'", operation.Description)
		}

		if len(operation.Tags) != 1 || operation.Tags[0] != "users" {
			t.Errorf("Expected tags ['users'], got %v", operation.Tags)
		}

		// Check default responses
		if operation.Responses["200"].Description != "Successful response" {
			t.Error("Expected 200 response to be added")
		}

		if operation.Responses["400"].Description != "Bad Request" {
			t.Error("Expected 400 response to be added")
		}

		if operation.Responses["500"].Description != "Internal Server Error" {
			t.Error("Expected 500 response to be added")
		}
	})

	t.Run("Process operation with parameters", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		handler := func(c *gin.Context) {}

		paramsSpec := &goop.OpenAPISchema{
			Type: "object",
			Properties: map[string]*goop.OpenAPISchema{
				"id": {Type: "string"},
			},
		}

		querySpec := &goop.OpenAPISchema{
			Type: "object",
			Properties: map[string]*goop.OpenAPISchema{
				"filter": {Type: "string"},
				"sort":   {Type: "string"},
			},
			Required: []string{"filter"},
		}

		op := CompiledOperation{
			Method:      "GET",
			Path:        "/users/{id}",
			Summary:     "Get user",
			Handler:     handler,
			SuccessCode: 200,
			ParamsSpec:  paramsSpec,
			QuerySpec:   querySpec,
		}

		info := OperationInfo{
			Method:    op.Method,
			Path:      op.Path,
			Summary:   op.Summary,
			Operation: &op,
		}

		err := generator.Process(info)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		operation := generator.Spec.Paths["/users/{id}"]["get"]

		// Check parameters were extracted
		if len(operation.Parameters) == 0 {
			t.Error("Expected parameters to be extracted")
		}

		// Find path parameter
		var pathParam *OpenAPIParameter
		var queryParamFilter *OpenAPIParameter
		var queryParamSort *OpenAPIParameter

		for i := range operation.Parameters {
			param := &operation.Parameters[i]
			switch {
			case param.Name == "id" && param.In == "path":
				pathParam = param
			case param.Name == "filter" && param.In == "query":
				queryParamFilter = param
			case param.Name == "sort" && param.In == "query":
				queryParamSort = param
			}
		}

		if pathParam == nil {
			t.Error("Expected path parameter 'id' to be extracted")
		} else {
			if !pathParam.Required {
				t.Error("Expected path parameter to be required")
			}
			if pathParam.Schema.Type != "string" {
				t.Errorf("Expected path parameter type 'string', got '%s'", pathParam.Schema.Type)
			}
		}

		if queryParamFilter == nil {
			t.Error("Expected query parameter 'filter' to be extracted")
		} else if !queryParamFilter.Required {
			t.Error("Expected 'filter' query parameter to be required")
		}

		if queryParamSort == nil {
			t.Error("Expected query parameter 'sort' to be extracted")
		} else if queryParamSort.Required {
			t.Error("Expected 'sort' query parameter to be optional")
		}
	})

	t.Run("Process operation with request body", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		handler := func(c *gin.Context) {}

		bodySpec := &goop.OpenAPISchema{
			Type: "object",
			Properties: map[string]*goop.OpenAPISchema{
				"name":  {Type: "string"},
				"email": {Type: "string"},
			},
			Required: []string{"name", "email"},
		}

		op := CompiledOperation{
			Method:      "POST",
			Path:        "/users",
			Summary:     "Create user",
			Handler:     handler,
			SuccessCode: 201,
			BodySpec:    bodySpec,
		}

		info := OperationInfo{
			Method:    op.Method,
			Path:      op.Path,
			Summary:   op.Summary,
			Operation: &op,
			BodyInfo:  &goop.ValidationInfo{Required: true},
		}

		err := generator.Process(info)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		operation := generator.Spec.Paths["/users"]["post"]

		if operation.RequestBody == nil {
			t.Error("Expected request body to be set")
		}

		if !operation.RequestBody.Required {
			t.Error("Expected request body to be required")
		}

		jsonContent := operation.RequestBody.Content["application/json"]
		if jsonContent.Schema != bodySpec {
			t.Error("Expected request body schema to match")
		}
	})

	t.Run("Process operation with response schema", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		handler := func(c *gin.Context) {}

		responseSpec := &goop.OpenAPISchema{
			Type: "object",
			Properties: map[string]*goop.OpenAPISchema{
				"id":   {Type: "string"},
				"name": {Type: "string"},
			},
		}

		op := CompiledOperation{
			Method:       "GET",
			Path:         "/users/1",
			Summary:      "Get user",
			Handler:      handler,
			SuccessCode:  200,
			ResponseSpec: responseSpec,
		}

		info := OperationInfo{
			Method:    op.Method,
			Path:      op.Path,
			Summary:   op.Summary,
			Operation: &op,
		}

		err := generator.Process(info)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		operation := generator.Spec.Paths["/users/1"]["get"]

		successResponse := operation.Responses["200"]
		if successResponse.Description != "Successful response" {
			t.Error("Expected success response description")
		}

		jsonContent := successResponse.Content["application/json"]
		if jsonContent.Schema != responseSpec {
			t.Error("Expected response schema to match")
		}
	})

	t.Run("Process operation with security", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		handler := func(c *gin.Context) {}

		security := goop.SecurityRequirements{}.RequireScheme("apiKey", "read")

		op := CompiledOperation{
			Method:      "GET",
			Path:        "/secure",
			Summary:     "Secure endpoint",
			Handler:     handler,
			SuccessCode: 200,
			Security:    security,
		}

		info := OperationInfo{
			Method:    op.Method,
			Path:      op.Path,
			Summary:   op.Summary,
			Operation: &op,
		}

		err := generator.Process(info)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		operation := generator.Spec.Paths["/secure"]["get"]

		if len(operation.Security) != 1 {
			t.Errorf("Expected 1 security requirement, got %d", len(operation.Security))
		}

		if operation.Security[0]["apiKey"][0] != "read" {
			t.Errorf("Expected security scope 'read', got %v", operation.Security[0]["apiKey"])
		}
	})
}

// TestOpenAPISpecWriting tests spec output functionality
func TestOpenAPISpecWriting(t *testing.T) {
	t.Run("Write to writer", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		var buf bytes.Buffer
		err := generator.WriteToWriter(&buf)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Parse the JSON to verify it's valid
		var spec map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &spec)
		if err != nil {
			t.Errorf("Expected valid JSON, got error: %v", err)
		}

		if spec["openapi"] != "3.1.0" {
			t.Errorf("Expected OpenAPI version '3.1.0', got %v", spec["openapi"])
		}

		info := spec["info"].(map[string]interface{})
		if info["title"] != "Test API" {
			t.Errorf("Expected title 'Test API', got %v", info["title"])
		}
	})

	t.Run("Write to file", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		// Create temporary file
		tmpFile := "/tmp/test_openapi.json"
		defer os.Remove(tmpFile)

		err := generator.WriteToFile(tmpFile)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Read back and verify
		data, err := os.ReadFile(tmpFile)
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}

		var spec map[string]interface{}
		err = json.Unmarshal(data, &spec)
		if err != nil {
			t.Errorf("Expected valid JSON, got error: %v", err)
		}

		if spec["openapi"] != "3.1.0" {
			t.Errorf("Expected OpenAPI version '3.1.0', got %v", spec["openapi"])
		}
	})

	t.Run("Write to invalid path", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		err := generator.WriteToFile("/invalid/path/file.json")

		if err == nil {
			t.Error("Expected error for invalid file path")
		}
	})
}

// TestGetSpec tests spec retrieval
func TestGetSpec(t *testing.T) {
	t.Run("Get spec returns the internal spec", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		spec := generator.GetSpec()

		if spec != generator.Spec {
			t.Error("Expected GetSpec to return the internal spec")
		}

		if spec.OpenAPI != "3.1.0" {
			t.Errorf("Expected OpenAPI version '3.1.0', got '%s'", spec.OpenAPI)
		}

		if spec.Info.Title != "Test API" {
			t.Errorf("Expected title 'Test API', got '%s'", spec.Info.Title)
		}
	})
}

// TestParameterExtraction tests parameter extraction methods
func TestParameterExtraction(t *testing.T) {
	t.Run("Extract path parameters", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		schema := &goop.OpenAPISchema{
			Type: "object",
			Properties: map[string]*goop.OpenAPISchema{
				"id":       {Type: "string"},
				"category": {Type: "string"},
				"other":    {Type: "string"}, // Not in path
			},
		}

		params := generator.extractPathParameters("/users/{id}/categories/{category}", schema)

		if len(params) != 2 {
			t.Errorf("Expected 2 path parameters, got %d", len(params))
		}

		// Check parameters
		paramNames := make(map[string]bool)
		for _, param := range params {
			paramNames[param.Name] = true

			if param.In != "path" {
				t.Errorf("Expected parameter in 'path', got '%s'", param.In)
			}

			if !param.Required {
				t.Errorf("Expected path parameter '%s' to be required", param.Name)
			}
		}

		if !paramNames["id"] {
			t.Error("Expected 'id' path parameter")
		}

		if !paramNames["category"] {
			t.Error("Expected 'category' path parameter")
		}

		if paramNames["other"] {
			t.Error("Did not expect 'other' parameter (not in path)")
		}
	})

	t.Run("Extract query parameters", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		schema := &goop.OpenAPISchema{
			Type: "object",
			Properties: map[string]*goop.OpenAPISchema{
				"filter": {Type: "string"},
				"sort":   {Type: "string"},
				"limit":  {Type: "integer"},
			},
			Required: []string{"filter"},
		}

		params := generator.extractQueryParameters(schema)

		if len(params) != 3 {
			t.Errorf("Expected 3 query parameters, got %d", len(params))
		}

		// Check parameters
		paramMap := make(map[string]OpenAPIParameter)
		for _, param := range params {
			paramMap[param.Name] = param

			if param.In != "query" {
				t.Errorf("Expected parameter in 'query', got '%s'", param.In)
			}
		}

		// Check required parameter
		if filterParam, exists := paramMap["filter"]; exists {
			if !filterParam.Required {
				t.Error("Expected 'filter' parameter to be required")
			}
		} else {
			t.Error("Expected 'filter' parameter")
		}

		// Check optional parameter
		if sortParam, exists := paramMap["sort"]; exists {
			if sortParam.Required {
				t.Errorf("Expected 'sort' parameter to be optional, but required=%v", sortParam.Required)
			}
		} else {
			t.Error("Expected 'sort' parameter")
		}
	})

	t.Run("Extract header parameters", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		schema := &goop.OpenAPISchema{
			Type: "object",
			Properties: map[string]*goop.OpenAPISchema{
				"X-Client-ID":      {Type: "string"},
				"X-Client-Version": {Type: "string"},
			},
			Required: []string{"X-Client-ID"},
		}

		params := generator.extractHeaderParameters(schema)

		if len(params) != 2 {
			t.Errorf("Expected 2 header parameters, got %d", len(params))
		}

		// Check parameters
		paramMap := make(map[string]OpenAPIParameter)
		for _, param := range params {
			paramMap[param.Name] = param

			if param.In != "header" {
				t.Errorf("Expected parameter in 'header', got '%s'", param.In)
			}
		}

		// Check required header
		if clientIDParam, exists := paramMap["X-Client-ID"]; exists {
			if !clientIDParam.Required {
				t.Error("Expected 'X-Client-ID' header to be required")
			}
		} else {
			t.Error("Expected 'X-Client-ID' header parameter")
		}

		// Check optional header
		if versionParam, exists := paramMap["X-Client-Version"]; exists {
			if versionParam.Required {
				t.Error("Expected 'X-Client-Version' header to be optional")
			}
		} else {
			t.Error("Expected 'X-Client-Version' header parameter")
		}
	})
}
