# a2t Protocol Design

## Philosophy

The a2t (Agent-to-Tool) protocol is built on three core principles:

1. **Simplicity**: Only what's needed for tool calling, nothing more
2. **Statelessness**: Every request is independent, no connection management
3. **Scalability**: Works from 1 tool to millions, from simple to complex

## Why Not MCP?

While MCP (Model Context Protocol) is powerful, it has some limitations:

- **Stateful**: Requires persistent connections via stdio or SSE
- **Complex**: Includes prompts, resources, and tool calling in one protocol
- **Limited organization**: No built-in concept of groups or hierarchical organization
- **No search**: Requires client-side filtering

a2t addresses these by focusing solely on tools and providing first-class support for organization and discovery.

## Architecture

### Three Layers

```
┌─────────────────────────────────────┐
│         HTTP REST API               │  ← Transport Layer
├─────────────────────────────────────┤
│    Server + Provider Interfaces     │  ← Abstraction Layer
├─────────────────────────────────────┤
│   SimpleProvider / GroupProvider    │  ← Implementation Layer
└─────────────────────────────────────┘
```

### Provider Pattern

The provider pattern allows implementations to be swapped without changing the HTTP layer:

```go
type ToolProvider interface {
    GetCapabilities() *Capabilities
    ListTools(ctx, groupID, offset, limit) (*ToolsResponse, error)
    ExecuteTool(ctx, *ExecuteRequest) (*ExecuteResponse, error)
}
```

Implementations can add features via interface composition:

```go
type GroupProvider interface {
    ToolProvider
    ListGroups(...) (*GroupsResponse, error)
    GetGroup(...) (*Group, error)
}

type SearchProvider interface {
    ToolProvider
    Search(...) (*SearchResponse, error)
}
```

## Capabilities Discovery

The `.well-known/a2t-capabilities.json` endpoint is central to the protocol. It tells clients:

1. What features are available (groups, search, dynamic tools)
2. What endpoints to use
3. What limits apply

This allows clients to adapt to different server implementations without hardcoding assumptions.

## Meta Responses

Meta responses solve the "dynamic discovery" problem. When a tool executes, it can return:

1. **New tools**: `tools_added` meta type provides tools the client didn't know about
2. **Group updates**: `group_refresh` indicates groups have changed and should be re-fetched

This enables workflows like:
- A tool that searches a plugin marketplace and returns newly installed tools
- A tool that loads a configuration file and exposes its options as new tools
- A tool that discovers available services and creates tools for each

## Scaling Strategy

### Simple (1-100 tools)
- No groups
- Return all tools in one `/tools` request
- Direct tool execution

### Medium (100-10,000 tools)
- Groups organize tools by domain
- Lazy loading via `/groups/{id}/tools`
- Optional search for quick discovery

### Large (10,000+ tools)
- Deep group hierarchies
- Pagination on all endpoints
- Search becomes primary discovery mechanism
- Consider database-backed provider implementation

## Implementation Patterns

### In-Memory Provider (SimpleProvider)

Best for:
- Development and testing
- Small tool catalogs
- Static tool sets

Trade-offs:
- Fast
- Simple
- Limited to memory constraints
- All tools loaded at startup

### Database-Backed Provider

Best for:
- Large tool catalogs
- Dynamic tool registration
- Multi-tenant systems

Trade-offs:
- Scales to millions of tools
- Supports complex queries
- Adds latency
- Requires infrastructure

### Federated Provider

Best for:
- Aggregating multiple a2t servers
- Distributed tool ecosystems
- Enterprise environments

Pattern:
```go
type FederatedProvider struct {
    upstreams []*a2t.Client
}

func (p *FederatedProvider) ListTools(...) {
    // Query all upstreams in parallel
    // Merge and deduplicate results
}
```

## Security Considerations

### Authentication

a2t is transport-agnostic. Add authentication at the HTTP layer:

```go
server := a2t.NewServer(provider)
handler := AuthMiddleware(server.Handler())
http.ListenAndServe(":8080", handler)
```

### Authorization

Implement in the provider:

```go
func (p *AuthProvider) ExecuteTool(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
    user := ctx.Value("user")
    if !p.authz.CanExecute(user, req.Tool) {
        return a2t.NewExecuteError("unauthorized", "Access denied")
    }
    return p.delegate.ExecuteTool(ctx, req)
}
```

### Rate Limiting

Add rate limiting middleware or implement in provider using context:

```go
func (p *RateLimitedProvider) ExecuteTool(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
    if !p.limiter.Allow(ctx) {
        return a2t.NewExecuteError("rate_limited", "Too many requests")
    }
    return p.delegate.ExecuteTool(ctx, req)
}
```

## Future Considerations

### Versioning

Capabilities include a version field. Future versions could:
- Add new meta response types
- Introduce new fields (backward compatible)
- Add optional query parameters

### Webhooks

For long-running tools, consider a webhook pattern:

```json
{
  "result": null,
  "meta": {
    "type": "async",
    "webhook_url": "https://server.com/webhooks/123",
    "estimated_completion": "2024-01-01T12:00:00Z"
  }
}
```

### Batch Execution

Allow executing multiple tools in one request:

```json
POST /execute/batch
{
  "executions": [
    {"tool": "add", "parameters": {"a": 1, "b": 2}},
    {"tool": "multiply", "parameters": {"a": 3, "b": 4}}
  ]
}
```

### Tool Dependencies

Express tool dependencies in metadata:

```json
{
  "name": "upload_file",
  "requires": ["create_session"],
  "input_schema": {...}
}
```

## Comparison Table

| Aspect | a2t | MCP | OpenAPI |
|--------|-----|-----|---------|
| Purpose | Tool calling | Multi-purpose protocol | API documentation |
| State | Stateless | Stateful | Stateless |
| Transport | HTTP | stdio/SSE | HTTP |
| Discovery | Capabilities file | Negotiation | OpenAPI spec |
| Organization | Groups | None | Tags |
| Search | Built-in | None | External |
| Complexity | Low | Medium | High |

## Conclusion

a2t provides a focused, scalable solution for tool calling between AI agents and tool providers. By keeping the protocol simple and stateless while providing powerful organization and discovery features, it serves both simple and complex use cases effectively.
