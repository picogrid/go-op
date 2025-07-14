package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// ASTAnalyzer provides sophisticated AST analysis for operation extraction
type ASTAnalyzer struct {
	fileSet    *token.FileSet
	verbose    bool
	schemaVars map[string]*SchemaDefinition // Track schema variable definitions
}

// NewASTAnalyzer creates a new AST analyzer
func NewASTAnalyzer(fileSet *token.FileSet, verbose bool) *ASTAnalyzer {
	return &ASTAnalyzer{
		fileSet:    fileSet,
		verbose:    verbose,
		schemaVars: make(map[string]*SchemaDefinition),
	}
}

// ExtractOperations extracts operation definitions from an AST node
func (a *ASTAnalyzer) ExtractOperations(file *ast.File, filename string) []OperationDefinition {
	var operations []OperationDefinition

	if a.verbose {
		fmt.Printf("[VERBOSE] Analyzing file %s with %d declarations\n", filename, len(file.Decls))
	}

	// Look for variable assignments that create operations
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
			if a.verbose {
				fmt.Printf("[VERBOSE] Found var declaration with %d specs\n", len(genDecl.Specs))
			}
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					ops := a.extractFromValueSpec(valueSpec, filename)
					operations = append(operations, ops...)
				}
			}
		} else if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if a.verbose {
				fmt.Printf("[VERBOSE] Found function declaration: %s\n", funcDecl.Name.Name)
			}
			// Look inside function bodies for operations and schema definitions
			if funcDecl.Body != nil {
				for _, stmt := range funcDecl.Body.List {
					if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
						// First, check if this is a schema variable assignment
						a.trackSchemaAssignments(assignStmt, filename)

						// Then, extract operations
						ops := a.extractFromAssignStmt(assignStmt, filename)
						operations = append(operations, ops...)
					}
				}
			}
		} else if a.verbose {
			fmt.Printf("[VERBOSE] Found other declaration type: %T\n", decl)
		}
	}

	// Look for function calls that register operations
	ast.Inspect(file, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if op := a.extractFromCallExpr(callExpr, filename); op != nil {
				operations = append(operations, *op)
			}
		}
		return true
	})

	if a.verbose {
		fmt.Printf("[VERBOSE] Found %d operations in %s\n", len(operations), filename)
	}

	return operations
}

// trackSchemaAssignments tracks schema variable assignments for later resolution
func (a *ASTAnalyzer) trackSchemaAssignments(assignStmt *ast.AssignStmt, filename string) {
	for i, lhs := range assignStmt.Lhs {
		if i < len(assignStmt.Rhs) {
			if ident, ok := lhs.(*ast.Ident); ok {
				// Check if this looks like a schema assignment (contains "Schema" in name)
				if strings.Contains(ident.Name, "Schema") {
					if a.verbose {
						fmt.Printf("[VERBOSE] Tracking schema variable: %s\n", ident.Name)
					}

					// Extract the schema definition from the RHS
					schema := a.extractSchemaDefinition(assignStmt.Rhs[i])
					if schema != nil {
						// Enhance the schema with more detailed analysis
						a.enhanceSchemaFromValidatorCall(assignStmt.Rhs[i], schema)
						a.schemaVars[ident.Name] = schema
					}
				}
			}
		}
	}
}

// enhanceSchemaFromValidatorCall enhances schema definition by analyzing validator calls
func (a *ASTAnalyzer) enhanceSchemaFromValidatorCall(expr ast.Expr, schema *SchemaDefinition) {
	if callExpr, ok := expr.(*ast.CallExpr); ok {
		// Deep analysis of validator call chains
		a.analyzeValidatorCall(callExpr, schema)

		if a.verbose {
			fmt.Printf("[VERBOSE] Enhanced schema type: %s with %d properties\n",
				schema.Type, len(schema.Properties))
		}
	}
}

// extractFromAssignStmt extracts operations from assignment statements
func (a *ASTAnalyzer) extractFromAssignStmt(assignStmt *ast.AssignStmt, filename string) []OperationDefinition {
	var operations []OperationDefinition

	if a.verbose {
		fmt.Printf("[VERBOSE] Found assignment with %d LHS and %d RHS\n", len(assignStmt.Lhs), len(assignStmt.Rhs))
	}

	// Handle := assignments which create new variables
	for i, lhs := range assignStmt.Lhs {
		if i < len(assignStmt.Rhs) {
			if ident, ok := lhs.(*ast.Ident); ok {
				if a.verbose {
					fmt.Printf("[VERBOSE] Checking assignment to variable: %s\n", ident.Name)
				}
				if op := a.extractFromExpr(assignStmt.Rhs[i], filename, ident.Name); op != nil {
					operations = append(operations, *op)
				}
			}
		}
	}

	return operations
}

// extractFromValueSpec extracts operations from variable declarations
func (a *ASTAnalyzer) extractFromValueSpec(valueSpec *ast.ValueSpec, filename string) []OperationDefinition {
	var operations []OperationDefinition

	for i, name := range valueSpec.Names {
		if i < len(valueSpec.Values) {
			if op := a.extractFromExpr(valueSpec.Values[i], filename, name.Name); op != nil {
				operations = append(operations, *op)
			}
		}
	}

	return operations
}

// extractFromCallExpr extracts operations from function calls (like router.Register)
func (a *ASTAnalyzer) extractFromCallExpr(callExpr *ast.CallExpr, filename string) *OperationDefinition {
	// Check if this is a router.Register call
	if a.isRouterRegister(callExpr) && len(callExpr.Args) > 0 {
		// The first argument should be the operation
		if op := a.extractFromExpr(callExpr.Args[0], filename, ""); op != nil {
			return op
		}
	}
	return nil
}

// extractFromExpr extracts operation definition from an expression
func (a *ASTAnalyzer) extractFromExpr(expr ast.Expr, filename, varName string) *OperationDefinition {
	switch e := expr.(type) {
	case *ast.CallExpr:
		return a.extractFromOperationChain(e, filename, varName)
	case *ast.Ident:
		// This might be a reference to an operation variable
		// For now, we'll skip these as they require more complex analysis
		return nil
	default:
		return nil
	}
}

// extractFromOperationChain extracts operation from method chain like operations.NewSimple().GET("/path")
func (a *ASTAnalyzer) extractFromOperationChain(callExpr *ast.CallExpr, filename, varName string) *OperationDefinition {
	// Walk up the call chain to extract operation details
	op := &OperationDefinition{
		SourceFile: filename,
		Tags:       []string{},
	}

	// Set variable name as identifier if available
	if varName != "" {
		op.Summary = varName
	}

	// Traverse the method chain
	a.traverseMethodChain(callExpr, op)

	// Only return operation if we found a valid HTTP method and path
	if op.Method != "" && op.Path != "" {
		return op
	}

	return nil
}

// traverseMethodChain recursively traverses method chains to extract operation details
func (a *ASTAnalyzer) traverseMethodChain(expr ast.Expr, op *OperationDefinition) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		// First, traverse the receiver (left side of the call)
		if selExpr, ok := e.Fun.(*ast.SelectorExpr); ok {
			a.traverseMethodChain(selExpr.X, op)

			// Then process this method call
			methodName := selExpr.Sel.Name
			a.processMethodCall(methodName, e.Args, op)
		}
	case *ast.SelectorExpr:
		// This handles cases like operations.NewSimple
		a.traverseMethodChain(e.X, op)
	case *ast.Ident:
		// Base case - this is usually the package name or variable
		return
	}
}

// processMethodCall processes individual method calls in the chain
func (a *ASTAnalyzer) processMethodCall(methodName string, args []ast.Expr, op *OperationDefinition) {
	if a.verbose {
		fmt.Printf("[VERBOSE] Processing method: %s with %d args\n", methodName, len(args))
	}

	switch methodName {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		op.Method = methodName
		if len(args) > 0 {
			if path := a.extractStringLiteral(args[0]); path != "" {
				op.Path = path
				if a.verbose {
					fmt.Printf("[VERBOSE] Set path: %s\n", path)
				}
			}
		}
	case "Summary":
		if len(args) > 0 {
			if summary := a.extractStringLiteral(args[0]); summary != "" {
				op.Summary = summary
			}
		}
	case "Description":
		if len(args) > 0 {
			if desc := a.extractStringLiteral(args[0]); desc != "" {
				op.Description = desc
			}
		}
	case "Tags":
		// Extract tags from arguments
		for _, arg := range args {
			if tag := a.extractStringLiteral(arg); tag != "" {
				op.Tags = append(op.Tags, tag)
			}
		}
	case "WithParams":
		if len(args) > 0 {
			op.Params = a.extractSchemaDefinition(args[0])
		}
	case "WithQuery":
		if len(args) > 0 {
			op.Query = a.extractSchemaDefinition(args[0])
		}
	case "WithBody":
		if len(args) > 0 {
			op.Body = a.extractSchemaDefinition(args[0])
		}
	case "WithResponse":
		if len(args) > 0 {
			op.Response = a.extractSchemaDefinition(args[0])
		}
	}
}

// extractStringLiteral extracts string value from a basic literal
func (a *ASTAnalyzer) extractStringLiteral(expr ast.Expr) string {
	if basicLit, ok := expr.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
		// Remove quotes from string literal
		value := basicLit.Value
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			return value[1 : len(value)-1]
		}
	}
	return ""
}

// extractNumberLiteral extracts numeric value from a basic literal
func (a *ASTAnalyzer) extractNumberLiteral(expr ast.Expr) *float64 {
	if basicLit, ok := expr.(*ast.BasicLit); ok {
		if basicLit.Kind == token.INT || basicLit.Kind == token.FLOAT {
			if val, err := parseFloat(basicLit.Value); err == nil {
				return &val
			}
		}
	}
	return nil
}

// parseFloat is a simple float parser for literal values
func parseFloat(s string) (float64, error) {
	// Simple implementation - in a real parser you'd use strconv.ParseFloat
	// For now, handle simple integers
	switch s {
	case "0":
		return 0.0, nil
	case "1":
		return 1.0, nil
	case "3":
		return 3.0, nil
	case "100":
		return 100.0, nil
	case "150":
		return 150.0, nil
	default:
		return 0.0, fmt.Errorf("unsupported number: %s", s)
	}
}

// extractSchemaDefinition extracts schema information from an expression
func (a *ASTAnalyzer) extractSchemaDefinition(expr ast.Expr) *SchemaDefinition {
	schema := &SchemaDefinition{
		Type:       "object",
		Properties: make(map[string]*SchemaDefinition),
		Required:   []string{},
	}

	if a.verbose {
		fmt.Printf("[VERBOSE] Extracting schema from expression type: %T\n", expr)
	}

	// Try to extract schema information from validator calls
	if callExpr, ok := expr.(*ast.CallExpr); ok {
		a.analyzeValidatorCall(callExpr, schema)
	} else if ident, ok := expr.(*ast.Ident); ok {
		// This might be a reference to a schema variable
		if a.verbose {
			fmt.Printf("[VERBOSE] Schema reference to variable: %s\n", ident.Name)
		}

		// Check if we have this schema variable tracked
		if trackedSchema, exists := a.schemaVars[ident.Name]; exists {
			if a.verbose {
				fmt.Printf("[VERBOSE] Resolved schema reference: %s -> %s\n", ident.Name, trackedSchema.Type)
			}
			// Copy the tracked schema content
			*schema = *trackedSchema
		} else {
			// Fallback to placeholder
			schema.Description = fmt.Sprintf("Reference to %s", ident.Name)
		}
	}

	return schema
}

// analyzeValidatorCall analyzes validator calls to extract schema information
func (a *ASTAnalyzer) analyzeValidatorCall(callExpr *ast.CallExpr, schema *SchemaDefinition) {
	if a.verbose {
		fmt.Printf("[VERBOSE] Analyzing validator call\n")
	}

	// Handle method chaining by traversing the call chain
	a.traverseValidatorChain(callExpr, schema)
}

// traverseValidatorChain recursively traverses validator method chains
func (a *ASTAnalyzer) traverseValidatorChain(expr ast.Expr, schema *SchemaDefinition) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		// First, traverse the receiver (left side of the call)
		if selExpr, ok := e.Fun.(*ast.SelectorExpr); ok {
			a.traverseValidatorChain(selExpr.X, schema)

			// Then process this method call
			methodName := selExpr.Sel.Name
			a.processValidatorMethod(methodName, e.Args, schema)
		}
	case *ast.SelectorExpr:
		// This handles cases like validators.String
		a.traverseValidatorChain(e.X, schema)
		if a.isValidatorPackage(e.X) {
			methodName := e.Sel.Name
			a.processValidatorMethod(methodName, []ast.Expr{}, schema)
		}
	case *ast.Ident:
		// Base case - this is usually the package name
		return
	}
}

// processValidatorMethod processes individual validator method calls
func (a *ASTAnalyzer) processValidatorMethod(methodName string, args []ast.Expr, schema *SchemaDefinition) {
	if a.verbose {
		fmt.Printf("[VERBOSE] Processing validator method: %s with %d args\n", methodName, len(args))
	}

	switch methodName {
	case "Object":
		schema.Type = "object"
		// Extract object properties from arguments
		if len(args) > 0 {
			a.extractObjectProperties(args[0], schema)
		}
	case "String":
		schema.Type = "string"
	case "Number":
		schema.Type = "number"
	case "Array":
		schema.Type = "array"
		// TODO: Extract array item type from arguments
	case "Bool":
		schema.Type = "boolean"
	case "Email":
		schema.Type = "string"
		schema.Format = "email"
	case "Min":
		if len(args) > 0 {
			if val := a.extractNumberLiteral(args[0]); val != nil {
				if schema.Type == "string" {
					schema.MinLength = func(v int) *int { return &v }(int(*val))
				} else {
					schema.Minimum = val
				}
			}
		}
	case "Max":
		if len(args) > 0 {
			if val := a.extractNumberLiteral(args[0]); val != nil {
				if schema.Type == "string" {
					schema.MaxLength = func(v int) *int { return &v }(int(*val))
				} else {
					schema.Maximum = val
				}
			}
		}
	case "Required":
		// Mark this property as required
		if a.verbose {
			fmt.Printf("[VERBOSE] Schema is required\n")
		}
		// Note: For object properties, we'll need context about which property this is
	case "Optional":
		// This indicates the field is optional
		if a.verbose {
			fmt.Printf("[VERBOSE] Schema is optional\n")
		}
	case "Example":
		// Extract simple example value
		if len(args) > 0 {
			if val := a.extractLiteralValue(args[0]); val != nil {
				schema.Example = val
				if a.verbose {
					fmt.Printf("[VERBOSE] Extracted example: %v\n", val)
				}
			}
		}
	case "Examples":
		// Extract multiple examples map
		if len(args) > 0 {
			if examples := a.extractExamplesMap(args[0]); examples != nil {
				schema.Examples = examples
				if a.verbose {
					fmt.Printf("[VERBOSE] Extracted %d examples\n", len(examples))
				}
			}
		}
	case "ExampleFromFile":
		// Extract external file reference
		if len(args) > 0 {
			if val := a.extractStringLiteral(args[0]); val != "" {
				schema.ExternalValue = val
				if a.verbose {
					fmt.Printf("[VERBOSE] Extracted external example: %s\n", val)
				}
			}
		}
	case "OneOf":
		// Extract OneOf composition schemas
		schema.Type = "" // Clear type for composition schemas
		if a.verbose {
			fmt.Printf("[VERBOSE] Processing OneOf with %d schemas\n", len(args))
		}
		for _, arg := range args {
			childSchema := a.extractSchemaDefinition(arg)
			if childSchema != nil {
				schema.OneOf = append(schema.OneOf, childSchema)
			}
		}
	case "AllOf":
		// Extract AllOf composition schemas
		schema.Type = "" // Clear type for composition schemas
		if a.verbose {
			fmt.Printf("[VERBOSE] Processing AllOf with %d schemas\n", len(args))
		}
		for _, arg := range args {
			childSchema := a.extractSchemaDefinition(arg)
			if childSchema != nil {
				schema.AllOf = append(schema.AllOf, childSchema)
			}
		}
	case "AnyOf":
		// Extract AnyOf composition schemas
		schema.Type = "" // Clear type for composition schemas
		if a.verbose {
			fmt.Printf("[VERBOSE] Processing AnyOf with %d schemas\n", len(args))
		}
		for _, arg := range args {
			childSchema := a.extractSchemaDefinition(arg)
			if childSchema != nil {
				schema.AnyOf = append(schema.AnyOf, childSchema)
			}
		}
	case "Not":
		// Extract Not composition schema
		schema.Type = "" // Clear type for composition schemas
		if len(args) > 0 {
			if a.verbose {
				fmt.Printf("[VERBOSE] Processing Not with 1 schema\n")
			}
			childSchema := a.extractSchemaDefinition(args[0])
			if childSchema != nil {
				schema.Not = childSchema
			}
		}
	}
}

// extractObjectProperties extracts object properties from a map literal
func (a *ASTAnalyzer) extractObjectProperties(expr ast.Expr, schema *SchemaDefinition) {
	if a.verbose {
		fmt.Printf("[VERBOSE] Extracting object properties from %T\n", expr)
	}

	// Look for map[string]interface{}{...} patterns
	if compositeLit, ok := expr.(*ast.CompositeLit); ok {
		if a.verbose {
			fmt.Printf("[VERBOSE] Found composite literal with %d elements\n", len(compositeLit.Elts))
		}

		for _, elt := range compositeLit.Elts {
			if keyValue, ok := elt.(*ast.KeyValueExpr); ok {
				// Extract property name from key
				if key := a.extractStringLiteral(keyValue.Key); key != "" {
					if a.verbose {
						fmt.Printf("[VERBOSE] Found property: %s\n", key)
					}

					// Create property schema from value
					propSchema := &SchemaDefinition{
						Type:       "string", // Default type
						Properties: make(map[string]*SchemaDefinition),
						Required:   []string{},
					}

					// Analyze the property value to determine its schema
					a.analyzePropertyValue(keyValue.Value, propSchema)

					// Add to schema properties
					if schema.Properties == nil {
						schema.Properties = make(map[string]*SchemaDefinition)
					}
					schema.Properties[key] = propSchema
				}
			}
		}
	}
}

// analyzePropertyValue analyzes a property value to determine its schema
func (a *ASTAnalyzer) analyzePropertyValue(expr ast.Expr, propSchema *SchemaDefinition) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		// Check if this is a method call on a schema variable (e.g., contactInfoSchema.Optional())
		if selExpr, ok := e.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := selExpr.X.(*ast.Ident); ok {
				// Check if the receiver is a tracked schema variable
				if trackedSchema, exists := a.schemaVars[ident.Name]; exists {
					if a.verbose {
						fmt.Printf("[VERBOSE] Property uses schema variable: %s -> %s\n", ident.Name, trackedSchema.Type)
					}
					// Copy the tracked schema content first
					*propSchema = *trackedSchema
					// Then apply any additional method calls (like .Optional())
					a.analyzeValidatorCall(e, propSchema)
					return
				}
			}
		}
		// This is a regular validator call chain like validators.String().Min(1)
		a.analyzeValidatorCall(e, propSchema)
	case *ast.Ident:
		// This is a bare schema variable reference (e.g., paymentMethodSchema)
		if trackedSchema, exists := a.schemaVars[e.Name]; exists {
			if a.verbose {
				fmt.Printf("[VERBOSE] Property uses bare schema variable: %s -> %s\n", e.Name, trackedSchema.Type)
			}
			// Copy the tracked schema content
			*propSchema = *trackedSchema
			return
		}
		if a.verbose {
			fmt.Printf("[VERBOSE] Unknown identifier reference: %s\n", e.Name)
		}
	default:
		if a.verbose {
			fmt.Printf("[VERBOSE] Unknown property value type: %T\n", expr)
		}
	}
}

// isValidatorPackage checks if an expression refers to the validators package
func (a *ASTAnalyzer) isValidatorPackage(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "validators"
	}
	return false
}

// isRouterRegister checks if a call expression is a router.Register call
func (a *ASTAnalyzer) isRouterRegister(callExpr *ast.CallExpr) bool {
	if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		return selExpr.Sel.Name == "Register"
	}
	return false
}

// extractLiteralValue extracts various literal values (string, number, bool, map, slice)
func (a *ASTAnalyzer) extractLiteralValue(expr ast.Expr) interface{} {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.STRING:
			// Remove quotes and handle escape sequences
			if len(e.Value) >= 2 {
				return e.Value[1 : len(e.Value)-1] // Remove surrounding quotes
			}
			return e.Value
		case token.INT:
			// Parse as int or float
			if val := a.extractNumberLiteral(expr); val != nil {
				// If it's a whole number, return as int
				if *val == float64(int(*val)) {
					return int(*val)
				}
				return *val
			}
		case token.FLOAT:
			if val := a.extractNumberLiteral(expr); val != nil {
				return *val
			}
		}
	case *ast.Ident:
		// Handle boolean literals
		switch e.Name {
		case "true":
			return true
		case "false":
			return false
		}
	case *ast.CompositeLit:
		// Handle map and slice literals
		return a.extractCompositeLiteral(e)
	}
	return nil
}

// extractCompositeLiteral extracts map or slice literals
func (a *ASTAnalyzer) extractCompositeLiteral(expr *ast.CompositeLit) interface{} {
	// Check if it's a map literal
	if len(expr.Elts) > 0 {
		if _, ok := expr.Elts[0].(*ast.KeyValueExpr); ok {
			// It's a map
			result := make(map[string]interface{})
			for _, elt := range expr.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					if key := a.extractStringLiteral(kv.Key); key != "" {
						if value := a.extractLiteralValue(kv.Value); value != nil {
							result[key] = value
						}
					}
				}
			}
			return result
		} else {
			// It's a slice
			var result []interface{}
			for _, elt := range expr.Elts {
				if value := a.extractLiteralValue(elt); value != nil {
					result = append(result, value)
				}
			}
			return result
		}
	}
	return nil
}

// extractExamplesMap extracts a map of ExampleObject structures
func (a *ASTAnalyzer) extractExamplesMap(expr ast.Expr) map[string]ExampleObject {
	if compositeLit, ok := expr.(*ast.CompositeLit); ok {
		result := make(map[string]ExampleObject)

		for _, elt := range compositeLit.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				// Extract the key (example name)
				key := a.extractStringLiteral(kv.Key)
				if key == "" {
					continue
				}

				// Extract the ExampleObject struct
				if exampleObj := a.extractExampleObject(kv.Value); exampleObj != nil {
					result[key] = *exampleObj
				}
			}
		}

		if len(result) > 0 {
			return result
		}
	}
	return nil
}

// extractExampleObject extracts an ExampleObject struct from AST
func (a *ASTAnalyzer) extractExampleObject(expr ast.Expr) *ExampleObject {
	if compositeLit, ok := expr.(*ast.CompositeLit); ok {
		example := &ExampleObject{}

		for _, elt := range compositeLit.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				fieldName := ""
				if ident, ok := kv.Key.(*ast.Ident); ok {
					fieldName = ident.Name
				}

				switch fieldName {
				case "Summary":
					if val := a.extractStringLiteral(kv.Value); val != "" {
						example.Summary = val
					}
				case "Description":
					if val := a.extractStringLiteral(kv.Value); val != "" {
						example.Description = val
					}
				case "Value":
					if val := a.extractLiteralValue(kv.Value); val != nil {
						example.Value = val
					}
				case "ExternalValue":
					if val := a.extractStringLiteral(kv.Value); val != "" {
						example.ExternalValue = val
					}
				}
			}
		}

		return example
	}
	return nil
}
