package users

import (
   "context"
   "encoding/json"
   "fmt"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
   "github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// searchUserSchema defines the input JSON schema for search_user
var searchUserSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "filter":        { "type": "string" },
    "limit":         { "type": "integer" },
    "cursor":        { "type": "string" }
  },
  "required": ["filter"],
  "additionalProperties": false
}`)

// FetchUsersFunc fetches paged users with optional filter, limit, and cursor
type FetchUsersFunc func(ctx context.Context, environmentID, filter string, limit int32, cursor string) management.EntityArrayPagedIterator

// SearchUserTool implements Tool for searching users
type SearchUserTool struct {
   fetcher     FetchUsersFunc
   environment string
}

// NewSearchUserTool constructs a new SearchUserTool
func NewSearchUserTool(fetcher FetchUsersFunc, environment string) tools.Tool {
   return &SearchUserTool{fetcher: fetcher, environment: environment}
}

// Name returns the tool name
func (t *SearchUserTool) Name() string { return "search_user" }

// Description returns a human-readable description
func (t *SearchUserTool) Description() string { return "Search users via PingOne with filtering" }

// InputSchema returns the JSON schema for tool arguments
func (t *SearchUserTool) InputSchema() json.RawMessage { return searchUserSchema }

// Run executes the tool logic
func (t *SearchUserTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   // Extract filter
   rawFilter, ok := args["filter"]
   if !ok {
       return nil, fmt.Errorf("missing required argument: filter")
   }
   filter, ok := rawFilter.(string)
   if !ok || filter == "" {
       return nil, fmt.Errorf("filter must be a non-empty string")
   }
   // Extract optional limit
   var limit int32
   if rawLimit, ok := args["limit"]; ok {
       switch v := rawLimit.(type) {
       case float64:
           limit = int32(v)
       case int:
           limit = int32(v)
       case int32:
           limit = v
       case int64:
           limit = int32(v)
       default:
           return nil, fmt.Errorf("limit must be a number")
       }
   }
   // Extract optional cursor
   var cursor string
   if rawCursor, ok := args["cursor"]; ok {
       s, ok2 := rawCursor.(string)
       if !ok2 {
           return nil, fmt.Errorf("cursor must be a string")
       }
       cursor = s
   }
   // Fetch pages of users
   iterator := t.fetcher(ctx, t.environment, filter, limit, cursor)
   var all []interface{}
   var fetchErr error
   iterator(func(pc management.PagedCursor, err error) bool {
       if err != nil {
           fetchErr = err
           return false
       }
       if pc.EntityArray != nil && pc.EntityArray.Embedded != nil {
           for _, u := range pc.EntityArray.Embedded.Users {
               data, _ := json.Marshal(u)
               var m map[string]interface{}
               _ = json.Unmarshal(data, &m)
               all = append(all, m)
           }
       }
       return true
   })
   if fetchErr != nil {
       return nil, fetchErr
   }
   return map[string]interface{}{"users": all}, nil
}