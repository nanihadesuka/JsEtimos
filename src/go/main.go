// JSEtimos - copyright nani 2021
//
// Go port of jsetimos.ts: lexer (lexer.go), parser (parser.go) and
// interpreter (interpreter.go).
//
// The TypeScript implementation leans on the JavaScript runtime for most of the
// value semantics (dynamic typing, `+` coercion, loose `==`, number printing,
// UTF-16 strings, object key order...). values.go reproduces those semantics
// explicitly so programs behave the same, including the known quirks.
//
// This file: command line entry point and host system access (node.js equivalent).
//
// Build (from the repository root): scripts/build_go.sh, or go build -C src/go -o ../../build/jsetimos.exe .
// Run:   build/jsetimos fileName | build/jsetimos (shell mode) | build/jsetimos --inputMode "print(1)"
// Paths are resolved from the working directory, so run it from the repository root.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
			// --flag or --name=value
			if name, value, found := strings.Cut(item[2:], "="); found {
				attributes[name] = value
			} else {
				attributes[item[2:]] = item[2:]
			}
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
		// Relative to the base directory, like any other written file
		dumpASTFile = fileName + ".ast.yml"
	}

	// The lexer takes the directory and the bare file name separately,
	// otherwise the relative directory ends up twice in the file path
	parts := strings.Split(fileName, "/")
	dir := extSystem.dirName()
	if relativeDir := strings.Join(parts[:len(parts)-1], "/"); relativeDir != "" {
		dir += "/" + relativeDir
	}

	lexer := newLexer(text, dir, parts[len(parts)-1])
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
