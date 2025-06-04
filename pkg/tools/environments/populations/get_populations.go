package populations

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// getPopulationsSchema defines the input JSON schema for get_environment_populations
var getPopulationsSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "environment_id": { "type": "string" }
  },
  "required": ["environment_id"],
  "additionalProperties": false
}`)

// FetchPopulationsFunc fetches paged populations for an environment
type FetchPopulationsFunc func(ctx context.Context, environmentID string) management.EntityArrayPagedIterator

// GetPopulationsTool lists all populations in an environment
// GetPopulationsTool lists all populations in an environment
type GetPopulationsTool struct {
   fetcher FetchPopulationsFunc
}

// NewGetPopulationsTool constructs a new GetPopulationsTool
func NewGetPopulationsTool(fetcher FetchPopulationsFunc) tools.Tool {
   return &GetPopulationsTool{fetcher: fetcher}
}

// Name returns the tool name
func (t *GetPopulationsTool) Name() string {
   return "get_environment_populations"
}

// Description returns a human-readable description
func (t *GetPopulationsTool) Description() string {
   return "Retrieve all populations for an environment"
}

// InputSchema returns the JSON schema for tool arguments
func (t *GetPopulationsTool) InputSchema() json.RawMessage {
   return getPopulationsSchema
}

// Run executes the tool logic
func (t *GetPopulationsTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   rawEnv, ok := args["environment_id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: environment_id")
   }
   envID, ok := rawEnv.(string)
   if !ok || envID == "" {
       return nil, fmt.Errorf("environment_id must be a non-empty string")
   }
   iterator := t.fetcher(ctx, envID)
   var out []map[string]interface{}
   var fetchErr error
   iterator(func(cursor management.PagedCursor, err error) bool {
       if err != nil {
           fetchErr = err
           return false
       }
       if cursor.EntityArray != nil && cursor.EntityArray.Embedded != nil {
           for _, pop := range cursor.EntityArray.Embedded.Populations {
               data, _ := json.Marshal(pop)
               var m map[string]interface{}
               json.Unmarshal(data, &m)
               out = append(out, m)
           }
       }
       return true
   })
   if fetchErr != nil {
       return nil, fetchErr
   }
   // Convert to slice of interface{} for JSON array
   pops := make([]interface{}, len(out))
   for i, m := range out {
       pops[i] = m
   }
   return map[string]interface{}{"populations": pops}, nil
}