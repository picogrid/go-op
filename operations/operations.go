package operations

// Operations creates a new operation builder.
// This is the primary entry point for defining API operations.
//
// Example usage:
//
//	getUserOp := operations.NewSimple().
//		GET("/users/{id}").
//		Summary("Get user by ID").
//		WithParams(validators.Object(map[string]goop.EnhancedSchema{
//			"id": validators.String().UUID().Required(),
//		})).
//		WithResponse(validators.Object(map[string]goop.EnhancedSchema{
//			"id": validators.String().UUID(),
//			"name": validators.String().Min(1),
//		})).
//		Handler(handleGetUser)
//
// Deprecated: Use NewSimple() instead for MVP
func Operations() *SimpleOperationBuilder {
	return NewSimple()
}

// Common HTTP status codes for convenience
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusAccepted            = 202
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusMethodNotAllowed    = 405
	StatusConflict            = 409
	StatusUnprocessableEntity = 422
	StatusInternalServerError = 500
	StatusNotImplemented      = 501
	StatusBadGateway          = 502
	StatusServiceUnavailable  = 503
)
