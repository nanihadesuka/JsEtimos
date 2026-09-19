// JSEtimos - copyright nani 2021
//
// Go port of jsetimos.ts (lexer, parser and interpreter in a single file).
//
// The TypeScript implementation leans on the JavaScript runtime for most of the
// value semantics (dynamic typing, `+` coercion, loose `==`, number printing,
// UTF-16 strings, object key order...). This port reproduces those semantics
// explicitly so programs behave the same, including the known quirks.
//
// Build: go build jsetimos.go
// Run:   jsetimos fileName | jsetimos (shell mode) | jsetimos --inputMode "print(1)"

package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
)

// ─────────────────────────────────────────────────────────────────────────────
// External system (node.js equivalent)
// ─────────────────────────────────────────────────────────────────────────────

type ProgramArgs struct {
	fileName  string
	dumpAST   bool
	dumpFile  string
	shellMode bool
	inputMode bool
	inputText string
}

type ExtSystem struct {
	redirectPrint         bool
	redirectPrintCallback func(output string)
	shellOnLineCallback   func(line string)
}

const escapeChar = string(rune(27))

var colors = struct{ reset, red string }{escapeChar + "[0m", escapeChar + "[31;1m"}

var extSystem = &ExtSystem{}

func (e *ExtSystem) readFile(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(&JSError{"Error", err.Error()})
	}
	return string(data)
}

// The TS version has a typo (`baseDir == baseDir ?? ...`) that makes the base
// directory "undefined"; here the intended default (dirName) is used.
func (e *ExtSystem) writeFile(data string, file string, baseDir string) {
	if baseDir == "" {
		baseDir = e.dirName()
	}
	_ = os.WriteFile(baseDir+"/"+file, []byte(data), 0o644)
}

func (e *ExtSystem) existFile(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func (e *ExtSystem) print(msg string) {
	if e.redirectPrint {
		if e.redirectPrintCallback != nil {
			e.redirectPrintCallback(msg)
		}
	} else {
		fmt.Fprintln(os.Stdout, msg)
	}
}

func (e *ExtSystem) formatColor(text string, color string) string {
	return color + text + colors.reset
}

// Equivalent of node's __dirname: the working directory the interpreter runs from.
func (e *ExtSystem) dirName() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

func (e *ExtSystem) programArgs() ProgramArgs {
	attributes := map[string]string{}
	for _, item := range os.Args[1:] {
		if strings.HasPrefix(item, "--") {
			attributes[item[2:]] = item[2:]
		} else if strings.HasPrefix(item, "-") {
			continue
		} else {
			attributes["text"] = item
		}
	}

	inputMode := attributes["inputMode"] != ""
	args := ProgramArgs{
		dumpAST:   attributes["dumpAST"] != "",
		dumpFile:  attributes["dumpFile"],
		shellMode: attributes["shellMode"] != "",
		inputMode: inputMode,
	}
	if inputMode {
		args.inputText = attributes["text"]
	} else {
		args.fileName = attributes["text"]
	}
	return args
}

func (e *ExtSystem) shellLoadInterface() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for scanner.Scan() {
		e.shellOnLineCallback(scanner.Text())
	}
}

func (e *ExtSystem) shellStdoutWrite(text string) {
	fmt.Fprint(os.Stdout, text)
}

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

// ─────────────────────────────────────────────────────────────────────────────
// Tokens
// ─────────────────────────────────────────────────────────────────────────────

type TokenType string

const (
	TK_START_SYMBOL_CHARS TokenType = "START_SYMBOL_CHARS"

	TK_PLUSPLUS   TokenType = "++"
	TK_MINUSMINUS TokenType = "--"

	TK_PLUS  TokenType = "+"
	TK_MINUS TokenType = "-"
	TK_MUL   TokenType = "*"
	TK_DIV   TokenType = "/"
	TK_POWER TokenType = "**"

	TK_SCALAR_PLUS  TokenType = ".+"
	TK_SCALAR_MINUS TokenType = ".-"
	TK_SCALAR_MUL   TokenType = ".*"
	TK_SCALAR_DIV   TokenType = "./"

	TK_COMPOUND_PLUS  TokenType = "+="
	TK_COMPOUND_MINUS TokenType = "-="
	TK_COMPOUND_MUL   TokenType = "*="
	TK_COMPOUND_DIV   TokenType = "/="
	TK_COMPOUND_POWER TokenType = "**="

	TK_LESS             TokenType = "<"
	TK_GREATER          TokenType = ">"
	TK_EQUAL            TokenType = "=="
	TK_NOT_EQUAL        TokenType = "!="
	TK_LESS_OR_EQUAL    TokenType = "<="
	TK_GREATER_OR_EQUAL TokenType = ">="

	TK_LPARE       TokenType = "("
	TK_RPARE       TokenType = ")"
	TK_LBRACKETSQR TokenType = "["
	TK_RBRACKETSQR TokenType = "]"
	TK_LBRACKET    TokenType = "{"
	TK_RBRACKET    TokenType = "}"

	TK_ASSIGN       TokenType = "="
	TK_COMMA        TokenType = ","
	TK_DOLLAR       TokenType = "$"
	TK_DOUBLE_QUOTE TokenType = `"`
	TK_SEMICOLON    TokenType = ";"
	TK_COLON        TokenType = ":"
	TK_MODULO       TokenType = "%"
	TK_PIPE         TokenType = "|"
	TK_DOT          TokenType = "."
	TK_DOT_THREE    TokenType = "..."
	TK_ARROW_RIGHT  TokenType = "->"
	TK_QUESTION     TokenType = "?"

	TK_END_SYMBOL_CHARS TokenType = "END_SYMBOL_CHARS"

	TK_FUNCTION TokenType = "fun"
	TK_RETURN   TokenType = "return"
	TK_IF       TokenType = "if"
	TK_ELIF     TokenType = "elif"
	TK_ELSE     TokenType = "else"
	TK_TRUE     TokenType = "true"
	TK_FALSE    TokenType = "false"
	TK_AND      TokenType = "and"
	TK_NOT      TokenType = "not"
	TK_OR       TokenType = "or"
	TK_DO       TokenType = "do"
	TK_FOR      TokenType = "for"
	TK_AS       TokenType = "as"
	TK_MAP      TokenType = "map"    // this should be an ID
	TK_FILTER   TokenType = "filter" // this should be an ID
	TK_IMPORT   TokenType = "import" // this should be an ID

	TK_END_FIRST_CHAR_IS_ALPHA TokenType = "END_FIRST_CHAR_IS_ALPHA"

	TK_REAL   TokenType = "REAL"
	TK_STRING TokenType = "STRING"

	TK_ID  TokenType = "ID"
	TK_EOL TokenType = "EOL"
	TK_EOF TokenType = "EOF"

	// END SPECIAL TOKENS

	TK_VAL      TokenType = "val"
	TK_CONST    TokenType = "const"
	TK_VAR      TokenType = "var"
	TK_CONTINUE TokenType = "continue"
	TK_BREAK    TokenType = "break"

	// END UNIMPLEMENTED TOKENS
)

// Declaration order matters: it defines the symbol and keyword groups.
var allTokens = []TokenType{
	TK_START_SYMBOL_CHARS,
	TK_PLUSPLUS, TK_MINUSMINUS,
	TK_PLUS, TK_MINUS, TK_MUL, TK_DIV, TK_POWER,
	TK_SCALAR_PLUS, TK_SCALAR_MINUS, TK_SCALAR_MUL, TK_SCALAR_DIV,
	TK_COMPOUND_PLUS, TK_COMPOUND_MINUS, TK_COMPOUND_MUL, TK_COMPOUND_DIV, TK_COMPOUND_POWER,
	TK_LESS, TK_GREATER, TK_EQUAL, TK_NOT_EQUAL, TK_LESS_OR_EQUAL, TK_GREATER_OR_EQUAL,
	TK_LPARE, TK_RPARE, TK_LBRACKETSQR, TK_RBRACKETSQR, TK_LBRACKET, TK_RBRACKET,
	TK_ASSIGN, TK_COMMA, TK_DOLLAR, TK_DOUBLE_QUOTE, TK_SEMICOLON, TK_COLON, TK_MODULO,
	TK_PIPE, TK_DOT, TK_DOT_THREE, TK_ARROW_RIGHT, TK_QUESTION,
	TK_END_SYMBOL_CHARS,
	TK_FUNCTION, TK_RETURN, TK_IF, TK_ELIF, TK_ELSE, TK_TRUE, TK_FALSE, TK_AND, TK_NOT,
	TK_OR, TK_DO, TK_FOR, TK_AS, TK_MAP, TK_FILTER, TK_IMPORT,
	TK_END_FIRST_CHAR_IS_ALPHA,
	TK_REAL, TK_STRING, TK_ID, TK_EOL, TK_EOF,
	TK_VAL, TK_CONST, TK_VAR, TK_CONTINUE, TK_BREAK,
}

type symbolTrie map[uint16]symbolTrie

var itKeywordRe = regexp.MustCompile(`^it\d*$`)

var langChars = struct {
	SYMBOL_CHAR_TREE             symbolTrie
	MULTIPLE_CHAR_FIRST_IS_ALPHA map[string]bool
	RESERVED_KEYWORDS            map[string]bool
}{symbolTrie{}, map[string]bool{}, map[string]bool{}}

func isKeywordIt(keyword string) bool { return itKeywordRe.MatchString(keyword) }

func isKeywordReserved(keyword string) bool {
	return langChars.RESERVED_KEYWORDS[keyword] || isKeywordIt(keyword)
}

func init() {
	indexOfToken := func(t TokenType) int {
		for i, v := range allTokens {
			if v == t {
				return i
			}
		}
		return -1
	}
	between := func(start, end TokenType) []TokenType {
		return allTokens[indexOfToken(start)+1 : indexOfToken(end)]
	}

	symbolChars := between(TK_START_SYMBOL_CHARS, TK_END_SYMBOL_CHARS)
	for _, key := range between(TK_END_SYMBOL_CHARS, TK_END_FIRST_CHAR_IS_ALPHA) {
		langChars.MULTIPLE_CHAR_FIRST_IS_ALPHA[string(key)] = true
	}
	for _, key := range symbolChars {
		langChars.RESERVED_KEYWORDS[string(key)] = true
	}
	for key := range langChars.MULTIPLE_CHAR_FIRST_IS_ALPHA {
		langChars.RESERVED_KEYWORDS[key] = true
	}

	for _, key := range symbolChars {
		node := langChars.SYMBOL_CHAR_TREE
		for _, letter := range u16(string(key)) {
			if _, ok := node[letter]; !ok {
				node[letter] = symbolTrie{}
			}
			node = node[letter]
		}
	}

	// Token sanity check
	tokens := map[TokenType]bool{}
	for _, value := range allTokens {
		if tokens[value] {
			panic(fmt.Sprintf("Duplicated token '%s', check TokenType enum", value))
		}
		tokens[value] = true
	}
}

type Token struct {
	Type  TokenType `js:"type"`
	Value string    `js:"value"`
}

func (t *Token) toString() string {
	return fmt.Sprintf("Token(type='%s', value='%s')", t.Type, t.Value)
}

func (t *Token) is(tokenTypes ...TokenType) bool {
	for _, tokenType := range tokenTypes {
		if tokenType == t.Type {
			return true
		}
	}
	return false
}

func (t *Token) copy() *Token { return &Token{t.Type, t.Value} }

// ─────────────────────────────────────────────────────────────────────────────
// Lexer
// ─────────────────────────────────────────────────────────────────────────────

const noChar = -1 // JavaScript's "" character (out of text)

type lexerState struct {
	absolutePosition int
	positionInLine   int
	line             int
	character        int

	prevAbsolutePosition int
	prevPositionInLine   int
	prevLine             int
	prevCharacter        int
}

type Lexer struct {
	lexerState
	rawText   string
	text      []uint16
	directory string
	file      string
	path      string
}

func newLexer(text string, directory string, file string) *Lexer {
	l := &Lexer{
		lexerState: lexerState{
			absolutePosition: -1, positionInLine: -1, line: 0, character: noChar,
			prevAbsolutePosition: -1, prevPositionInLine: -1, prevLine: 0, prevCharacter: noChar,
		},
		rawText:   text,
		text:      u16(text),
		directory: directory,
		file:      file,
		path:      directory + "/" + file,
	}
	l.advance()
	return l
}

func (l *Lexer) getStateCopy() lexerState  { return l.lexerState }
func (l *Lexer) setState(state lexerState) { l.lexerState = state }

func charStr(c int) string {
	if c == noChar {
		return ""
	}
	return fromU16([]uint16{uint16(c)})
}

func padEnd(s string, n int) string {
	if l := jsLen(s); l < n {
		return s + strings.Repeat(" ", n-l)
	}
	return s
}

func (l *Lexer) formatErrorMessage(msg string, errorLength int) string {
	return msg + "\n\n" +
		l.getCurrentLineTextWithErrorPosition(errorLength) + "\n" +
		fmt.Sprintf("(%s:%d:%d)\n", l.path, l.prevLine+1, l.prevPositionInLine)
}

func (l *Lexer) formatErrorMessageLexer(msg string, errorLength int) string {
	return msg + "\n\n" +
		l.getCurrentLineTextWithErrorPositionLexer(errorLength) + "\n" +
		fmt.Sprintf("(%s:%d:%d)\n", l.path, l.line+1, l.positionInLine)
}

func (l *Lexer) error(msg string, errorLength int) {
	formattedMsg := l.formatErrorMessageLexer(msg, errorLength)
	panic(&LangError{"LexerError",
		extSystem.formatColor("Lexer error", colors.red) + "\n" +
			extSystem.formatColor(" ■ ", colors.red) + formattedMsg,
	})
}

func (l *Lexer) contextLines(line int) string {
	splitText := strings.Split(l.rawText, "\n")
	startLine := max(0, line-7)
	endLine := min(len(splitText), max(0, line+1))
	var out []string
	for i := startLine; i < endLine; i++ {
		out = append(out, padEnd(strconv.Itoa(i+1), 6)+splitText[i])
	}
	return strings.Join(out, "\n")
}

func (l *Lexer) getCurrentLineTextWithErrorPosition(errorLength int) string {
	spaceCount := max(0, l.prevPositionInLine-errorLength)
	underlineSize := max(1, errorLength)
	underLine := extSystem.formatColor(strings.Repeat("═", underlineSize), colors.red)
	return l.contextLines(l.prevLine) + "\n" +
		padEnd("", 6) + strings.Repeat(" ", spaceCount) + underLine
}

func (l *Lexer) getCurrentLineTextWithErrorPositionLexer(errorLength int) string {
	spaceCount := max(0, l.prevPositionInLine-errorLength)
	underlineSize := max(1, errorLength)
	underLine := extSystem.formatColor(strings.Repeat("═", underlineSize), colors.red)
	return l.contextLines(l.line) + "\n" +
		padEnd("", 6) + " " + strings.Repeat(" ", spaceCount) + underLine
}

func (l *Lexer) advance() {
	l.prevAbsolutePosition = l.absolutePosition
	l.prevPositionInLine = l.positionInLine
	l.prevLine = l.line
	l.prevCharacter = l.character

	if l.prevCharacter == '\n' {
		l.positionInLine = 0
		l.line += 1
	}
	l.absolutePosition += 1
	l.positionInLine += 1
	if l.absolutePosition < len(l.text) {
		l.character = int(l.text[l.absolutePosition])
	} else {
		l.character = noChar
	}
}

func (l *Lexer) peekCharacter(next int) int {
	pos := l.absolutePosition + next
	if pos < 0 || pos >= len(l.text) {
		return noChar
	}
	return int(l.text[pos])
}

func (l *Lexer) skipWhitespace() {
	for l.character == ' ' || l.character == '\t' || l.character == '\r' {
		l.advance()
	}
}

func isNewLine(char int) bool { return char == '\n' || char == '\r' }

func (l *Lexer) skipLineComment() {
	if l.character == '/' && l.peekCharacter(1) == '/' {
		l.advance()
		l.advance()
		for !(isNewLine(l.character) || l.character == noChar) {
			l.advance()
		}
	}
}

func (l *Lexer) skipMultilineComment() {
	if l.character == '/' && l.peekCharacter(1) == '*' {
		isEnd := func() bool { return (l.character == '*' && l.peekCharacter(1) == '/') || l.character == noChar }
		l.advance()
		l.advance()
		for !isEnd() {
			l.advance()
		}
		l.advance()
		l.advance()
	}
}

func isDigit(c int) bool    { return c >= '0' && c <= '9' }
func isAlpha(c int) bool    { return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') }
func isAlphaNum(c int) bool { return isAlpha(c) || isDigit(c) || c == '_' }

func (l *Lexer) getNumber() string {
	output := charStr(l.character)
	l.advance()

	for isDigit(l.character) {
		output += charStr(l.character)
		l.advance()
	}

	if l.character == '.' {
		output += charStr(l.character)
		l.advance()
		if !isDigit(l.character) {
			l.error(fmt.Sprintf("Invalid character '%s' in decimal number expression after . (dot), expected at least one digit.", charStr(l.character)), 1)
		}
		for isDigit(l.character) {
			output += charStr(l.character)
			l.advance()
		}
	}
	return output
}

func (l *Lexer) getAlphaNum() string {
	output := charStr(l.character)
	l.advance()
	for isAlphaNum(l.character) {
		output += charStr(l.character)
		l.advance()
	}
	return output
}

// Looks for entry string characters then
// entry interpolation or terminal string character.
func (l *Lexer) getStringStart() string {
	output := []uint16{uint16(l.character)}
	startLine := l.line
	startPositionInLine := l.positionInLine

	for {
		l.advance()
		if l.character == '\\' && l.peekCharacter(1) == 'n' {
			output = append(output, '\n')
			l.advance()
		} else if l.character != noChar {
			output = append(output, uint16(l.character))
		}

		if l.character == '"' {
			break
		}
		if l.character == '{' && l.peekCharacter(-1) == '$' {
			break
		}
		if l.character == noChar {
			break
		}
	}

	if l.character == noChar {
		l.error(fmt.Sprintf("Invalid string expression, "+
			"reached end of file but didn't close string expression, "+
			"must end in '\"' or '${'. "+
			"String starting location at %d:%d.", startLine, startPositionInLine), 1)
	}

	l.advance()
	return fromU16(output)
}

// Expects to be called after a TokenType.RBRACKET
// Checks for previous terminal interpolation string character then
// looks for entry interpolation or terminal string character.
func (l *Lexer) getStringContinuation() string {
	if l.peekCharacter(-1) != '}' {
		l.error("getStringContinuation only can be run after } character.", 1)
	}

	var output []uint16
	startLine := l.line
	startPositionInLine := l.positionInLine

	for {
		if l.character == '\\' && l.peekCharacter(1) == 'n' {
			output = append(output, '\n')
			l.advance()
			l.advance()
			continue
		}

		if l.character != noChar {
			output = append(output, uint16(l.character))
		}

		if l.character == '"' {
			break
		}
		if l.character == '{' && l.peekCharacter(-1) == '$' {
			break
		}
		if l.character == noChar {
			break
		}

		l.advance()
	}

	if l.character == noChar {
		l.error(fmt.Sprintf("Invalid string expression, "+
			"reached end of file but didn't close string expression, "+
			"must end in '\"' or '${'. "+
			"String starting location at %d:%d", startLine, startPositionInLine), 1)
	}

	l.advance()
	return fromU16(output)
}

func (l *Lexer) peekToken(count int) *Token {
	state := l.getStateCopy()
	for i := 0; i < count-1; i++ {
		l.getNextToken()
	}
	token := l.getNextToken()
	l.setState(state)
	return token
}

// Called by the parser after string interpolation } token
func (l *Lexer) getNextTokenStringContinuation() *Token {
	return &Token{TK_STRING, l.getStringContinuation()}
}

func (l *Lexer) symbolSeek() *Token {
	var symbol []uint16
	node := langChars.SYMBOL_CHAR_TREE
	for l.character != noChar {
		next, ok := node[uint16(l.character)]
		if !ok {
			break
		}
		node = next
		symbol = append(symbol, uint16(l.character))
		l.advance()
	}
	s := fromU16(symbol)
	return &Token{TokenType(s), s}
}

func (l *Lexer) getNextToken() *Token {
	l.skipWhitespace()
	l.skipLineComment()
	l.skipWhitespace()
	l.skipMultilineComment()
	l.skipWhitespace()

	if l.character == '\n' {
		l.advance()
		return &Token{TK_EOL, string(TK_EOL)}
	}
	if l.character == noChar {
		l.advance()
		return &Token{TK_EOF, string(TK_EOF)}
	}

	if l.character == '"' {
		return &Token{TK_STRING, l.getStringStart()}
	}

	if _, ok := langChars.SYMBOL_CHAR_TREE[uint16(l.character)]; ok {
		return l.symbolSeek()
	}

	if isDigit(l.character) {
		return &Token{TK_REAL, l.getNumber()}
	}
	if isAlpha(l.character) {
		id := l.getAlphaNum()
		if langChars.MULTIPLE_CHAR_FIRST_IS_ALPHA[id] {
			return &Token{TokenType(id), id}
		}
		return &Token{TK_ID, id}
	}

	l.error(fmt.Sprintf("Invalid character: '%s'.", charStr(l.character)), 1)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// AST nodes
// ─────────────────────────────────────────────────────────────────────────────
//
// As in the TS version, some nodes double as runtime values: List/ListEval,
// Dictionary and the function nodes. The `js` tags give the property names
// (in declaration order) used when a node is converted to text (JSON dump).

type Node interface{ astNode() }

type NoOp struct{}
type ID struct {
	Op   *Token `js:"op"`
	Name string `js:"name"`
}

// They are shims, in reality they are converted to:
// string, number, boolean, undefined, null types
type StringOp struct {
	Strings        []string `js:"strings"`
	Interpolations []any    `js:"interpolations"`
}
type Real struct {
	Op *Token `js:"op"`
}
type Bool struct {
	Op *Token `js:"op"`
}
type Undefined struct{}
type Null struct{}

// List is both the AST list node and the runtime list value.
// Eval marks a ListEval (already evaluated, visiting it returns itself).
type List struct {
	Nodes []any `js:"nodes"`
	Eval  bool
}

type Dictionary struct {
	Nodes *OMap `js:"nodes"`
}

type BlockOp struct {
	List []any `js:"list"`
}
type UnaryOp struct {
	Op    *Token `js:"op"`
	Right any    `js:"right"`
}
type BinaryOp struct {
	Left  any    `js:"left"`
	Op    *Token `js:"op"`
	Right any    `js:"right"`
}
type ShortcircuitOp struct {
	Left  any    `js:"left"`
	Op    *Token `js:"op"`
	Right any    `js:"right"`
}
type ComparisonOp struct {
	Nodes       []any    `js:"nodes"`
	Comparators []*Token `js:"comparators"`
}
type CompoundOp struct {
	Left  *ID    `js:"left"`
	Op    *Token `js:"op"`
	Right any    `js:"right"`
}
type VariableIncrementPrefixOp struct {
	Op    *Token `js:"op"`
	Right *ID    `js:"right"`
}
type VariableIncrementPostfixOp struct {
	Op   *Token `js:"op"`
	Left *ID    `js:"left"`
}
type SpreadListOp struct {
	List any `js:"list"`
}
type AssignOp struct {
	Id         *ID `js:"id"`
	Expression any `js:"expression"`
}
type MultipleAssignOp struct {
	Ids        []*ID `js:"ids"`
	Expression any   `js:"expression"`
}
type PipeNodeOp struct {
	Expression any       `js:"expression"`
	PassType   TokenType `js:"passType"` // "" means null
}
type PipeOp struct {
	Entry any           `js:"entry"`
	Nodes []*PipeNodeOp `js:"nodes"`
}
type RangeOp struct {
	Start any `js:"start"`
	Step  any `js:"step"` // nil means null
	End   any `js:"end"`
}
type ForLoopOp struct {
	Setup        any `js:"setup"`
	RunCondition any `js:"runCondition"`
	Increment    any `js:"increment"`
	Body         any `js:"body"`
}
type Conditional struct {
	Conditions []any `js:"conditions"`
	Bodies     []any `js:"bodies"`
}
type ReturnOp struct {
	Node any `js:"node"`
}

type NamedFunction struct {
	StackNamespace int    `js:"stackNamespace"`
	Name           string `js:"name"`
	Parameters     []*ID  `js:"parameters"`
	Block          any    `js:"block"`
}
type AnonymousFunction struct {
	StackNamespace int    `js:"stackNamespace"`
	Name           string `js:"name"`
	Parameters     []*ID  `js:"parameters"`
	Block          any    `js:"block"`
}
type NativeFunction struct {
	StackNamespace int    `js:"stackNamespace"`
	Name           string `js:"name"`
	Parameters     []*ID  `js:"parameters"`
	Fun            func(args []any) any
}

type FunctionCall struct {
	Name string `js:"name"`
	Args *List  `js:"args"`
}
type PrintselfCall struct {
	Args *List  `js:"args"`
	Text string `js:"text"`
}
type AssertCall struct {
	Args   *List  `js:"args"`
	Text   string `js:"text"`
	File   string `js:"file"`
	Line   int    `js:"line"`
	Column int    `js:"column"`
}
type CallOp struct {
	Func any `js:"func"`
	Args any `js:"args"`
}

type AccessIndexOp struct {
	List  any `js:"list"`
	Index any `js:"index"`
}
type AccessKeyOp struct {
	Object any    `js:"object"`
	Key    string `js:"key"`
}
type AccessMethodOp struct {
	Object any    `js:"object"`
	Key    string `js:"key"`
}

type NamedImportOp struct {
	ModuleAST    any    `js:"moduleAST"`
	ModuleSymbol int    `js:"moduleSymbol"`
	ModuleAlias  string `js:"moduleAlias"`
	FilePath     string `js:"filePath"`
}
type NamedImportLoadOp struct {
	ModuleSymbol int    `js:"moduleSymbol"`
	ModuleAlias  string `js:"moduleAlias"`
}

func (*NoOp) astNode()                       {}
func (*ID) astNode()                         {}
func (*StringOp) astNode()                   {}
func (*Real) astNode()                       {}
func (*Bool) astNode()                       {}
func (*Undefined) astNode()                  {}
func (*Null) astNode()                       {}
func (*List) astNode()                       {}
func (*Dictionary) astNode()                 {}
func (*BlockOp) astNode()                    {}
func (*UnaryOp) astNode()                    {}
func (*BinaryOp) astNode()                   {}
func (*ShortcircuitOp) astNode()             {}
func (*ComparisonOp) astNode()               {}
func (*CompoundOp) astNode()                 {}
func (*VariableIncrementPrefixOp) astNode()  {}
func (*VariableIncrementPostfixOp) astNode() {}
func (*SpreadListOp) astNode()               {}
func (*AssignOp) astNode()                   {}
func (*MultipleAssignOp) astNode()           {}
func (*PipeNodeOp) astNode()                 {}
func (*PipeOp) astNode()                     {}
func (*RangeOp) astNode()                    {}
func (*ForLoopOp) astNode()                  {}
func (*Conditional) astNode()                {}
func (*ReturnOp) astNode()                   {}
func (*NamedFunction) astNode()              {}
func (*AnonymousFunction) astNode()          {}
func (*NativeFunction) astNode()             {}
func (*FunctionCall) astNode()               {}
func (*PrintselfCall) astNode()              {}
func (*AssertCall) astNode()                 {}
func (*CallOp) astNode()                     {}
func (*AccessIndexOp) astNode()              {}
func (*AccessKeyOp) astNode()                {}
func (*AccessMethodOp) astNode()             {}
func (*NamedImportOp) astNode()              {}
func (*NamedImportLoadOp) astNode()          {}

// BaseFunction: NamedFunction, AnonymousFunction and NativeFunction
type BaseFunction interface {
	Node
	header() (stackNamespace int, name string, parameters []*ID)
}

func (f *NamedFunction) header() (int, string, []*ID) { return f.StackNamespace, f.Name, f.Parameters }
func (f *AnonymousFunction) header() (int, string, []*ID) {
	return f.StackNamespace, f.Name, f.Parameters
}
func (f *NativeFunction) header() (int, string, []*ID) { return f.StackNamespace, f.Name, f.Parameters }

func newID(token *Token) *ID        { return &ID{token, token.Value} }
func newList(nodes []any) *List     { return &List{Nodes: nodes} }
func newListEval(nodes []any) *List { return &List{Nodes: nodes, Eval: true} }
func isList(v any) bool             { _, ok := v.(*List); return ok }
func isBaseFunction(v any) bool     { _, ok := v.(BaseFunction); return ok }

func className(v any) string {
	if l, ok := v.(*List); ok && l.Eval {
		return "ListEval"
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name()
}

func (l *List) toString() string { return l.toStringIdent(0) }

func (l *List) toStringIdent(ident int) string {
	parts := make([]string, len(l.Nodes))
	for i, v := range l.Nodes {
		switch x := v.(type) {
		case *Dictionary:
			parts[i] = x.toStringIdent(ident + 2)
		case *List:
			parts[i] = "(" + x.toStringIdent(ident+2) + ")"
		default:
			parts[i] = methodToString(v)
		}
	}
	return strings.Join(parts, ", ")
}

func (d *Dictionary) toString() string { return d.toStringIdent(0) }

func (d *Dictionary) toStringIdent(ident int) string {
	keys := d.Nodes.orderedKeys()
	first := ident == 0
	repr := func(v any, disp int) string {
		switch x := v.(type) {
		case *Dictionary:
			return x.toStringIdent(disp)
		case *List:
			return x.toStringIdent(disp)
		}
		return methodToString(v)
	}

	if len(keys) <= 0 {
		return "{  }"
	}

	space := strings.Repeat(" ", ident)
	space2 := strings.Repeat(" ", ident+2)
	lines := make([]string, len(keys))
	for i, k := range keys {
		v, _ := d.Nodes.get(k)
		lines[i] = space2 + k + " = " + repr(v, ident+2)
	}
	closing := space
	if first {
		closing = ""
	}
	return "{\n" + strings.Join(lines, "\n") + "\n" + closing + "}"
}

// ASTNode.toString(): "\n" + class name + " " + JSON.stringify(this, null, 2)
func astToString(v any) string {
	return "\n" + className(v) + " " + jsonStringify(v)
}

// JSON.stringify(value, null, 2)
func jsonStringify(v any) string {
	s, ok := jsonValue(v, "")
	if !ok {
		return "undefined"
	}
	return s
}

func jsonQuote(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\b':
			sb.WriteString(`\b`)
		case '\f':
			sb.WriteString(`\f`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			if r < 0x20 {
				sb.WriteString(fmt.Sprintf(`\u%04x`, r))
			} else {
				sb.WriteRune(r)
			}
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

type jsonMember struct {
	key   string
	value any
}

func jsonObject(members []jsonMember, indent string) string {
	inner := indent + "  "
	var parts []string
	for _, m := range members {
		if s, ok := jsonValue(m.value, inner); ok {
			parts = append(parts, inner+jsonQuote(m.key)+": "+s)
		}
	}
	if len(parts) == 0 {
		return "{}"
	}
	return "{\n" + strings.Join(parts, ",\n") + "\n" + indent + "}"
}

func jsonArray(items []any, indent string) string {
	if len(items) == 0 {
		return "[]"
	}
	inner := indent + "  "
	parts := make([]string, len(items))
	for i, item := range items {
		s, ok := jsonValue(item, inner)
		if !ok {
			s = "null"
		}
		parts[i] = inner + s
	}
	return "[\n" + strings.Join(parts, ",\n") + "\n" + indent + "]"
}

// Returns false when the value is not serializable (undefined, functions).
func jsonValue(v any, indent string) (string, bool) {
	switch x := v.(type) {
	case nil, nullT:
		return "null", true
	case undefinedT:
		return "", false
	case bool:
		if x {
			return "true", true
		}
		return "false", true
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return "null", true
		}
		return numToString(x), true
	case int:
		return strconv.Itoa(x), true
	case string:
		return jsonQuote(x), true
	case TokenType:
		if x == "" {
			return "null", true
		}
		return jsonQuote(string(x)), true
	case *OMap:
		var members []jsonMember
		for _, k := range x.orderedKeys() {
			value, _ := x.get(k)
			members = append(members, jsonMember{k, value})
		}
		return jsonObject(members, indent), true
	case *StackRecord:
		return jsonObject([]jsonMember{{"memory", newOMap()}}, indent), true // Map serializes as {}
	case *Stack:
		records := make([]any, len(x.stackRecordList))
		for i, r := range x.stackRecordList {
			records[i] = r
		}
		return jsonObject([]jsonMember{
			{"pathFile", x.PathFile}, {"symbol", float64(x.Symbol)},
			{"stackRecordList", records}, {"activeRecord", x.activeRecord},
		}, indent), true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Func:
		return "", false
	case reflect.Slice:
		items := make([]any, rv.Len())
		for i := range items {
			items[i] = rv.Index(i).Interface()
		}
		return jsonArray(items, indent), true
	case reflect.Pointer:
		if rv.IsNil() {
			return "null", true
		}
		elem := rv.Elem()
		if elem.Kind() != reflect.Struct {
			return "{}", true
		}
		var members []jsonMember
		for i := 0; i < elem.NumField(); i++ {
			name := elem.Type().Field(i).Tag.Get("js")
			if name == "" {
				continue
			}
			members = append(members, jsonMember{name, elem.Field(i).Interface()})
		}
		return jsonObject(members, indent), true
	}
	return "{}", true
}

// ─────────────────────────────────────────────────────────────────────────────
// Stack (scopes)
// ─────────────────────────────────────────────────────────────────────────────

func newNativeFunction(symbol int, name string, parameters []string, fun func(args []any) any) *NativeFunction {
	ids := make([]*ID, len(parameters))
	for i, p := range parameters {
		ids[i] = newID(&Token{TK_ID, p})
	}
	return &NativeFunction{symbol, name, ids, fun}
}

func argAt(args []any, i int) any {
	if i < len(args) {
		return args[i]
	}
	return jsUndefined
}

func writeGlobalUtils(stack *Stack) {
	printFn := func(args []any) any {
		extSystem.print(toStr(argAt(args, 0)))
		return jsUndefined
	}

	_print := newNativeFunction(stack.Symbol, "print", []string{"msg"}, printFn)
	printself := newNativeFunction(stack.Symbol, "printself", []string{"msg"}, printFn)
	assert := newNativeFunction(stack.Symbol, "assert", []string{"msg"}, printFn)
	_printToFile := newNativeFunction(stack.Symbol, "printToFile", []string{"data", "file"}, func(args []any) any {
		extSystem.writeFile(toStr(argAt(args, 0)), toStr(argAt(args, 1)), "")
		return jsUndefined
	})

	stack.set(_print.Name, _print)
	stack.set(printself.Name, printself)
	stack.set(assert.Name, assert)
	stack.set(_printToFile.Name, _printToFile)
}

type StackRecord struct {
	keys   []string
	memory map[string]any
}

func newStackRecord() *StackRecord { return &StackRecord{memory: map[string]any{}} }

func (r *StackRecord) get(id string) any {
	if v, ok := r.memory[id]; ok {
		return v
	}
	return jsUndefined
}

func (r *StackRecord) set(id string, value any) {
	if _, ok := r.memory[id]; !ok {
		r.keys = append(r.keys, id)
	}
	r.memory[id] = value
}

func (r *StackRecord) has(id string) bool { _, ok := r.memory[id]; return ok }

type Stack struct {
	PathFile        string
	Symbol          int
	stackRecordList []*StackRecord
	activeRecord    *StackRecord
}

// Modified on the parser stage only (when loading a new module)
var namespaceGlobalCount = 0
var absoluteFilePathToSymbol = map[string]int{}

// Modified on the parse(for checking) and interpreter stage (overrites parse data) only
var importModules = map[int]*Stack{}

func stackReset() {
	namespaceGlobalCount = 0
	absoluteFilePathToSymbol = map[string]int{}
	importModules = map[int]*Stack{}
}

func newStack(pathFile string, symbol int) *Stack {
	s := &Stack{PathFile: pathFile, Symbol: symbol}
	s.push()
	writeGlobalUtils(s)
	return s
}

func (s *Stack) isGlobalScope() bool { return len(s.stackRecordList) == 1 }

func (s *Stack) getGlobalVariablesKeys() []string {
	return append([]string{}, s.stackRecordList[0].keys...)
}

func (s *Stack) push() {
	record := newStackRecord()
	s.stackRecordList = append(s.stackRecordList, record)
	s.activeRecord = record
}

func (s *Stack) pop() {
	if len(s.stackRecordList) > 0 {
		s.stackRecordList = s.stackRecordList[:len(s.stackRecordList)-1]
	}
	if len(s.stackRecordList) > 0 {
		s.activeRecord = s.stackRecordList[len(s.stackRecordList)-1]
	} else {
		s.activeRecord = nil
	}
}

func (s *Stack) get(id string) any {
	for i := len(s.stackRecordList) - 1; i > -1; i-- {
		if s.stackRecordList[i].has(id) {
			return s.stackRecordList[i].get(id)
		}
	}
	return jsUndefined
}

func (s *Stack) has(id string) bool {
	for i := len(s.stackRecordList) - 1; i > -1; i-- {
		if s.stackRecordList[i].has(id) {
			return true
		}
	}
	return false
}

func (s *Stack) hasInSameRecord(id string) bool { return s.activeRecord.has(id) }

func (s *Stack) set(id string, value any) { s.activeRecord.set(id, value) }

// Note: sets the value on every record that already holds the name (as the TS version does)
func (s *Stack) setOnExisting(id string, value any) {
	for i := len(s.stackRecordList) - 1; i > -1; i-- {
		if s.stackRecordList[i].has(id) {
			s.stackRecordList[i].set(id, value)
		}
	}
}

func (s *Stack) getOwnIts() []string {
	var its []string
	for _, k := range s.activeRecord.keys {
		if isKeywordIt(k) {
			its = append(its, k)
		}
	}
	return its
}

// ─────────────────────────────────────────────────────────────────────────────
// Parser
// ─────────────────────────────────────────────────────────────────────────────

/* Extra grammar notation
   · [TOKEN] The token is not to be consumed (not "eat" call) in the currently parsing call
   · {call} The call is to be run in a new child stack scope (push > call > pop) in the currently running call
   · func_a< func_b > Signifies the composition of functions, func_a( () => func_b() )
   · $var Make reference to a local varaible o function
*/

type EatMode int

const (
	EatDefault EatMode = iota
	EatStringContinuation
)

// nVals == noVals is the TS `nVals = undefined`
const noVals = -1

type parserState struct {
	token      *Token
	lexerState lexerState
}

type Parser struct {
	token *Token
	lexer *Lexer
	stack *Stack
}

func newParser(lexer *Lexer, stack *Stack) *Parser {
	p := &Parser{lexer: lexer, stack: stack}
	p.token = lexer.getNextToken()
	return p
}

// Doesn't save stack !!!
func (p *Parser) getState() parserState {
	return parserState{p.token.copy(), p.lexer.getStateCopy()}
}

// Doesn't restore stack !!!
func (p *Parser) restoreState(state parserState) {
	p.token.Type = state.token.Type
	p.token.Value = state.token.Value
	p.lexer.setState(state.lexerState)
}

func (p *Parser) peekToken(count int) *Token { return p.lexer.peekToken(count) }

func (p *Parser) is(tokenTypes ...TokenType) bool { return p.token.is(tokenTypes...) }

func (p *Parser) isChain(tokenTypes ...TokenType) bool {
	if !p.is(tokenTypes[0]) {
		return false
	}
	for i := 1; i < len(tokenTypes); i++ {
		if tokenTypes[i] != p.peekToken(i).Type {
			return false
		}
	}
	return true
}

func (p *Parser) error(msg string, errorLength int) {
	formattedMsg := p.lexer.formatErrorMessage(msg, errorLength)
	panic(&LangError{"ParserError",
		extSystem.formatColor("Parser error", colors.red) + "\n" +
			extSystem.formatColor(" ■ ", colors.red) + formattedMsg,
	})
}

func (p *Parser) eat(tokenType TokenType) { p.eatMode(tokenType, EatDefault) }

func (p *Parser) eatMode(tokenType TokenType, mode EatMode) {
	if p.is(tokenType) {
		if mode == EatDefault {
			p.token = p.lexer.getNextToken()
		} else if mode == EatStringContinuation {
			p.token = p.lexer.getNextTokenStringContinuation()
		} else {
			p.error(fmt.Sprintf("Invalid EatMode %d", mode), 1)
		}
	} else {
		p.error(fmt.Sprintf("Invalid token %s, expected token type '%s'.", p.token.toString(), tokenType), jsLen(p.token.Value))
	}
}

func (p *Parser) eatEOL() {
	if p.is(TK_EOL) {
		p.eat(TK_EOL)
	}
}

func (p *Parser) eatEOLS() {
	for p.is(TK_EOL) {
		p.eat(TK_EOL)
	}
}

func (p *Parser) anonymousFunctionParameters() []*ID {
	// (ID (COMMA ID)* ARROW_RIGHT)?
	var parameters []*ID
	ids := map[string]bool{}

	if p.is(TK_ID) {
		parameter := newID(p.token)
		p.eat(TK_ID)
		parameters = append(parameters, parameter)
		ids[parameter.Name] = true

		for p.is(TK_COMMA) {
			p.eat(TK_COMMA)
			parameter := newID(p.token)
			if ids[parameter.Name] {
				p.error(fmt.Sprintf("Duplicate identifier '%s'.", parameter.Name), jsLen(parameter.Name))
			}
			p.eat(TK_ID)
			parameters = append(parameters, parameter)
		}

		p.eat(TK_ARROW_RIGHT)
	}
	return parameters
}

func (p *Parser) hasAnonymousFunctionParameters() bool {
	// (ID (COMMA ID)* ARROW_RIGHT)?
	if !p.is(TK_ID) {
		return false
	}

	initialState := p.getState()
	p.eat(TK_ID)
	for p.is(TK_COMMA) {
		p.eat(TK_COMMA)
		if !p.is(TK_ID) {
			break
		}
		p.eat(TK_ID)
	}

	hasParameters := p.is(TK_ARROW_RIGHT)
	p.restoreState(initialState)
	return hasParameters
}

func (p *Parser) anonymousFunction(noParams bool) *AnonymousFunction {
	// {LBRACKET anonymousFunctionParameters blockTo  RBRACKET}
	p.stack.push()
	p.eat(TK_LBRACKET)

	hasParams := p.hasAnonymousFunctionParameters()
	if noParams && hasParams {
		p.error("This anonymous function can't have parameters.", 1)
	}

	var parameters []*ID
	if hasParams {
		parameters = p.anonymousFunctionParameters()
	}

	for _, parameter := range parameters {
		p.stack.set(parameter.Name, jsUndefined)
	}
	block := p.blockTo(TK_RBRACKET)
	its := p.stack.getOwnIts()
	if len(parameters) > 0 && len(its) > 0 {
		p.error(fmt.Sprintf("Not allowed to have explicit parameters and also use special 'it' variables %s.", strings.Join(its, ",")), 1)
	}

	if len(its) > 0 {
		// Array.prototype.sort() default ordering compares the numbers as strings
		sorted := make([]float64, len(its))
		for i, v := range its {
			if v[2:] == "" {
				sorted[i] = 0
			} else {
				sorted[i] = jsParseFloat(v[2:]) // digits only, same as parseInt
			}
		}
		sort.SliceStable(sorted, func(i, j int) bool {
			return compareU16(numToString(sorted[i]), numToString(sorted[j])) < 0
		})
		nOfIts := int(sorted[len(sorted)-1]) + 1
		parameters = make([]*ID, nOfIts)
		for i := range parameters {
			name := "it"
			if i != 0 {
				name = "it" + strconv.Itoa(i)
			}
			parameters[i] = newID(&Token{TK_ID, name})
		}
	}

	p.stack.pop()
	namespace := -1
	if p.stack.isGlobalScope() {
		namespace = p.stack.Symbol
	}
	return &AnonymousFunction{namespace, "", parameters, block}
}

func (p *Parser) pipeExpression() *PipeNodeOp {
	// (MAP|FILTER)? ([LBRACKET] anonymousFunction  | ID access)

	var passType TokenType
	if p.is(TK_MAP, TK_FILTER) {
		passType = p.token.Type
		p.eat(passType)
	}

	if p.is(TK_LBRACKET) {
		return &PipeNodeOp{p.anonymousFunction(false), passType}
	}

	if p.is(TK_ID) {
		id := newID(p.token)
		if !p.stack.has(id.Name) {
			p.error(fmt.Sprintf("Function '%s' not defined.", id.Name), jsLen(id.Name))
		}

		fun := p.stack.get(id.Name)

		// Dynamic (error if any will be in the runtime stage)
		switch fun.(type) {
		case *Dictionary, *List, *Stack:
			dynamicFun := p.accessor(1)
			return &PipeNodeOp{dynamicFun, passType}
		}

		f, ok := fun.(BaseFunction)
		if !ok {
			p.error(fmt.Sprintf("Identifier '%s' is not a function.", id.Name), jsLen(id.Name))
		}

		p.eat(TK_ID)
		_, name, _ := f.header()
		return &PipeNodeOp{&FunctionCall{name, newList(nil)}, passType}
	}

	if p.is(TK_LBRACKET) {
		p.error("Anonymous function bracket must always start in the same line as the pipe operator.", jsLen(p.token.Value))
	}

	p.error(fmt.Sprintf("Invalid token %s for pipe node expression.", p.token.toString()), jsLen(p.token.Value))
	return nil
}

// Pipe horizontal travserse with optional accessors
func (p *Parser) expression(nVals int) any {
	// listExpression (PIPE EOL? ( isAccessor getAccessor | pipeExpression))*

	expression := p.listExpression(nVals)

	if !p.is(TK_PIPE) {
		return expression
	}

	for p.is(TK_PIPE) {
		var pipeNodes []*PipeNodeOp
		for p.is(TK_PIPE) {
			p.eat(TK_PIPE)
			p.eatEOL()
			if p.isAccessor() {
				break
			}
			pipeNodes = append(pipeNodes, p.pipeExpression())
		}

		if len(pipeNodes) == 0 {
			expression = p.getAccessors(expression)
		} else {
			expression = p.getAccessors(&PipeOp{expression, pipeNodes})
		}
	}

	return expression
}

// An expression can return multiple values
func (p *Parser) listExpression(nVals int) any {
	// DOT_THREE? disjunction (COMMA (EOL? disjunction))*
	expects := nVals != noVals

	const expandList = false // Buggy don't

	var nodes []any
	nodes = append(nodes, p.disjunction(nVals))

	makeList := false
	for p.is(TK_COMMA) {
		if expects && len(nodes) >= nVals {
			p.error(fmt.Sprintf("Expected %d expressions but got %d. %s", nVals, len(nodes)+1, p.token.toString()), jsLen(p.token.Value))
		}

		if p.isChain(TK_COMMA, TK_RPARE) {
			p.eat(TK_COMMA)
			makeList = true
			break
		}
		p.eat(TK_COMMA)
		p.eatEOL()
		nodes = append(nodes, p.disjunction(nVals))
	}

	if expects && len(nodes) < nVals {
		p.error(fmt.Sprintf("Expected %d expressions but got %d. %s", nVals, len(nodes), p.token.toString()), jsLen(p.token.Value))
	}

	var res any
	if len(nodes) == 1 {
		res = nodes[0]
	} else {
		res = newList(nodes)
	}
	if makeList {
		res = newList([]any{res})
	}
	if expandList {
		res = &SpreadListOp{res}
	}
	return res
}

func (p *Parser) disjunction(nVals int) any {
	// conjunction (OR EOL? conjunction)*
	node := p.conjunction(nVals)
	for p.is(TK_OR) {
		token := p.token
		p.eat(token.Type)
		p.eatEOL()
		node = &ShortcircuitOp{node, token, p.conjunction(nVals)}
	}
	return node
}

func (p *Parser) conjunction(nVals int) any {
	// comparision ( AND EOL? comparision)*
	node := p.comparision(nVals)
	for p.is(TK_AND) {
		token := p.token
		p.eat(token.Type)
		p.eatEOL()
		node = &ShortcircuitOp{node, token, p.comparision(nVals)}
	}
	return node
}

func (p *Parser) comparision(nVals int) any {
	// additive (( < | > | <= | >= | == | != ) EOL? additive)*
	var nodes []any
	var comparators []*Token
	nodes = append(nodes, p.additive(nVals))

	for p.is(TK_LESS, TK_GREATER, TK_EQUAL, TK_NOT_EQUAL, TK_LESS_OR_EQUAL, TK_GREATER_OR_EQUAL) {
		token := p.token
		p.eat(token.Type)
		p.eatEOL()
		comparators = append(comparators, token)
		nodes = append(nodes, p.additive(nVals))
	}

	if len(nodes) == 1 {
		return nodes[0]
	}
	return &ComparisonOp{nodes, comparators}
}

func (p *Parser) additive(nVals int) any {
	// multiplicative ((PLUS|MINUS) EOL? multiplicative)*
	node := p.multiplicative(nVals)
	for p.is(TK_PLUS, TK_MINUS) {
		token := p.token
		p.eat(token.Type)
		p.eatEOL()
		node = &BinaryOp{node, token, p.multiplicative(nVals)}
	}
	return node
}

func (p *Parser) multiplicative(nVals int) any {
	// power ((MUL|DIV|MODULO) EOL? power)*
	node := p.power(nVals)
	for p.is(TK_MUL, TK_DIV, TK_MODULO) {
		token := p.token
		p.eat(token.Type)
		p.eatEOL()
		node = &BinaryOp{node, token, p.power(nVals)}
	}
	return node
}

func (p *Parser) power(nVals int) any {
	// accessor (POWER accessor)*
	node := p.accessor(nVals)
	for p.is(TK_POWER) {
		token := p.token
		p.eat(token.Type)
		p.eatEOL()
		node = &BinaryOp{node, token, p.accessor(nVals)}
	}
	return node
}

func (p *Parser) isAccessor() bool {
	return p.is(TK_DOT, TK_LBRACKETSQR, TK_LPARE, TK_QUESTION)
}

func (p *Parser) getAccessors(node any) any {
	/*
	   (
	       DOT (ID|REAL) |
	       LBRACKETSQR EOL? expression EOL? RBRACKETSQR |
	       LPARE EOLS? (expression EOLS?)? RPARE |
	       QUESTION ID (parameters)?
	   )*
	*/
	for p.isAccessor() {
		if p.is(TK_LBRACKETSQR) {
			p.eat(TK_LBRACKETSQR)
			p.eatEOL()
			node = &AccessIndexOp{node, p.expression(noVals)}
			p.eatEOL()
			p.eat(TK_RBRACKETSQR)
		} else if p.is(TK_DOT) {
			p.eat(TK_DOT)
			key := p.token.Value
			if !p.is(TK_ID, TK_REAL) {
				p.error("Invalid token, access key needs to be a string or integer number.", 1)
			}
			if p.is(TK_REAL) && strings.Contains(key, ".") {
				p.error("Access key invalid: can't be a number with decimals.", 1)
			}
			p.eat(p.token.Type)
			node = &AccessKeyOp{node, key}
		} else if p.is(TK_LPARE) {
			p.eat(TK_LPARE)
			p.eatEOLS()
			var parameters any
			if p.is(TK_RPARE) {
				parameters = newList(nil)
			} else {
				parameters = p.expression(noVals)
			}
			p.eatEOLS()
			p.eat(TK_RPARE)
			node = &CallOp{node, parameters}
		} else if p.is(TK_QUESTION) {
			p.eat(TK_QUESTION)
			id := newID(p.token)
			p.eat(TK_ID)
			node = &AccessMethodOp{node, id.Name}
		}
	}
	return node
}

func (p *Parser) accessor(nVals int) any {
	// range getAccessors
	return p.getAccessors(p.rangeOp(nVals))
}

func (p *Parser) rangeOp(nVals int) any {
	// factor (: factor (: factor)?)?
	node1 := p.factor(nVals)
	if !p.is(TK_COLON) {
		return node1
	}

	p.eat(TK_COLON)
	node2 := p.factor(noVals)
	if !p.is(TK_COLON) {
		return &RangeOp{node1, nil, node2}
	}

	p.eat(TK_COLON)
	node3 := p.factor(noVals)
	return &RangeOp{node1, node2, node3}
}

func (p *Parser) factor(nVals int) any {
	/*
	   (PLUSPLUS|MINUSMINUS) ID |
	   (PLUS|MINUS) EOL? factor |
	   NOT factor |
	   REAL |
	   (TRUE|FALSE) |
	   ID [LPARE] functionCall |
	   ID (PLUSPLUS|MINUSMINUS) |
	   ID LBRACKETSQR EOL? expression EOL? RBRACKETSQR |
	   ID |
	   LPARE EOL? expression EOL? RPARE |
	   [IF] conditionalExpression |
	   [STRING] string |
	   [RETURN] return |
	   DO block |
	*/
	if p.is(TK_PLUSPLUS, TK_MINUSMINUS) {
		token := p.token
		p.eat(token.Type)
		id := newID(p.token)
		p.eat(TK_ID)
		return &VariableIncrementPrefixOp{token, id}
	}
	if p.is(TK_PLUS, TK_MINUS) {
		token := p.token
		p.eat(token.Type)
		p.eatEOL()
		return &UnaryOp{token, p.factor(nVals)}
	}
	if p.is(TK_NOT) {
		token := p.token
		p.eat(token.Type)
		return &UnaryOp{token, p.factor(nVals)}
	} else if p.is(TK_REAL) {
		token := p.token
		p.eat(TK_REAL)
		return &Real{token}
	} else if p.is(TK_TRUE, TK_FALSE) {
		token := p.token
		p.eat(token.Type)
		return &Bool{token}
	} else if p.is(TK_ID) {
		id := newID(p.token)

		if id.Name == "undefined" {
			p.eat(TK_ID)
			return &Undefined{}
		}

		if id.Name == "null" {
			p.eat(TK_ID)
			return &Null{}
		}

		isIt := isKeywordIt(id.Name)
		isArguments := id.Name == "arguments"
		// Special family of variables it, it1, it2, it3 , etc
		// that represent the scope arguments for anonymous functions.
		// They can't be reassigned or modified, only accessed

		if id.Name == "it0" {
			p.error("First 'it' value doesn't use index, try renaming 'it0' to 'it'.", jsLen(id.Name))
		}

		if isIt || isArguments {
			if p.stack.isGlobalScope() {
				p.error(fmt.Sprintf("Not allowed to access special variable '%s' in the global scope.", id.Name), jsLen(id.Name))
			}
			// Here, the name is "marked" so can be later known will be accessed
			p.stack.set(id.Name, jsNull)
		}

		if !p.stack.has(id.Name) {
			p.error(fmt.Sprintf("Trying to use undefined variable '%s'.", id.Name), jsLen(id.Name))
		}

		p.eat(TK_ID)

		if p.is(TK_LPARE) {
			if isIt {
				p.error(fmt.Sprintf("Trying to use the variable '%s' as a function.", id.Name), 1)
			}
			return p.functionCall(id)
		} else if p.is(TK_PLUSPLUS, TK_MINUSMINUS) {
			if isIt {
				p.error(fmt.Sprintf("Not allowed to modify special variable '%s', only allowed to access.", id.Name), jsLen(id.Name))
			}
			token := p.token
			p.eat(token.Type)
			return &VariableIncrementPostfixOp{token, id}
		} else if p.is(TK_LBRACKETSQR) {
			p.eat(TK_LBRACKETSQR)
			p.eatEOLS()
			node := p.expression(noVals)
			p.eatEOLS()
			p.eat(TK_RBRACKETSQR)
			return &AccessIndexOp{id, node}
		}
		return id
	} else if p.is(TK_LPARE) {
		p.eat(TK_LPARE)
		p.eatEOL()
		node := p.expression(noVals)
		p.eatEOL()
		p.eat(TK_RPARE)
		return node
	} else if p.is(TK_IF) {
		return p.conditionalExpression(nVals)
	} else if p.is(TK_RETURN) {
		return p.returnStatement(nVals)
	} else if p.is(TK_STRING) {
		return p.string()
	} else if p.is(TK_DO) {
		p.eat(TK_DO)
		return p.anonymousFunction(true)
	} else if p.is(TK_LBRACKET) {
		return p.dictionary()
	}

	p.error(fmt.Sprintf("Expected expression here but found none. %s.", p.token.toString()), 1)
	return nil
}

func (p *Parser) dictionary() any {
	// LBRACKET EOL? ( (ID|REAL) ASSIGN ( fun anonymousFunction | expression) (EOL|SEMICOLON)? EOLS? )* EOL? RBRACKET
	p.eat(TK_LBRACKET)
	p.eatEOL()
	object := newOMap()
	ids := map[string]bool{}
	for p.is(TK_ID, TK_REAL) {
		if p.is(TK_REAL) && strings.Contains(p.token.Value, ".") {
			p.error("No decimal numbers allowed as dictionary keys.", 1)
		}

		id := p.token.Value
		if ids[id] {
			p.error(fmt.Sprintf("Duplicated key %s.", id), 1)
		} else {
			ids[id] = true
		}

		p.eat(p.token.Type)
		p.eat(TK_ASSIGN)

		isFun := p.is(TK_FUNCTION)
		if isFun {
			p.eat(TK_FUNCTION)
		}

		var node any
		if isFun {
			node = p.anonymousFunction(false)
		} else {
			node = p.expression(noVals)
		}

		if p.is(TK_RBRACKET) {
			object.set(id, node)
			break
		}

		if p.is(TK_ID, TK_REAL) {
			object.set(id, node)
			continue
		}

		if p.is(TK_SEMICOLON) {
			p.eat(TK_SEMICOLON)
		} else {
			p.eat(TK_EOL)
		}

		p.eatEOLS()
		object.set(id, node)
	}
	p.eatEOL()
	p.eat(TK_RBRACKET)

	return &Dictionary{object}
}

func (p *Parser) importStatement() any {
	// IMPORT ID ( DOT ID )* (AS ID)? EOL
	if !p.stack.isGlobalScope() {
		p.error("Import can only be called in the global scope.", jsLen(p.token.Value))
	}

	p.eat(TK_IMPORT)

	ids := []*ID{}
	for {
		token := p.token
		p.eat(TK_ID)
		ids = append(ids, newID(token))
		if !p.is(TK_DOT) {
			break
		}
		p.eat(TK_DOT)
	}

	list := make([]string, len(ids))
	for i, v := range ids {
		list[i] = v.Name
	}
	relativeName := strings.Join(list, ".")
	relativeDir := strings.Join(list[:len(list)-1], "/")
	name := list[len(list)-1]
	fileName := name + ".jk"

	absoluteDir := p.lexer.directory
	if relativeDir != "" {
		absoluteDir += "/" + relativeDir
	}
	absolutePath := absoluteDir + "/" + fileName

	if !extSystem.existFile(absolutePath) {
		p.error(fmt.Sprintf("Import '%s' doesn't exist. Failed to access file in %s.", relativeName, absolutePath), jsLen(relativeName))
	}

	moduleAlias := name
	if p.is(TK_AS) {
		p.eat(TK_AS)
		moduleAlias = p.token.Value
		p.eat(TK_ID)
		p.eat(TK_EOL)
	}

	if p.stack.has(moduleAlias) {
		p.error(fmt.Sprintf("Trying to use duplicated name %s.", moduleAlias), jsLen(moduleAlias))
	}

	// Module AST tree and stack are already parsed, so just add the name
	if moduleSymbol, ok := absoluteFilePathToSymbol[absolutePath]; ok {
		p.stack.set(moduleAlias, moduleStackOrUndefined(moduleSymbol))
		return &NamedImportLoadOp{moduleSymbol, moduleAlias}
	}

	namespaceGlobalCount++
	moduleSymbol := namespaceGlobalCount
	absoluteFilePathToSymbol[absolutePath] = moduleSymbol
	moduleStack := newStack(absolutePath, moduleSymbol)

	rawText := extSystem.readFile(absolutePath)
	lexer := newLexer(rawText, absoluteDir, fileName)
	parser := newParser(lexer, moduleStack)
	moduleAST := parser.parse()
	p.stack.set(moduleAlias, moduleStack)
	importModules[moduleSymbol] = parser.stack
	return &NamedImportOp{moduleAST, moduleSymbol, moduleAlias, absolutePath}
}

// Map.get() returns undefined for missing keys
func moduleStackOrUndefined(symbol int) any {
	if s, ok := importModules[symbol]; ok {
		return s
	}
	return jsUndefined
}

func (p *Parser) string() *StringOp {
	// STRING ( EOLS? expression EOLS? RBRACKET STRING )*

	// string syntax example: "hello ${ callA() } ${ 2*1 } world"

	type extracted struct {
		isTerminal bool
		string     string
	}
	extract := func(raw string) extracted {
		// no terminal: ${  terminal: "
		u := u16(raw)
		isTerminal := len(u) > 0 && u[len(u)-1] == '"'
		start := 0
		if len(u) > 0 && u[0] == '"' {
			start = 1
		}
		end := len(u) - 2
		if isTerminal {
			end = len(u) - 1
		}
		return extracted{isTerminal, sliceString(raw, float64(start), float64(end))}
	}

	var listStrings []string
	var listExpressions []any

	data := extract(p.token.Value)
	p.eat(TK_STRING)

	listStrings = append(listStrings, data.string)

	for !data.isTerminal {
		p.eatEOLS()
		listExpressions = append(listExpressions, p.expression(1))
		p.eatEOLS()
		p.eatMode(TK_RBRACKET, EatStringContinuation)
		data := extract(p.token.Value)
		p.eat(TK_STRING)

		listStrings = append(listStrings, data.string)
		if data.isTerminal {
			break
		}
	}

	return &StringOp{listStrings, listExpressions}
}

func (p *Parser) optionalParenthesis(fn func(has bool) any) any {
	//  fn | LPARE EOLS? fn EOLS? RPARE
	if p.is(TK_LPARE) {
		p.eat(TK_LPARE)
		p.eatEOLS()
		node := fn(true)
		p.eatEOLS()
		p.eat(TK_RPARE)
		return node
	}
	return fn(false)
}

// Meant to always return something (must have terminal else expression)
func (p *Parser) conditionalExpression(nVals int) *Conditional {
	return p.conditionalStatement(true, nVals)
}

func (p *Parser) blockOrExpression(nVals int) any {
	// [LBRACKET] block | expression
	if p.is(TK_LBRACKET) {
		return p.block(nVals)
	}
	return p.expression(nVals)
}

func (p *Parser) conditionalBody(nVals int) any {
	if p.is(TK_LBRACKET) {
		return p.block(nVals)
	}
	return p.blockSubroutine(nVals)
}

func (p *Parser) conditionalStatement(mustHaveElse bool, nVals int) *Conditional {
	/*
	   IF optionalParenthesis<expression> EOL? blockOrExpression EOL?
	   (ELIF optionalParenthesis<expression> EOL? blockOrExpression EOL?)*
	   (ELSE EOL? blockOrExpression)?
	*/
	var conditions []any
	var bodies []any

	p.eat(TK_IF)
	conditions = append(conditions, p.optionalParenthesis(func(bool) any { return p.expression(1) }))
	p.eatEOL()
	bodies = append(bodies, p.conditionalBody(nVals))
	p.eatEOL()

	for p.is(TK_ELIF) {
		p.eat(TK_ELIF)
		conditions = append(conditions, p.optionalParenthesis(func(bool) any { return p.expression(1) }))
		p.eatEOL()
		bodies = append(bodies, p.conditionalBody(nVals))
		p.eatEOL()
	}

	if mustHaveElse || p.is(TK_ELSE) {
		p.eat(TK_ELSE)
		p.eatEOL()
		bodies = append(bodies, p.conditionalBody(nVals))
	}

	return &Conditional{conditions, bodies}
}

func (p *Parser) functionCall(id *ID) any {
	// functionCallArgumnets
	fun, ok := p.stack.get(id.Name).(BaseFunction)

	// Dynamic
	if !ok {
		p.eat(TK_LPARE)
		p.eatEOLS()
		var parameters any
		if p.is(TK_RPARE) {
			parameters = newList(nil)
		} else {
			parameters = p.expression(noVals)
		}
		p.eatEOLS()
		p.eat(TK_RPARE)
		return &CallOp{id, parameters}
	}

	if id.Name == "printself" || id.Name == "assert" {
		startPos := p.lexer.absolutePosition
		args := p.functionCallArgumnets(fun)
		endPos := p.lexer.absolutePosition
		text := p.sourceText(startPos-1, endPos)
		if id.Name == "printself" {
			return &PrintselfCall{args, text}
		}
		return &AssertCall{args, text, p.lexer.path, p.lexer.prevLine, p.lexer.prevPositionInLine}
	}

	return &FunctionCall{id.Name, p.functionCallArgumnets(fun)}
}

// text.slice(start, end).trim().slice(1, -1)
func (p *Parser) sourceText(start int, end int) string {
	from, to := sliceBounds(len(p.lexer.text), float64(start), float64(end))
	text := jsTrim(fromU16(p.lexer.text[from:to]))
	return sliceString(text, 1.0, -1.0)
}

func (p *Parser) functionCallArgumnets(funDeclaration BaseFunction) *List {
	// LPARE EOLS? (expression EOLS?)? RPARE
	_, name, IDs := funDeclaration.header()
	var args []any

	p.eat(TK_LPARE)
	p.eatEOLS()

	if len(IDs) != 0 {
		if len(IDs) == 1 {
			expressions := p.expression(noVals)
			if list, ok := expressions.(*List); ok && len(list.Nodes) != 1 {
				p.error(fmt.Sprintf("%s expected %d argument, %d provided.", name, 1, len(list.Nodes)), 1)
			}
			args = append(args, expressions)
		} else {
			expressions := p.expression(len(IDs))
			list, ok := expressions.(*List)
			if !ok {
				p.error(fmt.Sprintf("%s expected %d arguments, %d provided.", name, len(IDs), 1), 1)
			}
			if len(IDs) != len(list.Nodes) {
				p.error(fmt.Sprintf("%s expected %d arguments, %d provided.", name, len(IDs), len(list.Nodes)), 1)
			}
			args = append(args, list.Nodes...)
		}
		p.eatEOLS()
	}

	p.eat(TK_RPARE)
	return newList(args)
}

func (p *Parser) functionDeclaration() *NamedFunction {
	// FUNCTION ID functionParameters {block}
	functionToken := p.token
	p.eat(TK_FUNCTION)
	name := p.token.Value
	// Checks
	{
		if p.stack.hasInSameRecord(name) {
			p.error(fmt.Sprintf("Trying redefine declaration %s as function name.", name), jsLen(name))
		} else {
			p.stack.set(name, functionToken)
		}
		if isKeywordReserved(name) {
			p.error(fmt.Sprintf("Function name can't be '%s', reserved keyword.", name), jsLen(name))
		}
	}

	p.eat(TK_ID)
	parameters := p.functionParameters()
	namespace := -1
	if p.stack.isGlobalScope() {
		namespace = p.stack.Symbol
	}
	funDeclaration := &NamedFunction{namespace, name, parameters, nil}
	p.stack.set(name, funDeclaration)

	p.stack.push()
	for _, parameter := range parameters {
		p.stack.set(parameter.Name, jsUndefined)
	}

	funDeclaration.Block = p.block(noVals)

	if its := p.stack.getOwnIts(); len(its) > 0 {
		p.error(fmt.Sprintf("Named functions can't use special 'it' variables %s only anonymous functions can.", strings.Join(its, ",")), 1)
	}

	p.stack.pop()
	return funDeclaration
}

func (p *Parser) functionParameters() []*ID {
	// LPARE EOLS? (ID (COMMA EOLS? ID)* EOLS?)? RPARE EOL?
	var parameters []*ID
	ids := map[string]bool{}
	p.eat(TK_LPARE)
	p.eatEOLS()
	if p.is(TK_ID) {
		parameter := newID(p.token)
		p.eat(TK_ID)
		parameters = append(parameters, parameter)
		ids[parameter.Name] = true

		for p.is(TK_COMMA) {
			p.eat(TK_COMMA)
			p.eatEOLS()
			parameter := newID(p.token)
			if ids[parameter.Name] {
				p.error(fmt.Sprintf("Duplicate identifier '%s'.", parameter.Name), jsLen(parameter.Name))
			}
			p.eat(TK_ID)
			parameters = append(parameters, parameter)
		}
		p.eatEOLS()
	}
	p.eat(TK_RPARE)
	p.eatEOL()
	return parameters
}

func (p *Parser) returnStatement(nVals int) *ReturnOp {
	// RETURN ([EOL | RBRACKET | RPAREN] | SEMICOLON | [IF] conditionalExpression | expression)
	p.eat(TK_RETURN)

	if p.is(TK_EOL, TK_RBRACKET, TK_RPARE) {
		return &ReturnOp{&NoOp{}}
	}
	if p.is(TK_SEMICOLON) {
		p.eat(p.token.Type)
		return &ReturnOp{&NoOp{}}
	}
	if p.is(TK_IF) {
		return &ReturnOp{p.conditionalExpression(nVals)}
	}
	return &ReturnOp{p.expression(nVals)}
}

func (p *Parser) block(nVals int) any {
	//  [LBRACKET] blockFromTo [RBRACKET]
	return p.blockFromTo(TK_LBRACKET, TK_RBRACKET)
}

func (p *Parser) blockFromTo(start TokenType, end TokenType) any {
	//  $start ( EOL+ | blockSubroutine )* $end
	p.eat(start)
	return p.blockTo(end)
}

func (p *Parser) blockTo(end TokenType) any {
	//  ( EOL+ | blockSubroutine )* $end
	var nodes []any
	for !p.is(end) {
		if p.is(TK_EOL) {
			p.eatEOLS()
		} else {
			nodes = append(nodes, p.blockSubroutine(noVals))
		}
	}
	p.eat(end)

	switch len(nodes) {
	case 0:
		return &NoOp{}
	case 1:
		return nodes[0]
	}
	return &BlockOp{nodes}
}

// All the usual procedures you can do in a line of code
func (p *Parser) blockSubroutine(nVals int) any {
	/*
	   [FUNCTION] functionDeclaration |
	   [FOR] forLoop |
	   [RETURN] expression |
	   [IF] conditionalStatement |
	   [IMPORT] import |
	   [ID (COMPOUND_DIV|COMPOUND_MUL|COMPOUND_PLUS|COMPOUND_MINUS|COMPOUND_POWER)] expression |
	   [ID] testAssignation assignation |
	   expression
	*/
	if p.is(TK_FUNCTION) {
		return p.functionDeclaration()
	}
	if p.is(TK_FOR) {
		return p.forLoop()
	}
	if p.is(TK_RETURN) {
		return p.returnStatement(nVals)
	}
	if p.is(TK_IF) {
		return p.conditionalStatement(false, nVals)
	}
	if p.is(TK_IMPORT) {
		return p.importStatement()
	}

	if p.is(TK_ID) {
		if p.peekToken(1).is(TK_COMPOUND_DIV, TK_COMPOUND_MUL, TK_COMPOUND_PLUS, TK_COMPOUND_MINUS, TK_COMPOUND_POWER) {
			return p.compoundOperation()
		}
		if p.isAssignation() {
			return p.assignation()
		}
	}

	return p.expression(nVals)
}

func (p *Parser) isAssignation() bool {
	// ID ((COMMA EOLS? ID) ASSIGN)?
	if !p.is(TK_ID) {
		return false
	}

	initialState := p.getState()
	p.eat(TK_ID)
	for p.is(TK_COMMA) {
		p.eat(TK_COMMA)
		p.eatEOLS()
		if !p.is(TK_ID) {
			break
		}
		p.eat(TK_ID)
	}

	is := p.is(TK_ASSIGN)
	p.restoreState(initialState)
	return is
}

func (p *Parser) compoundOperation() any {
	// ID (COMPOUND_DIV|COMPOUND_MUL|COMPOUND_PLUS|COMPOUND_MINUS) expression
	id := newID(p.token)

	if isKeywordIt(id.Name) {
		p.error(fmt.Sprintf("Not allowed to modify special variable %s, inmmutable variable.", id.Name), jsLen(id.Name))
	}
	if isKeywordReserved(id.Name) {
		p.error(fmt.Sprintf("'%s' is a reserved keyword, can't redefine.", id.Name), jsLen(id.Name))
	}
	if !p.stack.has(id.Name) {
		p.error(fmt.Sprintf("'%s' not declared.", id.Name), jsLen(id.Name))
	}

	p.eat(TK_ID)

	if p.is(TK_COMPOUND_DIV, TK_COMPOUND_MUL, TK_COMPOUND_PLUS, TK_COMPOUND_MINUS, TK_COMPOUND_POWER) {
		token := p.token
		p.eat(p.token.Type)
		return &CompoundOp{id, token, p.expression(noVals)}
	}

	p.error(fmt.Sprintf("Invalid token for compound operation %s.", p.token.toString()), jsLen(p.token.Value))
	return nil
}

func (p *Parser) checkForIllegalReassignation(name string) {
	if isKeywordReserved(name) {
		p.error(fmt.Sprintf("'%s' is a reserved keyword, can't redefine.", name), jsLen(name))
	}
	if _, ok := p.stack.get(name).(*Stack); ok {
		p.error(fmt.Sprintf("Can't reassign variable '%s', in use by module as alias.", name), jsLen(name))
	}
}

func (p *Parser) assignation() any {
	// ID (COMMA (ID))* ASSIGN EOL? expression+
	var ids []*ID

	idNode := newID(p.token)
	p.checkForIllegalReassignation(idNode.Name)

	p.eat(TK_ID)
	ids = append(ids, idNode)
	p.stack.set(idNode.Name, idNode)

	for !p.is(TK_ASSIGN) {
		p.eat(TK_COMMA)
		idNode := newID(p.token)
		p.checkForIllegalReassignation(idNode.Name)
		p.eat(TK_ID)
		ids = append(ids, idNode)
		p.stack.set(idNode.Name, idNode)
	}

	p.eat(TK_ASSIGN)
	p.eatEOL()

	// This part might be a good reason to do an aditional parser
	// pass as you can't know statically the number of returns
	// directly (for now this is checked at runtime)

	expression := p.expression(noVals)

	// CASE SIGNLE ASSIGMENT
	if len(ids) == 1 {
		p.stack.set(idNode.Name, expression)
		return &AssignOp{ids[0], expression}
	}

	// CASE MULTIPLE ASSIGMENT
	if list, ok := expression.(*List); ok {
		if len(ids) != len(list.Nodes) {
			p.error(fmt.Sprintf("Assignment needs %d expressions, %d given.", len(ids), len(list.Nodes)), 1)
		}
		for i := range ids {
			p.stack.set(ids[i].Name, list.Nodes[i])
		}
		return &MultipleAssignOp{ids, expression}
	}

	return &MultipleAssignOp{ids, expression}
}

func (p *Parser) forLoop() any {
	/*
	   FOR
	   optionalParenthesis
	   <
	       assignation? SEMICOLON
	       expression? SEMICOLON
	       blockOrExpression?
	   >
	   EOL? ([LBRACKET] block | blockSubroutine) EOL?
	*/
	runBeforeSEMICOLON := func(fn func() any) any {
		var node any
		if p.is(TK_SEMICOLON) {
			node = &NoOp{}
		} else {
			node = fn()
		}
		p.eat(TK_SEMICOLON)
		return node
	}

	p.eat(TK_FOR)

	loopOp := p.optionalParenthesis(func(has bool) any {
		setup := runBeforeSEMICOLON(func() any { return p.assignation() })
		runCondition := runBeforeSEMICOLON(func() any { return p.expression(1) })
		var increment any
		if has && p.is(TK_RPARE) {
			increment = &NoOp{}
		} else {
			increment = p.blockOrExpression(noVals)
		}
		return &ForLoopOp{setup, runCondition, increment, nil}
	}).(*ForLoopOp)

	p.eatEOL()
	if p.is(TK_LBRACKET) {
		loopOp.Body = p.block(noVals)
	} else {
		loopOp.Body = p.blockSubroutine(noVals)
	}
	p.eatEOL()

	return loopOp
}

func (p *Parser) program() any {
	// ( EOL+ | blockSubroutine EOL )* EOF
	var nodes []any
	for !p.is(TK_EOF) {
		if p.is(TK_EOL) {
			p.eatEOLS()
		} else {
			nodes = append(nodes, p.blockSubroutine(noVals))
		}
	}
	p.eat(TK_EOF)

	switch len(nodes) {
	case 0:
		return &NoOp{}
	case 1:
		return nodes[0]
	}
	return &BlockOp{nodes}
}

func (p *Parser) parse() any { return p.program() }

// ─────────────────────────────────────────────────────────────────────────────
// Interpreter
// ─────────────────────────────────────────────────────────────────────────────

type Interpreter struct {
	filePath string
	symbol   int
	stack    *Stack
}

func newInterpreter(filePath string, symbol int) *Interpreter {
	return &Interpreter{filePath, symbol, newStack(filePath, symbol)}
}

func (in *Interpreter) error(msg string) {
	panic(&LangError{"RuntimeError",
		extSystem.formatColor("Runtime error", colors.red) + "\n" +
			extSystem.formatColor(" ■ ", colors.red) + msg + "\n",
	})
}

func (in *Interpreter) visit(node any) any {
	switch n := node.(type) {
	case nil:
		return jsNull
	case *List:
		if n.Eval {
			// The list has already been evaluated
			return n
		}
		return in.visitList(n)
	case *ReturnOp:
		return in.visitReturnOp(n)
	case *Real:
		return jsParseFloat(n.Op.Value)
	case *NamedImportLoadOp:
		in.stack.set(n.ModuleAlias, moduleStackOrUndefined(n.ModuleSymbol))
		return jsUndefined
	case *NamedImportOp:
		return in.visitNamedImportOp(n)
	case *Bool:
		return n.Op.Type == TK_TRUE
	case *NoOp:
		return jsUndefined
	case *UnaryOp:
		return in.visitUnaryOp(n)
	case *AccessIndexOp:
		return in.visitAccessIndexOp(n)
	case *AccessMethodOp:
		return in.visitAccessMethodOp(n)
	case *AccessKeyOp:
		return in.visitAccessKeyOp(n)
	case *Dictionary:
		return in.visitDictionary(n)
	case *CallOp:
		return in.visitCallOp(n)
	case *PipeOp:
		return in.visitPipeOp(n)
	case *VariableIncrementPrefixOp:
		return in.visitVariableIncrementPrefixOp(n)
	case *VariableIncrementPostfixOp:
		return in.visitVariableIncrementPostfixOp(n)
	case *AssignOp:
		in.stack.set(n.Id.Name, in.visit(n.Expression))
		return jsUndefined
	case *MultipleAssignOp:
		return in.visitMultipleAssignOp(n)
	case *ComparisonOp:
		return in.visitComparisonOp(n)
	case *BinaryOp:
		return in.visitBinaryOp(n)
	case *CompoundOp:
		return in.visitCompoundOp(n)
	case *ShortcircuitOp:
		return in.visitShortcircuitOp(n)
	case *NamedFunction:
		in.stack.set(n.Name, n)
		return jsUndefined
	case *FunctionCall:
		return in.executeFunction(in.stack.get(n.Name), n.Args)
	case *AnonymousFunction:
		return in.scopedReturn(func() any { return in.visit(n.Block) })
	case *SpreadListOp:
		res, ok := in.visit(n.List).(*List)
		if !ok {
			in.error("Can't spread, object is not a list.")
		}
		return newListEval(res.Nodes)
	case *PrintselfCall:
		evaluatedArgs := in.visitList(n.Args)
		extSystem.print(n.Text + " ⟶  " + toStr(argAt(evaluatedArgs.Nodes, 0)))
		return jsUndefined
	case *AssertCall:
		return in.visitAssertCall(n)
	case *Conditional:
		return in.visitConditional(n)
	case *ForLoopOp:
		in.visit(n.Setup)
		for looseEq(in.visit(n.RunCondition), true) {
			in.visit(n.Body)
			in.visit(n.Increment)
		}
		return jsUndefined
	case *RangeOp:
		return in.visitRangeOp(n)
	case *BlockOp:
		var last any = jsUndefined
		for _, child := range n.List {
			last = in.visit(child)
		}
		return last
	case *StringOp:
		output := n.Strings[0]
		for i, interpolation := range n.Interpolations {
			output += toStr(in.visit(interpolation))
			output += n.Strings[i+1]
		}
		return output
	case *Undefined:
		return jsUndefined
	case *Null:
		return jsNull
	case *ID:
		return in.visitID(n)
	case Node:
		in.error(fmt.Sprintf("Trying to visit not defined node type: %s for node %s.", className(n), toStr(n)))
	}
	// Not an AST node: already a value
	return node
}

func (in *Interpreter) visitReturnOp(node *ReturnOp) any {
	// Throw ReturnOp so it can signal the
	// capturing parent a return call has taken place
	// but keep vising the child node so a final
	// value can be found
	node.Node = in.visit(node.Node)
	panic(node)
}

func (in *Interpreter) visitNamedImportOp(node *NamedImportOp) any {
	interpreter := newInterpreter(node.FilePath, node.ModuleSymbol)
	interpreter.interpret(node.ModuleAST, false, "")
	importModules[node.ModuleSymbol] = interpreter.stack
	in.stack.set(node.ModuleAlias, interpreter.stack)
	return jsUndefined
}

func (in *Interpreter) visitUnaryOp(node *UnaryOp) any {
	right := in.visit(node.Right)

	if list, ok := right.(*List); ok {
		nodes := make([]any, len(list.Nodes))
		switch node.Op.Type {
		case TK_PLUS:
			return right
		case TK_MINUS:
			for i, v := range list.Nodes {
				nodes[i] = -toNum(v)
			}
			return newList(nodes)
		case TK_NOT:
			for i, v := range list.Nodes {
				nodes[i] = !truthy(v)
			}
			return newList(nodes)
		}
	}

	switch node.Op.Type {
	case TK_PLUS:
		return right
	case TK_MINUS:
		return -toNum(right)
	case TK_NOT:
		return !truthy(right)
	}

	in.error(fmt.Sprintf("Token type %s not implemented.", node.Op.Type))
	return nil
}

func (in *Interpreter) listIndex(object *List, index any) any {
	if n, ok := index.(float64); ok {
		if n < 0 {
			n = float64(len(object.Nodes)) + n
		}
		if i, ok := indexOf(len(object.Nodes), n); ok {
			return object.Nodes[i]
		}
		return jsUndefined
	}

	if list, ok := index.(*List); ok {
		start := argAt(list.Nodes, 0)
		end := argAt(list.Nodes, 1)
		if looseEq(end, -1.0) {
			return newList(sliceList(object.Nodes, start, jsUndefined))
		}
		if cmpLT(end, -1.0) {
			return newList(sliceList(object.Nodes, start, jsAdd(end, 1.0)))
		}
		return newList(sliceList(object.Nodes, start, end))
	}
	in.error("Index value in not integer or range.")
	return nil
}

func (in *Interpreter) stringIndex(object string, index any) any {
	if n, ok := index.(float64); ok {
		u := u16(object)
		if n < 0 {
			n = float64(len(u)) + n
		}
		if i, ok := indexOf(len(u), n); ok {
			return fromU16(u[i : i+1])
		}
		return jsUndefined
	}

	if list, ok := index.(*List); ok {
		start := argAt(list.Nodes, 0)
		end := argAt(list.Nodes, 1)
		if looseEq(end, -1.0) {
			return sliceString(object, start, jsUndefined)
		}
		if cmpLT(end, -1.0) {
			return sliceString(object, start, jsAdd(end, 1.0))
		}
		return sliceString(object, start, end)
	}
	in.error("Index value in not integer or range.")
	return nil
}

func (in *Interpreter) visitAccessIndexOp(node *AccessIndexOp) any {
	object := in.visit(node.List)
	if list, ok := object.(*List); ok {
		return in.listIndex(list, in.visit(node.Index))
	}
	if s, ok := object.(string); ok {
		return in.stringIndex(s, in.visit(node.Index))
	}
	in.error("Object is not a list or string, can't acccess index.")
	return nil
}

func sumNodes(nodes []any) any {
	var acc any = 0.0
	for _, v := range nodes {
		acc = jsAdd(acc, v)
	}
	return acc
}

// Math.max / Math.min over the values (ToNumber each)
func mathMaxMin(nodes []any, isMax bool) float64 {
	res := math.Inf(-1)
	if !isMax {
		res = math.Inf(1)
	}
	isNaN := false
	for _, v := range nodes {
		n := toNum(v)
		switch {
		case math.IsNaN(n):
			isNaN = true
		case isMax && (n > res || (n == 0 && res == 0 && !math.Signbit(n))):
			res = n
		case !isMax && (n < res || (n == 0 && res == 0 && math.Signbit(n))):
			res = n
		}
	}
	if isNaN {
		return math.NaN()
	}
	return res
}

func firstOr(nodes []any) any { return argAt(nodes, 0) }
func lastOr(nodes []any) any  { return argAt(nodes, len(nodes)-1) }

// Defined only for the primitives types:
// list, dict, string, number, bool, undefined and null
func (in *Interpreter) visitAccessMethodOp(node *AccessMethodOp) any {
	object := in.visit(node.Object)

	switch obj := object.(type) {
	case *List:
		nodes := obj.Nodes
		switch node.Key {
		case "toString":
			return toStr(obj)
		case "toDict":
			dict := newOMap()
			for _, v := range nodes {
				entry, ok := v.(*List)
				if !ok {
					throwTypeError("Iterator value " + toStr(v) + " is not an entry object")
				}
				dict.set(toStr(argAt(entry.Nodes, 0)), argAt(entry.Nodes, 1))
			}
			return &Dictionary{dict}
		case "size":
			return float64(len(nodes))
		case "sort":
			var defined, undefinedItems []any
			for _, v := range nodes {
				if isUndefined(v) {
					undefinedItems = append(undefinedItems, v)
				} else {
					defined = append(defined, v)
				}
			}
			key := func(v any) any {
				if l, ok := v.(*List); ok {
					return argAt(l.Nodes, 0)
				}
				return v
			}
			sort.SliceStable(defined, func(i, j int) bool {
				return toNum(jsSub(key(defined[i]), key(defined[j]))) < 0
			})
			return newList(append(defined, undefinedItems...))
		case "reverse":
			reversed := make([]any, len(nodes))
			for i, v := range nodes {
				reversed[len(nodes)-1-i] = v
			}
			return newList(reversed)
		case "flat":
			var flat []any
			for _, v := range nodes {
				if l, ok := v.(*List); ok {
					flat = append(flat, l.Nodes...)
				} else {
					flat = append(flat, v)
				}
			}
			return newList(flat)
		case "is":
			return "list"
		case "first":
			return firstOr(nodes)
		case "last":
			return lastOr(nodes)
		case "zip":
			if len(nodes) != 2 {
				in.error("Zip call needs a list with two sub lists of the same size.")
			}
			first, ok1 := nodes[0].(*List)
			second, ok2 := nodes[1].(*List)
			if !ok1 || !ok2 {
				in.error("Zip call needs a list with two sub lists of the same size.")
			}
			if len(first.Nodes) != len(second.Nodes) {
				in.error("Zip call needs a list with two sub lists of the same size.")
			}
			zipped := make([]any, len(first.Nodes))
			for i, v := range first.Nodes {
				zipped[i] = newList([]any{v, second.Nodes[i]})
			}
			return newList(zipped)
		case "lastIndex":
			return float64(len(nodes) - 1)
		case "isEmpty":
			return len(nodes) == 0
		case "isNotEmpty":
			return len(nodes) != 0
		case "sum":
			return sumNodes(nodes)
		case "max":
			return mathMaxMin(nodes, true)
		case "min":
			return mathMaxMin(nodes, false)
		case "average":
			if len(nodes) == 0 {
				return 0.0
			}
			return jsDiv(sumNodes(nodes), float64(len(nodes)))
		case "randomItem":
			if len(nodes) == 0 {
				return jsUndefined
			}
			return nodes[int(math.Floor(rand.Float64()*float64(len(nodes))))]
		case "dropFirst":
			return newList(sliceList(nodes, 1.0, jsUndefined))
		case "dropLast":
			return newList(sliceList(nodes, 0.0, -1.0))
		case "hypot":
			var acc any = 0.0
			for _, c := range nodes {
				acc = jsAdd(acc, jsMul(c, c))
			}
			return math.Sqrt(toNum(acc))
		case "group":
			dict := &Dictionary{newOMap()}
			for _, item := range nodes {
				pair, ok := item.(*List)
				if !ok || len(pair.Nodes) != 2 {
					in.error("Group operation. List items must be list of size 2: (key, value).")
				}
				key := pair.Nodes[0]
				if t := jsType(key); t != "string" && t != "number" {
					in.error("Group operation. List item key must be 'string' or 'number'.")
				}
				value := pair.Nodes[1]
				k := toStr(key)
				if existing, ok := dict.Nodes.get(k); ok {
					group := existing.(*List)
					group.Nodes = append(group.Nodes, value)
				} else {
					dict.Nodes.set(k, newList([]any{value}))
				}
			}
			return dict
		}
		in.error(fmt.Sprintf("Method '%s' not found in object type 'list'.", node.Key))

	case *Dictionary:
		keys := obj.Nodes.orderedKeys()
		switch node.Key {
		case "toString":
			return toStr(obj)
		case "toList":
			entries := make([]any, len(keys))
			for i, k := range keys {
				v, _ := obj.Nodes.get(k)
				entries[i] = newList([]any{k, v})
			}
			return newList(entries)
		case "size":
			return float64(len(keys))
		case "keys":
			list := make([]any, len(keys))
			for i, k := range keys {
				list[i] = k
			}
			return newList(list)
		case "values":
			list := make([]any, len(keys))
			for i, k := range keys {
				list[i], _ = obj.Nodes.get(k)
			}
			return newList(list)
		case "is":
			return "dict"
		case "isEmpty":
			return len(keys) == 0
		case "isNotEmpty":
			return len(keys) != 0
		}
		in.error(fmt.Sprintf("Method '%s' not found in object type 'dict' (dictionary).", node.Key))

	case string:
		u := u16(obj)
		switch node.Key {
		case "toString":
			return obj
		case "toNumber":
			return jsParseFloat(obj)
		case "toList":
			list := make([]any, len(u))
			for i := range u {
				list[i] = fromU16(u[i : i+1])
			}
			return newList(list)
		case "size":
			return float64(len(u))
		case "toLowerCase":
			return strings.ToLower(obj)
		case "toUpperCase":
			return strings.ToUpper(obj)
		case "trim":
			return jsTrim(obj)
		case "trimEnd":
			return jsTrimEnd(obj)
		case "trimStart":
			return jsTrimStart(obj)
		case "capitalize":
			if len(u) == 0 {
				return ""
			}
			return strings.ToUpper(fromU16(u[:1])) + fromU16(u[1:])
		case "is":
			return "string"
		case "lines":
			parts := strings.Split(obj, "\n")
			list := make([]any, len(parts))
			for i, part := range parts {
				list[i] = part
			}
			return newList(list)
		case "isEmpty":
			return obj == ""
		case "isNotEmpty":
			return obj != ""
		case "isBlank":
			return jsTrim(obj) == ""
		case "isNotBlank":
			return jsTrim(obj) != ""
		case "first":
			if len(u) == 0 {
				return jsUndefined
			}
			return fromU16(u[:1])
		case "last":
			if len(u) == 0 {
				return jsUndefined
			}
			return fromU16(u[len(u)-1:])
		case "hashCode":
			// Java's equivalent
			var hash int32
			for _, c := range u {
				hash = (hash << 5) - hash + int32(c) // int32
			}
			return float64(hash)
		}
		in.error(fmt.Sprintf("Method '%s' not found in object type 'string'.", node.Key))

	case float64:
		switch node.Key {
		case "toString":
			return numToString(obj)
		case "toBool":
			return truthy(obj)
		case "toList":
			n := arrayLength(math.Floor(math.Abs(obj)))
			list := make([]any, n)
			for i := range list {
				list[i] = float64(i)
			}
			return newList(list)
		case "trunc":
			return math.Trunc(obj)
		case "round":
			return mathRound(obj)
		case "ceil":
			return math.Ceil(obj)
		case "floor":
			return math.Floor(obj)
		case "abs":
			return math.Abs(obj)
		case "sign":
			return mathSign(obj)
		case "cos":
			return math.Cos(obj)
		case "sin":
			return math.Sin(obj)
		case "tan":
			return math.Tan(obj)
		case "cosh":
			return math.Cosh(obj)
		case "sinh":
			return math.Sinh(obj)
		case "tanh":
			return math.Tanh(obj)
		case "random":
			return rand.Float64() * obj
		case "randomInt":
			return math.Floor(rand.Float64() * (obj + 1))
		case "randomSign":
			return (1 - math.Floor(rand.Float64()*2)*2) * obj
		case "sqrt":
			return math.Sqrt(obj)
		case "negative":
			return -obj
		case "isNegative":
			return obj < 0
		case "isPositive":
			return obj >= 0
		case "isInfinite":
			return math.IsInf(obj, 0)
		case "isFinite":
			// `object !== NaN` is always true in the original, so NaN counts as finite
			return !math.IsInf(obj, 0)
		case "isNaN":
			// `object === NaN` is always false in the original
			return false
		case "is":
			return "number"
		case "rad":
			return obj * math.Pi / 180
		case "deg":
			return obj * 180 / math.Pi
		case "pi":
			return obj * math.Pi
		}
		in.error(fmt.Sprintf("Method '%s' not found in object type 'number': %s", node.Key, numToString(obj)))

	case bool:
		switch node.Key {
		case "toString":
			return toStr(obj)
		case "toNumber":
			return boolToNum(obj)
		case "flip":
			return !obj
		case "is":
			return "bool"
		}
		in.error(fmt.Sprintf("Method '%s' not found in object type 'bool': %s", node.Key, toStr(obj)))

	case undefinedT:
		switch node.Key {
		case "toString", "is":
			return "undefined"
		}
		in.error(fmt.Sprintf("Method '%s' not found in object type 'undefined': undefined", node.Key))

	case nullT:
		switch node.Key {
		case "toString", "is":
			return "null"
		}
		in.error(fmt.Sprintf("Method '%s' not found in object type 'null': null", node.Key))

	case *Stack:
		switch node.Key {
		case "keys":
			keys := obj.getGlobalVariablesKeys()
			list := make([]any, len(keys))
			for i, k := range keys {
				list[i] = k
			}
			return newList(list)
		case "is":
			return "dict"
		}
		in.error(fmt.Sprintf("Method '%s' not found in object type 'null': %s", node.Key, toStr(obj)))
	}

	if node.Key == "toString" {
		return toStr(object)
	}

	in.error(fmt.Sprintf("Can't call method '%s'' from unknown object type '%s'", node.Key, jsTypeof(object)))
	return nil
}

func (in *Interpreter) visitAccessKeyOp(node *AccessKeyOp) any {
	object := in.visit(node.Object)
	if dict, ok := object.(*Dictionary); ok {
		if v, ok := dict.Nodes.get(node.Key); ok {
			return v
		}
		return jsUndefined
	}
	if stack, ok := object.(*Stack); ok {
		if !stack.has(node.Key) {
			name := "undefined"
			if id, ok := node.Object.(*ID); ok {
				name = id.Name
			}
			in.error(fmt.Sprintf("Module %s: %s.%s not found.", name, name, node.Key))
		}
		return stack.get(node.Key)
	}
	in.error("Object is not a dictionary o module, can't acccess by key.")
	return nil
}

func (in *Interpreter) visitList(node *List) *List {
	nodes := make([]any, len(node.Nodes))
	for i, v := range node.Nodes {
		nodes[i] = in.visit(v)
	}
	return newListEval(nodes)
}

func (in *Interpreter) visitDictionary(node *Dictionary) any {
	nodes := newOMap()
	for _, key := range node.Nodes.orderedKeys() {
		value, _ := node.Nodes.get(key)
		if _, ok := value.(*AnonymousFunction); ok {
			nodes.set(key, value)
		} else {
			nodes.set(key, in.visit(value))
		}
	}
	return &Dictionary{nodes}
}

func asArgs(v any) *List {
	if list, ok := v.(*List); ok {
		return list
	}
	return newList([]any{v})
}

func (in *Interpreter) visitCallOp(node *CallOp) any {
	fun := in.visit(node.Func)
	if isBaseFunction(fun) {
		return in.executeFunction(fun, asArgs(node.Args))
	}
	in.error(fmt.Sprintf("Object is not an function, can't execute a call. Got '%s' instead.", toStr(fun)))
	return nil
}

func (in *Interpreter) visitPipeOp(pipe *PipeOp) any {
	itemExec := func(fun any, arg any) any {
		return in.executeFunction(fun, asArgs(arg))
	}

	left := in.visit(pipe.Entry)
	for _, pipeNode := range pipe.Nodes {
		args := asArgs(left)

		var fun any
		switch expression := pipeNode.Expression.(type) {
		case *FunctionCall:
			fun = in.stack.get(expression.Name)
		case *AnonymousFunction:
			fun = expression
		default:
			fun = in.visit(pipeNode.Expression)
			if !isBaseFunction(fun) {
				in.error(fmt.Sprintf("Invalid pipe operator, not a function: '%s'.", toStr(fun)))
			}
		}

		switch pipeNode.PassType {
		case TK_MAP:
			mapped := make([]any, len(args.Nodes))
			for i, arg := range args.Nodes {
				mapped[i] = itemExec(fun, arg)
			}
			left = newList(mapped)
		case TK_FILTER:
			var filtered []any
			for _, arg := range args.Nodes {
				if truthy(itemExec(fun, arg)) {
					filtered = append(filtered, arg)
				}
			}
			left = newList(filtered)
		default:
			left = in.executeFunction(fun, args)
		}
	}

	return left
}

func (in *Interpreter) visitVariableIncrementPrefixOp(node *VariableIncrementPrefixOp) any {
	currentVal := in.visitID(node.Right)

	switch node.Op.Type {
	case TK_PLUSPLUS:
		newVal := jsAdd(currentVal, 1.0)
		in.stack.set(node.Right.Name, newVal)
		return newVal
	case TK_MINUSMINUS:
		newVal := jsSub(currentVal, 1.0)
		in.stack.set(node.Right.Name, newVal)
		return newVal
	}

	in.error(fmt.Sprintf("Token type %s not implemented.", node.Op.Type))
	return nil
}

func (in *Interpreter) visitVariableIncrementPostfixOp(node *VariableIncrementPostfixOp) any {
	currentVal := in.visitID(node.Left)

	switch node.Op.Type {
	case TK_PLUSPLUS:
		in.stack.set(node.Left.Name, jsAdd(currentVal, 1.0))
		return currentVal
	case TK_MINUSMINUS:
		in.stack.set(node.Left.Name, jsSub(currentVal, 1.0))
		return currentVal
	}

	in.error(fmt.Sprintf("Token type %s not implemented.", node.Op.Type))
	return nil
}

func idNames(ids []*ID) string {
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = id.Name
	}
	return strings.Join(names, ",")
}

func (in *Interpreter) visitMultipleAssignOp(node *MultipleAssignOp) any {
	value := in.visit(node.Expression)
	res, ok := value.(*List)
	if !ok {
		in.error(fmt.Sprintf("Assignment %s expected %d values got one: '%s'.", idNames(node.Ids), len(node.Ids), toStr(value)))
	}

	if len(node.Ids) > len(res.Nodes) {
		in.error(fmt.Sprintf("Invalid assignment, expected %d values for %s expression got %d: %s.",
			len(node.Ids), idNames(node.Ids), len(res.Nodes), arrayToString(res.Nodes)))
	}

	for i, id := range node.Ids {
		in.stack.set(id.Name, res.Nodes[i])
	}
	return jsUndefined
}

func listsLooseEqual(left, right *List) bool {
	if len(left.Nodes) != len(right.Nodes) {
		return false
	}
	for i, v := range left.Nodes {
		if !looseEq(v, right.Nodes[i]) {
			return false
		}
	}
	return true
}

func (in *Interpreter) visitComparisonOp(node *ComparisonOp) any {
	left := in.visit(node.Nodes[0])
	for i, comparator := range node.Comparators {
		right := in.visit(node.Nodes[i+1])
		leftList, isListL := left.(*List)
		rightList, isListR := right.(*List)

		var holds bool
		switch comparator.Type {
		case TK_EQUAL:
			if isListL && isListR {
				holds = listsLooseEqual(leftList, rightList)
			} else {
				holds = looseEq(left, right)
			}
		case TK_NOT_EQUAL:
			if isListL && isListR {
				holds = !listsLooseEqual(leftList, rightList)
			} else {
				holds = !looseEq(left, right)
			}
		case TK_LESS:
			holds = cmpLT(left, right)
		case TK_LESS_OR_EQUAL:
			holds = cmpLE(left, right)
		case TK_GREATER:
			holds = cmpGT(left, right)
		case TK_GREATER_OR_EQUAL:
			holds = cmpGE(left, right)
		default:
			in.error(fmt.Sprintf("Token type %s not implemented", comparator.Type))
		}
		if !holds {
			return false
		}
		left = right
	}
	return true
}

// new Array(times).fill(0).flatMap(_ => nodes)
func repeatNodes(nodes []any, times any) []any {
	n := arrayLength(times)
	var out []any
	for i := 0; i < n; i++ {
		out = append(out, nodes...)
	}
	return out
}

func (in *Interpreter) visitBinaryOp(node *BinaryOp) any {
	left := in.visit(node.Left)
	right := in.visit(node.Right)
	leftList, isListL := left.(*List)
	rightList, isListR := right.(*List)
	op := node.Op.Type

	if isListL && !isListR {
		if op == TK_PLUS {
			return newList(append(append([]any{}, leftList.Nodes...), right))
		}
		if op == TK_MUL {
			return newList(repeatNodes(leftList.Nodes, right))
		}
	}
	if !isListL && isListR {
		if op == TK_PLUS {
			return newList(append([]any{left}, rightList.Nodes...))
		}
		if op == TK_MUL {
			return newList(repeatNodes(rightList.Nodes, left))
		}
	}
	if isListL && isListR {
		if op == TK_PLUS {
			return newList(append(append([]any{}, leftList.Nodes...), rightList.Nodes...))
		}
	}

	leftString, isStringL := left.(string)
	rightString, isStringR := right.(string)
	if !isStringL && isStringR && op == TK_MUL {
		return jsRepeat(rightString, left)
	}
	if isStringL && !isStringR && op == TK_MUL {
		return jsRepeat(leftString, right)
	}

	switch op {
	case TK_PLUS:
		return jsAdd(left, right)
	case TK_MINUS:
		return jsSub(left, right)
	case TK_MUL:
		return jsMul(left, right)
	case TK_DIV:
		return jsDiv(left, right)
	case TK_MODULO:
		return jsMod(left, right)
	case TK_POWER:
		return jsPow(left, right)
	}

	in.error(fmt.Sprintf("Token type %s not implemented.", op))
	return nil
}

func (in *Interpreter) visitCompoundOp(node *CompoundOp) any {
	id := node.Left
	value := in.visitID(id)
	compoundValue := in.visit(node.Right)

	valueList, isListL := value.(*List)
	compoundList, isListR := compoundValue.(*List)

	switch node.Op.Type {
	case TK_COMPOUND_PLUS:
		if isListL && !isListR {
			in.stack.setOnExisting(id.Name, newList(append(append([]any{}, valueList.Nodes...), compoundValue)))
			return jsUndefined
		}
		if isListL && isListR {
			in.stack.setOnExisting(id.Name, newList(append(append([]any{}, valueList.Nodes...), compoundList.Nodes...)))
			return jsUndefined
		}
		if !isListL && !isListR {
			in.stack.setOnExisting(id.Name, jsAdd(value, compoundValue))
			return jsUndefined
		}
	case TK_COMPOUND_MINUS:
		if !isListL && !isListR {
			in.stack.setOnExisting(id.Name, jsSub(value, compoundValue))
			return jsUndefined
		}
	case TK_COMPOUND_MUL:
		if isListL && !isListR {
			in.stack.setOnExisting(id.Name, newList(repeatNodes(valueList.Nodes, compoundValue)))
			return jsUndefined
		}
		if !isListL && !isListR {
			in.stack.setOnExisting(id.Name, jsMul(value, compoundValue))
			return jsUndefined
		}
	case TK_COMPOUND_DIV:
		if !isListL && !isListR {
			in.stack.setOnExisting(id.Name, jsDiv(value, compoundValue))
			return jsUndefined
		}
	case TK_COMPOUND_POWER:
		if !isListL && !isListR {
			in.stack.setOnExisting(id.Name, jsPow(value, compoundValue))
			return jsUndefined
		}
	}

	kind := func(isList bool) string {
		if isList {
			return "list"
		}
		return "scalar"
	}
	in.error(fmt.Sprintf("Invalid token %s operation for combination of %s with %s.", node.Op.toString(), kind(isListL), kind(isListR)))
	return nil
}

func (in *Interpreter) visitShortcircuitOp(node *ShortcircuitOp) any {
	left := in.visit(node.Left)
	switch node.Op.Type {
	case TK_AND:
		if !truthy(left) {
			return false
		}
		return truthy(in.visit(node.Right))
	case TK_OR:
		if truthy(left) {
			return true
		}
		return truthy(in.visit(node.Right))
	}
	in.error(fmt.Sprintf("Token type %s not implemented.", node.Op.Type))
	return nil
}

func (in *Interpreter) executeFunction(funValue any, args *List) (result any) {
	evaluatedArgs := in.visitList(args)
	currentScope := in.stack

	fun, ok := funValue.(BaseFunction)
	if !ok {
		// What the JavaScript runtime reports when a non function reaches this point
		if isNullish(funValue) {
			throwTypeError(fmt.Sprintf("Cannot read properties of %s (reading 'stackNamespace')", toStr(funValue)))
		}
		throwTypeError("Cannot read properties of undefined (reading 'pop')")
	}
	stackNamespace, _, parameters := fun.header()

	// -1 means the base stack scope is the same
	// (happens when a function is not defined in global scope)
	if stackNamespace != -1 {
		stack, ok := importModules[stackNamespace]
		if !ok {
			throwTypeError("Cannot read properties of undefined (reading 'pop')")
		}
		in.stack = stack
	}
	in.stack.push()

	defer func() {
		recovered := recover()
		in.stack.pop()
		in.stack = currentScope
		if recovered != nil {
			if value, ok := recovered.(*ReturnOp); ok {
				result = value.Node
				return
			}
			panic(recovered)
		}
	}()

	in.stack.set("arguments", evaluatedArgs)
	for i, parameter := range parameters {
		in.stack.set(parameter.Name, argAt(evaluatedArgs.Nodes, i))
	}

	switch f := fun.(type) {
	case *NamedFunction:
		return in.visit(f.Block)
	case *AnonymousFunction:
		return in.visit(f.Block)
	case *NativeFunction:
		return f.Fun(evaluatedArgs.Nodes)
	}
	in.error("FunctionCall, invalid instanceof type.")
	return nil
}

func (in *Interpreter) visitAssertCall(call *AssertCall) any {
	res := in.visitList(call.Args)
	if value, ok := argAt(res.Nodes, 0).(bool); ok && value {
		return jsUndefined
	}

	location := fmt.Sprintf("(%s:%d:%d)", call.File, call.Line+1, call.Column)
	extSystem.print(
		extSystem.formatColor("FAILED ASSERT:", colors.red) + "  " + call.Text + "\n" +
			location,
	)
	return jsUndefined
}

func (in *Interpreter) visitConditional(node *Conditional) any {
	for i, condition := range node.Conditions {
		if truthy(in.visit(condition)) {
			return in.visit(node.Bodies[i])
		}
	}
	if len(node.Conditions) != len(node.Bodies) {
		return in.visit(node.Bodies[len(node.Bodies)-1])
	}
	return jsUndefined
}

func (in *Interpreter) visitRangeOp(node *RangeOp) any {
	var list []any
	start := in.visit(node.Start)
	stepValue := in.visit(node.Step)
	if isNullish(stepValue) {
		stepValue = 1.0
	}
	step := math.Abs(toNum(stepValue))
	if step == 0 {
		in.error("Range step can't be zero.")
	}
	end := in.visit(node.End)

	invStep := 1.0 / step
	if cmpLT(start, end) {
		rangeSize := toNum(jsSub(end, start)) * invStep
		for i := 0.0; i <= rangeSize; i++ {
			list = append(list, jsAdd(start, i*step))
		}
	} else {
		rangeSize := toNum(jsSub(start, end)) * invStep
		for i := 0.0; i <= rangeSize; i++ {
			list = append(list, jsSub(start, i*step))
		}
	}
	return newList(list)
}

func (in *Interpreter) visitID(node *ID) any {
	if in.stack.has(node.Name) {
		return in.stack.get(node.Name)
	}
	in.error(fmt.Sprintf("Variable '%s' is not defined\n(%s)", node.Name, in.stack.PathFile))
	return nil
}

func (in *Interpreter) returnCatcher(fn func() any) (result any) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if value, ok := recovered.(*ReturnOp); ok {
				result = value.Node
				return
			}
			panic(recovered)
		}
	}()
	return fn()
}

func (in *Interpreter) scopedReturn(fn func() any) any {
	in.stack.push()
	res := in.returnCatcher(fn)
	in.stack.pop()
	return res
}

func (in *Interpreter) interpret(tree any, dumpAST bool, dumpFile string) any {
	if dumpAST {
		_ = os.WriteFile(dumpFile, []byte(jsonStringify(tree)), 0o644)
	}
	return in.returnCatcher(func() any { return in.visit(tree) })
}

// ─────────────────────────────────────────────────────────────────────────────
// Entry point
// ─────────────────────────────────────────────────────────────────────────────

type JSEtimos struct {
	args ProgramArgs
}

func describeUnknownError(e any) string {
	if err, ok := e.(error); ok {
		return err.Error()
	}
	return fmt.Sprint(e)
}

func (j *JSEtimos) shellMode() {
	interpreter := newInterpreter(extSystem.dirName(), 0)
	globalStack := newStack(extSystem.dirName(), 0)
	importModules[namespaceGlobalCount] = interpreter.stack

	extSystem.shellOnLineCallback = func(line string) {
		func() {
			defer func() {
				if e := recover(); e != nil {
					if langError, ok := e.(*LangError); ok {
						extSystem.print(langError.message)
					} else {
						extSystem.print("[[Unknown error, failed to execute]]\n\n" + describeUnknownError(e))
					}
				}
			}()
			lexer := newLexer(line, extSystem.dirName(), ":shell")
			parser := newParser(lexer, globalStack)
			result := interpreter.interpret(parser.parse(), false, "")
			if !isUndefined(result) {
				extSystem.print(toStr(result))
			}
		}()
		extSystem.shellStdoutWrite(">>> ")
	}

	extSystem.print("JSEtimos v0.0.1 - nani [shell mode] Hello there :)")
	extSystem.shellStdoutWrite(">>> ")
	extSystem.shellLoadInterface()
}

func (j *JSEtimos) runText(text string, fileName string, filePath string) {
	dumpAST := j.args.dumpAST
	dumpASTFile := j.args.dumpFile
	if dumpASTFile == "" {
		dumpASTFile = filePath + ".ast.yml"
	}

	parts := strings.Split(fileName, "/")
	dir := extSystem.dirName() + "/" + strings.Join(parts[:len(parts)-1], "/")

	lexer := newLexer(text, dir, fileName)
	parser := newParser(lexer, newStack(filePath, 0))
	interpreter := newInterpreter(filePath, 0)
	importModules[namespaceGlobalCount] = interpreter.stack

	result := interpreter.interpret(parser.parse(), dumpAST, dumpASTFile)
	if !isUndefined(result) {
		extSystem.print(toStr(result))
	}
}

func (j *JSEtimos) fileMode(fileName string) {
	filePath := extSystem.dirName() + "/" + fileName
	stackReset()
	text := extSystem.readFile(filePath)
	j.runText(text, fileName, filePath)
}

func (j *JSEtimos) inputMode(text string) {
	fileName := "input"
	stackReset()
	j.runText(text, fileName, extSystem.dirName()+"/"+fileName)
}

func (j *JSEtimos) run() {
	defer func() {
		if e := recover(); e != nil {
			if langError, ok := e.(*LangError); ok {
				j.printError(langError)
				return
			}
			fmt.Fprintln(os.Stderr, "Uncaught "+describeUnknownError(e))
			os.Exit(1)
		}
	}()

	if j.args.fileName != "" {
		j.fileMode(j.args.fileName)
	} else if j.args.inputMode {
		j.inputMode(j.args.inputText)
	} else {
		j.shellMode()
	}
}

func (j *JSEtimos) printError(e *LangError) {
	extSystem.print(e.message)
}

func main() {
	etimos := &JSEtimos{extSystem.programArgs()}
	etimos.run()
}
