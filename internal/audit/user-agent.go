// Copyright © 2025 Ping Identity Corporation

package audit

import "fmt"

func PingOneAPIUserAgent(serverVersion string) string {
	return fmt.Sprintf("pingone-mcp-server/%s", serverVersion)
}
