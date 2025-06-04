package tools

import (
   "context"
   "encoding/json"
   "testing"

)

// dummyTool is a no-op Tool implementation for testing registry
type dummyTool struct{ name string }

func (d *dummyTool) Name() string               { return d.name }
func (d *dummyTool) Description() string        { return "desc" }
func (d *dummyTool) InputSchema() json.RawMessage { return json.RawMessage(`{}`) }
func (d *dummyTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
   return map[string]interface{}{"ok": true}, nil
}

func TestRegistryBasics(t *testing.T) {
   // ensure clean state
   Reset()
   if _, found := Get("any"); found {
       t.Fatal("expected no tool registered")
   }
   // register tools
   t1 := &dummyTool{name: "tool1"}
   t2 := &dummyTool{name: "tool2"}
   Register(t1)
   Register(t2)
   // Get by name
   if got, ok := Get("tool1"); !ok || got.Name() != t1.Name() {
       t.Errorf("Get(tool1) = %v, %v; want name %q, true", got, ok, t1.Name())
   }
   if _, ok := Get("missing"); ok {
       t.Error("Get(missing) unexpectedly found a tool")
   }
   // List contains both names
   list := List()
   seen := map[string]bool{}
   for _, ti := range list {
       seen[ti.Name()] = true
   }
   if !seen["tool1"] || !seen["tool2"] {
       t.Errorf("List names = %v; want tool1 and tool2", seen)
   }
   // Reset clears registry
   Reset()
   if l := List(); len(l) != 0 {
       t.Errorf("after Reset, List length = %d; want 0", len(l))
   }
}
