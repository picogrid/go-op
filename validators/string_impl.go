package validators

import (
	"fmt"
	"net/url"
	"regexp"

	goop "github.com/picogrid/go-op"
)

// Core string schema struct (unexported)
// This contains all the validation configuration and is wrapped by state-specific types
type stringSchema struct {
	minLength     int
	maxLength     int
	required      bool
	pattern       *regexp.Regexp
	emailFormat   bool
	urlFormat     bool
	constValue    *string
	customFunc    func(string) error
	optional      bool
	defaultValue  *string
	customError   map[string]string
	example       interface{}
	examples      map[string]ExampleObject
	externalValue string
}

// ExampleObject represents an example value with metadata
type ExampleObject struct {
	Summary       string      `json:"summary,omitempty"`
	Description   string      `json:"description,omitempty"`
	Value         interface{} `json:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty"`
}

// State wrapper types for compile-time safety
type requiredStringSchema struct {
	*stringSchema
}

type optionalStringSchema struct {
	*stringSchema
}

// StringBuilder implementation (initial state)
// These methods return StringBuilder to allow continued configuration

func (s *stringSchema) Min(length int) StringBuilder {
	s.minLength = length
	return s
}

func (s *stringSchema) Max(length int) StringBuilder {
	s.maxLength = length
	return s
}

func (s *stringSchema) Pattern(pattern string) StringBuilder {
	// Handle potential regex compilation errors gracefully
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// Instead of panicking, create a pattern that always fails
		// This allows the schema to be created but validation will fail with a clear message
		s.pattern = regexp.MustCompile(`$^`) // This pattern never matches anything
		if s.customError == nil {
			s.customError = make(map[string]string)
		}
		s.customError[errorKeys.Pattern] = fmt.Sprintf("invalid regex pattern: %v", err)
	} else {
		s.pattern = compiled
	}
	return s
}

func (s *stringSchema) Email() StringBuilder {
	s.emailFormat = true
	return s
}

func (s *stringSchema) URL() StringBuilder {
	s.urlFormat = true
	return s
}

func (s *stringSchema) Const(value string) StringBuilder {
	s.constValue = &value
	return s
}

func (s *stringSchema) Custom(fn func(string) error) StringBuilder {
	s.customFunc = fn
	return s
}

// State transition methods - these change the return type to enforce compile-time safety
func (s *stringSchema) Required() RequiredStringBuilder {
	s.required = true
	s.optional = false
	return &requiredStringSchema{s}
}

func (s *stringSchema) Optional() OptionalStringBuilder {
	s.optional = true
	s.required = false
	return &optionalStringSchema{s}
}

// Error message methods for StringBuilder
func (s *stringSchema) WithMessage(validationType, message string) StringBuilder {
	if s.customError == nil {
		s.customError = make(map[string]string)
	}
	s.customError[validationType] = message
	return s
}

func (s *stringSchema) WithMinLengthMessage(message string) StringBuilder {
	return s.WithMessage(errorKeys.MinLength, message)
}

func (s *stringSchema) WithMaxLengthMessage(message string) StringBuilder {
	return s.WithMessage(errorKeys.MaxLength, message)
}

func (s *stringSchema) WithPatternMessage(message string) StringBuilder {
	return s.WithMessage(errorKeys.Pattern, message)
}

func (s *stringSchema) WithEmailMessage(message string) StringBuilder {
	return s.WithMessage(errorKeys.Email, message)
}

func (s *stringSchema) WithURLMessage(message string) StringBuilder {
	return s.WithMessage(errorKeys.URL, message)
}

// RequiredStringBuilder implementation
// These methods return RequiredStringBuilder to maintain the required state

func (r *requiredStringSchema) Min(length int) RequiredStringBuilder {
	r.minLength = length
	return r
}

func (r *requiredStringSchema) Max(length int) RequiredStringBuilder {
	r.maxLength = length
	return r
}

func (r *requiredStringSchema) Pattern(pattern string) RequiredStringBuilder {
	// Handle potential regex compilation errors gracefully
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		r.pattern = regexp.MustCompile(`$^`) // This pattern never matches anything
		if r.customError == nil {
			r.customError = make(map[string]string)
		}
		r.customError[errorKeys.Pattern] = fmt.Sprintf("invalid regex pattern: %v", err)
	} else {
		r.pattern = compiled
	}
	return r
}

func (r *requiredStringSchema) Email() RequiredStringBuilder {
	r.emailFormat = true
	return r
}

func (r *requiredStringSchema) URL() RequiredStringBuilder {
	r.urlFormat = true
	return r
}

func (r *requiredStringSchema) Const(value string) RequiredStringBuilder {
	r.constValue = &value
	return r
}

func (r *requiredStringSchema) Custom(fn func(string) error) RequiredStringBuilder {
	r.customFunc = fn
	return r
}

// Error message methods for RequiredStringBuilder
func (r *requiredStringSchema) WithMessage(validationType, message string) RequiredStringBuilder {
	if r.customError == nil {
		r.customError = make(map[string]string)
	}
	r.customError[validationType] = message
	return r
}

func (r *requiredStringSchema) WithMinLengthMessage(message string) RequiredStringBuilder {
	return r.WithMessage(errorKeys.MinLength, message)
}

func (r *requiredStringSchema) WithMaxLengthMessage(message string) RequiredStringBuilder {
	return r.WithMessage(errorKeys.MaxLength, message)
}

func (r *requiredStringSchema) WithPatternMessage(message string) RequiredStringBuilder {
	return r.WithMessage(errorKeys.Pattern, message)
}

func (r *requiredStringSchema) WithEmailMessage(message string) RequiredStringBuilder {
	return r.WithMessage(errorKeys.Email, message)
}

func (r *requiredStringSchema) WithURLMessage(message string) RequiredStringBuilder {
	return r.WithMessage(errorKeys.URL, message)
}

func (r *requiredStringSchema) WithRequiredMessage(message string) RequiredStringBuilder {
	return r.WithMessage(errorKeys.Required, message)
}

// OptionalStringBuilder implementation
// These methods return OptionalStringBuilder to maintain the optional state

func (o *optionalStringSchema) Min(length int) OptionalStringBuilder {
	o.minLength = length
	return o
}

func (o *optionalStringSchema) Max(length int) OptionalStringBuilder {
	o.maxLength = length
	return o
}

func (o *optionalStringSchema) Pattern(pattern string) OptionalStringBuilder {
	// Handle potential regex compilation errors gracefully
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		o.pattern = regexp.MustCompile(`$^`) // This pattern never matches anything
		if o.customError == nil {
			o.customError = make(map[string]string)
		}
		o.customError[errorKeys.Pattern] = fmt.Sprintf("invalid regex pattern: %v", err)
	} else {
		o.pattern = compiled
	}
	return o
}

func (o *optionalStringSchema) Email() OptionalStringBuilder {
	o.emailFormat = true
	return o
}

func (o *optionalStringSchema) URL() OptionalStringBuilder {
	o.urlFormat = true
	return o
}

func (o *optionalStringSchema) Const(value string) OptionalStringBuilder {
	o.constValue = &value
	return o
}

func (o *optionalStringSchema) Custom(fn func(string) error) OptionalStringBuilder {
	o.customFunc = fn
	return o
}

// Default is only available on optional builders - this is the key DX improvement!
func (o *optionalStringSchema) Default(value string) OptionalStringBuilder {
	o.defaultValue = &value
	return o
}

// Error message methods for OptionalStringBuilder
func (o *optionalStringSchema) WithMessage(validationType, message string) OptionalStringBuilder {
	if o.customError == nil {
		o.customError = make(map[string]string)
	}
	o.customError[validationType] = message
	return o
}

func (o *optionalStringSchema) WithMinLengthMessage(message string) OptionalStringBuilder {
	return o.WithMessage(errorKeys.MinLength, message)
}

func (o *optionalStringSchema) WithMaxLengthMessage(message string) OptionalStringBuilder {
	return o.WithMessage(errorKeys.MaxLength, message)
}

func (o *optionalStringSchema) WithPatternMessage(message string) OptionalStringBuilder {
	return o.WithMessage(errorKeys.Pattern, message)
}

func (o *optionalStringSchema) WithEmailMessage(message string) OptionalStringBuilder {
	return o.WithMessage(errorKeys.Email, message)
}

func (o *optionalStringSchema) WithURLMessage(message string) OptionalStringBuilder {
	return o.WithMessage(errorKeys.URL, message)
}

// Validation methods - these are the final methods in the builder chain
func (r *requiredStringSchema) Validate(data interface{}) error {
	return r.validate(data)
}

func (o *optionalStringSchema) Validate(data interface{}) error {
	return o.validate(data)
}

// Core validation logic (shared between required and optional)
func (s *stringSchema) validate(data interface{}) error {
	// Handle nil values
	if data == nil {
		if s.required {
			return goop.NewValidationError("", nil, s.getErrorMessage(errorKeys.Required, "field is required"))
		}
		if s.defaultValue != nil {
			return s.validate(*s.defaultValue)
		}
		if s.optional {
			return nil
		}
		return goop.NewValidationError("", nil, s.getErrorMessage(errorKeys.Required, "field is required"))
	}

	// Type check
	str, ok := data.(string)
	if !ok {
		return goop.NewValidationError(fmt.Sprintf("%v", data), data,
			s.getErrorMessage(errorKeys.Type, "invalid type, expected string"))
	}

	// Handle empty strings
	if str == "" {
		if s.required {
			return goop.NewValidationError("", str,
				s.getErrorMessage(errorKeys.Required, "string is required"))
		}
		if s.defaultValue != nil {
			return s.validate(*s.defaultValue)
		}
		if s.optional {
			return nil
		}
	}

	// Length validations
	if s.minLength > 0 && len(str) < s.minLength {
		return goop.NewValidationError(str, str,
			s.getErrorMessage(errorKeys.MinLength,
				fmt.Sprintf("string is too short, minimum length is %d", s.minLength)))
	}

	if s.maxLength > 0 && len(str) > s.maxLength {
		return goop.NewValidationError(str, str,
			s.getErrorMessage(errorKeys.MaxLength,
				fmt.Sprintf("string is too long, maximum length is %d", s.maxLength)))
	}

	// Pattern validation
	if s.pattern != nil && !s.pattern.MatchString(str) {
		return goop.NewValidationError(str, str,
			s.getErrorMessage(errorKeys.Pattern, "string does not match required pattern"))
	}

	// Email validation
	if s.emailFormat && !isValidEmail(str) {
		return goop.NewValidationError(str, str,
			s.getErrorMessage(errorKeys.Email, "invalid email format"))
	}

	// URL validation
	if s.urlFormat && !isValidURL(str) {
		return goop.NewValidationError(str, str,
			s.getErrorMessage(errorKeys.URL, "invalid URL format"))
	}

	// Const validation
	if s.constValue != nil && str != *s.constValue {
		return goop.NewValidationError(str, str,
			s.getErrorMessage(errorKeys.Const, fmt.Sprintf("value must be exactly '%s'", *s.constValue)))
	}

	// Custom validation
	if s.customFunc != nil {
		if err := s.customFunc(str); err != nil {
			return err
		}
	}

	return nil
}

// Example methods for StringBuilder
func (s *stringSchema) Example(value interface{}) StringBuilder {
	s.example = value
	return s
}

func (s *stringSchema) Examples(examples map[string]ExampleObject) StringBuilder {
	s.examples = examples
	return s
}

func (s *stringSchema) ExampleFromFile(path string) StringBuilder {
	s.externalValue = path
	return s
}

// Example methods for RequiredStringBuilder
func (r *requiredStringSchema) Example(value interface{}) RequiredStringBuilder {
	r.example = value
	return r
}

func (r *requiredStringSchema) Examples(examples map[string]ExampleObject) RequiredStringBuilder {
	r.examples = examples
	return r
}

func (r *requiredStringSchema) ExampleFromFile(path string) RequiredStringBuilder {
	r.externalValue = path
	return r
}

// Example methods for OptionalStringBuilder
func (o *optionalStringSchema) Example(value interface{}) OptionalStringBuilder {
	o.example = value
	return o
}

func (o *optionalStringSchema) Examples(examples map[string]ExampleObject) OptionalStringBuilder {
	o.examples = examples
	return o
}

func (o *optionalStringSchema) ExampleFromFile(path string) OptionalStringBuilder {
	o.externalValue = path
	return o
}

// Helper methods (unexported)
func (s *stringSchema) getErrorMessage(validationType, defaultMessage string) string {
	if s.customError != nil {
		if msg, exists := s.customError[validationType]; exists {
			return msg
		}
	}
	return defaultMessage
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email) && len(email) <= 254
}

func isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && u.Scheme != "" && u.Host != ""
}
