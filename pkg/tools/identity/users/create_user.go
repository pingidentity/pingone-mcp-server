package users

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

var createUserSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "username":       { "type": "string" },
    "email":          { "type": "string", "format": "email" },
    "environment_id": { "type": "string" }
  },
  "required": ["username", "email", "environment_id"]
}`)

type CreateUserTool struct {
   client      tools.PingOneClient
   environment string
}

// NewCreateUserTool constructs a new CreateUserTool
func NewCreateUserTool(client tools.PingOneClient, environment string) tools.Tool {
   return &CreateUserTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *CreateUserTool) Name() string {
   return "create_user"
}

// Description returns a human-readable description
func (t *CreateUserTool) Description() string {
   return "Create a new user in PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *CreateUserTool) InputSchema() json.RawMessage {
   return createUserSchema
}

// Run executes the tool logic
func (t *CreateUserTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   // Extract target environment ID
   rawEnv, ok := args["environment_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: environment_id")
   }
   envID, ok := rawEnv.(string)
   if !ok || envID == "" {
       return nil, fmt.Errorf("environment_id must be a non-empty string")
   }
   // Validate and extract username
   rawUsername, ok := args["username"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: username")
   }
   username, ok := rawUsername.(string)
   if !ok || username == "" {
       return nil, fmt.Errorf("username must be a non-empty string")
   }
   // Validate and extract email
   rawEmail, ok := args["email"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: email")
   }
   email, ok := rawEmail.(string)
   if !ok || email == "" {
       return nil, fmt.Errorf("email must be a non-empty string")
   }
   // Construct user model
   user := management.User{
       Username: username,
       Email:    email,
   }
   // Call PingOne API with target environment
   created, err := t.client.CreateUser(ctx, envID, user)
   if err != nil {
       return nil, err
   }
   // Build output map
   output := map[string]interface{}{}
   if created.Id != nil {
       output["id"] = *created.Id
   }
   output["username"] = created.Username
   output["email"] = created.Email
   return output, nil
}