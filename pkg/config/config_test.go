package config

import (
   "fmt"
   "io/ioutil"
   "net/http"
   "os"
   "strings"
   "testing"
)

// TestLoadConfig_MissingEnv verifies that LoadConfig returns an error if required env vars are unset
func TestLoadConfig_MissingEnv(t *testing.T) {
   // Unset relevant environment variables
   os.Unsetenv("PINGONE_CLIENT_ID")
   os.Unsetenv("PINGONE_CLIENT_SECRET")
   os.Unsetenv("PINGONE_ENV_ID")

   _, err := LoadConfig()
   if err == nil {
       t.Fatal("expected error when env vars are missing, got nil")
   }
}
// TestFetchAccessToken_DiscoveryBadStatus simulates non-200 discovery response
func TestFetchAccessToken_DiscoveryBadStatus(t *testing.T) {
   origGet, origClient := httpGet, httpClient
   defer func() { httpGet, httpClient = origGet, origClient }()
   httpGet = func(url string) (*http.Response, error) {
       body := `error` + ""
       return &http.Response{StatusCode: http.StatusInternalServerError, Body: ioutil.NopCloser(strings.NewReader(body))}, nil
   }
   cfg := &Config{ClientID: "cid", ClientSecret: "cs", EnvironmentID: "env"}
   _, err := FetchAccessToken(cfg)
   if err == nil || !strings.Contains(err.Error(), "failed to fetch well-known configuration") {
       t.Errorf("expected discovery status error, got %v", err)
   }
}

// TestFetchAccessToken_MissingTokenEndpoint simulates discovery JSON without token_endpoint
func TestFetchAccessToken_MissingTokenEndpoint(t *testing.T) {
   origGet, origClient := httpGet, httpClient
   defer func() { httpGet, httpClient = origGet, origClient }()
   httpGet = func(url string) (*http.Response, error) {
       body := `{}`
       return &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader(body))}, nil
   }
   cfg := &Config{ClientID: "cid", ClientSecret: "cs", EnvironmentID: "env"}
   _, err := FetchAccessToken(cfg)
   if err == nil || !strings.Contains(err.Error(), "token_endpoint not found") {
       t.Errorf("expected missing token_endpoint error, got %v", err)
   }
}

// TestFetchAccessToken_TokenHTTPError simulates token endpoint HTTP error
func TestFetchAccessToken_TokenHTTPError(t *testing.T) {
   origGet, origClient := httpGet, httpClient
   defer func() { httpGet, httpClient = origGet, origClient }()
   httpGet = func(url string) (*http.Response, error) {
       return &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader(`{"token_endpoint":"https://tok"}`))}, nil
   }
   httpClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
       return &http.Response{StatusCode: http.StatusBadRequest, Body: ioutil.NopCloser(strings.NewReader("denied"))}, nil
   })}
   cfg := &Config{ClientID: "cid", ClientSecret: "cs", EnvironmentID: "env"}
   _, err := FetchAccessToken(cfg)
   if err == nil || !strings.Contains(err.Error(), "token endpoint returned") {
       t.Errorf("expected token endpoint error, got %v", err)
   }
}

// TestFetchAccessToken_MissingAccessToken simulates missing access_token in response
func TestFetchAccessToken_MissingAccessToken(t *testing.T) {
   origGet, origClient := httpGet, httpClient
   defer func() { httpGet, httpClient = origGet, origClient }()
   httpGet = func(url string) (*http.Response, error) {
       return &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader(`{"token_endpoint":"https://tok"}`))}, nil
   }
   httpClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
       return &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader(`{}`))}, nil
   })}
   cfg := &Config{ClientID: "cid", ClientSecret: "cs", EnvironmentID: "env"}
   _, err := FetchAccessToken(cfg)
   if err == nil || !strings.Contains(err.Error(), "access_token missing") {
       t.Errorf("expected missing access_token error, got %v", err)
   }
}
// roundTripperFunc is a helper to mock http.Client Transport
type roundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper
func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
   return f(req)
}

// TestFetchAccessToken_Success exercises the happy path with default region
func TestFetchAccessToken_Success(t *testing.T) {
   // Backup and restore HTTP functions
   origGet := httpGet
   origClient := httpClient
   defer func() { httpGet = origGet; httpClient = origClient }()

   // Mock discovery endpoint
   httpGet = func(url string) (*http.Response, error) {
       // Expect default region "com"
       if !strings.Contains(url, "auth.pingone.com/test-env") {
           t.Fatalf("unexpected discovery URL: %s", url)
       }
       body := `{"token_endpoint":"https://example.com/token"}`
       return &http.Response{
           StatusCode: http.StatusOK,
           Body:       ioutil.NopCloser(strings.NewReader(body)),
           Header:     make(http.Header),
       }, nil
   }
   // Mock token endpoint
   httpClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
       if req.URL.String() != "https://example.com/token" {
           return nil, fmt.Errorf("unexpected token URL: %s", req.URL)
       }
       // Check basic auth header present
       user, pass, ok := req.BasicAuth()
       if !ok || user != "cid" || pass != "cs" {
           return nil, fmt.Errorf("invalid basic auth: %s/%s", user, pass)
       }
       body := `{"access_token":"tok123"}`
       return &http.Response{
           StatusCode: http.StatusOK,
           Body:       ioutil.NopCloser(strings.NewReader(body)),
           Header:     make(http.Header),
       }, nil
   })}

   // Prepare config and environment
   os.Unsetenv("PINGONE_REGION")
   cfg := &Config{ClientID: "cid", ClientSecret: "cs", EnvironmentID: "test-env"}
   tok, err := FetchAccessToken(cfg)
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   if tok != "tok123" {
       t.Fatalf("expected token tok123, got %s", tok)
   }
}

// TestLoadConfig_Success verifies that LoadConfig loads vars correctly when they are set
func TestLoadConfig_Success(t *testing.T) {
   // Backup and restore
   origID, hadID := os.LookupEnv("PINGONE_CLIENT_ID")
   origSecret, hadSecret := os.LookupEnv("PINGONE_CLIENT_SECRET")
   origEnv, hadEnv := os.LookupEnv("PINGONE_ENV_ID")
   defer func() {
       if hadID {
           os.Setenv("PINGONE_CLIENT_ID", origID)
       } else {
           os.Unsetenv("PINGONE_CLIENT_ID")
       }
       if hadSecret {
           os.Setenv("PINGONE_CLIENT_SECRET", origSecret)
       } else {
           os.Unsetenv("PINGONE_CLIENT_SECRET")
       }
       if hadEnv {
           os.Setenv("PINGONE_ENV_ID", origEnv)
       } else {
           os.Unsetenv("PINGONE_ENV_ID")
       }
   }()

   // Set test values
   os.Setenv("PINGONE_CLIENT_ID", "test-id")
   os.Setenv("PINGONE_CLIENT_SECRET", "test-secret")
   os.Setenv("PINGONE_ENV_ID", "test-env")

   cfg, err := LoadConfig()
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   if cfg.ClientID != "test-id" {
       t.Errorf("expected ClientID=test-id, got %s", cfg.ClientID)
   }
   if cfg.ClientSecret != "test-secret" {
       t.Errorf("expected ClientSecret=test-secret, got %s", cfg.ClientSecret)
   }
   if cfg.EnvironmentID != "test-env" {
       t.Errorf("expected EnvironmentID=test-env, got %s", cfg.EnvironmentID)
   }
}