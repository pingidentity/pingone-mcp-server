package users

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeUpdateClient implements PingOneClient for testing UpdateUserTool
type fakeUpdateClient struct {
   expectedEnv    string
   expectedUserID string
   expectedUser   management.User
   returnedUser   management.User
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeUpdateClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeUpdateClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
    return management.Environment{}, nil
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeUpdateClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}
// Stub methods to satisfy the PingOneClient interface
func (f *fakeUpdateClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeUpdateClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeUpdateClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeUpdateClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
// RecoverUserPassword stub to satisfy PingOneClient
func (f *fakeUpdateClient) RecoverUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
// UnlockUserPassword stub to satisfy PingOneClient
func (f *fakeUpdateClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
// ResetUserPassword stub to satisfy PingOneClient
func (f *fakeUpdateClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}

// UpdateUser checks that the correct parameters are passed
func (f *fakeUpdateClient) UpdateUser(ctx context.Context, envID, userID string, user management.User) (management.User, error) {
   if envID != f.expectedEnv {
       return management.User{}, fmt.Errorf("unexpected environmentID: %s", envID)
   }
   if userID != f.expectedUserID {
       return management.User{}, fmt.Errorf("unexpected userID: %s", userID)
   }
   if user.Username != f.expectedUser.Username {
       return management.User{}, fmt.Errorf("unexpected username: %s", user.Username)
   }
   if user.Email != f.expectedUser.Email {
       return management.User{}, fmt.Errorf("unexpected email: %s", user.Email)
   }
   return f.returnedUser, nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeUpdateClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeUpdateClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// GetUserPasswordState stub to satisfy PingOneClient
func (f *fakeUpdateClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
// AddUserToGroup stub to satisfy PingOneClient
func (f *fakeUpdateClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreatePopulation stub to satisfy PingOneClient
func (f *fakeUpdateClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
// DeletePopulation stub to satisfy PingOneClient
func (f *fakeUpdateClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeUpdateClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeUpdateClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// RemoveUserFromGroup stub to satisfy PingOneClient
func (f *fakeUpdateClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeUpdateClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeUpdateClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeUpdateClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}

func TestUpdateUserTool_Success(t *testing.T) {
   id := "user123"
   expected := management.User{Username: "alice2", Email: "alice2@example.com"}
   returned := management.User{Id: &id, Username: expected.Username, Email: expected.Email}
   fake := &fakeUpdateClient{
       expectedEnv:    "env1",
       expectedUserID: id,
       expectedUser:   expected,
       returnedUser:   returned,
   }
   tool := NewUpdateUserTool(fake, "default_env")

   args := map[string]interface{}{
       "environment_id": "env1",
       "id":             id,
       "username":       expected.Username,
       "email":          expected.Email,
   }
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   if out["id"] != id {
       t.Errorf("expected id %s, got %v", id, out["id"])
   }
   if out["username"] != expected.Username {
       t.Errorf("expected username %s, got %v", expected.Username, out["username"])
   }
   if out["email"] != expected.Email {
       t.Errorf("expected email %s, got %v", expected.Email, out["email"])
   }
}

func TestUpdateUserTool_MissingArgs(t *testing.T) {
   fake := &fakeUpdateClient{}
   tool := NewUpdateUserTool(fake, "")

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