package goop

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// SecuritySchemeType represents the type of security scheme as defined in OpenAPI 3.1
type SecuritySchemeType string

const (
	// APIKeyScheme represents API key authentication
	APIKeyScheme SecuritySchemeType = "apiKey"
	// HTTPScheme represents HTTP authentication schemes (Basic, Bearer, etc.)
	HTTPScheme SecuritySchemeType = "http"
	// OAuth2Scheme represents OAuth 2.0 authentication
	OAuth2Scheme SecuritySchemeType = "oauth2"
	// OpenIDConnectScheme represents OpenID Connect Discovery
	OpenIDConnectScheme SecuritySchemeType = "openIdConnect"
	// MutualTLSScheme represents mutual TLS authentication
	MutualTLSScheme SecuritySchemeType = "mutualTLS"
)

// APIKeyLocation represents where an API key can be located
type APIKeyLocation string

const (
	// HeaderLocation indicates API key is in a header
	HeaderLocation APIKeyLocation = "header"
	// QueryLocation indicates API key is in a query parameter
	QueryLocation APIKeyLocation = "query"
	// CookieLocation indicates API key is in a cookie
	CookieLocation APIKeyLocation = "cookie"
)

// OAuth2FlowType represents the OAuth 2.0 flow types
type OAuth2FlowType string

const (
	// AuthorizationCodeFlow represents the authorization code flow
	AuthorizationCodeFlow OAuth2FlowType = "authorizationCode"
	// ImplicitFlow represents the implicit flow (deprecated but supported)
	ImplicitFlow OAuth2FlowType = "implicit"
	// PasswordFlow represents the resource owner password flow (not recommended)
	PasswordFlow OAuth2FlowType = "password"
	// ClientCredentialsFlow represents the client credentials flow
	ClientCredentialsFlow OAuth2FlowType = "clientCredentials"
)

// SecurityScheme represents a security scheme that can be applied to operations
type SecurityScheme interface {
	// GetType returns the security scheme type
	GetType() SecuritySchemeType
	// Validate validates the security scheme configuration
	Validate() error
	// ToOpenAPI converts the security scheme to OpenAPI format
	ToOpenAPI() SecuritySchemeObject
}

// SecurityRequirement represents a security requirement for an operation
// Key is the security scheme name, value is array of scopes (for OAuth2/OpenID) or roles
type SecurityRequirement map[string][]string

// SecurityRequirements represents multiple security requirements with OR logic between them
type SecurityRequirements []SecurityRequirement

// RequireScheme adds a security requirement for a specific scheme with optional scopes
func (sr SecurityRequirements) RequireScheme(schemeName string, scopes ...string) SecurityRequirements {
	requirement := SecurityRequirement{schemeName: scopes}
	return append(sr, requirement)
}

// RequireAny adds security requirements where any one can satisfy the authentication (OR logic)
func (sr SecurityRequirements) RequireAny(schemes ...SecurityRequirement) SecurityRequirements {
	return append(sr, schemes...)
}

// RequireAll adds a security requirement where all schemes must be satisfied (AND logic)
func (sr SecurityRequirements) RequireAll(schemes ...SecurityRequirement) SecurityRequirements {
	if len(schemes) == 0 {
		return sr
	}
	
	// Merge all schemes into a single requirement (AND logic)
	merged := SecurityRequirement{}
	for _, scheme := range schemes {
		for name, scopes := range scheme {
			merged[name] = scopes
		}
	}
	
	return append(sr, merged)
}

// NoAuth creates an empty security requirement (removes all authentication)
func NoAuth() SecurityRequirements {
	return SecurityRequirements{SecurityRequirement{}}
}

// APIKeySecurityScheme represents API key authentication
type APIKeySecurityScheme struct {
	Name        string         `json:"name" yaml:"name"`
	In          APIKeyLocation `json:"in" yaml:"in"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
}

// GetType returns the security scheme type
func (a *APIKeySecurityScheme) GetType() SecuritySchemeType {
	return APIKeyScheme
}

// Validate validates the API key security scheme
func (a *APIKeySecurityScheme) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("apiKey security scheme requires 'name' field")
	}
	
	switch a.In {
	case HeaderLocation, QueryLocation, CookieLocation:
		// Valid locations
	default:
		return fmt.Errorf("apiKey 'in' field must be 'header', 'query', or 'cookie', got: %s", a.In)
	}
	
	return nil
}

// ToOpenAPI converts to OpenAPI format
func (a *APIKeySecurityScheme) ToOpenAPI() SecuritySchemeObject {
	return SecuritySchemeObject{
		Type:        string(APIKeyScheme),
		Name:        a.Name,
		In:          string(a.In),
		Description: a.Description,
	}
}

// HTTPSecurityScheme represents HTTP authentication schemes
type HTTPSecurityScheme struct {
	Scheme       string `json:"scheme" yaml:"scheme"`
	BearerFormat string `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
}

// GetType returns the security scheme type
func (h *HTTPSecurityScheme) GetType() SecuritySchemeType {
	return HTTPScheme
}

// Validate validates the HTTP security scheme
func (h *HTTPSecurityScheme) Validate() error {
	if h.Scheme == "" {
		return fmt.Errorf("http security scheme requires 'scheme' field")
	}
	
	// Common HTTP authentication schemes
	validSchemes := map[string]bool{
		"basic":    true,
		"bearer":   true,
		"digest":   true,
		"negotiate": true,
		"oauth":    true,
	}
	
	scheme := strings.ToLower(h.Scheme)
	if !validSchemes[scheme] {
		// Allow custom schemes but warn about common ones
		return nil
	}
	
	return nil
}

// ToOpenAPI converts to OpenAPI format
func (h *HTTPSecurityScheme) ToOpenAPI() SecuritySchemeObject {
	return SecuritySchemeObject{
		Type:         string(HTTPScheme),
		Scheme:       h.Scheme,
		BearerFormat: h.BearerFormat,
		Description:  h.Description,
	}
}

// OAuth2Flow represents a single OAuth 2.0 flow
type OAuth2Flow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes" yaml:"scopes"`
}

// Validate validates an OAuth2 flow based on its type
func (f *OAuth2Flow) Validate(flowType OAuth2FlowType) error {
	if f.Scopes == nil {
		return fmt.Errorf("oauth2 flow requires 'scopes' field")
	}
	
	switch flowType {
	case ImplicitFlow:
		if f.AuthorizationURL == "" {
			return fmt.Errorf("implicit flow requires 'authorizationUrl'")
		}
	case PasswordFlow:
		if f.TokenURL == "" {
			return fmt.Errorf("password flow requires 'tokenUrl'")
		}
	case ClientCredentialsFlow:
		if f.TokenURL == "" {
			return fmt.Errorf("clientCredentials flow requires 'tokenUrl'")
		}
	case AuthorizationCodeFlow:
		if f.AuthorizationURL == "" {
			return fmt.Errorf("authorizationCode flow requires 'authorizationUrl'")
		}
		if f.TokenURL == "" {
			return fmt.Errorf("authorizationCode flow requires 'tokenUrl'")
		}
	}
	
	// Validate URLs if provided
	urls := []string{f.AuthorizationURL, f.TokenURL, f.RefreshURL}
	for _, urlStr := range urls {
		if urlStr != "" {
			if _, err := url.Parse(urlStr); err != nil {
				return fmt.Errorf("invalid URL '%s': %v", urlStr, err)
			}
		}
	}
	
	return nil
}

// OAuth2Flows represents all OAuth 2.0 flows for a security scheme
type OAuth2Flows struct {
	Implicit          *OAuth2Flow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuth2Flow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuth2Flow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuth2Flow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// Validate validates all OAuth2 flows
func (f *OAuth2Flows) Validate() error {
	flowCount := 0
	
	if f.Implicit != nil {
		if err := f.Implicit.Validate(ImplicitFlow); err != nil {
			return fmt.Errorf("implicit flow validation failed: %v", err)
		}
		flowCount++
	}
	
	if f.Password != nil {
		if err := f.Password.Validate(PasswordFlow); err != nil {
			return fmt.Errorf("password flow validation failed: %v", err)
		}
		flowCount++
	}
	
	if f.ClientCredentials != nil {
		if err := f.ClientCredentials.Validate(ClientCredentialsFlow); err != nil {
			return fmt.Errorf("clientCredentials flow validation failed: %v", err)
		}
		flowCount++
	}
	
	if f.AuthorizationCode != nil {
		if err := f.AuthorizationCode.Validate(AuthorizationCodeFlow); err != nil {
			return fmt.Errorf("authorizationCode flow validation failed: %v", err)
		}
		flowCount++
	}
	
	if flowCount == 0 {
		return fmt.Errorf("oauth2 security scheme requires at least one flow")
	}
	
	return nil
}

// OAuth2SecurityScheme represents OAuth 2.0 authentication
type OAuth2SecurityScheme struct {
	Flows       OAuth2Flows `json:"flows" yaml:"flows"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
}

// GetType returns the security scheme type
func (o *OAuth2SecurityScheme) GetType() SecuritySchemeType {
	return OAuth2Scheme
}

// Validate validates the OAuth2 security scheme
func (o *OAuth2SecurityScheme) Validate() error {
	return o.Flows.Validate()
}

// ToOpenAPI converts to OpenAPI format
func (o *OAuth2SecurityScheme) ToOpenAPI() SecuritySchemeObject {
	return SecuritySchemeObject{
		Type:        string(OAuth2Scheme),
		Flows:       &o.Flows,
		Description: o.Description,
	}
}

// OpenIDConnectSecurityScheme represents OpenID Connect Discovery
type OpenIDConnectSecurityScheme struct {
	OpenIDConnectURL string `json:"openIdConnectUrl" yaml:"openIdConnectUrl"`
	Description      string `json:"description,omitempty" yaml:"description,omitempty"`
}

// GetType returns the security scheme type
func (o *OpenIDConnectSecurityScheme) GetType() SecuritySchemeType {
	return OpenIDConnectScheme
}

// Validate validates the OpenID Connect security scheme
func (o *OpenIDConnectSecurityScheme) Validate() error {
	if o.OpenIDConnectURL == "" {
		return fmt.Errorf("openIdConnect security scheme requires 'openIdConnectUrl' field")
	}
	
	if _, err := url.Parse(o.OpenIDConnectURL); err != nil {
		return fmt.Errorf("invalid openIdConnectUrl '%s': %v", o.OpenIDConnectURL, err)
	}
	
	return nil
}

// ToOpenAPI converts to OpenAPI format
func (o *OpenIDConnectSecurityScheme) ToOpenAPI() SecuritySchemeObject {
	return SecuritySchemeObject{
		Type:             string(OpenIDConnectScheme),
		OpenIdConnectUrl: o.OpenIDConnectURL,
		Description:      o.Description,
	}
}

// MutualTLSSecurityScheme represents mutual TLS authentication
type MutualTLSSecurityScheme struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// GetType returns the security scheme type
func (m *MutualTLSSecurityScheme) GetType() SecuritySchemeType {
	return MutualTLSScheme
}

// Validate validates the mutual TLS security scheme
func (m *MutualTLSSecurityScheme) Validate() error {
	// No additional validation required for mutualTLS
	return nil
}

// ToOpenAPI converts to OpenAPI format
func (m *MutualTLSSecurityScheme) ToOpenAPI() SecuritySchemeObject {
	return SecuritySchemeObject{
		Type:        string(MutualTLSScheme),
		Description: m.Description,
	}
}

// SecuritySchemeObject represents the OpenAPI 3.1 Security Scheme Object
type SecuritySchemeObject struct {
	Type             string       `json:"type" yaml:"type"`
	Description      string       `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string       `json:"name,omitempty" yaml:"name,omitempty"`
	In               string       `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme           string       `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string       `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flows            *OAuth2Flows `json:"flows,omitempty" yaml:"flows,omitempty"`
	OpenIdConnectUrl string       `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
}

// ValidateSecuritySchemeName validates that a security scheme name follows OpenAPI 3.1 rules
func ValidateSecuritySchemeName(name string) error {
	// Component names must match the regex: ^[a-zA-Z0-9\.\-_]+$
	matched, err := regexp.MatchString(`^[a-zA-Z0-9\.\-_]+$`, name)
	if err != nil {
		return fmt.Errorf("regex error: %v", err)
	}
	if !matched {
		return fmt.Errorf("security scheme name '%s' must match pattern ^[a-zA-Z0-9\\.\\-_]+$", name)
	}
	return nil
}

// Security helper functions for common patterns

// NewAPIKeyHeader creates a new API key security scheme for headers
func NewAPIKeyHeader(name, description string) *APIKeySecurityScheme {
	return &APIKeySecurityScheme{
		Name:        name,
		In:          HeaderLocation,
		Description: description,
	}
}

// NewAPIKeyQuery creates a new API key security scheme for query parameters
func NewAPIKeyQuery(name, description string) *APIKeySecurityScheme {
	return &APIKeySecurityScheme{
		Name:        name,
		In:          QueryLocation,
		Description: description,
	}
}

// NewBearerAuth creates a new Bearer token HTTP security scheme
func NewBearerAuth(format, description string) *HTTPSecurityScheme {
	return &HTTPSecurityScheme{
		Scheme:       "bearer",
		BearerFormat: format,
		Description:  description,
	}
}

// NewBasicAuth creates a new Basic HTTP security scheme
func NewBasicAuth(description string) *HTTPSecurityScheme {
	return &HTTPSecurityScheme{
		Scheme:      "basic",
		Description: description,
	}
}

// NewOAuth2AuthorizationCode creates OAuth2 security scheme with authorization code flow
func NewOAuth2AuthorizationCode(authURL, tokenURL, refreshURL string, scopes map[string]string, description string) *OAuth2SecurityScheme {
	return &OAuth2SecurityScheme{
		Flows: OAuth2Flows{
			AuthorizationCode: &OAuth2Flow{
				AuthorizationURL: authURL,
				TokenURL:         tokenURL,
				RefreshURL:       refreshURL,
				Scopes:           scopes,
			},
		},
		Description: description,
	}
}

// NewOAuth2ClientCredentials creates OAuth2 security scheme with client credentials flow
func NewOAuth2ClientCredentials(tokenURL, refreshURL string, scopes map[string]string, description string) *OAuth2SecurityScheme {
	return &OAuth2SecurityScheme{
		Flows: OAuth2Flows{
			ClientCredentials: &OAuth2Flow{
				TokenURL:   tokenURL,
				RefreshURL: refreshURL,
				Scopes:     scopes,
			},
		},
		Description: description,
	}
}