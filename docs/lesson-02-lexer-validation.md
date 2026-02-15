# Lesson 02: Validating the Lexer Token Stream

**Goal**: Adapt the provided Lemon grammar lexer and write comprehensive tests
that validate it produces the exact token stream the parser will expect.

## What you will build in this lesson

- Integrate the provided lexer into `internal/lex/lexer.go`
- A `Tokenize(filename string, src []byte) ([]Token, error)` function
- Detailed unit tests that validate tokenization of representative Lemon
  grammar snippets
- Ensure the lexer correctly distinguishes terminals, nonterminals, directives,
  code blocks, and punctuation

## Key concepts

**Lexer adaptation**: The lexer will be provided as existing Go code. Your job
is to ensure it produces tokens matching the `TokenType` constants from Lesson
01. If the existing lexer uses different type names or categories, you will map
them to the Lesson 01 contract.

**Token stream validation**: Before building the parser, we need confidence
that the lexer emits the right tokens in the right order. We test this by
feeding known grammar snippets and comparing the output token sequence.

**Code blocks are opaque**: When the lexer encounters `{ ... }`, it should
emit a single `TOKEN_CODE_BLOCK` whose literal contains everything between
(and including) the braces. The lexer must handle nested braces within code
blocks.

**Comments are discarded**: Lemon-style comments (`/* ... */` and `// ...`)
should be consumed by the lexer and not emitted as tokens.

## Inputs / Outputs

### Files to create / modify

| File                         | Action | Purpose                          |
|------------------------------|--------|----------------------------------|
| `internal/lex/lexer.go`      | Create | Lexer implementation             |
| `internal/lex/lexer_test.go` | Create | Comprehensive token stream tests |

### Function signatures

```go
// Tokenize scans the source and returns all tokens including a final TOKEN_EOF.
// The filename is used only for Position fields in the returned tokens.
func Tokenize(filename string, src []byte) ([]Token, error)
```

## Algorithm sketch

The lexer is a single-pass scanner over the byte slice:

```
function Tokenize(filename, src):
    tokens = []
    pos = 0
    line = 1, col = 1
    while pos < len(src):
        skip whitespace and comments (update line/col)
        if at end: break
        ch = src[pos]
        switch:
            case ch == '%':  scan directive keyword
            case ch == '{':  scan code block (track nesting depth)
            case ch == ':' and peek '::=':  emit TOKEN_COLONCOLON_EQ
            case ch == '.':  emit TOKEN_DOT
            case ch == '|':  emit TOKEN_PIPE
            case ch == '(' / ')' / '[' / ']' / ',':  emit punctuation
            case ch == '"':  scan quoted string
            case isUpper(ch):  scan identifier → TOKEN_TERMINAL
            case isLower(ch):  scan identifier → TOKEN_NONTERMINAL
            default: emit TOKEN_ERROR
        append token to tokens
    append TOKEN_EOF
    return tokens
```

**Invariant**: The returned slice always ends with exactly one `TOKEN_EOF`.

**Termination**: The loop advances `pos` by at least 1 on each iteration
(whitespace skip or token scan). Since `src` is finite, the loop terminates.

## Repository changes

| Action | File                         |
|--------|------------------------------|
| Create | `internal/lex/lexer.go`      |
| Create | `internal/lex/lexer_test.go` |

## Unit tests (pass/fail gate)

### Test 1: Simple rule tokenization

Input:
```
expr ::= expr PLUS term.
```

Expected tokens (types only):
```
TOKEN_NONTERMINAL  ("expr")
TOKEN_COLONCOLON_EQ ("::=")
TOKEN_NONTERMINAL  ("expr")
TOKEN_TERMINAL     ("PLUS")
TOKEN_NONTERMINAL  ("term")
TOKEN_DOT          (".")
TOKEN_EOF
```

### Test 2: Rule with action

Input:
```
expr(A) ::= expr(B) PLUS term(C). { A = B + C; }
```

Expected tokens:
```
TOKEN_NONTERMINAL ("expr")
TOKEN_LPAREN      ("(")
TOKEN_TERMINAL    ("A")  -- or TOKEN_NONTERMINAL depending on case
TOKEN_RPAREN      (")")
TOKEN_COLONCOLON_EQ ("::=")
TOKEN_NONTERMINAL ("expr")
TOKEN_LPAREN      ("(")
TOKEN_TERMINAL    ("B")
TOKEN_RPAREN      (")")
TOKEN_TERMINAL    ("PLUS")
TOKEN_NONTERMINAL ("term")
TOKEN_LPAREN      ("(")
TOKEN_TERMINAL    ("C")
TOKEN_RPAREN      (")")
TOKEN_DOT         (".")
TOKEN_CODE_BLOCK  ("{ A = B + C; }")
TOKEN_EOF
```

Note: In Lemon, single uppercase letters used as aliases in rules (like `A`,
`B`, `C`) are still identifiers. The lexer should classify them as
`TOKEN_TERMINAL` since they start with uppercase. The grammar parser (Lesson
04) will interpret context.

### Test 3: Directive tokenization

Input:
```
%left PLUS MINUS.
%left TIMES DIVIDE.
%token_type { int }
```

Expected:
```
TOKEN_DIR_LEFT      ("%left")
TOKEN_TERMINAL      ("PLUS")
TOKEN_TERMINAL      ("MINUS")
TOKEN_DOT           (".")
TOKEN_DIR_LEFT      ("%left")
TOKEN_TERMINAL      ("TIMES")
TOKEN_TERMINAL      ("DIVIDE")
TOKEN_DOT           (".")
TOKEN_DIR_TOKEN_TYPE ("%token_type")
TOKEN_CODE_BLOCK    ("{ int }")
TOKEN_EOF
```

### Test 4: Comments are skipped

Input:
```
// This is a comment
expr ::= term. /* another comment */
```

Expected:
```
TOKEN_NONTERMINAL  ("expr")
TOKEN_COLONCOLON_EQ ("::=")
TOKEN_NONTERMINAL  ("term")
TOKEN_DOT          (".")
TOKEN_EOF
```

### Test 5: Nested braces in code block

Input:
```
expr ::= IDENT. { if (x > 0) { y = x; } }
```

Expected: The code block token's literal should be `{ if (x > 0) { y = x; } }`.

### Test 6: Empty input

Input: `""`

Expected: `TOKEN_EOF` only.

### Test 7: Position tracking

Input:
```
expr ::= term.
factor ::= IDENT.
```

The first token (`expr`) should have Line=1, Column=1.
The `factor` token on the second line should have Line=2, Column=1.

### Implementation hint for tests

```go
func TestSimpleRule(t *testing.T) {
    src := []byte("expr ::= expr PLUS term.")
    tokens, err := Tokenize("test.y", src)
    if err != nil {
        t.Fatalf("Tokenize error: %v", err)
    }
    expected := []TokenType{
        TOKEN_NONTERMINAL, TOKEN_COLONCOLON_EQ,
        TOKEN_NONTERMINAL, TOKEN_TERMINAL, TOKEN_NONTERMINAL,
        TOKEN_DOT, TOKEN_EOF,
    }
    if len(tokens) != len(expected) {
        t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
    }
    for i, want := range expected {
        if tokens[i].Type != want {
            t.Errorf("token[%d].Type = %v, want %v (literal=%q)",
                i, tokens[i].Type, want, tokens[i].Literal)
        }
    }
}
```

## Common mistakes

1. **Not handling nested braces in code blocks**: `{ if (x) { y; } }` must
   be a single `TOKEN_CODE_BLOCK`. Track brace depth.

2. **Classifying aliases wrong**: In `expr(A)`, `A` is an identifier starting
   with uppercase, so the lexer classifies it as `TOKEN_TERMINAL`. Context
   interpretation happens in the grammar parser, not the lexer.

3. **Forgetting the trailing EOF**: Every token stream must end with
   `TOKEN_EOF`. Check this in every test.

4. **Wrong position tracking after multi-line code blocks**: If a code block
   spans multiple lines, the position for the next token must account for
   all the newlines inside the block.

5. **Not skipping `\r\n` as a single newline on Windows**: If you want
   portability, handle `\r\n` and `\r` as line terminators.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-02-lexer-validation.md and
implement exactly what it describes.

Scope:
- Create internal/lex/lexer.go with the Tokenize function.
- The lexer must handle: identifiers (terminal=uppercase, nonterminal=lowercase),
  directives (%left, %right, %nonassoc, %token_type, %type, %start_symbol,
  %name, %include, %code, %default_type, %extra_argument, %token_prefix,
  %fallback, %wildcard, %destructor, %syntax_error, %parse_accept,
  %parse_failure, %stack_overflow), code blocks with nested braces,
  punctuation (::=, ., |, (, ), [, ], ,), quoted strings, and comments
  (// and /* */).
- Create internal/lex/lexer_test.go with all seven tests described in the
  lesson: simple rule, rule with action, directive tokenization, comments
  skipped, nested braces, empty input, and position tracking.
- Use only the Go standard library.
- The returned token slice must always end with TOKEN_EOF.
- Keep the code readable and explicit.

Run `go test ./internal/lex/...` and ensure all tests pass.
```

## Checkpoint questions

1. Why does the token stream contract require `TOKEN_EOF` as the last token?
2. How does the lexer determine whether an identifier is a terminal or
   nonterminal?
3. What happens if a code block `{ ... }` contains string literals with
   braces inside them (e.g., `"}"`)? Is this a concern for our lexer?
4. Why do we test position tracking? When will positions matter later?
5. If the input contains `%unknown_directive`, what token should the lexer emit?
