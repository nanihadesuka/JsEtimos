# JsEtimos language reference

JsEtimos is a small dynamic language mixing ideas from Python (scoping, imports, little boilerplate), JavaScript (values and type conversion) and Kotlin (lambdas, `if` expressions, built-in methods), plus pipes and `?` methods.

This page describes the language as the interpreter implements it. Both implementations (TypeScript and Go) behave the same. See [building.md](building.md) to run programs.

Examples show results in comments. The code blocks are marked as Kotlin, which is close enough for highlighting.

```kotlin
import stdlib.math

fun square(x) { x * x }

values = 1:5 | map square | filter { it > 5 }
print("squares above 5: ${values}, sum ${values | math.sum}")
// squares above 5: 9, 16, 25, sum 50
```

- [Basics](#basics)
- [Values and types](#values-and-types)
- [Operators](#operators)
- [Strings](#strings)
- [Lists](#lists)
- [Ranges](#ranges)
- [Dictionaries](#dictionaries)
- [Control flow](#control-flow)
- [Functions](#functions)
- [Anonymous functions](#anonymous-functions)
- [Pipes](#pipes)
- [Built-in functions](#built-in-functions)
- [Built-in methods](#built-in-methods)
- [Imports](#imports)
- [Limitations and gotchas](#limitations-and-gotchas)

## Basics

### Statements and lines

There is one statement per line. `;` is **not** a statement separator; it's only used inside dictionaries, after `return`, and in `for` headers.

An expression continues on the next line when the line ends with an operator, `,`, `|`, `(` or `[`:

```kotlin
total = 1 +
    2        // 3

list = 1, 2,
    3        // 1, 2, 3
```

A line that *starts* with an operator is a new statement, so put the operator at the end of the line.

### Comments

```kotlin
// line comment
/* block
   comment */
```

### Names

Names start with a letter, followed by letters, digits or `_` (`total`, `row2`, `max_value`).

These words are reserved: `fun`, `return`, `if`, `elif`, `else`, `true`, `false`, `and`, `not`, `or`, `do`, `for`, `as`, `map`, `filter`, `import`, and `it`, `it1`, `it2`, ... (see [anonymous functions](#anonymous-functions)). `undefined`, `null` and `arguments` have special meanings.

### Variables

There is no declaration keyword; assigning creates the variable. Several variables can be assigned at once from a list:

```kotlin
a = 1
b, c = 2, 3
b, c = c, b      // swap: b is 3, c is 2
x, y = (10, 20)
```

A name must be assigned before the line that uses it, otherwise the program doesn't start (`Trying to use undefined variable`).

## Values and types

| Type | Examples | `?is` |
|---|---|---|
| number | `3`, `2.5`, `-1` | `"number"` |
| string | `"hello"`, `"total: ${n}"` | `"string"` |
| bool | `true`, `false` | `"bool"` |
| undefined | `undefined` | `"undefined"` |
| null | `null` | `"null"` |
| list | `1, 2, 3` | `"list"` |
| dictionary | `{ a = 1; b = 2 }` | `"dict"` |
| function | named functions, lambdas | — |
| module | an imported file | `"dict"` |

Numbers are 64-bit floating point. Number literals are digits with an optional decimal part (`12`, `0.5`); there's no exponent notation (`1e5`) and no leading dot (`.5`). Write negative numbers with a minus sign.

`undefined` is what you get for missing things: a list index out of range, a missing dictionary key, a missing function argument.

### Truthiness

`if`, `for`, `and`, `or` and `not` treat `false`, `0`, `""`, `undefined`, `null` and NaN as false. Everything else is true, including empty lists and dictionaries.

## Operators

### Arithmetic

| Operator | Meaning | Example |
|---|---|---|
| `+` `-` `*` `/` | add, subtract, multiply, divide | `10 / 4` is `2.5` |
| `%` | remainder (keeps the sign of the left side) | `-7 % 3` is `-1` |
| `**` | power | `2 ** 10` is `1024` |
| `-x`, `+x` | negate, identity | |

Division by zero gives `Infinity` (or NaN for `0/0`), not an error.

`**` is evaluated left to right: `2 ** 3 ** 2` is `(2 ** 3) ** 2`, which is `64`.

### Increment and compound assignment

```kotlin
n = 5
n++          // value 5, n becomes 6
++n          // value 7, n becomes 7
n--          // value 7, n becomes 6
n += 4       // 10   (also -=, *=, /=, **=)
```

`+=` also appends to lists (`list += 4`, `list += (5, 6)`), `*=` repeats a list, and `+=` concatenates strings.

### Comparison

`==`, `!=`, `<`, `>`, `<=`, `>=`. Comparisons can be chained: `1 < x < 10` means `1 < x and x < 10`.

Strings compare alphabetically (`"a" < "b"`). Lists compare element by element with `==` and `!=`:

```kotlin
(1, 2) == (1, 2)      // true
(1, 2) != (1, 2, 3)   // true
```

Lists *inside* lists are compared by identity, so `((1, 2),) == ((1, 2),)` is `false`.

### Logic

`and`, `or` and `not`. `and`/`or` only evaluate the right side when needed, and always return `true` or `false`:

```kotlin
0 or "x"      // true
1 and 2       // true
not ""        // true
```

### Type conversion

Mixed types convert like in JavaScript:

| Expression | Result | Why |
|---|---|---|
| `"5" + 1` | `"51"` | `+` with a string joins text |
| `1 + 2 + "a"` | `"3a"` | evaluated left to right |
| `"10" - 4` | `6` | `-`, `/`, `%`, `**` convert to numbers |
| `"5" * 2` | `"55"` | `*` with a string repeats it |
| `true + 1` | `2` | `true` is 1, `false` is 0 |
| `null + 1` | `1` | `null` is 0 |
| `undefined + 1` | NaN | |
| `"5" == 5` | `true` | `==` converts before comparing |
| `null == undefined` | `true` | |
| `"10" < "9"` | `true` | both strings: compared as text |
| `10 > "9"` | `true` | a number is involved: compared as numbers |

### Precedence

From loosest to tightest:

| Level | Operators |
|---|---|
| 1 | `\|` (pipes) |
| 2 | `,` (building lists) |
| 3 | `or` |
| 4 | `and` |
| 5 | `==` `!=` `<` `>` `<=` `>=` |
| 6 | `+` `-` |
| 7 | `*` `/` `%` |
| 8 | `**` |
| 9 | `.key`, `[index]`, `(arguments)`, `?method` |
| 10 | `a:b` ranges |
| 11 | unary `-` `+` `not`, `++` `--` |

So `1, 2 | f` pipes the whole list `1, 2` into `f`. Unary operators bind tightest: `-2 ** 2` is `(-2) ** 2`, which is `4`, and `-3.4?abs` is `3.4`. Range bounds are single values: write `1:(n + 1)`, not `1:n + 1`.

## Strings

Strings use double quotes and can span several lines. `\n` is the only escape sequence (a newline). A string can't contain `"`.

`${ }` inserts the value of any expression, including other strings:

```kotlin
name = "world"
"hello ${name}!"               // "hello world!"
"${ 2 * 21 } and ${ "a" * 3 }" // "42 and aaa"
```

Indexing counts from 0; negative indexes count from the end. Two values in the brackets give a slice (see [Lists](#lists) for the exact rules):

```kotlin
s = "hello"
s[0]        // "h"
s[-1]       // "o"
s[1, 3]     // "el"
s[1, -1]    // "ello"
s?size      // 5
```

`"ab" * 3` is `"ababab"`. See [string methods](#string).

## Lists

Commas build lists. Parentheses are only needed for grouping or nesting:

```kotlin
a = 1, 2, 3
b = (1, (2, 3), "x")
single = (5,)          // a list with one element
```

### Indexing and slices

```kotlin
list = 10, 20, 30, 40
list[0]          // 10
list[-1]         // 40
list[9]          // undefined
list[1, 3]       // 20, 30         end index excluded
list[1, -1]      // 20, 30, 40     -1 means "to the end"
list[0, -2]      // 10, 20, 30     other negative ends are included
grid = (1, 2), (3, 4)
grid[1][0]       // 3
```

### List operators

| Expression | Result |
|---|---|
| `(1, 2) + 3` | `1, 2, 3` |
| `0 + (1, 2)` | `0, 1, 2` |
| `(1, 2) + (3, 4)` | `1, 2, 3, 4` |
| `(1, 2) * 2` | `1, 2, 1, 2` |
| `-(1, 2)` | `-1, -2` |
| `not (1, 0)` | `false, true` |

Lists are values: methods and operators return new lists. See [list methods](#list).

## Ranges

`start:end` and `start:step:end` build a list of numbers. Both ends are included, the range counts down when `start > end`, and the sign of the step is ignored:

```kotlin
1:5          // 1, 2, 3, 4, 5
3:1          // 3, 2, 1
0:2:7        // 0, 2, 4, 6
0:0.25:1     // 0, 0.25, 0.5, 0.75, 1
```

## Dictionaries

Entries are `key = value`, separated by new lines or `;`. Keys are names or whole numbers:

```kotlin
box = {
    name = "box"
    size = 1, 2, 3
    10 = "ten"
    inner = { deep = true }
    area = fun { w, h -> w * h }
}

box.name            // "box"
box.10              // "ten"
box.inner.deep      // true
box.area(2, 3)      // 6
box.missing         // undefined
box?keys            // 10, name, size, inner, area
```

Number keys come first, in ascending order; the other keys keep their order. Keys are stored as strings (`box?keys` returns strings).

A function stored in a dictionary is written `fun { parameters -> body }`, like an [anonymous function](#anonymous-functions). Dictionaries can't be changed after they're created. See [dictionary methods](#dictionary).

## Control flow

### if / elif / else

As a statement:

```kotlin
if x > 10
    print("big")
elif x > 5 {
    print("medium")
    print("still medium")
}
else print("small")
```

Parentheses around the condition are optional. A branch is a single line or a `{ }` block.

`if` is also an expression. As a value it must have an `else`:

```kotlin
size = if x > 10 "big" else "small"
a, b = if flag 1, 2 else 3, 4
```

### for

`for setup; condition; step body` works like C's `for`, with optional parentheses:

```kotlin
for i = 0; i < 3; i++
    print(i)

for (j = 10; j > 0; { j -= 3 }) print(j)

for a, b = 0, 1; a < 50; { a, b = b, a + b } {
    print(a)
}
```

The step is an expression, or a `{ }` block for statements such as assignments (`j -= 3` on its own isn't an expression). The setup and step can be left empty; an empty condition counts as false, so the loop doesn't run. The condition uses [truthiness](#truthiness). Variables created in the setup still exist after the loop.

There's no `while`, `break` or `continue`.

### do blocks

`do { }` runs a block immediately, in its own scope, and gives the value of its last line (or of a `return` inside it):

```kotlin
result = do {
    tmp = 6
    tmp * 7
}                    // 42; tmp doesn't exist here
```

## Functions

```kotlin
fun area(w, h) {
    w * h
}

area(2, 3)     // 6
```

A function returns the value of its last line, or the value given to `return`:

```kotlin
fun sign(n) {
    if n < 0 return "negative"
    if n > 0 return "positive"
    "zero"
}
```

`return` on its own returns `undefined`. A function can return several values as a list: `return a, b`.

Calls are checked when the program starts: a named function must be called with exactly as many arguments as it has parameters. To accept any number of arguments, declare no parameters, read the special list `arguments`, and pass the values through a [pipe](#pipes):

```kotlin
fun count() { arguments?size }
(1, 2, 3) | count      // 3
```

### Scope

- Assigning a variable inside a function or block creates a local variable; a variable with the same name outside isn't changed.
- Reading a variable looks in the local scope first, then outwards.
- Compound assignment (`+=`, `++`, ...) updates the nearest existing variable, so it can change an outer variable.

```kotlin
counter = 0
fun increment() { counter += 1 }
increment()          // counter is now 1

fun shadow() {
    counter = 100    // a new local variable
}
shadow()             // counter is still 1
```

Functions can be nested and recursive:

```kotlin
fun fib(n) { if n < 2 n else fib(n - 1) + fib(n - 2) }
fib(10)              // 55
```

### Functions as values

A function's name without parentheses is the function itself. It can be stored, passed as an argument, and called later:

```kotlin
fun double(v) { v * 2 }
fun apply(fn, value) { fn(value) }

apply(double, 3)         // 6
(double, 3) | apply      // 6
```

## Anonymous functions

Anonymous functions (lambdas) are written in braces. They can appear in three places: after a pipe `|`, as a dictionary value (`fun { ... }`), and as `do { }` blocks (which take no parameters).

```kotlin
(1, 2) | { a, b -> a + b }     // 3: two parameters
5 | { it * 2 }                 // 10: one implicit parameter
```

Without an explicit parameter list, `it`, `it1`, `it2`, ... are the first, second, third, ... arguments:

```kotlin
(2, 3) | { it * it1 }        // 6
```

`arguments` holds all the arguments as a list, in both named and anonymous functions. `it` and `arguments` can't be used outside functions.

## Pipes

`value | f` calls `f` with `value`. If the value is a list, its elements become the arguments:

```kotlin
fun add(a, b) { a + b }

5 | { it * 2 }                   // 10
(2, 3) | add                     // 5
(2, 3) | add | { it * 10 }       // 50
```

Pipes accept named functions, lambdas, and functions in modules or dictionaries (`| math.sum`, `| box.area`). Extra arguments are ignored.

`map` and `filter` apply the function to each element:

```kotlin
1:5 | map { it * it }                    // 1, 4, 9, 16, 25
1:10 | filter { it % 3 == 0 }            // 3, 6, 9
((1, 2), (3, 4)) | map { a, b -> a * b } // 2, 12
```

When the elements are lists themselves, each one is spread into the arguments, as in the last example.

A `?method`, `.key` or `[index]` right after `|` applies to the value flowing through the pipe:

```kotlin
1:4 | map { it * 10 } | ?last      // 40
```

A pipe can continue on the next line after the `|`, including a lambda's opening `{`.

## Built-in functions

| Function | Description |
|---|---|
| `print(value)` | Prints one value. To print several, use a string with `${ }` or a list. |
| `printself(expression)` | Prints the expression's source text and its value: `printself(1 + 2)` prints `1 + 2 ⟶  3`. |
| `assert(condition)` | Prints `FAILED ASSERT` with the expression and its location when the condition isn't exactly `true`. Doesn't stop the program. |
| `printToFile(data, path)` | Writes `data` as text to `path`, relative to the base directory. |

## Built-in methods

Methods are called with `?` and take no arguments: `value?method`.

### All values

| Method | Result |
|---|---|
| `?is` | The type name (see [Values and types](#values-and-types)). Not available on functions. |
| `?toString` | The value as text, as `print` shows it. Not available on modules. |

### Number

| Method | Result |
|---|---|
| `?toBool` | `false` for 0 and NaN, `true` otherwise |
| `?toList` | `0, 1, ..., n - 1` (uses the whole part of the absolute value) |
| `?trunc`, `?floor`, `?ceil`, `?round` | Rounding: toward zero, down, up, to nearest (halves round up: `-2.5?round` is `-2`) |
| `?abs`, `?sign`, `?negative`, `?sqrt` | Absolute value, -1/0/1, negated value, square root |
| `?sin`, `?cos`, `?tan`, `?sinh`, `?cosh`, `?tanh` | Trigonometry, in radians |
| `?rad`, `?deg`, `?pi` | Degrees to radians, radians to degrees, the value times π |
| `?random` | Random number from 0 up to (not including) the value |
| `?randomInt` | Random whole number from 0 to the value (inclusive) |
| `?randomSign` | The value or its negative, at random |
| `?isNegative`, `?isPositive` | `< 0`, `>= 0` |
| `?isInfinite`, `?isFinite`, `?isNaN` | Special value checks |

### String

| Method | Result |
|---|---|
| `?size` | Number of characters |
| `?toNumber` | The number at the start of the text (`"3.5kg"?toNumber` is `3.5`), or NaN |
| `?toList` | List of characters |
| `?lines` | List of lines |
| `?toLowerCase`, `?toUpperCase`, `?capitalize` | Case changes; `capitalize` uppercases the first character |
| `?trim`, `?trimStart`, `?trimEnd` | Remove surrounding whitespace |
| `?first`, `?last` | First or last character (`undefined` if empty) |
| `?isEmpty`, `?isNotEmpty` | Compares with `""` |
| `?isBlank`, `?isNotBlank` | Like `isEmpty`, ignoring whitespace |
| `?hashCode` | Java-style 32-bit hash |

### Bool

| Method | Result |
|---|---|
| `?toNumber` | 1 or 0 |
| `?flip` | The opposite value |

### List

| Method | Result |
|---|---|
| `?size`, `?lastIndex` | Number of elements, index of the last one |
| `?first`, `?last` | First or last element (`undefined` if empty) |
| `?isEmpty`, `?isNotEmpty` | Checks the size |
| `?reverse` | Elements in reverse order |
| `?sort` | Sorted numerically, ascending. Lists inside are sorted by their first element. |
| `?flat` | Nested lists flattened one level: `(1, (2, 3))?flat` is `1, 2, 3` |
| `?dropFirst`, `?dropLast` | Without the first or last element |
| `?sum`, `?average`, `?max`, `?min` | Sum, mean, maximum, minimum |
| `?hypot` | Square root of the sum of squares |
| `?randomItem` | A random element |
| `?zip` | From a list of two equal-length lists, the list of pairs: `((1, 2), (3, 4))?zip` is `(1, 3), (2, 4)` |
| `?toDict` | From a list of `(key, value)` pairs, a dictionary |
| `?group` | From a list of `(key, value)` pairs, a dictionary with the values collected per key: `(("a", 1), ("b", 2), ("a", 3))?group` is `{ a = 1, 3; b = 2 }` |

### Dictionary

| Method | Result |
|---|---|
| `?size` | Number of keys |
| `?keys`, `?values` | Lists of keys (as strings) and values |
| `?toList` | List of `(key, value)` pairs |
| `?isEmpty`, `?isNotEmpty` | Checks the size |

### Module

| Method | Result |
|---|---|
| `?keys` | Names defined by the module, including the built-in functions |

## Imports

```kotlin
import stdlib.math
import stdlib.math as m

math.clamp(15, 0, 10)      // 10
m.pi                       // 3.141592653589793
(4, 9, 2) | math.max       // 9
```

`import a.b.c` loads the file `a/b/c.jk`, looked up next to the importing file first, then from the directory the interpreter was started in. The module is available under its last name (`c`) or under the name given with `as`.

Imports are only allowed at the top level of a file (not inside functions or blocks). A module is loaded once, even when several files import it. Everything a module defines at its top level is available as `module.name`.

### stdlib/math

`stdlib/math.jk` is the only standard module:

| Name | Description |
|---|---|
| `pi` | 3.141592653589793 |
| `sum`, `mul`, `average`, `min`, `max` | Over all arguments, usually through a pipe: `(1, 2, 3) \| math.sum` |
| `minMax` | The minimum and maximum of all arguments, as a list |
| `clamp(value, from, to)` | `value` limited to the range `from`...`to` |

## Limitations and gotchas

- One statement per line; `;` doesn't separate statements.
- No `while`, `break`, `continue`, classes, exceptions, or exponent number literals (`1e5`).
- Strings can't contain `"`, and `\n` is the only escape.
- Lambdas can only be written after `|`, as `fun { }` dictionary values, or as `do { }`. `f = { a -> a }` is an error.
- `print` takes exactly one value.
- A named function must be called with exactly its number of parameters. For a variable number of arguments, use `arguments` and pipes.
- `*` with a string repeats it: `"5" * 2` is `"55"`, not `10`.
- `**` is evaluated left to right (`2 ** 3 ** 2` is 64), and unary minus binds tighter than it (`-2 ** 2` is 4).
- A line starting with an operator is a new statement. End the previous line with the operator to continue an expression.
- Runtime errors say what went wrong but not where. Syntax errors show the line and position.
