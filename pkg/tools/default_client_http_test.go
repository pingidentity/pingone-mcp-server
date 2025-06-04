package tools

import (
   "bytes"
   "context"
   "io"
   "net/http"
   "strings"
   "testing"

   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// defaultRoundTripperFunc mocks HTTP transport for default client tests
type defaultRoundTripperFunc func(req *http.Request) (*http.Response, error)
func (f defaultRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// makeClient returns a *defaultClient with all HTTP calls stubbed to return JSON payload
func makeClient(t *testing.T, suffix string, payload []byte) *defaultClient {
   rt := defaultRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
       path := req.URL.Path
       trimmed := strings.TrimPrefix(path, "/v1")
       if !strings.HasPrefix(trimmed, suffix) {
           t.Errorf("unexpected path %s; want prefix %s", req.URL.Path, suffix)
       }
       return &http.Response{
           StatusCode: http.StatusOK,
           Header:     http.Header{"Content-Type": {"application/json"}},
           Body:       io.NopCloser(bytes.NewReader(payload)),
       }, nil
   })
   cfg := management.NewConfiguration()
   cfg.HTTPClient = &http.Client{Transport: rt}
   sdk := management.NewAPIClient(cfg)
   return NewPingOneClient(sdk).(*defaultClient)
}

// TestDefaultClientExercises all wrapper methods for coverage
func TestDefaultClientExercises(t *testing.T) {
   ctx := context.Background()
   // 1. User CRUD + password
   payloadUser := []byte(`{"email":"","username":""}`)
   c := makeClient(t, "/environments/e/users", payloadUser)
   _, _ = c.CreateUser(ctx, "e", management.User{})
   _, _ = c.GetUser(ctx, "e", "u")
   _, _ = c.UpdateUser(ctx, "e", "u", management.User{})
   _ = c.DeleteUser(ctx, "e", "u")
   _ = c.UnlockUserPassword(ctx, "e", "u")
   _ = c.ResetUserPassword(ctx, "e", "u", "pw")
   _, _ = c.UpdateUserEnabled(ctx, "e", "u", true)
   _, _ = c.GetUserPasswordState(ctx, "e", "u")

   // 2. Group membership
   payloadGM := []byte(`{"id":"g"}`)
   c = makeClient(t, "/environments/e/users/u/memberOfGroups", payloadGM)
   _, _ = c.AddUserToGroup(ctx, "e", "u", "g")
   c = makeClient(t, "/environments/e/users/u/memberOfGroups/g", []byte(`{}`))
   _ = c.RemoveUserFromGroup(ctx, "e", "u", "g")

   // 3. Populations
   payloadPop := []byte(`{"name":"p"}`)
   c = makeClient(t, "/environments/e/populations", payloadPop)
   _, _ = c.CreatePopulation(ctx, "e", management.Population{Name: "p"})
   c = makeClient(t, "/environments/e/populations/p", []byte(`{}`))
   _ = c.DeletePopulation(ctx, "e", "p")

   // 4. Groups
   payloadGrp := []byte(`{"name":"g"}`)
   c = makeClient(t, "/environments/e/groups", payloadGrp)
   _, _ = c.CreateGroup(ctx, "e", management.Group{Name: "g"})
   c = makeClient(t, "/environments/e/groups/g", []byte(`{}`))
   _ = c.DeleteGroup(ctx, "e", "g")
   _, _ = c.UpdateGroup(ctx, "e", "g", management.Group{Name: "g"})
   _, _ = c.GetGroup(ctx, "e", "g")

   // 5. Licenses
   payloadLic := []byte(`{"id":"l"}`)
   c = makeClient(t, "/organizations/o/licenses/l", payloadLic)
   _, _ = c.GetLicense(ctx, "o", "l")

   // 6. Environment lifecycle
   // Create
   payloadEnv := []byte(`{"id":"E","license":{"id":"L"},"name":"N","region":"r","type":"t"}`)
   c = makeClient(t, "/environments", payloadEnv)
   _, _ = c.CreateEnvironment(ctx, management.Environment{Id: defaultPtrString("E")})
   // Get
   c = makeClient(t, "/environments/E", payloadEnv)
   _, _ = c.GetEnvironment(ctx, "E")
   // Delete
   c = makeClient(t, "/environments/E", []byte(`{}`))
   _ = c.DeleteEnvironment(ctx, "E")
   // Update status
   payloadStatus := []byte(`{"id":"E","status":"ACTIVE"}`)
   c = makeClient(t, "/environments/E/status", payloadStatus)
   _, _ = c.UpdateEnvironmentStatus(ctx, "E", management.EnumEnvironmentStatus("ACTIVE"))
}

// defaultPtrString helper
func defaultPtrString(s string) *string { return &s }