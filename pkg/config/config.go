package config

import (
   "encoding/json"
   "fmt"
   "io/ioutil"
   "net/http"
   "net/url"
   "os"
   "strings"
)

// httpGet and httpClient are overridden in tests to simulate HTTP behavior
var (
   httpGet    = http.Get
   httpClient = http.DefaultClient
)

// Config holds PingOne credentials and environment ID
type Config struct {
   // ClientID is the PingOne OAuth client ID
   ClientID string
   // ClientSecret is the PingOne OAuth client secret
   ClientSecret string
   // EnvironmentID is the PingOne environment identifier
   EnvironmentID string
}

// LoadConfig reads required PingOne variables from environment and returns a Config or error
func LoadConfig() (*Config, error) {
   clientID := os.Getenv("PINGONE_CLIENT_ID")
   if clientID == "" {
       return nil, fmt.Errorf("missing environment variable PINGONE_CLIENT_ID")
   }
   clientSecret := os.Getenv("PINGONE_CLIENT_SECRET")
   if clientSecret == "" {
       return nil, fmt.Errorf("missing environment variable PINGONE_CLIENT_SECRET")
   }
   envID := os.Getenv("PINGONE_ENV_ID")
   if envID == "" {
       return nil, fmt.Errorf("missing environment variable PINGONE_ENV_ID")
   }
   return &Config{
       ClientID:     clientID,
       ClientSecret: clientSecret,
       EnvironmentID: envID,
   }, nil
}
// FetchAccessToken discovers the token endpoint via the OpenID Connect discovery
// and obtains an access token using the client credentials flow.
func FetchAccessToken(cfg *Config) (string, error) {
   // Determine region (default to "com")
   region := os.Getenv("PINGONE_REGION")
   if region == "" {
       region = "com"
   }
   // Construct .well-known URL
   wellKnownURL := fmt.Sprintf(
       "https://auth.pingone.%s/%s/as/.well-known/openid-configuration",
       region, cfg.EnvironmentID,
   )
   // Fetch discovery document
   // Fetch discovery document
   resp, err := httpGet(wellKnownURL)
   if err != nil {
       return "", fmt.Errorf("failed to fetch well-known configuration from %s: %w", wellKnownURL, err)
   }
   defer resp.Body.Close()
   if resp.StatusCode != http.StatusOK {
       body, _ := ioutil.ReadAll(resp.Body)
       return "", fmt.Errorf(
           "failed to fetch well-known configuration: %s %s",
           resp.Status, strings.TrimSpace(string(body)),
       )
   }
   // Parse discovery document
   var disc struct {
       TokenEndpoint string `json:"token_endpoint"`
   }
   if err := json.NewDecoder(resp.Body).Decode(&disc); err != nil {
       return "", fmt.Errorf("failed to parse well-known configuration: %w", err)
   }
   if disc.TokenEndpoint == "" {
       return "", fmt.Errorf("token_endpoint not found in well-known configuration")
   }
   // Prepare client credentials request using HTTP Basic auth
   form := url.Values{}
   form.Set("grant_type", "client_credentials")
   // Build HTTP request
   req, err := http.NewRequest("POST", disc.TokenEndpoint, strings.NewReader(form.Encode()))
   if err != nil {
       return "", fmt.Errorf("failed to create token request: %w", err)
   }
   req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
   req.SetBasicAuth(cfg.ClientID, cfg.ClientSecret)
   // Execute token request
   tokenResp, err := httpClient.Do(req)
   if err != nil {
       return "", fmt.Errorf("failed to request access token: %w", err)
   }
   defer tokenResp.Body.Close()
   if tokenResp.StatusCode != http.StatusOK {
       body, _ := ioutil.ReadAll(tokenResp.Body)
       return "", fmt.Errorf(
           "token endpoint returned %s: %s",
           tokenResp.Status, strings.TrimSpace(string(body)),
       )
   }
   // Decode token response
   var tr struct {
       AccessToken string `json:"access_token"`
   }
   if err := json.NewDecoder(tokenResp.Body).Decode(&tr); err != nil {
       return "", fmt.Errorf("failed to parse token response: %w", err)
   }
   if tr.AccessToken == "" {
       return "", fmt.Errorf("access_token missing in token response")
   }
   return tr.AccessToken, nil
}