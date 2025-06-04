package tools

import (
   "context"
   "encoding/json"
   "fmt"
   "testing"

  "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeGetLicenseClient implements PingOneClient for testing GetLicenseTool
type fakeGetLicenseClient struct {
   stubClient
   expectedOrg    string
   expectedID     string
   returnedLicense management.License
}

// GetLicense checks inputs and returns a predefined license
func (f *fakeGetLicenseClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   if organizationID != f.expectedOrg {
       return nil, fmt.Errorf("unexpected organizationID: %s", organizationID)
   }
   if licenseID != f.expectedID {
       return nil, fmt.Errorf("unexpected licenseID: %s", licenseID)
   }
   data, _ := json.Marshal(f.returnedLicense)
   var m map[string]interface{}
   json.Unmarshal(data, &m)
   return m, nil
}

// stub other PingOneClient methods
func (f *fakeGetLicenseClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) { return management.User{}, nil }
func (f *fakeGetLicenseClient) DeleteUser(ctx context.Context, environmentID, userID string) error { return nil }
func (f *fakeGetLicenseClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) { return management.User{}, nil }
func (f *fakeGetLicenseClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) { return management.User{}, nil }
func (f *fakeGetLicenseClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error { return nil }
func (f *fakeGetLicenseClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error { return nil }
func (f *fakeGetLicenseClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) { return management.UserEnabled{}, nil }
func (f *fakeGetLicenseClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeGetLicenseClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeGetLicenseClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error { return nil }
func (f *fakeGetLicenseClient) CreatePopulation(ctx context.Context, environmentID string, pop management.Population) (management.Population, error) { return management.Population{}, nil }
func (f *fakeGetLicenseClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error { return nil }
func (f *fakeGetLicenseClient) CreateGroup(ctx context.Context, environmentID string, grp management.Group) (management.Group, error) { return management.Group{}, nil }
func (f *fakeGetLicenseClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error { return nil }
// UpdateGroup stub to satisfy PingOneClient
func (f *fakeGetLicenseClient) UpdateGroup(ctx context.Context, environmentID, groupID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
func (f *fakeGetLicenseClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeGetLicenseClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) { return nil, nil }
func (f *fakeGetLicenseClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) { return management.Environment{}, nil }
func (f *fakeGetLicenseClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) { return management.Environment{}, nil }
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeGetLicenseClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeGetLicenseClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}

func TestGetLicenseTool_Success(t *testing.T) {
   org := "org1"
   licID := "L123"
   name := "MyLicense"
   lic := management.NewLicense(name)
   lic.Id = &licID
   fake := &fakeGetLicenseClient{expectedOrg: org, expectedID: licID, returnedLicense: *lic}
   tool := NewGetLicenseTool(fake)

   args := map[string]interface{}{ "organization_id": org, "id": licID }
   out, err := tool.Run(context.Background(), args)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   if out["id"] != licID {
       t.Errorf("expected id %s, got %v", licID, out["id"])
   }
   if out["name"] != name {
       t.Errorf("expected name %s, got %v", name, out["name"])
   }
}

func TestGetLicenseTool_MissingArgs(t *testing.T) {
   fake := &fakeGetLicenseClient{}
   tool := NewGetLicenseTool(fake)
   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing organization_id, got nil")
   }
   if _, err := tool.Run(context.Background(), map[string]interface{}{ "organization_id": "org1" }); err == nil {
       t.Fatal("expected error for missing id, got nil")
   }
}