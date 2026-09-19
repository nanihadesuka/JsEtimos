nodemon -e jk \
    --exec "node build/jsetimos.js playground.jk" \
    --watch build/jsetimos.js \
    --watch stdlib \
    --watch playground.jk
