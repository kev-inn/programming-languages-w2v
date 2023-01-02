from antlr4 import *
from .parsers.CPP14Lexer import CPP14Lexer
from .parsers.CPP14ParserListener import CPP14ParserListener
from .parsers.CPP14Parser import CPP14Parser

# some common tokens all languages should have in commin
STRING_LITERAL_TOKEN = "STRING_LITERAL"
INT_LITERAL_TOKEN = "INT_LITERAL"
FLOAT_LITERAL_TOKEN = "FLOAT_LITERAL"
BOOL_LITERAL_TOKEN = "BOOL_LITERAL"
VARIABLE_TOKEN = "VARIABLE"


class CppTokenizer(CPP14ParserListener):
    def __init__(self):
        self.tokens = []
        self.replace_identifiers = [[]]
        self.is_variable_declaration = False
        self.skip_terminal = False

        self.literal_token_replacement = {
            CPP14Lexer.StringLiteral: STRING_LITERAL_TOKEN,
            CPP14Lexer.CharacterLiteral: STRING_LITERAL_TOKEN,
            CPP14Lexer.IntegerLiteral: INT_LITERAL_TOKEN,
            CPP14Lexer.FloatingLiteral: FLOAT_LITERAL_TOKEN,
            CPP14Lexer.BooleanLiteral: BOOL_LITERAL_TOKEN,
        }

    def visitTerminal(self, node: TerminalNode):
        if self.skip_terminal:
            self.skip_terminal = False
            return
        if node.symbol.type in self.literal_token_replacement:
            replacement = self.literal_token_replacement[node.symbol.type]
            self.tokens.append(replacement)
            return

        # generic
        self.tokens.append(node.getText())

    def enterCompoundStatement(self, ctx: CPP14Parser.CompoundStatementContext):
        self.replace_identifiers.append([])

    def exitCompoundStatement(self, ctx: CPP14Parser.CompoundStatementContext):
        self.replace_identifiers.pop()

    def enterSimpleDeclaration(self, ctx: CPP14Parser.SimpleDeclarationContext):
        if ctx.declSpecifierSeq():
            self.is_variable_declaration = True

    def exitSimpleDeclaration(self, ctx: CPP14Parser.SimpleDeclarationContext):
        self.is_variable_declaration = False

    def enterDeclaratorid(self, ctx: CPP14Parser.DeclaratoridContext):
        if self.is_variable_declaration:
            self.replace_identifiers[-1].append(ctx.getText())

    def enterIdExpression(self, ctx: CPP14Parser.IdExpressionContext):
        if any([ctx.getText() in scope for scope in self.replace_identifiers]):
            self.skip_terminal = True
            self.tokens.append(VARIABLE_TOKEN)
