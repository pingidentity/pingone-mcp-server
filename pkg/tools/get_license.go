package tools

import (
   "context"
   "encoding/json"
   "fmt"
)

// getLicenseSchema defines the input JSON schema for get_license
var getLicenseSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "organization_id": { "type": "string" },
    "id":              { "type": "string" }
  },
  "required": ["organization_id", "id"],
  "additionalProperties": false
}`)

// GetLicenseTool implements Tool for retrieving a license
type GetLicenseTool struct {
   client PingOneClient
}

// NewGetLicenseTool constructs a new GetLicenseTool
func NewGetLicenseTool(client PingOneClient) Tool {
   return &GetLicenseTool{client: client}
}

// Name returns the tool name
func (t *GetLicenseTool) Name() string {
   return "get_license"
}

// Description returns a human-readable description
func (t *GetLicenseTool) Description() string {
   return "Retrieve a license via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *GetLicenseTool) InputSchema() json.RawMessage {
   return getLicenseSchema
}

// Run executes the tool logic
func (t *GetLicenseTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   rawOrg, ok := args["organization_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: organization_id")
   }
   orgID, ok := rawOrg.(string)
   if !ok || orgID == "" {
       return nil, fmt.Errorf("organization_id must be a non-empty string")
   }
   rawID, ok := args["id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: id")
   }
   licID, ok := rawID.(string)
   if !ok || licID == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   result, err := t.client.GetLicense(ctx, orgID, licID)
   if err != nil {
       return nil, err
   }
   return result, nil
}