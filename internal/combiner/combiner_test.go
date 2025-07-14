package combiner

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
		Title:   "Test API",
		Version: "1.0.0",
	}

	combiner := New(config)

	if combiner.config != config {
		t.Errorf("Expected config to be set")
	}

	if len(combiner.inputFiles) != 0 {
		t.Errorf("Expected empty input files")
	}

	if len(combiner.specs) != 0 {
		t.Errorf("Expected empty specs")
	}
}

func TestAddInputFile(t *testing.T) {
	combiner := New(&Config{})

	combiner.AddInputFile("test1.yaml")
	combiner.AddInputFile("test2.yaml")

	if len(combiner.inputFiles) != 2 {
		t.Errorf("Expected 2 input files, got %d", len(combiner.inputFiles))
	}

	if combiner.inputFiles[0] != "test1.yaml" {
		t.Errorf("Expected first file to be test1.yaml")
	}

	if combiner.inputFiles[1] != "test2.yaml" {
		t.Errorf("Expected second file to be test2.yaml")
	}
}

func TestExtractServiceName(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"user-service.yaml", "user"},
		{"order-service.json", "order"},
		{"user-api.yaml", "user"},
		{"order.api.json", "order"},
		{"simple.yaml", "simple"},
		{"/path/to/user-service.yaml", "user"},
	}

	combiner := New(&Config{})

	for _, test := range tests {
		result := combiner.extractServiceName(test.filename)
		if result != test.expected {
			t.Errorf("extractServiceName(%s) = %s, expected %s", test.filename, result, test.expected)
		}
	}
}

func TestTransformPath(t *testing.T) {
	tests := []struct {
		originalPath string
		prefix       string
		baseURL      string
		expected     string
	}{
		{"/users", "", "", "/users"},
		{"/users", "/api/v1", "", "/api/v1/users"},
		{"/users", "", "/api", "/api/users"},
		{"/users", "/v1", "/api", "/api/v1/users"},
		{"//users", "", "", "/users"},
		{"/users", "//v1", "", "/v1/users"},
	}

	for _, test := range tests {
		combiner := New(&Config{BaseURL: test.baseURL})
		specMeta := &SpecWithMetadata{
			Prefix: test.prefix,
		}

		result := combiner.transformPath(test.originalPath, specMeta)
		if result != test.expected {
			t.Errorf("transformPath(%s, prefix=%s, baseURL=%s) = %s, expected %s",
				test.originalPath, test.prefix, test.baseURL, result, test.expected)
		}
	}
}

func TestAddServiceTag(t *testing.T) {
	combiner := New(&Config{})

	tests := []struct {
		tags        []string
		serviceName string
		expected    []string
	}{
		{[]string{}, "user", []string{"service:user"}},
		{[]string{"auth"}, "user", []string{"service:user", "auth"}},
		{[]string{"service:user"}, "user", []string{"service:user"}},
		{[]string{"auth", "public"}, "order", []string{"service:order", "auth", "public"}},
	}

	for _, test := range tests {
		result := combiner.addServiceTag(test.tags, test.serviceName)
		if len(result) != len(test.expected) {
			t.Errorf("addServiceTag(%v, %s) length = %d, expected %d",
				test.tags, test.serviceName, len(result), len(test.expected))
			continue
		}

		for i, tag := range result {
			if tag != test.expected[i] {
				t.Errorf("addServiceTag result[%d] = %s, expected %s", i, tag, test.expected[i])
			}
		}
	}
}

func TestFindOperationSource(t *testing.T) {
	combiner := New(&Config{})

	tests := []struct {
		operation operations.OpenAPIOperation
		expected  string
	}{
		{operations.OpenAPIOperation{Tags: []string{"service:user", "auth"}}, "user"},
		{operations.OpenAPIOperation{Tags: []string{"auth", "service:order"}}, "order"},
		{operations.OpenAPIOperation{Tags: []string{"auth"}}, "unknown"},
		{operations.OpenAPIOperation{}, "unknown"},
	}

	for _, test := range tests {
		result := combiner.findOperationSource(test.operation)
		if result != test.expected {
			t.Errorf("findOperationSource() = %s, expected %s", result, test.expected)
		}
	}
}

func TestFilterMethodsByTags(t *testing.T) {
	combiner := New(&Config{})

	methods := map[string]operations.OpenAPIOperation{
		"get":    {Tags: []string{"public", "auth"}},
		"post":   {Tags: []string{"admin"}},
		"put":    {Tags: []string{"public"}},
		"delete": {Tags: []string{"admin", "dangerous"}},
	}

	// Test include tags
	combiner.config.IncludeTags = []string{"public"}
	filtered := combiner.filterMethodsByTags(methods)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 methods with 'public' tag, got %d", len(filtered))
	}

	// Test exclude tags
	combiner.config.IncludeTags = []string{}
	combiner.config.ExcludeTags = []string{"admin"}
	filtered = combiner.filterMethodsByTags(methods)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 methods without 'admin' tag, got %d", len(filtered))
	}

	// Test both include and exclude
	combiner.config.IncludeTags = []string{"public"}
	combiner.config.ExcludeTags = []string{"auth"}
	filtered = combiner.filterMethodsByTags(methods)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 method with 'public' but not 'auth', got %d", len(filtered))
	}
}

func TestLoadSingleSpec(t *testing.T) {
	// Create temporary test files
	tempDir := t.TempDir()

	// Test YAML file
	yamlFile := filepath.Join(tempDir, "test.yaml")
	yamlContent := `
openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      summary: Test endpoint
      responses:
        '200':
          description: Success
`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}

	// Test JSON file
	jsonFile := filepath.Join(tempDir, "test.json")
	jsonContent := `{
  "openapi": "3.1.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "paths": {
    "/test": {
      "get": {
        "summary": "Test endpoint",
        "responses": {
          "200": {
            "description": "Success"
          }
        }
      }
    }
  }
}`
	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to create test JSON file: %v", err)
	}

	combiner := New(&Config{})

	// Test loading YAML
	spec, err := combiner.loadSingleSpec(yamlFile)
	if err != nil {
		t.Errorf("Failed to load YAML spec: %v", err)
	}
	if spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", spec.Info.Title)
	}

	// Test loading JSON
	spec, err = combiner.loadSingleSpec(jsonFile)
	if err != nil {
		t.Errorf("Failed to load JSON spec: %v", err)
	}
	if spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", spec.Info.Title)
	}

	// Test loading non-existent file
	_, err = combiner.loadSingleSpec("non-existent.yaml")
	if err == nil {
		t.Errorf("Expected error loading non-existent file")
	}
}

func TestCombineSpecs(t *testing.T) {
	// Create test specs
	spec1 := &operations.OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: operations.OpenAPIInfo{
			Title:   "Service 1",
			Version: "1.0.0",
		},
		Paths: map[string]map[string]operations.OpenAPIOperation{
			"/users": {
				"get": {
					Summary: "Get users",
					Tags:    []string{"users"},
					Responses: map[string]operations.OpenAPIResponse{
						"200": {Description: "Success"},
					},
				},
			},
		},
	}

	spec2 := &operations.OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: operations.OpenAPIInfo{
			Title:   "Service 2",
			Version: "1.0.0",
		},
		Paths: map[string]map[string]operations.OpenAPIOperation{
			"/orders": {
				"post": {
					Summary: "Create order",
					Tags:    []string{"orders"},
					Responses: map[string]operations.OpenAPIResponse{
						"201": {Description: "Created"},
					},
				},
			},
		},
	}

	combiner := New(&Config{
		Title:   "Combined API",
		Version: "2.0.0",
		ServicePrefix: map[string]string{
			"service1": "/v1",
			"service2": "/v2",
		},
	})

	combiner.specs = []*SpecWithMetadata{
		{
			Spec:        spec1,
			ServiceName: "service1",
			Prefix:      "/v1",
			SourceFile:  "service1.yaml",
		},
		{
			Spec:        spec2,
			ServiceName: "service2",
			Prefix:      "/v2",
			SourceFile:  "service2.yaml",
		},
	}

	err := combiner.CombineSpecs()
	if err != nil {
		t.Errorf("Failed to combine specs: %v", err)
	}

	// Check combined result
	if combiner.combined.Info.Title != "Combined API" {
		t.Errorf("Expected title 'Combined API', got '%s'", combiner.combined.Info.Title)
	}

	if combiner.combined.Info.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", combiner.combined.Info.Version)
	}

	// Check paths were combined with prefixes
	if _, exists := combiner.combined.Paths["/v1/users"]; !exists {
		t.Errorf("Expected path '/v1/users' to exist")
	}

	if _, exists := combiner.combined.Paths["/v2/orders"]; !exists {
		t.Errorf("Expected path '/v2/orders' to exist")
	}

	// Check service tags were added
	if op, exists := combiner.combined.Paths["/v1/users"]["get"]; exists {
		found := false
		for _, tag := range op.Tags {
			if tag == "service:service1" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected service tag 'service:service1' to be added")
		}
	}
}

func TestValidateOutput(t *testing.T) {
	tests := []struct {
		name      string
		spec      *operations.OpenAPISpec
		expectErr bool
	}{
		{
			name:      "nil spec",
			spec:      nil,
			expectErr: true,
		},
		{
			name: "missing OpenAPI version",
			spec: &operations.OpenAPISpec{
				Info: operations.OpenAPIInfo{Title: "Test", Version: "1.0.0"},
			},
			expectErr: true,
		},
		{
			name: "missing title",
			spec: &operations.OpenAPISpec{
				OpenAPI: "3.1.0",
				Info:    operations.OpenAPIInfo{Version: "1.0.0"},
			},
			expectErr: true,
		},
		{
			name: "missing version",
			spec: &operations.OpenAPISpec{
				OpenAPI: "3.1.0",
				Info:    operations.OpenAPIInfo{Title: "Test"},
			},
			expectErr: true,
		},
		{
			name: "no paths",
			spec: &operations.OpenAPISpec{
				OpenAPI: "3.1.0",
				Info:    operations.OpenAPIInfo{Title: "Test", Version: "1.0.0"},
				Paths:   map[string]map[string]operations.OpenAPIOperation{},
			},
			expectErr: true,
		},
		{
			name: "path with no operations",
			spec: &operations.OpenAPISpec{
				OpenAPI: "3.1.0",
				Info:    operations.OpenAPIInfo{Title: "Test", Version: "1.0.0"},
				Paths: map[string]map[string]operations.OpenAPIOperation{
					"/test": {},
				},
			},
			expectErr: true,
		},
		{
			name: "operation with no responses",
			spec: &operations.OpenAPISpec{
				OpenAPI: "3.1.0",
				Info:    operations.OpenAPIInfo{Title: "Test", Version: "1.0.0"},
				Paths: map[string]map[string]operations.OpenAPIOperation{
					"/test": {
						"get": {
							Summary:   "Test",
							Responses: map[string]operations.OpenAPIResponse{},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "valid spec",
			spec: &operations.OpenAPISpec{
				OpenAPI: "3.1.0",
				Info:    operations.OpenAPIInfo{Title: "Test", Version: "1.0.0"},
				Paths: map[string]map[string]operations.OpenAPIOperation{
					"/test": {
						"get": {
							Summary: "Test",
							Responses: map[string]operations.OpenAPIResponse{
								"200": {Description: "Success"},
							},
						},
					},
				},
			},
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			combiner := New(&Config{})
			combiner.combined = test.spec

			err := combiner.ValidateOutput()
			if test.expectErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !test.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestWriteOutput(t *testing.T) {
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
	combiner := New(&Config{
		OutputFile: yamlFile,
		Format:     "yaml",
	})
	combiner.combined = spec

	err := combiner.WriteOutput()
	if err != nil {
		t.Errorf("Failed to write YAML output: %v", err)
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

	// Test JSON output
	jsonFile := filepath.Join(tempDir, "output.json")
	combiner.config.OutputFile = jsonFile
	combiner.config.Format = "json"

	err = combiner.WriteOutput()
	if err != nil {
		t.Errorf("Failed to write JSON output: %v", err)
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

	// Test unsupported format
	combiner.config.Format = "xml"
	err = combiner.WriteOutput()
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected unsupported format error")
	}
}

func TestLoadFromConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create test spec files
	spec1File := filepath.Join(tempDir, "user-service.yaml")
	spec1Content := `
openapi: 3.1.0
info:
  title: User Service
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
`
	if err := os.WriteFile(spec1File, []byte(spec1Content), 0644); err != nil {
		t.Fatalf("Failed to create spec file: %v", err)
	}

	spec2File := filepath.Join(tempDir, "order-service.yaml")
	spec2Content := `
openapi: 3.1.0
info:
  title: Order Service
  version: 1.0.0
paths:
  /orders:
    get:
      summary: Get orders
`
	if err := os.WriteFile(spec2File, []byte(spec2Content), 0644); err != nil {
		t.Fatalf("Failed to create spec file: %v", err)
	}

	// Create services config file
	configFile := filepath.Join(tempDir, "services.yaml")
	configContent := `
title: Platform API
version: 3.0.0
base_url: /api

services:
  - name: user
    spec_file: ` + spec1File + `
    path_prefix: /v1/users
  - name: order
    spec_file: ` + spec2File + `
    path_prefix: /v1/orders

settings:
  merge_schemas: true
  validate_output: true
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test loading from config
	combiner := New(&Config{
		ConfigFile: configFile,
		Verbose:    true,
	})

	err := combiner.LoadFromConfig()
	if err != nil {
		t.Errorf("Failed to load from config: %v", err)
	}

	// Verify files were added
	if len(combiner.inputFiles) != 2 {
		t.Errorf("Expected 2 input files, got %d", len(combiner.inputFiles))
	}

	// Verify service prefixes were set
	if combiner.config.ServicePrefix["user"] != "/v1/users" {
		t.Errorf("Expected user service prefix '/v1/users', got '%s'", combiner.config.ServicePrefix["user"])
	}

	if combiner.config.ServicePrefix["order"] != "/v1/orders" {
		t.Errorf("Expected order service prefix '/v1/orders', got '%s'", combiner.config.ServicePrefix["order"])
	}
}

func TestGetStats(t *testing.T) {
	combiner := New(&Config{})

	// Set some stats
	combiner.stats = CombinationStats{
		InputFiles:       3,
		ServicesCombined: 2,
		TotalOperations:  10,
		TotalPaths:       5,
		MergedSchemas:    7,
		Conflicts:        1,
	}

	stats := combiner.GetStats()

	if stats.InputFiles != 3 {
		t.Errorf("Expected InputFiles = 3, got %d", stats.InputFiles)
	}

	if stats.ServicesCombined != 2 {
		t.Errorf("Expected ServicesCombined = 2, got %d", stats.ServicesCombined)
	}

	if stats.TotalOperations != 10 {
		t.Errorf("Expected TotalOperations = 10, got %d", stats.TotalOperations)
	}
}

func TestCombineSpecsPaths(t *testing.T) {
	// Test path conflict handling
	spec1 := &operations.OpenAPISpec{
		OpenAPI: "3.1.0",
		Info:    operations.OpenAPIInfo{Title: "Service 1", Version: "1.0.0"},
		Paths: map[string]map[string]operations.OpenAPIOperation{
			"/users": {
				"get": {
					Summary: "Get users from service 1",
					Tags:    []string{"users"},
					Responses: map[string]operations.OpenAPIResponse{
						"200": {Description: "Success"},
					},
				},
			},
		},
	}

	spec2 := &operations.OpenAPISpec{
		OpenAPI: "3.1.0",
		Info:    operations.OpenAPIInfo{Title: "Service 2", Version: "1.0.0"},
		Paths: map[string]map[string]operations.OpenAPIOperation{
			"/users": {
				"get": {
					Summary: "Get users from service 2",
					Tags:    []string{"users-v2"},
					Responses: map[string]operations.OpenAPIResponse{
						"200": {Description: "Success"},
					},
				},
			},
		},
	}

	combiner := New(&Config{
		Title:   "Combined API",
		Version: "1.0.0",
		Verbose: true,
	})

	combiner.specs = []*SpecWithMetadata{
		{
			Spec:        spec1,
			ServiceName: "service1",
			SourceFile:  "service1.yaml",
		},
		{
			Spec:        spec2,
			ServiceName: "service2",
			SourceFile:  "service2.yaml",
		},
	}

	err := combiner.CombineSpecs()
	if err != nil {
		t.Errorf("Failed to combine specs: %v", err)
	}

	// The second service should override the first
	if op, exists := combiner.combined.Paths["/users"]["get"]; exists {
		if op.Summary != "Get users from service 2" {
			t.Errorf("Expected service 2 to override, got summary: %s", op.Summary)
		}

		// Should have service:service2 tag
		found := false
		for _, tag := range op.Tags {
			if tag == "service:service2" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected service:service2 tag")
		}
	} else {
		t.Errorf("Expected /users GET operation to exist")
	}
}

func TestWriteOutputWithDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Test creating nested directories
	outputFile := filepath.Join(tempDir, "nested", "dir", "output.yaml")

	combiner := New(&Config{
		OutputFile: outputFile,
		Format:     "yaml",
	})

	combiner.combined = &operations.OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: operations.OpenAPIInfo{
			Title:   "Test",
			Version: "1.0.0",
		},
		Paths: map[string]map[string]operations.OpenAPIOperation{},
	}

	err := combiner.WriteOutput()
	if err != nil {
		t.Errorf("Failed to write output with nested directories: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created")
	}
}
