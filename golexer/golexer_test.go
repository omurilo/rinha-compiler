package golexer

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/davyxu/golexer"
	"github.com/omurilo/rinha-compiler/lexer"
)

// const (
// 	Token_EOF = iota
// 	Token_Unknown
// 	Token_Numeral
// 	Token_String
// 	Token_WhiteSpace
// 	Token_LineEnd
// 	Token_UnixStyleComment
// 	Token_Identifier
// 	Token_Semicolon
// 	Token_If
// 	Token_First
// 	Token_Second
// 	Token_Print
// 	Token_Add
// 	Token_Sub
// 	Token_Mul
// 	Token_Div
// 	Token_Rem
// 	Token_Eq
// 	Token_Let
// 	Token_Fn
// 	Token_Comma
// 	Token_Lte
// 	Token_Gte
// 	Token_Lt
// 	Token_Gt
// 	Token_Neq
// 	Token_Assignment
// 	Token_Arrow
// 	Token_LParen
// 	Token_RParen
// 	Token_LBrace
// 	Token_RBrace
// )
//
// type CustomParser struct {
// 	*golexer.Parser
// }
//
// func NewCustomParser() *CustomParser {
// 	l := golexer.NewLexer()
//
// 	l.AddMatcher(golexer.NewNumeralMatcher(Token_Numeral))
// 	l.AddMatcher(golexer.NewStringMatcher(Token_String))
//
// 	l.AddIgnoreMatcher(golexer.NewWhiteSpaceMatcher(Token_WhiteSpace))
// 	l.AddIgnoreMatcher(golexer.NewLineEndMatcher(Token_LineEnd))
// 	l.AddIgnoreMatcher(golexer.NewUnixStyleCommentMatcher(Token_UnixStyleComment))
//
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Semicolon, ";"))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Comma, ","))
//
// 	l.AddMatcher(golexer.NewKeywordMatcher(Token_If, "if"))
// 	l.AddMatcher(golexer.NewKeywordMatcher(Token_Print, "print"))
// 	l.AddMatcher(golexer.NewKeywordMatcher(Token_First, "first"))
// 	l.AddMatcher(golexer.NewKeywordMatcher(Token_Second, "second"))
// 	l.AddMatcher(golexer.NewKeywordMatcher(Token_Let, "let"))
// 	l.AddMatcher(golexer.NewKeywordMatcher(Token_Fn, "fn"))
//
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Add, "+"))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Sub, "-"))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Mul, "*"))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Div, "/"))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Rem, "%"))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Arrow, "=>"))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Lt, "<"))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Gt, ">"))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Lte, "<="))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Gte, ">="))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Eq, "=="))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Neq, "!="))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_Assignment, "="))
//
// 	l.AddMatcher(golexer.NewSignMatcher(Token_LParen, "("))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_RParen, ")"))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_LBrace, "{"))
// 	l.AddMatcher(golexer.NewSignMatcher(Token_RBrace, "}"))
//
// 	l.AddMatcher(golexer.NewIdentifierMatcher(Token_Identifier))
//
// 	l.AddMatcher(golexer.NewUnknownMatcher(Token_Unknown))
//
// 	return &CustomParser{
// 		Parser: golexer.NewParser(l, "custom"),
// 	}
// }

func TestPrintHello(t *testing.T) {
	p := lexer.NewCustomParser("test")

	defer golexer.ErrorCatcher(func(err error) {
		t.Error(err.Error())
	})

	p.Lexer().Start(`print("Hello world!")`)

	p.NextToken()

	var b bytes.Buffer

	b.WriteString("===\n")

	for p.TokenID() != 0 {

		b.WriteString(fmt.Sprintf("MatcherName: '%s' Value: '%s'\n", p.MatcherName(), p.TokenValue()))

		p.NextToken()

	}

	b.WriteString("===\n")

	fmt.Println(b.String())
}

func TestFibonacciTrampoline(t *testing.T) {
	p := lexer.NewCustomParser("test")

	defer golexer.ErrorCatcher(func(err error) {
		t.Error(err.Error())
	})

	p.Lexer().Start(`let fib = fn(n, a, b) => {
    if (n == 0) {
        a
    } else {
      fn() => { fib(n, b, a + b) }
    }
  };
  print("fib: " + fib(10, 0, 1))`)

	p.NextToken()

	var b bytes.Buffer

	b.WriteString("===\n")

	for p.TokenID() != 0 {

		b.WriteString(fmt.Sprintf("MatcherName: '%s' Value: '%s'\n", p.MatcherName(), p.TokenValue()))

		p.NextToken()

	}

	b.WriteString("===\n")

	fmt.Println(b.String())
}

func TestTokenizeFibTrampoline(t *testing.T) {
	p := lexer.NewCustomParser("test")

	defer golexer.ErrorCatcher(func(err error) {
		t.Error(err.Error())
	})

	p.Lexer().Start(`let fib = fn(n, a, b) => {
    if (n == 0) {
        a
    } else {
      fn() => { fib(n, b, a + b) }
    }
  };
  print("fib: " + fib(10, 0, 1))`)

	token := p.Next()

	var b bytes.Buffer

	b.WriteString("===\n")

	for p.TokenID() != 0 {

		b.WriteString(fmt.Sprintf("%v\n", token))

		token = p.Next()

	}

	b.WriteString("===\n")

	fmt.Println(b.String())
}

func TestTokenizeSymbols(t *testing.T) {
	p := lexer.NewCustomParser("test")

	defer golexer.ErrorCatcher(func(err error) {
		t.Error(err.Error())
	})

	p.Lexer().Start(`if (1 > 2 || 3 < 2 || 4 >= 5 || 5 != 4 || 6 <= 5 && 1 == 1) {
    print("hello world")
  } else {
      print("welcome to hell")
    }
`)

	token := p.Next()

	var b bytes.Buffer

	b.WriteString("===\n")

	for p.TokenID() != 0 {

		b.WriteString(fmt.Sprintf("%v\n", token))

		token = p.Next()

	}

	b.WriteString("===\n")

	fmt.Println(b.String())
}
