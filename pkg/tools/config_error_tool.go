package tools

import (
	"context"
	"encoding/json"
)

// ConfigurationErrorTool is displayed when PingOne configuration is missing
type ConfigurationErrorTool struct{}

func (t *ConfigurationErrorTool) Name() string {
	return "configuration_status"
}

func (t *ConfigurationErrorTool) Description() string {
	return "Check PingOne MCP server configuration status"
}

func (t *ConfigurationErrorTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {},
		"required": []
	}`)
}

func (t *ConfigurationErrorTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"status": "error",
		"message": "PingOne configuration is missing or invalid. Please set the following environment variables:",
		"required_variables": []string{
			"PINGONE_CLIENT_ID",
			"PINGONE_CLIENT_SECRET", 
			"PINGONE_ENV_ID",
		},
		"optional_variables": []string{
			"PINGONE_REGION (defaults to 'com')",
			"PINGONE_MCP_ALLOW_MUTATION (set to 'true' to enable write operations)",
		},
		"instructions": "Set these environment variables and restart the server to access PingOne tools.",
	}, nil
}