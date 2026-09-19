// JSEtimos - copyright nani 2021
//
// Recursive descent parser: tokens to AST, with static checks.

package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

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
		sorted := make([]float64, len(its))
		for i, v := range its {
			if v[2:] == "" {
				sorted[i] = 0
			} else {
				sorted[i] = jsParseFloat(v[2:]) // digits only, same as parseInt
			}
		}
		sort.Float64s(sorted)
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

	// Modules are looked up next to the importing file first,
	// then from the base directory (where the program was started)
	importDir := func(baseDir string) string {
		if relativeDir == "" {
			return baseDir
		}
		return baseDir + "/" + relativeDir
	}
	absoluteDir := importDir(p.lexer.directory)
	if !extSystem.existFile(absoluteDir + "/" + fileName) {
		baseImportDir := importDir(extSystem.dirName())
		if baseImportDir == absoluteDir || !extSystem.existFile(baseImportDir+"/"+fileName) {
			tried := absoluteDir + "/" + fileName
			if baseImportDir != absoluteDir {
				tried += " or " + baseImportDir + "/" + fileName
			}
			p.error(fmt.Sprintf("Import '%s' doesn't exist. Failed to access file in %s.", relativeName, tried), jsLen(relativeName))
		}
		absoluteDir = baseImportDir
	}
	absolutePath := absoluteDir + "/" + fileName

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
