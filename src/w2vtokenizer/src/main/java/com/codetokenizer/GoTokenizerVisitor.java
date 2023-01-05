package com.codetokenizer;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Stream;
import java.util.stream.Collectors;

import org.antlr.v4.runtime.tree.TerminalNode;

import com.codetokenizer.GoParser.ConstSpecContext;
import com.codetokenizer.GoParser.FieldDeclContext;
import com.codetokenizer.GoParser.IntegerContext;
import com.codetokenizer.GoParser.ParameterDeclContext;
import com.codetokenizer.GoParser.ShortVarDeclContext;
import com.codetokenizer.GoParser.String_Context;
import com.codetokenizer.GoParser.VarSpecContext;

public class GoTokenizerVisitor extends GoParserBaseVisitor<List<String>> {
    List<Set<String>> replace_identifiers = new ArrayList<>();
    Map<Integer, String> literal_token_replacement;

    public GoTokenizerVisitor() {
        replace_identifiers.add(new HashSet<>());
        literal_token_replacement = new HashMap<>();
        literal_token_replacement.put(GoLexer.FLOAT_LIT, Tokenizer.FLOAT_LITERAL_TOKEN);
    }

    private boolean isVariableToReplace(String identifier) {
        for (var scope : replace_identifiers) {
            if (scope.contains(identifier))
                return true;
        }
        return false;
    }

    @Override
    public List<String> visitTerminal(TerminalNode node) {
        String replacement = literal_token_replacement.get(node.getSymbol().getType());
        if (replacement != null) {
            return List.of(replacement);
        }
        if (node.getSymbol().getType() == GoLexer.IDENTIFIER && isVariableToReplace(node.getText())) {
            return List.of(Tokenizer.VARIABLE_TOKEN);
        }
        if (node.getSymbol().getType() == GoLexer.L_CURLY) {
            replace_identifiers.add(new HashSet<>());
        } else if (node.getSymbol().getType() == GoLexer.R_CURLY) {
            replace_identifiers.remove(replace_identifiers.size() - 1);
        }
        return List.of(node.getText());
    }

    @Override
    public List<String> visitConstSpec(ConstSpecContext ctx) {
        var names = visitIdentifierList(ctx.identifierList());
        replace_identifiers.get(replace_identifiers.size() - 1).addAll(names);
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitVarSpec(VarSpecContext ctx) {
        var names = visitIdentifierList(ctx.identifierList());
        replace_identifiers.get(replace_identifiers.size() - 1).addAll(names);
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitFieldDecl(FieldDeclContext ctx) {
        var names = visitIdentifierList(ctx.identifierList());
        replace_identifiers.get(replace_identifiers.size() - 1).addAll(names);
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitInteger(IntegerContext ctx) {
        return List.of(Tokenizer.INT_LITERAL_TOKEN);
    }

    @Override
    public List<String> visitParameterDecl(ParameterDeclContext ctx) {
        if (ctx.identifierList() != null) {
            var names = visitIdentifierList(ctx.identifierList());
            replace_identifiers.get(replace_identifiers.size() - 1).addAll(names);
        }
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitShortVarDecl(ShortVarDeclContext ctx) {
        var names = visitIdentifierList(ctx.identifierList());
        replace_identifiers.get(replace_identifiers.size() - 1).addAll(names);
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitString_(String_Context ctx) {
        return List.of(Tokenizer.STRING_LITERAL_TOKEN);
    }

    @Override
    protected List<String> aggregateResult(List<String> aggregate, List<String> nextResult) {
        return Stream.concat(aggregate.stream(), nextResult.stream()).collect(Collectors.toList());
    }

    @Override
    protected List<String> defaultResult() {
        return List.of();
    }
}
