package a2t

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v5emb"
	"github.com/swaggest/usecase"
)

// Server is an HTTP server that exposes a ToolProvider with OpenAPI documentation.
type Server struct {
	provider ToolProvider
	service  *web.Service
}

// ListToolsInput represents input for listing tools.
type ListToolsInput struct {
	Q      string `query:"q" description:"Search query to filter tools by name or description"`
	Offset int    `query:"offset" description:"Pagination offset"`
	Limit  int    `query:"limit" description:"Maximum number of tools to return" default:"100"`
}

// ListGroupsInput represents input for listing groups.
type ListGroupsInput struct {
	Q        string `query:"q" description:"Search query to filter groups by name or description"`
	ParentID string `query:"parent_id" description:"Filter groups by parent ID"`
	Offset   int    `query:"offset" description:"Pagination offset"`
	Limit    int    `query:"limit" description:"Maximum number of groups to return" default:"50"`
}

// ExecuteToolInput represents input for executing a tool.
type ExecuteToolInput struct {
	Name   string                 `path:"name" description:"Tool name"`
	Params map[string]interface{} `json:"-"` // Body
}

// ListGroupToolsInput represents input for listing tools in a group.
type ListGroupToolsInput struct {
	ID     string `path:"id" description:"Group ID"`
	Q      string `query:"q" description:"Search query to filter tools by name or description"`
	Offset int    `query:"offset" description:"Pagination offset"`
	Limit  int    `query:"limit" description:"Maximum number of tools to return" default:"100"`
}

// ExecuteGroupToolInput represents input for executing a tool in a group.
type ExecuteGroupToolInput struct {
	ID     string                 `path:"id" description:"Group ID"`
	Name   string                 `path:"name" description:"Tool name"`
	Params map[string]interface{} `json:"-"` // Body
}

// NewServer creates a new a2t HTTP server with OpenAPI documentation.
func NewServer(provider ToolProvider) *Server {
	service := web.NewService(openapi3.NewReflector())

	// Set API information
	service.OpenAPISchema().SetTitle("a2t - Agent-to-Tool Protocol")
	service.OpenAPISchema().SetDescription("A simple, stateless protocol for tool calling between AI agents and tool providers")
	service.OpenAPISchema().SetVersion("1.0.0")

	s := &Server{
		provider: provider,
		service:  service,
	}

	// Register routes
	s.registerRoutes()

	return s
}

func (s *Server) registerRoutes() {
	caps := s.provider.GetCapabilities()

	// Well-known capabilities endpoint
	s.service.Get("/.well-known/a2t-capabilities.json", s.capabilitiesUsecase())

	// Tools endpoints
	s.service.Get(caps.Endpoints.Tools, s.listToolsUsecase())
	s.service.Post(caps.Endpoints.Tools+"/{name}", s.executeToolUsecase())

	// Group endpoints (if enabled)
	if caps.Features.Groups {
		s.service.Get(caps.Endpoints.Groups, s.listGroupsUsecase())
		s.service.Get(caps.Endpoints.Groups+"/{id}/tools", s.listGroupToolsUsecase())
		s.service.Post(caps.Endpoints.Groups+"/{id}/tools/{name}", s.executeGroupToolUsecase())
	}

	// Swagger UI endpoint
	s.service.Docs("/docs", swgui.New)
}

// capabilitiesUsecase returns the server's capabilities.
func (s *Server) capabilitiesUsecase() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input struct{}, output *Capabilities) error {
		caps := s.provider.GetCapabilities()
		*output = *caps
		return nil
	})

	u.SetTags("Capabilities")
	u.SetTitle("Get Capabilities")
	u.SetDescription("Returns the server's capabilities including supported features and endpoints")

	return u
}

// listToolsUsecase lists all available tools.
func (s *Server) listToolsUsecase() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input ListToolsInput, output *ToolsResponse) error {
		limit := input.Limit
		if limit == 0 {
			limit = 100
		}

		resp, err := s.provider.ListTools(ctx, "", input.Q, input.Offset, limit)
		if err != nil {
			return err
		}

		*output = *resp
		return nil
	})

	u.SetTags("Tools")
	u.SetTitle("List Tools")
	u.SetDescription("Returns all available tools with optional search filtering")

	return u
}

// executeToolUsecase executes a specific tool.
func (s *Server) executeToolUsecase() usecase.Interactor {
	type input struct {
		ExecuteToolInput
		Body map[string]interface{}
	}

	u := usecase.NewInteractor(func(ctx context.Context, in input, output *ExecuteResponse) error {
		// Use Body from the request
		params := in.Body
		if params == nil {
			params = make(map[string]interface{})
		}

		resp, err := s.provider.ExecuteTool(ctx, in.Name, params)
		if err != nil {
			return err
		}

		*output = *resp
		return nil
	})

	u.SetTags("Tools")
	u.SetTitle("Execute Tool")
	u.SetDescription("Executes a specific tool with the provided parameters")

	return u
}

// listGroupsUsecase lists all available groups.
func (s *Server) listGroupsUsecase() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input ListGroupsInput, output *GroupsResponse) error {
		groupProvider, ok := s.provider.(GroupProvider)
		if !ok {
			return fmt.Errorf("groups not supported")
		}

		limit := input.Limit
		if limit == 0 {
			limit = 50
		}

		resp, err := groupProvider.ListGroups(ctx, input.ParentID, input.Q, input.Offset, limit)
		if err != nil {
			return err
		}

		*output = *resp
		return nil
	})

	u.SetTags("Groups")
	u.SetTitle("List Groups")
	u.SetDescription("Returns all available groups with optional search filtering")

	return u
}

// listGroupToolsUsecase lists tools in a specific group.
func (s *Server) listGroupToolsUsecase() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input ListGroupToolsInput, output *ToolsResponse) error {
		groupProvider, ok := s.provider.(GroupProvider)
		if !ok {
			return fmt.Errorf("groups not supported")
		}

		limit := input.Limit
		if limit == 0 {
			limit = 100
		}

		resp, err := groupProvider.ListTools(ctx, input.ID, input.Q, input.Offset, limit)
		if err != nil {
			return err
		}

		*output = *resp
		return nil
	})

	u.SetTags("Groups", "Tools")
	u.SetTitle("List Group Tools")
	u.SetDescription("Returns all tools in a specific group with optional search filtering")

	return u
}

// executeGroupToolUsecase executes a tool within a specific group.
func (s *Server) executeGroupToolUsecase() usecase.Interactor {
	type input struct {
		ExecuteGroupToolInput
		Body map[string]interface{}
	}

	u := usecase.NewInteractor(func(ctx context.Context, in input, output *ExecuteResponse) error {
		groupProvider, ok := s.provider.(GroupProvider)
		if !ok {
			return fmt.Errorf("groups not supported")
		}

		// Use Body from the request
		params := in.Body
		if params == nil {
			params = make(map[string]interface{})
		}

		resp, err := groupProvider.ExecuteTool(ctx, in.Name, params)
		if err != nil {
			return err
		}

		*output = *resp
		return nil
	})

	u.SetTags("Groups", "Tools")
	u.SetTitle("Execute Group Tool")
	u.SetDescription("Executes a specific tool within a group context")

	return u
}

// Handler returns the http.Handler for the server.
func (s *Server) Handler() http.Handler {
	return s.service
}

// ListenAndServe starts the server on the specified address.
func (s *Server) ListenAndServe(addr string) error {
	// Parse host from addr
	host := addr
	if strings.HasPrefix(addr, ":") {
		host = "localhost" + addr
	}

	fmt.Printf("a2t server listening on %s\n", addr)
	fmt.Printf("Capabilities: http://%s/.well-known/a2t-capabilities.json\n", host)
	fmt.Printf("OpenAPI JSON: http://%s/docs/openapi.json\n", host)
	fmt.Printf("Swagger UI: http://%s/docs\n", host)
	fmt.Println()

	return http.ListenAndServe(addr, s.service)
}
