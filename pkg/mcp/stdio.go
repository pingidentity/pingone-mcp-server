package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// StdioServer handles MCP communication over stdin/stdout
type StdioServer struct {
	handler *JSONRPCHandler
	input   io.Reader
	output  io.Writer
}

// NewStdioServer creates a new stdio-based MCP server
func NewStdioServer(serverVersion string) *StdioServer {
	return &StdioServer{
		handler: NewJSONRPCHandler(serverVersion),
		input:   os.Stdin,
		output:  os.Stdout,
	}
}

// NewStdioServerWithIO creates a stdio server with custom input/output (for testing)
func NewStdioServerWithIO(serverVersion string, input io.Reader, output io.Writer) *StdioServer {
	return &StdioServer{
		handler: NewJSONRPCHandler(serverVersion),
		input:   input,
		output:  output,
	}
}

// Run starts the stdio server loop
func (s *StdioServer) Run(ctx context.Context) error {
	log.Printf("Starting MCP stdio server...")
	
	scanner := bufio.NewScanner(s.input)
	encoder := json.NewEncoder(s.output)
	
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		// Log incoming request for debugging (to stderr so it doesn't interfere with stdout)
		log.Printf("Received: %s", line)
		
		// Parse JSON-RPC request
		req, err := ParseJSONRPCRequest([]byte(line))
		if err != nil {
			// Send parse error response
			errorResp := NewJSONRPCError(nil, ErrorCodeParseError, 
				fmt.Sprintf("parse error: %v", err), nil)
			if encErr := encoder.Encode(errorResp); encErr != nil {
				log.Printf("Failed to send parse error response: %v", encErr)
			}
			continue
		}
		
		// Handle the request
		resp := s.handler.HandleRequest(ctx, req)
		
		// Send response
		if err := encoder.Encode(resp); err != nil {
			log.Printf("Failed to send response: %v", err)
			// Try to send an internal error response
			errorResp := NewJSONRPCError(req.ID, ErrorCodeInternalError, 
				"failed to encode response", err.Error())
			_ = encoder.Encode(errorResp)
		}
		
		// Log outgoing response for debugging
		if respBytes, err := json.Marshal(resp); err == nil {
			log.Printf("Sent: %s", string(respBytes))
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stdio scanner error: %w", err)
	}
	
	return nil
}