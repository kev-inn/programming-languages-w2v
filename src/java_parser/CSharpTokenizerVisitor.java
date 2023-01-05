import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import org.antlr.v4.runtime.tree.TerminalNode;

public class CSharpTokenizerVisitor extends CSharpParserBaseVisitor<List<String>> {
    List<Set<String>> replace_identifiers = new ArrayList<>();
    Map<Integer, String> literal_token_replacement;

    public CSharpTokenizerVisitor() {
        replace_identifiers.add(new HashSet<>());
        literal_token_replacement = new HashMap<>();
        literal_token_replacement.put(CSharpLexer.CHARACTER_LITERAL, Tokenizer.STRING_LITERAL_TOKEN);
        literal_token_replacement.put(CSharpLexer.INTEGER_LITERAL, Tokenizer.INT_LITERAL_TOKEN);
        literal_token_replacement.put(CSharpLexer.HEX_INTEGER_LITERAL, Tokenizer.INT_LITERAL_TOKEN);
        literal_token_replacement.put(CSharpLexer.BIN_INTEGER_LITERAL, Tokenizer.INT_LITERAL_TOKEN);
        literal_token_replacement.put(CSharpLexer.REAL_LITERAL, Tokenizer.FLOAT_LITERAL_TOKEN);
    }

    @Override
    public List<String> visitTerminal(TerminalNode node) {
        String replacement = literal_token_replacement.get(node.getSymbol().getType());
        if (replacement != null) {
            return List.of(replacement);
        }
        if (node.getSymbol().getType() == CSharpLexer.OPEN_BRACE) {
            replace_identifiers.add(new HashSet<>());
        } else if (node.getSymbol().getType() == CSharpLexer.CLOSE_BRACE) {
            replace_identifiers.remove(replace_identifiers.size() - 1);
        }
        return List.of(node.getText());
    }

    @Override
    public List<String> visitBoolean_literal(CSharpParser.Boolean_literalContext ctx) {
        return List.of(Tokenizer.BOOL_LITERAL_TOKEN);
    }

    @Override
    public List<String> visitString_literal(CSharpParser.String_literalContext ctx) {
        return List.of(Tokenizer.STRING_LITERAL_TOKEN);
    }

    @Override
    protected List<String> defaultResult() {
        return List.of();
    }

    private boolean isVariableToReplace(String identifier) {
        for (var scope : replace_identifiers) {
            if (scope.contains(identifier))
                return true;
        }
        return false;
    }

    @Override
    public List<String> visitIdentifier(CSharpParser.IdentifierContext ctx) {
        if (isVariableToReplace(ctx.getText()))
            return List.of(Tokenizer.VARIABLE_TOKEN);
        return List.of(ctx.getText());
    }

    @Override
    protected List<String> aggregateResult(List<String> aggregate, List<String> nextResult) {
        return Stream.concat(aggregate.stream(), nextResult.stream()).collect(Collectors.toList());
    }

    private void addVariableToScope(String variableName) {
        replace_identifiers.get(replace_identifiers.size() - 1).add(variableName);
    }

    @Override
    public List<String> visitConstant_declarator(CSharpParser.Constant_declaratorContext ctx) {
        addVariableToScope(ctx.identifier().getText());
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitLocal_variable_declarator(CSharpParser.Local_variable_declaratorContext ctx) {
        addVariableToScope(ctx.identifier().getText());
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitMember_declarator(CSharpParser.Member_declaratorContext ctx) {
        addVariableToScope(ctx.identifier().getText());
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitVariable_declarator(CSharpParser.Variable_declaratorContext ctx) {
        addVariableToScope(ctx.identifier().getText());
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitFixed_pointer_declarator(CSharpParser.Fixed_pointer_declaratorContext ctx) {
        addVariableToScope(ctx.identifier().getText());
        return visitChildren(ctx);
    }

    @Override
    public List<String> visitArg_declaration(CSharpParser.Arg_declarationContext ctx) {
        addVariableToScope(ctx.identifier().getText());
        return visitChildren(ctx);
    }

}
