package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/picogrid/go-op/operations"
	ginadapter "github.com/picogrid/go-op/operations/adapters/gin"
	"github.com/picogrid/go-op/validators"
)

// This example demonstrates comprehensive middleware patterns with go-op and Gin
// It shows both global middleware (applied to all routes) and per-operation middleware
// using the new WithMiddleware functionality.

// Domain models
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"` // "user", "admin", "super_admin"
}

type Company struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Settings struct {
	NotificationsEnabled bool              `json:"notifications_enabled"`
	Theme                string            `json:"theme"`
	Preferences          map[string]string `json:"preferences"`
}

// Request/Response types
type CreateSettingsRequest struct {
	NotificationsEnabled bool              `json:"notifications_enabled"`
	Theme                string            `json:"theme"`
	Preferences          map[string]string `json:"preferences,omitempty"`
}

type SettingsResponse struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	UserID    string    `json:"user_id"`
	Settings  Settings  `json:"settings"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Global middleware - applied to all routes
func RequestLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health", "/metrics"},
	})
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-COMPANY-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	}
}

// Authentication middleware - validates JWT tokens
func RequireJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization header",
				"code":  "AUTH_MISSING",
			})
			c.Abort()
			return
		}

		// Simple Bearer token validation (in real app, validate JWT)
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format",
				"code":  "AUTH_INVALID_FORMAT",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Simulate JWT validation and user extraction
		var user User
		switch token {
		case "user_token":
			user = User{ID: "usr_123", Email: "user@example.com", Role: "user"}
		case "admin_token":
			user = User{ID: "usr_456", Email: "admin@example.com", Role: "admin"}
		case "super_admin_token":
			user = User{ID: "usr_789", Email: "super@example.com", Role: "super_admin"}
		default:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "AUTH_INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Store user in context for downstream handlers
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Next()
	}
}

// Company context middleware - extracts company ID from header
func RequireCompanyID() gin.HandlerFunc {
	return func(c *gin.Context) {
		companyID := c.GetHeader("X-COMPANY-ID")
		if companyID == "" {
			// Also check query parameter as fallback
			companyID = c.Query("company_id")
		}

		if companyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Company ID is required",
				"code":  "COMPANY_ID_MISSING",
			})
			c.Abort()
			return
		}

		// Validate company ID format (UUID)
		if _, err := uuid.Parse(companyID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid company ID format",
				"code":  "COMPANY_ID_INVALID",
			})
			c.Abort()
			return
		}

		// Simulate company validation and storage
		company := Company{
			ID:   companyID,
			Name: fmt.Sprintf("Company %s", companyID[:8]),
		}

		c.Set("company", company)
		c.Set("company_id", companyID)
		c.Next()
	}
}

// Role-based access control middleware
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "User context not found",
				"code":  "USER_CONTEXT_MISSING",
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user context",
				"code":  "USER_CONTEXT_INVALID",
			})
			c.Abort()
			return
		}

		// Check if user has any of the allowed roles
		hasRole := false
		for _, role := range allowedRoles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Insufficient permissions. Required roles: %v, user role: %s", allowedRoles, user.Role),
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Scope-based access control middleware (similar to your RequireOrgScopes)
func RequireScopes(requiredScopes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "User context not found",
				"code":  "USER_CONTEXT_MISSING",
			})
			c.Abort()
			return
		}

		user := userInterface.(User)

		// Simulate scope resolution based on user role and org
		var userScopes []string
		switch user.Role {
		case "user":
			userScopes = []string{"settings:read"}
		case "admin":
			userScopes = []string{"settings:read", "settings:write", "users:read"}
		case "super_admin":
			userScopes = []string{"settings:read", "settings:write", "users:read", "users:write", "admin:all"}
		}

		// Check if user has all required scopes
		hasAllScopes := true
		missingScopes := []string{}

		for _, required := range requiredScopes {
			hasScope := false
			for _, userScope := range userScopes {
				if userScope == required || userScope == "admin:all" {
					hasScope = true
					break
				}
			}
			if !hasScope {
				hasAllScopes = false
				missingScopes = append(missingScopes, required)
			}
		}

		if !hasAllScopes {
			c.JSON(http.StatusForbidden, gin.H{
				"error":       "Insufficient scopes",
				"code":        "INSUFFICIENT_SCOPES",
				"required":    requiredScopes,
				"missing":     missingScopes,
				"user_scopes": userScopes,
			})
			c.Abort()
			return
		}

		// Store validated scopes for downstream use
		c.Set("validated_scopes", requiredScopes)
		c.Next()
	}
}

// Rate limiting middleware (simple version)
func RateLimitMiddleware() gin.HandlerFunc {
	// In real app, use redis or similar for distributed rate limiting
	return func(c *gin.Context) {
		// Simulate rate limiting check
		c.Header("X-RateLimit-Remaining", "99")
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()))
		c.Next()
	}
}

// Business logic handlers (pure functions - no HTTP concerns)
func getSettingsHandler(ctx context.Context, params struct{}, query struct{}, body struct{}) (SettingsResponse, error) {
	// Simulate database lookup
	return SettingsResponse{
		ID:        "set_" + uuid.New().String()[:8],
		CompanyID: "comp_123",
		UserID:    "usr_123",
		Settings: Settings{
			NotificationsEnabled: true,
			Theme:                "dark",
			Preferences: map[string]string{
				"language": "en",
				"timezone": "UTC",
			},
		},
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}, nil
}

func createSettingsHandler(ctx context.Context, params struct{}, query struct{}, body CreateSettingsRequest) (SettingsResponse, error) {
	// Simulate settings creation
	return SettingsResponse{
		ID:        "set_" + uuid.New().String()[:8],
		CompanyID: "comp_123",
		UserID:    "usr_123",
		Settings:  Settings(body),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func getUserInfoHandler(ctx context.Context, params struct{}, query struct{}, body struct{}) (User, error) {
	// This would typically get user from context or database
	return User{
		ID:    "usr_123",
		Email: "user@example.com",
		Role:  "admin",
	}, nil
}

// Gin handler adapters that extract context and call business logic
func ginGetSettings(c *gin.Context) {
	result, err := getSettingsHandler(c.Request.Context(), struct{}{}, struct{}{}, struct{}{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func ginCreateSettings(c *gin.Context) {
	var req CreateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := createSettingsHandler(c.Request.Context(), struct{}{}, struct{}{}, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func ginGetUserInfo(c *gin.Context) {
	result, err := getUserInfoHandler(c.Request.Context(), struct{}{}, struct{}{}, struct{}{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func main() {
	// Create Gin engine
	engine := gin.New()

	// Apply global middleware using GetEngine().Use() pattern
	engine.Use(gin.Recovery())
	engine.Use(RequestLoggingMiddleware())
	engine.Use(CORSMiddleware())
	engine.Use(SecurityHeadersMiddleware())
	engine.Use(RateLimitMiddleware())

	// Global auth middleware - all routes require authentication
	engine.Use(RequireJWTAuth())

	// Global org context - all routes require company ID
	engine.Use(RequireCompanyID())

	// Create OpenAPI generator
	openAPIGen := operations.NewOpenAPIGenerator("Middleware Patterns API", "1.0.0")
	openAPIGen.SetDescription("Comprehensive example of middleware patterns with go-op and Gin")

	// Create go-op router
	router := ginadapter.NewGinRouter(engine, openAPIGen)

	// Define schemas for validation and OpenAPI generation
	createSettingsSchema := validators.Object(map[string]interface{}{
		"notifications_enabled": validators.Bool().Required(),
		"theme": validators.String().
			Pattern("^(light|dark|auto)$").
			Required(),
		"preferences": validators.Object(map[string]interface{}{}).
			Optional(),
	}).Required()

	settingsResponseSchema := validators.Object(map[string]interface{}{
		"id":         validators.String().Required(),
		"company_id": validators.String().Required(),
		"user_id":    validators.String().Required(),
		"settings": validators.Object(map[string]interface{}{
			"notifications_enabled": validators.Bool().Required(),
			"theme":                 validators.String().Required(),
			"preferences":           validators.Object(map[string]interface{}{}).Required(),
		}).Required(),
		"created_at": validators.String().Required(),
		"updated_at": validators.String().Required(),
	}).Required()

	userResponseSchema := validators.Object(map[string]interface{}{
		"id":    validators.String().Required(),
		"email": validators.String().Email().Required(),
		"role":  validators.String().Required(),
	}).Required()

	// Define operations with per-operation middleware using WithMiddleware

	// 1. Get settings - requires read scope only
	getSettingsOp := operations.NewSimple().
		GET("/api/v1/settings").
		Summary("Get application settings").
		Description("Retrieves settings for the current company").
		Tags("Settings").
		WithResponse(settingsResponseSchema).
		Handler(router.WithMiddleware(
			ginGetSettings,
			RequireScopes([]string{"settings:read"}),
		))

	// 2. Create settings - requires write scope and admin role
	createSettingsOp := operations.NewSimple().
		POST("/api/v1/settings").
		Summary("Create application settings").
		Description("Creates new settings for the current company").
		Tags("Settings").
		WithBody(createSettingsSchema).
		WithResponse(settingsResponseSchema).
		Handler(router.WithMiddleware(
			ginCreateSettings,
			RequireRole("admin", "super_admin"),       // Must be admin or super_admin
			RequireScopes([]string{"settings:write"}), // Must have write scope
		))

	// 3. Get user info - public endpoint (only needs basic auth from global middleware)
	getUserInfoOp := operations.NewSimple().
		GET("/v1/user/me").
		Summary("Get current user information").
		Description("Returns information about the currently authenticated user").
		Tags("User").
		WithResponse(userResponseSchema).
		Handler(ginGetUserInfo) // No additional middleware - global auth is sufficient

	// 4. Admin-only endpoint - requires super admin role
	adminOnlyOp := operations.NewSimple().
		GET("/v1/admin/status").
		Summary("Get admin status").
		Description("Admin-only endpoint for system status").
		Tags("Admin").
		WithResponse(validators.Object(map[string]interface{}{
			"status": validators.String().Required(),
			"uptime": validators.String().Required(),
		}).Required()).
		Handler(router.WithMiddleware(
			func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"status": "healthy",
					"uptime": "24h",
				})
			},
			RequireRole("super_admin"), // Only super admins can access
		))

	// Register all operations
	// Register all operations using variadic Register method
	router.Register(
		getSettingsOp,
		createSettingsOp,
		getUserInfoOp,
		adminOnlyOp,
	)

	// Health check endpoint (traditional Gin - no go-op)
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "middleware-patterns-example",
		})
	})

	// OpenAPI spec endpoint
	engine.GET("/openapi.json", router.ServeSpec(openAPIGen))

	// Documentation endpoint
	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "Middleware Patterns Example",
			"patterns": gin.H{
				"global_middleware": []string{
					"Recovery", "Logging", "CORS", "Security Headers",
					"Rate Limiting", "JWT Auth", "Org Context",
				},
				"per_operation_middleware": []string{
					"Role-based Access Control", "Scope-based Access Control",
				},
			},
			"endpoints": gin.H{
				"health":     "GET /health",
				"openapi":    "GET /openapi.json",
				"settings":   "GET /api/v1/settings (scope: settings:read)",
				"create":     "POST /api/v1/settings (role: admin+, scope: settings:write)",
				"user_info":  "GET /v1/user/me (basic auth only)",
				"admin_only": "GET /v1/admin/status (role: super_admin)",
			},
			"test_tokens": gin.H{
				"user":        "user_token",
				"admin":       "admin_token",
				"super_admin": "super_admin_token",
			},
		})
	})

	fmt.Println("🚀 Middleware Patterns Example starting on :8080")
	fmt.Println()
	fmt.Println("📝 Example demonstrates:")
	fmt.Println("  • Global middleware with engine.Use() for all routes")
	fmt.Println("  • Per-operation middleware with router.WithMiddleware()")
	fmt.Println("  • Role-based and scope-based access control")
	fmt.Println("  • Integration with go-op validation and OpenAPI generation")
	fmt.Println()
	fmt.Println("🔗 Available endpoints:")
	fmt.Println("  • GET  /                     - API documentation")
	fmt.Println("  • GET  /health               - Health check")
	fmt.Println("  • GET  /openapi.json         - OpenAPI specification")
	fmt.Println("  • GET  /api/v1/settings    - Get settings (scope: settings:read)")
	fmt.Println("  • POST /api/v1/settings    - Create settings (role: admin+, scope: settings:write)")
	fmt.Println("  • GET  /v1/user/me           - User info (basic auth)")
	fmt.Println("  • GET  /v1/admin/status      - Admin status (role: super_admin)")
	fmt.Println()
	fmt.Println("🔑 Test with different tokens:")
	fmt.Println("  curl -H 'Authorization: Bearer user_token' -H 'X-COMPANY-ID: 550e8400-e29b-41d4-a716-446655440000' http://localhost:8080/v1/user/me")
	fmt.Println("  curl -H 'Authorization: Bearer admin_token' -H 'X-COMPANY-ID: 550e8400-e29b-41d4-a716-446655440000' http://localhost:8080/api/v1/settings")
	fmt.Println("  curl -H 'Authorization: Bearer super_admin_token' -H 'X-COMPANY-ID: 550e8400-e29b-41d4-a716-446655440000' http://localhost:8080/v1/admin/status")

	engine.Run(":8080")
}
