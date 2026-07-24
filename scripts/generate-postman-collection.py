#!/usr/bin/env python3
"""
Generate a Postman Collection v2.1 JSON from:
  - test/api/test-cases.md (markdown table of test cases)
  - api/swagger/swagger.yaml (endpoint definitions for auth/params context)

Usage:
  python3 scripts/generate-postman-collection.py [--input-md test/api/test-cases.md]
                                                 [--swagger api/swagger/swagger.yaml]
                                                 [--output test/api/postman-collection.json]
"""

import argparse
import json
import re
import sys
from pathlib import Path

try:
    import yaml
except ImportError:
    print("Error: PyYAML is required. Install with: pip install pyyaml", file=sys.stderr)
    sys.exit(1)


# ---------------------------------------------------------------------------
# Markdown table parser
# ---------------------------------------------------------------------------

def parse_md_table(md_path: Path) -> list[dict]:
    """Parse the test case markdown table and return a list of test case dicts."""
    text = md_path.read_text()
    rows = []

    # Find the table section: lines that start with "| TC-"
    for line in text.splitlines():
        stripped = line.strip()
        if not stripped.startswith("| TC-"):
            continue

        # Split on | but handle backtick-delimited content carefully
        cells = _split_table_row(stripped)
        if len(cells) < 10:
            continue

        tc = {
            "id": cells[1].strip(),
            "endpoint": cells[2].strip(),
            "method": cells[3].strip().upper(),
            "scenario": cells[4].strip(),
            "type": cells[5].strip(),
            "precondition": cells[6].strip(),
            "request_raw": cells[7].strip(),
            "expected_status": cells[8].strip(),
            "expected_response": cells[9].strip(),
        }
        rows.append(tc)

    return rows


def _split_table_row(line: str) -> list[str]:
    """Split a markdown table row into cells, respecting backtick boundaries."""
    cells = []
    current = []
    in_backtick = False

    for ch in line:
        if ch == "`":
            in_backtick = not in_backtick
            current.append(ch)
        elif ch == "|" and not in_backtick:
            cells.append("".join(current))
            current = []
        else:
            current.append(ch)

    # Don't forget trailing content (last cell before newline)
    if current:
        cells.append("".join(current))

    return cells


# ---------------------------------------------------------------------------
# Swagger spec loader
# ---------------------------------------------------------------------------

def load_swagger(spec_path: Path) -> dict:
    """Load swagger spec and index endpoints by (method, path_pattern)."""
    with open(spec_path) as f:
        spec = yaml.safe_load(f)

    base_path = spec.get("basePath", "").rstrip("/")
    indexed = {}

    for path, methods in spec.get("paths", {}).items():
        for method, details in methods.items():
            method = method.upper()

            # Determine if this endpoint requires auth
            has_security = "security" in details

            # Collect path parameter names
            path_params = []
            body_params = []
            for p in details.get("parameters", []):
                if p.get("in") == "path":
                    path_params.append(p["name"])
                elif p.get("in") == "body":
                    body_params.append(p)

            # Collect responses for status code reference
            responses = {}
            for code, resp_data in details.get("responses", {}).items():
                responses[code] = resp_data

            indexed[(method, path)] = {
                "path": path,
                "method": method,
                "requires_auth": has_security,
                "path_params": path_params,
                "body_params": body_params,
                "responses": responses,
            }

    return indexed


# ---------------------------------------------------------------------------
# Request body/params parser
# ---------------------------------------------------------------------------

def _extract_backtick_balanced(raw: str, start: int) -> tuple[str | None, int]:
    """
    Extract content between backticks starting at `start` (pointing to first
    char after the opening backtick). Uses brace-balancing so nested JSON
    objects do not prematurely close the backtick span. Returns (content, end_pos)
    where end_pos is the index of the closing backtick, or (None, -1) on failure.
    """
    depth = 0
    min_depth = 0  # track how far below 0 we go
    in_str = False
    escape = False
    for i in range(start, len(raw)):
        ch = raw[i]
        if ch == "`" and not in_str and depth <= 0:
            # Accept backtick as closing when depth ≤ 0 (balanced or over-closed)
            return raw[start:i], i
        if escape:
            escape = False
            continue
        if ch == "\\":
            escape = True
            continue
        if ch == '"' and not escape:
            in_str = not in_str
        if not in_str:
            if ch == "{":
                depth += 1
            elif ch == "}":
                depth -= 1
                min_depth = min(min_depth, depth)
    return None, -1


def _extract_json_body(raw: str) -> tuple[str | None, str]:
    """
    Extract JSON body and query string from the raw request column.
    Returns (body_json_string_or_None, query_string).
    """
    body = None
    query = ""

    # ---- Parse "Body:" explicitly (preferred) ----
    body_match = re.search(r"Body:\s*`", raw)
    if body_match:
        body, _ = _extract_backtick_balanced(raw, body_match.end())

    # ---- Fallback: plain JSON in backticks (no "Body:" prefix) ----
    if body is None:
        json_match = re.search(r"`(\{)", raw)
        if json_match:
            body, _ = _extract_backtick_balanced(raw, json_match.end())

    # ---- Parse "Query:" pattern ----
    query_match = re.search(r"Query:\s*`([^`]+)`", raw)
    if query_match:
        query = query_match.group(1).strip()
        if query.startswith("?"):
            query = query[1:]

    return body, query


def _parse_request_params(tc: dict, swagger_entry: dict | None) -> dict:
    """
    Parse the raw request column into actionable request parts.
    Returns dict with keys: body, query_params, path_values, headers, url_path
    """
    raw = tc["request_raw"]
    method = tc["method"]
    endpoint = tc["endpoint"]

    result = {
        "body": None,
        "query_params": "",
        "path_values": {},
        "headers": [],
        "url_path": endpoint,  # defaults
    }

    # ---- Determine if this test has auth ----
    precondition = tc["precondition"].lower()
    type_label = tc["type"].lower()
    is_auth_required = True

    # No auth if: precondition is "none", OR type is security and mentions "missing authorization"
    if precondition == "none":
        is_auth_required = False
    if type_label == "security":
        if "missing authorization" in tc["scenario"].lower():
            is_auth_required = False
        if "invalid" in tc["scenario"].lower() and "token" in tc["scenario"].lower():
            is_auth_required = False

    # Override with swagger info if available
    if swagger_entry:
        if not swagger_entry["requires_auth"]:
            is_auth_required = False
        # If swagger says it needs auth but our precondition says None → still
        # keep requires_auth true (test is about hitting it without auth)

    result["requires_auth"] = is_auth_required

    # ---- Handle "No body" and "No query params" ----
    if raw.lower() in ("no body", "no query params", ""):
        return result

    # ---- Special case: Header injection for invalid tokens ----
    header_match = re.search(r"Header:\s*`([^`]+)`", raw)
    if header_match:
        result["headers"].append({"key": "Authorization", "value": "Bearer invalid_token"})

    # ---- Parse path params ----
    # Pattern: Path: `id=X`, Path: `username=Y`
    path_matches = re.findall(r"Path:\s*`([^`]+)`", raw)
    for pm in path_matches:
        # Could be "id=1" or "id={placeholder}" or "username=johndoe"
        parts = pm.split(",")
        for part in parts:
            part = part.strip()
            if "=" in part:
                k, v = part.split("=", 1)
                k = k.strip()
                v = v.strip()
                result["path_values"][k] = v

    # ---- Parse body and query ----
    body, query = _extract_json_body(raw)
    if body:
        result["body"] = body
    if query:
        result["query_params"] = query

    # ---- Build url_path by substituting path params ----
    url_path = endpoint
    for k, v in result["path_values"].items():
        placeholder = "{" + k + "}"
        if placeholder in url_path:
            url_path = url_path.replace(placeholder, v)
        else:
            # Append as query param if not a path param
            pass

    result["url_path"] = url_path

    return result


# ---------------------------------------------------------------------------
# Status code parsing
# ---------------------------------------------------------------------------

def _parse_expected_status(raw: str) -> int | str:
    """Parse expected status: '200', '201', '400 or 404', etc."""
    raw = raw.strip()
    # Try simple integer
    try:
        return int(raw)
    except ValueError:
        pass
    # Multi-status: "200 or 401"
    if " or " in raw:
        return [int(s.strip()) for s in raw.split(" or ")]
    return raw


# ---------------------------------------------------------------------------
# Test script generation
# ---------------------------------------------------------------------------

def _generate_test_script(tc: dict, status, expected_response: str, body_str: str | None) -> str:
    """Generate a Postman pm.test() JavaScript string."""
    parts = []

    # Status check
    if isinstance(status, list):
        codes = ", ".join(str(s) for s in status)
        parts.append(
            f'pm.test("Status code is one of [{codes}]", function () {{'
            f"\n    pm.expect(pm.response.code).to.be.oneOf({json.dumps(status)});"
            f"\n}});"
        )
    else:
        parts.append(
            f'pm.test("Status code is {status}", function () {{'
            f"\n    pm.response.to.have.status({status});"
            f"\n}});"
        )

    # Response time
    parts.append(
        'pm.test("Response time is less than 1000ms", function () {'
        "\n    pm.expect(pm.response.responseTime).to.be.below(1000);"
        "\n});"
    )

    # ---- Response schema / field validation ----
    resp = expected_response.strip()

    # Detect ProblemDetail pattern
    is_problem_detail = (
        '"type"' in resp
        and '"title"' in resp
        and '"status"' in resp
        and '"detail"' in resp
    )

    if is_problem_detail:
        parts.append(
            'pm.test("Response is ProblemDetail", function () {'
            "\n    const r = pm.response.json();"
            '\n    pm.expect(r).to.have.property("status");'
            '\n    pm.expect(r).to.have.property("title");'
            "\n});"
        )

    # Detect JSON schema pattern: `{"key":"type","key2":"type"}`
    json_schema_match = re.match(r"^`(\{.*\})`$", resp)
    if json_schema_match:
        schema_str = json_schema_match.group(1)
        try:
            schema = json.loads(schema_str)
        except json.JSONDecodeError:
            schema = None

        if schema and isinstance(schema, dict):
            parts.append(
                'pm.test("Response matches expected schema", function () {'
                "\n    const r = pm.response.json();"
            )
            for key, val in schema.items():
                js_type = _json_type_to_js(val)
                parts.append(f'    pm.expect(r).to.have.property("{key}");')
                if js_type:
                    parts.append(
                        f"    pm.expect(r.{key}).to.be.a"
                        f'("{js_type}");'
                    )
            parts.append("});")

    # Detect "Array of X objects" or "Array (may be empty)"
    elif resp.lower().startswith("array"):
        parts.append(
            'pm.test("Response is an array", function () {'
            "\n    pm.expect(pm.response.json()).to.be.an("
            '"array");'
            "\n});"
        )

    # Detect "Empty array `[]`"
    elif "empty array" in resp.lower() or resp == "`[]`":
        parts.append(
            'pm.test("Response is an array", function () {'
            "\n    const r = pm.response.json();"
            '\n    pm.expect(r).to.be.an("array");'
            "\n});"
        )

    # Detect JSON response with backtick-wrapped JSON like `{"message":"...","username":"..."}`
    elif resp.startswith("`{") and resp.endswith("}`"):
        inner = resp[1:-1]  # strip backticks
        try:
            example = json.loads(inner)
        except json.JSONDecodeError:
            example = None

        if example and isinstance(example, dict):
            parts.append(
                'pm.test("Response has expected fields", function () {'
                "\n    const r = pm.response.json();"
            )
            for key, val in example.items():
                parts.append(f'    pm.expect(r).to.have.property("{key}");')
                js_type = _json_type_to_js(val)
                if js_type:
                    parts.append(
                        f"    pm.expect(r.{key}).to.be.a"
                        f'("{js_type}");'
                    )
            parts.append("});")

    # Detect "PostResponse object with `...`" or "Object with keys: `...`"
    elif "object with" in resp.lower() and "`" in resp:
        key_match = re.search(r"`([a-z_,]+)`", resp, re.IGNORECASE)
        if key_match:
            keys = [k.strip() for k in key_match.group(1).split(",")]
            parts.append(
                'pm.test("Response has expected fields", function () {'
                "\n    const r = pm.response.json();"
            )
            for key in keys:
                parts.append(f'    pm.expect(r).to.have.property("{key}");')
            parts.append("});")

    # Detect "(no schema, success)" or "(object)" or "JSON object"
    elif (
        "(no schema" in resp.lower()
        or resp in ("(object)", "`{}` (object)")
        or resp.lower().startswith("json object")
        or resp.lower() == "empty or success response"
    ):
        if status not in (400, 401, 403, 404, 500):
            if isinstance(status, int):
                code = status
            elif isinstance(status, list):
                code = status[0]
            else:
                code = 200
            if code == 201:
                parts.append(
                    'pm.test("Response received successfully", function () {'
                    f"\n    pm.expect(pm.response.code).to.equal(201);"
                    "\n});"
                )
            else:
                parts.append(
                    'pm.test("Response is valid object", function () {'
                    "\n    const r = pm.response.json();"
                    '\n    pm.expect(r).to.be.an("object");'
                    "\n});"
                )

    # ---- Token extraction for login ----
    tc_scenario_lower = tc["scenario"].lower()
    tc_type_lower = tc["type"].lower()
    is_contract_test = tc_type_lower == "contract"
    is_login_endpoint = tc["method"] == "POST" and "/auth/login" in tc["endpoint"]
    is_positive_or_edge = tc_type_lower in ("positive", "edge case")

    if is_login_endpoint and is_positive_or_edge and not is_contract_test and "200" in str(status):
        parts.append(
            "// Extract token from cookies set by the API\n"
            "const cookies = pm.cookies.toObject();\n"
            'if (cookies.access_token) {\n'
            '    pm.collectionVariables.set("token", cookies.access_token);\n'
            "}"
        )

    return "\n".join(parts)


def _json_type_to_js(val) -> str | None:
    """Map a JSON literal to a JavaScript typeof string."""
    if val == "...":
        return None  # placeholder, don't assert type
    if isinstance(val, bool):
        return "boolean"
    if isinstance(val, int):
        return "number"
    if isinstance(val, float):
        return "number"
    if isinstance(val, str):
        return "string"
    if isinstance(val, list):
        return None  # could be array or object
    if isinstance(val, dict):
        return "object"
    return None


def type_label(tc: dict) -> str:
    return tc.get("type", "").strip()


# ---------------------------------------------------------------------------
# Postman collection builder
# ---------------------------------------------------------------------------

def build_collection(test_cases: list[dict], swagger_index: dict) -> dict:
    """Build the Postman Collection v2.1 JSON structure."""
    collection = {
        "info": {
            "name": "User Service API - Test Cases",
            "description": (
                "Postman collection auto-generated from "
                "test/api/test-cases.md and api/swagger/swagger.yaml"
            ),
            "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
        },
        "variable": [
            {"key": "base_url", "value": "http://localhost:8080"},
            {"key": "token", "value": ""},
            {"key": "test_email", "value": "test@example.com"},
            {"key": "test_password", "value": "password123"},
        ],
        "item": [],
    }

    # Group by endpoint
    groups: dict[str, list[dict]] = {}
    for tc in test_cases:
        endpoint = tc["endpoint"]
        # Normalize endpoint for grouping: strip leading / and replace {placeholder} with {param}
        key = endpoint.lstrip("/")
        if key not in groups:
            groups[key] = []
        groups[key].append(tc)

    for folder_name, cases in groups.items():
        folder = {"name": folder_name, "item": []}

        for tc in cases:
            method = tc["method"]
            endpoint = tc["endpoint"]
            status = _parse_expected_status(tc["expected_status"])

            # Find swagger entry
            swagger_key = _find_swagger_key(endpoint, swagger_index)
            swagger_entry = swagger_index.get(swagger_key)

            # Parse request params
            params = _parse_request_params(tc, swagger_entry)

            # Build URL
            url_path = params["url_path"]
            full_url = "{{base_url}}" + url_path
            if params["query_params"]:
                full_url += "?" + params["query_params"]

            # Parse URL path segments
            path_segments = url_path.strip("/").split("/")

            # Build request
            request = {
                "method": method,
                "header": [
                    {"key": "Content-Type", "value": "application/json"}
                ],
                "url": {
                    "raw": full_url,
                    "host": ["{{base_url}}"],
                    "path": path_segments,
                },
            }

            # Auth
            if params["requires_auth"]:
                request["auth"] = {
                    "type": "bearer",
                    "bearer": [
                        {
                            "key": "token",
                            "value": "{{token}}",
                            "type": "string",
                        }
                    ],
                }
            else:
                request["auth"] = {"type": "noauth"}

            # Extra headers (e.g., invalid token for security tests)
            if params["headers"]:
                request["header"].extend(params["headers"])

            # Body
            if params["body"] is not None:
                request["body"] = {
                    "mode": "raw",
                    "raw": params["body"],
                }
            elif method in ("POST", "PUT", "PATCH") and params["body"] is None:
                # Some POST/PUT with path params only (e.g., follow) have no body
                pass

            # Build test script
            test_script = _generate_test_script(
                tc, status, tc["expected_response"], params["body"]
            )

            name = f"{tc['id']} - {tc['scenario']}"
            item = {
                "name": name,
                "request": request,
                "response": [],
                "event": [
                    {
                        "listen": "test",
                        "script": {
                            "exec": [test_script],
                            "type": "text/javascript",
                        },
                    }
                ],
            }

            folder["item"].append(item)

        collection["item"].append(folder)

    return collection


def _find_swagger_key(endpoint: str, swagger_index: dict) -> tuple | None:
    """
    Find the swagger key that matches the test case endpoint.
    Test case endpoints have {placeholder} but swagger paths have {param}.
    Both use {name} syntax so they should match directly,
    but we try exact match first, then fallback.
    """
    # Try all known (method, path) combos
    candidates = []
    for (method, path) in swagger_index:
        # Normalize both for comparison
        norm_ep = endpoint.rstrip("/")
        norm_path = path.rstrip("/")
        if norm_ep == norm_path:
            candidates.append((method, path))

    # Return the one matching a method in the candidates
    # (same endpoint can have GET/POST/PUT/DELETE)
    for (method, path) in candidates:
        return (method, path)

    # Fallback: try matching by base path ignoring path param values
    ep_parts = endpoint.strip("/").split("/")
    for (method, path) in swagger_index:
        sw_parts = path.strip("/").split("/")
        if len(ep_parts) != len(sw_parts):
            continue
        match = True
        for a, b in zip(ep_parts, sw_parts):
            if a == b:
                continue
            if a.startswith("{") and b.startswith("{"):
                continue
            match = False
            break
        if match:
            return (method, path)

    return None


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------


def main():
    parser = argparse.ArgumentParser(
        description="Generate Postman Collection from test case markdown"
    )
    parser.add_argument(
        "--input-md",
        default="test/api/test-cases.md",
        help="Path to test cases markdown file",
    )
    parser.add_argument(
        "--swagger",
        default="api/swagger/swagger.yaml",
        help="Path to swagger spec YAML",
    )
    parser.add_argument(
        "--output",
        default="test/api/postman-collection.json",
        help="Output path for Postman collection JSON",
    )
    args = parser.parse_args()

    repo_root = Path(__file__).resolve().parent.parent
    md_path = repo_root / args.input_md
    swagger_path = repo_root / args.swagger
    out_path = repo_root / args.output

    # Validate inputs
    if not md_path.exists():
        print(f"Error: Test cases file not found: {md_path}", file=sys.stderr)
        sys.exit(1)
    if not swagger_path.exists():
        print(f"Error: Swagger spec not found: {swagger_path}", file=sys.stderr)
        sys.exit(1)

    # Parse
    print(f"Parsing test cases from: {md_path}")
    test_cases = parse_md_table(md_path)
    print(f"  Found {len(test_cases)} test cases")

    print(f"Loading swagger spec from: {swagger_path}")
    swagger_index = load_swagger(swagger_path)
    print(f"  Found {len(swagger_index)} endpoint definitions")

    # Build collection
    collection = build_collection(test_cases, swagger_index)

    total_requests = sum(
        len(folder["item"]) for folder in collection["item"]
    )
    print(f"Generated collection with {len(collection['item'])} folders, {total_requests} requests")

    # Write output
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with open(out_path, "w") as f:
        json.dump(collection, f, indent=2)
    print(f"Written to: {out_path}")


if __name__ == "__main__":
    main()
