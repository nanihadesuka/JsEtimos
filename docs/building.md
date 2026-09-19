# Building and running

JsEtimos has two implementations of the same language, which are kept behaving identically:

| Implementation | Source | Build output | Used by |
|---|---|---|---|
| TypeScript | `src/typescript/jsetimos.ts` | `build/jsetimos.js` | node, the browser playground (`index.html`), 0 A.D. |
| Go | `src/go/` | `build/jsetimos` (`build/jsetimos.exe` on Windows) | the command line |

Run every command from the repository root: program files, imports and written files are resolved relative to the directory you run the interpreter from.

## Requirements

- **Node.js** 14.6 or newer, to run `build/jsetimos.js`.
- **npm**, only to rebuild `build/jsetimos.js`. `build_ts.sh` installs the pinned TypeScript 3.8.3 by itself (newer versions change the generated code).
- **Go** 1.21 or newer, only for the Go implementation.
- **bash** for the scripts (on Windows: Git Bash).

## Trying it

Open `index.html` in a browser, or use the online version at <https://nanihadesuka.github.io/JsEtimos/>.

## Running programs

Both implementations take the same arguments:

| Mode | Node | Go |
|---|---|---|
| Run a file | `node build/jsetimos.js file.jk` | `build/jsetimos file.jk` |
| Shell (REPL) | `node build/jsetimos.js` | `build/jsetimos` |
| Run code from the command line | `node build/jsetimos.js --inputMode "print(1)"` | `build/jsetimos --inputMode "print(1)"` |

Extra options:

- `--dumpAST` writes the parsed syntax tree to `<file>.ast.yml`.
- `--dumpFile=path` changes where `--dumpAST` writes it.

JsEtimos source files use the `.jk` extension.

## Imports

`import a.b` loads `a/b.jk`. It is looked up next to the importing file first, then from the base directory (where the interpreter was started). That's why a test in `tests/` can `import stdlib.math`.

See [language.md](language.md#imports) for the import syntax.

## Building

### TypeScript

After changing `src/typescript/jsetimos.ts`:

```
./build_ts.sh           # builds build/jsetimos.js
./build_ts.sh --test    # builds, then runs the tests on node
```

`build/jsetimos.js` is committed, because `index.html` (and GitHub Pages) loads it. Everything else in `build/` is ignored by git.

The source has known type errors (ES2019 APIs such as `flatMap` on an ES6 target). The TypeScript compiler still writes the output; `build_ts.sh` only fails if no output is written.

### Go

```
./build_go.sh           # builds build/jsetimos (build/jsetimos.exe on Windows)
./build_go.sh --test    # builds, then runs the tests on the Go binary
```

## Tests

| Tests | Command |
|---|---|
| Language coverage (`tests/lang_basic_coverage.jk`) | `node build/jsetimos.js tests/lang_basic_coverage.jk` or `build/jsetimos tests/lang_basic_coverage.jk` |
| Bug regression tests, one file per fixed bug (`tests/bugs/`) | `tests/bugs/run.sh [node\|go\|both]` |

A passing test file prints nothing. A failing `assert` prints `FAILED ASSERT` with the expression and its location.

`./build_ts.sh --test` and `./build_go.sh --test` run both test sets for their implementation.

## Development helpers

- `run.sh` reruns `playground.jk` on every change (needs [nodemon](https://nodemon.io/)).
- `run_db.sh` does the same with `--dumpAST`.

## Editor support

There is no dedicated syntax highlighter. The Kotlin highlighter of most editors works well enough for `.jk` files.
