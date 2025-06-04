package tools

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "net/http"
    "strings"
    "testing"

    "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// roundTripperFunc implements http.RoundTripper for testing
type roundTripperFunc func(req *http.Request) (*http.Response, error)
func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// makeTestClient returns a *defaultClient using a fake HTTP transport
func makeTestClient(t *testing.T, path string, payload []byte) *defaultClient {
   rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
        // allow /v1 prefix by matching suffix
        if !strings.HasSuffix(req.URL.Path, path) {
            t.Errorf("unexpected path %s; want suffix %s", req.URL.Path, path)
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

// TestCreateEnvironment verifies environment creation
func TestCreateEnvironment(t *testing.T) {
    path := "/environments"
    // payload with required fields
    payload := map[string]interface{}{
        "id":      "EID",
        "license": map[string]interface{}{"id": "L2"},
        "name":    "EnvName",
        "region":  "r1",
        "type":    "t1",
    }
    raw, _ := json.Marshal(payload)
    c := makeTestClient(t, path, raw)

    region := "r1"
    envModel := management.NewEnvironment(
        *management.NewEnvironmentLicense("L2"),
        "EnvName",
        management.StringAsEnvironmentRegion(&region),
        management.EnumEnvironmentType("t1"),
    )
    envModel.Id = nil
    out, err := c.CreateEnvironment(context.Background(), *envModel)
    if err != nil {
        t.Fatal("CreateEnvironment error:", err)
    }
    if out.Name != "EnvName" {
        t.Errorf("expected name EnvName, got %q", out.Name)
    }
    if out.License.Id != "L2" {
        t.Errorf("expected license.id L2, got %s", out.License.Id)
    }
    if out.Id == nil || *out.Id != "EID" {
        t.Errorf("expected id EID, got %v", out.Id)
    }
}

// TestGetEnvironment verifies retrieval
func TestGetEnvironment(t *testing.T) {
    id := "E2"
    path := "/environments/" + id
    payload := map[string]interface{}{"id": id, "name": "Env2", "license": map[string]interface{}{"id": "L3"}, "region": "r2", "type": "t2"}
    raw, _ := json.Marshal(payload)
    c := makeTestClient(t, path, raw)

    out, err := c.GetEnvironment(context.Background(), id)
    if err != nil {
        t.Fatal("GetEnvironment error:", err)
    }
    if out.Id == nil || *out.Id != id {
        t.Errorf("expected id %s, got %v", id, out.Id)
    }
    if out.Name != "Env2" {
        t.Errorf("expected name Env2, got %q", out.Name)
    }
    if out.License.Id != "L3" {
        t.Errorf("expected license.id L3, got %s", out.License.Id)
    }
}

// TestDeleteEnvironment verifies deletion succeeds
func TestDeleteEnvironment(t *testing.T) {
    id := "E3"
    path := "/environments/" + id
    c := makeTestClient(t, path, []byte(`{}`))
    if err := c.DeleteEnvironment(context.Background(), id); err != nil {
        t.Fatal("DeleteEnvironment error:", err)
    }
}

// TestUpdateEnvironmentStatus verifies status update
func TestUpdateEnvironmentStatus(t *testing.T) {
    id := "E4"
    path := "/environments/" + id + "/status"
    payload := map[string]interface{}{"id": id, "status": "ACTIVE"}
    raw, _ := json.Marshal(payload)
    c := makeTestClient(t, path, raw)

    out, err := c.UpdateEnvironmentStatus(context.Background(), id, management.EnumEnvironmentStatus("ACTIVE"))
    if err != nil {
        t.Fatal("UpdateEnvironmentStatus error:", err)
    }
    if out.Id == nil || *out.Id != id {
        t.Errorf("expected id %s, got %v", id, out.Id)
    }
    if out.Status == nil || string(*out.Status) != "ACTIVE" {
        t.Errorf("expected status ACTIVE, got %v", out.Status)
    }
}
// CRUD tests for users
func TestCreateUser(t *testing.T) {
    env := "envA"
    path := "/environments/" + env + "/users"
    id := "U100"
    u := management.NewUser("e@x.com", "user1")
    u.Id = &id
    raw, _ := json.Marshal(u)
    c := makeTestClient(t, path, raw)

    out, err := c.CreateUser(context.Background(), env, *u)
    if err != nil {
        t.Fatal("CreateUser error:", err)
    }
    if out.Email != u.Email {
        t.Errorf("expected email %q, got %q", u.Email, out.Email)
    }
    if out.Username != u.Username {
        t.Errorf("expected username %q, got %q", u.Username, out.Username)
    }
    if out.Id == nil || *out.Id != id {
        t.Errorf("expected id %q, got %v", id, out.Id)
    }
}

func TestGetUser(t *testing.T) {
    env := "envB"
    uid := "U200"
    path := "/environments/" + env + "/users/" + uid
    u := management.NewUser("g@x.com", "user2")
    u.Id = &uid
    raw, _ := json.Marshal(u)
    c := makeTestClient(t, path, raw)

    out, err := c.GetUser(context.Background(), env, uid)
    if err != nil {
        t.Fatal("GetUser error:", err)
    }
    if out.Email != u.Email {
        t.Errorf("expected email %q, got %q", u.Email, out.Email)
    }
    if out.Username != u.Username {
        t.Errorf("expected username %q, got %q", u.Username, out.Username)
    }
    if out.Id == nil || *out.Id != uid {
        t.Errorf("expected id %q, got %v", uid, out.Id)
    }
}

func TestUpdateUser(t *testing.T) {
    env := "envC"
    uid := "U300"
    path := "/environments/" + env + "/users/" + uid
    u := management.NewUser("upd@x.com", "user3")
    u.Id = &uid
    raw, _ := json.Marshal(u)
    c := makeTestClient(t, path, raw)

    out, err := c.UpdateUser(context.Background(), env, uid, *u)
    if err != nil {
        t.Fatal("UpdateUser error:", err)
    }
    if out.Email != u.Email {
        t.Errorf("expected email %q, got %q", u.Email, out.Email)
    }
    if out.Username != u.Username {
        t.Errorf("expected username %q, got %q", u.Username, out.Username)
    }
    if out.Id == nil || *out.Id != uid {
        t.Errorf("expected id %q, got %v", uid, out.Id)
    }
}

func TestDeleteUser(t *testing.T) {
    env := "envD"
    uid := "U400"
    path := "/environments/" + env + "/users/" + uid
    raw := []byte(`{}`)
    c := makeTestClient(t, path, raw)

    err := c.DeleteUser(context.Background(), env, uid)
    if err != nil {
        t.Fatal("DeleteUser error:", err)
    }
}

func TestGetUserPasswordState(t *testing.T) {
    // prepare a fake JSON body
    body := map[string]interface{}{ "locked": false, "lastChanged": "2025-01-01T00:00:00Z" }
    raw, _ := json.Marshal(body)
    // our SDK call path for GetUserPasswordState is
    // /environments/{env}/users/{userID}/password
    env, uid := "env123", "user456"
    path := "/environments/" + env + "/users/" + uid + "/password"
    c := makeTestClient(t, path, raw)

    got, err := c.GetUserPasswordState(context.Background(), env, uid)
    if err != nil {
        t.Fatal("unexpected error:", err)
    }
    if got["locked"] != false {
        t.Errorf("expected locked=false, got %v", got["locked"])
    }
    if got["lastChanged"] != "2025-01-01T00:00:00Z" {
        t.Errorf("expected lastChanged, got %v", got["lastChanged"])
    }
}

func TestAddUserToGroup(t *testing.T) {
    // fake GroupMembership JSON: only id field is required
    gm := management.NewGroupMembership("grp789")
    rawObj, _ := json.Marshal(gm)
    // endpoint path
    env, uid := "E1", "U1"
    path := "/environments/" + env + "/users/" + uid + "/memberOfGroups"
    c := makeTestClient(t, path, rawObj)

    out, err := c.AddUserToGroup(context.Background(), env, uid, "grp789")
    if err != nil {
        t.Fatal("unexpected error:", err)
    }
    if out["id"] != "grp789" {
        t.Errorf("expected id=grp789, got %v", out["id"])
    }
}

func TestGetLicense(t *testing.T) {
    // fake license as a generic map (License has fields id and name)
    name := "enterprise"
    lic := map[string]interface{}{ "id": "L42", "name": name }
    raw, _ := json.Marshal(lic)
    // /organizations/{org}/licenses/{licID}
    org, lid := "ORG1", "L42"
    path := "/organizations/" + org + "/licenses/" + lid
    c := makeTestClient(t, path, raw)

    m, err := c.GetLicense(context.Background(), org, lid)
    if err != nil {
        t.Fatal(err)
    }
    if m["id"] != "L42" {
        t.Errorf("expected id L42, got %v", m["id"])
    }
    if m["name"] != name {
        t.Errorf("expected name %q, got %v", name, m["name"])
    }
}

// TestUnlockUserPassword verifies the unlock call succeeds
func TestUnlockUserPassword(t *testing.T) {
    env, uid := "envX", "UX"
    path := "/environments/" + env + "/users/" + uid + "/password"
    c := makeTestClient(t, path, []byte(`{}`))
    if err := c.UnlockUserPassword(context.Background(), env, uid); err != nil {
        t.Fatal("UnlockUserPassword error:", err)
    }
}

// TestResetUserPassword verifies the reset call succeeds
func TestResetUserPassword(t *testing.T) {
    env, uid := "envY", "UY"
    path := "/environments/" + env + "/users/" + uid + "/password"
    c := makeTestClient(t, path, []byte(`{}`))
    if err := c.ResetUserPassword(context.Background(), env, uid, "newpass"); err != nil {
        t.Fatal("ResetUserPassword error:", err)
    }
}

// TestUpdateUserEnabled verifies enabling/disabling a user
func TestUpdateUserEnabled(t *testing.T) {
    env, uid := "envZ", "UZ"
    path := "/environments/" + env + "/users/" + uid + "/enabled"
    // return JSON {"enabled":false}
    c := makeTestClient(t, path, []byte(`{"enabled":false}`))
    ue, err := c.UpdateUserEnabled(context.Background(), env, uid, false)
    if err != nil {
        t.Fatal("UpdateUserEnabled error:", err)
    }
    if ue.Enabled == nil || *ue.Enabled != false {
        t.Errorf("expected enabled=false, got %v", ue.Enabled)
    }
}

// TestRemoveUserFromGroup verifies removal succeeds
func TestRemoveUserFromGroup(t *testing.T) {
    env, uid, gid := "envG", "UG", "GID"
    path := "/environments/" + env + "/users/" + uid + "/memberOfGroups/" + gid
    c := makeTestClient(t, path, []byte(`{}`))
    if err := c.RemoveUserFromGroup(context.Background(), env, uid, gid); err != nil {
        t.Fatal("RemoveUserFromGroup error:", err)
    }
}

// TestCreatePopulation verifies population creation
func TestCreatePopulation(t *testing.T) {
    env := "envP"
    name := "popA"
    path := "/environments/" + env + "/populations"
    pop := management.NewPopulation(name)
    id := "P1"
    pop.Id = &id
    raw, _ := json.Marshal(pop)
    c := makeTestClient(t, path, raw)
    got, err := c.CreatePopulation(context.Background(), env, *pop)
    if err != nil {
        t.Fatal("CreatePopulation error:", err)
    }
    if got.Name != pop.Name {
        t.Errorf("expected name %q, got %q", pop.Name, got.Name)
    }
    if got.Id == nil || *got.Id != id {
        t.Errorf("expected id %q, got %v", id, got.Id)
    }
}

// TestDeletePopulation verifies deletion succeeds
func TestDeletePopulation(t *testing.T) {
    env := "envDP"
    pid := "DP1"
    path := "/environments/" + env + "/populations/" + pid
    c := makeTestClient(t, path, []byte(`{}`))
    if err := c.DeletePopulation(context.Background(), env, pid); err != nil {
        t.Fatal("DeletePopulation error:", err)
    }
}

// TestCreateGroup verifies group creation
func TestCreateGroup(t *testing.T) {
    env := "envCG"
    name := "groupX"
    path := "/environments/" + env + "/groups"
    grp := management.NewGroup(name)
    id := "GX1"
    grp.Id = &id
    raw, _ := json.Marshal(grp)
    c := makeTestClient(t, path, raw)
    got, err := c.CreateGroup(context.Background(), env, *grp)
    if err != nil {
        t.Fatal("CreateGroup error:", err)
    }
    if got.Name != name {
        t.Errorf("expected name %q, got %q", name, got.Name)
    }
    if got.Id == nil || *got.Id != id {
        t.Errorf("expected id %q, got %v", id, got.Id)
    }
}

// TestDeleteGroup verifies deletion succeeds
func TestDeleteGroup(t *testing.T) {
    env := "envDG"
    gid := "DG1"
    path := "/environments/" + env + "/groups/" + gid
    c := makeTestClient(t, path, []byte(`{}`))
    if err := c.DeleteGroup(context.Background(), env, gid); err != nil {
        t.Fatal("DeleteGroup error:", err)
    }
}

// TestUpdateGroup verifies update succeeds
func TestUpdateGroup(t *testing.T) {
    env := "envUG"
    gid := "UG1"
    path := "/environments/" + env + "/groups/" + gid
    name := "groupY"
    grp := management.NewGroup(name)
    grp.Id = &gid
    raw, _ := json.Marshal(grp)
    c := makeTestClient(t, path, raw)
    got, err := c.UpdateGroup(context.Background(), env, gid, *grp)
    if err != nil {
        t.Fatal("UpdateGroup error:", err)
    }
    if got.Name != name {
        t.Errorf("expected name %q, got %q", name, got.Name)
    }
}

// TestGetGroupClient verifies GetGroup returns a map
func TestGetGroupClient(t *testing.T) {
    env := "envGG"
    gid := "GG1"
    path := "/environments/" + env + "/groups/" + gid
    name := "groupZ"
    grp := management.NewGroup(name)
    grp.Id = &gid
    raw, _ := json.Marshal(grp)
    c := makeTestClient(t, path, raw)
    out, err := c.GetGroup(context.Background(), env, gid)
    if err != nil {
        t.Fatal("GetGroup error:", err)
    }
    if out["id"] != gid || out["name"] != name {
        t.Errorf("unexpected map: %v", out)
    }
}

// TestGetPopulationClient verifies GetPopulation returns a map
func TestGetPopulationClient(t *testing.T) {
    env := "envGP"
    pid := "GP1"
    path := "/environments/" + env + "/populations/" + pid
    name := "popZ"
    pop := management.NewPopulation(name)
    pop.Id = &pid
    raw, _ := json.Marshal(pop)
    c := makeTestClient(t, path, raw)
    out, err := c.GetPopulation(context.Background(), env, pid)
    if err != nil {
        t.Fatal("GetPopulation error:", err)
    }
    if out["id"] != pid || out["name"] != name {
        t.Errorf("unexpected map: %v", out)
    }
}