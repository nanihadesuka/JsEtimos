#!/usr/bin/env bash
# Builds the Go implementation (src/go) into the repository root.
#
# Usage:
#   ./build_go.sh          build ./jsetimos (./jsetimos.exe on Windows)
#   ./build_go.sh --test   build, then run tests/lang_basic_coverage.jk and the bug
#                          regression tests (tests/bugs)

set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
binary="jsetimos$(go env GOEXE)"

go build -C "$root/src/go" -o "$root/$binary" .
echo "Built $binary"

if [[ "${1:-}" == "--test" ]]; then
    cd "$root"
    output="$("./$binary" tests/lang_basic_coverage.jk)"
    if [[ -n "$output" ]]; then
        echo "$output"
        echo "Tests failed"
        exit 1
    fi
    echo "Tests passed"

    if ! bug_output="$(tests/bugs/run.sh go)"; then
        echo "$bug_output"
        echo "Bug regression tests failed"
        exit 1
    fi
    echo "Bug regression tests passed"
fi
