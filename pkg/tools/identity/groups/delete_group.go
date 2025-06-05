package groups

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// deleteGroupSchema defines the input JSON schema for delete_group
var deleteGroupSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "id":             { "type": "string" }
  },
  "required": ["environment_id", "id"]
}`)

type DeleteGroupTool struct {
   client      tools.PingOneClient
   environment string
}

// NewDeleteGroupTool constructs a new DeleteGroupTool
func NewDeleteGroupTool(client tools.PingOneClient, environment string) tools.Tool {
   return &DeleteGroupTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *DeleteGroupTool) Name() string {
   return "delete_group"
}

// Description returns a human-readable description
func (t *DeleteGroupTool) Description() string {
   return "Delete a group via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *DeleteGroupTool) InputSchema() json.RawMessage {
   return deleteGroupSchema
}

// Run executes the tool logic
func (t *DeleteGroupTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
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
   groupID, ok := rawID.(string)
   if !ok || groupID == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   if err := t.client.DeleteGroup(ctx, envID, groupID); err != nil {
       return nil, err
   }
   return map[string]interface{}{"success": true}, nil
}