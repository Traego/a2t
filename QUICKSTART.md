# a2t Quickstart Guide

Get started with the a2t protocol in minutes.

## Installation

```bash
go get github.com/traego/a2t
```

## Simple Example

Create a basic tool server with a few tools:

```go
package main

import (
    "context"
    "log"

    "github.com/traego/a2t"
)

func main() {
    // Create a simple provider
    provider := a2t.NewSimpleProvider(nil)

    // Register a tool
    addTool := a2t.NewTool("add", "Add two numbers").
        WithProperty("a", "number", "First number", true).
        WithProperty("b", "number", "Second number", true)

    provider.RegisterTool(addTool, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
        a := params["a"].(float64)
        b := params["b"].(float64)
        return a + b, nil
    })

    // Start server
    server := a2t.NewServer(provider)
    log.Fatal(server.ListenAndServe(":8080"))
}
```

## Testing Your Server

```bash
# Get capabilities
curl http://localhost:8080/.well-known/a2t-capabilities.json

# List tools
curl http://localhost:8080/tools

# Search for tools
curl 'http://localhost:8080/tools?q=add'

# Execute a tool
curl -X POST http://localhost:8080/tools/add \
  -H "Content-Type: application/json" \
  -d '{"a":5,"b":3}'
```

## With Groups

For organizing tools into groups:

```go
// Create provider with groups
capabilities := a2t.NewCapabilities().WithGroups("")
provider := a2t.NewGroupProvider(capabilities)

// Register a group
mathGroup := a2t.NewGroup("math", "Math Tools", "Mathematical operations")
provider.RegisterGroup(mathGroup)

// Register tool in group
addTool := a2t.NewTool("add", "Add numbers").
    WithProperty("a", "number", "First number", true).
    WithProperty("b", "number", "Second number", true).
    WithGroup("math")

provider.RegisterTool(addTool, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    a := params["a"].(float64)
    b := params["b"].(float64)
    return a + b, nil
})
```

## Dynamic Tool Discovery

Return meta responses to inform clients about new tools:

```go
provider.RegisterTool(discoveryTool, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    newTool := a2t.NewTool("subtract", "Subtract numbers").
        WithProperty("a", "number", "First number", true).
        WithProperty("b", "number", "Second number", true)

    return map[string]interface{}{
        "result": "Discovered new tool",
        "meta": map[string]interface{}{
            "type": "tools_added",
            "tools": []interface{}{newTool},
        },
    }, nil
})
```

## Running Examples

```bash
# Simple example
cd examples/simple
go run main.go

# Advanced example with groups
cd examples/advanced
go run main.go
```

## Next Steps

- Read the [README](README.md) for full protocol details
- Check out the [examples](examples/) directory
- Implement custom providers for your use case
- Add search capabilities with the `SearchProvider` interface

## Client Implementation

To implement a client:

1. Fetch `/.well-known/a2t-capabilities.json` to discover features
2. List tools from `/tools` or `/groups/{id}/tools`
3. Execute tools via POST to `/execute`
4. Handle meta responses for dynamic tool discovery

Example in Python:

```python
import requests

# Discover capabilities
caps = requests.get('http://localhost:8080/.well-known/a2t-capabilities.json').json()
print(f"Server version: {caps['version']}")

# List tools
tools = requests.get('http://localhost:8080/tools').json()
for tool in tools['tools']:
    print(f"Tool: {tool['name']} - {tool['description']}")

# Execute a tool
result = requests.post('http://localhost:8080/tools/add', json={
    'a': 10,
    'b': 5
}).json()
print(f"Result: {result['result']}")

# Search for tools
tools = requests.get('http://localhost:8080/tools', params={'q': 'weather'}).json()
print(f"Found {len(tools['tools'])} tools")
```
