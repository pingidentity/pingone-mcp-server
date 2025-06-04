package environments

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// getEnvironmentSchema defines the input JSON schema for get_environment
var getEnvironmentSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "id": { "type": "string" }
  },
  "required": ["id"]
}`)

// GetEnvironmentTool implements Tool for retrieving an environment
type GetEnvironmentTool struct {
   client tools.PingOneClient
}

// NewGetEnvironmentTool constructs a new GetEnvironmentTool
func NewGetEnvironmentTool(client tools.PingOneClient) tools.Tool {
   return &GetEnvironmentTool{client: client}
}

// Name returns the tool name
func (t *GetEnvironmentTool) Name() string {
   return "get_environment"
}

// Description returns a human-readable description
func (t *GetEnvironmentTool) Description() string {
   return "Retrieve an environment via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *GetEnvironmentTool) InputSchema() json.RawMessage {
   return getEnvironmentSchema
}

// Run executes the tool logic
func (t *GetEnvironmentTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   rawID, ok := args["id"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: id")
   }
   id, ok := rawID.(string)
   if !ok || id == "" {
       return nil, fmt.Errorf("id must be a non-empty string")
   }
   env, err := t.client.GetEnvironment(ctx, id)
   if err != nil {
       return nil, err
   }
   buf, err := json.Marshal(env)
   if err != nil {
       return nil, fmt.Errorf("failed to marshal environment: %w", err)
   }
   var output map[string]interface{}
   if err := json.Unmarshal(buf, &output); err != nil {
       return nil, fmt.Errorf("failed to unmarshal environment into map: %w", err)
   }
   return output, nil
}