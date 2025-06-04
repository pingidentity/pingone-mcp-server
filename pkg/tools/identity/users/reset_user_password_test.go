package users

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeResetClient implements PingOneClient for testing ResetUserPasswordTool
type fakeResetClient struct {
   expectedEnv     string
   expectedUserID  string
   expectedPassword string
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeResetClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}

// ResetUserPassword verifies parameters and simulates password reset
func (f *fakeResetClient) ResetUserPassword(ctx context.Context, envID, userID, newPassword string) error {
   if envID != f.expectedEnv {
       return fmt.Errorf("unexpected environmentID: %s", envID)
   }
   if userID != f.expectedUserID {
       return fmt.Errorf("unexpected userID: %s", userID)
   }
   if newPassword != f.expectedPassword {
       return fmt.Errorf("unexpected password: %s", newPassword)
   }
   return nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeResetClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
    return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeResetClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
    return management.Environment{}, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeResetClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeResetClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
    return management.Environment{}, nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeResetClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeResetClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}

// Stub other PingOneClient methods
func (f *fakeResetClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeResetClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeResetClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeResetClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeResetClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
func (f *fakeResetClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
// GetUserPasswordState stub to satisfy PingOneClient
func (f *fakeResetClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
// AddUserToGroup stub to satisfy PingOneClient
func (f *fakeResetClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreatePopulation stub to satisfy PingOneClient
func (f *fakeResetClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
// DeletePopulation stub to satisfy PingOneClient
func (f *fakeResetClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeResetClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeResetClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeResetClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// RemoveUserFromGroup stub to satisfy PingOneClient
func (f *fakeResetClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}

func TestResetUserPasswordTool_Success(t *testing.T) {
   id := "user123"
   pwd := "newPass!"
   fake := &fakeResetClient{expectedEnv: "env1", expectedUserID: id, expectedPassword: pwd}
   tool := NewResetUserPasswordTool(fake, "default_env")

   args := map[string]interface{}{
       "environment_id": "env1",
       "id":             id,
       "password":       pwd,
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

func TestResetUserPasswordTool_MissingArgs(t *testing.T) {
   fake := &fakeResetClient{}
   tool := NewResetUserPasswordTool(fake, "")

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
   // Missing password
   _, err = tool.Run(context.Background(), map[string]interface{}{"environment_id": "env1", "id": "user123"})
   if err == nil {
       t.Fatal("expected error for missing password, got nil")
   }
}