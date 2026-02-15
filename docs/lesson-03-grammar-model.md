# Lesson 03: Grammar Model — Symbols and Productions

**Goal**: Define the in-memory representation of a context-free grammar:
symbols (terminals and nonterminals), production rules, and the Grammar
container that holds them all.

## What you will build in this lesson

- A `Symbol` type that represents either a terminal or nonterminal
- A `SymbolTable` that assigns stable integer IDs to symbols
- A `Rule` struct representing a production (LHS → RHS₁ RHS₂ … RHSₙ)
- A `Grammar` struct that holds the symbol table, rules, and metadata
- Builder methods to add symbols and rules programmatically (for testing
  without needing the lexer/parser yet)

## Key concepts

**Context-free grammar (CFG)**: A set of rules of the form `A → α` where `A`
is a nonterminal and `α` is a sequence of zero or more symbols (terminals and
nonterminals). A terminal is a symbol that appears literally in the input
(like `PLUS` or `INTEGER`). A nonterminal is a symbol that can be expanded
by applying a rule.

**Symbol IDs**: Each symbol gets a unique integer ID. This lets us use arrays
and bitsets later (for FIRST/FOLLOW) instead of maps with string keys.
Convention: terminal IDs start at 1, nonterminal IDs start after the last
terminal. Symbol ID 0 is reserved for `$` (the end-of-input marker).

**Productions**: Each rule has a left-hand side (a single nonterminal) and a
right-hand side (a list of symbol IDs). Rules also carry an index (their
position in the grammar's rule list) and an optional code action (a string
of Go code to execute on reduce).

**Augmented grammar**: LALR parsers need a start rule `S' → S $`. We add
this automatically when the grammar is finalized. The learner does not write
it in the grammar file.

## Inputs / Outputs

### Files to create

| File | Purpose |
|------|---------|
| `internal/grammar/symbol.go` | Symbol, SymbolKind, SymbolTable |
| `internal/grammar/rule.go` | Rule struct |
| `internal/grammar/grammar.go` | Grammar struct with builder methods |
| `internal/grammar/grammar_test.go` | Tests for model construction |

### Types defined

```go
// internal/grammar/symbol.go

type SymbolKind int

const (
    SymbolTerminal    SymbolKind = iota
    SymbolNonterminal
)

type Symbol struct {
    ID   int
    Name string
    Kind SymbolKind
}

type SymbolTable struct {
    symbols []*Symbol         // indexed by ID
    byName  map[string]*Symbol
}

// Methods:
// AddTerminal(name string) *Symbol
// AddNonterminal(name string) *Symbol
// Lookup(name string) (*Symbol, bool)
// Terminal(id int) *Symbol
// Nonterminal(id int) *Symbol
// Terminals() []*Symbol
// Nonterminals() []*Symbol
// NumSymbols() int
```

```go
// internal/grammar/rule.go

type Rule struct {
    Index  int       // position in the grammar's rule list
    LHS    *Symbol   // left-hand side nonterminal
    RHS    []*Symbol // right-hand side symbols (may be empty for ε-rules)
    Action string    // Go code for the reduce action (may be empty)
}
```

```go
// internal/grammar/grammar.go

type Grammar struct {
    Symbols *SymbolTable
    Rules   []*Rule
    Start   *Symbol // set during finalization
    EOF     *Symbol // the $ marker, always ID 0
}

// Methods:
// NewGrammar() *Grammar
// AddTerminal(name string) *Symbol
// AddNonterminal(name string) *Symbol
// AddRule(lhsName string, rhsNames []string, action string) (*Rule, error)
// RulesFor(nt *Symbol) []*Rule   // all rules with nt as LHS
```

## Algorithm sketch

No complex algorithm. The key operations are:

**AddTerminal/AddNonterminal**:
1. If name already exists in `byName`, return existing symbol (and verify
   kind matches; error if mismatch).
2. Allocate new ID (sequential), create Symbol, add to `symbols` and `byName`.

**AddRule**:
1. Look up LHS symbol; it must be a nonterminal.
2. Look up each RHS symbol; they must exist in the symbol table.
3. Create a `Rule` with the next sequential index.
4. Append to `Grammar.Rules`.

**Invariants**:
- Symbol ID 0 is always the EOF marker `$`.
- Every rule's LHS is a nonterminal.
- No duplicate symbol names.
- Rule indices are 0, 1, 2, … (sequential, no gaps).

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/grammar/symbol.go` |
| Create | `internal/grammar/rule.go` |
| Create | `internal/grammar/grammar.go` |
| Create | `internal/grammar/grammar_test.go` |

## Unit tests (pass/fail gate)

### Test 1: Symbol table basics

```go
func TestSymbolTable(t *testing.T) {
    g := NewGrammar()
    plus := g.AddTerminal("PLUS")
    minus := g.AddTerminal("MINUS")
    expr := g.AddNonterminal("expr")

    if plus.Kind != SymbolTerminal {
        t.Error("PLUS should be terminal")
    }
    if expr.Kind != SymbolNonterminal {
        t.Error("expr should be nonterminal")
    }
    // IDs should be unique
    if plus.ID == minus.ID || plus.ID == expr.ID {
        t.Error("symbol IDs must be unique")
    }
    // Lookup
    s, ok := g.Symbols.Lookup("PLUS")
    if !ok || s != plus {
        t.Error("Lookup(PLUS) failed")
    }
    // EOF is always present
    if g.EOF == nil || g.EOF.ID != 0 {
        t.Error("EOF marker should have ID 0")
    }
}
```

### Test 2: Add duplicate symbol

```go
func TestDuplicateSymbol(t *testing.T) {
    g := NewGrammar()
    s1 := g.AddTerminal("PLUS")
    s2 := g.AddTerminal("PLUS")
    if s1 != s2 {
        t.Error("adding same terminal twice should return same symbol")
    }
}
```

### Test 3: Rule construction

```go
func TestAddRule(t *testing.T) {
    g := NewGrammar()
    g.AddTerminal("PLUS")
    g.AddNonterminal("expr")
    g.AddNonterminal("term")

    r, err := g.AddRule("expr", []string{"expr", "PLUS", "term"}, "")
    if err != nil {
        t.Fatal(err)
    }
    if r.LHS.Name != "expr" {
        t.Error("LHS should be expr")
    }
    if len(r.RHS) != 3 {
        t.Errorf("RHS length = %d, want 3", len(r.RHS))
    }
    if r.Index != 0 {
        t.Errorf("first rule index = %d, want 0", r.Index)
    }
}
```

### Test 4: Rule with unknown symbol fails

```go
func TestAddRuleUnknownSymbol(t *testing.T) {
    g := NewGrammar()
    g.AddNonterminal("expr")
    _, err := g.AddRule("expr", []string{"UNKNOWN"}, "")
    if err == nil {
        t.Error("expected error for unknown symbol in RHS")
    }
}
```

### Test 5: Rule with terminal as LHS fails

```go
func TestAddRuleTerminalLHS(t *testing.T) {
    g := NewGrammar()
    g.AddTerminal("PLUS")
    _, err := g.AddRule("PLUS", []string{}, "")
    if err == nil {
        t.Error("expected error for terminal as LHS")
    }
}
```

### Test 6: RulesFor returns correct subset

```go
func TestRulesFor(t *testing.T) {
    g := NewGrammar()
    g.AddTerminal("PLUS")
    g.AddTerminal("NUM")
    g.AddNonterminal("expr")
    g.AddNonterminal("term")

    g.AddRule("expr", []string{"expr", "PLUS", "term"}, "")
    g.AddRule("expr", []string{"term"}, "")
    g.AddRule("term", []string{"NUM"}, "")

    expr, _ := g.Symbols.Lookup("expr")
    rules := g.RulesFor(expr)
    if len(rules) != 2 {
        t.Errorf("RulesFor(expr) = %d rules, want 2", len(rules))
    }
}
```

## Common mistakes

1. **Forgetting the EOF symbol at ID 0**: The augmented grammar needs `$`.
   Always pre-create it in `NewGrammar()`.

2. **Using string keys everywhere**: String lookups are fine for building the
   grammar, but later algorithms need fast integer-indexed access. That's why
   symbols have IDs.

3. **Allowing terminals as LHS of rules**: Only nonterminals can appear on
   the left-hand side. `AddRule` must check this.

4. **Not handling epsilon rules**: A rule with an empty RHS (like `opt_list →`)
   is valid. `RHS` can be an empty slice.

5. **Making symbol IDs non-deterministic**: Use sequential allocation so the
   grammar model is reproducible. Do not use map iteration order for ID
   assignment.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-03-grammar-model.md and implement
exactly what it describes.

Scope:
- Create internal/grammar/symbol.go with SymbolKind, Symbol, and SymbolTable.
- Create internal/grammar/rule.go with the Rule struct.
- Create internal/grammar/grammar.go with the Grammar struct and builder
  methods: NewGrammar, AddTerminal, AddNonterminal, AddRule, RulesFor.
- NewGrammar() must pre-create the EOF symbol with ID 0.
- Create internal/grammar/grammar_test.go with all six tests from the lesson.
- Use only the Go standard library.
- Keep code simple and explicit. No generics gymnastics.
- Leave TODO comments for: "// TODO(lesson-05): precedence and associativity"
  and "// TODO(lesson-06): grammar validation".

Run `go test ./internal/grammar/...` and ensure all tests pass.
```

## Checkpoint questions

1. Why does the EOF symbol need a fixed ID (0)?
2. What is an "augmented grammar" and why do LALR parsers need one?
3. Can a rule have an empty right-hand side? What does that mean linguistically?
4. Why do we assign integer IDs to symbols instead of just using names?
5. If two different rules have the same LHS and RHS, are they the same rule?
