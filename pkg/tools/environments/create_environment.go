package environments

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
   "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// createEnvironmentSchema defines the input JSON schema for create_environment
var createEnvironmentSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "name":        { "type": "string" },
    "description": { "type": "string" }
  },
  "required": ["name"]
}`)

// EnvironmentCreator defines the minimal interface to create an environment
type EnvironmentCreator interface {
   CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error)
}

// CreateEnvironmentTool implements Tool for creating an environment
type CreateEnvironmentTool struct {
   creator EnvironmentCreator
}

// NewCreateEnvironmentTool constructs a new CreateEnvironmentTool
// defaultEnvID is used to fetch current environment settings
// NewCreateEnvironmentTool constructs a new CreateEnvironmentTool
// NewCreateEnvironmentTool constructs a new CreateEnvironmentTool
func NewCreateEnvironmentTool(creator EnvironmentCreator) tools.Tool {
   return &CreateEnvironmentTool{creator: creator}
}

// Name returns the tool name
func (t *CreateEnvironmentTool) Name() string {
   return "create_environment"
}

// Description returns a human-readable description
func (t *CreateEnvironmentTool) Description() string {
   return "Create a new environment via PingOne"
}

// InputSchema returns the JSON schema for tool arguments
func (t *CreateEnvironmentTool) InputSchema() json.RawMessage {
   return createEnvironmentSchema
}

// Run executes the tool logic
func (t *CreateEnvironmentTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   rawName, ok := args["name"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: name")
   }
   name, ok := rawName.(string)
   if !ok || name == "" {
       return nil, fmt.Errorf("name must be a non-empty string")
   }
   var descPtr *string
   if rawDesc, ok2 := args["description"]; ok2 {
       if desc, ok3 := rawDesc.(string); ok3 {
           descPtr = &desc
       }
   }
   // Build environment model with defaults, then set name and optional description
   envModel := management.NewEnvironmentWithDefaults()
   envModel.Name = name
   if descPtr != nil {
       envModel.Description = descPtr
   }
   created, err := t.creator.CreateEnvironment(ctx, *envModel)
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