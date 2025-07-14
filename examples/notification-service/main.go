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

// Notification represents a notification in the system
type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Type      string                 `json:"type"`
	Channel   string                 `json:"channel"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Priority  string                 `json:"priority"`
	Status    string                 `json:"status"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
	SentAt    *time.Time             `json:"sent_at,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Template represents a notification template
type Template struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Channel   string                 `json:"channel"`
	Subject   string                 `json:"subject"`
	Body      string                 `json:"body"`
	Variables []string               `json:"variables"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	IsActive  bool                   `json:"is_active"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Request types
type SendNotificationRequest struct {
	UserID   string                 `json:"user_id"`
	Type     string                 `json:"type"`
	Channel  string                 `json:"channel"`
	Title    string                 `json:"title"`
	Message  string                 `json:"message"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Priority string                 `json:"priority"`
}

type SendBulkNotificationRequest struct {
	UserIDs    []string               `json:"user_ids"`
	Type       string                 `json:"type"`
	Channel    string                 `json:"channel"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Priority   string                 `json:"priority"`
	ScheduleAt *time.Time             `json:"schedule_at,omitempty"`
}

type SendTemplatedNotificationRequest struct {
	UserID     string                 `json:"user_id"`
	TemplateID string                 `json:"template_id"`
	Variables  map[string]interface{} `json:"variables"`
	Priority   string                 `json:"priority"`
	ScheduleAt *time.Time             `json:"schedule_at,omitempty"`
}

type CreateTemplateRequest struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Channel   string                 `json:"channel"`
	Subject   string                 `json:"subject"`
	Body      string                 `json:"body"`
	Variables []string               `json:"variables"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateTemplateRequest struct {
	Name      *string                `json:"name,omitempty"`
	Subject   *string                `json:"subject,omitempty"`
	Body      *string                `json:"body,omitempty"`
	Variables []string               `json:"variables,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	IsActive  *bool                  `json:"is_active,omitempty"`
}

// Path parameters
type GetNotificationParams struct {
	ID string `json:"id" uri:"id"`
}

type MarkReadParams struct {
	ID string `json:"id" uri:"id"`
}

type GetTemplateParams struct {
	ID string `json:"id" uri:"id"`
}

type UpdateTemplateParams struct {
	ID string `json:"id" uri:"id"`
}

type DeleteTemplateParams struct {
	ID string `json:"id" uri:"id"`
}

// Query parameters
type ListNotificationsQuery struct {
	UserID   string `json:"user_id" form:"user_id"`
	Type     string `json:"type" form:"type"`
	Channel  string `json:"channel" form:"channel"`
	Status   string `json:"status" form:"status"`
	Priority string `json:"priority" form:"priority"`
	IsRead   *bool  `json:"is_read" form:"is_read"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	SortBy   string `json:"sort_by" form:"sort_by"`
}

type GetNotificationStatsQuery struct {
	UserID   string `json:"user_id" form:"user_id"`
	DateFrom string `json:"date_from" form:"date_from"`
	DateTo   string `json:"date_to" form:"date_to"`
	GroupBy  string `json:"group_by" form:"group_by"`
	Channel  string `json:"channel" form:"channel"`
	Type     string `json:"type" form:"type"`
}

type ListTemplatesQuery struct {
	Type     string `json:"type" form:"type"`
	Channel  string `json:"channel" form:"channel"`
	IsActive *bool  `json:"is_active" form:"is_active"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
}

// Response types
type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	TotalCount    int            `json:"total_count"`
	UnreadCount   int            `json:"unread_count"`
	Page          int            `json:"page"`
	PageSize      int            `json:"page_size"`
	HasNext       bool           `json:"has_next"`
}

type BulkNotificationResponse struct {
	JobID       string     `json:"job_id"`
	TotalCount  int        `json:"total_count"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	Status      string     `json:"status"`
}

type NotificationStatsResponse struct {
	TotalSent      int                    `json:"total_sent"`
	TotalRead      int                    `json:"total_read"`
	ReadRate       float64                `json:"read_rate"`
	ChannelStats   []ChannelStats         `json:"channel_stats"`
	TypeStats      []TypeStats            `json:"type_stats"`
	TimeSeriesData []StatsTimeSeriesPoint `json:"time_series_data"`
}

type ChannelStats struct {
	Channel   string  `json:"channel"`
	TotalSent int     `json:"total_sent"`
	TotalRead int     `json:"total_read"`
	ReadRate  float64 `json:"read_rate"`
}

type TypeStats struct {
	Type      string  `json:"type"`
	TotalSent int     `json:"total_sent"`
	TotalRead int     `json:"total_read"`
	ReadRate  float64 `json:"read_rate"`
}

type StatsTimeSeriesPoint struct {
	Date      string `json:"date"`
	SentCount int    `json:"sent_count"`
	ReadCount int    `json:"read_count"`
}

type TemplateListResponse struct {
	Templates  []Template `json:"templates"`
	TotalCount int        `json:"total_count"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	HasNext    bool       `json:"has_next"`
}

// Business logic handlers
func sendNotificationHandler(ctx context.Context, params struct{}, query struct{}, body SendNotificationRequest) (Notification, error) {
	now := time.Now()
	return Notification{
		ID:        fmt.Sprintf("notif_%d", now.Unix()),
		UserID:    body.UserID,
		Type:      body.Type,
		Channel:   body.Channel,
		Title:     body.Title,
		Message:   body.Message,
		Data:      body.Data,
		Priority:  body.Priority,
		Status:    "sent",
		SentAt:    &now,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func sendBulkNotificationHandler(ctx context.Context, params struct{}, query struct{}, body SendBulkNotificationRequest) (BulkNotificationResponse, error) {
	return BulkNotificationResponse{
		JobID:       fmt.Sprintf("job_%d", time.Now().Unix()),
		TotalCount:  len(body.UserIDs),
		ScheduledAt: body.ScheduleAt,
		Status:      "queued",
	}, nil
}

func sendTemplatedNotificationHandler(ctx context.Context, params struct{}, query struct{}, body SendTemplatedNotificationRequest) (Notification, error) {
	now := time.Now()
	return Notification{
		ID:        fmt.Sprintf("notif_%d", now.Unix()),
		UserID:    body.UserID,
		Type:      "templated",
		Channel:   "email",
		Title:     "Templated Notification",
		Message:   "This is a templated notification",
		Priority:  body.Priority,
		Status:    "sent",
		SentAt:    &now,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func getNotificationHandler(ctx context.Context, params GetNotificationParams, query struct{}, body struct{}) (Notification, error) {
	now := time.Now()
	return Notification{
		ID:        params.ID,
		UserID:    "usr_123",
		Type:      "order_update",
		Channel:   "email",
		Title:     "Your order has been shipped",
		Message:   "Your order #12345 has been shipped and is on its way",
		Priority:  "high",
		Status:    "sent",
		SentAt:    &now,
		CreatedAt: now.Add(-2 * time.Hour),
		UpdatedAt: now,
	}, nil
}

func markNotificationReadHandler(ctx context.Context, params MarkReadParams, query struct{}, body struct{}) (Notification, error) {
	notification, _ := getNotificationHandler(ctx, GetNotificationParams{ID: params.ID}, struct{}{}, struct{}{})
	now := time.Now()
	notification.Status = "read"
	notification.ReadAt = &now
	notification.UpdatedAt = now
	return notification, nil
}

func listNotificationsHandler(ctx context.Context, params struct{}, query ListNotificationsQuery, body struct{}) (NotificationListResponse, error) {
	notifications := []Notification{
		{
			ID:        "notif_123",
			UserID:    query.UserID,
			Type:      "order_update",
			Channel:   "email",
			Title:     "Order Shipped",
			Message:   "Your order has been shipped",
			Priority:  "high",
			Status:    "sent",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	return NotificationListResponse{
		Notifications: notifications,
		TotalCount:    1,
		UnreadCount:   0,
		Page:          query.Page,
		PageSize:      query.PageSize,
		HasNext:       false,
	}, nil
}

func getNotificationStatsHandler(ctx context.Context, params struct{}, query GetNotificationStatsQuery, body struct{}) (NotificationStatsResponse, error) {
	return NotificationStatsResponse{
		TotalSent: 1250,
		TotalRead: 950,
		ReadRate:  76.0,
		ChannelStats: []ChannelStats{
			{Channel: "email", TotalSent: 800, TotalRead: 650, ReadRate: 81.25},
			{Channel: "sms", TotalSent: 300, TotalRead: 250, ReadRate: 83.33},
			{Channel: "push", TotalSent: 150, TotalRead: 50, ReadRate: 33.33},
		},
		TypeStats: []TypeStats{
			{Type: "order_update", TotalSent: 600, TotalRead: 500, ReadRate: 83.33},
			{Type: "promotion", TotalSent: 400, TotalRead: 280, ReadRate: 70.0},
			{Type: "reminder", TotalSent: 250, TotalRead: 170, ReadRate: 68.0},
		},
		TimeSeriesData: []StatsTimeSeriesPoint{
			{Date: "2024-01-01", SentCount: 120, ReadCount: 95},
			{Date: "2024-01-02", SentCount: 140, ReadCount: 110},
		},
	}, nil
}

func createTemplateHandler(ctx context.Context, params struct{}, query struct{}, body CreateTemplateRequest) (Template, error) {
	now := time.Now()
	return Template{
		ID:        fmt.Sprintf("tpl_%d", now.Unix()),
		Name:      body.Name,
		Type:      body.Type,
		Channel:   body.Channel,
		Subject:   body.Subject,
		Body:      body.Body,
		Variables: body.Variables,
		Metadata:  body.Metadata,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func getTemplateHandler(ctx context.Context, params GetTemplateParams, query struct{}, body struct{}) (Template, error) {
	now := time.Now()
	return Template{
		ID:        params.ID,
		Name:      "Order Confirmation",
		Type:      "order_update",
		Channel:   "email",
		Subject:   "Order Confirmation - {{order_id}}",
		Body:      "Dear {{customer_name}}, your order {{order_id}} has been confirmed.",
		Variables: []string{"order_id", "customer_name"},
		IsActive:  true,
		CreatedAt: now.Add(-48 * time.Hour),
		UpdatedAt: now,
	}, nil
}

func updateTemplateHandler(ctx context.Context, params UpdateTemplateParams, query struct{}, body UpdateTemplateRequest) (Template, error) {
	template, _ := getTemplateHandler(ctx, GetTemplateParams{ID: params.ID}, struct{}{}, struct{}{})

	if body.Name != nil {
		template.Name = *body.Name
	}
	if body.Subject != nil {
		template.Subject = *body.Subject
	}
	if body.Body != nil {
		template.Body = *body.Body
	}
	if body.Variables != nil {
		template.Variables = body.Variables
	}
	if body.Metadata != nil {
		template.Metadata = body.Metadata
	}
	if body.IsActive != nil {
		template.IsActive = *body.IsActive
	}

	template.UpdatedAt = time.Now()
	return template, nil
}

func deleteTemplateHandler(ctx context.Context, params DeleteTemplateParams, query struct{}, body struct{}) (struct{}, error) {
	return struct{}{}, nil
}

func listTemplatesHandler(ctx context.Context, params struct{}, query ListTemplatesQuery, body struct{}) (TemplateListResponse, error) {
	templates := []Template{
		{
			ID:        "tpl_123",
			Name:      "Welcome Email",
			Type:      "welcome",
			Channel:   "email",
			Subject:   "Welcome to our platform",
			Body:      "Welcome {{name}}, thank you for joining us!",
			Variables: []string{"name"},
			IsActive:  true,
			CreatedAt: time.Now().Add(-72 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	return TemplateListResponse{
		Templates:  templates,
		TotalCount: 1,
		Page:       query.Page,
		PageSize:   query.PageSize,
		HasNext:    false,
	}, nil
}

func main() {
	engine := gin.Default()
	openAPIGen := operations.NewOpenAPIGenerator("Notification Service API", "2.1.0")
	router := operations.NewRouter(engine, openAPIGen)

	// Define complex schemas with nested objects and arrays
	sendNotificationBodySchema := validators.Object(map[string]interface{}{
		"user_id":  validators.String().Min(1).Pattern("^usr_[a-zA-Z0-9]+$").Required(),
		"type":     validators.String().Required(),
		"channel":  validators.String().Required(),
		"title":    validators.String().Min(1).Max(200).Required(),
		"message":  validators.String().Min(1).Max(2000).Required(),
		"data":     validators.Object(map[string]interface{}{}).Optional(),
		"priority": validators.String().Optional().Default("normal"),
	}).Required()

	sendBulkNotificationBodySchema := validators.Object(map[string]interface{}{
		"user_ids":    validators.Array(validators.String().Pattern("^usr_[a-zA-Z0-9]+$")).Required(),
		"type":        validators.String().Required(),
		"channel":     validators.String().Required(),
		"title":       validators.String().Min(1).Max(200).Required(),
		"message":     validators.String().Min(1).Max(2000).Required(),
		"data":        validators.Object(map[string]interface{}{}).Optional(),
		"priority":    validators.String().Optional().Default("normal"),
		"schedule_at": validators.String().Optional(),
	}).Required()

	sendTemplatedNotificationBodySchema := validators.Object(map[string]interface{}{
		"user_id":     validators.String().Min(1).Pattern("^usr_[a-zA-Z0-9]+$").Required(),
		"template_id": validators.String().Min(1).Pattern("^tpl_[a-zA-Z0-9]+$").Required(),
		"variables":   validators.Object(map[string]interface{}{}).Required(),
		"priority":    validators.String().Optional().Default("normal"),
		"schedule_at": validators.String().Optional(),
	}).Required()

	createTemplateBodySchema := validators.Object(map[string]interface{}{
		"name":      validators.String().Min(1).Max(100).Required(),
		"type":      validators.String().Required(),
		"channel":   validators.String().Required(),
		"subject":   validators.String().Min(1).Max(300).Required(),
		"body":      validators.String().Min(1).Max(10000).Required(),
		"variables": validators.Array(validators.String().Pattern("^[a-zA-Z_][a-zA-Z0-9_]*$")).Optional(),
		"metadata":  validators.Object(map[string]interface{}{}).Optional(),
	}).Required()

	updateTemplateBodySchema := validators.Object(map[string]interface{}{
		"name":      validators.String().Min(1).Max(100).Optional(),
		"subject":   validators.String().Min(1).Max(300).Optional(),
		"body":      validators.String().Min(1).Max(10000).Optional(),
		"variables": validators.Array(validators.String().Pattern("^[a-zA-Z_][a-zA-Z0-9_]*$")).Optional(),
		"metadata":  validators.Object(map[string]interface{}{}).Optional(),
		"is_active": validators.Bool().Optional(),
	}).Optional()

	notificationParamsSchema := validators.Object(map[string]interface{}{
		"id": validators.String().Min(1).Pattern("^notif_[a-zA-Z0-9]+$").Required(),
	}).Required()

	templateParamsSchema := validators.Object(map[string]interface{}{
		"id": validators.String().Min(1).Pattern("^tpl_[a-zA-Z0-9]+$").Required(),
	}).Required()

	listNotificationsQuerySchema := validators.Object(map[string]interface{}{
		"user_id":   validators.String().Pattern("^usr_[a-zA-Z0-9]+$").Optional(),
		"type":      validators.String().Optional(),
		"channel":   validators.String().Optional(),
		"status":    validators.String().Optional(),
		"priority":  validators.String().Optional(),
		"is_read":   validators.Bool().Optional(),
		"page":      validators.Number().Min(1).Optional().Default(1),
		"page_size": validators.Number().Min(1).Max(100).Optional().Default(20),
		"sort_by":   validators.String().Optional().Default("created_at"),
	}).Optional()

	statsQuerySchema := validators.Object(map[string]interface{}{
		"user_id":   validators.String().Pattern("^usr_[a-zA-Z0-9]+$").Optional(),
		"date_from": validators.String().Required(),
		"date_to":   validators.String().Required(),
		"group_by":  validators.String().Optional().Default("day"),
		"channel":   validators.String().Optional(),
		"type":      validators.String().Optional(),
	}).Required()

	listTemplatesQuerySchema := validators.Object(map[string]interface{}{
		"type":      validators.String().Optional(),
		"channel":   validators.String().Optional(),
		"is_active": validators.Bool().Optional(),
		"page":      validators.Number().Min(1).Optional().Default(1),
		"page_size": validators.Number().Min(1).Max(50).Optional().Default(20),
	}).Optional()

	notificationResponseSchema := validators.Object(map[string]interface{}{
		"id":         validators.String().Min(1).Required(),
		"user_id":    validators.String().Min(1).Required(),
		"type":       validators.String().Required(),
		"channel":    validators.String().Required(),
		"title":      validators.String().Min(1).Required(),
		"message":    validators.String().Min(1).Required(),
		"data":       validators.Object(map[string]interface{}{}).Optional(),
		"priority":   validators.String().Required(),
		"status":     validators.String().Required(),
		"read_at":    validators.String().Optional(),
		"sent_at":    validators.String().Optional(),
		"created_at": validators.String().Required(),
		"updated_at": validators.String().Required(),
	}).Required()

	// Define operations with comprehensive documentation
	sendNotificationOp := operations.NewSimple().
		POST("/notifications").
		Summary("Send notification").
		Description("Sends a single notification to a user via specified channel").
		Tags("notifications", "messaging").
		WithBody(sendNotificationBodySchema).
		WithResponse(notificationResponseSchema).
		Handler(operations.CreateValidatedHandler(sendNotificationHandler, nil, nil, sendNotificationBodySchema, notificationResponseSchema))

	sendBulkOp := operations.NewSimple().
		POST("/notifications/bulk").
		Summary("Send bulk notifications").
		Description("Sends notifications to multiple users at once, with optional scheduling").
		Tags("notifications", "bulk", "messaging").
		WithBody(sendBulkNotificationBodySchema).
		WithResponse(validators.Object(map[string]interface{}{
			"job_id":       validators.String().Min(1).Required(),
			"total_count":  validators.Number().Min(1).Required(),
			"scheduled_at": validators.String().Optional(),
			"status":       validators.String().Required(),
		}).Required()).
		Handler(operations.CreateValidatedHandler(sendBulkNotificationHandler, nil, nil, sendBulkNotificationBodySchema, nil))

	sendTemplatedOp := operations.NewSimple().
		POST("/notifications/templated").
		Summary("Send templated notification").
		Description("Sends a notification using a predefined template with variable substitution").
		Tags("notifications", "templates", "messaging").
		WithBody(sendTemplatedNotificationBodySchema).
		WithResponse(notificationResponseSchema).
		Handler(operations.CreateValidatedHandler(sendTemplatedNotificationHandler, nil, nil, sendTemplatedNotificationBodySchema, notificationResponseSchema))

	getNotificationOp := operations.NewSimple().
		GET("/notifications/{id}").
		Summary("Get notification").
		Description("Retrieves a specific notification by its ID").
		Tags("notifications").
		WithParams(notificationParamsSchema).
		WithResponse(notificationResponseSchema).
		Handler(operations.CreateValidatedHandler(getNotificationHandler, notificationParamsSchema, nil, nil, notificationResponseSchema))

	markReadOp := operations.NewSimple().
		PATCH("/notifications/{id}/read").
		Summary("Mark notification as read").
		Description("Marks a notification as read by the user").
		Tags("notifications", "status").
		WithParams(notificationParamsSchema).
		WithResponse(notificationResponseSchema).
		Handler(operations.CreateValidatedHandler(markNotificationReadHandler, notificationParamsSchema, nil, nil, notificationResponseSchema))

	listNotificationsOp := operations.NewSimple().
		GET("/notifications").
		Summary("List notifications").
		Description("Retrieves notifications with filtering, pagination, and sorting").
		Tags("notifications", "filtering").
		WithQuery(listNotificationsQuerySchema).
		WithResponse(validators.Object(map[string]interface{}{
			"notifications": validators.Array(notificationResponseSchema).Required(),
			"total_count":   validators.Number().Min(0).Required(),
			"unread_count":  validators.Number().Min(0).Required(),
			"page":          validators.Number().Min(1).Required(),
			"page_size":     validators.Number().Min(1).Required(),
			"has_next":      validators.Bool().Required(),
		}).Required()).
		Handler(operations.CreateValidatedHandler(listNotificationsHandler, nil, listNotificationsQuerySchema, nil, nil))

	getStatsOp := operations.NewSimple().
		GET("/analytics/notifications").
		Summary("Get notification statistics").
		Description("Retrieves comprehensive notification analytics and statistics").
		Tags("analytics", "statistics", "reporting").
		WithQuery(statsQuerySchema).
		WithResponse(validators.Object(map[string]interface{}{
			"total_sent":       validators.Number().Min(0).Required(),
			"total_read":       validators.Number().Min(0).Required(),
			"read_rate":        validators.Number().Min(0).Max(100).Required(),
			"channel_stats":    validators.Array(validators.Object(map[string]interface{}{})).Required(),
			"type_stats":       validators.Array(validators.Object(map[string]interface{}{})).Required(),
			"time_series_data": validators.Array(validators.Object(map[string]interface{}{})).Required(),
		}).Required()).
		Handler(operations.CreateValidatedHandler(getNotificationStatsHandler, nil, statsQuerySchema, nil, nil))

	createTemplateOp := operations.NewSimple().
		POST("/templates").
		Summary("Create notification template").
		Description("Creates a new notification template for reuse").
		Tags("templates", "management").
		WithBody(createTemplateBodySchema).
		WithResponse(validators.Object(map[string]interface{}{
			"id":         validators.String().Min(1).Required(),
			"name":       validators.String().Min(1).Required(),
			"type":       validators.String().Required(),
			"channel":    validators.String().Required(),
			"subject":    validators.String().Min(1).Required(),
			"body":       validators.String().Min(1).Required(),
			"variables":  validators.Array(validators.String()).Required(),
			"metadata":   validators.Object(map[string]interface{}{}).Optional(),
			"is_active":  validators.Bool().Required(),
			"created_at": validators.String().Required(),
			"updated_at": validators.String().Required(),
		}).Required()).
		Handler(operations.CreateValidatedHandler(createTemplateHandler, nil, nil, createTemplateBodySchema, nil))

	getTemplateOp := operations.NewSimple().
		GET("/templates/{id}").
		Summary("Get template").
		Description("Retrieves a specific template by its ID").
		Tags("templates").
		WithParams(templateParamsSchema).
		WithResponse(validators.Object(map[string]interface{}{
			"id":         validators.String().Min(1).Required(),
			"name":       validators.String().Min(1).Required(),
			"type":       validators.String().Required(),
			"channel":    validators.String().Required(),
			"subject":    validators.String().Min(1).Required(),
			"body":       validators.String().Min(1).Required(),
			"variables":  validators.Array(validators.String()).Required(),
			"metadata":   validators.Object(map[string]interface{}{}).Optional(),
			"is_active":  validators.Bool().Required(),
			"created_at": validators.String().Required(),
			"updated_at": validators.String().Required(),
		}).Required()).
		Handler(operations.CreateValidatedHandler(getTemplateHandler, templateParamsSchema, nil, nil, nil))

	updateTemplateOp := operations.NewSimple().
		PUT("/templates/{id}").
		Summary("Update template").
		Description("Updates an existing template").
		Tags("templates", "management").
		WithParams(templateParamsSchema).
		WithBody(updateTemplateBodySchema).
		WithResponse(validators.Object(map[string]interface{}{
			"id":         validators.String().Min(1).Required(),
			"name":       validators.String().Min(1).Required(),
			"type":       validators.String().Required(),
			"channel":    validators.String().Required(),
			"subject":    validators.String().Min(1).Required(),
			"body":       validators.String().Min(1).Required(),
			"variables":  validators.Array(validators.String()).Required(),
			"metadata":   validators.Object(map[string]interface{}{}).Optional(),
			"is_active":  validators.Bool().Required(),
			"created_at": validators.String().Required(),
			"updated_at": validators.String().Required(),
		}).Required()).
		Handler(operations.CreateValidatedHandler(updateTemplateHandler, templateParamsSchema, nil, updateTemplateBodySchema, nil))

	deleteTemplateOp := operations.NewSimple().
		DELETE("/templates/{id}").
		Summary("Delete template").
		Description("Deletes a template").
		Tags("templates", "management").
		WithParams(templateParamsSchema).
		Handler(operations.CreateValidatedHandler(deleteTemplateHandler, templateParamsSchema, nil, nil, nil))

	listTemplatesOp := operations.NewSimple().
		GET("/templates").
		Summary("List templates").
		Description("Retrieves templates with filtering and pagination").
		Tags("templates", "filtering").
		WithQuery(listTemplatesQuerySchema).
		WithResponse(validators.Object(map[string]interface{}{
			"templates":   validators.Array(validators.Object(map[string]interface{}{})).Required(),
			"total_count": validators.Number().Min(0).Required(),
			"page":        validators.Number().Min(1).Required(),
			"page_size":   validators.Number().Min(1).Required(),
			"has_next":    validators.Bool().Required(),
		}).Required()).
		Handler(operations.CreateValidatedHandler(listTemplatesHandler, nil, listTemplatesQuerySchema, nil, nil))

	// Register all operations
	router.Register(sendNotificationOp)
	router.Register(sendBulkOp)
	router.Register(sendTemplatedOp)
	router.Register(getNotificationOp)
	router.Register(markReadOp)
	router.Register(listNotificationsOp)
	router.Register(getStatsOp)
	router.Register(createTemplateOp)
	router.Register(getTemplateOp)
	router.Register(updateTemplateOp)
	router.Register(deleteTemplateOp)
	router.Register(listTemplatesOp)

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "notification-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	fmt.Println("🚀 Notification Service starting on :8003")
	fmt.Println("📚 Generate OpenAPI spec: go-op generate -i ./examples/notification-service -o ./notification-service.yaml")
	engine.Run(":8003")
}
