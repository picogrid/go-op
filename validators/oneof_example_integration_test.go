package validators

import (
	"testing"

	goop "github.com/picogrid/go-op"
)

func TestOneOfWithExampleIntegration(t *testing.T) {
	// Test OneOf schema with examples on individual schemas
	userSchema := Object(map[string]interface{}{
		"type": String().Required(),
		"id":   Number().Required(),
		"name": String().Required(),
	}).Example(map[string]interface{}{
		"type": "user",
		"id":   123,
		"name": "John Doe",
	}).Required()

	adminSchema := Object(map[string]interface{}{
		"type":        String().Required(),
		"id":          Number().Required(),
		"name":        String().Required(),
		"permissions": Array(String()).Required(),
	}).Example(map[string]interface{}{
		"type":        "admin",
		"id":          456,
		"name":        "Jane Admin",
		"permissions": []string{"read", "write", "delete"},
	}).Required()

	// Create OneOf schema with examples
	oneOfSchema := OneOf(userSchema, adminSchema).Required()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := oneOfSchema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("OneOf schema does not implement EnhancedSchema interface")
	}

	// Generate OpenAPI schema
	openAPISchema := enhancedSchema.ToOpenAPISchema()

	// Verify OneOf structure
	if len(openAPISchema.OneOf) != 2 {
		t.Errorf("Expected 2 OneOf schemas, got %d", len(openAPISchema.OneOf))
	}

	// Verify examples are preserved in OneOf schemas
	for i, subSchema := range openAPISchema.OneOf {
		if subSchema.Example == nil {
			t.Errorf("OneOf schema %d should have an example", i)
		}

		// Verify the examples contain expected structure
		if example, ok := subSchema.Example.(map[string]interface{}); ok {
			if example["type"] == nil {
				t.Errorf("OneOf schema %d example should have 'type' field", i)
			}
			if example["id"] == nil {
				t.Errorf("OneOf schema %d example should have 'id' field", i)
			}
			if example["name"] == nil {
				t.Errorf("OneOf schema %d example should have 'name' field", i)
			}

			// Check specific examples
			if i == 0 && example["type"] != "user" {
				t.Errorf("First OneOf schema should be user type, got %v", example["type"])
			}
			if i == 1 && example["type"] != "admin" {
				t.Errorf("Second OneOf schema should be admin type, got %v", example["type"])
			}
		} else {
			t.Errorf("OneOf schema %d example should be a map, got %T", i, subSchema.Example)
		}
	}
}

func TestOneOfWithMultipleExamplesIntegration(t *testing.T) {
	// Test OneOf with multiple examples per schema
	stringExamples := map[string]ExampleObject{
		"short": {
			Summary:     "Short string",
			Description: "A short example string",
			Value:       "hello",
		},
		"long": {
			Summary:     "Long string",
			Description: "A longer example string",
			Value:       "this is a much longer example string",
		},
	}

	numberExamples := map[string]ExampleObject{
		"small": {
			Summary:     "Small number",
			Description: "A small number example",
			Value:       1,
		},
		"large": {
			Summary:     "Large number",
			Description: "A large number example",
			Value:       1000000,
		},
	}

	stringSchema := String().Min(1).Examples(stringExamples).Required()
	numberSchema := Number().Min(0).Examples(numberExamples).Required()

	// Create OneOf with multiple examples
	oneOfSchema := OneOf(stringSchema, numberSchema).Required()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := oneOfSchema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("OneOf schema does not implement EnhancedSchema interface")
	}

	// Generate OpenAPI schema
	openAPISchema := enhancedSchema.ToOpenAPISchema()

	// Verify OneOf structure
	if len(openAPISchema.OneOf) != 2 {
		t.Errorf("Expected 2 OneOf schemas, got %d", len(openAPISchema.OneOf))
	}

	// Verify that schemas maintain their individual properties
	// First schema should be string
	if openAPISchema.OneOf[0].Type != "string" {
		t.Errorf("First OneOf schema should be string type, got %s", openAPISchema.OneOf[0].Type)
	}

	// Second schema should be number
	if openAPISchema.OneOf[1].Type != "number" {
		t.Errorf("Second OneOf schema should be number type, got %s", openAPISchema.OneOf[1].Type)
	}
}

func TestOneOfValidationWithExamples(t *testing.T) {
	// Test that OneOf schemas with examples still validate correctly
	emailSchema := String().Email().Example("user@example.com").Required()
	phoneSchema := String().Pattern(`^\+?[1-9]\d{1,14}$`).Example("+1234567890").Required()

	contactSchema := OneOf(emailSchema, phoneSchema).Required()

	// Test validation with email example
	emailExample := "test@example.com"
	if err := contactSchema.Validate(emailExample); err != nil {
		t.Errorf("Email example should validate successfully: %v", err)
	}

	// Test validation with phone example
	phoneExample := "+1234567890"
	if err := contactSchema.Validate(phoneExample); err != nil {
		t.Errorf("Phone example should validate successfully: %v", err)
	}

	// Test validation with invalid input
	invalidExample := "not-email-or-phone"
	if err := contactSchema.Validate(invalidExample); err == nil {
		t.Error("Invalid input should fail validation")
	}
}

func TestNestedOneOfWithExamples(t *testing.T) {
	// Test nested OneOf with examples
	basicUserSchema := Object(map[string]interface{}{
		"type": String().Required(),
		"name": String().Required(),
	}).Example(map[string]interface{}{
		"type": "basic",
		"name": "Basic User",
	}).Required()

	premiumUserSchema := Object(map[string]interface{}{
		"type":    String().Required(),
		"name":    String().Required(),
		"premium": Bool().Required(),
	}).Example(map[string]interface{}{
		"type":    "premium",
		"name":    "Premium User",
		"premium": true,
	}).Required()

	userTypeSchema := OneOf(basicUserSchema, premiumUserSchema).Required()

	// Create a schema that has OneOf as a property
	responseSchema := Object(map[string]interface{}{
		"status": String().Required(),
		"user":   userTypeSchema,
	}).Example(map[string]interface{}{
		"status": "success",
		"user": map[string]interface{}{
			"type":    "premium",
			"name":    "Premium User",
			"premium": true,
		},
	}).Required()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := responseSchema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("Response schema does not implement EnhancedSchema interface")
	}

	// Generate OpenAPI schema
	openAPISchema := enhancedSchema.ToOpenAPISchema()

	// Verify structure
	if openAPISchema.Type != "object" {
		t.Errorf("Expected object type, got %s", openAPISchema.Type)
	}

	// Verify example is present
	if openAPISchema.Example == nil {
		t.Error("Response schema should have an example")
	}

	// Verify user property has OneOf
	if userProp, ok := openAPISchema.Properties["user"]; ok {
		if len(userProp.OneOf) != 2 {
			t.Errorf("User property should have 2 OneOf schemas, got %d", len(userProp.OneOf))
		}
	} else {
		t.Error("Response schema should have 'user' property")
	}
}

func TestOneOfExampleValidation(t *testing.T) {
	// Test that examples in OneOf schemas are valid according to their own constraints

	// Create schemas with specific constraints and examples that should match
	strictStringSchema := String().Min(10).Max(20).Example("exactly10chars").Required()
	strictNumberSchema := Number().Min(100).Max(200).Example(150).Required()

	oneOfSchema := OneOf(strictStringSchema, strictNumberSchema).Required()

	// Validate the examples against their respective schemas
	if err := strictStringSchema.Validate("exactly10chars"); err != nil {
		t.Errorf("String example should be valid for its schema: %v", err)
	}

	if err := strictNumberSchema.Validate(150); err != nil {
		t.Errorf("Number example should be valid for its schema: %v", err)
	}

	// Test validation through OneOf
	if err := oneOfSchema.Validate("exactly10chars"); err != nil {
		t.Errorf("String example should validate through OneOf: %v", err)
	}

	if err := oneOfSchema.Validate(150); err != nil {
		t.Errorf("Number example should validate through OneOf: %v", err)
	}
}

func TestComplexOneOfExampleScenario(t *testing.T) {
	// Test a complex real-world scenario: payment method OneOf with examples

	creditCardSchema := Object(map[string]interface{}{
		"type":        String().Required(),
		"card_number": String().Pattern(`^\d{16}$`).Required(),
		"expiry":      String().Pattern(`^\d{2}/\d{2}$`).Required(),
		"cvv":         String().Pattern(`^\d{3,4}$`).Required(),
	}).Example(map[string]interface{}{
		"type":        "credit_card",
		"card_number": "1234567890123456",
		"expiry":      "12/25",
		"cvv":         "123",
	}).Required()

	paypalSchema := Object(map[string]interface{}{
		"type":  String().Required(),
		"email": String().Email().Required(),
	}).Example(map[string]interface{}{
		"type":  "paypal",
		"email": "user@example.com",
	}).Required()

	bankTransferSchema := Object(map[string]interface{}{
		"type":           String().Required(),
		"account_number": String().Pattern(`^\d{10,12}$`).Required(),
		"routing_number": String().Pattern(`^\d{9}$`).Required(),
	}).Example(map[string]interface{}{
		"type":           "bank_transfer",
		"account_number": "1234567890",
		"routing_number": "123456789",
	}).Required()

	paymentMethodSchema := OneOf(creditCardSchema, paypalSchema, bankTransferSchema).Required()

	// Type assert to access OpenAPI generation methods
	enhancedSchema, ok := paymentMethodSchema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("Payment method schema does not implement EnhancedSchema interface")
	}

	// Generate OpenAPI schema
	openAPISchema := enhancedSchema.ToOpenAPISchema()

	// Verify OneOf structure
	if len(openAPISchema.OneOf) != 3 {
		t.Errorf("Expected 3 OneOf schemas (credit card, paypal, bank transfer), got %d", len(openAPISchema.OneOf))
	}

	// Verify each OneOf schema has examples
	expectedTypes := []string{"credit_card", "paypal", "bank_transfer"}
	for i, subSchema := range openAPISchema.OneOf {
		if subSchema.Example == nil {
			t.Errorf("OneOf schema %d should have an example", i)
			continue
		}

		if example, ok := subSchema.Example.(map[string]interface{}); ok {
			actualType, hasType := example["type"]
			if !hasType {
				t.Errorf("OneOf schema %d example should have 'type' field", i)
				continue
			}

			if actualType != expectedTypes[i] {
				t.Errorf("OneOf schema %d should have type '%s', got '%v'", i, expectedTypes[i], actualType)
			}
		} else {
			t.Errorf("OneOf schema %d example should be a map, got %T", i, subSchema.Example)
		}
	}

	// Test validation with examples
	creditCardExample := map[string]interface{}{
		"type":        "credit_card",
		"card_number": "1234567890123456",
		"expiry":      "12/25",
		"cvv":         "123",
	}

	if err := paymentMethodSchema.Validate(creditCardExample); err != nil {
		t.Errorf("Credit card example should validate: %v", err)
	}

	paypalExample := map[string]interface{}{
		"type":  "paypal",
		"email": "user@example.com",
	}

	if err := paymentMethodSchema.Validate(paypalExample); err != nil {
		t.Errorf("PayPal example should validate: %v", err)
	}
}
