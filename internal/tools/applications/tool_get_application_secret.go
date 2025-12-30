// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

var GetApplicationSecretDef = types.ToolDefinition{
	McpTool: &mcp.Tool{
		Name:        "get_application_secret",
		Title:       "Get PingOne Application Client Secret",
		Description: "Prompts the user to securely retrieve the client secret for an OIDC application from the PingOne console. The secret is NOT returned to the agent - this tool only facilitates the user accessing it directly. Use after creating or updating an application that requires a client secret.",
		InputSchema: schema.MustGenerateSchema[GetApplicationSecretInput](),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	},
}

type GetApplicationSecretInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	ApplicationId uuid.UUID `json:"applicationId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne application"`
}

// GetApplicationSecretHandler prompts the user to retrieve the application secret
func GetApplicationSecretHandler(initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetApplicationSecretInput,
) (
	*mcp.CallToolResult,
	any,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetApplicationSecretInput) (*mcp.CallToolResult, any, error) {
		ctx = initialize.InitializeToolInvocation(ctx, GetApplicationSecretDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetApplicationSecretDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		// retrieve the PINGONE_MCP_ENVIRONMENT_ID from env var
		mcpAdminEnvironmentID := os.Getenv("PINGONE_MCP_ENVIRONMENT_ID")
		if mcpAdminEnvironmentID == "" {
			return nil, nil, fmt.Errorf("admin environment ID not set in PINGONE_MCP_ENVIRONMENT_ID")
		}
		rootDomain := os.Getenv("PINGONE_ROOT_DOMAIN")
		if rootDomain == "" {
			return nil, nil, fmt.Errorf("root domain not set in PINGONE_ROOT_DOMAIN")
		}

		// Prompt user to retrieve the secret via elicitation
		var elicitErr error
		if req.Session != nil {
			elicitID := input.ApplicationId.String()
			_, elicitErr = req.Session.Elicit(
				ctx,
				&mcp.ElicitParams{
					Message:       "Securely retrieve the client secret from the PingOne console",
					URL:           fmt.Sprintf("https://console.%s/index.html?env=%s#/%s/applications/%s/CONFIGURATION", rootDomain, mcpAdminEnvironmentID, input.EnvironmentId.String(), input.ApplicationId.String()),
					ElicitationID: elicitID,
				},
			)
		}

		var resultText string
		if elicitErr != nil {
			resultText = "Secret retrieval elicitation failed. User must manually retrieve the client secret from the PingOne console. Agent cannot and should not access the secret. User must store it securely."
		} else {
			resultText = "Secret retrieval elicitation completed. The client secret is NOT returned in tool responses. User has been prompted to retrieve it directly from the PingOne console. Agent cannot and should not access the secret. User must store it securely."
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: resultText,
				},
			},
		}, nil, nil
	}
}
