package operations

import (
	"encoding/json"
	"testing"

	goop "github.com/picogrid/go-op"
)

func TestOpenAPI31ComprehensiveFeatures(t *testing.T) {
	t.Run("OpenAPI License Validation", func(t *testing.T) {
		// Test valid license with name only
		license := &OpenAPILicense{Name: "MIT"}
		if err := license.Validate(); err != nil {
			t.Errorf("Expected valid license, got error: %v", err)
		}

		// Test valid license with name and identifier
		license = &OpenAPILicense{Name: "MIT", Identifier: "MIT"}
		if err := license.Validate(); err != nil {
			t.Errorf("Expected valid license with identifier, got error: %v", err)
		}

		// Test valid license with name and URL
		license = &OpenAPILicense{Name: "MIT", URL: "https://opensource.org/licenses/MIT"}
		if err := license.Validate(); err != nil {
			t.Errorf("Expected valid license with URL, got error: %v", err)
		}

		// Test invalid license - missing name
		license = &OpenAPILicense{Identifier: "MIT"}
		if err := license.Validate(); err == nil {
			t.Error("Expected error for license without name")
		}

		// Test invalid license - both identifier and URL
		license = &OpenAPILicense{
			Name:       "MIT",
			Identifier: "MIT",
			URL:        "https://opensource.org/licenses/MIT",
		}
		if err := license.Validate(); err == nil {
			t.Error("Expected error for license with both identifier and URL")
		}
	})

	t.Run("Component Key Validation", func(t *testing.T) {
		validKeys := []string{
			"User",
			"User_1",
			"User_Name",
			"user-name",
			"my.org.User",
			"Component123",
			"API_Key_Schema",
		}

		for _, key := range validKeys {
			if err := ValidateComponentKey(key); err != nil {
				t.Errorf("Expected valid component key '%s', got error: %v", key, err)
			}
		}

		invalidKeys := []string{
			"User@Name",
			"User Name",
			"User#Schema",
			"",
			"User$",
			"User!",
			"User/Path",
		}

		for _, key := range invalidKeys {
			if err := ValidateComponentKey(key); err == nil {
				t.Errorf("Expected invalid component key '%s' to fail validation", key)
			}
		}
	})

	t.Run("OpenAPI Media Type Validation", func(t *testing.T) {
		// Test valid media type with example only
		mediaType := &OpenAPIMediaType{
			Example: "test value",
		}
		if err := mediaType.Validate(); err != nil {
			t.Errorf("Expected valid media type with example, got error: %v", err)
		}

		// Test valid media type with examples only
		mediaType = &OpenAPIMediaType{
			Examples: map[string]OpenAPIExample{
				"test": {Value: "test value"},
			},
		}
		if err := mediaType.Validate(); err != nil {
			t.Errorf("Expected valid media type with examples, got error: %v", err)
		}

		// Test invalid media type with both example and examples
		mediaType = &OpenAPIMediaType{
			Example: "test value",
			Examples: map[string]OpenAPIExample{
				"test": {Value: "test value"},
			},
		}
		if err := mediaType.Validate(); err == nil {
			t.Error("Expected error for media type with both example and examples")
		}
	})

	t.Run("Complete OpenAPI Spec Structure", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		// Set all OpenAPI 3.1 fields
		generator.SetDescription("A comprehensive test API")
		generator.SetSummary("Test API Summary")
		generator.SetTermsOfService("https://example.com/terms")

		contact := &OpenAPIContact{
			Name:  "API Support",
			Email: "support@example.com",
			URL:   "https://example.com/support",
		}
		generator.SetContact(contact)

		license := &OpenAPILicense{
			Name:       "MIT",
			Identifier: "MIT",
		}
		if err := generator.SetLicense(license); err != nil {
			t.Errorf("Failed to set license: %v", err)
		}

		// Add server with variables
		server := OpenAPIServer{
			URL:         "https://{environment}.example.com/{version}",
			Description: "Test server with variables",
			Variables: map[string]OpenAPIServerVariable{
				"environment": {
					Default:     "api",
					Enum:        []string{"api", "staging"},
					Description: "Environment name",
				},
				"version": {
					Default:     "v1",
					Description: "API version",
				},
			},
		}
		generator.AddServer(server)

		// Add tag with external docs
		tag := OpenAPITag{
			Name:        "test",
			Description: "Test operations",
			ExternalDocs: &OpenAPIExternalDocs{
				Description: "Test documentation",
				URL:         "https://example.com/docs",
			},
		}
		generator.AddTag(tag)

		// Set external docs
		generator.SetExternalDocs(&OpenAPIExternalDocs{
			Description: "Complete API documentation",
			URL:         "https://example.com/docs",
		})

		// Set JSON Schema dialect
		generator.SetJsonSchemaDialect("https://json-schema.org/draft/2020-12/schema")

		// Add webhook
		webhook := OpenAPIWebhook{
			Operations: map[string]OpenAPIOperation{
				"post": {
					Summary:     "Webhook notification",
					Description: "Receives webhook notifications",
				},
			},
		}
		generator.AddWebhook("notification", webhook)

		// Verify spec structure
		spec := generator.GetSpec()

		if spec.Info.Summary != "Test API Summary" {
			t.Errorf("Expected summary 'Test API Summary', got '%s'", spec.Info.Summary)
		}

		if spec.Info.TermsOfService != "https://example.com/terms" {
			t.Errorf("Expected terms of service, got '%s'", spec.Info.TermsOfService)
		}

		if spec.Info.Contact == nil || spec.Info.Contact.Name != "API Support" {
			t.Error("Expected contact information")
		}

		if spec.Info.License == nil || spec.Info.License.Name != "MIT" {
			t.Error("Expected license information")
		}

		if len(spec.Servers) != 1 {
			t.Errorf("Expected 1 server, got %d", len(spec.Servers))
		}

		if spec.Servers[0].Variables == nil || len(spec.Servers[0].Variables) != 2 {
			t.Error("Expected server variables")
		}

		if len(spec.Tags) != 1 {
			t.Errorf("Expected 1 tag, got %d", len(spec.Tags))
		}

		if spec.ExternalDocs == nil || spec.ExternalDocs.URL != "https://example.com/docs" {
			t.Error("Expected external docs")
		}

		if spec.JsonSchemaDialect != "https://json-schema.org/draft/2020-12/schema" {
			t.Error("Expected JSON Schema dialect")
		}

		if len(spec.Webhooks) != 1 {
			t.Errorf("Expected 1 webhook, got %d", len(spec.Webhooks))
		}
	})

	t.Run("Components Object Complete Structure", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")
		spec := generator.GetSpec()

		// Verify all component types are initialized
		if spec.Components.Schemas == nil {
			t.Error("Expected schemas map to be initialized")
		}
		if spec.Components.SecuritySchemes == nil {
			t.Error("Expected securitySchemes map to be initialized")
		}
		if spec.Components.Responses == nil {
			t.Error("Expected responses map to be initialized")
		}
		if spec.Components.Parameters == nil {
			t.Error("Expected parameters map to be initialized")
		}
		if spec.Components.Examples == nil {
			t.Error("Expected examples map to be initialized")
		}
		if spec.Components.RequestBodies == nil {
			t.Error("Expected requestBodies map to be initialized")
		}
		if spec.Components.Headers == nil {
			t.Error("Expected headers map to be initialized")
		}
		if spec.Components.Links == nil {
			t.Error("Expected links map to be initialized")
		}
		if spec.Components.Callbacks == nil {
			t.Error("Expected callbacks map to be initialized")
		}
		if spec.Components.PathItems == nil {
			t.Error("Expected pathItems map to be initialized")
		}
	})

	t.Run("Security Scheme Integration", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")

		// Add API key security scheme
		apiKeyScheme := goop.NewAPIKeyHeader("X-API-Key", "API key authentication")
		err := generator.AddSecurityScheme("apiKey", apiKeyScheme)
		if err != nil {
			t.Errorf("Failed to add API key security scheme: %v", err)
		}

		// Add Bearer authentication
		bearerScheme := goop.NewBearerAuth("JWT", "JWT token authentication")
		err = generator.AddSecurityScheme("bearerAuth", bearerScheme)
		if err != nil {
			t.Errorf("Failed to add bearer auth security scheme: %v", err)
		}

		// Add OAuth2 authentication
		oauth2Scheme := goop.NewOAuth2AuthorizationCode(
			"https://example.com/oauth/authorize",
			"https://example.com/oauth/token",
			"https://example.com/oauth/refresh",
			map[string]string{
				"read":  "Read access",
				"write": "Write access",
			},
			"OAuth2 authentication",
		)
		err = generator.AddSecurityScheme("oauth2", oauth2Scheme)
		if err != nil {
			t.Errorf("Failed to add OAuth2 security scheme: %v", err)
		}

		// Set global security requirements
		globalSecurity := goop.SecurityRequirements{}.
			RequireScheme("bearerAuth").
			RequireAny(goop.SecurityRequirement{"oauth2": {"read", "write"}})
		generator.SetGlobalSecurity(globalSecurity)

		// Verify security schemes
		schemes := generator.ListSecuritySchemes()
		if len(schemes) != 3 {
			t.Errorf("Expected 3 security schemes, got %d", len(schemes))
		}

		spec := generator.GetSpec()
		if len(spec.Components.SecuritySchemes) != 3 {
			t.Errorf("Expected 3 security schemes in spec, got %d", len(spec.Components.SecuritySchemes))
		}

		if len(spec.Security) != 2 {
			t.Errorf("Expected 2 global security requirements, got %d", len(spec.Security))
		}
	})

	t.Run("JSON Serialization Compatibility", func(t *testing.T) {
		generator := NewOpenAPIGenerator("Test API", "1.0.0")
		generator.SetDescription("Test description")
		generator.SetSummary("Test summary")

		// Add contact and license
		contact := &OpenAPIContact{
			Name:  "Support",
			Email: "support@example.com",
		}
		generator.SetContact(contact)

		license := &OpenAPILicense{
			Name: "MIT",
			URL:  "https://opensource.org/licenses/MIT",
		}
		generator.SetLicense(license)

		spec := generator.GetSpec()

		// Test JSON serialization
		jsonData, err := json.Marshal(spec)
		if err != nil {
			t.Errorf("Failed to marshal OpenAPI spec to JSON: %v", err)
		}

		// Test JSON deserialization
		var deserializedSpec OpenAPISpec
		err = json.Unmarshal(jsonData, &deserializedSpec)
		if err != nil {
			t.Errorf("Failed to unmarshal OpenAPI spec from JSON: %v", err)
		}

		// Verify key fields are preserved
		if deserializedSpec.Info.Title != "Test API" {
			t.Error("Title not preserved in JSON round-trip")
		}

		if deserializedSpec.Info.Summary != "Test summary" {
			t.Error("Summary not preserved in JSON round-trip")
		}

		if deserializedSpec.Info.Contact == nil || deserializedSpec.Info.Contact.Email != "support@example.com" {
			t.Error("Contact not preserved in JSON round-trip")
		}

		if deserializedSpec.Info.License == nil || deserializedSpec.Info.License.URL != "https://opensource.org/licenses/MIT" {
			t.Error("License not preserved in JSON round-trip")
		}
	})
}
