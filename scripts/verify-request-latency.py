#!/usr/bin/env python3
"""Verify authenticated Project One read/write p95 latency budgets."""

import argparse
import http.cookiejar
import json
import math
import os
import sys
import time
import urllib.error
import urllib.request


def percentile_95(values):
    """Return the nearest-rank 95th percentile."""
    return sorted(values)[max(0, math.ceil(0.95 * len(values)) - 1)]


def request(opener, url, method="GET", body=None, timeout=5.0):
    data = None if body is None else json.dumps(body).encode("utf-8")
    headers = {"Content-Type": "application/json"} if data is not None else {}
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    started = time.perf_counter()
    try:
        with opener.open(req, timeout=timeout) as response:
            response.read()
            status = response.status
    except urllib.error.HTTPError as error:
        error.read()
        status = error.code
    return status, (time.perf_counter() - started) * 1000


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-url", default=os.getenv("API_BASE_URL", "http://localhost:8080"))
    parser.add_argument("--samples", type=int, default=int(os.getenv("LATENCY_SAMPLES", "20")))
    parser.add_argument("--email", default=os.getenv("PROJECT1_EMAIL"))
    parser.add_argument("--password", default=os.getenv("PROJECT1_PASSWORD"))
    args = parser.parse_args()
    if not args.email or not args.password:
        parser.error("set PROJECT1_EMAIL and PROJECT1_PASSWORD (or pass --email and --password)")
    if args.samples < 2:
        parser.error("--samples must be at least 2")

    base = args.base_url.rstrip("/")
    opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
    status, _ = request(opener, base + "/auth/login", "POST", {"email": args.email, "password": args.password})
    if status != 200:
        raise SystemExit("FAIL: login returned HTTP %d" % status)

    reads = []
    writes = []
    run_id = int(time.time() * 1000)
    for sequence in range(args.samples):
        status, elapsed = request(opener, base + "/notifications")
        if status != 200:
            raise SystemExit("FAIL: authenticated read returned HTTP %d" % status)
        reads.append(elapsed)

        status, elapsed = request(
            opener,
            base + "/posts",
            "POST",
            {
                "title": "Latency acceptance %d-%d" % (run_id, sequence),
                "content": "Representative authenticated write latency sample.",
                "tags": ["acceptance"],
            },
        )
        if status != 201:
            raise SystemExit("FAIL: authenticated write returned HTTP %d" % status)
        writes.append(elapsed)

    read_p95 = percentile_95(reads)
    write_p95 = percentile_95(writes)
    print("authenticated read p95: %.2f ms (must be < 200 ms)" % read_p95)
    print("authenticated write p95: %.2f ms (must be < 500 ms)" % write_p95)
    if read_p95 >= 200 or write_p95 >= 500:
        print("FAIL: request latency budget exceeded", file=sys.stderr)
        return 1
    print("Request latency check passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
