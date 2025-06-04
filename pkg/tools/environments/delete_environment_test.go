package environments

import (
   "context"
   "errors"
   "testing"
)

// stubDeleter is a test double for EnvironmentDeleter
type stubDeleter struct {
   calledID string
   err      error
}

func (s *stubDeleter) DeleteEnvironment(ctx context.Context, environmentID string) error {
   s.calledID = environmentID
   return s.err
}

func TestDeleteEnvironmentTool_Run(t *testing.T) {
   ctx := context.Background()
   stub := &stubDeleter{}
   tool := &DeleteEnvironmentTool{deleter: stub}

   // Missing id
   if _, err := tool.Run(ctx, map[string]interface{}{}); err == nil || err.Error() != "missing required argument: id" {
       t.Errorf("expected missing argument error, got %v", err)
   }

   // Wrong type for id
   args := map[string]interface{}{"id": 123}
   if _, err := tool.Run(ctx, args); err == nil || err.Error() != "id must be a non-empty string" {
       t.Errorf("expected type error for id, got %v", err)
   }

   // Deletion error propagation
   stub.err = errors.New("delete failed")
   args = map[string]interface{}{"id": "env123"}
   if _, err := tool.Run(ctx, args); err == nil || err.Error() != "delete failed" {
       t.Errorf("expected deletion error, got %v", err)
   }

   // Success
   stub.err = nil
   output, err := tool.Run(ctx, map[string]interface{}{"id": "env123"})
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   if success, ok := output["success"].(bool); !ok || !success {
       t.Errorf("expected success=true, got %v", output["success"])
   }
   if stub.calledID != "env123" {
       t.Errorf("expected DeleteEnvironment called with env123, got %s", stub.calledID)
   }
}