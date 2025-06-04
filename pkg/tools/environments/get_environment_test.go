package environments

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeGetEnvironmentClient implements PingOneClient for testing GetEnvironmentTool
type fakeGetEnvironmentClient struct {
   expectedID    string
   returnedEnv   management.Environment
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeGetEnvironmentClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeGetEnvironmentClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
    return management.Environment{}, nil
}

// GetEnvironment checks input and returns a predefined environment
func (f *fakeGetEnvironmentClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   if environmentID != f.expectedID {
       return management.Environment{}, fmt.Errorf("unexpected environmentID: %s", environmentID)
   }
   return f.returnedEnv, nil
}

// stub other PingOneClient methods
func (f *fakeGetEnvironmentClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) { return management.User{}, nil }
func (f *fakeGetEnvironmentClient) DeleteUser(ctx context.Context, environmentID, userID string) error { return nil }
func (f *fakeGetEnvironmentClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) { return management.User{}, nil }
func (f *fakeGetEnvironmentClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) { return management.User{}, nil }
func (f *fakeGetEnvironmentClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error { return nil }
func (f *fakeGetEnvironmentClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPwd string) error { return nil }
func (f *fakeGetEnvironmentClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) { return management.UserEnabled{}, nil }
func (f *fakeGetEnvironmentClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeGetEnvironmentClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeGetEnvironmentClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error { return nil }
func (f *fakeGetEnvironmentClient) CreatePopulation(ctx context.Context, environmentID string, pop management.Population) (management.Population, error) { return management.Population{}, nil }
func (f *fakeGetEnvironmentClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error { return nil }
func (f *fakeGetEnvironmentClient) CreateGroup(ctx context.Context, environmentID string, grp management.Group) (management.Group, error) { return management.Group{}, nil }
func (f *fakeGetEnvironmentClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error { return nil }
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeGetEnvironmentClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
func (f *fakeGetEnvironmentClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeGetEnvironmentClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeGetEnvironmentClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) { return management.Environment{}, nil }
// GetLicense stub to satisfy PingOneClient
func (f *fakeGetEnvironmentClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}

func TestGetEnvironmentTool_Success(t *testing.T) {
   id := "env123"
   name := "envName"
   desc := "description"
   envModel := management.NewEnvironmentWithDefaults()
   envModel.Id = &id
   envModel.Name = name
   envModel.Description = &desc
   // Populate required fields for JSON marshaling
   envModel.Region = management.EnumRegionCodeAsEnvironmentRegion(management.ENUMREGIONCODE_NA.Ptr())
   envModel.Type = management.ENUMENVIRONMENTTYPE_PRODUCTION
   envModel.License = *management.NewEnvironmentLicense("")
   fake := &fakeGetEnvironmentClient{expectedID: id, returnedEnv: *envModel}
   tool := NewGetEnvironmentTool(fake)

   args := map[string]interface{}{"id": id}
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

func TestGetEnvironmentTool_MissingArgs(t *testing.T) {
   fake := &fakeGetEnvironmentClient{}
   tool := NewGetEnvironmentTool(fake)
   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing id, got nil")
   }
}