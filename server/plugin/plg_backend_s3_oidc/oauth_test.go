package plg_backend_s3_oidc

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestCanAssumeRole(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		decodedKcToken map[string]interface{}
		roleArn        string
		expected       bool
	}{
		{
			name: "User can assume role",
			decodedKcToken: map[string]interface{}{
				"resource_access": map[string]interface{}{
					"works": map[string]interface{}{
						"roles": []interface{}{"role1", "role2"},
					},
				},
			},
			roleArn:  "role1",
			expected: true,
		},
		{
			name: "User cannot assume role",
			decodedKcToken: map[string]interface{}{
				"resource_access": map[string]interface{}{
					"works": map[string]interface{}{
						"roles": []interface{}{"role1", "role2"},
					},
				},
			},
			roleArn:  "role3",
			expected: false,
		},
		{
			name: "Empty roles",
			decodedKcToken: map[string]interface{}{
				"resource_access": map[string]interface{}{
					"works": map[string]interface{}{
						"roles": []interface{}{},
					},
				},
			},
			roleArn:  "role1",
			expected: false,
		},
		{
			name: "20241101_access_token",
			decodedKcToken: map[string]interface{}{
				"exp":           1730517335,
				"iat":           1730474136,
				"auth_time":     1730474135,
				"jti":           "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"iss":           "https://keycloak.example.org/realms/datajoint",
				"aud":           "account",
				"sub":           "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"typ":           "Bearer",
				"azp":           "works",
				"session_state": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"acr":           "1",
				"realm_access": map[string]interface{}{
					"roles": []interface{}{
						"datajoint",
						"acme-lab_ephys",
						"offline_access",
						"acme-lab-test_ephys",
						"default-roles-datajoint",
						"uma_authorization",
					},
				},
				"resource_access": map[string]interface{}{
					"works": map[string]interface{}{
						"roles": []interface{}{
							"acme-lab-test_guest",
							"acme-lab-test_ephys_researcher",
						},
					},
					"account": map[string]interface{}{
						"roles": []interface{}{
							"manage-account",
							"manage-account-links",
							"view-profile",
						},
					},
				},
				"scope":          "openid email profile groups",
				"sid":            "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"email_verified": true,
				"name":           "Test User",
				"groups": []interface{}{
					"datajoint",
					"acme-lab_ephys",
					"offline_access",
					"acme-lab-test_ephys",
					"default-roles-datajoint",
					"uma_authorization",
				},
				"preferred_username": "test-user",
			},
			roleArn:  "dev_fakeproject_researcher",
			expected: true,
		},
		{
			name: "20241104_access_token",
			decodedKcToken: map[string]interface{}{
				"exp":           1730774655,
				"iat":           1730731455,
				"auth_time":     1730731455,
				"jti":           "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"iss":           "https://keycloak.example.org/realms/datajoint",
				"aud":           "account",
				"sub":           "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"typ":           "Bearer",
				"azp":           "works",
				"session_state": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"acr":           "1",
				"realm_access": map[string]interface{}{
					"roles": []interface{}{
						"datajoint",
						"acme-lab_ephys",
						"offline_access",
						"acme-lab-test_ephys",
						"default-roles-datajoint",
						"uma_authorization",
					},
				},
				"resource_access": map[string]interface{}{
					"works": map[string]interface{}{
						"roles": []interface{}{
							"acme-lab-test_guest",
							"acme-lab-test_ephys_researcher",
						},
					},
					"account": map[string]interface{}{
						"roles": []interface{}{
							"manage-account",
							"manage-account-links",
							"view-profile",
						},
					},
				},
				"scope":          "openid email profile groups",
				"sid":            "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"email_verified": true,
				"name":           "Test User",
				"groups": []interface{}{
					"datajoint",
					"acme-lab_ephys",
					"offline_access",
					"acme-lab-test_ephys",
					"default-roles-datajoint",
					"uma_authorization",
				},
				"preferred_username": "example",
				"given_name":         "Test",
				"family_name":        "User",
				"email":              "example@example.com",
			},
			roleArn:  "dev_fakeproject_researcher",
			expected: true,
		},
		{
			name: "20241107_access_token",
			decodedKcToken: map[string]interface{}{
				"exp":           1730774655,
				"iat":           1730731455,
				"auth_time":     1730731455,
				"jti":           "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"iss":           "https://keycloak.example.org/realms/datajoint",
				"aud":           "account",
				"sub":           "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"typ":           "Bearer",
				"azp":           "works",
				"session_state": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"acr":           "1",
				"realm_access": map[string]interface{}{
					"roles": []interface{}{
						"datajoint",
						"acme-lab_ephys",
						"offline_access",
						"acme-lab-test_ephys",
						"default-roles-datajoint",
						"uma_authorization",
					},
				},
				"resource_access": map[string]interface{}{
					"works": map[string]interface{}{
						"roles": []interface{}{
							"acme-lab-test_guest",
							"acme-lab-test_ephys_researcher",
						},
					},
					"account": map[string]interface{}{
						"roles": []interface{}{
							"manage-account",
							"manage-account-links",
							"view-profile",
						},
					},
				},
				"scope":          "openid email profile groups",
				"sid":            "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"email_verified": true,
				"name":           "Test User",
				"groups": []interface{}{
					"datajoint",
					"acme-lab_ephys",
					"offline_access",
					"acme-lab-test_ephys",
					"default-roles-datajoint",
					"uma_authorization",
				},
				"preferred_username": "test-user",
				"email":              "example@example.com",
			},
			roleArn:  "dev_fakeproject_researcher",
			expected: true,
		},
		{
			name:           "No resource_access",
			decodedKcToken: map[string]interface{}{},
			roleArn:        "role1",
			expected:       false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canAssumeRole(tt.decodedKcToken, tt.roleArn, "datajoint")
			if result != tt.expected {
				t.Errorf("canAssumeRole() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCanAssumeRoleFromRawToken(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		rawToken string
		roleArn  string
		expected bool
	}{
		{
			name:     "support group",
			rawToken: fakeToken(withProjectRoles),
			roleArn:  "dev_fakeproject_researcher",
			expected: true,
		},
		{
			name:     "support group",
			rawToken: fakeToken(withProjectRoles),
			roleArn:  "dev_fakeproject_researcher",
			expected: true,
		},
		{
			name:     "support group",
			rawToken: fakeToken(withoutProjectRoles),
			roleArn:  "dev_fakeproject_researcher",
			expected: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decodedKcToken, decodeErr := decodeToken(tt.rawToken)
			if decodeErr != nil {
				t.Errorf("Error decoding Keycloak token: %v", decodeErr)
			}
			result := canAssumeRole(decodedKcToken, tt.roleArn, "datajoint")
			if result != tt.expected {
				t.Errorf("canAssumeRoleFromRawToken() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// fakeToken encodes claims as an unsigned access token. decodeToken reads only
// the payload, so the tests need no real (and personal) Keycloak token.
func fakeToken(claims map[string]interface{}) string {
	enc := func(v interface{}) string {
		b, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	return enc(map[string]string{"alg": "none", "typ": "JWT"}) + "." + enc(claims) + ".sig"
}

// A Keycloak access token's claims, shaped like a real one's, for a made-up
// user in the support group: with roles on a project, and without.
var supportGroups = []interface{}{
	"datajoint", "acme-lab_ephys", "offline_access", "acme-lab-test_ephys",
	"default-roles-datajoint", "uma_authorization",
}

func tokenClaims(resourceAccess map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"iss":                "https://keycloak.example.org/realms/datajoint",
		"aud":                "account",
		"azp":                "works",
		"typ":                "Bearer",
		"scope":              "openid email profile groups",
		"preferred_username": "test-user",
		"name":               "Test User",
		"email":              "test-user@example.org",
		"groups":             supportGroups,
		"realm_access":       map[string]interface{}{"roles": supportGroups},
		"resource_access":    resourceAccess,
	}
}

var accountRoles = map[string]interface{}{
	"roles": []interface{}{"manage-account", "manage-account-links", "view-profile"},
}

var withProjectRoles = tokenClaims(map[string]interface{}{
	"works": map[string]interface{}{
		"roles": []interface{}{"acme-lab-test_guest", "acme-lab-test_ephys_researcher"},
	},
	"account": accountRoles,
})

var withoutProjectRoles = tokenClaims(map[string]interface{}{"account": accountRoles})
