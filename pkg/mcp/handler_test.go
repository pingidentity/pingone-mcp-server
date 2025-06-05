package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// stubTool implements tools.Tool for testing
type stubTool struct{}

func (s *stubTool) Name() string        { return "echo" }
func (s *stubTool) Description() string { return "echoes a message" }
func (s *stubTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"msg":{"type":"string"}},"required":["msg"]}`)
}

// TestRunHandler_MethodNotAllowed ensures that non-POST requests to /run return 405
func TestRunHandler_MethodNotAllowed(t *testing.T) {
	router := NewRouter("testkey", false)
	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/run", nil)
	req.Header.Set("X-API-Key", "testkey")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != http.MethodPost {
		t.Errorf("expected Allow header %q, got %q", http.MethodPost, allow)
	}
}
func (s *stubTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	m, ok := args["msg"].(string)
	if !ok {
		return nil, fmt.Errorf("msg must be a string")
	}
	return map[string]interface{}{"reply": m}, nil
}

// TestToolsHandler_NonEmpty verifies /tools returns registered tools
func TestToolsHandler_NonEmpty(t *testing.T) {
	// reset registry and add two stub tools
	tools.Reset()
	tools.Register(&stubTool{})
	tools.Register(&stubTool{}) // same name allowed, but registry uses map, so last wins
	router := NewRouter("testkey", false)
	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/tools", nil)
	req.Header.Set("X-API-Key", "testkey")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var list []ToolDescriptor
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if len(list) != 1 {
		// stubTool.Name() == "echo"
		t.Errorf("expected 1 tool descriptor, got %d", len(list))
	}
}

// TestInitializeHandler ensures the /mcp/v1/initialize endpoint returns valid metadata
func TestInitializeHandler(t *testing.T) {
	router := NewRouter("testkey", false)
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/initialize", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var meta ServerMetadata
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if meta.Name == "" {
		t.Error("metadata Name is empty; expected non-empty")
	}
	if meta.Version == "" {
		t.Error("metadata Version is empty; expected non-empty")
	}
	if meta.Description == "" {
		t.Error("metadata Description is empty; expected non-empty")
	}
}

// TestToolsHandler_Empty verifies that /tools returns an empty list initially
func TestToolsHandler_Empty(t *testing.T) {
	tools.Reset()
	router := NewRouter("testkey", false)
	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/tools", nil)
	req.Header.Set("X-API-Key", "testkey")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var toolsList []ToolDescriptor
	if err := json.Unmarshal(rec.Body.Bytes(), &toolsList); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if len(toolsList) != 0 {
		t.Errorf("expected empty tool list, got %d", len(toolsList))
	}
}

// TestRunHandler_UnknownTool verifies that invoking a non-existent tool returns 404
func TestRunHandler_UnknownTool(t *testing.T) {
	router := NewRouter("testkey", false)
	body := []byte("{\"tool\":\"nope\",\"input\":{}}")
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/run", bytes.NewReader(body))
	req.Header.Set("X-API-Key", "testkey")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected non-empty error message")
	}
}

// TestRunHandler_Success verifies running a registered tool returns its output
func TestRunHandler_Success(t *testing.T) {
	// Reset and register stub tool
	tools.Reset()
	tools.Register(&stubTool{})

	router := NewRouter("testkey", false)
	body := []byte("{\"tool\":\"echo\",\"input\":{\"msg\":\"hello\"}}")
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/run", bytes.NewReader(body))
	req.Header.Set("X-API-Key", "testkey")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var resp RunResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp.Output["reply"] != "hello" {
		t.Errorf("expected reply 'hello', got %v", resp.Output["reply"])
	}
}

// TestRunHandler_BadJSON returns 400 on invalid JSON
func TestRunHandler_BadJSON(t *testing.T) {
	router := NewRouter("testkey", false)
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/run", strings.NewReader("{bad json}"))
	req.Header.Set("X-API-Key", "testkey")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Errorf("invalid JSON error response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected non-empty error message for bad JSON")
	}
}

// TestRunHandler_RunError returns 500 on tool.Run error
func TestRunHandler_RunError(t *testing.T) {
	// register a stub tool that returns an error
	tools.Reset()
	tools.Register(&stubTool{})
	// override stubTool.Run to return an error
	errTool := &stubTool{}
	tools.Reset()
	tools.Register(errTool)
	// define a tool name that exists but errors
	// use stubTool which errors on missing msg
	router := NewRouter("testkey", false)
	// invoke without required input to trigger error
	body := []byte("{\"tool\":\"echo\",\"input\":{}}")
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/run", bytes.NewReader(body))
	req.Header.Set("X-API-Key", "testkey")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Errorf("invalid JSON error response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected non-empty error message for run error")
	}
}

// TestHandleJSONRPCHTTP_MethodNotAllowed tests non-POST requests to JSON-RPC endpoint
func TestHandleJSONRPCHTTP_MethodNotAllowed(t *testing.T) {
	router := NewRouter("testkey", false)
	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/jsonrpc", nil)
	req.Header.Set("X-API-Key", "testkey")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != http.MethodPost {
		t.Errorf("expected Allow header %q, got %q", http.MethodPost, allow)
	}
}

// TestHandleJSONRPCHTTP_Initialize tests JSON-RPC initialize method via HTTP
func TestHandleJSONRPCHTTP_Initialize(t *testing.T) {
	router := NewRouter("testkey", false)
	body := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/jsonrpc", bytes.NewReader(body))
	req.Header.Set("X-API-Key", "testkey")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	
	var response JSONRPCResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse JSON-RPC response: %v", err)
	}
	
	if response.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", response.JSONRPC)
	}
	if response.Error != nil {
		t.Errorf("unexpected error: %v", response.Error)
	}
}

// TestHandleJSONRPCHTTP_ToolsList tests JSON-RPC tools/list method via HTTP
func TestHandleJSONRPCHTTP_ToolsList(t *testing.T) {
	tools.Reset()
	tools.Register(&stubTool{})
	
	router := NewRouter("testkey", false)
	body := []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/jsonrpc", bytes.NewReader(body))
	req.Header.Set("X-API-Key", "testkey")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	
	var response JSONRPCResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse JSON-RPC response: %v", err)
	}
	
	if response.Error != nil {
		t.Errorf("unexpected error: %v", response.Error)
	}
	
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result to be object, got %T", response.Result)
	}
	
	toolsList, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatalf("expected tools array in result")
	}
	
	if len(toolsList) != 1 {
		t.Errorf("expected 1 tool, got %d", len(toolsList))
	}
}

// TestHandleJSONRPCHTTP_ToolsCall tests JSON-RPC tools/call method via HTTP
func TestHandleJSONRPCHTTP_ToolsCall(t *testing.T) {
	tools.Reset()
	tools.Register(&stubTool{})
	
	router := NewRouter("testkey", false)
	body := []byte(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"msg":"test message"}}}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/jsonrpc", bytes.NewReader(body))
	req.Header.Set("X-API-Key", "testkey")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	
	var response JSONRPCResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse JSON-RPC response: %v", err)
	}
	
	if response.Error != nil {
		t.Errorf("unexpected error: %v", response.Error)
	}
}

// TestHandleJSONRPCHTTP_InvalidJSON tests handling of malformed JSON
func TestHandleJSONRPCHTTP_InvalidJSON(t *testing.T) {
	router := NewRouter("testkey", false)
	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/jsonrpc", bytes.NewReader(body))
	req.Header.Set("X-API-Key", "testkey")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	
	var response JSONRPCResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse JSON-RPC error response: %v", err)
	}
	
	if response.Error == nil {
		t.Error("expected error for invalid JSON")
	}
	if response.Error.Code != ErrorCodeParseError {
		t.Errorf("expected parse error code %d, got %d", ErrorCodeParseError, response.Error.Code)
	}
}

// TestHandleJSONRPCHTTP_UnknownMethod tests handling of unknown JSON-RPC methods
func TestHandleJSONRPCHTTP_UnknownMethod(t *testing.T) {
	router := NewRouter("testkey", false)
	body := []byte(`{"jsonrpc":"2.0","id":4,"method":"unknown/method","params":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/jsonrpc", bytes.NewReader(body))
	req.Header.Set("X-API-Key", "testkey")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
	
	var response JSONRPCResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse JSON-RPC error response: %v", err)
	}
	
	if response.Error == nil {
		t.Error("expected error for unknown method")
	}
	if response.Error.Code != ErrorCodeMethodNotFound {
		t.Errorf("expected method not found error code %d, got %d", ErrorCodeMethodNotFound, response.Error.Code)
	}
}

// TestAuthMiddleware_MissingAPIKey tests authentication with missing API key
func TestAuthMiddleware_MissingAPIKey(t *testing.T) {
	router := NewRouter("testkey", false)
	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/tools", nil)
	// No X-API-Key header
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

// TestAuthMiddleware_WrongAPIKey tests authentication with wrong API key
func TestAuthMiddleware_WrongAPIKey(t *testing.T) {
	router := NewRouter("testkey", false)
	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/tools", nil)
	req.Header.Set("X-API-Key", "wrongkey")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

// TestRouterInsecureMode tests router in insecure mode (no auth required)
func TestRouterInsecureMode(t *testing.T) {
	router := NewRouter("", true) // insecure mode
	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/tools", nil)
	// No X-API-Key header, but insecure mode should allow it
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 in insecure mode, got %d", rec.Code)
	}
}

// TestRouterInsecureMode_JSONRPC tests JSON-RPC endpoint in insecure mode
func TestRouterInsecureMode_JSONRPC(t *testing.T) {
	router := NewRouter("", true) // insecure mode
	body := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/jsonrpc", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No X-API-Key header, but insecure mode should allow it
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 in insecure mode, got %d", rec.Code)
	}
}

// TestToolsHandler_MethodNotAllowed tests non-GET requests to tools endpoint
func TestToolsHandler_MethodNotAllowed(t *testing.T) {
	router := NewRouter("testkey", false)
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/tools", nil)
	req.Header.Set("X-API-Key", "testkey")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != http.MethodGet {
		t.Errorf("expected Allow header %q, got %q", http.MethodGet, allow)
	}
}
