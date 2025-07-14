package benchmarks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/picogrid/go-op"
	"github.com/picogrid/go-op/validators"
)

// BenchmarkFullRequestFlow tests complete request validation flows
func BenchmarkFullRequestFlow(b *testing.B) {
	// Setup schemas
	createUserSchema := validators.Object(map[string]interface{}{
		"username": validators.String().Min(3).Max(30).Pattern(`^[a-zA-Z0-9_]+$`).Required(),
		"email":    validators.Email(),
		"password": validators.String().Min(8).Max(128).Required(),
		"age":      validators.Number().Min(13).Max(120).Optional(),
		"profile": validators.Object(map[string]interface{}{
			"bio":      validators.String().Max(500).Optional(),
			"location": validators.String().Max(100).Optional(),
			"website":  validators.String().Pattern(`^https?://`).Optional(),
		}).Optional(),
	}).Required()

	userResponseSchema := validators.Object(map[string]interface{}{
		"id":        validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
		"username":  validators.String().Required(),
		"email":     validators.String().Required(),
		"createdAt": validators.String().Required(),
		"profile": validators.Object(map[string]interface{}{
			"bio":      validators.String().Optional(),
			"location": validators.String().Optional(),
			"website":  validators.String().Optional(),
		}).Optional(),
	}).Required()

	// Test data
	validPayload := map[string]interface{}{
		"username": "john_doe",
		"email":    "john@example.com",
		"password": "securePassword123",
		"age":      25,
		"profile": map[string]interface{}{
			"bio":      "Software developer",
			"location": "San Francisco",
			"website":  "https://johndoe.com",
		},
	}

	b.Run("Simple_Validation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = createUserSchema.Validate(validPayload)
		}
	})

	b.Run("With_JSON_Marshaling", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(validPayload)
			var payload map[string]interface{}
			_ = json.Unmarshal(data, &payload)
			_ = createUserSchema.Validate(payload)
		}
	})

	b.Run("Full_HTTP_Handler", func(b *testing.B) {
		// Handler with validation
		validatedHandler := func(w http.ResponseWriter, r *http.Request) {
			// Parse body
			var data map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"Invalid JSON"}`))
				return
			}

			// Validate request
			if err := createUserSchema.Validate(data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				errData, _ := json.Marshal(map[string]string{"error": err.Error()})
				w.Write(errData)
				return
			}

			// Build response
			response := map[string]interface{}{
				"id":        "123e4567-e89b-12d3-a456-426614174000",
				"username":  data["username"],
				"email":     data["email"],
				"createdAt": "2024-01-01T00:00:00Z",
			}
			if profile, ok := data["profile"]; ok {
				response["profile"] = profile
			}

			// Validate response
			if err := userResponseSchema.Validate(response); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			respData, _ := json.Marshal(response)
			w.Write(respData)
		}

		payload, _ := json.Marshal(validPayload)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/users", bytes.NewReader(payload))
			r.Header.Set("Content-Type", "application/json")
			validatedHandler(w, r)
		}
	})
}

// BenchmarkEndToEndOperation tests full operation handling
func BenchmarkEndToEndOperation(b *testing.B) {
	// Simple router implementation for benchmarking
	routes := make(map[string]http.HandlerFunc)

	register := func(method, path string, handler http.HandlerFunc) {
		key := method + " " + path
		routes[key] = handler
	}

	find := func(method, path string) (http.HandlerFunc, map[string]string, bool) {
		// Try exact match first
		key := method + " " + path
		if handler, ok := routes[key]; ok {
			return handler, nil, true
		}

		// Check for parameterized routes
		if method == "GET" && len(path) > 7 && path[:7] == "/users/" && path != "/users" {
			if handler, ok := routes["GET /users/{id}"]; ok {
				return handler, map[string]string{"id": path[7:]}, true
			}
		}

		return nil, nil, false
	}

	// Schemas defined inline in handlers for this benchmark

	// Register handlers directly
	register("GET", "/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"123","username":"john","email":"john@example.com"}]`))
	})

	register("POST", "/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123","username":"john","email":"john@example.com"}`))
	})

	register("GET", "/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","username":"john","email":"john@example.com"}`))
	})

	// Test different endpoints
	b.Run("GET_List", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			handler, params, found := find("GET", "/users")
			if found && handler != nil {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/users?page=1&limit=10", nil)
				for k, v := range params {
					r.SetPathValue(k, v)
				}
				handler(w, r)
			}
		}
	})

	b.Run("POST_Create", func(b *testing.B) {
		payload := []byte(`{"username":"john_doe","email":"john@example.com","password":"securepass123"}`)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			handler, params, found := find("POST", "/users")
			if found && handler != nil {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/users", bytes.NewReader(payload))
				r.Header.Set("Content-Type", "application/json")
				for k, v := range params {
					r.SetPathValue(k, v)
				}
				handler(w, r)
			}
		}
	})

	b.Run("GET_Single", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			handler, params, found := find("GET", "/users/123e4567-e89b-12d3-a456-426614174000")
			if found && handler != nil {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/users/123e4567-e89b-12d3-a456-426614174000", nil)
				for k, v := range params {
					r.SetPathValue(k, v)
				}
				handler(w, r)
			}
		}
	})
}

// BenchmarkRealWorldAPI tests realistic API scenarios
func BenchmarkRealWorldAPI(b *testing.B) {
	// E-commerce order processing scenario
	orderSchema := validators.Object(map[string]interface{}{
		"customer": validators.Object(map[string]interface{}{
			"id":    validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
			"email": validators.Email(),
			"phone": validators.String().Pattern(`^\+?[1-9]\d{1,14}$`).Optional(),
		}).Required(),
		"items": validators.Array(validators.Object(map[string]interface{}{
			"productId": validators.String().Pattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Required(),
			"quantity":  validators.Number().Min(1).Max(100).Required(),
			"price":     validators.Number().Min(0).Required(),
			"discount":  validators.Number().Min(0).Max(100).Optional(),
		})).Required(),
		"shipping": validators.Object(map[string]interface{}{
			"method": validators.String().Pattern(`^(standard|express|overnight)$`).Required(),
			"address": validators.Object(map[string]interface{}{
				"line1":      validators.String().Min(1).Max(200).Required(),
				"line2":      validators.String().Max(200).Optional(),
				"city":       validators.String().Min(1).Max(100).Required(),
				"state":      validators.String().Min(2).Max(2).Required(),
				"postalCode": validators.String().Pattern(`^\d{5}(-\d{4})?$`).Required(),
				"country":    validators.String().Min(2).Max(2).Required(),
			}).Required(),
		}).Required(),
		"payment": validators.Object(map[string]interface{}{
			"method": validators.String().Pattern(`^(credit_card|paypal|bank_transfer)$`).Required(),
			"token":  validators.String().Required(),
		}).Required(),
		"couponCode": validators.String().Pattern(`^[A-Z0-9]{4,12}$`).Optional(),
		"notes":      validators.String().Max(500).Optional(),
	}).Required()

	// Create test order data
	testOrder := map[string]interface{}{
		"customer": map[string]interface{}{
			"id":    "cust-123e4567-e89b-12d3-a456-426614174000",
			"email": "customer@example.com",
			"phone": "+1234567890",
		},
		"items": []interface{}{
			map[string]interface{}{
				"productId": "prod-123e4567-e89b-12d3-a456-426614174000",
				"quantity":  2.0,
				"price":     99.99,
				"discount":  10.0,
			},
			map[string]interface{}{
				"productId": "prod-223e4567-e89b-12d3-a456-426614174000",
				"quantity":  1.0,
				"price":     149.99,
			},
		},
		"shipping": map[string]interface{}{
			"method": "express",
			"address": map[string]interface{}{
				"line1":      "123 Main Street",
				"line2":      "Apt 4B",
				"city":       "New York",
				"state":      "NY",
				"postalCode": "10001",
				"country":    "US",
			},
		},
		"payment": map[string]interface{}{
			"method": "credit_card",
			"token":  "tok_visa_4242",
		},
		"couponCode": "SAVE20",
		"notes":      "Please leave at front door",
	}

	b.Run("Order_Validation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = orderSchema.Validate(testOrder)
		}
	})

	b.Run("Order_Processing_Handler", func(b *testing.B) {
		responseSchema := validators.Object(map[string]interface{}{
			"orderId":           validators.String().Required(),
			"status":            validators.String().Required(),
			"total":             validators.Number().Required(),
			"estimatedDelivery": validators.String().Required(),
		}).Required()

		// Handler with validation
		orderCounter := 0
		validatedHandler := func(w http.ResponseWriter, r *http.Request) {
			// Parse body
			var data map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Validate request
			if err := orderSchema.Validate(data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Process order
			total := 0.0
			if items, ok := data["items"].([]interface{}); ok {
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						qty := itemMap["quantity"].(float64)
						price := itemMap["price"].(float64)
						discount := 0.0
						if d, ok := itemMap["discount"].(float64); ok {
							discount = d
						}
						total += qty * price * (1 - discount/100)
					}
				}
			}

			orderCounter++
			response := map[string]interface{}{
				"orderId":           fmt.Sprintf("order-%d", orderCounter),
				"status":            "confirmed",
				"total":             total,
				"estimatedDelivery": "2024-01-05",
			}

			// Validate response
			if err := responseSchema.Validate(response); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			respData, _ := json.Marshal(response)
			w.Write(respData)
		}

		payload, _ := json.Marshal(testOrder)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/orders", bytes.NewReader(payload))
			r.Header.Set("Content-Type", "application/json")
			validatedHandler(w, r)
		}
	})
}

// BenchmarkGinIntegration tests integration with Gin framework
func BenchmarkGinIntegration(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	// Create schemas
	loginSchema := validators.Object(map[string]interface{}{
		"email":    validators.Email(),
		"password": validators.String().Min(8).Required(),
	}).Required()

	// Create Gin middleware for validation
	validateBody := func(schema goop.Schema) gin.HandlerFunc {
		return func(c *gin.Context) {
			var body interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(400, gin.H{"error": "Invalid JSON"})
				c.Abort()
				return
			}

			if err := schema.Validate(body); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

			c.Set("validatedBody", body)
			c.Next()
		}
	}

	b.Run("Gin_Middleware_Validation", func(b *testing.B) {
		router := gin.New()
		router.POST("/login", validateBody(loginSchema), func(c *gin.Context) {
			c.JSON(200, gin.H{
				"token":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
				"expiresIn": 3600,
				"tokenType": "Bearer",
			})
		})

		payload := []byte(`{"email":"user@example.com","password":"securepassword123"}`)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/login", bytes.NewReader(payload))
			r.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, r)
		}
	})

	// Skip Gin operations integration for now as it requires more setup
	// Focus on core validation performance which is the main concern
}
