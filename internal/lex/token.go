// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package lex

import "fmt"

//go:generate stringer --type TokenType

// Position records where a token was found in the source.
type Position struct {
	File   string
	Line   int
	Column int
}

func (p Position) IsZero() bool { return p == Position{} }

func (p Position) String() string {
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
}

// TokenType classifies a token.
type TokenType int

const (
	// Special
	TOKEN_EOF   TokenType = iota
	TOKEN_ERROR           // lexer error

	// Identifiers
	TOKEN_TERMINAL    // UPPER_CASE identifier (e.g., PLUS, INTEGER)
	TOKEN_NONTERMINAL // lower_case identifier (e.g., expr, stmt)

	// Punctuation
	TOKEN_COLONCOLON_EQ // ::=
	TOKEN_DOT           // . (rule terminator)
	TOKEN_PIPE          // | (alternative, if supported)
	TOKEN_LPAREN        // (
	TOKEN_RPAREN        // )
	TOKEN_LBRACKET      // [
	TOKEN_RBRACKET      // ]
	TOKEN_COMMA         // ,

	// Directives
	TOKEN_DIR_CODE // %code
	TOKEN_DIR_DEFAULT_DESTRUCTOR
	TOKEN_DIR_DEFAULT_TYPE // %default_type
	TOKEN_DIR_ENDIF
	TOKEN_DIR_EXTRA_ARGUMENT // %extra_argument
	TOKEN_DIR_EXTRA_CONTEXT
	TOKEN_DIR_INCLUDE // %include
	TOKEN_DIR_IFDEF
	TOKEN_DIR_IFNDEF
	TOKEN_DIR_LEFT     // %left
	TOKEN_DIR_NAME     // %name
	TOKEN_DIR_NONASSOC // %nonassoc
	TOKEN_DIR_RIGHT    // %right
	TOKEN_DIR_STACK_SIZE
	TOKEN_DIR_START_SYMBOL // %start_symbol
	TOKEN_DIR_TOKEN_CLASS
	TOKEN_DIR_TOKEN_DESTRUCTOR
	TOKEN_DIR_TOKEN_PREFIX   // %token_prefix
	TOKEN_DIR_TOKEN_TYPE     // %token_type
	TOKEN_DIR_TYPE           // %type
	TOKEN_DIR_FALLBACK       // %fallback
	TOKEN_DIR_WILDCARD       // %wildcard
	TOKEN_DIR_DESTRUCTOR     // %destructor
	TOKEN_DIR_SYNTAX_ERROR   // %syntax_error
	TOKEN_DIR_PARSE_ACCEPT   // %parse_accept
	TOKEN_DIR_PARSE_FAILURE  // %parse_failure
	TOKEN_DIR_STACK_OVERFLOW // %stack_overflow
	TOKEN_DIR_GENERIC        // unknown %directive

	// Code blocks
	TOKEN_CODE_BLOCK // { ... } (Go/C code in actions or directives)

	// Alias
	TOKEN_STRING // "quoted string" (alias for a token)
)

// Token is a single lexical unit from a Lemon grammar file.
type Token struct {
	Type    TokenType
	Literal string   // the raw text
	Pos     Position // where it appeared

	// The Leading and Trailing Trivia aren't used yet.
	// They're used to rebuild the original source for error reporting.
	LeadingTrivia  []*Span
	TrailingTrivia []*Span
}

type Span struct {
	Line  int // 1-based
	Col   int // 1-based, in UTF-8 code points
	Type  TokenType
	Value string // the raw text
}

// Length returns the length of the span.
func (s *Span) Length() int {
	if s == nil {
		return 0
	}
	return len(s.Value)
}

// LineNo returns the line number of the start of the span
func (s *Span) LineNo() int {
	if s == nil {
		return 0
	}
	return s.Line
}

// ColNo returns the column number of the start of the span
func (s *Span) ColNo() int {
	if s == nil {
		return 0
	}
	return s.Col
}

// Bytes is a helper for diagnostics / debugging.
func (s *Span) Bytes() []byte {
	if s == nil {
		return nil
	}
	return []byte(s.Value)
}
