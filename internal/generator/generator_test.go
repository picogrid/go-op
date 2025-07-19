package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/picogrid/go-op/operations"
)

func TestNew(t *testing.T) {
	config := &Config{
		InputDir:   "/test/input",
		OutputFile: "test.yaml",
		Title:      "Test API",
		Version:    "1.0.0",
	}

	gen := New(config)

	if gen.config != config {
		t.Errorf("Expected config to be set")
	}

	if gen.fileSet == nil {
		t.Errorf("Expected fileSet to be initialized")
	}

	if len(gen.operations) != 0 {
		t.Errorf("Expected empty operations")
	}

	if len(gen.schemas) != 0 {
		t.Errorf("Expected empty schemas")
	}
}

func TestGetTitle(t *testing.T) {
	tests := []struct {
		config   *Config
		expected string
	}{
		{
			config: &Config{
				Title: "Custom API",
			},
			expected: "Custom API",
		},
		{
			config: &Config{
				InputDir: "/path/to/user-service",
			},
			expected: "User Service API",
		},
		{
			config: &Config{
				InputDir: "/path/to/order-management",
			},
			expected: "Order Management API",
		},
		{
			config: &Config{
				InputDir: ".",
			},
			expected: "Generated API",
		},
		{
			config: &Config{
				InputDir: "/",
			},
			expected: "Generated API",
		},
	}

	for _, test := range tests {
		gen := New(test.config)
		title := gen.getTitle()
		if title != test.expected {
			t.Errorf("getTitle() with InputDir=%s = %s, expected %s",
				test.config.InputDir, title, test.expected)
		}
	}
}

func TestIsPropertyRequired(t *testing.T) {
	gen := New(&Config{})

	required := []string{"name", "email", "id"}

	tests := []struct {
		propName string
		expected bool
	}{
		{"name", true},
		{"email", true},
		{"id", true},
		{"age", false},
		{"address", false},
	}

	for _, test := range tests {
		result := gen.isPropertyRequired(test.propName, required)
		if result != test.expected {
			t.Errorf("isPropertyRequired(%s) = %v, expected %v",
				test.propName, result, test.expected)
		}
	}
}

func TestConvertSchemaToOpenAPI(t *testing.T) {
	gen := New(&Config{})

	// Test simple string schema
	stringSchema := &SchemaDefinition{
		Type:        "string",
		MinLength:   intPtr(3),
		MaxLength:   intPtr(50),
		Pattern:     "^[a-zA-Z]+$",
		Description: "A string field",
	}

	openAPISchema := gen.convertSchemaToOpenAPI(stringSchema)

	if openAPISchema.Type != "string" {
		t.Errorf("Expected type 'string', got '%s'", openAPISchema.Type)
	}

	if *openAPISchema.MinLength != 3 {
		t.Errorf("Expected MinLength 3, got %d", *openAPISchema.MinLength)
	}

	if *openAPISchema.MaxLength != 50 {
		t.Errorf("Expected MaxLength 50, got %d", *openAPISchema.MaxLength)
	}

	if openAPISchema.Pattern != "^[a-zA-Z]+$" {
		t.Errorf("Expected pattern '^[a-zA-Z]+$', got '%s'", openAPISchema.Pattern)
	}

	// Test number schema
	numberSchema := &SchemaDefinition{
		Type:    "number",
		Minimum: floatPtr(0),
		Maximum: floatPtr(100),
	}

	openAPINumber := gen.convertSchemaToOpenAPI(numberSchema)

	if openAPINumber.Type != "number" {
		t.Errorf("Expected type 'number', got '%s'", openAPINumber.Type)
	}

	if *openAPINumber.Minimum != 0 {
		t.Errorf("Expected Minimum 0, got %f", *openAPINumber.Minimum)
	}

	if *openAPINumber.Maximum != 100 {
		t.Errorf("Expected Maximum 100, got %f", *openAPINumber.Maximum)
	}

	// Test object schema
	objectSchema := &SchemaDefinition{
		Type: "object",
		Properties: map[string]*SchemaDefinition{
			"name": {
				Type:        "string",
				Description: "User name",
			},
			"age": {
				Type:    "number",
				Minimum: floatPtr(0),
			},
		},
		Required: []string{"name"},
	}

	openAPIObject := gen.convertSchemaToOpenAPI(objectSchema)

	if openAPIObject.Type != "object" {
		t.Errorf("Expected type 'object', got '%s'", openAPIObject.Type)
	}

	if len(openAPIObject.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(openAPIObject.Properties))
	}

	if openAPIObject.Properties["name"].Type != "string" {
		t.Errorf("Expected name property to be string")
	}

	if len(openAPIObject.Required) != 1 || openAPIObject.Required[0] != "name" {
		t.Errorf("Expected required array to contain 'name'")
	}

	// Test array schema
	arraySchema := &SchemaDefinition{
		Type: "array",
		Items: &SchemaDefinition{
			Type: "string",
		},
	}

	openAPIArray := gen.convertSchemaToOpenAPI(arraySchema)

	if openAPIArray.Type != "array" {
		t.Errorf("Expected type 'array', got '%s'", openAPIArray.Type)
	}

	if openAPIArray.Items == nil || openAPIArray.Items.Type != "string" {
		t.Errorf("Expected array items to be string type")
	}
}

func TestConvertSchemaToRequestBody(t *testing.T) {
	gen := New(&Config{})

	schema := &SchemaDefinition{
		Type: "object",
		Properties: map[string]*SchemaDefinition{
			"name": {Type: "string"},
		},
	}

	requestBody := gen.convertSchemaToRequestBody(schema)

	if !requestBody.Required {
		t.Errorf("Expected request body to be required")
	}

	if len(requestBody.Content) != 1 {
		t.Errorf("Expected 1 content type, got %d", len(requestBody.Content))
	}

	if _, exists := requestBody.Content["application/json"]; !exists {
		t.Errorf("Expected application/json content type")
	}

	mediaType := requestBody.Content["application/json"]
	if mediaType.Schema == nil {
		t.Errorf("Expected schema to be set")
	}

	if mediaType.Schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", mediaType.Schema.Type)
	}
}

func TestAddParametersFromSchema(t *testing.T) {
	gen := New(&Config{})

	// Test path parameters - path parameters should always be required per OpenAPI spec
	pathSchema := &SchemaDefinition{
		Type: "object",
		Properties: map[string]*SchemaDefinition{
			"id": {
				Type:        "string",
				Description: "User ID",
			},
			"category": {
				Type:        "string",
				Description: "Category ID",
			},
		},
		Required: []string{"id", "category"}, // Both are required in schema
	}

	operation := operations.OpenAPIOperation{
		Parameters: []operations.OpenAPIParameter{},
	}

	// Test path parameters
	gen.addParametersFromSchema(pathSchema, "path", &operation)

	if len(operation.Parameters) != 2 {
		t.Errorf("Expected 2 path parameters, got %d", len(operation.Parameters))
	}

	// Check parameters by name (order is not guaranteed due to map iteration)
	paramsByName := make(map[string]operations.OpenAPIParameter)
	for _, param := range operation.Parameters {
		paramsByName[param.Name] = param
	}

	// Check id parameter
	idParam, hasId := paramsByName["id"]
	if !hasId {
		t.Errorf("Expected to find 'id' parameter")
	} else {
		if idParam.In != "path" {
			t.Errorf("Expected id parameter in 'path', got '%s'", idParam.In)
		}
		if !idParam.Required {
			t.Errorf("Expected id path parameter to be required")
		}
	}

	// Check category parameter
	categoryParam, hasCategory := paramsByName["category"]
	if !hasCategory {
		t.Errorf("Expected to find 'category' parameter")
	} else {
		if categoryParam.In != "path" {
			t.Errorf("Expected category parameter in 'path', got '%s'", categoryParam.In)
		}
		if !categoryParam.Required {
			t.Errorf("Expected category path parameter to be required")
		}
	}

	// Test query parameters - these can be optional
	querySchema := &SchemaDefinition{
		Type: "object",
		Properties: map[string]*SchemaDefinition{
			"search": {
				Type:        "string",
				Description: "Search term",
			},
			"limit": {
				Type:    "number",
				Minimum: floatPtr(1),
				Maximum: floatPtr(100),
			},
		},
		Required: []string{"search"}, // Only search is required
	}

	operation.Parameters = []operations.OpenAPIParameter{}
	gen.addParametersFromSchema(querySchema, "query", &operation)

	if len(operation.Parameters) != 2 {
		t.Errorf("Expected 2 query parameters, got %d", len(operation.Parameters))
	}

	// Find search parameter
	var searchParam, limitParam *operations.OpenAPIParameter
	for i := range operation.Parameters {
		if operation.Parameters[i].Name == "search" {
			searchParam = &operation.Parameters[i]
		}
		if operation.Parameters[i].Name == "limit" {
			limitParam = &operation.Parameters[i]
		}
	}

	if searchParam == nil {
		t.Errorf("Expected to find search parameter")
	} else {
		if searchParam.In != "query" {
			t.Errorf("Expected search parameter in 'query', got '%s'", searchParam.In)
		}
		if !searchParam.Required {
			t.Errorf("Expected search parameter to be required")
		}
	}

	if limitParam == nil {
		t.Errorf("Expected to find limit parameter")
	} else {
		if limitParam.In != "query" {
			t.Errorf("Expected limit parameter in 'query', got '%s'", limitParam.In)
		}
		if limitParam.Required {
			t.Errorf("Expected limit parameter to be optional")
		}
	}
}

func TestAddOperationToSpec(t *testing.T) {
	gen := New(&Config{})
	gen.spec = &operations.OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: operations.OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: make(map[string]map[string]operations.OpenAPIOperation),
	}

	// Create test operation
	op := OperationDefinition{
		Method:      "POST",
		Path:        "/users",
		Summary:     "Create user",
		Description: "Creates a new user",
		Tags:        []string{"users"},
		Params: &SchemaDefinition{
			Type: "object",
			Properties: map[string]*SchemaDefinition{
				"tenant": {Type: "string"},
			},
			Required: []string{"tenant"},
		},
		Query: &SchemaDefinition{
			Type: "object",
			Properties: map[string]*SchemaDefinition{
				"notify": {Type: "boolean"},
			},
		},
		Body: &SchemaDefinition{
			Type: "object",
			Properties: map[string]*SchemaDefinition{
				"name":  {Type: "string"},
				"email": {Type: "string", Format: "email"},
			},
			Required: []string{"name", "email"},
		},
		Response: &SchemaDefinition{
			Type: "object",
			Properties: map[string]*SchemaDefinition{
				"id":      {Type: "string"},
				"name":    {Type: "string"},
				"email":   {Type: "string"},
				"created": {Type: "string", Format: "date-time"},
			},
		},
		SourceFile: "users.go",
		LineNumber: 42,
	}

	gen.addOperationToSpec(op)

	// Verify path was created
	if _, exists := gen.spec.Paths["/users"]; !exists {
		t.Errorf("Expected path '/users' to exist")
	}

	// Verify operation was added
	if _, exists := gen.spec.Paths["/users"]["post"]; !exists {
		t.Errorf("Expected POST operation to exist")
	}

	operation := gen.spec.Paths["/users"]["post"]

	// Check operation details
	if operation.Summary != "Create user" {
		t.Errorf("Expected summary 'Create user', got '%s'", operation.Summary)
	}

	if operation.Description != "Creates a new user" {
		t.Errorf("Expected description 'Creates a new user', got '%s'", operation.Description)
	}

	if len(operation.Tags) != 1 || operation.Tags[0] != "users" {
		t.Errorf("Expected tags ['users'], got %v", operation.Tags)
	}

	// Check parameters
	pathParams := 0
	queryParams := 0
	for _, param := range operation.Parameters {
		switch param.In {
		case "path":
			pathParams++
		case "query":
			queryParams++
		}
	}

	if pathParams != 1 {
		t.Errorf("Expected 1 path parameter, got %d", pathParams)
	}

	if queryParams != 1 {
		t.Errorf("Expected 1 query parameter, got %d", queryParams)
	}

	// Check request body
	if operation.RequestBody == nil {
		t.Errorf("Expected request body to be set")
	} else if !operation.RequestBody.Required {
		t.Errorf("Expected request body to be required")
	}

	// Check response
	if _, exists := operation.Responses["200"]; !exists {
		t.Errorf("Expected 200 response to exist")
	}

	response := operation.Responses["200"]
	if response.Description != "Successful response" {
		t.Errorf("Expected response description 'Successful response', got '%s'", response.Description)
	}

	// Test adding operation without response
	opNoResponse := OperationDefinition{
		Method:  "GET",
		Path:    "/health",
		Summary: "Health check",
	}

	gen.addOperationToSpec(opNoResponse)

	if _, exists := gen.spec.Paths["/health"]["get"]; !exists {
		t.Errorf("Expected GET /health to exist")
	}

	healthOp := gen.spec.Paths["/health"]["get"]
	if _, exists := healthOp.Responses["200"]; !exists {
		t.Errorf("Expected default 200 response")
	}
}

func TestGenerateSpec(t *testing.T) {
	config := &Config{
		Title:       "Test API",
		Version:     "2.0.0",
		Description: "Test API Description",
		Servers:     []string{"https://api.example.com", "http://localhost:8080"},
	}

	gen := New(config)

	// Add some test operations
	gen.operations = []OperationDefinition{
		{
			Method:  "GET",
			Path:    "/users",
			Summary: "List users",
			Tags:    []string{"users"},
		},
		{
			Method:  "POST",
			Path:    "/users",
			Summary: "Create user",
			Tags:    []string{"users"},
		},
		{
			Method:  "GET",
			Path:    "/orders",
			Summary: "List orders",
			Tags:    []string{"orders"},
		},
	}

	err := gen.GenerateSpec()
	if err != nil {
		t.Errorf("Failed to generate spec: %v", err)
	}

	// Verify spec details
	if gen.spec.OpenAPI != "3.1.0" {
		t.Errorf("Expected OpenAPI version '3.1.0', got '%s'", gen.spec.OpenAPI)
	}

	if gen.spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", gen.spec.Info.Title)
	}

	if gen.spec.Info.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", gen.spec.Info.Version)
	}

	if gen.spec.Info.Description != "Test API Description" {
		t.Errorf("Expected description 'Test API Description', got '%s'", gen.spec.Info.Description)
	}

	// Check servers
	if len(gen.spec.Servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(gen.spec.Servers))
	}

	if gen.spec.Servers[0].URL != "https://api.example.com" {
		t.Errorf("Expected first server URL 'https://api.example.com', got '%s'", gen.spec.Servers[0].URL)
	}

	// Check paths
	if len(gen.spec.Paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(gen.spec.Paths))
	}

	// Check operations were added
	if _, exists := gen.spec.Paths["/users"]["get"]; !exists {
		t.Errorf("Expected GET /users to exist")
	}

	if _, exists := gen.spec.Paths["/users"]["post"]; !exists {
		t.Errorf("Expected POST /users to exist")
	}

	if _, exists := gen.spec.Paths["/orders"]["get"]; !exists {
		t.Errorf("Expected GET /orders to exist")
	}

	// Check stats
	if gen.stats.PathCount != 2 {
		t.Errorf("Expected PathCount 2, got %d", gen.stats.PathCount)
	}
}

func TestWriteSpec(t *testing.T) {
	tempDir := t.TempDir()

	spec := &operations.OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: operations.OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]map[string]operations.OpenAPIOperation{
			"/test": {
				"get": {
					Summary: "Test endpoint",
					Responses: map[string]operations.OpenAPIResponse{
						"200": {Description: "Success"},
					},
				},
			},
		},
	}

	// Test YAML output
	yamlFile := filepath.Join(tempDir, "output.yaml")
	gen := New(&Config{
		OutputFile: yamlFile,
		Format:     "yaml",
	})
	gen.spec = spec

	err := gen.WriteSpec()
	if err != nil {
		t.Errorf("Failed to write YAML spec: %v", err)
	}

	// Verify YAML file exists and can be parsed
	yamlData, err := os.ReadFile(yamlFile)
	if err != nil {
		t.Errorf("Failed to read YAML output: %v", err)
	}

	var yamlSpec operations.OpenAPISpec
	if err := yaml.Unmarshal(yamlData, &yamlSpec); err != nil {
		t.Errorf("Failed to parse YAML output: %v", err)
	}

	if yamlSpec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API' in YAML output")
	}

	// Test JSON output
	jsonFile := filepath.Join(tempDir, "output.json")
	gen.config.OutputFile = jsonFile
	gen.config.Format = "json"

	err = gen.WriteSpec()
	if err != nil {
		t.Errorf("Failed to write JSON spec: %v", err)
	}

	// Verify JSON file exists and can be parsed
	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Errorf("Failed to read JSON output: %v", err)
	}

	var jsonSpec operations.OpenAPISpec
	if err := json.Unmarshal(jsonData, &jsonSpec); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
	}

	if jsonSpec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API' in JSON output")
	}

	// Test unsupported format
	gen.config.Format = "xml"
	err = gen.WriteSpec()
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected unsupported format error")
	}

	// Test creating nested directories
	nestedFile := filepath.Join(tempDir, "nested", "dir", "output.yaml")
	gen.config.OutputFile = nestedFile
	gen.config.Format = "yaml"

	err = gen.WriteSpec()
	if err != nil {
		t.Errorf("Failed to write spec with nested directories: %v", err)
	}

	if _, err := os.Stat(nestedFile); os.IsNotExist(err) {
		t.Errorf("Nested output file was not created")
	}
}

func TestScanFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test Go file
	goFile := filepath.Join(tempDir, "test.go")
	goContent := `
package main

import "github.com/picogrid/go-op/operations"
import "github.com/picogrid/go-op/validators"

var getUserOperation = operations.NewSimple().
	GET("/users/{id}").
	Summary("Get user by ID").
	WithParams(validators.Object(map[string]interface{}{
		"id": validators.String().Required(),
	})).
	WithResponse(validators.Object(map[string]interface{}{
		"id": validators.String(),
		"name": validators.String(),
	}))
`
	if err := os.WriteFile(goFile, []byte(goContent), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	gen := New(&Config{
		Verbose: true,
	})

	err := gen.scanFile(goFile)
	if err != nil {
		t.Errorf("Failed to scan file: %v", err)
	}

	if gen.stats.FileCount != 1 {
		t.Errorf("Expected FileCount 1, got %d", gen.stats.FileCount)
	}
}

func TestScanOperations(t *testing.T) {
	tempDir := t.TempDir()

	// Create test directory structure
	mainFile := filepath.Join(tempDir, "main.go")
	mainContent := `
package main

var op1 = operations.NewSimple().GET("/test")
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0o644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Create test file (should be skipped)
	testFile := filepath.Join(tempDir, "main_test.go")
	testContent := `
package main

var testOp = operations.NewSimple().GET("/test-endpoint")
`
	if err := os.WriteFile(testFile, []byte(testContent), 0o644); err != nil {
		t.Fatalf("Failed to create main_test.go: %v", err)
	}

	// Create vendor file (should be skipped)
	vendorDir := filepath.Join(tempDir, "vendor", "github.com", "test")
	os.MkdirAll(vendorDir, 0o755)
	vendorFile := filepath.Join(vendorDir, "vendor.go")
	vendorContent := `
package test

var vendorOp = operations.NewSimple().GET("/vendor")
`
	if err := os.WriteFile(vendorFile, []byte(vendorContent), 0o644); err != nil {
		t.Fatalf("Failed to create vendor file: %v", err)
	}

	// Create non-Go file (should be skipped)
	txtFile := filepath.Join(tempDir, "readme.txt")
	if err := os.WriteFile(txtFile, []byte("readme"), 0o644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}

	gen := New(&Config{
		InputDir: tempDir,
		Verbose:  true,
	})

	err := gen.ScanOperations()
	if err != nil {
		t.Errorf("Failed to scan operations: %v", err)
	}

	// Should only scan main.go
	if gen.stats.FileCount != 1 {
		t.Errorf("Expected to scan 1 file, scanned %d", gen.stats.FileCount)
	}
}

func TestGetStats(t *testing.T) {
	gen := New(&Config{})

	// Set some stats
	gen.stats = GenerationStats{
		OperationCount: 10,
		SchemaCount:    5,
		PathCount:      7,
		FileCount:      3,
	}

	stats := gen.GetStats()

	if stats.OperationCount != 10 {
		t.Errorf("Expected OperationCount 10, got %d", stats.OperationCount)
	}

	if stats.SchemaCount != 5 {
		t.Errorf("Expected SchemaCount 5, got %d", stats.SchemaCount)
	}

	if stats.PathCount != 7 {
		t.Errorf("Expected PathCount 7, got %d", stats.PathCount)
	}

	if stats.FileCount != 3 {
		t.Errorf("Expected FileCount 3, got %d", stats.FileCount)
	}
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
