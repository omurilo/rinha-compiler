package lexer

import (
	"fmt"
	"unicode/utf8"

	"github.com/davyxu/golexer"
	"github.com/omurilo/rinha-compiler/ast"
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
	Token_Tuple
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
	lineOffsets []int
	startCols   []int // Col value at the start of each line (golexer-inflated)
	source      string
}

// buildLineInfo simulates golexer's LineEndMatcher to compute, for each line:
//   - lineOffsets[i]: byte offset of the first char of line i+1
//   - startCols[i]:   the Col value at the start of line i+1
//
// golexer's LineEndMatcher processes ALL consecutive newline chars in one call,
// calling increaseLine() once per \n (resets Col=1) then ConsumeMulti(runLen)
// where runLen = total chars in the run (including \r). This means the starting
// Col for lines inside a multi-newline run is 1+runLen, not 1.
func buildLineInfo(source string) (lineOffsets []int, startCols []int) {
	lineOffsets = []int{0}
	startCols = []int{1} // line 1: Col starts at 1 (never reset)

	i := 0
	for i < len(source) {
		if source[i] == '\n' || source[i] == '\r' {
			runStart := i
			for i < len(source) && (source[i] == '\n' || source[i] == '\r') {
				i++
			}
			runLen := i - runStart
			startingCol := 1 + runLen

			for j := runStart; j < i; j++ {
				if source[j] == '\n' {
					lineOffsets = append(lineOffsets, j+1)
					startCols = append(startCols, startingCol)
				}
			}
		} else {
			i++
		}
	}
	return
}

func NewCustomParser(filename string) *CustomParser {
	l := golexer.NewLexer()

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

	l.AddMatcher(golexer.NewNumeralMatcher(Token_Numeral))
	l.AddMatcher(golexer.NewStringMatcher(Token_String))

	l.AddMatcher(golexer.NewIdentifierMatcher(Token_Identifier))

	l.AddMatcher(golexer.NewUnknownMatcher(Token_Unknown))

	return &CustomParser{
		Parser: golexer.NewParser(l, filename),
	}
}

func Initialize(program string, filename string) *CustomParser {
	p := NewCustomParser(filename)
	p.Lexer().Start(program)
	p.source = program
	p.lineOffsets, p.startCols = buildLineInfo(program)
	return p
}

func (p *CustomParser) Next() Token {
	p.NextToken()
	return p.parseTokenMatcher()
}

func (p *CustomParser) Tokenize() {
	token := p.Next()
	for p.TokenID() != 0 {
		fmt.Println(token)
		token = p.Next()
	}
}

// tokenLoc computes the byte-offset location of the current token.
// rawLen is the number of RUNES the token occupies in source (value + 2 for strings).
// We convert rune offset to byte offset to handle multi-byte UTF-8 characters.
func (p *CustomParser) tokenLoc(rawRuneLen int) ast.Location {
	pos := p.TokenPos()
	end := 0
	if pos.Line >= 1 && pos.Line-1 < len(p.lineOffsets) {
		lineByteStart := p.lineOffsets[pos.Line-1]
		startCol := p.startCols[pos.Line-1]
		runeOffset := pos.Col - startCol // runes consumed on this line up to end of token
		end = p.runeOffsetToByteOffset(lineByteStart, runeOffset)
	}
	start := end - runeLen2ByteLen(p.source, end, rawRuneLen)
	if start < 0 {
		start = 0
	}
	return ast.Location{Filename: pos.SourceName, Start: uint32(start), End: uint32(end)}
}

// runeOffsetToByteOffset converts a rune count from lineByteStart to a byte offset.
func (p *CustomParser) runeOffsetToByteOffset(lineByteStart, runeCount int) int {
	bytePos := lineByteStart
	for i := 0; i < runeCount && bytePos < len(p.source); i++ {
		_, size := utf8.DecodeRuneInString(p.source[bytePos:])
		bytePos += size
	}
	return bytePos
}

// runeLen2ByteLen returns the byte length of rawRuneLen runes ending at byteEnd.
func runeLen2ByteLen(source string, byteEnd, runeLen int) int {
	byteStart := byteEnd
	for i := 0; i < runeLen && byteStart > 0; i++ {
		_, size := utf8.DecodeLastRuneInString(source[:byteStart])
		byteStart -= size
	}
	return byteEnd - byteStart
}

func (p *CustomParser) parseTokenMatcher() Token {
	var token Token

	if p.TokenID() != 0 {
		val := p.TokenValue()
		runeLen := utf8.RuneCountInString(val)
		switch p.MatcherName() {
		case "NumeralMatcher":
			token = Token{Type: ":NUMBER", Value: val, Location: p.tokenLoc(runeLen)}
		case "StringMatcher":
			token = Token{Type: ":STRING", Value: val, Location: p.tokenLoc(runeLen + 2)}
		case "SignMatcher":
			token = Token{Type: SYMBOLS[val], Value: val, Location: p.tokenLoc(runeLen)}
		case "KeywordMatcher":
			token = Token{Type: KEYWORDS[val], Value: val, Location: p.tokenLoc(runeLen)}
		case "IdentifierMatcher":
			token = Token{Type: ":IDENTIFIER", Value: val, Location: p.tokenLoc(runeLen)}
		}
	}

	return token
}
