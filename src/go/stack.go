// JSEtimos - copyright nani 2021
//
// Scopes, module registry and the built-in functions.

package main

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

// Updates the nearest scope that holds the name
func (s *Stack) setOnExisting(id string, value any) {
	for i := len(s.stackRecordList) - 1; i > -1; i-- {
		if s.stackRecordList[i].has(id) {
			s.stackRecordList[i].set(id, value)
			return
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
