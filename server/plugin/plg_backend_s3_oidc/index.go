package plg_backend_s3_oidc

import (
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	. "github.com/mickael-kerjean/filestash/server/common"

	plg_s3 "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_s3"

	"io"
	"os"
)

var S3Cache AppCache

type S3OIDCBackend struct {
	plg_s3.S3Backend
}

func init() {
	Backend.Register("s3_oidc", S3OIDCBackend{})
	S3Cache = NewAppCache(2, 1)

	Hooks.Register.Onload(func() {
		loadS3OIDCConfig()
	})
}

func getConfigParams(baseName string, paramNames []string) (map[string]string, error) {
	keys := []string{}
	for _, p := range paramNames {
		keys = append(keys, baseName+"."+p)
	}
	out := map[string]string{}
	for _, k := range keys {
		v := Config.Get(k).String()
		if v == "" {
			log.Printf("ERROR: Missing config param: %s.%s", baseName, k)
			return nil, ErrNotValid
		}
		out[k] = strings.TrimSpace(v)
	}
	return out, nil
}

func constructPath(bucket string, project string) string {
	return "/" + bucket + "/" + project + "/"
}

// Get project name from role ARN
// e.g. arn:aws:iam::111111111111:role/debug_invite-user-test_researcher
// returns "debug_invite-user-test"
func getProjectFromRoleArn(roleArn string) string {
	parts := strings.Split(roleArn, "/")
	if len(parts) < 2 {
		log.Printf("ERROR: Unable to parse role ARN: %s", roleArn)
		return ""
	}
	parts = strings.Split(parts[1], "_")
	if len(parts) < 2 {
		log.Printf("ERROR: Unable to parse role ARN: %s", roleArn)
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "_")
}

func useProjectGuestRole(project string, roleArn string) string {
	if project == "" {
		return roleArn
	}
	parts := strings.Split(roleArn, "/")
	if len(parts) < 2 {
		log.Printf("ERROR: Unable to parse role ARN: %s", roleArn)
		return roleArn
	}
	return parts[0] + "/" + project + "_guest"
}

func (this S3OIDCBackend) Init(params map[string]string, app *App) (IBackend, error) {
	// Clean and validate params
	for k, v := range params {
		params[k] = strings.TrimSpace(v)
	}
	requiredParams := []string{"token"}
	for _, p := range requiredParams {
		if params[p] == "" {
			return nil, ErrNotValid
		}
	}

	// Use config value for role ARN if not provided
	if params["role_arn"] == "" {
		params["role_arn"] = strings.TrimSpace(Config.Get("s3_oidc.role_arn").String())
		if params["project"] != "" {
			// If project is passed but role ARN was not, then use the default
			// role ARN as a template for role name {project}_guest
			params["role_arn"] = useProjectGuestRole(params["project"], params["role_arn"])
		}
	}

	// Assemble params from config
	configParamNames := []string{
		"client_id", "client_secret",
		"sts_duration", "client_audience", "token_endpoint", "support_group",
	}
	configParams, err := getConfigParams("s3_oidc", configParamNames)
	if err != nil {
		return nil, err
	}
	durationSeconds, err := strconv.ParseInt(configParams["s3_oidc.sts_duration"], 10, 64)
	if err != nil {
		log.Printf("ERROR: Error parsing STS duration: %s", err)
		return nil, ErrNotValid
	}

	// Use config value for path if not provided
	project := getProjectFromRoleArn(params["role_arn"])
	if params["bucket"] == "" {
		params["bucket"] = Config.Get("s3_oidc.bucket").String()
	}
	if (params["path"] == "" || params["path"] == "/") && project != "" {
		params["path"] = constructPath(params["bucket"], project)
	}
	// log.Printf("project: %s, path: %s", project, params["path"])

	// Use config value for region if not provided
	if params["region"] == "" {
		params["region"] = strings.TrimSpace(Config.Get("s3_oidc.region").String())
	}

	// Get S3 credentials from token
	accessKey, secretKey, sessionToken, err := getSessionFromToken(
		params["token"], params["role_arn"],
		configParams["s3_oidc.client_id"], configParams["s3_oidc.client_secret"],
		configParams["s3_oidc.client_audience"], durationSeconds,
		configParams["s3_oidc.token_endpoint"], configParams["s3_oidc.support_group"],
	)
	if err != nil {
		return nil, err
	}

	// Set the parameters for the S3 backend
	params["type"] = "s3"
	params["access_key_id"] = accessKey
	params["secret_access_key"] = secretKey
	params["session_token"] = sessionToken
	params["session_role_arn"] = params["role_arn"]
	// params["role_arn"] = ""

	return this.S3Backend.Init(params, app)
}

func (this S3OIDCBackend) LoginForm() Form {
	return Form{
		Elmnts: []FormElement{
			FormElement{
				Name:  "type",
				Type:  "hidden",
				Value: "s3_oidc",
			},
			FormElement{
				Name:        "token",
				Type:        "password",
				Placeholder: "DataJoint Works KeyCloak access token*",
			},
			FormElement{
				Id:          "s3_role_arn",
				Name:        "role_arn",
				Type:        "text",
				Placeholder: "Role ARN*",
			},
			FormElement{
				Name:        "advanced",
				Type:        "enable",
				Placeholder: "Advanced",
				Target: []string{
					"s3_region", "s3_endpoint", "s3_session_token",
					"s3_path", "s3_encryption_key", "s3_number_thread",
				},
			},
			FormElement{
				Id:          "s3_region",
				Name:        "region",
				Type:        "text",
				Placeholder: "Region",
			},
			FormElement{
				Id:          "s3_endpoint",
				Name:        "endpoint",
				Type:        "text",
				Placeholder: "Endpoint",
			},
			FormElement{
				Id:          "s3_path",
				Name:        "path",
				Type:        "text",
				Placeholder: "Path",
			},
			FormElement{
				Id:          "s3_encryption_key",
				Name:        "encryption_key",
				Type:        "text",
				Placeholder: "Encryption Key",
			},
			FormElement{
				Id:          "s3_number_thread",
				Name:        "number_thread",
				Type:        "text",
				Placeholder: "Num. Thread",
			},
		},
	}
}

func (this S3OIDCBackend) Meta(path string) Metadata {
	return this.S3Backend.Meta(path)
}

func (this S3OIDCBackend) Ls(path string) (files []os.FileInfo, err error) {
	return this.S3Backend.Ls(path)
}

func (this S3OIDCBackend) Cat(path string) (io.ReadCloser, error) {
	return this.S3Backend.Cat(path)
}

func (this S3OIDCBackend) Mkdir(path string) error {
	return this.S3Backend.Mkdir(path)
}

func (this S3OIDCBackend) Rm(path string) error {
	return this.S3Backend.Rm(path)
}

func (this S3OIDCBackend) Mv(from string, to string) error {
	return this.S3Backend.Mv(from, to)
}

func (this S3OIDCBackend) Touch(path string) error {
	return this.S3Backend.Touch(path)
}

func (this S3OIDCBackend) Save(path string, file io.Reader) error {
	return this.S3Backend.Save(path, file)
}

func (this S3OIDCBackend) createSession(bucket string) *session.Session {
	return this.S3Backend.CreateSession(bucket)
}

func (s S3OIDCBackend) path(p string) plg_s3.S3Path {
	return s.S3Backend.Path(p)
}
