package groups

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// updateGroupSchema defines the input JSON schema for update_group
var updateGroupSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "id":             { "type": "string" },
    "name":           { "type": "string" },
    "description":    { "type": "string" }
  },
  "required": ["environment_id", "id"]
}`)

type UpdateGroupTool struct {
   client      tools.PingOneClient
   environment string
}

// NewUpdateGroupTool constructs a new UpdateGroupTool
func NewUpdateGroupTool(client tools.PingOneClient, environment string) tools.Tool {
   return &UpdateGroupTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *UpdateGroupTool) Name() string {
   return "update_group"
}

// Description returns a human-readable description
func (t *UpdateGroupTool) Description() string {
   return "Update a group via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *UpdateGroupTool) InputSchema() json.RawMessage {
   return updateGroupSchema
}

// Run executes the tool logic
func (t *UpdateGroupTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
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
   grpID, ok := rawID.(string)
   if !ok || grpID == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   var grp management.Group
   if rawName, exists := args["name"]; exists {
       name, ok := rawName.(string)
       if !ok || name == "" {
           return nil, fmt.Errorf("name must be a non-empty string")
       }
       grp.Name = name
   }
   if rawDesc, exists := args["description"]; exists {
       desc, ok := rawDesc.(string)
       if !ok {
           return nil, fmt.Errorf("description must be a string")
       }
       grp.Description = &desc
   }
   updated, err := t.client.UpdateGroup(ctx, envID, grpID, grp)
   if err != nil {
       return nil, err
   }
   data, err := json.Marshal(updated)
   if err != nil {
       return nil, err
   }
   var result map[string]interface{}
   if err := json.Unmarshal(data, &result); err != nil {
       return nil, err
   }
   return result, nil
}