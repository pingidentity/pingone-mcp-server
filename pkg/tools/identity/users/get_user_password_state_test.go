package users

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeGetPasswordStateClient implements PingOneClient for testing GetUserPasswordStateTool
type fakeGetPasswordStateClient struct {
   expectedEnv    string
   expectedUserID string
   returnedState  map[string]interface{}
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}
// Stub AddUserToGroup stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// DeletePopulation stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// CreatePopulation stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
// Stub RemoveUserFromGroup stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}

// GetUserPasswordState checks inputs and returns a predefined state
func (f *fakeGetPasswordStateClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   if environmentID != f.expectedEnv {
       return nil, fmt.Errorf("unexpected environmentID: %s", environmentID)
   }
   if userID != f.expectedUserID {
       return nil, fmt.Errorf("unexpected userID: %s", userID)
   }
   return f.returnedState, nil
}

// Stub other PingOneClient methods
func (f *fakeGetPasswordStateClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeGetPasswordStateClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeGetPasswordStateClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeGetPasswordStateClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeGetPasswordStateClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeGetPasswordStateClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeGetPasswordStateClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) DeleteEnvironment(ctx context.Context, environmentID string) error {
   return nil
}
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeGetPasswordStateClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}

func TestGetUserPasswordStateTool_Success(t *testing.T) {
   // Arrange
   env := "env1"
   id := "user123"
   state := map[string]interface{}{"status": "ready", "forceChange": false}
   fake := &fakeGetPasswordStateClient{expectedEnv: env, expectedUserID: id, returnedState: state}
   tool := NewGetUserPasswordStateTool(fake, "unused_env")

   // Act
   args := map[string]interface{}{"environment_id": env, "id": id}
   out, err := tool.Run(context.Background(), args)

   // Assert
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   for k, v := range state {
       if out[k] != v {
           t.Errorf("expected %s=%v, got %v", k, v, out[k])
       }
   }
}

func TestGetUserPasswordStateTool_MissingArgs(t *testing.T) {
   fake := &fakeGetPasswordStateClient{}
   tool := NewGetUserPasswordStateTool(fake, "")

   // Missing environment_id
   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }
   // Missing id
   if _, err := tool.Run(context.Background(), map[string]interface{}{"environment_id": "e1"}); err == nil {
       t.Fatal("expected error for missing id, got nil")
   }
}