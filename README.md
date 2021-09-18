

### **[Click here to try it !](https://nanihadesuka.github.io/JsEtimos/)**

### About

JsEtimos is an interpreted language based on a mix of python, javascript and kotlin.

Python for the scope rules, minimum boilerplate and the import system.  
Javascript for the list and array objects and dynamism.  
Kotlin for the lambda and conditional expressions and useful built-in methods.  
Also has some extra syntax rules and code flow as pipes and special "?" methods.

The language is executed using a custom lexer, [recursive descent parser](https://en.wikipedia.org/wiki/Recursive_descent_parser), [AST](https://en.wikipedia.org/wiki/Abstract_syntax_tree) tree and finally an interepreter.

The parser is of the type [PEG](https://en.wikipedia.org/wiki/Parsing_expression_grammar) with quasi context-free syntax.

The lexer and parser can analize statically the validity of the syntax (and some semantics) and give contextual errors with error line and cause.

### Running

The interpreter validates the code and data flow at runtime given the possible dynamism of the language. The runtime will give informative errors but won't show you any contextual information or where (position in the program) the error has occured.

The interpreter can be run in three modes:  
- from a file,
- shell mode (REPL)
- from a string given in the command line.  

For more info for how to run it read the INSTRUCTIONS file.

As the implementation has been done in Typescript it can run on the browser, node.js and even on the game [0 A.D](https://play0ad.com/) which uses the spidermonkey js engine (nice game btw, check it out :-> )

### Notice
This is a toy language, the implementation is still missing many things, has some bugs and inconsistencies and is not optimized so be gente with it :)

