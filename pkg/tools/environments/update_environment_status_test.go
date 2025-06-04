package environments

import (
   "context"
   "errors"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// stubUpdater implements EnvironmentStatusUpdater for testing
type stubUpdater struct {
   calledID     string
   calledStatus management.EnumEnvironmentStatus
   returnEnv    management.Environment
   returnErr    error
}

func (s *stubUpdater) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   s.calledID = environmentID
   s.calledStatus = status
   return s.returnEnv, s.returnErr
}

func TestUpdateEnvironmentStatusTool_Run(t *testing.T) {
   ctx := context.Background()
   stub := &stubUpdater{}
   tool := NewUpdateEnvironmentStatusTool(stub)

   // Missing id
   if _, err := tool.Run(ctx, map[string]interface{}{}); err == nil || err.Error() != "missing required argument: id" {
       t.Errorf("expected missing id error, got %v", err)
   }

   // Missing status
   if _, err := tool.Run(ctx, map[string]interface{}{"id": "env1"}); err == nil || err.Error() != "missing required argument: status" {
       t.Errorf("expected missing status error, got %v", err)
   }

   // Wrong type id
   if _, err := tool.Run(ctx, map[string]interface{}{"id": 123, "status": "DELETE_PENDING"}); err == nil || err.Error() != "id must be a non-empty string" {
       t.Errorf("expected type error for id, got %v", err)
   }

   // Wrong type status
   if _, err := tool.Run(ctx, map[string]interface{}{"id": "env1", "status": 456}); err == nil || err.Error() != "status must be a non-empty string" {
       t.Errorf("expected type error for status, got %v", err)
   }

   // Invalid status
   if _, err := tool.Run(ctx, map[string]interface{}{"id": "env1", "status": "BAD"}); err == nil {
       t.Errorf("expected invalid status error, got nil")
   }

   // Propagate error
   stub.returnErr = errors.New("update failed")
   if _, err := tool.Run(ctx, map[string]interface{}{"id": "env1", "status": "ACTIVE"}); err == nil || err.Error() != "update failed" {
       t.Errorf("expected update failed error, got %v", err)
   }

   // Success
   stub.returnErr = nil
   // Ensure returnedEnv has required fields for JSON marshalling
   e := management.Environment{Id: ptrString("env1"), Name: "env1"}
   e.Region = management.EnumRegionCodeAsEnvironmentRegion(management.ENUMREGIONCODE_NA.Ptr())
   e.Type = management.ENUMENVIRONMENTTYPE_PRODUCTION
   e.License = *management.NewEnvironmentLicense("")
   // Set status to match requested
   e.Status = management.ENUMENVIRONMENTSTATUS_DELETE_PENDING.Ptr()
   stub.returnEnv = e
   out, err := tool.Run(ctx, map[string]interface{}{"id": "env1", "status": "DELETE_PENDING"})
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   if out["id"] != "env1" {
       t.Errorf("expected id env1, got %v", out["id"])
   }
   if out["status"] != "DELETE_PENDING" {
       t.Errorf("expected status DELETE_PENDING, got %v", out["status"])
   }
}

// helper to get pointer to string
func ptrString(s string) *string { return &s }