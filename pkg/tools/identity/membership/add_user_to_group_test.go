package membership

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeAddClient implements PingOneClient for testing AddUserToGroupTool
type fakeAddClient struct {
   expectedEnv     string
   expectedUserID  string
   expectedGroupID string
   returnedState   map[string]interface{}
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeAddClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreatePopulation stub to satisfy PingOneClient
func (f *fakeAddClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
// DeletePopulation stub to satisfy PingOneClient
func (f *fakeAddClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeAddClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeAddClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// RemoveUserFromGroup stub to satisfy PingOneClient
func (f *fakeAddClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeAddClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}

// AddUserToGroup checks inputs and returns a predefined state
func (f *fakeAddClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   if environmentID != f.expectedEnv {
       return nil, fmt.Errorf("unexpected environmentID: %s", environmentID)
   }
   if userID != f.expectedUserID {
       return nil, fmt.Errorf("unexpected userID: %s", userID)
   }
   if groupID != f.expectedGroupID {
       return nil, fmt.Errorf("unexpected groupID: %s", groupID)
   }
   return f.returnedState, nil
}

// Stub other PingOneClient methods
func (f *fakeAddClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeAddClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeAddClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeAddClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeAddClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeAddClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
func (f *fakeAddClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
func (f *fakeAddClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeAddClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeAddClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeAddClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeAddClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeAddClient) DeleteEnvironment(ctx context.Context, environmentID string) error {
   return nil
}
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeAddClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}

func TestAddUserToGroupTool_Success(t *testing.T) {
   env := "env1"
   userID := "user123"
   groupID := "group456"
   state := map[string]interface{}{"id": groupID, "role": "member"}
   fake := &fakeAddClient{expectedEnv: env, expectedUserID: userID, expectedGroupID: groupID, returnedState: state}
   tool := NewAddUserToGroupTool(fake, "unused_env")

   args := map[string]interface{}{"environment_id": env, "user_id": userID, "group_id": groupID}
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   for k, v := range state {
       if out[k] != v {
           t.Errorf("expected %s=%v, got %v", k, v, out[k])
       }
   }
}

func TestAddUserToGroupTool_MissingArgs(t *testing.T) {
   fake := &fakeAddClient{}
   tool := NewAddUserToGroupTool(fake, "")

   // Missing environment_id
   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }
   // Missing user_id
   if _, err := tool.Run(context.Background(), map[string]interface{}{"environment_id": "e1"}); err == nil {
       t.Fatal("expected error for missing id, got nil")
   }
   // Missing group_id
   if _, err := tool.Run(context.Background(), map[string]interface{}{"environment_id": "e1", "user_id": "u1"}); err == nil {
       t.Fatal("expected error for missing group_id, got nil")
   }
}