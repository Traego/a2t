package a2t

// NewMetaToolsAdded creates a meta response for dynamically discovered tools.
func NewMetaToolsAdded(tools ...Tool) *MetaResponse {
	return &MetaResponse{
		Type: "tools_added",
		Data: map[string]interface{}{
			"tools": tools,
		},
	}
}

// NewMetaGroupRefresh creates a meta response indicating groups need refreshing.
func NewMetaGroupRefresh(groupIDs ...string) *MetaResponse {
	return &MetaResponse{
		Type: "group_refresh",
		Data: map[string]interface{}{
			"group_ids": groupIDs,
		},
	}
}

// WithMeta adds a meta response to an ExecuteResponse.
func (r *ExecuteResponse) WithMeta(meta *MetaResponse) *ExecuteResponse {
	r.Meta = meta
	return r
}

// NewExecuteResponse creates a new execute response with the given result.
func NewExecuteResponse(result interface{}) *ExecuteResponse {
	return &ExecuteResponse{
		Result: result,
	}
}

// NewExecuteError creates a new execute response with an error.
func NewExecuteError(code, message string) *ExecuteResponse {
	return &ExecuteResponse{
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}
