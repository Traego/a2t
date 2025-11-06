package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/traego/a2t"
)

func main() {
	// Create a provider with groups and dynamic tools
	capabilities := a2t.NewCapabilities().
		WithGroups("").
		WithDynamicTools().
		WithLimits(&a2t.LimitsConfig{
			MaxToolsPerRequest:  100,
			MaxGroupsPerRequest: 50,
		})

	provider := a2t.NewGroupProvider(capabilities)

	// Create groups
	mathGroup := a2t.NewGroup("math", "Mathematics", "Mathematical operations").
		WithToolCount(3)

	stringGroup := a2t.NewGroup("string", "String Operations", "Tools for manipulating strings").
		WithToolCount(2)

	provider.RegisterGroup(mathGroup)
	provider.RegisterGroup(stringGroup)

	// Register math tools
	addTool := a2t.NewTool("add", "Add two numbers").
		WithProperty("a", "number", "First number", true).
		WithProperty("b", "number", "Second number", true).
		WithGroup("math")

	provider.RegisterTool(addTool, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		a, _ := params["a"].(float64)
		b, _ := params["b"].(float64)
		return a + b, nil
	})

	multiplyTool := a2t.NewTool("multiply", "Multiply two numbers").
		WithProperty("a", "number", "First number", true).
		WithProperty("b", "number", "Second number", true).
		WithGroup("math")

	provider.RegisterTool(multiplyTool, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		a, _ := params["a"].(float64)
		b, _ := params["b"].(float64)
		return a * b, nil
	})

	// Special tool that demonstrates meta responses by dynamically adding a subtract tool
	discoverTool := a2t.NewTool("discover_math_tools", "Discover additional math tools").
		WithGroup("math")

	provider.RegisterTool(discoverTool, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		// Return a response with meta information about new tools
		subtractTool := a2t.NewTool("subtract", "Subtract two numbers").
			WithProperty("a", "number", "First number", true).
			WithProperty("b", "number", "Second number", true).
			WithGroup("math")

		// In a real implementation, we would register this tool for future use
		// For now, we just return it in the meta response

		return map[string]interface{}{
			"result": "Discovered 1 new math tool",
			"meta": map[string]interface{}{
				"type": "tools_added",
				"tools": []interface{}{
					map[string]interface{}{
						"name":        subtractTool.Name,
						"description": subtractTool.Description,
						"input_schema": subtractTool.InputSchema,
						"group_id":    subtractTool.GroupID,
					},
				},
			},
		}, nil
	})

	// Register string tools
	uppercaseTool := a2t.NewTool("uppercase", "Convert text to uppercase").
		WithProperty("text", "string", "Text to convert", true).
		WithGroup("string")

	provider.RegisterTool(uppercaseTool, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		text, _ := params["text"].(string)
		return strings.ToUpper(text), nil
	})

	reverseTool := a2t.NewTool("reverse", "Reverse a string").
		WithProperty("text", "string", "Text to reverse", true).
		WithGroup("string")

	provider.RegisterTool(reverseTool, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		text, _ := params["text"].(string)
		runes := []rune(text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	})

	// Create and start server
	server := a2t.NewServer(provider)

	fmt.Println("Starting advanced a2t server with groups and OpenAPI/Swagger...")
	fmt.Println("\nAPI Documentation:")
	fmt.Println("  Swagger UI: http://localhost:8081/docs")
	fmt.Println("  OpenAPI JSON: http://localhost:8081/docs/openapi.json")
	fmt.Println("\nTry:")
	fmt.Println("  # Get capabilities")
	fmt.Println("  curl http://localhost:8081/.well-known/a2t-capabilities.json")
	fmt.Println()
	fmt.Println("  # List all groups")
	fmt.Println("  curl http://localhost:8081/groups")
	fmt.Println()
	fmt.Println("  # Search groups")
	fmt.Println("  curl http://localhost:8081/groups?q=math")
	fmt.Println()
	fmt.Println("  # List tools in math group")
	fmt.Println("  curl http://localhost:8081/groups/math/tools")
	fmt.Println()
	fmt.Println("  # Search tools")
	fmt.Println("  curl http://localhost:8081/tools?q=add")
	fmt.Println()
	fmt.Println("  # Execute a tool")
	fmt.Println("  curl -X POST http://localhost:8081/tools/add -d '{\"a\":10,\"b\":5}'")
	fmt.Println()
	fmt.Println("  # Execute tool from group")
	fmt.Println("  curl -X POST http://localhost:8081/groups/math/tools/multiply -d '{\"a\":6,\"b\":7}'")
	fmt.Println()
	fmt.Println("  # Discover new tools (meta response)")
	fmt.Println("  curl -X POST http://localhost:8081/tools/discover_math_tools -d '{}'")
	fmt.Println()

	if err := server.ListenAndServe(":8081"); err != nil {
		log.Fatal(err)
	}
}
