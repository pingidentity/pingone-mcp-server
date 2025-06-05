package mcp

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// ServerMetadata describes the MCP server to clients
type ServerMetadata struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
// loggingResponseWriter wraps http.ResponseWriter to capture status code
// and provide implicit status capture on writes.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write captures implicit HTTP 200 status on first write if WriteHeader wasn't called.
func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.statusCode == 0 {
		lrw.statusCode = http.StatusOK
	}
	return lrw.ResponseWriter.Write(b)
}

// NewRouter returns an HTTP handler with MCP endpoints registered.
// If allowInsecure is true, API key auth is disabled for protected endpoints.
func NewRouter(apiKey string, allowInsecure bool) http.Handler {
	mux := http.NewServeMux()
	// Public endpoint: initialization (legacy REST)
	mux.HandleFunc("/mcp/v1/initialize", initializeHandler)

	// JSON-RPC endpoint for MCP compliance
	jsonrpcHandler := NewJSONRPCHandler("0.1.0")
	jsonrpcHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		handleJSONRPCHTTP(w, r, jsonrpcHandler)
	}

	if allowInsecure {
		mux.HandleFunc("/mcp/v1/tools", toolsHandler)
		mux.HandleFunc("/mcp/v1/run", runHandler)
		mux.HandleFunc("/mcp/v1/jsonrpc", jsonrpcHandlerFunc)
		return mux
	}

	mux.Handle("/mcp/v1/tools", authMiddleware(apiKey, http.HandlerFunc(toolsHandler)))
	mux.Handle("/mcp/v1/run", authMiddleware(apiKey, http.HandlerFunc(runHandler)))
	mux.Handle("/mcp/v1/jsonrpc", authMiddleware(apiKey, http.HandlerFunc(jsonrpcHandlerFunc)))
	return mux
}

// initializeHandler returns basic server metadata
func initializeHandler(w http.ResponseWriter, r *http.Request) {
	meta := ServerMetadata{
		Name:        "PingOne MCP Server",
		Version:     "0.1.0",
		Description: "MCP endpoint for PingOne user and group management",
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(meta)
}

// ToolDescriptor describes a registered tool
type ToolDescriptor struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// RunRequest is the payload for /run
type RunRequest struct {
	Tool  string                 `json:"tool"`
	Input map[string]interface{} `json:"input"`
}

// RunResponse is the successful response from /run
type RunResponse struct {
	Output map[string]interface{} `json:"output"`
}

// ErrorResponse is the error payload
type ErrorResponse struct {
	Error string `json:"error"`
}

// toolsHandler returns all registered tools
func toolsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	list := tools.List()
	descriptors := make([]ToolDescriptor, 0, len(list))
	for _, t := range list {
		descriptors = append(descriptors, ToolDescriptor{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.InputSchema(),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(descriptors)
}

// runHandler executes a specified tool
func runHandler(origW http.ResponseWriter, r *http.Request) {
	// wrap ResponseWriter to capture status code
	lrw := &loggingResponseWriter{ResponseWriter: origW}
	// track execution time
	start := time.Now()
	// at end, log one summary line
	var req RunRequest
	defer func() {
		log.Printf("MCP method=%s tool=%s status=%d elapsed=%s", r.Method, req.Tool, lrw.statusCode, time.Since(start))
	}()
	// Only allow POST. For non-POST, return 405 and capture attempted tool name for logging
	if r.Method != http.MethodPost {
		// attempt to extract the tool name from URL query parameters
		toolName := r.URL.Query().Get("tool")
		// fallback to JSON body if query param is not set
		if toolName == "" && r.Body != nil {
			var peek struct {
				Tool string `json:"tool"`
			}
			if bodyBytes, err := io.ReadAll(r.Body); err == nil {
				_ = json.Unmarshal(bodyBytes, &peek)
				toolName = peek.Tool
			}
		}
		req.Tool = toolName
		lrw.Header().Set("Allow", http.MethodPost)
		lrw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// decode request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lrw.Header().Set("Content-Type", "application/json")
		lrw.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(lrw).Encode(ErrorResponse{Error: "invalid request"})
		return
	}
	// look up tool
	tool, ok := tools.Get(req.Tool)
	if !ok {
		lrw.Header().Set("Content-Type", "application/json")
		lrw.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(lrw).Encode(ErrorResponse{Error: "tool not found"})
		return
	}
	// execute tool
	output, err := tool.Run(r.Context(), req.Input)
	if err != nil {
		lrw.Header().Set("Content-Type", "application/json")
		// If we have a bodyErr, return its status and body
		type bodyErr interface {
			Body() []byte
			Error() string
		}
		if apiErr, ok := err.(bodyErr); ok {
			parts := strings.SplitN(apiErr.Error(), " ", 2)
			if code, parseErr := strconv.Atoi(parts[0]); parseErr == nil {
				lrw.WriteHeader(code)
			} else {
				lrw.WriteHeader(http.StatusInternalServerError)
			}
			lrw.Write(apiErr.Body())
		} else {
			lrw.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(lrw).Encode(ErrorResponse{Error: err.Error()})
		}
		return
	}
	// success
	lrw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(lrw).Encode(RunResponse{Output: output})
}

// handleJSONRPCHTTP handles JSON-RPC requests over HTTP
func handleJSONRPCHTTP(w http.ResponseWriter, r *http.Request, handler *JSONRPCHandler) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(NewJSONRPCError(nil, ErrorCodeParseError, "failed to read request body", err.Error()))
		return
	}

	// Parse JSON-RPC request
	req, err := ParseJSONRPCRequest(body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(NewJSONRPCError(nil, ErrorCodeParseError, err.Error(), nil))
		return
	}

	// Handle request
	resp := handler.HandleRequest(r.Context(), req)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if resp.Error != nil {
		// Set appropriate HTTP status code based on JSON-RPC error
		switch resp.Error.Code {
		case ErrorCodeMethodNotFound:
			w.WriteHeader(http.StatusNotFound)
		case ErrorCodeInvalidParams:
			w.WriteHeader(http.StatusBadRequest)
		case ErrorCodeInternalError:
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// authMiddleware enforces API key via X-API-Key header
func authMiddleware(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
