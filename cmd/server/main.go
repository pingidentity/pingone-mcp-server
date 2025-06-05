package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pingidentity/pingone-mcp-server/pkg/config"
	"github.com/pingidentity/pingone-mcp-server/pkg/mcp"
	"github.com/pingidentity/pingone-mcp-server/pkg/tools"
	// identity user search tool
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/pkg/tools/environments"
	"github.com/pingidentity/pingone-mcp-server/pkg/tools/environments/populations"
	"github.com/pingidentity/pingone-mcp-server/pkg/tools/identity/groups"
	"github.com/pingidentity/pingone-mcp-server/pkg/tools/identity/membership"
	"github.com/pingidentity/pingone-mcp-server/pkg/tools/identity/users"
)

// tokenRefreshingTransport wraps an HTTP RoundTripper and on 401 responses
// will automatically fetch a new access token and retry the request.
type tokenRefreshingTransport struct {
	config *config.Config
	base   http.RoundTripper
	mu     sync.Mutex
	token  string
}

// RoundTrip implements http.RoundTripper.
func (t *tokenRefreshingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Prepare request with current token
	r1 := req.Clone(req.Context())
	t.mu.Lock()
	r1.Header.Set("Authorization", "Bearer "+t.token)
	t.mu.Unlock()
	// First attempt
	resp, err := t.base.RoundTrip(r1)
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}
	// On 401, close and discard body
	resp.Body.Close()
	// Refresh token
	start := time.Now()
	newToken, err := config.FetchAccessToken(t.config)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("MCP get token status=500 elapsed=%s", elapsed)
		return resp, err
	}
	log.Printf("MCP get token status=200 elapsed=%s", elapsed)
	// Store new token
	t.mu.Lock()
	t.token = newToken
	t.mu.Unlock()
	// Retry original request with new token
	r2 := req.Clone(req.Context())
	r2.Header.Set("Authorization", "Bearer "+newToken)
	return t.base.RoundTrip(r2)
}

// loggingTransport wraps HTTP requests to dump requests and responses when debugging.
type loggingTransport struct {
	base http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if dump, err := httputil.DumpRequestOut(req, true); err == nil {
		fmt.Printf(">>> Request >>>\n%s\n", dump)
	} else {
		log.Printf("Error dumping request: %v", err)
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		log.Printf("Error in RoundTrip: %v", err)
		return resp, err
	}
	if dump, err := httputil.DumpResponse(resp, true); err == nil {
		fmt.Printf("<<< Response <<<\n%s\n", dump)
	} else {
		log.Printf("Error dumping response: %v", err)
	}
	return resp, nil
}

// getEnvString returns the environment variable value or the default value if not set
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns the environment variable value as a boolean or the default value if not set
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// setEnvFromFlag sets an environment variable from a flag value if env var doesn't exist
func setEnvFromFlag(key string, value interface{}) {
	if os.Getenv(key) != "" {
		return // Don't override existing environment variables
	}
	
	switch v := value.(type) {
	case bool:
		os.Setenv(key, strconv.FormatBool(v))
	case string:
		if v != "" {
			os.Setenv(key, v)
		}
	}
}

// getAPIKey loads the API key from a file or environment, or generates and persists a new one.
func getAPIKey() string {
	// Use current environment variable (which may have been set by flag)
	keyPath := os.Getenv("PINGONE_MCP_API_KEY_PATH")
	// Try reading existing file
	if data, err := os.ReadFile(keyPath); err == nil {
		key := strings.TrimSpace(string(data))
		if key != "" {
			return key
		}
	}
	// Fallback to environment variable
	if key := os.Getenv("PINGONE_MCP_API_KEY"); key != "" {
		// Persist to file for next runs
		_ = os.WriteFile(keyPath, []byte(key+"\n"), 0600)
		return key
	}
	// Generate new key
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("failed to generate API key: %v", err)
	}
	key := hex.EncodeToString(b)
	if err := os.WriteFile(keyPath, []byte(key+"\n"), 0600); err != nil {
		log.Printf("warning: unable to write API key to %s: %v", keyPath, err)
	}
	return key
}

func main() {
	// Server version
	const serverVersion = "0.1.0"
	
	// Parse command-line flags with environment variable fallbacks
	debugAPI := flag.Bool("debug-api", getEnvBool("PINGONE_MCP_DEBUG_API", false), "log API requests and responses to PingOne")
	transportMode := flag.String("transport", getEnvString("PINGONE_MCP_TRANSPORT", "stdio"), "transport mode: 'stdio' for Claude Desktop, 'http' for REST API (default: stdio)")
	allowMutation := flag.Bool("allow-mutation", getEnvBool("PINGONE_MCP_ALLOW_MUTATION", false), "enable mutation tools (create, update, delete operations)")
	allowInsecure := flag.Bool("allow-insecure", getEnvBool("PINGONE_MCP_ALLOW_INSECURE", false), "disable API key requirement for HTTP mode")
	serverPort := flag.String("server-port", getEnvString("PINGONE_MCP_SERVER_PORT", "8080"), "HTTP server port (HTTP mode only)")
	apiKeyPath := flag.String("api-key-path", getEnvString("PINGONE_MCP_API_KEY_PATH", "pingone-mcp-server-api.key"), "path to API key file (HTTP mode only)")
	
	// PingOne configuration flags
	clientID := flag.String("client-id", getEnvString("PINGONE_CLIENT_ID", ""), "PingOne OAuth client ID")
	clientSecret := flag.String("client-secret", getEnvString("PINGONE_CLIENT_SECRET", ""), "PingOne OAuth client secret")
	envID := flag.String("env-id", getEnvString("PINGONE_ENV_ID", ""), "PingOne environment ID")
	region := flag.String("region", getEnvString("PINGONE_REGION", "com"), "PingOne region (com, eu, ca, asia)")
	
	flag.Parse()

	// Override environment variables with command-line values if provided
	setEnvFromFlag("PINGONE_MCP_DEBUG_API", *debugAPI)
	setEnvFromFlag("PINGONE_MCP_TRANSPORT", *transportMode)
	setEnvFromFlag("PINGONE_MCP_ALLOW_MUTATION", *allowMutation)
	setEnvFromFlag("PINGONE_MCP_ALLOW_INSECURE", *allowInsecure)
	setEnvFromFlag("PINGONE_MCP_SERVER_PORT", *serverPort)
	setEnvFromFlag("PINGONE_MCP_API_KEY_PATH", *apiKeyPath)
	setEnvFromFlag("PINGONE_CLIENT_ID", *clientID)
	setEnvFromFlag("PINGONE_CLIENT_SECRET", *clientSecret)
	setEnvFromFlag("PINGONE_ENV_ID", *envID)
	setEnvFromFlag("PINGONE_REGION", *region)
	// Load configuration - handle gracefully for MCP mode
	cfg, err := config.LoadConfig()
	var token string
	var client tools.PingOneClient
	
	if err != nil {
		if *transportMode == "stdio" {
			// In stdio mode, log the error but continue - tools will return proper MCP errors
			log.Printf("Warning: PingOne configuration not available: %v", err)
			log.Printf("Tools will return appropriate error messages when called")
			// Create a nil client - tools should handle this gracefully
			client = nil
		} else {
			log.Fatalf("failed to load config: %v", err)
		}
	} else {
		// Try to obtain access token
		token, err = config.FetchAccessToken(cfg)
		if err != nil {
			if *transportMode == "stdio" {
				log.Printf("Warning: Failed to obtain PingOne access token: %v", err)
				log.Printf("Tools will return appropriate error messages when called")
				client = nil
			} else {
				log.Fatalf("failed to obtain access token: %v", err)
			}
		}
	}
	// Initialize PingOne SDK client if we have valid configuration
	if client == nil && cfg != nil && token != "" {
		sdkCfg := management.NewConfiguration()
		// Build base transport with token refreshing
		baseTransport := &tokenRefreshingTransport{config: cfg, base: http.DefaultTransport, token: token}
		// If debug API is enabled, wrap in logging transport
		var transport http.RoundTripper = baseTransport
		if *debugAPI {
			transport = &loggingTransport{base: transport}
		}
		// Use custom HTTP client
		sdkCfg.HTTPClient = &http.Client{Transport: transport}
		sdkClient := management.NewAPIClient(sdkCfg)
		// Wrap SDK with our client interface
		client = tools.NewPingOneClient(sdkClient)
	}
	// Configuration reporting
	if !*allowMutation {
		log.Printf("Mutation tools disabled (read-only mode). Set PINGONE_MCP_ALLOW_MUTATION=true to enable.")
	}

	if !*allowInsecure {
		log.Printf("API key required (secure mode). Set PINGONE_MCP_ALLOW_INSECURE=true to disable API key requirement.")
	}

	// Register tools only if we have a valid client and configuration
	if client != nil && cfg != nil {
		// For now, skip SDK-dependent tools in favor of simpler approach
		// In a future version, we could expose GetSDKClient() method on the interface

		// Read-only tools (always registered)
		tools.Register(users.NewGetUserTool(client, cfg.EnvironmentID))
		tools.Register(users.NewGetUserPasswordStateTool(client, cfg.EnvironmentID))
		tools.Register(populations.NewGetPopulationTool(client, cfg.EnvironmentID))
		tools.Register(groups.NewGetGroupTool(client, cfg.EnvironmentID))
		tools.Register(environments.NewGetEnvironmentTool(client))
		tools.Register(environments.NewGetEnvironmentBomTool(client))
		tools.Register(tools.NewGetLicenseTool(client))
	} else {
		// Register a dummy tool that explains the configuration issue
		tools.Register(&tools.ConfigurationErrorTool{})
	}

	// Mutation tools (register only if allowed and we have a valid client)
	if *allowMutation && client != nil && cfg != nil {
		tools.Register(users.NewCreateUserTool(client, cfg.EnvironmentID))
		tools.Register(users.NewDeleteUserTool(client, cfg.EnvironmentID))
		tools.Register(users.NewUpdateUserTool(client, cfg.EnvironmentID))
		tools.Register(users.NewSetUserEnabledTool(client, cfg.EnvironmentID))
		tools.Register(users.NewUnlockUserPasswordTool(client, cfg.EnvironmentID))
		tools.Register(users.NewResetUserPasswordTool(client, cfg.EnvironmentID))
		tools.Register(membership.NewAddUserToGroupTool(client, cfg.EnvironmentID))
		tools.Register(membership.NewRemoveUserFromGroupTool(client, cfg.EnvironmentID))
		tools.Register(populations.NewCreatePopulationTool(client, cfg.EnvironmentID))
		tools.Register(populations.NewDeletePopulationTool(client, cfg.EnvironmentID))
		tools.Register(groups.NewCreateGroupTool(client, cfg.EnvironmentID))
		tools.Register(groups.NewDeleteGroupTool(client, cfg.EnvironmentID))
		tools.Register(groups.NewUpdateGroupTool(client, cfg.EnvironmentID))
		tools.Register(environments.NewCreateEnvironmentTool(client))
		tools.Register(environments.NewDeleteEnvironmentTool(client))
		tools.Register(environments.NewUpdateEnvironmentStatusTool(client))
	}
	// Print startup stats
	toolList := tools.List()
	names := make([]string, len(toolList))
	for i, t := range toolList {
		names[i] = t.Name()
	}
	sort.Strings(names)
	log.Printf("Server version %s", serverVersion)
	log.Printf("Registered %d tools: %s", len(names), strings.Join(names, ", "))

	// Start server based on transport mode
	if *transportMode == "stdio" {
		log.Printf("Starting MCP server with stdio transport for Claude Desktop")
		stdioServer := mcp.NewStdioServer(serverVersion)
		if err := stdioServer.Run(context.Background()); err != nil {
			log.Fatalf("stdio server error: %v", err)
		}
	} else {
		// HTTP mode - keep existing behavior for backward compatibility
		var apiKey string
		if !*allowInsecure {
			// Generate and display API key for authenticating MCP calls (use header X-API-Key)
			apiKey = getAPIKey()
			log.Printf("API key: %s", apiKey)
		}

		// Start HTTP server
		router := mcp.NewRouter(apiKey, *allowInsecure)
		addr := ":" + *serverPort
		log.Printf("PingOne MCP Server listening on %s (transport: %s, insecure mode: %v)", addr, *transportMode, *allowInsecure)
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}
