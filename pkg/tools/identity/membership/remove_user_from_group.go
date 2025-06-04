package membership

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// removeUserFromGroupSchema defines the input JSON schema for remove_user_from_group
var removeUserFromGroupSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "user_id":        { "type": "string" },
    "group_id":       { "type": "string" }
  },
  "required": ["environment_id", "user_id", "group_id"]
}`)

type RemoveUserFromGroupTool struct {
   client      tools.PingOneClient
   environment string
}

// NewRemoveUserFromGroupTool constructs a new RemoveUserFromGroupTool
func NewRemoveUserFromGroupTool(client tools.PingOneClient, environment string) tools.Tool {
   return &RemoveUserFromGroupTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *RemoveUserFromGroupTool) Name() string {
   return "remove_user_from_group"
}

// Description returns a human-readable description
func (t *RemoveUserFromGroupTool) Description() string {
   return "Remove a user from a group via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *RemoveUserFromGroupTool) InputSchema() json.RawMessage {
   return removeUserFromGroupSchema
}

// Run executes the tool logic
func (t *RemoveUserFromGroupTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   // Extract environment_id
   rawEnv, ok := args["environment_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: environment_id")
   }
   envID, ok := rawEnv.(string)
   if !ok || envID == "" {
       return nil, fmt.Errorf("environment_id must be a non-empty string")
   }
   // Extract user_id
   rawUID, ok := args["user_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: user_id")
   }
   userID, ok := rawUID.(string)
   if !ok || userID == "" {
       return nil, fmt.Errorf("user_id must be a non-empty string")
   }
   // Extract group_id
   rawG, ok := args["group_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: group_id")
   }
   groupID, ok := rawG.(string)
   if !ok || groupID == "" {
       return nil, fmt.Errorf("group_id must be a non-empty string")
   }
   // Call PingOne API to remove user from group
   if err := t.client.RemoveUserFromGroup(ctx, envID, userID, groupID); err != nil {
       return nil, err
   }
   // Return success indicator
   return map[string]interface{}{"success": true}, nil
}