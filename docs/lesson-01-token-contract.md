# Lesson 01: Token Types and the Lexer Contract

**Goal**: Define the token types, the `Token` struct, and the contract between
the lexer and the parser so that every later lesson has a stable foundation.

## What you will build in this lesson

- A `TokenType` enum (Go `int` constants) covering all token kinds that appear
  in Lemon grammar files
- A `Token` struct with type, literal value, and source location
- A `TokenStream` interface that the parser will consume
- A set of representative token type constants with `String()` for debugging

## Key concepts

**Terminals vs. nonterminals in the grammar file**: When Lemon reads a `.y`
file, it sees terminal symbol names (UPPER_CASE by convention), nonterminal
symbol names (lower_case), directives (`%left`, `%type`, etc.), C/Go code
blocks in braces, punctuation (`.`, `::=`, `|`), and comments. Our lexer must
classify each of these into a token type.

**The token stream contract**: The parser will consume tokens one at a time.
Each token must carry:
1. Its type (what kind of thing it is)
2. Its literal text (the actual characters from the source)
3. Its position (file, line, column) for error messages

**Lexer drives the parser**: In Lemon's architecture, the lexer scans the
input and calls `Parse(tokenType, tokenValue)` for each token. We mirror this:
the lexer iterates over tokens and feeds them to the parser. The token types
defined here are the shared vocabulary.

## Inputs / Outputs

### Files to create

| File | Purpose |
|------|---------|
| `internal/lex/token.go` | `TokenType` constants, `Token` struct, `Position` struct |
| `internal/lex/token_test.go` | Tests for token type string representation |

### Types defined

```go
// internal/lex/token.go
package lex

// Position records where a token was found in the source.
type Position struct {
    File   string
    Line   int
    Column int
}

// TokenType classifies a token.
type TokenType int

const (
    // Special
    TOKEN_EOF     TokenType = iota
    TOKEN_ERROR             // lexer error

    // Identifiers
    TOKEN_TERMINAL      // UPPER_CASE identifier (e.g., PLUS, INTEGER)
    TOKEN_NONTERMINAL   // lower_case identifier (e.g., expr, stmt)

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
    TOKEN_DIR_CODE         // %code
    TOKEN_DIR_DEFAULT_TYPE // %default_type
    TOKEN_DIR_EXTRA_ARGUMENT // %extra_argument
    TOKEN_DIR_INCLUDE      // %include
    TOKEN_DIR_LEFT         // %left
    TOKEN_DIR_NAME         // %name
    TOKEN_DIR_NONASSOC     // %nonassoc
    TOKEN_DIR_RIGHT        // %right
    TOKEN_DIR_START_SYMBOL // %start_symbol
    TOKEN_DIR_TOKEN_PREFIX // %token_prefix
    TOKEN_DIR_TOKEN_TYPE   // %token_type
    TOKEN_DIR_TYPE         // %type
    TOKEN_DIR_FALLBACK     // %fallback
    TOKEN_DIR_WILDCARD     // %wildcard
    TOKEN_DIR_DESTRUCTOR   // %destructor
    TOKEN_DIR_SYNTAX_ERROR // %syntax_error
    TOKEN_DIR_PARSE_ACCEPT // %parse_accept
    TOKEN_DIR_PARSE_FAILURE // %parse_failure
    TOKEN_DIR_STACK_OVERFLOW // %stack_overflow

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
}
```

## Algorithm sketch

No algorithm in this lesson — we are defining data types only. The key
invariants are:

1. Every `TokenType` constant has a human-readable `String()` method.
2. `TOKEN_EOF` is always 0 (the zero value signals end-of-input).
3. Terminal identifiers start with an uppercase letter; nonterminal
   identifiers start with a lowercase letter. The lexer (not defined yet)
   will enforce this; the token types just record the classification.

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/lex/token.go` |
| Create | `internal/lex/token_test.go` |

No other files are modified.

## Unit tests (pass/fail gate)

### Test 1: TokenType String coverage

Every `TokenType` constant must return a non-empty, unique string from its
`String()` method. No constant should return `"TokenType(N)"` (the default
for unstringered ints).

```go
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
```

### Test 2: Token zero value is EOF

```go
func TestTokenZeroValueIsEOF(t *testing.T) {
    var tok Token
    if tok.Type != TOKEN_EOF {
        t.Errorf("zero-value Token.Type = %v, want TOKEN_EOF", tok.Type)
    }
}
```

### Test 3: Position formatting

```go
func TestPositionString(t *testing.T) {
    p := Position{File: "calc.y", Line: 10, Column: 5}
    s := p.String() // should return "calc.y:10:5"
    if s != "calc.y:10:5" {
        t.Errorf("Position.String() = %q, want %q", s, "calc.y:10:5")
    }
}
```

## Common mistakes

1. **Forgetting TOKEN_EOF = 0**: If EOF is not the zero value, a default
   `Token{}` will have a confusing type. Always make EOF `iota` (which is 0).

2. **Not distinguishing terminal from nonterminal identifiers**: Both are
   "identifiers" to a generic lexer. You need two token types because the
   parser treats them differently.

3. **Omitting directives**: Lemon has many `%` directives. Missing one means
   the parser can't handle real grammar files. List them all up front.

4. **No `String()` method on TokenType**: Debugging without human-readable
   token names is painful. Implement `String()` using a switch or a map.

5. **Confusing "token type" (our enum) with "terminal symbol" (grammar
   concept)**: `TOKEN_TERMINAL` is the lexer's classification of an
   UPPER_CASE identifier. The actual terminal symbols (PLUS, INTEGER, etc.)
   are grammar-level concepts built later.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-01-token-contract.md and implement
exactly what it describes.

Scope:
- Create internal/lex/token.go with the TokenType constants, Token struct,
  and Position struct as specified in the lesson.
- Implement String() methods on both TokenType and Position.
- Create internal/lex/token_test.go with the three tests described in the
  lesson: TestTokenTypeString, TestTokenZeroValueIsEOF, TestPositionString.
- Use only the Go standard library.
- Keep the code simple, readable, and well-formatted.
- Do NOT implement the lexer yet — that is a later lesson.
- Leave a TODO comment: "// TODO(lesson-02): lexer implementation"

Run `go test ./internal/lex/...` and ensure all tests pass.
```

## Checkpoint questions

1. Why must `TOKEN_EOF` be the zero value of `TokenType`?
2. What is the difference between `TOKEN_TERMINAL` and `TOKEN_NONTERMINAL`
   at the lexer level?
3. Why do we need `TOKEN_CODE_BLOCK` as a separate type instead of just
   treating braces as punctuation?
4. If a grammar file contains `%left PLUS MINUS`, how many tokens would the
   lexer emit for that line?
5. Why does the `Token` struct carry a `Position` instead of just a line number?
