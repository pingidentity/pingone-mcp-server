package populations

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// createPopulationSchema defines the input JSON schema for create_population
var createPopulationSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" },
    "name":           { "type": "string" },
    "description":    { "type": "string" }
  },
  "required": ["environment_id", "name"]
}`)

type CreatePopulationTool struct {
   client      tools.PingOneClient
   environment string
}

// NewCreatePopulationTool constructs a new CreatePopulationTool
func NewCreatePopulationTool(client tools.PingOneClient, environment string) tools.Tool {
   return &CreatePopulationTool{client: client, environment: environment}
}

// Name returns the tool name
func (t *CreatePopulationTool) Name() string {
   return "create_population"
}

// Description returns a human-readable description
func (t *CreatePopulationTool) Description() string {
   return "Create a population in PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *CreatePopulationTool) InputSchema() json.RawMessage {
   return createPopulationSchema
}

// Run executes the tool logic
func (t *CreatePopulationTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   // Extract environment_id
   rawEnv, ok := args["environment_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: environment_id")
   }
   envID, ok := rawEnv.(string)
   if !ok || envID == "" {
       return nil, fmt.Errorf("environment_id must be a non-empty string")
   }
   // Extract name
   rawName, ok := args["name"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: name")
   }
   name, ok := rawName.(string)
   if !ok || name == "" {
       return nil, fmt.Errorf("name must be a non-empty string")
   }
   // Optional description
   var descPtr *string
   if rawDesc, ok := args["description"]; ok {
       if desc, ok2 := rawDesc.(string); ok2 {
           descPtr = &desc
       }
   }
   // Build Population model
   pop := management.NewPopulation(name)
   if descPtr != nil {
       pop.Description = descPtr
   }
   // Call API
   created, err := t.client.CreatePopulation(ctx, envID, *pop)
   if err != nil {
       return nil, err
   }
   // Convert to generic map
   var result map[string]interface{}
   data, err := json.Marshal(created)
   if err != nil {
       return nil, err
   }
   if err := json.Unmarshal(data, &result); err != nil {
       return nil, err
   }
   return result, nil
}