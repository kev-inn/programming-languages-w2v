import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import org.antlr.v4.runtime.tree.TerminalNode;

public class CppParseListener extends CPP14ParserBaseListener {
    public static String STRING_LITERAL_TOKEN = "STRING_LITERAL";
    public static String INT_LITERAL_TOKEN = "INT_LITERAL";
    public static String FLOAT_LITERAL_TOKEN = "FLOAT_LITERAL";
    public static String BOOL_LITERAL_TOKEN = "BOOL_LITERAL";
    public static String VARIABLE_TOKEN = "VARIABLE";

    public List<String> tokens = new ArrayList<>();
    List<Set<String>> replace_identifiers = new ArrayList<>();
    boolean is_variable_declaration = false;
    boolean skip_terminal = false;
    Map<Integer, String> literal_token_replacement;

    public CppParseListener() {
        literal_token_replacement = new HashMap<Integer, String>();
        literal_token_replacement.put(CPP14Lexer.StringLiteral, STRING_LITERAL_TOKEN);
        literal_token_replacement.put(CPP14Lexer.CharacterLiteral, STRING_LITERAL_TOKEN);
        literal_token_replacement.put(CPP14Lexer.IntegerLiteral, INT_LITERAL_TOKEN);
        literal_token_replacement.put(CPP14Lexer.FloatingLiteral, FLOAT_LITERAL_TOKEN);
        literal_token_replacement.put(CPP14Lexer.BooleanLiteral, BOOL_LITERAL_TOKEN);
    }

    @Override
    public void visitTerminal(TerminalNode node) {
        if (skip_terminal) {
            skip_terminal = false;
            return;
        }
        String replacement = literal_token_replacement.get(node.getSymbol().getType());
        if (replacement != null) {
            tokens.add(replacement);
            return;
        }

        tokens.add(node.getText());
    }

    @Override
    public void enterCompoundStatement(CPP14Parser.CompoundStatementContext ctx) {
        replace_identifiers.add(new HashSet<>());
    }

    @Override
    public void exitCompoundStatement(CPP14Parser.CompoundStatementContext ctx) {
        replace_identifiers.remove(replace_identifiers.size() - 1);
    }

    @Override
    public void enterSimpleDeclaration(CPP14Parser.SimpleDeclarationContext ctx) {
        if (ctx.declSpecifierSeq() != null)
            is_variable_declaration = true;
    }

    @Override
    public void exitSimpleDeclaration(CPP14Parser.SimpleDeclarationContext ctx) {
        is_variable_declaration = false;
    }

    @Override
    public void enterDeclaratorid(CPP14Parser.DeclaratoridContext ctx) {
        if (is_variable_declaration)
            replace_identifiers.get(replace_identifiers.size() - 1).add(ctx.toString());
    }

    @Override
    public void enterIdExpression(CPP14Parser.IdExpressionContext ctx) {
        for (Set<String> scope : replace_identifiers) {
            if (scope.contains(ctx.toString())) {
                skip_terminal = true;
                tokens.add(VARIABLE_TOKEN);
            }
        }
    }

}
