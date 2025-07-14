package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/picogrid/go-op/operations"
	"github.com/picogrid/go-op/validators"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" uri:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Age       int       `json:"age"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	Password  string `json:"password"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Age       *int    `json:"age,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// GetUserParams represents path parameters for getting a user
type GetUserParams struct {
	ID string `json:"id" uri:"id"`
}

// UpdateUserParams represents path parameters for updating a user
type UpdateUserParams struct {
	ID string `json:"id" uri:"id"`
}

// DeleteUserParams represents path parameters for deleting a user
type DeleteUserParams struct {
	ID string `json:"id" uri:"id"`
}

// ListUsersQuery represents query parameters for listing users
type ListUsersQuery struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Search   string `json:"search" form:"search"`
	IsActive *bool  `json:"is_active" form:"is_active"`
	MinAge   *int   `json:"min_age" form:"min_age"`
	MaxAge   *int   `json:"max_age" form:"max_age"`
}

// UserListResponse represents the response for listing users
type UserListResponse struct {
	Users      []User `json:"users"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	HasNext    bool   `json:"has_next"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Business logic handlers
func createUserHandler(ctx context.Context, params struct{}, query struct{}, body CreateUserRequest) (User, error) {
	// Simulate user creation
	return User{
		ID:        "usr_123",
		Email:     body.Email,
		Username:  body.Username,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Age:       body.Age,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func getUserHandler(ctx context.Context, params GetUserParams, query struct{}, body struct{}) (User, error) {
	// Simulate getting user from database
	return User{
		ID:        params.ID,
		Email:     "john.doe@example.com",
		Username:  "johndoe",
		FirstName: "John",
		LastName:  "Doe",
		Age:       30,
		IsActive:  true,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}, nil
}

func updateUserHandler(ctx context.Context, params UpdateUserParams, query struct{}, body UpdateUserRequest) (User, error) {
	// Simulate updating user
	user := User{
		ID:        params.ID,
		Email:     "john.doe@example.com",
		Username:  "johndoe",
		FirstName: "John",
		LastName:  "Doe",
		Age:       30,
		IsActive:  true,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	// Apply updates
	if body.Email != nil {
		user.Email = *body.Email
	}
	if body.FirstName != nil {
		user.FirstName = *body.FirstName
	}
	if body.LastName != nil {
		user.LastName = *body.LastName
	}
	if body.Age != nil {
		user.Age = *body.Age
	}
	if body.IsActive != nil {
		user.IsActive = *body.IsActive
	}

	return user, nil
}

func deleteUserHandler(ctx context.Context, params DeleteUserParams, query struct{}, body struct{}) (struct{}, error) {
	// Simulate user deletion
	return struct{}{}, nil
}

func listUsersHandler(ctx context.Context, params struct{}, query ListUsersQuery, body struct{}) (UserListResponse, error) {
	// Simulate listing users with pagination and filtering
	users := []User{
		{
			ID:        "usr_123",
			Email:     "john.doe@example.com",
			Username:  "johndoe",
			FirstName: "John",
			LastName:  "Doe",
			Age:       30,
			IsActive:  true,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "usr_456",
			Email:     "jane.smith@example.com",
			Username:  "janesmith",
			FirstName: "Jane",
			LastName:  "Smith",
			Age:       25,
			IsActive:  true,
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},
	}

	return UserListResponse{
		Users:      users,
		TotalCount: 2,
		Page:       query.Page,
		PageSize:   query.PageSize,
		HasNext:    false,
	}, nil
}

func main() {
	// Create Gin engine
	engine := gin.Default()

	// Create OpenAPI generator
	openAPIGen := operations.NewOpenAPIGenerator("User Service API", "1.0.0")

	// Create router with generators
	router := operations.NewRouter(engine, openAPIGen)

	// Define schemas using go-op validators
	createUserBodySchema := validators.Object(map[string]interface{}{
		"email":      validators.String().Email().Required(),
		"username":   validators.String().Min(3).Max(50).Pattern("^[a-zA-Z0-9_]+$").Required(),
		"first_name": validators.String().Min(1).Max(100).Required(),
		"last_name":  validators.String().Min(1).Max(100).Required(),
		"age":        validators.Number().Min(13).Max(120).Required(),
		"password":   validators.String().Min(8).Max(128).Required(),
	}).Required()

	updateUserBodySchema := validators.Object(map[string]interface{}{
		"email":      validators.String().Email().Optional(),
		"first_name": validators.String().Min(1).Max(100).Optional(),
		"last_name":  validators.String().Min(1).Max(100).Optional(),
		"age":        validators.Number().Min(13).Max(120).Optional(),
		"is_active":  validators.Bool().Optional(),
	}).Optional()

	userParamsSchema := validators.Object(map[string]interface{}{
		"id": validators.String().Min(1).Pattern("^usr_[a-zA-Z0-9]+$").Required(),
	}).Required()

	listUsersQuerySchema := validators.Object(map[string]interface{}{
		"page":      validators.Number().Min(1).Optional().Default(1),
		"page_size": validators.Number().Min(1).Max(100).Optional().Default(20),
		"search":    validators.String().Min(1).Max(255).Optional(),
		"is_active": validators.Bool().Optional(),
		"min_age":   validators.Number().Min(0).Max(120).Optional(),
		"max_age":   validators.Number().Min(0).Max(120).Optional(),
	}).Optional()

	userResponseSchema := validators.Object(map[string]interface{}{
		"id":         validators.String().Min(1).Required(),
		"email":      validators.String().Email().Required(),
		"username":   validators.String().Min(1).Required(),
		"first_name": validators.String().Min(1).Required(),
		"last_name":  validators.String().Min(1).Required(),
		"age":        validators.Number().Min(0).Required(),
		"is_active":  validators.Bool().Required(),
		"created_at": validators.String().Required(),
		"updated_at": validators.String().Required(),
	}).Required()

	userListResponseSchema := validators.Object(map[string]interface{}{
		"users":       validators.Array(userResponseSchema).Required(),
		"total_count": validators.Number().Min(0).Required(),
		"page":        validators.Number().Min(1).Required(),
		"page_size":   validators.Number().Min(1).Required(),
		"has_next":    validators.Bool().Required(),
	}).Required()

	// Define operations
	createUserOp := operations.NewSimple().
		POST("/users").
		Summary("Create a new user").
		Description("Creates a new user account with the provided information").
		Tags("users").
		WithBody(createUserBodySchema).
		WithResponse(userResponseSchema).
		Handler(operations.CreateValidatedHandler(
			createUserHandler,
			nil,
			nil,
			createUserBodySchema,
			userResponseSchema,
		))

	getUserOp := operations.NewSimple().
		GET("/users/{id}").
		Summary("Get user by ID").
		Description("Retrieves a specific user by their unique identifier").
		Tags("users").
		WithParams(userParamsSchema).
		WithResponse(userResponseSchema).
		Handler(operations.CreateValidatedHandler(
			getUserHandler,
			userParamsSchema,
			nil,
			nil,
			userResponseSchema,
		))

	updateUserOp := operations.NewSimple().
		PUT("/users/{id}").
		Summary("Update user").
		Description("Updates an existing user with the provided information").
		Tags("users").
		WithParams(userParamsSchema).
		WithBody(updateUserBodySchema).
		WithResponse(userResponseSchema).
		Handler(operations.CreateValidatedHandler(
			updateUserHandler,
			userParamsSchema,
			nil,
			updateUserBodySchema,
			userResponseSchema,
		))

	deleteUserOp := operations.NewSimple().
		DELETE("/users/{id}").
		Summary("Delete user").
		Description("Permanently deletes a user account").
		Tags("users").
		WithParams(userParamsSchema).
		Handler(operations.CreateValidatedHandler(
			deleteUserHandler,
			userParamsSchema,
			nil,
			nil,
			nil,
		))

	listUsersOp := operations.NewSimple().
		GET("/users").
		Summary("List users").
		Description("Retrieves a paginated list of users with optional filtering").
		Tags("users").
		WithQuery(listUsersQuerySchema).
		WithResponse(userListResponseSchema).
		Handler(operations.CreateValidatedHandler(
			listUsersHandler,
			nil,
			listUsersQuerySchema,
			nil,
			userListResponseSchema,
		))

	// Register operations
	router.Register(createUserOp)
	router.Register(getUserOp)
	router.Register(updateUserOp)
	router.Register(deleteUserOp)
	router.Register(listUsersOp)

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "user-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	fmt.Println("🚀 User Service starting on :8001")
	fmt.Println("📚 Generate OpenAPI spec: go-op generate -i ./examples/user-service -o ./user-service.yaml")
	engine.Run(":8001")
}
