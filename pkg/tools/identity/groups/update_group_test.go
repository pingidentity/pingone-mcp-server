package groups

import (
   "context"
   "errors"
   "fmt"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// fakeUpdateGroupClient implements PingOneClient for testing UpdateGroupTool
type fakeUpdateGroupClient struct {
   expectedEnv   string
   expectedID    string
   expectedGroup management.Group
   returnedGroup management.Group
   err           error
}
// GetPopulation stub to satisfy PingOneClient
func (f *fakeUpdateGroupClient) GetPopulation(ctx context.Context, environmentID, populationID string) (map[string]interface{}, error) {
    return nil, nil
}
// DeleteEnvironment stub to satisfy PingOneClient
func (f *fakeUpdateGroupClient) DeleteEnvironment(ctx context.Context, environmentID string) error { return nil }
// UpdateEnvironmentStatus stub to satisfy PingOneClient
func (f *fakeUpdateGroupClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
   return management.Environment{}, nil
}

// UpdateGroup checks inputs and returns a predefined group or error
func (f *fakeUpdateGroupClient) UpdateGroup(ctx context.Context, environmentID, groupID string, grp management.Group) (management.Group, error) {
   if environmentID != f.expectedEnv {
       return management.Group{}, fmt.Errorf("unexpected environmentID: %s", environmentID)
   }
   if groupID != f.expectedID {
       return management.Group{}, fmt.Errorf("unexpected groupID: %s", groupID)
   }
   if f.expectedGroup.Name != "" {
       if grp.Name != f.expectedGroup.Name {
           return management.Group{}, fmt.Errorf("unexpected name: %v", grp.Name)
       }
   } else if grp.Name != "" {
       return management.Group{}, fmt.Errorf("unexpected name: %v", grp.Name)
   }
   if f.expectedGroup.Description != nil {
       if grp.Description == nil || *grp.Description != *f.expectedGroup.Description {
           return management.Group{}, fmt.Errorf("unexpected description: %v", grp.Description)
       }
   } else if grp.Description != nil {
       return management.Group{}, fmt.Errorf("unexpected description: %v", *grp.Description)
   }
   return f.returnedGroup, f.err
}

// Stub other PingOneClient methods to satisfy interface
func (f *fakeUpdateGroupClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeUpdateGroupClient) DeleteUser(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeUpdateGroupClient) GetUser(ctx context.Context, environmentID, userID string) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeUpdateGroupClient) UpdateUser(ctx context.Context, environmentID, userID string, user management.User) (management.User, error) {
   return management.User{}, nil
}
func (f *fakeUpdateGroupClient) UnlockUserPassword(ctx context.Context, environmentID, userID string) error {
   return nil
}
func (f *fakeUpdateGroupClient) ResetUserPassword(ctx context.Context, environmentID, userID, newPassword string) error {
   return nil
}
func (f *fakeUpdateGroupClient) UpdateUserEnabled(ctx context.Context, environmentID, userID string, enabled bool) (management.UserEnabled, error) {
   return management.UserEnabled{}, nil
}
func (f *fakeUpdateGroupClient) GetUserPasswordState(ctx context.Context, environmentID, userID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeUpdateGroupClient) AddUserToGroup(ctx context.Context, environmentID, userID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeUpdateGroupClient) RemoveUserFromGroup(ctx context.Context, environmentID, userID, groupID string) error {
   return nil
}
func (f *fakeUpdateGroupClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   return management.Population{}, nil
}
func (f *fakeUpdateGroupClient) DeletePopulation(ctx context.Context, environmentID, populationID string) error {
   return nil
}
func (f *fakeUpdateGroupClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   return management.Group{}, nil
}
func (f *fakeUpdateGroupClient) DeleteGroup(ctx context.Context, environmentID, groupID string) error {
   return nil
}
func (f *fakeUpdateGroupClient) GetGroup(ctx context.Context, environmentID, groupID string) (map[string]interface{}, error) {
   return nil, nil
}
func (f *fakeUpdateGroupClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
   return management.Environment{}, nil
}
func (f *fakeUpdateGroupClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
   return management.Environment{}, nil
}
func (f *fakeUpdateGroupClient) GetLicense(ctx context.Context, organizationID, licenseID string) (map[string]interface{}, error) {
   return nil, nil
}

func TestUpdateGroupTool_Success(t *testing.T) {
   envID := "env1"
   grpID := "grp1"
   name := "newName"
   desc := "newDesc"
   expected := management.Group{}
   expected.Name = name
   expected.Description = &desc
   returned := management.Group{}
   returned.Id = &grpID
   returned.Name = name
   returned.Description = &desc
   fake := &fakeUpdateGroupClient{expectedEnv: envID, expectedID: grpID, expectedGroup: expected, returnedGroup: returned, err: nil}
   tool := NewUpdateGroupTool(fake, envID)

   args := map[string]interface{}{
       "environment_id": envID,
       "id":             grpID,
       "name":           name,
       "description":    desc,
   }
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

func TestUpdateGroupTool_Error(t *testing.T) {
   envID := "env1"
   grpID := "grp1"
   fakeErr := errors.New("fail")
   fake := &fakeUpdateGroupClient{expectedEnv: envID, expectedID: grpID, err: fakeErr}
   tool := NewUpdateGroupTool(fake, envID)

   _, err := tool.Run(context.Background(), map[string]interface{}{
       "environment_id": envID,
       "id":             grpID,
   })
   if err == nil || err.Error() != fakeErr.Error() {
       t.Errorf("expected error %v, got %v", fakeErr, err)
   }
}

func TestUpdateGroupTool_MissingArgs(t *testing.T) {
   fake := &fakeUpdateGroupClient{}
   tool := NewUpdateGroupTool(fake, "")
   if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing environment_id, got nil")
   }
   if _, err := tool.Run(context.Background(), map[string]interface{}{"environment_id": "e1"}); err == nil {
       t.Fatal("expected error for missing id, got nil")
   }
}