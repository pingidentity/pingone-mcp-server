package populations

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// getPopulationSchema defines the input JSON schema for get_population
var getPopulationSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "id":             { "type": "string" }
  },
  "required": ["environment_id", "id"]
}`)

type GetPopulationTool struct {
   client      tools.PingOneClient
   environment string
}

// NewGetPopulationTool constructs a new GetPopulationTool
func NewGetPopulationTool(client tools.PingOneClient, environment string) tools.Tool {
   return &GetPopulationTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *GetPopulationTool) Name() string {
   return "get_population"
}

// Description returns a human-readable description
func (t *GetPopulationTool) Description() string {
   return "Retrieve a population via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *GetPopulationTool) InputSchema() json.RawMessage {
   return getPopulationSchema
}

// Run executes the tool logic
func (t *GetPopulationTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
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
   // Call PingOne API
   result, err := t.client.GetPopulation(ctx, envID, popID)
   if err != nil {
       return nil, err
   }
   return result, nil
}