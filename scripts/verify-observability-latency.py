#!/usr/bin/env python3
"""Verify the observability latency budget of a running Project One API.

Each endpoint is attempted 100 times and at least 99 attempts must complete
within five seconds:

  * GET /metrics  authenticated with the dedicated monitoring credential
  * GET /healthz  process liveness
  * GET /status   readiness, where 200 and 503 are both valid answers

Usage:
  scripts/verify-observability-latency.py [base_url]

Environment overrides:
  API_BASE_URL              default http://localhost:8080
  REQUESTS                  attempts per endpoint, default 100
  REQUIRED                  attempts that must stay within budget, default 99
  BUDGET_SECONDS            per-request budget, default 5
  MONITORING_USERNAME       default projectone-metrics
  MONITORING_PASSWORD       used when set, otherwise read from the file below
  MONITORING_PASSWORD_FILE  default deployments/observability/secrets/metrics-password

Run it from the repository root against a healthy stack.
"""

import argparse
import base64
import os
import sys
import time
import urllib.error
import urllib.request

DEFAULT_BASE_URL = "http://localhost:8080"
DEFAULT_USERNAME = "projectone-metrics"
DEFAULT_PASSWORD_FILE = "deployments/observability/secrets/metrics-password"

# A scrape must be authorized. Readiness may legitimately report 503 while a
# required dependency is unavailable, so both answers count as a valid reply.
EXPECTED_STATUS = {
    "metrics": (200,),
    "healthz": (200,),
    "status": (200, 503),
}


def env_number(name, default, cast):
    """Read a numeric environment override, failing clearly on bad input."""
    raw = os.environ.get(name)
    if raw is None or raw == "":
        return default
    try:
        return cast(raw)
    except ValueError:
        raise SystemExit("error: %s must be a number, got %r" % (name, raw))


def parse_args():
    parser = argparse.ArgumentParser(
        description="Verify that health and metrics endpoints stay within the latency budget.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument(
        "base_url",
        nargs="?",
        default=os.environ.get("API_BASE_URL") or DEFAULT_BASE_URL,
        help="API base URL (default: %(default)s)",
    )
    parser.add_argument(
        "--requests",
        type=int,
        default=env_number("REQUESTS", 100, int),
        help="attempts per endpoint (default: %(default)s)",
    )
    parser.add_argument(
        "--required",
        type=int,
        default=env_number("REQUIRED", 99, int),
        help="attempts that must stay within budget (default: %(default)s)",
    )
    parser.add_argument(
        "--budget-seconds",
        type=float,
        default=env_number("BUDGET_SECONDS", 5.0, float),
        help="per-request budget in seconds (default: %(default)s)",
    )
    parser.add_argument(
        "--username",
        default=os.environ.get("MONITORING_USERNAME") or DEFAULT_USERNAME,
        help="monitoring username used for /metrics (default: %(default)s)",
    )
    parser.add_argument(
        "--password",
        default=os.environ.get("MONITORING_PASSWORD", ""),
        help="monitoring password; falls back to --password-file",
    )
    parser.add_argument(
        "--password-file",
        default=os.environ.get("MONITORING_PASSWORD_FILE") or DEFAULT_PASSWORD_FILE,
        help="file holding the monitoring password (default: %(default)s)",
    )
    return parser.parse_args()


def resolve_password(args):
    """Return the monitoring password, read from its mounted secret file when needed."""
    if args.password:
        return args.password

    if not os.path.isfile(args.password_file):
        raise SystemExit(
            "error: set MONITORING_PASSWORD or create %s" % args.password_file
        )

    with open(args.password_file, encoding="utf-8") as handle:
        password = handle.read().strip()
    if not password:
        raise SystemExit("error: the monitoring password file is empty: %s" % args.password_file)
    return password


def attempt(url, timeout, username=None, password=None):
    """Perform one request and return its status code and elapsed seconds.

    Returns a status of None when the request never produced a reply, which
    counts as a failed attempt.
    """
    request = urllib.request.Request(url)
    if username is not None:
        credentials = base64.b64encode(("%s:%s" % (username, password)).encode("utf-8"))
        request.add_header("Authorization", "Basic " + credentials.decode("ascii"))

    started = time.perf_counter()
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            response.read()
            status = response.status
    except urllib.error.HTTPError as error:
        # An HTTP error is still a timely, valid reply for unauthenticated
        # probes; the expected status list decides whether it is acceptable.
        error.read()
        status = error.code
    except (urllib.error.URLError, OSError):
        status = None
    return status, time.perf_counter() - started


def verify_route(route, url, args, password, username=None):
    """Attempt one endpoint repeatedly and report whether it met the budget."""
    expected = EXPECTED_STATUS[route]
    within_budget = 0

    for _ in range(args.requests):
        status, elapsed = attempt(
            url,
            timeout=args.budget_seconds,
            username=username,
            password=password,
        )
        if status in expected and elapsed <= args.budget_seconds:
            within_budget += 1

    print(
        "%-9s %3d/%d attempts within %gs (expected %s)"
        % (
            route,
            within_budget,
            args.requests,
            args.budget_seconds,
            "/".join(str(code) for code in expected),
        )
    )

    if within_budget < args.required:
        print(
            "FAIL: %s completed %d of %d attempts in time, need at least %d"
            % (route, within_budget, args.requests, args.required),
            file=sys.stderr,
        )
        return False
    return True


def main():
    args = parse_args()
    base_url = args.base_url.rstrip("/")

    if args.requests < 1 or args.required < 1 or args.required > args.requests:
        raise SystemExit("error: required must be between 1 and requests")

    password = resolve_password(args)

    print("Checking observability latency against %s" % base_url)

    results = [
        verify_route(
            "metrics",
            base_url + "/metrics",
            args,
            password,
            username=args.username,
        ),
        verify_route("healthz", base_url + "/healthz", args, password),
        verify_route("status", base_url + "/status", args, password),
    ]

    failures = results.count(False)
    if failures:
        print(
            "Observability latency check FAILED (%d endpoint(s) out of budget)" % failures,
            file=sys.stderr,
        )
        return 1

    print("Observability latency check passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
