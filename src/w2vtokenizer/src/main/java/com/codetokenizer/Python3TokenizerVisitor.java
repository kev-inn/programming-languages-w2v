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

import com.codetokenizer.Python3Parser.Atom_exprContext;
import com.codetokenizer.Python3Parser.Capture_patternContext;
import com.codetokenizer.Python3Parser.ClassdefContext;
import com.codetokenizer.Python3Parser.Expr_stmtContext;
import com.codetokenizer.Python3Parser.FuncdefContext;
import com.codetokenizer.Python3Parser.LambdefContext;
import com.codetokenizer.Python3Parser.Lambdef_nocondContext;
import com.codetokenizer.Python3Parser.Match_stmtContext;
import com.codetokenizer.Python3Parser.StringsContext;
import com.codetokenizer.Python3Parser.TfpdefContext;
import com.codetokenizer.Python3Parser.VfpdefContext;
import com.codetokenizer.Python3Parser.With_itemContext;

public class Python3TokenizerVisitor extends Python3ParserBaseVisitor<List<String>> {

    List<Set<String>> replace_identifiers = new ArrayList<>();
    Map<Integer, String> literal_token_replacement;
    boolean is_variable_declaration = false;

    public Python3TokenizerVisitor() {
        replace_identifiers.add(new HashSet<>());
        literal_token_replacement = new HashMap<>();
        literal_token_replacement.put(Python3Lexer.STRING, Tokenizer.STRING_LITERAL_TOKEN);
        // TODO: maybe we can somehow differentiate between float and int
        literal_token_replacement.put(Python3Lexer.NUMBER, Tokenizer.FLOAT_LITERAL_TOKEN);
        literal_token_replacement.put(Python3Lexer.TRUE, Tokenizer.BOOL_LITERAL_TOKEN);
        literal_token_replacement.put(Python3Lexer.FALSE, Tokenizer.BOOL_LITERAL_TOKEN);
    }

    @Override
    public List<String> visitVfpdef(VfpdefContext ctx) {
        addVariableToScope(ctx.getText());
        return super.visitVfpdef(ctx);
    }

    private boolean isVariableToReplace(String identifier) {
        for (var scope : replace_identifiers) {
            if (scope.contains(identifier))
                return true;
        }
        return false;
    }

    @Override
    protected List<String> defaultResult() {
        return List.of();
    }

    @Override
    public List<String> visitClassdef(ClassdefContext ctx) {
        replace_identifiers.add(new HashSet<>());
        var result = visitChildren(ctx);
        replace_identifiers.remove(replace_identifiers.size() - 1);
        return result;
    }

    @Override
    public List<String> visitFuncdef(FuncdefContext ctx) {
        replace_identifiers.add(new HashSet<>());
        var result = visitChildren(ctx);
        replace_identifiers.remove(replace_identifiers.size() - 1);
        return result;
    }

    @Override
    public List<String> visitMatch_stmt(Match_stmtContext ctx) {
        replace_identifiers.add(new HashSet<>());
        var result = visitChildren(ctx);
        replace_identifiers.remove(replace_identifiers.size() - 1);
        return result;
    }

    @Override
    public List<String> visitCapture_pattern(Capture_patternContext ctx) {
        addVariableToScope(ctx.getText());
        return super.visitCapture_pattern(ctx);
    }

    @Override
    public List<String> visitStrings(StringsContext ctx) {
        return List.of(Tokenizer.STRING_LITERAL_TOKEN);
    }

    @Override
    public List<String> visitTfpdef(TfpdefContext ctx) {
        addVariableToScope(ctx.name().getText());
        return super.visitTfpdef(ctx);
    }

    @Override
    public List<String> visitWith_item(With_itemContext ctx) {
        if (ctx.expr() != null) {
            addVariableToScope(ctx.expr().getText());
        }
        return super.visitWith_item(ctx);
    }

    @Override
    protected List<String> aggregateResult(List<String> aggregate, List<String> nextResult) {
        return Stream.concat(aggregate.stream(), nextResult.stream()).collect(Collectors.toList());
    }

    private void addVariableToScope(String variableName) {
        replace_identifiers.get(replace_identifiers.size() - 1).add(variableName);
    }

    @Override
    public List<String> visitExpr_stmt(Expr_stmtContext ctx) {
        is_variable_declaration = true;
        visitTestlist_star_expr(ctx.testlist_star_expr(0));
        is_variable_declaration = false;
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitAtom_expr(Atom_exprContext ctx) {
        if (!is_variable_declaration) {
            return visitChildren(ctx);
        }
        if (ctx.trailer() != null && !ctx.trailer().isEmpty()) {
            var name = ctx.trailer(ctx.trailer().size() - 1).name();
            if (name != null) {
                addVariableToScope(name.getText());
            }
        } else {
            addVariableToScope(ctx.atom().getText());
        }
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitLambdef(LambdefContext ctx) {
        replace_identifiers.add(new HashSet<>());
        var result = visitChildren(ctx);
        replace_identifiers.remove(replace_identifiers.size() - 1);
        return result;
    }

    @Override
    public List<String> visitLambdef_nocond(Lambdef_nocondContext ctx) {
        replace_identifiers.add(new HashSet<>());
        var result = visitChildren(ctx);
        replace_identifiers.remove(replace_identifiers.size() - 1);
        return result;
    }

    @Override
    public List<String> visitTerminal(TerminalNode node) {
        String replacement = literal_token_replacement.get(node.getSymbol().getType());
        if (replacement != null) {
            return List.of(replacement);
        }
        if (node.getSymbol().getType() == Python3Lexer.NAME && isVariableToReplace(node.getText())) {
            return List.of(Tokenizer.VARIABLE_TOKEN);
        }
        return List.of(node.getText());
    }

}
