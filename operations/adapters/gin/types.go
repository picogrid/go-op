package gin

import (
	"github.com/gin-gonic/gin"

	goop "github.com/picogrid/go-op"
)

// GinHandler represents a Gin handler function for maximum performance
// This is what gets registered with the Gin router - no reflection needed
type GinHandler = gin.HandlerFunc

// GinRouter wraps a Gin engine to provide go-op routing functionality
type GinRouter struct {
	engine     *gin.Engine
	generators []goop.Generator
	operations []goop.CompiledOperation
}

// NewGinRouter creates a new Gin-based router with the specified engine and generators
func NewGinRouter(engine *gin.Engine, generators ...goop.Generator) *GinRouter {
	return &GinRouter{
		engine:     engine,
		generators: generators,
		operations: make([]goop.CompiledOperation, 0),
	}
}

// GetEngine returns the underlying Gin engine
func (r *GinRouter) GetEngine() *gin.Engine {
	return r.engine
}
