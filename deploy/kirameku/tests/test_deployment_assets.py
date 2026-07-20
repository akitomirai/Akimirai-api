import importlib.util
from http.cookiejar import Cookie, CookieJar
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


class DeploymentAssetTests(unittest.TestCase):
    def test_site_access_cookie_jar_clone_is_independent(self):
        script = ROOT / "scripts" / "verify-site-access.py"
        spec = importlib.util.spec_from_file_location("verify_site_access", script)
        self.assertIsNotNone(spec)
        module = importlib.util.module_from_spec(spec)
        assert spec.loader is not None
        spec.loader.exec_module(module)

        source = CookieJar()
        source.set_cookie(
            Cookie(
                version=0,
                name="kirameku_site_access",
                value="opaque",
                port=None,
                port_specified=False,
                domain="akimirai.xyz",
                domain_specified=False,
                domain_initial_dot=False,
                path="/",
                path_specified=True,
                secure=True,
                expires=None,
                discard=True,
                comment=None,
                comment_url=None,
                rest={"HttpOnly": None, "SameSite": "lax"},
                rfc2109=False,
            )
        )

        clone = module.clone_cookie_jar(source)
        clone.clear()
        self.assertEqual(len(list(source)), 1)
        self.assertEqual(len(list(clone)), 0)

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

    def test_nginx_enforces_site_access_without_exposing_all_auth_routes(self):
        nginx = (ROOT / "nginx" / "akimirai.xyz.conf").read_text(encoding="utf-8")
        rate_limit = (ROOT / "nginx" / "kirameku-access-limit.conf").read_text(
            encoding="utf-8"
        )
        activator = (ROOT / "scripts" / "activate-site-access.sh").read_text(
            encoding="utf-8"
        )

        for contract in (
            "location = /_site_access_verify",
            "internal;",
            "auth_request /_site_access_verify;",
            "location = /access",
            "location = /api/site-access/login",
            "location = /api/site-access/logout",
            "error_page 401 = @site_access_login",
            "limit_req zone=kirameku_site_access",
            "limit_req_status 429",
        ):
            self.assertIn(contract, nginx)
        self.assertIn("limit_req_zone", rate_limit)
        self.assertNotIn("location ^~ /api/site-access/", nginx)
        self.assertIn("nginx -t", activator)
        self.assertIn("akimirai.xyz.before-site-access", activator)

    def test_release_install_does_not_enable_gate_before_password_bootstrap(self):
        installer = (ROOT / "scripts" / "install-release.sh").read_text(encoding="utf-8")
        readme = (ROOT / "README.md").read_text(encoding="utf-8")

        self.assertNotIn("/etc/nginx/sites-available/akimirai.xyz", installer)
        self.assertIn("bootstrap_site_access.py --password-stdin", readme)
        self.assertIn("activate-site-access.sh", readme)

    def test_site_access_verifier_checks_cookie_and_logout_contracts(self):
        verifier = (ROOT / "scripts" / "verify-site-access.py").read_text(
            encoding="utf-8"
        )
        for contract in (
            "getpass(",
            "cookie.domain_specified",
            "cookie.secure",
            '"HttpOnly"',
            '"SameSite"',
            'expect(request("/api/site-access/logout"',
            '"--check-rotation"',
            '"stale cookie"',
            "clone.set_cookie(copy(cookie))",
            'print("site access verification passed")',
        ):
            self.assertIn(contract, verifier)
        self.assertNotIn("print(password", verifier)


if __name__ == "__main__":
    unittest.main()
