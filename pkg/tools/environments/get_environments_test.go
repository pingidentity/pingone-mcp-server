package environments

import (
  "context"
  "errors"
  "testing"

  "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeEnvFetcher simulates paged environment retrieval
type fakeEnvFetcher struct {
  arr *management.EntityArray
  err error
}

func (f *fakeEnvFetcher) Fetch(ctx context.Context) management.EntityArrayPagedIterator {
  called := false
  return func(yield func(management.PagedCursor, error) bool) {
    if called {
      return
    }
    called = true
    yield(management.PagedCursor{EntityArray: f.arr}, f.err)
  }
}

func TestGetEnvironmentsTool_Success(t *testing.T) {
  // Setup fake environments
  id1, id2 := "e1", "e2"
  env1 := management.NewEnvironmentWithDefaults()
  env1.Id = &id1
  env1.Name = "Env1"
  env1.Region = management.EnumRegionCodeAsEnvironmentRegion(management.ENUMREGIONCODE_NA.Ptr())
  env1.Type = management.ENUMENVIRONMENTTYPE_PRODUCTION
  env1.License = *management.NewEnvironmentLicense("")

  env2 := management.NewEnvironmentWithDefaults()
  env2.Id = &id2
  env2.Name = "Env2"
  env2.Region = management.EnumRegionCodeAsEnvironmentRegion(management.ENUMREGIONCODE_NA.Ptr())
  env2.Type = management.ENUMENVIRONMENTTYPE_PRODUCTION
  env2.License = *management.NewEnvironmentLicense("")

  arr := &management.EntityArray{Embedded: &management.EntityArrayEmbedded{Environments: []management.Environment{*env1, *env2}}}
  fake := &fakeEnvFetcher{arr: arr, err: nil}
  tool := NewGetEnvironmentsTool(fake.Fetch)

  out, err := tool.Run(context.Background(), nil)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  rawSlice, ok := out["environments"].([]interface{})
  if !ok {
    t.Fatalf("expected []interface{}, got %T", out["environments"])
  }
  if len(rawSlice) != 2 {
    t.Fatalf("expected 2, got %d", len(rawSlice))
  }
  m1 := rawSlice[0].(map[string]interface{})
  if m1["id"] != id1 || m1["name"] != "Env1" {
    t.Errorf("unexpected first env: %v", m1)
  }
}

func TestGetEnvironmentsTool_Error(t *testing.T) {
  fake := &fakeEnvFetcher{arr: nil, err: errors.New("fail")}
  tool := NewGetEnvironmentsTool(fake.Fetch)
  if _, err := tool.Run(context.Background(), nil); err == nil || err.Error() != "fail" {
    t.Errorf("expected fail, got %v", err)
  }
}