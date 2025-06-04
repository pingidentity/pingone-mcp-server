package environments

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// EnvironmentStatusUpdater defines the minimal interface to update an environment status
type EnvironmentStatusUpdater interface {
   UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error)
}

// updateEnvironmentStatusSchema defines the input JSON schema for update_environment_status
var updateEnvironmentStatusSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "id":     { "type": "string" },
    "status": { "type": "string", "enum": ["ACTIVE", "DELETE_PENDING"] }
  },
  "required": ["id", "status"]
}`)

// UpdateEnvironmentStatusTool implements Tool for updating an environment's status
type UpdateEnvironmentStatusTool struct {
   updater EnvironmentStatusUpdater
}

// NewUpdateEnvironmentStatusTool constructs a new UpdateEnvironmentStatusTool
// updater must implement UpdateEnvironmentStatus
func NewUpdateEnvironmentStatusTool(updater EnvironmentStatusUpdater) tools.Tool {
   return &UpdateEnvironmentStatusTool{updater: updater}
}

// Name returns the tool name
func (t *UpdateEnvironmentStatusTool) Name() string {
   return "update_environment_status"
}

// Description returns a human-readable description
func (t *UpdateEnvironmentStatusTool) Description() string {
   return "Update the status of an environment"
}

// InputSchema returns the JSON schema for tool arguments
func (t *UpdateEnvironmentStatusTool) InputSchema() json.RawMessage {
   return updateEnvironmentStatusSchema
}

// Run executes the tool logic
func (t *UpdateEnvironmentStatusTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   // Extract environment ID
   rawID, ok := args["id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: id")
   }
   id, ok := rawID.(string)
   if !ok || id == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   // Extract and validate status
   rawStatus, ok := args["status"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: status")
   }
   statusStr, ok := rawStatus.(string)
   if !ok || statusStr == "" {
       return nil, fmt.Errorf("status must be a non-empty string")
   }
   statusPtr, err := management.NewEnumEnvironmentStatusFromValue(statusStr)
   if err != nil {
       return nil, err
   }
   // Call client to update status
   updatedEnv, err := t.updater.UpdateEnvironmentStatus(ctx, id, *statusPtr)
   if err != nil {
       return nil, err
   }
   // Marshal response to generic map
   data, err := json.Marshal(updatedEnv)
   if err != nil {
       return nil, err
   }
   var result map[string]interface{}
   if err := json.Unmarshal(data, &result); err != nil {
       return nil, err
   }
   return result, nil
}