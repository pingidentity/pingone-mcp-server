package users

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// unlockUserPasswordSchema defines the input schema for unlocking a user's password
var unlockUserPasswordSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "id":             { "type": "string" }
  },
  "required": ["environment_id", "id"]
}`)

type UnlockUserPasswordTool struct {
   client      tools.PingOneClient
   environment string
}

// NewUnlockUserPasswordTool constructs a new UnlockUserPasswordTool
func NewUnlockUserPasswordTool(client tools.PingOneClient, environment string) tools.Tool {
   return &UnlockUserPasswordTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *UnlockUserPasswordTool) Name() string {
   return "unlock_user_password"
}

// Description returns a human-readable description
func (t *UnlockUserPasswordTool) Description() string {
   return "Unlock (recover) a user's password via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *UnlockUserPasswordTool) InputSchema() json.RawMessage {
   return unlockUserPasswordSchema
}

// Run executes the tool logic
func (t *UnlockUserPasswordTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
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
   userID, ok := rawID.(string)
   if !ok || userID == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   if err := t.client.UnlockUserPassword(ctx, envID, userID); err != nil {
       return nil, err
   }
   return map[string]interface{}{"success": true}, nil
}