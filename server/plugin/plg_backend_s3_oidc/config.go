package plg_backend_s3_oidc

import (
	"fmt"
	"os"

	. "github.com/mickael-kerjean/filestash/server/common"
)

var client_id = func() string {
	return Config.Get("s3_oidc.client_id").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "client_id"
		f.Type = "text"
		f.Description = "Client ID for OIDC (required)"
		f.Default = os.Getenv("FS_S3_OIDC_CLIENT_ID")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = "works"
		}
		return f
	}).String()
}

var client_secret = func() string {
	return Config.Get("s3_oidc.client_secret").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "client_secret"
		f.Type = "password"
		f.Description = "Client Secret for OIDC (required)"
		f.Default = os.Getenv("FS_S3_OIDC_CLIENT_SECRET")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default set in environment")
		} else {
			f.Placeholder = ""
		}
		return f
	}).String()
}

var client_audience = func() string {
	return Config.Get("s3_oidc.client_audience").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "client_audience"
		f.Type = "text"
		f.Description = "Client Audience for OIDC (required)"
		f.Default = os.Getenv("FS_S3_OIDC_CLIENT_AUD")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = "workss3"
		}
		return f
	}).String()
}

var token_endpoint = func() string {
	return Config.Get("s3_oidc.token_endpoint").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "token_endpoint"
		f.Type = "text"
		f.Description = "Endpoint for OIDC token exchange (required)"
		f.Default = os.Getenv("FS_S3_OIDC_TOKEN_ENDPOINT")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = "https://keycloak-qa.datajoint.io/realms/datajoint/protocol/openid-connect/token"
		}
		return f
	}).String()
}

var role_arn = func() string {
	return Config.Get("s3_oidc.role_arn").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "role_arn"
		f.Type = "text"
		f.Description = "Default role ARN for OIDC"
		f.Default = os.Getenv("FS_S3_OIDC_DEFAULT_ROLE_ARN")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = "arn:aws:iam::111111111111:role/debug_invite-user-test_researcher"
		}
		return f
	}).String()
}

var support_group = func() string {
	return Config.Get("s3_oidc.support_group").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "support_group"
		f.Type = "text"
		f.Description = "Name of the support group that allows assumption of any role ARN"
		f.Default = os.Getenv("FS_S3_OIDC_SUPPORT_GROUP")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = "datajoint"
		}
		return f
	}).String()
}

var bucket = func() string {
	return Config.Get("s3_oidc.bucket").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "bucket"
		f.Type = "text"
		f.Description = "Default bucket name if `path` is not provided"
		f.Default = os.Getenv("FS_S3_OIDC_DEFAULT_BUCKET")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = "dj-sciops"
		}
		return f
	}).String()
}

var region = func() string {
	return Config.Get("s3_oidc.region").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "region"
		f.Type = "text"
		f.Description = "Default bucket region if `region` is not provided"
		f.Default = os.Getenv("FS_S3_OIDC_DEFAULT_REGION")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = "us-east-2"
		}
		return f
	}).String()
}

var sts_duration = func() string {
	return Config.Get("s3_oidc.sts_duration").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "sts_duration"
		f.Type = "text"
		f.Description = "Duration in seconds for the STS token"
		f.Default = os.Getenv("FS_S3_OIDC_STS_DURATION")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = "3600"
		}
		return f
	}).String()
}

var iframe_override = func() string {
	return Config.Get("features.protection.iframe").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "iframe"
		f.Type = "text"
		f.Description = "list of domains who can use the application from an iframe. eg: https://www.filestash.app http://example.com"
		f.Default = os.Getenv("FS_IFRAME_OVERRIDE")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = ""
		}
		return f
	}).String()
}

var credential_mode = func() string {
	return Config.Get("s3_oidc.credential_mode").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "credential_mode"
		f.Type = "text"
		f.Description = "How S3 credentials are obtained: 'web_identity' (default: token exchange + STS AssumeRoleWithWebIdentity) or 'assume_role' (verify the token locally, then STS AssumeRole with the server's own AWS credentials)"
		f.Default = os.Getenv("FS_S3_OIDC_CREDENTIAL_MODE")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = "web_identity"
		}
		return f
	}).String()
}

var role_arn_prefix = func() string {
	return Config.Get("s3_oidc.role_arn_prefix").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "role_arn_prefix"
		f.Type = "text"
		f.Description = "assume_role mode only (required there): role ARNs must start with this, e.g. arn:aws:iam::111111111111:role/myorg_"
		f.Default = os.Getenv("FS_S3_OIDC_ROLE_ARN_PREFIX")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = ""
		}
		return f
	}).String()
}

var issuer = func() string {
	return Config.Get("s3_oidc.issuer").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Name = "issuer"
		f.Type = "text"
		f.Description = "assume_role mode only (optional): token issuer; derived from the token endpoint when empty"
		f.Default = os.Getenv("FS_S3_OIDC_ISSUER")
		if f.Default != "" {
			f.Placeholder = fmt.Sprintf("Default: '%s'", f.Default)
		} else {
			f.Placeholder = ""
		}
		return f
	}).String()
}

func loadS3OIDCConfig() {
	client_id()
	client_secret()
	client_audience()
	token_endpoint()
	role_arn()
	support_group()
	bucket()
	region()
	sts_duration()
	iframe_override()
	credential_mode()
	role_arn_prefix()
	issuer()
}
