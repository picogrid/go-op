package validators

import (
	"errors"
	"testing"
)

func TestStringValidator_Default(t *testing.T) {
	// Default is only available on optional builders
	v := String().Optional().Default("default_value")

	// Test with nil input
	err := v.Validate(nil)
	if err != nil {
		t.Errorf("Expected no error for nil input with default, but got %v", err)
	}

	// Test with empty string input
	err = v.Validate("")
	if err != nil {
		t.Errorf("Expected no error for empty string input with default, but got %v", err)
	}

	// Test with a non-empty value
	err = v.Validate("some_value")
	if err != nil {
		t.Errorf("Expected no error for non-empty input, but got %v", err)
	}
}

func TestStringValidator_Pattern(t *testing.T) {
	// Test optional pattern
	vOptional := String().Pattern(`^[a-z]+$`).Optional()

	// Test with a valid string
	err := vOptional.Validate("abc")
	if err != nil {
		t.Errorf("Expected no error for valid string, but got %v", err)
	}

	// Test with an invalid string
	err = vOptional.Validate("aBc")
	if err == nil {
		t.Errorf("Expected an error for invalid string, but got nil")
	}

	// Test with another invalid string
	err = vOptional.Validate("ab1")
	if err == nil {
		t.Errorf("Expected an error for invalid string, but got nil")
	}

	// Test with nil
	err = vOptional.Validate(nil)
	if err != nil {
		t.Errorf("Expected no error for nil input on optional field, but got %v", err)
	}

	// Test required pattern
	vRequired := String().Pattern(`^[a-z]+$`).Required()
	err = vRequired.Validate(nil)
	if err == nil {
		t.Errorf("Expected an error for nil input on required field, but got nil")
	}

	err = vRequired.Validate("abc")
	if err != nil {
		t.Errorf("Expected no error for valid string on required field, but got %v", err)
	}
}

func TestStringValidator_URL(t *testing.T) {
	v := String().URL().Required()

	// Test with a valid URL
	err := v.Validate("https://example.com")
	if err != nil {
		t.Errorf("Expected no error for valid URL, but got %v", err)
	}

	// Test with an invalid URL
	err = v.Validate("not-a-url")
	if err == nil {
		t.Errorf("Expected an error for invalid URL, but got nil")
	}
}

func TestStringValidator_CustomMessages(t *testing.T) {
	v := String().Min(5).WithMinLengthMessage("too short").Required()
	err := v.Validate("abc")
	if err == nil {
		t.Errorf("Expected an error for string shorter than min length, but got nil")
	}
	expectedError := `Field: abc, Error: too short`
	if err.Error() != expectedError {
		t.Errorf("Expected custom error message '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestStringValidator_Email(t *testing.T) {
	v := String().Email().Required()

	// Test with a valid email
	err := v.Validate("test@example.com")
	if err != nil {
		t.Errorf("Expected no error for valid email, but got %v", err)
	}

	// Test with an invalid email
	err = v.Validate("not-an-email")
	if err == nil {
		t.Errorf("Expected an error for invalid email, but got nil")
	}
}

func TestStringValidator_Custom(t *testing.T) {
	customErr := errors.New("custom validation failed")
	v := String().Custom(func(s string) error {
		if s == "invalid" {
			return customErr
		}
		return nil
	}).Required()

	// Test with a valid string
	err := v.Validate("valid")
	if err != nil {
		t.Errorf("Expected no error for valid custom string, but got %v", err)
	}

	// Test with an invalid string
	err = v.Validate("invalid")
	if err != customErr {
		t.Errorf("Expected custom error, but got %v", err)
	}
}

func TestStringValidator_InvalidType(t *testing.T) {
	v := String().Required()
	err := v.Validate(123)
	if err == nil {
		t.Errorf("Expected an error for invalid type, but got nil")
	}
}

func TestStringValidator_InvalidRegex(t *testing.T) {
	// This pattern is invalid because of the unclosed parenthesis
	v := String().Pattern(`(^[a-z]+$`).Required()
	err := v.Validate("abc")
	if err == nil {
		t.Errorf("Expected an error for invalid regex pattern, but got nil")
	}
}
