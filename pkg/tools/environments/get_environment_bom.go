package environments

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// getEnvironmentBomSchema defines the input JSON schema for get_environment_bom
var getEnvironmentBomSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "id": { "type": "string" }
  },
  "required": ["id"],
  "additionalProperties": false
}`)

// GetEnvironmentBomTool retrieves the bill of materials for an environment
type GetEnvironmentBomTool struct {
   client tools.PingOneClient
}

// NewGetEnvironmentBomTool constructs a new GetEnvironmentBomTool
func NewGetEnvironmentBomTool(client tools.PingOneClient) tools.Tool {
   return &GetEnvironmentBomTool{client: client}
}

// Name returns the tool name
func (t *GetEnvironmentBomTool) Name() string {
   return "get_environment_bom"
}

// Description returns a human-readable description
func (t *GetEnvironmentBomTool) Description() string {
   return "Retrieve the bill of materials for an environment"
}

// InputSchema returns the JSON schema for tool arguments
func (t *GetEnvironmentBomTool) InputSchema() json.RawMessage {
   return getEnvironmentBomSchema
}

// Run executes the tool logic
func (t *GetEnvironmentBomTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   rawID, ok := args["id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: id")
   }
   id, ok := rawID.(string)
   if !ok || id == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   env, err := t.client.GetEnvironment(ctx, id)
   if err != nil {
       return nil, err
   }
   if env.BillOfMaterials == nil {
       return nil, fmt.Errorf("no bill of materials for environment %s", id)
   }
   data, err := json.Marshal(env.BillOfMaterials)
   if err != nil {
       return nil, fmt.Errorf("failed to marshal bill of materials: %w", err)
   }
   var bom map[string]interface{}
   if err := json.Unmarshal(data, &bom); err != nil {
       return nil, fmt.Errorf("failed to unmarshal bill of materials: %w", err)
   }
   return map[string]interface{}{"bill_of_materials": bom}, nil
}