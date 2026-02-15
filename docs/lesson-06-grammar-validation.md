# Lesson 06: Grammar Validation and Error Diagnostics

**Goal**: Validate the parsed grammar for common errors — missing start
symbol, unreachable symbols, unused rules, undefined nonterminals — and
produce structured diagnostics.

## What you will build in this lesson

- Start symbol selection: explicit (`%start_symbol`) or implicit (LHS of
  first rule)
- Augmented grammar creation: add `S' → S $` rule
- Reachability analysis: flag symbols not reachable from the start symbol
- Productivity analysis: flag nonterminals that can never derive a terminal
  string
- Undefined nonterminal detection: a nonterminal used in RHS but never in LHS
- A `Validate(*Grammar) []Diagnostic` function
- A `Finalize(*Grammar) error` function that runs validation and augments
  the grammar

## Key concepts

**Start symbol**: Lemon uses the first rule's LHS as the start symbol unless
`%start_symbol` overrides it. After selection, we create the augmented start
rule `S' → S $`.

**Reachability**: A symbol is reachable if there's a derivation chain from
the start symbol to a rule containing that symbol. Unreachable symbols
indicate dead code in the grammar.

**Productivity**: A nonterminal is productive if it can eventually derive a
string of only terminals. If nonterminal `A` has rules that all reference
other unproductive nonterminals, `A` is unproductive — the grammar can never
complete a parse involving `A`.

**Structured diagnostics**: Each issue is reported as a `Diagnostic` with
position, severity, and a clear message. Validation collects all issues
rather than stopping at the first one.

## Inputs / Outputs

### Files to create / modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/grammar/validate.go` | Create | Validation logic |
| `internal/grammar/validate_test.go` | Create | Validation tests |
| `internal/grammar/grammar.go` | Modify | Add Finalize method, augmented start rule |

### Functions

```go
// Validate checks the grammar for common problems and returns diagnostics.
// It does not modify the grammar.
func Validate(g *Grammar) []Diagnostic

// Finalize selects the start symbol, adds the augmented start rule,
// runs validation, and returns an error if there are any SeverityError
// diagnostics.
func (g *Grammar) Finalize() ([]Diagnostic, error)
```

## Algorithm sketch

**Start symbol selection**:
```
function selectStart(grammar):
    if %start_symbol directive exists:
        sym = lookup directive value
        if sym is not a nonterminal: error
        grammar.Start = sym
    else if len(grammar.Rules) > 0:
        grammar.Start = grammar.Rules[0].LHS
    else:
        error "grammar has no rules"
```

**Augmented start rule**:
```
    augStart = grammar.AddNonterminal("$accept")  // or S'
    grammar.AddRule("$accept", [grammar.Start.Name], "")
    // This becomes rule 0 (shift to front, or track separately)
```

**Reachability** (BFS from start):
```
function findReachable(grammar):
    visited = set{grammar.Start}
    queue = [grammar.Start]
    while queue not empty:
        sym = dequeue
        for each rule where sym == LHS:
            for each s in rule.RHS:
                if s not in visited:
                    visited.add(s)
                    queue.append(s)
    return visited
```

**Productivity** (fixed-point iteration):
```
function findProductive(grammar):
    productive = set of all terminals  // terminals are trivially productive
    changed = true
    while changed:
        changed = false
        for each rule:
            if all symbols in rule.RHS are productive:
                if rule.LHS not in productive:
                    productive.add(rule.LHS)
                    changed = true
    return productive
```
Terminates because: productive set grows monotonically and is bounded by the
number of symbols.

**Undefined nonterminals**:
```
    for each nonterminal nt:
        if no rule has nt as LHS:
            emit error "nonterminal X is used but has no rules"
```

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/grammar/validate.go` |
| Create | `internal/grammar/validate_test.go` |
| Modify | `internal/grammar/grammar.go` — add Finalize, AugmentedStart field |

## Unit tests (pass/fail gate)

### Test 1: Start symbol from first rule

```go
func TestStartSymbolDefault(t *testing.T) {
    src := []byte(`
expr ::= expr PLUS term.
expr ::= term.
term ::= NUM.
`)
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, _ := ParseGrammar(tokens)
    diags, err := g.Finalize()
    if err != nil {
        t.Fatal(err)
    }
    _ = diags
    if g.Start.Name != "expr" {
        t.Errorf("start = %q, want %q", g.Start.Name, "expr")
    }
}
```

### Test 2: Explicit start symbol directive

```go
func TestStartSymbolDirective(t *testing.T) {
    src := []byte(`
%start_symbol program.
program ::= expr.
expr ::= NUM.
`)
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, _ := ParseGrammar(tokens)
    g.Finalize()
    if g.Start.Name != "program" {
        t.Errorf("start = %q, want %q", g.Start.Name, "program")
    }
}
```

### Test 3: Augmented start rule exists

```go
func TestAugmentedStartRule(t *testing.T) {
    src := []byte("expr ::= NUM.")
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, _ := ParseGrammar(tokens)
    g.Finalize()

    // Find the augmented rule
    found := false
    for _, r := range g.Rules {
        if r.LHS.Name == "$accept" {
            found = true
            if len(r.RHS) != 1 || r.RHS[0].Name != "expr" {
                t.Error("augmented rule should be $accept -> expr")
            }
        }
    }
    if !found {
        t.Error("augmented start rule not found")
    }
}
```

### Test 4: Unreachable symbol warning

```go
func TestUnreachableSymbol(t *testing.T) {
    src := []byte(`
expr ::= NUM.
orphan ::= THING.
`)
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, _ := ParseGrammar(tokens)
    diags, _ := g.Finalize()

    found := false
    for _, d := range diags {
        if strings.Contains(d.Message, "orphan") &&
           strings.Contains(d.Message, "unreachable") {
            found = true
        }
    }
    if !found {
        t.Error("expected warning about unreachable 'orphan'")
    }
}
```

### Test 5: Undefined nonterminal error

```go
func TestUndefinedNonterminal(t *testing.T) {
    src := []byte("expr ::= expr PLUS missing_nt.")
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, _ := ParseGrammar(tokens)
    diags, err := g.Finalize()

    if err == nil {
        t.Error("expected error for undefined nonterminal")
    }
    found := false
    for _, d := range diags {
        if d.Severity == SeverityError &&
           strings.Contains(d.Message, "missing_nt") {
            found = true
        }
    }
    if !found {
        t.Error("expected diagnostic about missing_nt")
    }
}
```

### Test 6: Unproductive nonterminal warning

```go
func TestUnproductiveNonterminal(t *testing.T) {
    // 'loop' can never produce a terminal string
    src := []byte(`
expr ::= NUM.
expr ::= loop.
loop ::= loop.
`)
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, _ := ParseGrammar(tokens)
    diags, _ := g.Finalize()

    found := false
    for _, d := range diags {
        if strings.Contains(d.Message, "loop") &&
           strings.Contains(d.Message, "unproductive") {
            found = true
        }
    }
    if !found {
        t.Error("expected warning about unproductive 'loop'")
    }
}
```

## Common mistakes

1. **Forgetting to add the augmented start rule**: The LALR construction
   (Lessons 10–14) depends on `$accept → S`. Without it, the algorithm
   has no accept state.

2. **Making unreachable symbols an error instead of a warning**: Unreachable
   symbols are suspicious but not fatal. Use `SeverityWarning`.

3. **Confusing undefined with unreachable**: A nonterminal with no rules is
   *undefined* (error). A nonterminal that has rules but can't be reached
   from the start is *unreachable* (warning).

4. **Not iterating productivity to a fixed point**: One pass isn't enough.
   A nonterminal might become productive only after another one does.

5. **Modifying the grammar during validation**: `Validate` should be
   read-only. `Finalize` handles modification (adding the augmented rule)
   before calling `Validate`.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-06-grammar-validation.md and
implement exactly what it describes.

Scope:
- Create internal/grammar/validate.go with Validate function implementing:
  start symbol selection, reachability analysis (BFS), productivity analysis
  (fixed-point), and undefined nonterminal detection.
- Add a Finalize method to Grammar that adds the augmented start rule
  ($accept -> StartSymbol), runs Validate, and returns diagnostics + error.
- Create internal/grammar/validate_test.go with all six tests.
- Unreachable symbols are warnings; undefined nonterminals are errors;
  unproductive nonterminals are warnings.
- Use only the Go standard library.
- Keep code explicit and clear.

Run `go test ./internal/grammar/...` and ensure all tests pass.
```

## Checkpoint questions

1. Why does the LALR algorithm require an augmented start rule?
2. What is the difference between an unreachable symbol and an undefined
   nonterminal?
3. Why does the productivity algorithm use fixed-point iteration?
4. Should an empty grammar (no rules) be an error or just a warning?
5. When `Finalize` adds the augmented rule, should it go at the beginning
   or end of the rule list? Does it matter?
