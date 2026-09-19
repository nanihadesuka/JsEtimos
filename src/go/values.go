// JSEtimos - copyright nani 2021
//
// JavaScript value semantics the TypeScript implementation got from the JS runtime.

package main

import (
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
)

// ─────────────────────────────────────────────────────────────────────────────
// Errors
// ─────────────────────────────────────────────────────────────────────────────

// LangError covers LexerError, ParserError and RuntimeError.
type LangError struct {
	kind    string
	message string
}

func (e *LangError) Error() string { return e.message }

// JSError stands in for errors the JavaScript runtime itself would throw
// (TypeError, RangeError...). They are not language errors, so they crash
// the program just like an uncaught exception in node.
type JSError struct {
	name    string
	message string
}

func (e *JSError) Error() string { return e.name + ": " + e.message }

func throwTypeError(msg string)  { panic(&JSError{"TypeError", msg}) }
func throwRangeError(msg string) { panic(&JSError{"RangeError", msg}) }

// ─────────────────────────────────────────────────────────────────────────────
// JavaScript value semantics
// ─────────────────────────────────────────────────────────────────────────────

// Primitive values: float64 (number), string, bool, undefinedT, nullT.
// Everything else (lists, dictionaries, functions, modules) is an object.
type undefinedT struct{}
type nullT struct{}

var jsUndefined any = undefinedT{}
var jsNull any = nullT{}

func isUndefined(v any) bool { _, ok := v.(undefinedT); return ok }
func isNullish(v any) bool {
	switch v.(type) {
	case undefinedT, nullT, nil:
		return true
	}
	return false
}

// jsType is `typeof` except that null gets its own tag.
func jsType(v any) string {
	switch v.(type) {
	case float64:
		return "number"
	case string:
		return "string"
	case bool:
		return "boolean"
	case undefinedT:
		return "undefined"
	case nullT, nil:
		return "null"
	}
	return "object"
}

func jsTypeof(v any) string {
	t := jsType(v)
	if t == "null" {
		return "object"
	}
	return t
}

// Number.prototype.toString()
func numToString(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case f == 0:
		return "0"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case f < 0:
		return "-" + numToString(-f)
	}
	s := strconv.FormatFloat(f, 'e', -1, 64)
	mantissa, expText, _ := strings.Cut(s, "e")
	digits := strings.Replace(mantissa, ".", "", 1)
	exp, _ := strconv.Atoi(expText)
	k := len(digits)
	n := exp + 1

	switch {
	case k <= n && n <= 21:
		return digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		return digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		return "0." + strings.Repeat("0", -n) + digits
	}
	sign := "+"
	e := n - 1
	if e < 0 {
		sign = "-"
		e = -e
	}
	if k == 1 {
		return digits + "e" + sign + strconv.Itoa(e)
	}
	return digits[:1] + "." + digits[1:] + "e" + sign + strconv.Itoa(e)
}

func isJSSpace(r rune) bool {
	switch r {
	case 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x20, 0xa0, 0x1680, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000, 0xfeff:
		return true
	}
	return unicode.Is(unicode.Zs, r)
}

func jsTrim(s string) string      { return strings.TrimFunc(s, isJSSpace) }
func jsTrimStart(s string) string { return strings.TrimLeftFunc(s, isJSSpace) }
func jsTrimEnd(s string) string   { return strings.TrimRightFunc(s, isJSSpace) }

var strictDecimalRe = regexp.MustCompile(`^[+-]?(Infinity|(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?)$`)
var prefixDecimalRe = regexp.MustCompile(`^[+-]?(Infinity|(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?)`)

func parseDecimal(text string) float64 {
	if strings.HasSuffix(text, "Infinity") {
		if strings.HasPrefix(text, "-") {
			return math.Inf(-1)
		}
		return math.Inf(1)
	}
	f, _ := strconv.ParseFloat(text, 64)
	return f
}

// StringToNumber (used by Number(x), arithmetic and loose equality)
func strToNumber(s string) float64 {
	t := jsTrim(s)
	if t == "" {
		return 0
	}
	if len(t) > 2 && t[0] == '0' {
		base := 0
		switch t[1] {
		case 'x', 'X':
			base = 16
		case 'o', 'O':
			base = 8
		case 'b', 'B':
			base = 2
		}
		if base != 0 {
			value := 0.0
			for _, c := range t[2:] {
				d, err := strconv.ParseInt(string(c), base, 64)
				if err != nil {
					return math.NaN()
				}
				value = value*float64(base) + float64(d)
			}
			return value
		}
	}
	if !strictDecimalRe.MatchString(t) {
		return math.NaN()
	}
	return parseDecimal(t)
}

// Global parseFloat
func jsParseFloat(s string) float64 {
	m := prefixDecimalRe.FindString(jsTrimStart(s))
	if m == "" {
		return math.NaN()
	}
	return parseDecimal(m)
}

// JavaScript strings are UTF-16 sequences; length, indexing and slicing work on code units.
func u16(s string) []uint16     { return utf16.Encode([]rune(s)) }
func fromU16(u []uint16) string { return string(utf16.Decode(u)) }
func jsLen(s string) int        { return len(u16(s)) }
func compareU16(a, b string) int {
	ua, ub := u16(a), u16(b)
	for i := 0; i < len(ua) && i < len(ub); i++ {
		if ua[i] != ub[i] {
			if ua[i] < ub[i] {
				return -1
			}
			return 1
		}
	}
	return len(ua) - len(ub)
}

// ToPrimitive: objects convert through their toString() method.
func toPrimitive(v any) any {
	switch x := v.(type) {
	case float64, string, bool, undefinedT, nullT:
		return v
	case nil:
		return jsNull
	case *List:
		return x.toString()
	case *Dictionary:
		return x.toString()
	case *Token:
		return x.toString()
	case Node:
		return astToString(x)
	}
	return "[object Object]"
}

// ToString, as done by template literals and String(x)
func toStr(v any) string {
	switch x := toPrimitive(v).(type) {
	case float64:
		return numToString(x)
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case undefinedT:
		return "undefined"
	}
	return "null"
}

// Array.prototype.toString() of a raw JS array (not a List): joins with "," and
// turns undefined/null into empty strings
func arrayToString(values []any) string {
	parts := make([]string, len(values))
	for i, v := range values {
		if !isNullish(v) {
			parts[i] = toStr(v)
		}
	}
	return strings.Join(parts, ",")
}

// value.toString(): same as toStr but throws on undefined and null
func methodToString(v any) string {
	if isNullish(v) {
		name := "null"
		if isUndefined(v) {
			name = "undefined"
		}
		throwTypeError("Cannot read properties of " + name + " (reading 'toString')")
	}
	return toStr(v)
}

// ToNumber
func toNum(v any) float64 {
	switch x := toPrimitive(v).(type) {
	case float64:
		return x
	case string:
		return strToNumber(x)
	case bool:
		if x {
			return 1
		}
		return 0
	case undefinedT:
		return math.NaN()
	}
	return 0
}

// ToIntegerOrInfinity
func toIntegerOrInf(v any) float64 {
	n := toNum(v)
	if math.IsNaN(n) {
		return 0
	}
	return math.Trunc(n)
}

// ToBoolean
func truthy(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return !(x == 0 || math.IsNaN(x))
	case string:
		return x != ""
	case undefinedT, nullT, nil:
		return false
	}
	return true
}

func boolToNum(v any) any {
	if v.(bool) {
		return 1.0
	}
	return 0.0
}

// Loose equality (==)
func looseEq(a, b any) bool {
	ta, tb := jsType(a), jsType(b)
	if ta == tb {
		switch ta {
		case "undefined", "null":
			return true
		case "object":
			return a == b // identity
		}
		return a == b
	}
	switch {
	case (ta == "undefined" && tb == "null") || (ta == "null" && tb == "undefined"):
		return true
	case ta == "number" && tb == "string":
		return a.(float64) == strToNumber(b.(string))
	case ta == "string" && tb == "number":
		return strToNumber(a.(string)) == b.(float64)
	case ta == "boolean":
		return looseEq(boolToNum(a), b)
	case tb == "boolean":
		return looseEq(a, boolToNum(b))
	case (ta == "number" || ta == "string") && tb == "object":
		return looseEq(a, toPrimitive(b))
	case ta == "object" && (tb == "number" || tb == "string"):
		return looseEq(toPrimitive(a), b)
	}
	return false
}

// Abstract relational comparison a < b, second result is false when undefined (NaN involved)
func jsLess(a, b any) (bool, bool) {
	pa, pb := toPrimitive(a), toPrimitive(b)
	if sa, ok := pa.(string); ok {
		if sb, ok := pb.(string); ok {
			return compareU16(sa, sb) < 0, true
		}
	}
	na, nb := toNum(pa), toNum(pb)
	if math.IsNaN(na) || math.IsNaN(nb) {
		return false, false
	}
	return na < nb, true
}

func cmpLT(a, b any) bool { r, _ := jsLess(a, b); return r }
func cmpGT(a, b any) bool { r, _ := jsLess(b, a); return r }
func cmpLE(a, b any) bool { r, ok := jsLess(b, a); return ok && !r }
func cmpGE(a, b any) bool { r, ok := jsLess(a, b); return ok && !r }

func jsAdd(a, b any) any {
	pa, pb := toPrimitive(a), toPrimitive(b)
	_, sa := pa.(string)
	_, sb := pb.(string)
	if sa || sb {
		return toStr(pa) + toStr(pb)
	}
	return toNum(pa) + toNum(pb)
}

func jsSub(a, b any) any { return toNum(a) - toNum(b) }
func jsMul(a, b any) any { return toNum(a) * toNum(b) }
func jsDiv(a, b any) any { return toNum(a) / toNum(b) }
func jsMod(a, b any) any { return math.Mod(toNum(a), toNum(b)) }
func jsPow(a, b any) any { return mathPow(toNum(a), toNum(b)) }

// Math.pow (differs from Go for NaN exponents and ±1 bases with infinite exponents)
func mathPow(x, y float64) float64 {
	if math.IsNaN(y) {
		return math.NaN()
	}
	if y == 0 {
		return 1
	}
	if (x == 1 || x == -1) && math.IsInf(y, 0) {
		return math.NaN()
	}
	return math.Pow(x, y)
}

// Math.round
func mathRound(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) || x == 0 {
		return x
	}
	if x > 0 && x < 0.5 {
		return 0
	}
	if x < 0 && x >= -0.5 {
		return math.Copysign(0, -1)
	}
	r := math.Floor(x)
	if x-r >= 0.5 {
		r += 1
	}
	return r
}

// Math.sign
func mathSign(x float64) float64 {
	if math.IsNaN(x) || x == 0 {
		return x
	}
	if x > 0 {
		return 1
	}
	return -1
}

// new Array(v).length
func arrayLength(v any) int {
	n, ok := v.(float64)
	if !ok {
		return 1
	}
	if n < 0 || n != math.Trunc(n) || n > 4294967295 {
		throwRangeError("Invalid array length")
	}
	return int(n)
}

// String.prototype.repeat
func jsRepeat(s string, count any) string {
	n := toIntegerOrInf(count)
	if n < 0 || math.IsInf(n, 0) {
		throwRangeError("Invalid count value")
	}
	if n > 0 && float64(jsLen(s))*n > (1<<29)-24 {
		throwRangeError("Invalid string length")
	}
	return strings.Repeat(s, int(n))
}

// Array.prototype.slice / String.prototype.slice index resolution
func sliceBounds(length int, start any, end any) (int, int) {
	resolve := func(v any, def int) int {
		if isUndefined(v) {
			return def
		}
		rel := toIntegerOrInf(v)
		if rel < 0 {
			return int(math.Max(float64(length)+rel, 0))
		}
		return int(math.Min(rel, float64(length)))
	}
	from := resolve(start, 0)
	to := resolve(end, length)
	if to < from {
		to = from
	}
	return from, to
}

func sliceList(nodes []any, start any, end any) []any {
	from, to := sliceBounds(len(nodes), start, end)
	return append([]any{}, nodes[from:to]...)
}

func sliceString(s string, start any, end any) string {
	u := u16(s)
	from, to := sliceBounds(len(u), start, end)
	return fromU16(u[from:to])
}

// object[index] for arrays and strings: only exact integer indexes hit an element
func indexOf(length int, index float64) (int, bool) {
	if index != math.Trunc(index) || index < 0 || index >= float64(length) {
		return 0, false
	}
	return int(index), true
}

// ─────────────────────────────────────────────────────────────────────────────
// Ordered object map (JS property order: integer keys ascending, then insertion order)
// ─────────────────────────────────────────────────────────────────────────────

type OMap struct {
	keys   []string
	values map[string]any
}

func newOMap() *OMap { return &OMap{values: map[string]any{}} }

func arrayIndexKey(k string) (uint64, bool) {
	n, err := strconv.ParseUint(k, 10, 32)
	if err != nil || n >= 4294967295 || strconv.FormatUint(n, 10) != k {
		return 0, false
	}
	return n, true
}

func (o *OMap) set(k string, v any) {
	if _, ok := o.values[k]; !ok {
		o.keys = append(o.keys, k)
	}
	o.values[k] = v
}

func (o *OMap) get(k string) (any, bool) {
	v, ok := o.values[k]
	return v, ok
}

func (o *OMap) size() int { return len(o.keys) }

func (o *OMap) orderedKeys() []string {
	var indexes, others []string
	for _, k := range o.keys {
		if _, ok := arrayIndexKey(k); ok {
			indexes = append(indexes, k)
		} else {
			others = append(others, k)
		}
	}
	sort.Slice(indexes, func(i, j int) bool {
		a, _ := arrayIndexKey(indexes[i])
		b, _ := arrayIndexKey(indexes[j])
		return a < b
	})
	return append(indexes, others...)
}
