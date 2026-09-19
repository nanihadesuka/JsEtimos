// JSEtimos - copyright nani 2021
//
// Token types and lexer.

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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
