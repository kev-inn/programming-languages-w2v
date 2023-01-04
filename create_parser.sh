#!/usr/bin/env bash

antlr4 -Dlanguage=Python3 -o src/parsers grammars/C.g4
antlr4 -Dlanguage=Python3 -o src/parsers grammars/CPP14Parser.g4 grammars/CPP14Lexer.g4
antlr4 -Dlanguage=Java -o src/java_parser grammars/CPP14Parser.g4 grammars/CPP14Lexer.g4

# todo: adjust this for linux environment
javac -cp ./src/java_parser/antlr-runtime-4.11.1.jar;env/share/py4j/py4j0.10.9.7.jar -d ./src/java_parser/build ./src/java_parser/*.java
