// JSEtimos - copyright nani 2021

interface ProgramArgs
{
    fileName: string | undefined
    dumpAST: boolean
    dumpFile: string | undefined
    shellMode: boolean
    inputMode: boolean
    inputText: string | undefined
}

interface ExtSystem
{
    dirName(): string
    programArgs(): ProgramArgs

    readFile(filePath: string): string
    writeFile(data: string, file: string, baseDir: string | undefined)
    existFile(filePath: string): boolean

    print(msg: string): void
    prettyPrintToString(msg: any): string

    redirectPrint: boolean
    redirectPrintCallback: ((output: string) => void) | undefined

    formatColor(text: string, color: any): string
    colors: {
        red: any,
    }

    shell_interface: undefined | any
    shell_onLineCallback: (line: string) => void
    shell_loadInterface(): void
    shell_loaded: boolean
    shell_stdoutWrite(text: string)
}

class ExtSystem_NodeJs implements ExtSystem
{
    constructor()
    {
        try
        { require('source-map-support').install() }
        catch (e)
        { this.print("WARNING: 'source-map-support' javascript library not installed (needed for TypeScript stacktraces) ") }
    }

    fs = require('fs')
    util = require("util")

    readFile(filePath: string): string
    {
        return this.fs.readFileSync(filePath, "utf-8")
    }
    writeFile(data: string, file: string, baseDir: string | undefined = undefined)
    {
        baseDir == baseDir ?? this.dirName()
        return this.fs.writeFile(baseDir + "/" + file, data, () => { })
    }
    existFile(filePath: string): boolean
    {
        return this.fs.existsSync(filePath)
    }

    print(msg: string)
    {
        if (this.redirectPrint)
            this.redirectPrintCallback?.(msg)
        else
            console.log(msg)
    }

    redirectPrint = false
    redirectPrintCallback: (output: string) => void = undefined

    prettyPrintToString(msg: any, color = true): string
    {
        return this.util.inspect(msg, { depth: null, colors: color, compact: 2 })
    }
    formatColor(text, color: any): string
    {
        return `${ color }${ text }${ this.colors.reset }`
    }

    colors = {
        reset: "\u001b[0m",
        red: "\u001b[31;1m",
    }

    dirName(): string
    {
        return __dirname
    }

    programArgs(): ProgramArgs
    {
        const list = process.argv.slice(2)
        const attributes: any = {}
        for (let item of list)
        {
            if (item.startsWith("--"))
                attributes[ item.slice(2) ] = item.slice(2)
            else if (item.startsWith("-"))
                continue
            else
                attributes[ "text" ] = item
        }

        const inputMode = !!attributes.inputMode

        return {
            fileName: inputMode ? "" : attributes.text,
            dumpAST: !!attributes.dumpAST,
            dumpFile: attributes.dumpFile,
            shellMode: attributes.shellMode,
            inputMode: inputMode,
            inputText: inputMode ? attributes.text : ""
        }
    }

    shell_interface: undefined | any = undefined
    shell_onLineCallback: (line: string) => void

    shell_loaded = false
    shell_loadInterface()
    {
        if (this.shell_loaded)
            return
        this.shell_loaded = true

        this.shell_interface = require("readline").createInterface({
            input: process.stdin,
            output: process.stdout,
            terminal: false
        })

        this.shell_interface?.on("line", (line: string) => this.shell_onLineCallback(line))
    }
    shell_stdoutWrite(text: string)
    {
        process.stdout.write(text)
    }
}

class ExtSystem_Pyrogenesis0ad implements ExtSystem
{
    constructor()
    { }

    readFile(filePath: string): string
    {
        return ""
    }
    writeFile(data: string, file: string, baseDir: string | undefined = undefined)
    {
        baseDir == baseDir ?? this.dirName()
        return ""
    }
    existFile(filePath: string): boolean
    {
        return false
    }

    print(msg: string)
    {
        if (this.redirectPrint)
            this.redirectPrintCallback?.(msg)
        else
            global.warn(msg)
    }

    redirectPrint = false
    redirectPrintCallback: (output: string) => void = undefined

    prettyPrintToString(msg: any, color = true): string
    {
        return msg.map(v => `${ v }`)
    }
    formatColor(text: string, color: any): string
    {
        return `${ color }${ text }${ this.colors.reset }`
    }

    colors = {
        reset: "",
        red: "",
    }

    dirName(): string
    {
        return "___"
    }

    programArgs(): ProgramArgs
    {
        return {
            fileName: undefined,
            dumpAST: false,
            dumpFile: undefined,
            shellMode: true,
            inputMode: false,
            inputText: ""
        }
    }

    shell_loaded = false
    shell_interface: any = undefined
    shell_onLineCallback: (line: string) => void
    shell_loadInterface()
    {
        if (this.shell_loaded)
            return

        this.shell_loaded = true
        this.shell_interface.onPress = () =>
        {
            const line = this.shell_interface.caption
            this.shell_onLineCallback(line)
        }
    }
    shell_stdoutWrite(text: string)
    {
        global.warn(text)
    }
}

class ExtSystem_Browser implements ExtSystem
{
    constructor()
    { }

    readFile(filePath: string): string
    {
        return ""
    }
    writeFile(data: string, file: string, baseDir: string | undefined = undefined)
    {
        baseDir == baseDir ?? this.dirName()
        return ""
    }
    existFile(filePath: string): boolean
    {
        return false
    }

    print(msg: string)
    {
        if (this.redirectPrint)
            this.redirectPrintCallback?.(msg)
        else
            console.log(msg)
    }

    redirectPrint = false
    redirectPrintCallback: (output: string) => void = undefined

    prettyPrintToString(msg: any, color = true): string
    {
        return msg.map(v => `${ v }`)
    }
    formatColor(text: string, color: any): string
    {
        return `${ color }${ text }${ this.colors.reset }`
    }

    colors = {
        reset: "</span>",
        red: '<span style="color: #cc0000">',
    }

    dirName(): string
    {
        return "___"
    }

    programArgs(): ProgramArgs
    {
        return {
            fileName: undefined,
            dumpAST: false,
            dumpFile: undefined,
            shellMode: false,
            inputMode: true,
            inputText: `print("hello")`
        }
    }

    shell_loaded = false
    shell_interface: any = undefined
    shell_onLineCallback: (line: string) => void
    shell_loadInterface()
    {
        if (this.shell_loaded)
            return

        this.shell_loaded = true
        this.shell_interface.onPress = () =>
        {
            const line = this.shell_interface.caption
            this.shell_onLineCallback(line)
        }
    }
    shell_stdoutWrite(text: string)
    {
        console.log(text)
    }
}

enum TokenType
{
    START_SYMBOL_CHARS = "START_SYMBOL_CHARS",

    PLUSPLUS = "++",
    MINUSMINUS = "--",

    PLUS = "+",
    MINUS = "-",
    MUL = "*",
    DIV = "/",
    POWER = "**",

    SCALAR_PLUS = ".+",
    SCALAR_MINUS = ".-",
    SCALAR_MUL = ".*",
    SCALAR_DIV = "./",

    COMPOUND_PLUS = "+=",
    COMPOUND_MINUS = "-=",
    COMPOUND_MUL = "*=",
    COMPOUND_DIV = "/=",
    COMPOUND_POWER = "**=",

    LESS = "<",
    GREATER = ">",
    EQUAL = "==",
    NOT_EQUAL = "!=",
    LESS_OR_EQUAL = "<=",
    GREATER_OR_EQUAL = ">=",

    LPARE = "(",
    RPARE = ")",
    LBRACKETSQR = "[",
    RBRACKETSQR = "]",
    LBRACKET = "{",
    RBRACKET = "}",

    ASSIGN = "=",
    COMMA = ",",
    DOLLAR = "$",
    DOUBLE_QUOTE = `"`,
    SEMICOLON = ";",
    COLON = ":",
    MODULO = "%",
    PIPE = "|",
    DOT = ".",
    DOT_THREE = "...",
    ARROW_RIGHT = "->",
    QUESTION = "?",

    END_SYMBOL_CHARS = "END_SYMBOL_CHARS",

    FUNCTION = "fun",
    RETURN = "return",
    IF = "if",
    ELIF = "elif",
    ELSE = "else",
    TRUE = "true",
    FALSE = "false",
    AND = "and",
    NOT = "not",
    OR = "or",
    DO = "do",
    FOR = "for",
    AS = "as",
    MAP = "map", // this should be an ID
    FILTER = "filter", // this should be an ID
    IMPORT = "import", // this should be an ID

    END_FIRST_CHAR_IS_ALPHA = "END_FIRST_CHAR_IS_ALPHA",

    REAL = "REAL",
    STRING = "STRING",

    ID = "ID",
    EOL = "EOL",
    EOF = "EOF",

    // END SPECIAL TOKENS

    VAL = "val",
    CONST = "const",
    VAR = "var",
    CONTINUE = "continue",
    BREAK = "break",

    // END UNIMPLEMENTED TOKENS
}

var langChars = {
    SYMBOL_CHAR_TREE: {},
    MULTIPLE_CHAR_FIRST_IS_ALPHA: new Set<string>(),
    RESERVED_KEYWORDS: new Set<string>(),
    isKeywordIt: function (keyword: string): boolean
    {
        return /^it\d*$/.test(keyword)
    },
    isKeywordReserved: function (keyword: string): boolean
    {
        return this.RESERVED_KEYWORDS.has(keyword) ? true : this.isKeywordIt(keyword)
    }
}


{
    const allTokens = Object.values(TokenType)

    const set = (set: Set<string>, start: TokenType, end: TokenType) =>
    {
        const list = allTokens.slice(allTokens.indexOf(start) + 1, allTokens.indexOf(end))
        for (let key of list)
            set.add(key)
    }

    const SYMBOL_CHAR: Set<string> = new Set()
    set(SYMBOL_CHAR, TokenType.START_SYMBOL_CHARS, TokenType.END_SYMBOL_CHARS)
    set(langChars.MULTIPLE_CHAR_FIRST_IS_ALPHA, TokenType.END_SYMBOL_CHARS, TokenType.END_FIRST_CHAR_IS_ALPHA)

    for (let key of SYMBOL_CHAR) langChars.RESERVED_KEYWORDS.add(key)
    for (let key of langChars.MULTIPLE_CHAR_FIRST_IS_ALPHA) langChars.RESERVED_KEYWORDS.add(key)

    for (let key of SYMBOL_CHAR)
    {
        let node = langChars.SYMBOL_CHAR_TREE
        for (let letter of key.split(""))
        {
            if (!(letter in node))
                node[ letter ] = {}

            node = node[ letter ]
        }
    }

    // Token sanity check, typescript allows "enum" with keys with same value
    const tokens = new Set<string>()
    for (let value of Object.values(TokenType))
    {
        if (tokens.has(value))
            throw Error(`Duplicated token '${ value }', check TokenType enum`)
        else
            tokens.add(value)
    }
}

class LangError extends Error
{
    constructor(msg: string)
    {
        super(msg)
        Object.setPrototypeOf(this, new.target.prototype)
    }
}

class LexerError extends LangError { constructor(msg: string) { super(msg) } }
class ParserError extends LangError { constructor(msg: string) { super(msg) } }
class RuntimeError extends LangError { constructor(msg: string) { super(msg) } }

class Token
{
    constructor(public type: TokenType, public value: string) { }
    toString() { return `Token(type='${ this.type }', value='${ this.value }')` }
    is(...tokenTypes: TokenType[]): boolean
    {
        for (let type of tokenTypes)
            if (type == this.type)
                return true
        return false
    }

    copy()
    {
        return new Token(this.type, this.value)
    }
}

class Lexer
{
    absolutePosition: number = -1
    positionInLine: number = -1
    line: number = 0
    character: string = ""

    prev_absolutePosition: number = -1
    prev_positionInLine: number = -1
    prev_line: number = 0
    prev_character: string = ""

    savedState: object = {}
    path: string

    constructor(public text: string, public directory: string, public file: string)
    {
        this.path = `${ directory }/${ file }`
        this.advance()
    }

    getStateCopy()
    {
        return {
            absolutePosition: this.absolutePosition,
            positionInLine: this.positionInLine,
            line: this.line,
            character: this.character,
            prev_absolutePosition: this.prev_absolutePosition,
            prev_positionInLine: this.prev_positionInLine,
            prev_line: this.prev_line,
            prev_character: this.prev_character,
        }
    }

    setState(state)
    {
        for (let prop in state)
            this[ prop ] = state[ prop ]
    }

    formatErrorMessage(msg: string, errorLength: number = 1): string
    {
        return `${ msg }\n\n` +
            this.getCurrentLineTextWithErrorPosition(errorLength) + "\n" +
            `(${ this.path }:${ this.prev_line + 1 }:${ this.prev_positionInLine })\n`
    }

    private formatErrorMessageLexer(msg: string, errorLength: number = 1): string
    {
        return `${ msg }\n\n` +
            this.getCurrentLineTextWithErrorPositionLexer(errorLength) + "\n" +
            `(${ this.path }:${ this.line + 1 }:${ this.positionInLine })\n`
    }

    error(msg: string, errorLength: number = 1)
    {
        const formattedMsg = this.formatErrorMessageLexer(msg, errorLength)
        throw new LexerError(
            extSystem.formatColor("Lexer error", extSystem.colors.red) + "\n" +
            extSystem.formatColor(" ■ ", extSystem.colors.red) + formattedMsg
        )
    }

    getCurrentLineTextWithErrorPosition(errorLength: number = 1): string
    {
        const spaceCount = Math.max(0, this.prev_positionInLine - errorLength)
        const underlineSize = Math.max(1, errorLength)
        const splitText = this.text.split("\n")

        const startLine = Math.max(0, this.prev_line - 7)
        const context = splitText
            .slice(startLine, Math.max(0, this.prev_line + 1))
            .map((t, i) => `${ startLine + i + 1 }`.padEnd(6) + t)
            .join("\n")

        const underLine = extSystem.formatColor("═".repeat(underlineSize), extSystem.colors.red)

        return context + "\n" +
            "".padEnd(6) + " ".repeat(spaceCount) + underLine
    }

    private getCurrentLineTextWithErrorPositionLexer(errorLength: number = 1): string
    {
        const spaceCount = Math.max(0, this.prev_positionInLine - errorLength)
        const underlineSize = Math.max(1, errorLength)
        const splitText = this.text.split("\n")

        const startLine = Math.max(0, this.line - 7)
        const context = splitText
            .slice(startLine, Math.max(0, this.line + 1))
            .map((t, i) => `${ startLine + i + 1 }`.padEnd(6) + t)
            .join("\n")

        const underLine = extSystem.formatColor("═".repeat(underlineSize), extSystem.colors.red)

        return context + "\n" +
            "".padEnd(6) + " " + " ".repeat(spaceCount) + underLine
    }

    advance()
    {
        this.prev_absolutePosition = this.absolutePosition
        this.prev_positionInLine = this.positionInLine
        this.prev_line = this.line
        this.prev_character = this.character

        if (this.prev_character == "\n")
        {
            this.positionInLine = 0
            this.line += 1
        }
        this.absolutePosition += 1
        this.positionInLine += 1
        this.character = this.absolutePosition < this.text.length ? this.text[ this.absolutePosition ] : ""
    }

    peekCharacter(next: number)
    {
        const pos = this.absolutePosition + next
        return this.text[ pos ] ?? ""
    }

    skipWhitespace()
    {
        while (this.character == " " || this.character == "\t" || this.character == "\r")
            this.advance()
    }

    isNewLine(char: string): boolean { return char == "\n" || char == "\r" }

    skipLineComment()
    {
        if (this.character == "/" && this.peekCharacter(1) == "/")
        {
            const isEnd = () => this.isNewLine(this.character) || this.character == ""
            this.advance()
            this.advance()
            while (!isEnd())
                this.advance()
        }
    }

    skipMultilineComment()
    {
        if (this.character == "/" && this.peekCharacter(1) == "*")
        {
            const isEnd = () => (this.character == "*" && this.peekCharacter(1) == "/") || this.character == ""
            this.advance()
            this.advance()
            while (!isEnd())
                this.advance()
            this.advance()
            this.advance()
        }
    }

    isDigit(char: string) { return /\d/.test(char) }
    isAlpha(char: string) { return /[A-Za-z]/.test(char) }
    isAlphaNum(char: string) { return /[A-Za-z0-9_]/.test(char) }
    isWhitespace(char: string) { return /\s/.test(char) }

    getNumber(): string
    {
        let output: string = this.character
        this.advance()

        while (this.isDigit(this.character))
        {
            output += this.character
            this.advance()
        }

        if (this.character === ".")
        {
            output += this.character
            this.advance()
            if (!this.isDigit(this.character))
                this.error(`Invalid character '${ this.character }' in decimal number expression after . (dot), expected at least one digit.`)

            while (this.isDigit(this.character))
            {
                output += this.character
                this.advance()
            }
        }
        return output
    }

    getAlphaNum(): string
    {
        let output: string = this.character
        this.advance()
        while (this.isAlphaNum(this.character))
        {
            output += this.character
            this.advance()
        }
        return output
    }

    /**
     * Looks for entry string characters then
     * entry interpolation or terminal string character.
     */
    getStringStart(): string
    {
        let output: string = this.character
        const startLine = this.line
        const startPositionInLine = this.positionInLine

        while (true)
        {
            this.advance()
            if (this.character == "\\" && this.peekCharacter(1) == "n")
            {
                output += "\n"
                this.advance()
            }
            else
                output += this.character

            if (this.character == TokenType.DOUBLE_QUOTE) break
            if (this.character == "{" && this.peekCharacter(-1) == "$") break
            if (this.character == "") break
        }

        if (this.character == "") this.error(
            `Invalid string expression, ` +
            `reached end of file but didn't close string expression, ` +
            `must end in '"' or '\${'. ` +
            `String starting location at ${ startLine }:${ startPositionInLine }.`
            , 1
        )

        this.advance()

        return output
    }

    /**
     Expects to be called after a TokenType.RBRACKET
     Checks for previous terminal interpolation string character then
     looks for entry interpolation or terminal string character.
     */
    getStringContinuation(): string
    {
        if (this.peekCharacter(-1) != "}")
            this.error("getStringContinuation only can be run after } character.", 1)

        let output: string = ""
        const startLine = this.line
        const startPositionInLine = this.positionInLine

        while (true)
        {
            if (this.character == "\\" && this.peekCharacter(1) == "n")
            {
                output += "\n"
                this.advance()
                this.advance()
                continue
            }

            output += this.character

            if (this.character == TokenType.DOUBLE_QUOTE) break
            if (this.character == "{" && this.peekCharacter(-1) == "$") break
            if (this.character == "") break

            this.advance()
        }

        if (this.character == "") this.error(
            `Invalid string expression, ` +
            `reached end of file but didn't close string expression, ` +
            `must end in '"' or '\${'. ` +
            `String starting location at ${ startLine }:${ startPositionInLine }`
            , 1
        )

        this.advance()
        return output
    }

    peekToken(count: number = 1)
    {
        const state = this.getStateCopy()
        for (let i = 0; i < count - 1; i++)
            this.getNextToken()
        const token = this.getNextToken()
        this.setState(state)
        return token
    }

    // Called by the parser after string interpolation } token
    getNextTokenStringContinuation(): Token
    {
        return new Token(TokenType.STRING, this.getStringContinuation())
    }

    symbolSeek()
    {
        let symbol = ""
        let node = langChars.SYMBOL_CHAR_TREE
        while (this.character in node)
        {
            node = node[ this.character ]
            symbol += this.character
            this.advance()
        }
        return new Token(symbol as TokenType, symbol)
    }

    getNextToken(): Token
    {
        this.skipWhitespace()
        this.skipLineComment()
        this.skipWhitespace()
        this.skipMultilineComment()
        this.skipWhitespace()

        if (this.character === "\n") { this.advance(); return new Token(TokenType.EOL, TokenType.EOL) }
        if (this.character === "") { this.advance(); return new Token(TokenType.EOF, TokenType.EOF) }

        if (this.character == TokenType.DOUBLE_QUOTE)
            return new Token(TokenType.STRING, this.getStringStart())

        if (this.character in langChars.SYMBOL_CHAR_TREE)
            return this.symbolSeek()

        if (this.isDigit(this.character)) return new Token(TokenType.REAL, this.getNumber())
        if (this.isAlpha(this.character))
        {
            const id = this.getAlphaNum()

            if (langChars.MULTIPLE_CHAR_FIRST_IS_ALPHA.has(id))
                return new Token(id as TokenType, id)
            else
                return new Token(TokenType.ID, id)
        }

        this.error(`Invalid character: '${ this.character }'.`)
    }
}

class ASTNode
{
    toString()
    {
        return "\n" + this.constructor.name + " " + JSON.stringify(this, null, 2)
    }
}

class NoOp extends ASTNode { }
class ID extends ASTNode { constructor(public op: Token, public name: string = op.value) { super() } }

// They are shims, in reality they are converted to javascript:
// string, number, boolean, undefined, null types
class StringOp extends ASTNode { constructor(public strings: string[], public interpolations: ASTNode[]) { super() } }
class Real extends ASTNode { constructor(public op: Token) { super() } }
class Bool extends ASTNode { constructor(public op: Token) { super() } }
class Undefined extends ASTNode { }
class Null extends ASTNode { }

class List extends ASTNode
{
    constructor(public nodes: ASTNode[]) { super() }

    public toString()
    {
        return this.toStringIdent(0)
    }

    public toStringIdent(ident: number)
    {
        const repr = (v: ASTNode, disp: number) => v instanceof Dictionary ?
            v.toStringIdent(disp) :
            v instanceof List ?
                "(" + v.toStringIdent(disp) + ")" :
                v.toString()

        return this.nodes.map(v => repr(v, ident + 2)).join(", ")
    }

    public sumScalar(scalar: number) { return new List(this.nodes.map(v => (v as number) + scalar)) }
    public restScalar(scalar: number) { return new List(this.nodes.map(v => (v as number) - scalar)) }
    public multScalar(scalar: number) { return new List(this.nodes.map(v => (v as number) * scalar)) }
    public divScalar(scalar: number) { return new List(this.nodes.map(v => (v as number) / scalar)) }
    public modScalar(scalar: number) { return new List(this.nodes.map(v => (v as number) % scalar)) }
    public powScalar(scalar: number) { return new List(this.nodes.map(v => Math.pow(v as number, scalar))) }

    public restScalarRL(scalarL: number) { return new List(this.nodes.map(v => scalarL - (v as number))) }
    public divScalarRL(scalarL: number) { return new List(this.nodes.map(v => scalarL / (v as number))) }
    public modScalarRL(scalarL: number) { return new List(this.nodes.map(v => scalarL % (v as number))) }
    public powScalarRL(scalarL: number) { return new List(this.nodes.map(v => Math.pow(scalarL, v as number))) }

    public sumList(vector: List) { return new List(this.nodes.map((v, i) => (v as number) + (vector.nodes[ i ] as number))) }
    public restList(vector: List) { return new List(this.nodes.map((v, i) => (v as number) - (vector.nodes[ i ] as number))) }
    public multList(vector: List) { return new List(this.nodes.map((v, i) => (v as number) * (vector.nodes[ i ] as number))) }
    public divList(vector: List) { return new List(this.nodes.map((v, i) => (v as number) / (vector.nodes[ i ] as number))) }
    public modList(vector: List) { return new List(this.nodes.map((v, i) => (v as number) % (vector.nodes[ i ] as number))) }
    public powList(vector: List) { return new List(this.nodes.map((v, i) => Math.pow(v as number, vector.nodes[ i ] as number))) }
}

// Used so when the List is visited it doesn't cause an infinite loop
class ListEval extends List { }

class Dictionary extends ASTNode
{
    constructor(public nodes: {}) { super() }

    public toString()
    {
        return this.toStringIdent(0)
    }

    public toStringIdent(ident: number)
    {
        const list = Object.keys(this.nodes)
        const first = ident == 0
        const disp = 2

        const repr = (v: ASTNode, disp: number) => v instanceof Dictionary ?
            v.toStringIdent(disp) :
            v instanceof List ?
                v.toStringIdent(disp) :
                v.toString()

        if (list.length <= 0)
            return "{ " +
                list.map(k => `${ k } = ${ repr(this.nodes[ k ], ident + 2) }`).join() +
                " }"

        const space = " ".repeat(ident)
        const space2 = " ".repeat(ident + disp)

        return "{\n" +
            list.map(k => space2 + `${ k } = ${ repr(this.nodes[ k ], ident + 2) }`).join("\n") + "\n" +
            (first ? "" : space) + "}"
    }
}

class BlockOp extends ASTNode { constructor(public list: ASTNode[]) { super() } }
class UnaryOp extends ASTNode { constructor(public op: Token, public right: ASTNode) { super() } }
class BinaryOp extends ASTNode { constructor(public left: ASTNode, public op: Token, public right: ASTNode) { super() } }
class ShortcircuitOp extends ASTNode { constructor(public left: ASTNode, public op: Token, public right: ASTNode) { super() } }
class ComparisonOp extends ASTNode { constructor(public nodes: ASTNode[], public comparators: Token[]) { super() } }
class CompoundOp extends ASTNode { constructor(public left: ID, public op: Token, public right: ASTNode) { super() } }
class VariableIncrementPrefixOp extends ASTNode { constructor(public op: Token, public right: ID) { super() } }
class VariableIncrementPostfixOp extends ASTNode { constructor(public op: Token, public left: ID) { super() } }
class SpreadListOp extends ASTNode { constructor(public list: ASTNode) { super() } }

class AssignOp extends ASTNode { constructor(public id: ID, public expression: ASTNode) { super() } }
class MultipleAssignOp extends ASTNode { constructor(public ids: ID[], public expression: ASTNode) { super() } }
class PipeNodeOp extends ASTNode { constructor(public expression: FunctionCall | AnonymousFunction | ASTNode, public passType: TokenType | null) { super() } }
class PipeOp extends ASTNode { constructor(public entry: ASTNode, public nodes: PipeNodeOp[]) { super() } }
class RangeOp extends ASTNode { constructor(public start: ASTNode, public step: ASTNode, public end: ASTNode,) { super() } }
class ForLoopOp extends ASTNode { constructor(public setup: ASTNode, public runCondition: ASTNode, public increment: ASTNode, public body: ASTNode) { super() } }
class Conditional extends ASTNode { constructor(public conditions: ASTNode[], public bodies: ASTNode[]) { super() } }
class ReturnOp extends ASTNode { constructor(public node: ASTNode) { super() } }

abstract class BaseFunction extends ASTNode { constructor(public stackNamespace: number, public name: string, public parameters: ID[]) { super() } }
class NamedFunction extends BaseFunction { constructor(public stackNamespace: number, public name: string, public parameters: ID[], public block: ASTNode) { super(stackNamespace, name, parameters) } }
class AnonymousFunction extends BaseFunction { constructor(public stackNamespace: number, public parameters: ID[], public block: ASTNode) { super(stackNamespace, "", parameters) } }
class NativeFunction extends BaseFunction { constructor(public stackNamespace: number, public name: string, public parameters: ID[], public fun: (...args: any) => any) { super(stackNamespace, name, parameters) } }

class FunctionCall extends ASTNode { constructor(public name: string, public args: List) { super() } }
class PrintselfCall extends ASTNode { constructor(public args: List, public text: string) { super() } }
class AssertCall extends ASTNode { constructor(public args: List, public text: string, public file: string, public line: number, public column: number) { super() } }
class CallOp extends ASTNode { constructor(public func: ASTNode, public args: ASTNode) { super() } }

class AccessIndexOp extends ASTNode { constructor(public list: ASTNode, public index: ASTNode) { super() } }
class AccessKeyOp extends ASTNode { constructor(public object: ASTNode, public key: string) { super() } }
class AccessMethodOp extends ASTNode { constructor(public object: ASTNode, public key: string) { super() } }

class NamedImportOp extends ASTNode { constructor(public moduleAST: ASTNode, public moduleSymbol: number, public moduleAlias: string, public filePath: string) { super() } }
class NamedImportLoadOp extends ASTNode { constructor(public moduleSymbol: number, public moduleAlias: string) { super() } }

function writeGlobalUtils(stack: Stack)
{
    const _print = new NativeFunction(
        stack.symbol,
        "print",
        [ "msg" ].map(name => new ID(new Token(TokenType.ID, name))),
        (msg: string) => { extSystem.print(`${ msg }`) }
    )

    const printself = new NativeFunction(
        stack.symbol,
        "printself",
        [ "msg" ].map(name => new ID(new Token(TokenType.ID, name))),
        (msg: string) => { extSystem.print(`${ msg }`) }
    )

    const assert = new NativeFunction(
        stack.symbol,
        "assert",
        [ "msg" ].map(name => new ID(new Token(TokenType.ID, name))),
        (msg: string) => { extSystem.print(`${ msg }`) }
    )

    const _printToFile = new NativeFunction(
        stack.symbol,
        "printToFile",
        [ "data", "file" ].map(name => new ID(new Token(TokenType.ID, name))),
        (data: any, file: string) => { extSystem.writeFile(`${ data }`, file, undefined) }
    )

    stack.set(_print.name, _print)
    stack.set(printself.name, printself)
    stack.set(assert.name, assert)
    stack.set(_printToFile.name, _printToFile)
}

class StackRecord
{
    memory: Map<string, any> = new Map()
    get(id: string): any { return this.memory.get(id) }
    set(id: string, value: any) { this.memory.set(id, value) }
    has(id: string): boolean { return this.memory.has(id) }
}

class Stack
{
    // Modified on the parser stage only (when loading a new module)
    public static namespaceGlobalCount: number = 0
    public static absoluteFilePath_to_symbol = new Map<string, number>()

    // Modified on the parse(for checking) and interpreter stage (overrites parse data) only
    public static importModules = new Map<number, Stack>()

    public static reset()
    {
        Stack.namespaceGlobalCount = 0
        Stack.absoluteFilePath_to_symbol = new Map<string, number>()
        Stack.importModules = new Map<number, Stack>()
    }

    private stackRecordList: StackRecord[] = []

    activeRecord: StackRecord

    constructor(public pathFile: string, public symbol: number)
    {
        this.push()
        writeGlobalUtils(this)
    }

    isGlobalScope() { return this.stackRecordList.length == 1 }


    getGlobalVariablesKeys(): string[]
    {
        return [ ...this.stackRecordList[ 0 ].memory.keys() ]
    }

    push()
    {
        const stackRecord = new StackRecord()
        this.stackRecordList.push(stackRecord)
        this.activeRecord = stackRecord
    }

    pop()
    {
        this.stackRecordList.pop()
        this.activeRecord = this.stackRecordList[ this.stackRecordList.length - 1 ]
    }

    get(id: string): any | NamedFunction | NativeFunction
    {
        for (let i = this.stackRecordList.length - 1; i > -1; i--)
            if (this.stackRecordList[ i ].has(id))
                return this.stackRecordList[ i ].get(id)
    }

    has(id: string): boolean
    {
        for (let i = this.stackRecordList.length - 1; i > -1; i--)
            if (this.stackRecordList[ i ].has(id))
                return true

        return false
    }

    hasInSameRecord(id: string): boolean
    {
        return this.activeRecord.has(id)
    }

    set(id: string, value: any | NamedFunction | NativeFunction)
    {
        this.activeRecord.set(id, value)
    }

    setOnExisting(id: string, value: any | NamedFunction | NativeFunction)
    {
        for (let i = this.stackRecordList.length - 1; i > -1; i--)
            if (this.stackRecordList[ i ].has(id))
                this.stackRecordList[ i ].set(id, value)
    }

    hasOwn(id: string): boolean
    {
        return this.activeRecord.has(id)
    }

    getOwnIts(): string[]
    {
        return [ ...this.activeRecord.memory.keys() ].filter(v => langChars.isKeywordIt(v))
    }
}

/* Extra grammar notation
   · [TOKEN] The token is not to be consumed (not "eat" call) in the currently parsing call
   · {call} The call is to be run in a new child stack scope (push > call > pop) in the currently running call
   · func_a< func_b > Signifies the composition of functions, func_a( () => func_b() )
   · $var Make reference to a local varaible o function
*/

enum EatMode
{
    Default,
    StringContinuation,
}

class Parser
{
    private token: Token

    // Doesn't save stack !!!
    private getState()
    {
        return {
            token: this.token.copy(),
            lexerState: this.lexer.getStateCopy()
        }
    }

    // Doesn't restore stack !!!
    private restoreState(savedState: any)
    {
        this.token.type = savedState.token.type
        this.token.value = savedState.token.value
        this.lexer.setState(savedState.lexerState)
    }

    peekToken(count: number = 1): Token
    {
        return this.lexer.peekToken(count)
    }

    is(...tokenTypes: TokenType[]): boolean
    {
        for (let type of tokenTypes)
            if (type == this.token.type)
                return true
        return false
    }

    isChain(...tokenTypes: TokenType[]): boolean
    {
        if (!this.is(tokenTypes[ 0 ]))
            return false

        for (let i = 1; i < tokenTypes.length; i++)
            if (tokenTypes[ i ] !== this.peekToken(i).type)
                return false

        return true
    }

    constructor(public lexer: Lexer, public stack: Stack)
    {
        this.token = this.lexer.getNextToken()
    }

    error(msg: string, errorLength: number = 1)
    {
        const formattedMsg = this.lexer.formatErrorMessage(msg, errorLength)
        throw new ParserError(
            extSystem.formatColor("Parser error", extSystem.colors.red) + "\n" +
            extSystem.formatColor(" ■ ", extSystem.colors.red) + formattedMsg
        )
    }

    eat(tokenType: TokenType, mode: EatMode = EatMode.Default)
    {
        if (this.is(tokenType))
        {
            if (mode == EatMode.Default)
                this.token = this.lexer.getNextToken()
            else if (mode == EatMode.StringContinuation)
                this.token = this.lexer.getNextTokenStringContinuation()
            else
                this.error(`Invalid EatMode ${ mode }`)
        }
        else
            this.error(`Invalid token ${ this.token }, expected token type '${ tokenType }'.`, this.token.value.length)

        // log("EAT", tokenType.padEnd(5), "GET", this.token, "MODE ", mode)
    }

    eatEOL(optional: boolean = true)
    {
        if (!optional)
            this.eat(TokenType.EOL)
        else if (this.is(TokenType.EOL))
            this.eat(TokenType.EOL)
    }

    eatEOLS()
    {
        while (this.is(TokenType.EOL))
            this.eat(TokenType.EOL)
    }

    anonymousFunctionParameters(): ID[]
    {
        // (ID (COMMA ID)* ARROW_RIGHT)?
        let parameters: ID[] = []
        let ids: Set<string> = new Set()

        if (this.is(TokenType.ID))
        {
            const parameter = new ID(this.token)
            this.eat(TokenType.ID)
            parameters.push(parameter)

            ids.add(parameter.name)

            while (this.is(TokenType.COMMA))
            {
                this.eat(TokenType.COMMA)
                const parameter = new ID(this.token)
                if (ids.has(parameter.name))
                    this.error(`Duplicate identifier '${ parameter.name }'.`, parameter.name.length)
                this.eat(TokenType.ID)
                parameters.push(parameter)
            }

            this.eat(TokenType.ARROW_RIGHT)
        }

        return parameters
    }


    hasAnonymousFunctionParameters(): Boolean
    {
        // (ID (COMMA ID)* ARROW_RIGHT)?


        if (!this.is(TokenType.ID))
            return false

        const initialState = this.getState()
        this.eat(TokenType.ID)
        while (this.is(TokenType.COMMA))
        {
            this.eat(TokenType.COMMA)
            if (!this.is(TokenType.ID))
                break
            this.eat(TokenType.ID)
        }

        const hasParameters = this.is(TokenType.ARROW_RIGHT)
        this.restoreState(initialState)
        return hasParameters
    }

    anonymousFunction(noParams: boolean = false): AnonymousFunction
    {
        // {LBRACKET anonymousFunctionParameters blockTo  RBRACKET}
        this.stack.push()
        this.eat(TokenType.LBRACKET)

        let hasParams = this.hasAnonymousFunctionParameters()
        if (noParams && hasParams)
            this.error(`This anonymous function can't have parameters.`)

        let parameters = hasParams ? this.anonymousFunctionParameters() : []

        for (let parameter of parameters)
            this.stack.set(parameter.name, undefined)
        const block = this.blockTo(TokenType.RBRACKET)
        const its = this.stack.getOwnIts()
        if (parameters.length && its.length)
            this.error(`Not allowed to have explicit parameters and also use special 'it' variables ${ its }.`)

        if (its.length)
        {
            const sorted = its.map(v => v.slice(2)).map(v => v == "" ? 0 : parseInt(v)).sort()
            const nOfIts = sorted[ sorted.length - 1 ] + 1
            parameters = new Array(nOfIts).fill(0).
                map((_, i) => i == 0 ? "it" : `it${ i }`).
                map((v) => new ID(new Token(TokenType.ID, v)))
        }

        this.stack.pop()
        return new AnonymousFunction(this.stack.isGlobalScope() ? this.stack.symbol : -1, parameters, block)
    }

    pipeExpression(): PipeNodeOp
    {
        // (MAP|FILTER)? ([LBRACKET] anonymousFunction  | ID access)

        const passType: TokenType = this.is(TokenType.MAP, TokenType.FILTER) ? this.token.type : null
        if (passType != null) this.eat(passType)

        if (this.is(TokenType.LBRACKET))
            return new PipeNodeOp(this.anonymousFunction(), passType)

        if (this.is(TokenType.ID))
        {
            const id = new ID(this.token)
            if (!this.stack.has(id.name))
                throw this.error(`Function '${ id.name }' not defined.`, id.name.length)

            const fun = this.stack.get(id.name)

            // Dynamic (error if any will be in the runtime stage)
            if (fun instanceof Dictionary || fun instanceof List || fun instanceof Stack)
            {
                const dynamicFun = this.accessor(1)
                return new PipeNodeOp(dynamicFun, passType)
            }

            if (!(fun instanceof BaseFunction))
                throw this.error(`Identifier '${ id.name }' is not a function.`, id.name.length)

            this.eat(TokenType.ID)
            return new PipeNodeOp(new FunctionCall(fun.name, new List([])), passType)
        }

        if (this.is(TokenType.LBRACKET))
            this.error(`Anonymous function bracket must always start in the same line as the pipe operator.`, this.token.value.length)

        this.error(`Invalid token ${ this.token } for pipe node expression.`, this.token.value.length)
    }

    // Pipe horizontal travserse with optional accessors
    expression(nVals: number = undefined): ASTNode
    {
        // listExpression (PIPE EOL? ( isAccessor getAccessor | pipeExpression))*

        let expression = this.listExpression(nVals)

        if (!this.is(TokenType.PIPE))
            return expression

        while (this.is(TokenType.PIPE))
        {
            const pipeNodes: PipeNodeOp[] = []
            while (this.is(TokenType.PIPE))
            {
                this.eat(TokenType.PIPE)
                this.eatEOL()
                if (this.isAccessor())
                    break
                pipeNodes.push(this.pipeExpression())
            }

            if (pipeNodes.length == 0)
                expression = this.getAccessors(expression)
            else
                expression = this.getAccessors(new PipeOp(expression, pipeNodes))
        }

        return expression
    }

    // An expression can return multiple values
    listExpression(nVals: number = undefined): ASTNode
    {
        // DOT_THREE? disjunction (COMMA (EOL? disjunction))*
        const expects: boolean = nVals !== undefined

        const expandList = false // Buggy don't
        // const expandList = this.is(TokenType.DOT_THREE)
        // if (expandList) this.eat(TokenType.DOT_THREE)

        let nodes: ASTNode[] = []
        nodes.push(this.disjunction(nVals))

        let makeList = false
        while (this.is(TokenType.COMMA))
        {
            if (expects && nodes.length >= nVals)
                this.error(`Expected ${ nVals } expressions but got ${ nodes.length + 1 }. ${ this.token }`, this.token.value.length)

            if (this.isChain(TokenType.COMMA, TokenType.RPARE))
            {
                this.eat(TokenType.COMMA)
                makeList = true
                break
            }
            this.eat(TokenType.COMMA)
            this.eatEOL()
            const node = this.disjunction(nVals)
            nodes.push(node)
        }

        if (expects && nodes.length < nVals)
            this.error(`Expected ${ nVals } expressions but got ${ nodes.length }. ${ this.token }`, this.token.value.length)

        let res = nodes.length == 1 ? nodes[ 0 ] : new List(nodes)
        if (makeList) res = new List([ res ])
        if (expandList) res = new SpreadListOp(res)

        return res
    }

    disjunction(nVals: number = undefined): ASTNode
    {
        // conjunction (OR EOL? conjunction)*
        let node = this.conjunction(nVals)
        while (this.is(TokenType.OR))
        {
            const token = this.token
            this.eat(token.type)
            this.eatEOL()
            node = new ShortcircuitOp(node, token, this.conjunction(nVals))
        }
        return node
    }

    conjunction(nVals: number = undefined): ASTNode
    {
        // comparision ( AND EOL? comparision)*
        let node = this.comparision(nVals)
        while (this.is(TokenType.AND))
        {
            const token = this.token
            this.eat(token.type)
            this.eatEOL()
            node = new ShortcircuitOp(node, token, this.comparision(nVals))
        }
        return node
    }

    comparision(nVals: number = undefined): ASTNode
    {
        // additive (( < | > | <= | >= | == | != ) EOL? additive)*
        let nodes: ASTNode[] = []
        let comparators: Token[] = []
        nodes.push(this.additive(nVals))

        while (this.is(
            TokenType.LESS,
            TokenType.GREATER,
            TokenType.EQUAL,
            TokenType.NOT_EQUAL,
            TokenType.LESS_OR_EQUAL,
            TokenType.GREATER_OR_EQUAL
        ))
        {
            const token = this.token
            this.eat(token.type)
            this.eatEOL()
            comparators.push(token)
            nodes.push(this.additive(nVals))
        }

        if (nodes.length == 1)
            return nodes[ 0 ]

        return new ComparisonOp(nodes, comparators)
    }

    additive(nVals: number = undefined): ASTNode
    {
        // multiplicative ((PLUS|MINUS) EOL? multiplicative)*
        let node = this.multiplicative(nVals)
        while (this.is(TokenType.PLUS, TokenType.MINUS))
        {
            const token = this.token
            this.eat(token.type)
            this.eatEOL()
            node = new BinaryOp(node, token, this.multiplicative(nVals))
        }
        return node
    }

    multiplicative(nVals: number = undefined): ASTNode
    {
        // power ((MUL|DIV|MODULO) EOL? power)*
        let node = this.power(nVals)
        while (this.is(TokenType.MUL, TokenType.DIV, TokenType.MODULO))
        {
            const token = this.token
            this.eat(token.type)
            this.eatEOL()
            node = new BinaryOp(node, token, this.power(nVals))
        }
        return node
    }

    power(nVals: number = undefined): ASTNode
    {
        // accessor (POWER accessor)*
        let node = this.accessor(nVals)
        while (this.is(TokenType.POWER))
        {
            const token = this.token
            this.eat(token.type)
            this.eatEOL()
            node = new BinaryOp(node, token, this.accessor(nVals))
        }
        return node
    }

    isAccessor(): boolean
    {
        return this.is(TokenType.DOT, TokenType.LBRACKETSQR, TokenType.LPARE, TokenType.QUESTION)
    }

    getAccessors(node: ASTNode): ASTNode
    {
        /*
            (
                DOT (ID|REAL) |
                LBRACKETSQR EOL? expression EOL? RBRACKETSQR |
                LPARE EOLS? (expression EOLS?)? RPARE |
                QUESTION ID (parameters)?
            )*
        */
        while (this.isAccessor())
        {
            if (this.is(TokenType.LBRACKETSQR))
            {
                this.eat(TokenType.LBRACKETSQR)
                this.eatEOL()
                node = new AccessIndexOp(node, this.expression())
                this.eatEOL()
                this.eat(TokenType.RBRACKETSQR)
            }
            else if (this.is(TokenType.DOT))
            {
                this.eat(TokenType.DOT)
                const key = this.token.value
                if (!this.is(TokenType.ID, TokenType.REAL))
                    this.error("Invalid token, access key needs to be a string or integer number.")
                if (this.is(TokenType.REAL) && key.includes("."))
                    this.error("Access key invalid: can't be a number with decimals.")
                this.eat(this.token.type)
                node = new AccessKeyOp(node, key)
            }
            else if (this.is(TokenType.LPARE))
            {
                this.eat(TokenType.LPARE)
                this.eatEOLS()
                const isEmpty = this.is(TokenType.RPARE)
                const parameters = isEmpty ? new List([]) : this.expression()
                this.eatEOLS()
                this.eat(TokenType.RPARE)
                node = new CallOp(node, parameters)
            }
            else if (this.is(TokenType.QUESTION))
            {
                this.eat(TokenType.QUESTION)
                const id = new ID(this.token)
                this.eat(TokenType.ID)
                node = new AccessMethodOp(node, id.name)
            }
        }
        return node
    }

    accessor(nVals: number = undefined): ASTNode
    {
        /* range getAccessors
        */
        let node = this.range(nVals)
        return this.getAccessors(node)
    }

    range(nVals: number = undefined): ASTNode
    {
        // factor (: factor (: factor)?)?

        const node1 = this.factor(nVals)
        if (!this.is(TokenType.COLON))
            return node1

        this.eat(TokenType.COLON)
        const node2 = this.factor()
        if (!this.is(TokenType.COLON))
            return new RangeOp(node1, null, node2)

        this.eat(TokenType.COLON)
        const node3 = this.factor()
        return new RangeOp(node1, node2, node3)
    }

    factor(nVals: number = undefined): ASTNode
    {
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
        if (this.is(TokenType.PLUSPLUS, TokenType.MINUSMINUS))
        {
            const token = this.token
            this.eat(token.type)
            const id = new ID(this.token)
            this.eat(TokenType.ID)
            return new VariableIncrementPrefixOp(token, id)
        }
        if (this.is(TokenType.PLUS, TokenType.MINUS))
        {
            const token = this.token
            this.eat(token.type)
            this.eatEOL()
            return new UnaryOp(token, this.factor(nVals))
        }
        if (this.is(TokenType.PLUS, TokenType.MINUS))
        {
            const token = this.token
            this.eat(token.type)
            this.eatEOL()
            return new UnaryOp(token, this.factor(nVals))
        }
        if (this.is(TokenType.NOT))
        {
            const token = this.token
            this.eat(token.type)
            return new UnaryOp(token, this.factor(nVals))
        }
        else if (this.is(TokenType.REAL))
        {
            const token = this.token
            this.eat(TokenType.REAL)
            return new Real(token)
        }
        else if (this.is(TokenType.TRUE, TokenType.FALSE))
        {
            const token = this.token
            this.eat(token.type)
            return new Bool(token)
        }
        else if (this.is(TokenType.ID))
        {
            const id = new ID(this.token)

            if (id.name == "undefined")
            {
                this.eat(TokenType.ID)
                return new Undefined()
            }

            if (id.name == "null")
            {
                this.eat(TokenType.ID)
                return new Null()
            }

            const isIt = langChars.isKeywordIt(id.name)
            const isArguments = id.name == "arguments"
            // Special family of variables it, it1, it2, it3 , etc
            // that represent the scope arguments for anonymous functions.
            // They can't be reassigned or modified, only accessed

            if (id.name == "it0")
                this.error(`First 'it' value doesn't use index, try renaming 'it0' to 'it'.`, id.name.length)

            if (isIt || isArguments)
            {
                if (this.stack.isGlobalScope())
                    this.error(`Not allowed to access special variable '${ id.name }' in the global scope.`, id.name.length)

                // Here, the name is "marked" so can be later known will be accessed
                this.stack.set(id.name, null)
            }

            if (!this.stack.has(id.name))
                this.error(`Trying to use undefined variable '${ id.name }'.`, id.name.length)

            this.eat(TokenType.ID)

            if (this.is(TokenType.LPARE))
            {
                if (isIt)
                    this.error(`Trying to use the variable '${ id.name }' as a function.`, 1)

                return this.functionCall(id)
            }
            else if (this.is(TokenType.PLUSPLUS, TokenType.MINUSMINUS))
            {
                if (isIt)
                    this.error(`Not allowed to modify special variable '${ id.name }', only allowed to access.`, id.name.length)

                const token = this.token
                this.eat(token.type)
                return new VariableIncrementPostfixOp(token, id)
            }
            else if (this.is(TokenType.LBRACKETSQR))
            {
                this.eat(TokenType.LBRACKETSQR)
                this.eatEOLS()
                const node = this.expression()
                this.eatEOLS()
                this.eat(TokenType.RBRACKETSQR)
                return new AccessIndexOp(id, node)
            }
            else
                return id
        }
        else if (this.is(TokenType.LPARE))
        {
            this.eat(TokenType.LPARE)
            this.eatEOL()
            let node = this.expression()
            this.eatEOL()
            this.eat(TokenType.RPARE)
            return node
        }
        else if (this.is(TokenType.IF))
        {
            return this.conditionalExpression(nVals)
        }
        else if (this.is(TokenType.RETURN))
        {
            return this.return(nVals)
        }
        else if (this.is(TokenType.STRING))
        {
            return this.string()
        }
        else if (this.is(TokenType.DO))
        {
            this.eat(TokenType.DO)
            return this.anonymousFunction(true)
        }
        else if (this.is(TokenType.LBRACKET))
        {
            return this.dictionary()
        }

        this.error(`Expected expression here but found none. ${ this.token }.`, 1)
    }

    dictionary(): ASTNode
    {
        // LBRACKET EOL? ( (ID|REAL) ASSIGN ( fun anonymousFunction | expression) (EOL|SEMICOLON)? EOLS? )* EOL? RBRACKET
        this.eat(TokenType.LBRACKET)
        this.eatEOL()
        const object = {}
        const ids = new Set<string>()
        while (this.is(TokenType.ID, TokenType.REAL))
        {
            if (this.is(TokenType.REAL) && this.token.value.includes("."))
                this.error("No decimal numbers allowed as dictionary keys.")

            const id = `${ this.token.value }`
            if (ids.has(id))
                this.error(`Duplicated key ${ id }.`)
            else
                ids.add(id)

            this.eat(this.token.type)
            this.eat(TokenType.ASSIGN)

            const isFun = this.is(TokenType.FUNCTION)
            if (isFun) this.eat(TokenType.FUNCTION)

            const node = isFun ? this.anonymousFunction() : this.expression()

            if (this.is(TokenType.RBRACKET))
            {
                object[ id ] = node
                break
            }

            if (this.is(TokenType.ID, TokenType.REAL))
            {
                object[ id ] = node
                continue
            }

            if (this.is(TokenType.SEMICOLON))
                this.eat(TokenType.SEMICOLON)
            else
                this.eat(TokenType.EOL)

            this.eatEOLS()
            object[ id ] = node
        }
        this.eatEOL()
        this.eat(TokenType.RBRACKET)

        return new Dictionary(object)
    }

    import(): ASTNode
    {
        // IMPORT ID ( DOT ID )* (AS ID)? EOL
        if (!this.stack.isGlobalScope())
            this.error(`Import can only be called in the global scope.`, this.token.value.length)

        this.eat(TokenType.IMPORT)

        const ids = this.doWhileEat(TokenType.DOT, () =>
        {
            const token = this.token
            this.eat(TokenType.ID)
            return new ID(token)
        })

        const list = ids.map(v => v.name)
        const relativeName = list.join(".")
        const relativeDir = list.slice(0, -1).join("/")
        const name = list[ list.length - 1 ]
        const fileName = name + ".jk"

        const absoluteDir = this.lexer.directory + (relativeDir == "" ? "" : "/" + relativeDir)
        const absolutePath = absoluteDir + "/" + fileName

        if (!extSystem.existFile(absolutePath))
            this.error(`Import '${ relativeName }' doesn't exist. Failed to access file in ${ absolutePath }.`, relativeName.length)

        let moduleAlias = name
        if (this.is(TokenType.AS))
        {
            this.eat(TokenType.AS)
            moduleAlias = this.token.value
            this.eat(TokenType.ID)
            this.eat(TokenType.EOL)
        }

        if (this.stack.has(moduleAlias))
            this.error(`Trying to use duplicated name ${ moduleAlias }.`, moduleAlias.length)

        // Module AST tree and stack are already parsed, so just add the name
        if (Stack.absoluteFilePath_to_symbol.has(absolutePath))
        {
            const moduleSymbol = Stack.absoluteFilePath_to_symbol.get(absolutePath)
            const moduleStack = Stack.importModules.get(moduleSymbol)
            this.stack.set(moduleAlias, moduleStack)
            return new NamedImportLoadOp(moduleSymbol, moduleAlias)
        }

        const moduleSymbol = ++Stack.namespaceGlobalCount
        Stack.absoluteFilePath_to_symbol.set(absolutePath, moduleSymbol)
        const moduleStack = new Stack(absolutePath, moduleSymbol)

        const rawText = extSystem.readFile(absolutePath)
        const lexer = new Lexer(rawText, absoluteDir, fileName)
        const parser = new Parser(lexer, moduleStack)
        const module_ast = parser.parse()
        this.stack.set(moduleAlias, moduleStack)
        Stack.importModules.set(moduleSymbol, parser.stack)
        return new NamedImportOp(module_ast, moduleSymbol, moduleAlias, absolutePath)
    }

    // Eats the token given !!
    doWhileEat<T extends ASTNode>(tokenType: TokenType, fn: () => T): T[]
    {
        const list: T[] = [ fn() ]
        while (this.is(tokenType))
        {
            this.eat(tokenType)
            list.push(fn())
        }
        return list
    }

    string(): StringOp
    {
        // STRING ( EOLS? expression EOLS? RBRACKET STRING )*

        // string syntax example: "hello ${ callA() } ${ 2*1 } world"

        const extract = (raw: string) =>
        {
            // no terminal: ${  terminal: "
            const isTerminal = raw[ raw.length - 1 ] == '"'
            const start = raw[ 0 ] == `"` ? 1 : 0
            const end = raw.length - (isTerminal ? 1 : 2)
            return {
                isTerminal: isTerminal,
                string: raw.slice(start, end)
            }
        }

        const listStrings: string[] = []
        const listExpressions: ASTNode[] = []

        const data = extract(this.token.value)
        this.eat(TokenType.STRING)

        listStrings.push(data.string)

        while (!data.isTerminal)
        {
            this.eatEOLS()
            listExpressions.push(this.expression(1))
            this.eatEOLS()
            this.eat(TokenType.RBRACKET, EatMode.StringContinuation)
            const data = extract(this.token.value)
            this.eat(TokenType.STRING)

            listStrings.push(data.string)
            if (data.isTerminal)
                break
        }

        return new StringOp(listStrings, listExpressions)
    }

    optionalParenthesis<T extends ASTNode>(fn: (has: boolean) => T)
    {
        //  fn | LPARE EOLS? fn EOLS? RPARE

        if (this.is(TokenType.LPARE))
        {
            this.eat(TokenType.LPARE)
            this.eatEOLS()
            const node = fn(true)
            this.eatEOLS()
            this.eat(TokenType.RPARE)
            return node
        }
        return fn(false)
    }

    optionalBrackets<T extends ASTNode>(fn: (has: boolean) => T)
    {
        //  fn | LBRACKET EOLS? fn EOLS? RBRACKET

        if (this.is(TokenType.LBRACKET))
        {
            this.eat(TokenType.LBRACKET)
            this.eatEOLS()
            const node = fn(true)
            this.eatEOLS()
            this.eat(TokenType.RBRACKET)
            return node
        }
        return fn(true)
    }

    // Meant to always return something (must have terminal else expression)
    conditionalExpression(nVals: number = undefined): Conditional
    {
        return this.conditionalStatement(true, nVals)
    }

    blockOrExpression(nVals: number = undefined): ASTNode
    {
        // [LBRACKET] block | expression
        if (this.is(TokenType.LBRACKET))
            return this.block(nVals)
        else
            return this.expression(nVals)
    }

    conditionalStatement(mustHaveElse: boolean = false, nVals: number = undefined): Conditional
    {
        /*
            IF optionalParenthesis<expression> EOL? blockOrExpression EOL?
            (ELIF optionalParenthesis<expression> EOL? blockOrExpression EOL?)*
            (ELSE EOL? blockOrExpression)?
        */

        let conditions: ASTNode[] = []
        let bodies: ASTNode[] = []

        this.eat(TokenType.IF)
        conditions.push(this.optionalParenthesis(() => this.expression(1)))
        this.eatEOL()
        bodies.push(this.is(TokenType.LBRACKET) ? this.block(nVals) : this.blockSubroutine(nVals))
        this.eatEOL()

        while (this.is(TokenType.ELIF))
        {
            this.eat(TokenType.ELIF)
            conditions.push(this.optionalParenthesis(() => this.expression(1)))
            this.eatEOL()
            bodies.push(this.is(TokenType.LBRACKET) ? this.block(nVals) : this.blockSubroutine(nVals))
            this.eatEOL()
        }

        if (mustHaveElse || this.is(TokenType.ELSE))
        {
            this.eat(TokenType.ELSE)
            this.eatEOL()
            bodies.push(this.is(TokenType.LBRACKET) ? this.block(nVals) : this.blockSubroutine(nVals))
        }

        return new Conditional(conditions, bodies)
    }


    functionCall(id: ID): ASTNode
    {
        // functionCallArgumnets
        const fun = this.stack.get(id.name)

        // Dynamic
        if (!(fun instanceof BaseFunction))
        {
            this.eat(TokenType.LPARE)
            this.eatEOLS()
            const isEmpty = this.is(TokenType.RPARE)
            const parameters = isEmpty ? new List([]) : this.expression()
            this.eatEOLS()
            this.eat(TokenType.RPARE)
            return new CallOp(id, parameters)
        }

        if (id.name === "printself")
        {
            const startPos = this.lexer.absolutePosition
            const args = this.functionCallArgumnets(fun)
            const endPos = this.lexer.absolutePosition
            const text = this.lexer.text.slice(startPos - 1, endPos).trim().slice(1, -1)
            return new PrintselfCall(args, text)
        }

        if (id.name === "assert")
        {
            const startPos = this.lexer.absolutePosition
            const args = this.functionCallArgumnets(fun)
            const endPos = this.lexer.absolutePosition
            const text = this.lexer.text.slice(startPos - 1, endPos).trim().slice(1, -1)
            return new AssertCall(args, text, this.lexer.path, this.lexer.prev_line, this.lexer.prev_positionInLine)
        }

        return new FunctionCall(id.name, this.functionCallArgumnets(fun))
    }

    functionCallArgumnets(funDeclaration: BaseFunction): List
    {
        // LPARE EOLS? (expression EOLS?)? RPARE
        let IDs = funDeclaration.parameters
        let args: ASTNode[] = []

        this.eat(TokenType.LPARE)
        this.eatEOLS()

        if (IDs.length != 0)
        {
            if (IDs.length == 1)
            {
                const expressions = this.expression()
                if (expressions instanceof List && expressions.nodes.length != 1)
                    throw this.error(`${ funDeclaration.name } expected ${ 1 } argument, ${ expressions.nodes.length } provided.`, 1)

                args.push(expressions)
            }
            else
            {
                const expressions = this.expression(IDs.length)
                if (!(expressions instanceof List))
                    throw this.error(`${ funDeclaration.name } expected ${ IDs.length } arguments, ${ 1 } provided.`, 1)

                if (IDs.length !== expressions.nodes.length)
                    throw this.error(`${ funDeclaration.name } expected ${ IDs.length } arguments, ${ expressions.nodes.length } provided.`, 1)

                args.push(...expressions.nodes)
            }
            this.eatEOLS()
        }

        this.eat(TokenType.RPARE)
        return new List(args)
    }

    functionDeclaration(): NamedFunction
    {
        // FUNCTION ID functionParameters {block}
        const functionToken = this.token
        this.eat(TokenType.FUNCTION)
        const name = this.token.value
        // Checks
        {
            if (this.stack.hasInSameRecord(name))
                this.error(`Trying redefine declaration ${ name } as function name.`, name.length)
            else this.stack.set(name, functionToken)
            if (langChars.isKeywordReserved(name))
                this.error(`Function name can't be '${ name }', reserved keyword.`, name.length)
        }

        this.eat(TokenType.ID)
        const parameters = this.functionParameters()
        const funDeclaration = new NamedFunction(this.stack.isGlobalScope() ? this.stack.symbol : -1, name, parameters, null)
        this.stack.set(name, funDeclaration)

        this.stack.push()
        for (let parameter of parameters)
            this.stack.set(parameter.name, undefined)

        funDeclaration.block = this.block()

        const its = this.stack.getOwnIts()
        if (its.length)
            this.error(`Named functions can't use special 'it' variables ${ its } only anonymous functions can.`)

        this.stack.pop()

        return funDeclaration
    }

    functionParameters(): ID[]
    {
        // LPARE EOLS? (ID (COMMA EOLS? ID)* EOLS?)? RPARE EOL?

        let parameters: ID[] = []
        let ids: Set<string> = new Set()
        this.eat(TokenType.LPARE)
        this.eatEOLS()
        if (this.is(TokenType.ID))
        {
            const parameter = new ID(this.token)
            this.eat(TokenType.ID)
            parameters.push(parameter)

            ids.add(parameter.name)

            while (this.is(TokenType.COMMA))
            {
                this.eat(TokenType.COMMA)
                this.eatEOLS()
                const parameter = new ID(this.token)
                if (ids.has(parameter.name))
                    this.error(`Duplicate identifier '${ parameter.name }'.`, parameter.name.length)
                this.eat(TokenType.ID)
                parameters.push(parameter)
            }
            this.eatEOLS()
        }
        this.eat(TokenType.RPARE)
        this.eatEOL()
        return parameters
    }

    return(nVals: number = undefined): ReturnOp
    {
        // RETURN ([EOL | RBRACKET | RPAREN] | SEMICOLON | [IF] conditionalExpression | expression)
        this.eat(TokenType.RETURN)

        if (this.is(TokenType.EOL, TokenType.RBRACKET, TokenType.RPARE))
            return new ReturnOp(new NoOp())
        if (this.is(TokenType.SEMICOLON))
        {
            this.eat(this.token.type)
            return new ReturnOp(new NoOp())
        }
        if (this.is(TokenType.IF))
            return new ReturnOp(this.conditionalExpression(nVals))
        else
            return new ReturnOp(this.expression(nVals))
    }

    block(nVals: number = undefined): ASTNode
    {
        //  [LBRACKET] blockFromTo [RBRACKET]
        return this.blockFromTo(TokenType.LBRACKET, TokenType.RBRACKET, nVals)
    }

    blockFromTo(start: TokenType, end: TokenType, nVals: number = undefined): ASTNode
    {
        //  $start ( EOL+ | blockSubroutine )* $end

        this.eat(start)
        return this.blockTo(end)
    }

    blockTo(end: TokenType, nVals: number = undefined): ASTNode
    {
        //  ( EOL+ | blockSubroutine )* $end
        const nodes = []
        while (!this.is(end))
        {
            if (this.is(TokenType.EOL))
                this.eatEOLS()
            else
                nodes.push(this.blockSubroutine(nVals))
        }
        this.eat(end)

        if (nodes.length === 0) return new NoOp()
        if (nodes.length === 1) return nodes[ 0 ]
        else return new BlockOp(nodes)
    }


    // All the usual procedures you can do in a line of code
    blockSubroutine(nVals: number = undefined): ASTNode
    {
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

        if (this.is(TokenType.FUNCTION))
            return this.functionDeclaration()
        if (this.is(TokenType.FOR))
            return this.forLoop()
        if (this.is(TokenType.RETURN))
            return this.return(nVals)
        if (this.is(TokenType.IF))
            return this.conditionalStatement(false, nVals)
        if (this.is(TokenType.IMPORT))
            return this.import()

        if (this.is(TokenType.ID))
        {
            if (this.peekToken().is(TokenType.COMPOUND_DIV, TokenType.COMPOUND_MUL, TokenType.COMPOUND_PLUS, TokenType.COMPOUND_MINUS, TokenType.COMPOUND_POWER))
                return this.compoundOperation()
            if (this.isAssignation())
                return this.assignation()
        }

        return this.expression(nVals)
    }

    isAssignation(): boolean
    {
        // ID ((COMMA EOLS? ID) ASSIGN)?
        if (!this.is(TokenType.ID))
            return false

        const initialState = this.getState()
        this.eat(TokenType.ID)
        while (this.is(TokenType.COMMA))
        {
            this.eat(TokenType.COMMA)
            this.eatEOLS()
            if (!this.is(TokenType.ID))
                break
            this.eat(TokenType.ID)
        }

        const is = this.is(TokenType.ASSIGN)
        this.restoreState(initialState)
        return is
    }

    compoundOperation(): ASTNode
    {
        // ID (COMPOUND_DIV|COMPOUND_MUL|COMPOUND_PLUS|COMPOUND_MINUS) expression
        const id = new ID(this.token)

        if (langChars.isKeywordIt(id.name))
            this.error(`Not allowed to modify special variable ${ id.name }, inmmutable variable.`, id.name.length)

        if (langChars.isKeywordReserved(id.name))
            this.error(`'${ id.name }' is a reserved keyword, can't redefine.`, id.name.length)

        if (!this.stack.has(id.name))
            this.error(`'${ id.name }' not declared.`, id.name.length)

        this.eat(TokenType.ID)

        if (this.is(
            TokenType.COMPOUND_DIV,
            TokenType.COMPOUND_MUL,
            TokenType.COMPOUND_PLUS,
            TokenType.COMPOUND_MINUS,
            TokenType.COMPOUND_POWER
        ))
        {
            const token = this.token
            this.eat(this.token.type)
            return new CompoundOp(id, token, this.expression())
        }

        this.error(`Invalid token for compound operation ${ this.token }.`, this.token.value.length)

    }

    _checkForIllegalReassignation(name: string)
    {
        if (langChars.isKeywordReserved(name))
            this.error(`'${ name }' is a reserved keyword, can't redefine.`, name.length)

        if (this.stack.get(name) instanceof Stack)
            this.error(`Can't reassign variable '${ name }', in use by module as alias.`, name.length)
    }

    assignation(): ASTNode
    {
        // ID (COMMA (ID))* ASSIGN EOL? expression+

        const ids: ID[] = []

        const idNode = new ID(this.token)
        this._checkForIllegalReassignation(idNode.name)

        this.eat(TokenType.ID)
        ids.push(idNode)
        this.stack.set(idNode.name, idNode)

        while (!this.is(TokenType.ASSIGN))
        {
            this.eat(TokenType.COMMA)
            const idNode = new ID(this.token)
            this._checkForIllegalReassignation(idNode.name)
            this.eat(TokenType.ID)
            ids.push(idNode)
            this.stack.set(idNode.name, idNode)
        }

        this.eat(TokenType.ASSIGN)
        this.eatEOL()

        // This part might be a good reason to do an aditional parser
        // pass as you can't know statically the number of returns
        // directly (for now this is checked at runtime)

        const expression = this.expression()

        // CASE SIGNLE ASSIGMENT
        if (ids.length == 1)
        {
            this.stack.set(idNode.name, expression)
            return new AssignOp(ids[ 0 ], expression)
        }

        // CASE MULTIPLE ASSIGMENT
        if (expression instanceof List)
        {
            if (ids.length != expression.nodes.length)
                this.error(`Assignment needs ${ ids.length } expressions, ${ expression.nodes.length } given.`)

            for (let i = 0; i < ids.length; i++)
                this.stack.set(ids[ i ].name, expression.nodes[ i ])

            return new MultipleAssignOp(ids, expression)
        }

        return new MultipleAssignOp(ids, expression)
    }

    forLoop(): ASTNode
    {
        /*
            FOR
            {
                optionalParenthesis
                <
                    assignation? SEMICOLON
                    expression? SEMICOLON
                    blockOrExpression?
                >
                EOL? ([LBRACKET] block | blockSubroutine) EOL?
            }
        */
        const runBeforeSEMICOLON = (fn: () => ASTNode): ASTNode =>
        {
            const node = this.is(TokenType.SEMICOLON) ? new NoOp() : fn()
            this.eat(TokenType.SEMICOLON)
            return node
        }

        this.eat(TokenType.FOR)

        const loopOp = this.optionalParenthesis((has) =>
        {
            const setup = runBeforeSEMICOLON(() => this.assignation())
            const runCondition = runBeforeSEMICOLON(() => this.expression(1))
            const increment = has && this.is(TokenType.RPARE) ? new NoOp() : this.blockOrExpression()
            return new ForLoopOp(setup, runCondition, increment, null)
        })

        this.eatEOL()
        loopOp.body = this.is(TokenType.LBRACKET) ? this.block() : this.blockSubroutine()
        this.eatEOL()

        return loopOp
    }

    program(): ASTNode
    {
        // ( EOL+ | blockSubroutine EOL )* EOF

        const nodes = []
        while (!this.is(TokenType.EOF))
        {
            if (this.is(TokenType.EOL))
                this.eatEOLS()
            else
                nodes.push(this.blockSubroutine())
        }
        this.eat(TokenType.EOF)

        if (nodes.length === 0) return new NoOp()
        if (nodes.length === 1) return nodes[ 0 ]
        else return new BlockOp(nodes)
    }

    parse(): ASTNode
    {
        return this.program()
    }
}

class Interpreter
{
    constructor(public filePath: string, public symbol: number) { }

    stack = new Stack(this.filePath, this.symbol)

    error(msg: string)
    {
        throw new RuntimeError(
            extSystem.formatColor("Runtime error", extSystem.colors.red) + "\n" +
            extSystem.formatColor(" ■ ", extSystem.colors.red) + msg + "\n"
        )
    }

    visit(node: any): any
    {
        if (!(node instanceof ASTNode))
            return node

        const call = this[ `visit_${ node.constructor.name }` ]
        if (call != null) return call.bind(this)(node)
        this.error(`Trying to visit not defined node type: ${ node.constructor.name } for node ${ node }.`)
    }

    visit_ListEval(node: ListEval): any
    {
        // The list has already been evaluated
        return node
    }

    visit_ReturnOp(node: ReturnOp): any
    {
        // Throw ReturnOp so it can signal the
        // capturing parent a return call has taken place
        // but keep vising the child node so a final
        // value can be found
        node.node = this.visit(node.node)
        throw node
    }

    visit_Real(node: Real): any
    {
        return parseFloat(node.op.value)
    }

    visit_NamedImportLoadOp(node: NamedImportLoadOp): any
    {
        // Already loaded somewhere else, just need to get from importModules
        const moduleAlias = node.moduleAlias
        const moduleSymbol = node.moduleSymbol
        const moduleStack = Stack.importModules.get(moduleSymbol)
        this.stack.set(moduleAlias, moduleStack)
    }

    visit_NamedImportOp(node: NamedImportOp): any
    {
        const ast = node.moduleAST
        const moduleAlias = node.moduleAlias
        const moduleSymbol = node.moduleSymbol
        const interpreter = new Interpreter(node.filePath, moduleSymbol)
        interpreter.interpret(ast)
        const stack = interpreter.stack
        Stack.importModules.set(moduleSymbol, interpreter.stack)
        this.stack.set(moduleAlias, stack)
    }

    visit_Bool(node: Bool): any
    {
        return node.op.type == TokenType.TRUE ? true : false
    }

    visit_NoOp(node: NoOp): void { }

    visit_UnaryOp(node: UnaryOp): any
    {
        const right = this.visit(node.right)

        if (right instanceof List)
        {
            if (node.op.type == TokenType.PLUS) return right
            if (node.op.type == TokenType.MINUS) return new List(right.nodes.map(v => -v))
            if (node.op.type == TokenType.NOT) return new List(right.nodes.map(v => !v))
        }

        if (node.op.type == TokenType.PLUS) return right
        if (node.op.type == TokenType.MINUS) return - right
        if (node.op.type == TokenType.NOT) return !right

        this.error(`Token type ${ node.op.type } not implemented.`)
    }


    listIndex(object: List, index: ASTNode): any
    {
        if (typeof index == "number")
        {
            if (index < 0)
                return object.nodes[ object.nodes.length + index ]
            return object.nodes[ index ]
        }

        if (index instanceof List)
        {
            const start = index.nodes[ 0 ] as number
            const end = index.nodes[ 1 ] as number
            if (end == -1)
                return new List(object.nodes.slice(start))
            if (end < -1)
                return new List(object.nodes.slice(start, end + 1))
            return new List(object.nodes.slice(start, end))
        }
        throw this.error(`Index value in not integer or range.`)

    }

    stringIndex(object: String, index: ASTNode): any
    {
        if (typeof index == "number")
        {
            if (index < 0)
                return object[ object.length + index ]
            return object[ index ]
        }

        if (index instanceof List)
        {
            const start = index.nodes[ 0 ] as number
            const end = index.nodes[ 1 ] as number
            if (end == -1)
                return object.slice(start)
            if (end < -1)
                return object.slice(start, end + 1)
            return object.slice(start, end)
        }
        throw this.error(`Index value in not integer or range.`)
    }

    visit_AccessIndexOp(node: AccessIndexOp): any
    {
        const object = this.visit(node.list)
        if (object instanceof List)
            return this.listIndex(object, this.visit(node.index))

        if (typeof object == "string")
            return this.stringIndex(object, this.visit(node.index))

        throw this.error(`Object is not a list or string, can't acccess index.`)
    }

    // Defined only for the primitives types:
    // list, dict, string, number, bool, undefined and null
    visit_AccessMethodOp(node: AccessMethodOp): any
    {
        const object = this.visit(node.object)

        if (object instanceof List)
        {
            switch (node.key)
            {
                case "toString": return `${ object }`
                case "toDict": return new Dictionary(Object.fromEntries(object.nodes.map(v => (v as List).nodes)))
                case "size": return object.nodes.length
                case "sort": return new List([ ...object.nodes ].sort((a, b) =>
                    (a instanceof List ? a.nodes[ 0 ] : a) as number -
                    ((b instanceof List ? b.nodes[ 0 ] : b) as number))
                )
                case "reverse": return new List([ ...object.nodes ].reverse())
                case "flat": return new List(object.nodes.flatMap(v => v instanceof List ? v.nodes : v))
                case "is": return "list"
                case "first": return object.nodes[ 0 ]
                case "last": return object.nodes[ object.nodes.length - 1 ]
                case "zip": {
                    if (object.nodes.length != 2)
                        throw this.error(`Zip call needs a list with two sub lists of the same size.`)
                    const first = object.nodes[ 0 ] as List
                    const second = object.nodes[ 1 ] as List
                    if (!(first instanceof List) || !(second instanceof List))
                        throw this.error(`Zip call needs a list with two sub lists of the same size.`)
                    if (first.nodes.length != second.nodes.length)
                        throw this.error(`Zip call needs a list with two sub lists of the same size.`)
                    return new List(first.nodes.map((v, i) => new List([ v, second.nodes[ i ] ])))
                }
                case "lastIndex": return object.nodes.length - 1
                case "isEmpty": return object.nodes.length == 0
                case "isNotEmpty": return object.nodes.length != 0
                case "sum": return object.nodes.reduce((a, b) => (a as any) + (b as number), 0 as number)
                case "max": return Math.max.apply(null, object.nodes)
                case "min": return Math.min.apply(null, object.nodes)
                case "average": return object.nodes.length == 0 ? 0 : object.nodes.reduce((a, b) => (a as any) + (b as number), 0 as number) / object.nodes.length
                case "randomItem": return object.nodes[ Math.floor(Math.random() * (object.nodes.length)) ]
                case "dropFirst": return new List(object.nodes.slice(1))
                case "dropLast": return new List(object.nodes.slice(0, -1))
                case "hypot": return Math.sqrt(object.nodes.reduce((a, c) => (a as any) + (c as number) * (c as number), 0 as number))
                case "group": {
                    const dict = new Dictionary({})
                    for (let item of object.nodes)
                    {
                        if (!(item instanceof List) || item.nodes.length != 2)
                            throw this.error(`Group operation. List items must be list of size 2: (key, value).`)
                        const key = item.nodes[ 0 ]
                        if (typeof key != "string" && typeof key != "number")
                            throw this.error(`Group operation. List item key must be 'string' or 'number'.`)
                        const value = item.nodes[ 1 ]
                        if (key in dict.nodes)
                            dict.nodes[ key ].nodes.push(value)
                        else
                            dict.nodes[ key ] = new List([ value ])
                    }
                    return dict
                }
            }

            this.error(`Method '${ node.key }' not found in object type 'list'.`)
        }

        if (object instanceof Dictionary)
        {
            switch (node.key)
            {
                case "toString": return `${ object }`
                case "toList": return new List(Object.entries(object.nodes).map(v => new List(v)))
                case "size": return Object.keys(object.nodes).length
                case "keys": return new List(Object.keys(object.nodes))
                case "values": return new List(Object.values(object.nodes))
                case "is": return "dict"
                case "isEmpty": return Object.keys(object.nodes).length == 0
                case "isNotEmpty": return Object.keys(object.nodes).length != 0
            }

            this.error(`Method '${ node.key }' not found in object type 'dict' (dictionary).`)
        }

        if (typeof object == "string")
        {
            switch (node.key)
            {
                case "toString": return object.toString()
                case "toNumber": return parseFloat(object)
                case "toList": return new List(object.split(""))
                case "size": return object.length
                case "toLowerCase": return object.toLowerCase()
                case "toUpperCase": return object.toUpperCase()
                case "trim": return object.trim()
                case "trimEnd": return object.trimEnd()
                case "trimStart": return object.trimStart()
                case "capitalize": return object.length == 0 ? "" : object[ 0 ]?.toUpperCase() + object.slice(1)
                case "is": return "string"
                case "lines": return new List(object.split("\n"))
                case "isEmpty": return object === ""
                case "isNotEmpty": return object !== ""
                case "isBlank": return object.trim() === ""
                case "isNotBlank": return object.trim() !== ""
                case "first": return object[ 0 ]
                case "last": return object[ object.length - 1 ]
                case "hashCode": {
                    // Java's equivalent
                    if (object.length == 0)
                        return 0
                    let hash = 0
                    for (let i = 0; i < object.length; ++i)
                        hash = (((hash << 5) - hash) + object.charCodeAt(i)) | 0  // int32
                    return hash;
                }
            }
            this.error(`Method '${ node.key }' not found in object type 'string'.`)
        }

        if (typeof object == "number")
        {
            switch (node.key)
            {
                case "toString": return object.toString()
                case "toBool": return !object ? false : true
                case "toList": return new List(new Array(Math.floor(Math.abs(object))).fill(0).map((_, i) => i))
                case "trunc": return Math.trunc(object)
                case "round": return Math.round(object)
                case "ceil": return Math.ceil(object)
                case "floor": return Math.floor(object)
                case "abs": return Math.abs(object)
                case "sign": return Math.sign(object)
                case "cos": return Math.cos(object)
                case "sin": return Math.sin(object)
                case "tan": return Math.tan(object)
                case "cosh": return Math.cosh(object)
                case "sinh": return Math.sinh(object)
                case "tanh": return Math.tanh(object)
                case "random": return Math.random() * object
                case "randomInt": return Math.floor(Math.random() * (object + 1))
                case "randomSign": return (1 - Math.floor(Math.random() * 2) * 2) * object
                case "sqrt": return Math.sqrt(object)
                case "negative": return -object
                case "isNegative": return object < 0
                case "isPositive": return object >= 0
                case "isInfinite": return object === Infinity || object === -Infinity
                case "isFinite": return object !== Infinity && object !== -Infinity && object !== NaN
                case "isNaN": return object === NaN
                case "is": return "number"
                case "rad": return object * Math.PI / 180
                case "deg": return object * 180 / Math.PI
                case "pi": return object * Math.PI
            }
            this.error(`Method '${ node.key }' not found in object type 'number': ${ object }`)
        }

        if (typeof object == "boolean")
        {
            switch (node.key)
            {
                case "toString": return object ? "true" : "false"
                case "toNumber": return object ? 1 : 0
                case "flip": return object ? false : true
                case "is": return "bool"
            }
            this.error(`Method '${ node.key }' not found in object type 'bool': ${ object }`)
        }

        if (object === undefined)
        {
            switch (node.key)
            {
                case "toString": return "undefined"
                case "is": return "undefined"
            }
            this.error(`Method '${ node.key }' not found in object type 'undefined': ${ object }`)
        }

        if (object === null)
        {
            switch (node.key)
            {
                case "toString": return "null"
                case "is": return "null"
            }
            this.error(`Method '${ node.key }' not found in object type 'null': ${ object }`)
        }

        if (object instanceof Stack)
        {
            switch (node.key)
            {
                case "keys": return new List(object.getGlobalVariablesKeys())
                case "is": return "dict"
            }
            this.error(`Method '${ node.key }' not found in object type 'null': ${ object }`)
        }

        if (node.key == "toString") return `${ object }`

        this.error(`Can't call method '${ node.key }'' from unknown object type '${ typeof object }'`)
    }

    visit_AccessKeyOp(node: AccessKeyOp): any
    {
        const object = this.visit(node.object)
        if (object instanceof Dictionary)
            return object.nodes[ node.key ]
        if (object instanceof Stack)
        {
            if (!object.has(node.key))
                this.error(`Module ${ (node.object as ID).name }: ${ (node.object as ID).name }.${ node.key } not found.`)
            return object.get(node.key)
        }
        throw this.error(`Object is not a dictionary o module, can't acccess by key.`)
    }

    visit_List(node: List): ListEval
    {
        return new ListEval(node.nodes.map(v => this.visit(v)))
    }

    visit_Dictionary(node: Dictionary): any
    {
        for (let key in node.nodes)
        {
            if (node.nodes[ key ] instanceof AnonymousFunction)
                continue
            node.nodes[ key ] = this.visit(node.nodes[ key ])
        }
        return node
    }

    visit_CallOp(node: CallOp): any
    {
        const fun = this.visit(node.func)
        if (fun instanceof BaseFunction)
        {
            const args = node.args instanceof List ? node.args : new List([ node.args ])
            return this.executeFunction(fun, args)
        }
        throw this.error(`Object is not an function, can't execute a call. Got '${ fun }' instead.`)
    }

    visit_PipeOp(pipe: PipeOp): any
    {
        const itemExec = (fun: BaseFunction, arg: ASTNode) => this.executeFunction(
            fun,
            arg instanceof List ? arg : new List([ arg ])
        )

        let left = this.visit(pipe.entry)
        for (let pipeNode of pipe.nodes)
        {
            const args = left instanceof List ? left : new List([ left ])
            if (pipeNode.expression instanceof FunctionCall)
            {
                const fun = this.stack.get(pipeNode.expression.name) as BaseFunction
                if (pipeNode.passType == TokenType.MAP)
                    left = new List(args.nodes.map(arg => itemExec(fun, arg)))
                else if (pipeNode.passType == TokenType.FILTER)
                    left = new List(args.nodes.filter(arg => itemExec(fun, arg)))
                else
                    left = this.executeFunction(fun, args)
            }
            else if (pipeNode.expression instanceof AnonymousFunction)
            {
                const fun = pipeNode.expression
                if (pipeNode.passType == TokenType.MAP)
                    left = new List(args.nodes.map(arg => itemExec(fun, arg)))
                else if (pipeNode.passType == TokenType.FILTER)
                    left = new List(args.nodes.filter(arg => itemExec(fun, arg)))
                else
                    left = this.executeFunction(fun, args)
            }
            else
            {
                const fun = this.visit(pipeNode.expression)
                if (!(fun instanceof BaseFunction))
                    throw this.error(`Invalid pipe operator, not a function: '${ fun }'.`)

                if (pipeNode.passType == TokenType.MAP)
                    left = new List(args.nodes.map(arg => itemExec(fun, arg)))
                else if (pipeNode.passType == TokenType.FILTER)
                    left = new List(args.nodes.filter(arg => itemExec(fun, arg)))
                else
                    left = this.executeFunction(fun, args)
            }
        }

        return left
    }

    visit_VariableIncrementPrefixOp(node: VariableIncrementPrefixOp): any
    {
        const currentVal = this.visit_ID(node.right)

        if (node.op.type == TokenType.PLUSPLUS)
        {
            const newVal = currentVal + 1
            this.stack.set(node.right.name, newVal)
            return newVal
        }
        if (node.op.type == TokenType.MINUSMINUS)
        {
            const newVal = currentVal - 1
            this.stack.set(node.right.name, newVal)
            return newVal
        }

        this.error(`Token type ${ node.op.type } not implemented.`)
    }

    visit_VariableIncrementPostfixOp(node: VariableIncrementPostfixOp): any
    {
        const currentVal = this.visit_ID(node.left)

        if (node.op.type == TokenType.PLUSPLUS)
        {
            this.stack.set(node.left.name, currentVal + 1)
            return currentVal
        }
        if (node.op.type == TokenType.MINUSMINUS)
        {
            this.stack.set(node.left.name, currentVal - 1)
            return currentVal
        }

        this.error(`Token type ${ node.op.type } not implemented.`)
    }

    visit_AssignOp(node: AssignOp): any
    {
        const right = this.visit(node.expression)
        this.stack.set(node.id.name, right)
        return
    }

    visit_MultipleAssignOp(node: MultipleAssignOp): any
    {
        const res = this.visit(node.expression)
        if (!(res instanceof List))
            throw this.error(`Assignment ${ node.ids.map(v => v.name) } expected ${ node.ids.length } values got one: '${ res }'.`)

        if (node.ids.length > res.nodes.length)
            throw this.error(`Invalid assignment, expected ${ node.ids.length } values for ${ node.ids.map(v => v.name) } expression got ${ res.nodes.length }: ${ res.nodes }.`)


        for (let i = 0; i < node.ids.length; i++)
            this.stack.set(node.ids[ i ].name, res.nodes[ i ])
    }

    visit_ComparisonOp(node: ComparisonOp): any
    {
        let left = this.visit(node.nodes[ 0 ])
        for (let i = 0; i < node.comparators.length; i++)
        {
            let right = this.visit(node.nodes[ i + 1 ])
            const type = node.comparators[ i ].type

            switch (type)
            {
                case TokenType.EQUAL: {
                    if (left instanceof List && right instanceof List)
                    {
                        if (left.nodes.length == right.nodes.length)
                        {
                            if (left.nodes.every((v, i) => v == right.nodes[ i ])) break
                            else return false
                        }
                        else return false
                    }
                    else if (left == right) break
                    else return false
                }
                case TokenType.NOT_EQUAL: {
                    if (left instanceof List && right instanceof List)
                    {
                        if (left.nodes.length == right.nodes.length)
                        {
                            if (left.nodes.every((v, i) => v == right.nodes[ i ])) return false
                            else break
                        }
                        else break
                    }
                    else if (left != right) break
                    else return false
                }
                case TokenType.LESS: {
                    if (left < right) break
                    else return false
                }
                case TokenType.LESS_OR_EQUAL: {
                    if (left <= right) break
                    else return false
                }
                case TokenType.GREATER: {
                    if (left > right) break
                    else return false
                }
                case TokenType.GREATER_OR_EQUAL: {
                    if (left >= right) break
                    else return false
                }
                default: this.error(`Token type ${ type } not implemented`)
            }
            left = right
        }

        return true
    }

    visit_BinaryOp(node: BinaryOp): any
    {
        const left = this.visit(node.left)
        const right = this.visit(node.right)
        const isListL = left instanceof List
        const isListR = right instanceof List
        if (isListL && !isListR)
        {
            if (node.op.type === TokenType.PLUS) return new List([ ...left.nodes, right ])
            if (node.op.type === TokenType.MUL) return new List(new Array(right).fill(0).flatMap(_ => left.nodes))
        }
        if (!isListL && isListR)
        {
            if (node.op.type === TokenType.PLUS) return new List([ left, ...right.nodes ])
            if (node.op.type === TokenType.MUL) return new List(new Array(left).fill(0).flatMap(_ => right.nodes))
        }
        if (isListL && isListR)
        {
            if (node.op.type === TokenType.PLUS) return new List([ ...left.nodes, ...right.nodes ])
        }

        const isStringL = typeof left === "string"
        const isStringR = typeof right === "string"
        if (!isStringL && isStringR)
        {
            if (node.op.type === TokenType.MUL) return right.repeat(left)
        }
        if (isStringL && !isStringR)
        {
            if (node.op.type === TokenType.MUL) return left.repeat(right)
        }

        if (node.op.type === TokenType.PLUS) return left + right
        if (node.op.type === TokenType.MINUS) return left - right
        if (node.op.type === TokenType.MUL) return left * right
        if (node.op.type === TokenType.DIV) return left / right
        if (node.op.type === TokenType.MODULO) return left % right
        if (node.op.type === TokenType.POWER) return Math.pow(left, right)

        this.error(`Token type ${ node.op.type } not implemented.`)
    }

    visit_CompoundOp(node: CompoundOp): any
    {
        const id = node.left
        const value = this.visit_ID(id)
        const compoundValue = this.visit(node.right)

        const isListL = value instanceof List
        const isListR = compoundValue instanceof List

        switch (node.op.type)
        {
            case TokenType.COMPOUND_PLUS: {
                if (isListL && !isListR)
                    return this.stack.setOnExisting(id.name, new List([ ...value.nodes, compoundValue ]))
                if (isListL && isListR)
                    return this.stack.setOnExisting(id.name, new List([ ...value.nodes, ...compoundValue.nodes ]))
                if (!isListL && !isListR)
                    return this.stack.setOnExisting(id.name, value + compoundValue)
                break
            }
            case TokenType.COMPOUND_MINUS: {
                if (!isListL && !isListR)
                    return this.stack.setOnExisting(id.name, value - compoundValue)
                break
            }
            case TokenType.COMPOUND_MUL: {
                if (isListL && !isListR)
                    return this.stack.setOnExisting(id.name, new List(new Array(compoundValue).fill(0).flatMap(_ => value.nodes)))
                if (!isListL && !isListR)
                    return this.stack.setOnExisting(id.name, value * compoundValue)
                break
            }
            case TokenType.COMPOUND_DIV: {
                if (!isListL && !isListR)
                    return this.stack.setOnExisting(id.name, value / compoundValue)
                break
            }
            case TokenType.COMPOUND_POWER: {
                if (!isListL && !isListR)
                    return this.stack.setOnExisting(id.name, Math.pow(value, compoundValue))
                break
            }
        }

        this.error(`Invalid token ${ node.op } operation for combination of ${ isListL ? "list" : "scalar" } with ${ isListR ? "list" : "scalar" }.`)
    }

    visit_ShortcircuitOp(node: ShortcircuitOp): any
    {
        const left = this.visit(node.left)
        if (node.op.type == TokenType.AND)
        {
            if (!left)
                return false

            const right = this.visit(node.right)
            return !!right
        }
        if (node.op.type == TokenType.OR)
        {
            if (!!left)
                return true

            const right = this.visit(node.right)
            return !!right
        }

        this.error(`Token type ${ node.op.type } not implemented.`)
    }

    visit_NamedFunction(node: NamedFunction): any
    {
        this.stack.set(node.name, node)
    }

    visit_FunctionCall(call: FunctionCall): any
    {
        const fun = this.stack.get(call.name) as BaseFunction
        return this.executeFunction(fun, call.args)
    }

    executeFunction(fun: BaseFunction, args: List)
    {
        const evaluatedArgs = this.visit_List(args)
        const currentScope = this.stack

        // -1 means the base stack scope is the same
        // (happens when a function is not defined in global scope)
        if (fun.stackNamespace != -1)
            this.stack = Stack.importModules.get(fun.stackNamespace)
        this.stack.push()

        try
        {
            this.stack.set("arguments", evaluatedArgs)

            for (let i = 0; i < fun.parameters.length; i++)
                this.stack.set(fun.parameters[ i ].name, evaluatedArgs.nodes[ i ])

            if (fun instanceof NamedFunction || fun instanceof AnonymousFunction)
                return this.visit(fun.block)
            else if (fun instanceof NativeFunction)
                return fun.fun.apply(fun.fun, evaluatedArgs.nodes)
            else
                this.error("FunctionCall, invalid instanceof type.")
        }
        catch (value)
        {
            if (value instanceof ReturnOp) return value.node
            throw value
        }
        finally
        {
            this.stack.pop()
            this.stack = currentScope
        }
    }

    visit_AnonymousFunction(node: AnonymousFunction)
    {
        return this.scopedReturn(() => { return this.visit(node.block) })
    }

    visit_SpreadListOp(node: SpreadListOp): any
    {
        const res = this.visit(node.list)
        if (!(res instanceof List))
            throw this.error("Can't spread, object is not a list.")
        return res.nodes
    }

    visit_PrintselfCall(call: PrintselfCall): any
    {
        const evaluatedArgs = this.visit_List(call.args)
        const msg = evaluatedArgs.nodes[ 0 ]
        extSystem.print(`${ call.text } ⟶  ${ msg }`)
    }

    visit_AssertCall(call: AssertCall): any
    {
        const res = this.visit_List(call.args)
        if (res.nodes[ 0 ] === true) return

        const location = `(${ call.file }:${ call.line + 1 }:${ call.column })`
        extSystem.print(
            extSystem.formatColor("FAILED ASSERT:", extSystem.colors.red) + "  " + call.text + "\n" +
            location
        )
    }

    visit_Conditional(node: Conditional): any
    {
        for (let i = 0; i < node.conditions.length; i++)
            if (this.visit(node.conditions[ i ]))
                return this.visit(node.bodies[ i ])

        if (node.conditions.length != node.bodies.length)
            return this.visit(node.bodies[ node.bodies.length - 1 ])
    }

    visit_ForLoopOp(node: ForLoopOp): any
    {
        this.visit(node.setup)
        while (this.visit(node.runCondition) == true)
        {
            this.visit(node.body)
            this.visit(node.increment)
        }
    }

    visit_RangeOp(node: RangeOp): any
    {
        const list = []
        const start = this.visit(node.start) as number
        const step = Math.abs(this.visit(node.step) ?? 1)
        if (step === 0) this.error("Range step can't be zero.")
        const end = this.visit(node.end) as number

        if (start < end)
        {
            const inv_step = 1.0 / step
            const range = (end - start) * inv_step
            list.length == range + 1
            for (let i = 0; i <= range; i++)
                list.push(start + i * step)
            return new List(list)
        }
        else
        {
            const inv_step = 1.0 / step
            const range = (start - end) * inv_step
            list.length == range + 1
            for (let i = 0; i <= range; i++)
                list.push(start - i * step)
            return new List(list)
        }
    }

    visit_BlockOp(blockNode: BlockOp): any
    {
        const lastIndex = blockNode.list.length - 1
        for (let i = 0; i <= lastIndex; i++)
        {
            const node = this.visit(blockNode.list[ i ])
            if (i == lastIndex)
                return node
        }
    }

    visit_StringOp(stringOp: StringOp): string
    {
        let output = stringOp.strings[ 0 ]
        for (let i = 0; i < stringOp.interpolations.length; i++)
        {
            const node = this.visit(stringOp.interpolations[ i ])
            output += String(node)
            output += stringOp.strings[ i + 1 ]
        }
        return output
    }

    visit_Undefined(node: Undefined): any
    {
        return undefined
    }

    visit_Null(node: Null): any
    {
        return null
    }

    visit_ID(node: ID): any
    {
        if (this.stack.has(node.name))
            return this.stack.get(node.name)
        else
            this.error(`Variable '${ node.name }' is not defined` + "\n" +
                `(${ this.stack.pathFile })`)
    }

    returnCatcher(fn: () => ASTNode)
    {
        try { return fn() }
        catch (value)
        {
            if (value instanceof ReturnOp) return value.node
            throw value
        }
    }

    scopedReturn(fn: () => ASTNode)
    {
        this.stack.push()
        const res = this.returnCatcher(() => fn())
        this.stack.pop()
        return res
    }

    interpret(tree: ASTNode, dumpAST = false, dumpFile: string = undefined): any
    {
        if (dumpAST)
            extSystem.writeFile(extSystem.prettyPrintToString(tree), dumpFile, undefined)

        return this.returnCatcher(() => this.visit(tree))
    }
}


class JSEtimos
{
    constructor(public args: ProgramArgs) { }

    shellMode()
    {
        const interpreter = new Interpreter(extSystem.dirName(), 0)
        const globalStack = new Stack(extSystem.dirName(), 0)
        Stack.importModules.set(Stack.namespaceGlobalCount, interpreter.stack)

        extSystem.shell_onLineCallback = (line: string) =>
        {
            try
            {
                const lexer = new Lexer(line, extSystem.dirName(), ":shell")
                const parser = new Parser(lexer, globalStack)
                const result = interpreter.interpret(parser.parse())
                if (result !== undefined)
                    extSystem.print(`${ result }`)
            }
            catch (e)
            {
                if (e instanceof LexerError || e instanceof ParserError || e instanceof RuntimeError)
                    extSystem.print(e.message)
                else
                    extSystem.print("[[Unknown error, failed to execute]]\n\n" + e)
            }

            extSystem.shell_stdoutWrite(">>> ")
        }

        extSystem.print("JSEtimos v0.0.1 - nani [shell mode] Hello there :)")
        extSystem.shell_stdoutWrite(">>> ")
        extSystem.shell_loadInterface()
    }

    fileMode(fileName: string)
    {
        const filePath = extSystem.dirName() + "/" + fileName
        const dumpAST = this.args.dumpAST
        const dumpASTFile = this.args.dumpFile ?? filePath + ".ast.yml"

        const dir = extSystem.dirName() + "/" + fileName.split("/").slice(0, -1).join("/")

        Stack.reset()
        const text: string = extSystem.readFile(filePath)
        const lexer = new Lexer(text, dir, fileName)
        const parser = new Parser(lexer, new Stack(filePath, 0))
        const interpreter = new Interpreter(filePath, 0)
        Stack.importModules.set(Stack.namespaceGlobalCount, interpreter.stack)

        const result = interpreter.interpret(parser.parse(), dumpAST, dumpASTFile)
        if (result !== undefined)
            extSystem.print(`${ result }`)
    }

    inputMode(text: string)
    {
        const fileName = "input"
        const filePath = extSystem.dirName() + "/" + fileName
        const dumpAST = this.args.dumpAST
        const dumpASTFile = this.args.dumpFile ?? filePath + ".ast.yml"

        const dir = extSystem.dirName() + "/" + fileName.split("/").slice(0, -1).join("/")

        Stack.reset()
        const lexer = new Lexer(text, dir, fileName)
        const parser = new Parser(lexer, new Stack(filePath, 0))
        const interpreter = new Interpreter(filePath, 0)
        Stack.importModules.set(Stack.namespaceGlobalCount, interpreter.stack)

        const result = interpreter.interpret(parser.parse(), dumpAST, dumpASTFile)
        if (result !== undefined)
            extSystem.print(`${ result }`)
    }

    run()
    {
        try
        {
            if (this.args.fileName) this.fileMode(this.args.fileName)
            else if (this.args.inputMode) this.inputMode(this.args.inputText)
            else this.shellMode()
        }
        catch (e)
        {
            if (e instanceof LexerError || e instanceof ParserError || e instanceof RuntimeError)
                this.printError(e)
            else
                throw e
        }
    }

    printError(e: Error)
    {
        // In chrome the message is duplicated
        extSystem.print(e.message)
        const stack = e.stack.replace(e.message,"")
        extSystem.print(stack)
    }
}

var extSystem: ExtSystem = undefined
var etimos: JSEtimos = undefined

// node
if ((typeof process !== "undefined") && (process.release.name === "node"))
{
    extSystem = new ExtSystem_NodeJs()
    // extSystem.print("Loaded in node.js")
    etimos = new JSEtimos(extSystem.programArgs())
    etimos.run()
}
// Pyrogenesis
else if (typeof global !== "undefined" && global[ "Engine" ] !== undefined && global[ "Engine" ][ "ProfileStart" ] != undefined)
{
    extSystem = new ExtSystem_Pyrogenesis0ad()
    // extSystem.print("Loaded in pyrogenesis engine")
    // Manual
}
// Browser
else
{
    extSystem = new ExtSystem_Browser()
    // extSystem.print("Loaded in the browser")
    // Manual
}

