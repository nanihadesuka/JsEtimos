# Bugs audit

Bugs found in the original TypeScript interpreter (`jsetimos.ts`) while porting it to Go (`src/go/`).

**All of them are now fixed in both implementations.** Each bug has a regression test in `tests/bugs/`
that asserts the correct behaviour. The tests failed before the fixes and pass now:

```
tests/bugs/run.sh          # both implementations
tests/bugs/run.sh node     # jsetimos.js only
tests/bugs/run.sh go       # Go only
```

Entries marked **Verified** were first reproduced by running the original with node.
Line numbers (`jsetimos.ts:NNN`) refer to the source before the fixes (commit `554eafc`).

## Summary

| # | Bug | Severity | Test |
|---|-----|----------|------|
| 1 | `return` caches the first value it returns | High | `bug01_return_caches_value.jk` |
| 2 | `printToFile` never writes anything | High | `bug02_print_to_file.jk` |
| 3 | `--dumpAST` / `--dumpFile` never write anything | Medium | `bug03a_dump_ast_default_path.jk`, `bug03b_dump_ast_custom_path.jk` |
| 4 | Printing a list or dict containing `undefined`/`null` crashes | Medium | `bug04_print_undefined_in_collections.jk` |
| 5 | `?isNaN` is always `false`, `?isFinite` is `true` for NaN | Medium | `bug05_isnan_isfinite.jk` |
| 6 | Compound assignment updates every scope with that name | Medium | `bug06_compound_assignment_shadowing.jk` |
| 7 | `for` loop condition must be loosely `== true` | Low | `bug07_for_condition_truthiness.jk` |
| 8 | `it10`+ parameters are sorted as text | Low | `bug08_it_parameters_order.jk` |
| 9 | `?group` crashes on keys like `toString` | Low | `bug09_group_builtin_keys.jk` |
| 10 | Dictionary access exposes JavaScript internals | Low | `bug10_dict_js_internals.jk` |
| 11 | Modules report errors as type `'null'` | Cosmetic | `bug11_module_error_type.jk` |
| 12 | Function values can't be passed through pipes as plain values | Low | `bug12_function_values_in_pipes.jk` |
| 13 | A runtime error inside `do { }` leaves an extra scope pushed | Low | `bug13_shell_scope_after_error.jk` |
| 14 | Dead or disabled code | Cosmetic | — |
| 15 | Error locations repeat the directory of files outside the root | Low | `bug15_error_location_path.jk` |

## Details

### 1. `return` caches the first value it returns — High, Verified

`visit_ReturnOp` (`jsetimos.ts:2440`) overwrote the syntax tree node with the value it computed:

```ts
node.node = this.visit(node.node)
throw node
```

The next time the same `return` ran, it "evaluated" the stored value instead of the original expression, so every later call returned the first result:

```
fun f(a) { return a*2 }
print(f(1))   // 2
print(f(5))   // 2  (expected 10)
```

**Fix:** a return statement now throws a separate `ReturnValue` object carrying the value; the AST is never modified. Go: `ReturnValue` in `interpreter.go`.

### 2. `printToFile` never writes anything — High, Verified

In `ExtSystem_NodeJs.writeFile` (`jsetimos.ts:59`) the default directory was a no-op comparison instead of an assignment:

```ts
baseDir == baseDir ?? this.dirName()   // should be: baseDir = baseDir ?? this.dirName()
```

`baseDir` stayed `undefined`, so the path became `undefined/<file>`. The write failed, and the empty callback swallowed the error. The same typo was in the Pyrogenesis and Browser adapters (`jsetimos.ts:158`, `jsetimos.ts:240`).

**Fix:** assignment in all three adapters. The Go port never had this bug.

### 3. `--dumpAST` / `--dumpFile` never write anything — Medium, Verified

- `writeFile` had the bug above.
- The default dump path was already absolute (`filePath + ".ast.yml"`), but `writeFile` prefixes a directory.
- `programArgs` (`jsetimos.ts:104`) stored flags as `attributes[name] = name`, so there was no way to give `--dumpFile` a path.

**Fix:** arguments accept `--name=value` (`--dumpFile=path/out.yml`). Dump paths are relative to the base directory like every other written file, and the default is `<file>.ast.yml`. The node dump is written without terminal colours. Go writes a JSON dump.

### 4. Printing a list or dict containing `undefined`/`null` crashes — Medium, Verified

`List.toStringIdent` and `Dictionary.toStringIdent` (`jsetimos.ts:863`, `jsetimos.ts:910`) called `v.toString()` on every element, which throws a JavaScript `TypeError` for `undefined`/`null`:

```
l = 1,undefined
print(l)   // TypeError: Cannot read properties of undefined (reading 'toString')
```

**Fix:** `String(v)`, which prints `1, undefined`.

### 5. `?isNaN` is always `false`, `?isFinite` is `true` for NaN — Medium, Verified

`jsetimos.ts:2698-2699` compared against `NaN`, which is never equal to anything:

```ts
case "isFinite": return object !== Infinity && object !== -Infinity && object !== NaN
case "isNaN": return object === NaN
```

**Fix:** `Number.isFinite(object)` and `Number.isNaN(object)`.

### 6. Compound assignment updates every scope with that name — Medium, Verified

`Stack.setOnExisting` (`jsetimos.ts:1083`) kept looping after the first match, so `x += 1` wrote into every scope that had an `x`, overwriting shadowed outer variables:

```
x = 10
fun shadow() {
    x = 1
    x += 5
}
shadow()
print(x)   // 6 (expected 10)
```

**Fix:** stop at the nearest scope that holds the name.

### 7. `for` loop condition must be loosely `== true` — Low, Verified

`visit_ForLoopOp` (`jsetimos.ts:3171`) looped `while (this.visit(node.runCondition) == true)`. Because of JavaScript's loose equality, a condition such as `3` or `"abc"` stopped the loop, unlike `if`, which uses normal truthiness.

**Fix:** `while (this.visit(node.runCondition))`. `for i = 3; i; i-- ...` now runs 3 times.

### 8. `it10`+ parameters are sorted as text — Low, Verified

`anonymousFunction` (`jsetimos.ts:1280`) sorted the `it` indexes with the default `.sort()`, which compares numbers as strings (`[0, 10, 2]`). A lambda using `it2` and `it10` got 3 parameters, and `it10` was undefined at runtime.

**Fix:** numeric sort (`.sort((a, b) => a - b)`, Go: `sort.Float64s`).

### 9. `?group` crashes on keys like `toString` — Low, Verified

`?group` (`jsetimos.ts:2607`) checked `key in dict.nodes`, which also sees built-in object properties, so `"toString"` was treated as an existing group and `.push` was called on a function:

```
(("toString",1),("a",2))?group   // TypeError: Cannot read properties of undefined (reading 'push')
```

**Fix:** `Object.prototype.hasOwnProperty.call(dict.nodes, key)`. The Go port never had this bug.

### 10. Dictionary access exposes JavaScript internals — Low, Verified

Key access (`jsetimos.ts:2759`) read any property of the underlying JavaScript object:

```
d = {a=1}
print(d.constructor)   // function Object() { [native code] }
```

**Fix:** only the dictionary's own keys are returned; anything else is `undefined`. A key the user defines (e.g. `{ constructor = 1 }`) still works. The Go port never had this bug.

### 11. Modules report errors as type `'null'` — Cosmetic, Verified

The module branch of `visit_AccessMethodOp` (`jsetimos.ts:2747`) copied the `null` branch's message:

```
import stdlib.math
math?toString   // Method 'toString' not found in object type 'null': [object Object]
```

**Fix:** `Method 'toString' not found in object type 'module': <module file path>`.

### 12. Function values can't be passed through pipes as plain values — Low, Verified

`executeFunction` always evaluated its arguments again (`visit_List`). That's correct for a call written in the source (the arguments are AST), but pipes pass values that are already evaluated. A function value in a pipe got "visited":

- a native function (`print`) raised `Trying to visit not defined node type: NativeFunction`
- an anonymous function was executed instead of passed
- a named function re-declared itself and became `undefined`

```
fun double(v) { v * 2 }
fun apply(fn, value) { fn(value) }
(double, 3) | apply   // Object is not an function ... Got 'undefined' instead.
```

**Fix:** `executeFunction` takes an `argsEvaluated` flag; pipes, `map` and `filter` pass their values without evaluating them again.
Side effect: a dictionary piped into a function is now passed as-is instead of being copied. Nothing in the language mutates dictionaries, so this isn't observable.

### 13. A runtime error inside `do { }` leaves an extra scope pushed — Low, Verified

`scopedReturn` (`jsetimos.ts:3258`) only popped the scope when the block finished normally or returned. In shell mode, which keeps running after errors, variables assigned before the error stayed visible:

```
>>> secret = 0
>>> do { for i = 0; i < 1; undefined?fail secret = 5 }
(runtime error)
>>> print(secret)
5   (expected 0)
```

**Fix:** pop the scope in a `finally` block (Go: `defer`).

### 14. Dead or disabled code — Cosmetic

Removed:
- the unreachable second copy of the `PLUS`/`MINUS` branch in `factor()` (`jsetimos.ts:1617`)
- the never-called `List.sumScalar`, `multList`, ... methods (`jsetimos.ts:872-889`)
- the no-op statements `list.length == range + 1` in `visit_RangeOp` (`jsetimos.ts:3190`)

Kept on purpose, because removing them would change how the lexer splits symbols:
- the reserved scalar operator tokens (`.+`, `.-`, `.*`, `./`)
- the disabled list spread (`...`, `SpreadListOp`, marked `// Buggy don't` in `listExpression`)

`ProgramArgs.shellMode` is still parsed but unused.

### 15. Error locations repeat the directory of files outside the root — Low, Verified

Found while writing the regression tests. `fileMode` built the lexer's directory from the relative folder of the file, then passed the full relative file name again, so locations repeated the folder:

```
node jsetimos.js tests/bugs/some_test.jk
FAILED ASSERT: false
(C:\...\JsEtimos/tests/bugs/tests/bugs/some_test.jk:5:14)
```

Files in the root got a double slash instead (`JsEtimos//playground.jk`).

**Fix:** the lexer gets the directory and the bare file name, without a trailing slash for files in the root. Locations are now `JsEtimos/tests/bugs/some_test.jk:5:14` and `JsEtimos/playground.jk`.

## Not bugs, but worth knowing

- **Floating point results depend on the runtime.** On Windows, node's `**` uses the C runtime `pow`, which isn't always correctly rounded. `sin`/`cos`/`tan` also differ from Go's `math` package in the last bit. The unit tests compare those values with an epsilon.
- **`;` is not a statement separator.** `fun f() { a = 1; b = 2 }` is a parse error. `;` only works inside dictionaries, after `return`, and in `for` headers.
- **Rebuilding `jsetimos.js`** needs TypeScript 3.8.3 and the Node type definitions, and must use `tsc -p .`. The source has 9 type errors (ES2019 APIs such as `flatMap` and `Object.fromEntries` on an ES6 target), so `tsc -b` refuses to emit. See `INSTRUCTIONS`.
