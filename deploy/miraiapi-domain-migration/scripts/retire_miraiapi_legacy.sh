#!/usr/bin/env bash
set -euo pipefail

enabled_path="${MIRAIAPI_LEGACY_ENABLED_PATH:-/etc/nginx/sites-enabled/akimirai.xyz-miraiapi-compat}"
state_dir="${MIRAIAPI_MIGRATION_STATE_DIR:-/var/lib/miraiapi-domain-migration}"
nginx_bin="${NGINX_BIN:-nginx}"
systemctl_bin="${SYSTEMCTL_BIN:-systemctl}"

mkdir -p "$state_dir"
exec 9>"$state_dir/retire.lock"
flock -n 9 || {
    echo "legacy retirement is already running" >&2
    exit 0
}

if [[ ! -e "$enabled_path" && ! -L "$enabled_path" ]]; then
    echo "legacy compatibility vhost is already absent"
    exit 0
fi

timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_path="$state_dir/$(basename "$enabled_path").pre-retire.$timestamp"
mv -T -- "$enabled_path" "$backup_path"

restore_legacy() {
    if [[ ! -e "$enabled_path" && ! -L "$enabled_path" ]]; then
        mv -T -- "$backup_path" "$enabled_path"
    fi
}

if ! "$nginx_bin" -t; then
    restore_legacy
    if "$nginx_bin" -t; then
        echo "nginx validation failed after removing the legacy vhost; compatibility vhost restored" >&2
    else
        echo "legacy compatibility vhost restored, but nginx configuration still fails validation" >&2
    fi
    exit 1
fi

if ! "$systemctl_bin" reload nginx; then
    restore_legacy
    if ! "$nginx_bin" -t; then
        echo "legacy compatibility vhost restored after reload failure, but nginx validation now fails" >&2
    fi
    "$systemctl_bin" reload nginx || true
    echo "nginx reload failed; legacy compatibility vhost restored" >&2
    exit 1
fi

date -u +%Y-%m-%dT%H:%M:%SZ >"$state_dir/retired-at"
echo "legacy compatibility vhost retired; backup retained at $backup_path"
