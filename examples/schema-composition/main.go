package main

import (
	"fmt"

	"github.com/picogrid/go-op/operations"
	"github.com/picogrid/go-op/validators"
)

func main() {
	fmt.Println("🔗 OpenAPI 3.1 Schema Composition Examples")
	fmt.Println("========================================")

	// Create OpenAPI generator
	generator := operations.NewOpenAPIGenerator("Schema Composition Examples", "1.0.0")
	generator.SetDescription("Examples demonstrating OpenAPI 3.1 schema composition features")

	// Example 1: Basic schema with OpenAPI 3.1 features
	fmt.Println("\n1. Basic Schema with OpenAPI 3.1 Features:")
	basicSchema := createBasicSchema()
	printSchema("Basic Product Schema", basicSchema)

	// Example 2: Schema composition with inheritance
	fmt.Println("\n2. Schema Composition with Inheritance:")
	composedSchema := createComposedSchema()
	printSchema("Composed Product Schema", composedSchema)

	// Example 3: Polymorphic schemas
	fmt.Println("\n3. Polymorphic Schema Examples:")
	polymorphicSchemas := createPolymorphicSchemas()
	for name, schema := range polymorphicSchemas {
		printSchema(name, schema)
	}

	// Example 4: Advanced validation schemas
	fmt.Println("\n4. Advanced Validation Examples:")
	advancedSchemas := createAdvancedValidationSchemas()
	for name, schema := range advancedSchemas {
		printSchema(name, schema)
	}

	fmt.Println("\n✅ All schema examples generated successfully!")
	fmt.Println("\nTo see these schemas in a real API, check out the advanced-api example:")
	fmt.Println("  go run ./examples/advanced-api/main.go")
}

// createBasicSchema demonstrates basic OpenAPI 3.1 schema features
func createBasicSchema() interface{} {
	// Note: This demonstrates the schema structure that would be generated
	// The actual OpenAPI Fixed Fields are added during the generation process
	return validators.Object(map[string]interface{}{
		"id": validators.String().
			Min(1).
			Pattern("^prod_[a-zA-Z0-9]+$").
			Required(),
		"name": validators.String().
			Min(1).
			Max(200).
			Required(),
		"price": validators.Number().
			Min(0).
			Required(),
		"currency": validators.String().
			Min(3).
			Max(3).
			Required(),
		"tags": validators.Array(validators.String().Min(1).Max(30)).
			Optional(),
		"metadata": validators.Object(map[string]interface{}{
			"weight":   validators.Number().Min(0).Optional(),
			"category": validators.String().Min(1).Optional(),
		}).Optional(),
	}).Required()
}

// createComposedSchema demonstrates schema composition
func createComposedSchema() interface{} {
	// This demonstrates how schemas can be composed
	// The actual composition would use allOf in the OpenAPI spec
	return validators.Object(map[string]interface{}{
		// Base entity fields (would be in baseEntity schema)
		"id":         validators.String().Min(1).Required(),
		"created_at": validators.String().Required(),
		"updated_at": validators.String().Required(),
		// Product-specific fields (would be in productFields schema)
		"name":     validators.String().Min(1).Max(200).Required(),
		"price":    validators.Number().Min(0).Required(),
		"currency": validators.String().Min(3).Max(3).Required(),
		"in_stock": validators.Bool().Required(),
	}).Required()
}

// createPolymorphicSchemas demonstrates polymorphic schema patterns
func createPolymorphicSchemas() map[string]interface{} {
	schemas := make(map[string]interface{})

	// Base payment method
	schemas["PaymentMethod"] = validators.Object(map[string]interface{}{
		"type":   validators.String().Required(),
		"amount": validators.Number().Min(0).Required(),
	}).Required()

	// Credit card payment (extends PaymentMethod)
	schemas["CreditCardPayment"] = validators.Object(map[string]interface{}{
		"type":         validators.String().Required(),
		"amount":       validators.Number().Min(0).Required(),
		"card_number":  validators.String().Min(16).Max(19).Pattern("^[0-9]+$").Required(),
		"expiry_month": validators.Number().Min(1).Max(12).Required(),
		"expiry_year":  validators.Number().Min(2024).Required(),
	}).Required()

	// Bank transfer payment (extends PaymentMethod)
	schemas["BankTransferPayment"] = validators.Object(map[string]interface{}{
		"type":           validators.String().Required(),
		"amount":         validators.Number().Min(0).Required(),
		"account_number": validators.String().Min(8).Max(20).Required(),
		"routing_number": validators.String().Min(9).Max(9).Required(),
	}).Required()

	return schemas
}

// createAdvancedValidationSchemas demonstrates advanced validation features
func createAdvancedValidationSchemas() map[string]interface{} {
	schemas := make(map[string]interface{})

	// Strict numeric validation
	schemas["PrecisePrice"] = validators.Number().
		Min(0).
		Max(999999.99).
		Required()

	// Array with constraints
	schemas["TagList"] = validators.Array(validators.String().Min(1).Max(30)).
		Optional()

	// Flexible object with additional properties
	schemas["FlexibleMetadata"] = validators.Object(map[string]interface{}{
		"required_field": validators.String().Required(),
	}).Optional()

	// Conditional schema (would be oneOf in OpenAPI)
	schemas["ConditionalData"] = validators.Object(map[string]interface{}{
		"type": validators.String().Required(),
		"data": validators.Object(map[string]interface{}{}).Optional(),
	}).Required()

	return schemas
}

// printSchema prints a schema in a readable format
func printSchema(name string, schema interface{}) {
	fmt.Printf("\n%s:\n", name)

	// For demonstration purposes, we'll show the structure
	// In a real implementation, this would generate proper OpenAPI JSON
	switch s := schema.(type) {
	case validators.RequiredObjectBuilder:
		fmt.Printf("  Type: object (required)\n")
		fmt.Printf("  Description: Example schema showcasing OpenAPI 3.1 features\n")
	case validators.OptionalObjectBuilder:
		fmt.Printf("  Type: object (optional)\n")
		fmt.Printf("  Description: Example schema showcasing OpenAPI 3.1 features\n")
	default:
		fmt.Printf("  Type: %T\n", s)
	}

	fmt.Printf("  OpenAPI 3.1 Features:\n")
	fmt.Printf("  • Enhanced JSON Schema validation\n")
	fmt.Printf("  • Rich metadata support\n")
	fmt.Printf("  • Schema composition capabilities\n")
}

// demonstrateOpenAPIGeneration shows how these schemas would be used in OpenAPI generation
func demonstrateOpenAPIGeneration() {
	fmt.Println("\n📋 OpenAPI Generation Example:")
	fmt.Println("==============================")

	// Create a generator
	generator := operations.NewOpenAPIGenerator("Schema Demo API", "1.0.0")

	// Set comprehensive metadata
	generator.SetDescription("API demonstrating advanced schema composition")
	generator.SetContact(&operations.OpenAPIContact{
		Name:  "Schema Team",
		Email: "schemas@example.com",
	})

	// Add tags
	generator.AddTag(operations.OpenAPITag{
		Name:        "schemas",
		Description: "Schema composition examples",
	})

	fmt.Println("✅ Generator configured with:")
	fmt.Println("  • Enhanced metadata")
	fmt.Println("  • Contact information")
	fmt.Println("  • Organized tags")
	fmt.Println("  • Ready for complex schema composition")
}

// Example of how to create helper functions for common schema patterns
func createCommonSchemas() map[string]interface{} {
	return map[string]interface{}{
		"ID": validators.String().
			Min(1).
			Pattern("^[a-zA-Z0-9_-]+$").
			Required(),

		"Timestamp": validators.String().
			Required(),

		"Money": validators.Number().
			Min(0).
			Required(),

		"Email": validators.String().
			Email().
			Required(),

		"URL": validators.String().
			URL().
			Optional(),
	}
}
