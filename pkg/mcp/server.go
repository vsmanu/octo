package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcpsdk "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/manu/octo/pkg/config"
	"github.com/manu/octo/pkg/satellite"
)

// NewServer initializes and returns an MCP server configured for Octo
func NewServer(cfgMgr *config.Manager, satMgr *satellite.Manager) *server.MCPServer {
	srv := server.NewMCPServer(
		"octo-master",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Tool: get_config
	getConfigTool := mcpsdk.NewTool("get_config",
		mcpsdk.WithDescription("Returns the full configuration of the Octo master node"),
	)
	srv.AddTool(getConfigTool, func(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		cfg := cfgMgr.GetConfig()
		cfgJSON, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("failed to encode config: %v", err)), nil
		}
		return mcpsdk.NewToolResultText(string(cfgJSON)), nil
	})

	// Tool: list_endpoints
	listEndpointsTool := mcpsdk.NewTool("list_endpoints",
		mcpsdk.WithDescription("Returns the list of endpoints currently monitored by Octo"),
	)
	srv.AddTool(listEndpointsTool, func(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		cfg := cfgMgr.GetConfig()
		endpointsJSON, err := json.MarshalIndent(cfg.Endpoints, "", "  ")
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("failed to encode endpoints: %v", err)), nil
		}
		return mcpsdk.NewToolResultText(string(endpointsJSON)), nil
	})

	// Tool: list_satellites
	listSatellitesTool := mcpsdk.NewTool("list_satellites",
		mcpsdk.WithDescription("Returns the list of registered satellites and their current status"),
	)
	srv.AddTool(listSatellitesTool, func(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		sats := satMgr.GetAllSatellites()
		satsJSON, err := json.MarshalIndent(sats, "", "  ")
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("failed to encode satellites: %v", err)), nil
		}
		return mcpsdk.NewToolResultText(string(satsJSON)), nil
	})

	return srv
}
