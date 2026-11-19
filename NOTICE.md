# Notice

This is a modified version of [Filestash](https://github.com/mickael-kerjean/filestash), licensed under
the GNU Affero General Public License, version 3 (see [LICENSE](LICENSE)). DataJoint modified it from
2024-10-29 onward, and the modified version is distributed under the same license.

The source of the version you are running is linked from its About page (`/about`); the source of
every image DataJoint released from 0.2.7 on is published at https://github.com/datajoint-company/filestash-source.

This tree is the complete source of the version DataJoint released as image 0.2.12 (2026-11-19). Example
settings and documentation name example hosts rather than DataJoint's own.

## Modifications

| Date | What changed |
|---|---|
| 2024-10 to 2024-11 | The `s3_oidc` storage backend (`server/plugin/plg_backend_s3_oidc`): S3 access with short-lived credentials exchanged for the user's OpenID Connect token, with the role checked against the token's groups and roles; its settings from environment variables |
| 2024-11 | The interface trimmed for use inside another application: no home button on error pages and no logout button in the breadcrumb; iframe protection settable from the environment; file permissions set by the user's role |
| 2025-01 to 2025-02 | Bucket parameters passed through; the admin password from the environment; the `s3` and `s3_oidc` backends enabled by default |
| 2025-06 | Continuous integration fixes |
| 2026-09 | An opt-in `assume_role` credential mode for the `s3_oidc` backend |
| 2026-10 | Eight settings of the file browser settable from the environment, with sharing off by default; custom CSS documented |
| 2026-10-21 | This notice; the About page links the source of this version (`FS_SOURCE_URL`); the build can be given its commit (`BUILD_REF`) |
| 2026-11 | Copying a file or folder path to the clipboard; a notice when an upload into `inbox/` starts (`FS_UPLOAD_NOTICE`); a tree of folders and files beside the file list, with the number of files in each opened folder; the `s3_oidc` backend reports an identity provider or STS it cannot reach as a connection error instead of "Permission Denied"; TeX removed from the image, so org-mode export to PDF and Beamer is no longer available |

