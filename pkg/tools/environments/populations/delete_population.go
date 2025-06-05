package populations

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// deletePopulationSchema defines the input JSON schema for delete_population
var deletePopulationSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id":  { "type": "string" },
    "id":              { "type": "string" }
  },
  "required": ["environment_id", "id"]
}`)

type DeletePopulationTool struct {
   client      tools.PingOneClient
   environment string
}

// NewDeletePopulationTool constructs a new DeletePopulationTool
func NewDeletePopulationTool(client tools.PingOneClient, environment string) tools.Tool {
   return &DeletePopulationTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *DeletePopulationTool) Name() string {
   return "delete_population"
}

// Description returns a human-readable description
func (t *DeletePopulationTool) Description() string {
   return "Delete a population via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *DeletePopulationTool) InputSchema() json.RawMessage {
   return deletePopulationSchema
}

// Run executes the tool logic
func (t *DeletePopulationTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   // Extract environment_id
   rawEnv, ok := args["environment_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: environment_id")
   }
   envID, ok := rawEnv.(string)
   if !ok || envID == "" {
       return nil, fmt.Errorf("environment_id must be a non-empty string")
   }
   // Extract id
   rawID, ok := args["id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: id")
   }
   popID, ok := rawID.(string)
   if !ok || popID == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   // Call PingOne API to delete the population
   if err := t.client.DeletePopulation(ctx, envID, popID); err != nil {
       return nil, err
   }
   // Return success indicator
   return map[string]interface{}{"success": true}, nil
}