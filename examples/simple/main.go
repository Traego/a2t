package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/traego/a2t"
)

func main() {
	// Create a simple provider with basic capabilities
	capabilities := a2t.NewCapabilities()
	provider := a2t.NewSimpleProvider(capabilities)

	// Register a simple calculator tool
	addTool := a2t.NewTool("add", "Add two numbers together").
		WithProperty("a", "number", "First number", true).
		WithProperty("b", "number", "Second number", true)

	provider.RegisterTool(addTool, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		a, _ := params["a"].(float64)
		b, _ := params["b"].(float64)
		return a + b, nil
	})

	// Register a weather tool
	weatherTool := a2t.NewTool("get_weather", "Get current weather for a location").
		WithProperty("location", "string", "City name or coordinates", true)

	provider.RegisterTool(weatherTool, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		location := params["location"].(string)
		// Simulated weather data
		return map[string]interface{}{
			"location":    location,
			"temperature": 72,
			"condition":   "Sunny",
			"timestamp":   time.Now().Format(time.RFC3339),
		}, nil
	})

	// Register a time tool
	timeTool := a2t.NewTool("get_time", "Get current time in a timezone").
		WithProperty("timezone", "string", "IANA timezone (e.g., America/New_York)", false)

	provider.RegisterTool(timeTool, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		timezone := "UTC"
		if tz, ok := params["timezone"].(string); ok {
			timezone = tz
		}

		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone: %s", timezone)
		}

		return map[string]interface{}{
			"timezone": timezone,
			"time":     time.Now().In(loc).Format(time.RFC3339),
		}, nil
	})

	// Create and start server
	server := a2t.NewServer(provider)

	fmt.Println("Starting simple a2t server with OpenAPI/Swagger documentation...")
	fmt.Println("\nAPI Documentation:")
	fmt.Println("  Swagger UI: http://localhost:8080/docs")
	fmt.Println("  OpenAPI JSON: http://localhost:8080/docs/openapi.json")
	fmt.Println("\nTry:")
	fmt.Println("  curl http://localhost:8080/.well-known/a2t-capabilities.json")
	fmt.Println("  curl http://localhost:8080/tools")
	fmt.Println("  curl http://localhost:8080/tools?q=weather")
	fmt.Println("  curl -X POST http://localhost:8080/tools/add -d '{\"a\":5,\"b\":3}'")
	fmt.Println("  curl -X POST http://localhost:8080/tools/get_weather -d '{\"location\":\"San Francisco\"}'")
	fmt.Println()

	if err := server.ListenAndServe(":8080"); err != nil {
		log.Fatal(err)
	}
}
