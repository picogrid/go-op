package goop

import (
	"fmt"
	"testing"
)

// TestSecurityRequirements tests the SecurityRequirements builder methods
func TestSecurityRequirements(t *testing.T) {
	t.Run("RequireScheme adds security requirement", func(t *testing.T) {
		req := SecurityRequirements{}.RequireScheme("apiKey", "read", "write")
		
		if len(req) != 1 {
			t.Errorf("Expected 1 requirement, got %d", len(req))
		}
		
		if len(req[0]["apiKey"]) != 2 {
			t.Errorf("Expected 2 scopes, got %d", len(req[0]["apiKey"]))
		}
		
		if req[0]["apiKey"][0] != "read" || req[0]["apiKey"][1] != "write" {
			t.Errorf("Expected scopes [read, write], got %v", req[0]["apiKey"])
		}
	})

	t.Run("RequireScheme with no scopes", func(t *testing.T) {
		req := SecurityRequirements{}.RequireScheme("apiKey")
		
		if len(req) != 1 {
			t.Errorf("Expected 1 requirement, got %d", len(req))
		}
		
		if len(req[0]["apiKey"]) != 0 {
			t.Errorf("Expected 0 scopes, got %d", len(req[0]["apiKey"]))
		}
	})

	t.Run("RequireScheme with empty scheme name", func(t *testing.T) {
		req := SecurityRequirements{}.RequireScheme("", "scope")
		
		if len(req) != 1 {
			t.Errorf("Expected 1 requirement, got %d", len(req))
		}
		
		// Empty scheme name should still create requirement
		if _, exists := req[0][""]; !exists {
			t.Error("Expected empty scheme name to be allowed")
		}
	})

	t.Run("RequireAny creates OR logic", func(t *testing.T) {
		scheme1 := SecurityRequirement{"apiKey": []string{}}
		scheme2 := SecurityRequirement{"oauth2": []string{"read"}}
		
		req := SecurityRequirements{}.RequireAny(scheme1, scheme2)
		
		if len(req) != 2 {
			t.Errorf("Expected 2 requirements (OR logic), got %d", len(req))
		}
		
		// First requirement should have apiKey
		if _, exists := req[0]["apiKey"]; !exists {
			t.Error("Expected first requirement to have apiKey")
		}
		
		// Second requirement should have oauth2
		if _, exists := req[1]["oauth2"]; !exists {
			t.Error("Expected second requirement to have oauth2")
		}
	})

	t.Run("RequireAny with empty slice", func(t *testing.T) {
		req := SecurityRequirements{}.RequireAny()
		
		if len(req) != 0 {
			t.Errorf("Expected 0 requirements for empty slice, got %d", len(req))
		}
	})

	t.Run("RequireAll creates AND logic", func(t *testing.T) {
		scheme1 := SecurityRequirement{"apiKey": []string{}}
		scheme2 := SecurityRequirement{"oauth2": []string{"read"}}
		
		req := SecurityRequirements{}.RequireAll(scheme1, scheme2)
		
		if len(req) != 1 {
			t.Errorf("Expected 1 requirement (AND logic), got %d", len(req))
		}
		
		// Single requirement should have both schemes
		if _, exists := req[0]["apiKey"]; !exists {
			t.Error("Expected requirement to have apiKey")
		}
		
		if _, exists := req[0]["oauth2"]; !exists {
			t.Error("Expected requirement to have oauth2")
		}
		
		if req[0]["oauth2"][0] != "read" {
			t.Errorf("Expected oauth2 scope 'read', got %v", req[0]["oauth2"])
		}
	})

	t.Run("RequireAll with empty slice", func(t *testing.T) {
		req := SecurityRequirements{}.RequireAll()
		
		if len(req) != 0 {
			t.Errorf("Expected 0 requirements for empty slice, got %d", len(req))
		}
	})

	t.Run("RequireAll with overlapping scheme names", func(t *testing.T) {
		scheme1 := SecurityRequirement{"oauth2": []string{"read"}}
		scheme2 := SecurityRequirement{"oauth2": []string{"write"}}
		
		req := SecurityRequirements{}.RequireAll(scheme1, scheme2)
		
		if len(req) != 1 {
			t.Errorf("Expected 1 requirement, got %d", len(req))
		}
		
		// Second scheme should overwrite first
		if len(req[0]["oauth2"]) != 1 || req[0]["oauth2"][0] != "write" {
			t.Errorf("Expected oauth2 scope ['write'], got %v", req[0]["oauth2"])
		}
	})

	t.Run("NoAuth creates empty requirement", func(t *testing.T) {
		req := NoAuth()
		
		if len(req) != 1 {
			t.Errorf("Expected 1 requirement, got %d", len(req))
		}
		
		if len(req[0]) != 0 {
			t.Errorf("Expected empty requirement, got %v", req[0])
		}
	})

	t.Run("Complex requirement chaining", func(t *testing.T) {
		req := SecurityRequirements{}.
			RequireScheme("apiKey").
			RequireScheme("oauth2", "read").
			RequireAny(SecurityRequirement{"bearer": []string{}})
		
		if len(req) != 3 {
			t.Errorf("Expected 3 requirements, got %d", len(req))
		}
	})
}

// TestAPIKeySecurityScheme tests API key security scheme validation and conversion
func TestAPIKeySecurityScheme(t *testing.T) {
	t.Run("GetType returns correct type", func(t *testing.T) {
		scheme := &APIKeySecurityScheme{}
		if scheme.GetType() != APIKeyScheme {
			t.Errorf("Expected type %s, got %s", APIKeyScheme, scheme.GetType())
		}
	})

	t.Run("Valid header API key scheme", func(t *testing.T) {
		scheme := &APIKeySecurityScheme{
			Name:        "X-API-Key",
			In:          HeaderLocation,
			Description: "API key in header",
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Valid scheme should not return error: %v", err)
		}
	})

	t.Run("Valid query API key scheme", func(t *testing.T) {
		scheme := &APIKeySecurityScheme{
			Name: "api_key",
			In:   QueryLocation,
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Valid scheme should not return error: %v", err)
		}
	})

	t.Run("Valid cookie API key scheme", func(t *testing.T) {
		scheme := &APIKeySecurityScheme{
			Name: "session_id",
			In:   CookieLocation,
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Valid scheme should not return error: %v", err)
		}
	})

	t.Run("Empty name fails validation", func(t *testing.T) {
		scheme := &APIKeySecurityScheme{
			Name: "",
			In:   HeaderLocation,
		}
		
		err := scheme.Validate()
		if err == nil {
			t.Error("Expected validation to fail for empty name")
		}
		
		expectedMsg := "apiKey security scheme requires 'name' field"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Invalid location fails validation", func(t *testing.T) {
		scheme := &APIKeySecurityScheme{
			Name: "api_key",
			In:   APIKeyLocation("invalid"),
		}
		
		err := scheme.Validate()
		if err == nil {
			t.Error("Expected validation to fail for invalid location")
		}
		
		if !contains(err.Error(), "apiKey 'in' field must be") {
			t.Errorf("Expected location validation error, got: %v", err)
		}
	})

	t.Run("ToOpenAPI conversion", func(t *testing.T) {
		scheme := &APIKeySecurityScheme{
			Name:        "X-API-Key",
			In:          HeaderLocation,
			Description: "API key authentication",
		}
		
		openapi := scheme.ToOpenAPI()
		
		if openapi.Type != string(APIKeyScheme) {
			t.Errorf("Expected type %s, got %s", APIKeyScheme, openapi.Type)
		}
		
		if openapi.Name != "X-API-Key" {
			t.Errorf("Expected name 'X-API-Key', got '%s'", openapi.Name)
		}
		
		if openapi.In != string(HeaderLocation) {
			t.Errorf("Expected in '%s', got '%s'", HeaderLocation, openapi.In)
		}
		
		if openapi.Description != "API key authentication" {
			t.Errorf("Expected description 'API key authentication', got '%s'", openapi.Description)
		}
	})
}

// TestHTTPSecurityScheme tests HTTP security scheme validation and conversion
func TestHTTPSecurityScheme(t *testing.T) {
	t.Run("GetType returns correct type", func(t *testing.T) {
		scheme := &HTTPSecurityScheme{}
		if scheme.GetType() != HTTPScheme {
			t.Errorf("Expected type %s, got %s", HTTPScheme, scheme.GetType())
		}
	})

	t.Run("Valid basic auth scheme", func(t *testing.T) {
		scheme := &HTTPSecurityScheme{
			Scheme:      "basic",
			Description: "Basic authentication",
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Valid scheme should not return error: %v", err)
		}
	})

	t.Run("Valid bearer auth scheme", func(t *testing.T) {
		scheme := &HTTPSecurityScheme{
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Bearer token authentication",
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Valid scheme should not return error: %v", err)
		}
	})

	t.Run("Valid custom scheme", func(t *testing.T) {
		scheme := &HTTPSecurityScheme{
			Scheme:      "custom",
			Description: "Custom authentication",
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Custom scheme should be allowed: %v", err)
		}
	})

	t.Run("Case insensitive scheme validation", func(t *testing.T) {
		schemes := []string{"BASIC", "Bearer", "DIGEST", "Negotiate", "OAUTH"}
		
		for _, scheme := range schemes {
			t.Run("scheme_"+scheme, func(t *testing.T) {
				httpScheme := &HTTPSecurityScheme{
					Scheme: scheme,
				}
				
				err := httpScheme.Validate()
				if err != nil {
					t.Errorf("Scheme '%s' should be valid: %v", scheme, err)
				}
			})
		}
	})

	t.Run("Empty scheme fails validation", func(t *testing.T) {
		scheme := &HTTPSecurityScheme{
			Scheme: "",
		}
		
		err := scheme.Validate()
		if err == nil {
			t.Error("Expected validation to fail for empty scheme")
		}
		
		expectedMsg := "http security scheme requires 'scheme' field"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("ToOpenAPI conversion", func(t *testing.T) {
		scheme := &HTTPSecurityScheme{
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Bearer token authentication",
		}
		
		openapi := scheme.ToOpenAPI()
		
		if openapi.Type != string(HTTPScheme) {
			t.Errorf("Expected type %s, got %s", HTTPScheme, openapi.Type)
		}
		
		if openapi.Scheme != "bearer" {
			t.Errorf("Expected scheme 'bearer', got '%s'", openapi.Scheme)
		}
		
		if openapi.BearerFormat != "JWT" {
			t.Errorf("Expected bearer format 'JWT', got '%s'", openapi.BearerFormat)
		}
		
		if openapi.Description != "Bearer token authentication" {
			t.Errorf("Expected description 'Bearer token authentication', got '%s'", openapi.Description)
		}
	})
}

// TestOAuth2Flow tests OAuth2 flow validation
func TestOAuth2Flow(t *testing.T) {
	t.Run("Implicit flow validation", func(t *testing.T) {
		flow := &OAuth2Flow{
			AuthorizationURL: "https://auth.example.com/oauth/authorize",
			Scopes:           map[string]string{"read": "Read access"},
		}
		
		err := flow.Validate(ImplicitFlow)
		if err != nil {
			t.Errorf("Valid implicit flow should not return error: %v", err)
		}
	})

	t.Run("Implicit flow missing authorization URL", func(t *testing.T) {
		flow := &OAuth2Flow{
			Scopes: map[string]string{"read": "Read access"},
		}
		
		err := flow.Validate(ImplicitFlow)
		if err == nil {
			t.Error("Expected validation to fail for missing authorization URL")
		}
		
		if !contains(err.Error(), "implicit flow requires 'authorizationUrl'") {
			t.Errorf("Expected authorization URL error, got: %v", err)
		}
	})

	t.Run("Password flow validation", func(t *testing.T) {
		flow := &OAuth2Flow{
			TokenURL: "https://auth.example.com/oauth/token",
			Scopes:   map[string]string{"read": "Read access"},
		}
		
		err := flow.Validate(PasswordFlow)
		if err != nil {
			t.Errorf("Valid password flow should not return error: %v", err)
		}
	})

	t.Run("Password flow missing token URL", func(t *testing.T) {
		flow := &OAuth2Flow{
			Scopes: map[string]string{"read": "Read access"},
		}
		
		err := flow.Validate(PasswordFlow)
		if err == nil {
			t.Error("Expected validation to fail for missing token URL")
		}
		
		if !contains(err.Error(), "password flow requires 'tokenUrl'") {
			t.Errorf("Expected token URL error, got: %v", err)
		}
	})

	t.Run("Client credentials flow validation", func(t *testing.T) {
		flow := &OAuth2Flow{
			TokenURL: "https://auth.example.com/oauth/token",
			Scopes:   map[string]string{"admin": "Admin access"},
		}
		
		err := flow.Validate(ClientCredentialsFlow)
		if err != nil {
			t.Errorf("Valid client credentials flow should not return error: %v", err)
		}
	})

	t.Run("Client credentials flow missing token URL", func(t *testing.T) {
		flow := &OAuth2Flow{
			Scopes: map[string]string{"admin": "Admin access"},
		}
		
		err := flow.Validate(ClientCredentialsFlow)
		if err == nil {
			t.Error("Expected validation to fail for missing token URL")
		}
		
		if !contains(err.Error(), "clientCredentials flow requires 'tokenUrl'") {
			t.Errorf("Expected token URL error, got: %v", err)
		}
	})

	t.Run("Authorization code flow validation", func(t *testing.T) {
		flow := &OAuth2Flow{
			AuthorizationURL: "https://auth.example.com/oauth/authorize",
			TokenURL:         "https://auth.example.com/oauth/token",
			RefreshURL:       "https://auth.example.com/oauth/refresh",
			Scopes:           map[string]string{"read": "Read access"},
		}
		
		err := flow.Validate(AuthorizationCodeFlow)
		if err != nil {
			t.Errorf("Valid authorization code flow should not return error: %v", err)
		}
	})

	t.Run("Authorization code flow missing authorization URL", func(t *testing.T) {
		flow := &OAuth2Flow{
			TokenURL: "https://auth.example.com/oauth/token",
			Scopes:   map[string]string{"read": "Read access"},
		}
		
		err := flow.Validate(AuthorizationCodeFlow)
		if err == nil {
			t.Error("Expected validation to fail for missing authorization URL")
		}
		
		if !contains(err.Error(), "authorizationCode flow requires 'authorizationUrl'") {
			t.Errorf("Expected authorization URL error, got: %v", err)
		}
	})

	t.Run("Authorization code flow missing token URL", func(t *testing.T) {
		flow := &OAuth2Flow{
			AuthorizationURL: "https://auth.example.com/oauth/authorize",
			Scopes:           map[string]string{"read": "Read access"},
		}
		
		err := flow.Validate(AuthorizationCodeFlow)
		if err == nil {
			t.Error("Expected validation to fail for missing token URL")
		}
		
		if !contains(err.Error(), "authorizationCode flow requires 'tokenUrl'") {
			t.Errorf("Expected token URL error, got: %v", err)
		}
	})

	t.Run("Nil scopes fails validation", func(t *testing.T) {
		flow := &OAuth2Flow{
			TokenURL: "https://auth.example.com/oauth/token",
			Scopes:   nil,
		}
		
		err := flow.Validate(PasswordFlow)
		if err == nil {
			t.Error("Expected validation to fail for nil scopes")
		}
		
		if !contains(err.Error(), "oauth2 flow requires 'scopes' field") {
			t.Errorf("Expected scopes error, got: %v", err)
		}
	})

	// Note: Go's url.Parse is very permissive, so we focus on testing valid URLs
	// rather than trying to find URLs that cause parse errors

	t.Run("Valid URLs pass validation", func(t *testing.T) {
		validURLs := []string{
			"https://example.com",
			"http://localhost:8080",
			"https://auth.example.com/oauth/token",
			"http://127.0.0.1:3000/auth",
		}
		
		for _, validURL := range validURLs {
			t.Run("url_"+validURL, func(t *testing.T) {
				flow := &OAuth2Flow{
					AuthorizationURL: validURL,
					TokenURL:         validURL,
					RefreshURL:       validURL,
					Scopes:           map[string]string{"read": "Read access"},
				}
				
				err := flow.Validate(AuthorizationCodeFlow)
				if err != nil {
					t.Errorf("Valid URL should pass validation: %s, error: %v", validURL, err)
				}
			})
		}
	})
}

// TestOAuth2Flows tests OAuth2 flows validation
func TestOAuth2Flows(t *testing.T) {
	t.Run("No flows fails validation", func(t *testing.T) {
		flows := &OAuth2Flows{}
		
		err := flows.Validate()
		if err == nil {
			t.Error("Expected validation to fail when no flows are defined")
		}
		
		if !contains(err.Error(), "oauth2 security scheme requires at least one flow") {
			t.Errorf("Expected 'at least one flow' error, got: %v", err)
		}
	})

	t.Run("Single valid flow passes validation", func(t *testing.T) {
		flows := &OAuth2Flows{
			Implicit: &OAuth2Flow{
				AuthorizationURL: "https://auth.example.com/oauth/authorize",
				Scopes:           map[string]string{"read": "Read access"},
			},
		}
		
		err := flows.Validate()
		if err != nil {
			t.Errorf("Single valid flow should pass: %v", err)
		}
	})

	t.Run("Multiple valid flows pass validation", func(t *testing.T) {
		flows := &OAuth2Flows{
			AuthorizationCode: &OAuth2Flow{
				AuthorizationURL: "https://auth.example.com/oauth/authorize",
				TokenURL:         "https://auth.example.com/oauth/token",
				Scopes:           map[string]string{"read": "Read access"},
			},
			ClientCredentials: &OAuth2Flow{
				TokenURL: "https://auth.example.com/oauth/token",
				Scopes:   map[string]string{"admin": "Admin access"},
			},
		}
		
		err := flows.Validate()
		if err != nil {
			t.Errorf("Multiple valid flows should pass: %v", err)
		}
	})

	t.Run("Invalid implicit flow fails validation", func(t *testing.T) {
		flows := &OAuth2Flows{
			Implicit: &OAuth2Flow{
				// Missing authorization URL
				Scopes: map[string]string{"read": "Read access"},
			},
		}
		
		err := flows.Validate()
		if err == nil {
			t.Error("Expected validation to fail for invalid implicit flow")
		}
		
		if !contains(err.Error(), "implicit flow validation failed") {
			t.Errorf("Expected implicit flow error, got: %v", err)
		}
	})

	t.Run("Invalid password flow fails validation", func(t *testing.T) {
		flows := &OAuth2Flows{
			Password: &OAuth2Flow{
				// Missing token URL
				Scopes: map[string]string{"read": "Read access"},
			},
		}
		
		err := flows.Validate()
		if err == nil {
			t.Error("Expected validation to fail for invalid password flow")
		}
		
		if !contains(err.Error(), "password flow validation failed") {
			t.Errorf("Expected password flow error, got: %v", err)
		}
	})

	t.Run("Invalid client credentials flow fails validation", func(t *testing.T) {
		flows := &OAuth2Flows{
			ClientCredentials: &OAuth2Flow{
				// Missing token URL
				Scopes: map[string]string{"admin": "Admin access"},
			},
		}
		
		err := flows.Validate()
		if err == nil {
			t.Error("Expected validation to fail for invalid client credentials flow")
		}
		
		if !contains(err.Error(), "clientCredentials flow validation failed") {
			t.Errorf("Expected clientCredentials flow error, got: %v", err)
		}
	})

	t.Run("Invalid authorization code flow fails validation", func(t *testing.T) {
		flows := &OAuth2Flows{
			AuthorizationCode: &OAuth2Flow{
				// Missing both URLs
				Scopes: map[string]string{"read": "Read access"},
			},
		}
		
		err := flows.Validate()
		if err == nil {
			t.Error("Expected validation to fail for invalid authorization code flow")
		}
		
		if !contains(err.Error(), "authorizationCode flow validation failed") {
			t.Errorf("Expected authorizationCode flow error, got: %v", err)
		}
	})
}

// TestOAuth2SecurityScheme tests OAuth2 security scheme
func TestOAuth2SecurityScheme(t *testing.T) {
	t.Run("GetType returns correct type", func(t *testing.T) {
		scheme := &OAuth2SecurityScheme{}
		if scheme.GetType() != OAuth2Scheme {
			t.Errorf("Expected type %s, got %s", OAuth2Scheme, scheme.GetType())
		}
	})

	t.Run("Valid OAuth2 scheme", func(t *testing.T) {
		scheme := &OAuth2SecurityScheme{
			Flows: OAuth2Flows{
				AuthorizationCode: &OAuth2Flow{
					AuthorizationURL: "https://auth.example.com/oauth/authorize",
					TokenURL:         "https://auth.example.com/oauth/token",
					Scopes:           map[string]string{"read": "Read access"},
				},
			},
			Description: "OAuth2 authentication",
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Valid OAuth2 scheme should not return error: %v", err)
		}
	})

	t.Run("Invalid flows fail validation", func(t *testing.T) {
		scheme := &OAuth2SecurityScheme{
			Flows: OAuth2Flows{
				// No flows defined
			},
			Description: "OAuth2 authentication",
		}
		
		err := scheme.Validate()
		if err == nil {
			t.Error("Expected validation to fail for invalid flows")
		}
	})

	t.Run("ToOpenAPI conversion", func(t *testing.T) {
		flows := OAuth2Flows{
			AuthorizationCode: &OAuth2Flow{
				AuthorizationURL: "https://auth.example.com/oauth/authorize",
				TokenURL:         "https://auth.example.com/oauth/token",
				Scopes:           map[string]string{"read": "Read access"},
			},
		}
		
		scheme := &OAuth2SecurityScheme{
			Flows:       flows,
			Description: "OAuth2 authentication",
		}
		
		openapi := scheme.ToOpenAPI()
		
		if openapi.Type != string(OAuth2Scheme) {
			t.Errorf("Expected type %s, got %s", OAuth2Scheme, openapi.Type)
		}
		
		if openapi.Flows == nil {
			t.Error("Expected flows to be present")
		}
		
		if openapi.Flows.AuthorizationCode == nil {
			t.Error("Expected authorization code flow to be present")
		}
		
		if openapi.Description != "OAuth2 authentication" {
			t.Errorf("Expected description 'OAuth2 authentication', got '%s'", openapi.Description)
		}
	})
}

// TestOpenIDConnectSecurityScheme tests OpenID Connect security scheme
func TestOpenIDConnectSecurityScheme(t *testing.T) {
	t.Run("GetType returns correct type", func(t *testing.T) {
		scheme := &OpenIDConnectSecurityScheme{}
		if scheme.GetType() != OpenIDConnectScheme {
			t.Errorf("Expected type %s, got %s", OpenIDConnectScheme, scheme.GetType())
		}
	})

	t.Run("Valid OpenID Connect scheme", func(t *testing.T) {
		scheme := &OpenIDConnectSecurityScheme{
			OpenIDConnectURL: "https://auth.example.com/.well-known/openid_configuration",
			Description:      "OpenID Connect authentication",
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Valid OpenID Connect scheme should not return error: %v", err)
		}
	})

	t.Run("Empty URL fails validation", func(t *testing.T) {
		scheme := &OpenIDConnectSecurityScheme{
			OpenIDConnectURL: "",
			Description:      "OpenID Connect authentication",
		}
		
		err := scheme.Validate()
		if err == nil {
			t.Error("Expected validation to fail for empty URL")
		}
		
		expectedMsg := "openIdConnect security scheme requires 'openIdConnectUrl' field"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	// Note: Go's url.Parse is very permissive, so we focus on empty URL validation
	// rather than trying to find URLs that cause parse errors

	t.Run("ToOpenAPI conversion", func(t *testing.T) {
		scheme := &OpenIDConnectSecurityScheme{
			OpenIDConnectURL: "https://auth.example.com/.well-known/openid_configuration",
			Description:      "OpenID Connect authentication",
		}
		
		openapi := scheme.ToOpenAPI()
		
		if openapi.Type != string(OpenIDConnectScheme) {
			t.Errorf("Expected type %s, got %s", OpenIDConnectScheme, openapi.Type)
		}
		
		if openapi.OpenIdConnectUrl != "https://auth.example.com/.well-known/openid_configuration" {
			t.Errorf("Expected URL 'https://auth.example.com/.well-known/openid_configuration', got '%s'", openapi.OpenIdConnectUrl)
		}
		
		if openapi.Description != "OpenID Connect authentication" {
			t.Errorf("Expected description 'OpenID Connect authentication', got '%s'", openapi.Description)
		}
	})
}

// TestMutualTLSSecurityScheme tests Mutual TLS security scheme
func TestMutualTLSSecurityScheme(t *testing.T) {
	t.Run("GetType returns correct type", func(t *testing.T) {
		scheme := &MutualTLSSecurityScheme{}
		if scheme.GetType() != MutualTLSScheme {
			t.Errorf("Expected type %s, got %s", MutualTLSScheme, scheme.GetType())
		}
	})

	t.Run("Validate always passes", func(t *testing.T) {
		scheme := &MutualTLSSecurityScheme{
			Description: "Mutual TLS authentication",
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Mutual TLS validation should always pass: %v", err)
		}
	})

	t.Run("Validate passes with empty description", func(t *testing.T) {
		scheme := &MutualTLSSecurityScheme{}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Mutual TLS validation should always pass even with empty description: %v", err)
		}
	})

	t.Run("ToOpenAPI conversion", func(t *testing.T) {
		scheme := &MutualTLSSecurityScheme{
			Description: "Mutual TLS authentication",
		}
		
		openapi := scheme.ToOpenAPI()
		
		if openapi.Type != string(MutualTLSScheme) {
			t.Errorf("Expected type %s, got %s", MutualTLSScheme, openapi.Type)
		}
		
		if openapi.Description != "Mutual TLS authentication" {
			t.Errorf("Expected description 'Mutual TLS authentication', got '%s'", openapi.Description)
		}
	})

	t.Run("ToOpenAPI conversion with empty description", func(t *testing.T) {
		scheme := &MutualTLSSecurityScheme{}
		
		openapi := scheme.ToOpenAPI()
		
		if openapi.Type != string(MutualTLSScheme) {
			t.Errorf("Expected type %s, got %s", MutualTLSScheme, openapi.Type)
		}
		
		if openapi.Description != "" {
			t.Errorf("Expected empty description, got '%s'", openapi.Description)
		}
	})
}

// TestValidateSecuritySchemeName tests security scheme name validation
func TestValidateSecuritySchemeName(t *testing.T) {
	t.Run("Valid names pass validation", func(t *testing.T) {
		validNames := []string{
			"apiKey",
			"api_key",
			"api-key",
			"api.key",
			"ApiKey123",
			"123",
			"a",
			"_",
			"-",
			".",
			"a1b2c3",
			"very.long-name_with123numbers",
		}
		
		for _, name := range validNames {
			t.Run("name_"+name, func(t *testing.T) {
				err := ValidateSecuritySchemeName(name)
				if err != nil {
					t.Errorf("Valid name '%s' should pass validation: %v", name, err)
				}
			})
		}
	})

	t.Run("Invalid names fail validation", func(t *testing.T) {
		invalidNames := []string{
			"",          // Empty
			" ",         // Space
			"api key",   // Space in middle
			"api@key",   // @ character
			"api#key",   // # character
			"api%key",   // % character
			"api key ",  // Trailing space
			" apikey",   // Leading space
			"api\tkey",  // Tab character
			"api\nkey",  // Newline character
			"api+key",   // + character
			"api=key",   // = character
			"api/key",   // / character
		}
		
		for _, name := range invalidNames {
			t.Run("invalid_name_"+fmt.Sprintf("%q", name), func(t *testing.T) {
				err := ValidateSecuritySchemeName(name)
				if err == nil {
					t.Errorf("Invalid name '%s' should fail validation", name)
				}
				
				if !contains(err.Error(), "must match pattern") {
					t.Errorf("Expected pattern validation error, got: %v", err)
				}
			})
		}
	})
}

// TestSecurityHelperFunctions tests helper functions for common security patterns
func TestSecurityHelperFunctions(t *testing.T) {
	t.Run("NewAPIKeyHeader creates valid scheme", func(t *testing.T) {
		scheme := NewAPIKeyHeader("X-API-Key", "API key in header")
		
		if scheme.Name != "X-API-Key" {
			t.Errorf("Expected name 'X-API-Key', got '%s'", scheme.Name)
		}
		
		if scheme.In != HeaderLocation {
			t.Errorf("Expected location '%s', got '%s'", HeaderLocation, scheme.In)
		}
		
		if scheme.Description != "API key in header" {
			t.Errorf("Expected description 'API key in header', got '%s'", scheme.Description)
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Generated scheme should be valid: %v", err)
		}
	})

	t.Run("NewAPIKeyQuery creates valid scheme", func(t *testing.T) {
		scheme := NewAPIKeyQuery("api_key", "API key in query")
		
		if scheme.Name != "api_key" {
			t.Errorf("Expected name 'api_key', got '%s'", scheme.Name)
		}
		
		if scheme.In != QueryLocation {
			t.Errorf("Expected location '%s', got '%s'", QueryLocation, scheme.In)
		}
		
		if scheme.Description != "API key in query" {
			t.Errorf("Expected description 'API key in query', got '%s'", scheme.Description)
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Generated scheme should be valid: %v", err)
		}
	})

	t.Run("NewBearerAuth creates valid scheme", func(t *testing.T) {
		scheme := NewBearerAuth("JWT", "Bearer token authentication")
		
		if scheme.Scheme != "bearer" {
			t.Errorf("Expected scheme 'bearer', got '%s'", scheme.Scheme)
		}
		
		if scheme.BearerFormat != "JWT" {
			t.Errorf("Expected bearer format 'JWT', got '%s'", scheme.BearerFormat)
		}
		
		if scheme.Description != "Bearer token authentication" {
			t.Errorf("Expected description 'Bearer token authentication', got '%s'", scheme.Description)
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Generated scheme should be valid: %v", err)
		}
	})

	t.Run("NewBasicAuth creates valid scheme", func(t *testing.T) {
		scheme := NewBasicAuth("Basic authentication")
		
		if scheme.Scheme != "basic" {
			t.Errorf("Expected scheme 'basic', got '%s'", scheme.Scheme)
		}
		
		if scheme.Description != "Basic authentication" {
			t.Errorf("Expected description 'Basic authentication', got '%s'", scheme.Description)
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Generated scheme should be valid: %v", err)
		}
	})

	t.Run("NewOAuth2AuthorizationCode creates valid scheme", func(t *testing.T) {
		scopes := map[string]string{
			"read":  "Read access",
			"write": "Write access",
		}
		
		scheme := NewOAuth2AuthorizationCode(
			"https://auth.example.com/oauth/authorize",
			"https://auth.example.com/oauth/token",
			"https://auth.example.com/oauth/refresh",
			scopes,
			"OAuth2 authorization code flow",
		)
		
		if scheme.Description != "OAuth2 authorization code flow" {
			t.Errorf("Expected description 'OAuth2 authorization code flow', got '%s'", scheme.Description)
		}
		
		if scheme.Flows.AuthorizationCode == nil {
			t.Error("Expected authorization code flow to be present")
		}
		
		flow := scheme.Flows.AuthorizationCode
		if flow.AuthorizationURL != "https://auth.example.com/oauth/authorize" {
			t.Errorf("Expected authorization URL 'https://auth.example.com/oauth/authorize', got '%s'", flow.AuthorizationURL)
		}
		
		if flow.TokenURL != "https://auth.example.com/oauth/token" {
			t.Errorf("Expected token URL 'https://auth.example.com/oauth/token', got '%s'", flow.TokenURL)
		}
		
		if flow.RefreshURL != "https://auth.example.com/oauth/refresh" {
			t.Errorf("Expected refresh URL 'https://auth.example.com/oauth/refresh', got '%s'", flow.RefreshURL)
		}
		
		if len(flow.Scopes) != 2 {
			t.Errorf("Expected 2 scopes, got %d", len(flow.Scopes))
		}
		
		if flow.Scopes["read"] != "Read access" {
			t.Errorf("Expected scope 'read' with description 'Read access', got '%s'", flow.Scopes["read"])
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Generated scheme should be valid: %v", err)
		}
	})

	t.Run("NewOAuth2ClientCredentials creates valid scheme", func(t *testing.T) {
		scopes := map[string]string{
			"admin": "Admin access",
		}
		
		scheme := NewOAuth2ClientCredentials(
			"https://auth.example.com/oauth/token",
			"https://auth.example.com/oauth/refresh",
			scopes,
			"OAuth2 client credentials flow",
		)
		
		if scheme.Description != "OAuth2 client credentials flow" {
			t.Errorf("Expected description 'OAuth2 client credentials flow', got '%s'", scheme.Description)
		}
		
		if scheme.Flows.ClientCredentials == nil {
			t.Error("Expected client credentials flow to be present")
		}
		
		flow := scheme.Flows.ClientCredentials
		if flow.TokenURL != "https://auth.example.com/oauth/token" {
			t.Errorf("Expected token URL 'https://auth.example.com/oauth/token', got '%s'", flow.TokenURL)
		}
		
		if flow.RefreshURL != "https://auth.example.com/oauth/refresh" {
			t.Errorf("Expected refresh URL 'https://auth.example.com/oauth/refresh', got '%s'", flow.RefreshURL)
		}
		
		if len(flow.Scopes) != 1 {
			t.Errorf("Expected 1 scope, got %d", len(flow.Scopes))
		}
		
		if flow.Scopes["admin"] != "Admin access" {
			t.Errorf("Expected scope 'admin' with description 'Admin access', got '%s'", flow.Scopes["admin"])
		}
		
		err := scheme.Validate()
		if err != nil {
			t.Errorf("Generated scheme should be valid: %v", err)
		}
	})

	t.Run("Helper functions with empty parameters", func(t *testing.T) {
		t.Run("Empty name and description", func(t *testing.T) {
			scheme := NewAPIKeyHeader("", "")
			if scheme.Name != "" || scheme.Description != "" {
				t.Error("Empty parameters should be preserved")
			}
		})

		t.Run("Empty bearer format", func(t *testing.T) {
			scheme := NewBearerAuth("", "test")
			if scheme.BearerFormat != "" {
				t.Error("Empty bearer format should be preserved")
			}
		})

		t.Run("Empty URLs and nil scopes", func(t *testing.T) {
			scheme := NewOAuth2AuthorizationCode("", "", "", nil, "")
			if scheme.Flows.AuthorizationCode.Scopes != nil {
				t.Error("Nil scopes should be preserved")
			}
		})
	})
}

// TestSecuritySchemePolymorphism tests polymorphic behavior through SecurityScheme interface
func TestSecuritySchemePolymorphism(t *testing.T) {
	schemes := []struct {
		name   string
		scheme SecurityScheme
	}{
		{
			"APIKey",
			&APIKeySecurityScheme{
				Name: "X-API-Key",
				In:   HeaderLocation,
			},
		},
		{
			"HTTP",
			&HTTPSecurityScheme{
				Scheme: "bearer",
			},
		},
		{
			"OAuth2",
			&OAuth2SecurityScheme{
				Flows: OAuth2Flows{
					AuthorizationCode: &OAuth2Flow{
						AuthorizationURL: "https://auth.example.com/oauth/authorize",
						TokenURL:         "https://auth.example.com/oauth/token",
						Scopes:           map[string]string{"read": "Read access"},
					},
				},
			},
		},
		{
			"OpenIDConnect",
			&OpenIDConnectSecurityScheme{
				OpenIDConnectURL: "https://auth.example.com/.well-known/openid_configuration",
			},
		},
		{
			"MutualTLS",
			&MutualTLSSecurityScheme{},
		},
	}

	for _, test := range schemes {
		t.Run(test.name+"_interface_compliance", func(t *testing.T) {
			// Test GetType method
			schemeType := test.scheme.GetType()
			if schemeType == "" {
				t.Error("GetType should return non-empty scheme type")
			}

			// Test Validate method
			err := test.scheme.Validate()
			if err != nil {
				t.Errorf("Valid scheme should pass validation: %v", err)
			}

			// Test ToOpenAPI method
			openapi := test.scheme.ToOpenAPI()
			if openapi.Type == "" {
				t.Error("ToOpenAPI should return object with non-empty Type")
			}
			
			if openapi.Type != string(schemeType) {
				t.Errorf("OpenAPI type '%s' should match scheme type '%s'", openapi.Type, schemeType)
			}
		})
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && 
		 (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		  func() bool {
		  	for i := 1; i <= len(s)-len(substr); i++ {
		  		if s[i:i+len(substr)] == substr {
		  			return true
		  		}
		  	}
		  	return false
		  }())))
}