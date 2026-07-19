# MiraiAPI Domain Migration Runbook

This directory owns the repeatable assets for moving MiraiAPI from
`akimirai.xyz` to `miraiapi.cloud` while preserving explicit legacy API paths for
30 days.

## Safety Boundary

- Do not touch unrelated `*.akimirai.xyz` vhosts.
- Do not build the application on the production ECS host.
- Keep the active old vhost until the new DNS-only origin passes all health
  gates.
- Do not orange-cloud the apex until the origin certificate and Full (strict)
  validation succeed.

## Render

Use the actual cutover timestamp, including its UTC offset:

```bash
python deploy/miraiapi-domain-migration/scripts/render_templates.py \
  --cutover 2026-07-19T14:30:00+08:00 \
  --source-dir deploy/miraiapi-domain-migration \
  --output-dir .runtime/miraiapi-domain-migration
```

The rendered manifest records the exact 30-day sunset time. Do not edit the
rendered Sunset header or timer independently.

Generate the Cloudflare real-IP include from official lists:

```bash
python deploy/miraiapi-domain-migration/scripts/fetch_cloudflare_ips.py \
  --output .runtime/miraiapi-domain-migration/nginx/cloudflare-real-ip.conf
```

## Production Backup Gate

Before changing the host, capture timestamped copies of:

- `/etc/nginx/sites-enabled/akimirai.xyz`
- `/etc/nginx/sites-available/akimirai.xyz`
- `/opt/akimirai/data/config.yaml`
- the PostgreSQL rows `api_base_url` and
  `balance_low_notify_recharge_url`
- `systemctl status akimirai.service nginx`
- existing migration-specific systemd units/timers

Store the backup path and hashes in the task evidence before continuing.

## Origin Preparation

1. Install `miraiapi.cloud-bootstrap.conf` as the only new-domain vhost.
2. Create `/var/www/letsencrypt` and verify an ACME challenge file over HTTP.
3. Obtain a public certificate:

   ```bash
   certbot certonly --webroot -w /var/www/letsencrypt \
     -d miraiapi.cloud -d www.miraiapi.cloud
   ```

4. Install `miraiapi.cloud.conf`, `miraiapi-proxy.conf` and
   `cloudflare-real-ip.conf`.
5. Run `nginx -t`, reload Nginx, then verify:

   ```bash
   curl -fsS https://miraiapi.cloud/health
   curl -fsS https://miraiapi.cloud/admin/usage | grep -F '<div id="app"></div>'
   curl -fsS https://miraiapi.cloud/images/ | grep -F '<div id="root"></div>'
   ```

## Application Cutover

- Set `server.frontend_url` to `https://miraiapi.cloud`.
- Keep both old and new HTTPS origins in CORS during the compatibility period.
- Update PostgreSQL `api_base_url` to `https://miraiapi.cloud/` and
  `balance_low_notify_recharge_url` to `https://miraiapi.cloud` in one
  transaction.
- Restart `akimirai.service` once, then validate public settings.

## Emergency Primary Re-cutover

If the current primary hostname becomes blocked, treat the replacement as a
new promotion without extending the existing legacy compatibility period.

1. Publish the replacement apex and `www` records, then verify the authoritative
   nameservers plus at least two independent recursive resolvers.
2. Install the HTTP bootstrap, prove the ACME challenge path, issue the
   replacement certificate and pass `/health`, `/admin/usage` and `/images/`
   before changing application settings.
3. Update YAML `frontend_url`/CORS and PostgreSQL `api_base_url`/
   `balance_low_notify_recharge_url`, then restart the application once.
4. Change the existing legacy compatibility `Link`, new-origin header and UI
   redirect directly to the replacement hostname. Keep the original `Sunset`
   and retirement timer unchanged.
5. Disable the superseded primary vhost after the replacement passes strict TLS,
   public settings, CORS and authenticated API gates. Keep its certificate,
   disabled config and timestamped backup only for the rollback window.

The 2026-07-19 emergency re-cutover promoted `miraiapi.cloud` using this
sequence.

## Legacy Compatibility

Install the rendered compatibility vhost as:

```text
/etc/nginx/sites-enabled/akimirai.xyz-miraiapi-compat
```

Disable the previous catch-all old-domain vhost only after the new-domain
health gates pass. The compatibility vhost proxies API/setup/health and
rent-ledger paths; all other routes use a path-preserving 308 redirect to
`miraiapi.cloud`.

Install the retirement script and systemd units, then verify the scheduled
time with:

```bash
systemctl list-timers miraiapi-legacy-retire.timer --all
```

The retirement script removes only the compatibility vhost. If `nginx -t` or
reload fails, it restores that vhost automatically.

## Cloudflare Handoff

1. Set SSL/TLS mode to **Full (strict)**.
2. Ensure WebSockets are enabled.
3. Ensure API paths are not covered by JS/Managed Challenge or Cache Everything
   rules.
4. Orange-cloud `www` first and validate it as the canary.
5. Orange-cloud the apex and verify Cloudflare Anycast DNS, `CF-Ray`, SSE,
   WebSocket, 42 MB request bodies and real client IP.
6. Enable DNSSEC after steady-state validation and publish the DS record at the
   registrar.

Cloudflare Free allows 100 MB uploads and has a 120-second origin read timeout.
Streaming handlers must send the first byte before that limit and maintain SSE
heartbeats. WebSocket clients must reconnect after edge connection resets.

## Rollback

1. Switch Cloudflare records to DNS-only if the edge is implicated.
2. Restore the timestamped old Nginx enabled file and disable the new vhost.
3. Restore YAML and the two PostgreSQL settings.
4. Disable the migration timer and remove only migration-specific units.
5. Run `nginx -t`, restart `akimirai.service`, and re-run `/health`,
   `/admin/usage` and `/images/` on the old domain.
