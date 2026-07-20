#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
retire_script="$root_dir/scripts/retire_miraiapi_legacy.sh"
test_root="$(mktemp -d)"
trap 'rm -rf -- "$test_root"' EXIT

write_fake_nginx() {
    local path="$1"
    local result="$2"
    cat >"$path" <<EOF
#!/usr/bin/env bash
exit $result
EOF
    chmod +x "$path"
}

write_fake_systemctl() {
    local path="$1"
    local result="$2"
    cat >"$path" <<EOF
#!/usr/bin/env bash
exit $result
EOF
    chmod +x "$path"
}

run_retire() {
    local case_dir="$1"
    MIRAIAPI_LEGACY_ENABLED_PATH="$case_dir/enabled" \
    MIRAIAPI_MIGRATION_STATE_DIR="$case_dir/state" \
    NGINX_BIN="$case_dir/fake-nginx" \
    SYSTEMCTL_BIN="$case_dir/fake-systemctl" \
        "$retire_script"
}

success_dir="$test_root/success"
mkdir -p "$success_dir"
touch "$success_dir/enabled"
write_fake_nginx "$success_dir/fake-nginx" 0
write_fake_systemctl "$success_dir/fake-systemctl" 0
run_retire "$success_dir"
[[ ! -e "$success_dir/enabled" ]]
[[ -f "$success_dir/state/retired-at" ]]

nginx_fail_dir="$test_root/nginx-fail"
mkdir -p "$nginx_fail_dir"
touch "$nginx_fail_dir/enabled"
write_fake_nginx "$nginx_fail_dir/fake-nginx" 1
write_fake_systemctl "$nginx_fail_dir/fake-systemctl" 0
if run_retire "$nginx_fail_dir"; then
    echo "retirement unexpectedly succeeded when nginx validation failed" >&2
    exit 1
fi
[[ -e "$nginx_fail_dir/enabled" ]]
[[ ! -e "$nginx_fail_dir/state/retired-at" ]]

reload_fail_dir="$test_root/reload-fail"
mkdir -p "$reload_fail_dir"
touch "$reload_fail_dir/enabled"
write_fake_nginx "$reload_fail_dir/fake-nginx" 0
write_fake_systemctl "$reload_fail_dir/fake-systemctl" 1
if run_retire "$reload_fail_dir"; then
    echo "retirement unexpectedly succeeded when nginx reload failed" >&2
    exit 1
fi
[[ -e "$reload_fail_dir/enabled" ]]
[[ ! -e "$reload_fail_dir/state/retired-at" ]]

echo "retirement integration tests passed"
