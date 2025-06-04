package tools

import (
  "context"
  "errors"
  "testing"

  "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeLicenser implements a fake page fetcher
type fakeLicenser struct {
  arr *management.EntityArray
  err error
}

func (f *fakeLicenser) Read(ctx context.Context, org string) management.EntityArrayPagedIterator {
  called := false
  return func(yield func(management.PagedCursor, error) bool) {
      if called {
          return
      }
      called = true
      yield(management.PagedCursor{EntityArray: f.arr}, f.err)
  }
}

func TestGetLicensesTool_Success(t *testing.T) {
  // Create two license objects with IDs
  lic1 := management.NewLicense("L1")
  id1 := "L1"
  lic1.Id = &id1
  lic2 := management.NewLicense("L2")
  id2 := "L2"
  lic2.Id = &id2
  arr := &management.EntityArray{Embedded: &management.EntityArrayEmbedded{Licenses: []management.License{*lic1, *lic2}}}
  fake := &fakeLicenser{arr: arr, err: nil}
  tool := NewGetLicensesTool(fake.Read, "org1")

  out, err := tool.Run(context.Background(), nil)
  if err != nil {
      t.Fatalf("unexpected error: %v", err)
  }
  rawSlice, ok := out["licenses"].([]map[string]interface{})
  if !ok {
      // handle []interface{}
      tmp, _ := out["licenses"].([]interface{})
      rawSlice = make([]map[string]interface{}, len(tmp))
      for i, v := range tmp {
          rawSlice[i] = v.(map[string]interface{})
      }
  }
  if len(rawSlice) != 2 {
      t.Fatalf("expected 2 licenses, got %d", len(rawSlice))
  }
  if rawSlice[0]["id"] != "L1" || rawSlice[1]["id"] != "L2" {
      t.Errorf("unexpected licenses: %v", rawSlice)
  }
}

func TestGetLicensesTool_Error(t *testing.T) {
  fake := &fakeLicenser{arr: nil, err: errors.New("fail")}
  tool := NewGetLicensesTool(fake.Read, "org1")
  if _, err := tool.Run(context.Background(), nil); err == nil || err.Error() != "fail" {
      t.Fatalf("expected error 'fail', got %v", err)
  }
}