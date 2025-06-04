package populations

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeCreatePopulationClient implements PingOneClient for testing CreatePopulationTool
type fakeCreatePopulationClient struct {
   expectedEnv string
   expectedPop management.Population
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeCreatePopulationClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeCreatePopulationClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
    return management.Environment{}, nil
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeCreatePopulationClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}

// CreatePopulation checks inputs and returns a predefined population
func (f *fakeCreatePopulationClient) CreatePopulation(ctx context.Context, environmentID string, pop management.Population) (management.Population, error) {
   if environmentID != f.expectedEnv {
       return management.Population{}, fmt.Errorf("unexpected environmentID: %s", environmentID)
   }
   if pop.Name != f.expectedPop.Name {
       return management.Population{}, fmt.Errorf("unexpected name: %s", pop.Name)
   }
   if (pop.Description == nil && f.expectedPop.Description != nil) || (pop.Description != nil && f.expectedPop.Description == nil) {
       return management.Population{}, fmt.Errorf("unexpected description: %v", pop.Description)
   }
   if pop.Description != nil && f.expectedPop.Description != nil && *pop.Description != *f.expectedPop.Description {
       return management.Population{}, fmt.Errorf("unexpected description: %s", *pop.Description)
   }
   // Simulate server-assigned ID
   return f.expectedPop, nil
}

// Stub other PingOneClient methods to satisfy interface
func (f *fakeCreatePopulationClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeCreatePopulationClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeCreatePopulationClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeCreatePopulationClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeCreatePopulationClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeCreatePopulationClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
func (f *fakeCreatePopulationClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
func (f *fakeCreatePopulationClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeCreatePopulationClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeCreatePopulationClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}
// DeletePopulation stub to satisfy PingOneClient
func (f *fakeCreatePopulationClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeCreatePopulationClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
    return nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeCreatePopulationClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeCreatePopulationClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeCreatePopulationClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeCreatePopulationClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeCreatePopulationClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeCreatePopulationClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}

func TestCreatePopulationTool_Success(t *testing.T) {
   env := "env1"
   name := "myPop"
   desc := "a desc"
   expected := management.Population{Name: name}
   expected.SetDescription(desc)
   // populate ID
   id := "pop123"
   expected.Id = &id
   fake := &fakeCreatePopulationClient{expectedEnv: env, expectedPop: expected}
   tool := NewCreatePopulationTool(fake, env)

   args := map[string]interface{}{ 
       "environment_id": env,
       "name":           name,
       "description":    desc,
   }
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

func TestCreatePopulationTool_MissingArgs(t *testing.T) {
   fake := &fakeCreatePopulationClient{}
   tool := NewCreatePopulationTool(fake, "")
   // missing environment_id
   if _, err := tool.Run(context.Background(), map[string]interface{}{ "name": "n" }); err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }
   // missing name
   if _, err := tool.Run(context.Background(), map[string]interface{}{ "environment_id": "e" }); err == nil {
       t.Fatal("expected error for missing name, got nil")
   }
}