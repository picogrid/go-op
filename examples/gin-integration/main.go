package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	goop "github.com/picogrid/go-op"
	"github.com/picogrid/go-op/operations"
	ginadapter "github.com/picogrid/go-op/operations/adapters/gin"
	"github.com/picogrid/go-op/validators"
)

// This example shows how to integrate go-op with an existing Gin application
// It demonstrates:
// 1. Gradual adoption - mixing traditional Gin handlers with go-op validated handlers
// 2. Middleware preservation - keeping existing auth, logging, CORS middleware
// 3. Type-safe validation with automatic OpenAPI generation
// 4. Migration patterns from manual validation to go-op schemas

// Domain models matching an existing system (e.g., Platform Platform)
type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	UserRole  string    `json:"user_role"` // STANDARD, GLOBAL_ADMIN
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DTOs matching existing patterns
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required,min=1,max=255"`
	LastName  string `json:"last_name" binding:"required,min=1,max=255"`
}

type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty" binding:"omitempty,email"`
	FirstName *string `json:"first_name,omitempty" binding:"omitempty,min=1,max=255"`
	LastName  *string `json:"last_name,omitempty" binding:"omitempty,min=1,max=255"`
}

type PaginatedResponse[T any] struct {
	Results    []T                    `json:"results"`
	TotalCount int                    `json:"total_count"`
	Paging     map[string]interface{} `json:"paging"`
}

// Simulated service layer (like Platform's container pattern)
type UserService struct {
	// In real app: database, cache, etc.
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	return &User{
		ID:        uuid.New(),
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		UserRole:  "STANDARD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	return &User{
		ID:        id,
		Email:     "john.doe@example.com",
		FirstName: "John",
		LastName:  "Doe",
		UserRole:  "STANDARD",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error) {
	user := &User{
		ID:        id,
		Email:     "john.doe@example.com",
		FirstName: "John",
		LastName:  "Doe",
		UserRole:  "STANDARD",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) (*PaginatedResponse[User], error) {
	users := []User{
		{
			ID:        uuid.New(),
			Email:     "john.doe@example.com",
			FirstName: "John",
			LastName:  "Doe",
			UserRole:  "STANDARD",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Email:     "jane.smith@example.com",
			FirstName: "Jane",
			LastName:  "Smith",
			UserRole:  "GLOBAL_ADMIN",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now(),
		},
	}

	return &PaginatedResponse[User]{
		Results:    users,
		TotalCount: 2,
		Paging: map[string]interface{}{
			"next":     nil,
			"previous": nil,
		},
	}, nil
}

// Middleware examples (matching existing auth patterns)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simulate JWT validation
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		// In real app: validate JWT, extract claims
		c.Set("user_id", "usr_123")
		c.Set("scopes", []string{"users:read", "users:write"})
		c.Next()
	}
}

func RequireScopes(scopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userScopes, exists := c.Get("scopes")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No scopes found"})
			c.Abort()
			return
		}

		// Check if user has required scopes
		hasScope := false
		for _, required := range scopes {
			for _, userScope := range userScopes.([]string) {
				if userScope == required {
					hasScope = true
					break
				}
			}
		}

		if !hasScope {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Traditional Gin handlers (before go-op)
func traditionalCreateUserHandler(userService *UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := userService.CreateUser(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, user)
	}
}

func traditionalGetUserHandler(userService *UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		user, err := userService.GetUser(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

// Define go-op schemas using type-safe ForStruct pattern
func getUserSchemas() (
	userParamsSchema goop.Schema,
	createUserBodySchema goop.Schema,
	updateUserBodySchema goop.Schema,
	userResponseSchema goop.Schema,
	listQuerySchema goop.Schema,
) {
	// UUID validation pattern
	uuidPattern := "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"

	// Path parameters
	userParamsSchema = validators.Object(map[string]interface{}{
		"id": validators.String().
			Pattern(uuidPattern).
			Example("a1b2c3d4-e5f6-7890-1234-567890abcdef").
			Required(),
	}).Required()

	// NEW: Using ForStruct for type-safe validation
	// This provides compile-time safety and better refactoring support

	// Create user request - using ForStruct with CreateUserRequest type
	createUserBodySchema = validators.ForStruct[CreateUserRequest]().
		Field("email", validators.String().Email().
			Example("jane.doe@example.com").
			Required()).
		Field("first_name", validators.String().Min(1).Max(255).
			Example("Jane").
			Required()).
		Field("last_name", validators.String().Min(1).Max(255).
			Example("Doe").
			Required()).
		Required().
		Schema()

	// Update user request - using ForStruct with UpdateUserRequest type
	updateUserBodySchema = validators.ForStruct[UpdateUserRequest]().
		Field("email", validators.String().Email().
			Example("john.doe.updated@example.com").
			Optional()).
		Field("first_name", validators.String().Min(1).Max(255).
			Example("Johnny").
			Optional()).
		Field("last_name", validators.String().Min(1).Max(255).
			Example("Doe Jr.").
			Optional()).
		Optional().
		Schema()

	// User response - using ForStruct with User type
	userResponseSchema = validators.ForStruct[User]().
		Field("id", validators.String().
			Pattern(uuidPattern).
			Example("a1b2c3d4-e5f6-7890-1234-567890abcdef").
			Required()).
		Field("email", validators.String().Email().
			Example("john.doe@example.com").
			Required()).
		Field("first_name", validators.String().
			Example("John").
			Required()).
		Field("last_name", validators.String().
			Example("Doe").
			Required()).
		Field("user_role", validators.String().
			Pattern("^(STANDARD|GLOBAL_ADMIN)$").
			Example("STANDARD").
			Required()).
		Field("created_at", validators.String().
			Pattern("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}Z$").
			Example("2024-02-16T21:45:33Z").
			Required()).
		Field("updated_at", validators.String().
			Pattern("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}Z$").
			Example("2024-02-16T21:45:33Z").
			Required()).
		Required().
		Schema()

	// List query parameters (using inline struct type for demonstration)
	type ListQuery struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}

	listQuerySchema = validators.ForStruct[ListQuery]().
		Field("limit", validators.Number().Min(1).Max(1000).
			Example(10).
			Optional().Default(10)).
		Field("offset", validators.Number().Min(0).
			Example(0).
			Optional().Default(0)).
		Optional().
		Schema()

	return
}

// go-op validated handlers that call the service layer
func createGoOpHandlers(userService *UserService) (
	createHandler operations.Handler[struct{}, struct{}, CreateUserRequest, User],
	getHandler operations.Handler[struct{ ID string }, struct{}, struct{}, User],
	updateHandler operations.Handler[struct{ ID string }, struct{}, UpdateUserRequest, User],
	listHandler operations.Handler[struct{}, struct {
		Limit  int
		Offset int
	}, struct{}, PaginatedResponse[User]],
) {
	// Create user handler
	createHandler = func(ctx context.Context, params struct{}, query struct{}, body CreateUserRequest) (User, error) {
		user, err := userService.CreateUser(ctx, &body)
		if err != nil {
			return User{}, err
		}
		return *user, nil
	}

	// Get user handler
	getHandler = func(ctx context.Context, params struct{ ID string }, query struct{}, body struct{}) (User, error) {
		id, err := uuid.Parse(params.ID)
		if err != nil {
			return User{}, fmt.Errorf("invalid user ID format")
		}

		user, err := userService.GetUser(ctx, id)
		if err != nil {
			return User{}, err
		}
		return *user, nil
	}

	// Update user handler
	updateHandler = func(ctx context.Context, params struct{ ID string }, query struct{}, body UpdateUserRequest) (User, error) {
		id, err := uuid.Parse(params.ID)
		if err != nil {
			return User{}, fmt.Errorf("invalid user ID format")
		}

		user, err := userService.UpdateUser(ctx, id, &body)
		if err != nil {
			return User{}, err
		}
		return *user, nil
	}

	// List users handler
	listHandler = func(ctx context.Context, params struct{}, query struct {
		Limit  int
		Offset int
	}, body struct{}) (PaginatedResponse[User], error) {
		resp, err := userService.ListUsers(ctx, query.Limit, query.Offset)
		if err != nil {
			return PaginatedResponse[User]{}, err
		}
		return *resp, nil
	}

	return
}

func main() {
	// Initialize services (simulating DI container)
	userService := &UserService{}

	// Create Gin engine with existing middleware
	engine := gin.New()

	// Apply existing middleware (preserving current setup)
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	// In real app: add CORS, tracing, metrics, etc.

	// Create OpenAPI generator
	openAPIGen := operations.NewOpenAPIGenerator("User Management API", "1.0.0")
	openAPIGen.SetDescription("User management service with gradual go-op adoption")
	openAPIGen.SetContact(&operations.OpenAPIContact{
		Name:  "API Support",
		Email: "support@example.com",
	})

	// Add security schemes matching existing auth
	bearerScheme := &goop.HTTPSecurityScheme{
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT token for authentication",
	}
	openAPIGen.AddSecurityScheme("BearerAuth", bearerScheme)

	// Create go-op router wrapping Gin
	goOpRouter := ginadapter.NewGinRouter(engine, openAPIGen)

	// Get schemas and handlers
	userParamsSchema, createUserBodySchema, updateUserBodySchema, userResponseSchema, listQuerySchema := getUserSchemas()
	createHandler, getHandler, updateHandler, listHandler := createGoOpHandlers(userService)

	// Example: Gradual migration - some endpoints use traditional Gin, others use go-op

	// Traditional Gin endpoints (existing code)
	v1 := engine.Group("/v1")
	{
		// These endpoints use traditional Gin validation
		v1.POST("/users", AuthMiddleware(), RequireScopes("users:write"), traditionalCreateUserHandler(userService))
		v1.GET("/users/:id", AuthMiddleware(), RequireScopes("users:read"), traditionalGetUserHandler(userService))
	}

	// go-op validated endpoints (new or migrated)
	// Create operations with proper security requirements
	createUserOp := operations.NewSimple().
		POST("/v2/users").
		Summary("Create user").
		Description("Creates a new user with automatic validation and OpenAPI generation").
		Tags("users").
		RequireAuth("BearerAuth"). // Automatic security requirement
		WithBody(createUserBodySchema).
		WithResponse(userResponseSchema).
		Handler(func(c *gin.Context) {
			// Apply auth middleware inline for this example
			AuthMiddleware()(c)
			if c.IsAborted() {
				return
			}
			RequireScopes("users:write")(c)
			if c.IsAborted() {
				return
			}

			// Use go-op validated handler
			ginadapter.CreateValidatedHandler(
				createHandler,
				nil,
				nil,
				createUserBodySchema,
				userResponseSchema,
			)(c)
		})

	getUserOp := operations.NewSimple().
		GET("/v2/users/{id}").
		Summary("Get user by ID").
		Description("Retrieves a specific user by their ID with automatic validation").
		Tags("users").
		RequireAuth("BearerAuth").
		WithParams(userParamsSchema).
		WithResponse(userResponseSchema).
		Handler(func(c *gin.Context) {
			// Apply auth middleware
			AuthMiddleware()(c)
			if c.IsAborted() {
				return
			}
			RequireScopes("users:read")(c)
			if c.IsAborted() {
				return
			}

			// Use go-op validated handler
			ginadapter.CreateValidatedHandler(
				getHandler,
				userParamsSchema,
				nil,
				nil,
				userResponseSchema,
			)(c)
		})

	updateUserOp := operations.NewSimple().
		PUT("/v2/users/{id}").
		Summary("Update user").
		Description("Updates a user's information with partial update support").
		Tags("users").
		RequireAuth("BearerAuth").
		WithParams(userParamsSchema).
		WithBody(updateUserBodySchema).
		WithResponse(userResponseSchema).
		Handler(func(c *gin.Context) {
			// Apply auth middleware
			AuthMiddleware()(c)
			if c.IsAborted() {
				return
			}
			RequireScopes("users:read", "users:write")(c)
			if c.IsAborted() {
				return
			}

			// Use go-op validated handler
			ginadapter.CreateValidatedHandler(
				updateHandler,
				userParamsSchema,
				nil,
				updateUserBodySchema,
				userResponseSchema,
			)(c)
		})

	listUsersOp := operations.NewSimple().
		GET("/v2/users").
		Summary("List users").
		Description("Retrieves a paginated list of all users").
		Tags("users").
		RequireAuth("BearerAuth").
		WithQuery(listQuerySchema).
		WithResponse(validators.Object(map[string]interface{}{
			"results":     validators.Array(userResponseSchema).Required(),
			"total_count": validators.Number().Required(),
			"paging": validators.Object(map[string]interface{}{
				"next":     validators.String().Optional(),
				"previous": validators.String().Optional(),
			}).Required(),
		}).Required()).
		Handler(func(c *gin.Context) {
			// Apply auth middleware
			AuthMiddleware()(c)
			if c.IsAborted() {
				return
			}
			RequireScopes("users:read")(c)
			if c.IsAborted() {
				return
			}

			// Use go-op validated handler
			ginadapter.CreateValidatedHandler(
				listHandler,
				nil,
				listQuerySchema,
				nil,
				nil, // Response validation handled by operation
			)(c)
		})

	// Register go-op operations
	goOpRouter.Register(createUserOp)
	goOpRouter.Register(getUserOp)
	goOpRouter.Register(updateUserOp)
	goOpRouter.Register(listUsersOp)

	// Health check (traditional Gin handler)
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"versions": gin.H{
				"v1": "traditional Gin validation",
				"v2": "go-op validated endpoints",
			},
		})
	})

	// OpenAPI spec endpoint
	engine.GET("/openapi.json", goOpRouter.ServeSpec(openAPIGen))

	// Add a demonstration endpoint that showcases type-safe validation
	engine.POST("/demo/validate", func(c *gin.Context) {
		// Get raw request data
		var rawData map[string]interface{}
		if err := c.ShouldBindJSON(&rawData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		// Use type-safe ValidateStruct function
		user, err := validators.ValidateStruct[CreateUserRequest](createUserBodySchema, rawData)
		if err != nil {
			// go-op provides detailed validation errors
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": err.Error(),
			})
			return
		}

		// At this point, 'user' is a properly typed *CreateUserRequest
		c.JSON(http.StatusOK, gin.H{
			"message":   "Validation successful!",
			"validated": user,
			"type_safe": fmt.Sprintf("Email: %s, Name: %s %s", user.Email, user.FirstName, user.LastName),
		})
	})

	fmt.Println("🚀 Gin Integration Example starting on :8003")
	fmt.Println("📝 Traditional endpoints: /v1/users (Gin validation)")
	fmt.Println("✨ go-op endpoints: /v2/users (automatic validation + OpenAPI)")
	fmt.Println("🧪 Type-safe demo: POST /demo/validate")
	fmt.Println("📚 OpenAPI spec: http://localhost:8003/openapi.json")
	fmt.Println("🏥 Health check: http://localhost:8003/health")
	fmt.Println()
	fmt.Println("This example demonstrates:")
	fmt.Println("1. Gradual migration from traditional Gin to go-op")
	fmt.Println("2. Preserving existing middleware and auth patterns")
	fmt.Println("3. Type-safe validation with ForStruct and ValidateStruct")
	fmt.Println("4. Side-by-side operation of old and new endpoints")
	fmt.Println()
	fmt.Println("Key ForStruct benefits:")
	fmt.Println("- Compile-time type safety with your existing structs")
	fmt.Println("- IDE autocompletion for struct fields")
	fmt.Println("- Refactoring safety - renaming struct fields will show errors")
	fmt.Println("- Zero runtime reflection for maximum performance")

	engine.Run(":8003")
}
