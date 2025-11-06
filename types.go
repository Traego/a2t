package a2t

import (
	"encoding/json"
	"fmt"
)

// Tool represents a callable function that an AI agent can invoke.
// The schema matches standard LLM tool calling formats (OpenAI, Anthropic, etc.)
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
	GroupID     string                 `json:"group_id,omitempty"`
}

// Group organizes tools hierarchically.
type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id,omitempty"`
	ToolCount   int    `json:"tool_count"`
}

// Capabilities declares what features a server supports.
type Capabilities struct {
	Version   string               `json:"version"`
	Features  FeatureSet           `json:"features"`
	Endpoints EndpointConfig       `json:"endpoints"`
	Limits    *LimitsConfig        `json:"limits,omitempty"`
}

// FeatureSet defines which optional features are enabled.
type FeatureSet struct {
	Groups       bool `json:"groups"`
	Search       bool `json:"search"`
	DynamicTools bool `json:"dynamic_tools"`
}

// EndpointConfig defines the URL paths for each endpoint.
type EndpointConfig struct {
	Tools  string `json:"tools"`
	Groups string `json:"groups,omitempty"`
}

// LimitsConfig defines server-side limits.
type LimitsConfig struct {
	MaxToolsPerRequest   int `json:"max_tools_per_request,omitempty"`
	MaxGroupsPerRequest  int `json:"max_groups_per_request,omitempty"`
	MaxSearchResults     int `json:"max_search_results,omitempty"`
}

// ExecuteResponse is the response from tool execution.
type ExecuteResponse struct {
	Result interface{}   `json:"result"`
	Error  *ErrorDetail  `json:"error,omitempty"`
	Meta   *MetaResponse `json:"meta,omitempty"`
}

// ErrorDetail provides structured error information.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *ErrorDetail) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// MetaResponse contains metadata that clients can intercept.
type MetaResponse struct {
	Type interface{} `json:"type"`
	Data interface{} `json:"data"`
}

// MetaToolsAdded is metadata for dynamically discovered tools.
type MetaToolsAdded struct {
	Type  string `json:"type"` // "tools_added"
	Tools []Tool `json:"tools"`
}

// MetaGroupRefresh is metadata indicating groups need refreshing.
type MetaGroupRefresh struct {
	Type     string   `json:"type"` // "group_refresh"
	GroupIDs []string `json:"group_ids"`
}

// ToolsResponse is the response for listing tools.
type ToolsResponse struct {
	Tools  []Tool         `json:"tools"`
	Total  int            `json:"total,omitempty"`
	Offset int            `json:"offset,omitempty"`
	Limit  int            `json:"limit,omitempty"`
}

// GroupsResponse is the response for listing groups.
type GroupsResponse struct {
	Groups []Group `json:"groups"`
	Total  int     `json:"total,omitempty"`
	Offset int     `json:"offset,omitempty"`
	Limit  int     `json:"limit,omitempty"`
}

// NewCapabilities creates a basic capabilities configuration.
func NewCapabilities() *Capabilities {
	return &Capabilities{
		Version: "1.0",
		Features: FeatureSet{
			Groups:       false,
			Search:       false,
			DynamicTools: false,
		},
		Endpoints: EndpointConfig{
			Tools: "/tools",
		},
	}
}

// WithGroups enables groups feature.
func (c *Capabilities) WithGroups(endpoint string) *Capabilities {
	c.Features.Groups = true
	if endpoint == "" {
		endpoint = "/groups"
	}
	c.Endpoints.Groups = endpoint
	return c
}

// WithSearch enables search feature.
func (c *Capabilities) WithSearch() *Capabilities {
	c.Features.Search = true
	return c
}

// WithDynamicTools enables dynamic tool discovery.
func (c *Capabilities) WithDynamicTools() *Capabilities {
	c.Features.DynamicTools = true
	return c
}

// WithLimits sets server limits.
func (c *Capabilities) WithLimits(limits *LimitsConfig) *Capabilities {
	c.Limits = limits
	return c
}

// ToJSON serializes capabilities to JSON.
func (c *Capabilities) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// NewTool creates a new tool with the given name and description.
func NewTool(name, description string) *Tool {
	return &Tool{
		Name:        name,
		Description: description,
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
	}
}

// WithProperty adds a property to the tool's input schema.
func (t *Tool) WithProperty(name, propType, description string, required bool) *Tool {
	props := t.InputSchema["properties"].(map[string]interface{})
	props[name] = map[string]interface{}{
		"type":        propType,
		"description": description,
	}

	if required {
		reqs := t.InputSchema["required"].([]string)
		t.InputSchema["required"] = append(reqs, name)
	}

	return t
}

// WithGroup sets the group ID for the tool.
func (t *Tool) WithGroup(groupID string) *Tool {
	t.GroupID = groupID
	return t
}

// NewGroup creates a new group.
func NewGroup(id, name, description string) *Group {
	return &Group{
		ID:          id,
		Name:        name,
		Description: description,
	}
}

// WithParent sets the parent group ID.
func (g *Group) WithParent(parentID string) *Group {
	g.ParentID = parentID
	return g
}

// WithToolCount sets the number of tools in the group.
func (g *Group) WithToolCount(count int) *Group {
	g.ToolCount = count
	return g
}
