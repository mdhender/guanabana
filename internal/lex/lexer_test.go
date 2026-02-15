// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package lex

import "testing"

func TestSimpleRule(t *testing.T) {
	src := []byte("expr ::= expr PLUS term.")
	tokens, err := Tokenize("test.y", src)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}
	expected := []Token{
		{Type: TOKEN_NONTERMINAL, Literal: `expr`},
		{Type: TOKEN_COLONCOLON_EQ},
		{Type: TOKEN_NONTERMINAL},
		{Type: TOKEN_TERMINAL, Literal: `PLUS`},
		{Type: TOKEN_NONTERMINAL, Literal: `term`},
		{Type: TOKEN_DOT},
		{Type: TOKEN_EOF},
	}
	if len(tokens) != len(expected) {
		for i, tok := range tokens {
			t.Errorf("%d: got %+v\n", i, tok)
		}
		t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
	}
	for i, want := range expected {
		if tokens[i].Type != want.Type {
			t.Errorf("token[%d].Type = %v, want %v (literal=%q)",
				i, tokens[i].Type, want.Type, tokens[i].Literal)
		}
		if want.Literal != "" && tokens[i].Literal != want.Literal {
			t.Errorf("token[%d].Literal = %q, want %q", i, tokens[i].Literal, want.Literal)
		}
	}
}

func TestRuleWithAction(t *testing.T) {
	src := []byte("expr(A) ::= expr(B) PLUS term(C). { A = B + C; }")
	tokens, err := Tokenize("test.y", src)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}
	expected := []Token{
		{Type: TOKEN_NONTERMINAL, Literal: `expr`},
		{Type: TOKEN_LPAREN},
		{Type: TOKEN_TERMINAL, Literal: `A`},
		{Type: TOKEN_RPAREN},
		{Type: TOKEN_COLONCOLON_EQ},
		{Type: TOKEN_NONTERMINAL, Literal: `expr`},
		{Type: TOKEN_LPAREN},
		{Type: TOKEN_TERMINAL, Literal: `B`},
		{Type: TOKEN_RPAREN},
		{Type: TOKEN_TERMINAL, Literal: `PLUS`},
		{Type: TOKEN_NONTERMINAL, Literal: `term`},
		{Type: TOKEN_LPAREN},
		{Type: TOKEN_TERMINAL, Literal: `C`},
		{Type: TOKEN_RPAREN},
		{Type: TOKEN_DOT},
		{Type: TOKEN_CODE_BLOCK, Literal: `{ A = B + C; }`},
		{Type: TOKEN_EOF},
	}
	if len(tokens) != len(expected) {
		for i, tok := range tokens {
			t.Errorf("%d: got %+v\n", i, tok)
		}
		t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
	}
	for i, want := range expected {
		if tokens[i].Type != want.Type {
			t.Errorf("token[%d].Type = %v, want %v (literal=%q)",
				i, tokens[i].Type, want.Type, tokens[i].Literal)
		}
		if want.Literal != "" && tokens[i].Literal != want.Literal {
			t.Errorf("token[%d].Literal = %q, want %q", i, tokens[i].Literal, want.Literal)
		}
	}
}

func TestDirectiveTokenization(t *testing.T) {
	src := []byte(`%left PLUS MINUS.
%left TIMES DIVIDE.
%token_type { int }`)
	tokens, err := Tokenize("test.y", src)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}
	expected := []Token{
		{Type: TOKEN_DIR_LEFT},
		{Type: TOKEN_TERMINAL},
		{Type: TOKEN_TERMINAL},
		{Type: TOKEN_DOT},
		{Type: TOKEN_DIR_LEFT},
		{Type: TOKEN_TERMINAL},
		{Type: TOKEN_TERMINAL},
		{Type: TOKEN_DOT},
		{Type: TOKEN_DIR_TOKEN_TYPE},
		{Type: TOKEN_CODE_BLOCK, Literal: `{ int }`},
		{Type: TOKEN_EOF},
	}
	if len(tokens) != len(expected) {
		for i, tok := range tokens {
			t.Errorf("%d: got %+v\n", i, tok)
		}
		t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
	}
	for i, want := range expected {
		if tokens[i].Type != want.Type {
			t.Errorf("token[%d].Type = %v, want %v (literal=%q)",
				i, tokens[i].Type, want.Type, tokens[i].Literal)
		}
		if want.Literal != "" && tokens[i].Literal != want.Literal {
			t.Errorf("token[%d].Literal = %q, want %q", i, tokens[i].Literal, want.Literal)
		}
	}
}

func TestCommentsAreSkipped(t *testing.T) {
	src := []byte(`// This is a comment
expr ::= term. /* another comment */`)
	tokens, err := Tokenize("test.y", src)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}
	expected := []Token{
		{Type: TOKEN_NONTERMINAL, Literal: `expr`},
		{Type: TOKEN_COLONCOLON_EQ},
		{Type: TOKEN_NONTERMINAL, Literal: `term`},
		{Type: TOKEN_DOT},
		{Type: TOKEN_EOF},
	}
	if len(tokens) != len(expected) {
		for i, tok := range tokens {
			t.Errorf("%d: got %+v\n", i, tok)
		}
		t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
	}
	for i, want := range expected {
		if tokens[i].Type != want.Type {
			t.Errorf("token[%d].Type = %v, want %v (literal=%q)",
				i, tokens[i].Type, want.Type, tokens[i].Literal)
		}
		if want.Literal != "" && tokens[i].Literal != want.Literal {
			t.Errorf("token[%d].Literal = %q, want %q", i, tokens[i].Literal, want.Literal)
		}
	}
}

func TestNestedBracesInCodeBlock(t *testing.T) {
	src := []byte(`
expr ::= IDENT. {
	if (x > 0) {
		y = x;
	}
}`)
	tokens, err := Tokenize("test.y", src)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}
	expected := []Token{
		{Type: TOKEN_NONTERMINAL, Literal: `expr`},
		{Type: TOKEN_COLONCOLON_EQ},
		{Type: TOKEN_TERMINAL, Literal: `IDENT`},
		{Type: TOKEN_DOT},
		{Type: TOKEN_CODE_BLOCK, Literal: "{\n\tif (x > 0) {\n\t\ty = x;\n\t}\n}"},
		{Type: TOKEN_EOF},
	}
	if len(tokens) != len(expected) {
		for i, tok := range tokens {
			t.Errorf("%d: got %+v\n", i, tok)
		}
		t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
	}
	for i, want := range expected {
		if tokens[i].Type != want.Type {
			t.Errorf("token[%d].Type = %v, want %v (literal=%q)",
				i, tokens[i].Type, want.Type, tokens[i].Literal)
		}
		if want.Literal != "" && tokens[i].Literal != want.Literal {
			t.Errorf("token[%d].Literal = %q, want %q", i, tokens[i].Literal, want.Literal)
		}
	}
}

func TestBracesInStringsInCodeBlock(t *testing.T) {
	src := []byte(`expr ::= IDENT. { x = "{"; }`)
	tokens, err := Tokenize("test.y", src)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}
	expected := []Token{
		{Type: TOKEN_NONTERMINAL, Literal: `expr`},
		{Type: TOKEN_COLONCOLON_EQ},
		{Type: TOKEN_TERMINAL, Literal: `IDENT`},
		{Type: TOKEN_DOT},
		{Type: TOKEN_CODE_BLOCK, Literal: `{ x = "}"; }`},
		{Type: TOKEN_EOF},
	}
	if len(tokens) != len(expected) {
		for i, tok := range tokens {
			t.Errorf("%d: got %+v\n", i, tok)
		}
		t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
	}
	for i, want := range expected {
		if tokens[i].Type != want.Type {
			t.Errorf("token[%d].Type = %v, want %v (literal=%q)",
				i, tokens[i].Type, want.Type, tokens[i].Literal)
		}
		if want.Literal != "" && tokens[i].Literal != want.Literal {
			t.Errorf("token[%d].Literal = %q, want %q", i, tokens[i].Literal, want.Literal)
		}
	}
}

func TestEmptyInput(t *testing.T) {
	src := []byte(``)
	tokens, err := Tokenize("test.y", src)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}
	expected := []Token{
		{Type: TOKEN_EOF},
	}
	if len(tokens) != len(expected) {
		for i, tok := range tokens {
			t.Errorf("%d: got %+v\n", i, tok)
		}
		t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
	}
	for i, want := range expected {
		if tokens[i].Type != want.Type {
			t.Errorf("token[%d].Type = %v, want %v (literal=%q)",
				i, tokens[i].Type, want.Type, tokens[i].Literal)
		}
		if want.Literal != "" && tokens[i].Literal != want.Literal {
			t.Errorf("token[%d].Literal = %q, want %q", i, tokens[i].Literal, want.Literal)
		}
	}
}

func TestPositionTracking(t *testing.T) {
	src := []byte(`expr ::= term.
factor ::= IDENT.`)
	tokens, err := Tokenize("test.y", src)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}
	expected := []Token{
		{Type: TOKEN_NONTERMINAL, Pos: Position{File: "test.y", Line: 1, Column: 1}},
		{Type: TOKEN_COLONCOLON_EQ, Pos: Position{File: "test.y", Line: 1, Column: 6}},
		{Type: TOKEN_NONTERMINAL, Pos: Position{File: "test.y", Line: 1, Column: 10}},
		{Type: TOKEN_DOT, Pos: Position{File: "test.y", Line: 1, Column: 14}},
		{Type: TOKEN_NONTERMINAL, Pos: Position{File: "test.y", Line: 2, Column: 1}},
		{Type: TOKEN_COLONCOLON_EQ, Pos: Position{File: "test.y", Line: 2, Column: 8}},
		{Type: TOKEN_TERMINAL, Pos: Position{File: "test.y", Line: 2, Column: 12}},
		{Type: TOKEN_DOT, Pos: Position{File: "test.y", Line: 2, Column: 17}},
		{Type: TOKEN_EOF, Pos: Position{File: "test.y", Line: 2, Column: 17}},
	}
	if len(tokens) != len(expected) {
		for i, tok := range tokens {
			t.Errorf("%d: got %+v\n", i, tok)
		}
		t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
	}
	for i, want := range expected {
		if tokens[i].Type != want.Type {
			t.Errorf("token[%d].Type = %v, want %v (literal=%q)",
				i, tokens[i].Type, want.Type, tokens[i].Literal)
		}
		if want.Literal != "" && tokens[i].Literal != want.Literal {
			t.Errorf("token[%d].Literal = %q, want %q", i, tokens[i].Literal, want.Literal)
		}
		if !want.Pos.IsZero() && tokens[i].Pos.String() != want.Pos.String() {
			t.Errorf("token[%d].Pos = %q, want %q", i, tokens[i].Pos.String(), want.Pos.String())
		}
	}
}
