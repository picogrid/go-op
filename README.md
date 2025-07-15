# go-op

[![Go CI](https://github.com/picogrid/go-op/actions/workflows/go.yml/badge.svg)](https://github.com/picogrid/go-op/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/picogrid/go-op.svg)](https://pkg.go.dev/github.com/picogrid/go-op)
[![Go Report Card](https://goreportcard.com/badge/github.com/picogrid/go-op)](https://goreportcard.com/report/github.com/picogrid/go-op)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**go-op** (Go Operations & Parsing) is a comprehensive API framework for building type-safe APIs with build-time OpenAPI 3.1 generation. It combines powerful validation with automatic API documentation generation using Go AST analysis, achieving zero runtime reflection for maximum performance.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Components](#components)
  - [Validation Framework](#validation-framework)
  - [CLI Tool](#cli-tool)
  - [OpenAPI Generation](#openapi-generation)
  - [Framework Integration](#framework-integration)
- [OpenAPI 3.1 Support](#openapi-31-support)
- [Examples](#examples)
- [Performance](#performance)
- [Advanced Usage](#advanced-usage)
- [CLI Reference](#cli-reference)
- [Development](#development)
- [Contributing](#contributing)

---

## Overview

go-op is designed for teams building microservices that need:
- **Type-safe validation** with compile-time guarantees
- **Automatic API documentation** that stays in sync with code
- **High performance** with zero runtime reflection
- **Microservices architecture** support

The framework consists of three main components:
1. **Validation Library** - Type-safe validation with fluent API
2. **CLI Tool** - Build-time OpenAPI spec generation
3. **Web Framework Integration** - Seamless Gin router integration

## Key Features

### 🚀 Build-Time OpenAPI Generation
- Generate OpenAPI 3.1 specs from Go source code using AST analysis
- No runtime overhead - all generation happens at build time
- Automatic schema extraction from validation chains

### ⚡ Zero Runtime Reflection
- Maximum performance with compile-time validation
- Type-safe struct validation using Go generics
- Zero-allocation paths for simple types

### 🛠️ Type-Safe API Framework
- Fluent validation chains with comprehensive error handling
- Generic struct validation with compile-time type safety
- Automatic request/response validation middleware

### 🏗️ Microservices Ready
- Multi-service spec combination and CLI tools
- Configuration-based service composition
- Independent service development and deployment

### 📋 OpenAPI 3.1 Compliant
- Full specification support with JSON Schema Draft 2020-12
- Advanced features: oneOf, allOf, anyOf schemas
- Enhanced metadata, examples, and documentation

### 🔗 Framework Integration
- Seamless Gin router integration
- Automatic validation middleware
- Type-safe handler functions

---

## Architecture

### Core Design Philosophy

go-op follows a **schema-first, build-time generation** approach:

```
Go Source Code → AST Analysis → OpenAPI 3.1 Spec
      ↓              ↓              ↓
  Validation    CLI Tool      Documentation
   Runtime      Build Time     API Portal
```

### Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        go-op Framework                      │
├─────────────────────────────────────────────────────────────┤
│  Application Layer                                          │
│  ┌─────────────────┐  ┌─────────────────┐                 │
│  │   Web Handlers  │  │   CLI Tool      │                 │
│  │   (Your Code)   │  │   (goop)        │                 │
│  └─────────────────┘  └─────────────────┘                 │
├─────────────────────────────────────────────────────────────┤
│  Framework Layer                                            │
│  ┌─────────────────┐  ┌─────────────────┐                 │
│  │   Operations    │  │   AST Analyzer  │                 │
│  │   (API Ops)     │  │   (Generator)   │                 │
│  └─────────────────┘  └─────────────────┘                 │
├─────────────────────────────────────────────────────────────┤
│  Validation Layer                                           │
│  ┌─────────────────┐  ┌─────────────────┐                 │
│  │   Validators    │  │   OpenAPI       │                 │
│  │   (Type-Safe)   │  │   Extensions    │                 │
│  └─────────────────┘  └─────────────────┘                 │
├─────────────────────────────────────────────────────────────┤
│  Core Layer                                                 │
│  ┌─────────────────┐  ┌─────────────────┐                 │
│  │   Schema        │  │   Errors        │                 │
│  │   (Interface)   │  │   (Validation)  │                 │
│  └─────────────────┘  └─────────────────┘                 │
└─────────────────────────────────────────────────────────────┘
```

---

## Installation

### Install the Library

```bash
go get github.com/picogrid/go-op
```

### Install the CLI Tool

```bash
go install github.com/picogrid/go-op/cmd/goop@latest
```

---

## Quick Start

### 1. Create a Type-Safe API Service

```go
package main

import (
    "context"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/picogrid/go-op/operations"
    "github.com/picogrid/go-op/validators"
)

// Define your data structures
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
    
    // Define type-safe validation schemas
    createUserSchema := validators.ForStruct[CreateUserRequest]().
        Field("email", validators.Email()).
        Field("username", validators.String().Min(3).Max(50).Required()).
        Field("age", validators.Number().Min(18).Max(120).Required()).
        Build()
    
    userResponseSchema := validators.ForStruct[User]().
        Field("id", validators.String().Min(1).Required()).
        Field("email", validators.Email()).
        Field("username", validators.String().Min(1).Required()).
        Field("created_at", validators.String().Required()).
        Build()
    
    // Define API operation
    createUser := operations.NewSimple().
        POST("/users").
        Summary("Create a new user").
        Tags("users").
        WithBody(createUserSchema).
        WithResponse(userResponseSchema).
        Handler(operations.CreateValidatedHandler(
            createUserHandler,
            nil, nil, createUserSchema, userResponseSchema,
        ))
    
    router.Register(createUser)
    engine.Run(":8080")
}

// Type-safe handler - validation is automatic
func createUserHandler(ctx context.Context, params struct{}, query struct{}, body CreateUserRequest) (User, error) {
    return User{
        ID:        "usr_123",
        Email:     body.Email,
        Username:  body.Username,
        CreatedAt: time.Now(),
    }, nil
}
```

### 2. Generate OpenAPI Specification

```bash
# Generate OpenAPI spec from your service
goop generate -i ./my-service -o ./api-spec.yaml -t "My API" -V "1.0.0"
```

### 3. Combine Multiple Microservices

```bash
# Generate individual service specs
goop generate -i ./user-service -o ./user-api.yaml
goop generate -i ./order-service -o ./order-api.yaml

# Combine into unified platform API
goop combine -o ./platform-api.yaml -t "Platform API" -V "2.0.0" \
    ./user-api.yaml ./order-api.yaml
```

---

## Components

### Validation Framework

The validation framework provides type-safe validation with a fluent API:

#### String Validation
```go
// Basic string validation
schema := validators.String().
    Min(5).                    // Minimum length
    Max(100).                  // Maximum length  
    Pattern(`^[a-zA-Z0-9]+$`). // Regex pattern
    Required()                 // Non-empty required

// Specialized string validators
emailSchema := validators.Email()
urlSchema := validators.URL()
```

#### Number Validation
```go
schema := validators.Number().
    Min(0).        // Minimum value
    Max(100).      // Maximum value
    Integer().     // Must be integer
    Positive().    // Must be positive
    MultipleOf(5). // Must be multiple of 5
    Required()
```

#### Array Validation
```go
schema := validators.Array(validators.String()).
    MinItems(1).     // Minimum array length
    MaxItems(10).    // Maximum array length
    UniqueItems().   // All items must be unique
    Required()
```

#### Object Validation
```go
schema := validators.Object(map[string]interface{}{
    "name": validators.String().Required(),
    "age":  validators.Number().Min(0),
    "tags": validators.Array(validators.String()),
}).
MinProperties(1).  // Minimum properties
MaxProperties(10). // Maximum properties
Required()
```

#### Type-Safe Struct Validation (Recommended)
```go
type User struct {
    Email    string   `json:"email"`
    Username string   `json:"username"`
    Age      int      `json:"age"`
    Tags     []string `json:"tags"`
}

// Method 1: Builder pattern
userSchema := validators.ForStruct[User]().
    Field("email", validators.Email()).
    Field("username", validators.String().Min(3).Max(50).Required()).
    Field("age", validators.Number().Min(18).Max(120).Required()).
    Field("tags", validators.Array(validators.String()).Optional()).
    Build()

// Type-safe validation with typed results
user, err := validators.ValidateStruct[User](userSchema, requestData)
// user is now *User type with compile-time safety
```

### CLI Tool

The `goop` CLI tool provides build-time OpenAPI spec generation:

#### Generate Command
```bash
goop generate [flags]

Flags:
  -i, --input string       Source directory to scan
  -o, --output string      Output file path for OpenAPI spec
  -t, --title string       API title
  -V, --version string     API version
  -d, --description string API description
  -f, --format string      Output format (yaml/json)
  -v, --verbose           Enable verbose logging
```

#### Combine Command
```bash
goop combine [flags] [spec-files...]

Flags:
  -o, --output string           Output file path
  -c, --config string          Configuration file path
  -t, --title string           Combined API title
  -V, --version string         Combined API version
  -b, --base-url string        Base URL for all paths
  -f, --format string          Output format (yaml/json)
  -p, --service-prefix strings Service prefix mappings
      --include-tags strings   Include only specific tags
      --exclude-tags strings   Exclude specific tags
```

#### Configuration-Based Combination

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
goop combine -c ./services.yaml -o ./platform-api.yaml
```

### OpenAPI Generation

The AST analyzer extracts OpenAPI schemas from Go source code:

#### How It Works
1. **AST Parsing**: Analyzes Go source files for validator usage
2. **Schema Extraction**: Converts validator chains to OpenAPI schemas
3. **Operation Discovery**: Finds API operations and their schemas
4. **Spec Generation**: Produces complete OpenAPI 3.1 specification

#### Supported Patterns
```go
// The CLI detects these patterns automatically:

// Variable definitions
userSchema := validators.String().Min(3).Required()

// Inline usage in operations
operation := operations.NewSimple().
    WithBody(validators.Object(map[string]interface{}{
        "email": validators.Email(),
    }))

// Struct validators
userValidator := validators.ForStruct[User]().
    Field("email", validators.Email()).
    Build()
```

### Framework Integration

#### Gin Integration

go-op provides seamless Gin router integration:

```go
// Standard Gin setup
engine := gin.Default()

// go-op router with OpenAPI generation
openAPIGen := operations.NewOpenAPIGenerator("My API", "1.0.0")
router := operations.NewRouter(engine, openAPIGen)

// Register operations - validation is automatic
router.Register(createUserOp)
router.Register(getUserOp)
router.Register(updateUserOp)

// OpenAPI spec is generated automatically
spec := openAPIGen.Generate()
```

#### Automatic Validation Middleware

The framework provides automatic request/response validation:

```go
// Handler with automatic validation
handler := operations.CreateValidatedHandler(
    businessLogicHandler,  // Your business logic
    paramsSchema,         // Path parameter validation
    querySchema,          // Query parameter validation  
    bodySchema,           // Request body validation
    responseSchema,       // Response validation
)

// The middleware:
// 1. Validates incoming request
// 2. Calls your handler with typed data
// 3. Validates outgoing response
// 4. Returns appropriate errors
```

---

## OpenAPI 3.1 Support

go-op provides comprehensive OpenAPI 3.1 support with full JSON Schema Draft 2020-12 compatibility:

### Supported Features

| Feature | Status | Description |
|---------|--------|-------------|
| **Core OpenAPI 3.1** | ✅ | Full specification compliance |
| **JSON Schema 2020-12** | ✅ | Modern schema validation |
| **Schema Composition** | ✅ | oneOf, allOf, anyOf support |
| **Advanced Validation** | ✅ | Pattern properties, conditionals |
| **Enhanced Metadata** | ✅ | Rich documentation and examples |
| **Fixed Fields** | ✅ | Complete OpenAPI field support |
| **Server Variables** | ✅ | Dynamic server configuration |
| **Links & Callbacks** | ✅ | Advanced API relationships |
| **Content Types** | ✅ | Multiple media type support |
| **Security Schemes** | ✅ | OAuth2, JWT, API keys |

### Schema Composition

```go
// OneOf validation - value must match exactly one schema
statusSchema := validators.OneOf([]interface{}{
    validators.String().Const("pending"),
    validators.String().Const("approved"), 
    validators.String().Const("rejected"),
})

// AllOf validation - value must match all schemas
userSchema := validators.AllOf([]interface{}{
    baseUserSchema,    // Common fields
    extendedUserSchema, // Additional fields
})

// AnyOf validation - value must match at least one schema
searchSchema := validators.AnyOf([]interface{}{
    validators.String().Min(1),           // Text search
    validators.Number().Integer().Min(0), // ID search
})
```

### Enhanced Metadata

```go
openAPIGen := operations.NewOpenAPIGenerator("E-commerce API", "2.1.0")

// Rich API information
openAPIGen.SetDescription("A comprehensive e-commerce API")
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
```

### Server Configuration

```go
openAPIGen.AddServer(operations.OpenAPIServer{
    URL:         "https://{environment}.api.example.com/{version}",
    Description: "API server with configurable environment",
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

### Content Type Support

```go
// Multiple content types for requests
operation := operations.NewSimple().
    POST("/users").
    WithBody(userSchema).
    ContentTypes([]string{
        "application/json",
        "application/xml",
        "multipart/form-data",
    })
```

---

## Examples

The [examples directory](./examples/) contains comprehensive demonstrations:

### User Service
Complete CRUD operations with authentication patterns:
- User registration and login
- Profile management
- Password validation
- JWT token handling

### Order Service  
E-commerce processing with complex nested schemas:
- Order creation and management
- Product catalog integration
- Payment processing
- Order status tracking

### Notification Service
Multi-channel messaging with templates:
- Email notifications
- SMS alerts
- Push notifications
- Template management

### Advanced API
Showcase of OpenAPI 3.1 features:
- Complex schema composition
- Advanced validation patterns
- Rich metadata and documentation
- Server variables and configuration

### Running Examples

```bash
# Start individual services
go run ./examples/user-service/main.go      # Port 8001
go run ./examples/order-service/main.go     # Port 8002
go run ./examples/notification-service/main.go # Port 8003

# Generate API documentation
goop generate -i ./examples/user-service -o ./user-api.yaml
goop generate -i ./examples/order-service -o ./order-api.yaml

# Test API endpoints
curl -X GET http://localhost:8001/health
curl -X POST http://localhost:8001/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","age":25}'
```

---

## Performance

go-op is designed for high performance with minimal overhead:

### Zero Runtime Reflection
- All schema analysis happens at build time
- Runtime validation uses direct field access
- No reflection-based type inspection

### Optimized Validation Paths
- Zero-allocation validation for simple types
- Efficient memory usage with object pooling
- Concurrent validation for large datasets

### Benchmark Results
```
Struct validation:    ~142 ns/op,   192 B/op,    3 allocs/op
Map validation:     ~2,932 ns/op, 6,430 B/op,   78 allocs/op
Performance gain:        20x faster,   33x less memory usage

OpenAPI generation:   ~1.2ms for typical service
AST analysis:         ~850μs for 1000 validators
Spec serialization:   ~350μs for complete spec
```

### Performance Tips

1. **Use Struct Validation**: 20x faster than map-based validation
2. **Reuse Validators**: Create validators once, use many times
3. **Batch Operations**: Use concurrent validation for arrays
4. **Optimize Patterns**: Simple patterns are faster than complex regex

```go
// Fast: Reusable typed validator
validateUser := validators.TypedValidator[User](userSchema)
user, err := validateUser(data) // ~142 ns/op

// Slower: Map-based validation  
err := userSchema.Validate(data) // ~2,932 ns/op
```

---

## Advanced Usage

### Custom Generators

Extend go-op to support additional output formats:

```go
// Implement the Generator interface
type GRPCGenerator struct {
    // Your implementation
}

func (g *GRPCGenerator) Generate(operations []Operation) ([]byte, error) {
    // Generate .proto files instead of OpenAPI
    return protoContent, nil
}

// Use with CLI or programmatically
generator := &GRPCGenerator{}
content, err := generator.Generate(operations)
```

### Custom Validators

Create domain-specific validators:

```go
// Custom validation function
func ValidateISBN(isbn string) error {
    // ISBN validation logic
    if !isValidISBN(isbn) {
        return goop.NewValidationError("isbn", isbn, "invalid ISBN format")
    }
    return nil
}

// Use in schema
bookSchema := validators.Object(map[string]interface{}{
    "isbn": validators.String().Custom(ValidateISBN).Required(),
    "title": validators.String().Min(1).Required(),
})
```

### Middleware Integration

Create custom middleware for advanced features:

```go
// Rate limiting middleware
func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Rate limiting logic
        c.Next()
    }
}

// Authentication middleware
func AuthMiddleware(userSchema goop.Schema) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Authentication and user validation
        c.Next()
    }
}

// Combine with go-op
router.Use(RateLimitMiddleware())
router.Use(AuthMiddleware(userSchema))
router.Register(operation)
```

### Testing Strategies

Comprehensive testing approaches:

```go
// Test validation schemas
func TestUserValidation(t *testing.T) {
    schema := validators.ForStruct[User]().
        Field("email", validators.Email()).
        Build()
    
    validUser := User{Email: "test@example.com"}
    user, err := validators.ValidateStruct[User](schema, validUser)
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
}

// Test OpenAPI generation
func TestOpenAPIGeneration(t *testing.T) {
    generator := operations.NewOpenAPIGenerator("Test API", "1.0.0")
    operation := operations.NewSimple().GET("/test")
    
    spec := generator.Generate()
    assert.Contains(t, spec, "/test")
}

// Integration testing
func TestAPIEndpoint(t *testing.T) {
    router := setupTestRouter()
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/users", nil)
    
    router.ServeHTTP(w, req)
    assert.Equal(t, 200, w.Code)
}
```

---

## CLI Reference

### Generate Command

Generate OpenAPI specifications from Go source code:

```bash
goop generate [flags]
```

**Flags:**
- `-i, --input string`: Source directory to scan for operations
- `-o, --output string`: Output file path for OpenAPI spec
- `-t, --title string`: API title
- `-V, --version string`: API version  
- `-d, --description string`: API description
- `-f, --format string`: Output format (yaml/json), default: yaml
- `-v, --verbose`: Enable verbose logging

**Examples:**
```bash
# Basic generation
goop generate -i ./service -o ./api.yaml -t "My API" -V "1.0.0"

# With description and JSON output
goop generate -i ./service -o ./api.json -t "My API" -V "1.0.0" \
  -d "A comprehensive API" -f json

# Verbose output for debugging
goop generate -i ./service -o ./api.yaml -t "My API" -V "1.0.0" --verbose
```

### Combine Command

Combine multiple OpenAPI specifications:

```bash
goop combine [flags] [spec-files...]
```

**Flags:**
- `-o, --output string`: Output file path
- `-c, --config string`: Configuration file path
- `-t, --title string`: Combined API title
- `-V, --version string`: Combined API version
- `-b, --base-url string`: Base URL for all paths
- `-f, --format string`: Output format (yaml/json), default: yaml
- `-p, --service-prefix strings`: Service prefix mappings (service:prefix)
- `--include-tags strings`: Include only specific tags
- `--exclude-tags strings`: Exclude specific tags
- `-v, --verbose`: Enable verbose logging

**Examples:**
```bash
# Combine multiple specs
goop combine -o ./platform.yaml -t "Platform API" -V "2.0.0" \
  ./user-api.yaml ./order-api.yaml

# With configuration file
goop combine -c ./services.yaml -o ./platform.yaml

# With service prefixes
goop combine -o ./api.yaml -p user-service:/users -p order-service:/orders \
  ./user-api.yaml ./order-api.yaml

# Filter by tags
goop combine -o ./public-api.yaml --include-tags public,external \
  ./internal-api.yaml ./public-api.yaml
```

### Configuration File Format

```yaml
# services.yaml
title: "Platform API"
version: "2.0.0"
description: "Combined microservices API"
base_url: "/api/v2"

# Global settings
contact:
  name: "API Team"
  email: "api@example.com"
  url: "https://example.com/contact"

license:
  name: "MIT"
  url: "https://opensource.org/licenses/MIT"

# Service definitions
services:
  - name: "user-service"
    spec_file: "./user-api.yaml"
    path_prefix: "/users"
    description: "User management operations"
    tags:
      - "users"
      - "authentication"
    
  - name: "order-service"
    spec_file: "./order-api.yaml" 
    path_prefix: "/orders"
    description: "Order processing operations"
    tags:
      - "orders"
      - "payments"

# Filtering options
include_tags:
  - "public"
  - "v2"
  
exclude_tags:
  - "internal"
  - "deprecated"
```

---

## Development

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/picogrid/go-op.git
cd go-op

# Install dependencies
go mod tidy

# Install development tools
make install-tools

# Run tests
make test

# Run benchmarks
make benchmark
```

### Available Make Targets

```bash
# Development workflow
make dev-setup          # First-time setup
make quick-check        # Fast feedback: fmt + vet + test  
make pre-commit         # Full pre-commit validation
make ci-test           # Full CI simulation

# Code quality
make fmt               # Format with gofumpt
make lint              # Run golangci-lint
make lint-fix          # Auto-fix linting issues
make security          # Security analysis
make quality-check     # All quality checks

# Testing
make test              # Basic test suite
make test-all          # Complete testing with race detection
make test-examples     # Test example services
make benchmark         # Performance benchmarks
make benchmark-compare # Compare with baseline

# OpenAPI validation
make validate-openapi       # Full validation
make validate-openapi-quick # Quick Redocly validation

# Maintenance
make clean             # Clean build artifacts
make tidy              # Clean dependencies and format
make deps-update       # Update all dependencies
```

### Building the CLI

```bash
# Build for current platform
go build -o goop ./cmd/goop

# Build for multiple platforms
make build-all

# Install globally
go install ./cmd/goop
```

### Running Tests

```bash
# Unit tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Race detection
go test -race ./...

# Benchmarks
go test -bench=. ./benchmarks

# Integration tests
./scripts/test-microservices.sh
```

### Project Structure

```
go-op/
├── README.md                    # This file
├── CLAUDE.md                   # Claude Code instructions  
├── Makefile                    # Development automation
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
│
├── cmd/goop/                   # CLI tool
│   ├── main.go                 # CLI entry point
│   └── cmd/                    # Command implementations
│       ├── generate.go         # Generate command
│       ├── combine.go          # Combine command
│       └── root.go            # Root command setup
│
├── validators/                 # Validation framework
│   ├── validators.go          # Public API
│   ├── *_interfaces.go        # Type-safe interfaces
│   ├── *_impl.go             # Implementation
│   ├── openapi_extensions.go # OpenAPI schema generation
│   └── struct_builder.go     # Generic struct validation
│
├── operations/                # API operations framework  
│   ├── types.go              # Core types and interfaces
│   ├── simple_builder.go     # Operation builder
│   ├── router.go             # Gin integration
│   └── openapi_generator.go  # OpenAPI generation
│
├── internal/                  # Internal packages
│   ├── generator/            # AST analysis and generation
│   │   ├── generator.go      # Main generator
│   │   ├── ast_analyzer.go   # Go AST analysis
│   │   └── config.go         # Configuration
│   └── combiner/             # Spec combination
│       ├── combiner.go       # Main combiner
│       └── config.go         # Configuration
│
├── examples/                  # Example services
│   ├── user-service/         # User management
│   ├── order-service/        # E-commerce orders
│   ├── notification-service/ # Multi-channel notifications
│   └── services.yaml         # Multi-service config
│
├── benchmarks/               # Performance tests
├── scripts/                  # Development scripts
└── docs/                     # Documentation
```

---

## Contributing

We welcome contributions! Here's how to get started:

### Getting Started

1. **Fork the repository**
2. **Clone your fork**: `git clone https://github.com/your-username/go-op.git`
3. **Create a feature branch**: `git checkout -b feature/amazing-feature`
4. **Set up development environment**: `make dev-setup`

### Development Guidelines

- **Code Style**: Run `make fmt` before committing
- **Testing**: Ensure `make test-all` passes
- **Linting**: Fix issues found by `make lint`
- **Documentation**: Update README for new features
- **Examples**: Add examples for significant features

### Submitting Changes

1. **Commit your changes**: `git commit -m 'Add amazing feature'`
2. **Push to your branch**: `git push origin feature/amazing-feature`
3. **Open a Pull Request**

### Pull Request Requirements

- [ ] Tests pass (`make test-all`)
- [ ] Linting passes (`make lint`)
- [ ] Security checks pass (`make security`)
- [ ] Examples work (`make test-examples`)
- [ ] Documentation updated
- [ ] Changelog entry added

### Reporting Issues

When reporting issues, please include:
- Go version (`go version`)
- go-op version
- Minimal reproduction case
- Expected vs actual behavior
- Relevant error messages

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Inspiration

Inspired by:
- [Zod](https://github.com/colinhacks/zod) for TypeScript validation patterns
- [Zod-Go](https://github.com/aymaneallaoui/zod-go) for Go validation concepts
- Modern API development practices and OpenAPI 3.1 specification

---

## Support

- **Documentation**: [GitHub Wiki](https://github.com/picogrid/go-op/wiki)
- **Issues**: [GitHub Issues](https://github.com/picogrid/go-op/issues)
- **Discussions**: [GitHub Discussions](https://github.com/picogrid/go-op/discussions)
- **Examples**: [Examples Directory](./examples/)
