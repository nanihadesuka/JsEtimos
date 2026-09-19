// JSEtimos - copyright nani 2021
//
// Tree walking interpreter.

package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
)

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
