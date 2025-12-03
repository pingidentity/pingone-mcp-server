// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
)

type ReadApplicationModel struct {
	ApplicationExternalLink        *management.ApplicationExternalLink        `json:"applicationExternalLink,omitempty"`
	ApplicationOIDC                *management.ApplicationOIDC                `json:"applicationOIDC,omitempty"`
	ApplicationPingOneAdminConsole *management.ApplicationPingOneAdminConsole `json:"applicationPingOneAdminConsole,omitempty"`
	ApplicationPingOnePortal       *management.ApplicationPingOnePortal       `json:"applicationPingOnePortal,omitempty"`
	ApplicationPingOneSelfService  *management.ApplicationPingOneSelfService  `json:"applicationPingOneSelfService,omitempty"`
	ApplicationSAML                *management.ApplicationSAML                `json:"applicationSAML,omitempty"`
	ApplicationWSFED               *management.ApplicationWSFED               `json:"applicationWSFED,omitempty"`
}

type UpdateApplicationModel struct {
	ApplicationExternalLink       *management.ApplicationExternalLink       `json:"applicationExternalLink,omitempty"`
	ApplicationOIDC               *management.ApplicationOIDC               `json:"applicationOIDC,omitempty"`
	ApplicationPingOnePortal      *management.ApplicationPingOnePortal      `json:"applicationPingOnePortal,omitempty"`
	ApplicationPingOneSelfService *management.ApplicationPingOneSelfService `json:"applicationPingOneSelfService,omitempty"`
	ApplicationSAML               *management.ApplicationSAML               `json:"applicationSAML,omitempty"`
	ApplicationWSFED              *management.ApplicationWSFED              `json:"applicationWSFED,omitempty"`
}

type OIDCApplicationModel struct {
	ApplicationOIDC *management.ApplicationOIDC `json:"applicationOIDC,omitempty"`
}

// filterApplicationLinks removes _links field from all application types
func filterApplicationLinks(model ReadApplicationModel) ReadApplicationModel {
	if model.ApplicationExternalLink != nil {
		model.ApplicationExternalLink.Links = nil
	}
	if model.ApplicationOIDC != nil {
		model.ApplicationOIDC.Links = nil
	}
	// ApplicationPingOneAdminConsole doesn't have Links field
	if model.ApplicationPingOnePortal != nil {
		model.ApplicationPingOnePortal.Links = nil
	}
	if model.ApplicationPingOneSelfService != nil {
		model.ApplicationPingOneSelfService.Links = nil
	}
	if model.ApplicationSAML != nil {
		model.ApplicationSAML.Links = nil
	}
	if model.ApplicationWSFED != nil {
		model.ApplicationWSFED.Links = nil
	}
	return model
}

func ReadApplicationModelFromSDKReadResponse(sdkApp management.ReadOneApplication200Response) ReadApplicationModel {
	model := ReadApplicationModel{
		ApplicationExternalLink:        sdkApp.ApplicationExternalLink,
		ApplicationOIDC:                sdkApp.ApplicationOIDC,
		ApplicationPingOneAdminConsole: sdkApp.ApplicationPingOneAdminConsole,
		ApplicationPingOnePortal:       sdkApp.ApplicationPingOnePortal,
		ApplicationPingOneSelfService:  sdkApp.ApplicationPingOneSelfService,
		ApplicationSAML:                sdkApp.ApplicationSAML,
		ApplicationWSFED:               sdkApp.ApplicationWSFED,
	}
	return filterApplicationLinks(model)
}

func UpdateApplicationModelFromSDKReadResponse(sdkApp management.ReadOneApplication200Response) UpdateApplicationModel {
	model := UpdateApplicationModel{
		ApplicationExternalLink:       sdkApp.ApplicationExternalLink,
		ApplicationOIDC:               sdkApp.ApplicationOIDC,
		ApplicationPingOnePortal:      sdkApp.ApplicationPingOnePortal,
		ApplicationPingOneSelfService: sdkApp.ApplicationPingOneSelfService,
		ApplicationSAML:               sdkApp.ApplicationSAML,
		ApplicationWSFED:              sdkApp.ApplicationWSFED,
	}
	// Filter out _links field from all application types
	if model.ApplicationExternalLink != nil {
		model.ApplicationExternalLink.Links = nil
	}
	if model.ApplicationOIDC != nil {
		model.ApplicationOIDC.Links = nil
	}
	if model.ApplicationPingOnePortal != nil {
		model.ApplicationPingOnePortal.Links = nil
	}
	if model.ApplicationPingOneSelfService != nil {
		model.ApplicationPingOneSelfService.Links = nil
	}
	if model.ApplicationSAML != nil {
		model.ApplicationSAML.Links = nil
	}
	if model.ApplicationWSFED != nil {
		model.ApplicationWSFED.Links = nil
	}
	return model
}

func UpdateApplicationModelToSDKUpdateRequest(model UpdateApplicationModel) management.UpdateApplicationRequest {
	return management.UpdateApplicationRequest{
		ApplicationExternalLink:       model.ApplicationExternalLink,
		ApplicationOIDC:               model.ApplicationOIDC,
		ApplicationPingOnePortal:      model.ApplicationPingOnePortal,
		ApplicationPingOneSelfService: model.ApplicationPingOneSelfService,
		ApplicationSAML:               model.ApplicationSAML,
		ApplicationWSFED:              model.ApplicationWSFED,
	}
}

const getApplicationOperation = "get"
const putApplicationOperation = "put"
const postApplicationOperation = "post"

// Creates a JSON schema for an application model with oneOf constraint
func MustGenerateApplicationModelSchema(operation string) *jsonschema.Schema {
	externalLinkSchema := schema.MustGenerateSchema[management.ApplicationExternalLink]()
	oidcSchema := schema.MustGenerateSchema[management.ApplicationOIDC]()
	samlSchema := schema.MustGenerateSchema[management.ApplicationSAML]()
	wsfedSchema := schema.MustGenerateSchema[management.ApplicationWSFED]()

	props := map[string]*jsonschema.Schema{
		"applicationExternalLink": externalLinkSchema,
		"applicationOIDC":         oidcSchema,
		"applicationSAML":         samlSchema,
		"applicationWSFED":        wsfedSchema,
	}

	// Require the schema to match exactly one of the application types
	oneOf := []*jsonschema.Schema{
		{Required: []string{"applicationExternalLink"}},
		{Required: []string{"applicationOIDC"}},
		{Required: []string{"applicationSAML"}},
		{Required: []string{"applicationWSFED"}},
	}

	if operation == getApplicationOperation || operation == putApplicationOperation {
		// Add P1 apps that can be read/updated
		portalSchema := schema.MustGenerateSchema[management.ApplicationPingOnePortal]()
		selfServiceSchema := schema.MustGenerateSchema[management.ApplicationPingOneSelfService]()

		props["applicationPingOnePortal"] = portalSchema
		props["applicationPingOneSelfService"] = selfServiceSchema

		oneOf = append(oneOf,
			&jsonschema.Schema{Required: []string{"applicationPingOnePortal"}},
			&jsonschema.Schema{Required: []string{"applicationPingOneSelfService"}},
		)
	}

	if operation == getApplicationOperation {
		// Add P1 Admin Console app that can only be read
		adminConsoleSchema := schema.MustGenerateSchema[management.ApplicationPingOneAdminConsole]()
		props["applicationPingOneAdminConsole"] = adminConsoleSchema
		oneOf = append(oneOf,
			&jsonschema.Schema{Required: []string{"applicationPingOneAdminConsole"}},
		)
	}

	return &jsonschema.Schema{
		Type:       "object",
		Properties: props,
		OneOf:      oneOf,
	}
}

// Creates a JSON schema for ReadApplicationModel with oneOf constraint
func MustGenerateReadApplicationModelSchema() *jsonschema.Schema {
	return MustGenerateApplicationModelSchema(getApplicationOperation)
}

// Creates a JSON schema for CreateApplicationModel with oneOf constraint
func MustGenerateCreateApplicationModelSchema() *jsonschema.Schema {
	return MustGenerateApplicationModelSchema(postApplicationOperation)
}

// Creates a JSON schema for UpdateApplicationModel with oneOf constraint
func MustGenerateUpdateApplicationModelSchema() *jsonschema.Schema {
	return MustGenerateApplicationModelSchema(putApplicationOperation)
}
