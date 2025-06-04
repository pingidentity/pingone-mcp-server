package users

import (
   "bytes"
   "context"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeFetch simulates paged user fetching
func fakeFetch(ctx context.Context, environmentID, filter string, limit int32, cursor string) management.EntityArrayPagedIterator {
   first := true
   return func(yield func(management.PagedCursor, error) bool) {
       if !first {
           return
       }
       first = false
       user := management.NewUser("u@x.com", "u")
       id := "uid"
       user.Id = &id
       arr := &management.EntityArray{Embedded: &management.EntityArrayEmbedded{Users: []management.User{*user}}}
       yield(management.PagedCursor{EntityArray: arr}, nil)
   }
}

func TestSearchUserTool_Success(t *testing.T) {
   tool := NewSearchUserTool(fakeFetch, "env1")
   schema := tool.InputSchema()
   if !bytes.Contains(schema, []byte("filter")) || !bytes.Contains(schema, []byte("limit")) {
       t.Errorf("schema missing fields: %s", schema)
   }
   args := map[string]interface{}{"filter": "username eq \"u\"", "limit": 1, "cursor": ""}
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   users, ok := out["users"].([]interface{})
   if !ok || len(users) != 1 {
       t.Fatalf("expected 1 user, got %v", out["users"])
   }
   uMap := users[0].(map[string]interface{})
   if uMap["username"] != "u" {
       t.Errorf("expected username 'u', got %v", uMap["username"])
   }
}

func TestSearchUserTool_MissingFilter(t *testing.T) {
   tool := NewSearchUserTool(fakeFetch, "env1")
   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing filter, got nil")
   }
}