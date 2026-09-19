#!/usr/bin/env bash
# Compiles the TypeScript implementation (src/typescript) into ./jsetimos.js,
# used by node (node jsetimos.js ...) and by index.html.
#
# Usage:
#   ./build_ts.sh          build jsetimos.js
#   ./build_ts.sh --test   build, then run tests/lang_basic_coverage.jk and the bug
#                          regression tests (tests/bugs)
#
# Uses the pinned TypeScript 3.8.3 (newer versions change the generated code),
# installed on first use into src/typescript/node_modules.

set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project="$root/src/typescript"

if [[ ! -d "$project/node_modules" ]]; then
    (cd "$project" && npm install --no-audit --no-fund --silent)
fi

# The source has known type errors (ES2019 APIs on an ES6 target). tsc still
# writes the output, so only a missing output counts as a failure.
rm -f "$root/jsetimos.js"
(cd "$project" && npx --no-install tsc -p .) > /dev/null || true
if [[ ! -f "$root/jsetimos.js" ]]; then
    (cd "$project" && npx --no-install tsc -p .) || true
    echo "Build failed"
    exit 1
fi
echo "Built jsetimos.js"

if [[ "${1:-}" == "--test" ]]; then
    cd "$root"
    output="$(node jsetimos.js tests/lang_basic_coverage.jk | grep -v "source-map-support" || true)"
    if [[ -n "$output" ]]; then
        echo "$output"
        echo "Tests failed"
        exit 1
    fi
    echo "Tests passed"

    if ! bug_output="$(tests/bugs/run.sh node)"; then
        echo "$bug_output"
        echo "Bug regression tests failed"
        exit 1
    fi
    echo "Bug regression tests passed"
fi
