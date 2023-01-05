package com.codetokenizer;

import java.util.List;
import java.util.function.Supplier;

import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.atn.PredictionMode;
import org.antlr.v4.runtime.tree.AbstractParseTreeVisitor;
import org.antlr.v4.runtime.tree.ParseTree;
import py4j.GatewayServer;

public class Tokenizer {
    public static String STRING_LITERAL_TOKEN = "STRING_LITERAL";
    public static String INT_LITERAL_TOKEN = "INT_LITERAL";
    public static String FLOAT_LITERAL_TOKEN = "FLOAT_LITERAL";
    public static String BOOL_LITERAL_TOKEN = "BOOL_LITERAL";
    public static String VARIABLE_TOKEN = "VARIABLE";

    public static void main(String[] args) {
        Tokenizer app = new Tokenizer();
        GatewayServer server = new GatewayServer(app);
        server.start();
    }

    public static List<String> tokenizeCpp(String code) {
        var lexer = new CPP14Lexer(CharStreams.fromString(code));
        var parser = new CPP14Parser(new CommonTokenStream(lexer));
        var visitor = new CppTokenizerVisitor();

        return tokenize(lexer, parser, visitor, parser::translationUnit);
    }

    public static List<String> tokenizeCsharp(String code) {
        var lexer = new CSharpLexer(CharStreams.fromString(code));
        var parser = new CSharpParser(new CommonTokenStream(lexer));
        var visitor = new CSharpTokenizerVisitor();

        return tokenize(lexer, parser, visitor, parser::compilation_unit);
    }

    public static List<String> tokenizeGo(String code) {
        var lexer = new GoLexer(CharStreams.fromString(code));
        var parser = new GoParser(new CommonTokenStream(lexer));
        var visitor = new GoTokenizerVisitor();

        return tokenize(lexer, parser, visitor, parser::sourceFile);
    }

    private static List<String> tokenize(Lexer lexer, Parser parser,
            AbstractParseTreeVisitor<List<String>> visitor, Supplier<ParseTree> startRule) {
        lexer.removeErrorListeners();
        parser.removeErrorListeners();

        parser.getInterpreter().setPredictionMode(PredictionMode.SLL);
        parser.setErrorHandler(new BailErrorStrategy());
        ParseTree tree;

        try {
            tree = startRule.get();
            return visitor.visit(tree);
        } catch (Exception e) {
        }
        try {
            lexer.reset();
            parser.reset();
            parser.setErrorHandler(new DefaultErrorStrategy());
            parser.getInterpreter().setPredictionMode(PredictionMode.LL);

            tree = startRule.get();
            return visitor.visit(tree);
        } catch (Exception e) {
        }
        return List.of();
    }
}