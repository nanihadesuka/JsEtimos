# Bugs audit

Bugs found in the original TypeScript interpreter (`jsetimos.ts`) while porting it to Go (`jsetimos.go`).

Every entry marked **Verified** was reproduced by running the original with node (`node jsetimos.js ...`).
The **Go port** line says whether `jsetimos.go` keeps the bug (so both behave the same) or differs.

## Summary

| # | Bug | Severity | Go port |
|---|-----|----------|---------|
| 1 | `return` caches the first value it returns | High | Kept |
| 2 | `printToFile` never writes anything | High | Fixed |
| 3 | `--dumpAST` / `--dumpFile` never write anything | Medium | Fixed (writes JSON) |
| 4 | Printing a list or dict containing `undefined`/`null` crashes | Medium | Kept |
| 5 | `?isNaN` is always `false`, `?isFinite` is `true` for NaN | Medium | Kept |
| 6 | Compound assignment updates every scope with that name | Medium | Kept |
| 7 | `for` loop condition must be loosely `== true` | Low | Kept |
| 8 | `it10`+ parameters are sorted as text | Low | Kept |
| 9 | `?group` crashes on keys like `toString` | Low | Differs (no crash) |
| 10 | Dictionary access exposes JavaScript internals | Low | Differs (returns `undefined`) |
| 11 | Modules report errors as type `'null'` | Cosmetic | Kept |
| 12 | Function values can't be passed through pipes as plain values | Low | Kept |
| 13 | A runtime error inside `do { }` leaves an extra scope pushed | Low | Kept |
| 14 | Dead or disabled code | Cosmetic | Mostly dropped |

## Details

### 1. `return` caches the first value it returns — High, Verified

`visit_ReturnOp` (`jsetimos.ts:2440`) overwrites the syntax tree node with the value it computed:

```ts
node.node = this.visit(node.node)
throw node
```

The next time the same `return` runs, it "evaluates" the stored value instead of the original expression, so every later call returns the first result.

```
fun f(a) { return a*2 }
print(f(1))   // 2
print(f(5))   // 2  (expected 10)
```

Functions that don't use `return` (the last expression is the result) work correctly.
**Fix:** keep the evaluated value in a separate object that is thrown, for example `throw new ReturnValue(this.visit(node.node))`.
**Go port:** kept for parity (`visitReturnOp`).

### 2. `printToFile` never writes anything — High, Verified

In `ExtSystem_NodeJs.writeFile` (`jsetimos.ts:59`) the default directory is a no-op comparison instead of an assignment:

```ts
baseDir == baseDir ?? this.dirName()   // should be: baseDir = baseDir ?? this.dirName()
```

`baseDir` stays `undefined`, so the file path becomes `undefined/<file>`. The write fails, and the error is swallowed by the empty callback `() => { }`. The same typo is in the Pyrogenesis and Browser adapters (`jsetimos.ts:158`, `jsetimos.ts:240`).
`printToFile("data", "out.txt")` creates no file.
**Go port:** fixed; writes to `<working dir>/<file>`.

### 3. `--dumpAST` / `--dumpFile` never write anything — Medium, Verified

- `writeFile` has the bug above.
- The default dump path is already absolute (`filePath + ".ast.yml"`) but `writeFile` prefixes it with a directory again.
- `programArgs` (`jsetimos.ts:104`) stores flags as `attributes[name] = name`, so `--dumpFile` sets the dump file to the literal string `"dumpFile"`; there is no way to pass a path.

`node jsetimos.js playground.jk --dumpAST` creates no file.
**Go port:** `--dumpAST` writes a JSON dump to `<file>.ast.yml`. The `--dumpFile` parsing quirk is kept.

### 4. Printing a list or dict containing `undefined`/`null` crashes — Medium, Verified

`List.toStringIdent` and `Dictionary.toStringIdent` (`jsetimos.ts:863`, `jsetimos.ts:910`) call `v.toString()` on every element, which throws a JavaScript `TypeError` for `undefined`/`null`. That isn't a language error, so the interpreter crashes with a node stack trace.

```
l = 1,undefined
print(l)   // TypeError: Cannot read properties of undefined (reading 'toString')
```

**Fix:** use `String(v)` instead of `v.toString()`.
**Go port:** kept (exits with `Uncaught TypeError`).

### 5. `?isNaN` is always `false`, `?isFinite` is `true` for NaN — Medium, Verified

`jsetimos.ts:2698-2699` compare against `NaN`, which is never equal to anything:

```ts
case "isFinite": return object !== Infinity && object !== -Infinity && object !== NaN
case "isNaN": return object === NaN
```

`(0/0)?isNaN` gives `false` and `(0/0)?isFinite` gives `true`.
**Fix:** use `Number.isNaN(object)` and `Number.isFinite(object)`.
**Go port:** kept.

### 6. Compound assignment updates every scope with that name — Medium, by inspection

`Stack.setOnExisting` (`jsetimos.ts:1083`) keeps looping after the first match, so `x += 1` writes the new value into every scope that has an `x`, not only the nearest one. A shadowed outer variable gets overwritten.
**Fix:** `return` after the first `set`.
**Go port:** kept.

### 7. `for` loop condition must be loosely `== true` — Low, by inspection

`visit_ForLoopOp` (`jsetimos.ts:3171`) loops `while (this.visit(node.runCondition) == true)`. Because of JavaScript's loose equality, only `true` or values equal to `1` keep the loop going. A condition that evaluates to `2`, a non-empty string or a list stops it, unlike `if`, which uses normal truthiness.
**Fix:** `while (this.visit(node.runCondition))`.
**Go port:** kept.

### 8. `it10`+ parameters are sorted as text — Low, by inspection

`anonymousFunction` (`jsetimos.ts:1280`) sorts the `it` indexes with the default `.sort()`, which compares them as strings (`[0, 10, 2]`). A lambda that uses `it10` and `it2` gets 3 parameters instead of 11, so `it10` is never bound.
**Fix:** `.sort((a, b) => a - b)`.
**Go port:** kept.

### 9. `?group` crashes on keys like `toString` — Low, Verified

`?group` (`jsetimos.ts:2607`) checks `key in dict.nodes`. The `in` operator also sees the object's built-in properties, so a key such as `"toString"` or `"constructor"` is treated as already present, and `.push` is called on a function.

```
(("toString",1),("a",2))?group   // TypeError: Cannot read properties of undefined (reading 'push')
```

**Fix:** use `Object.prototype.hasOwnProperty.call(dict.nodes, key)` or a `Map`.
**Go port:** differs; these keys work like any other key.

### 10. Dictionary access exposes JavaScript internals — Low, Verified

Dictionaries are plain JavaScript objects, so key access (`jsetimos.ts:2759`) also finds built-in properties:

```
d = {a=1}
print(d.constructor)   // function Object() { [native code] }
```

**Fix:** create dictionaries with `Object.create(null)` or use a `Map`.
**Go port:** differs; unknown keys return `undefined`.

### 11. Modules report errors as type `'null'` — Cosmetic, Verified

The module branch of `visit_AccessMethodOp` (`jsetimos.ts:2747`) copies the `null` branch's error message:

```
import stdlib.math
math?toString   // Method 'toString' not found in object type 'null': [object Object]
```

**Go port:** kept.

### 12. Function values can't be passed through pipes as plain values — Low, Verified

Pipe and call arguments are evaluated again by `executeFunction` (`visit_List`). A function value inside the argument list gets "visited":

- a native function (`print`) raises `Trying to visit not defined node type: NativeFunction`
- an anonymous function is executed instead of being passed
- a named function re-declares itself and becomes `undefined`

```
x = print
y = x | { it }   // Runtime error: Trying to visit not defined node type: NativeFunction ...
```

**Go port:** kept.

### 13. A runtime error inside `do { }` leaves an extra scope pushed — Low, by inspection

`scopedReturn` (`jsetimos.ts:3258`) pops the scope only when the block finishes normally or returns; other errors skip `this.stack.pop()`. In shell mode the interpreter keeps running after the error, with one extra scope still on the stack. I couldn't produce a visible symptom in a quick test.
**Fix:** wrap the call in `try { ... } finally { this.stack.pop() }`.
**Go port:** kept.

### 14. Dead or disabled code — Cosmetic

- `factor()` checks the `PLUS`/`MINUS` branch twice (`jsetimos.ts:1610` and `jsetimos.ts:1617`); the second copy is unreachable.
- The scalar operators (`.+`, `.-`, `.*`, `./`) are lexed and reserved, but the parser never uses them; the matching `List.sumScalar`, `multList`, ... methods (`jsetimos.ts:872-889`) are never called.
- List spreading (`...`, `SpreadListOp`) is disabled in `listExpression` with the comment `// Buggy don't`.
- `visit_RangeOp` has a no-op statement `list.length == range + 1` (`jsetimos.ts:3190`).
- `ProgramArgs.shellMode` is parsed but never used.

**Go port:** the unused scalar list methods are not ported; the rest is mirrored so the grammar behaves the same.

## Not bugs, but worth knowing

- **Floating point results depend on the runtime.** On Windows, node's `**` uses the C runtime `pow`, which isn't always correctly rounded. `sin`/`cos`/`tan` also differ from Go's `math` package in the last bit. The unit tests now compare those values with an epsilon.
- **`;` is not a statement separator.** `fun f() { a = 1; b = 2 }` is a parse error. `;` only works inside dictionaries, after `return`, and in `for` headers.
