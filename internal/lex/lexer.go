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
	s := &scanner.Scanner{}
	_, err = s.Init(r)
	if err != nil {
		return nil, err
	}
	ch := s.Scan()
	for ; ch != scanner.EOF; ch = s.Scan() {
		//value := scanner.TokenString(ch)
		//log.Printf("ch is %d, tok is %q\n", ch, value)
		tt, literal := TOKEN_ERROR, s.TokenText()
		pos := Position{File: s.Filename, Line: s.Line, Column: s.Column}
		switch ch {
		case '.':
			tt = TOKEN_DOT
		case '(':
			tt = TOKEN_LPAREN
		case ')':
			tt = TOKEN_RPAREN
		case '[':
			tt = TOKEN_LBRACKET
		case ']':
			tt = TOKEN_RBRACKET
		case ',':
			tt = TOKEN_COMMA
		case '|':
			tt = TOKEN_PIPE

		case scanner.Action:
			tt = TOKEN_CODE_BLOCK
		case scanner.Code:
			tt = TOKEN_DIR_CODE
		case scanner.DefaultDestructor:
			tt = TOKEN_DIR_DEFAULT_DESTRUCTOR
		case scanner.DefaultType:
			tt = TOKEN_DIR_DEFAULT_TYPE
		case scanner.Destructor:
			tt = TOKEN_DIR_DESTRUCTOR
		case scanner.EndIf:
			tt = TOKEN_DIR_ENDIF
		case scanner.ExtraArgument:
			tt = TOKEN_DIR_EXTRA_ARGUMENT
		case scanner.ExtraContext:
			tt = TOKEN_DIR_EXTRA_CONTEXT
		case scanner.Fallback:
			tt = TOKEN_DIR_FALLBACK
		case scanner.IfDef:
			tt = TOKEN_DIR_IFDEF
		case scanner.IfNDdef:
			tt = TOKEN_DIR_IFNDEF
		case scanner.Include:
			tt = TOKEN_DIR_INCLUDE
		case scanner.Is:
			tt = TOKEN_COLONCOLON_EQ
		case scanner.Left:
			tt = TOKEN_DIR_LEFT
		case scanner.Name:
			tt = TOKEN_DIR_NAME
		case scanner.NonAssoc:
			tt = TOKEN_DIR_NONASSOC
		case scanner.NonTerminal:
			tt = TOKEN_NONTERMINAL
		case scanner.ParseAccept:
			tt = TOKEN_DIR_PARSE_ACCEPT
		case scanner.ParseFailure:
			tt = TOKEN_DIR_PARSE_FAILURE
		case scanner.Right:
			tt = TOKEN_DIR_RIGHT
		case scanner.StackSize:
			tt = TOKEN_DIR_STACK_SIZE
		case scanner.StackOverflow:
			tt = TOKEN_DIR_STACK_OVERFLOW
		case scanner.StartSymbol:
			tt = TOKEN_DIR_START_SYMBOL
		case scanner.String:
			tt = TOKEN_STRING
		case scanner.SyntaxError:
			tt = TOKEN_DIR_SYNTAX_ERROR
		case scanner.Terminal:
			tt = TOKEN_TERMINAL
		case scanner.TokenClass:
			tt = TOKEN_DIR_TOKEN_CLASS
		case scanner.TokenDestructor:
			tt = TOKEN_DIR_TOKEN_DESTRUCTOR
		case scanner.TokenPrefix:
			tt = TOKEN_DIR_TOKEN_PREFIX
		case scanner.TokenType:
			tt = TOKEN_DIR_TOKEN_TYPE
		case scanner.Type:
			tt = TOKEN_DIR_TYPE
		case scanner.Wildcard:
			tt = TOKEN_DIR_WILDCARD
		case scanner.Directive:
			tt = TOKEN_DIR_GENERIC
		default:
			tt = TOKEN_ERROR
		}
		tokens = append(tokens, Token{
			Type:    tt,
			Literal: literal,
			Pos:     pos,
		})
	}
	if ch != scanner.EOF {
		return nil, fmt.Errorf("scanner did not return EOF")
	}
	tokens = append(tokens, Token{
		Type: TOKEN_EOF,
		Pos: Position{
			File:   s.Filename,
			Line:   s.Line,
			Column: s.Column,
		},
	})
	return tokens, nil
}
