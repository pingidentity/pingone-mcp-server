package tools

import (
   "context"
   "encoding/json"
)

// Tool defines the interface each MCP tool must implement
type Tool interface {
   // Name returns the unique tool name
   Name() string
   // Description returns a human-readable description
   Description() string
   // InputSchema returns the JSON schema for the tool's input
   InputSchema() json.RawMessage
   // Run executes the tool with given args and returns output or error
   Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error)
}

// registry holds registered tools by name
var registry = make(map[string]Tool)

// Register adds a tool to the global registry
func Register(t Tool) {
   registry[t.Name()] = t
}

// Get retrieves a tool by name
func Get(name string) (Tool, bool) {
   t, ok := registry[name]
   return t, ok
}

// List returns all registered tools
func List() []Tool {
   tools := make([]Tool, 0, len(registry))
   for _, t := range registry {
       tools = append(tools, t)
   }
   return tools
}

// Reset clears the tool registry (for testing)
func Reset() {
   registry = make(map[string]Tool)
}