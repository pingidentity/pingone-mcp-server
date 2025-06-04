package populations

import (
   "context"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeDeletePopulationClient implements PingOneClient for testing DeletePopulationTool
type fakeDeletePopulationClient struct {
   expectedEnv        string
   expectedPopulation string
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeDeletePopulationClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
   return nil, nil
}
// CreateGroup stub to satisfy PingOneClient
func (f *fakeDeletePopulationClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}

// DeletePopulation checks inputs and simulates deletion
func (f *fakeDeletePopulationClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   if environmentID != f.expectedEnv {
       return fmt.Errorf("unexpected environmentID: %s", environmentID)
   }
   if populationID != f.expectedPopulation {
       return fmt.Errorf("unexpected populationID: %s", populationID)
   }
   return nil
}
// DeleteGroup stub to satisfy PingOneClient
func (f *fakeDeletePopulationClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
    return nil
}
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeDeletePopulationClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeDeletePopulationClient) DeleteEnvironment(ctx context.Context, environmentID string) error {
   return nil
}
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeDeletePopulationClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}
// CreateEnvironment stub to satisfy PingOneClient
func (f *fakeDeletePopulationClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
    return management.Environment{}, nil
}
// GetEnvironment stub to satisfy PingOneClient
func (f *fakeDeletePopulationClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
    return management.Environment{}, nil
}
// GetGroup stub to satisfy PingOneClient
func (f *fakeDeletePopulationClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
// GetLicense stub to satisfy PingOneClient
func (f *fakeDeletePopulationClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}

// Stub other PingOneClient methods to satisfy interface
func (f *fakeDeletePopulationClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeDeletePopulationClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeDeletePopulationClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeDeletePopulationClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeDeletePopulationClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeDeletePopulationClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
func (f *fakeDeletePopulationClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
func (f *fakeDeletePopulationClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeDeletePopulationClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeDeletePopulationClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}
func (f *fakeDeletePopulationClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}

func TestDeletePopulationTool_Success(t *testing.T) {
   env := "env1"
   popID := "pop123"
   fake := &fakeDeletePopulationClient{expectedEnv: env, expectedPopulation: popID}
   tool := NewDeletePopulationTool(fake, "unused_env")

   args := map[string]interface{}{ "environment_id": env, "id": popID }
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   success, ok := out["success"].(bool)
   if !ok || !success {
       t.Errorf("expected success=true, got %v", out["success"])
   }
}

func TestDeletePopulationTool_MissingArgs(t *testing.T) {
   fake := &fakeDeletePopulationClient{}
   tool := NewDeletePopulationTool(fake, "")

   // Missing environment_id
   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }
   // Missing id
   if _, err := tool.Run(context.Background(), map[string]interface{}{ "environment_id": "env1" }); err == nil {
       t.Fatal("expected error for missing id, got nil")
   }
}