package plg_backend_s3_oidc

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	. "github.com/mickael-kerjean/filestash/server/common"
)

// A service that cannot be reached shows as ErrNotReachable; a service that
// answers no stays ErrPermissionDenied.

func TestIsUnreachable(t *testing.T) {
	transport := &url.Error{Op: "Post", URL: "https://idp.example.test/token", Err: errors.New("x509: certificate signed by unknown authority")}
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"http client error", transport, true},
		{"wrapped http client error", fmt.Errorf("failed to send request: %w", transport), true},
		{"AWS request error", awserr.New(request.ErrCodeRequestError, "send request failed", transport), true},
		{"wrapped AWS request error", fmt.Errorf("failed to assume role: %w", awserr.New(request.ErrCodeRequestError, "send request failed", transport)), true},
		{"AWS access denied", awserr.New("AccessDenied", "not authorized to perform sts:AssumeRole", nil), false},
		{"refusal from the token endpoint", fmt.Errorf("non-200 status code: 403"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUnreachable(tt.err); got != tt.want {
				t.Errorf("isUnreachable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestGetSessionFromTokenUnreachable(t *testing.T) {
	tok := fakeToken(withProjectRoles)
	role := "dev_fakeproject_researcher"

	// The default client does not trust httptest's certificate, as when the image has no CA bundle.
	untrusted := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer untrusted.Close()
	if _, _, _, err := getSessionFromToken(tok, role, "works", "secret", "workss3", 900, untrusted.URL, "datajoint"); err != ErrNotReachable {
		t.Errorf("untrusted certificate: got %v, want %v", err, ErrNotReachable)
	}

	refusing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"access_denied"}`, http.StatusForbidden)
	}))
	defer refusing.Close()
	if _, _, _, err := getSessionFromToken(tok, role, "works", "secret", "workss3", 900, refusing.URL, "datajoint"); err != ErrPermissionDenied {
		t.Errorf("token exchange refused: got %v, want %v", err, ErrPermissionDenied)
	}
}

func TestGetSessionFromTokenAssumeRoleUnreachable(t *testing.T) {
	key := newKey(t)
	var hits int32
	srv := jwksServer(t, &key.PublicKey, "k1", &hits)
	defer srv.Close()
	iss := srv.URL + "/realms/test"
	signingKeys = &jwksCache{}

	fake := &fakeSTS{}
	orig := newSTSClient
	newSTSClient = func(string) stsiface.STSAPI { return fake }
	defer func() { newSTSClient = orig }()

	claims := baseClaims(time.Now())
	claims["iss"] = iss
	tok := signToken(t, key, "RS256", "k1", claims)
	role := "arn:aws:iam::111111111111:role/myorg_flow-cytometry_researcher"
	prefix := "arn:aws:iam::111111111111:role/myorg_"

	fake.err = awserr.New(request.ErrCodeRequestError, "send request failed",
		&url.Error{Op: "Post", URL: "https://sts.amazonaws.com/", Err: errors.New("x509: certificate signed by unknown authority")})
	if _, _, _, err := getSessionFromTokenAssumeRole(tok, role, "works", "", iss, prefix, "", 900); err != ErrNotReachable {
		t.Errorf("STS unreachable: got %v, want %v", err, ErrNotReachable)
	}
	fake.err = awserr.New("AccessDenied", "not authorized to perform sts:AssumeRole", nil)
	if _, _, _, err := getSessionFromTokenAssumeRole(tok, role, "works", "", iss, prefix, "", 900); err != ErrPermissionDenied {
		t.Errorf("STS refused: got %v, want %v", err, ErrPermissionDenied)
	}

	// The identity provider's signing keys cannot be fetched and none are cached.
	srv.Close()
	signingKeys = &jwksCache{}
	if _, _, _, err := getSessionFromTokenAssumeRole(tok, role, "works", "", iss, prefix, "", 900); err != ErrNotReachable {
		t.Errorf("identity provider unreachable: got %v, want %v", err, ErrNotReachable)
	}
}
