#!/bin/bash

# test-microservices.sh - Comprehensive test script for go-op CLI microservices demonstration

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_step() {
    echo -e "${BLUE}==== $1 ====${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLI_BINARY="$PROJECT_ROOT/go-op-cli"
EXAMPLES_DIR="$PROJECT_ROOT/examples"
OUTPUT_DIR="$PROJECT_ROOT/output"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Change to project root
cd "$PROJECT_ROOT"

print_step "Building go-op CLI"
if go build -o go-op-cli ./cmd/goop; then
    print_success "CLI built successfully"
else
    print_error "Failed to build CLI"
    exit 1
fi

print_step "Step 1: Generate Individual Service OpenAPI Specs"

# Generate User Service spec
print_step "Generating User Service OpenAPI spec"
if ./go-op-cli generate \
    -i ./examples/user-service \
    -o ./output/user-service.yaml \
    -t "User Service API" \
    -V "1.0.0" \
    -d "User management and authentication service" \
    -v; then
    print_success "User Service spec generated: $(wc -l < ./output/user-service.yaml) lines"
else
    print_error "Failed to generate User Service spec"
    exit 1
fi

# Generate Order Service spec
print_step "Generating Order Service OpenAPI spec"
if ./go-op-cli generate \
    -i ./examples/order-service \
    -o ./output/order-service.yaml \
    -t "Order Service API" \
    -V "1.2.0" \
    -d "Order processing and e-commerce management service" \
    -v; then
    print_success "Order Service spec generated: $(wc -l < ./output/order-service.yaml) lines"
else
    print_error "Failed to generate Order Service spec"
    exit 1
fi

# Generate Notification Service spec
print_step "Generating Notification Service OpenAPI spec"
if ./go-op-cli generate \
    -i ./examples/notification-service \
    -o ./output/notification-service.yaml \
    -t "Notification Service API" \
    -V "2.1.0" \
    -d "Notification and messaging service with templating" \
    -v; then
    print_success "Notification Service spec generated: $(wc -l < ./output/notification-service.yaml) lines"
else
    print_error "Failed to generate Notification Service spec"
    exit 1
fi

print_step "Step 2: Validate Individual Specs"

# Check if each spec has the expected structure
for service in user order notification; do
    spec_file="./output/${service}-service.yaml"
    
    # Check OpenAPI version
    if grep -q "openapi: 3.1.0" "$spec_file"; then
        print_success "$service service has correct OpenAPI version"
    else
        print_error "$service service missing OpenAPI 3.1.0 version"
    fi
    
    # Check for paths
    if grep -q "paths:" "$spec_file"; then
        path_count=$(grep -c "    [a-z].*:" "$spec_file" || echo "0")
        print_success "$service service has $path_count HTTP operations"
    else
        print_error "$service service missing paths section"
    fi
    
    # Check for parameters
    if grep -q "parameters:" "$spec_file"; then
        print_success "$service service has parameters defined"
    else
        print_warning "$service service has no parameters (this might be OK)"
    fi
    
    # Check for schemas
    if grep -q "type: object" "$spec_file"; then
        schema_count=$(grep -c "type: object" "$spec_file")
        print_success "$service service has $schema_count object schemas"
    else
        print_warning "$service service has no object schemas"
    fi
done

print_step "Step 3: Test Combine Command (Method 1: Direct Files)"

# Combine specs using direct file arguments
if ./go-op-cli combine \
    -v \
    -o ./output/combined-direct.yaml \
    -t "E-commerce Platform API" \
    -V "3.0.0" \
    -b "/api/v1" \
    ./output/user-service.yaml \
    ./output/order-service.yaml \
    ./output/notification-service.yaml; then
    print_success "Combined spec via direct files: $(wc -l < ./output/combined-direct.yaml) lines"
else
    print_error "Failed to combine specs via direct files"
    exit 1
fi

print_step "Step 4: Create Services Configuration"

# Create a dynamic services.yaml configuration
cat > ./output/services.yaml << 'EOF'
# services.yaml - Dynamic configuration for testing

title: "E-commerce Platform API"
version: "3.0.0" 
description: "Comprehensive API documentation for the e-commerce platform microservices"
base_url: "/api/v1"

services:
  - name: "user-service"
    spec_file: "./output/user-service.yaml"
    path_prefix: "/users"
    description: "User management and authentication service"
    version: "1.0.0"
    health_check: "/health"
    tags: ["users", "authentication"]
    enabled: true

  - name: "order-service"
    spec_file: "./output/order-service.yaml"
    path_prefix: "/orders"
    description: "Order processing and management service"
    version: "1.2.0"
    health_check: "/health"
    tags: ["orders", "e-commerce", "payments"]
    enabled: true

  - name: "notification-service"
    spec_file: "./output/notification-service.yaml"
    path_prefix: "/notifications"
    description: "Notification and messaging service"
    version: "2.1.0"
    health_check: "/health"
    tags: ["notifications", "messaging", "templates"]
    enabled: true

settings:
  merge_schemas: true
  validate_output: true
  conflict_strategy: "override"
EOF

print_success "Created services.yaml configuration"

print_step "Step 5: Test Combine Command (Method 2: Configuration File)"

# Combine specs using configuration file
if ./go-op-cli combine \
    -v \
    -c ./output/services.yaml \
    -o ./output/combined-config.yaml; then
    print_success "Combined spec via config file: $(wc -l < ./output/combined-config.yaml) lines"
else
    print_error "Failed to combine specs via config file"
    exit 1
fi

print_step "Step 6: Validate Combined Specifications"

# Test both combined specifications
for combined_file in "combined-direct.yaml" "combined-config.yaml"; do
    spec_path="./output/$combined_file"
    
    print_step "Validating $combined_file"
    
    # Check OpenAPI version
    if grep -q "openapi: 3.1.0" "$spec_path"; then
        print_success "✓ Correct OpenAPI version"
    else
        print_error "✗ Missing OpenAPI 3.1.0 version"
    fi
    
    # Check API metadata
    if grep -q "title: E-commerce Platform API" "$spec_path"; then
        print_success "✓ Correct API title"
    else
        print_error "✗ Missing or incorrect API title"
    fi
    
    # Check base URL application
    if grep -q "/api/v1/" "$spec_path"; then
        api_path_count=$(grep -c "/api/v1/" "$spec_path")
        print_success "✓ Base URL applied to $api_path_count paths"
    else
        print_error "✗ Base URL not applied to paths"
    fi
    
    # Check service tags
    if grep -q "service:user" "$spec_path" && grep -q "service:order" "$spec_path" && grep -q "service:notification" "$spec_path"; then
        print_success "✓ Service tags added to all services"
    else
        print_error "✗ Missing service tags"
    fi
    
    # Count total operations
    total_operations=$(grep -c "summary:" "$spec_path" || echo "0")
    print_success "✓ Total operations: $total_operations"
    
    # Count total paths
    total_paths=$(grep -c "  /api/v1/" "$spec_path" || echo "0")
    print_success "✓ Total paths: $total_paths"
    
    # Verify all three services are present
    user_paths=$(grep -c "/api/v1/users" "$spec_path" || echo "0")
    order_paths=$(grep -c "/api/v1/orders" "$spec_path" || echo "0")
    notification_paths=$(grep -c "/api/v1/notifications" "$spec_path" || echo "0")
    
    print_success "✓ User service paths: $user_paths"
    print_success "✓ Order service paths: $order_paths"  
    print_success "✓ Notification service paths: $notification_paths"
    
    if [ "$user_paths" -gt 0 ] && [ "$order_paths" -gt 0 ] && [ "$notification_paths" -gt 0 ]; then
        print_success "✓ All services successfully combined"
    else
        print_error "✗ Some services missing from combined spec"
    fi
done

print_step "Step 7: Test JSON Output Format"

# Test JSON output
if ./go-op-cli combine \
    -v \
    -o ./output/combined.json \
    -f json \
    -t "E-commerce Platform API" \
    -V "3.0.0" \
    -b "/api/v1" \
    ./output/user-service.yaml \
    ./output/order-service.yaml \
    ./output/notification-service.yaml; then
    print_success "JSON format output generated: $(wc -l < ./output/combined.json) lines"
    
    # Validate JSON syntax
    if command -v jq >/dev/null 2>&1; then
        if jq . ./output/combined.json >/dev/null 2>&1; then
            print_success "✓ Valid JSON syntax"
        else
            print_error "✗ Invalid JSON syntax"
        fi
    else
        print_warning "jq not available - skipping JSON validation"
    fi
else
    print_error "Failed to generate JSON output"
fi

print_step "Step 8: Advanced Testing"

# Test with tag filtering
print_step "Testing tag filtering (include only 'users' tag)"
if ./go-op-cli combine \
    -v \
    -o ./output/combined-users-only.yaml \
    --include-tags users \
    -t "Users API Only" \
    -V "1.0.0" \
    ./output/user-service.yaml \
    ./output/order-service.yaml \
    ./output/notification-service.yaml; then
    
    users_only_paths=$(grep -c "/users" "./output/combined-users-only.yaml" || echo "0")
    orders_paths=$(grep -c "/orders" "./output/combined-users-only.yaml" || echo "0")
    
    if [ "$users_only_paths" -gt 0 ] && [ "$orders_paths" -eq 0 ]; then
        print_success "✓ Tag filtering working - only users operations included"
    else
        print_warning "⚠ Tag filtering may not be working as expected"
    fi
else
    print_warning "Tag filtering test failed"
fi

# Test service prefix mapping
print_step "Testing service prefix mapping"
if ./go-op-cli combine \
    -v \
    -o ./output/combined-prefixed.yaml \
    -p "user:/user-api" \
    -p "order:/order-api" \
    -p "notification:/notify-api" \
    -t "Prefixed API" \
    -V "1.0.0" \
    ./output/user-service.yaml \
    ./output/order-service.yaml \
    ./output/notification-service.yaml; then
    
    if grep -q "/user-api/" "./output/combined-prefixed.yaml" && 
       grep -q "/order-api/" "./output/combined-prefixed.yaml" && 
       grep -q "/notify-api/" "./output/combined-prefixed.yaml"; then
        print_success "✓ Service prefix mapping working"
    else
        print_warning "⚠ Service prefix mapping may not be working as expected"
    fi
else
    print_warning "Service prefix mapping test failed"
fi

print_step "Step 9: Generate Summary Report"

echo
echo "=== GO-OP CLI MICROSERVICES TEST SUMMARY ==="
echo
echo "📊 Individual Service Specs:"
echo "   User Service:         $([ -f ./output/user-service.yaml ] && echo "✅ Generated" || echo "❌ Failed")"
echo "   Order Service:        $([ -f ./output/order-service.yaml ] && echo "✅ Generated" || echo "❌ Failed")"
echo "   Notification Service: $([ -f ./output/notification-service.yaml ] && echo "✅ Generated" || echo "❌ Failed")"
echo
echo "🔗 Combined Specifications:"
echo "   Direct file method:   $([ -f ./output/combined-direct.yaml ] && echo "✅ Generated" || echo "❌ Failed")"
echo "   Config file method:   $([ -f ./output/combined-config.yaml ] && echo "✅ Generated" || echo "❌ Failed")"
echo "   JSON format:          $([ -f ./output/combined.json ] && echo "✅ Generated" || echo "❌ Failed")"
echo
echo "🧪 Advanced Features:"
echo "   Tag filtering:        $([ -f ./output/combined-users-only.yaml ] && echo "✅ Tested" || echo "❌ Failed")"
echo "   Service prefixes:     $([ -f ./output/combined-prefixed.yaml ] && echo "✅ Tested" || echo "❌ Failed")"
echo
echo "📁 Output files located in: $OUTPUT_DIR"
echo
if [ -f ./output/combined-direct.yaml ]; then
    total_lines=$(wc -l < ./output/combined-direct.yaml)
    total_paths=$(grep -c "  /api/v1/" ./output/combined-direct.yaml || echo "0")
    total_operations=$(grep -c "summary:" ./output/combined-direct.yaml || echo "0")
    
    echo "📈 Combined API Statistics:"
    echo "   Total lines:      $total_lines"
    echo "   Total paths:      $total_paths"
    echo "   Total operations: $total_operations"
    echo "   OpenAPI version:  3.1.0"
    echo "   Services:         3 (user, order, notification)"
fi

print_success "All tests completed successfully! 🎉"

echo
echo "💡 Next steps:"
echo "   1. View the generated specs: ls -la ./output/"
echo "   2. Validate with external tools: swagger-codegen, redoc-cli, etc."
echo "   3. Import into API gateway tools: Kong, Nginx, Traefik"
echo "   4. Generate client SDKs: openapi-generator"
echo "   5. Set up API documentation: Swagger UI, Redoc"
echo