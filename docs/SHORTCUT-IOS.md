# FreshVPS Admin Shortcut (iOS / iPadOS / macOS)

Goal: one operator Shortcut that talks to the **VPN-only API** after **Face ID** and a **local session token file**. End users never use this Shortcut.

## Security model (agreed)

1. API only reachable when you are on **your VPN** (or SSH tunnel to `127.0.0.1:8787`).
2. **Session token** from `freshvps-vpn session` or Telegram `/session` — not a permanent master in the Shortcut body.
3. **Face ID / passcode** action at the start of the Shortcut.
4. Token stored in a **local file** on the device (On My iPhone / On My Mac preferred over iCloud).

## One-time: mint a session on the VPS

```bash
sudo freshvps-vpn session 72
# copy the token line only
```

Or Telegram (operator chat): `/session 72`

On the iPhone: Files → **On My iPhone** → folder `FreshVPS` → file `session.txt` containing only the token (no spaces/newlines issues).

## Build the Shortcut on iPhone (detailed)

### A. Create

1. Open **Shortcuts** → **+** → name it `FreshVPS Admin`.
2. Add action **Authenticate** (or “Lock Screen” / Face ID / passcode — exact label depends on iOS language).
   - If authentication fails, Shortcut must stop.

### B. Read token

3. **Get File** → select `On My iPhone/FreshVPS/session.txt` (or “Ask Each Time” if you prefer typing).
4. **Get Text from Input** (file body = token).
5. **Set Variable** → `Token`.

### C. Choose action

6. **Choose from Menu**:
   - List users
   - Add user
   - Disable user
   - Link + QR for user

### D. API base URL

While on VPN, base is typically:

- `http://127.0.0.1:8787` if you use **SSH local forward**: `ssh -L 8787:127.0.0.1:8787 user@VPS`
- or `http://<VPN-internal-IP>:8787` if you later bind API to the VPN interface

Default install binds **`127.0.0.1:8787`**. From iPhone the practical path is:

1. Connect **VLESS to this VPS**, **and**
2. Either change `VPN_API_BIND` to the VPN tunnel IP later, **or** use a small SSH tunnel app, **or** set bind to `0.0.0.0` only on a **VPN interface** (advanced).

Recommended v1 operator flow from phone without SSH app:

- Prefer **Telegram bot** for add/list/QR when away from Mac.
- Prefer **Shortcuts + SSH tunnel** or Mac Shortcuts when at desk.

Alternatively set in config before install:

```bash
VPN_API_BIND=0.0.0.0   # then firewall allow only from VPN subnet — careful
```

Document your chosen bind in `/etc/freshvps/READY.txt`.

### E. HTTP examples in Shortcut

**List users**

- Get Contents of URL
  - URL: `http://BASE/vpn/users`
  - Method: GET
  - Headers: `Authorization` = `Bearer ` + Variable Token

**Add user**

- URL: `http://BASE/vpn/users`
- Method: POST
- Headers: Authorization as above, Content-Type `application/json`
- Body: `{"name":"alice","note":"phone"}`
- Show Result (JSON includes `link`)
- Optional: Get Contents of URL `.../vpn/users/alice/qr` → Show/Quick Look image

**Disable**

- POST `http://BASE/vpn/users/alice/disable`

### F. QR display

Server generates PNG. Shortcut only **downloads and shows** it — no local QR encoding required.

## Export and host on VPS (so TG can redistribute)

1. On the device where the Shortcut works: **Share** → **Export** / **Anyone** (or “People who know me”) → save `FreshVPS-Admin.shortcut`.
2. Copy to VPS:

```bash
sudo mkdir -p /opt/freshvps/shortcuts
sudo install -m 644 FreshVPS-Admin.shortcut /opt/freshvps/shortcuts/
```

3. Optional IMPORT helper (when API reachable):

```bash
# After you know how phone reaches API:
echo 'shortcuts://import-shortcut?url=URL_ENCODED_HTTPS_OR_HTTP_TO_FILE&name=FreshVPS%20Admin' \
  | sudo tee /opt/freshvps/shortcuts/IMPORT.txt
```

Apple import scheme:

```text
shortcuts://import-shortcut?url=<url-to-.shortcut>&name=FreshVPS%20Admin
```

4. Telegram `/shortcut` sends `IMPORT.txt` if present.

**Note:** VPS does not generate a full `.shortcut` binary from scratch in v1; it **stores and serves** your exported template (no secrets inside).

## macOS / iPadOS

Same Shortcut syncs via iCloud if you use the same Apple ID — still keep **session file local** or re-mint per device. Prefer On My Mac file, not Desktop iCloud.

## Checklist

- [ ] Face ID at start
- [ ] Token only from local file / Ask
- [ ] No master token pasted into Shortcut actions as plain text
- [ ] API only while on operator VPN / tunnel
- [ ] Exported template on VPS without secrets
- [ ] `/session` when token expires
