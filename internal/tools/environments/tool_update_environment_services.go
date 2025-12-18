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

var UpdateEnvironmentServicesDef = types.ToolDefinition{
	McpTool: &mcp.Tool{
		Name:         "update_environment_services",
		Title:        "Update PingOne Environment Services by ID",
		Description:  "Update the services assigned to a PingOne environment (update's the environment's Bill of Materials) by the environment's unique ID. IMPORTANT: when changing the services for an environment, include any optional fields you wish to retain from the existing configuration, as omitting them remove those fields from the configuration.",
		InputSchema:  mustGenerateUpdateEnvironmentServicesInputSchema(),
		OutputSchema: schema.MustGenerateSchema[UpdateEnvironmentServicesOutput](),
	},
}

const NeoServiceValue = "NEO"

type EnvironmentServiceInput struct {
	Type      string                                              `json:"type" jsonschema:"REQUIRED. The product type value. Note that 'NEO' represents both 'PING_ONE_VERIFY' and 'PING_ONE_CREDENTIALS' services."`
	Bookmarks []pingone.EnvironmentBillOfMaterialsProductBookmark `json:"bookmarks,omitempty" jsonschema:"OPTIONAL. Custom bookmarks. Up to five can be specified per product."`
	Console   *pingone.EnvironmentBillOfMaterialsProductConsole   `json:"console,omitempty" jsonschema:"OPTIONAL. Link to your administrative console for the product, whether the product is in the PingOne platform, PingCloud, a private cloud, or on-premises. If specified, must be an RFC 2396-compliant URI with a maximum length of 1024 characters."`
	Tags      []string                                            `json:"tags,omitempty" jsonschema:"OPTIONAL. The set of tags for the PingOne products to be initially configured. The currently supported value is DAVINCI_MINIMAL (only valid when the product type is PING_ONE_DAVINCI). This indicates that DaVinci is to be configured with a minimal set of resources."`
}

type UpdateEnvironmentServicesInput struct {
	EnvironmentId uuid.UUID                 `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	Services      []EnvironmentServiceInput `json:"services" jsonschema:"REQUIRED. The services enabled for the environment. Note that 'NEO' represents both 'PING_ONE_VERIFY' and 'PING_ONE_CREDENTIALS' services."`
}

type UpdateEnvironmentServicesOutput struct {
	Services pingone.EnvironmentBillOfMaterialsResponse `json:"services" jsonschema:"The updated bill of materials for the environment, including products and solution type"`
}

func mustGenerateUpdateEnvironmentServicesInputSchema() *jsonschema.Schema {
	baseSchema := schema.MustGenerateSchema[UpdateEnvironmentServicesInput]()

	if baseSchema.Properties == nil {
		panic("baseSchema.Properties is nil when generating UpdateEnvironmentServicesInput schema")
	}

	// Add enum values to the services field
	servicesSchema, exists := baseSchema.Properties["services"]
	if !exists || servicesSchema == nil || servicesSchema.Items == nil {
		panic("services property not found in UpdateEnvironmentServicesInput schema")
	}
	if servicesSchema.Items.Properties == nil || servicesSchema.Items.Properties["type"] == nil {
		panic("type property not found in services item schema for UpdateEnvironmentServicesInput")
	}
	var itemsEnum []any
	for _, val := range pingone.AllowedEnvironmentBillOfMaterialsProductTypeEnumValues {
		itemsEnum = append(itemsEnum, string(val))
	}
	// Add Neo value, representing Verify and Credentials combined
	itemsEnum = append(itemsEnum, NeoServiceValue)
	servicesSchema.Items.Properties["type"].Enum = itemsEnum

	return baseSchema
}

// Converts input struct into client SDK's EnvironmentBillOfMaterialsProduct struct
func (i EnvironmentServiceInput) toBOMProductWithType(productType pingone.EnvironmentBillOfMaterialsProductType) pingone.EnvironmentBillOfMaterialsProduct {
	return pingone.EnvironmentBillOfMaterialsProduct{
		Type:      productType,
		Bookmarks: i.Bookmarks,
		Console:   i.Console,
		Tags:      i.Tags,
	}
}

// UpdateEnvironmentServicesHandler updates PingOne environment services by ID using the provided client
func UpdateEnvironmentServicesHandler(environmentsClientFactory EnvironmentsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdateEnvironmentServicesInput,
) (
	*mcp.CallToolResult,
	*UpdateEnvironmentServicesOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateEnvironmentServicesInput) (*mcp.CallToolResult, *UpdateEnvironmentServicesOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, UpdateEnvironmentServicesDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdateEnvironmentServicesDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := environmentsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdateEnvironmentServicesDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		// First, get current environment services to preserve existing configurations
		currentServices, httpResponse, err := client.GetEnvironmentServices(ctx, input.EnvironmentId)
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
		inputProductsByType := make(map[pingone.EnvironmentBillOfMaterialsProductType]pingone.EnvironmentBillOfMaterialsProduct)
		for _, service := range input.Services {
			if service.Type == NeoServiceValue {
				// Expand Neo into its constituent services
				inputProductsByType[pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_CREDENTIALS] =
					service.toBOMProductWithType(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_CREDENTIALS)
				inputProductsByType[pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_VERIFY] =
					service.toBOMProductWithType(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_VERIFY)
			} else {
				productType, err := pingone.NewEnvironmentBillOfMaterialsProductTypeFromValue(service.Type)
				if err != nil {
					toolErr := errs.NewToolError(UpdateEnvironmentServicesDef.McpTool.Name, fmt.Errorf("invalid service value: %s", service.Type))
					errs.Log(ctx, toolErr)
					return nil, nil, toolErr
				}
				inputProductsByType[*productType] = service.toBOMProductWithType(*productType)
			}
		}

		logger.FromContext(ctx).Debug("Updating environment services",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.Int("productCount", len(input.Services)))

		// Build request struct, preserving existing product configurations where they exist
		var replaceRequest pingone.EnvironmentBillOfMaterialsReplaceRequest
		for productType := range inputProductsByType {
			if existingProduct, exists := currentProductsByType[productType]; exists {
				// Update any optional fields from input, but otherwise preserve existing configuration
				existingProduct.Bookmarks = inputProductsByType[productType].Bookmarks
				existingProduct.Console = inputProductsByType[productType].Console
				existingProduct.Tags = inputProductsByType[productType].Tags
				replaceRequest.Products = append(replaceRequest.Products, existingProduct)
			} else {
				// Create a new product entry
				replaceRequest.Products = append(replaceRequest.Products, inputProductsByType[productType])
			}
		}

		// Call the API to update the environment services
		services, httpResponse, err := client.UpdateEnvironmentServices(ctx, input.EnvironmentId, &replaceRequest)
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

		result := &UpdateEnvironmentServicesOutput{
			Services: *services,
		}

		return nil, result, nil
	}
}
