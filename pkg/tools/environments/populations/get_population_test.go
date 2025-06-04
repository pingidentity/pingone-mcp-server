package populations

import (
   "context"
   "encoding/json"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeGetPopulationClient implements PingOneClient for testing GetPopulationTool
type fakeGetPopulationClient struct {
   expectedEnv        string
   expectedPopulation string
   returnedPopulation management.Population
}

// GetPopulation checks inputs and returns a predefined population
func (f *fakeGetPopulationClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   if environmentID != f.expectedEnv {
       return nil, fmt.Errorf("unexpected environmentID: %s", environmentID)
   }
   if populationID != f.expectedPopulation {
       return nil, fmt.Errorf("unexpected populationID: %s", populationID)
   }
   // Convert returnedPopulation to map[string]interface{}
   data, _ := json.Marshal(f.returnedPopulation)
   var result map[string]interface{}
   _ = json.Unmarshal(data, &result)
   return result, nil
}

// Stub other PingOneClient methods to satisfy interface
func (f *fakeGetPopulationClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeGetPopulationClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeGetPopulationClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeGetPopulationClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeGetPopulationClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeGetPopulationClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
func (f *fakeGetPopulationClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
func (f *fakeGetPopulationClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeGetPopulationClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeGetPopulationClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}
func (f *fakeGetPopulationClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
func (f *fakeGetPopulationClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeGetPopulationClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeGetPopulationClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeGetPopulationClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeGetPopulationClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeGetPopulationClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeGetPopulationClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeGetPopulationClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeGetPopulationClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeGetPopulationClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}

func TestGetPopulationTool_Success(t *testing.T) {
   env := "env1"
   popID := "pop123"
   name := "popName"
   desc := "desc"
   pop := management.Population{Name: name}
   pop.Id = &popID
   pop.Description = &desc
   fake := &fakeGetPopulationClient{expectedEnv: env, expectedPopulation: popID, returnedPopulation: pop}
   tool := NewGetPopulationTool(fake, "unused_env")

   args := map[string]interface{}{"environment_id": env, "id": popID}
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   if out["id"] != popID {
       t.Errorf("expected id %s, got %v", popID, out["id"])
   }
   if out["name"] != name {
       t.Errorf("expected name %s, got %v", name, out["name"])
   }
   if out["description"] != desc {
       t.Errorf("expected description %s, got %v", desc, out["description"])
   }
}

func TestGetPopulationTool_MissingArgs(t *testing.T) {
   fake := &fakeGetPopulationClient{}
   tool := NewGetPopulationTool(fake, "")
   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }
   if _, err := tool.Run(context.Background(), map[string]interface{}{"environment_id": "env1"}); err == nil {
       t.Fatal("expected error for missing id, got nil")
   }
}