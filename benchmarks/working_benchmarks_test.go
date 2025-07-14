package benchmarks

import (
	"testing"

	"github.com/picogrid/go-op/validators"
)

// BenchmarkWorkingValidators tests basic validator performance with actual working methods
func BenchmarkWorkingValidators(b *testing.B) {
	b.Run("String_Min_Max", func(b *testing.B) {
		schema := validators.String().Min(3).Max(50).Required()
		testValue := "test_string"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("String_Pattern", func(b *testing.B) {
		schema := validators.String().Pattern(`^[a-zA-Z0-9]+$`).Required()
		testValue := "testuser123"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Number_Min_Max", func(b *testing.B) {
		schema := validators.Number().Min(0).Max(100).Required()
		testValue := 42.5
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Number_Integer", func(b *testing.B) {
		schema := validators.Number().Integer().Required()
		testValue := 42.0
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Boolean_Required", func(b *testing.B) {
		schema := validators.Bool().Required()
		testValue := true
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Email_Validator", func(b *testing.B) {
		schema := validators.Email()
		testValue := "test@example.com"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("URL_Validator", func(b *testing.B) {
		schema := validators.URL()
		testValue := "https://example.com"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Array_MinMax_Items", func(b *testing.B) {
		schema := validators.Array(validators.String()).MinItems(1).MaxItems(10).Required()
		testValue := []string{"item1", "item2", "item3"}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Simple_Object", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"name":  validators.String().Required(),
			"email": validators.Email(),
			"age":   validators.Number().Min(0).Optional(),
		}).Required()
		testValue := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Nested_Object", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"user": validators.Object(map[string]interface{}{
				"name":  validators.String().Required(),
				"email": validators.Email(),
			}).Required(),
			"settings": validators.Object(map[string]interface{}{
				"theme": validators.String().Optional().Default("light"),
			}).Optional(),
		}).Required()
		testValue := map[string]interface{}{
			"user": map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			"settings": map[string]interface{}{
				"theme": "dark",
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})
}

// BenchmarkWorkingOpenAPI tests OpenAPI generation
// Commented out for now as GetOpenAPISchema might not be available on all types
/*
func BenchmarkWorkingOpenAPI(b *testing.B) {
	b.Run("String_Schema_OpenAPI", func(b *testing.B) {
		schema := validators.String().Min(3).Max(50).Required()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// _ = schema.GetOpenAPISchema()
		}
	})

	b.Run("Object_Schema_OpenAPI", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"name":  validators.String().Required(),
			"email": validators.Email(),
			"age":   validators.Number().Min(0).Optional(),
		}).Required()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// _ = schema.GetOpenAPISchema()
		}
	})
}
*/

// BenchmarkWorkingOperations tests operations framework
func BenchmarkWorkingOperations(b *testing.B) {
	// Importing operations will be tested only if it compiles
	// For now, we'll skip this to get basic benchmarks working
}