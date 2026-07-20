import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


class DeploymentAssetTests(unittest.TestCase):
    def test_nginx_routes_frontend_backend_admin_and_uploads(self):
        nginx = (ROOT / "nginx" / "akimirai.xyz.conf").read_text(encoding="utf-8")
        for contract in (
            "server_name akimirai.xyz www.akimirai.xyz",
            "server_name www.akimirai.xyz",
            "server_name akimirai.xyz",
            "return 308 https://akimirai.xyz$request_uri",
            "proxy_pass http://127.0.0.1:3001",
            "proxy_pass http://127.0.0.1:8000",
            "location ^~ /api/",
            "location ^~ /uploads/",
            "location ^~ /admin/",
            "client_max_body_size 12m",
            'proxy_cache off',
        ):
            self.assertIn(contract, nginx)

    def test_systemd_services_are_loopback_and_resource_bounded(self):
        backend = (ROOT / "systemd" / "kirameku-backend.service").read_text(encoding="utf-8")
        frontend = (ROOT / "systemd" / "kirameku-frontend.service").read_text(encoding="utf-8")
        self.assertIn("--host 127.0.0.1 --port 8000", backend)
        self.assertIn("MemoryMax=512M", backend)
        self.assertIn("Environment=HOSTNAME=127.0.0.1", frontend)
        self.assertIn("Environment=PORT=3001", frontend)
        self.assertIn("MemoryMax=768M", frontend)
        self.assertIn("SuccessExitStatus=130", frontend)
        self.assertIn("ReadWritePaths=/opt/kirameku/current/frontend/.next/cache", frontend)
        self.assertIn("ReadWritePaths=/opt/kirameku/current/frontend/.next/server/app", frontend)
        self.assertIn("ReadWritePaths=/opt/kirameku/current/frontend/.next/server/pages", frontend)
        self.assertIn("NoNewPrivileges=true", backend)
        self.assertIn("NoNewPrivileges=true", frontend)
        for contract in (
            "CapabilityBoundingSet=",
            "PrivateDevices=true",
            "ProtectKernelTunables=true",
            "ProtectProc=invisible",
            "RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6",
            "RestrictNamespaces=true",
            "SystemCallArchitectures=native",
        ):
            self.assertIn(contract, backend)
            self.assertIn(contract, frontend)

    def test_runbook_forbids_server_builds_and_preserves_state(self):
        readme = (ROOT / "README.md").read_text(encoding="utf-8")
        self.assertIn("Build on a workstation, never on the production ECS host", readme)
        self.assertIn("Database and upload directories are persistent", readme)
        self.assertNotIn("docker compose", readme.lower())

    def test_release_installer_is_verified_atomic_and_rollback_safe(self):
        installer = (ROOT / "scripts" / "install-release.sh").read_text(encoding="utf-8")
        for contract in (
            "sha256sum -c -",
            "refusing to replace the active release",
            "node_modules/.pnpm",
            "schema_exists=",
            "to_regclass('public.\\\"user\\\"')",
            "mv -Tf /opt/kirameku/current.new /opt/kirameku/current",
            "rollback_on_error",
            "nginx -t",
            "wait_for_url http://127.0.0.1:8000/api/health",
            "wait_for_url http://127.0.0.1:3001/",
        ):
            self.assertIn(contract, installer)


if __name__ == "__main__":
    unittest.main()
