// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/go-openapi/jsonpointer"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

var PatchApplicationByIdDef = types.ToolDefinition{
	IsReadOnly: false,
	McpTool: &mcp.Tool{
		Name:        "patch_application_by_id",
		Title:       "Patch PingOne Application by ID",
		Description: "Patch an existing application within a specified PingOne environment.",
		InputSchema: schema.MustGenerateSchema[PatchApplicationByIdInput](),
		//TODO change this probably
		OutputSchema: mustGenerateUpdateApplicationByIdSchema[PatchApplicationByIdOutput](),
	},
}

const OperationUpdate = "UPDATE"
const OperationDelete = "DELETE"

type Patch struct {
	Operation string      `json:"operation" jsonschema:"REQUIRED. The operation to perform: UPDATE or DELETE"`
	Path      string      `json:"path" jsonschema:"REQUIRED. The JSON Pointer path (RFC 6901) to the field to update or delete"`
	Value     interface{} `json:"value,omitempty" jsonschema:"The value to set for UPDATE operations"`
}

type PatchApplicationByIdInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	ApplicationId uuid.UUID `json:"applicationId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne application"`
	Patches       []Patch   `json:"patches" jsonschema:"REQUIRED. The list of patch operations to apply to the application"`
}

type PatchApplicationByIdOutput struct {
	Application UpdateApplicationModel `json:"application" jsonschema:"The updated application configuration details"`
}

func PatchApplicationByIdHandler(applicationsClientFactory ApplicationsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input PatchApplicationByIdInput,
) (
	*mcp.CallToolResult,
	*PatchApplicationByIdOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input PatchApplicationByIdInput) (*mcp.CallToolResult, *PatchApplicationByIdOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, PatchApplicationByIdDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		//TODO validate patches

		client, err := applicationsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Patching application",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("applicationId", input.ApplicationId.String()),
		)

		// Get current application config
		currentApplication, httpResponse, err := client.GetApplication(ctx, input.EnvironmentId, input.ApplicationId)
		logger.LogHttpResponse(ctx, httpResponse)
		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}
		if currentApplication == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no application data in get response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		// Find the specific application type in the current config
		var currentAppJson any
		switch {
		case currentApplication.ApplicationExternalLink != nil:
			currentAppJson = currentApplication.ApplicationExternalLink
		case currentApplication.ApplicationOIDC != nil:
			currentAppJson = currentApplication.ApplicationOIDC
		case currentApplication.ApplicationPingOnePortal != nil:
			currentAppJson = currentApplication.ApplicationPingOnePortal
		case currentApplication.ApplicationPingOneSelfService != nil:
			currentAppJson = currentApplication.ApplicationPingOneSelfService
		case currentApplication.ApplicationSAML != nil:
			currentAppJson = currentApplication.ApplicationSAML
		case currentApplication.ApplicationWSFED != nil:
			currentAppJson = currentApplication.ApplicationWSFED
		default:
			toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, fmt.Errorf("unsupported application type for application ID %s", input.ApplicationId.String()))
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		// Apply patches to json based on current config

		//TODO this won't handle if the LLM tries to set a field that doesn't exist in our model
		patchedJsonDoc, err := applyPatchesToJSON(currentAppJson, input.Patches)
		if err != nil {
			toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, fmt.Errorf("failed to apply patches to application json: %w", err))
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}
		//TODO is this necessary, or can I just cast it
		patchedJson, err := json.Marshal(patchedJsonDoc)
		if err != nil {
			toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, fmt.Errorf("failed to marshal patched application json: %w", err))
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		// TODO convert patched json to UpdateApplicationModel
		var updatedModel UpdateApplicationModel
		switch {
		case currentApplication.ApplicationExternalLink != nil:
			var updatedApp management.ApplicationExternalLink
			err = json.Unmarshal(patchedJson, &updatedApp)
			if err != nil {
				toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, fmt.Errorf("failed to unmarshal patched application json to external link type: %w", err))
				errs.Log(ctx, toolErr)
				return nil, nil, toolErr
			}
			updatedModel.ApplicationExternalLink = &updatedApp
		case currentApplication.ApplicationOIDC != nil:
			var updatedApp management.ApplicationOIDC
			err = json.Unmarshal(patchedJson, &updatedApp)
			if err != nil {
				toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, fmt.Errorf("failed to unmarshal patched application json to OIDC type: %w", err))
				errs.Log(ctx, toolErr)
				return nil, nil, toolErr
			}
			updatedModel.ApplicationOIDC = &updatedApp
		case currentApplication.ApplicationPingOnePortal != nil:
			var updatedApp management.ApplicationPingOnePortal
			err = json.Unmarshal(patchedJson, &updatedApp)
			if err != nil {
				toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, fmt.Errorf("failed to unmarshal patched application json to PingOne Portal type: %w", err))
				errs.Log(ctx, toolErr)
				return nil, nil, toolErr
			}
			updatedModel.ApplicationPingOnePortal = &updatedApp
		case currentApplication.ApplicationPingOneSelfService != nil:
			var updatedApp management.ApplicationPingOneSelfService
			err = json.Unmarshal(patchedJson, &updatedApp)
			if err != nil {
				toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, fmt.Errorf("failed to unmarshal patched application json to PingOne Self-Service type: %w", err))
				errs.Log(ctx, toolErr)
				return nil, nil, toolErr
			}
			updatedModel.ApplicationPingOneSelfService = &updatedApp
		case currentApplication.ApplicationSAML != nil:
			var updatedApp management.ApplicationSAML
			err = json.Unmarshal(patchedJson, &updatedApp)
			if err != nil {
				toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, fmt.Errorf("failed to unmarshal patched application json to SAML type: %w", err))
				errs.Log(ctx, toolErr)
				return nil, nil, toolErr
			}
			updatedModel.ApplicationSAML = &updatedApp
		case currentApplication.ApplicationWSFED != nil:
			var updatedApp management.ApplicationWSFED
			err = json.Unmarshal(patchedJson, &updatedApp)
			if err != nil {
				toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, fmt.Errorf("failed to unmarshal patched application json to WS-FED type: %w", err))
				errs.Log(ctx, toolErr)
				return nil, nil, toolErr
			}
			updatedModel.ApplicationWSFED = &updatedApp
		default:
			toolErr := errs.NewToolError(PatchApplicationByIdDef.McpTool.Name, fmt.Errorf("unsupported application type for application ID %s during patching", input.ApplicationId.String()))
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		updateRequest := UpdateApplicationModelToSDKUpdateRequest(updatedModel)

		// Call the API to update the application
		applicationResponse, httpResponse, err := client.UpdateApplicationById(ctx, input.EnvironmentId, input.ApplicationId, updateRequest)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}
		if applicationResponse == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no application data in put response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Application patched successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("applicationId", input.ApplicationId.String()),
		)

		result := &PatchApplicationByIdOutput{
			Application: UpdateApplicationModelFromSDKReadResponse(*applicationResponse),
		}

		return nil, result, nil
	}
}

func applyPatchesToJSON(baseDocument any, patches []Patch) (any, error) {
	for _, patch := range patches {
		pointer, err := jsonpointer.New(patch.Path)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON Pointer path '%s': %w", patch.Path, err)
		}
		switch patch.Operation {
		case OperationUpdate:
			baseDocument, err = pointer.Set(baseDocument, patch.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to apply UPDATE patch at path '%s': %w", patch.Path, err)
			}
		case OperationDelete:
			baseDocument, err = pointer.Set(baseDocument, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to apply DELETE patch at path '%s': %w", patch.Path, err)
			}
		default:
			return nil, fmt.Errorf("unsupported patch operation '%s' at path '%s'", patch.Operation, patch.Path)
		}
	}
	return baseDocument, nil
}
