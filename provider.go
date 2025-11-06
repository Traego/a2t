package a2t

import (
	"context"
	"strings"
)

// ToolProvider is the main interface that tool implementations must satisfy.
type ToolProvider interface {
	// GetCapabilities returns the server's capabilities.
	GetCapabilities() *Capabilities

	// ListTools returns available tools.
	// If groupID is provided, only tools in that group are returned.
	// If query is provided, tools are filtered by name/description.
	ListTools(ctx context.Context, groupID, query string, offset, limit int) (*ToolsResponse, error)

	// ExecuteTool executes a tool and returns the result.
	// The result can include meta responses for dynamic tool discovery.
	ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*ExecuteResponse, error)
}

// GroupProvider is an optional interface for providers that support groups.
type GroupProvider interface {
	ToolProvider

	// ListGroups returns available groups.
	// If parentID is provided, only child groups are returned.
	// If query is provided, groups are filtered by name/description.
	ListGroups(ctx context.Context, parentID, query string, offset, limit int) (*GroupsResponse, error)

	// GetGroup returns a specific group by ID.
	GetGroup(ctx context.Context, groupID string) (*Group, error)
}

// ToolExecutor is a function that executes a tool.
type ToolExecutor func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// SimpleProvider is a basic in-memory implementation of ToolProvider.
type SimpleProvider struct {
	capabilities *Capabilities
	tools        map[string]*Tool
	executors    map[string]ToolExecutor
}

// NewSimpleProvider creates a new simple provider.
func NewSimpleProvider(capabilities *Capabilities) *SimpleProvider {
	if capabilities == nil {
		capabilities = NewCapabilities()
	}
	return &SimpleProvider{
		capabilities: capabilities,
		tools:        make(map[string]*Tool),
		executors:    make(map[string]ToolExecutor),
	}
}

// RegisterTool registers a tool with its executor function.
func (p *SimpleProvider) RegisterTool(tool *Tool, executor ToolExecutor) {
	p.tools[tool.Name] = tool
	p.executors[tool.Name] = executor
}

// GetCapabilities returns the provider's capabilities.
func (p *SimpleProvider) GetCapabilities() *Capabilities {
	return p.capabilities
}

// ListTools returns all registered tools.
func (p *SimpleProvider) ListTools(ctx context.Context, groupID, query string, offset, limit int) (*ToolsResponse, error) {
	var tools []Tool
	for _, tool := range p.tools {
		// Filter by group
		if groupID != "" && tool.GroupID != groupID {
			continue
		}

		// Filter by search query
		if query != "" {
			if !matchesQuery(tool.Name, tool.Description, query) {
				continue
			}
		}

		tools = append(tools, *tool)
	}

	total := len(tools)

	// Apply pagination
	if offset >= len(tools) {
		tools = []Tool{}
	} else {
		end := offset + limit
		if limit == 0 || end > len(tools) {
			end = len(tools)
		}
		tools = tools[offset:end]
	}

	return &ToolsResponse{
		Tools:  tools,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}, nil
}

// ExecuteTool executes a registered tool.
func (p *SimpleProvider) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*ExecuteResponse, error) {
	executor, ok := p.executors[toolName]
	if !ok {
		return &ExecuteResponse{
			Error: &ErrorDetail{
				Code:    "tool_not_found",
				Message: "Tool not found: " + toolName,
			},
		}, nil
	}

	result, err := executor(ctx, params)
	if err != nil {
		return &ExecuteResponse{
			Error: &ErrorDetail{
				Code:    "execution_error",
				Message: err.Error(),
			},
		}, nil
	}

	return &ExecuteResponse{
		Result: result,
	}, nil
}

// GroupProviderImpl extends SimpleProvider with group support.
type GroupProviderImpl struct {
	*SimpleProvider
	groups map[string]*Group
}

// NewGroupProvider creates a provider with group support.
func NewGroupProvider(capabilities *Capabilities) *GroupProviderImpl {
	if capabilities == nil {
		capabilities = NewCapabilities()
	}
	capabilities.WithGroups("")

	return &GroupProviderImpl{
		SimpleProvider: NewSimpleProvider(capabilities),
		groups:         make(map[string]*Group),
	}
}

// RegisterGroup registers a group.
func (p *GroupProviderImpl) RegisterGroup(group *Group) {
	p.groups[group.ID] = group
}

// ListGroups returns all registered groups.
func (p *GroupProviderImpl) ListGroups(ctx context.Context, parentID, query string, offset, limit int) (*GroupsResponse, error) {
	var groups []Group
	for _, group := range p.groups {
		// Filter by parent
		if parentID != "" && group.ParentID != parentID {
			continue
		}

		// Filter by search query
		if query != "" {
			if !matchesQuery(group.Name, group.Description, query) {
				continue
			}
		}

		groups = append(groups, *group)
	}

	total := len(groups)

	// Apply pagination
	if offset >= len(groups) {
		groups = []Group{}
	} else {
		end := offset + limit
		if limit == 0 || end > len(groups) {
			end = len(groups)
		}
		groups = groups[offset:end]
	}

	return &GroupsResponse{
		Groups: groups,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}, nil
}

// GetGroup returns a specific group.
func (p *GroupProviderImpl) GetGroup(ctx context.Context, groupID string) (*Group, error) {
	group, ok := p.groups[groupID]
	if !ok {
		return nil, &ErrorDetail{
			Code:    "group_not_found",
			Message: "Group not found: " + groupID,
		}
	}
	return group, nil
}

// matchesQuery checks if a name or description matches a search query.
// This is a simple case-insensitive substring match.
func matchesQuery(name, description, query string) bool {
	query = strings.ToLower(query)
	name = strings.ToLower(name)
	description = strings.ToLower(description)

	return strings.Contains(name, query) || strings.Contains(description, query)
}
