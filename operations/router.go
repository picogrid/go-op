package operations

import (
	"fmt"

	goop "github.com/picogrid/go-op"
)

// Router provides zero-reflection operation registration and handler creation
// This is the core component that enables high-performance API operations
// It is framework-agnostic and works with any HTTP framework through adapters
type Router struct {
	generators []Generator
	operations []CompiledOperation
}

// NewRouter creates a new framework-agnostic router with the specified generators
func NewRouter(generators ...Generator) *Router {
	return &Router{
		generators: generators,
		operations: make([]CompiledOperation, 0),
	}
}

// Register registers a compiled operation with the router
// This method performs zero reflection and maximum performance registration
func (r *Router) Register(op CompiledOperation) error {
	// Store the operation for generator processing
	r.operations = append(r.operations, op)

	// Process with all generators (build-time analysis)
	info := OperationInfo{
		Method:      op.Method,
		Path:        op.Path,
		Summary:     op.Summary,
		Description: op.Description,
		Tags:        op.Tags,
		Security:    op.Security,
		Operation:   &op,
	}

	// Extract validation info if schemas are present and enhanced
	if op.ParamsSchema != nil {
		if enhanced, ok := op.ParamsSchema.(goop.EnhancedSchema); ok {
			info.ParamsInfo = enhanced.GetValidationInfo()
		}
	}
	if op.QuerySchema != nil {
		if enhanced, ok := op.QuerySchema.(goop.EnhancedSchema); ok {
			info.QueryInfo = enhanced.GetValidationInfo()
		}
	}
	if op.BodySchema != nil {
		if enhanced, ok := op.BodySchema.(goop.EnhancedSchema); ok {
			info.BodyInfo = enhanced.GetValidationInfo()
		}
	}
	if op.ResponseSchema != nil {
		if enhanced, ok := op.ResponseSchema.(goop.EnhancedSchema); ok {
			info.ResponseInfo = enhanced.GetValidationInfo()
		}
	}
	if op.HeaderSchema != nil {
		if enhanced, ok := op.HeaderSchema.(goop.EnhancedSchema); ok {
			info.HeaderInfo = enhanced.GetValidationInfo()
		}
	}

	// Process with all generators
	for _, generator := range r.generators {
		if err := generator.Process(info); err != nil {
			return fmt.Errorf("generator processing failed: %w", err)
		}
	}

	return nil
}

// GetOperations returns all registered operations
// Useful for build-time analysis and spec generation
func (r *Router) GetOperations() []CompiledOperation {
	// Return a copy to prevent external modification
	operations := make([]CompiledOperation, len(r.operations))
	copy(operations, r.operations)
	return operations
}
