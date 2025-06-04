package users

import (
   "bytes"
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeClient implements PingOneClient for testing CreateUserTool
type fakeClient struct {
   expectedEnv  string
   expectedUser management.User
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}
// Stub methods to satisfy the PingOneClient interface
func (f *fakeClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
// RecoverUserPassword stub to satisfy PingOneClient
func (f *fakeClient) RecoverUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
// UnlockUserPassword stub to satisfy PingOneClient
func (f *fakeClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
// ResetUserPassword stub to satisfy PingOneClient
func (f *fakeClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
func (f *fakeClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
// GetUserPasswordState stub to satisfy PingOneClient
func (f *fakeClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
// AddUserToGroup stub to satisfy PingOneClient
func (f *fakeClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// RemoveUserFromGroup stub to satisfy PingOneClient
func (f *fakeClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}
// CreatePopulation stub to satisfy PingOneClient
func (f *fakeClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
// DeletePopulation stub to satisfy PingOneClient
func (f *fakeClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}

func (f *fakeClient) CreateUser(ctx context.Context, envID string, user management.User) (management.User, error) {
   if envID != f.expectedEnv {
       return management.User{}, fmt.Errorf("unexpected environmentID: %s", envID)
   }
   if user.Username != f.expectedUser.Username {
       return management.User{}, fmt.Errorf("unexpected username: %s", user.Username)
   }
   if user.Email != f.expectedUser.Email {
       return management.User{}, fmt.Errorf("unexpected email: %s", user.Email)
   }
   id := "id123"
   return management.User{Id: &id, Username: user.Username, Email: user.Email}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeClient) DeleteEnvironment(ctx context.Context, environmentID string) error {
   return nil
}
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}

func TestCreateUserTool_Success(t *testing.T) {
   // Prepare fake client and tool
   expected := management.User{Username: "alice", Email: "alice@example.com"}
   fake := &fakeClient{expectedEnv: "env1", expectedUser: expected}
   tool := NewCreateUserTool(fake, "env1")

   // Check InputSchema contains username and email keys
   schema := tool.InputSchema()
   if !bytes.Contains(schema, []byte("username")) || !bytes.Contains(schema, []byte("email")) {
       t.Errorf("InputSchema missing required properties: %s", schema)
   }

   // Run with valid args
   args := map[string]interface{}{
       "environment_id": "env1",
       "username":       "alice",
       "email":          "alice@example.com",
   }
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   // Validate output
   if out["id"] != "id123" {
       t.Errorf("expected id 'id123', got %v", out["id"])
   }
   if out["username"] != "alice" {
       t.Errorf("expected username 'alice', got %v", out["username"])
   }
   if out["email"] != "alice@example.com" {
       t.Errorf("expected email 'alice@example.com', got %v", out["email"])
   }
}

func TestCreateUserTool_MissingArgs(t *testing.T) {
   tool := NewCreateUserTool(nil, "env1")
   _, err := tool.Run(context.Background(), map[string]interface{}{ /* missing username, email */ })
   if err == nil {
       t.Fatal("expected error for missing arguments, got nil")
   }
}