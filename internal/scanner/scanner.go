// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package scanner provides a scanner and tokenizer for UTF-8-encoded
// Guanabana grammar files.
package scanner

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

// Position represents a location in the source.
// A position is valid if Line > 0.
type Position struct {
	Filename string // filename, if any
	Offset   int    // byte offset, starting at 0
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1 (character count per line)
}

// IsValid reports whether the position is valid.
func (pos *Position) IsValid() bool { return pos.Line > 0 }

func (pos Position) String() string {
	s := pos.Filename
	if s == "" {
		s = "<input>"
	}
	if pos.IsValid() {
		s += fmt.Sprintf(":%d:%d", pos.Line, pos.Column)
	}
	return s
}

// Predefined mode bits to control recognition of tokens.
// Unrecognized tokens are returned as their individual Unicode characters.
const (
	ScanIdents    = 1 << -Ident
	ScanStrings   = 1 << -String
	ScanComments  = 1 << -Comment
	SkipComments  = 1 << -skipComment // if set with ScanComments, comments become white space
	DefaultTokens = ScanIdents | ScanStrings | ScanComments | SkipComments
	GoTokens      = DefaultTokens // compatibility alias
)

// The result of Scan is one of these tokens or a Unicode character.
const (
	EOF = -(iota + 1)
	Ident
	Int
	Float
	Char
	String
	RawString
	Comment
	skipComment
	Action
	Code
	DefaultDestructor
	DefaultType
	Destructor
	Directive
	EndIf
	ExtraArgument
	ExtraContext
	Fallback
	IfDef
	IfNDdef
	Include
	Is
	Left
	Name
	NonAssoc
	NonTerminal
	ParseAccept
	ParseFailure
	Period
	Right
	StackOverflow
	StackSize
	StartSymbol
	SyntaxError
	Terminal
	TokenClass
	TokenDestructor
	TokenPrefix
	TokenType
	Type
	Wildcard
)

var tokenString = map[rune]string{
	EOF:               "EOF",
	Ident:             "Ident",
	Int:               "Int",
	Float:             "Float",
	Char:              "Char",
	String:            "String",
	RawString:         "RawString",
	Comment:           "Comment",
	Action:            "Action",
	Code:              "Code",
	DefaultDestructor: "DefaultDestructor",
	DefaultType:       "DefaultType",
	Destructor:        "Destructor",
	Directive:         "Directive",
	EndIf:             "EndIf",
	ExtraArgument:     "ExtraArgument",
	ExtraContext:      "ExtraContext",
	Fallback:          "Fallback",
	IfDef:             "IfDef",
	IfNDdef:           "IfNDef",
	Include:           "Include",
	Is:                "::=",
	Left:              "Left",
	Name:              "Name",
	NonAssoc:          "NonAssoc",
	NonTerminal:       "NonTerminal",
	ParseAccept:       "ParseAccept",
	ParseFailure:      "ParseFailure",
	Period:            ".",
	Right:             "Right",
	StackOverflow:     "StackOverflow",
	StackSize:         "StackSize",
	StartSymbol:       "StartSymbol",
	SyntaxError:       "SyntaxError",
	Terminal:          "Terminal",
	TokenClass:        "TokenClass",
	TokenDestructor:   "TokenDestructor",
	TokenPrefix:       "TokenPrefix",
	TokenType:         "TokenType",
	Type:              "Type",
	Wildcard:          "Wildcard",
}

// TokenString returns a printable string for a token or Unicode character.
func TokenString(tok rune) string {
	if s, found := tokenString[tok]; found {
		return s
	}
	return fmt.Sprintf("%q", string(tok))
}

// DefaultWhitespace is the default value for the Scanner's Whitespace field.
const DefaultWhitespace = 1<<'\t' | 1<<'\n' | 1<<'\r' | 1<<' '

const GoWhitespace = DefaultWhitespace // compatibility alias

// A Scanner implements reading of Unicode characters and tokens from an io.Reader.
// mdhender: updated to read the entire buffer into srcBuf on initialization.
type Scanner struct {
	// Source buffer
	srcBuf []byte
	srcPos int // reading position (srcBuf index)

	// Source position
	line        int // line count
	column      int // character count
	lastLineLen int // length of last line in characters (for correct column reporting)
	lastCharLen int // length of last character in bytes

	// Token text buffer
	tokBuf bytes.Buffer
	tokPos int // token text tail position (srcBuf index); valid if >= 0
	tokEnd int // token text tail end (srcBuf index)

	// One character look-ahead
	ch rune // character before current srcPos

	// Error is called for each error encountered.
	Error func(s *Scanner, msg string)

	// ErrorCount is incremented by one for each error encountered.
	ErrorCount int

	// The Mode field controls which tokens are recognized.
	// The field may be changed at any time.
	Mode uint

	// The Whitespace field controls which characters are recognized
	// as white space.
	Whitespace uint64

	// IsIdentRune is a predicate controlling the characters accepted
	// as the ith rune in an identifier.
	IsIdentRune func(ch rune, i int) bool

	// Start position of most recently scanned token; set by Scan.
	Position

	// ErrorLog captures all error messages, usually one line per message.
	ErrorLog *bytes.Buffer
}

// Init initializes a Scanner with a new source and returns s.
// If Mode is 0, it is set to DefaultTokens.
// If Whitespace is 0, it is set to DefaultWhitespace.
func (s *Scanner) Init(r io.Reader) (*Scanner, error) {
	if buf, err := io.ReadAll(r); err != nil {
		return nil, err
	} else {
		s.srcBuf = buf
	}
	s.srcPos = 0

	s.line = 1
	s.column = 0
	s.lastLineLen = 0
	s.lastCharLen = 0

	s.tokPos = -1
	s.ch = -2 // no char read yet, not EOF

	s.ErrorLog = &bytes.Buffer{}
	if s.Error == nil {
		s.Error = func(pS *Scanner, pMsg string) {
			pos := pS.Position
			if !pos.IsValid() {
				pos = pS.Pos()
			}
			s.ErrorLog.WriteString(fmt.Sprintf("%s: %s\n", pos, pMsg))
		}
	}

	s.ErrorCount = 0
	if s.Mode == 0 {
		s.Mode = DefaultTokens
	}
	if s.Whitespace == 0 {
		s.Whitespace = DefaultWhitespace
	}
	s.Line = 0 // invalidate token position

	return s, nil
}

// next reads and returns the next Unicode character.
func (s *Scanner) next() rune {
	if !(s.srcPos < len(s.srcBuf)) {
		s.lastCharLen = 0
		return EOF
	}

	ch, width := utf8.DecodeRune(s.srcBuf[s.srcPos:])
	if ch == utf8.RuneError && width == 1 {
		s.srcPos += width
		s.lastCharLen = width
		s.column++
		s.error("illegal UTF-8 encoding")
		return ch
	}

	s.srcPos += width
	s.lastCharLen = width
	s.column++

	switch ch {
	case 0:
		s.error("illegal character NUL")
	case '\n':
		s.line++
		s.lastLineLen = s.column
		s.column = 0
	}

	return ch
}

// Next reads and returns the next Unicode character.
func (s *Scanner) Next() rune {
	s.tokPos = -1 // don't collect token text
	s.Line = 0    // invalidate token position
	ch := s.Peek()
	if ch != EOF {
		s.ch = s.next()
	}
	return ch
}

// Peek returns the next Unicode character in the source without advancing the scanner.
func (s *Scanner) Peek() rune {
	if s.ch == -2 {
		s.ch = s.next()
		if s.ch == '\uFEFF' {
			s.ch = s.next() // ignore BOM
		}
	}
	return s.ch
}

func (s *Scanner) error(msg string) {
	s.ErrorCount++
	if s.Error == nil {
		return
	}
	s.Error(s, msg)
}

func (s *Scanner) isIdentRune(ch rune, i int) bool {
	if s.IsIdentRune != nil {
		return s.IsIdentRune(ch, i)
	}
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
}

func (s *Scanner) scanAction() rune {
	ch := s.next()
	level := 1
	for level > 0 && ch != EOF {
		if ch == '{' {
			level++
		} else if ch == '}' {
			level--
		}
		ch = s.next()
	}
	if level != 0 {
		s.error("unterminated action")
	}
	return ch
}

func (s *Scanner) scanDirective() rune {
	ch := s.next()
	for i := 1; s.isIdentRune(ch, i); i++ {
		ch = s.next()
	}
	return ch
}

func (s *Scanner) scanIdentifier() rune {
	ch := s.next()
	for i := 1; s.isIdentRune(ch, i); i++ {
		ch = s.next()
	}
	return ch
}

func (s *Scanner) scanString(quote rune) {
	ch := s.next() // read character after quote
	for ch != quote {
		if ch == '\n' || ch < 0 {
			s.error("literal not terminated")
			return
		}
		if ch == '\\' {
			ch = s.next()
			if ch < 0 {
				s.error("literal not terminated")
				return
			}
			ch = s.next()
		} else {
			ch = s.next()
		}
	}
}

func (s *Scanner) scanComment(ch rune) rune {
	if ch == '/' {
		ch = s.next() // read character after "//"
		for ch != '\n' && ch >= 0 {
			ch = s.next()
		}
		return ch
	}

	ch = s.next() // read character after "/*"
	for {
		if ch < 0 {
			s.error("comment not terminated")
			break
		}
		ch0 := ch
		ch = s.next()
		if ch0 == '*' && ch == '/' {
			ch = s.next()
			break
		}
	}
	return ch
}

// Scan reads the next token or Unicode character from source and returns it.
func (s *Scanner) Scan() rune {
	ch := s.Peek()

	s.tokPos = -1
	s.Line = 0

redo:
	for ch <= ' ' && s.Whitespace&(1<<uint(ch)) != 0 {
		ch = s.next()
	}

	s.tokBuf.Reset()
	s.tokPos = s.srcPos - s.lastCharLen

	s.Offset = s.tokPos
	if s.column > 0 {
		s.Line = s.line
		s.Column = s.column
	} else {
		s.Line = s.line - 1
		s.Column = s.lastLineLen
	}

	tok := ch
	switch {
	case s.isIdentRune(ch, 0):
		if s.Mode&ScanIdents != 0 {
			if unicode.IsUpper(tok) {
				tok = Terminal
			} else if unicode.IsLower(tok) {
				tok = NonTerminal
			} else {
				tok = Ident
			}
			ch = s.scanIdentifier()
		} else {
			ch = s.next()
		}
	default:
		switch ch {
		case EOF:
			break
		case '"':
			if s.Mode&ScanStrings != 0 {
				s.scanString('"')
				tok = String
			}
			ch = s.next()
		case '/':
			ch = s.next()
			if (ch == '/' || ch == '*') && s.Mode&ScanComments != 0 {
				if s.Mode&SkipComments != 0 {
					s.tokPos = -1 // don't collect token text
					ch = s.scanComment(ch)
					goto redo
				}
				ch = s.scanComment(ch)
				tok = Comment
			}
		case ':':
			ch = s.next()
			if ch == ':' {
				ch = s.next()
				if ch == '=' {
					tok = Is
					ch = s.next()
				} else {
					s.error("expected '=' after '::'")
					tok = ':'
				}
			} else {
				tok = ':'
			}
		case '{':
			tok = Action
			ch = s.scanAction()
		case '%':
			ch = s.scanDirective()
			s.tokEnd = s.srcPos - s.lastCharLen
			switch s.TokenText() {
			case "%code":
				tok = Code
			case "%default_destructor":
				tok = DefaultDestructor
			case "%default_type":
				tok = DefaultType
			case "%destructor":
				tok = Destructor
			case "%endif":
				tok = EndIf
			case "%extra_argument":
				tok = ExtraArgument
			case "%extra_context":
				tok = ExtraContext
			case "%fallback":
				tok = Fallback
			case "%ifdef":
				tok = IfDef
			case "%ifndef":
				tok = IfNDdef
			case "%include":
				tok = Include
			case "%left":
				tok = Left
			case "%name":
				tok = Name
			case "%nonassoc":
				tok = NonAssoc
			case "%parse_accept":
				tok = ParseAccept
			case "%parse_failure":
				tok = ParseFailure
			case "%right":
				tok = Right
			case "%stack_overflow":
				tok = StackOverflow
			case "%stack_size":
				tok = StackSize
			case "%start_symbol":
				tok = StartSymbol
			case "%syntax_error":
				tok = SyntaxError
			case "%token_class":
				tok = TokenClass
			case "%token_destructor":
				tok = TokenDestructor
			case "%token_prefix":
				tok = TokenPrefix
			case "%token_type":
				tok = TokenType
			case "%type":
				tok = Type
			case "%wildcard":
				tok = Wildcard
			default:
				tok = Directive
			}
		default:
			ch = s.next()
		}
	}

	s.tokEnd = s.srcPos - s.lastCharLen
	s.ch = ch
	return tok
}

// Pos returns the position of the character immediately after
// the character or token returned by the last call to Next or Scan.
func (s *Scanner) Pos() (pos Position) {
	pos.Filename = s.Filename
	pos.Offset = s.srcPos - s.lastCharLen
	switch {
	case s.column > 0:
		pos.Line = s.line
		pos.Column = s.column
	case s.lastLineLen > 0:
		pos.Line = s.line - 1
		pos.Column = s.lastLineLen
	default:
		pos.Line = 1
		pos.Column = 1
	}
	return
}

// TokenText returns the string corresponding to the most recently scanned token.
// Valid after calling Scan().
func (s *Scanner) TokenText() string {
	if s.tokPos < 0 {
		return ""
	}

	if s.tokEnd < 0 {
		s.tokEnd = s.tokPos
	}

	if s.tokBuf.Len() == 0 {
		return string(s.srcBuf[s.tokPos:s.tokEnd])
	}

	s.tokBuf.Write(s.srcBuf[s.tokPos:s.tokEnd])
	s.tokPos = s.tokEnd // ensure idempotency of TokenText() call
	return s.tokBuf.String()
}
