// Copyright © 2025 Ping Identity Corporation

package applications_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/applications"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetApplicationSecretHandler(t *testing.T) {
	tests := []struct {
		name  string
		input applications.GetApplicationSecretInput
	}{
		{
			name: "Success - Prompts user to retrieve application secret",
			input: applications.GetApplicationSecretInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
			},
		},
		{
			name: "Success - Valid UUIDs provided",
			input: applications.GetApplicationSecretInput{
				EnvironmentId: uuid.New(),
				ApplicationId: uuid.New(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup - use empty request since Elicit is optional
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			initializeAuthContext := testutils.MockContextInitializer()

			// Execute
			handler := applications.GetApplicationSecretHandler(initializeAuthContext)
			result, output, err := handler(ctx, req, tt.input)

			// Assert - should succeed even if Elicit fails (nil session)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Nil(t, output)

			// Verify result contains appropriate messaging about secret
			require.Len(t, result.Content, 1)
			textContent, ok := result.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			assert.Contains(t, textContent.Text, "secret")
			assert.Contains(t, textContent.Text, "Agent cannot and should not access")
		})
	}
}

func TestGetApplicationSecretHandler_AuthError(t *testing.T) {
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Mock auth context initializer that returns error
	testErr := assert.AnError
	initializeAuthContext := testutils.MockContextInitializerWithError(testErr)

	input := applications.GetApplicationSecretInput{
		EnvironmentId: testEnvironmentId,
		ApplicationId: uuid.New(),
	}

	handler := applications.GetApplicationSecretHandler(initializeAuthContext)
	result, output, err := handler(ctx, req, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, output)
}
