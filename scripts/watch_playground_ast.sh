#!/usr/bin/env bash
# Like watch_playground.sh, also writing the syntax tree to playground.jk.ast.yml.
# Needs nodemon (npm install -g nodemon).
#
# Usage: scripts/watch_playground_ast.sh

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

nodemon -e jk \
    --exec "node build/jsetimos.js playground.jk --dumpAST" \
    --watch build/jsetimos.js \
    --watch stdlib \
    --watch playground.jk
