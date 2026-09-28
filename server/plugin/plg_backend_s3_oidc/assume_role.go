package plg_backend_s3_oidc

// Credential mode "assume_role".
//
// The default mode ("web_identity") exchanges the user's token for a second
// audience and hands it to STS AssumeRoleWithWebIdentity, so AWS validates
// the token against the IdP's published signing keys. That needs STS to reach
// the IdP from the public internet. Deployments whose IdP is reachable only
// on a private network -- or whose account denies web-identity federation --
// cannot use it.
//
// In "assume_role" mode the plugin does the validation itself:
//
//  1. verify the token's signature against the IdP's JWKS (fetched from the
//     issuer over the network the pod is on) and check iss / exp / azp;
//  2. authorise the requested role by EXACT match against the verified
//     claims -- there is no AWS trust condition behind this check, so the
//     suffix matching the default mode tolerates is not acceptable here;
//  3. call plain sts:AssumeRole into that role with the pod's own
//     credentials (e.g. IRSA). The target role must trust the pod's role.
//
// Nothing here runs unless FS_S3_OIDC_CREDENTIAL_MODE=assume_role.

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	. "github.com/mickael-kerjean/filestash/server/common"
)

const (
	credentialModeWebIdentity = "web_identity"
	credentialModeAssumeRole  = "assume_role"

	// A role assumed with credentials that are themselves a role session is
	// "role chaining", which AWS caps at one hour.
	maxChainedSessionSeconds = 3600

	tokenClockLeeway = 30 * time.Second
	jwksMaxAge       = 10 * time.Minute
	jwksMinRefetch   = 30 * time.Second
)

var roleArnPattern = regexp.MustCompile(`^arn:aws[a-z-]*:iam::[0-9]{12}:role/[\w+=,.@/-]+$`)
var sessionNameUnsafe = regexp.MustCompile(`[^\w+=,.@-]`)

// parseCredentialMode maps the configured value to a mode. Unset means the
// default; anything unrecognised is an error, so a typo fails closed instead
// of silently falling back.
func parseCredentialMode(v string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", credentialModeWebIdentity:
		return credentialModeWebIdentity, nil
	case credentialModeAssumeRole:
		return credentialModeAssumeRole, nil
	}
	return "", fmt.Errorf("unknown credential mode %q (want %q or %q)", v, credentialModeWebIdentity, credentialModeAssumeRole)
}

// issuerFromTokenEndpoint derives the issuer from a Keycloak-style token
// endpoint (<issuer>/protocol/openid-connect/token) unless an explicit
// issuer is configured.
func issuerFromTokenEndpoint(tokenEndpoint, override string) (string, error) {
	if o := strings.TrimRight(strings.TrimSpace(override), "/"); o != "" {
		return o, nil
	}
	const suffix = "/protocol/openid-connect/token"
	ep := strings.TrimSpace(tokenEndpoint)
	if !strings.HasSuffix(ep, suffix) {
		return "", fmt.Errorf("cannot derive issuer from token endpoint %q; set the issuer explicitly", ep)
	}
	return strings.TrimSuffix(ep, suffix), nil
}

// ---------------------------------------------------------------- JWKS

type jwksCache struct {
	mu          sync.Mutex
	issuer      string
	keys        map[string]*rsa.PublicKey
	fetched     time.Time
	lastAttempt time.Time
}

var signingKeys = &jwksCache{}
var httpClient = &http.Client{Timeout: 10 * time.Second}

func (c *jwksCache) key(issuer, kid string, now time.Time) (*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	sameIssuer := c.issuer == issuer
	k, known := c.keys[kid]
	known = known && sameIssuer
	if known && now.Sub(c.fetched) <= jwksMaxAge {
		return k, nil
	}
	// Refetch when the keys are stale or the kid is unknown (key rotation),
	// but at most once per jwksMinRefetch, so bad tokens or an IdP outage
	// cannot turn the plugin into a load generator against the IdP.
	if !sameIssuer || c.lastAttempt.IsZero() || now.Sub(c.lastAttempt) >= jwksMinRefetch {
		c.lastAttempt = now
		keys, err := fetchSigningKeys(issuer)
		if err == nil {
			c.issuer, c.keys, c.fetched = issuer, keys, now
			k, known = keys[kid]
		} else if !known {
			return nil, err
		} // else: keep serving the cached key through an IdP blip
	}
	if !known {
		return nil, fmt.Errorf("no signing key with kid %q", kid)
	}
	return k, nil
}

func getJSON(url string, out interface{}) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return json.Unmarshal(body, out)
}

func fetchSigningKeys(issuer string) (map[string]*rsa.PublicKey, error) {
	var disc struct {
		Issuer  string `json:"issuer"`
		JwksURI string `json:"jwks_uri"`
	}
	if err := getJSON(issuer+"/.well-known/openid-configuration", &disc); err != nil {
		return nil, fmt.Errorf("discovery: %v", err)
	}
	if disc.Issuer != issuer {
		return nil, fmt.Errorf("discovery issuer %q != expected %q", disc.Issuer, issuer)
	}
	if disc.JwksURI == "" {
		return nil, fmt.Errorf("discovery document has no jwks_uri")
	}
	var set struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := getJSON(disc.JwksURI, &set); err != nil {
		return nil, fmt.Errorf("jwks: %v", err)
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range set.Keys {
		if k.Kty != "RSA" || (k.Use != "" && k.Use != "sig") || k.Kid == "" {
			continue
		}
		n, err1 := base64.RawURLEncoding.DecodeString(k.N)
		e, err2 := base64.RawURLEncoding.DecodeString(k.E)
		if err1 != nil || err2 != nil || len(e) == 0 || len(e) > 4 {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("jwks has no usable RSA signing keys")
	}
	return keys, nil
}

// ---------------------------------------------------------------- token

var rsaAlgs = map[string]crypto.Hash{"RS256": crypto.SHA256, "RS384": crypto.SHA384, "RS512": crypto.SHA512}

func digest(h crypto.Hash, data []byte) []byte {
	switch h {
	case crypto.SHA384:
		s := sha512.Sum384(data)
		return s[:]
	case crypto.SHA512:
		s := sha512.Sum512(data)
		return s[:]
	}
	s := sha256.Sum256(data)
	return s[:]
}

// verifyToken checks signature, issuer, expiry and authorised party, and
// returns the claims only if all pass.
func verifyToken(token, issuer string, allowedAzp []string, now time.Time,
	keyFor func(kid string) (*rsa.PublicKey, error)) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed token")
	}
	hb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("malformed header")
	}
	var hdr struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(hb, &hdr); err != nil {
		return nil, fmt.Errorf("malformed header")
	}
	hash, ok := rsaAlgs[hdr.Alg]
	if !ok {
		return nil, fmt.Errorf("unsupported alg %q", hdr.Alg)
	}
	if hdr.Kid == "" {
		return nil, fmt.Errorf("token has no kid")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("malformed signature")
	}
	pub, err := keyFor(hdr.Kid)
	if err != nil {
		return nil, err
	}
	if err := rsa.VerifyPKCS1v15(pub, hash, digest(hash, []byte(parts[0]+"."+parts[1])), sig); err != nil {
		return nil, fmt.Errorf("bad signature")
	}

	claims, err := decodeToken(token)
	if err != nil {
		return nil, err
	}
	if iss, _ := claims["iss"].(string); iss != issuer {
		return nil, fmt.Errorf("issuer %q != %q", iss, issuer)
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("token has no exp")
	}
	if now.After(time.Unix(int64(exp), 0).Add(tokenClockLeeway)) {
		return nil, fmt.Errorf("token expired")
	}
	if nbf, ok := claims["nbf"].(float64); ok && now.Add(tokenClockLeeway).Before(time.Unix(int64(nbf), 0)) {
		return nil, fmt.Errorf("token not yet valid")
	}
	azp, _ := claims["azp"].(string)
	if !contains(allowedAzp, azp) {
		return nil, fmt.Errorf("authorised party %q not allowed", azp)
	}
	return claims, nil
}

// ---------------------------------------------------------------- authorisation

func stringsAt(m map[string]interface{}, path ...string) []string {
	var cur interface{} = m
	for _, p := range path {
		mm, ok := cur.(map[string]interface{})
		if !ok {
			return nil
		}
		cur = mm[p]
	}
	arr, _ := cur.([]interface{})
	out := []string{}
	for _, v := range arr {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// authorizeRoleStrict is the exact-match form of canAssumeRole. A role is
// allowed when, in the verified claims:
//   - resource_access.<client>.roles contains the role name exactly, or
//   - groups contains <group> and the role name is exactly "<group>_guest", or
//   - groups contains the configured support group.
//
// The role ARN must also start with the configured prefix, which bounds the
// roles this deployment will ever request (the pod role's own policy is the
// AWS-side bound).
func authorizeRoleStrict(claims map[string]interface{}, roleArn, rolePrefix, clientID, supportGroup string) error {
	if rolePrefix == "" {
		return fmt.Errorf("no role ARN prefix configured")
	}
	if !roleArnPattern.MatchString(roleArn) || !strings.HasPrefix(roleArn, rolePrefix) {
		return fmt.Errorf("role %q is outside the allowed prefix", roleArn)
	}
	roleName := roleArn[strings.LastIndex(roleArn, "/")+1:]
	for _, r := range stringsAt(claims, "resource_access", clientID, "roles") {
		if r == roleName {
			return nil
		}
	}
	for _, g := range stringsAt(claims, "groups") {
		if supportGroup != "" && g == supportGroup {
			return nil
		}
		if g+"_guest" == roleName {
			return nil
		}
	}
	return fmt.Errorf("no claim grants role %q", roleName)
}

// ---------------------------------------------------------------- STS

var newSTSClient = func(region string) stsiface.STSAPI {
	cfg := aws.Config{}
	if region != "" {
		cfg.Region = aws.String(region)
	}
	// Default credential chain: on EKS with IRSA this is the pod's role.
	return sts.New(session.Must(session.NewSession(&cfg)))
}

func sessionName(claims map[string]interface{}) string {
	u, _ := claims["preferred_username"].(string)
	u = sessionNameUnsafe.ReplaceAllString(u, "-")
	if len(u) > 64 {
		u = u[:64]
	}
	if len(u) < 2 {
		u = "filestash"
	}
	return u
}

func getSessionFromTokenAssumeRole(
	kcToken, roleArn, clientID, supportGroup, issuer, rolePrefix, region string,
	durationSeconds int64,
) (string, string, string, error) {
	now := time.Now()
	claims, err := verifyToken(kcToken, issuer, []string{clientID}, now,
		func(kid string) (*rsa.PublicKey, error) { return signingKeys.key(issuer, kid, now) })
	if err != nil {
		log.Printf("ERROR: token rejected: %v", err)
		return "", "", "", ErrPermissionDenied
	}
	if err := authorizeRoleStrict(claims, roleArn, rolePrefix, clientID, supportGroup); err != nil {
		log.Printf("ERROR: user %q not authorised: %v", claims["preferred_username"], err)
		return "", "", "", ErrPermissionDenied
	}
	if durationSeconds > maxChainedSessionSeconds {
		log.Printf("WARN: sts_duration %d capped to %d (role chaining limit)", durationSeconds, maxChainedSessionSeconds)
		durationSeconds = maxChainedSessionSeconds
	}
	out, err := newSTSClient(region).AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(sessionName(claims)),
		DurationSeconds: aws.Int64(durationSeconds),
	})
	if err != nil {
		log.Printf("ERROR: AssumeRole %s: %v", roleArn, err)
		return "", "", "", ErrPermissionDenied
	}
	c := out.Credentials
	return aws.StringValue(c.AccessKeyId), aws.StringValue(c.SecretAccessKey), aws.StringValue(c.SessionToken), nil
}
