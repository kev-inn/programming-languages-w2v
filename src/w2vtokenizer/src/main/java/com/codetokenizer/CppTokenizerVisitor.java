package com.codetokenizer;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import org.antlr.v4.runtime.tree.TerminalNode;

public class CppTokenizerVisitor extends CPP14ParserBaseVisitor<List<String>> {
    List<Set<String>> replace_identifiers = new ArrayList<>();
    boolean is_variable_declaration = false;
    Map<Integer, String> literal_token_replacement;

    public CppTokenizerVisitor() {
        replace_identifiers.add(new HashSet<>());
        literal_token_replacement = new HashMap<>();
        literal_token_replacement.put(CPP14Lexer.StringLiteral, Tokenizer.STRING_LITERAL_TOKEN);
        literal_token_replacement.put(CPP14Lexer.CharacterLiteral, Tokenizer.STRING_LITERAL_TOKEN);
        literal_token_replacement.put(CPP14Lexer.IntegerLiteral, Tokenizer.INT_LITERAL_TOKEN);
        literal_token_replacement.put(CPP14Lexer.FloatingLiteral, Tokenizer.FLOAT_LITERAL_TOKEN);
        literal_token_replacement.put(CPP14Lexer.BooleanLiteral, Tokenizer.BOOL_LITERAL_TOKEN);
    }

    @Override
    public List<String> visitTerminal(TerminalNode node) {
        String replacement = literal_token_replacement.get(node.getSymbol().getType());
        if (replacement != null) {
            return List.of(replacement);
        }
        if (node.getSymbol().getType() == CPP14Lexer.Identifier) {
            for (Set<String> scope : replace_identifiers) {
                if (scope.contains(node.getText())) {
                    return List.of(Tokenizer.VARIABLE_TOKEN);
                }
            }
        }
        return List.of(node.getText());
    }

    @Override
    public List<String> visitCompoundStatement(CPP14Parser.CompoundStatementContext ctx) {
        replace_identifiers.add(new HashSet<>());
        var children = visitChildren(ctx);
        replace_identifiers.remove(replace_identifiers.size() - 1);
        return children;
    }

    @Override
    public List<String> visitSimpleDeclaration(CPP14Parser.SimpleDeclarationContext ctx) {
        if (ctx.declSpecifierSeq() != null) {
            is_variable_declaration = true;
        }
        var children = visitChildren(ctx);
        is_variable_declaration = false;
        return children;
    }

    @Override
    public List<String> visitDeclaratorid(CPP14Parser.DeclaratoridContext ctx) {
        if (is_variable_declaration)
            addVariableToScope(ctx.getText());
        return visitChildren(ctx);
    }

    @Override
    protected List<String> defaultResult() {
        return List.of();
    }

    @Override
    protected List<String> aggregateResult(List<String> aggregate, List<String> nextResult) {
        return Stream.concat(aggregate.stream(), nextResult.stream()).collect(Collectors.toList());
    }

    private void addVariableToScope(String variableName) {
        replace_identifiers.get(replace_identifiers.size() - 1).add(variableName);
    }
}
