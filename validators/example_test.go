package validators

import (
	"testing"

	goop "github.com/picogrid/go-op"
)

func TestStringExampleFunctionality(t *testing.T) {
	// Test basic example functionality
	schema := String().Min(3).Example("test-example").Required()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := schema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("Schema does not implement EnhancedSchema interface")
	}

	// Generate OpenAPI schema
	openAPISchema := enhancedSchema.ToOpenAPISchema()

	// Verify the example is present
	if openAPISchema.Example == nil {
		t.Error("Expected example to be present in OpenAPI schema")
	}

	if openAPISchema.Example != "test-example" {
		t.Errorf("Expected example to be 'test-example', got %v", openAPISchema.Example)
	}

	// Verify other properties are still present
	if openAPISchema.Type != "string" {
		t.Errorf("Expected type to be 'string', got %s", openAPISchema.Type)
	}

	if openAPISchema.MinLength == nil || *openAPISchema.MinLength != 3 {
		t.Error("Expected minLength to be 3")
	}
}

func TestNumberExampleFunctionality(t *testing.T) {
	// Test number example functionality
	schema := Number().Min(1).Max(100).Example(42.5).Required()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := schema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("Schema does not implement EnhancedSchema interface")
	}

	// Generate OpenAPI schema
	openAPISchema := enhancedSchema.ToOpenAPISchema()

	// Verify the example is present
	if openAPISchema.Example == nil {
		t.Error("Expected example to be present in OpenAPI schema")
	}

	if openAPISchema.Example != 42.5 {
		t.Errorf("Expected example to be 42.5, got %v", openAPISchema.Example)
	}

	// Verify other properties are still present
	if openAPISchema.Type != "number" {
		t.Errorf("Expected type to be 'number', got %s", openAPISchema.Type)
	}
}

func TestObjectExampleFunctionality(t *testing.T) {
	// Test object example functionality
	exampleObj := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}

	schema := Object(map[string]interface{}{
		"name": String().Required(),
		"age":  Number().Required(),
	}).Example(exampleObj).Required()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := schema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("Schema does not implement EnhancedSchema interface")
	}

	// Generate OpenAPI schema
	openAPISchema := enhancedSchema.ToOpenAPISchema()

	// Verify the example is present
	if openAPISchema.Example == nil {
		t.Error("Expected example to be present in OpenAPI schema")
	}

	// Verify other properties are still present
	if openAPISchema.Type != "object" {
		t.Errorf("Expected type to be 'object', got %s", openAPISchema.Type)
	}

	if len(openAPISchema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(openAPISchema.Properties))
	}
}

func TestArrayExampleFunctionality(t *testing.T) {
	// Test array example functionality
	exampleArray := []interface{}{"item1", "item2", "item3"}

	schema := Array(String()).MinItems(1).Example(exampleArray).Required()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := schema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("Schema does not implement EnhancedSchema interface")
	}

	// Generate OpenAPI schema
	openAPISchema := enhancedSchema.ToOpenAPISchema()

	// Verify the example is present
	if openAPISchema.Example == nil {
		t.Error("Expected example to be present in OpenAPI schema")
	}

	// Verify other properties are still present
	if openAPISchema.Type != "array" {
		t.Errorf("Expected type to be 'array', got %s", openAPISchema.Type)
	}
}

func TestBoolExampleFunctionality(t *testing.T) {
	// Test bool example functionality
	schema := Bool().Example(true).Required()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := schema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("Schema does not implement EnhancedSchema interface")
	}

	// Generate OpenAPI schema
	openAPISchema := enhancedSchema.ToOpenAPISchema()

	// Verify the example is present
	if openAPISchema.Example == nil {
		t.Error("Expected example to be present in OpenAPI schema")
	}

	if openAPISchema.Example != true {
		t.Errorf("Expected example to be true, got %v", openAPISchema.Example)
	}

	// Verify other properties are still present
	if openAPISchema.Type != "boolean" {
		t.Errorf("Expected type to be 'boolean', got %s", openAPISchema.Type)
	}
}

func TestExamplesMapFunctionality(t *testing.T) {
	// Test multiple examples functionality
	examples := map[string]ExampleObject{
		"valid_email": {
			Summary:     "Valid email example",
			Description: "An example of a valid email address",
			Value:       "user@example.com",
		},
		"corporate_email": {
			Summary:     "Corporate email example",
			Description: "An example of a corporate email",
			Value:       "jane.doe@company.com",
		},
	}

	schema := String().Email().Examples(examples).Required()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := schema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("Schema does not implement EnhancedSchema interface")
	}

	// This test just verifies that the Examples method works without errors
	// The OpenAPI schema doesn't currently expose the examples map, but the
	// method should work for future AST extraction

	// Verify the schema still generates correctly
	openAPISchema := enhancedSchema.ToOpenAPISchema()
	if openAPISchema.Type != "string" {
		t.Errorf("Expected type to be 'string', got %s", openAPISchema.Type)
	}

	if openAPISchema.Format != "email" {
		t.Errorf("Expected format to be 'email', got %s", openAPISchema.Format)
	}
}

func TestExampleFromFileFunctionality(t *testing.T) {
	// Test external example functionality
	schema := String().ExampleFromFile("./examples/user.json").Required()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := schema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("Schema does not implement EnhancedSchema interface")
	}

	// This test just verifies that the ExampleFromFile method works without errors
	// The external value should be stored for AST extraction

	// Verify the schema still generates correctly
	openAPISchema := enhancedSchema.ToOpenAPISchema()
	if openAPISchema.Type != "string" {
		t.Errorf("Expected type to be 'string', got %s", openAPISchema.Type)
	}
}

func TestExampleWithOptionalSchema(t *testing.T) {
	// Test example functionality with optional schema
	schema := String().Min(1).Example("optional-example").Optional()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := schema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("Schema does not implement EnhancedSchema interface")
	}

	// Generate OpenAPI schema
	openAPISchema := enhancedSchema.ToOpenAPISchema()

	// Verify the example is present
	if openAPISchema.Example == nil {
		t.Error("Expected example to be present in OpenAPI schema")
	}

	if openAPISchema.Example != "optional-example" {
		t.Errorf("Expected example to be 'optional-example', got %v", openAPISchema.Example)
	}
}
