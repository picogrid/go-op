package validators_test

import (
	"encoding/json"
	"testing"

	"github.com/picogrid/go-op/validators"
)

// Test structs
type User struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Age      int    `json:"age"`
}

type Profile struct {
	Bio    string `json:"bio"`
	Avatar string `json:"avatar"`
}

type ComplexUser struct {
	ID      string   `json:"id"`
	Email   string   `json:"email"`
	Profile *Profile `json:"profile"`
	Tags    []string `json:"tags"`
	Active  bool     `json:"active"`
}

func TestStructValidator(t *testing.T) {
	// Create a schema using StructValidator
	userSchema := validators.StructValidator(func(u *User) map[string]interface{} {
		return map[string]interface{}{
			"email":    validators.Email(),
			"username": validators.String().Min(3).Max(50).Required(),
			"age":      validators.Number().Min(18).Max(120).Required(),
		}
	})

	// Helper function to convert struct to map via JSON
	structToMap := func(v interface{}) map[string]interface{} {
		data, _ := json.Marshal(v)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		return m
	}

	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name: "valid user struct",
			data: structToMap(User{
				Email:    "test@example.com",
				Username: "testuser",
				Age:      25,
			}),
			wantErr: false,
		},
		{
			name: "valid user map",
			data: map[string]interface{}{
				"email":    "test@example.com",
				"username": "testuser",
				"age":      25,
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			data: structToMap(User{
				Email:    "invalid-email",
				Username: "testuser",
				Age:      25,
			}),
			wantErr: true,
		},
		{
			name: "username too short",
			data: structToMap(User{
				Email:    "test@example.com",
				Username: "ab",
				Age:      25,
			}),
			wantErr: true,
		},
		{
			name: "age too young",
			data: structToMap(User{
				Email:    "test@example.com",
				Username: "testuser",
				Age:      17,
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := userSchema.Validate(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStruct(t *testing.T) {
	userSchema := validators.StructValidator(func(u *User) map[string]interface{} {
		return map[string]interface{}{
			"email":    validators.Email(),
			"username": validators.String().Min(3).Max(50).Required(),
			"age":      validators.Number().Min(18).Max(120).Required(),
		}
	})

	t.Run("valid struct value", func(t *testing.T) {
		user := User{
			Email:    "test@example.com",
			Username: "testuser",
			Age:      25,
		}

		result, err := validators.ValidateStruct[User](userSchema, user)
		if err != nil {
			t.Errorf("ValidateStruct() unexpected error: %v", err)
		}
		if result == nil {
			t.Error("ValidateStruct() returned nil result")
		} else if result.Email != user.Email || result.Username != user.Username || result.Age != user.Age {
			t.Error("ValidateStruct() returned incorrect data")
		}
	})

	t.Run("valid struct pointer", func(t *testing.T) {
		user := &User{
			Email:    "test@example.com",
			Username: "testuser",
			Age:      25,
		}

		result, err := validators.ValidateStruct[User](userSchema, user)
		if err != nil {
			t.Errorf("ValidateStruct() unexpected error: %v", err)
		}
		if result == nil {
			t.Error("ValidateStruct() returned nil result")
		} else if result.Email != user.Email {
			t.Error("ValidateStruct() returned incorrect data")
		}
	})

	t.Run("invalid data type", func(t *testing.T) {
		data := "not a user"
		_, err := validators.ValidateStruct[User](userSchema, data)
		if err == nil {
			t.Error("ValidateStruct() expected type error")
		}
	})

	t.Run("validation error", func(t *testing.T) {
		user := User{
			Email:    "invalid-email",
			Username: "testuser",
			Age:      25,
		}

		_, err := validators.ValidateStruct[User](userSchema, user)
		if err == nil {
			t.Error("ValidateStruct() expected validation error")
		}
	})
}

func TestForStruct(t *testing.T) {
	// Helper function to convert struct to map via JSON
	structToMap := func(v interface{}) map[string]interface{} {
		data, _ := json.Marshal(v)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		return m
	}

	t.Run("basic builder usage", func(t *testing.T) {
		schema := validators.ForStruct[User]().
			Field("email", validators.Email()).
			Field("username", validators.String().Min(3).Max(50).Required()).
			Field("age", validators.Number().Min(18).Max(120).Required()).
			Required()

		user := User{
			Email:    "test@example.com",
			Username: "testuser",
			Age:      25,
		}

		err := schema.Build().Validate(structToMap(user))
		if err != nil {
			t.Errorf("Build() validation error: %v", err)
		}
	})

	t.Run("builder with Fields method", func(t *testing.T) {
		schema := validators.ForStruct[User]().
			Fields(map[string]interface{}{
				"email":    validators.Email(),
				"username": validators.String().Min(3).Max(50).Required(),
				"age":      validators.Number().Min(18).Max(120).Required(),
			}).
			Build()

		user := User{
			Email:    "test@example.com",
			Username: "testuser",
			Age:      25,
		}

		err := schema.Validate(structToMap(user))
		if err != nil {
			t.Errorf("Fields() validation error: %v", err)
		}
	})

	t.Run("strict mode", func(t *testing.T) {
		schema := validators.ForStruct[User]().
			Field("email", validators.Email()).
			Field("username", validators.String().Required()).
			Strict().
			Build()

		// Test with extra field
		data := map[string]interface{}{
			"email":    "test@example.com",
			"username": "testuser",
			"extra":    "field",
		}

		err := schema.Validate(data)
		if err == nil {
			t.Error("Strict() expected error for extra field")
		}
	})

	t.Run("optional struct", func(t *testing.T) {
		schema := validators.ForStruct[User]().
			Field("email", validators.Email()).
			Optional()

		// nil should be valid for optional
		err := schema.Build().Validate(nil)
		if err != nil {
			t.Errorf("Optional() unexpected error for nil: %v", err)
		}
	})

	t.Run("custom error", func(t *testing.T) {
		schema := validators.ForStruct[User]().
			Field("email", validators.Email()).
			CustomError("required", "User information is required").
			Required()

		err := schema.Build().Validate(nil)
		if err == nil {
			t.Error("CustomError() expected error")
		}
		// Note: Checking exact error message would require accessing internal error structure
	})
}

func TestTypedValidator(t *testing.T) {
	schema := validators.ForStruct[User]().
		Field("email", validators.Email()).
		Field("username", validators.String().Min(3).Max(50).Required()).
		Field("age", validators.Number().Min(18).Max(120).Required()).
		Build()

	validateUser := validators.TypedValidator[User](schema)

	t.Run("valid user", func(t *testing.T) {
		user := User{
			Email:    "test@example.com",
			Username: "testuser",
			Age:      25,
		}

		result, err := validateUser(user)
		if err != nil {
			t.Errorf("TypedValidator() unexpected error: %v", err)
		}
		if result == nil || result.Email != user.Email {
			t.Error("TypedValidator() incorrect result")
		}
	})

	t.Run("invalid user", func(t *testing.T) {
		user := User{
			Email:    "invalid",
			Username: "testuser",
			Age:      25,
		}

		_, err := validateUser(user)
		if err == nil {
			t.Error("TypedValidator() expected validation error")
		}
	})
}

func TestMustValidateStruct(t *testing.T) {
	schema := validators.ForStruct[User]().
		Field("email", validators.Email()).
		Build()

	t.Run("valid data", func(t *testing.T) {
		user := User{Email: "test@example.com"}

		// Should not panic
		result := validators.MustValidateStruct[User](schema, user)
		if result == nil || result.Email != user.Email {
			t.Error("MustValidateStruct() incorrect result")
		}
	})

	t.Run("invalid data panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustValidateStruct() expected panic")
			}
		}()

		user := User{Email: "invalid"}
		validators.MustValidateStruct[User](schema, user)
	})
}

func TestComplexStructValidation(t *testing.T) {
	// Test with nested structs
	profileSchema := validators.Object(map[string]interface{}{
		"bio":    validators.String().Max(500).Optional(),
		"avatar": validators.String().Optional(),
	})

	complexSchema := validators.ForStruct[ComplexUser]().
		Field("id", validators.String().Required()).
		Field("email", validators.Email()).
		Field("profile", profileSchema.Optional()).
		Field("tags", validators.Array(validators.String()).Optional()).
		Field("active", validators.Bool().Required()).
		Build()

	t.Run("valid complex struct", func(t *testing.T) {
		user := ComplexUser{
			ID:    "user123",
			Email: "test@example.com",
			Profile: &Profile{
				Bio:    "Test bio",
				Avatar: "avatar.jpg",
			},
			Tags:   []string{"tag1", "tag2"},
			Active: true,
		}

		result, err := validators.ValidateStruct[ComplexUser](complexSchema, user)
		if err != nil {
			t.Errorf("Complex validation error: %v", err)
		}
		if result == nil || result.ID != user.ID {
			t.Error("Complex validation incorrect result")
		}
	})

	t.Run("nil optional fields", func(t *testing.T) {
		user := ComplexUser{
			ID:      "user123",
			Email:   "test@example.com",
			Profile: nil,
			Tags:    nil,
			Active:  true,
		}

		result, err := validators.ValidateStruct[ComplexUser](complexSchema, user)
		if err != nil {
			t.Errorf("Complex validation with nil fields error: %v", err)
		}
		if result == nil {
			t.Error("Complex validation returned nil")
		}
	})
}

func TestJSONUnmarshalWithValidation(t *testing.T) {
	schema := validators.ForStruct[User]().
		Field("email", validators.Email()).
		Field("username", validators.String().Min(3).Max(50).Required()).
		Field("age", validators.Number().Min(18).Max(120).Required()).
		Build()

	jsonData := `{
		"email": "test@example.com",
		"username": "testuser",
		"age": 25
	}`

	var user User
	err := json.Unmarshal([]byte(jsonData), &user)
	if err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}

	result, err := validators.ValidateStruct[User](schema, user)
	if err != nil {
		t.Errorf("Validation after JSON unmarshal error: %v", err)
	}
	if result == nil || result.Email != "test@example.com" {
		t.Error("Incorrect result after JSON unmarshal")
	}
}

func TestWithFields(t *testing.T) {
	// Helper function to convert struct to map via JSON
	structToMap := func(v interface{}) map[string]interface{} {
		data, _ := json.Marshal(v)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		return m
	}

	schema := validators.WithFields(
		validators.MapField[User]("email", validators.Email()),
		validators.MapField[User]("username", validators.String().Min(3).Required()),
		validators.MapField[User]("age", validators.Number().Min(18).Required()),
	)

	user := User{
		Email:    "test@example.com",
		Username: "testuser",
		Age:      25,
	}

	err := schema.Validate(structToMap(user))
	if err != nil {
		t.Errorf("WithFields() validation error: %v", err)
	}
}
