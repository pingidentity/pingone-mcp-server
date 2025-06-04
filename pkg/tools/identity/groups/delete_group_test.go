package groups

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeDeleteGroupClient implements PingOneClient for testing DeleteGroupTool
type fakeDeleteGroupClient struct {
   expectedEnv    string
   expectedGroupID string
}

// DeleteGroup checks inputs and simulates deletion
func (f *fakeDeleteGroupClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   if environmentID != f.expectedEnv {
       return fmt.Errorf("unexpected environmentID: %s", environmentID)
   }
   if groupID != f.expectedGroupID {
       return fmt.Errorf("unexpected groupID: %s", groupID)
   }
   return nil
}

// Stub other PingOneClient methods
func (f *fakeDeleteGroupClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeDeleteGroupClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeDeleteGroupClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeDeleteGroupClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeDeleteGroupClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeDeleteGroupClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
func (f *fakeDeleteGroupClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
func (f *fakeDeleteGroupClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeDeleteGroupClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeDeleteGroupClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}
func (f *fakeDeleteGroupClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
func (f *fakeDeleteGroupClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeDeleteGroupClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeDeleteGroupClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeDeleteGroupClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeDeleteGroupClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeDeleteGroupClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
func (f *fakeDeleteGroupClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeDeleteGroupClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeDeleteGroupClient) DeleteEnvironment(ctx context.Context, environmentID string) error {
   return nil
}
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeDeleteGroupClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}

func TestDeleteGroupTool_Success(t *testing.T) {
   env := "env1"
   groupID := "grp123"
   fake := &fakeDeleteGroupClient{expectedEnv: env, expectedGroupID: groupID}
   tool := NewDeleteGroupTool(fake, env)

   args := map[string]interface{}{ "environment_id": env, "id": groupID }
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   success, ok := out["success"].(bool)
   if !ok || !success {
       t.Errorf("expected success=true, got %v", out["success"])
   }
}

func TestDeleteGroupTool_MissingArgs(t *testing.T) {
   fake := &fakeDeleteGroupClient{}
   tool := NewDeleteGroupTool(fake, "unused_env")

   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }
   if _, err := tool.Run(context.Background(), map[string]interface{}{ "environment_id": "env1" }); err == nil {
       t.Fatal("expected error for missing id, got nil")
   }
}