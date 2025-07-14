package benchmarks

import (
	"fmt"
	"testing"

	"github.com/picogrid/go-op/validators"
)

// BenchmarkStringValidators tests string validation performance
func BenchmarkStringValidators(b *testing.B) {
	b.Run("Min_Length", func(b *testing.B) {
		schema := validators.String().Min(5).Required()
		testValue := "hello world"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Max_Length", func(b *testing.B) {
		schema := validators.String().Max(100).Required()
		testValue := "hello world"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Min_Max_Length", func(b *testing.B) {
		schema := validators.String().Min(5).Max(100).Required()
		testValue := "hello world"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Pattern_Simple", func(b *testing.B) {
		schema := validators.String().Pattern(`^[a-zA-Z]+$`).Required()
		testValue := "helloworld"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Pattern_Complex", func(b *testing.B) {
		schema := validators.String().Pattern(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]+$`).Required()
		testValue := "Password123!"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Email_Valid", func(b *testing.B) {
		schema := validators.Email()
		testValue := "test@example.com"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("URL_Valid", func(b *testing.B) {
		schema := validators.URL()
		testValue := "https://example.com/path"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Optional_With_Default", func(b *testing.B) {
		schema := validators.String().Optional().Default("default")
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(nil)
		}
	})
}

// BenchmarkNumberValidators tests number validation performance
func BenchmarkNumberValidators(b *testing.B) {
	b.Run("Min_Check", func(b *testing.B) {
		schema := validators.Number().Min(0).Required()
		testValue := 42.5
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Max_Check", func(b *testing.B) {
		schema := validators.Number().Max(100).Required()
		testValue := 42.5
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Range_Check", func(b *testing.B) {
		schema := validators.Number().Min(0).Max(100).Required()
		testValue := 42.5
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Integer_Check", func(b *testing.B) {
		schema := validators.Number().Integer().Required()
		testValue := 42.0
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Positive_Number", func(b *testing.B) {
		schema := validators.Number().Positive().Required()
		testValue := 42.5
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Negative_Number", func(b *testing.B) {
		schema := validators.Number().Negative().Required()
		testValue := -42.5
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Type_Conversion_Int", func(b *testing.B) {
		schema := validators.Number().Required()
		testValue := 42
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Type_Conversion_Float32", func(b *testing.B) {
		schema := validators.Number().Required()
		testValue := float32(42.5)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})
}

// BenchmarkArrayValidators tests array validation performance
func BenchmarkArrayValidators(b *testing.B) {
	b.Run("Small_Array_5", func(b *testing.B) {
		schema := validators.Array(validators.String()).Required()
		testValue := []string{"a", "b", "c", "d", "e"}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Medium_Array_50", func(b *testing.B) {
		schema := validators.Array(validators.String()).Required()
		testValue := make([]string, 50)
		for i := range testValue {
			testValue[i] = fmt.Sprintf("item%d", i)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Large_Array_500", func(b *testing.B) {
		schema := validators.Array(validators.String()).Required()
		testValue := make([]string, 500)
		for i := range testValue {
			testValue[i] = fmt.Sprintf("item%d", i)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Array_With_Complex_Items", func(b *testing.B) {
		itemSchema := validators.Object(map[string]interface{}{
			"id":   validators.String().Required(),
			"name": validators.String().Min(1).Required(),
			"age":  validators.Number().Min(0).Required(),
		}).Required()
		schema := validators.Array(itemSchema).Required()

		testValue := []map[string]interface{}{
			{"id": "1", "name": "John", "age": 30},
			{"id": "2", "name": "Jane", "age": 25},
			{"id": "3", "name": "Bob", "age": 35},
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("MinItems_Check", func(b *testing.B) {
		schema := validators.Array(validators.String()).MinItems(3).Required()
		testValue := []string{"a", "b", "c", "d", "e"}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("MaxItems_Check", func(b *testing.B) {
		schema := validators.Array(validators.String()).MaxItems(10).Required()
		testValue := []string{"a", "b", "c", "d", "e"}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})
}

// BenchmarkObjectValidators tests object validation performance
func BenchmarkObjectValidators(b *testing.B) {
	b.Run("Simple_Object_3_Fields", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"name":  validators.String().Required(),
			"email": validators.Email(),
			"age":   validators.Number().Min(0).Required(),
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

	b.Run("Complex_Object_10_Fields", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"id":        validators.String().Required(),
			"username":  validators.String().Min(3).Max(50).Required(),
			"email":     validators.Email(),
			"age":       validators.Number().Min(0).Max(150).Integer().Required(),
			"active":    validators.Bool().Required(),
			"score":     validators.Number().Min(0).Max(100).Required(),
			"tags":      validators.Array(validators.String()).MaxItems(10).Optional(),
			"createdAt": validators.String().Required(),
			"updatedAt": validators.String().Required(),
			"metadata":  validators.Object(map[string]interface{}{}).Optional(),
		}).Required()

		testValue := map[string]interface{}{
			"id":        "usr123",
			"username":  "johndoe",
			"email":     "john@example.com",
			"age":       30,
			"active":    true,
			"score":     85.5,
			"tags":      []string{"user", "premium"},
			"createdAt": "2024-01-01",
			"updatedAt": "2024-01-15",
			"metadata":  map[string]interface{}{},
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Nested_Object_2_Levels", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"user": validators.Object(map[string]interface{}{
				"name":  validators.String().Required(),
				"email": validators.Email(),
			}).Required(),
			"settings": validators.Object(map[string]interface{}{
				"theme":    validators.String().Optional().Default("light"),
				"language": validators.String().Optional().Default("en"),
			}).Optional(),
		}).Required()

		testValue := map[string]interface{}{
			"user": map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			"settings": map[string]interface{}{
				"theme":    "dark",
				"language": "en",
			},
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Deeply_Nested_5_Levels", func(b *testing.B) {
		level5 := validators.Object(map[string]interface{}{
			"value": validators.String().Required(),
		}).Required()

		level4 := validators.Object(map[string]interface{}{
			"nested": level5,
			"data":   validators.String().Required(),
		}).Required()

		level3 := validators.Object(map[string]interface{}{
			"nested": level4,
			"data":   validators.String().Required(),
		}).Required()

		level2 := validators.Object(map[string]interface{}{
			"nested": level3,
			"data":   validators.String().Required(),
		}).Required()

		schema := validators.Object(map[string]interface{}{
			"nested": level2,
			"data":   validators.String().Required(),
		}).Required()

		testValue := map[string]interface{}{
			"data": "level1",
			"nested": map[string]interface{}{
				"data": "level2",
				"nested": map[string]interface{}{
					"data": "level3",
					"nested": map[string]interface{}{
						"data": "level4",
						"nested": map[string]interface{}{
							"value": "level5",
						},
					},
				},
			},
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Partial_Object", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"name":  validators.String().Optional(),
			"email": validators.String().Optional(),
			"age":   validators.Number().Optional(),
		}).Partial().Required()

		testValue := map[string]interface{}{
			"name": "John Doe",
			// email and age are omitted
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Strict_Object", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"name":  validators.String().Required(),
			"email": validators.Email(),
		}).Strict().Required()

		testValue := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			// No extra fields
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})
}

// BenchmarkErrorHandling tests error generation and handling
func BenchmarkErrorHandling(b *testing.B) {
	b.Run("Single_String_Error", func(b *testing.B) {
		schema := validators.String().Min(10).Required()
		testValue := "short"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := schema.Validate(testValue)
			_ = err
		}
	})

	b.Run("Multiple_Field_Errors", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"username": validators.String().Min(3).Max(20).Pattern(`^[a-zA-Z0-9]+$`).Required(),
			"email":    validators.Email(),
			"password": validators.String().Min(8).Required(),
			"age":      validators.Number().Min(18).Max(120).Integer().Required(),
		}).Required()

		invalidData := map[string]interface{}{
			"username": "u",             // Too short
			"email":    "invalid-email", // Invalid format
			"password": "weak",          // Too short
			"age":      150,             // Too high
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := schema.Validate(invalidData)
			_ = err
		}
	})

	b.Run("Nested_Errors", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"user": validators.Object(map[string]interface{}{
				"profile": validators.Object(map[string]interface{}{
					"name":  validators.String().Min(1).Required(),
					"email": validators.Email(),
				}).Required(),
			}).Required(),
		}).Required()

		invalidData := map[string]interface{}{
			"user": map[string]interface{}{
				"profile": map[string]interface{}{
					"name":  "",             // Empty
					"email": "not-an-email", // Invalid
				},
			},
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := schema.Validate(invalidData)
			_ = err
		}
	})

	b.Run("Error_Message_Formatting", func(b *testing.B) {
		schema := validators.String().
			Min(5).WithMinLengthMessage("Must be at least {min} characters").
			Max(20).WithMaxLengthMessage("Cannot exceed {max} characters").
			Pattern(`^[a-zA-Z0-9_]+$`).WithPatternMessage("Only alphanumeric and underscore allowed").
			Required().WithRequiredMessage("This field is required")

		testValues := []string{
			"ab",                               // Too short
			"verylongusernamethatexceedslimit", // Too long
			"invalid-chars!",                   // Pattern mismatch
			"",                                 // Empty
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			value := testValues[i%len(testValues)]
			err := schema.Validate(value)
			if err != nil {
				_ = err.Error()
			}
		}
	})
}

// BenchmarkMixedScenarios tests realistic mixed validation scenarios
func BenchmarkMixedScenarios(b *testing.B) {
	b.Run("User_Registration_Schema", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"username": validators.String().Min(3).Max(30).Pattern(`^[a-zA-Z0-9_]+$`).Required(),
			"email":    validators.Email(),
			"password": validators.String().Min(8).Pattern(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)`).Required(),
			"profile": validators.Object(map[string]interface{}{
				"firstName": validators.String().Min(1).Max(50).Required(),
				"lastName":  validators.String().Min(1).Max(50).Required(),
				"age":       validators.Number().Min(13).Max(120).Integer().Optional(),
				"bio":       validators.String().Max(500).Optional(),
			}).Required(),
			"preferences": validators.Object(map[string]interface{}{
				"newsletter": validators.Bool().Optional().Default(true),
				"theme":      validators.String().Optional().Default("light"),
			}).Optional(),
		}).Required()

		validData := map[string]interface{}{
			"username": "john_doe_123",
			"email":    "john@example.com",
			"password": "SecurePass123",
			"profile": map[string]interface{}{
				"firstName": "John",
				"lastName":  "Doe",
				"age":       25,
				"bio":       "Software developer",
			},
			"preferences": map[string]interface{}{
				"newsletter": false,
				"theme":      "dark",
			},
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(validData)
		}
	})

	b.Run("E_Commerce_Order", func(b *testing.B) {
		addressSchema := validators.Object(map[string]interface{}{
			"street":     validators.String().Required(),
			"city":       validators.String().Required(),
			"state":      validators.String().Min(2).Max(2).Required(),
			"postalCode": validators.String().Pattern(`^\d{5}$`).Required(),
		}).Required()

		itemSchema := validators.Object(map[string]interface{}{
			"productId": validators.String().Required(),
			"quantity":  validators.Number().Min(1).Integer().Required(),
			"price":     validators.Number().Min(0).Required(),
		}).Required()

		schema := validators.Object(map[string]interface{}{
			"orderId":    validators.String().Required(),
			"customerId": validators.String().Required(),
			"items":      validators.Array(itemSchema).MinItems(1).Required(),
			"shipping":   addressSchema,
			"billing":    addressSchema,
			"total":      validators.Number().Min(0).Required(),
			"tax":        validators.Number().Min(0).Required(),
			"status":     validators.String().Pattern(`^(pending|processing|shipped|delivered)$`).Required(),
		}).Required()

		validOrder := map[string]interface{}{
			"orderId":    "ORD-12345",
			"customerId": "CUST-67890",
			"items": []map[string]interface{}{
				{"productId": "PROD-001", "quantity": 2, "price": 29.99},
				{"productId": "PROD-002", "quantity": 1, "price": 49.99},
			},
			"shipping": map[string]interface{}{
				"street":     "123 Main St",
				"city":       "Springfield",
				"state":      "IL",
				"postalCode": "62701",
			},
			"billing": map[string]interface{}{
				"street":     "123 Main St",
				"city":       "Springfield",
				"state":      "IL",
				"postalCode": "62701",
			},
			"total":  109.97,
			"tax":    8.80,
			"status": "pending",
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(validOrder)
		}
	})
}

// BenchmarkConcurrentValidation tests thread safety and performance
func BenchmarkComprehensiveConcurrentValidation(b *testing.B) {
	schema := validators.String().Min(3).Max(50).Required()
	testValue := "concurrent_test"

	b.Run("Sequential", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(testValue)
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = schema.Validate(testValue)
			}
		})
	})
}
