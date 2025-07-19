package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/picogrid/go-op/operations"
	ginadapter "github.com/picogrid/go-op/operations/adapters/gin"
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

// ContactInfo represents different ways to contact a user (showcases OneOf)
type ContactInfo struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// NotificationPreference represents different notification types (showcases OneOf)
type NotificationPreference struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings"`
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

	// Add enhanced OpenAPI metadata using new Fixed Fields features
	openAPIGen.SetDescription("A comprehensive user management service with full CRUD operations")
	openAPIGen.SetSummary("User Management API")
	openAPIGen.SetTermsOfService("https://example.com/terms")
	openAPIGen.SetContact(&operations.OpenAPIContact{
		Name:  "API Support",
		Email: "support@example.com",
		URL:   "https://example.com/support",
	})
	openAPIGen.SetLicense(&operations.OpenAPILicense{
		Name: "MIT License",
		URL:  "https://opensource.org/licenses/MIT",
	})

	// Add global tags
	openAPIGen.AddTag(operations.OpenAPITag{
		Name:        "users",
		Description: "Operations related to user management",
		ExternalDocs: &operations.OpenAPIExternalDocs{
			Description: "User management documentation",
			URL:         "https://example.com/docs/users",
		},
	})

	// Add global external documentation
	openAPIGen.SetExternalDocs(&operations.OpenAPIExternalDocs{
		Description: "Complete API documentation",
		URL:         "https://example.com/docs",
	})

	// Add server configuration
	openAPIGen.AddServer(operations.OpenAPIServer{
		URL:         "https://api.example.com/{version}",
		Description: "Production server",
		Variables: map[string]operations.OpenAPIServerVariable{
			"version": {
				Default:     "v1",
				Enum:        []string{"v1", "v2"},
				Description: "API version",
			},
		},
	})

	// Create router with generators
	router := ginadapter.NewGinRouter(engine, openAPIGen)

	// ===== OneOf Schema Examples =====
	// These demonstrate complex OneOf patterns for API flexibility

	// Contact method OneOf - email, phone, or social media
	emailContactSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^email$").
			Example("email").
			Required(),
		"email": validators.String().Email().
			Example("contact@example.com").
			Required(),
		"verified": validators.Bool().
			Example(true).
			Optional(),
	}).Example(map[string]interface{}{
		"type":     "email",
		"email":    "contact@example.com",
		"verified": true,
	}).Required()

	phoneContactSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^phone$").
			Example("phone").
			Required(),
		"phone": validators.String().Pattern(`^\+?[1-9]\d{1,14}$`).
			Example("+1234567890").
			Required(),
		"country_code": validators.String().Pattern("^[A-Z]{2}$").
			Example("US").
			Optional(),
	}).Example(map[string]interface{}{
		"type":         "phone",
		"phone":        "+1234567890",
		"country_code": "US",
	}).Required()

	socialContactSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^social$").
			Example("social").
			Required(),
		"platform": validators.String().
			Examples(map[string]validators.ExampleObject{
				"twitter": {
					Summary:     "Twitter handle",
					Description: "Twitter/X platform contact",
					Value:       "twitter",
				},
				"linkedin": {
					Summary:     "LinkedIn profile",
					Description: "LinkedIn professional network",
					Value:       "linkedin",
				},
				"github": {
					Summary:     "GitHub profile",
					Description: "GitHub developer profile",
					Value:       "github",
				},
			}).
			Required(),
		"handle": validators.String().Min(1).
			Example("@johndoe").
			Required(),
	}).Example(map[string]interface{}{
		"type":     "social",
		"platform": "twitter",
		"handle":   "@johndoe",
	}).Required()

	// OneOf contact schema combining all contact methods
	contactInfoSchema := validators.OneOf(
		emailContactSchema,
		phoneContactSchema,
		socialContactSchema,
	).Required()

	// Notification preferences OneOf - email, SMS, push, or webhook
	emailNotificationSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^email$").
			Example("email").
			Required(),
		"email_address": validators.String().Email().
			Example("notifications@example.com").
			Required(),
		"frequency": validators.String().
			Examples(map[string]validators.ExampleObject{
				"immediate": {
					Summary:     "Immediate notifications",
					Description: "Send notifications immediately as events occur",
					Value:       "immediate",
				},
				"daily": {
					Summary:     "Daily digest",
					Description: "Send a daily summary of notifications",
					Value:       "daily",
				},
				"weekly": {
					Summary:     "Weekly summary",
					Description: "Send a weekly compilation of notifications",
					Value:       "weekly",
				},
			}).
			Optional().Default("immediate"),
	}).Example(map[string]interface{}{
		"type":          "email",
		"email_address": "notifications@example.com",
		"frequency":     "immediate",
	}).Required()

	smsNotificationSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^sms$").
			Example("sms").
			Required(),
		"phone_number": validators.String().Pattern(`^\+?[1-9]\d{1,14}$`).
			Example("+1234567890").
			Required(),
		"carrier": validators.String().
			Example("verizon").
			Optional(),
	}).Example(map[string]interface{}{
		"type":         "sms",
		"phone_number": "+1234567890",
		"carrier":      "verizon",
	}).Required()

	pushNotificationSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^push$").
			Example("push").
			Required(),
		"device_tokens": validators.Array(validators.String()).
			Example([]interface{}{
				"device_token_abc123",
				"device_token_def456",
			}).
			Required(),
		"sound": validators.Bool().
			Example(true).
			Optional().Default(true),
		"badge": validators.Bool().
			Example(true).
			Optional().Default(true),
	}).Example(map[string]interface{}{
		"type": "push",
		"device_tokens": []interface{}{
			"device_token_abc123",
			"device_token_def456",
		},
		"sound": true,
		"badge": true,
	}).Required()

	webhookNotificationSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^webhook$").
			Example("webhook").
			Required(),
		"url": validators.String().Pattern(`^https?://`).
			Example("https://api.example.com/webhooks/notifications").
			Required(),
		"secret": validators.String().Min(16).
			Example("webhook_secret_key_12345").
			Optional(),
		"headers": validators.Object(map[string]interface{}{
			"Authorization": validators.String().Optional(),
			"Content-Type":  validators.String().Optional(),
		}).Optional(),
	}).Example(map[string]interface{}{
		"type":   "webhook",
		"url":    "https://api.example.com/webhooks/notifications",
		"secret": "webhook_secret_key_12345",
		"headers": map[string]interface{}{
			"Authorization": "Bearer token123",
			"Content-Type":  "application/json",
		},
	}).Required()

	// OneOf notification preference schema
	notificationPreferenceSchema := validators.OneOf(
		emailNotificationSchema,
		smsNotificationSchema,
		pushNotificationSchema,
		webhookNotificationSchema,
	).Required()

	// User profile update with OneOf fields
	updateUserProfileBodySchema := validators.Object(map[string]interface{}{
		"basic_info": validators.Object(map[string]interface{}{
			"first_name": validators.String().Min(1).Max(100).
				Example("Jane").
				Optional(),
			"last_name": validators.String().Min(1).Max(100).
				Example("Smith").
				Optional(),
			"age": validators.Number().Min(13).Max(120).
				Example(28).
				Optional(),
		}).Optional(),
		"contact_method":          contactInfoSchema.Optional(),
		"notification_preference": notificationPreferenceSchema.Optional(),
		"is_active": validators.Bool().
			Example(true).
			Optional(),
	}).Example(map[string]interface{}{
		"basic_info": map[string]interface{}{
			"first_name": "Jane",
			"last_name":  "Smith",
			"age":        28,
		},
		"contact_method": map[string]interface{}{
			"type":     "email",
			"email":    "jane.smith@example.com",
			"verified": true,
		},
		"notification_preference": map[string]interface{}{
			"type":          "email",
			"email_address": "jane.notifications@example.com",
			"frequency":     "daily",
		},
		"is_active": true,
	}).Optional()

	// Define schemas using go-op validators with comprehensive examples and OpenAPI 3.1 features
	createUserBodySchema := validators.Object(map[string]interface{}{
		"email": validators.String().Email().
			Example("john.doe@example.com").
			Required(),
		"username": validators.String().Min(3).Max(50).Pattern("^[a-zA-Z0-9_]+$").
			Examples(map[string]validators.ExampleObject{
				"simple": {
					Summary:     "Simple username",
					Description: "A basic alphanumeric username",
					Value:       "johndoe",
				},
				"with_underscore": {
					Summary:     "Username with underscore",
					Description: "Username containing underscores",
					Value:       "john_doe_123",
				},
			}).
			Required(),
		"first_name": validators.String().Min(1).Max(100).
			Example("John").
			Required(),
		"last_name": validators.String().Min(1).Max(100).
			Example("Doe").
			Required(),
		"age": validators.Number().
			Integer().           // Ensure whole years only
			MultipleOf(1.0).     // OpenAPI 3.1: Age must be whole years
			ExclusiveMin(12.0).  // OpenAPI 3.1: Must be over 12 (13+)
			ExclusiveMax(150.0). // OpenAPI 3.1: Must be under 150
			Example(25).
			Required(),
		"password": validators.String().Min(8).Max(128).
			Examples(map[string]validators.ExampleObject{
				"strong": {
					Summary:     "Strong password",
					Description: "A secure password with mixed characters",
					Value:       "MyStr0ngP@ssw0rd!",
				},
				"simple": {
					Summary:     "Simple password",
					Description: "A basic but valid password",
					Value:       "password123",
				},
			}).
			Required(),
		"preferences": validators.Object(map[string]interface{}{
			"language":      validators.String().Optional().Default("en"),
			"timezone":      validators.String().Optional(),
			"notifications": validators.Bool().Optional().Default(true),
			"theme":         validators.String().Optional().Default("light"),
		}).
			MinProperties(1).MaxProperties(4). // OpenAPI 3.1: Flexible preferences
			Optional(),
		"interests": validators.Array(validators.String().Min(1).Max(50).Required()).
			UniqueItems(). // OpenAPI 3.1: No duplicate interests
			MinItems(0).MaxItems(10).
			Example([]interface{}{"technology", "music", "sports"}).
			Optional(),
	}).Example(map[string]interface{}{
		"email":      "john.doe@example.com",
		"username":   "johndoe",
		"first_name": "John",
		"last_name":  "Doe",
		"age":        25,
		"password":   "MyStr0ngP@ssw0rd!",
		"preferences": map[string]interface{}{
			"language":      "en",
			"timezone":      "America/New_York",
			"notifications": true,
		},
		"interests": []interface{}{"technology", "music", "sports"},
	}).Required()

	updateUserBodySchema := validators.Object(map[string]interface{}{
		"email": validators.String().Email().
			Example("jane.smith@example.com").
			Optional(),
		"first_name": validators.String().Min(1).Max(100).
			Example("Jane").
			Optional(),
		"last_name": validators.String().Min(1).Max(100).
			Example("Smith").
			Optional(),
		"age": validators.Number().Min(13).Max(120).
			Example(28).
			Optional(),
		"is_active": validators.Bool().
			Examples(map[string]validators.ExampleObject{
				"active": {
					Summary:     "Active user",
					Description: "User account is active and can access the system",
					Value:       true,
				},
				"inactive": {
					Summary:     "Inactive user",
					Description: "User account is disabled",
					Value:       false,
				},
			}).
			Optional(),
	}).Example(map[string]interface{}{
		"email":      "jane.smith@example.com",
		"first_name": "Jane",
		"last_name":  "Smith",
		"age":        28,
		"is_active":  true,
	}).Optional()

	userParamsSchema := validators.Object(map[string]interface{}{
		"id": validators.String().Min(1).Pattern("^usr_[a-zA-Z0-9]+$").
			Examples(map[string]validators.ExampleObject{
				"basic": {
					Summary:     "Basic user ID",
					Description: "Standard user identifier format",
					Value:       "usr_123",
				},
				"alphanumeric": {
					Summary:     "Alphanumeric user ID",
					Description: "User ID with mixed letters and numbers",
					Value:       "usr_abc123def",
				},
			}).
			Required(),
	}).Example(map[string]interface{}{
		"id": "usr_123",
	}).Required()

	listUsersQuerySchema := validators.Object(map[string]interface{}{
		"page": validators.Number().Min(1).
			Example(1).
			Optional().Default(1),
		"page_size": validators.Number().Min(1).Max(100).
			Examples(map[string]validators.ExampleObject{
				"small": {
					Summary:     "Small page size",
					Description: "Useful for mobile apps or limited bandwidth",
					Value:       10,
				},
				"default": {
					Summary:     "Default page size",
					Description: "Standard page size for most use cases",
					Value:       20,
				},
				"large": {
					Summary:     "Large page size",
					Description: "For dashboards or bulk operations",
					Value:       50,
				},
			}).
			Optional().Default(20),
		"search": validators.String().Min(1).Max(255).
			Examples(map[string]validators.ExampleObject{
				"name_search": {
					Summary:     "Search by name",
					Description: "Search for users by first or last name",
					Value:       "John",
				},
				"email_search": {
					Summary:     "Search by email",
					Description: "Search for users by email domain",
					Value:       "@example.com",
				},
			}).
			Optional(),
		"is_active": validators.Bool().
			Example(true).
			Optional(),
		"min_age": validators.Number().Min(0).Max(120).
			Example(18).
			Optional(),
		"max_age": validators.Number().Min(0).Max(120).
			Example(65).
			Optional(),
	}).Example(map[string]interface{}{
		"page":      1,
		"page_size": 20,
		"search":    "John",
		"is_active": true,
		"min_age":   18,
		"max_age":   65,
	}).Optional()

	userResponseSchema := validators.Object(map[string]interface{}{
		"id": validators.String().Min(1).
			Example("usr_123").
			Required(),
		"email": validators.String().Email().
			Example("john.doe@example.com").
			Required(),
		"username": validators.String().Min(1).
			Example("johndoe").
			Required(),
		"first_name": validators.String().Min(1).
			Example("John").
			Required(),
		"last_name": validators.String().Min(1).
			Example("Doe").
			Required(),
		"age": validators.Number().Min(0).
			Example(25).
			Required(),
		"is_active": validators.Bool().
			Example(true).
			Required(),
		"created_at": validators.String().
			Example("2024-01-15T10:30:00Z").
			Required(),
		"updated_at": validators.String().
			Example("2024-01-15T14:45:30Z").
			Required(),
	}).Example(map[string]interface{}{
		"id":         "usr_123",
		"email":      "john.doe@example.com",
		"username":   "johndoe",
		"first_name": "John",
		"last_name":  "Doe",
		"age":        25,
		"is_active":  true,
		"created_at": "2024-01-15T10:30:00Z",
		"updated_at": "2024-01-15T14:45:30Z",
	}).Required()

	userListResponseSchema := validators.Object(map[string]interface{}{
		"users": validators.Array(userResponseSchema).
			Example([]interface{}{
				map[string]interface{}{
					"id":         "usr_123",
					"email":      "john.doe@example.com",
					"username":   "johndoe",
					"first_name": "John",
					"last_name":  "Doe",
					"age":        25,
					"is_active":  true,
					"created_at": "2024-01-15T10:30:00Z",
					"updated_at": "2024-01-15T14:45:30Z",
				},
				map[string]interface{}{
					"id":         "usr_456",
					"email":      "jane.smith@example.com",
					"username":   "janesmith",
					"first_name": "Jane",
					"last_name":  "Smith",
					"age":        28,
					"is_active":  true,
					"created_at": "2024-01-14T09:15:00Z",
					"updated_at": "2024-01-15T11:20:15Z",
				},
			}).
			Required(),
		"total_count": validators.Number().Min(0).
			Example(42).
			Required(),
		"page": validators.Number().Min(1).
			Example(1).
			Required(),
		"page_size": validators.Number().Min(1).
			Example(20).
			Required(),
		"has_next": validators.Bool().
			Examples(map[string]validators.ExampleObject{
				"more_pages": {
					Summary:     "More pages available",
					Description: "There are additional pages of results",
					Value:       true,
				},
				"last_page": {
					Summary:     "Last page",
					Description: "This is the final page of results",
					Value:       false,
				},
			}).
			Required(),
	}).Example(map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"id":         "usr_123",
				"email":      "john.doe@example.com",
				"username":   "johndoe",
				"first_name": "John",
				"last_name":  "Doe",
				"age":        25,
				"is_active":  true,
				"created_at": "2024-01-15T10:30:00Z",
				"updated_at": "2024-01-15T14:45:30Z",
			},
			map[string]interface{}{
				"id":         "usr_456",
				"email":      "jane.smith@example.com",
				"username":   "janesmith",
				"first_name": "Jane",
				"last_name":  "Smith",
				"age":        28,
				"is_active":  true,
				"created_at": "2024-01-14T09:15:00Z",
				"updated_at": "2024-01-15T11:20:15Z",
			},
		},
		"total_count": 42,
		"page":        1,
		"page_size":   20,
		"has_next":    true,
	}).Required()

	// Define operations
	createUserOp := operations.NewSimple().
		POST("/users").
		Summary("Create a new user").
		Description("Creates a new user account with the provided information").
		Tags("users").
		WithBody(createUserBodySchema).
		WithResponse(userResponseSchema).
		Handler(ginadapter.CreateValidatedHandler(
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
		Handler(ginadapter.CreateValidatedHandler(
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
		Handler(ginadapter.CreateValidatedHandler(
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
		Handler(ginadapter.CreateValidatedHandler(
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
		Handler(ginadapter.CreateValidatedHandler(
			listUsersHandler,
			nil,
			listUsersQuerySchema,
			nil,
			userListResponseSchema,
		))

	// New operation showcasing OneOf functionality
	updateUserProfileOp := operations.NewSimple().
		PATCH("/users/{id}/profile").
		Summary("Update user profile with flexible contact and notification settings").
		Description("Updates user profile with OneOf support for contact methods and notification preferences. "+
			"This endpoint demonstrates how OneOf schemas enable flexible API design where clients can "+
			"choose between different data structures (email/phone/social contact, email/SMS/push/webhook notifications).").
		Tags("users", "profile", "oneof-example").
		WithParams(userParamsSchema).
		WithBody(updateUserProfileBodySchema).
		WithResponse(userResponseSchema).
		Handler(ginadapter.CreateValidatedHandler(
			updateUserHandler, // Reuse existing handler for demo
			userParamsSchema,
			nil,
			updateUserProfileBodySchema,
			userResponseSchema,
		))

	// ===== API Version Operation with OpenAPI 3.1 const validation =====
	apiVersionResponseSchema := validators.Object(map[string]interface{}{
		"api_version": validators.String().Const("v1.0").Required().
			Example("v1.0"),
		"service_name": validators.String().Const("user-service").Required().
			Example("user-service"),
		"build_number": validators.String().Pattern(`^\d+\.\d+\.\d+$`).Required().
			Example("1.2.3"),
		"features": validators.Array(validators.String().Required()).
			UniqueItems(). // OpenAPI 3.1: No duplicate features
			MinItems(1).
			Example([]interface{}{"crud", "validation", "openapi31", "oneof"}).
			Required(),
		"supported_formats": validators.Array(validators.String().Required()).
			UniqueItems(). // OpenAPI 3.1: No duplicate formats
			Example([]interface{}{"json", "yaml"}).
			Required(),
	}).Example(map[string]interface{}{
		"api_version":       "v1.0",
		"service_name":      "user-service",
		"build_number":      "1.2.3",
		"features":          []interface{}{"crud", "validation", "openapi31", "oneof"},
		"supported_formats": []interface{}{"json", "yaml"},
	}).Required()

	getAPIVersionOp := operations.NewSimple().
		GET("/api-info").
		Summary("Get API version and service information").
		Description("Returns the current API version with const validation and service features using OpenAPI 3.1 uniqueItems").
		Tags("meta").
		WithResponse(apiVersionResponseSchema).
		Handler(ginadapter.CreateValidatedHandler(
			func(ctx context.Context, params struct{}, query struct{}, body struct{}) (map[string]interface{}, error) {
				return map[string]interface{}{
					"api_version":       "v1.0",
					"service_name":      "user-service",
					"build_number":      "1.2.3",
					"features":          []interface{}{"crud", "validation", "openapi31", "oneof"},
					"supported_formats": []interface{}{"json", "yaml"},
				}, nil
			},
			nil,
			nil,
			nil,
			apiVersionResponseSchema,
		))

	// Register operations
	router.Register(createUserOp)
	router.Register(getUserOp)
	router.Register(updateUserOp)
	router.Register(deleteUserOp)
	router.Register(listUsersOp)
	router.Register(updateUserProfileOp) // OneOf showcase operation
	router.Register(getAPIVersionOp)     // OpenAPI 3.1 features showcase

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
