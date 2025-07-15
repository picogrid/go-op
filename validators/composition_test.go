package validators

import (
	"testing"

	goop "github.com/picogrid/go-op"
)

func TestOneOfValidation(t *testing.T) {
	// Create test schemas
	stringSchema := String().Min(1).Required()
	numberSchema := Number().Min(0).Required()

	// OneOf schema - should validate string OR number, but not both patterns
	oneOfSchema := OneOf(stringSchema, numberSchema).Required()

	// Test valid string
	if err := oneOfSchema.Validate("hello"); err != nil {
		t.Errorf("Expected valid string to pass OneOf validation: %v", err)
	}

	// Test valid number
	if err := oneOfSchema.Validate(42); err != nil {
		t.Errorf("Expected valid number to pass OneOf validation: %v", err)
	}

	// Test invalid data (empty string fails string validation, negative number fails number validation)
	if err := oneOfSchema.Validate(""); err == nil {
		t.Error("Expected empty string to fail OneOf validation")
	}

	// Test proper OneOf case with distinct schema types
	stringSchema2 := String().Pattern("^test$").Required()
	numberSchema2 := Number().Min(100).Max(200).Required()
	properOneOf := OneOf(stringSchema2, numberSchema2).Required()

	// Test with valid string
	if err := properOneOf.Validate("test"); err != nil {
		t.Errorf("Expected 'test' to pass OneOf validation: %v", err)
	}

	// Test with valid number
	if err := properOneOf.Validate(150); err != nil {
		t.Errorf("Expected 150 to pass OneOf validation: %v", err)
	}
}

func TestAllOfValidation(t *testing.T) {
	// Create schemas that can be combined
	minLengthSchema := String().Min(5).Required()
	maxLengthSchema := String().Max(10).Required()
	patternSchema := String().Pattern("^[a-z]+$").Required()

	// AllOf schema - must satisfy ALL constraints
	allOfSchema := AllOf(minLengthSchema, maxLengthSchema, patternSchema).Required()

	// Test valid data that satisfies all constraints
	if err := allOfSchema.Validate("hello"); err != nil {
		t.Errorf("Expected 'hello' to pass AllOf validation: %v", err)
	}

	// Test data that fails one constraint (too short)
	if err := allOfSchema.Validate("hi"); err == nil {
		t.Error("Expected 'hi' to fail AllOf validation (too short)")
	}

	// Test data that fails one constraint (too long)
	if err := allOfSchema.Validate("verylongstring"); err == nil {
		t.Error("Expected 'verylongstring' to fail AllOf validation (too long)")
	}

	// Test data that fails one constraint (invalid pattern)
	if err := allOfSchema.Validate("HELLO"); err == nil {
		t.Error("Expected 'HELLO' to fail AllOf validation (uppercase not allowed)")
	}
}

func TestAnyOfValidation(t *testing.T) {
	// Create different schema types
	emailSchema := String().Email().Required()
	phoneSchema := String().Pattern(`^\+?[1-9]\d{1,14}$`).Required()
	urlSchema := String().Pattern(`^https?://`).Required()

	// AnyOf schema - must satisfy AT LEAST ONE constraint
	anyOfSchema := AnyOf(emailSchema, phoneSchema, urlSchema).Required()

	// Test valid email
	if err := anyOfSchema.Validate("user@example.com"); err != nil {
		t.Errorf("Expected valid email to pass AnyOf validation: %v", err)
	}

	// Test valid phone
	if err := anyOfSchema.Validate("+1234567890"); err != nil {
		t.Errorf("Expected valid phone to pass AnyOf validation: %v", err)
	}

	// Test valid URL
	if err := anyOfSchema.Validate("https://example.com"); err != nil {
		t.Errorf("Expected valid URL to pass AnyOf validation: %v", err)
	}

	// Test data that doesn't match any schema
	if err := anyOfSchema.Validate("just text"); err == nil {
		t.Error("Expected 'just text' to fail AnyOf validation (doesn't match any pattern)")
	}
}

func TestNotValidation(t *testing.T) {
	// Create a schema to negate
	bannedWordsSchema := String().Pattern("(badword|spam|evil)").Required()

	// Not schema - should NOT match the pattern
	notSchema := Not(bannedWordsSchema).Required()

	// Test valid data (doesn't contain banned words)
	if err := notSchema.Validate("hello world"); err != nil {
		t.Errorf("Expected 'hello world' to pass Not validation: %v", err)
	}

	// Test invalid data (contains banned word)
	if err := notSchema.Validate("this is spam"); err == nil {
		t.Error("Expected 'this is spam' to fail Not validation (contains banned word)")
	}

	// Test another banned word
	if err := notSchema.Validate("badword here"); err == nil {
		t.Error("Expected 'badword here' to fail Not validation (contains banned word)")
	}
}

func TestCompositionWithOptionalSchemas(t *testing.T) {
	// Test OneOf with optional schema
	stringSchema := String().Required()
	numberSchema := Number().Required()
	optionalOneOf := OneOf(stringSchema, numberSchema).Optional()

	// Test with nil value (should be valid for optional)
	if err := optionalOneOf.Validate(nil); err != nil {
		t.Errorf("Expected nil to be valid for optional OneOf: %v", err)
	}

	// Test with default value
	defaultOneOf := OneOf(stringSchema, numberSchema).Optional().Default("default")
	if err := defaultOneOf.Validate(nil); err != nil {
		t.Errorf("Expected nil with default to be valid: %v", err)
	}
}

func TestCompositionOpenAPIGeneration(t *testing.T) {
	// Test OneOf OpenAPI generation
	stringSchema := String().Min(1).Required()
	numberSchema := Number().Min(0).Required()
	oneOfSchema := OneOf(stringSchema, numberSchema).Required()

	// Type assert to access OpenAPI generation
	enhancedSchema, ok := oneOfSchema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("OneOf schema does not implement EnhancedSchema interface")
	}

	// Generate OpenAPI schema
	openAPISchema := enhancedSchema.ToOpenAPISchema()

	// Verify OneOf is present
	if len(openAPISchema.OneOf) != 2 {
		t.Errorf("Expected 2 OneOf schemas, got %d", len(openAPISchema.OneOf))
	}

	// Test AllOf OpenAPI generation
	allOfSchema := AllOf(stringSchema, numberSchema).Required()
	enhancedAllOf, ok := allOfSchema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("AllOf schema does not implement EnhancedSchema interface")
	}

	allOfOpenAPI := enhancedAllOf.ToOpenAPISchema()
	if len(allOfOpenAPI.AllOf) != 2 {
		t.Errorf("Expected 2 AllOf schemas, got %d", len(allOfOpenAPI.AllOf))
	}

	// Test AnyOf OpenAPI generation
	anyOfSchema := AnyOf(stringSchema, numberSchema).Required()
	enhancedAnyOf, ok := anyOfSchema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("AnyOf schema does not implement EnhancedSchema interface")
	}

	anyOfOpenAPI := enhancedAnyOf.ToOpenAPISchema()
	if len(anyOfOpenAPI.AnyOf) != 2 {
		t.Errorf("Expected 2 AnyOf schemas, got %d", len(anyOfOpenAPI.AnyOf))
	}

	// Test Not OpenAPI generation
	notSchema := Not(stringSchema).Required()
	enhancedNot, ok := notSchema.(goop.EnhancedSchema)
	if !ok {
		t.Fatal("Not schema does not implement EnhancedSchema interface")
	}

	notOpenAPI := enhancedNot.ToOpenAPISchema()
	if notOpenAPI.Not == nil {
		t.Error("Expected Not schema to be present in OpenAPI")
	}
}

func TestComplexCompositionScenarios(t *testing.T) {
	// Test nested composition (OneOf inside AllOf)
	userBaseSchema := Object(map[string]interface{}{
		"name": String().Required(),
		"id":   String().Pattern("^usr_").Required(),
	}).Required()

	adminSchema := Object(map[string]interface{}{
		"permissions": Array(String()).Required(),
	}).Required()

	customerSchema := Object(map[string]interface{}{
		"orders": Array(String()).Required(),
	}).Required()

	// User must be base user AND either admin OR customer
	complexSchema := AllOf(
		userBaseSchema,
		OneOf(adminSchema, customerSchema),
	).Required()

	// Test valid admin user
	adminUser := map[string]interface{}{
		"name":        "Admin User",
		"id":          "usr_admin123",
		"permissions": []string{"read", "write"},
	}

	if err := complexSchema.Validate(adminUser); err != nil {
		t.Errorf("Expected admin user to be valid: %v", err)
	}

	// Test valid customer user
	customerUser := map[string]interface{}{
		"name":   "Customer User",
		"id":     "usr_customer456",
		"orders": []string{"ord_123", "ord_456"},
	}

	if err := complexSchema.Validate(customerUser); err != nil {
		t.Errorf("Expected customer user to be valid: %v", err)
	}

	// Test invalid user (missing role-specific fields)
	baseOnlyUser := map[string]interface{}{
		"name": "Base User",
		"id":   "usr_base789",
	}

	if err := complexSchema.Validate(baseOnlyUser); err == nil {
		t.Error("Expected base-only user to be invalid (missing role)")
	}

	// Test invalid user (has both admin and customer fields - violates OneOf)
	hybridUser := map[string]interface{}{
		"name":        "Hybrid User",
		"id":          "usr_hybrid999",
		"permissions": []string{"read"},
		"orders":      []string{"ord_789"},
	}

	if err := complexSchema.Validate(hybridUser); err == nil {
		t.Error("Expected hybrid user to be invalid (matches both admin and customer)")
	}
}

func TestCompositionErrorHandling(t *testing.T) {
	// Test composition with invalid schema types
	invalidSchema := "not a schema"
	oneOfWithInvalid := OneOf(String().Required(), invalidSchema).Required()

	if err := oneOfWithInvalid.Validate("test"); err == nil {
		t.Error("Expected OneOf with invalid schema to fail")
	}

	// Test Not with multiple schemas (should fail)
	invalidNotSchema := &compositionSchema{
		compositionType: CompositionTypeNot,
		schemas:         []interface{}{String().Required(), Number().Required()},
	}

	if err := invalidNotSchema.Validate("test"); err == nil {
		t.Error("Expected Not with multiple schemas to fail")
	}

	// Test unknown composition type
	unknownSchema := &compositionSchema{
		compositionType: CompositionType("unknown"),
		schemas:         []interface{}{String().Required()},
	}

	if err := unknownSchema.Validate("test"); err == nil {
		t.Error("Expected unknown composition type to fail")
	}
}

func TestPolymorphicUserTypes(t *testing.T) {
	// Real-world example: Different user types with distinct schemas (not using AllOf to avoid conflicts)
	basicUserSchema := Object(map[string]interface{}{
		"type":  String().Pattern("^basic$").Required(),
		"id":    String().Required(),
		"email": String().Email().Required(),
	}).Required()

	premiumUserSchema := Object(map[string]interface{}{
		"type":          String().Pattern("^premium$").Required(),
		"id":            String().Required(),
		"email":         String().Email().Required(),
		"subscription":  String().Required(),
		"premium_since": String().Required(),
	}).Required()

	adminUserSchema := Object(map[string]interface{}{
		"type":        String().Pattern("^admin$").Required(),
		"id":          String().Required(),
		"email":       String().Email().Required(),
		"permissions": Array(String()).Required(),
		"admin_level": Number().Min(1).Max(10).Required(),
	}).Required()

	// OneOf for polymorphic user types
	userSchema := OneOf(basicUserSchema, premiumUserSchema, adminUserSchema).Required()

	// Test basic user
	basicUser := map[string]interface{}{
		"type":  "basic",
		"id":    "usr_123",
		"email": "basic@example.com",
	}

	if err := userSchema.Validate(basicUser); err != nil {
		t.Errorf("Expected basic user to be valid: %v", err)
	}

	// Test premium user
	premiumUser := map[string]interface{}{
		"type":          "premium",
		"id":            "usr_456",
		"email":         "premium@example.com",
		"subscription":  "monthly",
		"premium_since": "2024-01-01",
	}

	if err := userSchema.Validate(premiumUser); err != nil {
		t.Errorf("Expected premium user to be valid: %v", err)
	}

	// Test admin user
	adminUser := map[string]interface{}{
		"type":        "admin",
		"id":          "usr_789",
		"email":       "admin@example.com",
		"permissions": []string{"read", "write", "delete"},
		"admin_level": 5,
	}

	if err := userSchema.Validate(adminUser); err != nil {
		t.Errorf("Expected admin user to be valid: %v", err)
	}

	// Test invalid user (wrong type)
	invalidUser := map[string]interface{}{
		"type":  "invalid",
		"id":    "usr_999",
		"email": "invalid@example.com",
	}

	if err := userSchema.Validate(invalidUser); err == nil {
		t.Error("Expected invalid user type to fail validation")
	}
}
