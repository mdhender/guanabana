# Lesson 07: The Lemon Grammar of Lemon

**Goal**: Write a meaningful subset of the Lemon grammar format *as a Lemon
grammar file*, parse it through the full pipeline (lexer → parser →
validation), and confirm the lexer→parser integration is correct.

## What you will build in this lesson

- A Lemon grammar file (`testdata/lemon.y`) that describes the syntax of
  Lemon grammar files themselves (a self-describing subset)
- An integration test that tokenizes, parses, and validates this self-grammar
- Verification that the token stream matches what the grammar parser expects
- Confidence that the lexer and parser are correctly integrated

## Key concepts

**Self-describing grammar**: Writing the grammar of the grammar file format in
its own notation is a powerful validation technique. If the tools can process
the grammar that describes themselves, it's strong evidence that the pipeline
works correctly. We don't need full self-hosting — just enough of the Lemon
syntax to exercise every token type and directive.

**Integration testing**: Lessons 01–06 tested individual components. This
lesson tests the full pipeline: bytes → tokens → grammar model → validation.
Bugs in the contract between components surface here.

**The subset we model**: Our self-grammar covers:
- Grammar rules with `::=` and `.` terminators
- Parenthesized aliases
- Code block actions
- Precedence directives (`%left`, `%right`, `%nonassoc`)
- `%token_type` and `%start_symbol` directives
- Comments

## Inputs / Outputs

### Files to create

| File | Action | Purpose |
|------|--------|---------|
| `testdata/lemon.y` | Create | Self-grammar file |
| `internal/grammar/integration_test.go` | Create | Integration tests |

### The self-grammar

The grammar below describes a simplified Lemon grammar file format:

```
// testdata/lemon.y
// A Lemon grammar that describes the Lemon grammar file format (subset).

%token_type { Token }
%start_symbol grammar_file.

%left DIR_LEFT DIR_RIGHT DIR_NONASSOC.

grammar_file ::= item_list.

item_list ::= item_list item.
item_list ::= .

item ::= rule.
item ::= directive.

rule ::= NONTERMINAL opt_alias COLONCOLON_EQ rhs_list DOT opt_code_block.

opt_alias ::= LPAREN TERMINAL RPAREN.
opt_alias ::= .

rhs_list ::= rhs_list rhs_symbol.
rhs_list ::= .

rhs_symbol ::= TERMINAL opt_alias.
rhs_symbol ::= NONTERMINAL opt_alias.

opt_code_block ::= CODE_BLOCK.
opt_code_block ::= .

directive ::= prec_directive.
directive ::= token_type_directive.
directive ::= start_directive.

prec_directive ::= DIR_LEFT terminal_list DOT.
prec_directive ::= DIR_RIGHT terminal_list DOT.
prec_directive ::= DIR_NONASSOC terminal_list DOT.

terminal_list ::= terminal_list TERMINAL.
terminal_list ::= TERMINAL.

token_type_directive ::= DIR_TOKEN_TYPE CODE_BLOCK.

start_directive ::= DIR_START_SYMBOL NONTERMINAL DOT.
```

## Algorithm sketch

No new algorithms. This lesson exercises existing code:

```
function testSelfGrammar():
    src = readFile("testdata/lemon.y")

    // Step 1: tokenize
    tokens = lex.Tokenize("lemon.y", src)
    verify no TOKEN_ERROR in tokens
    verify TOKEN_EOF at end

    // Step 2: parse
    grammar, diagnostics, err = grammar.ParseGrammar(tokens)
    verify err == nil
    verify no SeverityError diagnostics

    // Step 3: validate
    diags, err = grammar.Finalize()
    verify err == nil

    // Step 4: check grammar properties
    verify grammar has expected number of nonterminals
    verify grammar has expected number of terminals
    verify grammar has expected number of rules
    verify start symbol == "grammar_file"
```

## Repository changes

| Action | File |
|--------|------|
| Create | `testdata/lemon.y` |
| Create | `internal/grammar/integration_test.go` |

## Unit tests (pass/fail gate)

### Test 1: Self-grammar tokenizes without errors

```go
func TestSelfGrammarTokenizes(t *testing.T) {
    src, err := os.ReadFile("../../testdata/lemon.y")
    if err != nil {
        t.Fatal(err)
    }
    tokens, err := lex.Tokenize("lemon.y", src)
    if err != nil {
        t.Fatal(err)
    }
    for _, tok := range tokens {
        if tok.Type == lex.TOKEN_ERROR {
            t.Errorf("lexer error at %v: %q", tok.Pos, tok.Literal)
        }
    }
    last := tokens[len(tokens)-1]
    if last.Type != lex.TOKEN_EOF {
        t.Error("last token should be EOF")
    }
}
```

### Test 2: Self-grammar parses without errors

```go
func TestSelfGrammarParses(t *testing.T) {
    src, _ := os.ReadFile("../../testdata/lemon.y")
    tokens, _ := lex.Tokenize("lemon.y", src)
    g, diags, err := ParseGrammar(tokens)
    if err != nil {
        t.Fatalf("parse error: %v", err)
    }
    for _, d := range diags {
        if d.Severity == SeverityError {
            t.Errorf("error diagnostic: %v", d)
        }
    }
    if len(g.Rules) == 0 {
        t.Error("expected rules, got none")
    }
}
```

### Test 3: Self-grammar validates

```go
func TestSelfGrammarValidates(t *testing.T) {
    src, _ := os.ReadFile("../../testdata/lemon.y")
    tokens, _ := lex.Tokenize("lemon.y", src)
    g, _, _ := ParseGrammar(tokens)
    diags, err := g.Finalize()
    if err != nil {
        t.Fatalf("validation error: %v", err)
    }
    for _, d := range diags {
        if d.Severity == SeverityError {
            t.Errorf("validation error: %v", d)
        }
    }
}
```

### Test 4: Self-grammar has expected structure

```go
func TestSelfGrammarStructure(t *testing.T) {
    src, _ := os.ReadFile("../../testdata/lemon.y")
    tokens, _ := lex.Tokenize("lemon.y", src)
    g, _, _ := ParseGrammar(tokens)
    g.Finalize()

    if g.Start.Name != "grammar_file" {
        t.Errorf("start = %q, want %q", g.Start.Name, "grammar_file")
    }

    // Count nonterminals (excluding $accept)
    ntCount := 0
    for _, s := range g.Symbols.Nonterminals() {
        if s.Name != "$accept" {
            ntCount++
        }
    }
    // Expected nonterminals: grammar_file, item_list, item, rule,
    // opt_alias, rhs_list, rhs_symbol, opt_code_block, directive,
    // prec_directive, token_type_directive, start_directive, terminal_list
    if ntCount < 10 {
        t.Errorf("nonterminal count = %d, want >= 10", ntCount)
    }
}
```

### Test 5: Token stream snapshot for a known rule

```go
func TestTokenStreamForRule(t *testing.T) {
    // Verify exact token sequence for a single well-known line
    src := []byte("rule ::= NONTERMINAL opt_alias COLONCOLON_EQ rhs_list DOT opt_code_block.")
    tokens, _ := lex.Tokenize("test.y", src)

    expected := []lex.TokenType{
        lex.TOKEN_NONTERMINAL,   // "rule"
        lex.TOKEN_COLONCOLON_EQ, // "::="
        lex.TOKEN_TERMINAL,      // "NONTERMINAL"
        lex.TOKEN_NONTERMINAL,   // "opt_alias"
        lex.TOKEN_TERMINAL,      // "COLONCOLON_EQ"
        lex.TOKEN_NONTERMINAL,   // "rhs_list"
        lex.TOKEN_TERMINAL,      // "DOT"
        lex.TOKEN_NONTERMINAL,   // "opt_code_block"
        lex.TOKEN_DOT,           // "."
        lex.TOKEN_EOF,
    }

    if len(tokens) != len(expected) {
        t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
    }
    for i, want := range expected {
        if tokens[i].Type != want {
            t.Errorf("token[%d] = %v (%q), want %v",
                i, tokens[i].Type, tokens[i].Literal, want)
        }
    }
}
```

## Common mistakes

1. **Using lowercase names for terminals in the self-grammar**: In the
   self-grammar, the terminal symbols are things like `NONTERMINAL`,
   `COLONCOLON_EQ`, `DOT` — these are the token type names, and they must
   be uppercase for the lexer to classify them as terminals.

2. **Not handling the self-grammar's comments**: The self-grammar file has
   `//` comments. Make sure the lexer skips them.

3. **Wrong file path in tests**: The testdata directory is at the repo root.
   From `internal/grammar/`, the relative path is `../../testdata/lemon.y`.

4. **Expecting exact rule/symbol counts**: The grammar may evolve. Test for
   minimum counts or structural properties rather than exact numbers.

5. **Circular validation**: The self-grammar is tested by our parser, which
   doesn't yet use the generated parser. This is not truly self-hosting — it's
   a validation exercise. Don't confuse the two.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-07-self-grammar.md and implement
exactly what it describes.

Scope:
- Create testdata/lemon.y with the self-grammar file shown in the lesson.
  The grammar should describe a meaningful subset of the Lemon grammar format
  using Lemon syntax.
- Create internal/grammar/integration_test.go with all five tests:
  tokenization, parsing, validation, structure check, and token stream
  snapshot.
- The tests must read testdata/lemon.y and run the full pipeline.
- Use only the Go standard library.
- Do NOT modify any existing source files — only create new files.

Run `go test ./internal/grammar/...` and ensure all tests pass.
```

## Checkpoint questions

1. Why is writing a self-describing grammar a useful validation technique?
2. Does our self-grammar need to be *complete* (describe every Lemon feature)?
   Why or why not?
3. In the self-grammar, `NONTERMINAL` is a terminal symbol. Why isn't this
   confusing to the parser?
4. What kinds of bugs would the integration test catch that unit tests
   from earlier lessons would miss?
5. Could we use the parser generator we're building to parse its own grammar
   files? At what point in the curriculum would that become possible?
