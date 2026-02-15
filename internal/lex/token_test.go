// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package lex

import (
	"testing"
)

func TestTokenTypeString(t *testing.T) {
	types := []TokenType{
		TOKEN_EOF, TOKEN_ERROR, TOKEN_TERMINAL, TOKEN_NONTERMINAL,
		TOKEN_COLONCOLON_EQ, TOKEN_DOT, TOKEN_PIPE,
		TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_LBRACKET, TOKEN_RBRACKET, TOKEN_COMMA,
		TOKEN_DIR_LEFT, TOKEN_DIR_RIGHT, TOKEN_DIR_NONASSOC,
		TOKEN_DIR_TOKEN_TYPE, TOKEN_DIR_TYPE, TOKEN_DIR_START_SYMBOL,
		TOKEN_DIR_NAME, TOKEN_DIR_INCLUDE, TOKEN_DIR_CODE,
		TOKEN_DIR_DEFAULT_TYPE, TOKEN_DIR_EXTRA_ARGUMENT,
		TOKEN_DIR_TOKEN_PREFIX, TOKEN_DIR_FALLBACK, TOKEN_DIR_WILDCARD,
		TOKEN_DIR_DESTRUCTOR, TOKEN_DIR_SYNTAX_ERROR,
		TOKEN_DIR_PARSE_ACCEPT, TOKEN_DIR_PARSE_FAILURE, TOKEN_DIR_STACK_OVERFLOW,
		TOKEN_CODE_BLOCK, TOKEN_STRING,
	}
	seen := map[string]bool{}
	for _, tt := range types {
		s := tt.String()
		if s == "" {
			t.Errorf("TokenType %d has empty string", tt)
		}
		if seen[s] {
			t.Errorf("duplicate String() value: %q", s)
		}
		seen[s] = true
	}
}

func TestTokenZeroValueIsEOF(t *testing.T) {
	var tok Token
	if tok.Type != TOKEN_EOF {
		t.Errorf("zero-value Token.Type = %v, want TOKEN_EOF", tok.Type)
	}
}

func TestPositionString(t *testing.T) {
	p := Position{File: "calc.y", Line: 10, Column: 5}
	s := p.String() // should return "calc.y:10:5"
	if s != "calc.y:10:5" {
		t.Errorf("Position.String() = %q, want %q", s, "calc.y:10:5")
	}
}
