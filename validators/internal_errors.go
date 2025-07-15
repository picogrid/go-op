package validators

// Internal error keys (unexported to keep package interface clean)
// These are used internally by all validators to maintain consistency
var errorKeys = struct {
	// Common validation errors
	Required string
	Type     string
	Custom   string

	// String validation errors
	MinLength string
	MaxLength string
	Pattern   string
	Email     string
	URL       string
	Const     string

	// Number validation errors
	Min          string
	Max          string
	ExclusiveMin string
	ExclusiveMax string
	MultipleOf   string
	Integer      string
	Positive     string
	Negative     string

	// Array validation errors
	MinItems    string
	MaxItems    string
	Contains    string
	UniqueItems string

	// Object validation errors
	UnknownKey    string
	MissingKey    string
	InvalidShape  string
	MinProperties string
	MaxProperties string

	// Boolean validation errors
	InvalidBoolean string
}{
	// Common
	Required: "required",
	Type:     "type",
	Custom:   "custom",

	// String
	MinLength: "minLength",
	MaxLength: "maxLength",
	Pattern:   "pattern",
	Email:     "email",
	URL:       "url",
	Const:     "const",

	// Number
	Min:          "min",
	Max:          "max",
	ExclusiveMin: "exclusiveMin",
	ExclusiveMax: "exclusiveMax",
	MultipleOf:   "multipleOf",
	Integer:      "integer",
	Positive:     "positive",
	Negative:     "negative",

	// Array
	MinItems:    "minItems",
	MaxItems:    "maxItems",
	Contains:    "contains",
	UniqueItems: "uniqueItems",

	// Object
	UnknownKey:    "unknownKey",
	MissingKey:    "missingKey",
	InvalidShape:  "invalidShape",
	MinProperties: "minProperties",
	MaxProperties: "maxProperties",

	// Boolean
	InvalidBoolean: "invalidBoolean",
}

// ErrorKeys provides autocompletion for error keys.
// This is a clean way to access error keys with IDE autocompletion.
type ErrorKeys struct{}

// Common error keys available across all validators
func (ErrorKeys) Required() string { return errorKeys.Required }
func (ErrorKeys) Type() string     { return errorKeys.Type }
func (ErrorKeys) Custom() string   { return errorKeys.Custom }

// String-specific error keys
func (ErrorKeys) MinLength() string { return errorKeys.MinLength }
func (ErrorKeys) MaxLength() string { return errorKeys.MaxLength }
func (ErrorKeys) Pattern() string   { return errorKeys.Pattern }
func (ErrorKeys) Email() string     { return errorKeys.Email }
func (ErrorKeys) URL() string       { return errorKeys.URL }
func (ErrorKeys) Const() string     { return errorKeys.Const }

// Number-specific error keys
func (ErrorKeys) Min() string          { return errorKeys.Min }
func (ErrorKeys) Max() string          { return errorKeys.Max }
func (ErrorKeys) ExclusiveMin() string { return errorKeys.ExclusiveMin }
func (ErrorKeys) ExclusiveMax() string { return errorKeys.ExclusiveMax }
func (ErrorKeys) MultipleOf() string   { return errorKeys.MultipleOf }
func (ErrorKeys) Integer() string      { return errorKeys.Integer }
func (ErrorKeys) Positive() string     { return errorKeys.Positive }
func (ErrorKeys) Negative() string     { return errorKeys.Negative }

// Array-specific error keys
func (ErrorKeys) MinItems() string    { return errorKeys.MinItems }
func (ErrorKeys) MaxItems() string    { return errorKeys.MaxItems }
func (ErrorKeys) Contains() string    { return errorKeys.Contains }
func (ErrorKeys) UniqueItems() string { return errorKeys.UniqueItems }

// Object-specific error keys
func (ErrorKeys) UnknownKey() string    { return errorKeys.UnknownKey }
func (ErrorKeys) MissingKey() string    { return errorKeys.MissingKey }
func (ErrorKeys) InvalidShape() string  { return errorKeys.InvalidShape }
func (ErrorKeys) MinProperties() string { return errorKeys.MinProperties }
func (ErrorKeys) MaxProperties() string { return errorKeys.MaxProperties }

// Boolean-specific error keys
func (ErrorKeys) InvalidBoolean() string { return errorKeys.InvalidBoolean }

// Errors provides a global instance for accessing error keys with autocompletion.
// Usage: validators.Errors.MinLength(), validators.Errors.Required(), etc.
var Errors ErrorKeys

// Alternative approach: Individual constants for users who prefer this style.
// These provide shorter syntax: validators.ErrMinLength vs validators.Errors.MinLength()
const (
	// Common error constants
	ErrRequired = "required"
	ErrType     = "type"
	ErrCustom   = "custom"

	// String error constants
	ErrMinLength = "minLength"
	ErrMaxLength = "maxLength"
	ErrPattern   = "pattern"
	ErrEmail     = "email"
	ErrURL       = "url"
	ErrConst     = "const"

	// Number error constants
	ErrMin          = "min"
	ErrMax          = "max"
	ErrExclusiveMin = "exclusiveMin"
	ErrExclusiveMax = "exclusiveMax"
	ErrMultipleOf   = "multipleOf"
	ErrInteger      = "integer"
	ErrPositive     = "positive"
	ErrNegative     = "negative"

	// Array error constants
	ErrMinItems    = "minItems"
	ErrMaxItems    = "maxItems"
	ErrContains    = "contains"
	ErrUniqueItems = "uniqueItems"

	// Object error constants
	ErrUnknownKey    = "unknownKey"
	ErrMissingKey    = "missingKey"
	ErrInvalidShape  = "invalidShape"
	ErrMinProperties = "minProperties"
	ErrMaxProperties = "maxProperties"

	// Boolean error constants
	ErrInvalidBoolean = "invalidBoolean"
)
