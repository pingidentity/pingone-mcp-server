// Copyright © 2025 Ping Identity Corporation

package environments

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

var UpdateEnvironmentServicesByIdDef = types.ToolDefinition{
	IsReadOnly: false,
	McpTool: &mcp.Tool{
		Name:         "update_environment_services_by_id",
		Title:        "Update PingOne Environment Services by ID",
		Description:  "Update the services assigned to a PingOne environment (update's the environment's Bill of Materials) by the environment's unique ID.",
		InputSchema:  mustGenerateUpdateEnvironmentServicesByIdInputSchema(),
		OutputSchema: schema.MustGenerateSchema[UpdateEnvironmentServicesByIdOutput](),
	},
}

const NeoServiceValue = "NEO"

type UpdateEnvironmentServicesByIdInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	Services      []string  `json:"services" jsonschema:"REQUIRED. The product type values enabled for the environment. Note that 'NEO' represents both 'PING_ONE_VERIFY' and 'PING_ONE_CREDENTIALS' services."`
}

type UpdateEnvironmentServicesByIdOutput struct {
	Services pingone.EnvironmentBillOfMaterialsResponse `json:"services" jsonschema:"The updated bill of materials for the environment, including products and solution type"`
}

func mustGenerateUpdateEnvironmentServicesByIdInputSchema() *jsonschema.Schema {
	baseSchema := schema.MustGenerateSchema[UpdateEnvironmentServicesByIdInput]()

	if baseSchema.Properties == nil {
		panic("baseSchema.Properties is nil when generating UpdateEnvironmentServicesByIdInput schema")
	}

	// Add enum values to the services field
	servicesSchema, exists := baseSchema.Properties["services"]
	if !exists || servicesSchema == nil || servicesSchema.Items == nil {
		panic("services property not found in UpdateEnvironmentServicesByIdInput schema")
	}
	var itemsEnum []any
	for _, val := range pingone.AllowedEnvironmentBillOfMaterialsProductTypeEnumValues {
		itemsEnum = append(itemsEnum, string(val))
	}
	// Add Neo value, representing Verify and Credentials combined
	itemsEnum = append(itemsEnum, NeoServiceValue)
	servicesSchema.Items.Enum = itemsEnum

	return baseSchema
}

// UpdateEnvironmentServicesByIdHandler updates PingOne environment services by ID using the provided client
func UpdateEnvironmentServicesByIdHandler(environmentsClientFactory EnvironmentsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdateEnvironmentServicesByIdInput,
) (
	*mcp.CallToolResult,
	*UpdateEnvironmentServicesByIdOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateEnvironmentServicesByIdInput) (*mcp.CallToolResult, *UpdateEnvironmentServicesByIdOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, UpdateEnvironmentServicesByIdDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdateEnvironmentServicesByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := environmentsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdateEnvironmentServicesByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		// First, get current environment services to preserve existing configurations
		currentServices, httpResponse, err := client.GetEnvironmentServicesById(ctx, input.EnvironmentId)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if currentServices == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no services data in response from get"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		currentProductsByType := make(map[pingone.EnvironmentBillOfMaterialsProductType]pingone.EnvironmentBillOfMaterialsProduct)
		for _, product := range currentServices.Products {
			currentProductsByType[product.Type] = product
		}

		// Build list of desired EnvironmentBillOfMaterialsProductType values
		desiredProductTypes := make(map[pingone.EnvironmentBillOfMaterialsProductType]struct{})
		for _, service := range input.Services {
			if service == NeoServiceValue {
				// Expand Neo into its constituent services
				desiredProductTypes[pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_CREDENTIALS] = struct{}{}
				desiredProductTypes[pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_VERIFY] = struct{}{}
			} else {
				productType, err := pingone.NewEnvironmentBillOfMaterialsProductTypeFromValue(service)
				if err != nil {
					toolErr := errs.NewToolError(UpdateEnvironmentServicesByIdDef.McpTool.Name, fmt.Errorf("invalid service value: %s", service))
					errs.Log(ctx, toolErr)
					return nil, nil, toolErr
				}
				desiredProductTypes[*productType] = struct{}{}
			}
		}

		logger.FromContext(ctx).Debug("Updating environment services",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.Int("productCount", len(desiredProductTypes)))

		// Build request struct, preserving existing product configurations where they exist
		var replaceRequest pingone.EnvironmentBillOfMaterialsReplaceRequest
		for productType := range desiredProductTypes {
			if existingProduct, exists := currentProductsByType[productType]; exists {
				// Preserve the existing product configuration
				replaceRequest.Products = append(replaceRequest.Products, existingProduct)
			} else {
				// Create a new product with default configuration
				product := pingone.NewEnvironmentBillOfMaterialsProduct(productType)
				replaceRequest.Products = append(replaceRequest.Products, *product)
			}
		}

		// Call the API to update the environment services
		services, httpResponse, err := client.UpdateEnvironmentServicesById(ctx, input.EnvironmentId, &replaceRequest)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if services == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no services data in response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Environment services updated successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.Int("productCount", len(services.Products)))

		// Filter out _links field from response
		services.Links = nil

		result := &UpdateEnvironmentServicesByIdOutput{
			Services: *services,
		}

		return nil, result, nil
	}
}
