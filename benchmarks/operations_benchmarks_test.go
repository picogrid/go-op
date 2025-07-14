package benchmarks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/picogrid/go-op/operations"
	"github.com/picogrid/go-op/validators"
)

// BenchmarkOperationBuilding tests the performance of building operations
func BenchmarkOperationBuilding(b *testing.B) {
	// Simple operation
	b.Run("Simple_GET", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = operations.NewSimple().
				GET("/users").
				Summary("List users")
		}
	})

	// Operation with parameters
	b.Run("With_Params", func(b *testing.B) {
		paramsSchema := validators.Object(map[string]interface{}{
			"id": validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
		}).Required()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = operations.NewSimple().
				GET("/users/{id}").
				Summary("Get user by ID").
				WithParams(paramsSchema)
		}
	})

	// Operation with body and response
	b.Run("With_Body_And_Response", func(b *testing.B) {
		bodySchema := validators.Object(map[string]interface{}{
			"name":  validators.String().Min(1).Max(100).Required(),
			"email": validators.Email(),
			"age":   validators.Number().Min(18).Max(120).Optional(),
		}).Required()

		responseSchema := validators.Object(map[string]interface{}{
			"id":        validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
			"name":      validators.String().Required(),
			"email":     validators.String().Required(),
			"createdAt": validators.String().Required(),
		}).Required()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = operations.NewSimple().
				POST("/users").
				Summary("Create user").
				WithBody(bodySchema).
				WithResponse(responseSchema)
		}
	})

	// Complex operation with all components
	b.Run("Complex_Full_Operation", func(b *testing.B) {
		paramsSchema := validators.Object(map[string]interface{}{
			"id": validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
		}).Required()

		querySchema := validators.Object(map[string]interface{}{
			"include": validators.String().Pattern(`^(profile|posts|comments)$`).Optional(),
			"page":    validators.Number().Min(1).Optional(),
			"limit":   validators.Number().Min(1).Max(100).Optional(),
		}).Optional()

		bodySchema := validators.Object(map[string]interface{}{
			"name":  validators.String().Min(1).Max(100).Optional(),
			"email": validators.String().Pattern(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).Optional(),
			"bio":   validators.String().Max(500).Optional(),
		}).Required()

		responseSchema := validators.Object(map[string]interface{}{
			"id":        validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
			"name":      validators.String().Required(),
			"email":     validators.String().Required(),
			"bio":       validators.String().Optional(),
			"updatedAt": validators.String().Required(),
		}).Required()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = operations.NewSimple().
				PATCH("/users/{id}").
				Summary("Update user").
				Description("Partially update a user's information").
				WithParams(paramsSchema).
				WithQuery(querySchema).
				WithBody(bodySchema).
				WithResponse(responseSchema)
		}
	})
}

// BenchmarkRouterDispatch tests router performance
func BenchmarkRouterDispatch(b *testing.B) {
	// Simple router implementation for benchmarking
	routes := make(map[string]http.HandlerFunc)

	// Register test routes
	routePaths := []struct {
		method string
		path   string
	}{
		{"GET", "/users"},
		{"POST", "/users"},
		{"GET", "/users/{id}"},
		{"PUT", "/users/{id}"},
		{"DELETE", "/users/{id}"},
		{"GET", "/posts"},
		{"POST", "/posts"},
		{"GET", "/posts/{id}"},
		{"GET", "/posts/{id}/comments"},
		{"POST", "/posts/{id}/comments"},
	}

	for _, r := range routePaths {
		key := r.method + " " + r.path
		routes[key] = func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}
	}

	// Simple route finder
	find := func(method, path string) (http.HandlerFunc, map[string]string, bool) {
		// Try exact match
		key := method + " " + path
		if handler, ok := routes[key]; ok {
			return handler, nil, true
		}

		// Try parameterized matches
		if method == "GET" && len(path) > 7 {
			if path[:7] == "/users/" {
				if handler, ok := routes["GET /users/{id}"]; ok {
					return handler, map[string]string{"id": path[7:]}, true
				}
			} else if path[:7] == "/posts/" {
				parts := strings.Split(path[7:], "/")
				if len(parts) == 1 {
					if handler, ok := routes["GET /posts/{id}"]; ok {
						return handler, map[string]string{"id": parts[0]}, true
					}
				} else if len(parts) == 2 && parts[1] == "comments" {
					if handler, ok := routes["GET /posts/{id}/comments"]; ok {
						return handler, map[string]string{"id": parts[0]}, true
					}
				}
			}
		}

		return nil, nil, false
	}

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{"Static_Route", "GET", "/users"},
		{"Parameterized_Route", "GET", "/users/123e4567-e89b-12d3-a456-426614174000"},
		{"Nested_Route", "GET", "/posts/456/comments"},
		{"Non_Existent_Route", "GET", "/non/existent/path"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = find(tc.method, tc.path)
			}
		})
	}
}

// BenchmarkHandlerValidation tests validation overhead in handlers
func BenchmarkHandlerValidation(b *testing.B) {
	// Setup schemas
	paramsSchema := validators.Object(map[string]interface{}{
		"id": validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
	}).Required()

	bodySchema := validators.Object(map[string]interface{}{
		"name":  validators.String().Min(1).Max(100).Required(),
		"email": validators.Email(),
		"age":   validators.Number().Min(18).Max(120).Optional(),
	}).Required()

	// Test payloads defined inline where needed
	validResponse := `{"id":"123e4567-e89b-12d3-a456-426614174000","name":"John Doe","email":"john@example.com"}`

	// Test cases
	b.Run("No_Validation", func(b *testing.B) {
		// Direct handler without validation
		directHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(validResponse))
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/users", strings.NewReader(`{"name":"John Doe","email":"john@example.com"}`))
			directHandler(w, r)
		}
	})

	b.Run("With_Body_Validation", func(b *testing.B) {
		// Handler with manual validation
		validatedHandler := func(w http.ResponseWriter, r *http.Request) {
			// Parse body
			var data map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Validate body
			if err := bodySchema.Validate(data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(validResponse))
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/users", strings.NewReader(`{"name":"John Doe","email":"john@example.com"}`))
			validatedHandler(w, r)
		}
	})

	b.Run("With_Full_Validation", func(b *testing.B) {
		// Handler with full validation
		validatedHandler := func(w http.ResponseWriter, r *http.Request) {
			// Validate params
			params := map[string]interface{}{"id": r.PathValue("id")}
			if err := paramsSchema.Validate(params); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Parse and validate body
			var data map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if err := bodySchema.Validate(data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(validResponse))
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/users/123e4567-e89b-12d3-a456-426614174000", strings.NewReader(`{"name":"John Doe","email":"john@example.com"}`))
			r.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
			validatedHandler(w, r)
		}
	})
}

// BenchmarkRequestResponseValidation tests validation performance for different payload sizes
func BenchmarkRequestResponseValidation(b *testing.B) {
	// Small payload
	b.Run("Small_Payload", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"id":   validators.String().Required(),
			"name": validators.String().Required(),
		}).Required()

		data := map[string]interface{}{
			"id":   "123",
			"name": "Test",
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(data)
		}
	})

	// Medium payload
	b.Run("Medium_Payload", func(b *testing.B) {
		schema := validators.Object(map[string]interface{}{
			"id":          validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
			"name":        validators.String().Min(1).Max(100).Required(),
			"email":       validators.Email(),
			"age":         validators.Number().Min(18).Max(120).Optional(),
			"bio":         validators.String().Max(500).Optional(),
			"tags":        validators.Array(validators.String()).Optional(),
			"preferences": validators.Object(nil).Optional(),
		}).Required()

		data := map[string]interface{}{
			"id":    "123e4567-e89b-12d3-a456-426614174000",
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
			"bio":   "Software developer",
			"tags":  []interface{}{"golang", "backend", "api"},
			"preferences": map[string]interface{}{
				"theme":         "dark",
				"notifications": true,
			},
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(data)
		}
	})

	// Large payload
	b.Run("Large_Payload", func(b *testing.B) {
		// Create schema for a complex order
		schema := validators.Object(map[string]interface{}{
			"orderId": validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
			"customer": validators.Object(map[string]interface{}{
				"id":    validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
				"name":  validators.String().Required(),
				"email": validators.Email(),
			}).Required(),
			"items": validators.Array(validators.Object(map[string]interface{}{
				"productId": validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
				"name":      validators.String().Required(),
				"quantity":  validators.Number().Min(1).Required(),
				"price":     validators.Number().Min(0).Required(),
			})).Required(),
			"shipping": validators.Object(map[string]interface{}{
				"address": validators.String().Required(),
				"city":    validators.String().Required(),
				"country": validators.String().Required(),
				"zip":     validators.String().Required(),
			}).Required(),
			"total": validators.Number().Min(0).Required(),
		}).Required()

		// Create large test data
		items := make([]interface{}, 50)
		for i := 0; i < 50; i++ {
			items[i] = map[string]interface{}{
				"productId": fmt.Sprintf("prod-%d", i),
				"name":      fmt.Sprintf("Product %d", i),
				"quantity":  float64(i%5 + 1),
				"price":     float64(i*10 + 99),
			}
		}

		data := map[string]interface{}{
			"orderId": "123e4567-e89b-12d3-a456-426614174000",
			"customer": map[string]interface{}{
				"id":    "cust-123",
				"name":  "John Doe",
				"email": "john@example.com",
			},
			"items": items,
			"shipping": map[string]interface{}{
				"address": "123 Main St",
				"city":    "New York",
				"country": "USA",
				"zip":     "10001",
			},
			"total": 12345.67,
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(data)
		}
	})
}

// BenchmarkHTTPMethods tests performance across different HTTP methods
func BenchmarkHTTPMethods(b *testing.B) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	for _, method := range methods {
		b.Run(method, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				op := operations.NewSimple()
				switch method {
				case "GET":
					op.GET("/test")
				case "POST":
					op.POST("/test")
				case "PUT":
					op.PUT("/test")
				case "PATCH":
					op.PATCH("/test")
				case "DELETE":
					op.DELETE("/test")
				}
				_ = op
			}
		})
	}
}
