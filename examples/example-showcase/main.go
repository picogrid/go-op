package main

import (
	"fmt"

	goop "github.com/picogrid/go-op"
)

// ExampleShowcase demonstrates OpenAPI 3.1 Example Object functionality
func main() {
	fmt.Println("🎯 OpenAPI 3.1 Example Object Showcase")
	fmt.Println("=====================================")

	// 1. Simple Example - Single value
	fmt.Println("\n1. Simple Example Usage:")
	emailSchema := validators.String().
		Email().
		Example("user@example.com").
		Required()

	showSchema("Email with simple example", emailSchema)

	// 2. Multiple Examples - Rich documentation
	fmt.Println("\n2. Multiple Examples with Descriptions:")
	usernameSchema := validators.String().
		Min(3).Max(50).
		Pattern("^[a-zA-Z0-9_]+$").
		Examples(map[string]validators.ExampleObject{
			"simple": {
				Summary:     "Simple username",
				Description: "A basic alphanumeric username",
				Value:       "johndoe",
			},
			"with_underscore": {
				Summary:     "Username with underscore",
				Description: "Username containing underscores for readability",
				Value:       "john_doe_123",
			},
			"numeric": {
				Summary:     "Numeric username",
				Description: "Username with numbers",
				Value:       "user2024",
			},
		}).
		Required()

	showSchema("Username with multiple examples", usernameSchema)

	// 3. External File Reference
	fmt.Println("\n3. External File Reference:")
	configSchema := validators.Object(map[string]interface{}{
		"database_url": validators.String().Required(),
		"port":         validators.Number().Min(1024).Max(65535).Required(),
		"debug":        validators.Bool().Required(),
	}).ExampleFromFile("./config.json").Required()

	showSchema("Config with external example", configSchema)

	// 4. Complex Object with Nested Examples
	fmt.Println("\n4. Complex Object with Nested Examples:")
	userSchema := validators.Object(map[string]interface{}{
		"name": validators.String().
			Example("John Doe").
			Required(),
		"age": validators.Number().Min(18).Max(120).
			Examples(map[string]validators.ExampleObject{
				"young_adult": {
					Summary:     "Young adult",
					Description: "Typical age for college students",
					Value:       22,
				},
				"middle_aged": {
					Summary:     "Middle-aged",
					Description: "Professional working age",
					Value:       35,
				},
				"senior": {
					Summary:     "Senior citizen",
					Description: "Retirement age",
					Value:       65,
				},
			}).
			Required(),
		"preferences": validators.Object(map[string]interface{}{
			"newsletter": validators.Bool().
				Example(true).
				Required(),
			"theme": validators.String().
				Examples(map[string]validators.ExampleObject{
					"light": {
						Summary:     "Light theme",
						Description: "Bright theme for daytime use",
						Value:       "light",
					},
					"dark": {
						Summary:     "Dark theme",
						Description: "Dark theme for low-light environments",
						Value:       "dark",
					},
					"auto": {
						Summary:     "Auto theme",
						Description: "Automatically switch based on system settings",
						Value:       "auto",
					},
				}).
				Required(),
		}).Required(),
	}).Example(map[string]interface{}{
		"name": "John Doe",
		"age":  28,
		"preferences": map[string]interface{}{
			"newsletter": true,
			"theme":      "dark",
		},
	}).Required()

	showSchema("User with nested examples", userSchema)

	// 5. Array Examples
	fmt.Println("\n5. Array with Examples:")
	tagsSchema := validators.Array(
		validators.String().
			Examples(map[string]validators.ExampleObject{
				"tech": {
					Summary:     "Technology tag",
					Description: "Tags related to technology",
					Value:       "javascript",
				},
				"business": {
					Summary:     "Business tag",
					Description: "Tags related to business",
					Value:       "marketing",
				},
			}).
			Required(),
	).Example([]interface{}{
		"javascript", "react", "nodejs", "api",
	}).Required()

	showSchema("Tags array with examples", tagsSchema)

	fmt.Println("\n✅ Example showcase complete!")
	fmt.Println("\nNext steps:")
	fmt.Println("- These examples will be extracted by AST analysis")
	fmt.Println("- They'll appear in generated OpenAPI 3.1 specifications")
	fmt.Println("- They provide rich documentation for API consumers")
}

// showSchema demonstrates OpenAPI generation for a schema
func showSchema(title string, schema interface{}) {
	fmt.Printf("\n📋 %s:\n", title)

	// Try to convert to enhanced schema and show OpenAPI output
	if enhancedSchema, ok := schema.(goop.EnhancedSchema); ok {
		openAPISchema := enhancedSchema.ToOpenAPISchema()
		fmt.Printf("   Type: %s\n", openAPISchema.Type)

		if openAPISchema.Example != nil {
			fmt.Printf("   Example: %v\n", openAPISchema.Example)
		}

		if openAPISchema.Format != "" {
			fmt.Printf("   Format: %s\n", openAPISchema.Format)
		}

		if openAPISchema.MinLength != nil {
			fmt.Printf("   MinLength: %d\n", *openAPISchema.MinLength)
		}

		if openAPISchema.Pattern != "" {
			fmt.Printf("   Pattern: %s\n", openAPISchema.Pattern)
		}

		if openAPISchema.Properties != nil {
			fmt.Printf("   Properties: %d fields\n", len(openAPISchema.Properties))
		}
	} else {
		fmt.Println("   (Schema does not implement EnhancedSchema)")
	}
}
