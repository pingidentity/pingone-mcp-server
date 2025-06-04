package users

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeDeleteClient implements PingOneClient for testing DeleteUserTool
type fakeDeleteClient struct {
   expectedEnv    string
   expectedUserID string
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeDeleteClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}
// Stub methods to satisfy the PingOneClient interface
func (f *fakeDeleteClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeDeleteClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeDeleteClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
// RecoverUserPassword stub to satisfy PingOneClient
func (f *fakeDeleteClient) RecoverUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeDeleteClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
// UnlockUserPassword stub to satisfy PingOneClient
func (f *fakeDeleteClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
// ResetUserPassword stub to satisfy PingOneClient
func (f *fakeDeleteClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
func (f *fakeDeleteClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
// GetUserPasswordState stub to satisfy PingOneClient
func (f *fakeDeleteClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
// AddUserToGroup stub to satisfy PingOneClient
func (f *fakeDeleteClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreatePopulation stub to satisfy PingOneClient
func (f *fakeDeleteClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
// DeletePopulation stub to satisfy PingOneClient
func (f *fakeDeleteClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeDeleteClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeDeleteClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeDeleteClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeDeleteClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeDeleteClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeDeleteClient) DeleteEnvironment(ctx context.Context, environmentID string) error {
   return nil
}
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeDeleteClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeDeleteClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
// RemoveUserFromGroup stub to satisfy PingOneClient
func (f *fakeDeleteClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}

// DeleteUser checks the passed environment and userID match expectations
func (f *fakeDeleteClient) DeleteUser(ctx context.Context, envID, userID string) error {
   if envID != f.expectedEnv {
       return fmt.Errorf("unexpected environmentID: %s", envID)
   }
   if userID != f.expectedUserID {
       return fmt.Errorf("unexpected userID: %s", userID)
   }
   return nil
}

func TestDeleteUserTool_Success(t *testing.T) {
   fake := &fakeDeleteClient{expectedEnv: "env1", expectedUserID: "user123"}
   tool := NewDeleteUserTool(fake, "default_env")

   args := map[string]interface{}{
       "environment_id": "env1",
       "id":             "user123",
   }
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   success, ok := out["success"].(bool)
   if !ok || !success {
       t.Errorf("expected success=true, got %v", out["success"])
   }
}

func TestDeleteUserTool_MissingArgs(t *testing.T) {
   fake := &fakeDeleteClient{}
   tool := NewDeleteUserTool(fake, "")

   // Missing environment_id
   _, err := tool.Run(context.Background(), map[string]interface{}{ // no environment_id
       "id": "user123",
   })
   if err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }

   // Missing id
   _, err = tool.Run(context.Background(), map[string]interface{}{ // no id
       "environment_id": "env1",
   })
   if err == nil {
       t.Fatal("expected error for missing user_id, got nil")
   }
}