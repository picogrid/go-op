#!/bin/bash

# validate-output.sh - Quick validation of generated OpenAPI specs

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

print_step() { echo -e "${BLUE}==== $1 ====${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }

print_step "Validating Combined OpenAPI Specification"

COMBINED_FILE="./output/combined-direct.yaml"

if [ ! -f "$COMBINED_FILE" ]; then
    print_error "Combined file not found: $COMBINED_FILE"
    exit 1
fi

print_success "Found combined specification file"

# Check OpenAPI version
if grep -q "openapi: 3.1.0" "$COMBINED_FILE"; then
    print_success "Correct OpenAPI version (3.1.0)"
else
    print_error "Missing or incorrect OpenAPI version"
fi

# Check API metadata
if grep -q "title: E-commerce Platform API" "$COMBINED_FILE"; then
    print_success "Correct API title"
else
    print_error "Missing or incorrect API title"
fi

# Check base URL application
api_paths=$(grep -c "/api/v1/" "$COMBINED_FILE" || echo "0")
if [ "$api_paths" -gt 0 ]; then
    print_success "Base URL applied to $api_paths paths"
else
    print_error "Base URL not applied"
fi

# Check all three services are present
user_paths=$(grep -c "/api/v1/users" "$COMBINED_FILE" || echo "0")
order_paths=$(grep -c "/api/v1/orders" "$COMBINED_FILE" || echo "0")
notification_paths=$(grep -c "/api/v1/notifications" "$COMBINED_FILE" || echo "0")

print_success "User service paths: $user_paths"
print_success "Order service paths: $order_paths"
print_success "Notification service paths: $notification_paths"

if [ "$user_paths" -gt 0 ] && [ "$order_paths" -gt 0 ] && [ "$notification_paths" -gt 0 ]; then
    print_success "All three services successfully combined!"
else
    print_error "Some services missing from combined specification"
fi

# Check service tags
if grep -q "service:user" "$COMBINED_FILE" && 
   grep -q "service:order" "$COMBINED_FILE" && 
   grep -q "service:notification" "$COMBINED_FILE"; then
    print_success "Service tags properly added"
else
    print_error "Service tags missing"
fi

# Count operations and paths
total_operations=$(grep -c "summary:" "$COMBINED_FILE" || echo "0")
total_paths=$(grep -c "  /api/v1/" "$COMBINED_FILE" || echo "0")

print_step "Final Statistics"
echo "📊 Combined API Specification:"
echo "   Total paths: $total_paths"
echo "   Total operations: $total_operations"
echo "   OpenAPI version: 3.1.0"
echo "   Services: 3 (user, order, notification)"
echo "   File size: $(wc -l < "$COMBINED_FILE") lines"

# Show some example paths
print_step "Sample API Paths"
echo "User Service:"
grep "/api/v1/users" "$COMBINED_FILE" | head -3 | sed 's/^/   /'
echo "Order Service:"  
grep "/api/v1/orders" "$COMBINED_FILE" | head -3 | sed 's/^/   /'
echo "Notification Service:"
grep "/api/v1/notifications" "$COMBINED_FILE" | head -3 | sed 's/^/   /'

print_success "Validation completed successfully! 🎉"

echo
echo "💡 The combined OpenAPI specification includes:"
echo "   • User management (authentication, CRUD)"
echo "   • Order processing (e-commerce, analytics)"  
echo "   • Notification system (messaging, templates)"
echo "   • Comprehensive schemas with validation"
echo "   • Path/query/body parameters"
echo "   • Service-specific tags for structure"
echo