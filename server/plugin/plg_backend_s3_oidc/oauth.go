package plg_backend_s3_oidc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	slog "log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"

	. "github.com/mickael-kerjean/filestash/server/common"
)

// Get base URL from a request
func getBaseURL(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + req.Host
}

// POST a new session to the backend
func postSession(url string, body map[string]string) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v\n", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("Error creating request: %v\n", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("X-Requested-With", "XmlHttpRequest")
	req.Header.Set("Referrer", "http://localhost:8334/login")
	req.Header.Set("Referrer-Policy", "same-origin")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error making request: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Request failed with status: %s\n", resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v\n", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("Error unmarshalling response body: %v\n", err)
	}
	fmt.Printf("postSession response: %v\n", response)

	return nil
}

// getSTSwTest runs getSTS to get S3 STS token, then creates static credentials
func getSTSwTest(
	kcToken, roleArn,
	roleSessionName string, durationSeconds int64,
) (string, string, string, error) {
	// Exchange KeyCloak token for S3 STS token
	accessKey, secretKey, sessionToken, err := getSTS(kcToken, roleArn, roleSessionName, durationSeconds)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get STS token: %v", err)
	}

	// Create a new session with the temporary credentials
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(
			accessKey,     // Access key ID
			secretKey,     // Secret access key
			sessionToken), // Session token
	})
	if err != nil {
		slog.Fatalf("Failed to create session: %v", err)
	}

	slog.Printf("Access Key: %s", accessKey)
	slog.Printf("Secret Key: %s", secretKey)
	slog.Printf("Session Token: %s", sessionToken)

	// Create an S3 client using the session
	s3Client := s3.New(sess)

	// Example: List buckets
	result, err := s3Client.ListBuckets(nil)
	if err != nil {
		slog.Fatalf("Failed to list buckets: %v", err)
	}

	// Print the bucket names
	fmt.Println("Buckets:")
	for _, bucket := range result.Buckets {
		fmt.Printf("* %s\n", aws.StringValue(bucket.Name))
	}
	return accessKey, secretKey, sessionToken, nil
}

// getSTS exchanges a KeyCloak token for S3 STS token
func getSTS(
	kcToken, roleArn, roleSessionName string, durationSeconds int64,
) (string, string, string, error) {
	// Create a new session
	sess, err := session.NewSession()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create session: %v", err)
	}

	// Create a new STS client
	stsClient := sts.New(sess)

	// Assume role with web identity
	input := &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(roleArn),
		RoleSessionName:  aws.String(roleSessionName),
		WebIdentityToken: aws.String(kcToken),
		DurationSeconds:  aws.Int64(durationSeconds),
	}

	result, err := stsClient.AssumeRoleWithWebIdentity(input)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to assume role with web identity: %v", err)
	}

	// Extract credentials
	credentials := result.Credentials
	accessKey := aws.StringValue(credentials.AccessKeyId)
	secretKey := aws.StringValue(credentials.SecretAccessKey)
	sessionToken := aws.StringValue(credentials.SessionToken)

	// Log token expiration time
	// expiryTime := aws.TimeValue(credentials.Expiration).Format(time.RFC3339)
	// slog.Printf("STS token expire time: %s", expiryTime)

	return accessKey, secretKey, sessionToken, nil
}

// clientCredGrant gets a KeyCloak token for client workss3
// Using client credentials grant
func clientCredGrant() (string, error) {
	// Keycloak token endpoint
	tokenEndpoint := "https://accounts.datajoint.com/realms/datajoint/protocol/openid-connect/token"

	// Prepare the form data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", "workss3")
	data.Set("client_secret", "your-client-secret")

	// Create the HTTP request
	req, err := http.NewRequest("POST", tokenEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 status code: %d, body: %s", resp.StatusCode, body)
	}

	// Parse the JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// Extract the access token
	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access token not found in response")
	}

	return accessToken, nil
}

// exchangeToken exchanges a KeyCloak token for client works,
// for a KeyCloak token for client workss3
// See https://www.keycloak.org/securing-apps/token-exchange
func exchangeToken(
	kcToken string, clientID string, clientSecret string,
	clientAud string, tokenEndpoint string,
) (string, error) {
	// Keycloak token endpoint
	if clientID == "" || clientSecret == "" || clientAud == "" {
		return "", fmt.Errorf("clientID, clientSecret, or clientAud not set")
	}

	// Prepare the form data
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("subject_token", kcToken)
	data.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
	data.Set("requested_token_type", "urn:ietf:params:oauth:token-type:access_token")
	data.Set("audience", clientAud)

	// Create the HTTP request
	req, err := http.NewRequest("POST", tokenEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 status code: %d, body: %s", resp.StatusCode, body)
	}

	// Parse the JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// Extract the access token
	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access token not found in response")
	}

	return accessToken, nil
}

// decodeToken decodes a Keycloak access token
func decodeToken(kcToken string) (map[string]interface{}, error) {
	// Split the token into parts
	parts := strings.Split(kcToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}
	// Decode the payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return map[string]interface{}{"error": "failed to decode token payload"}, err
	}

	// Unmarshal the JSON payload
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token payload: %v", err)
	}
	return claims, nil
}

func parseRoleURN(urn string) (string, string, error) {
	parts := strings.Split(urn, "_")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid role URN format")
	}
	roleName := parts[len(parts)-1]
	project := strings.Join(parts[:len(parts)-1], "_")
	return project, roleName, nil
}

// contains checks if a slice contains a specific element
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Check if user can assume the role, based on the KeyCloak roles or groups
// defined in the token
func canAssumeRole(
	decodedKcToken map[string]interface{},
	roleArn string, supportGroupName string,
) bool {
	return canAssumeRoleFromResourceAccess(decodedKcToken, roleArn) ||
		canAssumeRoleFromGroups(decodedKcToken, roleArn, supportGroupName)
}

// Check if user can assume the role, based on the KeyCloak groups defined
// in the token. Also implements access for support group.
// This is to support legacy roles: if user is in group "dev_my-project", grant
// them access to role "dev_my-project_guest"
func canAssumeRoleFromGroups(
	decodedKcToken map[string]interface{},
	roleArn string, supportGroupName string,
) bool {
	groups, ok := decodedKcToken["groups"].([]interface{})
	if !ok {
		return false
	}
	log.Printf("User '%s' attempting to use role '%s' with groups: %v", decodedKcToken["preferred_username"], roleArn, groups)

	for _, group := range groups {
		// Perform a type assertion to convert each element to a string
		grpStr, ok := group.(string)
		if !ok {
			log.Printf("ERROR: Could not cast group to string: %v", group)
			continue
		}

		// Check for designated support group
		if supportGroupName != "" && grpStr == supportGroupName {
			return true
		}

		if strings.HasSuffix(roleArn, grpStr+"_guest") {
			return true
		}
	}

	return false
}

// Check if user can assume the role, based on the KeyCloak roles defined
// in the token
func canAssumeRoleFromResourceAccess(decodedKcToken map[string]interface{}, roleArn string) bool {
	resourceAccess, ok := decodedKcToken["resource_access"].(map[string]interface{})
	if !ok {
		return false
	}

	works, ok := resourceAccess["works"].(map[string]interface{})
	if !ok {
		return false
	}

	roles, ok := works["roles"].([]interface{})
	if !ok {
		return false
	}

	log.Printf("User '%s' attempting to use '%s' role with resource_access: %v",
		decodedKcToken["preferred_username"], roleArn, roles)
	recognizedRoles := []string{"researcher", "guest"}

	// Check if roleArn is in the roles slice
	for _, role := range roles {
		// Perform a type assertion to convert each element to a string
		roleStr, ok := role.(string)
		if !ok {
			log.Printf("ERROR: Could not cast role to string: %v", role)
			continue
		}
		if strings.HasSuffix(roleArn, roleStr) {
			return true
		}

		project, roleName, err := parseRoleURN(roleStr)
		if err != nil {
			log.Printf("ERROR: Could not parse role URN: %s", roleStr)
			continue
		}
		// Error if roleName is not recognized
		if !contains(recognizedRoles, roleName) {
			log.Printf("ERROR: Unrecognized role: %s", roleName)
			continue
		}
		if roleName == "reseacher" && strings.HasSuffix(roleArn, project+"_guest") {
			return true
		}
	}

	return false
}

// getSessionFromToken uses a KeyCloak access token to get a temporary
// AWS credentials for S3 access
func getSessionFromToken(
	kcToken string, roleArn string, clientID string, clientSecret string,
	clientAud string, durationSeconds int64, tokenEndpoint string,
	supportGroupName string,
) (string, string, string, error) {
	// 1. Decode the Keycloak token
	decodedKcToken, decodeErr := decodeToken(kcToken)
	if decodeErr != nil {
		slog.Printf("ERROR: Error decoding Keycloak token: %v", decodeErr)
		return "", "", "", ErrNotValid
	}

	// 2. Check that the user has permission to assume this role, i.e. that
	// the role ARN ends with one of the user's KeyCloak roles
	if !canAssumeRole(decodedKcToken, roleArn, supportGroupName) {
		slog.Printf("ERROR: User does not have permission to assume role: %s", roleArn)
		return "", "", "", ErrPermissionDenied
	}

	// 3. Exchange for workss3 token
	s3Token, exchangeErr := exchangeToken(kcToken, clientID, clientSecret, clientAud, tokenEndpoint)
	if exchangeErr != nil {
		slog.Printf("ERROR: Error getting token from token exchange flow: %v", exchangeErr)
		return "", "", "", ErrPermissionDenied
	}

	// 3. Client credentials grant
	// As an alternative to token exchange
	// ccToken, ccErr := clientCredGrant()
	// if ccErr != nil {
	// 	return "", "", "", fmt.Errorf("Error getting token from client credentials flow: %v", ccErr)
	// }
	// slog.Printf("workss3 client credentials token: %s", ccToken)

	// 4. Exchange for S3 STS token
	roleSessionName := decodedKcToken["preferred_username"].(string)
	if roleSessionName == "" {
		roleSessionName = "default"
	}
	accessKey, secretKey, sessionToken, err := getSTS(
		s3Token, roleArn, roleSessionName, durationSeconds,
	)
	if err != nil {
		slog.Printf("ERROR: Error getting STS token: %v", err)
		return "", "", "", ErrPermissionDenied
	}

	return accessKey, secretKey, sessionToken, nil
}
