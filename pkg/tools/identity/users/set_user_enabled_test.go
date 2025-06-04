package users

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeEnableClient implements PingOneClient for testing SetUserEnabledTool
type fakeEnableClient struct {
   expectedEnv     string
   expectedUserID  string
   expectedEnabled bool
   returnedState   management.UserEnabled
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeEnableClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeEnableClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
    return management.Environment{}, nil
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeEnableClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}
// Stub methods to satisfy the PingOneClient interface
func (f *fakeEnableClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeEnableClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeEnableClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeEnableClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
// RecoverUserPassword stub to satisfy PingOneClient
func (f *fakeEnableClient) RecoverUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
// UnlockUserPassword stub to satisfy PingOneClient
func (f *fakeEnableClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
// ResetUserPassword stub to satisfy PingOneClient
func (f *fakeEnableClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
// GetUserPasswordState stub to satisfy PingOneClient
func (f *fakeEnableClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
// AddUserToGroup stub to satisfy PingOneClient
func (f *fakeEnableClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreatePopulation stub to satisfy PingOneClient
func (f *fakeEnableClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
// DeletePopulation stub to satisfy PingOneClient
func (f *fakeEnableClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeEnableClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeEnableClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// RemoveUserFromGroup stub to satisfy PingOneClient
func (f *fakeEnableClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}

// UpdateUserEnabled checks inputs and returns preset state
func (f *fakeEnableClient) UpdateUserEnabled(ctx context.Context, envID, userID string, enabled bool) (management.UserEnabled, error) {
   if envID != f.expectedEnv {
       return management.UserEnabled{}, fmt.Errorf("unexpected environmentID: %s", envID)
   }
   if userID != f.expectedUserID {
       return management.UserEnabled{}, fmt.Errorf("unexpected userID: %s", userID)
   }
   if enabled != f.expectedEnabled {
       return management.UserEnabled{}, fmt.Errorf("unexpected enabled flag: %v", enabled)
   }
   return f.returnedState, nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeEnableClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeEnableClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeEnableClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeEnableClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeEnableClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}

func TestSetUserEnabledTool_Success(t *testing.T) {
   id := "user123"
   // Prepare expected UserEnabled with Enabled=true
   enabledVal := true
   expected := management.UserEnabled{Enabled: &enabledVal}
   fake := &fakeEnableClient{
       expectedEnv:     "env1",
       expectedUserID:  id,
       expectedEnabled: true,
       returnedState:   expected,
   }
   tool := NewSetUserEnabledTool(fake, "default_env")

   args := map[string]interface{}{
       "environment_id": "env1",
       "id":             id,
       "enabled":        true,
   }
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   // The returned map should contain the field enabled
   if out["enabled"] != true {
       t.Errorf("expected enabled true, got %v", out["enabled"])
   }
}

func TestSetUserEnabledTool_MissingArgs(t *testing.T) {
   fake := &fakeEnableClient{}
   tool := NewSetUserEnabledTool(fake, "")

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
   // Missing enabled
   _, err = tool.Run(context.Background(), map[string]interface{}{"environment_id": "env1", "id": "user123"})
   if err == nil {
       t.Fatal("expected error for missing enabled, got nil")
   }
}