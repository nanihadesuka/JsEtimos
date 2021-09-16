nodemon -e jk \
    --exec "node jsetimos.js playground.jk --dumpAST" \
    --watch jsetimos.js \
    --watch stdlib \
    --watch playground.jk
