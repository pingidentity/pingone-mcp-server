package groups

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// getGroupSchema defines the input JSON schema for get_group
var getGroupSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "id":             { "type": "string" }
  },
  "required": ["environment_id", "id"]
}`)

type GetGroupTool struct {
   client      tools.PingOneClient
   environment string
}

// NewGetGroupTool constructs a new GetGroupTool
func NewGetGroupTool(client tools.PingOneClient, environment string) tools.Tool {
   return &GetGroupTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *GetGroupTool) Name() string {
   return "get_group"
}

// Description returns a human-readable description
func (t *GetGroupTool) Description() string {
   return "Retrieve a group via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *GetGroupTool) InputSchema() json.RawMessage {
   return getGroupSchema
}

// Run executes the tool logic
func (t *GetGroupTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   rawEnv, ok := args["environment_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: environment_id")
   }
   envID, ok := rawEnv.(string)
   if !ok || envID == "" {
       return nil, fmt.Errorf("environment_id must be a non-empty string")
   }
   rawID, ok := args["id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: id")
   }
   grpID, ok := rawID.(string)
   if !ok || grpID == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   result, err := t.client.GetGroup(ctx, envID, grpID)
   if err != nil {
       return nil, err
   }
   return result, nil
}