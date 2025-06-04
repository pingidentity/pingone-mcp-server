package groups

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeCreateGroupClient implements PingOneClient for testing CreateGroupTool
type fakeCreateGroupClient struct {
   expectedEnv    string
   expectedGroup  management.Group
   returnedGroup  management.Group
}

// CreateGroup checks inputs and returns a predefined group
func (f *fakeCreateGroupClient) CreateGroup(ctx context.Context, environmentID string, grp management.Group) (management.Group, error) {
   if environmentID != f.expectedEnv {
       return management.Group{}, fmt.Errorf("unexpected environmentID: %s", environmentID)
   }
   if grp.Name != f.expectedGroup.Name {
       return management.Group{}, fmt.Errorf("unexpected name: %s", grp.Name)
   }
   // description may be nil
   if (grp.Description == nil) != (f.expectedGroup.Description == nil) {
       return management.Group{}, fmt.Errorf("unexpected description: %v", grp.Description)
   }
   if grp.Description != nil && f.expectedGroup.Description != nil && *grp.Description != *f.expectedGroup.Description {
       return management.Group{}, fmt.Errorf("unexpected description: %s", *grp.Description)
   }
   return f.returnedGroup, nil
}

// Stub other PingOneClient methods to satisfy interface
func (f *fakeCreateGroupClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeCreateGroupClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeCreateGroupClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeCreateGroupClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeCreateGroupClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeCreateGroupClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
func (f *fakeCreateGroupClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
func (f *fakeCreateGroupClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeCreateGroupClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeCreateGroupClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}
func (f *fakeCreateGroupClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
func (f *fakeCreateGroupClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeCreateGroupClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeCreateGroupClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
func (f *fakeCreateGroupClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeCreateGroupClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeCreateGroupClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeCreateGroupClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeCreateGroupClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeCreateGroupClient) DeleteEnvironment(ctx context.Context, environmentID string) error {
   return nil
}
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeCreateGroupClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}

func TestCreateGroupTool_Success(t *testing.T) {
   env := "env1"
   grpID := "grp123"
   name := "myGroup"
   desc := "group desc"
   expected := management.Group{Name: name}
   expected.Id = &grpID
   expected.Description = &desc
   fake := &fakeCreateGroupClient{expectedEnv: env, expectedGroup: expected, returnedGroup: expected}
   tool := NewCreateGroupTool(fake, env)

   args := map[string]interface{}{ "environment_id": env, "name": name, "description": desc }
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

func TestCreateGroupTool_MissingArgs(t *testing.T) {
   fake := &fakeCreateGroupClient{}
   tool := NewCreateGroupTool(fake, "")
   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }
   if _, err := tool.Run(context.Background(), map[string]interface{}{ "environment_id": "e1" }); err == nil {
       t.Fatal("expected error for missing name, got nil")
   }
}