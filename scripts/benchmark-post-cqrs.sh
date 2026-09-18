#!/usr/bin/env bash
set -euo pipefail

: "${BASE_URL:?set BASE_URL, e.g. http://localhost:8080}"
: "${PUBLIC_POST_ID:?set PUBLIC_POST_ID}"
: "${AUTHOR_POST_ID:?set AUTHOR_POST_ID}"
: "${ACCESS_TOKEN:?set ACCESS_TOKEN}"

requests=200
concurrency=20
workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT

measure() {
  local name=$1 url=$2 method=$3 limit=$4
  local times="$workdir/$name"
  export url method times
  seq "$requests" | xargs -P "$concurrency" -I{} sh -c '
    curl --silent --show-error --output /dev/null --write-out "%{time_total}\n" \
      -X "$method" -H "Authorization: Bearer $ACCESS_TOKEN" "$url" >> "$times"
  '
  local p95
  p95=$(sort -n "$times" | awk -v n="$requests" 'NR == int((n * 95 + 99) / 100) { print; exit }')
  printf '%s p95: %ss\n' "$name" "$p95"
  awk -v value="$p95" -v limit="$limit" 'BEGIN { exit value > limit }'
}

measure read "$BASE_URL/posts/$PUBLIC_POST_ID" GET 0.2
measure write "$BASE_URL/posts/$AUTHOR_POST_ID/likes" POST 0.5
