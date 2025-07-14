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

// PaymentMethod represents different payment methods (OneOf example)
type PaymentMethod struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// ShippingMethod represents different shipping options (OneOf example)
type ShippingMethod struct {
	Type          string  `json:"type"`
	Provider      string  `json:"provider"`
	Cost          float64 `json:"cost"`
	EstimatedDays int     `json:"estimated_days"`
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
	order, _ := getOrderHandler(ctx, GetOrderParams(params), struct{}{}, struct{}{})
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

	// Create OpenAPI generator with enhanced metadata
	openAPIGen := operations.NewOpenAPIGenerator("Order Service API", "1.2.0")

	// Demonstrate OpenAPI 3.1 Fixed Fields features
	openAPIGen.SetDescription("A comprehensive order processing service with advanced e-commerce features, complex nested schemas, and rich validation")
	openAPIGen.SetSummary("Order Processing & E-commerce API")
	openAPIGen.SetTermsOfService("https://api.ecommerce.com/terms")

	// Enhanced contact information
	openAPIGen.SetContact(&operations.OpenAPIContact{
		Name:  "E-commerce Order Team",
		Email: "orders@ecommerce.com",
		URL:   "https://api.ecommerce.com/support/orders",
	})

	// License information
	openAPIGen.SetLicense(&operations.OpenAPILicense{
		Name: "Apache 2.0",
		URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
	})

	// Global tags with external documentation
	openAPIGen.AddTag(operations.OpenAPITag{
		Name:        "orders",
		Description: "Order management operations including creation, status updates, and cancellation",
		ExternalDocs: &operations.OpenAPIExternalDocs{
			Description: "Order management documentation",
			URL:         "https://docs.ecommerce.com/orders",
		},
	})

	openAPIGen.AddTag(operations.OpenAPITag{
		Name:        "e-commerce",
		Description: "E-commerce specific operations for order processing",
		ExternalDocs: &operations.OpenAPIExternalDocs{
			Description: "E-commerce integration guide",
			URL:         "https://docs.ecommerce.com/integration",
		},
	})

	openAPIGen.AddTag(operations.OpenAPITag{
		Name:        "filtering",
		Description: "Advanced filtering and search operations for orders",
	})

	openAPIGen.AddTag(operations.OpenAPITag{
		Name:        "analytics",
		Description: "Order analytics and reporting operations",
		ExternalDocs: &operations.OpenAPIExternalDocs{
			Description: "Analytics API documentation",
			URL:         "https://docs.ecommerce.com/analytics",
		},
	})

	openAPIGen.AddTag(operations.OpenAPITag{
		Name:        "status",
		Description: "Order status management and tracking",
	})

	openAPIGen.AddTag(operations.OpenAPITag{
		Name:        "reporting",
		Description: "Business intelligence and reporting features",
	})

	// Global external documentation
	openAPIGen.SetExternalDocs(&operations.OpenAPIExternalDocs{
		Description: "Complete order service documentation with examples and integration guides",
		URL:         "https://docs.ecommerce.com/order-service",
	})

	// Server configuration with variables
	openAPIGen.AddServer(operations.OpenAPIServer{
		URL:         "https://{environment}.orders.ecommerce.com/{version}",
		Description: "Order service with configurable environment and API version",
		Variables: map[string]operations.OpenAPIServerVariable{
			"environment": {
				Default:     "api",
				Enum:        []string{"api", "staging", "sandbox"},
				Description: "Order service environment",
			},
			"version": {
				Default:     "v1",
				Enum:        []string{"v1", "v2"},
				Description: "API version",
			},
		},
	})

	// Set JSON Schema dialect
	openAPIGen.SetJsonSchemaDialect("https://json-schema.org/draft/2020-12/schema")

	router := operations.NewRouter(engine, openAPIGen)

	// ===== OneOf Schema Examples for E-commerce =====
	// These demonstrate complex OneOf patterns for flexible payment and shipping options

	// Payment method OneOf - credit card, PayPal, bank transfer, crypto
	creditCardPaymentSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^credit_card$").
			Example("credit_card").
			Required(),
		"card_number": validators.String().Pattern(`^\d{16}$`).
			Example("1234567890123456").
			Required(),
		"expiry_month": validators.Number().Min(1).Max(12).
			Example(12).
			Required(),
		"expiry_year": validators.Number().Min(2024).Max(2034).
			Example(2025).
			Required(),
		"cvv": validators.String().Pattern(`^\d{3,4}$`).
			Example("123").
			Required(),
		"cardholder_name": validators.String().Min(2).Max(100).
			Example("John Doe").
			Required(),
		"billing_zip": validators.String().Pattern(`^\d{5}(-\d{4})?$`).
			Example("10001").
			Optional(),
	}).Example(map[string]interface{}{
		"type":            "credit_card",
		"card_number":     "1234567890123456",
		"expiry_month":    12,
		"expiry_year":     2025,
		"cvv":             "123",
		"cardholder_name": "John Doe",
		"billing_zip":     "10001",
	}).Required()

	paypalPaymentSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^paypal$").
			Example("paypal").
			Required(),
		"email": validators.String().Email().
			Example("payment@example.com").
			Required(),
		"payer_id": validators.String().Min(10).Max(20).
			Example("PAYERID123456789").
			Optional(),
	}).Example(map[string]interface{}{
		"type":     "paypal",
		"email":    "payment@example.com",
		"payer_id": "PAYERID123456789",
	}).Required()

	bankTransferPaymentSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^bank_transfer$").
			Example("bank_transfer").
			Required(),
		"account_number": validators.String().Pattern(`^\d{10,12}$`).
			Example("1234567890").
			Required(),
		"routing_number": validators.String().Pattern(`^\d{9}$`).
			Example("123456789").
			Required(),
		"account_type": validators.String().
			Examples(map[string]validators.ExampleObject{
				"checking": {
					Summary:     "Checking account",
					Description: "Standard checking account for everyday transactions",
					Value:       "checking",
				},
				"savings": {
					Summary:     "Savings account",
					Description: "Savings account with higher interest",
					Value:       "savings",
				},
			}).
			Required(),
		"bank_name": validators.String().Min(2).Max(100).
			Example("Chase Bank").
			Optional(),
	}).Example(map[string]interface{}{
		"type":           "bank_transfer",
		"account_number": "1234567890",
		"routing_number": "123456789",
		"account_type":   "checking",
		"bank_name":      "Chase Bank",
	}).Required()

	cryptoPaymentSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^crypto$").
			Example("crypto").
			Required(),
		"currency": validators.String().
			Examples(map[string]validators.ExampleObject{
				"bitcoin": {
					Summary:     "Bitcoin",
					Description: "Bitcoin cryptocurrency payment",
					Value:       "BTC",
				},
				"ethereum": {
					Summary:     "Ethereum",
					Description: "Ethereum cryptocurrency payment",
					Value:       "ETH",
				},
				"litecoin": {
					Summary:     "Litecoin",
					Description: "Litecoin cryptocurrency payment",
					Value:       "LTC",
				},
			}).
			Required(),
		"wallet_address": validators.String().Min(26).Max(62).
			Example("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa").
			Required(),
		"network": validators.String().
			Example("mainnet").
			Optional().Default("mainnet"),
	}).Example(map[string]interface{}{
		"type":           "crypto",
		"currency":       "BTC",
		"wallet_address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
		"network":        "mainnet",
	}).Required()

	// OneOf payment method schema
	paymentMethodSchema := validators.OneOf(
		creditCardPaymentSchema,
		paypalPaymentSchema,
		bankTransferPaymentSchema,
		cryptoPaymentSchema,
	).Required()

	// Shipping method OneOf - standard, expedited, overnight, pickup
	standardShippingSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^standard$").
			Example("standard").
			Required(),
		"provider": validators.String().
			Examples(map[string]validators.ExampleObject{
				"usps": {
					Summary:     "USPS Standard",
					Description: "United States Postal Service standard shipping",
					Value:       "USPS",
				},
				"ups": {
					Summary:     "UPS Ground",
					Description: "UPS standard ground shipping",
					Value:       "UPS",
				},
				"fedex": {
					Summary:     "FedEx Ground",
					Description: "FedEx standard ground shipping",
					Value:       "FedEx",
				},
			}).
			Required(),
		"estimated_days": validators.Number().Min(3).Max(10).
			Example(5).
			Required(),
		"cost": validators.Number().Min(0).
			Example(9.99).
			Required(),
		"tracking_available": validators.Bool().
			Example(true).
			Optional().Default(true),
	}).Example(map[string]interface{}{
		"type":               "standard",
		"provider":           "UPS",
		"estimated_days":     5,
		"cost":               9.99,
		"tracking_available": true,
	}).Required()

	expeditedShippingSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^expedited$").
			Example("expedited").
			Required(),
		"provider": validators.String().
			Example("FedEx").
			Required(),
		"estimated_days": validators.Number().Min(1).Max(3).
			Example(2).
			Required(),
		"cost": validators.Number().Min(0).
			Example(19.99).
			Required(),
		"signature_required": validators.Bool().
			Example(false).
			Optional().Default(false),
	}).Example(map[string]interface{}{
		"type":               "expedited",
		"provider":           "FedEx",
		"estimated_days":     2,
		"cost":               19.99,
		"signature_required": false,
	}).Required()

	overnightShippingSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^overnight$").
			Example("overnight").
			Required(),
		"provider": validators.String().
			Example("FedEx").
			Required(),
		"estimated_days": validators.Number().Min(1).Max(1).
			Example(1).
			Required(),
		"cost": validators.Number().Min(0).
			Example(39.99).
			Required(),
		"delivery_time": validators.String().
			Examples(map[string]validators.ExampleObject{
				"morning": {
					Summary:     "Morning delivery",
					Description: "Delivery before 10:30 AM",
					Value:       "10:30 AM",
				},
				"afternoon": {
					Summary:     "Afternoon delivery",
					Description: "Delivery before 5:00 PM",
					Value:       "5:00 PM",
				},
			}).
			Optional(),
		"signature_required": validators.Bool().
			Example(true).
			Optional().Default(true),
	}).Example(map[string]interface{}{
		"type":               "overnight",
		"provider":           "FedEx",
		"estimated_days":     1,
		"cost":               39.99,
		"delivery_time":      "10:30 AM",
		"signature_required": true,
	}).Required()

	pickupShippingSchema := validators.Object(map[string]interface{}{
		"type": validators.String().Pattern("^pickup$").
			Example("pickup").
			Required(),
		"location": validators.String().Min(5).Max(200).
			Example("123 Store Street, New York, NY 10001").
			Required(),
		"store_hours": validators.String().
			Example("9:00 AM - 9:00 PM").
			Required(),
		"cost": validators.Number().Min(0).
			Example(0).
			Required(),
		"estimated_days": validators.Number().Min(1).Max(3).
			Example(2).
			Required(),
		"special_instructions": validators.String().Max(500).
			Example("Please bring photo ID for pickup verification").
			Optional(),
	}).Example(map[string]interface{}{
		"type":                 "pickup",
		"location":             "123 Store Street, New York, NY 10001",
		"store_hours":          "9:00 AM - 9:00 PM",
		"cost":                 0,
		"estimated_days":       2,
		"special_instructions": "Please bring photo ID for pickup verification",
	}).Required()

	// OneOf shipping method schema
	shippingMethodSchema := validators.OneOf(
		standardShippingSchema,
		expeditedShippingSchema,
		overnightShippingSchema,
		pickupShippingSchema,
	).Required()

	// Define comprehensive schemas with rich examples
	addressSchema := validators.Object(map[string]interface{}{
		"street": validators.String().Min(1).Max(255).
			Examples(map[string]validators.ExampleObject{
				"residential": {
					Summary:     "Residential address",
					Description: "Typical home address format",
					Value:       "123 Oak Street",
				},
				"apartment": {
					Summary:     "Apartment address",
					Description: "Address with apartment number",
					Value:       "456 Elm Ave, Apt 2B",
				},
				"business": {
					Summary:     "Business address",
					Description: "Commercial building address",
					Value:       "789 Corporate Blvd, Suite 100",
				},
			}).
			Required(),
		"city": validators.String().Min(1).Max(100).
			Example("New York").
			Required(),
		"state": validators.String().Min(2).Max(100).
			Examples(map[string]validators.ExampleObject{
				"abbreviated": {
					Summary:     "State abbreviation",
					Description: "Two-letter state code",
					Value:       "NY",
				},
				"full_name": {
					Summary:     "Full state name",
					Description: "Complete state name",
					Value:       "New York",
				},
			}).
			Required(),
		"postal_code": validators.String().Min(3).Max(20).Pattern("^[0-9A-Za-z\\s-]+$").
			Examples(map[string]validators.ExampleObject{
				"us_zip": {
					Summary:     "US ZIP code",
					Description: "Standard 5-digit US postal code",
					Value:       "10001",
				},
				"us_zip_plus4": {
					Summary:     "US ZIP+4 code",
					Description: "Extended 9-digit US postal code",
					Value:       "10001-1234",
				},
				"uk_postcode": {
					Summary:     "UK postcode",
					Description: "British postal code format",
					Value:       "SW1A 1AA",
				},
			}).
			Required(),
		"country": validators.String().Min(2).Max(3).Pattern("^[A-Z]{2,3}$").
			Examples(map[string]validators.ExampleObject{
				"us": {
					Summary:     "United States",
					Description: "ISO country code for USA",
					Value:       "US",
				},
				"uk": {
					Summary:     "United Kingdom",
					Description: "ISO country code for UK",
					Value:       "GB",
				},
				"canada": {
					Summary:     "Canada",
					Description: "ISO country code for Canada",
					Value:       "CA",
				},
			}).
			Required(),
	}).Example(map[string]interface{}{
		"street":      "123 Oak Street",
		"city":        "New York",
		"state":       "NY",
		"postal_code": "10001",
		"country":     "US",
	}).Required()

	createOrderItemSchema := validators.Object(map[string]interface{}{
		"product_id": validators.String().Min(1).Pattern("^prod_[a-zA-Z0-9]+$").
			Examples(map[string]validators.ExampleObject{
				"electronics": {
					Summary:     "Electronics product",
					Description: "ID for electronic devices and gadgets",
					Value:       "prod_headphones_123",
				},
				"clothing": {
					Summary:     "Clothing product",
					Description: "ID for apparel and fashion items",
					Value:       "prod_shirt_456",
				},
				"books": {
					Summary:     "Book product",
					Description: "ID for books and publications",
					Value:       "prod_book_789",
				},
			}).
			Required(),
		"quantity": validators.Number().Min(1).Max(100).
			Examples(map[string]validators.ExampleObject{
				"single": {
					Summary:     "Single item",
					Description: "Ordering just one unit",
					Value:       1,
				},
				"multiple": {
					Summary:     "Multiple items",
					Description: "Bulk order of several units",
					Value:       5,
				},
				"bulk": {
					Summary:     "Bulk order",
					Description: "Large quantity for wholesale",
					Value:       25,
				},
			}).
			Required(),
	}).Example(map[string]interface{}{
		"product_id": "prod_headphones_123",
		"quantity":   2,
	}).Required()

	createOrderBodySchema := validators.Object(map[string]interface{}{
		"user_id":          validators.String().Min(1).Pattern("^usr_[a-zA-Z0-9]+$").Required(),
		"items":            validators.Array(createOrderItemSchema).Required(),
		"currency":         validators.String().Min(3).Max(3).Pattern("^[A-Z]{3}$").Optional().Default("USD"),
		"shipping_address": addressSchema,
		"billing_address":  addressSchema,
		"payment_method":   paymentMethodSchema,
		"shipping_method":  shippingMethodSchema,
		"special_instructions": validators.String().Max(1000).
			Examples(map[string]validators.ExampleObject{
				"fragile": {
					Summary:     "Fragile item instructions",
					Description: "Instructions for handling fragile or delicate items",
					Value:       "Please handle with care - contains fragile electronics",
				},
				"gift": {
					Summary:     "Gift wrapping instructions",
					Description: "Instructions for gift wrapping and presentation",
					Value:       "Please gift wrap with premium paper and include gift receipt",
				},
			}).Optional(),
	}).Example(map[string]interface{}{
		"user_id": "usr_12345",
		"items": []map[string]interface{}{
			{
				"product_id": "prod_headphones_123",
				"quantity":   1,
			},
		},
		"currency": "USD",
		"shipping_address": map[string]interface{}{
			"street":      "123 Oak Street",
			"city":        "New York",
			"state":       "NY",
			"postal_code": "10001",
			"country":     "US",
		},
		"billing_address": map[string]interface{}{
			"street":      "123 Oak Street",
			"city":        "New York",
			"state":       "NY",
			"postal_code": "10001",
			"country":     "US",
		},
		"payment_method": map[string]interface{}{
			"type":            "credit_card",
			"card_number":     "1234567890123456",
			"cardholder_name": "John Doe",
			"cvv":             "123",
		},
		"shipping_method": map[string]interface{}{
			"type":     "expedited",
			"provider": "FedEx",
		},
		"special_instructions": "Please handle with care - contains fragile electronics",
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
		Description("Updates the status of an existing order (pending, confirmed, shipped, delivered, canceled)").
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

	// New operation showcasing OneOf for payment and shipping methods
	createAdvancedOrderOp := operations.NewSimple().
		POST("/orders/advanced").
		Summary("Create order with flexible payment and shipping options").
		Description("Creates an order with OneOf support for multiple payment methods (credit card, PayPal, bank transfer, crypto) "+
			"and shipping options (standard, expedited, overnight, pickup). This endpoint demonstrates how OneOf schemas "+
			"provide flexible API design for e-commerce platforms where customers can choose from various payment and shipping methods.").
		Tags("orders", "e-commerce", "oneof-example", "payment", "shipping").
		WithBody(createOrderBodySchema).
		WithResponse(orderResponseSchema).
		Handler(operations.CreateValidatedHandler(
			createOrderHandler, // Reuse existing handler for demo
			nil,
			nil,
			createOrderBodySchema,
			orderResponseSchema,
		))

	// Register operations
	router.Register(createOrderOp)
	router.Register(getOrderOp)
	router.Register(updateOrderStatusOp)
	router.Register(cancelOrderOp)
	router.Register(listOrdersOp)
	router.Register(getAnalyticsOp)
	router.Register(createAdvancedOrderOp) // OneOf showcase operation

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
