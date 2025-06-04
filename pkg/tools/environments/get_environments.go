package environments

import (
  "context"
  "encoding/json"

  "github.com/patrickcping/pingone-go-sdk-v2/management"
  "github.com/patrickcping/pingone-go-sdk-v2/pingone-mcp-server/pkg/tools"
)

// getEnvironmentsSchema defines the input schema for get_environments
var getEnvironmentsSchema = json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`)

// FetchEnvironmentsFunc fetches paged environments
type FetchEnvironmentsFunc func(ctx context.Context) management.EntityArrayPagedIterator

// GetEnvironmentsTool lists all environments
type GetEnvironmentsTool struct {
  fetcher FetchEnvironmentsFunc
}

// NewGetEnvironmentsTool constructs a new GetEnvironmentsTool
func NewGetEnvironmentsTool(fetcher FetchEnvironmentsFunc) tools.Tool {
  return &GetEnvironmentsTool{fetcher: fetcher}
}

// Name returns the tool name
func (t *GetEnvironmentsTool) Name() string { return "get_environments" }

// Description returns a human-readable description
func (t *GetEnvironmentsTool) Description() string { return "Retrieve all environments" }

// InputSchema returns the JSON schema for tool arguments
func (t *GetEnvironmentsTool) InputSchema() json.RawMessage { return getEnvironmentsSchema }

// Run executes the tool logic
func (t *GetEnvironmentsTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
  iterator := t.fetcher(ctx)
  var out []map[string]interface{}
  var fetchErr error
  iterator(func(cursor management.PagedCursor, err error) bool {
    if err != nil {
      fetchErr = err
      return false
    }
    if cursor.EntityArray != nil && cursor.EntityArray.Embedded != nil {
      for _, env := range cursor.EntityArray.Embedded.Environments {
        data, _ := json.Marshal(env)
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
  // Convert slice of maps to slice of interface{} for consistent JSON arrays
  envs := make([]interface{}, len(out))
  for i, m := range out {
    envs[i] = m
  }
  return map[string]interface{}{"environments": envs}, nil
}