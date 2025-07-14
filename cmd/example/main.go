package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/picogrid/go-op/operations"
	"github.com/picogrid/go-op/validators"
)

// Example data structures
type User struct {
	ID    string `json:"id" uri:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GetUserParams struct {
	ID string `json:"id" uri:"id"`
}

type GetUserQuery struct {
	IncludeProfile bool `json:"include_profile" form:"include_profile"`
}

// Business logic handlers (pure functions, no HTTP concerns)
func getUserHandler(ctx context.Context, params GetUserParams, query GetUserQuery, body struct{}) (User, error) {
	// Simulate getting user from database
	if params.ID == "" {
		return User{}, fmt.Errorf("user not found")
	}

	return User{
		ID:    params.ID,
		Name:  "John Doe",
		Email: "john@example.com",
	}, nil
}

func createUserHandler(ctx context.Context, params struct{}, query struct{}, body CreateUserRequest) (User, error) {
	// Simulate creating user in database
	return User{
		ID:    "123",
		Name:  body.Name,
		Email: body.Email,
	}, nil
}

func main() {
	// Create Gin engine
	engine := gin.Default()

	// Create OpenAPI generator
	openAPIGen := operations.NewOpenAPIGenerator("User API", "1.0.0")

	// Create router with generators
	router := operations.NewRouter(engine, openAPIGen)

	// Define schemas using existing go-op validators
	getUserParamsSchema := validators.Object(map[string]interface{}{
		"id": validators.String().Min(1).Required(),
	}).Required()

	getUserQuerySchema := validators.Object(map[string]interface{}{
		"include_profile": validators.Bool().Optional(),
	}).Optional()

	userResponseSchema := validators.Object(map[string]interface{}{
		"id":    validators.String().Min(1).Required(),
		"name":  validators.String().Min(1).Required(),
		"email": validators.String().Email().Required(),
	}).Required()

	createUserBodySchema := validators.Object(map[string]interface{}{
		"name":  validators.String().Min(1).Max(100).Required(),
		"email": validators.String().Email().Required(),
	}).Required()

	// Define operations using simple builders
	getUserOp := operations.NewSimple().
		GET("/users/{id}").
		Summary("Get user by ID").
		Description("Retrieve a user by their unique identifier").
		Tags("users").
		WithParams(getUserParamsSchema).
		WithQuery(getUserQuerySchema).
		WithResponse(userResponseSchema).
		Handler(operations.CreateValidatedHandler(
			getUserHandler,
			getUserParamsSchema,
			getUserQuerySchema,
			nil, // no body schema
			userResponseSchema,
		))

	createUserOp := operations.NewSimple().
		POST("/users").
		Summary("Create a new user").
		Description("Create a new user with the provided information").
		Tags("users").
		SuccessCode(operations.StatusCreated).
		WithBody(createUserBodySchema).
		WithResponse(userResponseSchema).
		Handler(operations.CreateValidatedHandler(
			createUserHandler,
			nil, // no params schema
			nil, // no query schema
			createUserBodySchema,
			userResponseSchema,
		))

	// Register operations with zero reflection
	if err := router.Register(getUserOp); err != nil {
		panic(fmt.Sprintf("Failed to register getUserOp: %v", err))
	}

	if err := router.Register(createUserOp); err != nil {
		panic(fmt.Sprintf("Failed to register createUserOp: %v", err))
	}

	// Serve OpenAPI specification
	engine.GET("/openapi.json", func(c *gin.Context) {
		if err := openAPIGen.WriteToWriter(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OpenAPI spec"})
			return
		}
		c.Header("Content-Type", "application/json")
	})

	// Serve Swagger UI (optional)
	engine.GET("/docs", func(c *gin.Context) {
		c.HTML(http.StatusOK, "swagger.html", gin.H{
			"title": "API Documentation",
		})
	})

	// Health check endpoint (simple, no validation needed)
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	fmt.Println("🚀 Server starting on :8080")
	fmt.Println("📚 API Documentation: http://localhost:8080/docs")
	fmt.Println("📋 OpenAPI Spec: http://localhost:8080/openapi.json")
	fmt.Println("❤️  Health Check: http://localhost:8080/health")
	fmt.Println()
	fmt.Println("Example requests:")
	fmt.Println("  GET    http://localhost:8080/users/123?include_profile=true")
	fmt.Println("  POST   http://localhost:8080/users")
	fmt.Println("         Body: {\"name\": \"Jane Doe\", \"email\": \"jane@example.com\"}")

	// Start server
	if err := engine.Run(":8080"); err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}
