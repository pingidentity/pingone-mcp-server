package environments

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeCreateEnvironmentClient implements PingOneClient for testing CreateEnvironmentTool
type fakeCreateEnvironmentClient struct {
   expectedEnvEnv management.Environment
   returnedEnv    management.Environment
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeCreateEnvironmentClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeCreateEnvironmentClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
    return management.Environment{}, nil
}

// CreateEnvironment checks inputs and returns a predefined environment
func (f *fakeCreateEnvironmentClient) CreateEnvironment(ctx context.Context, env management.Environment) (management.Environment, error) {
   if env.Name != f.expectedEnvEnv.Name {
       return management.Environment{}, fmt.Errorf("unexpected name: %s", env.Name)
   }
   if (env.Description == nil) != (f.expectedEnvEnv.Description == nil) {
       return management.Environment{}, fmt.Errorf("unexpected description: %v", env.Description)
   }
   if env.Description != nil && f.expectedEnvEnv.Description != nil && *env.Description != *f.expectedEnvEnv.Description {
       return management.Environment{}, fmt.Errorf("unexpected description: %s", *env.Description)
   }
   return f.returnedEnv, nil
}

// Stub other PingOneClient methods
func (f *fakeCreateEnvironmentClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) { return management.User{}, nil }
func (f *fakeCreateEnvironmentClient) DeleteUser(ctx context.Context, environmentID, userID string) error { return nil }
func (f *fakeCreateEnvironmentClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) { return management.User{}, nil }
func (f *fakeCreateEnvironmentClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) { return management.User{}, nil }
func (f *fakeCreateEnvironmentClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error { return nil }
func (f *fakeCreateEnvironmentClient) ResetUserPassword(ctx context.Context, environmentID, userID, password string) error { return nil }
func (f *fakeCreateEnvironmentClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) { return management.UserEnabled{}, nil }
func (f *fakeCreateEnvironmentClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeCreateEnvironmentClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeCreateEnvironmentClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error { return nil }
func (f *fakeCreateEnvironmentClient) CreatePopulation(ctx context.Context, environmentID string, pop management.Population) (management.Population, error) { return management.Population{}, nil }
func (f *fakeCreateEnvironmentClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error { return nil }
func (f *fakeCreateEnvironmentClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeCreateEnvironmentClient) CreateGroup(ctx context.Context, environmentID string, grp management.Group) (management.Group, error) { return management.Group{}, nil }
func (f *fakeCreateEnvironmentClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error { return nil }
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeCreateEnvironmentClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeCreateEnvironmentClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}

// GetEnvironment stub to satisfy PingOneClient
func (f *fakeCreateEnvironmentClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}

func TestCreateEnvironmentTool_Success(t *testing.T) {
   name := "envName"
   desc := "envDesc"
   // build expected input model
   expectedEnvModel := management.Environment{ Name: name, Description: &desc }
   // simulate server response with assigned ID
   serverEnv := expectedEnvModel
   id := "env123"
   serverEnv.Id = &id
   // Ensure returnedEnv has required fields for JSON marshalling
   serverEnv.Region = management.EnumRegionCodeAsEnvironmentRegion(management.ENUMREGIONCODE_NA.Ptr())
   serverEnv.Type = management.ENUMENVIRONMENTTYPE_PRODUCTION
   serverEnv.License = *management.NewEnvironmentLicense("")
   fake := &fakeCreateEnvironmentClient{ expectedEnvEnv: expectedEnvModel, returnedEnv: serverEnv }
   tool := NewCreateEnvironmentTool(fake)

   args := map[string]interface{}{ "name": name, "description": desc }
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   if out["id"] != id {
       t.Errorf("expected id %s, got %v", id, out["id"])
   }
   if out["name"] != name {
       t.Errorf("expected name %s, got %v", name, out["name"])
   }
   if out["description"] != desc {
       t.Errorf("expected description %s, got %v", desc, out["description"])
   }
}

func TestCreateEnvironmentTool_MissingArgs(t *testing.T) {
   fake := &fakeCreateEnvironmentClient{}
   tool := NewCreateEnvironmentTool(fake)
   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing name, got nil")
   }
}