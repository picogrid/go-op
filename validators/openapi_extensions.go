package validators

import (
	goop "github.com/picogrid/go-op"
)

// OpenAPI generation methods for stringSchema
// These methods enable build-time spec generation from existing validators

// ToOpenAPISchema generates OpenAPI 3.1 schema definition from string validation rules
func (s *stringSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	schema := &goop.OpenAPISchema{
		Type: "string",
	}

	// Add format constraints
	if s.emailFormat {
		schema.Format = "email"
	} else if s.urlFormat {
		schema.Format = "uri"
	}

	// Add length constraints
	if s.minLength > 0 {
		schema.MinLength = &s.minLength
	}
	if s.maxLength > 0 {
		schema.MaxLength = &s.maxLength
	}

	// Add pattern constraint
	if s.pattern != nil {
		schema.Pattern = s.pattern.String()
	}

	// Add const constraint
	if s.constValue != nil {
		schema.Const = *s.constValue
	}

	// Add default value for optional schemas
	if s.defaultValue != nil {
		schema.Default = *s.defaultValue
	}

	// Add example information
	if s.example != nil {
		schema.Example = s.example
	}

	return schema
}

// GetValidationInfo returns metadata about the validation configuration
func (s *stringSchema) GetValidationInfo() *goop.ValidationInfo {
	info := &goop.ValidationInfo{
		Required:    s.required,
		Optional:    s.optional,
		HasDefault:  s.defaultValue != nil,
		Constraints: make(map[string]interface{}),
	}

	if s.defaultValue != nil {
		info.DefaultValue = *s.defaultValue
	}

	// Store constraint information for build-time analysis
	if s.minLength > 0 {
		info.Constraints["minLength"] = s.minLength
	}
	if s.maxLength > 0 {
		info.Constraints["maxLength"] = s.maxLength
	}
	if s.pattern != nil {
		info.Constraints["pattern"] = s.pattern.String()
	}
	if s.emailFormat {
		info.Constraints["format"] = "email"
	}
	if s.urlFormat {
		info.Constraints["format"] = "uri"
	}

	return info
}

// OpenAPI generation methods for RequiredStringBuilder
func (r *requiredStringSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	return r.stringSchema.ToOpenAPISchema()
}

func (r *requiredStringSchema) GetValidationInfo() *goop.ValidationInfo {
	return r.stringSchema.GetValidationInfo()
}

// OpenAPI generation methods for OptionalStringBuilder
func (o *optionalStringSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	return o.stringSchema.ToOpenAPISchema()
}

func (o *optionalStringSchema) GetValidationInfo() *goop.ValidationInfo {
	return o.stringSchema.GetValidationInfo()
}

// OpenAPI generation methods for numberSchema

// ToOpenAPISchema generates OpenAPI 3.1 schema definition from number validation rules
func (n *numberSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	schema := &goop.OpenAPISchema{
		Type: "number",
	}

	// Set integer type if specified
	if n.integerOnly {
		schema.Type = "integer"
	}

	// Add range constraints
	if n.minValue != nil {
		schema.Minimum = n.minValue
	}
	if n.maxValue != nil {
		schema.Maximum = n.maxValue
	}

	// Add exclusive range constraints
	if n.exclusiveMinValue != nil {
		schema.ExclusiveMinimum = n.exclusiveMinValue
	}
	if n.exclusiveMaxValue != nil {
		schema.ExclusiveMaximum = n.exclusiveMaxValue
	}

	// Add multipleOf constraint
	if n.multipleOfValue != nil {
		schema.MultipleOf = n.multipleOfValue
	}

	// Handle positive/negative constraints
	if n.positiveOnly && (n.minValue == nil || *n.minValue <= 0) {
		zero := 0.0
		schema.Minimum = &zero
	}
	if n.negativeOnly && (n.maxValue == nil || *n.maxValue >= 0) {
		zero := 0.0
		schema.Maximum = &zero
	}

	// Add default value for optional schemas
	if n.defaultValue != nil {
		schema.Default = *n.defaultValue
	}

	// Add example information
	if n.example != nil {
		schema.Example = n.example
	}

	return schema
}

// GetValidationInfo returns metadata about the number validation configuration
func (n *numberSchema) GetValidationInfo() *goop.ValidationInfo {
	info := &goop.ValidationInfo{
		Required:    n.required,
		Optional:    n.optional,
		HasDefault:  n.defaultValue != nil,
		Constraints: make(map[string]interface{}),
	}

	if n.defaultValue != nil {
		info.DefaultValue = *n.defaultValue
	}

	// Store constraint information for build-time analysis
	if n.minValue != nil {
		info.Constraints["minimum"] = *n.minValue
	}
	if n.maxValue != nil {
		info.Constraints["maximum"] = *n.maxValue
	}
	if n.integerOnly {
		info.Constraints["type"] = "integer"
	}
	if n.positiveOnly {
		info.Constraints["positive"] = true
	}
	if n.negativeOnly {
		info.Constraints["negative"] = true
	}

	return info
}

// OpenAPI generation methods for RequiredNumberBuilder
func (r *requiredNumberSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	return r.numberSchema.ToOpenAPISchema()
}

func (r *requiredNumberSchema) GetValidationInfo() *goop.ValidationInfo {
	return r.numberSchema.GetValidationInfo()
}

// OpenAPI generation methods for OptionalNumberBuilder
func (o *optionalNumberSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	return o.numberSchema.ToOpenAPISchema()
}

func (o *optionalNumberSchema) GetValidationInfo() *goop.ValidationInfo {
	return o.numberSchema.GetValidationInfo()
}

// OpenAPI generation methods for arraySchema

// ToOpenAPISchema generates OpenAPI 3.1 schema definition from array validation rules
func (a *arraySchema) ToOpenAPISchema() *goop.OpenAPISchema {
	schema := &goop.OpenAPISchema{
		Type: "array",
	}

	// Add array size constraints
	if a.minItems > 0 {
		schema.MinItems = &a.minItems
	}
	if a.maxItems > 0 {
		schema.MaxItems = &a.maxItems
	}

	// Add uniqueItems constraint
	if a.uniqueItems {
		schema.UniqueItems = &a.uniqueItems
	}

	// Generate schema for array items
	if a.elementSchema != nil {
		if enhancedElement, ok := a.elementSchema.(goop.EnhancedSchema); ok {
			schema.Items = enhancedElement.ToOpenAPISchema()
		} else {
			// Fallback for non-enhanced schemas - basic type detection
			schema.Items = &goop.OpenAPISchema{Type: "string"} // Default fallback
		}
	}

	// Add default value for optional schemas
	if a.defaultValue != nil {
		schema.Default = a.defaultValue
	}

	// Add example information
	if a.example != nil {
		schema.Example = a.example
	}

	return schema
}

// GetValidationInfo returns metadata about the array validation configuration
func (a *arraySchema) GetValidationInfo() *goop.ValidationInfo {
	info := &goop.ValidationInfo{
		Required:    a.required,
		Optional:    a.optional,
		HasDefault:  a.defaultValue != nil,
		Constraints: make(map[string]interface{}),
	}

	if a.defaultValue != nil {
		info.DefaultValue = a.defaultValue
	}

	// Store constraint information for build-time analysis
	if a.minItems > 0 {
		info.Constraints["minItems"] = a.minItems
	}
	if a.maxItems > 0 {
		info.Constraints["maxItems"] = a.maxItems
	}
	if a.contains != nil {
		info.Constraints["contains"] = true
	}

	return info
}

// OpenAPI generation methods for RequiredArrayBuilder
func (r *requiredArraySchema) ToOpenAPISchema() *goop.OpenAPISchema {
	return r.arraySchema.ToOpenAPISchema()
}

func (r *requiredArraySchema) GetValidationInfo() *goop.ValidationInfo {
	return r.arraySchema.GetValidationInfo()
}

// OpenAPI generation methods for OptionalArrayBuilder
func (o *optionalArraySchema) ToOpenAPISchema() *goop.OpenAPISchema {
	return o.arraySchema.ToOpenAPISchema()
}

func (o *optionalArraySchema) GetValidationInfo() *goop.ValidationInfo {
	return o.arraySchema.GetValidationInfo()
}

// OpenAPI generation methods for objectSchema

// ToOpenAPISchema generates OpenAPI 3.1 schema definition from object validation rules
func (obj *objectSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	schema := &goop.OpenAPISchema{
		Type:       "object",
		Properties: make(map[string]*goop.OpenAPISchema),
		Required:   []string{},
	}

	// Process each property in the object schema
	for fieldName, fieldSchema := range obj.schema {
		if enhancedField, ok := fieldSchema.(goop.EnhancedSchema); ok {
			propertySchema := enhancedField.ToOpenAPISchema()
			schema.Properties[fieldName] = propertySchema

			// Check if this field is required
			validationInfo := enhancedField.GetValidationInfo()
			if validationInfo.Required {
				schema.Required = append(schema.Required, fieldName)
			}
		} else {
			// Fallback for non-enhanced schemas
			schema.Properties[fieldName] = &goop.OpenAPISchema{Type: "string"} // Default fallback
		}
	}

	// Add property count constraints
	if obj.minProperties > 0 {
		schema.MinProperties = &obj.minProperties
	}
	if obj.maxProperties > 0 {
		schema.MaxProperties = &obj.maxProperties
	}

	// Add example information
	if obj.example != nil {
		schema.Example = obj.example
	}

	return schema
}

// GetValidationInfo returns metadata about the object validation configuration
func (obj *objectSchema) GetValidationInfo() *goop.ValidationInfo {
	info := &goop.ValidationInfo{
		Required:    obj.required,
		Optional:    obj.optional,
		HasDefault:  false, // Objects typically don't have default values in this implementation
		Constraints: make(map[string]interface{}),
	}

	// Store schema structure information
	info.Constraints["properties"] = len(obj.schema)

	return info
}

// OpenAPI generation methods for RequiredObjectBuilder
func (r *requiredObjectSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	return r.objectSchema.ToOpenAPISchema()
}

func (r *requiredObjectSchema) GetValidationInfo() *goop.ValidationInfo {
	return r.objectSchema.GetValidationInfo()
}

// OpenAPI generation methods for OptionalObjectBuilder
func (o *optionalObjectSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	return o.objectSchema.ToOpenAPISchema()
}

func (o *optionalObjectSchema) GetValidationInfo() *goop.ValidationInfo {
	return o.objectSchema.GetValidationInfo()
}

// OpenAPI generation methods for boolSchema

// ToOpenAPISchema generates OpenAPI 3.1 schema definition from boolean validation rules
func (b *boolSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	schema := &goop.OpenAPISchema{
		Type: "boolean",
	}

	// Add default value for optional schemas
	if b.defaultValue != nil {
		schema.Default = *b.defaultValue
	}

	// Add example information
	if b.example != nil {
		schema.Example = b.example
	}

	return schema
}

// GetValidationInfo returns metadata about the boolean validation configuration
func (b *boolSchema) GetValidationInfo() *goop.ValidationInfo {
	info := &goop.ValidationInfo{
		Required:    b.required,
		Optional:    b.optional,
		HasDefault:  b.defaultValue != nil,
		Constraints: make(map[string]interface{}),
	}

	if b.defaultValue != nil {
		info.DefaultValue = *b.defaultValue
	}

	return info
}

// OpenAPI generation methods for RequiredBoolBuilder
func (r *requiredBoolSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	return r.boolSchema.ToOpenAPISchema()
}

func (r *requiredBoolSchema) GetValidationInfo() *goop.ValidationInfo {
	return r.boolSchema.GetValidationInfo()
}

// OpenAPI generation methods for OptionalBoolBuilder
func (o *optionalBoolSchema) ToOpenAPISchema() *goop.OpenAPISchema {
	return o.boolSchema.ToOpenAPISchema()
}

func (o *optionalBoolSchema) GetValidationInfo() *goop.ValidationInfo {
	return o.boolSchema.GetValidationInfo()
}

// Enhanced interfaces that extend the existing builders with OpenAPI generation
// These allow the builders to be used as EnhancedSchema

type EnhancedRequiredObjectBuilder interface {
	RequiredObjectBuilder
	goop.EnhancedSchema
}

type EnhancedOptionalObjectBuilder interface {
	OptionalObjectBuilder
	goop.EnhancedSchema
}

type EnhancedRequiredStringBuilder interface {
	RequiredStringBuilder
	goop.EnhancedSchema
}

type EnhancedOptionalStringBuilder interface {
	OptionalStringBuilder
	goop.EnhancedSchema
}

type EnhancedRequiredNumberBuilder interface {
	RequiredNumberBuilder
	goop.EnhancedSchema
}

type EnhancedOptionalNumberBuilder interface {
	OptionalNumberBuilder
	goop.EnhancedSchema
}

type EnhancedRequiredArrayBuilder interface {
	RequiredArrayBuilder
	goop.EnhancedSchema
}

type EnhancedOptionalArrayBuilder interface {
	OptionalArrayBuilder
	goop.EnhancedSchema
}

type EnhancedRequiredBoolBuilder interface {
	RequiredBoolBuilder
	goop.EnhancedSchema
}

type EnhancedOptionalBoolBuilder interface {
	OptionalBoolBuilder
	goop.EnhancedSchema
}

// Enhanced interface compliance check at compile time
var _ EnhancedRequiredStringBuilder = (*requiredStringSchema)(nil)
var _ EnhancedOptionalStringBuilder = (*optionalStringSchema)(nil)
var _ EnhancedRequiredNumberBuilder = (*requiredNumberSchema)(nil)
var _ EnhancedOptionalNumberBuilder = (*optionalNumberSchema)(nil)
var _ EnhancedRequiredArrayBuilder = (*requiredArraySchema)(nil)
var _ EnhancedOptionalArrayBuilder = (*optionalArraySchema)(nil)
var _ EnhancedRequiredObjectBuilder = (*requiredObjectSchema)(nil)
var _ EnhancedOptionalObjectBuilder = (*optionalObjectSchema)(nil)
var _ EnhancedRequiredBoolBuilder = (*requiredBoolSchema)(nil)
var _ EnhancedOptionalBoolBuilder = (*optionalBoolSchema)(nil)
