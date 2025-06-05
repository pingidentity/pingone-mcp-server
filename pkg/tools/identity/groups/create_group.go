package groups

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// createGroupSchema defines the input JSON schema for create_group
var createGroupSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "name":           { "type": "string" },
    "description":    { "type": "string" }
  },
  "required": ["environment_id", "name"]
}`)

type CreateGroupTool struct {
   client      tools.PingOneClient
   environment string
}

// NewCreateGroupTool constructs a new CreateGroupTool
func NewCreateGroupTool(client tools.PingOneClient, environment string) tools.Tool {
   return &CreateGroupTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *CreateGroupTool) Name() string {
   return "create_group"
}

// Description returns a human-readable description
func (t *CreateGroupTool) Description() string {
   return "Create a group via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *CreateGroupTool) InputSchema() json.RawMessage {
   return createGroupSchema
}

// Run executes the tool logic
func (t *CreateGroupTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   rawEnv, ok := args["environment_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: environment_id")
   }
   envID, ok := rawEnv.(string)
   if !ok || envID == "" {
       return nil, fmt.Errorf("environment_id must be a non-empty string")
   }
   rawName, ok := args["name"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: name")
   }
   name, ok := rawName.(string)
   if !ok || name == "" {
       return nil, fmt.Errorf("name must be a non-empty string")
   }
   var descPtr *string
   if rawDesc, ok := args["description"]; ok {
       if desc, ok2 := rawDesc.(string); ok2 {
           descPtr = &desc
       }
   }
   grp := management.NewGroup(name)
   if descPtr != nil {
       grp.Description = descPtr
   }
   created, err := t.client.CreateGroup(ctx, envID, *grp)
   if err != nil {
       return nil, err
   }
   data, err := json.Marshal(created)
   if err != nil {
       return nil, err
   }
   var result map[string]interface{}
   if err := json.Unmarshal(data, &result); err != nil {
       return nil, err
   }
   return result, nil
}