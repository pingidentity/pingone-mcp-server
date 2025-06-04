package populations

import (
   "context"
   "errors"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakePopFetcher simulates paged population retrieval
type fakePopFetcher struct {
   expectedEnv string
   arr         *management.EntityArray
   err         error
}

// Fetch returns a paged iterator that yields the preset EntityArray or error
func (f *fakePopFetcher) Fetch(ctx context.Context, environmentID string) management.EntityArrayPagedIterator {
   return func(yield func(management.PagedCursor, error) bool) {
       if environmentID != f.expectedEnv {
           yield(management.PagedCursor{}, errors.New("unexpected environment"))
           return
       }
       yield(management.PagedCursor{EntityArray: f.arr}, f.err)
   }
}

func TestGetPopulationsTool_Success(t *testing.T) {
   envID := "env1"
   popID := "pop1"
   pop := management.Population{Id: &popID}
   arr := &management.EntityArray{Embedded: &management.EntityArrayEmbedded{Populations: []management.Population{pop}}}
   fake := &fakePopFetcher{expectedEnv: envID, arr: arr, err: nil}
   tool := NewGetPopulationsTool(fake.Fetch)

   out, err := tool.Run(context.Background(), map[string]interface{}{ "environment_id": envID })
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   items, ok := out["populations"].([]interface{})
   if !ok {
       t.Fatalf("expected []interface{}, got %T", out["populations"])
   }
   if len(items) != 1 {
       t.Fatalf("expected 1 population, got %d", len(items))
   }
   first := items[0].(map[string]interface{})
   if first["id"] != popID {
       t.Errorf("expected id %s, got %v", popID, first["id"])
   }
}

func TestGetPopulationsTool_Error(t *testing.T) {
   envID := "env1"
   fake := &fakePopFetcher{expectedEnv: envID, arr: nil, err: errors.New("fail")}
   tool := NewGetPopulationsTool(fake.Fetch)
   if _, err := tool.Run(context.Background(), map[string]interface{}{ "environment_id": envID }); err == nil || err.Error() != "fail" {
       t.Errorf("expected fail, got %v", err)
   }
}