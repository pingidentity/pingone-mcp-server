package membership

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// addUserToGroupSchema defines the input schema for adding a user to a group
var addUserToGroupSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "user_id":        { "type": "string" },
    "group_id":       { "type": "string" }
  },
  "required": ["environment_id", "user_id", "group_id"]
}`)

type AddUserToGroupTool struct {
   client      tools.PingOneClient
   environment string
}

// NewAddUserToGroupTool constructs a new AddUserToGroupTool
func NewAddUserToGroupTool(client tools.PingOneClient, environment string) tools.Tool {
   return &AddUserToGroupTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *AddUserToGroupTool) Name() string {
   return "add_user_to_group"
}

// Description returns a human-readable description
func (t *AddUserToGroupTool) Description() string {
   return "Add a user to a group via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *AddUserToGroupTool) InputSchema() json.RawMessage {
   return addUserToGroupSchema
}

// Run executes the tool logic
func (t *AddUserToGroupTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
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
   // Extract group id
   rawGroup, ok := args["group_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: group_id")
   }
   groupID, ok := rawGroup.(string)
   if !ok || groupID == "" {
       return nil, fmt.Errorf("group_id must be a non-empty string")
   }
   // Call PingOne API for membership add
   state, err := t.client.AddUserToGroup(ctx, envID, userID, groupID)
   if err != nil {
       return nil, err
   }
   return state, nil
}