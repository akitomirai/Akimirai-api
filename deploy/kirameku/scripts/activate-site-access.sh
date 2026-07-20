#!/usr/bin/env bash
set -Eeuo pipefail

vhost_source="/tmp/akimirai.xyz.conf"
rate_source="/tmp/kirameku-access-limit.conf"
vhost_target="/etc/nginx/sites-available/akimirai.xyz"
rate_target="/etc/nginx/conf.d/kirameku-access-limit.conf"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
vhost_backup="/etc/nginx/sites-available/akimirai.xyz.before-site-access.${timestamp}"
rate_backup="/etc/nginx/conf.d/kirameku-access-limit.conf.before-site-access.${timestamp}"
installed=0
had_rate=0

for required in "$vhost_source" "$rate_source"; do
  [[ -f "$required" ]] || {
    echo "required deployment file is missing: $required" >&2
    exit 1
  }
done

table_exists="$(docker exec sub2api-postgres psql -U sub2api -d kirameku -Atc \
  "SELECT to_regclass('public.site_access_setting') IS NOT NULL")"
[[ "$table_exists" == t ]] || {
  echo "site access schema is not installed" >&2
  exit 1
}

password_configured="$(docker exec sub2api-postgres psql -U sub2api -d kirameku -Atc \
  "SELECT EXISTS (SELECT 1 FROM site_access_setting WHERE id = 1 AND length(password_hash) >= 20)")"
[[ "$password_configured" == t ]] || {
  echo "site access password is not configured" >&2
  exit 1
}

cp -a "$vhost_target" "$vhost_backup"
if [[ -f "$rate_target" ]]; then
  cp -a "$rate_target" "$rate_backup"
  had_rate=1
fi

restore_on_error() {
  local rc=$?
  trap - ERR
  if [[ $installed -eq 1 ]]; then
    install -o root -g root -m 0644 "$vhost_backup" "$vhost_target"
    if [[ $had_rate -eq 1 ]]; then
      install -o root -g root -m 0644 "$rate_backup" "$rate_target"
    else
      rm -f -- "$rate_target"
    fi
    nginx -t && systemctl reload nginx || true
  fi
  exit "$rc"
}
trap restore_on_error ERR

installed=1
install -o root -g root -m 0644 "$vhost_source" "$vhost_target"
install -o root -g root -m 0644 "$rate_source" "$rate_target"
ln -sfn "$vhost_target" /etc/nginx/sites-enabled/akimirai.xyz

nginx -t
systemctl reload nginx

trap - ERR
echo "site_access=active"
echo "nginx_backup=$vhost_backup"
