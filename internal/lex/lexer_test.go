// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package lex

import "testing"

func TestLexer_SimpleRule(t *testing.T) {
	input := "expr ::= expr PLUS term."
	expected := []struct {
		id    int
		Type  TokenType
		Value string
	}{
		{id: 1, Type: TOKEN_NONTERMINAL, Value: "expr"},
		{id: 2, Type: TOKEN_COLONCOLON_EQ, Value: "::="},
		{id: 3, Type: TOKEN_NONTERMINAL, Value: "expr"},
		{id: 4, Type: TOKEN_TERMINAL, Value: "PLUS"},
		{id: 5, Type: TOKEN_NONTERMINAL, Value: "term"},
		{id: 6, Type: TOKEN_DOT, Value: "."},
		{id: 7, Type: TOKEN_EOF, Value: ""},
	}
	tokens, err := Tokenize("<>", []byte(input))
	if err != nil {
		t.Fatalf("tokenize: failed %v\n", err)
	}
	for _, tc := range expected {
		token := tokens[0]
		tokens = tokens[1:]
		if token.Type != tc.Type {
			t.Fatalf("%d: want %q, got %q\n", tc.id, tc.Type, token.Type)
		}
	}
}
