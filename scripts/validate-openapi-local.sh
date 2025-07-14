#!/bin/bash

# validate-openapi-local.sh - Local OpenAPI validation using Redocly CLI
# This script provides local validation equivalent to CI pipeline

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Print functions
print_step() { echo -e "${BLUE}==== $1 ====${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLI_BINARY="$PROJECT_ROOT/go-op-cli"
SPECS_DIR="$PROJECT_ROOT/generated-specs"
DOCS_DIR="$PROJECT_ROOT/docs"

# Change to project root
cd "$PROJECT_ROOT"

print_step "Local OpenAPI Validation with Redocly CLI"

# Check if Node.js is available
if ! command -v node &> /dev/null; then
    print_error "Node.js is required but not installed. Please install Node.js first."
    exit 1
fi

# Install Redocly CLI if not present
print_step "Checking Redocly CLI Installation"
if ! command -v redocly &> /dev/null; then
    print_warning "Redocly CLI not found. Installing globally..."
    if npm install -g @redocly/cli; then
        print_success "Redocly CLI installed successfully"
    else
        print_error "Failed to install Redocly CLI"
        exit 1
    fi
else
    print_success "Redocly CLI already available"
fi

# Build CLI tool
print_step "Building go-op CLI Tool"
if go build -o go-op-cli ./cmd/goop; then
    print_success "CLI tool built successfully"
else
    print_error "Failed to build CLI tool"
    exit 1
fi

# Create output directories
mkdir -p "$SPECS_DIR"
mkdir -p "$DOCS_DIR"

# Generate individual service specs
print_step "Generating Individual Service OpenAPI Specs"

services=("user-service" "order-service" "notification-service")
service_titles=("User Service API" "Order Service API" "Notification Service API")
service_versions=("1.0.0" "1.2.0" "2.1.0")

for i in "${!services[@]}"; do
    service="${services[$i]}"
    title="${service_titles[$i]}"
    version="${service_versions[$i]}"
    
    print_step "Generating $service spec"
    if ./go-op-cli generate \
        -i "./examples/$service" \
        -o "$SPECS_DIR/$service.yaml" \
        -t "$title" \
        -V "$version" \
        -d "Generated API specification for $service" \
        -v; then
        lines=$(wc -l < "$SPECS_DIR/$service.yaml")
        print_success "$service spec generated ($lines lines)"
    else
        print_error "Failed to generate $service spec"
        exit 1
    fi
done

# Combine service specs
print_step "Combining Service Specifications"
if ./go-op-cli combine \
    -o "$SPECS_DIR/combined-platform.yaml" \
    -t "Platform API" \
    -V "3.0.0" \
    -b "/api/v1" \
    "$SPECS_DIR/user-service.yaml" \
    "$SPECS_DIR/order-service.yaml" \
    "$SPECS_DIR/notification-service.yaml" \
    -v; then
    lines=$(wc -l < "$SPECS_DIR/combined-platform.yaml")
    print_success "Combined platform spec generated ($lines lines)"
else
    print_error "Failed to combine service specs"
    exit 1
fi

# Validate individual specs with Redocly
print_step "Validating Individual Service Specifications"
validation_failed=false

for service in "${services[@]}"; do
    spec_file="$SPECS_DIR/$service.yaml"
    print_step "Validating $service.yaml"
    
    if redocly lint "$spec_file" --config=redocly.yaml --format=codeframe; then
        print_success "$service.yaml is valid"
    else
        print_error "$service.yaml has validation errors"
        validation_failed=true
    fi
done

# Validate combined spec
print_step "Validating Combined Platform Specification"
if redocly lint "$SPECS_DIR/combined-platform.yaml" --config=redocly.yaml --format=codeframe; then
    print_success "combined-platform.yaml is valid"
else
    print_error "combined-platform.yaml has validation errors"
    validation_failed=true
fi

# Exit if validation failed
if [ "$validation_failed" = true ]; then
    print_error "OpenAPI validation failed. Please fix the errors above."
    exit 1
fi

# Generate documentation
print_step "Generating API Documentation"
if redocly build-docs "$SPECS_DIR/combined-platform.yaml" --output "$DOCS_DIR/api-docs.html"; then
    print_success "API documentation generated at $DOCS_DIR/api-docs.html"
else
    print_warning "Failed to generate documentation (non-critical)"
fi

# Run additional checks
print_step "Running Additional Quality Checks"

# Check for common issues
print_step "Checking for Common Issues"
for spec in "$SPECS_DIR"/*.yaml; do
    basename=$(basename "$spec")
    
    # Check file size
    size=$(du -h "$spec" | cut -f1)
    lines=$(wc -l < "$spec")
    
    if [ "$lines" -gt 5000 ]; then
        print_warning "$basename is large ($lines lines) - consider splitting"
    else
        print_success "$basename size is reasonable ($size, $lines lines)"
    fi
    
    # Check for required fields
    if grep -q "required: true" "$spec"; then
        print_success "$basename has required field validations"
    else
        print_warning "$basename may be missing required field validations"
    fi
    
    # Check for path parameters
    if grep -q "in: path" "$spec"; then
        # Verify path parameters are marked as required
        if grep -A2 "in: path" "$spec" | grep -q "required: true"; then
            print_success "$basename path parameters are correctly marked as required"
        else
            print_error "$basename path parameters are not marked as required"
            validation_failed=true
        fi
    fi
done

# Generate summary report
print_step "Validation Summary Report"
echo
echo "=== OpenAPI Validation Results ==="
echo

# Individual specs summary
echo "📊 Individual Service Specifications:"
for service in "${services[@]}"; do
    spec_file="$SPECS_DIR/$service.yaml"
    if [ -f "$spec_file" ]; then
        lines=$(wc -l < "$spec_file")
        paths=$(grep -c "^  /" "$spec_file" || echo "0")
        operations=$(grep -E "^ +(get|post|put|patch|delete|head|options):" "$spec_file" | wc -l || echo "0")
        echo "   $service: ✅ $lines lines, $paths paths, $operations operations"
    else
        echo "   $service: ❌ Not generated"
    fi
done

echo
echo "🔗 Combined Platform Specification:"
if [ -f "$SPECS_DIR/combined-platform.yaml" ]; then
    lines=$(wc -l < "$SPECS_DIR/combined-platform.yaml")
    paths=$(grep -c "^  /" "$SPECS_DIR/combined-platform.yaml" || echo "0")
    operations=$(grep -E "^ +(get|post|put|patch|delete|head|options):" "$SPECS_DIR/combined-platform.yaml" | wc -l || echo "0")
    echo "   Platform API: ✅ $lines lines, $paths paths, $operations operations"
else
    echo "   Platform API: ❌ Not generated"
fi

echo
echo "📁 Generated Files:"
echo "   Specifications: $SPECS_DIR/"
echo "   Documentation: $DOCS_DIR/"
echo "   CLI Tool: $CLI_BINARY"

# Performance metrics
echo
echo "⏱️  Performance Metrics:"
echo "   Generated specs: $(ls -1 "$SPECS_DIR"/*.yaml | wc -l) files"
echo "   Documentation: $([ -f "$DOCS_DIR/api-docs.html" ] && echo "✅ Generated" || echo "❌ Failed")"
echo "   Total validation time: $(date '+%Y-%m-%d %H:%M:%S')"

if [ "$validation_failed" = true ]; then
    echo
    print_error "Validation completed with errors. Please review the issues above."
    exit 1
else
    echo
    print_success "All OpenAPI specifications are valid! 🎉"
fi

echo
echo "💡 Next Steps:"
echo "   1. View generated specs: ls -la $SPECS_DIR/"
echo "   2. Open documentation: open $DOCS_DIR/api-docs.html"
echo "   3. Run CI validation: git push (triggers GitHub Actions)"
echo "   4. Import to API tools: Postman, Insomnia, API Gateway"
echo "   5. Generate client SDKs: openapi-generator"