package groups

import (
   "context"
   "encoding/json"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeGetGroupClient implements PingOneClient for testing GetGroupTool
type fakeGetGroupClient struct {
   expectedEnv   string
   expectedGroup string
   returnedGroup management.Group
}

// GetGroup checks inputs and returns a predefined group
func (f *fakeGetGroupClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   if environmentID != f.expectedEnv {
       return nil, fmt.Errorf("unexpected environmentID: %s", environmentID)
   }
   if groupID != f.expectedGroup {
       return nil, fmt.Errorf("unexpected groupID: %s", groupID)
   }
   data, _ := json.Marshal(f.returnedGroup)
   var result map[string]interface{}
   _ = json.Unmarshal(data, &result)
   return result, nil
}

// Stub methods to satisfy PingOneClient
func (f *fakeGetGroupClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeGetGroupClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeGetGroupClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeGetGroupClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeGetGroupClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error { return nil }
func (f *fakeGetGroupClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPwd string) error { return nil }
func (f *fakeGetGroupClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) { return management.UserEnabled{}, nil }
func (f *fakeGetGroupClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeGetGroupClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeGetGroupClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error { return nil }
func (f *fakeGetGroupClient) CreatePopulation(ctx context.Context, environmentID string, pop management.Population) (management.Population, error) { return management.Population{}, nil }
func (f *fakeGetGroupClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error { return nil }
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeGetGroupClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeGetGroupClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
func (f *fakeGetGroupClient) CreateGroup(ctx context.Context, environmentID string, grp management.Group) (management.Group, error) { return f.returnedGroup, nil }
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeGetGroupClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error { return nil }
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeGetGroupClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return f.returnedGroup, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeGetGroupClient) DeleteEnvironment(ctx context.Context, environmentID string) error {
   return nil
}
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeGetGroupClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeGetGroupClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeGetGroupClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}

func TestGetGroupTool_Success(t *testing.T) {
   env := "env1"
   grpID := "grp123"
   name := "group1"
   desc := "desc"
   grp := management.NewGroup(name)
   grp.Id = &grpID
   grp.Description = &desc
   fake := &fakeGetGroupClient{expectedEnv: env, expectedGroup: grpID, returnedGroup: *grp}
   tool := NewGetGroupTool(fake, env)

   args := map[string]interface{}{"environment_id": env, "id": grpID}
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   if out["id"] != grpID {
       t.Errorf("expected id %s, got %v", grpID, out["id"])
   }
   if out["name"] != name {
       t.Errorf("expected name %s, got %v", name, out["name"])
   }
   if out["description"] != desc {
       t.Errorf("expected description %s, got %v", desc, out["description"])
   }
}

func TestGetGroupTool_MissingArgs(t *testing.T) {
   fake := &fakeGetGroupClient{}
   tool := NewGetGroupTool(fake, "unused_env")
   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }
   if _, err := tool.Run(context.Background(), map[string]interface{}{ "environment_id": "e1" }); err == nil {
       t.Fatal("expected error for missing id, got nil")
   }
}