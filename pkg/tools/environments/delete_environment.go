package environments

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// deleteEnvironmentSchema defines the input JSON schema for delete_environment
var deleteEnvironmentSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "id": { "type": "string" }
  },
  "required": ["id"]
}`)

// EnvironmentDeleter defines the minimal interface to delete an environment
type EnvironmentDeleter interface {
   DeleteEnvironment(ctx context.Context, environmentID string) error
}

// DeleteEnvironmentTool implements Tool for deleting an environment
type DeleteEnvironmentTool struct {
   deleter EnvironmentDeleter
}

// NewDeleteEnvironmentTool constructs a new DeleteEnvironmentTool
// client must implement DeleteEnvironment
func NewDeleteEnvironmentTool(client EnvironmentDeleter) tools.Tool {
   return &DeleteEnvironmentTool{deleter: client}
}

// Name returns the tool name
func (t *DeleteEnvironmentTool) Name() string {
   return "delete_environment"
}

// Description returns a human-readable description
func (t *DeleteEnvironmentTool) Description() string {
   return "Delete an environment via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *DeleteEnvironmentTool) InputSchema() json.RawMessage {
   return deleteEnvironmentSchema
}

// Run executes the tool logic
func (t *DeleteEnvironmentTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   rawID, ok := args["id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: id")
   }
   id, ok := rawID.(string)
   if !ok || id == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   if err := t.deleter.DeleteEnvironment(ctx, id); err != nil {
       return nil, err
   }
   return map[string]interface{}{"success": true}, nil
}