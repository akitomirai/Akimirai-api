#!/usr/bin/env bash
set -Eeuo pipefail

release="${1:?usage: install-release.sh RELEASE FRONTEND_SHA256 BACKEND_SHA256}"
frontend_sha="${2:?frontend sha256 required}"
backend_sha="${3:?backend sha256 required}"

release_root="/opt/kirameku/releases/$release"
frontend_archive="/tmp/kirameku-frontend-$release.tar.gz"
backend_archive="/tmp/kirameku-backend-$release.tar.gz"
previous_release="$(readlink -f /opt/kirameku/current 2>/dev/null || true)"
switched=0

rollback_on_error() {
  local rc=$?
  if [[ $switched -eq 1 && -n "$previous_release" && -d "$previous_release" ]]; then
    ln -sfn "$previous_release" /opt/kirameku/current.rollback
    mv -Tf /opt/kirameku/current.rollback /opt/kirameku/current
    systemctl restart kirameku-backend.service kirameku-frontend.service || true
  fi
  exit "$rc"
}
trap rollback_on_error ERR

wait_for_url() {
  local url=$1
  for _ in $(seq 1 30); do
    if curl -fsS "$url" >/dev/null; then
      return 0
    fi
    sleep 1
  done
  curl -fsS "$url" >/dev/null
}

single_pnpm_target() {
  local pattern=$1
  local -a matches
  mapfile -t matches < <(find "$release_root/frontend/node_modules/.pnpm" \
    -mindepth 3 -maxdepth 3 -type d -path "$pattern" | sort)
  [[ ${#matches[@]} -eq 1 ]] || {
    echo "expected one pnpm target for $pattern, found ${#matches[@]}" >&2
    return 1
  }
  printf '%s\n' "${matches[0]}"
}

printf '%s  %s\n' "$frontend_sha" "$frontend_archive" | sha256sum -c -
printf '%s  %s\n' "$backend_sha" "$backend_archive" | sha256sum -c -

resolved_release="$(readlink -m -- "$release_root")"
[[ "$resolved_release" == /opt/kirameku/releases/* ]] || {
  echo "refusing unexpected release path: $resolved_release" >&2
  exit 1
}
[[ -z "$previous_release" || "$resolved_release" != "$previous_release" ]] || {
  echo "refusing to replace the active release: $resolved_release" >&2
  exit 1
}

rm -rf -- "$resolved_release"
install -d -o root -g root -m 0755 "$release_root/backend" "$release_root/frontend"
tar -xzf "$backend_archive" -C "$release_root/backend"
tar -xzf "$frontend_archive" -C "$release_root/frontend"

# Windows pnpm uses junctions. A dereferenced tar contains duplicate top-level
# modules, so restore the portable Linux links to the bundled .pnpm store.
pnpm_root="$release_root/frontend/node_modules/.pnpm"
next_target="$(single_pnpm_target "$pnpm_root/next@*/node_modules/next")"
react_target="$(single_pnpm_target "$pnpm_root/react@*/node_modules/react")"
react_dom_target="$(single_pnpm_target "$pnpm_root/react-dom@*/node_modules/react-dom")"
for module in next react react-dom; do
  rm -rf -- "$release_root/frontend/node_modules/$module"
done
ln -s "${next_target#"$release_root/frontend/node_modules/"}" "$release_root/frontend/node_modules/next"
ln -s "${react_target#"$release_root/frontend/node_modules/"}" "$release_root/frontend/node_modules/react"
ln -s "${react_dom_target#"$release_root/frontend/node_modules/"}" "$release_root/frontend/node_modules/react-dom"

chown -R root:root "$release_root"
find "$release_root" -type d -exec chmod 0755 {} +
find "$release_root" -type f -exec chmod u=rw,go=r {} +
install -d -o kirameku -g kirameku -m 0750 \
  "$release_root/frontend/.next/cache" \
  "$release_root/frontend/.next/server/app" \
  "$release_root/frontend/.next/server/pages"
chown -R kirameku:kirameku \
  "$release_root/frontend/.next/cache" \
  "$release_root/frontend/.next/server/app" \
  "$release_root/frontend/.next/server/pages"

/opt/kirameku/venv/bin/pip install --no-cache-dir \
  -r "$release_root/backend/requirements.txt"

schema_exists="$(docker exec sub2api-postgres psql -U sub2api -d kirameku -Atc \
  "SELECT to_regclass('public.\"user\"') IS NOT NULL")"
if [[ "$schema_exists" == f ]]; then
  {
    printf 'SET ROLE kirameku;\n'
    cat "$release_root/backend/init_db.sql"
  } | docker exec -i sub2api-postgres psql -v ON_ERROR_STOP=1 \
    -U sub2api -d kirameku
fi

if [[ -d "$release_root/backend/migrations" ]]; then
  while IFS= read -r -d '' migration; do
    {
      printf 'SET ROLE kirameku;\n'
      cat "$migration"
    } | docker exec -i sub2api-postgres psql -v ON_ERROR_STOP=1 \
      -U sub2api -d kirameku
  done < <(find "$release_root/backend/migrations" -maxdepth 1 -type f \
    -name '*.sql' -print0 | sort -z)
fi

ln -sfn "$release_root" /opt/kirameku/current.new
mv -Tf /opt/kirameku/current.new /opt/kirameku/current
switched=1

install -o root -g root -m 0644 /tmp/kirameku-backend.service \
  /etc/systemd/system/kirameku-backend.service
install -o root -g root -m 0644 /tmp/kirameku-frontend.service \
  /etc/systemd/system/kirameku-frontend.service
systemctl daemon-reload
systemctl enable --now kirameku-backend.service kirameku-frontend.service
systemctl restart kirameku-backend.service kirameku-frontend.service

wait_for_url http://127.0.0.1:8000/api/health
wait_for_url http://127.0.0.1:3001/

nginx -t

rm -f "$frontend_archive" "$backend_archive" \
  /tmp/kirameku-backend.service /tmp/kirameku-frontend.service
trap - ERR

echo "release=$release"
systemctl --no-pager --full status \
  kirameku-backend.service kirameku-frontend.service | sed -n '1,28p'
