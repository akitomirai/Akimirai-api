# Kirameku production deployment

Kirameku is served at `https://akimirai.xyz` without adding a Docker workload.
Nginx terminates TLS, the Next.js standalone server listens on
`127.0.0.1:3001`, and FastAPI listens on `127.0.0.1:8000`. The backend uses a
dedicated database/user in the existing PostgreSQL service and stores uploaded
images under `/var/lib/kirameku/uploads`.

## Build order

Build on a workstation, never on the production ECS host:

1. Install frozen dependencies for `Kirameku-backend/admin` and build its
   `dist` directory.
2. Start a local backend so Next.js static generation can read empty/seed data.
3. Build `Kirameku` with `output: "standalone"`.
4. Copy `public` and `.next/static` into `.next/standalone`.
5. Package the backend source/admin dist and the standalone frontend separately.

Upload both archives plus the files under `systemd/` and `nginx/` to `/tmp`,
then run `scripts/install-release.sh RELEASE FRONTEND_SHA256 BACKEND_SHA256` as
root. The installer verifies both archives, restores portable pnpm links,
initializes only an empty database, applies idempotent SQL migrations, switches
the active release atomically, checks both loopback services, validates the
currently active Nginx configuration, and rolls back the symlink if a
post-switch step fails. It deliberately does not activate the password gate.

## Site access activation

Activate the password gate only after the new application release is healthy:

1. Load `/etc/kirameku/backend.env` into the root shell and change to
   `/opt/kirameku/current/backend`.
2. Pass the initial password through standard input to
   `scripts/bootstrap_site_access.py --password-stdin`. Do not place the
   password in shell history, process arguments, Git, or deployment logs.
3. Confirm the bootstrap reports a positive authentication version.
4. Upload `nginx/akimirai.xyz.conf` and
   `nginx/kirameku-access-limit.conf` to `/tmp`, install
   `scripts/activate-site-access.sh` with mode `0755`, then run it as root.

The activator refuses to change Nginx until a password hash exists, keeps a
timestamped `akimirai.xyz.before-site-access` backup, validates the complete
configuration, and restores the previous files if validation or reload fails.
The login endpoint is limited per source IP; `/access`, its static assets,
`/api/health`, login, logout, and ACME remain public. All other pages, admin
routes, APIs, and uploads require the secure site-access cookie.

Run `scripts/verify-site-access.py` from an interactive terminal after
activation. It prompts without echo and checks the anonymous/authenticated
matrix, production Cookie attributes, protected uploads, and logout without
printing the password or Cookie value.
Add `--check-rotation` to securely prompt for the admin username/password and
verify version invalidation by re-saving the same access password. The check
never changes the effective password and never prints the admin token.

## Production layout

- `/opt/kirameku/releases/<release>/backend`
- `/opt/kirameku/releases/<release>/frontend`
- `/opt/kirameku/current` -> active release
- `/opt/kirameku/venv` -> shared Python virtual environment
- `/etc/kirameku/backend.env` -> root-owned secrets, mode `0640`
- `/var/lib/kirameku/uploads` -> persistent uploaded images

The release tree stays read-only except for Next.js derived caches under
`.next/cache`, `.next/server/app`, and `.next/server/pages`. Those three
directories are owned by the service account because Next.js refreshes
prerendered output at runtime; they contain no source-of-truth data.
Both services run without Linux capabilities and with private devices,
restricted address families/namespaces, protected kernel interfaces, hidden
process metadata, and explicit memory/task ceilings.

The backend environment must explicitly set `DATABASE_URL`, `SECRET_KEY`,
`CORS_ORIGINS=https://akimirai.xyz`, `FRONTEND_ORIGIN=https://akimirai.xyz`,
`STORAGE_BACKEND=local`, `UPLOAD_DIR=/var/lib/kirameku/uploads`, and
`UPLOAD_BASE_URL=/uploads`. GitHub OAuth remains disabled until provider
credentials are supplied.

## Health gates

- `GET /` returns the Kirameku HTML shell.
- `GET /posts` returns a deep Next.js route.
- `GET /api/health` returns `{"status":"ok"}`.
- `GET /admin/` returns the Vue mount point.
- Administrator login succeeds without logging the password or token.
- An authenticated raster upload returns a same-origin `/uploads/...` URL and
  the returned URL is fetchable.
- A fake SVG or invalid image upload is rejected.
- Both services remain healthy after `systemctl restart`.
- MiraiAPI gates at `miraiapi.cloud` remain unchanged.
- Without a site cookie, `/` redirects to `/access`, while `/api/posts`
  returns `401` and never returns an HTML login page.
- A correct password sets a host-only, Secure, HttpOnly, SameSite=Lax cookie;
  a wrong password returns `401`, and repeated attempts are rate limited.
- Updating the password in admin invalidates older site-access cookies, and
  “锁定站点” clears the current browser cookie.

## Rollback

Keep the previous release directory and the timestamped pre-gate Nginx backup.
For a gate-related rollback, restore the previous Nginx files and reload Nginx
first so the site cannot remain locked behind an incompatible application.
Then atomically repoint
`/opt/kirameku/current`, run `systemctl restart kirameku-backend
kirameku-frontend`, validate the local health endpoints, then reload Nginx only
if its configuration changed. Database and upload directories are persistent
and are never deleted as part of a release rollback.
