// JSEtimos - copyright nani 2021
class ExtSystem_NodeJs {
    constructor() {
        this.fs = require('fs');
        this.util = require("util");
        this.redirectPrint = false;
        this.redirectPrintCallback = undefined;
        this.colors = {
            reset: "\u001b[0m",
            red: "\u001b[31;1m",
        };
        this.shell_interface = undefined;
        this.shell_loaded = false;
        try {
            require('source-map-support').install();
        }
        catch (e) {
            this.print("WARNING: 'source-map-support' javascript library not installed (needed for TypeScript stacktraces) ");
        }
    }
    readFile(filePath) {
        return this.fs.readFileSync(filePath, "utf-8");
    }
    writeFile(data, file, baseDir = undefined) {
        var _a;
        (_a = baseDir == baseDir) !== null && _a !== void 0 ? _a : this.dirName();
        return this.fs.writeFile(baseDir + "/" + file, data, () => { });
    }
    existFile(filePath) {
        return this.fs.existsSync(filePath);
    }
    print(msg) {
        var _a;
        if (this.redirectPrint)
            (_a = this.redirectPrintCallback) === null || _a === void 0 ? void 0 : _a.call(this, msg);
        else
            console.log(msg);
    }
    prettyPrintToString(msg, color = true) {
        return this.util.inspect(msg, { depth: null, colors: color, compact: 2 });
    }
    formatColor(text, color) {
        return `${color}${text}${this.colors.reset}`;
    }
    dirName() {
        return __dirname;
    }
    programArgs() {
        const list = process.argv.slice(2);
        const attributes = {};
        for (let item of list) {
            if (item.startsWith("--"))
                attributes[item.slice(2)] = item.slice(2);
            else if (item.startsWith("-"))
                continue;
            else
                attributes["text"] = item;
        }
        const inputMode = !!attributes.inputMode;
        return {
            fileName: inputMode ? "" : attributes.text,
            dumpAST: !!attributes.dumpAST,
            dumpFile: attributes.dumpFile,
            shellMode: attributes.shellMode,
            inputMode: inputMode,
            inputText: inputMode ? attributes.text : ""
        };
    }
    shell_loadInterface() {
        var _a;
        if (this.shell_loaded)
            return;
        this.shell_loaded = true;
        this.shell_interface = require("readline").createInterface({
            input: process.stdin,
            output: process.stdout,
            terminal: false
        });
        (_a = this.shell_interface) === null || _a === void 0 ? void 0 : _a.on("line", (line) => this.shell_onLineCallback(line));
    }
    shell_stdoutWrite(text) {
        process.stdout.write(text);
    }
}
class ExtSystem_Pyrogenesis0ad {
    constructor() {
        this.redirectPrint = false;
        this.redirectPrintCallback = undefined;
        this.colors = {
            reset: "",
            red: "",
        };
        this.shell_loaded = false;
        this.shell_interface = undefined;
    }
    readFile(filePath) {
        return "";
    }
    writeFile(data, file, baseDir = undefined) {
        var _a;
        (_a = baseDir == baseDir) !== null && _a !== void 0 ? _a : this.dirName();
        return "";
    }
    existFile(filePath) {
        return false;
    }
    print(msg) {
        var _a;
        if (this.redirectPrint)
            (_a = this.redirectPrintCallback) === null || _a === void 0 ? void 0 : _a.call(this, msg);
        else
            global.warn(msg);
    }
    prettyPrintToString(msg, color = true) {
        return msg.map(v => `${v}`);
    }
    formatColor(text, color) {
        return `${color}${text}${this.colors.reset}`;
    }
    dirName() {
        return "___";
    }
    programArgs() {
        return {
            fileName: undefined,
            dumpAST: false,
            dumpFile: undefined,
            shellMode: true,
            inputMode: false,
            inputText: ""
        };
    }
    shell_loadInterface() {
        if (this.shell_loaded)
            return;
        this.shell_loaded = true;
        this.shell_interface.onPress = () => {
            const line = this.shell_interface.caption;
            this.shell_onLineCallback(line);
        };
    }
    shell_stdoutWrite(text) {
        global.warn(text);
    }
}
class ExtSystem_Browser {
    constructor() {
        this.redirectPrint = false;
        this.redirectPrintCallback = undefined;
        this.colors = {
            reset: "</span>",
            red: '<span style="color: #cc0000">',
        };
        this.shell_loaded = false;
        this.shell_interface = undefined;
    }
    readFile(filePath) {
        return "";
    }
    writeFile(data, file, baseDir = undefined) {
        var _a;
        (_a = baseDir == baseDir) !== null && _a !== void 0 ? _a : this.dirName();
        return "";
    }
    existFile(filePath) {
        return false;
    }
    print(msg) {
        var _a;
        if (this.redirectPrint)
            (_a = this.redirectPrintCallback) === null || _a === void 0 ? void 0 : _a.call(this, msg);
        else
            console.log(msg);
    }
    prettyPrintToString(msg, color = true) {
        return msg.map(v => `${v}`);
    }
    formatColor(text, color) {
        return `${color}${text}${this.colors.reset}`;
    }
    dirName() {
        return "___";
    }
    programArgs() {
        return {
            fileName: undefined,
            dumpAST: false,
            dumpFile: undefined,
            shellMode: false,
            inputMode: true,
            inputText: `print("hello")`
        };
    }
    shell_loadInterface() {
        if (this.shell_loaded)
            return;
        this.shell_loaded = true;
        this.shell_interface.onPress = () => {
            const line = this.shell_interface.caption;
            this.shell_onLineCallback(line);
        };
    }
    shell_stdoutWrite(text) {
        console.log(text);
    }
}
var TokenType;
(function (TokenType) {
    TokenType["START_SYMBOL_CHARS"] = "START_SYMBOL_CHARS";
    TokenType["PLUSPLUS"] = "++";
    TokenType["MINUSMINUS"] = "--";
    TokenType["PLUS"] = "+";
    TokenType["MINUS"] = "-";
    TokenType["MUL"] = "*";
    TokenType["DIV"] = "/";
    TokenType["POWER"] = "**";
    TokenType["SCALAR_PLUS"] = ".+";
    TokenType["SCALAR_MINUS"] = ".-";
    TokenType["SCALAR_MUL"] = ".*";
    TokenType["SCALAR_DIV"] = "./";
    TokenType["COMPOUND_PLUS"] = "+=";
    TokenType["COMPOUND_MINUS"] = "-=";
    TokenType["COMPOUND_MUL"] = "*=";
    TokenType["COMPOUND_DIV"] = "/=";
    TokenType["COMPOUND_POWER"] = "**=";
    TokenType["LESS"] = "<";
    TokenType["GREATER"] = ">";
    TokenType["EQUAL"] = "==";
    TokenType["NOT_EQUAL"] = "!=";
    TokenType["LESS_OR_EQUAL"] = "<=";
    TokenType["GREATER_OR_EQUAL"] = ">=";
    TokenType["LPARE"] = "(";
    TokenType["RPARE"] = ")";
    TokenType["LBRACKETSQR"] = "[";
    TokenType["RBRACKETSQR"] = "]";
    TokenType["LBRACKET"] = "{";
    TokenType["RBRACKET"] = "}";
    TokenType["ASSIGN"] = "=";
    TokenType["COMMA"] = ",";
    TokenType["DOLLAR"] = "$";
    TokenType["DOUBLE_QUOTE"] = "\"";
    TokenType["SEMICOLON"] = ";";
    TokenType["COLON"] = ":";
    TokenType["MODULO"] = "%";
    TokenType["PIPE"] = "|";
    TokenType["DOT"] = ".";
    TokenType["DOT_THREE"] = "...";
    TokenType["ARROW_RIGHT"] = "->";
    TokenType["QUESTION"] = "?";
    TokenType["END_SYMBOL_CHARS"] = "END_SYMBOL_CHARS";
    TokenType["FUNCTION"] = "fun";
    TokenType["RETURN"] = "return";
    TokenType["IF"] = "if";
    TokenType["ELIF"] = "elif";
    TokenType["ELSE"] = "else";
    TokenType["TRUE"] = "true";
    TokenType["FALSE"] = "false";
    TokenType["AND"] = "and";
    TokenType["NOT"] = "not";
    TokenType["OR"] = "or";
    TokenType["DO"] = "do";
    TokenType["FOR"] = "for";
    TokenType["AS"] = "as";
    TokenType["MAP"] = "map";
    TokenType["FILTER"] = "filter";
    TokenType["IMPORT"] = "import";
    TokenType["END_FIRST_CHAR_IS_ALPHA"] = "END_FIRST_CHAR_IS_ALPHA";
    TokenType["REAL"] = "REAL";
    TokenType["STRING"] = "STRING";
    TokenType["ID"] = "ID";
    TokenType["EOL"] = "EOL";
    TokenType["EOF"] = "EOF";
    // END SPECIAL TOKENS
    TokenType["VAL"] = "val";
    TokenType["CONST"] = "const";
    TokenType["VAR"] = "var";
    TokenType["CONTINUE"] = "continue";
    TokenType["BREAK"] = "break";
    // END UNIMPLEMENTED TOKENS
})(TokenType || (TokenType = {}));
var langChars = {
    SYMBOL_CHAR_TREE: {},
    MULTIPLE_CHAR_FIRST_IS_ALPHA: new Set(),
    RESERVED_KEYWORDS: new Set(),
    isKeywordIt: function (keyword) {
        return /^it\d*$/.test(keyword);
    },
    isKeywordReserved: function (keyword) {
        return this.RESERVED_KEYWORDS.has(keyword) ? true : this.isKeywordIt(keyword);
    }
};
{
    const allTokens = Object.values(TokenType);
    const set = (set, start, end) => {
        const list = allTokens.slice(allTokens.indexOf(start) + 1, allTokens.indexOf(end));
        for (let key of list)
            set.add(key);
    };
    const SYMBOL_CHAR = new Set();
    set(SYMBOL_CHAR, TokenType.START_SYMBOL_CHARS, TokenType.END_SYMBOL_CHARS);
    set(langChars.MULTIPLE_CHAR_FIRST_IS_ALPHA, TokenType.END_SYMBOL_CHARS, TokenType.END_FIRST_CHAR_IS_ALPHA);
    for (let key of SYMBOL_CHAR)
        langChars.RESERVED_KEYWORDS.add(key);
    for (let key of langChars.MULTIPLE_CHAR_FIRST_IS_ALPHA)
        langChars.RESERVED_KEYWORDS.add(key);
    for (let key of SYMBOL_CHAR) {
        let node = langChars.SYMBOL_CHAR_TREE;
        for (let letter of key.split("")) {
            if (!(letter in node))
                node[letter] = {};
            node = node[letter];
        }
    }
    // Token sanity check, typescript allows "enum" with keys with same value
    const tokens = new Set();
    for (let value of Object.values(TokenType)) {
        if (tokens.has(value))
            throw Error(`Duplicated token '${value}', check TokenType enum`);
        else
            tokens.add(value);
    }
}
class LangError extends Error {
    constructor(msg) {
        super(msg);
        Object.setPrototypeOf(this, new.target.prototype);
    }
}
class LexerError extends LangError {
    constructor(msg) { super(msg); }
}
class ParserError extends LangError {
    constructor(msg) { super(msg); }
}
class RuntimeError extends LangError {
    constructor(msg) { super(msg); }
}
class Token {
    constructor(type, value) {
        this.type = type;
        this.value = value;
    }
    toString() { return `Token(type='${this.type}', value='${this.value}')`; }
    is(...tokenTypes) {
        for (let type of tokenTypes)
            if (type == this.type)
                return true;
        return false;
    }
    copy() {
        return new Token(this.type, this.value);
    }
}
class Lexer {
    constructor(text, directory, file) {
        this.text = text;
        this.directory = directory;
        this.file = file;
        this.absolutePosition = -1;
        this.positionInLine = -1;
        this.line = 0;
        this.character = "";
        this.prev_absolutePosition = -1;
        this.prev_positionInLine = -1;
        this.prev_line = 0;
        this.prev_character = "";
        this.savedState = {};
        this.path = `${directory}/${file}`;
        this.advance();
    }
    getStateCopy() {
        return {
            absolutePosition: this.absolutePosition,
            positionInLine: this.positionInLine,
            line: this.line,
            character: this.character,
            prev_absolutePosition: this.prev_absolutePosition,
            prev_positionInLine: this.prev_positionInLine,
            prev_line: this.prev_line,
            prev_character: this.prev_character,
        };
    }
    setState(state) {
        for (let prop in state)
            this[prop] = state[prop];
    }
    formatErrorMessage(msg, errorLength = 1) {
        return `${msg}\n\n` +
            this.getCurrentLineTextWithErrorPosition(errorLength) + "\n" +
            `(${this.path}:${this.prev_line + 1}:${this.prev_positionInLine})\n`;
    }
    formatErrorMessageLexer(msg, errorLength = 1) {
        return `${msg}\n\n` +
            this.getCurrentLineTextWithErrorPositionLexer(errorLength) + "\n" +
            `(${this.path}:${this.line + 1}:${this.positionInLine})\n`;
    }
    error(msg, errorLength = 1) {
        const formattedMsg = this.formatErrorMessageLexer(msg, errorLength);
        throw new LexerError(extSystem.formatColor("Lexer error", extSystem.colors.red) + "\n" +
            extSystem.formatColor(" ■ ", extSystem.colors.red) + formattedMsg);
    }
    getCurrentLineTextWithErrorPosition(errorLength = 1) {
        const spaceCount = Math.max(0, this.prev_positionInLine - errorLength);
        const underlineSize = Math.max(1, errorLength);
        const splitText = this.text.split("\n");
        const startLine = Math.max(0, this.prev_line - 7);
        const context = splitText
            .slice(startLine, Math.max(0, this.prev_line + 1))
            .map((t, i) => `${startLine + i + 1}`.padEnd(6) + t)
            .join("\n");
        const underLine = extSystem.formatColor("═".repeat(underlineSize), extSystem.colors.red);
        return context + "\n" +
            "".padEnd(6) + " ".repeat(spaceCount) + underLine;
    }
    getCurrentLineTextWithErrorPositionLexer(errorLength = 1) {
        const spaceCount = Math.max(0, this.prev_positionInLine - errorLength);
        const underlineSize = Math.max(1, errorLength);
        const splitText = this.text.split("\n");
        const startLine = Math.max(0, this.line - 7);
        const context = splitText
            .slice(startLine, Math.max(0, this.line + 1))
            .map((t, i) => `${startLine + i + 1}`.padEnd(6) + t)
            .join("\n");
        const underLine = extSystem.formatColor("═".repeat(underlineSize), extSystem.colors.red);
        return context + "\n" +
            "".padEnd(6) + " " + " ".repeat(spaceCount) + underLine;
    }
    advance() {
        this.prev_absolutePosition = this.absolutePosition;
        this.prev_positionInLine = this.positionInLine;
        this.prev_line = this.line;
        this.prev_character = this.character;
        if (this.prev_character == "\n") {
            this.positionInLine = 0;
            this.line += 1;
        }
        this.absolutePosition += 1;
        this.positionInLine += 1;
        this.character = this.absolutePosition < this.text.length ? this.text[this.absolutePosition] : "";
    }
    peekCharacter(next) {
        var _a;
        const pos = this.absolutePosition + next;
        return (_a = this.text[pos]) !== null && _a !== void 0 ? _a : "";
    }
    skipWhitespace() {
        while (this.character == " " || this.character == "\t" || this.character == "\r")
            this.advance();
    }
    isNewLine(char) { return char == "\n" || char == "\r"; }
    skipLineComment() {
        if (this.character == "/" && this.peekCharacter(1) == "/") {
            const isEnd = () => this.isNewLine(this.character) || this.character == "";
            this.advance();
            this.advance();
            while (!isEnd())
                this.advance();
        }
    }
    skipMultilineComment() {
        if (this.character == "/" && this.peekCharacter(1) == "*") {
            const isEnd = () => (this.character == "*" && this.peekCharacter(1) == "/") || this.character == "";
            this.advance();
            this.advance();
            while (!isEnd())
                this.advance();
            this.advance();
            this.advance();
        }
    }
    isDigit(char) { return /\d/.test(char); }
    isAlpha(char) { return /[A-Za-z]/.test(char); }
    isAlphaNum(char) { return /[A-Za-z0-9_]/.test(char); }
    isWhitespace(char) { return /\s/.test(char); }
    getNumber() {
        let output = this.character;
        this.advance();
        while (this.isDigit(this.character)) {
            output += this.character;
            this.advance();
        }
        if (this.character === ".") {
            output += this.character;
            this.advance();
            if (!this.isDigit(this.character))
                this.error(`Invalid character '${this.character}' in decimal number expression after . (dot), expected at least one digit.`);
            while (this.isDigit(this.character)) {
                output += this.character;
                this.advance();
            }
        }
        return output;
    }
    getAlphaNum() {
        let output = this.character;
        this.advance();
        while (this.isAlphaNum(this.character)) {
            output += this.character;
            this.advance();
        }
        return output;
    }
    /**
     * Looks for entry string characters then
     * entry interpolation or terminal string character.
     */
    getStringStart() {
        let output = this.character;
        const startLine = this.line;
        const startPositionInLine = this.positionInLine;
        while (true) {
            this.advance();
            if (this.character == "\\" && this.peekCharacter(1) == "n") {
                output += "\n";
                this.advance();
            }
            else
                output += this.character;
            if (this.character == TokenType.DOUBLE_QUOTE)
                break;
            if (this.character == "{" && this.peekCharacter(-1) == "$")
                break;
            if (this.character == "")
                break;
        }
        if (this.character == "")
            this.error(`Invalid string expression, ` +
                `reached end of file but didn't close string expression, ` +
                `must end in '"' or '\${'. ` +
                `String starting location at ${startLine}:${startPositionInLine}.`, 1);
        this.advance();
        return output;
    }
    /**
     Expects to be called after a TokenType.RBRACKET
     Checks for previous terminal interpolation string character then
     looks for entry interpolation or terminal string character.
     */
    getStringContinuation() {
        if (this.peekCharacter(-1) != "}")
            this.error("getStringContinuation only can be run after } character.", 1);
        let output = "";
        const startLine = this.line;
        const startPositionInLine = this.positionInLine;
        while (true) {
            if (this.character == "\\" && this.peekCharacter(1) == "n") {
                output += "\n";
                this.advance();
                this.advance();
                continue;
            }
            output += this.character;
            if (this.character == TokenType.DOUBLE_QUOTE)
                break;
            if (this.character == "{" && this.peekCharacter(-1) == "$")
                break;
            if (this.character == "")
                break;
            this.advance();
        }
        if (this.character == "")
            this.error(`Invalid string expression, ` +
                `reached end of file but didn't close string expression, ` +
                `must end in '"' or '\${'. ` +
                `String starting location at ${startLine}:${startPositionInLine}`, 1);
        this.advance();
        return output;
    }
    peekToken(count = 1) {
        const state = this.getStateCopy();
        for (let i = 0; i < count - 1; i++)
            this.getNextToken();
        const token = this.getNextToken();
        this.setState(state);
        return token;
    }
    // Called by the parser after string interpolation } token
    getNextTokenStringContinuation() {
        return new Token(TokenType.STRING, this.getStringContinuation());
    }
    symbolSeek() {
        let symbol = "";
        let node = langChars.SYMBOL_CHAR_TREE;
        while (this.character in node) {
            node = node[this.character];
            symbol += this.character;
            this.advance();
        }
        return new Token(symbol, symbol);
    }
    getNextToken() {
        this.skipWhitespace();
        this.skipLineComment();
        this.skipWhitespace();
        this.skipMultilineComment();
        this.skipWhitespace();
        if (this.character === "\n") {
            this.advance();
            return new Token(TokenType.EOL, TokenType.EOL);
        }
        if (this.character === "") {
            this.advance();
            return new Token(TokenType.EOF, TokenType.EOF);
        }
        if (this.character == TokenType.DOUBLE_QUOTE)
            return new Token(TokenType.STRING, this.getStringStart());
        if (this.character in langChars.SYMBOL_CHAR_TREE)
            return this.symbolSeek();
        if (this.isDigit(this.character))
            return new Token(TokenType.REAL, this.getNumber());
        if (this.isAlpha(this.character)) {
            const id = this.getAlphaNum();
            if (langChars.MULTIPLE_CHAR_FIRST_IS_ALPHA.has(id))
                return new Token(id, id);
            else
                return new Token(TokenType.ID, id);
        }
        this.error(`Invalid character: '${this.character}'.`);
    }
}
class ASTNode {
    toString() {
        return "\n" + this.constructor.name + " " + JSON.stringify(this, null, 2);
    }
}
class NoOp extends ASTNode {
}
class ID extends ASTNode {
    constructor(op, name = op.value) {
        super();
        this.op = op;
        this.name = name;
    }
}
// They are shims, in reality they are converted to javascript:
// string, number, boolean, undefined, null types
class StringOp extends ASTNode {
    constructor(strings, interpolations) {
        super();
        this.strings = strings;
        this.interpolations = interpolations;
    }
}
class Real extends ASTNode {
    constructor(op) {
        super();
        this.op = op;
    }
}
class Bool extends ASTNode {
    constructor(op) {
        super();
        this.op = op;
    }
}
class Undefined extends ASTNode {
}
class Null extends ASTNode {
}
class List extends ASTNode {
    constructor(nodes) {
        super();
        this.nodes = nodes;
    }
    toString() {
        return this.toStringIdent(0);
    }
    toStringIdent(ident) {
        const repr = (v, disp) => v instanceof Dictionary ?
            v.toStringIdent(disp) :
            v instanceof List ?
                "(" + v.toStringIdent(disp) + ")" :
                v.toString();
        return this.nodes.map(v => repr(v, ident + 2)).join(", ");
    }
    sumScalar(scalar) { return new List(this.nodes.map(v => v + scalar)); }
    restScalar(scalar) { return new List(this.nodes.map(v => v - scalar)); }
    multScalar(scalar) { return new List(this.nodes.map(v => v * scalar)); }
    divScalar(scalar) { return new List(this.nodes.map(v => v / scalar)); }
    modScalar(scalar) { return new List(this.nodes.map(v => v % scalar)); }
    powScalar(scalar) { return new List(this.nodes.map(v => Math.pow(v, scalar))); }
    restScalarRL(scalarL) { return new List(this.nodes.map(v => scalarL - v)); }
    divScalarRL(scalarL) { return new List(this.nodes.map(v => scalarL / v)); }
    modScalarRL(scalarL) { return new List(this.nodes.map(v => scalarL % v)); }
    powScalarRL(scalarL) { return new List(this.nodes.map(v => Math.pow(scalarL, v))); }
    sumList(vector) { return new List(this.nodes.map((v, i) => v + vector.nodes[i])); }
    restList(vector) { return new List(this.nodes.map((v, i) => v - vector.nodes[i])); }
    multList(vector) { return new List(this.nodes.map((v, i) => v * vector.nodes[i])); }
    divList(vector) { return new List(this.nodes.map((v, i) => v / vector.nodes[i])); }
    modList(vector) { return new List(this.nodes.map((v, i) => v % vector.nodes[i])); }
    powList(vector) { return new List(this.nodes.map((v, i) => Math.pow(v, vector.nodes[i]))); }
}
// Used so when the List is visited it doesn't cause an infinite loop
class ListEval extends List {
}
class Dictionary extends ASTNode {
    constructor(nodes) {
        super();
        this.nodes = nodes;
    }
    toString() {
        return this.toStringIdent(0);
    }
    toStringIdent(ident) {
        const list = Object.keys(this.nodes);
        const first = ident == 0;
        const disp = 2;
        const repr = (v, disp) => v instanceof Dictionary ?
            v.toStringIdent(disp) :
            v instanceof List ?
                v.toStringIdent(disp) :
                v.toString();
        if (list.length <= 0)
            return "{ " +
                list.map(k => `${k} = ${repr(this.nodes[k], ident + 2)}`).join() +
                " }";
        const space = " ".repeat(ident);
        const space2 = " ".repeat(ident + disp);
        return "{\n" +
            list.map(k => space2 + `${k} = ${repr(this.nodes[k], ident + 2)}`).join("\n") + "\n" +
            (first ? "" : space) + "}";
    }
}
class BlockOp extends ASTNode {
    constructor(list) {
        super();
        this.list = list;
    }
}
class UnaryOp extends ASTNode {
    constructor(op, right) {
        super();
        this.op = op;
        this.right = right;
    }
}
class BinaryOp extends ASTNode {
    constructor(left, op, right) {
        super();
        this.left = left;
        this.op = op;
        this.right = right;
    }
}
class ShortcircuitOp extends ASTNode {
    constructor(left, op, right) {
        super();
        this.left = left;
        this.op = op;
        this.right = right;
    }
}
class ComparisonOp extends ASTNode {
    constructor(nodes, comparators) {
        super();
        this.nodes = nodes;
        this.comparators = comparators;
    }
}
class CompoundOp extends ASTNode {
    constructor(left, op, right) {
        super();
        this.left = left;
        this.op = op;
        this.right = right;
    }
}
class VariableIncrementPrefixOp extends ASTNode {
    constructor(op, right) {
        super();
        this.op = op;
        this.right = right;
    }
}
class VariableIncrementPostfixOp extends ASTNode {
    constructor(op, left) {
        super();
        this.op = op;
        this.left = left;
    }
}
class SpreadListOp extends ASTNode {
    constructor(list) {
        super();
        this.list = list;
    }
}
class AssignOp extends ASTNode {
    constructor(id, expression) {
        super();
        this.id = id;
        this.expression = expression;
    }
}
class MultipleAssignOp extends ASTNode {
    constructor(ids, expression) {
        super();
        this.ids = ids;
        this.expression = expression;
    }
}
class PipeNodeOp extends ASTNode {
    constructor(expression, passType) {
        super();
        this.expression = expression;
        this.passType = passType;
    }
}
class PipeOp extends ASTNode {
    constructor(entry, nodes) {
        super();
        this.entry = entry;
        this.nodes = nodes;
    }
}
class RangeOp extends ASTNode {
    constructor(start, step, end) {
        super();
        this.start = start;
        this.step = step;
        this.end = end;
    }
}
class ForLoopOp extends ASTNode {
    constructor(setup, runCondition, increment, body) {
        super();
        this.setup = setup;
        this.runCondition = runCondition;
        this.increment = increment;
        this.body = body;
    }
}
class Conditional extends ASTNode {
    constructor(conditions, bodies) {
        super();
        this.conditions = conditions;
        this.bodies = bodies;
    }
}
class ReturnOp extends ASTNode {
    constructor(node) {
        super();
        this.node = node;
    }
}
class BaseFunction extends ASTNode {
    constructor(stackNamespace, name, parameters) {
        super();
        this.stackNamespace = stackNamespace;
        this.name = name;
        this.parameters = parameters;
    }
}
class NamedFunction extends BaseFunction {
    constructor(stackNamespace, name, parameters, block) {
        super(stackNamespace, name, parameters);
        this.stackNamespace = stackNamespace;
        this.name = name;
        this.parameters = parameters;
        this.block = block;
    }
}
class AnonymousFunction extends BaseFunction {
    constructor(stackNamespace, parameters, block) {
        super(stackNamespace, "", parameters);
        this.stackNamespace = stackNamespace;
        this.parameters = parameters;
        this.block = block;
    }
}
class NativeFunction extends BaseFunction {
    constructor(stackNamespace, name, parameters, fun) {
        super(stackNamespace, name, parameters);
        this.stackNamespace = stackNamespace;
        this.name = name;
        this.parameters = parameters;
        this.fun = fun;
    }
}
class FunctionCall extends ASTNode {
    constructor(name, args) {
        super();
        this.name = name;
        this.args = args;
    }
}
class PrintselfCall extends ASTNode {
    constructor(args, text) {
        super();
        this.args = args;
        this.text = text;
    }
}
class AssertCall extends ASTNode {
    constructor(args, text, file, line, column) {
        super();
        this.args = args;
        this.text = text;
        this.file = file;
        this.line = line;
        this.column = column;
    }
}
class CallOp extends ASTNode {
    constructor(func, args) {
        super();
        this.func = func;
        this.args = args;
    }
}
class AccessIndexOp extends ASTNode {
    constructor(list, index) {
        super();
        this.list = list;
        this.index = index;
    }
}
class AccessKeyOp extends ASTNode {
    constructor(object, key) {
        super();
        this.object = object;
        this.key = key;
    }
}
class AccessMethodOp extends ASTNode {
    constructor(object, key) {
        super();
        this.object = object;
        this.key = key;
    }
}
class NamedImportOp extends ASTNode {
    constructor(moduleAST, moduleSymbol, moduleAlias, filePath) {
        super();
        this.moduleAST = moduleAST;
        this.moduleSymbol = moduleSymbol;
        this.moduleAlias = moduleAlias;
        this.filePath = filePath;
    }
}
class NamedImportLoadOp extends ASTNode {
    constructor(moduleSymbol, moduleAlias) {
        super();
        this.moduleSymbol = moduleSymbol;
        this.moduleAlias = moduleAlias;
    }
}
function writeGlobalUtils(stack) {
    const _print = new NativeFunction(stack.symbol, "print", ["msg"].map(name => new ID(new Token(TokenType.ID, name))), (msg) => { extSystem.print(`${msg}`); });
    const printself = new NativeFunction(stack.symbol, "printself", ["msg"].map(name => new ID(new Token(TokenType.ID, name))), (msg) => { extSystem.print(`${msg}`); });
    const assert = new NativeFunction(stack.symbol, "assert", ["msg"].map(name => new ID(new Token(TokenType.ID, name))), (msg) => { extSystem.print(`${msg}`); });
    const _printToFile = new NativeFunction(stack.symbol, "printToFile", ["data", "file"].map(name => new ID(new Token(TokenType.ID, name))), (data, file) => { extSystem.writeFile(`${data}`, file, undefined); });
    stack.set(_print.name, _print);
    stack.set(printself.name, printself);
    stack.set(assert.name, assert);
    stack.set(_printToFile.name, _printToFile);
}
class StackRecord {
    constructor() {
        this.memory = new Map();
    }
    get(id) { return this.memory.get(id); }
    set(id, value) { this.memory.set(id, value); }
    has(id) { return this.memory.has(id); }
}
class Stack {
    constructor(pathFile, symbol) {
        this.pathFile = pathFile;
        this.symbol = symbol;
        this.stackRecordList = [];
        this.push();
        writeGlobalUtils(this);
    }
    static reset() {
        Stack.namespaceGlobalCount = 0;
        Stack.absoluteFilePath_to_symbol = new Map();
        Stack.importModules = new Map();
    }
    isGlobalScope() { return this.stackRecordList.length == 1; }
    getGlobalVariablesKeys() {
        return [...this.stackRecordList[0].memory.keys()];
    }
    push() {
        const stackRecord = new StackRecord();
        this.stackRecordList.push(stackRecord);
        this.activeRecord = stackRecord;
    }
    pop() {
        this.stackRecordList.pop();
        this.activeRecord = this.stackRecordList[this.stackRecordList.length - 1];
    }
    get(id) {
        for (let i = this.stackRecordList.length - 1; i > -1; i--)
            if (this.stackRecordList[i].has(id))
                return this.stackRecordList[i].get(id);
    }
    has(id) {
        for (let i = this.stackRecordList.length - 1; i > -1; i--)
            if (this.stackRecordList[i].has(id))
                return true;
        return false;
    }
    hasInSameRecord(id) {
        return this.activeRecord.has(id);
    }
    set(id, value) {
        this.activeRecord.set(id, value);
    }
    setOnExisting(id, value) {
        for (let i = this.stackRecordList.length - 1; i > -1; i--)
            if (this.stackRecordList[i].has(id))
                this.stackRecordList[i].set(id, value);
    }
    hasOwn(id) {
        return this.activeRecord.has(id);
    }
    getOwnIts() {
        return [...this.activeRecord.memory.keys()].filter(v => langChars.isKeywordIt(v));
    }
}
// Modified on the parser stage only (when loading a new module)
Stack.namespaceGlobalCount = 0;
Stack.absoluteFilePath_to_symbol = new Map();
// Modified on the parse(for checking) and interpreter stage (overrites parse data) only
Stack.importModules = new Map();
/* Extra grammar notation
   · [TOKEN] The token is not to be consumed (not "eat" call) in the currently parsing call
   · {call} The call is to be run in a new child stack scope (push > call > pop) in the currently running call
   · func_a< func_b > Signifies the composition of functions, func_a( () => func_b() )
   · $var Make reference to a local varaible o function
*/
var EatMode;
(function (EatMode) {
    EatMode[EatMode["Default"] = 0] = "Default";
    EatMode[EatMode["StringContinuation"] = 1] = "StringContinuation";
})(EatMode || (EatMode = {}));
class Parser {
    constructor(lexer, stack) {
        this.lexer = lexer;
        this.stack = stack;
        this.token = this.lexer.getNextToken();
    }
    // Doesn't save stack !!!
    getState() {
        return {
            token: this.token.copy(),
            lexerState: this.lexer.getStateCopy()
        };
    }
    // Doesn't restore stack !!!
    restoreState(savedState) {
        this.token.type = savedState.token.type;
        this.token.value = savedState.token.value;
        this.lexer.setState(savedState.lexerState);
    }
    peekToken(count = 1) {
        return this.lexer.peekToken(count);
    }
    is(...tokenTypes) {
        for (let type of tokenTypes)
            if (type == this.token.type)
                return true;
        return false;
    }
    isChain(...tokenTypes) {
        if (!this.is(tokenTypes[0]))
            return false;
        for (let i = 1; i < tokenTypes.length; i++)
            if (tokenTypes[i] !== this.peekToken(i).type)
                return false;
        return true;
    }
    error(msg, errorLength = 1) {
        const formattedMsg = this.lexer.formatErrorMessage(msg, errorLength);
        throw new ParserError(extSystem.formatColor("Parser error", extSystem.colors.red) + "\n" +
            extSystem.formatColor(" ■ ", extSystem.colors.red) + formattedMsg);
    }
    eat(tokenType, mode = EatMode.Default) {
        if (this.is(tokenType)) {
            if (mode == EatMode.Default)
                this.token = this.lexer.getNextToken();
            else if (mode == EatMode.StringContinuation)
                this.token = this.lexer.getNextTokenStringContinuation();
            else
                this.error(`Invalid EatMode ${mode}`);
        }
        else
            this.error(`Invalid token ${this.token}, expected token type '${tokenType}'.`, this.token.value.length);
        // log("EAT", tokenType.padEnd(5), "GET", this.token, "MODE ", mode)
    }
    eatEOL(optional = true) {
        if (!optional)
            this.eat(TokenType.EOL);
        else if (this.is(TokenType.EOL))
            this.eat(TokenType.EOL);
    }
    eatEOLS() {
        while (this.is(TokenType.EOL))
            this.eat(TokenType.EOL);
    }
    anonymousFunctionParameters() {
        // (ID (COMMA ID)* ARROW_RIGHT)?
        let parameters = [];
        let ids = new Set();
        if (this.is(TokenType.ID)) {
            const parameter = new ID(this.token);
            this.eat(TokenType.ID);
            parameters.push(parameter);
            ids.add(parameter.name);
            while (this.is(TokenType.COMMA)) {
                this.eat(TokenType.COMMA);
                const parameter = new ID(this.token);
                if (ids.has(parameter.name))
                    this.error(`Duplicate identifier '${parameter.name}'.`, parameter.name.length);
                this.eat(TokenType.ID);
                parameters.push(parameter);
            }
            this.eat(TokenType.ARROW_RIGHT);
        }
        return parameters;
    }
    hasAnonymousFunctionParameters() {
        // (ID (COMMA ID)* ARROW_RIGHT)?
        if (!this.is(TokenType.ID))
            return false;
        const initialState = this.getState();
        this.eat(TokenType.ID);
        while (this.is(TokenType.COMMA)) {
            this.eat(TokenType.COMMA);
            if (!this.is(TokenType.ID))
                break;
            this.eat(TokenType.ID);
        }
        const hasParameters = this.is(TokenType.ARROW_RIGHT);
        this.restoreState(initialState);
        return hasParameters;
    }
    anonymousFunction(noParams = false) {
        // {LBRACKET anonymousFunctionParameters blockTo  RBRACKET}
        this.stack.push();
        this.eat(TokenType.LBRACKET);
        let hasParams = this.hasAnonymousFunctionParameters();
        if (noParams && hasParams)
            this.error(`This anonymous function can't have parameters.`);
        let parameters = hasParams ? this.anonymousFunctionParameters() : [];
        for (let parameter of parameters)
            this.stack.set(parameter.name, undefined);
        const block = this.blockTo(TokenType.RBRACKET);
        const its = this.stack.getOwnIts();
        if (parameters.length && its.length)
            this.error(`Not allowed to have explicit parameters and also use special 'it' variables ${its}.`);
        if (its.length) {
            const sorted = its.map(v => v.slice(2)).map(v => v == "" ? 0 : parseInt(v)).sort();
            const nOfIts = sorted[sorted.length - 1] + 1;
            parameters = new Array(nOfIts).fill(0).
                map((_, i) => i == 0 ? "it" : `it${i}`).
                map((v) => new ID(new Token(TokenType.ID, v)));
        }
        this.stack.pop();
        return new AnonymousFunction(this.stack.isGlobalScope() ? this.stack.symbol : -1, parameters, block);
    }
    pipeExpression() {
        // (MAP|FILTER)? ([LBRACKET] anonymousFunction  | ID access)
        const passType = this.is(TokenType.MAP, TokenType.FILTER) ? this.token.type : null;
        if (passType != null)
            this.eat(passType);
        if (this.is(TokenType.LBRACKET))
            return new PipeNodeOp(this.anonymousFunction(), passType);
        if (this.is(TokenType.ID)) {
            const id = new ID(this.token);
            if (!this.stack.has(id.name))
                throw this.error(`Function '${id.name}' not defined.`, id.name.length);
            const fun = this.stack.get(id.name);
            // Dynamic (error if any will be in the runtime stage)
            if (fun instanceof Dictionary || fun instanceof List || fun instanceof Stack) {
                const dynamicFun = this.accessor(1);
                return new PipeNodeOp(dynamicFun, passType);
            }
            if (!(fun instanceof BaseFunction))
                throw this.error(`Identifier '${id.name}' is not a function.`, id.name.length);
            this.eat(TokenType.ID);
            return new PipeNodeOp(new FunctionCall(fun.name, new List([])), passType);
        }
        if (this.is(TokenType.LBRACKET))
            this.error(`Anonymous function bracket must always start in the same line as the pipe operator.`, this.token.value.length);
        this.error(`Invalid token ${this.token} for pipe node expression.`, this.token.value.length);
    }
    // Pipe horizontal travserse with optional accessors
    expression(nVals = undefined) {
        // listExpression (PIPE EOL? ( isAccessor getAccessor | pipeExpression))*
        let expression = this.listExpression(nVals);
        if (!this.is(TokenType.PIPE))
            return expression;
        while (this.is(TokenType.PIPE)) {
            const pipeNodes = [];
            while (this.is(TokenType.PIPE)) {
                this.eat(TokenType.PIPE);
                this.eatEOL();
                if (this.isAccessor())
                    break;
                pipeNodes.push(this.pipeExpression());
            }
            if (pipeNodes.length == 0)
                expression = this.getAccessors(expression);
            else
                expression = this.getAccessors(new PipeOp(expression, pipeNodes));
        }
        return expression;
    }
    // An expression can return multiple values
    listExpression(nVals = undefined) {
        // DOT_THREE? disjunction (COMMA (EOL? disjunction))*
        const expects = nVals !== undefined;
        const expandList = false; // Buggy don't
        // const expandList = this.is(TokenType.DOT_THREE)
        // if (expandList) this.eat(TokenType.DOT_THREE)
        let nodes = [];
        nodes.push(this.disjunction(nVals));
        let makeList = false;
        while (this.is(TokenType.COMMA)) {
            if (expects && nodes.length >= nVals)
                this.error(`Expected ${nVals} expressions but got ${nodes.length + 1}. ${this.token}`, this.token.value.length);
            if (this.isChain(TokenType.COMMA, TokenType.RPARE)) {
                this.eat(TokenType.COMMA);
                makeList = true;
                break;
            }
            this.eat(TokenType.COMMA);
            this.eatEOL();
            const node = this.disjunction(nVals);
            nodes.push(node);
        }
        if (expects && nodes.length < nVals)
            this.error(`Expected ${nVals} expressions but got ${nodes.length}. ${this.token}`, this.token.value.length);
        let res = nodes.length == 1 ? nodes[0] : new List(nodes);
        if (makeList)
            res = new List([res]);
        if (expandList)
            res = new SpreadListOp(res);
        return res;
    }
    disjunction(nVals = undefined) {
        // conjunction (OR EOL? conjunction)*
        let node = this.conjunction(nVals);
        while (this.is(TokenType.OR)) {
            const token = this.token;
            this.eat(token.type);
            this.eatEOL();
            node = new ShortcircuitOp(node, token, this.conjunction(nVals));
        }
        return node;
    }
    conjunction(nVals = undefined) {
        // comparision ( AND EOL? comparision)*
        let node = this.comparision(nVals);
        while (this.is(TokenType.AND)) {
            const token = this.token;
            this.eat(token.type);
            this.eatEOL();
            node = new ShortcircuitOp(node, token, this.comparision(nVals));
        }
        return node;
    }
    comparision(nVals = undefined) {
        // additive (( < | > | <= | >= | == | != ) EOL? additive)*
        let nodes = [];
        let comparators = [];
        nodes.push(this.additive(nVals));
        while (this.is(TokenType.LESS, TokenType.GREATER, TokenType.EQUAL, TokenType.NOT_EQUAL, TokenType.LESS_OR_EQUAL, TokenType.GREATER_OR_EQUAL)) {
            const token = this.token;
            this.eat(token.type);
            this.eatEOL();
            comparators.push(token);
            nodes.push(this.additive(nVals));
        }
        if (nodes.length == 1)
            return nodes[0];
        return new ComparisonOp(nodes, comparators);
    }
    additive(nVals = undefined) {
        // multiplicative ((PLUS|MINUS) EOL? multiplicative)*
        let node = this.multiplicative(nVals);
        while (this.is(TokenType.PLUS, TokenType.MINUS)) {
            const token = this.token;
            this.eat(token.type);
            this.eatEOL();
            node = new BinaryOp(node, token, this.multiplicative(nVals));
        }
        return node;
    }
    multiplicative(nVals = undefined) {
        // power ((MUL|DIV|MODULO) EOL? power)*
        let node = this.power(nVals);
        while (this.is(TokenType.MUL, TokenType.DIV, TokenType.MODULO)) {
            const token = this.token;
            this.eat(token.type);
            this.eatEOL();
            node = new BinaryOp(node, token, this.power(nVals));
        }
        return node;
    }
    power(nVals = undefined) {
        // accessor (POWER accessor)*
        let node = this.accessor(nVals);
        while (this.is(TokenType.POWER)) {
            const token = this.token;
            this.eat(token.type);
            this.eatEOL();
            node = new BinaryOp(node, token, this.accessor(nVals));
        }
        return node;
    }
    isAccessor() {
        return this.is(TokenType.DOT, TokenType.LBRACKETSQR, TokenType.LPARE, TokenType.QUESTION);
    }
    getAccessors(node) {
        /*
            (
                DOT (ID|REAL) |
                LBRACKETSQR EOL? expression EOL? RBRACKETSQR |
                LPARE EOLS? (expression EOLS?)? RPARE |
                QUESTION ID (parameters)?
            )*
        */
        while (this.isAccessor()) {
            if (this.is(TokenType.LBRACKETSQR)) {
                this.eat(TokenType.LBRACKETSQR);
                this.eatEOL();
                node = new AccessIndexOp(node, this.expression());
                this.eatEOL();
                this.eat(TokenType.RBRACKETSQR);
            }
            else if (this.is(TokenType.DOT)) {
                this.eat(TokenType.DOT);
                const key = this.token.value;
                if (!this.is(TokenType.ID, TokenType.REAL))
                    this.error("Invalid token, access key needs to be a string or integer number.");
                if (this.is(TokenType.REAL) && key.includes("."))
                    this.error("Access key invalid: can't be a number with decimals.");
                this.eat(this.token.type);
                node = new AccessKeyOp(node, key);
            }
            else if (this.is(TokenType.LPARE)) {
                this.eat(TokenType.LPARE);
                this.eatEOLS();
                const isEmpty = this.is(TokenType.RPARE);
                const parameters = isEmpty ? new List([]) : this.expression();
                this.eatEOLS();
                this.eat(TokenType.RPARE);
                node = new CallOp(node, parameters);
            }
            else if (this.is(TokenType.QUESTION)) {
                this.eat(TokenType.QUESTION);
                const id = new ID(this.token);
                this.eat(TokenType.ID);
                node = new AccessMethodOp(node, id.name);
            }
        }
        return node;
    }
    accessor(nVals = undefined) {
        /* range getAccessors
        */
        let node = this.range(nVals);
        return this.getAccessors(node);
    }
    range(nVals = undefined) {
        // factor (: factor (: factor)?)?
        const node1 = this.factor(nVals);
        if (!this.is(TokenType.COLON))
            return node1;
        this.eat(TokenType.COLON);
        const node2 = this.factor();
        if (!this.is(TokenType.COLON))
            return new RangeOp(node1, null, node2);
        this.eat(TokenType.COLON);
        const node3 = this.factor();
        return new RangeOp(node1, node2, node3);
    }
    factor(nVals = undefined) {
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
        if (this.is(TokenType.PLUSPLUS, TokenType.MINUSMINUS)) {
            const token = this.token;
            this.eat(token.type);
            const id = new ID(this.token);
            this.eat(TokenType.ID);
            return new VariableIncrementPrefixOp(token, id);
        }
        if (this.is(TokenType.PLUS, TokenType.MINUS)) {
            const token = this.token;
            this.eat(token.type);
            this.eatEOL();
            return new UnaryOp(token, this.factor(nVals));
        }
        if (this.is(TokenType.PLUS, TokenType.MINUS)) {
            const token = this.token;
            this.eat(token.type);
            this.eatEOL();
            return new UnaryOp(token, this.factor(nVals));
        }
        if (this.is(TokenType.NOT)) {
            const token = this.token;
            this.eat(token.type);
            return new UnaryOp(token, this.factor(nVals));
        }
        else if (this.is(TokenType.REAL)) {
            const token = this.token;
            this.eat(TokenType.REAL);
            return new Real(token);
        }
        else if (this.is(TokenType.TRUE, TokenType.FALSE)) {
            const token = this.token;
            this.eat(token.type);
            return new Bool(token);
        }
        else if (this.is(TokenType.ID)) {
            const id = new ID(this.token);
            if (id.name == "undefined") {
                this.eat(TokenType.ID);
                return new Undefined();
            }
            if (id.name == "null") {
                this.eat(TokenType.ID);
                return new Null();
            }
            const isIt = langChars.isKeywordIt(id.name);
            const isArguments = id.name == "arguments";
            // Special family of variables it, it1, it2, it3 , etc
            // that represent the scope arguments for anonymous functions.
            // They can't be reassigned or modified, only accessed
            if (id.name == "it0")
                this.error(`First 'it' value doesn't use index, try renaming 'it0' to 'it'.`, id.name.length);
            if (isIt || isArguments) {
                if (this.stack.isGlobalScope())
                    this.error(`Not allowed to access special variable '${id.name}' in the global scope.`, id.name.length);
                // Here, the name is "marked" so can be later known will be accessed
                this.stack.set(id.name, null);
            }
            if (!this.stack.has(id.name))
                this.error(`Trying to use undefined variable '${id.name}'.`, id.name.length);
            this.eat(TokenType.ID);
            if (this.is(TokenType.LPARE)) {
                if (isIt)
                    this.error(`Trying to use the variable '${id.name}' as a function.`, 1);
                return this.functionCall(id);
            }
            else if (this.is(TokenType.PLUSPLUS, TokenType.MINUSMINUS)) {
                if (isIt)
                    this.error(`Not allowed to modify special variable '${id.name}', only allowed to access.`, id.name.length);
                const token = this.token;
                this.eat(token.type);
                return new VariableIncrementPostfixOp(token, id);
            }
            else if (this.is(TokenType.LBRACKETSQR)) {
                this.eat(TokenType.LBRACKETSQR);
                this.eatEOLS();
                const node = this.expression();
                this.eatEOLS();
                this.eat(TokenType.RBRACKETSQR);
                return new AccessIndexOp(id, node);
            }
            else
                return id;
        }
        else if (this.is(TokenType.LPARE)) {
            this.eat(TokenType.LPARE);
            this.eatEOL();
            let node = this.expression();
            this.eatEOL();
            this.eat(TokenType.RPARE);
            return node;
        }
        else if (this.is(TokenType.IF)) {
            return this.conditionalExpression(nVals);
        }
        else if (this.is(TokenType.RETURN)) {
            return this.return(nVals);
        }
        else if (this.is(TokenType.STRING)) {
            return this.string();
        }
        else if (this.is(TokenType.DO)) {
            this.eat(TokenType.DO);
            return this.anonymousFunction(true);
        }
        else if (this.is(TokenType.LBRACKET)) {
            return this.dictionary();
        }
        this.error(`Expected expression here but found none. ${this.token}.`, 1);
    }
    dictionary() {
        // LBRACKET EOL? ( (ID|REAL) ASSIGN ( fun anonymousFunction | expression) (EOL|SEMICOLON)? EOLS? )* EOL? RBRACKET
        this.eat(TokenType.LBRACKET);
        this.eatEOL();
        const object = {};
        const ids = new Set();
        while (this.is(TokenType.ID, TokenType.REAL)) {
            if (this.is(TokenType.REAL) && this.token.value.includes("."))
                this.error("No decimal numbers allowed as dictionary keys.");
            const id = `${this.token.value}`;
            if (ids.has(id))
                this.error(`Duplicated key ${id}.`);
            else
                ids.add(id);
            this.eat(this.token.type);
            this.eat(TokenType.ASSIGN);
            const isFun = this.is(TokenType.FUNCTION);
            if (isFun)
                this.eat(TokenType.FUNCTION);
            const node = isFun ? this.anonymousFunction() : this.expression();
            if (this.is(TokenType.RBRACKET)) {
                object[id] = node;
                break;
            }
            if (this.is(TokenType.ID, TokenType.REAL)) {
                object[id] = node;
                continue;
            }
            if (this.is(TokenType.SEMICOLON))
                this.eat(TokenType.SEMICOLON);
            else
                this.eat(TokenType.EOL);
            this.eatEOLS();
            object[id] = node;
        }
        this.eatEOL();
        this.eat(TokenType.RBRACKET);
        return new Dictionary(object);
    }
    import() {
        // IMPORT ID ( DOT ID )* (AS ID)? EOL
        if (!this.stack.isGlobalScope())
            this.error(`Import can only be called in the global scope.`, this.token.value.length);
        this.eat(TokenType.IMPORT);
        const ids = this.doWhileEat(TokenType.DOT, () => {
            const token = this.token;
            this.eat(TokenType.ID);
            return new ID(token);
        });
        const list = ids.map(v => v.name);
        const relativeName = list.join(".");
        const relativeDir = list.slice(0, -1).join("/");
        const name = list[list.length - 1];
        const fileName = name + ".jk";
        const absoluteDir = this.lexer.directory + (relativeDir == "" ? "" : "/" + relativeDir);
        const absolutePath = absoluteDir + "/" + fileName;
        if (!extSystem.existFile(absolutePath))
            this.error(`Import '${relativeName}' doesn't exist. Failed to access file in ${absolutePath}.`, relativeName.length);
        let moduleAlias = name;
        if (this.is(TokenType.AS)) {
            this.eat(TokenType.AS);
            moduleAlias = this.token.value;
            this.eat(TokenType.ID);
            this.eat(TokenType.EOL);
        }
        if (this.stack.has(moduleAlias))
            this.error(`Trying to use duplicated name ${moduleAlias}.`, moduleAlias.length);
        // Module AST tree and stack are already parsed, so just add the name
        if (Stack.absoluteFilePath_to_symbol.has(absolutePath)) {
            const moduleSymbol = Stack.absoluteFilePath_to_symbol.get(absolutePath);
            const moduleStack = Stack.importModules.get(moduleSymbol);
            this.stack.set(moduleAlias, moduleStack);
            return new NamedImportLoadOp(moduleSymbol, moduleAlias);
        }
        const moduleSymbol = ++Stack.namespaceGlobalCount;
        Stack.absoluteFilePath_to_symbol.set(absolutePath, moduleSymbol);
        const moduleStack = new Stack(absolutePath, moduleSymbol);
        const rawText = extSystem.readFile(absolutePath);
        const lexer = new Lexer(rawText, absoluteDir, fileName);
        const parser = new Parser(lexer, moduleStack);
        const module_ast = parser.parse();
        this.stack.set(moduleAlias, moduleStack);
        Stack.importModules.set(moduleSymbol, parser.stack);
        return new NamedImportOp(module_ast, moduleSymbol, moduleAlias, absolutePath);
    }
    // Eats the token given !!
    doWhileEat(tokenType, fn) {
        const list = [fn()];
        while (this.is(tokenType)) {
            this.eat(tokenType);
            list.push(fn());
        }
        return list;
    }
    string() {
        // STRING ( EOLS? expression EOLS? RBRACKET STRING )*
        // string syntax example: "hello ${ callA() } ${ 2*1 } world"
        const extract = (raw) => {
            // no terminal: ${  terminal: "
            const isTerminal = raw[raw.length - 1] == '"';
            const start = raw[0] == `"` ? 1 : 0;
            const end = raw.length - (isTerminal ? 1 : 2);
            return {
                isTerminal: isTerminal,
                string: raw.slice(start, end)
            };
        };
        const listStrings = [];
        const listExpressions = [];
        const data = extract(this.token.value);
        this.eat(TokenType.STRING);
        listStrings.push(data.string);
        while (!data.isTerminal) {
            this.eatEOLS();
            listExpressions.push(this.expression(1));
            this.eatEOLS();
            this.eat(TokenType.RBRACKET, EatMode.StringContinuation);
            const data = extract(this.token.value);
            this.eat(TokenType.STRING);
            listStrings.push(data.string);
            if (data.isTerminal)
                break;
        }
        return new StringOp(listStrings, listExpressions);
    }
    optionalParenthesis(fn) {
        //  fn | LPARE EOLS? fn EOLS? RPARE
        if (this.is(TokenType.LPARE)) {
            this.eat(TokenType.LPARE);
            this.eatEOLS();
            const node = fn(true);
            this.eatEOLS();
            this.eat(TokenType.RPARE);
            return node;
        }
        return fn(false);
    }
    optionalBrackets(fn) {
        //  fn | LBRACKET EOLS? fn EOLS? RBRACKET
        if (this.is(TokenType.LBRACKET)) {
            this.eat(TokenType.LBRACKET);
            this.eatEOLS();
            const node = fn(true);
            this.eatEOLS();
            this.eat(TokenType.RBRACKET);
            return node;
        }
        return fn(true);
    }
    // Meant to always return something (must have terminal else expression)
    conditionalExpression(nVals = undefined) {
        return this.conditionalStatement(true, nVals);
    }
    blockOrExpression(nVals = undefined) {
        // [LBRACKET] block | expression
        if (this.is(TokenType.LBRACKET))
            return this.block(nVals);
        else
            return this.expression(nVals);
    }
    conditionalStatement(mustHaveElse = false, nVals = undefined) {
        /*
            IF optionalParenthesis<expression> EOL? blockOrExpression EOL?
            (ELIF optionalParenthesis<expression> EOL? blockOrExpression EOL?)*
            (ELSE EOL? blockOrExpression)?
        */
        let conditions = [];
        let bodies = [];
        this.eat(TokenType.IF);
        conditions.push(this.optionalParenthesis(() => this.expression(1)));
        this.eatEOL();
        bodies.push(this.is(TokenType.LBRACKET) ? this.block(nVals) : this.blockSubroutine(nVals));
        this.eatEOL();
        while (this.is(TokenType.ELIF)) {
            this.eat(TokenType.ELIF);
            conditions.push(this.optionalParenthesis(() => this.expression(1)));
            this.eatEOL();
            bodies.push(this.is(TokenType.LBRACKET) ? this.block(nVals) : this.blockSubroutine(nVals));
            this.eatEOL();
        }
        if (mustHaveElse || this.is(TokenType.ELSE)) {
            this.eat(TokenType.ELSE);
            this.eatEOL();
            bodies.push(this.is(TokenType.LBRACKET) ? this.block(nVals) : this.blockSubroutine(nVals));
        }
        return new Conditional(conditions, bodies);
    }
    functionCall(id) {
        // functionCallArgumnets
        const fun = this.stack.get(id.name);
        // Dynamic
        if (!(fun instanceof BaseFunction)) {
            this.eat(TokenType.LPARE);
            this.eatEOLS();
            const isEmpty = this.is(TokenType.RPARE);
            const parameters = isEmpty ? new List([]) : this.expression();
            this.eatEOLS();
            this.eat(TokenType.RPARE);
            return new CallOp(id, parameters);
        }
        if (id.name === "printself") {
            const startPos = this.lexer.absolutePosition;
            const args = this.functionCallArgumnets(fun);
            const endPos = this.lexer.absolutePosition;
            const text = this.lexer.text.slice(startPos - 1, endPos).trim().slice(1, -1);
            return new PrintselfCall(args, text);
        }
        if (id.name === "assert") {
            const startPos = this.lexer.absolutePosition;
            const args = this.functionCallArgumnets(fun);
            const endPos = this.lexer.absolutePosition;
            const text = this.lexer.text.slice(startPos - 1, endPos).trim().slice(1, -1);
            return new AssertCall(args, text, this.lexer.path, this.lexer.prev_line, this.lexer.prev_positionInLine);
        }
        return new FunctionCall(id.name, this.functionCallArgumnets(fun));
    }
    functionCallArgumnets(funDeclaration) {
        // LPARE EOLS? (expression EOLS?)? RPARE
        let IDs = funDeclaration.parameters;
        let args = [];
        this.eat(TokenType.LPARE);
        this.eatEOLS();
        if (IDs.length != 0) {
            if (IDs.length == 1) {
                const expressions = this.expression();
                if (expressions instanceof List && expressions.nodes.length != 1)
                    throw this.error(`${funDeclaration.name} expected ${1} argument, ${expressions.nodes.length} provided.`, 1);
                args.push(expressions);
            }
            else {
                const expressions = this.expression(IDs.length);
                if (!(expressions instanceof List))
                    throw this.error(`${funDeclaration.name} expected ${IDs.length} arguments, ${1} provided.`, 1);
                if (IDs.length !== expressions.nodes.length)
                    throw this.error(`${funDeclaration.name} expected ${IDs.length} arguments, ${expressions.nodes.length} provided.`, 1);
                args.push(...expressions.nodes);
            }
            this.eatEOLS();
        }
        this.eat(TokenType.RPARE);
        return new List(args);
    }
    functionDeclaration() {
        // FUNCTION ID functionParameters {block}
        const functionToken = this.token;
        this.eat(TokenType.FUNCTION);
        const name = this.token.value;
        // Checks
        {
            if (this.stack.hasInSameRecord(name))
                this.error(`Trying redefine declaration ${name} as function name.`, name.length);
            else
                this.stack.set(name, functionToken);
            if (langChars.isKeywordReserved(name))
                this.error(`Function name can't be '${name}', reserved keyword.`, name.length);
        }
        this.eat(TokenType.ID);
        const parameters = this.functionParameters();
        const funDeclaration = new NamedFunction(this.stack.isGlobalScope() ? this.stack.symbol : -1, name, parameters, null);
        this.stack.set(name, funDeclaration);
        this.stack.push();
        for (let parameter of parameters)
            this.stack.set(parameter.name, undefined);
        funDeclaration.block = this.block();
        const its = this.stack.getOwnIts();
        if (its.length)
            this.error(`Named functions can't use special 'it' variables ${its} only anonymous functions can.`);
        this.stack.pop();
        return funDeclaration;
    }
    functionParameters() {
        // LPARE EOLS? (ID (COMMA EOLS? ID)* EOLS?)? RPARE EOL?
        let parameters = [];
        let ids = new Set();
        this.eat(TokenType.LPARE);
        this.eatEOLS();
        if (this.is(TokenType.ID)) {
            const parameter = new ID(this.token);
            this.eat(TokenType.ID);
            parameters.push(parameter);
            ids.add(parameter.name);
            while (this.is(TokenType.COMMA)) {
                this.eat(TokenType.COMMA);
                this.eatEOLS();
                const parameter = new ID(this.token);
                if (ids.has(parameter.name))
                    this.error(`Duplicate identifier '${parameter.name}'.`, parameter.name.length);
                this.eat(TokenType.ID);
                parameters.push(parameter);
            }
            this.eatEOLS();
        }
        this.eat(TokenType.RPARE);
        this.eatEOL();
        return parameters;
    }
    return(nVals = undefined) {
        // RETURN ([EOL | RBRACKET | RPAREN] | SEMICOLON | [IF] conditionalExpression | expression)
        this.eat(TokenType.RETURN);
        if (this.is(TokenType.EOL, TokenType.RBRACKET, TokenType.RPARE))
            return new ReturnOp(new NoOp());
        if (this.is(TokenType.SEMICOLON)) {
            this.eat(this.token.type);
            return new ReturnOp(new NoOp());
        }
        if (this.is(TokenType.IF))
            return new ReturnOp(this.conditionalExpression(nVals));
        else
            return new ReturnOp(this.expression(nVals));
    }
    block(nVals = undefined) {
        //  [LBRACKET] blockFromTo [RBRACKET]
        return this.blockFromTo(TokenType.LBRACKET, TokenType.RBRACKET, nVals);
    }
    blockFromTo(start, end, nVals = undefined) {
        //  $start ( EOL+ | blockSubroutine )* $end
        this.eat(start);
        return this.blockTo(end);
    }
    blockTo(end, nVals = undefined) {
        //  ( EOL+ | blockSubroutine )* $end
        const nodes = [];
        while (!this.is(end)) {
            if (this.is(TokenType.EOL))
                this.eatEOLS();
            else
                nodes.push(this.blockSubroutine(nVals));
        }
        this.eat(end);
        if (nodes.length === 0)
            return new NoOp();
        if (nodes.length === 1)
            return nodes[0];
        else
            return new BlockOp(nodes);
    }
    // All the usual procedures you can do in a line of code
    blockSubroutine(nVals = undefined) {
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
            return this.functionDeclaration();
        if (this.is(TokenType.FOR))
            return this.forLoop();
        if (this.is(TokenType.RETURN))
            return this.return(nVals);
        if (this.is(TokenType.IF))
            return this.conditionalStatement(false, nVals);
        if (this.is(TokenType.IMPORT))
            return this.import();
        if (this.is(TokenType.ID)) {
            if (this.peekToken().is(TokenType.COMPOUND_DIV, TokenType.COMPOUND_MUL, TokenType.COMPOUND_PLUS, TokenType.COMPOUND_MINUS, TokenType.COMPOUND_POWER))
                return this.compoundOperation();
            if (this.isAssignation())
                return this.assignation();
        }
        return this.expression(nVals);
    }
    isAssignation() {
        // ID ((COMMA EOLS? ID) ASSIGN)?
        if (!this.is(TokenType.ID))
            return false;
        const initialState = this.getState();
        this.eat(TokenType.ID);
        while (this.is(TokenType.COMMA)) {
            this.eat(TokenType.COMMA);
            this.eatEOLS();
            if (!this.is(TokenType.ID))
                break;
            this.eat(TokenType.ID);
        }
        const is = this.is(TokenType.ASSIGN);
        this.restoreState(initialState);
        return is;
    }
    compoundOperation() {
        // ID (COMPOUND_DIV|COMPOUND_MUL|COMPOUND_PLUS|COMPOUND_MINUS) expression
        const id = new ID(this.token);
        if (langChars.isKeywordIt(id.name))
            this.error(`Not allowed to modify special variable ${id.name}, inmmutable variable.`, id.name.length);
        if (langChars.isKeywordReserved(id.name))
            this.error(`'${id.name}' is a reserved keyword, can't redefine.`, id.name.length);
        if (!this.stack.has(id.name))
            this.error(`'${id.name}' not declared.`, id.name.length);
        this.eat(TokenType.ID);
        if (this.is(TokenType.COMPOUND_DIV, TokenType.COMPOUND_MUL, TokenType.COMPOUND_PLUS, TokenType.COMPOUND_MINUS, TokenType.COMPOUND_POWER)) {
            const token = this.token;
            this.eat(this.token.type);
            return new CompoundOp(id, token, this.expression());
        }
        this.error(`Invalid token for compound operation ${this.token}.`, this.token.value.length);
    }
    _checkForIllegalReassignation(name) {
        if (langChars.isKeywordReserved(name))
            this.error(`'${name}' is a reserved keyword, can't redefine.`, name.length);
        if (this.stack.get(name) instanceof Stack)
            this.error(`Can't reassign variable '${name}', in use by module as alias.`, name.length);
    }
    assignation() {
        // ID (COMMA (ID))* ASSIGN EOL? expression+
        const ids = [];
        const idNode = new ID(this.token);
        this._checkForIllegalReassignation(idNode.name);
        this.eat(TokenType.ID);
        ids.push(idNode);
        this.stack.set(idNode.name, idNode);
        while (!this.is(TokenType.ASSIGN)) {
            this.eat(TokenType.COMMA);
            const idNode = new ID(this.token);
            this._checkForIllegalReassignation(idNode.name);
            this.eat(TokenType.ID);
            ids.push(idNode);
            this.stack.set(idNode.name, idNode);
        }
        this.eat(TokenType.ASSIGN);
        this.eatEOL();
        // This part might be a good reason to do an aditional parser
        // pass as you can't know statically the number of returns
        // directly (for now this is checked at runtime)
        const expression = this.expression();
        // CASE SIGNLE ASSIGMENT
        if (ids.length == 1) {
            this.stack.set(idNode.name, expression);
            return new AssignOp(ids[0], expression);
        }
        // CASE MULTIPLE ASSIGMENT
        if (expression instanceof List) {
            if (ids.length != expression.nodes.length)
                this.error(`Assignment needs ${ids.length} expressions, ${expression.nodes.length} given.`);
            for (let i = 0; i < ids.length; i++)
                this.stack.set(ids[i].name, expression.nodes[i]);
            return new MultipleAssignOp(ids, expression);
        }
        return new MultipleAssignOp(ids, expression);
    }
    forLoop() {
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
        const runBeforeSEMICOLON = (fn) => {
            const node = this.is(TokenType.SEMICOLON) ? new NoOp() : fn();
            this.eat(TokenType.SEMICOLON);
            return node;
        };
        this.eat(TokenType.FOR);
        const loopOp = this.optionalParenthesis((has) => {
            const setup = runBeforeSEMICOLON(() => this.assignation());
            const runCondition = runBeforeSEMICOLON(() => this.expression(1));
            const increment = has && this.is(TokenType.RPARE) ? new NoOp() : this.blockOrExpression();
            return new ForLoopOp(setup, runCondition, increment, null);
        });
        this.eatEOL();
        loopOp.body = this.is(TokenType.LBRACKET) ? this.block() : this.blockSubroutine();
        this.eatEOL();
        return loopOp;
    }
    program() {
        // ( EOL+ | blockSubroutine EOL )* EOF
        const nodes = [];
        while (!this.is(TokenType.EOF)) {
            if (this.is(TokenType.EOL))
                this.eatEOLS();
            else
                nodes.push(this.blockSubroutine());
        }
        this.eat(TokenType.EOF);
        if (nodes.length === 0)
            return new NoOp();
        if (nodes.length === 1)
            return nodes[0];
        else
            return new BlockOp(nodes);
    }
    parse() {
        return this.program();
    }
}
class Interpreter {
    constructor(filePath, symbol) {
        this.filePath = filePath;
        this.symbol = symbol;
        this.stack = new Stack(this.filePath, this.symbol);
    }
    error(msg) {
        throw new RuntimeError(extSystem.formatColor("Runtime error", extSystem.colors.red) + "\n" +
            extSystem.formatColor(" ■ ", extSystem.colors.red) + msg + "\n");
    }
    visit(node) {
        if (!(node instanceof ASTNode))
            return node;
        const call = this[`visit_${node.constructor.name}`];
        if (call != null)
            return call.bind(this)(node);
        this.error(`Trying to visit not defined node type: ${node.constructor.name} for node ${node}.`);
    }
    visit_ListEval(node) {
        // The list has already been evaluated
        return node;
    }
    visit_ReturnOp(node) {
        // Throw ReturnOp so it can signal the
        // capturing parent a return call has taken place
        // but keep vising the child node so a final
        // value can be found
        node.node = this.visit(node.node);
        throw node;
    }
    visit_Real(node) {
        return parseFloat(node.op.value);
    }
    visit_NamedImportLoadOp(node) {
        // Already loaded somewhere else, just need to get from importModules
        const moduleAlias = node.moduleAlias;
        const moduleSymbol = node.moduleSymbol;
        const moduleStack = Stack.importModules.get(moduleSymbol);
        this.stack.set(moduleAlias, moduleStack);
    }
    visit_NamedImportOp(node) {
        const ast = node.moduleAST;
        const moduleAlias = node.moduleAlias;
        const moduleSymbol = node.moduleSymbol;
        const interpreter = new Interpreter(node.filePath, moduleSymbol);
        interpreter.interpret(ast);
        const stack = interpreter.stack;
        Stack.importModules.set(moduleSymbol, interpreter.stack);
        this.stack.set(moduleAlias, stack);
    }
    visit_Bool(node) {
        return node.op.type == TokenType.TRUE ? true : false;
    }
    visit_NoOp(node) { }
    visit_UnaryOp(node) {
        const right = this.visit(node.right);
        if (right instanceof List) {
            if (node.op.type == TokenType.PLUS)
                return right;
            if (node.op.type == TokenType.MINUS)
                return new List(right.nodes.map(v => -v));
            if (node.op.type == TokenType.NOT)
                return new List(right.nodes.map(v => !v));
        }
        if (node.op.type == TokenType.PLUS)
            return right;
        if (node.op.type == TokenType.MINUS)
            return -right;
        if (node.op.type == TokenType.NOT)
            return !right;
        this.error(`Token type ${node.op.type} not implemented.`);
    }
    listIndex(object, index) {
        if (typeof index == "number") {
            if (index < 0)
                return object.nodes[object.nodes.length + index];
            return object.nodes[index];
        }
        if (index instanceof List) {
            const start = index.nodes[0];
            const end = index.nodes[1];
            if (end == -1)
                return new List(object.nodes.slice(start));
            if (end < -1)
                return new List(object.nodes.slice(start, end + 1));
            return new List(object.nodes.slice(start, end));
        }
        throw this.error(`Index value in not integer or range.`);
    }
    stringIndex(object, index) {
        if (typeof index == "number") {
            if (index < 0)
                return object[object.length + index];
            return object[index];
        }
        if (index instanceof List) {
            const start = index.nodes[0];
            const end = index.nodes[1];
            if (end == -1)
                return object.slice(start);
            if (end < -1)
                return object.slice(start, end + 1);
            return object.slice(start, end);
        }
        throw this.error(`Index value in not integer or range.`);
    }
    visit_AccessIndexOp(node) {
        const object = this.visit(node.list);
        if (object instanceof List)
            return this.listIndex(object, this.visit(node.index));
        if (typeof object == "string")
            return this.stringIndex(object, this.visit(node.index));
        throw this.error(`Object is not a list or string, can't acccess index.`);
    }
    // Defined only for the primitives types:
    // list, dict, string, number, bool, undefined and null
    visit_AccessMethodOp(node) {
        var _a;
        const object = this.visit(node.object);
        if (object instanceof List) {
            switch (node.key) {
                case "toString": return `${object}`;
                case "toDict": return new Dictionary(Object.fromEntries(object.nodes.map(v => v.nodes)));
                case "size": return object.nodes.length;
                case "sort": return new List([...object.nodes].sort((a, b) => (a instanceof List ? a.nodes[0] : a) -
                    (b instanceof List ? b.nodes[0] : b)));
                case "reverse": return new List([...object.nodes].reverse());
                case "flat": return new List(object.nodes.flatMap(v => v instanceof List ? v.nodes : v));
                case "is": return "list";
                case "first": return object.nodes[0];
                case "last": return object.nodes[object.nodes.length - 1];
                case "zip": {
                    if (object.nodes.length != 2)
                        throw this.error(`Zip call needs a list with two sub lists of the same size.`);
                    const first = object.nodes[0];
                    const second = object.nodes[1];
                    if (!(first instanceof List) || !(second instanceof List))
                        throw this.error(`Zip call needs a list with two sub lists of the same size.`);
                    if (first.nodes.length != second.nodes.length)
                        throw this.error(`Zip call needs a list with two sub lists of the same size.`);
                    return new List(first.nodes.map((v, i) => new List([v, second.nodes[i]])));
                }
                case "lastIndex": return object.nodes.length - 1;
                case "isEmpty": return object.nodes.length == 0;
                case "isNotEmpty": return object.nodes.length != 0;
                case "sum": return object.nodes.reduce((a, b) => a + b, 0);
                case "max": return Math.max.apply(null, object.nodes);
                case "min": return Math.min.apply(null, object.nodes);
                case "average": return object.nodes.length == 0 ? 0 : object.nodes.reduce((a, b) => a + b, 0) / object.nodes.length;
                case "randomItem": return object.nodes[Math.floor(Math.random() * (object.nodes.length))];
                case "dropFirst": return new List(object.nodes.slice(1));
                case "dropLast": return new List(object.nodes.slice(0, -1));
                case "hypot": return Math.sqrt(object.nodes.reduce((a, c) => a + c * c, 0));
                case "group": {
                    const dict = new Dictionary({});
                    for (let item of object.nodes) {
                        if (!(item instanceof List) || item.nodes.length != 2)
                            throw this.error(`Group operation. List items must be list of size 2: (key, value).`);
                        const key = item.nodes[0];
                        if (typeof key != "string" && typeof key != "number")
                            throw this.error(`Group operation. List item key must be 'string' or 'number'.`);
                        const value = item.nodes[1];
                        if (key in dict.nodes)
                            dict.nodes[key].nodes.push(value);
                        else
                            dict.nodes[key] = new List([value]);
                    }
                    return dict;
                }
            }
            this.error(`Method '${node.key}' not found in object type 'list'.`);
        }
        if (object instanceof Dictionary) {
            switch (node.key) {
                case "toString": return `${object}`;
                case "toList": return new List(Object.entries(object.nodes).map(v => new List(v)));
                case "size": return Object.keys(object.nodes).length;
                case "keys": return new List(Object.keys(object.nodes));
                case "values": return new List(Object.values(object.nodes));
                case "is": return "dict";
                case "isEmpty": return Object.keys(object.nodes).length == 0;
                case "isNotEmpty": return Object.keys(object.nodes).length != 0;
            }
            this.error(`Method '${node.key}' not found in object type 'dict' (dictionary).`);
        }
        if (typeof object == "string") {
            switch (node.key) {
                case "toString": return object.toString();
                case "toNumber": return parseFloat(object);
                case "toList": return new List(object.split(""));
                case "size": return object.length;
                case "toLowerCase": return object.toLowerCase();
                case "toUpperCase": return object.toUpperCase();
                case "trim": return object.trim();
                case "trimEnd": return object.trimEnd();
                case "trimStart": return object.trimStart();
                case "capitalize": return object.length == 0 ? "" : ((_a = object[0]) === null || _a === void 0 ? void 0 : _a.toUpperCase()) + object.slice(1);
                case "is": return "string";
                case "lines": return new List(object.split("\n"));
                case "isEmpty": return object === "";
                case "isNotEmpty": return object !== "";
                case "isBlank": return object.trim() === "";
                case "isNotBlank": return object.trim() !== "";
                case "first": return object[0];
                case "last": return object[object.length - 1];
                case "hashCode": {
                    // Java's equivalent
                    if (object.length == 0)
                        return 0;
                    let hash = 0;
                    for (let i = 0; i < object.length; ++i)
                        hash = (((hash << 5) - hash) + object.charCodeAt(i)) | 0; // int32
                    return hash;
                }
            }
            this.error(`Method '${node.key}' not found in object type 'string'.`);
        }
        if (typeof object == "number") {
            switch (node.key) {
                case "toString": return object.toString();
                case "toBool": return !object ? false : true;
                case "toList": return new List(new Array(Math.floor(Math.abs(object))).fill(0).map((_, i) => i));
                case "trunc": return Math.trunc(object);
                case "round": return Math.round(object);
                case "ceil": return Math.ceil(object);
                case "floor": return Math.floor(object);
                case "abs": return Math.abs(object);
                case "sign": return Math.sign(object);
                case "cos": return Math.cos(object);
                case "sin": return Math.sin(object);
                case "tan": return Math.tan(object);
                case "cosh": return Math.cosh(object);
                case "sinh": return Math.sinh(object);
                case "tanh": return Math.tanh(object);
                case "random": return Math.random() * object;
                case "randomInt": return Math.floor(Math.random() * (object + 1));
                case "randomSign": return (1 - Math.floor(Math.random() * 2) * 2) * object;
                case "sqrt": return Math.sqrt(object);
                case "negative": return -object;
                case "isNegative": return object < 0;
                case "isPositive": return object >= 0;
                case "isInfinite": return object === Infinity || object === -Infinity;
                case "isFinite": return object !== Infinity && object !== -Infinity && object !== NaN;
                case "isNaN": return object === NaN;
                case "is": return "number";
                case "rad": return object * Math.PI / 180;
                case "deg": return object * 180 / Math.PI;
                case "pi": return object * Math.PI;
            }
            this.error(`Method '${node.key}' not found in object type 'number': ${object}`);
        }
        if (typeof object == "boolean") {
            switch (node.key) {
                case "toString": return object ? "true" : "false";
                case "toNumber": return object ? 1 : 0;
                case "flip": return object ? false : true;
                case "is": return "bool";
            }
            this.error(`Method '${node.key}' not found in object type 'bool': ${object}`);
        }
        if (object === undefined) {
            switch (node.key) {
                case "toString": return "undefined";
                case "is": return "undefined";
            }
            this.error(`Method '${node.key}' not found in object type 'undefined': ${object}`);
        }
        if (object === null) {
            switch (node.key) {
                case "toString": return "null";
                case "is": return "null";
            }
            this.error(`Method '${node.key}' not found in object type 'null': ${object}`);
        }
        if (object instanceof Stack) {
            switch (node.key) {
                case "keys": return new List(object.getGlobalVariablesKeys());
                case "is": return "dict";
            }
            this.error(`Method '${node.key}' not found in object type 'null': ${object}`);
        }
        if (node.key == "toString")
            return `${object}`;
        this.error(`Can't call method '${node.key}'' from unknown object type '${typeof object}'`);
    }
    visit_AccessKeyOp(node) {
        const object = this.visit(node.object);
        if (object instanceof Dictionary)
            return object.nodes[node.key];
        if (object instanceof Stack) {
            if (!object.has(node.key))
                this.error(`Module ${node.object.name}: ${node.object.name}.${node.key} not found.`);
            return object.get(node.key);
        }
        throw this.error(`Object is not a dictionary o module, can't acccess by key.`);
    }
    visit_List(node) {
        return new ListEval(node.nodes.map(v => this.visit(v)));
    }
    visit_Dictionary(node) {
        for (let key in node.nodes) {
            if (node.nodes[key] instanceof AnonymousFunction)
                continue;
            node.nodes[key] = this.visit(node.nodes[key]);
        }
        return node;
    }
    visit_CallOp(node) {
        const fun = this.visit(node.func);
        if (fun instanceof BaseFunction) {
            const args = node.args instanceof List ? node.args : new List([node.args]);
            return this.executeFunction(fun, args);
        }
        throw this.error(`Object is not an function, can't execute a call. Got '${fun}' instead.`);
    }
    visit_PipeOp(pipe) {
        const itemExec = (fun, arg) => this.executeFunction(fun, arg instanceof List ? arg : new List([arg]));
        let left = this.visit(pipe.entry);
        for (let pipeNode of pipe.nodes) {
            const args = left instanceof List ? left : new List([left]);
            if (pipeNode.expression instanceof FunctionCall) {
                const fun = this.stack.get(pipeNode.expression.name);
                if (pipeNode.passType == TokenType.MAP)
                    left = new List(args.nodes.map(arg => itemExec(fun, arg)));
                else if (pipeNode.passType == TokenType.FILTER)
                    left = new List(args.nodes.filter(arg => itemExec(fun, arg)));
                else
                    left = this.executeFunction(fun, args);
            }
            else if (pipeNode.expression instanceof AnonymousFunction) {
                const fun = pipeNode.expression;
                if (pipeNode.passType == TokenType.MAP)
                    left = new List(args.nodes.map(arg => itemExec(fun, arg)));
                else if (pipeNode.passType == TokenType.FILTER)
                    left = new List(args.nodes.filter(arg => itemExec(fun, arg)));
                else
                    left = this.executeFunction(fun, args);
            }
            else {
                const fun = this.visit(pipeNode.expression);
                if (!(fun instanceof BaseFunction))
                    throw this.error(`Invalid pipe operator, not a function: '${fun}'.`);
                if (pipeNode.passType == TokenType.MAP)
                    left = new List(args.nodes.map(arg => itemExec(fun, arg)));
                else if (pipeNode.passType == TokenType.FILTER)
                    left = new List(args.nodes.filter(arg => itemExec(fun, arg)));
                else
                    left = this.executeFunction(fun, args);
            }
        }
        return left;
    }
    visit_VariableIncrementPrefixOp(node) {
        const currentVal = this.visit_ID(node.right);
        if (node.op.type == TokenType.PLUSPLUS) {
            const newVal = currentVal + 1;
            this.stack.set(node.right.name, newVal);
            return newVal;
        }
        if (node.op.type == TokenType.MINUSMINUS) {
            const newVal = currentVal - 1;
            this.stack.set(node.right.name, newVal);
            return newVal;
        }
        this.error(`Token type ${node.op.type} not implemented.`);
    }
    visit_VariableIncrementPostfixOp(node) {
        const currentVal = this.visit_ID(node.left);
        if (node.op.type == TokenType.PLUSPLUS) {
            this.stack.set(node.left.name, currentVal + 1);
            return currentVal;
        }
        if (node.op.type == TokenType.MINUSMINUS) {
            this.stack.set(node.left.name, currentVal - 1);
            return currentVal;
        }
        this.error(`Token type ${node.op.type} not implemented.`);
    }
    visit_AssignOp(node) {
        const right = this.visit(node.expression);
        this.stack.set(node.id.name, right);
        return;
    }
    visit_MultipleAssignOp(node) {
        const res = this.visit(node.expression);
        if (!(res instanceof List))
            throw this.error(`Assignment ${node.ids.map(v => v.name)} expected ${node.ids.length} values got one: '${res}'.`);
        if (node.ids.length > res.nodes.length)
            throw this.error(`Invalid assignment, expected ${node.ids.length} values for ${node.ids.map(v => v.name)} expression got ${res.nodes.length}: ${res.nodes}.`);
        for (let i = 0; i < node.ids.length; i++)
            this.stack.set(node.ids[i].name, res.nodes[i]);
    }
    visit_ComparisonOp(node) {
        let left = this.visit(node.nodes[0]);
        for (let i = 0; i < node.comparators.length; i++) {
            let right = this.visit(node.nodes[i + 1]);
            const type = node.comparators[i].type;
            switch (type) {
                case TokenType.EQUAL: {
                    if (left instanceof List && right instanceof List) {
                        if (left.nodes.length == right.nodes.length) {
                            if (left.nodes.every((v, i) => v == right.nodes[i]))
                                break;
                            else
                                return false;
                        }
                        else
                            return false;
                    }
                    else if (left == right)
                        break;
                    else
                        return false;
                }
                case TokenType.NOT_EQUAL: {
                    if (left instanceof List && right instanceof List) {
                        if (left.nodes.length == right.nodes.length) {
                            if (left.nodes.every((v, i) => v == right.nodes[i]))
                                return false;
                            else
                                break;
                        }
                        else
                            break;
                    }
                    else if (left != right)
                        break;
                    else
                        return false;
                }
                case TokenType.LESS: {
                    if (left < right)
                        break;
                    else
                        return false;
                }
                case TokenType.LESS_OR_EQUAL: {
                    if (left <= right)
                        break;
                    else
                        return false;
                }
                case TokenType.GREATER: {
                    if (left > right)
                        break;
                    else
                        return false;
                }
                case TokenType.GREATER_OR_EQUAL: {
                    if (left >= right)
                        break;
                    else
                        return false;
                }
                default: this.error(`Token type ${type} not implemented`);
            }
            left = right;
        }
        return true;
    }
    visit_BinaryOp(node) {
        const left = this.visit(node.left);
        const right = this.visit(node.right);
        const isListL = left instanceof List;
        const isListR = right instanceof List;
        if (isListL && !isListR) {
            if (node.op.type === TokenType.PLUS)
                return new List([...left.nodes, right]);
            if (node.op.type === TokenType.MUL)
                return new List(new Array(right).fill(0).flatMap(_ => left.nodes));
        }
        if (!isListL && isListR) {
            if (node.op.type === TokenType.PLUS)
                return new List([left, ...right.nodes]);
            if (node.op.type === TokenType.MUL)
                return new List(new Array(left).fill(0).flatMap(_ => right.nodes));
        }
        if (isListL && isListR) {
            if (node.op.type === TokenType.PLUS)
                return new List([...left.nodes, ...right.nodes]);
        }
        const isStringL = typeof left === "string";
        const isStringR = typeof right === "string";
        if (!isStringL && isStringR) {
            if (node.op.type === TokenType.MUL)
                return right.repeat(left);
        }
        if (isStringL && !isStringR) {
            if (node.op.type === TokenType.MUL)
                return left.repeat(right);
        }
        if (node.op.type === TokenType.PLUS)
            return left + right;
        if (node.op.type === TokenType.MINUS)
            return left - right;
        if (node.op.type === TokenType.MUL)
            return left * right;
        if (node.op.type === TokenType.DIV)
            return left / right;
        if (node.op.type === TokenType.MODULO)
            return left % right;
        if (node.op.type === TokenType.POWER)
            return Math.pow(left, right);
        this.error(`Token type ${node.op.type} not implemented.`);
    }
    visit_CompoundOp(node) {
        const id = node.left;
        const value = this.visit_ID(id);
        const compoundValue = this.visit(node.right);
        const isListL = value instanceof List;
        const isListR = compoundValue instanceof List;
        switch (node.op.type) {
            case TokenType.COMPOUND_PLUS: {
                if (isListL && !isListR)
                    return this.stack.setOnExisting(id.name, new List([...value.nodes, compoundValue]));
                if (isListL && isListR)
                    return this.stack.setOnExisting(id.name, new List([...value.nodes, ...compoundValue.nodes]));
                if (!isListL && !isListR)
                    return this.stack.setOnExisting(id.name, value + compoundValue);
                break;
            }
            case TokenType.COMPOUND_MINUS: {
                if (!isListL && !isListR)
                    return this.stack.setOnExisting(id.name, value - compoundValue);
                break;
            }
            case TokenType.COMPOUND_MUL: {
                if (isListL && !isListR)
                    return this.stack.setOnExisting(id.name, new List(new Array(compoundValue).fill(0).flatMap(_ => value.nodes)));
                if (!isListL && !isListR)
                    return this.stack.setOnExisting(id.name, value * compoundValue);
                break;
            }
            case TokenType.COMPOUND_DIV: {
                if (!isListL && !isListR)
                    return this.stack.setOnExisting(id.name, value / compoundValue);
                break;
            }
            case TokenType.COMPOUND_POWER: {
                if (!isListL && !isListR)
                    return this.stack.setOnExisting(id.name, Math.pow(value, compoundValue));
                break;
            }
        }
        this.error(`Invalid token ${node.op} operation for combination of ${isListL ? "list" : "scalar"} with ${isListR ? "list" : "scalar"}.`);
    }
    visit_ShortcircuitOp(node) {
        const left = this.visit(node.left);
        if (node.op.type == TokenType.AND) {
            if (!left)
                return false;
            const right = this.visit(node.right);
            return !!right;
        }
        if (node.op.type == TokenType.OR) {
            if (!!left)
                return true;
            const right = this.visit(node.right);
            return !!right;
        }
        this.error(`Token type ${node.op.type} not implemented.`);
    }
    visit_NamedFunction(node) {
        this.stack.set(node.name, node);
    }
    visit_FunctionCall(call) {
        const fun = this.stack.get(call.name);
        return this.executeFunction(fun, call.args);
    }
    executeFunction(fun, args) {
        const evaluatedArgs = this.visit_List(args);
        const currentScope = this.stack;
        // -1 means the base stack scope is the same
        // (happens when a function is not defined in global scope)
        if (fun.stackNamespace != -1)
            this.stack = Stack.importModules.get(fun.stackNamespace);
        this.stack.push();
        try {
            this.stack.set("arguments", evaluatedArgs);
            for (let i = 0; i < fun.parameters.length; i++)
                this.stack.set(fun.parameters[i].name, evaluatedArgs.nodes[i]);
            if (fun instanceof NamedFunction || fun instanceof AnonymousFunction)
                return this.visit(fun.block);
            else if (fun instanceof NativeFunction)
                return fun.fun.apply(fun.fun, evaluatedArgs.nodes);
            else
                this.error("FunctionCall, invalid instanceof type.");
        }
        catch (value) {
            if (value instanceof ReturnOp)
                return value.node;
            throw value;
        }
        finally {
            this.stack.pop();
            this.stack = currentScope;
        }
    }
    visit_AnonymousFunction(node) {
        return this.scopedReturn(() => { return this.visit(node.block); });
    }
    visit_SpreadListOp(node) {
        const res = this.visit(node.list);
        if (!(res instanceof List))
            throw this.error("Can't spread, object is not a list.");
        return res.nodes;
    }
    visit_PrintselfCall(call) {
        const evaluatedArgs = this.visit_List(call.args);
        const msg = evaluatedArgs.nodes[0];
        extSystem.print(`${call.text} ⟶  ${msg}`);
    }
    visit_AssertCall(call) {
        const res = this.visit_List(call.args);
        if (res.nodes[0] === true)
            return;
        const location = `(${call.file}:${call.line + 1}:${call.column})`;
        extSystem.print(extSystem.formatColor("FAILED ASSERT:", extSystem.colors.red) + "  " + call.text + "\n" +
            location);
    }
    visit_Conditional(node) {
        for (let i = 0; i < node.conditions.length; i++)
            if (this.visit(node.conditions[i]))
                return this.visit(node.bodies[i]);
        if (node.conditions.length != node.bodies.length)
            return this.visit(node.bodies[node.bodies.length - 1]);
    }
    visit_ForLoopOp(node) {
        this.visit(node.setup);
        while (this.visit(node.runCondition) == true) {
            this.visit(node.body);
            this.visit(node.increment);
        }
    }
    visit_RangeOp(node) {
        var _a;
        const list = [];
        const start = this.visit(node.start);
        const step = Math.abs((_a = this.visit(node.step)) !== null && _a !== void 0 ? _a : 1);
        if (step === 0)
            this.error("Range step can't be zero.");
        const end = this.visit(node.end);
        if (start < end) {
            const inv_step = 1.0 / step;
            const range = (end - start) * inv_step;
            list.length == range + 1;
            for (let i = 0; i <= range; i++)
                list.push(start + i * step);
            return new List(list);
        }
        else {
            const inv_step = 1.0 / step;
            const range = (start - end) * inv_step;
            list.length == range + 1;
            for (let i = 0; i <= range; i++)
                list.push(start - i * step);
            return new List(list);
        }
    }
    visit_BlockOp(blockNode) {
        const lastIndex = blockNode.list.length - 1;
        for (let i = 0; i <= lastIndex; i++) {
            const node = this.visit(blockNode.list[i]);
            if (i == lastIndex)
                return node;
        }
    }
    visit_StringOp(stringOp) {
        let output = stringOp.strings[0];
        for (let i = 0; i < stringOp.interpolations.length; i++) {
            const node = this.visit(stringOp.interpolations[i]);
            output += String(node);
            output += stringOp.strings[i + 1];
        }
        return output;
    }
    visit_Undefined(node) {
        return undefined;
    }
    visit_Null(node) {
        return null;
    }
    visit_ID(node) {
        if (this.stack.has(node.name))
            return this.stack.get(node.name);
        else
            this.error(`Variable '${node.name}' is not defined` + "\n" +
                `(${this.stack.pathFile})`);
    }
    returnCatcher(fn) {
        try {
            return fn();
        }
        catch (value) {
            if (value instanceof ReturnOp)
                return value.node;
            throw value;
        }
    }
    scopedReturn(fn) {
        this.stack.push();
        const res = this.returnCatcher(() => fn());
        this.stack.pop();
        return res;
    }
    interpret(tree, dumpAST = false, dumpFile = undefined) {
        if (dumpAST)
            extSystem.writeFile(extSystem.prettyPrintToString(tree), dumpFile, undefined);
        return this.returnCatcher(() => this.visit(tree));
    }
}
class JSEtimos {
    constructor(args) {
        this.args = args;
    }
    shellMode() {
        const interpreter = new Interpreter(extSystem.dirName(), 0);
        const globalStack = new Stack(extSystem.dirName(), 0);
        Stack.importModules.set(Stack.namespaceGlobalCount, interpreter.stack);
        extSystem.shell_onLineCallback = (line) => {
            try {
                const lexer = new Lexer(line, extSystem.dirName(), ":shell");
                const parser = new Parser(lexer, globalStack);
                const result = interpreter.interpret(parser.parse());
                if (result !== undefined)
                    extSystem.print(`${result}`);
            }
            catch (e) {
                if (e instanceof LexerError || e instanceof ParserError || e instanceof RuntimeError)
                    extSystem.print(e.message);
                else
                    extSystem.print("[[Unknown error, failed to execute]]\n\n" + e);
            }
            extSystem.shell_stdoutWrite(">>> ");
        };
        extSystem.print("JSEtimos v0.0.1 - nani [shell mode] Hello there :)");
        extSystem.shell_stdoutWrite(">>> ");
        extSystem.shell_loadInterface();
    }
    fileMode(fileName) {
        var _a;
        const filePath = extSystem.dirName() + "/" + fileName;
        const dumpAST = this.args.dumpAST;
        const dumpASTFile = (_a = this.args.dumpFile) !== null && _a !== void 0 ? _a : filePath + ".ast.yml";
        const dir = extSystem.dirName() + "/" + fileName.split("/").slice(0, -1).join("/");
        Stack.reset();
        const text = extSystem.readFile(filePath);
        const lexer = new Lexer(text, dir, fileName);
        const parser = new Parser(lexer, new Stack(filePath, 0));
        const interpreter = new Interpreter(filePath, 0);
        Stack.importModules.set(Stack.namespaceGlobalCount, interpreter.stack);
        const result = interpreter.interpret(parser.parse(), dumpAST, dumpASTFile);
        if (result !== undefined)
            extSystem.print(`${result}`);
    }
    inputMode(text) {
        var _a;
        const fileName = "input";
        const filePath = extSystem.dirName() + "/" + fileName;
        const dumpAST = this.args.dumpAST;
        const dumpASTFile = (_a = this.args.dumpFile) !== null && _a !== void 0 ? _a : filePath + ".ast.yml";
        const dir = extSystem.dirName() + "/" + fileName.split("/").slice(0, -1).join("/");
        Stack.reset();
        const lexer = new Lexer(text, dir, fileName);
        const parser = new Parser(lexer, new Stack(filePath, 0));
        const interpreter = new Interpreter(filePath, 0);
        Stack.importModules.set(Stack.namespaceGlobalCount, interpreter.stack);
        const result = interpreter.interpret(parser.parse(), dumpAST, dumpASTFile);
        if (result !== undefined)
            extSystem.print(`${result}`);
    }
    run() {
        try {
            if (this.args.fileName)
                this.fileMode(this.args.fileName);
            else if (this.args.inputMode)
                this.inputMode(this.args.inputText);
            else
                this.shellMode();
        }
        catch (e) {
            if (e instanceof LexerError || e instanceof ParserError || e instanceof RuntimeError)
                extSystem.print(e.stack);
            else
                throw e;
        }
    }
}
var extSystem = undefined;
var etimos = undefined;
// node
if ((typeof process !== "undefined") && (process.release.name === "node")) {
    extSystem = new ExtSystem_NodeJs();
    // extSystem.print("Loaded in node.js")
    etimos = new JSEtimos(extSystem.programArgs());
    etimos.run();
}
// Pyrogenesis
else if (typeof global !== "undefined" && global["Engine"] !== undefined && global["Engine"]["ProfileStart"] != undefined) {
    extSystem = new ExtSystem_Pyrogenesis0ad();
    // extSystem.print("Loaded in pyrogenesis engine")
    // Manual
}
// Browser
else {
    extSystem = new ExtSystem_Browser();
    // extSystem.print("Loaded in the browser")
    // Manual
}
//# sourceMappingURL=jsetimos.js.map