package users

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// resetUserPasswordSchema defines the input schema for admin password reset
var resetUserPasswordSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "id":             { "type": "string" },
    "password":       { "type": "string" }
  },
  "required": ["environment_id", "id", "password"]
}`)

type ResetUserPasswordTool struct {
   client      tools.PingOneClient
   environment string
}

// NewResetUserPasswordTool constructs a new ResetUserPasswordTool
func NewResetUserPasswordTool(client tools.PingOneClient, environment string) tools.Tool {
   return &ResetUserPasswordTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *ResetUserPasswordTool) Name() string {
   return "reset_user_password"
}

// Description returns a human-readable description
func (t *ResetUserPasswordTool) Description() string {
   return "Reset (admin) a user's password via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *ResetUserPasswordTool) InputSchema() json.RawMessage {
   return resetUserPasswordSchema
}

// Run executes the tool logic
func (t *ResetUserPasswordTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   // Extract environment_id
   rawEnv, ok := args["environment_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: environment_id")
   }
   envID, ok := rawEnv.(string)
   if !ok || envID == "" {
       return nil, fmt.Errorf("environment_id must be a non-empty string")
   }
   // Extract user id
   rawID, ok := args["id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: id")
   }
   userID, ok := rawID.(string)
   if !ok || userID == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   // Extract new password
   rawPwd, ok := args["password"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: password")
   }
   password, ok := rawPwd.(string)
   if !ok || password == "" {
       return nil, fmt.Errorf("password must be a non-empty string")
   }
   // Call PingOne API for reset
   if err := t.client.ResetUserPassword(ctx, envID, userID, password); err != nil {
       return nil, err
   }
   // Return success indicator
   return map[string]interface{}{"success": true}, nil
}