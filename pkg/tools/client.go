package tools

import (
   "context"
   "encoding/json"
   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// PingOneClient abstracts interactions with the PingOne Management API
type PingOneClient interface {
   // CreateUser creates a user in the specified environment
   CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error)
   // DeleteUser deletes a user in the specified environment
   DeleteUser(ctx context.Context, environmentID string, userID string) error
   // GetUser retrieves a user in the specified environment
   GetUser(ctx context.Context, environmentID string, userID string) (management.User, error)
  
   // UpdateUser applies a patch update to a user in the specified environment
   UpdateUser(ctx context.Context, environmentID string, userID string, user management.User) (management.User, error)
   // UnlockUserPassword invokes password recovery (unlock) for a user in the specified environment
   UnlockUserPassword(ctx context.Context, environmentID string, userID string) error
  
   // ResetUserPassword performs an administrative password reset (full set of new password) for a user in the specified environment
   ResetUserPassword(ctx context.Context, environmentID string, userID string, newPassword string) error
   // UpdateUserEnabled enables or disables a user in the specified environment
   UpdateUserEnabled(ctx context.Context, environmentID string, userID string, enabled bool) (management.UserEnabled, error)
  
   // GetUserPasswordState retrieves a user's password state in the specified environment
   GetUserPasswordState(ctx context.Context, environmentID string, userID string) (map[string]interface{}, error)
   // AddUserToGroup adds a user to a group in the specified environment
   AddUserToGroup(ctx context.Context, environmentID string, userID string, groupID string) (map[string]interface{}, error)
   // RemoveUserFromGroup removes a user from a group in the specified environment
   RemoveUserFromGroup(ctx context.Context, environmentID string, userID string, groupID string) error
   // CreatePopulation creates a new population in the specified environment
   CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error)
  
   // CreateGroup creates a new group in the specified environment
   CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error)
  
   // CreateEnvironment creates a new environment via the PingOne API
   CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error)
   // GetEnvironment retrieves an environment via the PingOne API
   GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error)
  // DeleteEnvironment deletes an environment via the PingOne API
  DeleteEnvironment(ctx context.Context, environmentID string) error
  // UpdateEnvironmentStatus updates an environment's status via the PingOne API
  UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error)
  
  
   // DeleteGroup deletes a group in the specified environment
   DeleteGroup(ctx context.Context, environmentID string, groupID string) error
  
   // GetGroup retrieves a group in the specified environment
   GetGroup(ctx context.Context, environmentID string, groupID string) (map[string]interface{}, error)
   // UpdateGroup updates a group in the specified environment
   UpdateGroup(ctx context.Context, environmentID string, groupID string, group management.Group) (management.Group, error)
  
   // GetLicense retrieves a license in the specified organization
   GetLicense(ctx context.Context, organizationID string, licenseID string) (map[string]interface{}, error)
  
   // DeletePopulation deletes a population in the specified environment
   DeletePopulation(ctx context.Context, environmentID string, populationID string) error
  
   // GetPopulation retrieves a population in the specified environment
   GetPopulation(ctx context.Context, environmentID string, populationID string) (map[string]interface{}, error)
}
// UpdateUserEnabled calls the SDK to enable or disable a user account
func (c *defaultClient) UpdateUserEnabled(ctx context.Context, environmentID string, userID string, enabled bool) (management.UserEnabled, error) {
   // Construct the payload
   ue := management.UserEnabled{Enabled: &enabled}
   updated, _, err := c.sdk.EnableUsersApi.UpdateUserEnabled(ctx, environmentID, userID).
       UserEnabled(ue).
       Execute()
   if err != nil {
       return management.UserEnabled{}, err
   }
   return *updated, nil
}
// UnlockUserPassword invokes password recovery (unlock) for a user in the specified environment
func (c *defaultClient) UnlockUserPassword(ctx context.Context, environmentID string, userID string) error {
   // Use the Password Locked Out endpoint for recovery
   _, err := c.sdk.UserPasswordsApi.EnvironmentsEnvironmentIDUsersUserIDPasswordPost(ctx, environmentID, userID).
       ContentType("application/vnd.pingidentity.password.recover+json").
       Body(map[string]interface{}{}).
       Execute()
   return err
}
// ResetUserPassword performs an administrative password reset (full set of new password)
func (c *defaultClient) ResetUserPassword(ctx context.Context, environmentID string, userID string, newPassword string) error {
   // Use the Password Set endpoint to fully set a new password with vendor media type
   // Send cleartext password + forceChange via PUT with Content-Type application/vnd.pingidentity.password.set+json
   _, err := c.sdk.UserPasswordsApi.EnvironmentsEnvironmentIDUsersUserIDPasswordPut(ctx, environmentID, userID).
       ContentType("application/vnd.pingidentity.password.set+json").
       Body(map[string]interface{}{ "value": newPassword, "forceChange": true }).
       Execute()
   return err
}

// GetUserPasswordState retrieves a user's password state via the PingOne API
func (c *defaultClient) GetUserPasswordState(ctx context.Context, environmentID string, userID string) (map[string]interface{}, error) {
   resp, err := c.sdk.UserPasswordsApi.EnvironmentsEnvironmentIDUsersUserIDPasswordGet(ctx, environmentID, userID).
       Execute()
   if err != nil {
       return nil, err
   }
   defer resp.Body.Close()
   var result map[string]interface{}
   if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
       return nil, err
   }
   return result, nil
}
// AddUserToGroup adds a user to a group via the PingOne API
func (c *defaultClient) AddUserToGroup(ctx context.Context, environmentID string, userID string, groupID string) (map[string]interface{}, error) {
   // Prepare request body
   // Use the generated GroupMembership model
   gm := management.NewGroupMembership(groupID)
   respObj, _, err := c.sdk.GroupMembershipApi.AddUserToGroup(ctx, environmentID, userID).
       GroupMembership(*gm).
       Execute()
   if err != nil {
       return nil, err
   }
   // Marshal response model to generic map
   var result map[string]interface{}
   data, err := json.Marshal(respObj)
   if err != nil {
       return nil, err
   }
   if err := json.Unmarshal(data, &result); err != nil {
       return nil, err
   }
   return result, nil
}
// RemoveUserFromGroup removes a user from a group via the PingOne API
func (c *defaultClient) RemoveUserFromGroup(ctx context.Context, environmentID string, userID string, groupID string) error {
   _, err := c.sdk.GroupMembershipApi.RemoveUserFromGroup(ctx, environmentID, userID, groupID).
       Execute()
   return err
}
// CreatePopulation creates a new population via the PingOne API
func (c *defaultClient) CreatePopulation(ctx context.Context, environmentID string, population management.Population) (management.Population, error) {
   created, _, err := c.sdk.PopulationsApi.CreatePopulation(ctx, environmentID).
       Population(population).
       Execute()
   if err != nil {
       return management.Population{}, err
   }
   return *created, nil
}
// DeletePopulation deletes a population via the PingOne API
func (c *defaultClient) DeletePopulation(ctx context.Context, environmentID string, populationID string) error {
   _, err := c.sdk.PopulationsApi.DeletePopulation(ctx, environmentID, populationID).
       Execute()
   return err
}
// CreateGroup creates a new group via the PingOne API
func (c *defaultClient) CreateGroup(ctx context.Context, environmentID string, group management.Group) (management.Group, error) {
   created, _, err := c.sdk.GroupsApi.CreateGroup(ctx, environmentID).
       Group(group).
       Execute()
   if err != nil {
       return management.Group{}, err
   }
   return *created, nil
}
// DeleteGroup deletes a group via the PingOne API
func (c *defaultClient) DeleteGroup(ctx context.Context, environmentID string, groupID string) error {
   _, err := c.sdk.GroupsApi.DeleteGroup(ctx, environmentID, groupID).
       Execute()
   return err
}
// UpdateGroup updates a group via the PingOne API
func (c *defaultClient) UpdateGroup(ctx context.Context, environmentID string, groupID string, group management.Group) (management.Group, error) {
   updated, _, err := c.sdk.GroupsApi.UpdateGroup(ctx, environmentID, groupID).
       Group(group).
       Execute()
   if err != nil {
       return management.Group{}, err
   }
   return *updated, nil
}
// GetGroup retrieves a group via the PingOne API
func (c *defaultClient) GetGroup(ctx context.Context, environmentID string, groupID string) (map[string]interface{}, error) {
   grp, _, err := c.sdk.GroupsApi.ReadOneGroup(ctx, environmentID, groupID).
       Execute()
   if err != nil {
       return nil, err
   }
   data, err := json.Marshal(grp)
   if err != nil {
       return nil, err
   }
   var result map[string]interface{}
   if err := json.Unmarshal(data, &result); err != nil {
       return nil, err
   }
   return result, nil
}
// GetPopulation retrieves a population via the PingOne API
func (c *defaultClient) GetPopulation(ctx context.Context, environmentID string, populationID string) (map[string]interface{}, error) {
   pop, _, err := c.sdk.PopulationsApi.ReadOnePopulation(ctx, environmentID, populationID).
       Execute()
   if err != nil {
       return nil, err
   }
   data, err := json.Marshal(pop)
   if err != nil {
       return nil, err
   }
   var result map[string]interface{}
   if err := json.Unmarshal(data, &result); err != nil {
       return nil, err
   }
   return result, nil
}
// GetLicense retrieves a license via the PingOne API
func (c *defaultClient) GetLicense(ctx context.Context, organizationID string, licenseID string) (map[string]interface{}, error) {
   lic, _, err := c.sdk.LicensesApi.ReadOneLicense(ctx, organizationID, licenseID).
       Execute()
   if err != nil {
       return nil, err
   }
   data, err := json.Marshal(lic)
   if err != nil {
       return nil, err
   }
   var result map[string]interface{}
   if err := json.Unmarshal(data, &result); err != nil {
       return nil, err
   }
   return result, nil
}

// defaultClient is the real implementation of PingOneClient using the SDK
type defaultClient struct {
   sdk *management.APIClient
}

// GetEnvironment retrieves an environment via the PingOne API
func (c *defaultClient) GetEnvironment(ctx context.Context, environmentID string) (management.Environment, error) {
    env, _, err := c.sdk.EnvironmentsApi.ReadOneEnvironment(ctx, environmentID).Execute()
    if err != nil {
        return management.Environment{}, err
    }
    return *env, nil
}
// DeleteEnvironment deletes an environment via the PingOne API
func (c *defaultClient) DeleteEnvironment(ctx context.Context, environmentID string) error {
    _, err := c.sdk.EnvironmentsApi.DeleteEnvironment(ctx, environmentID).Execute()
    return err
}

// NewPingOneClient returns a PingOneClient wrapping the given SDK client
func NewPingOneClient(sdk *management.APIClient) PingOneClient {
   return &defaultClient{sdk: sdk}
}

// CreateUser calls the SDK to create a new PingOne user
func (c *defaultClient) CreateUser(ctx context.Context, environmentID string, user management.User) (management.User, error) {
   created, _, err := c.sdk.UsersApi.CreateUser(ctx, environmentID).
       User(user).
       Execute()
   if err != nil {
       return management.User{}, err
   }
   return *created, nil
}
// DeleteUser calls the SDK to delete a PingOne user
func (c *defaultClient) DeleteUser(ctx context.Context, environmentID string, userID string) error {
   _, err := c.sdk.UsersApi.DeleteUser(ctx, environmentID, userID).
       Execute()
   return err
}
// GetUser calls the SDK to retrieve a PingOne user
func (c *defaultClient) GetUser(ctx context.Context, environmentID string, userID string) (management.User, error) {
   // Use ReadUser under the hood (SDK uses ReadUser for retrieving a single user)
   found, _, err := c.sdk.UsersApi.ReadUser(ctx, environmentID, userID).
       Execute()
   if err != nil {
       return management.User{}, err
   }
   return *found, nil
}
// CreateEnvironment creates a new environment via the PingOne API
func (c *defaultClient) CreateEnvironment(ctx context.Context, environment management.Environment) (management.Environment, error) {
    created, _, err := c.sdk.EnvironmentsApi.CreateEnvironmentActiveLicense(ctx).
        Environment(environment).
        Execute()
    if err != nil {
        return management.Environment{}, err
    }
    return *created, nil
}
// UpdateEnvironmentStatus updates an environment's status via the PingOne API
func (c *defaultClient) UpdateEnvironmentStatus(ctx context.Context, environmentID string, status management.EnumEnvironmentStatus) (management.Environment, error) {
    es := management.NewEnvironmentStatus(status)
    updated, _, err := c.sdk.EnvironmentsApi.UpdateEnvironmentStatus(ctx, environmentID).
        EnvironmentStatus(*es).
        Execute()
    if err != nil {
        return management.Environment{}, err
    }
    return *updated, nil
}
// UpdateUser calls the SDK to update a PingOne user via PATCH
func (c *defaultClient) UpdateUser(ctx context.Context, environmentID string, userID string, user management.User) (management.User, error) {
   updated, _, err := c.sdk.UsersApi.UpdateUserPatch(ctx, environmentID, userID).
       User(user).
       Execute()
   if err != nil {
       return management.User{}, err
   }
   return *updated, nil
}