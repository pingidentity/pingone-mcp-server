package users

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeUnlockClient implements PingOneClient for testing UnlockUserPasswordTool
type fakeUnlockClient struct {
   expectedEnv    string
   expectedUserID string
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeUnlockClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeUnlockClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
    return management.Environment{}, nil
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeUnlockClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}

// UnlockUserPassword checks the inputs and returns success
func (f *fakeUnlockClient) UnlockUserPassword(ctx context.Context, envID, userID string) error {
   if envID != f.expectedEnv {
       return fmt.Errorf("unexpected environmentID: %s", envID)
   }
   if userID != f.expectedUserID {
       return fmt.Errorf("unexpected userID: %s", userID)
   }
   return nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeUnlockClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeUnlockClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeUnlockClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeUnlockClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeUnlockClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}

// Stub other PingOneClient methods
func (f *fakeUnlockClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeUnlockClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeUnlockClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeUnlockClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeUnlockClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
// ResetUserPassword stub to satisfy PingOneClient
func (f *fakeUnlockClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
// GetUserPasswordState stub to satisfy PingOneClient
func (f *fakeUnlockClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
// AddUserToGroup stub to satisfy PingOneClient
func (f *fakeUnlockClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreatePopulation stub to satisfy PingOneClient
func (f *fakeUnlockClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
// DeletePopulation stub to satisfy PingOneClient
func (f *fakeUnlockClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeUnlockClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeUnlockClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// RemoveUserFromGroup stub to satisfy PingOneClient
func (f *fakeUnlockClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}

func TestUnlockUserPasswordTool_Success(t *testing.T) {
   id := "user123"
   fake := &fakeUnlockClient{expectedEnv: "env1", expectedUserID: id}
   tool := NewUnlockUserPasswordTool(fake, "default_env")

   args := map[string]interface{}{
       "environment_id": "env1",
       "id":             id,
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

func TestUnlockUserPasswordTool_MissingArgs(t *testing.T) {
   fake := &fakeUnlockClient{}
   tool := NewUnlockUserPasswordTool(fake, "")

   // Missing environment_id
   _, err := tool.Run(context.Background(), map[string]interface{}{ /* no environment_id */ })
   if err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }
   // Missing id
   _, err = tool.Run(context.Background(), map[string]interface{}{"environment_id": "env1"})
   if err == nil {
       t.Fatal("expected error for missing id, got nil")
   }
}