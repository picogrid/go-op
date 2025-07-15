package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// Tests for AST analyzer functions that extract values from Go source code for OpenAPI generation
func TestASTAnalyzer_ExtractLiteralValue(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewASTAnalyzer(fset, false)

	t.Run("extracts string literals with quote removal", func(t *testing.T) {
		src := `package main
		var s = "hello world"`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the string literal
		var stringLit *ast.BasicLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				stringLit = lit
				return false
			}
			return true
		})

		if stringLit == nil {
			t.Fatal("String literal not found")
		}

		value := analyzer.extractLiteralValue(stringLit)
		if value != "hello world" {
			t.Errorf("Expected 'hello world', got %v", value)
		}
	})

	t.Run("Extract number literal", func(t *testing.T) {
		src := `package main
		var n = 10`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the number literal
		var numberLit *ast.BasicLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.INT {
				numberLit = lit
				return false
			}
			return true
		})

		if numberLit == nil {
			t.Fatal("Number literal not found")
		}

		value := analyzer.extractLiteralValue(numberLit)
		if value != 10 {
			t.Errorf("Expected 10, got %v", value)
		}
	})

	t.Run("Extract float literal", func(t *testing.T) {
		src := `package main
		var f = 0.5`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the float literal
		var floatLit *ast.BasicLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.FLOAT {
				floatLit = lit
				return false
			}
			return true
		})

		if floatLit == nil {
			t.Fatal("Float literal not found")
		}

		value := analyzer.extractLiteralValue(floatLit)
		if value != 0.5 {
			t.Errorf("Expected 0.5, got %v", value)
		}
	})

	t.Run("Extract boolean literal", func(t *testing.T) {
		src := `package main
		var b = true`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the boolean literal (identifier)
		var boolLit *ast.Ident
		ast.Inspect(file, func(n ast.Node) bool {
			if ident, ok := n.(*ast.Ident); ok && ident.Name == "true" {
				boolLit = ident
				return false
			}
			return true
		})

		if boolLit == nil {
			t.Fatal("Boolean literal not found")
		}

		value := analyzer.extractLiteralValue(boolLit)
		if value != true {
			t.Errorf("Expected true, got %v", value)
		}
	})

	t.Run("Extract nil literal", func(t *testing.T) {
		src := `package main
		var n = nil`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the nil literal (identifier)
		var nilLit *ast.Ident
		ast.Inspect(file, func(n ast.Node) bool {
			if ident, ok := n.(*ast.Ident); ok && ident.Name == "nil" {
				nilLit = ident
				return false
			}
			return true
		})

		if nilLit == nil {
			t.Fatal("Nil literal not found")
		}

		value := analyzer.extractLiteralValue(nilLit)
		if value != nil {
			t.Errorf("Expected nil, got %v", value)
		}
	})

	t.Run("Extract unknown literal type", func(t *testing.T) {
		src := `package main
		var i = 'c'` // char literal

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the char literal
		var charLit *ast.BasicLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.CHAR {
				charLit = lit
				return false
			}
			return true
		})

		if charLit == nil {
			t.Fatal("Char literal not found")
		}

		value := analyzer.extractLiteralValue(charLit)
		if value != nil {
			t.Errorf("Expected nil for unsupported type, got %v", value)
		}
	})
}

// TestExtractCompositeLiteral tests the extractCompositeLiteral function (0% coverage)
func TestExtractCompositeLiteral(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewASTAnalyzer(fset, false)

	t.Run("Extract slice literal", func(t *testing.T) {
		src := `package main
		var arr = []string{"hello", "world"}`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the composite literal
		var compLit *ast.CompositeLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				compLit = lit
				return false
			}
			return true
		})

		if compLit == nil {
			t.Fatal("Composite literal not found")
		}

		value := analyzer.extractCompositeLiteral(compLit)
		if slice, ok := value.([]interface{}); ok {
			if len(slice) != 2 {
				t.Errorf("Expected slice of length 2, got %d", len(slice))
			}
			if slice[0] != "hello" || slice[1] != "world" {
				t.Errorf("Expected [hello, world], got %v", slice)
			}
		} else {
			t.Errorf("Expected slice, got %T", value)
		}
	})

	t.Run("Extract map literal", func(t *testing.T) {
		src := `package main
		var m = map[string]int{"key1": 1, "key2": 2}`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the composite literal
		var compLit *ast.CompositeLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				compLit = lit
				return false
			}
			return true
		})

		if compLit == nil {
			t.Fatal("Composite literal not found")
		}

		value := analyzer.extractCompositeLiteral(compLit)
		if m, ok := value.(map[string]interface{}); ok {
			if len(m) != 2 {
				t.Errorf("Expected map of length 2, got %d", len(m))
			}
			if m["key1"] != 1 || m["key2"] != 2 {
				t.Errorf("Expected map[key1:1 key2:2], got %v", m)
			}
		} else {
			t.Errorf("Expected map, got %T", value)
		}
	})

	t.Run("Extract struct-like map literal", func(t *testing.T) {
		src := `package main
		var p = map[string]interface{}{"Name": "John", "Age": 10}`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the composite literal
		var compLit *ast.CompositeLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				compLit = lit
				return false
			}
			return true
		})

		if compLit == nil {
			t.Fatal("Composite literal not found")
		}

		value := analyzer.extractCompositeLiteral(compLit)
		if m, ok := value.(map[string]interface{}); ok {
			if len(m) != 2 {
				t.Errorf("Expected map of length 2, got %d", len(m))
			}
			if m["Name"] != "John" || m["Age"] != 10 {
				t.Errorf("Expected map[Name:John Age:10], got %v", m)
			}
		} else {
			t.Errorf("Expected map, got %T", value)
		}
	})

	t.Run("Extract empty literal", func(t *testing.T) {
		src := `package main
		var arr = []string{}`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the composite literal
		var compLit *ast.CompositeLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				compLit = lit
				return false
			}
			return true
		})

		if compLit == nil {
			t.Fatal("Composite literal not found")
		}

		value := analyzer.extractCompositeLiteral(compLit)
		// For empty literals, the function returns nil
		if value != nil {
			t.Errorf("Expected nil for empty literal, got %v", value)
		}
	})
}

// TestExtractExamplesMap tests the extractExamplesMap function (0% coverage)
func TestExtractExamplesMap(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewASTAnalyzer(fset, false)

	t.Run("Extract examples map", func(t *testing.T) {
		src := `package main
		import "github.com/picogrid/go-op/validators"
		var examples = map[string]validators.ExampleObject{
			"simple": {Value: "test", Summary: "Simple example"},
			"complex": {Value: map[string]interface{}{"key": "value"}, Summary: "Complex example", Description: "A complex example"},
		}`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the composite literal for the examples map
		var examplesLit *ast.CompositeLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				// Look for the map type with ExampleObject
				if mapType, ok := lit.Type.(*ast.MapType); ok {
					if sel, ok := mapType.Value.(*ast.SelectorExpr); ok {
						if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "validators" && sel.Sel.Name == "ExampleObject" {
							examplesLit = lit
							return false
						}
					}
				}
			}
			return true
		})

		if examplesLit == nil {
			t.Fatal("Examples composite literal not found")
		}

		examples := analyzer.extractExamplesMap(examplesLit)
		if len(examples) != 2 {
			t.Errorf("Expected 2 examples, got %d", len(examples))
		}

		if simple, ok := examples["simple"]; ok {
			if simple.Value != "test" {
				t.Errorf("Expected simple.Value = 'test', got %v", simple.Value)
			}
			if simple.Summary != "Simple example" {
				t.Errorf("Expected simple.Summary = 'Simple example', got %v", simple.Summary)
			}
		} else {
			t.Error("'simple' example not found")
		}

		if complex, ok := examples["complex"]; ok {
			if complex.Summary != "Complex example" {
				t.Errorf("Expected complex.Summary = 'Complex example', got %v", complex.Summary)
			}
			if complex.Description != "A complex example" {
				t.Errorf("Expected complex.Description = 'A complex example', got %v", complex.Description)
			}
		} else {
			t.Error("'complex' example not found")
		}
	})

	t.Run("Extract empty examples map", func(t *testing.T) {
		src := `package main
		import "github.com/picogrid/go-op/validators"
		var examples = map[string]validators.ExampleObject{}`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the composite literal for the examples map
		var examplesLit *ast.CompositeLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				if mapType, ok := lit.Type.(*ast.MapType); ok {
					if sel, ok := mapType.Value.(*ast.SelectorExpr); ok {
						if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "validators" && sel.Sel.Name == "ExampleObject" {
							examplesLit = lit
							return false
						}
					}
				}
			}
			return true
		})

		if examplesLit == nil {
			t.Fatal("Examples composite literal not found")
		}

		examples := analyzer.extractExamplesMap(examplesLit)
		if len(examples) != 0 {
			t.Errorf("Expected 0 examples, got %d", len(examples))
		}
	})
}

// TestExtractExampleObject tests the extractExampleObject function (0% coverage)
func TestExtractExampleObject(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewASTAnalyzer(fset, false)

	t.Run("Extract complete example object", func(t *testing.T) {
		src := `package main
		import "github.com/picogrid/go-op/validators"
		var example = validators.ExampleObject{
			Value: "test value",
			Summary: "Test summary",
			Description: "Test description",
			ExternalValue: "https://example.com/external",
		}`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the composite literal for ExampleObject
		var exampleLit *ast.CompositeLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				if sel, ok := lit.Type.(*ast.SelectorExpr); ok {
					if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "validators" && sel.Sel.Name == "ExampleObject" {
						exampleLit = lit
						return false
					}
				}
			}
			return true
		})

		if exampleLit == nil {
			t.Fatal("ExampleObject composite literal not found")
		}

		example := analyzer.extractExampleObject(exampleLit)
		if example.Value != "test value" {
			t.Errorf("Expected Value = 'test value', got %v", example.Value)
		}
		if example.Summary != "Test summary" {
			t.Errorf("Expected Summary = 'Test summary', got %v", example.Summary)
		}
		if example.Description != "Test description" {
			t.Errorf("Expected Description = 'Test description', got %v", example.Description)
		}
		if example.ExternalValue != "https://example.com/external" {
			t.Errorf("Expected ExternalValue = 'https://example.com/external', got %v", example.ExternalValue)
		}
	})

	t.Run("Extract minimal example object", func(t *testing.T) {
		src := `package main
		import "github.com/picogrid/go-op/validators"
		var example = validators.ExampleObject{
			Value: 10,
		}`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the composite literal for ExampleObject
		var exampleLit *ast.CompositeLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				if sel, ok := lit.Type.(*ast.SelectorExpr); ok {
					if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "validators" && sel.Sel.Name == "ExampleObject" {
						exampleLit = lit
						return false
					}
				}
			}
			return true
		})

		if exampleLit == nil {
			t.Fatal("ExampleObject composite literal not found")
		}

		example := analyzer.extractExampleObject(exampleLit)
		if example.Value != 10 {
			t.Errorf("Expected Value = 10, got %v", example.Value)
		}
		if example.Summary != "" {
			t.Errorf("Expected empty Summary, got %v", example.Summary)
		}
		if example.Description != "" {
			t.Errorf("Expected empty Description, got %v", example.Description)
		}
		if example.ExternalValue != "" {
			t.Errorf("Expected empty ExternalValue, got %v", example.ExternalValue)
		}
	})

	t.Run("Extract example object with complex value", func(t *testing.T) {
		src := `package main
		import "github.com/picogrid/go-op/validators"
		var example = validators.ExampleObject{
			Value: map[string]interface{}{"name": "John", "age": 10},
			Summary: "User object",
		}`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the composite literal for ExampleObject
		var exampleLit *ast.CompositeLit
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				if sel, ok := lit.Type.(*ast.SelectorExpr); ok {
					if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "validators" && sel.Sel.Name == "ExampleObject" {
						exampleLit = lit
						return false
					}
				}
			}
			return true
		})

		if exampleLit == nil {
			t.Fatal("ExampleObject composite literal not found")
		}

		example := analyzer.extractExampleObject(exampleLit)
		if example.Summary != "User object" {
			t.Errorf("Expected Summary = 'User object', got %v", example.Summary)
		}

		// Check the complex value
		if valueMap, ok := example.Value.(map[string]interface{}); ok {
			if valueMap["name"] != "John" || valueMap["age"] != 10 {
				t.Errorf("Expected complex value map[name:John age:10], got %v", valueMap)
			}
		} else {
			t.Errorf("Expected complex value to be a map, got %T", example.Value)
		}
	})
}

// TestTraverseValidatorChainPartialCoverage tests missing branches in traverseValidatorChain
func TestTraverseValidatorChainPartialCoverage(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewASTAnalyzer(fset, false)

	t.Run("Handle unknown method call", func(t *testing.T) {
		src := `package main
		import "github.com/picogrid/go-op/validators"
		var schema = validators.String().UnknownMethod().Required()`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the call chain
		var callExpr *ast.CallExpr
		ast.Inspect(file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Required" {
					callExpr = call
					return false
				}
			}
			return true
		})

		if callExpr == nil {
			t.Fatal("Call expression not found")
		}

		// This should handle the unknown method gracefully
		// Create a basic schema to pass to the function
		schema := &SchemaDefinition{Type: "string"}
		analyzer.traverseValidatorChain(callExpr, schema)
		// If we get here without panic, the function handled the unknown method gracefully
	})

	t.Run("Handle nested call expressions", func(t *testing.T) {
		src := `package main
		import "github.com/picogrid/go-op/validators"
		var schema = validators.String().Pattern(getPattern()).Required()`

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Find the call chain
		var callExpr *ast.CallExpr
		ast.Inspect(file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Required" {
					callExpr = call
					return false
				}
			}
			return true
		})

		if callExpr == nil {
			t.Fatal("Call expression not found")
		}

		// This should handle function calls as arguments
		// Create a basic schema to pass to the function
		schema := &SchemaDefinition{Type: "string"}
		analyzer.traverseValidatorChain(callExpr, schema)
		// If we get here without panic, the function handled the nested call gracefully
	})
}
