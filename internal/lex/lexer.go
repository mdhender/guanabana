// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package lexers implements a lexer for turn reports.
// Returns tokens that contain copies from the input buffer.
package lex

import (
	"bytes"
	"fmt"

	"github.com/mdhender/guanabana/internal/scanner"
)

// TODO(lesson-02): lexer implementation

// Tokenize scans the source and returns all tokens including a final TOKEN_EOF.
// The filename is used only for Position fields in the returned tokens.
func Tokenize(filename string, src []byte) (tokens []Token, err error) {
	r := bytes.NewReader(src)
	s := &scanner.Scanner{Mode: scanner.ScanIdents}
	_, err = s.Init(r)
	if err != nil {
		return nil, err
	}
	ch := s.Scan()
	for ; ch != scanner.EOF; ch = s.Scan() {
		//value := scanner.TokenString(ch)
		//log.Printf("ch is %d, tok is %q\n", ch, value)
		tt := TOKEN_ERROR
		switch ch {
		case scanner.Is:
			tt = TOKEN_COLONCOLON_EQ
		case scanner.NonTerminal:
			tt = TOKEN_NONTERMINAL
		case scanner.Period:
			tt = TOKEN_DOT
		case scanner.Terminal:
			tt = TOKEN_TERMINAL
		default:
			tt = TOKEN_ERROR
		}
		tokens = append(tokens, Token{Type: tt})
	}
	if ch != scanner.EOF {
		return nil, fmt.Errorf("scanner did not return EOF")
	}
	tokens = append(tokens, Token{Type: TOKEN_EOF})
	return tokens, nil
}
