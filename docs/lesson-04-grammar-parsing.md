# Lesson 04: Parsing Lemon Grammar Files

**Goal**: Build a parser that consumes the token stream from Lesson 02 and
constructs a `Grammar` (from Lesson 03), handling rules and basic directives.

## What you will build in this lesson

- A `ParseGrammar(tokens []lex.Token) (*Grammar, []Diagnostic, error)` function
- Parsing of production rules: `lhs ::= rhs_symbol ... . { action }`
- Parsing of rule aliases: `lhs(A) ::= rhs(B) rhs(C) . { ... }`
- Handling of unknown or unrecognized tokens (emit diagnostic, skip gracefully)
- A `Diagnostic` type for structured error/warning messages

## Key concepts

**Recursive descent over a flat token stream**: Even though we're building a
parser generator, the parser for the `.y` grammar file itself is a simple
hand-written recursive descent parser. It reads tokens sequentially and
builds up the Grammar model.

**Rule structure in Lemon**:
```
lhs_symbol(ALIAS) ::= rhs_1(alias1) rhs_2(alias2) ... . { code }
```
- The LHS is always a nonterminal.
- Each RHS symbol may have a parenthesized alias.
- The rule ends with a `.` (dot).
- An optional code block follows the dot.

**Directives are deferred**: This lesson handles rules and recognizes
directives but does NOT fully process precedence or other directives.
It records them for later processing. We parse `%left`, `%right`, etc. into
a list of raw directive records.

**Diagnostics, not panics**: When the parser encounters something unexpected,
it emits a `Diagnostic` (with position, severity, message) and tries to
recover by skipping to the next `.` or directive.

## Inputs / Outputs

### Files to create / modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/grammar/diagnostic.go` | Create | Diagnostic type |
| `internal/grammar/parser.go` | Create | Grammar file parser |
| `internal/grammar/parser_test.go` | Create | Parser tests |

### Types and functions

```go
// internal/grammar/diagnostic.go

type Severity int

const (
    SeverityWarning Severity = iota
    SeverityError
)

type Diagnostic struct {
    Pos      lex.Position
    Severity Severity
    Message  string
}

func (d Diagnostic) String() string
```

```go
// internal/grammar/parser.go

// ParseGrammar reads a token stream and builds a Grammar.
// It returns the grammar, any diagnostics collected, and a fatal error
// if the input is completely unparseable.
func ParseGrammar(tokens []lex.Token) (*Grammar, []Diagnostic, error)
```

## Algorithm sketch

```
function ParseGrammar(tokens):
    grammar = NewGrammar()
    diagnostics = []
    pos = 0  // index into tokens

    while tokens[pos].Type != TOKEN_EOF:
        tok = tokens[pos]

        if tok is a directive token (%left, %right, etc.):
            record = parseDirective(tokens, &pos, &diagnostics)
            store record in grammar.rawDirectives
            continue

        if tok.Type == TOKEN_NONTERMINAL:
            parseRule(tokens, &pos, grammar, &diagnostics)
            continue

        // unexpected token — emit diagnostic and skip
        emit Diagnostic(warning, "unexpected token", tok.Pos)
        pos++

    return grammar, diagnostics, nil

function parseRule(tokens, pos, grammar, diagnostics):
    lhsName = tokens[*pos].Literal
    *pos++
    lhsAlias = ""
    if tokens[*pos].Type == TOKEN_LPAREN:
        *pos++  // skip (
        lhsAlias = tokens[*pos].Literal
        *pos++  // skip alias
        expect TOKEN_RPAREN, *pos++

    expect TOKEN_COLONCOLON_EQ, *pos++

    rhs = []
    aliases = []
    while tokens[*pos].Type != TOKEN_DOT and != TOKEN_EOF:
        symName = tokens[*pos].Literal
        *pos++
        alias = ""
        if tokens[*pos].Type == TOKEN_LPAREN:
            *pos++
            alias = tokens[*pos].Literal
            *pos++
            expect TOKEN_RPAREN, *pos++
        rhs.append(symName)
        aliases.append(alias)

    expect TOKEN_DOT, *pos++

    action = ""
    if tokens[*pos].Type == TOKEN_CODE_BLOCK:
        action = tokens[*pos].Literal
        *pos++

    // Ensure symbols exist
    grammar.AddNonterminal(lhsName)
    for each sym in rhs:
        if isUpper(sym[0]):
            grammar.AddTerminal(sym)
        else:
            grammar.AddNonterminal(sym)

    rule, err = grammar.AddRule(lhsName, rhs, action)
    if err: emit diagnostic

    // Store aliases on the rule for code generation
    rule.LHSAlias = lhsAlias
    rule.RHSAliases = aliases
```

**Invariants**:
- Every rule is terminated by `.` (TOKEN_DOT).
- If `.` is missing, emit a diagnostic and skip to the next dot or directive.
- The parser never panics.
- All symbols referenced in rules are auto-registered in the symbol table.

**Termination**: The outer loop advances `pos` on every iteration (either by
parsing a rule/directive or by skipping a token). Since the token list is
finite and ends with TOKEN_EOF, the loop terminates.

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/grammar/diagnostic.go` |
| Create | `internal/grammar/parser.go` |
| Create | `internal/grammar/parser_test.go` |
| Modify | `internal/grammar/rule.go` — add `LHSAlias string` and `RHSAliases []string` fields |

## Unit tests (pass/fail gate)

### Test 1: Parse a single rule

```go
func TestParseSingleRule(t *testing.T) {
    src := []byte("expr ::= expr PLUS term.")
    tokens, _ := lex.Tokenize("test.y", src)
    g, diags, err := ParseGrammar(tokens)
    if err != nil {
        t.Fatal(err)
    }
    if len(diags) > 0 {
        t.Errorf("unexpected diagnostics: %v", diags)
    }
    if len(g.Rules) != 1 {
        t.Fatalf("got %d rules, want 1", len(g.Rules))
    }
    r := g.Rules[0]
    if r.LHS.Name != "expr" {
        t.Errorf("LHS = %q, want %q", r.LHS.Name, "expr")
    }
    if len(r.RHS) != 3 {
        t.Errorf("RHS len = %d, want 3", len(r.RHS))
    }
}
```

### Test 2: Parse multiple rules

```go
func TestParseMultipleRules(t *testing.T) {
    src := []byte(`
expr ::= expr PLUS term.
expr ::= term.
term ::= NUM.
`)
    tokens, _ := lex.Tokenize("test.y", src)
    g, diags, err := ParseGrammar(tokens)
    if err != nil {
        t.Fatal(err)
    }
    if len(diags) > 0 {
        t.Errorf("unexpected diagnostics: %v", diags)
    }
    if len(g.Rules) != 3 {
        t.Fatalf("got %d rules, want 3", len(g.Rules))
    }
}
```

### Test 3: Parse rule with aliases

```go
func TestParseRuleWithAliases(t *testing.T) {
    src := []byte("expr(A) ::= expr(B) PLUS term(C). { A = B + C; }")
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, err := ParseGrammar(tokens)
    if err != nil {
        t.Fatal(err)
    }
    r := g.Rules[0]
    if r.LHSAlias != "A" {
        t.Errorf("LHSAlias = %q, want %q", r.LHSAlias, "A")
    }
    if r.Action == "" {
        t.Error("expected non-empty action")
    }
}
```

### Test 4: Directive is recognized but deferred

```go
func TestDirectiveRecognized(t *testing.T) {
    src := []byte(`
%left PLUS MINUS.
expr ::= term.
term ::= NUM.
`)
    tokens, _ := lex.Tokenize("test.y", src)
    g, diags, err := ParseGrammar(tokens)
    if err != nil {
        t.Fatal(err)
    }
    // No error diagnostics for recognized directives
    for _, d := range diags {
        if d.Severity == SeverityError {
            t.Errorf("unexpected error: %v", d)
        }
    }
    if len(g.Rules) != 2 {
        t.Fatalf("got %d rules, want 2", len(g.Rules))
    }
}
```

### Test 5: Auto-creation of symbols

```go
func TestAutoCreateSymbols(t *testing.T) {
    src := []byte("expr ::= expr PLUS term.")
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, _ := ParseGrammar(tokens)

    _, ok := g.Symbols.Lookup("PLUS")
    if !ok {
        t.Error("PLUS should be auto-created as terminal")
    }
    s, ok := g.Symbols.Lookup("expr")
    if !ok {
        t.Error("expr should be auto-created as nonterminal")
    }
    if s.Kind != SymbolNonterminal {
        t.Error("expr should be nonterminal")
    }
}
```

### Test 6: Missing dot emits diagnostic

```go
func TestMissingDotDiagnostic(t *testing.T) {
    src := []byte("expr ::= term")  // no dot!
    tokens, _ := lex.Tokenize("test.y", src)
    _, diags, _ := ParseGrammar(tokens)
    found := false
    for _, d := range diags {
        if d.Severity == SeverityError {
            found = true
        }
    }
    if !found {
        t.Error("expected error diagnostic for missing dot")
    }
}
```

## Common mistakes

1. **Not auto-registering symbols**: The parser must automatically call
   `AddTerminal` / `AddNonterminal` when it encounters symbol names in
   rules. Don't require symbols to be pre-declared.

2. **Confusing alias identifiers with symbols**: In `expr(A)`, `A` is an
   alias, not a symbol in the grammar. Don't add it to the symbol table.

3. **Crashing on malformed input**: Use diagnostics and skip-to-dot recovery.
   Never panic.

4. **Ignoring the code block after the dot**: The action code is optional
   but must be captured when present.

5. **Not storing rule order**: Rule indices matter for conflict resolution
   later. Rules must be stored in the order they appear.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-04-grammar-parsing.md and
implement exactly what it describes.

Scope:
- Create internal/grammar/diagnostic.go with Diagnostic type.
- Create internal/grammar/parser.go with ParseGrammar function.
- Modify internal/grammar/rule.go to add LHSAlias and RHSAliases fields.
- Create internal/grammar/parser_test.go with all six tests.
- The parser must: handle rules with ::= and . terminators, support
  parenthesized aliases, auto-register symbols, recognize directives
  (store raw directive data but do not fully process yet), emit Diagnostic
  structs on errors, and recover by skipping to next dot/directive.
- Use only the Go standard library.
- Keep code readable. Prefer explicit error handling over clever tricks.
- Leave TODO: "// TODO(lesson-05): process precedence directives"

Run `go test ./internal/grammar/...` and ensure all tests pass.
```

## Checkpoint questions

1. Why is a hand-written recursive descent parser appropriate for the grammar
   file, even though we're building a parser generator?
2. What is the recovery strategy when the parser encounters an unexpected token?
3. Why do we store aliases separately from the symbol list?
4. What happens if the same nonterminal appears as LHS in multiple rules?
5. Why must rule indices be sequential and deterministic?
