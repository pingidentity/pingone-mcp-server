package users

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// setUserEnabledSchema defines the input schema for enabling/disabling a user
var setUserEnabledSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "id":             { "type": "string" },
    "enabled":        { "type": "boolean" }
  },
  "required": ["environment_id", "id", "enabled"]
}`)

type SetUserEnabledTool struct {
   client      tools.PingOneClient
   environment string
}

// NewSetUserEnabledTool constructs a new SetUserEnabledTool
func NewSetUserEnabledTool(client tools.PingOneClient, environment string) tools.Tool {
   return &SetUserEnabledTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *SetUserEnabledTool) Name() string {
   return "set_user_enabled"
}

// Description returns a human-readable description
func (t *SetUserEnabledTool) Description() string {
   return "Enable or disable a user account in PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *SetUserEnabledTool) InputSchema() json.RawMessage {
   return setUserEnabledSchema
}

// Run executes the tool logic
func (t *SetUserEnabledTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
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
   // Extract enabled flag
   rawEnabled, ok := args["enabled"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: enabled")
   }
   enabled, ok := rawEnabled.(bool)
   if !ok {
       return nil, fmt.Errorf("enabled must be a boolean")
   }
   // Call PingOne API
   updated, err := t.client.UpdateUserEnabled(ctx, envID, userID, enabled)
   if err != nil {
       return nil, err
   }
   // Return full updated UserEnabled object
   buf, err := json.Marshal(updated)
   if err != nil {
       return nil, fmt.Errorf("failed to marshal updated state: %w", err)
   }
   var output map[string]interface{}
   if err := json.Unmarshal(buf, &output); err != nil {
       return nil, fmt.Errorf("failed to unmarshal updated state: %w", err)
   }
   return output, nil
}