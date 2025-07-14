package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/picogrid/go-op/operations"
	"github.com/picogrid/go-op/validators"
)

// Order represents an order in the system
type Order struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	Status          string      `json:"status"`
	Items           []OrderItem `json:"items"`
	TotalAmount     float64     `json:"total_amount"`
	Currency        string      `json:"currency"`
	ShippingAddress Address     `json:"shipping_address"`
	BillingAddress  Address     `json:"billing_address"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Subtotal  float64 `json:"subtotal"`
}

// Address represents a shipping or billing address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// CreateOrderRequest represents the request body for creating an order
type CreateOrderRequest struct {
	UserID          string            `json:"user_id"`
	Items           []CreateOrderItem `json:"items"`
	Currency        string            `json:"currency"`
	ShippingAddress Address           `json:"shipping_address"`
	BillingAddress  Address           `json:"billing_address"`
}

// CreateOrderItem represents an item in a create order request
type CreateOrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// UpdateOrderStatusRequest represents request to update order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

// Order path parameters
type GetOrderParams struct {
	ID string `json:"id" uri:"id"`
}

type UpdateOrderParams struct {
	ID string `json:"id" uri:"id"`
}

type CancelOrderParams struct {
	ID string `json:"id" uri:"id"`
}

// Order query parameters for listing/filtering
type ListOrdersQuery struct {
	UserID        string   `json:"user_id" form:"user_id"`
	Status        string   `json:"status" form:"status"`
	MinAmount     *float64 `json:"min_amount" form:"min_amount"`
	MaxAmount     *float64 `json:"max_amount" form:"max_amount"`
	CreatedAfter  *string  `json:"created_after" form:"created_after"`
	CreatedBefore *string  `json:"created_before" form:"created_before"`
	Page          int      `json:"page" form:"page"`
	PageSize      int      `json:"page_size" form:"page_size"`
	SortBy        string   `json:"sort_by" form:"sort_by"`
	SortOrder     string   `json:"sort_order" form:"sort_order"`
}

// Analytics query parameters
type OrderAnalyticsQuery struct {
	UserID   string `json:"user_id" form:"user_id"`
	DateFrom string `json:"date_from" form:"date_from"`
	DateTo   string `json:"date_to" form:"date_to"`
	GroupBy  string `json:"group_by" form:"group_by"`
	Currency string `json:"currency" form:"currency"`
}

// Response types
type OrderListResponse struct {
	Orders     []Order `json:"orders"`
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	HasNext    bool    `json:"has_next"`
}

type OrderAnalyticsResponse struct {
	TotalOrders    int                   `json:"total_orders"`
	TotalRevenue   float64               `json:"total_revenue"`
	AverageOrder   float64               `json:"average_order"`
	Currency       string                `json:"currency"`
	TopProducts    []ProductAnalytics    `json:"top_products"`
	TimeSeriesData []TimeSeriesDataPoint `json:"time_series_data"`
}

type ProductAnalytics struct {
	ProductID    string  `json:"product_id"`
	Name         string  `json:"name"`
	TotalSold    int     `json:"total_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

type TimeSeriesDataPoint struct {
	Date       string  `json:"date"`
	OrderCount int     `json:"order_count"`
	Revenue    float64 `json:"revenue"`
}

// Business logic handlers
func createOrderHandler(ctx context.Context, params struct{}, query struct{}, body CreateOrderRequest) (Order, error) {
	// Simulate order creation with item processing
	items := make([]OrderItem, len(body.Items))
	total := 0.0

	for i, item := range body.Items {
		price := 29.99 // Simulate product lookup
		subtotal := float64(item.Quantity) * price
		total += subtotal

		items[i] = OrderItem{
			ProductID: item.ProductID,
			Name:      fmt.Sprintf("Product %s", item.ProductID),
			Quantity:  item.Quantity,
			Price:     price,
			Subtotal:  subtotal,
		}
	}

	return Order{
		ID:              "ord_" + fmt.Sprintf("%d", time.Now().Unix()),
		UserID:          body.UserID,
		Status:          "pending",
		Items:           items,
		TotalAmount:     total,
		Currency:        body.Currency,
		ShippingAddress: body.ShippingAddress,
		BillingAddress:  body.BillingAddress,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

func getOrderHandler(ctx context.Context, params GetOrderParams, query struct{}, body struct{}) (Order, error) {
	// Simulate getting order from database
	return Order{
		ID:     params.ID,
		UserID: "usr_123",
		Status: "shipped",
		Items: []OrderItem{
			{
				ProductID: "prod_123",
				Name:      "Wireless Headphones",
				Quantity:  1,
				Price:     99.99,
				Subtotal:  99.99,
			},
		},
		TotalAmount: 99.99,
		Currency:    "USD",
		ShippingAddress: Address{
			Street:     "123 Main St",
			City:       "New York",
			State:      "NY",
			PostalCode: "10001",
			Country:    "US",
		},
		BillingAddress: Address{
			Street:     "123 Main St",
			City:       "New York",
			State:      "NY",
			PostalCode: "10001",
			Country:    "US",
		},
		CreatedAt: time.Now().Add(-48 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}, nil
}

func updateOrderStatusHandler(ctx context.Context, params UpdateOrderParams, query struct{}, body UpdateOrderStatusRequest) (Order, error) {
	// Simulate order status update
	order, _ := getOrderHandler(ctx, GetOrderParams{ID: params.ID}, struct{}{}, struct{}{})
	order.Status = body.Status
	order.UpdatedAt = time.Now()
	return order, nil
}

func cancelOrderHandler(ctx context.Context, params CancelOrderParams, query struct{}, body struct{}) (struct{}, error) {
	// Simulate order cancellation
	return struct{}{}, nil
}

func listOrdersHandler(ctx context.Context, params struct{}, query ListOrdersQuery, body struct{}) (OrderListResponse, error) {
	// Simulate listing orders with complex filtering
	orders := []Order{
		{
			ID:     "ord_123",
			UserID: query.UserID,
			Status: "completed",
			Items: []OrderItem{
				{ProductID: "prod_123", Name: "Laptop", Quantity: 1, Price: 999.99, Subtotal: 999.99},
			},
			TotalAmount: 999.99,
			Currency:    "USD",
			CreatedAt:   time.Now().Add(-72 * time.Hour),
			UpdatedAt:   time.Now().Add(-48 * time.Hour),
		},
	}

	return OrderListResponse{
		Orders:     orders,
		TotalCount: 1,
		Page:       query.Page,
		PageSize:   query.PageSize,
		HasNext:    false,
	}, nil
}

func getOrderAnalyticsHandler(ctx context.Context, params struct{}, query OrderAnalyticsQuery, body struct{}) (OrderAnalyticsResponse, error) {
	// Simulate analytics data
	return OrderAnalyticsResponse{
		TotalOrders:  150,
		TotalRevenue: 25000.50,
		AverageOrder: 166.67,
		Currency:     query.Currency,
		TopProducts: []ProductAnalytics{
			{ProductID: "prod_123", Name: "Laptop", TotalSold: 25, TotalRevenue: 12500.00},
			{ProductID: "prod_456", Name: "Mouse", TotalSold: 80, TotalRevenue: 2400.00},
		},
		TimeSeriesData: []TimeSeriesDataPoint{
			{Date: "2024-01-01", OrderCount: 15, Revenue: 2500.00},
			{Date: "2024-01-02", OrderCount: 22, Revenue: 3200.00},
		},
	}, nil
}

func main() {
	engine := gin.Default()
	openAPIGen := operations.NewOpenAPIGenerator("Order Service API", "1.2.0")
	router := operations.NewRouter(engine, openAPIGen)

	// Define comprehensive schemas
	addressSchema := validators.Object(map[string]interface{}{
		"street":      validators.String().Min(1).Max(255).Required(),
		"city":        validators.String().Min(1).Max(100).Required(),
		"state":       validators.String().Min(2).Max(100).Required(),
		"postal_code": validators.String().Min(3).Max(20).Pattern("^[0-9A-Za-z\\s-]+$").Required(),
		"country":     validators.String().Min(2).Max(3).Pattern("^[A-Z]{2,3}$").Required(),
	}).Required()

	createOrderItemSchema := validators.Object(map[string]interface{}{
		"product_id": validators.String().Min(1).Pattern("^prod_[a-zA-Z0-9]+$").Required(),
		"quantity":   validators.Number().Min(1).Max(100).Required(),
	}).Required()

	createOrderBodySchema := validators.Object(map[string]interface{}{
		"user_id":          validators.String().Min(1).Pattern("^usr_[a-zA-Z0-9]+$").Required(),
		"items":            validators.Array(createOrderItemSchema).Required(),
		"currency":         validators.String().Min(3).Max(3).Pattern("^[A-Z]{3}$").Optional().Default("USD"),
		"shipping_address": addressSchema,
		"billing_address":  addressSchema,
	}).Required()

	updateOrderStatusBodySchema := validators.Object(map[string]interface{}{
		"status": validators.String().Required(),
	}).Required()

	orderParamsSchema := validators.Object(map[string]interface{}{
		"id": validators.String().Min(1).Pattern("^ord_[a-zA-Z0-9]+$").Required(),
	}).Required()

	listOrdersQuerySchema := validators.Object(map[string]interface{}{
		"user_id":        validators.String().Pattern("^usr_[a-zA-Z0-9]+$").Optional(),
		"status":         validators.String().Optional(),
		"min_amount":     validators.Number().Min(0).Optional(),
		"max_amount":     validators.Number().Min(0).Optional(),
		"created_after":  validators.String().Optional(),
		"created_before": validators.String().Optional(),
		"page":           validators.Number().Min(1).Optional().Default(1),
		"page_size":      validators.Number().Min(1).Max(100).Optional().Default(20),
		"sort_by":        validators.String().Optional().Default("created_at"),
		"sort_order":     validators.String().Optional().Default("desc"),
	}).Optional()

	analyticsQuerySchema := validators.Object(map[string]interface{}{
		"user_id":   validators.String().Pattern("^usr_[a-zA-Z0-9]+$").Optional(),
		"date_from": validators.String().Required(),
		"date_to":   validators.String().Required(),
		"group_by":  validators.String().Optional().Default("day"),
		"currency":  validators.String().Pattern("^[A-Z]{3}$").Optional().Default("USD"),
	}).Required()

	orderItemSchema := validators.Object(map[string]interface{}{
		"product_id": validators.String().Min(1).Required(),
		"name":       validators.String().Min(1).Required(),
		"quantity":   validators.Number().Min(1).Required(),
		"price":      validators.Number().Min(0).Required(),
		"subtotal":   validators.Number().Min(0).Required(),
	}).Required()

	orderResponseSchema := validators.Object(map[string]interface{}{
		"id":               validators.String().Min(1).Required(),
		"user_id":          validators.String().Min(1).Required(),
		"status":           validators.String().Required(),
		"items":            validators.Array(orderItemSchema).Required(),
		"total_amount":     validators.Number().Min(0).Required(),
		"currency":         validators.String().Min(3).Max(3).Required(),
		"shipping_address": addressSchema,
		"billing_address":  addressSchema,
		"created_at":       validators.String().Required(),
		"updated_at":       validators.String().Required(),
	}).Required()

	orderListResponseSchema := validators.Object(map[string]interface{}{
		"orders":      validators.Array(orderResponseSchema).Required(),
		"total_count": validators.Number().Min(0).Required(),
		"page":        validators.Number().Min(1).Required(),
		"page_size":   validators.Number().Min(1).Required(),
		"has_next":    validators.Bool().Required(),
	}).Required()

	productAnalyticsSchema := validators.Object(map[string]interface{}{
		"product_id":    validators.String().Min(1).Required(),
		"name":          validators.String().Min(1).Required(),
		"total_sold":    validators.Number().Min(0).Required(),
		"total_revenue": validators.Number().Min(0).Required(),
	}).Required()

	timeSeriesDataSchema := validators.Object(map[string]interface{}{
		"date":        validators.String().Required(),
		"order_count": validators.Number().Min(0).Required(),
		"revenue":     validators.Number().Min(0).Required(),
	}).Required()

	analyticsResponseSchema := validators.Object(map[string]interface{}{
		"total_orders":     validators.Number().Min(0).Required(),
		"total_revenue":    validators.Number().Min(0).Required(),
		"average_order":    validators.Number().Min(0).Required(),
		"currency":         validators.String().Min(3).Max(3).Required(),
		"top_products":     validators.Array(productAnalyticsSchema).Required(),
		"time_series_data": validators.Array(timeSeriesDataSchema).Required(),
	}).Required()

	// Define operations with rich metadata
	createOrderOp := operations.NewSimple().
		POST("/orders").
		Summary("Create a new order").
		Description("Creates a new order with items, shipping and billing addresses").
		Tags("orders", "e-commerce").
		WithBody(createOrderBodySchema).
		WithResponse(orderResponseSchema).
		Handler(operations.CreateValidatedHandler(createOrderHandler, nil, nil, createOrderBodySchema, orderResponseSchema))

	getOrderOp := operations.NewSimple().
		GET("/orders/{id}").
		Summary("Get order by ID").
		Description("Retrieves a specific order with all its details including items and addresses").
		Tags("orders").
		WithParams(orderParamsSchema).
		WithResponse(orderResponseSchema).
		Handler(operations.CreateValidatedHandler(getOrderHandler, orderParamsSchema, nil, nil, orderResponseSchema))

	updateOrderStatusOp := operations.NewSimple().
		PATCH("/orders/{id}/status").
		Summary("Update order status").
		Description("Updates the status of an existing order (pending, confirmed, shipped, delivered, cancelled)").
		Tags("orders", "status").
		WithParams(orderParamsSchema).
		WithBody(updateOrderStatusBodySchema).
		WithResponse(orderResponseSchema).
		Handler(operations.CreateValidatedHandler(updateOrderStatusHandler, orderParamsSchema, nil, updateOrderStatusBodySchema, orderResponseSchema))

	cancelOrderOp := operations.NewSimple().
		DELETE("/orders/{id}").
		Summary("Cancel order").
		Description("Cancels an order if it hasn't been shipped yet").
		Tags("orders").
		WithParams(orderParamsSchema).
		Handler(operations.CreateValidatedHandler(cancelOrderHandler, orderParamsSchema, nil, nil, nil))

	listOrdersOp := operations.NewSimple().
		GET("/orders").
		Summary("List orders with filtering").
		Description("Retrieves orders with advanced filtering, pagination, and sorting options").
		Tags("orders", "filtering").
		WithQuery(listOrdersQuerySchema).
		WithResponse(orderListResponseSchema).
		Handler(operations.CreateValidatedHandler(listOrdersHandler, nil, listOrdersQuerySchema, nil, orderListResponseSchema))

	getAnalyticsOp := operations.NewSimple().
		GET("/analytics/orders").
		Summary("Get order analytics").
		Description("Retrieves comprehensive order analytics including revenue, top products, and time series data").
		Tags("analytics", "reporting").
		WithQuery(analyticsQuerySchema).
		WithResponse(analyticsResponseSchema).
		Handler(operations.CreateValidatedHandler(getOrderAnalyticsHandler, nil, analyticsQuerySchema, nil, analyticsResponseSchema))

	// Register operations
	router.Register(createOrderOp)
	router.Register(getOrderOp)
	router.Register(updateOrderStatusOp)
	router.Register(cancelOrderOp)
	router.Register(listOrdersOp)
	router.Register(getAnalyticsOp)

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "order-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	fmt.Println("🚀 Order Service starting on :8002")
	fmt.Println("📚 Generate OpenAPI spec: go-op generate -i ./examples/order-service -o ./order-service.yaml")
	engine.Run(":8002")
}
