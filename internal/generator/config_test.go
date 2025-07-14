package generator

import (
	"testing"
)

func TestConfigStructure(t *testing.T) {
	config := &Config{
		InputDir:    "/path/to/input",
		OutputFile:  "output.yaml",
		Format:      "json",
		Title:       "Test API",
		Version:     "2.0.0",
		Description: "Test API Description",
		Servers:     []string{"https://api.example.com", "http://localhost:8080"},
		Verbose:     true,
	}
	
	if config.InputDir != "/path/to/input" {
		t.Errorf("Expected InputDir to be '/path/to/input', got '%s'", config.InputDir)
	}
	
	if config.OutputFile != "output.yaml" {
		t.Errorf("Expected OutputFile to be 'output.yaml', got '%s'", config.OutputFile)
	}
	
	if config.Format != "json" {
		t.Errorf("Expected Format to be 'json', got '%s'", config.Format)
	}
	
	if config.Title != "Test API" {
		t.Errorf("Expected Title to be 'Test API', got '%s'", config.Title)
	}
	
	if config.Version != "2.0.0" {
		t.Errorf("Expected Version to be '2.0.0', got '%s'", config.Version)
	}
	
	if config.Description != "Test API Description" {
		t.Errorf("Expected Description to be 'Test API Description', got '%s'", config.Description)
	}
	
	if len(config.Servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(config.Servers))
	}
	
	if !config.Verbose {
		t.Errorf("Expected Verbose to be true")
	}
}

func TestGenerationStats(t *testing.T) {
	stats := GenerationStats{
		OperationCount: 10,
		SchemaCount:    5,
		PathCount:      7,
		FileCount:      3,
	}
	
	if stats.OperationCount != 10 {
		t.Errorf("Expected OperationCount to be 10, got %d", stats.OperationCount)
	}
	
	if stats.SchemaCount != 5 {
		t.Errorf("Expected SchemaCount to be 5, got %d", stats.SchemaCount)
	}
	
	if stats.PathCount != 7 {
		t.Errorf("Expected PathCount to be 7, got %d", stats.PathCount)
	}
	
	if stats.FileCount != 3 {
		t.Errorf("Expected FileCount to be 3, got %d", stats.FileCount)
	}
}

func TestConfigDefaults(t *testing.T) {
	// Test zero-value Config
	var config Config
	
	if config.InputDir != "" {
		t.Errorf("Expected InputDir to default to empty string")
	}
	
	if config.OutputFile != "" {
		t.Errorf("Expected OutputFile to default to empty string")
	}
	
	if config.Format != "" {
		t.Errorf("Expected Format to default to empty string")
	}
	
	if config.Title != "" {
		t.Errorf("Expected Title to default to empty string")
	}
	
	if config.Version != "" {
		t.Errorf("Expected Version to default to empty string")
	}
	
	if config.Description != "" {
		t.Errorf("Expected Description to default to empty string")
	}
	
	if len(config.Servers) != 0 {
		t.Errorf("Expected Servers to default to empty slice")
	}
	
	if config.Verbose {
		t.Errorf("Expected Verbose to default to false")
	}
}

func TestGenerationStatsDefaults(t *testing.T) {
	// Test zero-value GenerationStats
	var stats GenerationStats
	
	if stats.OperationCount != 0 {
		t.Errorf("Expected OperationCount to default to 0")
	}
	
	if stats.SchemaCount != 0 {
		t.Errorf("Expected SchemaCount to default to 0")
	}
	
	if stats.PathCount != 0 {
		t.Errorf("Expected PathCount to default to 0")
	}
	
	if stats.FileCount != 0 {
		t.Errorf("Expected FileCount to default to 0")
	}
}

func TestConfigWithPartialValues(t *testing.T) {
	// Test Config with only some fields set
	config := &Config{
		InputDir:   "/input",
		OutputFile: "api.yaml",
		Version:    "1.0.0",
	}
	
	// Verify set fields
	if config.InputDir != "/input" {
		t.Errorf("Expected InputDir to be '/input', got '%s'", config.InputDir)
	}
	
	if config.OutputFile != "api.yaml" {
		t.Errorf("Expected OutputFile to be 'api.yaml', got '%s'", config.OutputFile)
	}
	
	if config.Version != "1.0.0" {
		t.Errorf("Expected Version to be '1.0.0', got '%s'", config.Version)
	}
	
	// Verify unset fields remain at defaults
	if config.Format != "" {
		t.Errorf("Expected Format to be empty, got '%s'", config.Format)
	}
	
	if config.Title != "" {
		t.Errorf("Expected Title to be empty, got '%s'", config.Title)
	}
	
	if config.Description != "" {
		t.Errorf("Expected Description to be empty, got '%s'", config.Description)
	}
	
	if len(config.Servers) != 0 {
		t.Errorf("Expected Servers to be empty, got %v", config.Servers)
	}
	
	if config.Verbose {
		t.Errorf("Expected Verbose to be false")
	}
}