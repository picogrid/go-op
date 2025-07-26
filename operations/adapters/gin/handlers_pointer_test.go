package gin_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	ginadapter "github.com/picogrid/go-op/operations/adapters/gin"
	"github.com/picogrid/go-op/validators"
)

// TestCreateValidatedHandlerWithPointerFields tests the fix for the issue where
// ForStruct validators were failing with pointer fields, showing error messages like:
// "Field: {0x14000952180}, Error: invalid type, expected object"
func TestCreateValidatedHandlerWithPointerFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Define the exact structs from the issue report
	type Viewport struct {
		Bearing   *float64 `json:"bearing,omitempty"`
		Latitude  *float64 `json:"latitude,omitempty"`
		Longitude *float64 `json:"longitude,omitempty"`
		Pitch     *float64 `json:"pitch,omitempty"`
		Zoom      *float64 `json:"zoom,omitempty"`
	}

	type UpdateOrgSettingsRequest struct {
		Viewport *Viewport `json:"viewport,omitempty"`
	}

	type OrgSettingsResponse struct {
		OrganizationID string    `json:"organization_id"`
		Viewport       *Viewport `json:"viewport,omitempty"`
	}

	// Create validators exactly as shown in the issue
	ViewportValidator := validators.ForStruct[Viewport]().
		Field("bearing", validators.Number().Min(0).Max(360).Optional()).
		Field("latitude", validators.Number().Min(-90).Max(90).Optional()).
		Field("longitude", validators.Number().Min(-180).Max(180).Optional()).
		Field("pitch", validators.Number().Min(0).Max(60).Optional()).
		Field("zoom", validators.Number().Min(0).Max(24).Optional()).
		Optional()

	UpdateOrgSettingsRequestValidator := validators.ForStruct[UpdateOrgSettingsRequest]().
		Field("viewport", ViewportValidator)

	OrgSettingsResponseValidator := validators.ForStruct[OrgSettingsResponse]().
		Field("organization_id", validators.String().Required()).
		Field("viewport", ViewportValidator)

	// Create handler function as would be used in the issue scenario
	updateOrgSettings := func(
		ctx context.Context,
		params struct{},
		query struct{},
		body UpdateOrgSettingsRequest,
	) (OrgSettingsResponse, error) {
		return OrgSettingsResponse{
			OrganizationID: "org_123",
			Viewport:       body.Viewport,
		}, nil
	}

	// Create the validated handler - this is where the issue was occurring
	handler := ginadapter.CreateValidatedHandler(
		updateOrgSettings,
		nil, // params schema
		nil, // query schema
		UpdateOrgSettingsRequestValidator.Schema(), // This was failing
		OrgSettingsResponseValidator.Schema(),
	)

	t.Run("Valid Request - should pass", func(t *testing.T) {
		router := gin.New()
		router.POST("/settings", handler)

		reqBody := `{
			"viewport": {
				"bearing": 90.0,
				"latitude": 35.0,
				"longitude": -115.0
			}
		}`

		req := httptest.NewRequest("POST", "/settings", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK for valid request")

		var response OrgSettingsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "org_123", response.OrganizationID)
		assert.NotNil(t, response.Viewport)
		assert.Equal(t, 90.0, *response.Viewport.Bearing)
	})

	t.Run("Invalid Request - bearing exceeds max", func(t *testing.T) {
		router := gin.New()
		router.POST("/settings", handler)

		reqBody := `{
			"viewport": {
				"bearing": 400.0
			}
		}`

		req := httptest.NewRequest("POST", "/settings", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected 400 for invalid data")

		var errorResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "Request body validation failed", errorResp["error"])

		// Key assertions: error should be clear, not show memory addresses
		details := errorResp["details"].(string)
		assert.Contains(t, details, "viewport.bearing", "Error should mention field path")
		assert.Contains(t, details, "360", "Error should mention constraint")
		assert.NotContains(t, details, "0x", "Error should NOT contain memory addresses")
		assert.NotContains(t, details, "{0x", "Error should NOT contain pointer format")

		t.Logf("Clear error message: %s", details)
	})

	t.Run("Empty viewport - all optional fields", func(t *testing.T) {
		router := gin.New()
		router.POST("/settings", handler)

		reqBody := `{"viewport": {}}`

		req := httptest.NewRequest("POST", "/settings", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Empty viewport should be valid")
	})

	t.Run("Missing viewport - optional field", func(t *testing.T) {
		router := gin.New()
		router.POST("/settings", handler)

		reqBody := `{}`

		req := httptest.NewRequest("POST", "/settings", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Missing optional viewport should be valid")
	})

	t.Run("Multiple invalid fields", func(t *testing.T) {
		router := gin.New()
		router.POST("/settings", handler)

		reqBody := `{
			"viewport": {
				"bearing": 400.0,
				"latitude": 100.0,
				"pitch": 70.0
			}
		}`

		req := httptest.NewRequest("POST", "/settings", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		assert.NoError(t, err)

		details := errorResp["details"].(string)
		assert.Contains(t, details, "viewport.bearing")
		assert.Contains(t, details, "viewport.latitude")
		assert.Contains(t, details, "viewport.pitch")
		assert.NotContains(t, details, "0x", "No memory addresses in errors")

		t.Logf("Multiple field errors: %s", details)
	})
}

// TestRequiredPointerFields tests validation with required pointer fields
func TestRequiredPointerFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type Config struct {
		APIKey *string `json:"api_key"`
		Port   *int    `json:"port"`
	}

	ConfigValidator := validators.ForStruct[Config]().
		Field("api_key", validators.String().Min(10).Required()).
		Field("port", validators.Number().Min(1000).Max(9999).Required())

	handler := ginadapter.CreateValidatedHandler(
		func(ctx context.Context, _ struct{}, _ struct{}, body Config) (Config, error) {
			return body, nil
		},
		nil,
		nil,
		ConfigValidator.Schema(),
		ConfigValidator.Schema(),
	)

	t.Run("Missing required pointer fields", func(t *testing.T) {
		router := gin.New()
		router.POST("/config", handler)

		reqBody := `{}`
		req := httptest.NewRequest("POST", "/config", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &errorResp)

		details := errorResp["details"].(string)
		assert.Contains(t, details, "api_key")
		assert.Contains(t, details, "port")
		assert.Contains(t, details, "required")
		assert.NotContains(t, details, "0x")
	})

	t.Run("Valid required pointer fields", func(t *testing.T) {
		router := gin.New()
		router.POST("/config", handler)

		reqBody := `{"api_key": "valid_key_123", "port": 8080}`
		req := httptest.NewRequest("POST", "/config", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Config
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response.APIKey)
		assert.Equal(t, "valid_key_123", *response.APIKey)
		assert.NotNil(t, response.Port)
		assert.Equal(t, 8080, *response.Port)
	})
}
