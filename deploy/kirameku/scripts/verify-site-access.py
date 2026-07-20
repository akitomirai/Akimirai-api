#!/usr/bin/env python3
"""Verify the production site-access contract without logging secrets."""

from __future__ import annotations

import argparse
from copy import copy
from getpass import getpass
import json
import sys
import time
from http.cookiejar import CookieJar
from urllib.error import HTTPError
from urllib.parse import urljoin
from urllib.request import (
    HTTPCookieProcessor,
    HTTPRedirectHandler,
    Request,
    build_opener,
)


COOKIE_NAME = "kirameku_site_access"


class NoRedirect(HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def clone_cookie_jar(source: CookieJar) -> CookieJar:
    clone = CookieJar()
    for cookie in source:
        clone.set_cookie(copy(cookie))
    return clone


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Verify Kirameku site access")
    parser.add_argument("--base-url", default="https://akimirai.xyz")
    parser.add_argument(
        "--password-stdin",
        action="store_true",
        help="Read one password line from standard input instead of prompting",
    )
    parser.add_argument(
        "--check-rotation",
        action="store_true",
        help="Also verify admin-authorized version rotation using the same password",
    )
    return parser.parse_args()


def expect(actual: int, expected: int, label: str) -> None:
    if actual != expected:
        raise RuntimeError(f"{label}: expected {expected}, got {actual}")


def main() -> int:
    args = parse_args()
    base_url = args.base_url.rstrip("/") + "/"
    password = (
        sys.stdin.readline().rstrip("\r\n")
        if args.password_stdin
        else getpass("Site access password: ")
    )
    if not password:
        raise SystemExit("password is required")

    jar = CookieJar()
    opener = build_opener(HTTPCookieProcessor(jar), NoRedirect())

    def request(
        path: str,
        *,
        method: str = "GET",
        payload: dict | None = None,
        client=None,
        extra_headers: dict[str, str] | None = None,
    ):
        data = json.dumps(payload).encode("utf-8") if payload is not None else None
        headers = {"Content-Type": "application/json"} if data is not None else {}
        headers.update(extra_headers or {})
        req = Request(urljoin(base_url, path.lstrip("/")), data=data, headers=headers, method=method)
        try:
            with (client or opener).open(req, timeout=15) as response:
                return response.status, response.headers, response.read()
        except HTTPError as error:
            return error.code, error.headers, error.read()

    for path, status in (
        ("/", 302),
        ("/posts", 302),
        ("/admin/", 302),
        ("/uploads/site-access-missing.png", 302),
        ("/api/posts", 401),
        ("/api/health", 200),
        ("/access", 200),
        ("/images/1.webp", 200),
    ):
        expect(request(path)[0], status, f"anonymous {path}")

    wrong_password = password + "x" if len(password) < 128 else ("x" + password[1:])
    wrong_status, wrong_headers, _ = request(
        "/api/site-access/login",
        method="POST",
        payload={"password": wrong_password},
    )
    expect(wrong_status, 401, "wrong password")
    if wrong_headers.get("Set-Cookie"):
        raise RuntimeError("wrong password unexpectedly set a cookie")

    login_status, _, _ = request(
        "/api/site-access/login",
        method="POST",
        payload={"password": password},
    )
    expect(login_status, 204, "correct password")

    cookie = next((item for item in jar if item.name == COOKIE_NAME), None)
    if cookie is None:
        raise RuntimeError("site access cookie is missing")
    if cookie.domain_specified:
        raise RuntimeError("site access cookie must be host-only")
    if not cookie.secure or cookie.path != "/":
        raise RuntimeError("site access cookie has unsafe scope")
    if "HttpOnly" not in cookie._rest:
        raise RuntimeError("site access cookie is not HttpOnly")
    if str(cookie._rest.get("SameSite", "")).lower() != "lax":
        raise RuntimeError("site access cookie is not SameSite=Lax")
    if cookie.expires is None or cookie.expires < time.time() + 29 * 24 * 60 * 60:
        raise RuntimeError("site access cookie lifetime is shorter than 29 days")

    for path, status in (
        ("/", 200),
        ("/api/posts", 200),
        ("/admin/", 200),
        ("/uploads/site-access-missing.png", 404),
    ):
        expect(request(path)[0], status, f"authenticated {path}")

    if args.check_rotation:
        admin_username = getpass("Admin username: ")
        admin_password = getpass("Admin password: ")
        old_jar = clone_cookie_jar(jar)
        old_opener = build_opener(HTTPCookieProcessor(old_jar), NoRedirect())
        old_cookie_value = next(item.value for item in old_jar if item.name == COOKIE_NAME)

        auth_status, _, auth_body = request(
            "/api/auth/login",
            method="POST",
            payload={"username": admin_username, "password": admin_password},
        )
        expect(auth_status, 200, "admin login")
        token = json.loads(auth_body)["data"]["accessToken"]
        authorization = {"Authorization": f"Bearer {token}"}

        rotation_status, _, _ = request(
            "/api/site-access/admin/password",
            method="PUT",
            payload={"password": password, "confirm_password": password},
            extra_headers=authorization,
        )
        expect(rotation_status, 200, "password rotation")

        current_cookie_value = next(item.value for item in jar if item.name == COOKIE_NAME)
        if current_cookie_value == old_cookie_value:
            raise RuntimeError("password rotation did not refresh the current cookie")
        expect(
            request("/api/site-access/verify", client=old_opener)[0],
            401,
            "stale cookie",
        )
        expect(request("/api/site-access/verify")[0], 204, "refreshed cookie")

        fresh_jar = CookieJar()
        fresh_opener = build_opener(HTTPCookieProcessor(fresh_jar), NoRedirect())
        fresh_status, _, _ = request(
            "/api/site-access/login",
            method="POST",
            payload={"password": password},
            client=fresh_opener,
        )
        expect(fresh_status, 204, "login after rotation")

    expect(request("/api/site-access/logout", method="POST")[0], 204, "logout")
    expect(request("/")[0], 302, "post-logout root")
    print("site access verification passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
