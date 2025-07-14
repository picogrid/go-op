package validators

import (
	"errors"
	"testing"
)

func TestArrayValidator_Contains(t *testing.T) {
	v := Array(String()).Contains("must_have").Required()
	if err := v.Validate([]string{"a", "must_have", "b"}); err != nil {
		t.Errorf("Expected no error for array containing the value, but got %v", err)
	}
	if err := v.Validate([]string{"a", "b"}); err == nil {
		t.Errorf("Expected an error for array not containing the value, but got nil")
	}
}

func TestArrayValidator_NestedValidation(t *testing.T) {
	// Test an array of strings where each string must be an email
	v := Array(String().Email()).Required()
	if err := v.Validate([]string{"test@example.com", "another@test.com"}); err != nil {
		t.Errorf("Expected no error for valid nested elements, but got %v", err)
	}

	err := v.Validate([]string{"test@example.com", "not-an-email"})
	if err == nil {
		t.Errorf("Expected an error for invalid nested element, but got nil")
	}
}

func TestArrayValidator_CustomMessages(t *testing.T) {
	v := Array(String()).MaxItems(2).WithMaxItemsMessage("too many items").Required()
	err := v.Validate([]string{"a", "b", "c"})
	if err == nil {
		t.Errorf("Expected an error for array with too many items, but got nil")
	}
	expectedError := `Field: [a b c], Error: too many items`
	if err.Error() != expectedError {
		t.Errorf("Expected custom error message '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestArrayValidator_InvalidType(t *testing.T) {
	v := Array(String()).Required()
	err := v.Validate("not an array")
	if err == nil {
		t.Errorf("Expected an error for invalid type, but got nil")
	}
}

func TestArrayValidator_Custom(t *testing.T) {
	customErr := errors.New("custom validation failed")
	v := Array(String()).Custom(func(arr []interface{}) error {
		if len(arr) == 0 {
			return customErr
		}
		return nil
	}).Required()

	if err := v.Validate([]string{"a"}); err != nil {
		t.Errorf("Expected no error for valid custom array, but got %v", err)
	}

	if err := v.Validate([]string{}); err != customErr {
		t.Errorf("Expected custom error, but got %v", err)
	}
}
