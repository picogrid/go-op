package validators

import (
	"testing"
)

type TestUser struct {
	Username string `json:"username"`
	Age      int    `json:"age"`
}

func TestStructValidator(t *testing.T) {
	userSchema := StructValidator(func(u *TestUser) map[string]interface{} {
		return map[string]interface{}{
			"username": String().Required(),
			"age":      Number().Min(18).Required(),
		}
	})

	validData := map[string]interface{}{"username": "test", "age": 21}
	if err := userSchema.Validate(validData); err != nil {
		t.Errorf("Expected no error for valid struct data, but got %v", err)
	}

	invalidData := map[string]interface{}{"username": "test", "age": 17}
	if err := userSchema.Validate(invalidData); err == nil {
		t.Errorf("Expected an error for invalid struct data, but got nil")
	}
}

func TestValidateStruct(t *testing.T) {
	userSchema := ForStruct[TestUser]().
		Field("username", String().Required()).
		Field("age", Number().Min(18).Required()).
		Build()

	validData := map[string]interface{}{"username": "test", "age": 21}
	user, err := ValidateStruct[TestUser](userSchema, validData)
	if err != nil {
		t.Errorf("Expected no error for valid struct validation, but got %v", err)
	}
	if user.Username != "test" || user.Age != 21 {
		t.Errorf("Struct fields not populated correctly, got %+v", user)
	}
}

func TestForStructBuilder_Strict(t *testing.T) {
	userSchema := ForStruct[TestUser]().
		Field("username", String().Required()).
		Strict().
		Build()

	invalidData := map[string]interface{}{"username": "test", "age": 21}
	err := userSchema.Validate(invalidData)
	if err == nil {
		t.Errorf("Expected an error for unknown field in strict mode, but got nil")
	}
}
