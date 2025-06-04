package users

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// updateUserSchema defines the JSON schema for update_user input
var updateUserSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "id":             { "type": "string" },
    "username":       { "type": "string" },
    "email":          { "type": "string", "format": "email" }
  },
  "required": ["environment_id", "id"]
}`)

type UpdateUserTool struct {
   client      tools.PingOneClient
   environment string
}

// NewUpdateUserTool constructs a new UpdateUserTool
func NewUpdateUserTool(client tools.PingOneClient, environment string) tools.Tool {
   return &UpdateUserTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *UpdateUserTool) Name() string {
   return "update_user"
}

// Description returns a human-readable description
func (t *UpdateUserTool) Description() string {
   return "Update an existing user in PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *UpdateUserTool) InputSchema() json.RawMessage {
   return updateUserSchema
}

// Run executes the tool logic
func (t *UpdateUserTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
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
   // Build user model with updatable fields
   var user management.User
   if rawUsername, exists := args["username"]; exists {
       username, ok := rawUsername.(string)
       if !ok || username == "" {
           return nil, fmt.Errorf("username must be a non-empty string")
       }
       user.Username = username
   }
   if rawEmail, exists := args["email"]; exists {
       email, ok := rawEmail.(string)
       if !ok || email == "" {
           return nil, fmt.Errorf("email must be a non-empty string")
       }
       user.Email = email
   }
   // Call PingOne API to update the user
   updated, err := t.client.UpdateUser(ctx, envID, userID, user)
   if err != nil {
       return nil, err
   }
   // Return full updated user as map
   buf, err := json.Marshal(updated)
   if err != nil {
       return nil, fmt.Errorf("failed to marshal updated user: %w", err)
   }
   var output map[string]interface{}
   if err := json.Unmarshal(buf, &output); err != nil {
       return nil, fmt.Errorf("failed to unmarshal updated user: %w", err)
   }
   return output, nil
}