// Copyright © 2025 Ping Identity Corporation

package applications_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
)

var testEnvironmentId = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
var testAppId = uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

var (
	// Test applications with various configurations
	testOIDCApp = management.ReadOneApplication200Response{
		ApplicationOIDC: &management.ApplicationOIDC{
			Id:                      testutils.Pointer(testAppId.String()),
			Name:                    "Test OIDC Web App",
			Description:             testutils.Pointer("A test OIDC web application"),
			Enabled:                 true,
			Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
			Type:                    management.ENUMAPPLICATIONTYPE_WEB_APP,
			CreatedAt:               testutils.Pointer(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)),
			UpdatedAt:               testutils.Pointer(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)),
			GrantTypes:              []management.EnumApplicationOIDCGrantType{management.ENUMAPPLICATIONOIDCGRANTTYPE_AUTHORIZATION_CODE},
			RedirectUris:            []string{"https://example.com/callback"},
			TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_CLIENT_SECRET_BASIC,
		},
	}

	testOIDCAppOnlyRequiredFields = management.ReadOneApplication200Response{
		ApplicationOIDC: &management.ApplicationOIDC{
			Name:                    "Test Native App - Required Fields Only",
			Enabled:                 true,
			Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
			Type:                    management.ENUMAPPLICATIONTYPE_NATIVE_APP,
			GrantTypes:              []management.EnumApplicationOIDCGrantType{management.ENUMAPPLICATIONOIDCGRANTTYPE_AUTHORIZATION_CODE},
			TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_NONE,
		},
	}

	testSAMLApp = management.ReadOneApplication200Response{
		ApplicationSAML: &management.ApplicationSAML{
			Id:          testutils.Pointer(testAppId.String()),
			Name:        "Test SAML App",
			Description: testutils.Pointer("A test SAML application"),
			Enabled:     true,
			Protocol:    management.ENUMAPPLICATIONPROTOCOL_SAML,
			Type:        management.ENUMAPPLICATIONTYPE_WEB_APP,
			CreatedAt:   testutils.Pointer(time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)),
			UpdatedAt:   testutils.Pointer(time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)),
			AcsUrls:     []string{"https://saml.example.com/acs"},
		},
	}

	testSinglePageApp = management.ReadOneApplication200Response{
		ApplicationOIDC: &management.ApplicationOIDC{
			Id:                      testutils.Pointer(testAppId.String()),
			Name:                    "Test SPA",
			Description:             testutils.Pointer("A test single page application"),
			Enabled:                 false, // Disabled application
			Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
			Type:                    management.ENUMAPPLICATIONTYPE_SINGLE_PAGE_APP,
			CreatedAt:               testutils.Pointer(time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)),
			UpdatedAt:               testutils.Pointer(time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)),
			GrantTypes:              []management.EnumApplicationOIDCGrantType{management.ENUMAPPLICATIONOIDCGRANTTYPE_IMPLICIT, management.ENUMAPPLICATIONOIDCGRANTTYPE_AUTHORIZATION_CODE},
			RedirectUris:            []string{"https://spa.example.com/callback"},
			TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_CLIENT_SECRET_JWT,
		},
	}

	testExternalLinkApp = management.ReadOneApplication200Response{
		ApplicationExternalLink: &management.ApplicationExternalLink{
			Id:          testutils.Pointer(testAppId.String()),
			Name:        "Test External Link",
			Description: testutils.Pointer("An external link application"),
			Enabled:     true,
			Protocol:    management.ENUMAPPLICATIONPROTOCOL_EXTERNAL_LINK,
			Type:        management.ENUMAPPLICATIONTYPE_PORTAL_LINK_APP,
			CreatedAt:   testutils.Pointer(time.Date(2024, 1, 4, 12, 0, 0, 0, time.UTC)),
			UpdatedAt:   testutils.Pointer(time.Date(2024, 1, 4, 12, 0, 0, 0, time.UTC)),
			HomePageUrl: "https://external.example.com",
		},
	}

	testP1PortalApp = management.ReadOneApplication200Response{
		ApplicationPingOnePortal: &management.ApplicationPingOnePortal{
			Id:                      testutils.Pointer(testAppId.String()),
			Name:                    "PingOne Portal",
			Enabled:                 true,
			Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
			Type:                    management.ENUMAPPLICATIONTYPE_PING_ONE_PORTAL,
			TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_NONE,
			ApplyDefaultTheme:       true,
		},
	}

	testWSFEDApp = management.ReadOneApplication200Response{
		ApplicationWSFED: &management.ApplicationWSFED{
			Id:          testutils.Pointer(testAppId.String()),
			Name:        "Test WS-FED App",
			Description: testutils.Pointer("A test WS-FED application"),
			Enabled:     true,
			Protocol:    management.ENUMAPPLICATIONPROTOCOL_WS_FED,
			Type:        management.ENUMAPPLICATIONTYPE_WEB_APP,
			CreatedAt:   testutils.Pointer(time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC)),
			UpdatedAt:   testutils.Pointer(time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC)),
			DomainName:  "example.com",
			ReplyUrl:    "https://wsfed.example.com/reply",
			IdpSigning: management.ApplicationWSFEDAllOfIdpSigning{
				Algorithm: management.ENUMAPPLICATIONWSFEDIDPSIGNINGALGORITHM_SHA256WITH_RSA,
				Key: management.ApplicationWSFEDAllOfIdpSigningKey{
					Id: "test-signing-key-id",
				},
			},
			AudienceRestriction:         testutils.Pointer("urn:federation:MicrosoftOnline"),
			SloEndpoint:                 testutils.Pointer("https://wsfed.example.com/slo"),
			SubjectNameIdentifierFormat: testutils.Pointer(management.ENUMAPPLICATIONWSFEDSUBJECTNAMEIDENTIFIERFORMAT_EMAIL_ADDRESS),
		},
	}

	testP1SelfServiceApp = management.ReadOneApplication200Response{
		ApplicationPingOneSelfService: &management.ApplicationPingOneSelfService{
			Id:                       testutils.Pointer(testAppId.String()),
			Name:                     "PingOne Self Service",
			Description:              testutils.Pointer("A test PingOne Self Service application"),
			Enabled:                  true,
			Protocol:                 management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
			Type:                     management.ENUMAPPLICATIONTYPE_PING_ONE_SELF_SERVICE,
			CreatedAt:                testutils.Pointer(time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC)),
			UpdatedAt:                testutils.Pointer(time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC)),
			TokenEndpointAuthMethod:  management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_NONE,
			ApplyDefaultTheme:        true,
			EnableDefaultThemeFooter: testutils.Pointer(false),
		},
	}

	testP1AdminConsoleApp = management.ReadOneApplication200Response{
		ApplicationPingOneAdminConsole: &management.ApplicationPingOneAdminConsole{
			PkceEnforcement:         testutils.Pointer(management.ENUMAPPLICATIONOIDCPKCEOPTION_OPTIONAL),
			TokenEndpointAuthMethod: testutils.Pointer(management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_CLIENT_SECRET_BASIC),
		},
	}
)

// Helper functions to assert applications match expected values using go-cmp

func assertOIDCApplicationMatches(t *testing.T, expected *management.ApplicationOIDC, actual *management.ApplicationOIDC) {
	t.Helper()

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("OIDC application mismatch (-expected +actual):\n%s", diff)
	}
}
