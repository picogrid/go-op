package benchmarks

import (
	"testing"

	"github.com/picogrid/go-op"
	"github.com/picogrid/go-op/operations"
	"github.com/picogrid/go-op/validators"
)

// BenchmarkToOpenAPISchema tests schema to OpenAPI conversion performance
func BenchmarkToOpenAPISchema(b *testing.B) {
	// Simple types
	b.Run("String_Simple", func(b *testing.B) {
		schema := validators.String().Required()
		enhancedSchema := schema.(goop.EnhancedSchema)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = enhancedSchema.ToOpenAPISchema()
		}
	})

	b.Run("String_With_Constraints", func(b *testing.B) {
		schema := validators.String().
			Min(5).
			Max(100).
			Pattern(`^[a-zA-Z0-9_]+$`).
			Required()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	b.Run("Number_Simple", func(b *testing.B) {
		schema := validators.Number().Required()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	b.Run("Number_With_Constraints", func(b *testing.B) {
		schema := validators.Number().
			Min(0).
			Max(1000).
			Required()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	b.Run("Array_Simple", func(b *testing.B) {
		schema := validators.Array(validators.String()).Required()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	b.Run("Array_With_Constraints", func(b *testing.B) {
		schema := validators.Array(validators.String().Min(1).Max(50)).Required()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	// Object types
	b.Run("Object_Simple", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"id":   validators.String().Required(),
			"name": validators.String().Required(),
		}).Required()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	b.Run("Object_Medium", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"id":       validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
			"username": validators.String().Min(3).Max(30).Pattern(`^[a-zA-Z0-9_]+$`).Required(),
			"email":    validators.Email(),
			"age":      validators.Number().Min(18).Max(120).Optional(),
			"bio":      validators.String().Max(500).Optional(),
			"verified": validators.Bool().Optional(),
		}).Required()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	// Special validators
	b.Run("Email_Validator", func(b *testing.B) {
		schema := validators.Email()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	b.Run("UUID_Validator", func(b *testing.B) {
		schema := validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	b.Run("Pattern_Enum", func(b *testing.B) {
		schema := validators.String().Pattern(`^(draft|published|archived)$`).Required()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})
}

// BenchmarkComplexNestedSchema tests deeply nested schema generation
func BenchmarkComplexNestedSchema(b *testing.B) {
	b.Run("Nested_2_Levels", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"user": validators.Object(map[string]interface{}{
				"id":    validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
				"name":  validators.String().Required(),
				"email": validators.Email(),
			}).Required(),
			"metadata": validators.Object(map[string]interface{}{
				"createdAt": validators.String().Required(),
				"updatedAt": validators.String().Required(),
			}).Required(),
		}).Required()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	b.Run("Nested_3_Levels", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"order": validators.Object(map[string]interface{}{
				"id": validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
				"customer": validators.Object(map[string]interface{}{
					"id":   validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
					"name": validators.String().Required(),
					"address": validators.Object(map[string]interface{}{
						"street":  validators.String().Required(),
						"city":    validators.String().Required(),
						"country": validators.String().Required(),
					}).Required(),
				}).Required(),
				"items": validators.Array(validators.Object(map[string]interface{}{
					"productId": validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
					"quantity":  validators.Number().Min(1).Required(),
					"price":     validators.Number().Min(0).Required(),
				})).Required(),
			}).Required(),
		}).Required()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	b.Run("Array_Of_Objects", func(b *testing.B) {
		schema := validators.Array(
			validators.Object(map[string]interface{}{
				"id":       validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
				"title":    validators.String().Min(1).Max(200).Required(),
				"content":  validators.String().Min(1).Max(10000).Required(),
				"author":   validators.String().Required(),
				"tags":     validators.Array(validators.String()).Optional(),
				"metadata": validators.Object(nil).Optional(),
			}),
		).Required()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})
}

// BenchmarkOperationToOpenAPI tests operation building performance
// Note: Full OpenAPI generation would require proper operation setup
func BenchmarkOperationToOpenAPI(b *testing.B) {
	b.Run("Simple_Operation_Building", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = operations.NewSimple().
				GET("/users").
				Summary("List users")
		}
	})

	b.Run("Operation_With_Schemas", func(b *testing.B) {
		paramsSchema := validators.Object(map[string]interface{}{
			"id": validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
		}).Required()

		responseSchema := validators.Object(map[string]interface{}{
			"id":       validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
			"username": validators.String().Required(),
			"email":    validators.String().Required(),
		}).Required()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = operations.NewSimple().
				GET("/users/{id}").
				Summary("Get user by ID").
				WithParams(paramsSchema).
				WithResponse(responseSchema)
		}
	})
}

// BenchmarkMemoryUsage tests memory usage during generation
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("Large_Schema_Generation", func(b *testing.B) {
		// Create a large schema with many fields
		fields := make(map[string]interface{})
		for i := 0; i < 100; i++ {
			fieldName := string(rune('a'+i%26)) + string(rune('0'+i/26))
			switch i % 5 {
			case 0:
				fields[fieldName] = validators.String().Min(1).Max(100).Optional()
			case 1:
				fields[fieldName] = validators.Number().Min(0).Max(1000).Optional()
			case 2:
				fields[fieldName] = validators.Bool().Optional()
			case 3:
				fields[fieldName] = validators.Array(validators.String()).Optional()
			case 4:
				fields[fieldName] = validators.Object(nil).Optional()
			}
		}
		schema := validators.Object(fields).Required()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	b.Run("Many_Schema_Conversions", func(b *testing.B) {
		// Create many schemas and convert to OpenAPI
		schemas := make([]goop.EnhancedSchema, 100)
		for i := 0; i < 100; i++ {
			schemas[i] = validators.Object(map[string]interface{}{
				"id":    validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
				"name":  validators.String().Required(),
				"value": validators.Number().Min(0).Max(100).Required(),
			}).Required().(goop.EnhancedSchema)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, schema := range schemas {
				_ = schema.ToOpenAPISchema()
			}
		}
	})

	b.Run("Deep_Recursion_Schema", func(b *testing.B) {
		// Create a deeply recursive schema structure
		var createNestedObject func(depth int) interface{}
		createNestedObject = func(depth int) interface{} {
			if depth <= 0 {
				return validators.String().Required()
			}
			return validators.Object(map[string]interface{}{
				"value": validators.String().Required(),
				"child": createNestedObject(depth - 1),
			}).Required()
		}

		schema := createNestedObject(10) // 10 levels deep

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})
}

// BenchmarkOpenAPIExtensions tests various validation patterns
func BenchmarkOpenAPIExtensions(b *testing.B) {
	b.Run("String_With_Messages", func(b *testing.B) {
		schema := validators.String().
			Min(1).WithMinLengthMessage("Too short").
			Max(100).WithMaxLengthMessage("Too long").
			Required()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})

	b.Run("Complex_Object_Schema", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"type":      validators.String().Pattern(`^(email|sms|push)$`).Required(),
			"recipient": validators.String().Required(),
			"message":   validators.String().Max(1000).Required(),
			"metadata": validators.Object(map[string]interface{}{
				"priority": validators.String().Pattern(`^(low|medium|high)$`).Optional(),
				"retries":  validators.Number().Min(0).Max(10).Optional(),
			}).Optional(),
		}).Required()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if enhanced, ok := schema.(goop.EnhancedSchema); ok {
				_ = enhanced.ToOpenAPISchema()
			}
		}
	})
}
