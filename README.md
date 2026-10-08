![screenshot](https://raw.githubusercontent.com/mickael-kerjean/filestash_images/master/.assets/photo.jpg)

<p align="center">
    <a href="https://github.com/mickael-kerjean/contributors" alt="Contributors">
        <img src="https://img.shields.io/github/contributors/mickael-kerjean/filestash" style="max-width:100%;">
    </a>
    <a href="https://hub.docker.com/r/machines/filestash" alt="Docker Hub">
        <img src="https://img.shields.io/docker/pulls/machines/filestash" style="max-width:100%;">
    </a>
    <a href="https://kiwiirc.com/nextclient/#irc://irc.libera.chat/#filestash?nick=guest??" alt="Chat on IRC">
        <img src="https://img.shields.io/badge/IRC-%23filestash-brightgreen.svg" style="max-width:100%;">
    </a>
</p>

<p align="center">
    A Dropbox-like file manager that let you manage your data anywhere it is located:<br>
    <a href="https://www.filestash.app/ftp-client.html">FTP</a> • FTPS • <a href="https://www.filestash.app/ssh-file-transfer.html">SFTP</a> • <a href="https://www.filestash.app/webdav-client.html">WebDAV</a> • Git • <a href="https://www.filestash.app/s3-browser.html">S3</a> • NFS • <a href="https://www.filestash.app/smb-client.html">SMB</a> • Artifactory • <a href="https://www.filestash.app/ldap-browser.html">LDAP</a> • Mysql <br>
       Storj • CardDAV • CalDAV • Backblaze B2 • <a href="https://www.filestash.app/s3-browser.html">Minio</a> <br>
               Dropbox • Google Drive
</p>
<p align="center">
    <a href="http://demo.filestash.app">
      <img src="https://raw.githubusercontent.com/mickael-kerjean/filestash_images/master/.assets/button_demo.png" alt="demo button" />
    </a>
</p>

# Features
- Manage your files from a browser
- Authentication middleware to connect to various source of user
- Flexible Share mechanism
- Chromecast support for images, music, and videos
- Video player
- Video transcoding (mov, mkv, avi, mpeg, and more)
- Image viewer
- Image transcoding (raw images from Nikon, Canon, and more)
- Photo management
- Audio player
- Shared links are full fledge network drive
- Office documents (docx, xlsx and more)
- Full org mode client ([documentation](https://www.filestash.app/2018/05/31/release-note-v0.1/))
- User friendly
- Mobile friendly
- Customisable
- Plugins
- Super fast
- Upload files and folders
- Download as zip
- Multiple cloud providers and protocols, easily extensible
- Nyan cat loader
- Emacs, VIM or Sublime keybindings `;)`
- Pluggable Search (default is recursive search)
- .. and many more

# Developer Documentation

## Local Build

```console
❯ npm --version
10.2.3
❯ go version
go version go1.23.2 linux/amd64
❯ webpack --version

  System:
    OS: Linux 6.2 Ubuntu 23.04 23.04 (Lunar Lobster)
    CPU: (4) x64 Intel(R) Core(TM) i7-7600U CPU @ 2.80GHz
    Memory: 9.54 GB / 15.36 GB
  Binaries:
    Node: 18.19.0 - ~/.nvm/versions/node/v18.19.0/bin/node
    npm: 10.2.3 - ~/.nvm/versions/node/v18.19.0/bin/npm
  Browsers:
    Chrome: 127.0.6533.72
  Packages:
    babel-loader: ^7.1.5 => 7.1.5
    compression-webpack-plugin: ^1.1.11 => 1.1.12
    copy-webpack-plugin: ^4.5.2 => 4.6.0
    css-loader: ^0.28.10 => 0.28.11
    html-loader: ^0.4.5 => 0.4.5
    html-webpack-plugin: ^5.5.3 => 5.6.0
    little-loader: ^0.2.0 => 0.2.0
    sass-loader: ^7.2.0 => 7.3.1
    style-loader: ^0.20.2 => 0.20.3
    uglifyjs-webpack-plugin: ^1.2.5 => 1.3.0
    url-loader: ^0.6.2 => 0.6.2
    webpack: ^5.87.0 => 5.95.0
    webpack-bundle-analyzer: ^3.3.2 => 3.9.0
    webpack-cli: ^5.1.4 => 5.1.4
    worker-loader: ^2.0.0 => 2.0.0
```

```bash
sudo apt-get install ffmpeg libwebp-dev libheif-dev libjpeg-dev libpng-dev
npm install --legacy-peer-deps
make build_frontend
make build_init
go build tidy
make build_backend
make -B debug

# Optional: set s3_oidc defaults, can override at
# localhost:8334/admin/settings
export FS_S3_OIDC_CLIENT_ID='works'
export FS_S3_OIDC_CLIENT_SECRET='xxx'
export FS_S3_OIDC_CLIENT_AUD='workss3'
export FS_S3_OIDC_TOKEN_ENDPOINT='https://keycloak.example.org/realms/datajoint/protocol/openid-connect/token'
export FS_S3_OIDC_DEFAULT_ROLE_ARN='arn:aws:iam::111111111111:role/acme-lab_ephys_researcher'
export FS_S3_OIDC_DEFAULT_BUCKET='example-bucket'
export FS_S3_OIDC_DEFAULT_REGION='us-east-2'
export FS_S3_OIDC_SUPPORT_GROUP='datajoint'
export FS_S3_OIDC_STS_DURATION=3600
export FS_IFRAME_OVERRIDE='https://portal.example.org'
# Injected as /custom.css on both frontends; keeps UI tweaks in deployment config
export FS_CUSTOM_CSS=''
# Files-tab UX: a visible upload control, feedback after upload, list over grid
export FS_UPLOAD_BUTTON='true'
export FS_REFRESH_AFTER_UPLOAD='true'
export FS_FILEPAGE_DEFAULT_VIEW='list'
# Chunked uploads (MB; 0 = off), sort order, zip timeout (s), shared links.
# Chunking is read only by the canary frontend, so it has no effect while CANARY
# is unset, and resume-after-blip is unverified on canary too -- left at 0.
export FS_UPLOAD_CHUNK_SIZE='0'
export FS_FILEPAGE_DEFAULT_SORT='date'
export FS_ZIP_TIMEOUT='600'
export FS_SHARE_ENABLE='false'
export APPLICATION_URL=''
```


`FS_CUSTOM_CSS` is served as `/custom.css` on both frontends, so per-deployment UI
tweaks live in deployment config rather than in the Admin Console (whose state
directory a kubernetes deployment does not persist). A worked example:

```css
/* Quick Access is not wanted here; hiding the component also removes its
   "Frequently access folders will be shown here" placeholder and the 145px
   it reserves at the top of every root listing. */
.component_frequently-access { display: none; }

/* A permanent drop zone. filestash's own (.component_filezone) is rendered
   behind `NgIf cond={fileIsOver}`, so it does not exist in the DOM until a
   drag is already under way and cannot be made static from CSS. What makes a
   static one work is that connectDropFile() wraps .component_filesystem, so
   the whole file area is the drop target and a box drawn inside it is
   genuinely droppable. Anchored on .component_container to inherit the file
   list's width rather than bleeding to the viewport edges. */
.component_filesystem.is_folder > .component_container::before {
    content: "Drag and drop your folder here to upload it with everything inside — or use CREATE NEW FOLDER above, open it, and drop your files in.";
    display: block; margin: 14px auto 10px; max-width: 620px; padding: 25px 20px;
    border: 2px dashed; border-radius: 6px; text-align: center; font-weight: bold;
    font-size: .88em; line-height: 1.6; opacity: .6;
}
/* Mid-drag filestash renders .component_filezone in the same place; drop ours
   then so the two never stack. */
.component_filesystem.is_folder > .component_container:has(.component_filezone)::before {
    display: none;
}

/* Relabel the create controls. The stock rule
   `.component_submenu .component_container .menubar>span` sets font-size 12px
   and letter-spacing .06em, so it must be matched on specificity or the
   original text node survives and the replacement merely appends to it. The
   label is a bare text node with no element children, so zeroing the font
   collapses it and ::after carries the replacement. */
.component_submenu .component_container .menubar > span.button-new-file,
.component_submenu .component_container .menubar > span.button-new-folder {
    font-size: 0; letter-spacing: 0;
}
.component_submenu .component_container .menubar > span.button-new-file::after {
    content: "Create new file"; font-size: 12px; letter-spacing: .06em;
}
.component_submenu .component_container .menubar > span.button-new-folder::after {
    content: "Create new folder"; font-size: 12px; letter-spacing: .06em;
}
```

These selectors target the legacy frontend, the one served while `CANARY` is
unset. They are written against the compiled stylesheet's own specificity, so
they are more brittle than a code change — keep them few and deliberate.

Note that an `FS_*` variable sets a setting's **default**, not its value. A value
saved through the Admin Console is stored separately and still wins, so where the
state directory *is* persisted, a console edit will override the environment. Where
it is not persisted — the kubernetes deployment — a console edit appears to work and
is silently lost on restart.

```bash
./dist/filestash-debug
# 2024/10/22 21:02:51 SYST INFO [http] starting ...
# 2024/10/22 21:02:52 SYST INFO [http] listening on :8334
# 2024/10/22 21:02:55 HTTP 307 GET    0.1ms /login
# 2024/10/22 21:02:55 HTTP 200 GET    0.1ms /custom.css
# 2024/10/22 21:02:55 HTTP 401 GET    0.2ms /api/session
# 2024/10/22 21:02:55 HTTP 200 GET    1.6ms /api/config
# 2024/10/22 21:02:56 HTTP 304 GET    7.8ms /api/backend
# 2024/10/22 21:02:56 HTTP 200 GET    0.4ms /manifest.json
# 2024/10/22 21:02:57 HTTP 304 GET    0.1ms /sw_cache.js
```

Navigating to `http://localhost:8334/login#type=s3_oidc&token=xxx` will authorize the user and redirect to the S3 bucket. Can also pass `role_arn` to assume a specific role, `region` to specify a region e.g. if role doesn't have `GetBucketLocation`, `path` like `/example-bucket/acme-lab_ephys/` to set the root directory. Without the overrides in the query params, the application will fall back to the configuration values defined in the Admin Console and/or environment variables.

Alternatively, pass `project` only in the URL: `http://localhost:8334/login#type=s3_oidc&project=acme-lab_ephys&token=xxx`, which will use the `role_arn` and other defaults defined in the Admin Console and/or environment variables to get the `path`, default `region`, and `role_arn` for the user.

## How to Enable Internal Token Exchange in KeyCloak

1. Turn on "preview" features (possibly only `token-exchange` is enough) and `kubectl apply` the updated KeyCloak deployment.
2. Go to target client and turn on Permissions enabled switch in top level permissions tab.
3. Click on the `token-exchange` permission and note the client that it redirects to, in our case it was `realm-management`.
4. Go to `realm-management` Authorization -> Policies, then create a policy of Client type.
    1. Name is `target-client-exchange`, and set the client to the source client (`works` for us)
    2. Logic `positive`, and save the policy.
5. Go back to target client, `workss3`, Permissions top level tab, click `token-exchange` permission again, and then add the `target-client-exchange` policy to permission and save
    1. Permission should be unanimous
6. Test:

```bash
curl -L -X POST \
    -d "client_id=<source client ID, e.g. works>" \
    -d "client_secret=<source client secret>" \
    --data-urlencode "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
    -d "subject_token=<source client access token>" \
    --data-urlencode "requested_token_type=urn:ietf:params:oauth:token-type:refresh_token" \
    -d "audience=<target client>" \
    http://keycloak.example.org/realms/datajoint/protocol/openid-connect/token
```

This request should return an access and refresh token:

```json
{
  "access_token": "<workss3 access token>",
  "expires_in": 35415,
  "refresh_expires_in": 35415,
  "refresh_token": "<workss3 refresh token>",
  "token_type": "Bearer",
  "not-before-policy": 0,
  "session_state": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "scope": "email profile groups"
}
```

Access token decodes to:

```json
{
  "exp": 1729917199,
  "iat": 1729881784,
  "auth_time": 1729873999,
  "jti": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "iss": "https://keycloak.example.org/realms/datajoint",
  "aud": [
    "datajoint",
    "acme-lab_ephys",
    "account",
    "workss3"
  ],
  "sub": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "typ": "Bearer",
  "azp": "works",
  "session_state": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "acr": "1",
  "realm_access": {
    "roles": [
      "datajoint",
      "acme-lab_ephys",
      "offline_access",
      "default-roles-datajoint",
      "uma_authorization"
    ]
  },
  "resource_access": {
    "account": {
      "roles": [
        "manage-account",
        "manage-account-links",
        "view-profile"
      ]
    }
  },
  "scope": "email profile groups",
  "sid": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "email_verified": true,
  "name": "Test User",
  "groups": [
    "datajoint",
    "acme-lab_ephys",
    "offline_access",
    "default-roles-datajoint",
    "uma_authorization"
  ],
  "preferred_username": "test-user",
  "given_name": "Test",
  "family_name": "User",
  "email": "test-user@example.org"
}
```

For assume role with token in AWS, `workss3` has a User Role Mapping in dedicated scopes, that makes the access token's `aud` the same as the `groups`. Then AWS converts `aud` to `oaud` [under the hood](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_iam-condition-keys.html#condition-keys-wif). `aud` won't be the same as grouops, unless you attach a role to the group via the User Role Mapping.

# Documentation
- [Getting started](https://www.filestash.app/docs/)
- [Installation](https://www.filestash.app/docs/install-and-upgrade/)
- [FAQ](https://www.filestash.app/docs/faq/)

# Screenshots
<p align="center">
    <a href="https://demo.filestash.app">
        <img src="https://raw.githubusercontent.com/mickael-kerjean/filestash_images/master/.assets/navigation.gif" alt="user experience on navigation" />
    </a>
</p>
<p align="center">
    <a href="http://demo.filestash.app">
        <img src="https://raw.githubusercontent.com/mickael-kerjean/filestash_images/master/.assets/photo_management.gif" alt="user experience on medias" />
    </a>
</p>

# Core Principles and Ideas

Filestash started as an attempt to solve the "Dropbox problem" by abstracting the storage layer, allowing you to "bring your own backend" by implementing this interface:
```go
type IBackend interface {
	Init(params map[string]string, app *App) (IBackend, error) // constructor
	Ls(path string) ([]os.FileInfo, error)           // list files in a folder
	Cat(path string) (io.ReadCloser, error)          // download a file
	Mkdir(path string) error                         // create a folder
	Rm(path string) error                            // remove something
	Mv(from string, to string) error                 // rename something
	Save(path string, file io.Reader) error          // save a file
	Touch(path string) error                         // create a file
	LoginForm() Form                                 // dynamic form generation for the login
}
```

The software is effectively 2 distinct parts: Core and [Plugins](https://github.com/mickael-kerjean/filestash/tree/master/server/plugin). Plugins act as "Lego blocks" that shape all the opinions on how the file manager operates, such as who can do what and where (aka, the authorisation), how the search features work (e.g., recursive search, full text search, ...), how users are authenticated (e.g., passthrough to the storage layer, LDAP, OIDC, SAML, signed URLs), how the application behaves when opening different file types, etc...

The architecture is designed to enable implementers to build file managers that are fit for purpose. To cite an example, several control panels for elevators in Europe feature QR codes generated by Filestash. Technicians responsible for their maintenance scan these QR codes to access the complete history of the elevator, upload new records, access relevant external resources, and view a banner displaying location specific metadata from the ERP. What would have been a complex custom solution was made possible through the addition of two plugins: one for QR code generation and another for ERP integration. Everything else was off the shelf.

# Documentation
- [Getting started](https://www.filestash.app/docs/)
- [Installation](https://www.filestash.app/docs/install-and-upgrade/)
- [API](https://www.filestash.app/docs/api/)

# Support
- For companies -> [support contract](https://www.filestash.app/pricing/)
- For individuals -> [#filestash](https://kiwiirc.com/nextclient/#irc://irc.libera.chat/#filestash?nick=guest??) on IRC (libera.chat). To financially contribute to the project:
  - Bitcoin: `3LX5KGmSmHDj5EuXrmUvcg77EJxCxmdsgW`
  - [Open Collective](https://opencollective.com/filestash)

# Credits
- [Contributors](https://github.com/mickael-kerjean/filestash/graphs/contributors) and folks developing [awesome libraries](https://github.com/mickael-kerjean/filestash/blob/master/go.mod)
- This project is tested with BrowserStack
- Logo derived from the work of [ssnjrthegr8](https://github.com/ssnjrthegr8), Iconography from [flaticon](https://www.flaticon.com/), [fontawesome](https://fontawesome.com) and [material](https://material.io/icons/)
- [libvips](https://github.com/libvips/libvips) and [libraw](https://github.com/LibRaw/LibRaw). Those libraries are statically compiled in Filestash. Instructions to build Filestash is available [here](https://github.com/mickael-kerjean/filestash/blob/master/.drone.yml) and instructions to create your own static library for libvips and libraw is to be found [here](https://github.com/mickael-kerjean/filestash/tree/master/server/plugin/plg_image_light/deps)
