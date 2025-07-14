# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**go-op** (Go Operations & Parsing) is a comprehensive API framework for building type-safe APIs with build-time OpenAPI 3.1 generation. It combines powerful validation with automatic API documentation generation using Go AST analysis, achieving zero runtime reflection for maximum performance.

**Module name**: `github.com/picogrid/go-op`

## Essential Commands

### CLI Tool Development
```bash
# Build CLI tool
go build -o go-op-cli ./cmd/goop

# Generate OpenAPI spec from service
./go-op-cli generate -i ./examples/user-service -o ./user-api.yaml -t "User API" -V "1.0.0" -v

# Combine multiple service specs
./go-op-cli combine -o ./combined-api.yaml -t "Platform API" -V "3.0.0" -b "/api/v1" ./user-api.yaml ./order-api.yaml

# Test complete microservices workflow
./scripts/test-microservices.sh

# Single test file execution
go test ./validators -run TestStringValidation -v
go test ./operations -run TestNewRouter -v
```

### Development Workflow
```bash
# Setup and daily development
make dev-setup          # Install tools and dependencies
make quick-check        # Fast feedback: fmt + vet + test
make test               # Run basic tests
make test-all           # Run all tests with race detection and coverage
make fmt                # Format code with gofumpt

# Quality assurance
make lint               # Run golangci-lint
make lint-fix           # Run golangci-lint with automatic fixes
make security           # Run security checks (gosec, nancy)
make pre-commit         # Full pre-commit validation
make ci-test           # Simulate CI environment

# Performance monitoring
make benchmark          # Run performance benchmarks
make benchmark-cpu      # CPU profiling benchmarks
make benchmark-mem      # Memory profiling benchmarks
make benchmark-compare  # Save benchmarks for comparison
```

### Direct Go Commands
```bash
go mod tidy                     # Manage dependencies
go test ./...                   # Run all tests
go test -bench=. ./benchmarks   # Run benchmarks
go build ./...                  # Verify compilation
go test -race ./...            # Test for race conditions

# Fix lint issues before committing
go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

# Run specific test packages
go test ./validators -v
go test ./operations -v
go test ./internal/generator -v
```

## Architecture Overview

### Core Design: Build-Time Generation + Zero Reflection

The framework uses AST analysis to extract OpenAPI specifications from Go source code at build time, eliminating all runtime reflection:

```go
// Define schemas using fluent validators
userSchema := validators.Object(map[string]interface{}{
    "email":    validators.Email().Required(),
    "username": validators.String().Min(3).Max(50).Pattern("^[a-zA-Z0-9_]+$").Required(),
    "age":      validators.Number().Min(18).Max(120).Required(),
}).Required()

// Create type-safe operations
operation := operations.NewSimple().
    POST("/users").
    Summary("Create user").
    WithBody(userSchema).
    WithResponse(responseSchema).
    Handler(createUserHandler)

// CLI extracts schemas via AST analysis → OpenAPI 3.1 spec
```

### Key Components

- **`/goop/`**: Core validation and operations framework
  - `schema.go`: Base Schema interface with `Validate(data interface{}) error`
  - `errors.go`: ValidationError with nested error support and JSON serialization
  - `openapi.go`: OpenAPI schema generation interfaces and types
  - `validators/`: All validator implementations with OpenAPI extensions
  - `operations/`: API operation builders, routing, and OpenAPI generation

- **`/cmd/goop/`**: CLI tool for build-time generation
  - `main.go`: Cobra CLI entry point
  - `cmd/generate.go`: Generate OpenAPI specs from Go source
  - `cmd/combine.go`: Combine multiple service specs

- **`/internal/`**: Build-time analysis and generation
  - `generator/`: Go AST analysis and OpenAPI spec generation
  - `combiner/`: Multi-service specification combination

- **`/examples/`**: Comprehensive microservice demonstrations
  - `user-service/`: CRUD operations with authentication patterns
  - `order-service/`: E-commerce with complex nested schemas
  - `notification-service/`: Multi-channel messaging with templates

### Directory Structure
```
goop/                   # Core validation and API framework
├── schema.go          # Base Schema interface, validation errors
├── errors.go          # ValidationError types with JSON serialization
├── openapi.go         # OpenAPI 3.1 schema generation interfaces
├── validators/        # Type-specific validator implementations
│   ├── validators.go  # Entry points for all validator types
│   ├── *_interfaces.go # Type-safe builder interfaces
│   ├── *_impl.go      # Validation logic implementations
│   └── openapi_extensions.go # OpenAPI schema generation
└── operations/        # API operation framework
    ├── types.go       # Core operation types and interfaces
    ├── simple_builder.go # Operation builder with fluent API
    ├── router.go      # Zero-reflection router and handlers
    └── openapi_generator.go # OpenAPI 3.1 spec generation

cmd/goop/              # CLI tool for build-time generation
├── main.go           # Cobra CLI entry point
└── cmd/              # CLI commands
    ├── generate.go   # Generate OpenAPI from Go source
    ├── combine.go    # Combine multiple service specs
    └── root.go       # Root command configuration

internal/             # Build-time analysis (not imported by users)
├── generator/        # Go AST analysis and spec generation
│   ├── generator.go  # Main generation orchestration
│   ├── ast_analyzer.go # Sophisticated AST parsing
│   └── config.go     # Generation configuration
└── combiner/         # Multi-service spec combination
    ├── combiner.go   # Spec merging and path resolution
    └── config.go     # Combination configuration

examples/             # Microservice demonstrations
├── user-service/     # User management with CRUD patterns
├── order-service/    # E-commerce with nested schemas
├── notification-service/ # Multi-channel messaging
└── services.yaml     # Multi-service combination config

scripts/              # Development and testing scripts
├── test-microservices.sh # Comprehensive CLI testing
└── validate-output.sh    # OpenAPI spec validation
```

## Key Architecture Patterns

### 1. Type-Safe Builder Pattern
Prevents invalid method chaining at compile time through interface state management:

```go
// StringBuilder → RequiredStringBuilder | OptionalStringBuilder
schema := validators.String().Min(3).Required()  // ✓ Valid
// schema.Default("value")  // ✗ Compilation error - Required can't have defaults
```

### 2. Schema-First Development
Single source of truth for validation, documentation, and types:

```go
// Schema defines validation rules
userSchema := validators.Object(map[string]interface{}{
    "email": validators.Email().Required(),
    "age":   validators.Number().Min(18).Optional(),
}).Required()

// Same schema used for:
// 1. Runtime validation: userSchema.Validate(data)
// 2. OpenAPI generation: CLI extracts via AST → spec
// 3. Handler validation: automatic request/response validation
```

### 3. Zero-Reflection AST Analysis
Build-time schema extraction for maximum runtime performance:

```go
// CLI scans Go source for validator usage:
// validators.String().Min(3).Required() → OpenAPI schema:
// { "type": "string", "minLength": 3, "required": true }

// No runtime reflection = maximum performance
```

### 4. Generator Pattern
Extensible architecture supporting multiple output formats:

```go
type Generator interface {
    Generate(operations []Operation) ([]byte, error)
}

// Current: OpenAPIGenerator → OpenAPI 3.1 specs
// Future: GRPCGenerator → .proto files
//         AsyncAPIGenerator → AsyncAPI specs
```

### 5. Handler Separation
Pure business logic separated from transport concerns:

```go
// Pure business function - no HTTP knowledge
func createUserHandler(ctx context.Context, params struct{}, query struct{}, body CreateUserRequest) (User, error) {
    return User{ID: "usr_123", Email: body.Email}, nil
}

// Framework handles validation, serialization, HTTP concerns
operation.Handler(operations.CreateValidatedHandler(
    createUserHandler, paramsSchema, querySchema, bodySchema, responseSchema,
))
```

## Important Development Patterns

### CLI Tool Schema Detection
The AST analyzer tracks schema variables and resolves references:

```go
// Variable tracking
createUserBodySchema := validators.Object(map[string]interface{}{
    "email": validators.Email().Required(),
}).Required()

// Reference resolution  
operation := operations.NewSimple().
    WithBody(createUserBodySchema)  // ← CLI resolves this reference
```

### Microservices Workflow
```bash
# 1. Generate individual service specs
./go-op-cli generate -i ./examples/user-service -o ./user-api.yaml
./go-op-cli generate -i ./examples/order-service -o ./order-api.yaml
./go-op-cli generate -i ./examples/notification-service -o ./notification-api.yaml

# 2. Combine into unified platform API
./go-op-cli combine -c ./examples/services.yaml -o ./platform-api.yaml

# 3. Deploy to API gateway, generate SDKs, etc.
```

### Performance Optimization
- Zero-allocation validation paths for simple types
- Object pooling for memory efficiency
- Concurrent validation with `ValidateConcurrently()`
- AST analysis happens at build time, not runtime

## Testing Strategy

- **Unit tests**: Comprehensive coverage in `/tests/` and within packages
- **Integration tests**: Real microservice examples in `/examples/`
- **CLI testing**: End-to-end workflow testing in `/scripts/test-microservices.sh`
- **Performance tests**: Benchmarks in `/benchmarks/` with regression detection
- **Race condition testing**: `make test-race` for concurrent safety
- **AST analysis testing**: Verify schema extraction accuracy
- **OpenAPI validation**: External validation using Redocly CLI
- **Multi-platform builds**: Verify compilation across OS/architecture combinations

### Comprehensive Testing Workflow
```bash
# Full end-to-end testing
./scripts/test-microservices.sh  # Tests complete CLI workflow with 3 services

# The script tests:
# 1. Individual service spec generation
# 2. Spec validation and structure verification
# 3. Multi-service combination (direct files and config)
# 4. JSON output format
# 5. Advanced features (tag filtering, service prefixes)
# 6. Generated spec validation
```

## Development Guidelines

1. **CLI-First Development**: Test OpenAPI generation for any new validator features
2. **Zero Reflection**: Use build-time analysis, never runtime reflection
3. **Type Safety**: Leverage Go's type system to prevent invalid configurations
4. **Schema Evolution**: Ensure backward compatibility in OpenAPI output
5. **Performance**: Benchmark any changes affecting validation hot paths
6. **Microservices**: Test multi-service combination scenarios

## Build System & Development Environment

### Makefile Targets
The comprehensive Makefile provides organized development workflows:

```bash
# Essential development workflow
make dev-setup          # First-time setup: install tools, dependencies
make quick-check        # Fast feedback loop: fmt + vet + test
make pre-commit         # Pre-commit validation: quality + tests + OpenAPI
make ci-test           # Full CI simulation

# Code quality
make fmt               # Format with gofumpt
make lint              # Run golangci-lint
make lint-fix          # Auto-fix linting issues
make security          # Security analysis (gosec + nancy)
make quality-check     # All quality checks combined

# Testing
make test              # Basic test suite
make test-all          # Complete testing: unit + race + coverage
make test-examples     # Test example services
make benchmark         # Performance benchmarks
make benchmark-compare # Benchmark comparison with baseline

# OpenAPI validation
make validate-openapi       # Full validation with external tools
make validate-openapi-quick # Quick Redocly validation

# Maintenance
make clean             # Clean build artifacts
make tidy              # Clean dependencies and format
make deps-update       # Update all dependencies
```

### Required Tools
```bash
# Go tools (installed via make install-tools)
gofumpt                # Enhanced Go formatting
golangci-lint         # Comprehensive linting
gosec                 # Security analysis
nancy                 # Dependency vulnerability scanning

# External tools
npm install -g @redocly/cli  # OpenAPI validation
```

### Adding New Validator Types
1. Create interfaces in `*_interfaces.go` following the builder pattern
2. Implement validation logic in `*_impl.go`
3. Add OpenAPI schema generation in `openapi_extensions.go`
4. Update CLI AST analyzer to handle new validator methods
5. Add comprehensive tests and examples

### Extending CLI Capabilities
1. Follow Cobra command patterns in `/cmd/goop/cmd/`
2. Enhance AST analyzer in `/internal/generator/ast_analyzer.go` for new patterns
3. Update OpenAPI generator for new output requirements
4. Add integration tests in test scripts

## Import Patterns

When working with this codebase, use these import patterns:

```go
// Core framework imports
import (
    "github.com/picogrid/go-op/operations"
    "github.com/picogrid/go-op/validators"
)

// For CLI tool development
import (
    "github.com/picogrid/go-op/internal/generator"
    "github.com/picogrid/go-op/internal/combiner"
)

// For testing with the framework
import (
    "github.com/picogrid/go-op/goop"
)
```

## Common Development Tasks

### Testing a Single Package
```bash
# Test validators only
go test ./validators -v

# Test with race detection
go test ./operations -race -v

# Test specific function
go test ./validators -run TestStringValidation -v
go test ./operations -run TestNewRouter -v

# Run benchmarks
go test -bench=BenchmarkStringValidation ./benchmarks -benchmem
go test -bench=. ./benchmarks -benchmem

# Test examples
make test-examples
make examples-test  # Test compilation only
```

### Debugging CLI Tool
```bash
# Build and test CLI in one command
go build -o go-op-cli ./cmd/goop && ./go-op-cli generate -i ./examples/user-service -o test.yaml -v

# Debug AST analysis
GO_OP_DEBUG=1 ./go-op-cli generate -i ./examples/user-service -o test.yaml -v
```

### Working with Examples
```bash
# Run example services
go run ./examples/user-service/main.go      # Port 8001
go run ./examples/order-service/main.go     # Port 8002
go run ./examples/notification-service/main.go # Port 8003

# Test API endpoints
curl -X GET http://localhost:8001/health
curl -X POST http://localhost:8001/users -H "Content-Type: application/json" -d '{"email":"test@example.com","username":"testuser","first_name":"Test","last_name":"User","age":25,"password":"password123"}'

# OpenAPI validation
make validate-openapi           # Full validation with external tools
make validate-openapi-quick     # Quick validation using Redocly
```

## Code Quality Standards

- **Formatting**: gofumpt (stricter than gofmt)
- **Linting**: golangci-lint with comprehensive rules configured in `.golangci.yml`
- **Security**: gosec static analysis, nancy dependency scanning
- **Testing**: High coverage with race detection
- **Performance**: Benchmark-driven development with regression detection
- **OpenAPI Compliance**: Full OpenAPI 3.1 specification adherence
- **Build System**: Comprehensive Makefile with development workflow targets
- **CI/CD**: GitHub Actions for validation, multi-platform builds

## Troubleshooting

### Common Issues

1. **Import path errors**: Ensure you're using `github.com/picogrid/go-op` in all imports
2. **CLI command not found**: Build the CLI tool first: `go build -o go-op-cli ./cmd/goop`
3. **AST parsing fails**: Check that validator schemas are defined as variables, not inline
4. **Test failures**: Run `go mod tidy` and ensure all dependencies are up to date
5. **Lint errors**: Use format strings in fmt.Errorf: `fmt.Errorf("%s", msg)` instead of `fmt.Errorf(msg)`
6. **OpenAPI validation fails**: Install Redocly CLI with `npm install -g @redocly/cli`
7. **Build tools missing**: Run `make install-tools` to install required development tools
8. **Performance regression**: Use `make benchmark-compare` to compare with baseline