package users

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeGetClient implements PingOneClient for testing GetUserTool
type fakeGetClient struct {
   expectedEnv    string
   expectedUserID string
   returnedUser   management.User
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeGetClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}
// Stub methods to satisfy the PingOneClient interface
func (f *fakeGetClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeGetClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeGetClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
// RecoverUserPassword stub to satisfy PingOneClient
func (f *fakeGetClient) RecoverUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
// UnlockUserPassword stub to satisfy PingOneClient
func (f *fakeGetClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
// ResetUserPassword stub to satisfy PingOneClient
func (f *fakeGetClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
func (f *fakeGetClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeGetClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
// GetUserPasswordState stub to satisfy PingOneClient
func (f *fakeGetClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
// AddUserToGroup stub to satisfy PingOneClient
func (f *fakeGetClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreatePopulation stub to satisfy PingOneClient
func (f *fakeGetClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
// DeletePopulation stub to satisfy PingOneClient
func (f *fakeGetClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeGetClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeGetClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// RemoveUserFromGroup stub to satisfy PingOneClient
func (f *fakeGetClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeGetClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeGetClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeGetClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeGetClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeGetClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeGetClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}

// GetUser checks inputs and returns a predefined user
func (f *fakeGetClient) GetUser(ctx context.Context, envID, userID string) (management.User, error) {
   if envID != f.expectedEnv {
       return management.User{}, fmt.Errorf("unexpected environmentID: %s", envID)
   }
   if userID != f.expectedUserID {
       return management.User{}, fmt.Errorf("unexpected userID: %s", userID)
   }
   return f.returnedUser, nil
}

func TestGetUserTool_Success(t *testing.T) {
   // Arrange
   id := "user123"
   user := management.User{Id: &id, Username: "alice", Email: "alice@example.com"}
   fake := &fakeGetClient{expectedEnv: "env1", expectedUserID: id, returnedUser: user}
   tool := NewGetUserTool(fake, "default_env")

   // Act
   args := map[string]interface{}{"environment_id": "env1", "id": id}
   out, err := tool.Run(context.Background(), args)

   // Assert
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   if out["id"] != id {
       t.Errorf("expected id %s, got %v", id, out["id"])
   }
   if out["username"] != "alice" {
       t.Errorf("expected username alice, got %v", out["username"])
   }
   if out["email"] != "alice@example.com" {
       t.Errorf("expected email alice@example.com, got %v", out["email"])
   }
}

func TestGetUserTool_MissingArgs(t *testing.T) {
   fake := &fakeGetClient{}
   tool := NewGetUserTool(fake, "")

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