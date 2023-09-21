// YourLanguageLexer.l

%package rinhaLexer

// Tokens
%{
package rinhaLexer
%}

// Lexer rules
WS          [ \t\r\n]+
COMMENT     \/\/[^\n]*\n?
ML_COMMENT  \/\*[^*]*\*+([^\/][^*]*\*+)*\/

// Keywords
EXTERNAL    "external"
SEMI        ";"
TRUE        "true"
FALSE       "false"
LET         "let"
IF          "if"
ELSE        "else"
FN          "fn"
PRINT       "print"
FIRST       "first"
SECOND      "second"

// Operators
ADD         "+"
SUB         "-"
MUL         "*"
DIV         "/"
MOD         "%"
OR          "||"
AND         "&&"
EQ          "=="
NEQ         "!="
LTE         "<="
GTE         ">="
LT          "<"
GT          ">"

// Identifiers and References
IDENTIFIER  [a-zA-Z][a-zA-Z0-9/_]*
REFERENCE   {IDENTIFIER}

// Literals
INTEGER     [0-9]+
STRING_LITERAL  "\""([^"\\]|\\.)*\""

// Escape sequence
ESC         \\[\\"]

%%
