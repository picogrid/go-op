package validators

import (
	"errors"
	"testing"
)

func TestNumberValidator_PositiveNegative(t *testing.T) {
	// Test Positive
	vPositive := Number().Positive().Required()
	if err := vPositive.Validate(10); err != nil {
		t.Errorf("Expected no error for positive number, but got %v", err)
	}
	if err := vPositive.Validate(-10); err == nil {
		t.Errorf("Expected an error for non-positive number, but got nil")
	}
	if err := vPositive.Validate(0); err == nil {
		t.Errorf("Expected an error for zero, but got nil")
	}

	// Test Negative
	vNegative := Number().Negative().Required()
	if err := vNegative.Validate(-10); err != nil {
		t.Errorf("Expected no error for negative number, but got %v", err)
	}
	if err := vNegative.Validate(10); err == nil {
		t.Errorf("Expected an error for non-negative number, but got nil")
	}
	if err := vNegative.Validate(0); err == nil {
		t.Errorf("Expected an error for zero, but got nil")
	}
}

func TestNumberValidator_Integer(t *testing.T) {
	v := Number().Integer().Required()
	if err := v.Validate(10); err != nil {
		t.Errorf("Expected no error for integer, but got %v", err)
	}
	if err := v.Validate(10.5); err == nil {
		t.Errorf("Expected an error for non-integer, but got nil")
	}
}

func TestNumberValidator_CustomMessages(t *testing.T) {
	v := Number().Max(5).WithMaxMessage("too big").Required()
	err := v.Validate(10)
	if err == nil {
		t.Errorf("Expected an error for number greater than max, but got nil")
	}
	expectedError := `Field: 10, Error: too big`
	if err.Error() != expectedError {
		t.Errorf("Expected custom error message '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestNumberValidator_InvalidType(t *testing.T) {
	v := Number().Required()
	err := v.Validate("not a number")
	if err == nil {
		t.Errorf("Expected an error for invalid type, but got nil")
	}
}

func TestNumberValidator_Custom(t *testing.T) {
	customErr := errors.New("custom validation failed")
	v := Number().Custom(func(n float64) error {
		if n == 13 {
			return customErr
		}
		return nil
	}).Required()

	if err := v.Validate(10); err != nil {
		t.Errorf("Expected no error for valid custom number, but got %v", err)
	}

	if err := v.Validate(13); err != customErr {
		t.Errorf("Expected custom error, but got %v", err)
	}
}
