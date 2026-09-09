# Releases

## How releases are published

### Preferred (CI)

1. Bump `VERSION` on `main`.
2. Tag: `git tag v0.7.0 && git push origin v0.7.0`
3. Workflow [`.github/workflows/release-netductor.yml`](../.github/workflows/release-netductor.yml) builds Linux amd64/arm64 and creates a GitHub Release with assets via **`GITHUB_TOKEN`** (Actions — not a personal PAT in chat).

### Manual / agent (REST API)

When CI is not available:

1. `POST https://api.github.com/repos/{owner}/{repo}/releases` with JSON `{tag_name, name, body, prerelease}`.
2. Upload each file to `upload_url` from the response:
   `POST https://uploads.github.com/.../assets?name=netductor-linux-amd64`
   headers: `Authorization: Bearer <token>`, `Content-Type: application/octet-stream`.

Token needs **`contents: write`**. Same method was used earlier for `freshvps-tg` v0.5.7.

The in-chat GitHub **connector** can edit files but **cannot** create releases or upload assets — use CI or PAT+REST.

## Install CLI from a release

```bash
TAG=v0.7.0
curl -fsSL -o /usr/local/bin/netductor \
  "https://github.com/PavelNeyman/FreshVPS/releases/download/${TAG}/netductor-linux-amd64"
chmod 755 /usr/local/bin/netductor
netductor version
```

After the repository is renamed to `netductor`, update the URL path.
