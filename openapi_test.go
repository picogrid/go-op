package goop

import (
	"encoding/json"
	"testing"
)

// TestOpenAPISchema tests the OpenAPISchema struct
func TestOpenAPISchema(t *testing.T) {
	t.Run("OpenAPISchema creation and JSON serialization", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type:        "object",
			Description: "User schema",
			Properties: map[string]*OpenAPISchema{
				"name": {
					Type:      "string",
					MinLength: intPtr(2),
					MaxLength: intPtr(50),
				},
				"age": {
					Type:    "integer",
					Minimum: floatPtr(0),
					Maximum: floatPtr(120),
				},
			},
			Required: []string{"name", "age"},
		}

		// Test JSON marshaling
		jsonData, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal OpenAPISchema to JSON: %v", err)
		}

		// Test JSON unmarshaling
		var unmarshaled OpenAPISchema
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal OpenAPISchema from JSON: %v", err)
		}

		// Verify fields
		if unmarshaled.Type != "object" {
			t.Errorf("Expected Type 'object', got '%s'", unmarshaled.Type)
		}
		if unmarshaled.Description != "User schema" {
			t.Errorf("Expected Description 'User schema', got '%s'", unmarshaled.Description)
		}
		if len(unmarshaled.Required) != 2 {
			t.Errorf("Expected 2 required fields, got %d", len(unmarshaled.Required))
		}
		if len(unmarshaled.Properties) != 2 {
			t.Errorf("Expected 2 properties, got %d", len(unmarshaled.Properties))
		}
	})

	t.Run("OpenAPISchema with all fields", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type:        "string",
			Format:      "email",
			MinLength:   intPtr(5),
			MaxLength:   intPtr(100),
			Pattern:     "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			Enum:        []interface{}{"admin@example.com", "user@example.com"},
			Default:     "user@example.com",
			Description: "User email address",
			Example:     "john.doe@example.com",
		}

		// Verify all fields are set
		if schema.Type != "string" {
			t.Errorf("Expected Type 'string', got '%s'", schema.Type)
		}
		if schema.Format != "email" {
			t.Errorf("Expected Format 'email', got '%s'", schema.Format)
		}
		if *schema.MinLength != 5 {
			t.Errorf("Expected MinLength 5, got %d", *schema.MinLength)
		}
		if *schema.MaxLength != 100 {
			t.Errorf("Expected MaxLength 100, got %d", *schema.MaxLength)
		}
		if len(schema.Enum) != 2 {
			t.Errorf("Expected 2 enum values, got %d", len(schema.Enum))
		}
		if schema.Default != "user@example.com" {
			t.Errorf("Expected Default 'user@example.com', got '%v'", schema.Default)
		}
	})

	t.Run("OpenAPISchema with nested objects", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type: "object",
			Properties: map[string]*OpenAPISchema{
				"user": {
					Type: "object",
					Properties: map[string]*OpenAPISchema{
						"profile": {
							Type: "object",
							Properties: map[string]*OpenAPISchema{
								"name": {Type: "string"},
								"age":  {Type: "integer"},
							},
						},
					},
				},
			},
		}

		// Test deep nesting access
		userProp := schema.Properties["user"]
		if userProp == nil {
			t.Fatal("Expected user property to exist")
		}

		profileProp := userProp.Properties["profile"]
		if profileProp == nil {
			t.Fatal("Expected profile property to exist")
		}

		nameProp := profileProp.Properties["name"]
		if nameProp == nil || nameProp.Type != "string" {
			t.Error("Expected name property to be string type")
		}
	})

	t.Run("OpenAPISchema with arrays", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type: "array",
			Items: &OpenAPISchema{
				Type: "object",
				Properties: map[string]*OpenAPISchema{
					"id":   {Type: "integer"},
					"name": {Type: "string"},
				},
				Required: []string{"id"},
			},
		}

		if schema.Type != "array" {
			t.Errorf("Expected Type 'array', got '%s'", schema.Type)
		}
		if schema.Items == nil {
			t.Fatal("Expected Items to be set")
		}
		if schema.Items.Type != "object" {
			t.Errorf("Expected Items Type 'object', got '%s'", schema.Items.Type)
		}
		if len(schema.Items.Required) != 1 {
			t.Errorf("Expected 1 required field in items, got %d", len(schema.Items.Required))
		}
	})

	t.Run("OpenAPISchema with number constraints", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type:    "number",
			Minimum: floatPtr(-100.5),
			Maximum: floatPtr(100.5),
		}

		if *schema.Minimum != -100.5 {
			t.Errorf("Expected Minimum -100.5, got %f", *schema.Minimum)
		}
		if *schema.Maximum != 100.5 {
			t.Errorf("Expected Maximum 100.5, got %f", *schema.Maximum)
		}
	})
}

// TestValidationInfo tests the ValidationInfo struct
func TestValidationInfo(t *testing.T) {
	t.Run("ValidationInfo creation and field access", func(t *testing.T) {
		info := &ValidationInfo{
			Required:     true,
			Optional:     false,
			HasDefault:   true,
			DefaultValue: "default_value",
			Constraints: map[string]interface{}{
				"minLength": 5,
				"maxLength": 50,
				"pattern":   "^[a-zA-Z]+$",
			},
		}

		if !info.Required {
			t.Error("Expected Required to be true")
		}
		if info.Optional {
			t.Error("Expected Optional to be false")
		}
		if !info.HasDefault {
			t.Error("Expected HasDefault to be true")
		}
		if info.DefaultValue != "default_value" {
			t.Errorf("Expected DefaultValue 'default_value', got '%v'", info.DefaultValue)
		}
		if len(info.Constraints) != 3 {
			t.Errorf("Expected 3 constraints, got %d", len(info.Constraints))
		}
	})

	t.Run("ValidationInfo with various constraint types", func(t *testing.T) {
		info := &ValidationInfo{
			Constraints: map[string]interface{}{
				"minLength":   10,
				"maxLength":   100,
				"minimum":     0.0,
				"maximum":     1000.5,
				"pattern":     "^test",
				"enum":        []string{"a", "b", "c"},
				"required":    true,
				"multipleOf":  2.5,
				"uniqueItems": true,
				"minItems":    1,
				"maxItems":    10,
				"format":      "email",
			},
		}

		// Test integer constraints
		if minLength, ok := info.Constraints["minLength"]; !ok || minLength != 10 {
			t.Errorf("Expected minLength 10, got %v", minLength)
		}

		// Test float constraints
		if maximum, ok := info.Constraints["maximum"]; !ok || maximum != 1000.5 {
			t.Errorf("Expected maximum 1000.5, got %v", maximum)
		}

		// Test string constraints
		if pattern, ok := info.Constraints["pattern"]; !ok || pattern != "^test" {
			t.Errorf("Expected pattern '^test', got %v", pattern)
		}

		// Test boolean constraints
		if uniqueItems, ok := info.Constraints["uniqueItems"]; !ok || uniqueItems != true {
			t.Errorf("Expected uniqueItems true, got %v", uniqueItems)
		}

		// Test array constraints
		if enum, ok := info.Constraints["enum"]; !ok {
			t.Error("Expected enum constraint to exist")
		} else if enumSlice, ok := enum.([]string); !ok || len(enumSlice) != 3 {
			t.Errorf("Expected enum to be []string with 3 items, got %v", enum)
		}
	})

	t.Run("ValidationInfo with nil and empty values", func(t *testing.T) {
		info := &ValidationInfo{
			Required:     false,
			Optional:     true,
			HasDefault:   false,
			DefaultValue: nil,
			Constraints:  nil,
		}

		if info.Required {
			t.Error("Expected Required to be false")
		}
		if !info.Optional {
			t.Error("Expected Optional to be true")
		}
		if info.HasDefault {
			t.Error("Expected HasDefault to be false")
		}
		if info.DefaultValue != nil {
			t.Errorf("Expected DefaultValue to be nil, got %v", info.DefaultValue)
		}
		if info.Constraints != nil {
			t.Errorf("Expected Constraints to be nil, got %v", info.Constraints)
		}
	})

	t.Run("ValidationInfo edge cases", func(t *testing.T) {
		// Test with empty constraints map
		info := &ValidationInfo{
			Constraints: map[string]interface{}{},
		}
		if len(info.Constraints) != 0 {
			t.Errorf("Expected empty constraints map, got %d items", len(info.Constraints))
		}

		// Test with complex default value
		complexDefault := map[string]interface{}{
			"nested": "value",
			"array":  []int{1, 2, 3},
		}
		info.DefaultValue = complexDefault
		// For maps, we need to check content rather than direct comparison
		if defaultMap, ok := info.DefaultValue.(map[string]interface{}); !ok {
			t.Errorf("Expected DefaultValue to be map[string]interface{}, got %T", info.DefaultValue)
		} else {
			if defaultMap["nested"] != "value" {
				t.Error("Expected complex default value nested field to be preserved")
			}
			if arr, ok := defaultMap["array"].([]int); !ok || len(arr) != 3 || arr[0] != 1 {
				t.Error("Expected complex default value array field to be preserved")
			}
		}
	})
}

// TestOpenAPISchemaJSONEdgeCases tests JSON serialization edge cases
func TestOpenAPISchemaJSONEdgeCases(t *testing.T) {
	t.Run("OpenAPISchema with nil pointers", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type:      "string",
			MinLength: nil,
			MaxLength: nil,
			Minimum:   nil,
			Maximum:   nil,
		}

		jsonData, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal schema with nil pointers: %v", err)
		}

		var unmarshaled OpenAPISchema
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal schema with nil pointers: %v", err)
		}

		// Nil pointers should remain nil after round-trip
		if unmarshaled.MinLength != nil {
			t.Error("Expected MinLength to remain nil")
		}
		if unmarshaled.MaxLength != nil {
			t.Error("Expected MaxLength to remain nil")
		}
	})

	t.Run("OpenAPISchema with empty collections", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type:       "object",
			Properties: map[string]*OpenAPISchema{},
			Required:   []string{},
			Enum:       []interface{}{},
		}

		jsonData, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal schema with empty collections: %v", err)
		}

		// Check that empty collections are handled properly in JSON
		var jsonMap map[string]interface{}
		if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
			t.Fatalf("Failed to unmarshal to map: %v", err)
		}

		// Empty slices and maps might be omitted or included depending on omitempty
		t.Logf("JSON representation: %s", string(jsonData))
	})

	t.Run("OpenAPISchema with special characters", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type:        "string",
			Description: "Description with \"quotes\" and \nnewlines\ttabs",
			Pattern:     "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			Example:     "test@example.com\nwith\nnewlines",
		}

		jsonData, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal schema with special characters: %v", err)
		}

		var unmarshaled OpenAPISchema
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal schema with special characters: %v", err)
		}

		if unmarshaled.Description != schema.Description {
			t.Error("Description with special characters not preserved")
		}
		if unmarshaled.Pattern != schema.Pattern {
			t.Error("Pattern with special characters not preserved")
		}
	})
}

// TestOpenAPIIntegration tests integration scenarios
func TestOpenAPIIntegration(t *testing.T) {
	t.Run("Complex schema generation scenario", func(t *testing.T) {
		// Simulate a complex API schema
		userSchema := &OpenAPISchema{
			Type:        "object",
			Description: "User information",
			Properties: map[string]*OpenAPISchema{
				"id": {
					Type:        "integer",
					Description: "Unique user identifier",
					Minimum:     floatPtr(1),
				},
				"username": {
					Type:        "string",
					Description: "User's unique username",
					MinLength:   intPtr(3),
					MaxLength:   intPtr(20),
					Pattern:     "^[a-zA-Z0-9_]+$",
				},
				"email": {
					Type:        "string",
					Format:      "email",
					Description: "User's email address",
				},
				"profile": {
					Type:        "object",
					Description: "User profile information",
					Properties: map[string]*OpenAPISchema{
						"firstName": {Type: "string", MinLength: intPtr(1)},
						"lastName":  {Type: "string", MinLength: intPtr(1)},
						"age": {
							Type:    "integer",
							Minimum: floatPtr(0),
							Maximum: floatPtr(120),
						},
						"tags": {
							Type: "array",
							Items: &OpenAPISchema{
								Type:      "string",
								MinLength: intPtr(1),
							},
						},
					},
					Required: []string{"firstName", "lastName"},
				},
			},
			Required: []string{"id", "username", "email"},
		}

		// Test JSON serialization of complex schema
		jsonData, err := json.Marshal(userSchema)
		if err != nil {
			t.Fatalf("Failed to marshal complex schema: %v", err)
		}

		// Test JSON deserialization
		var unmarshaled OpenAPISchema
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal complex schema: %v", err)
		}

		// Verify complex structure integrity
		if len(unmarshaled.Required) != 3 {
			t.Errorf("Expected 3 required fields, got %d", len(unmarshaled.Required))
		}

		profileProp := unmarshaled.Properties["profile"]
		if profileProp == nil {
			t.Fatal("Expected profile property to exist")
		}

		if len(profileProp.Required) != 2 {
			t.Errorf("Expected 2 required fields in profile, got %d", len(profileProp.Required))
		}

		tagsProp := profileProp.Properties["tags"]
		if tagsProp == nil || tagsProp.Type != "array" {
			t.Error("Expected tags to be array type")
		}

		if tagsProp.Items == nil || tagsProp.Items.Type != "string" {
			t.Error("Expected tags items to be string type")
		}
	})

	t.Run("Validation info integration", func(t *testing.T) {
		// Test how ValidationInfo might be used with OpenAPISchema
		validationInfo := &ValidationInfo{
			Required:     true,
			Optional:     false,
			HasDefault:   true,
			DefaultValue: "default@example.com",
			Constraints: map[string]interface{}{
				"format":    "email",
				"minLength": 5,
				"maxLength": 100,
			},
		}

		// Convert validation info to OpenAPI schema
		schema := &OpenAPISchema{
			Type:   "string",
			Format: validationInfo.Constraints["format"].(string),
		}

		if minLen, ok := validationInfo.Constraints["minLength"].(int); ok {
			schema.MinLength = &minLen
		}
		if maxLen, ok := validationInfo.Constraints["maxLength"].(int); ok {
			schema.MaxLength = &maxLen
		}
		if validationInfo.HasDefault {
			schema.Default = validationInfo.DefaultValue
		}

		// Verify conversion
		if schema.Format != "email" {
			t.Errorf("Expected format 'email', got '%s'", schema.Format)
		}
		if *schema.MinLength != 5 {
			t.Errorf("Expected MinLength 5, got %d", *schema.MinLength)
		}
		if schema.Default != "default@example.com" {
			t.Errorf("Expected default 'default@example.com', got %v", schema.Default)
		}
	})

	t.Run("Schema composition and reusability", func(t *testing.T) {
		// Define reusable schemas
		stringSchema := &OpenAPISchema{
			Type:      "string",
			MinLength: intPtr(1),
		}

		idSchema := &OpenAPISchema{
			Type:    "integer",
			Minimum: floatPtr(1),
		}

		// Compose schemas
		userSchema := &OpenAPISchema{
			Type: "object",
			Properties: map[string]*OpenAPISchema{
				"id":   idSchema,
				"name": stringSchema,
				"email": {
					Type:   "string",
					Format: "email",
				},
			},
			Required: []string{"id", "name"},
		}

		// Test that shared schemas maintain integrity
		if userSchema.Properties["id"] != idSchema {
			t.Error("ID schema reference not maintained")
		}
		if userSchema.Properties["name"] != stringSchema {
			t.Error("String schema reference not maintained")
		}

		// Modify shared schema and verify it affects composed schema
		stringSchema.MaxLength = intPtr(50)
		if userSchema.Properties["name"].MaxLength == nil || *userSchema.Properties["name"].MaxLength != 50 {
			t.Error("Changes to shared schema not reflected in composed schema")
		}
	})
}

// Helper functions for creating pointers
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

// TestOpenAPIHelperFunctions tests helper functions
func TestOpenAPIHelperFunctions(t *testing.T) {
	t.Run("Helper functions for pointer creation", func(t *testing.T) {
		intVal := intPtr(42)
		if *intVal != 42 {
			t.Errorf("Expected intPtr(42) to create pointer to 42, got %d", *intVal)
		}

		floatVal := floatPtr(3.14)
		if *floatVal != 3.14 {
			t.Errorf("Expected floatPtr(3.14) to create pointer to 3.14, got %f", *floatVal)
		}
	})

	t.Run("Pointer comparison in schemas", func(t *testing.T) {
		schema1 := &OpenAPISchema{
			MinLength: intPtr(5),
			Minimum:   floatPtr(0.0),
		}

		schema2 := &OpenAPISchema{
			MinLength: intPtr(5),
			Minimum:   floatPtr(0.0),
		}

		// Values should be equal even though pointers are different
		if *schema1.MinLength != *schema2.MinLength {
			t.Error("MinLength values should be equal")
		}
		if *schema1.Minimum != *schema2.Minimum {
			t.Error("Minimum values should be equal")
		}

		// Pointers should be different
		if schema1.MinLength == schema2.MinLength {
			t.Error("MinLength pointers should be different")
		}
		if schema1.Minimum == schema2.Minimum {
			t.Error("Minimum pointers should be different")
		}
	})
}

// TestOpenAPIFixedFields tests the new OpenAPI 3.1 fixed fields
func TestOpenAPIFixedFields(t *testing.T) {
	t.Run("Numeric validation fields", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type:             "number",
			MultipleOf:       floatPtr(2.5),
			ExclusiveMinimum: floatPtr(0.0),
			ExclusiveMaximum: floatPtr(100.0),
		}

		if *schema.MultipleOf != 2.5 {
			t.Errorf("Expected MultipleOf 2.5, got %f", *schema.MultipleOf)
		}
		if *schema.ExclusiveMinimum != 0.0 {
			t.Errorf("Expected ExclusiveMinimum 0.0, got %f", *schema.ExclusiveMinimum)
		}
		if *schema.ExclusiveMaximum != 100.0 {
			t.Errorf("Expected ExclusiveMaximum 100.0, got %f", *schema.ExclusiveMaximum)
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal schema with numeric fields: %v", err)
		}

		var unmarshaled OpenAPISchema
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal schema with numeric fields: %v", err)
		}

		if *unmarshaled.MultipleOf != 2.5 {
			t.Errorf("Expected unmarshaled MultipleOf 2.5, got %f", *unmarshaled.MultipleOf)
		}
	})

	t.Run("Array validation fields", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type:        "array",
			MaxItems:    intPtr(10),
			MinItems:    intPtr(1),
			UniqueItems: boolPtr(true),
			Items: &OpenAPISchema{
				Type: "string",
			},
		}

		if *schema.MaxItems != 10 {
			t.Errorf("Expected MaxItems 10, got %d", *schema.MaxItems)
		}
		if *schema.MinItems != 1 {
			t.Errorf("Expected MinItems 1, got %d", *schema.MinItems)
		}
		if !*schema.UniqueItems {
			t.Error("Expected UniqueItems to be true")
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal schema with array fields: %v", err)
		}

		var unmarshaled OpenAPISchema
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal schema with array fields: %v", err)
		}

		if *unmarshaled.UniqueItems != true {
			t.Error("Expected unmarshaled UniqueItems to be true")
		}
	})

	t.Run("Object validation fields", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type:          "object",
			MaxProperties: intPtr(5),
			MinProperties: intPtr(1),
			AdditionalProperties: &OpenAPISchemaOrBool{
				Bool: boolPtr(false),
			},
		}

		if *schema.MaxProperties != 5 {
			t.Errorf("Expected MaxProperties 5, got %d", *schema.MaxProperties)
		}
		if *schema.MinProperties != 1 {
			t.Errorf("Expected MinProperties 1, got %d", *schema.MinProperties)
		}
		if *schema.AdditionalProperties.Bool != false {
			t.Error("Expected AdditionalProperties to be false")
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal schema with object fields: %v", err)
		}

		var unmarshaled OpenAPISchema
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal schema with object fields: %v", err)
		}

		if *unmarshaled.AdditionalProperties.Bool != false {
			t.Error("Expected unmarshaled AdditionalProperties to be false")
		}
	})

	t.Run("Schema composition fields", func(t *testing.T) {
		stringSchema := &OpenAPISchema{Type: "string"}
		numberSchema := &OpenAPISchema{Type: "number"}

		schema := &OpenAPISchema{
			AllOf: []*OpenAPISchema{stringSchema, numberSchema},
			OneOf: []*OpenAPISchema{stringSchema, numberSchema},
			AnyOf: []*OpenAPISchema{stringSchema, numberSchema},
			Not:   stringSchema,
		}

		if len(schema.AllOf) != 2 {
			t.Errorf("Expected AllOf length 2, got %d", len(schema.AllOf))
		}
		if len(schema.OneOf) != 2 {
			t.Errorf("Expected OneOf length 2, got %d", len(schema.OneOf))
		}
		if len(schema.AnyOf) != 2 {
			t.Errorf("Expected AnyOf length 2, got %d", len(schema.AnyOf))
		}
		if schema.Not != stringSchema {
			t.Error("Expected Not to reference stringSchema")
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal schema with composition fields: %v", err)
		}

		var unmarshaled OpenAPISchema
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal schema with composition fields: %v", err)
		}

		if len(unmarshaled.AllOf) != 2 {
			t.Errorf("Expected unmarshaled AllOf length 2, got %d", len(unmarshaled.AllOf))
		}
	})

	t.Run("Metadata fields", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type:       "string",
			Title:      "User Email",
			Const:      "admin@example.com",
			ReadOnly:   boolPtr(true),
			WriteOnly:  boolPtr(false),
			Deprecated: boolPtr(true),
		}

		if schema.Title != "User Email" {
			t.Errorf("Expected Title 'User Email', got '%s'", schema.Title)
		}
		if schema.Const != "admin@example.com" {
			t.Errorf("Expected Const 'admin@example.com', got '%v'", schema.Const)
		}
		if !*schema.ReadOnly {
			t.Error("Expected ReadOnly to be true")
		}
		if *schema.WriteOnly {
			t.Error("Expected WriteOnly to be false")
		}
		if !*schema.Deprecated {
			t.Error("Expected Deprecated to be true")
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal schema with metadata fields: %v", err)
		}

		var unmarshaled OpenAPISchema
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal schema with metadata fields: %v", err)
		}

		if unmarshaled.Title != "User Email" {
			t.Errorf("Expected unmarshaled Title 'User Email', got '%s'", unmarshaled.Title)
		}
		if !*unmarshaled.ReadOnly {
			t.Error("Expected unmarshaled ReadOnly to be true")
		}
	})

	t.Run("AdditionalProperties with schema", func(t *testing.T) {
		additionalSchema := &OpenAPISchema{
			Type:    "string",
			Pattern: "^[a-z]+$",
		}

		schema := &OpenAPISchema{
			Type: "object",
			AdditionalProperties: &OpenAPISchemaOrBool{
				Schema: additionalSchema,
			},
		}

		if schema.AdditionalProperties.Schema != additionalSchema {
			t.Error("Expected AdditionalProperties to reference schema")
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal schema with additionalProperties schema: %v", err)
		}

		var unmarshaled OpenAPISchema
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal schema with additionalProperties schema: %v", err)
		}

		if unmarshaled.AdditionalProperties.Schema == nil {
			t.Error("Expected unmarshaled AdditionalProperties to have schema")
		}
		if unmarshaled.AdditionalProperties.Schema.Type != "string" {
			t.Error("Expected unmarshaled AdditionalProperties schema type to be string")
		}
	})
}

// TestOpenAPISchemaOrBool tests the OpenAPISchemaOrBool type
func TestOpenAPISchemaOrBool(t *testing.T) {
	t.Run("OpenAPISchemaOrBool with boolean", func(t *testing.T) {
		schemaOrBool := &OpenAPISchemaOrBool{
			Bool: boolPtr(true),
		}

		// Test JSON marshaling
		jsonData, err := json.Marshal(schemaOrBool)
		if err != nil {
			t.Fatalf("Failed to marshal OpenAPISchemaOrBool with bool: %v", err)
		}

		expected := "true"
		if string(jsonData) != expected {
			t.Errorf("Expected JSON '%s', got '%s'", expected, string(jsonData))
		}

		// Test JSON unmarshaling
		var unmarshaled OpenAPISchemaOrBool
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal OpenAPISchemaOrBool with bool: %v", err)
		}

		if unmarshaled.Bool == nil || !*unmarshaled.Bool {
			t.Error("Expected unmarshaled Bool to be true")
		}
		if unmarshaled.Schema != nil {
			t.Error("Expected unmarshaled Schema to be nil")
		}
	})

	t.Run("OpenAPISchemaOrBool with schema", func(t *testing.T) {
		schema := &OpenAPISchema{
			Type:    "string",
			Pattern: "^[a-z]+$",
		}

		schemaOrBool := &OpenAPISchemaOrBool{
			Schema: schema,
		}

		// Test JSON marshaling
		jsonData, err := json.Marshal(schemaOrBool)
		if err != nil {
			t.Fatalf("Failed to marshal OpenAPISchemaOrBool with schema: %v", err)
		}

		// Test JSON unmarshaling
		var unmarshaled OpenAPISchemaOrBool
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal OpenAPISchemaOrBool with schema: %v", err)
		}

		if unmarshaled.Schema == nil {
			t.Error("Expected unmarshaled Schema to be set")
		}
		if unmarshaled.Schema.Type != "string" {
			t.Error("Expected unmarshaled Schema type to be string")
		}
		if unmarshaled.Bool != nil {
			t.Error("Expected unmarshaled Bool to be nil")
		}
	})

	t.Run("OpenAPISchemaOrBool with nil values", func(t *testing.T) {
		schemaOrBool := &OpenAPISchemaOrBool{}

		// Test JSON marshaling
		jsonData, err := json.Marshal(schemaOrBool)
		if err != nil {
			t.Fatalf("Failed to marshal OpenAPISchemaOrBool with nil values: %v", err)
		}

		expected := "null"
		if string(jsonData) != expected {
			t.Errorf("Expected JSON '%s', got '%s'", expected, string(jsonData))
		}
	})
}

// Helper function for creating bool pointers
func boolPtr(b bool) *bool {
	return &b
}
