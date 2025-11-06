# a2t - Agent-to-Tool Protocol

A simple, stateless protocol for tool calling between AI agents and tool providers.

## Overview

a2t (Agent-to-Tool) is a minimalist alternative to MCP that focuses exclusively on tool calling. It's designed to be:

- **Simple**: Only tools, groups, and search - nothing more
- **Stateless**: Every request is independent
- **Scalable**: From a single tool to millions, from no groups to thousands
- **Standards-compliant**: Tool schemas match LLM provider formats (OpenAI, Anthropic, etc.)

## Core Concepts

### Tools

Tools are functions that an AI agent can call. They follow standard LLM tool calling schemas:

```json
{
  "name": "get_weather",
  "description": "Get current weather for a location",
  "input_schema": {
    "type": "object",
    "properties": {
      "location": {
        "type": "string",
        "description": "City name or coordinates"
      }
    },
    "required": ["location"]
  }
}
```

### Groups

Groups organize tools hierarchically. They're optional but useful for:

- Namespacing tools by domain
- Lazy loading tools on demand
- Managing large tool catalogs

```json
{
  "id": "weather",
  "name": "Weather Tools",
  "description": "Tools for weather information",
  "tool_count": 5
}
```

### Capabilities

The well-known capabilities file (`.well-known/a2t-capabilities.json`) declares what a server supports:

```json
{
  "version": "1.0",
  "features": {
    "groups": true,
    "search": true,
    "dynamic_tools": true
  },
  "endpoints": {
    "tools": "/tools",
    "groups": "/groups"
  },
  "limits": {
    "max_tools_per_request": 100,
    "max_groups_per_request": 50
  }
}
```

## Protocol Flow

### Simple Instance (No Groups)

1. Client fetches `/.well-known/a2t-capabilities.json`
2. Client fetches `/tools` to get all available tools
3. Client calls `POST /tools/{name}` with parameters
4. Server returns result

### Complex Instance (With Groups & Search)

1. Client fetches capabilities
2. Client searches `/tools?q=weather` or `/groups?q=weather` to find relevant tools/groups
3. Client fetches `/groups/{id}/tools` for specific group
4. Client calls `POST /tools/{name}` or `POST /groups/{id}/tools/{name}` with parameters
5. Server may return meta responses (new tools, group refresh pointers)

## Meta Responses

Meta responses allow servers to dynamically provide additional tools or update the client's tool catalog:

```json
{
  "result": "...",
  "meta": {
    "type": "tools_added",
    "tools": [
      {
        "name": "new_tool",
        "description": "Dynamically discovered tool",
        "input_schema": {...}
      }
    ]
  }
}
```

Or point to groups that need refreshing:

```json
{
  "result": "...",
  "meta": {
    "type": "group_refresh",
    "group_ids": ["weather", "location"]
  }
}
```

## API Endpoints

### GET /.well-known/a2t-capabilities.json

Returns server capabilities.

### GET /tools

Returns all available tools (if no groups) or top-level tools.

Query parameters:
- `q`: Search query (optional) - filters tools by name/description
- `limit`: Max tools to return (optional)
- `offset`: Pagination offset (optional)

Response:
```json
{
  "tools": [...],
  "total": 42,
  "offset": 0,
  "limit": 100
}
```

### GET /groups

Returns available groups.

Query parameters:
- `q`: Search query (optional) - filters groups by name/description
- `parent_id`: Filter by parent group (optional)
- `limit`: Max groups to return (optional)
- `offset`: Pagination offset (optional)

Response:
```json
{
  "groups": [...],
  "total": 10,
  "offset": 0,
  "limit": 50
}
```

### GET /groups/{id}/tools

Returns tools in a specific group.

Query parameters:
- `q`: Search query (optional)
- `limit`: Max tools to return (optional)
- `offset`: Pagination offset (optional)

### POST /tools/{name}

Execute a tool.

Request body (tool parameters):
```json
{
  "location": "San Francisco"
}
```

Response:
```json
{
  "result": "Sunny, 72°F",
  "meta": {
    "type": "tools_added",
    "data": {...}
  }
}
```

### POST /groups/{id}/tools/{name}

Execute a tool within a specific group context. Same request/response format as `POST /tools/{name}`.

## Design Principles

1. **Stateless**: No sessions, no connection management
2. **HTTP-first**: Simple REST API that works anywhere
3. **Pagination**: All list endpoints support pagination
4. **Discovery**: Capabilities file makes features explicit
5. **Extensible**: Meta responses allow dynamic behavior
6. **Minimal**: Only what's needed for tool calling

## Comparison to MCP

| Feature | a2t | MCP |
|---------|-----|-----|
| Stateless | ✓ | ✗ (connection-based) |
| Tool calling | ✓ | ✓ |
| Prompts/Resources | ✗ | ✓ |
| Groups/Organization | ✓ | ✗ |
| Search | ✓ | ✗ |
| Dynamic discovery | ✓ | Limited |
| Protocol | HTTP/REST | JSON-RPC over stdio/SSE |

## Use Cases

### Simple: Personal Tool Server

A local server exposing 10-20 tools without groups or search.

### Medium: Domain-Specific API

A service offering 100-200 tools organized into groups by domain.

### Complex: Tool Marketplace

A platform with thousands of groups and millions of tools, requiring search and lazy loading.

## Getting Started

See the Go library and example implementations in this repository.

```bash
# Install library
go get github.com/yourorg/a2t

# Run example server
cd examples/simple
go run main.go
```

## License

MIT
