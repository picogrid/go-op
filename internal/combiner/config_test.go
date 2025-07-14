package combiner

import (
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	// Test default constants
	if DefaultTitle != "Combined API" {
		t.Errorf("Expected DefaultTitle to be 'Combined API', got '%s'", DefaultTitle)
	}

	if DefaultVersion != "1.0.0" {
		t.Errorf("Expected DefaultVersion to be '1.0.0', got '%s'", DefaultVersion)
	}

	if DefaultFormat != "yaml" {
		t.Errorf("Expected DefaultFormat to be 'yaml', got '%s'", DefaultFormat)
	}

	if DefaultOutputFile != "combined-api.yaml" {
		t.Errorf("Expected DefaultOutputFile to be 'combined-api.yaml', got '%s'", DefaultOutputFile)
	}

	if !DefaultMergeSchemas {
		t.Errorf("Expected DefaultMergeSchemas to be true")
	}

	if !DefaultValidateOutput {
		t.Errorf("Expected DefaultValidateOutput to be true")
	}
}

func TestConfigStructure(t *testing.T) {
	// Test Config initialization
	config := &Config{
		OutputFile:     "test.yaml",
		Format:         "json",
		Title:          "Test API",
		Version:        "2.0.0",
		BaseURL:        "/api/v2",
		ConfigFile:     "services.yaml",
		ServicePrefix:  map[string]string{"user": "/users"},
		IncludeTags:    []string{"public"},
		ExcludeTags:    []string{"internal"},
		MergeSchemas:   true,
		ValidateOutput: true,
		Verbose:        true,
	}

	if config.OutputFile != "test.yaml" {
		t.Errorf("Expected OutputFile to be 'test.yaml', got '%s'", config.OutputFile)
	}

	if config.ServicePrefix["user"] != "/users" {
		t.Errorf("Expected ServicePrefix['user'] to be '/users', got '%s'", config.ServicePrefix["user"])
	}
}

func TestServicesConfig(t *testing.T) {
	// Test ServicesConfig structure
	servicesConfig := ServicesConfig{
		Title:       "Platform API",
		Version:     "3.0.0",
		Description: "Combined platform API",
		BaseURL:     "/api/v3",
		Services: []ServiceConfig{
			{
				Name:        "user",
				SpecFile:    "user.yaml",
				PathPrefix:  "/users",
				Tags:        []string{"users", "auth"},
				Enabled:     true,
				Description: "User management service",
				HealthCheck: "/users/health",
				Version:     "1.0.0",
			},
			{
				Name:        "order",
				SpecFile:    "order.yaml",
				PathPrefix:  "/orders",
				Tags:        []string{"orders"},
				Enabled:     false,
				Description: "Order management service",
				HealthCheck: "/orders/health",
				Version:     "2.0.0",
			},
		},
		Settings: CombinationSettings{
			MergeSchemas:     true,
			ValidateOutput:   true,
			IncludeTags:      []string{"public"},
			ExcludeTags:      []string{"internal"},
			ConflictStrategy: "override",
		},
	}

	if servicesConfig.Title != "Platform API" {
		t.Errorf("Expected Title to be 'Platform API', got '%s'", servicesConfig.Title)
	}

	if len(servicesConfig.Services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(servicesConfig.Services))
	}

	// Test first service
	userService := servicesConfig.Services[0]
	if userService.Name != "user" {
		t.Errorf("Expected first service name to be 'user', got '%s'", userService.Name)
	}

	if !userService.Enabled {
		t.Errorf("Expected user service to be enabled")
	}

	if len(userService.Tags) != 2 {
		t.Errorf("Expected user service to have 2 tags, got %d", len(userService.Tags))
	}

	// Test second service
	orderService := servicesConfig.Services[1]
	if orderService.Name != "order" {
		t.Errorf("Expected second service name to be 'order', got '%s'", orderService.Name)
	}

	if orderService.Enabled {
		t.Errorf("Expected order service to be disabled")
	}

	// Test settings
	if servicesConfig.Settings.ConflictStrategy != "override" {
		t.Errorf("Expected ConflictStrategy to be 'override', got '%s'", servicesConfig.Settings.ConflictStrategy)
	}
}

func TestCombinationStats(t *testing.T) {
	stats := CombinationStats{
		InputFiles:       5,
		ServicesCombined: 4,
		TotalOperations:  20,
		TotalPaths:       10,
		MergedSchemas:    15,
		Conflicts:        2,
	}

	if stats.InputFiles != 5 {
		t.Errorf("Expected InputFiles to be 5, got %d", stats.InputFiles)
	}

	if stats.ServicesCombined != 4 {
		t.Errorf("Expected ServicesCombined to be 4, got %d", stats.ServicesCombined)
	}

	if stats.TotalOperations != 20 {
		t.Errorf("Expected TotalOperations to be 20, got %d", stats.TotalOperations)
	}

	if stats.TotalPaths != 10 {
		t.Errorf("Expected TotalPaths to be 10, got %d", stats.TotalPaths)
	}

	if stats.MergedSchemas != 15 {
		t.Errorf("Expected MergedSchemas to be 15, got %d", stats.MergedSchemas)
	}

	if stats.Conflicts != 2 {
		t.Errorf("Expected Conflicts to be 2, got %d", stats.Conflicts)
	}
}

func TestServiceConfigDefaults(t *testing.T) {
	// Test zero-value ServiceConfig
	var service ServiceConfig

	if service.Enabled {
		t.Errorf("Expected Enabled to default to false")
	}

	if service.Name != "" {
		t.Errorf("Expected Name to default to empty string")
	}

	if service.PathPrefix != "" {
		t.Errorf("Expected PathPrefix to default to empty string")
	}

	if len(service.Tags) != 0 {
		t.Errorf("Expected Tags to default to empty slice")
	}
}

func TestCombinationSettingsDefaults(t *testing.T) {
	// Test zero-value CombinationSettings
	var settings CombinationSettings

	if settings.MergeSchemas {
		t.Errorf("Expected MergeSchemas to default to false")
	}

	if settings.ValidateOutput {
		t.Errorf("Expected ValidateOutput to default to false")
	}

	if settings.ConflictStrategy != "" {
		t.Errorf("Expected ConflictStrategy to default to empty string")
	}

	if len(settings.IncludeTags) != 0 {
		t.Errorf("Expected IncludeTags to default to empty slice")
	}

	if len(settings.ExcludeTags) != 0 {
		t.Errorf("Expected ExcludeTags to default to empty slice")
	}
}
