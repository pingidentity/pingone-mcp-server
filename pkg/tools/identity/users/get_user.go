package users

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// getUserSchema defines the input JSON schema for get_user
var getUserSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "id":             { "type": "string" }
  },
  "required": ["environment_id", "id"]
}`)

type GetUserTool struct {
   client      tools.PingOneClient
   environment string
}

// NewGetUserTool constructs a new GetUserTool
func NewGetUserTool(client tools.PingOneClient, environment string) tools.Tool {
   return &GetUserTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *GetUserTool) Name() string {
   return "get_user"
}

// Description returns a human-readable description
func (t *GetUserTool) Description() string {
   return "Retrieve a user from PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *GetUserTool) InputSchema() json.RawMessage {
   return getUserSchema
}

// Run executes the tool logic
func (t *GetUserTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
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
   // Call PingOne API to get the user
   user, err := t.client.GetUser(ctx, envID, userID)
   if err != nil {
       return nil, err
   }
   // Serialize the full user object into a generic map to return all fields
   buf, err := json.Marshal(user)
   if err != nil {
       return nil, fmt.Errorf("failed to marshal user: %w", err)
   }
   var output map[string]interface{}
   if err := json.Unmarshal(buf, &output); err != nil {
       return nil, fmt.Errorf("failed to unmarshal user into map: %w", err)
   }
   return output, nil
}