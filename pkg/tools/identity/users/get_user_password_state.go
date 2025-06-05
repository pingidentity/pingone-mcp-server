package users

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// getUserPasswordStateSchema defines the input schema for reading a user's password state
var getUserPasswordStateSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "id":             { "type": "string" }
  },
  "required": ["environment_id", "id"]
}`)

type GetUserPasswordStateTool struct {
   client      tools.PingOneClient
   environment string
}

// NewGetUserPasswordStateTool constructs a new GetUserPasswordStateTool
func NewGetUserPasswordStateTool(client tools.PingOneClient, environment string) tools.Tool {
   return &GetUserPasswordStateTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *GetUserPasswordStateTool) Name() string {
   return "get_user_password_state"
}

// Description returns a human-readable description
func (t *GetUserPasswordStateTool) Description() string {
   return "Read a user's password state via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *GetUserPasswordStateTool) InputSchema() json.RawMessage {
   return getUserPasswordStateSchema
}

// Run executes the tool logic
func (t *GetUserPasswordStateTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
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
   // Call PingOne API for reading password state
   state, err := t.client.GetUserPasswordState(ctx, envID, userID)
   if err != nil {
       return nil, err
   }
   return state, nil
}