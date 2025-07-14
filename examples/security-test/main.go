package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	goop "github.com/picogrid/go-op"
	"github.com/picogrid/go-op/operations"
	"github.com/picogrid/go-op/validators"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type GetUserParams struct {
	ID string `json:"id" uri:"id"`
}

type CustomHeaders struct {
	RequestID string `json:"X-Request-ID" header:"X-Request-ID"`
	Version   string `json:"X-API-Version" header:"X-API-Version"`
}

func getUserHandler(ctx context.Context, params GetUserParams, query struct{}, body struct{}) (User, error) {
	return User{
		ID:       params.ID,
		Username: "testuser",
		Email:    "test@example.com",
	}, nil
}

func adminHandler(ctx context.Context, params struct{}, query struct{}, body struct{}) (map[string]string, error) {
	return map[string]string{
		"message": "Admin access granted",
		"level":   "administrator",
	}, nil
}

func main() {
	engine := gin.Default()
	
	// Create OpenAPI generator
	openAPIGen := operations.NewOpenAPIGenerator("Security Test API", "1.0.0")
	
	// Add security schemes
	apiKeyAuth := goop.NewAPIKeyHeader("X-API-Key", "API key authentication")
	err := openAPIGen.AddSecurityScheme("ApiKeyAuth", apiKeyAuth)
	if err != nil {
		panic(fmt.Sprintf("Failed to add API key security scheme: %v", err))
	}
	
	bearerAuth := goop.NewBearerAuth("JWT", "Bearer token authentication")
	err = openAPIGen.AddSecurityScheme("BearerAuth", bearerAuth)
	if err != nil {
		panic(fmt.Sprintf("Failed to add bearer security scheme: %v", err))
	}
	
	oauth2Auth := goop.NewOAuth2AuthorizationCode(
		"https://auth.example.com/oauth/authorize",
		"https://auth.example.com/oauth/token",
		"https://auth.example.com/oauth/refresh",
		map[string]string{
			"read":  "Read access to user data",
			"write": "Write access to user data",
			"admin": "Administrative access",
		},
		"OAuth2 authentication with various scopes",
	)
	err = openAPIGen.AddSecurityScheme("OAuth2", oauth2Auth)
	if err != nil {
		panic(fmt.Sprintf("Failed to add OAuth2 security scheme: %v", err))
	}
	
	// Set global security (optional - any of these can be used)
	globalSecurity := goop.SecurityRequirements{}.RequireScheme("ApiKeyAuth").RequireScheme("BearerAuth")
	openAPIGen.SetGlobalSecurity(globalSecurity)
	
	router := operations.NewRouter(engine, openAPIGen)
	
	// Define schemas
	userParamsSchema := validators.Object(map[string]interface{}{
		"id": validators.String().Min(1).Required(),
	}).Required()
	
	userResponseSchema := validators.Object(map[string]interface{}{
		"id":       validators.String().Required(),
		"username": validators.String().Required(),
		"email":    validators.Email(),
	}).Required()
	
	adminResponseSchema := validators.Object(map[string]interface{}{
		"message": validators.String().Required(),
		"level":   validators.String().Required(),
	}).Required()
	
	headerSchema := validators.Object(map[string]interface{}{
		"X-Request-ID":  validators.String().Optional(),
		"X-API-Version": validators.String().Optional().Default("v1"),
	}).Optional()
	
	// Public endpoint (no authentication required)
	publicOp := operations.NewSimple().
		GET("/public/users/{id}").
		Summary("Get user (public)").
		Description("Publicly accessible user information").
		Tags("public").
		WithParams(userParamsSchema).
		WithResponse(userResponseSchema).
		NoAuth(). // Explicitly mark as no auth required
		Handler(operations.CreateValidatedHandler(
			getUserHandler,
			userParamsSchema,
			nil,
			nil,
			userResponseSchema,
		))
	
	// API Key protected endpoint
	apiKeyOp := operations.NewSimple().
		GET("/api/users/{id}").
		Summary("Get user (API Key)").
		Description("Get user information with API key authentication").
		Tags("protected").
		WithParams(userParamsSchema).
		WithHeaders(headerSchema).
		WithResponse(userResponseSchema).
		RequireAPIKey("ApiKeyAuth"). // Require API key
		Handler(operations.CreateValidatedHandler(
			getUserHandler,
			userParamsSchema,
			nil,
			nil,
			userResponseSchema,
		))
	
	// Bearer token protected endpoint
	bearerOp := operations.NewSimple().
		GET("/bearer/users/{id}").
		Summary("Get user (Bearer)").
		Description("Get user information with Bearer token authentication").
		Tags("protected").
		WithParams(userParamsSchema).
		WithResponse(userResponseSchema).
		RequireBearer("BearerAuth"). // Require Bearer token
		Handler(operations.CreateValidatedHandler(
			getUserHandler,
			userParamsSchema,
			nil,
			nil,
			userResponseSchema,
		))
	
	// OAuth2 protected endpoint with specific scopes
	oauth2Op := operations.NewSimple().
		GET("/oauth2/users/{id}").
		Summary("Get user (OAuth2)").
		Description("Get user information with OAuth2 authentication and read scope").
		Tags("protected").
		WithParams(userParamsSchema).
		WithResponse(userResponseSchema).
		RequireOAuth2("OAuth2", "read"). // Require OAuth2 with read scope
		Handler(operations.CreateValidatedHandler(
			getUserHandler,
			userParamsSchema,
			nil,
			nil,
			userResponseSchema,
		))
	
	// Admin endpoint requiring OAuth2 with admin scope
	adminOp := operations.NewSimple().
		GET("/admin/status").
		Summary("Admin status").
		Description("Administrative endpoint requiring admin scope").
		Tags("admin").
		WithResponse(adminResponseSchema).
		RequireOAuth2("OAuth2", "admin"). // Require OAuth2 with admin scope
		Handler(operations.CreateValidatedHandler(
			adminHandler,
			nil,
			nil,
			nil,
			adminResponseSchema,
		))
	
	// Multi-auth endpoint (any of the specified auth methods can be used)
	multiAuthOp := operations.NewSimple().
		GET("/multi/users/{id}").
		Summary("Get user (Multiple auth options)").
		Description("Get user with multiple authentication options").
		Tags("flexible").
		WithParams(userParamsSchema).
		WithResponse(userResponseSchema).
		RequireAnyOf("ApiKeyAuth", "BearerAuth"). // Either API key OR Bearer token
		Handler(operations.CreateValidatedHandler(
			getUserHandler,
			userParamsSchema,
			nil,
			nil,
			userResponseSchema,
		))
	
	// Register all operations
	router.Register(publicOp)
	router.Register(apiKeyOp)
	router.Register(bearerOp)
	router.Register(oauth2Op)
	router.Register(adminOp)
	router.Register(multiAuthOp)
	
	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "security-test",
		})
	})
	
	// Serve OpenAPI spec
	engine.GET("/openapi.json", func(c *gin.Context) {
		if err := openAPIGen.WriteToWriter(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate OpenAPI spec",
			})
			return
		}
	})
	
	fmt.Println("🔒 Security Test API starting on :8002")
	fmt.Println("📚 OpenAPI spec available at: http://localhost:8002/openapi.json")
	fmt.Println("🔑 API Key endpoint: GET /api/users/{id} (requires X-API-Key header)")
	fmt.Println("🎫 Bearer endpoint: GET /bearer/users/{id} (requires Authorization: Bearer <token>)")
	fmt.Println("🔐 OAuth2 endpoint: GET /oauth2/users/{id} (requires OAuth2 with 'read' scope)")
	fmt.Println("👑 Admin endpoint: GET /admin/status (requires OAuth2 with 'admin' scope)")
	fmt.Println("🔄 Multi-auth endpoint: GET /multi/users/{id} (accepts API Key OR Bearer)")
	fmt.Println("🌍 Public endpoint: GET /public/users/{id} (no authentication)")
	
	engine.Run(":8002")
}