package lexer

import (
	"fmt"

	"github.com/davyxu/golexer"
	"github.com/omurilo/rinha-compiler/ast"
	// "github.com/omurilo/rinha-compiler/runtime"
)

var KEYWORDS = map[string]string{
	"print":  ":PRINT",
	"true":   ":TRUE",
	"false":  ":FALSE",
	"first":  ":FIRST",
	"second": ":SECOND",
	"if":     ":IF",
	"else":   ":ELSE",
	"fn":     ":FUNCTION",
	"let":    ":LET",
}

var SYMBOLS = map[string]string{
	"(":  ":LPAREN",
	")":  ":RPAREN",
	"+":  ":BINARY_OP",
	"-":  ":BINARY_OP",
	"*":  ":BINARY_OP",
	"/":  ":BINARY_OP",
	"%":  ":BINARY_OP",
	"==": ":BINARY_OP",
	"<":  ":BINARY_OP",
	">":  ":BINARY_OP",
	"=":  ":ASSIGNMENT",
	";":  ":SEMICOLON",
	"{":  ":LBRACE",
	"}":  ":RBRACE",
	",":  ":COMMA",
	"=>": ":ARROW",
	">=": ":BINARY_OP",
	"<=": ":BINARY_OP",
	"!=": ":BINARY_OP",
	"&&": ":BINARY_OP",
	"||": ":BINARY_OP",
}

type Token struct {
	Type     string
	Value    string
	Location ast.Location
}

const (
	Token_EOF = iota
	Token_Unknown
	Token_Numeral
	Token_String
	Token_WhiteSpace
	Token_LineEnd
	Token_UnixStyleComment
	Token_Identifier
	Token_Semicolon
	Token_If
	Token_Else
	Token_True
	Token_False
	Token_First
	Token_Second
	Token_Print
	Token_Add
	Token_Sub
	Token_Mul
	Token_Div
	Token_Rem
	Token_Eq
	Token_Let
	Token_Fn
	Token_Comma
	Token_Lte
	Token_Gte
	Token_Lt
	Token_Gt
	Token_Neq
	Token_Or
	Token_And
	Token_Assignment
	Token_Arrow
	Token_LParen
	Token_RParen
	Token_LBrace
	Token_RBrace
)

type CustomParser struct {
	*golexer.Parser
}

func NewCustomParser(filename string) *CustomParser {
	l := golexer.NewLexer()

	l.AddMatcher(golexer.NewNumeralMatcher(Token_Numeral))
	l.AddMatcher(golexer.NewStringMatcher(Token_String))

	l.AddIgnoreMatcher(golexer.NewWhiteSpaceMatcher(Token_WhiteSpace))
	l.AddIgnoreMatcher(golexer.NewLineEndMatcher(Token_LineEnd))
	l.AddIgnoreMatcher(golexer.NewUnixStyleCommentMatcher(Token_UnixStyleComment))

	l.AddMatcher(golexer.NewSignMatcher(Token_Semicolon, ";"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Comma, ","))

	l.AddMatcher(golexer.NewKeywordMatcher(Token_If, "if"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_Else, "else"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_Print, "print"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_First, "first"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_Second, "second"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_Let, "let"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_Fn, "fn"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_True, "true"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_False, "false"))

	l.AddMatcher(golexer.NewSignMatcher(Token_Add, "+"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Sub, "-"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Mul, "*"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Div, "/"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Rem, "%"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Arrow, "=>"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Lte, "<="))
	l.AddMatcher(golexer.NewSignMatcher(Token_Gte, ">="))
	l.AddMatcher(golexer.NewSignMatcher(Token_Eq, "=="))
	l.AddMatcher(golexer.NewSignMatcher(Token_Neq, "!="))
	l.AddMatcher(golexer.NewSignMatcher(Token_Lt, "<"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Gt, ">"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Assignment, "="))
	l.AddMatcher(golexer.NewSignMatcher(Token_Or, "||"))
	l.AddMatcher(golexer.NewSignMatcher(Token_And, "&&"))

	l.AddMatcher(golexer.NewSignMatcher(Token_LParen, "("))
	l.AddMatcher(golexer.NewSignMatcher(Token_RParen, ")"))
	l.AddMatcher(golexer.NewSignMatcher(Token_LBrace, "{"))
	l.AddMatcher(golexer.NewSignMatcher(Token_RBrace, "}"))

	l.AddMatcher(golexer.NewIdentifierMatcher(Token_Identifier))

	l.AddMatcher(golexer.NewUnknownMatcher(Token_Unknown))

	return &CustomParser{
		Parser: golexer.NewParser(l, filename),
	}
}

func Initialize(program string, filename string) *CustomParser {
	p := NewCustomParser(filename)
	p.Lexer().Start(program)

	return p
}

func (p *CustomParser) Next() Token {
	var token Token
	p.NextToken()

	if p.TokenID() != 0 {
		switch p.MatcherName() {
		case "NumeralMatcher":
			token = Token{Type: ":NUMBER", Value: p.TokenValue(), Location: parse_location(p.TokenPos())}
		case "StringMatcher":
			token = Token{Type: ":STRING", Value: p.TokenValue(), Location: parse_location(p.TokenPos())}
		case "SignMatcher":
			token = Token{Type: SYMBOLS[p.TokenValue()], Value: p.TokenValue(), Location: parse_location(p.TokenPos())}
		case "KeywordMatcher":
			token = Token{Type: KEYWORDS[p.TokenValue()], Value: p.TokenValue(), Location: parse_location(p.TokenPos())}
		case "IdentifierMatcher":
			token = Token{Type: ":IDENTIFIER", Value: p.TokenValue(), Location: parse_location(p.TokenPos())}
		}
	}

	return token
}

func (p *CustomParser) Tokenize() {
	token := p.Next()

	for p.TokenID() != 0 {
		fmt.Println(token)
		token = p.Next()
	}
}

func parse_location(position golexer.TokenPos) ast.Location {
	return ast.Location{Filename: position.SourceName, Start: uint32(position.Line), End: uint32(position.Col)}
}

// func parse_trash(token string) bool {
// 	trash_regex := regexp.MustCompile(`\s+|//[^\n\r]*[\n\r]*|/\*[^*]*\*+(?:[^/*][^*]*\*+)*/`)
// 	ok := trash_regex.MatchString(token)
// 	return ok
// }
//
// func parse_identifier(token string) bool {
// 	identifier_regex := regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9/_]*`)
// 	ok := identifier_regex.MatchString(token)
// 	return ok
// }
//
// func parse_strings(token string) bool {
// 	string_regex := regexp.MustCompile(`"(\\\\|\\"|[^"\\])*"`)
// 	ok := string_regex.MatchString(token)
// 	return ok
// }
//
// func parse_numbers(token string) bool {
// 	number_regex := regexp.MustCompile(`\d+`)
// 	ok := number_regex.MatchString(token)
// 	return ok
// }
//
// func parse_symbols(token string) bool {
// 	symbols_regex := regexp.MustCompile(`[\(\)\+\-\*\/\<\>\;\{\}\,]|\!?\=\=?\>?|\|\||\&\&`)
// 	ok := symbols_regex.MatchString(token)
// 	return ok
// }
//
// func parse_keywords(token string) bool {
// 	keywords_regex := regexp.MustCompile(`^(print|first|second|true|false|if|else|fn|let)`)
// 	ok := keywords_regex.MatchString(token)
// 	return ok
// }
