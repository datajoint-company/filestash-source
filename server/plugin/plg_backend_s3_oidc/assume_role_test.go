package plg_backend_s3_oidc

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

// ------------------------------------------------------------ helpers

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func signToken(t *testing.T, key *rsa.PrivateKey, alg, kid string, claims map[string]interface{}) string {
	t.Helper()
	h, _ := json.Marshal(map[string]string{"alg": alg, "kid": kid, "typ": "JWT"})
	c, _ := json.Marshal(claims)
	input := b64(h) + "." + b64(c)
	hash := rsaAlgs[alg]
	if hash == 0 {
		hash = crypto.SHA256
	}
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, hash, digest(hash, []byte(input)))
	if err != nil {
		t.Fatal(err)
	}
	return input + "." + b64(sig)
}

func newKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

const testIssuer = "https://idp.example.test/realms/test"

func baseClaims(now time.Time) map[string]interface{} {
	return map[string]interface{}{
		"iss":                testIssuer,
		"azp":                "works",
		"aud":                "account",
		"sub":                "00000000-0000-0000-0000-000000000000",
		"exp":                float64(now.Add(5 * time.Minute).Unix()),
		"iat":                float64(now.Unix()),
		"preferred_username": "alice",
		"resource_access": map[string]interface{}{
			"works": map[string]interface{}{"roles": []interface{}{"myorg_flow-cytometry_researcher"}},
		},
		"groups": []interface{}{"myorg_msd"},
	}
}

func keyring(k *rsa.PrivateKey, kid string) func(string) (*rsa.PublicKey, error) {
	return func(got string) (*rsa.PublicKey, error) {
		if got != kid {
			return nil, fmt.Errorf("no signing key with kid %q", got)
		}
		return &k.PublicKey, nil
	}
}

// ------------------------------------------------------------ mode + issuer

func TestParseCredentialMode(t *testing.T) {
	for in, want := range map[string]string{
		"": credentialModeWebIdentity, "web_identity": credentialModeWebIdentity,
		" WEB_IDENTITY ": credentialModeWebIdentity, "assume_role": credentialModeAssumeRole,
	} {
		got, err := parseCredentialMode(in)
		if err != nil || got != want {
			t.Errorf("parseCredentialMode(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"assumerole", "irsa", "true"} {
		if _, err := parseCredentialMode(bad); err == nil {
			t.Errorf("parseCredentialMode(%q) accepted an unknown mode", bad)
		}
	}
}

func TestIssuerFromTokenEndpoint(t *testing.T) {
	got, err := issuerFromTokenEndpoint(testIssuer+"/protocol/openid-connect/token", "")
	if err != nil || got != testIssuer {
		t.Errorf("derived %q, %v", got, err)
	}
	got, _ = issuerFromTokenEndpoint("anything", "https://other.test/realms/x/")
	if got != "https://other.test/realms/x" {
		t.Errorf("override not used: %q", got)
	}
	if _, err := issuerFromTokenEndpoint("https://idp.example.test/oauth/token", ""); err == nil {
		t.Error("non-Keycloak endpoint should need an explicit issuer")
	}
}

// ------------------------------------------------------------ verifyToken

func TestVerifyToken(t *testing.T) {
	now := time.Now()
	key, other := newKey(t), newKey(t)
	good := keyring(key, "k1")

	tok := signToken(t, key, "RS256", "k1", baseClaims(now))
	if _, err := verifyToken(tok, testIssuer, []string{"works"}, now, good); err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	for _, alg := range []string{"RS384", "RS512"} {
		if _, err := verifyToken(signToken(t, key, alg, "k1", baseClaims(now)), testIssuer, []string{"works"}, now, good); err != nil {
			t.Errorf("%s rejected: %v", alg, err)
		}
	}

	mut := func(f func(map[string]interface{})) map[string]interface{} {
		c := baseClaims(now)
		f(c)
		return c
	}
	parts := strings.Split(tok, ".")
	noneHdr := b64([]byte(`{"alg":"none","kid":"k1"}`))
	hsHdr := b64([]byte(`{"alg":"HS256","kid":"k1"}`))
	tampered := mut(func(c map[string]interface{}) { c["preferred_username"] = "mallory" })
	tb, _ := json.Marshal(tampered)

	cases := []struct {
		name, token string
	}{
		{"signed by another key", signToken(t, other, "RS256", "k1", baseClaims(now))},
		{"payload tampered after signing", parts[0] + "." + b64(tb) + "." + parts[2]},
		{"alg none", noneHdr + "." + parts[1] + "."},
		{"alg HS256", hsHdr + "." + parts[1] + "." + parts[2]},
		{"unknown kid", signToken(t, key, "RS256", "k2", baseClaims(now))},
		{"no kid", signToken(t, key, "RS256", "", baseClaims(now))},
		{"wrong issuer", signToken(t, key, "RS256", "k1", mut(func(c map[string]interface{}) { c["iss"] = "https://evil.test/realms/test" }))},
		{"expired", signToken(t, key, "RS256", "k1", mut(func(c map[string]interface{}) { c["exp"] = float64(now.Add(-time.Minute).Unix()) }))},
		{"no exp", signToken(t, key, "RS256", "k1", mut(func(c map[string]interface{}) { delete(c, "exp") }))},
		{"not yet valid", signToken(t, key, "RS256", "k1", mut(func(c map[string]interface{}) { c["nbf"] = float64(now.Add(time.Hour).Unix()) }))},
		{"wrong azp", signToken(t, key, "RS256", "k1", mut(func(c map[string]interface{}) { c["azp"] = "some-other-client" }))},
		{"no azp", signToken(t, key, "RS256", "k1", mut(func(c map[string]interface{}) { delete(c, "azp") }))},
		{"not a JWT", "abc.def"},
	}
	for _, c := range cases {
		if _, err := verifyToken(c.token, testIssuer, []string{"works"}, now, good); err == nil {
			t.Errorf("%s: accepted", c.name)
		}
	}

	// clock leeway: 10 s past exp is still accepted, 2 min is not
	skew := signToken(t, key, "RS256", "k1", mut(func(c map[string]interface{}) { c["exp"] = float64(now.Add(-10 * time.Second).Unix()) }))
	if _, err := verifyToken(skew, testIssuer, []string{"works"}, now, good); err != nil {
		t.Errorf("within leeway rejected: %v", err)
	}
}

// ------------------------------------------------------------ authorizeRoleStrict

func TestAuthorizeRoleStrict(t *testing.T) {
	const prefix = "arn:aws:iam::111111111111:role/myorg_"
	arn := func(name string) string { return "arn:aws:iam::111111111111:role/" + name }
	claims := baseClaims(time.Now())
	claims["groups"] = []interface{}{"myorg_msd", "myorg_cytometry"}

	allow := []struct{ name, role, support string }{
		{"exact client role", arn("myorg_flow-cytometry_researcher"), ""},
		{"group guest", arn("myorg_msd_guest"), ""},
		{"support group", arn("myorg_qpcr_admin"), "myorg_cytometry"},
	}
	for _, c := range allow {
		if err := authorizeRoleStrict(claims, c.role, prefix, "works", c.support); err != nil {
			t.Errorf("%s: denied: %v", c.name, err)
		}
	}

	deny := []struct{ name, role, prefix string }{
		{"guest of a different project", arn("myorg_flow-cytometry_guest"), prefix},
		{"role name with an extra leading prefix", arn("xmyorg_flow-cytometry_researcher"), "arn:aws:iam::111111111111:role/"},
		{"other role of the same project", arn("myorg_flow-cytometry_admin"), prefix},
		{"guest of a project with no group", arn("myorg_qpcr_guest"), prefix},
		{"outside the prefix", arn("otherorg_msd_guest"), prefix},
		{"other account", "arn:aws:iam::222222222222:role/myorg_msd_guest", prefix},
		{"not an ARN", "myorg_msd_guest", prefix},
		{"no prefix configured", arn("myorg_msd_guest"), ""},
	}
	for _, c := range deny {
		if err := authorizeRoleStrict(claims, c.role, c.prefix, "works", ""); err == nil {
			t.Errorf("%s: allowed %s", c.name, c.role)
		}
	}

	// Suffix look-alikes: one org's name is a suffix of another's. The
	// default mode's HasSuffix matching would grant both of these.
	lookalike := map[string]interface{}{
		"groups": []interface{}{"org_qpcr"},
		"resource_access": map[string]interface{}{
			"works": map[string]interface{}{"roles": []interface{}{"org_msd_researcher"}},
		},
	}
	for _, role := range []string{arn("myorg_qpcr_guest"), arn("myorg_msd_researcher")} {
		if err := authorizeRoleStrict(lookalike, role, prefix, "works", ""); err == nil {
			t.Errorf("suffix look-alike allowed %s", role)
		}
	}
	if err := authorizeRoleStrict(claims, arn("myorg_msd_guest"), prefix, "works", "myorg"); err != nil {
		t.Errorf("unrelated support group changed a normal grant: %v", err)
	}
}

// ------------------------------------------------------------ JWKS fetch + cache

func jwksServer(t *testing.T, key *rsa.PublicKey, kid string, hits *int32) *httptest.Server {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(hits, 1)
		switch r.URL.Path {
		case "/realms/test/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(map[string]string{"issuer": srv.URL + "/realms/test", "jwks_uri": srv.URL + "/certs"})
		case "/certs":
			json.NewEncoder(w).Encode(map[string]interface{}{"keys": []map[string]string{
				{"kid": "enc", "kty": "RSA", "use": "enc", "n": b64(key.N.Bytes()), "e": "AQAB"},
				{"kid": kid, "kty": "RSA", "use": "sig", "n": b64(key.N.Bytes()), "e": b64(big.NewInt(int64(key.E)).Bytes())},
			}})
		default:
			http.NotFound(w, r)
		}
	}))
	return srv
}

func TestJWKSCache(t *testing.T) {
	key := newKey(t)
	var hits int32
	srv := jwksServer(t, &key.PublicKey, "k1", &hits)
	defer srv.Close()
	iss := srv.URL + "/realms/test"
	c := &jwksCache{}
	now := time.Now()

	if _, err := c.key(iss, "k1", now); err != nil {
		t.Fatalf("first fetch: %v", err)
	}
	if _, err := c.key(iss, "k1", now.Add(time.Minute)); err != nil || atomic.LoadInt32(&hits) != 2 {
		t.Fatalf("cached lookup refetched (hits=%d, err=%v)", hits, err)
	}
	if _, err := c.key(iss, "enc", now); err == nil {
		t.Error("an encryption key was usable for signatures")
	}
	// unknown kid refetches at most once per jwksMinRefetch
	before := atomic.LoadInt32(&hits)
	for i := 0; i < 5; i++ {
		c.key(iss, "nope", now.Add(time.Second))
	}
	if got := atomic.LoadInt32(&hits) - before; got > 2 {
		t.Errorf("unknown kid caused %d fetches in the refetch window", got)
	}
	// IdP down: a cached key keeps working past its max age
	srv.Close()
	if _, err := c.key(iss, "k1", now.Add(jwksMaxAge+time.Minute)); err != nil {
		t.Errorf("cached key dropped during an IdP outage: %v", err)
	}
	// discovery issuer mismatch is refused
	var h2 int32
	srv2 := jwksServer(t, &key.PublicKey, "k1", &h2)
	defer srv2.Close()
	if _, err := fetchSigningKeys(srv2.URL + "/realms/test-other"); err == nil {
		t.Error("discovery for a different issuer accepted")
	}
}

// ------------------------------------------------------------ end to end (STS stubbed)

type fakeSTS struct {
	stsiface.STSAPI
	got *sts.AssumeRoleInput
	err error
}

func (f *fakeSTS) AssumeRole(in *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
	f.got = in
	if f.err != nil {
		return nil, f.err
	}
	return &sts.AssumeRoleOutput{Credentials: &sts.Credentials{
		AccessKeyId: aws.String("ASIATEST"), SecretAccessKey: aws.String("secret"), SessionToken: aws.String("session"),
	}}, nil
}

func TestGetSessionFromTokenAssumeRole(t *testing.T) {
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
	claims["preferred_username"] = "alice smith/../x"
	tok := signToken(t, key, "RS256", "k1", claims)
	role := "arn:aws:iam::111111111111:role/myorg_flow-cytometry_researcher"
	prefix := "arn:aws:iam::111111111111:role/myorg_"

	ak, sk, st, err := getSessionFromTokenAssumeRole(tok, role, "works", "", iss, prefix, "us-east-1", 43200)
	if err != nil || ak != "ASIATEST" || sk != "secret" || st != "session" {
		t.Fatalf("got %q %q %q %v", ak, sk, st, err)
	}
	if aws.StringValue(fake.got.RoleArn) != role {
		t.Errorf("assumed %s", aws.StringValue(fake.got.RoleArn))
	}
	if d := aws.Int64Value(fake.got.DurationSeconds); d != maxChainedSessionSeconds {
		t.Errorf("duration %d not capped to %d", d, maxChainedSessionSeconds)
	}
	if n := aws.StringValue(fake.got.RoleSessionName); n != "alice-smith-..-x" {
		t.Errorf("session name %q not sanitised", n)
	}

	// no STS call at all when the token or the role is refused
	fake.got = nil
	if _, _, _, err := getSessionFromTokenAssumeRole(tok, "arn:aws:iam::111111111111:role/myorg_flow-cytometry_admin", "works", "", iss, prefix, "", 900); err == nil || fake.got != nil {
		t.Error("unauthorised role reached STS")
	}
	forged := signToken(t, newKey(t), "RS256", "k1", claims)
	if _, _, _, err := getSessionFromTokenAssumeRole(forged, role, "works", "", iss, prefix, "", 900); err == nil || fake.got != nil {
		t.Error("forged token reached STS")
	}
	// STS failure surfaces as an error
	fake.err = fmt.Errorf("AccessDenied")
	if _, _, _, err := getSessionFromTokenAssumeRole(tok, role, "works", "", iss, prefix, "", 900); err == nil {
		t.Error("STS error swallowed")
	}
	// missing username falls back instead of panicking
	delete(claims, "preferred_username")
	if n := sessionName(claims); n != "filestash" {
		t.Errorf("fallback session name %q", n)
	}
}
