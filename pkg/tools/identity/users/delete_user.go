package users

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// deleteUserSchema defines the input JSON schema for delete_user
var deleteUserSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "id":             { "type": "string" }
  },
  "required": ["environment_id", "id"]
}`)

type DeleteUserTool struct {
   client      tools.PingOneClient
   environment string
}

// NewDeleteUserTool constructs a new DeleteUserTool
func NewDeleteUserTool(client tools.PingOneClient, environment string) tools.Tool {
   return &DeleteUserTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *DeleteUserTool) Name() string {
   return "delete_user"
}

// Description returns a human-readable description
func (t *DeleteUserTool) Description() string {
   return "Delete a user in PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *DeleteUserTool) InputSchema() json.RawMessage {
   return deleteUserSchema
}

// Run executes the tool logic
func (t *DeleteUserTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   // Extract target environment ID
   rawEnv, ok := args["environment_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: environment_id")
   }
   envID, ok := rawEnv.(string)
   if !ok || envID == "" {
       return nil, fmt.Errorf("environment_id must be a non-empty string")
   }
   // Extract user ID
   rawID, ok := args["id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: id")
   }
   userID, ok := rawID.(string)
   if !ok || userID == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   // Call PingOne API to delete the user
   if err := t.client.DeleteUser(ctx, envID, userID); err != nil {
       return nil, err
   }
   // Return success indicator
   return map[string]interface{}{"success": true}, nil
}