package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/picogrid/go-op/operations"
	ginadapter "github.com/picogrid/go-op/operations/adapters/gin"
	"github.com/picogrid/go-op/validators"
)

// Product represents a product in an e-commerce system
type Product struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Price      float64                `json:"price"`
	Currency   string                 `json:"currency"`
	Category   string                 `json:"category"`
	Tags       []string               `json:"tags"`
	Attributes map[string]interface{} `json:"attributes"`
	InStock    bool                   `json:"in_stock"`
	StockCount *int                   `json:"stock_count,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	IsActive   bool                   `json:"is_active"`
	Metadata   *ProductMetadata       `json:"metadata,omitempty"`
}

// ProductMetadata represents additional product metadata
type ProductMetadata struct {
	Weight       *float64    `json:"weight,omitempty"`
	Dimensions   *Dimensions `json:"dimensions,omitempty"`
	Manufacturer string      `json:"manufacturer,omitempty"`
	SKU          string      `json:"sku,omitempty"`
}

// Dimensions represents product dimensions
type Dimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"`
}

// CreateProductRequest represents the request body for creating a product
type CreateProductRequest struct {
	Name       string                 `json:"name"`
	Price      float64                `json:"price"`
	Currency   string                 `json:"currency"`
	Category   string                 `json:"category"`
	Tags       []string               `json:"tags,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	StockCount int                    `json:"stock_count"`
	Metadata   *ProductMetadata       `json:"metadata,omitempty"`
}

// UpdateProductRequest represents the request body for updating a product
type UpdateProductRequest struct {
	Name       *string                `json:"name,omitempty"`
	Price      *float64               `json:"price,omitempty"`
	Category   *string                `json:"category,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	StockCount *int                   `json:"stock_count,omitempty"`
	IsActive   *bool                  `json:"is_active,omitempty"`
	Metadata   *ProductMetadata       `json:"metadata,omitempty"`
}

// ProductSearchQuery represents query parameters for searching products
type ProductSearchQuery struct {
	Query     string   `json:"query" form:"query"`
	Category  string   `json:"category" form:"category"`
	MinPrice  *float64 `json:"min_price" form:"min_price"`
	MaxPrice  *float64 `json:"max_price" form:"max_price"`
	Tags      []string `json:"tags" form:"tags"`
	InStock   *bool    `json:"in_stock" form:"in_stock"`
	SortBy    string   `json:"sort_by" form:"sort_by"`
	SortOrder string   `json:"sort_order" form:"sort_order"`
	Page      int      `json:"page" form:"page"`
	PageSize  int      `json:"page_size" form:"page_size"`
}

// ProductListResponse represents the response for listing products
type ProductListResponse struct {
	Products   []Product            `json:"products"`
	TotalCount int                  `json:"total_count"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	HasNext    bool                 `json:"has_next"`
	Facets     *ProductSearchFacets `json:"facets,omitempty"`
}

// ProductSearchFacets represents search facets for filtering
type ProductSearchFacets struct {
	Categories []CategoryFacet `json:"categories"`
	PriceRange PriceRange      `json:"price_range"`
	Tags       []TagFacet      `json:"tags"`
}

// CategoryFacet represents a category facet
type CategoryFacet struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// TagFacet represents a tag facet
type TagFacet struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// PriceRange represents price range facet
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// GetProductParams represents path parameters for getting a product
type GetProductParams struct {
	ID string `json:"id" uri:"id"`
}

// Business logic handlers
func createProductHandler(ctx context.Context, params struct{}, query struct{}, body CreateProductRequest) (Product, error) {
	return Product{
		ID:         "prod_123",
		Name:       body.Name,
		Price:      body.Price,
		Currency:   body.Currency,
		Category:   body.Category,
		Tags:       body.Tags,
		Attributes: body.Attributes,
		InStock:    body.StockCount > 0,
		StockCount: &body.StockCount,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsActive:   true,
		Metadata:   body.Metadata,
	}, nil
}

func getProductHandler(ctx context.Context, params GetProductParams, query struct{}, body struct{}) (Product, error) {
	return Product{
		ID:       params.ID,
		Name:     "Sample Product",
		Price:    29.99,
		Currency: "USD",
		Category: "electronics",
		Tags:     []string{"popular", "featured"},
		Attributes: map[string]interface{}{
			"color": "blue",
			"size":  "medium",
		},
		InStock:    true,
		StockCount: intPtr(100),
		CreatedAt:  time.Now().Add(-24 * time.Hour),
		UpdatedAt:  time.Now(),
		IsActive:   true,
	}, nil
}

func updateProductHandler(ctx context.Context, params GetProductParams, query struct{}, body UpdateProductRequest) (Product, error) {
	product := Product{
		ID:        params.ID,
		Name:      "Sample Product",
		Price:     29.99,
		Currency:  "USD",
		Category:  "electronics",
		Tags:      []string{"popular"},
		InStock:   true,
		IsActive:  true,
		UpdatedAt: time.Now(),
	}

	if body.Name != nil {
		product.Name = *body.Name
	}
	if body.Price != nil {
		product.Price = *body.Price
	}
	if body.Category != nil {
		product.Category = *body.Category
	}
	if body.Tags != nil {
		product.Tags = body.Tags
	}
	if body.IsActive != nil {
		product.IsActive = *body.IsActive
	}

	return product, nil
}

func searchProductsHandler(ctx context.Context, params struct{}, query ProductSearchQuery, body struct{}) (ProductListResponse, error) {
	products := []Product{
		{
			ID:       "prod_1",
			Name:     "Laptop",
			Price:    999.99,
			Currency: "USD",
			Category: "electronics",
			Tags:     []string{"popular", "tech"},
			InStock:  true,
			IsActive: true,
		},
		{
			ID:       "prod_2",
			Name:     "Smartphone",
			Price:    599.99,
			Currency: "USD",
			Category: "electronics",
			Tags:     []string{"featured", "mobile"},
			InStock:  true,
			IsActive: true,
		},
	}

	return ProductListResponse{
		Products:   products,
		TotalCount: 2,
		Page:       query.Page,
		PageSize:   query.PageSize,
		HasNext:    false,
		Facets: &ProductSearchFacets{
			Categories: []CategoryFacet{
				{Name: "electronics", Count: 2},
			},
			PriceRange: PriceRange{Min: 599.99, Max: 999.99},
			Tags: []TagFacet{
				{Name: "popular", Count: 1},
				{Name: "featured", Count: 1},
			},
		},
	}, nil
}

func main() {
	// Create Gin engine
	engine := gin.Default()

	// Create OpenAPI generator with comprehensive metadata
	openAPIGen := operations.NewOpenAPIGenerator("Advanced E-commerce API", "2.1.0")

	// Demonstrate all new OpenAPI 3.1 Fixed Fields features
	openAPIGen.SetDescription("A comprehensive e-commerce API showcasing advanced OpenAPI 3.1 features including schema composition, complex validation, and rich metadata")
	openAPIGen.SetSummary("Advanced E-commerce API with OpenAPI 3.1 Features")
	openAPIGen.SetTermsOfService("https://api.example.com/terms")

	// Enhanced contact information
	openAPIGen.SetContact(&operations.OpenAPIContact{
		Name:  "E-commerce API Team",
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
		Description: "Product management operations with advanced search and filtering",
		ExternalDocs: &operations.OpenAPIExternalDocs{
			Description: "Product API documentation",
			URL:         "https://docs.example.com/products",
		},
	})

	openAPIGen.AddTag(operations.OpenAPITag{
		Name:        "search",
		Description: "Advanced search operations with faceted filtering",
		ExternalDocs: &operations.OpenAPIExternalDocs{
			Description: "Search API documentation",
			URL:         "https://docs.example.com/search",
		},
	})

	// Global external documentation
	openAPIGen.SetExternalDocs(&operations.OpenAPIExternalDocs{
		Description: "Complete API documentation with examples and tutorials",
		URL:         "https://docs.example.com/api",
	})

	// Server configuration with variables
	openAPIGen.AddServer(operations.OpenAPIServer{
		URL:         "https://{environment}.api.example.com/{version}",
		Description: "E-commerce API server",
		Variables: map[string]operations.OpenAPIServerVariable{
			"environment": {
				Default:     "production",
				Enum:        []string{"production", "staging", "development"},
				Description: "API environment",
			},
			"version": {
				Default:     "v2",
				Enum:        []string{"v1", "v2", "v3"},
				Description: "API version",
			},
		},
	})

	// Set JSON Schema dialect
	openAPIGen.SetJsonSchemaDialect("https://json-schema.org/draft/2020-12/schema")

	// Create router
	router := ginadapter.NewGinRouter(engine, openAPIGen)

	// Define schemas using available validator methods

	// Dimensions schema
	dimensionsSchema := validators.Object(map[string]interface{}{
		"length": validators.Number().Min(0).Required(),
		"width":  validators.Number().Min(0).Required(),
		"height": validators.Number().Min(0).Required(),
		"unit":   validators.String().Min(1).Required(),
	}).Required()

	// Product metadata schema
	metadataSchema := validators.Object(map[string]interface{}{
		"weight":       validators.Number().Min(0).Optional(),
		"dimensions":   dimensionsSchema,
		"manufacturer": validators.String().Min(1).Max(100).Optional(),
		"sku":          validators.String().Min(1).Max(50).Pattern("^[A-Z0-9-]+$").Optional(),
	}).Optional()

	// Create product schema
	createProductBodySchema := validators.Object(map[string]interface{}{
		"name":        validators.String().Min(1).Max(200).Required(),
		"price":       validators.Number().Min(0).Required(),
		"currency":    validators.String().Min(3).Max(3).Required(),
		"category":    validators.String().Min(1).Max(50).Pattern("^[a-z0-9-]+$").Required(),
		"tags":        validators.Array(validators.String().Min(1).Max(30)).Optional(),
		"attributes":  validators.Object(map[string]interface{}{}).Optional(),
		"stock_count": validators.Number().Min(0).Required(),
		"metadata":    metadataSchema,
	}).Required()

	// Update product schema
	updateProductBodySchema := validators.Object(map[string]interface{}{
		"name":        validators.String().Min(1).Max(200).Optional(),
		"price":       validators.Number().Min(0).Optional(),
		"category":    validators.String().Min(1).Max(50).Pattern("^[a-z0-9-]+$").Optional(),
		"tags":        validators.Array(validators.String().Min(1).Max(30)).Optional(),
		"attributes":  validators.Object(map[string]interface{}{}).Optional(),
		"stock_count": validators.Number().Min(0).Optional(),
		"is_active":   validators.Bool().Optional(),
		"metadata":    metadataSchema,
	}).Optional()

	// Product path parameters
	productParamsSchema := validators.Object(map[string]interface{}{
		"id": validators.String().Min(1).Pattern("^prod_[a-zA-Z0-9]+$").Required(),
	}).Required()

	// Search query schema
	searchQuerySchema := validators.Object(map[string]interface{}{
		"query":      validators.String().Min(1).Max(200).Optional(),
		"category":   validators.String().Min(1).Max(50).Pattern("^[a-z0-9-]+$").Optional(),
		"min_price":  validators.Number().Min(0).Optional(),
		"max_price":  validators.Number().Min(0).Optional(),
		"tags":       validators.Array(validators.String().Min(1).Max(30)).Optional(),
		"in_stock":   validators.Bool().Optional(),
		"sort_by":    validators.String().Min(1).Optional().Default("created_at"),
		"sort_order": validators.String().Min(1).Optional().Default("desc"),
		"page":       validators.Number().Min(1).Optional().Default(1),
		"page_size":  validators.Number().Min(1).Max(100).Optional().Default(20),
	}).Optional()

	// Product response schema
	productResponseSchema := validators.Object(map[string]interface{}{
		"id":          validators.String().Min(1).Required(),
		"name":        validators.String().Min(1).Max(200).Required(),
		"price":       validators.Number().Min(0).Required(),
		"currency":    validators.String().Min(3).Max(3).Required(),
		"category":    validators.String().Min(1).Max(50).Required(),
		"tags":        validators.Array(validators.String().Min(1).Max(30)).Optional(),
		"attributes":  validators.Object(map[string]interface{}{}).Optional(),
		"in_stock":    validators.Bool().Required(),
		"stock_count": validators.Number().Min(0).Optional(),
		"is_active":   validators.Bool().Required(),
		"created_at":  validators.String().Required(),
		"updated_at":  validators.String().Required(),
		"metadata":    metadataSchema,
	}).Required()

	// Product list response schema
	productListResponseSchema := validators.Object(map[string]interface{}{
		"products":    validators.Array(productResponseSchema).Required(),
		"total_count": validators.Number().Min(0).Required(),
		"page":        validators.Number().Min(1).Required(),
		"page_size":   validators.Number().Min(1).Required(),
		"has_next":    validators.Bool().Required(),
		"facets": validators.Object(map[string]interface{}{
			"categories": validators.Array(validators.Object(map[string]interface{}{
				"name":  validators.String().Min(1).Required(),
				"count": validators.Number().Min(0).Required(),
			})).Required(),
			"price_range": validators.Object(map[string]interface{}{
				"min": validators.Number().Min(0).Required(),
				"max": validators.Number().Min(0).Required(),
			}).Required(),
			"tags": validators.Array(validators.Object(map[string]interface{}{
				"name":  validators.String().Min(1).Required(),
				"count": validators.Number().Min(0).Required(),
			})).Required(),
		}).Optional(),
	}).Required()

	// Define operations with enhanced metadata

	// Create product operation
	createProductOp := operations.NewSimple().
		POST("/products").
		Summary("Create a new product").
		Description("Creates a new product in the catalog with comprehensive validation and metadata support").
		Tags("products").
		WithBody(createProductBodySchema).
		WithResponse(productResponseSchema).
		Handler(ginadapter.CreateValidatedHandler(
			createProductHandler,
			validators.Object(map[string]interface{}{}).Required(),
			validators.Object(map[string]interface{}{}).Required(),
			createProductBodySchema,
			productResponseSchema,
		))

	// Get product operation
	getProductOp := operations.NewSimple().
		GET("/products/{id}").
		Summary("Get product by ID").
		Description("Retrieves detailed information about a specific product").
		Tags("products").
		WithParams(productParamsSchema).
		WithResponse(productResponseSchema).
		Handler(ginadapter.CreateValidatedHandler(
			getProductHandler,
			productParamsSchema,
			validators.Object(map[string]interface{}{}).Required(),
			validators.Object(map[string]interface{}{}).Required(),
			productResponseSchema,
		))

	// Update product operation
	updateProductOp := operations.NewSimple().
		PUT("/products/{id}").
		Summary("Update product").
		Description("Updates an existing product with partial data support").
		Tags("products").
		WithParams(productParamsSchema).
		WithBody(updateProductBodySchema).
		WithResponse(productResponseSchema).
		Handler(ginadapter.CreateValidatedHandler(
			updateProductHandler,
			productParamsSchema,
			validators.Object(map[string]interface{}{}).Required(),
			updateProductBodySchema,
			productResponseSchema,
		))

	// Search products operation
	searchProductsOp := operations.NewSimple().
		GET("/products").
		Summary("Search products").
		Description("Advanced product search with filtering, sorting, and faceted results").
		Tags("products").
		WithQuery(searchQuerySchema).
		WithResponse(productListResponseSchema).
		Handler(ginadapter.CreateValidatedHandler(
			searchProductsHandler,
			validators.Object(map[string]interface{}{}).Required(),
			searchQuerySchema,
			validators.Object(map[string]interface{}{}).Required(),
			productListResponseSchema,
		))

	// Register operations
	router.Register(createProductOp)
	router.Register(getProductOp)
	router.Register(updateProductOp)
	router.Register(searchProductsOp)

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": "2.1.0",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Start server
	fmt.Println("🚀 Advanced E-commerce API starting on :8080")
	fmt.Println("📖 OpenAPI spec: http://localhost:8080/openapi.json")
	fmt.Println("🏥 Health check: http://localhost:8080/health")

	if err := engine.Run(":8080"); err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}
