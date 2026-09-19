#!/usr/bin/env bash
# Reruns playground.jk with node whenever it, the stdlib or build/jsetimos.js
# changes. Needs nodemon (npm install -g nodemon).
#
# Usage: scripts/watch_playground.sh

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

nodemon -e jk \
    --exec "node build/jsetimos.js playground.jk" \
    --watch build/jsetimos.js \
    --watch stdlib \
    --watch playground.jk
