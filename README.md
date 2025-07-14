# Go-Op

[![Go CI](https://github.com/picogrid/go-op/actions/workflows/go.yml/badge.svg)](https://github.com/picogrid/go-op/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/picogrid/go-op.svg)](https://pkg.go.dev/github.com/picogrid/go-op)
[![Go Report Card](https://goreportcard.com/badge/github.com/picogrid/go-op)](https://goreportcard.com/report/github.com/picogrid/go-op)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Go Operations & Parsing** - A comprehensive API framework for building type-safe APIs with build-time OpenAPI 3.1 generation. Go-Op combines powerful validation with automatic API documentation generation for microservices.

## Features

- **Build-Time OpenAPI Generation**: Generate OpenAPI 3.1 specs from Go source code using AST analysis
- **Zero Runtime Reflection**: Maximum performance with compile-time validation and schema extraction
- **Type-Safe API Framework**: Built on fluent validation chains with comprehensive error handling
- **Generic Struct Validation**: Compile-time type safety with Go generics - no runtime reflection needed
- **Microservices Ready**: Multi-service spec combination and CLI tools for complex architectures
- **OpenAPI 3.1 Compliant**: Full specification support with JSON Schema Draft 2020-12
- **Gin Integration**: Seamless router integration with automatic validation middleware
- **High Performance**: Optimized validation with zero-allocation paths for simple types
- **Extensible**: Support for future protocols (gRPC, MQTT, etc.) via generator pattern

## Installation

```bash
go get github.com/picogrid/go-op
```

## Quick Start

### 1. Type-Safe API Service (Recommended)

```go
package main

import (
    "context"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/picogrid/go-op/operations"
    "github.com/picogrid/go-op/validators"
)

type CreateUserRequest struct {
    Email    string `json:"email"`
    Username string `json:"username"`
    Age      int    `json:"age"`
}

type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    CreatedAt time.Time `json:"created_at"`
}

func main() {
    engine := gin.Default()
    
    // Create OpenAPI generator
    openAPIGen := operations.NewOpenAPIGenerator("My API", "1.0.0")
    router := operations.NewRouter(engine, openAPIGen)
    
    // Define type-safe validation schemas using generics
    userSchema := validators.StructValidator(func(u *CreateUserRequest) map[string]interface{} {
        return map[string]interface{}{
            "email":    validators.Email(),
            "username": validators.String().Min(3).Max(50).Required(),
            "age":      validators.Number().Min(18).Max(120).Required(),
        }
    })
    
    responseSchema := validators.StructValidator(func(u *User) map[string]interface{} {
        return map[string]interface{}{
            "id":         validators.String().Min(1).Required(),
            "email":      validators.Email(),
            "username":   validators.String().Min(1).Required(),
            "created_at": validators.String().Required(),
        }
    })
    
    // Define operation with automatic validation
    createUser := operations.NewSimple().
        POST("/users").
        Summary("Create a new user").
        Description("Creates a user account with type-safe validation").
        Tags("users").
        WithBody(userSchema).
        WithResponse(responseSchema).
        Handler(operations.CreateValidatedHandler(
            createUserHandler,
            nil,
            nil,
            userSchema,
            responseSchema,
        ))
    
    router.Register(createUser)
    engine.Run(":8080")
}

func createUserHandler(ctx context.Context, params struct{}, query struct{}, body CreateUserRequest) (User, error) {
    // Type-safe business logic - validation is handled automatically
    return User{
        ID:        "usr_123",
        Email:     body.Email,
        Username:  body.Username,
        CreatedAt: time.Now(),
    }, nil
}
```

### 2. Alternative Builder Pattern

```go
// Using the fluent builder pattern for struct validation
userSchema := validators.ForStruct[CreateUserRequest]().
    Field("email", validators.Email()).
    Field("username", validators.String().Min(3).Max(50).Required()).
    Field("age", validators.Number().Min(18).Max(120).Required()).
    Build()

// Type-safe validation with typed results
user, err := validators.ValidateStruct[CreateUserRequest](userSchema, requestData)
if err != nil {
    // Handle validation error
}
// user is now *CreateUserRequest type
```

### 3. Generate OpenAPI Specification

```bash
# Install CLI tool
go install github.com/picogrid/go-op/cmd/goop@latest

# Generate OpenAPI spec from your service
go-op generate \
    -i ./my-service \
    -o ./api-spec.yaml \
    -t "My API" \
    -V "1.0.0" \
    --verbose

# Result: Complete OpenAPI 3.1 spec with all endpoints, schemas, and validation rules
```

## Microservices Architecture

### Multiple Service Generation

```bash
# Generate specs for each microservice
go-op generate -i ./user-service -o ./user-api.yaml -t "User Service" -V "1.0.0"
go-op generate -i ./order-service -o ./order-api.yaml -t "Order Service" -V "1.2.0"
go-op generate -i ./notification-service -o ./notification-api.yaml -t "Notification Service" -V "2.1.0"

# Combine into unified API documentation
go-op combine \
    --output ./combined-api.yaml \
    --title "E-commerce Platform API" \
    --version "3.0.0" \
    --base-url "/api/v1" \
    ./user-api.yaml \
    ./order-api.yaml \
    ./notification-api.yaml
```

### Configuration-Based Combination

Create `services.yaml`:
```yaml
title: "E-commerce Platform API"
version: "3.0.0"
base_url: "/api/v1"

services:
  - name: "user-service"
    spec_file: "./user-api.yaml"
    path_prefix: "/users"
    description: "User management service"
    
  - name: "order-service"  
    spec_file: "./order-api.yaml"
    path_prefix: "/orders"
    description: "Order processing service"
```

```bash
go-op combine -c ./services.yaml -o ./platform-api.yaml
```

## Validation Types

### Type-Safe Struct Validation (Recommended)

```go
type User struct {
    Email    string   `json:"email"`
    Username string   `json:"username"`
    Age      int      `json:"age"`
    Tags     []string `json:"tags"`
}

// Method 1: StructValidator function
userSchema := validators.StructValidator(func(u *User) map[string]interface{} {
    return map[string]interface{}{
        "email":    validators.Email(),
        "username": validators.String().Min(3).Max(50).Required(),
        "age":      validators.Number().Min(18).Max(120).Required(),
        "tags":     validators.Array(validators.String()).Optional(),
    }
})

// Method 2: ForStruct builder pattern
userSchema := validators.ForStruct[User]().
    Field("email", validators.Email()).
    Field("username", validators.String().Min(3).Max(50).Required()).
    Field("age", validators.Number().Min(18).Max(120).Required()).
    Field("tags", validators.Array(validators.String()).Optional()).
    Build()

// Type-safe validation with typed results
user, err := validators.ValidateStruct[User](userSchema, requestData)
if err != nil {
    // Handle validation error
}
// user is now *User type with compile-time safety
```

### Traditional Map-Based Validation

```go
// String Validation
validators.String().
    Min(5).                          // Minimum length
    Max(100).                        // Maximum length  
    Pattern(`^[a-zA-Z0-9]+$`).      // Regex pattern
    Email().                         // Email format
    Required()                       // Non-empty required

// Number Validation
validators.Number().
    Min(0).                          // Minimum value
    Max(100).                        // Maximum value
    Integer().                       // Must be integer
    Required()                       // Required field

// Object Validation
validators.Object(map[string]interface{}{
    "name": validators.String().Required(),
    "age":  validators.Number().Min(0),
    "tags": validators.Array(validators.String()),
}).Required()

// Array Validation
validators.Array(validators.String()).
    Min(1).                         // Minimum array length
    Max(10)                         // Maximum array length
```

## CLI Commands

### Generate Command
```bash
go-op generate [flags]

Flags:
  -i, --input string       Source directory to scan for operations
  -o, --output string      Output file path for OpenAPI spec
  -t, --title string       API title
  -V, --version string     API version
  -d, --description string API description
  -f, --format string      Output format (yaml/json) (default "yaml")
  -v, --verbose           Enable verbose logging
```

### Combine Command
```bash
go-op combine [flags] [spec-files...]

Flags:
  -o, --output string           Output file path
  -c, --config string          Configuration file path
  -t, --title string           Combined API title
  -V, --version string         Combined API version
  -b, --base-url string        Base URL for all paths
  -f, --format string          Output format (yaml/json)
  -p, --service-prefix strings Service prefix mappings (service:prefix)
      --include-tags strings   Include only specific tags
      --exclude-tags strings   Exclude specific tags
  -v, --verbose               Enable verbose logging
```

## Advanced Features

### Type-Safe Path Parameters
```go
type UserParams struct {
    ID string `json:"id"`
}

paramsSchema := validators.StructValidator(func(p *UserParams) map[string]interface{} {
    return map[string]interface{}{
        "id": validators.String().Min(1).Pattern("^usr_[a-zA-Z0-9]+$").Required(),
    }
})

operation := operations.NewSimple().
    GET("/users/{id}").
    WithParams(paramsSchema).
    // Automatic path parameter extraction and type-safe validation
```

### Type-Safe Query Parameters
```go
type UserQuery struct {
    Page     int    `json:"page"`
    PageSize int    `json:"page_size"`
    Search   string `json:"search"`
}

querySchema := validators.StructValidator(func(q *UserQuery) map[string]interface{} {
    return map[string]interface{}{
        "page":      validators.Number().Min(1).Default(1).Optional(),
        "page_size": validators.Number().Min(1).Max(100).Default(20).Optional(),
        "search":    validators.String().Min(1).Max(255).Optional(),
    }
})

operation := operations.NewSimple().
    GET("/users").
    WithQuery(querySchema)
    // Automatic query parameter parsing with typed results
```

### Complex Nested Struct Schemas
```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    ZipCode string `json:"zip_code"`
}

type User struct {
    Email   string   `json:"email"`
    Address Address  `json:"address"`
    Tags    []string `json:"tags"`
}

addressSchema := validators.StructValidator(func(a *Address) map[string]interface{} {
    return map[string]interface{}{
        "street":   validators.String().Min(1).Required(),
        "city":     validators.String().Min(1).Required(),
        "zip_code": validators.String().Pattern(`^\d{5}$`).Required(),
    }
})

userSchema := validators.StructValidator(func(u *User) map[string]interface{} {
    return map[string]interface{}{
        "email":   validators.Email(),
        "address": addressSchema,
        "tags":    validators.Array(validators.String()).Optional(),
    }
})
```

### Performance-Optimized Validation Functions
```go
// Create reusable typed validators for better performance
validateUser := validators.TypedValidator[User](userSchema)
validateQuery := validators.TypedValidator[UserQuery](querySchema)

// Use in handlers
user, err := validateUser(requestData)
query, err := validateQuery(queryParams)
```

## OpenAPI 3.1 Fixed Fields Support

Go-Op provides comprehensive support for all OpenAPI 3.1 Fixed Fields, enabling rich API documentation with advanced validation and metadata.

### Enhanced API Metadata

Configure comprehensive API information with contact details, licensing, and external documentation:

```go
openAPIGen := operations.NewOpenAPIGenerator("E-commerce API", "2.1.0")

// Enhanced API information
openAPIGen.SetDescription("A comprehensive e-commerce API with advanced features")
openAPIGen.SetSummary("E-commerce Platform API")
openAPIGen.SetTermsOfService("https://api.example.com/terms")

// Contact information
openAPIGen.SetContact(&operations.OpenAPIContact{
    Name:  "API Support Team",
    Email: "api-support@example.com",
    URL:   "https://api.example.com/support",
})

// License information
openAPIGen.SetLicense(&operations.OpenAPILicense{
    Name: "Apache 2.0",
    URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
})

// Global tags with external documentation
openAPIGen.AddTag(operations.OpenAPITag{
    Name:        "products",
    Description: "Product management operations",
    ExternalDocs: &operations.OpenAPIExternalDocs{
        Description: "Product API documentation",
        URL:         "https://docs.example.com/products",
    },
})

// Global external documentation
openAPIGen.SetExternalDocs(&operations.OpenAPIExternalDocs{
    Description: "Complete API documentation",
    URL:         "https://docs.example.com/api",
})
```

### Server Configuration with Variables

Define flexible server configurations with environment variables:

```go
openAPIGen.AddServer(operations.OpenAPIServer{
    URL:         "https://{environment}.api.example.com/{version}",
    Description: "API server with configurable environment and version",
    Variables: map[string]operations.OpenAPIServerVariable{
        "environment": {
            Default:     "production",
            Enum:        []string{"production", "staging", "development"},
            Description: "API deployment environment",
        },
        "version": {
            Default:     "v2",
            Enum:        []string{"v1", "v2", "v3"},
            Description: "API version",
        },
    },
})
```

### Advanced Schema Validation

Leverage OpenAPI 3.1's extended JSON Schema features:

```go
// Numeric validation with exclusive bounds and multiples
priceSchema := validators.Number().
    Min(0).
    Max(999999.99).
    MultipleOf(0.01). // Enforce cents precision
    Required()

// Array validation with size and uniqueness constraints  
tagsSchema := validators.Array(validators.String().Min(1).Max(30)).
    MinItems(0).
    MaxItems(10).
    UniqueItems(true).
    Optional()

// Object validation with property constraints
attributesSchema := validators.Object(map[string]interface{}{}).
    MinProperties(0).
    MaxProperties(20).
    AdditionalProperties(validators.String().Min(1).Max(100)).
    Optional()
```

### Schema Composition

Use advanced composition patterns for flexible schema design:

```go
// Base schema for common fields
baseProductSchema := validators.Object(map[string]interface{}{
    "id":          validators.String().ReadOnly(true).Required(),
    "name":        validators.String().Min(1).Max(200).Required(),
    "created_at":  validators.String().Format("date-time").ReadOnly(true).Required(),
})

// Extended schema for additional fields
extendedProductSchema := validators.Object(map[string]interface{}{
    "tags":        validators.Array(validators.String()).UniqueItems(true).Optional(),
    "attributes":  validators.Object(map[string]interface{}{}).Optional(),
    "is_active":   validators.Bool().Required(),
})

// Combine schemas using allOf composition
productSchema := validators.AllOf([]interface{}{
    baseProductSchema,
    extendedProductSchema,
}).Title("Product").Description("Complete product information")
```

### Metadata and Deprecation

Mark schemas and fields with rich metadata:

```go
legacyFieldSchema := validators.String().
    Deprecated(true).
    Title("Legacy Field").
    Description("This field is deprecated and will be removed in v3.0").
    Optional()

readOnlySchema := validators.String().
    ReadOnly(true).
    Title("System Generated").
    Description("This field is automatically generated by the system").
    Required()

constSchema := validators.String().
    Const("application/json").
    Title("Content Type").
    Description("Fixed content type for this endpoint").
    Required()
```

### Enhanced Operation Metadata

Enrich your API operations with comprehensive documentation:

```go
createProductOp := operations.NewSimple().
    POST("/products").
    Summary("Create a new product").
    Description("Creates a new product with comprehensive validation").
    OperationId("createProduct").
    Tags("products").
    ExternalDocs(&operations.OpenAPIExternalDocs{
        Description: "Product creation guide",
        URL:         "https://docs.example.com/products/create",
    }).
    Deprecated(false).
    WithBody(productCreateSchema).
    WithResponse(productResponseSchema)
```

### Advanced Response Features

Define rich response schemas with headers and links:

```go
// Response with custom headers
productResponse := operations.OpenAPIResponse{
    Description: "Product created successfully",
    Headers: map[string]operations.OpenAPIHeader{
        "X-Rate-Limit": {
            Description: "Requests remaining in current window",
            Schema:      &goop.OpenAPISchema{Type: "integer"},
            Example:     100,
        },
        "Location": {
            Description: "URL of the created product",
            Schema:      &goop.OpenAPISchema{Type: "string", Format: "uri"},
            Example:     "/products/prod_123",
        },
    },
    Links: map[string]operations.OpenAPILink{
        "GetProduct": {
            OperationId: "getProduct",
            Parameters: map[string]interface{}{
                "id": "$response.body#/id",
            },
            Description: "Link to retrieve the created product",
        },
    },
}
```

### Enhanced Parameter Features

Define rich parameter schemas with comprehensive validation:

```go
// Query parameter with advanced validation
searchQuerySchema := validators.Object(map[string]interface{}{
    "query": validators.String().
        Min(1).
        Max(200).
        Optional(),
    "category": validators.String().
        Min(1).
        Max(50).
        Pattern("^[a-z0-9-]+$").
        Optional(),
    "price_min": validators.Number().
        Min(0).
        Optional(),
    "price_max": validators.Number().
        Min(0).
        Optional(),
    "page": validators.Number().
        Min(1).
        Default(1).
        Optional(),
    "limit": validators.Number().
        Min(1).
        Max(100).
        Default(20).
        Optional(),
}).Optional()

// Path parameter with pattern validation
productParamsSchema := validators.Object(map[string]interface{}{
    "id": validators.String().
        Min(1).
        Pattern("^prod_[a-zA-Z0-9]+$").
        Required(),
}).Required()
```

### Advanced Response Schemas

Create comprehensive response schemas with metadata:

```go
// Rich response schema with nested objects
productResponseSchema := validators.Object(map[string]interface{}{
    // Read-only system fields
    "id": validators.String().
        Min(1).
        Required(),
    "created_at": validators.String().
        Required(),
    "updated_at": validators.String().
        Required(),
    
    // User-modifiable fields
    "name": validators.String().
        Min(1).
        Max(200).
        Required(),
    "price": validators.Number().
        Min(0).
        Required(),
    "currency": validators.String().
        Min(3).
        Max(3).
        Required(),
    
    // Optional complex fields
    "tags": validators.Array(validators.String().Min(1).Max(30)).
        Optional(),
    "metadata": validators.Object(map[string]interface{}{
        "weight": validators.Number().Min(0).Optional(),
        "category": validators.String().Min(1).Optional(),
    }).Optional(),
    
    // Status fields
    "is_active": validators.Bool().Required(),
    "in_stock": validators.Bool().Required(),
}).Required()

// Paginated list response
listResponseSchema := validators.Object(map[string]interface{}{
    "items": validators.Array(productResponseSchema).Required(),
    "total_count": validators.Number().Min(0).Required(),
    "page": validators.Number().Min(1).Required(),
    "page_size": validators.Number().Min(1).Required(),
    "has_next": validators.Bool().Required(),
    "has_previous": validators.Bool().Required(),
}).Required()
```

### Real-World Usage Example

Here's how to use these schemas in a complete API operation:

```go
// Create a comprehensive API operation
searchProductsOp := operations.NewSimple().
    GET("/products").
    Summary("Search products").
    Description("Advanced product search with filtering, sorting, and pagination").
    Tags("products").
    WithQuery(searchQuerySchema).
    WithResponse(listResponseSchema).
    Handler(operations.CreateValidatedHandler(
        searchProductsHandler,
        nil, // No path params
        searchQuerySchema,
        nil, // No body
        listResponseSchema,
    ))

// Register the operation
router.Register(searchProductsOp)
```

## Example Microservices

See our [comprehensive examples](./examples/) demonstrating:

- **User Service**: CRUD operations with authentication patterns
- **Order Service**: E-commerce processing with complex nested schemas  
- **Notification Service**: Multi-channel messaging with templates
- **Advanced API**: Full OpenAPI 3.1 Fixed Fields showcase with enhanced metadata, server variables, and complex schema composition

Each example showcases different API patterns:
- Path parameters with validation patterns
- Complex query parameter filtering
- Nested request/response schemas
- Enum validation and constraints
- Optional vs required field handling
- **NEW**: OpenAPI 3.1 advanced features (contact info, licensing, schema composition, metadata)

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Generate API Documentation
on: [push]

jobs:
  generate-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.24
          
      - name: Install go-op CLI
        run: go install github.com/picogrid/go-op/cmd/goop@latest
        
      - name: Generate API specs
        run: |
          go-op generate -i ./user-service -o ./docs/user-api.yaml
          go-op generate -i ./order-service -o ./docs/order-api.yaml
          go-op combine -c ./services.yaml -o ./docs/platform-api.yaml
          
      - name: Deploy to API portal
        run: ./deploy-docs.sh
```

## Development

```bash
# Clone repository
git clone https://github.com/picogrid/go-op.git
cd go-op

# Install dependencies
go mod tidy

# Build CLI tool
go build -o go-op-cli ./cmd/goop

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./benchmarks

# Test microservices demo
./scripts/test-microservices.sh
```

## Performance

- **Zero runtime reflection** - Pure compile-time analysis using Go generics
- **Build-time generation** - No performance impact on running services  
- **Optimized validation** - Zero-allocation paths for simple types
- **AST-based extraction** - Fast and accurate schema generation
- **Type-safe struct validation** - 20x faster than map[string]interface{} validation
- **Concurrent processing** - Parallel validation for large datasets

### Benchmark Results
```
Struct validation:    ~142 ns/op,   192 B/op,    3 allocs/op
Map validation:     ~2,932 ns/op, 6,430 B/op,   78 allocs/op
Performance gain:        20x faster,   33x less memory usage
```

## Key Benefits

1. **Always In Sync**: API docs generated from actual source code
2. **Compile-Time Type Safety**: Full validation with Go generics and zero runtime reflection
3. **Zero Runtime Cost**: All generation happens at build time
4. **Microservice Ready**: Built-in multi-service support
5. **Standards Compliant**: Full OpenAPI 3.1 compatibility
6. **Developer Friendly**: Simple CLI with powerful features
7. **High Performance**: 20x faster struct validation with 33x less memory usage

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Inspiration

Inspired by [Zod](https://github.com/colinhacks/zod) for TypeScript and modern API development practices.

Based off [Zod-Go](https://github.com/aymaneallaoui/zod-go)

---