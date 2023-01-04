import java.util.ArrayList;
import java.util.List;

import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.atn.PredictionMode;
import org.antlr.v4.runtime.tree.ParseTree;
import org.antlr.v4.runtime.tree.ParseTreeWalker;
import py4j.GatewayServer;

public class CppTokenizer {
    public static void main(String[] args) {
        CppTokenizer app = new CppTokenizer();
        GatewayServer server = new GatewayServer(app);
        server.start();
    }

    public static List<String> tokenize(String code) {
        CPP14Lexer lexer = new CPP14Lexer(CharStreams.fromString(code));
        lexer.removeErrorListeners();
        CPP14Parser parser = new CPP14Parser(new CommonTokenStream(lexer));
        parser.removeErrorListeners();

        ParseTreeWalker walker = new ParseTreeWalker();
        CppParseListener listener = new CppParseListener();

        parser.getInterpreter().setPredictionMode(PredictionMode.SLL);
        parser.setErrorHandler(new BailErrorStrategy());
        ParseTree tree;

        try {
            tree = parser.translationUnit();

            walker.walk(listener, tree);
            return listener.tokens;
        } catch (Exception e) {
            try {
                lexer.reset();
                parser.reset();
                parser.setErrorHandler(new DefaultErrorStrategy());
                parser.getInterpreter().setPredictionMode(PredictionMode.LL);

                tree = parser.translationUnit();

                walker.walk(listener, tree);
                return listener.tokens;
            } catch (Exception e_) {

            }
        }
        return List.of();
    }
}