package membership

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeRemoveClient implements PingOneClient for testing RemoveUserFromGroupTool
type fakeRemoveClient struct {
   expectedEnv     string
   expectedUserID  string
   expectedGroupID string
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeRemoveClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreatePopulation stub to satisfy PingOneClient
func (f *fakeRemoveClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeRemoveClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeRemoveClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// DeletePopulation stub to satisfy PingOneClient
func (f *fakeRemoveClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeRemoveClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeRemoveClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}

// RemoveUserFromGroup verifies parameters and simulates removal
func (f *fakeRemoveClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   if environmentID != f.expectedEnv {
       return fmt.Errorf("unexpected environmentID: %s", environmentID)
   }
   if userID != f.expectedUserID {
       return fmt.Errorf("unexpected userID: %s", userID)
   }
   if groupID != f.expectedGroupID {
       return fmt.Errorf("unexpected groupID: %s", groupID)
   }
   return nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeRemoveClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}

// Stub other PingOneClient methods
func (f *fakeRemoveClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeRemoveClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeRemoveClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeRemoveClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeRemoveClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeRemoveClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
func (f *fakeRemoveClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
func (f *fakeRemoveClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeRemoveClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeRemoveClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeRemoveClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeRemoveClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeRemoveClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}

func TestRemoveUserFromGroupTool_Success(t *testing.T) {
   env := "env1"
   userID := "u123"
   groupID := "g456"
   fake := &fakeRemoveClient{expectedEnv: env, expectedUserID: userID, expectedGroupID: groupID}
   tool := NewRemoveUserFromGroupTool(fake, "unused_env")

   args := map[string]interface{}{"environment_id": env, "user_id": userID, "group_id": groupID}
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   success, ok := out["success"].(bool)
   if !ok || !success {
       t.Errorf("expected success=true, got %v", out["success"])
   }
}

func TestRemoveUserFromGroupTool_MissingArgs(t *testing.T) {
   fake := &fakeRemoveClient{}
   tool := NewRemoveUserFromGroupTool(fake, "")

   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }
   if _, err := tool.Run(context.Background(), map[string]interface{}{"environment_id": "e1"}); err == nil {
       t.Fatal("expected error for missing user_id, got nil")
   }
   if _, err := tool.Run(context.Background(), map[string]interface{}{"environment_id": "e1", "user_id": "u1"}); err == nil {
       t.Fatal("expected error for missing group_id, got nil")
   }
}