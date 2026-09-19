#!/usr/bin/env bash
# Runs the regression tests for fixed interpreter bugs (one tests/bugs/bug*.jk file per bug).
#
# Usage: tests/bugs/run.sh [node|go|both]   (default: both)
#
# Each tests/bugs/bug*.jk file is run on its own, so a crash only fails that test.
# A test passes when the interpreter exits with status 0 and prints nothing
# (no FAILED ASSERT, no error), unless the file header changes the expectations:
#
#   // args: <arguments>       extra command line arguments
#   // mode: shell             feed the file to the interactive shell instead
#   // expect-output: <text>   the output must contain <text> (can repeat);
#                              any other output is then allowed
#   // expect-file: <path>     after the run, <path> (relative to the repository
#                              root) must exist and not be empty (can repeat)
#   // reject-output: <text>   the output must not contain <text> (can repeat)

set -uo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$root"

impls="${1:-both}"
[[ "$impls" == "both" ]] && impls="node go"

binary="build/jsetimos$(go env GOEXE)"
if [[ " $impls " == *" go "* ]]; then
    ./build_go.sh > /dev/null || exit 1
fi

passed=0
failed=0

for impl in $impls; do
    if [[ "$impl" == "node" ]]; then
        cmd=(node build/jsetimos.js)
    else
        cmd=("$binary")
    fi

    for test in tests/bugs/bug*.jk; do
        args="$(sed -n 's|^// args: ||p' "$test")"
        mode="$(sed -n 's|^// mode: ||p' "$test")"
        mapfile -t expects < <(sed -n 's|^// expect-output: ||p' "$test" | tr -d '\r')
        mapfile -t files < <(sed -n 's|^// expect-file: ||p' "$test" | tr -d '\r')
        mapfile -t rejects < <(sed -n 's|^// reject-output: ||p' "$test" | tr -d '\r')

        for file in "${files[@]}"; do rm -f "$file"; done

        if [[ "$mode" == "shell" ]]; then
            output="$("${cmd[@]}" < "$test" 2>&1)"
        else
            # shellcheck disable=SC2086
            output="$("${cmd[@]}" "$test" $args 2>&1)"
        fi
        status=$?
        output="$(printf '%s' "$output" | grep -v "source-map-support")"

        problems=()
        [[ $status -ne 0 ]] && problems+=("exit status $status")
        if [[ ${#expects[@]} -eq 0 && -n "$output" ]]; then
            problems+=("unexpected output")
        fi
        for expect in "${expects[@]}"; do
            [[ "$output" != *"$expect"* ]] && problems+=("missing output: $expect")
        done
        for reject in "${rejects[@]}"; do
            [[ "$output" == *"$reject"* ]] && problems+=("unwanted output: $reject")
        done
        for file in "${files[@]}"; do
            [[ -s "$file" ]] || problems+=("file not written: $file")
            rm -f "$file"
        done

        if [[ ${#problems[@]} -eq 0 ]]; then
            passed=$((passed + 1))
            echo "PASS [$impl] $test"
        else
            failed=$((failed + 1))
            echo "FAIL [$impl] $test"
            for problem in "${problems[@]}"; do echo "    - $problem"; done
            [[ -n "$output" ]] && printf '%s\n' "$output" | head -n 8 | sed 's/^/    | /'
        fi
    done
done

echo
echo "$passed passed, $failed failed"
[[ $failed -eq 0 ]]
