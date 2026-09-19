// JSEtimos - copyright nani 2021
//
// AST node types (some double as runtime values), their text and JSON forms.

package main

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

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
			parts[i] = toStr(v)
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
		return toStr(v)
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
