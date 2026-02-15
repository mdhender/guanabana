# Lesson 08: Nullable and FIRST Sets

**Goal**: Compute the nullable set (which nonterminals can derive the empty
string) and FIRST sets (which terminals can begin a string derived from each
symbol).

## What you will build in this lesson

- A `ComputeNullable(g *grammar.Grammar) map[int]bool` function
- A `ComputeFirst(g *grammar.Grammar, nullable map[int]bool) map[int]map[int]bool` function
- FIRST for individual symbols and for sequences of symbols
- Unit tests with small grammars and hand-verified expected sets

## Key concepts

**Nullable**: A nonterminal `A` is nullable if there exists a derivation
`A ⇒* ε` (the empty string). Terminals are never nullable. A nonterminal
is nullable if at least one of its rules has an all-nullable RHS (including
an empty RHS).

**FIRST set**: For a symbol `X`:
- If `X` is a terminal: `FIRST(X) = {X}`
- If `X` is a nonterminal: `FIRST(X)` is the set of terminals that can begin
  a string derived from `X`

For a sequence `α = X₁ X₂ … Xₙ`:
- Start with FIRST(X₁)
- If X₁ is nullable, add FIRST(X₂)
- If X₁ and X₂ are both nullable, add FIRST(X₃)
- … and so on

**Why we need these**: The FOLLOW set computation (Lesson 09) needs FIRST.
The LALR lookahead computation (Lesson 12) needs both. Without correct
nullable and FIRST, the parse tables will be wrong.

## Inputs / Outputs

### Files to create

| File | Purpose |
|------|---------|
| `internal/analysis/nullable.go` | Nullable set computation |
| `internal/analysis/first.go` | FIRST set computation |
| `internal/analysis/analysis_test.go` | Tests for both |

### Function signatures

```go
package analysis

import "github.com/mdhender/guanabana/internal/grammar"

// ComputeNullable returns a set of symbol IDs that are nullable.
// Only nonterminals can be nullable. Terminals are never nullable.
func ComputeNullable(g *grammar.Grammar) map[int]bool

// ComputeFirst returns FIRST sets indexed by symbol ID.
// Each FIRST set is a set of terminal symbol IDs.
// For a terminal T, FIRST(T) = {T.ID}.
func ComputeFirst(g *grammar.Grammar, nullable map[int]bool) map[int]map[int]bool

// FirstOfSequence computes FIRST for a sequence of symbols.
// Returns the set of terminal IDs that can begin the sequence.
func FirstOfSequence(
    seq []*grammar.Symbol,
    first map[int]map[int]bool,
    nullable map[int]bool,
) map[int]bool
```

## Algorithm sketch

### Nullable (fixed-point iteration)

```
function ComputeNullable(grammar):
    nullable = {}

    // A nonterminal with an empty rule is immediately nullable
    for each rule in grammar:
        if len(rule.RHS) == 0:
            nullable[rule.LHS.ID] = true

    // Iterate until no changes
    changed = true
    while changed:
        changed = false
        for each rule in grammar:
            if rule.LHS.ID in nullable:
                continue  // already known nullable
            allNullable = true
            for each sym in rule.RHS:
                if sym.ID not in nullable:
                    allNullable = false
                    break
            if allNullable:
                nullable[rule.LHS.ID] = true
                changed = true

    return nullable
```

**Termination**: The nullable set grows monotonically (symbols are only added,
never removed) and is bounded by the number of nonterminals. Therefore the
loop terminates.

**Invariant**: After the algorithm, `nullable[id]` is true if and only if
there exists a derivation from the symbol to ε.

### FIRST sets (fixed-point iteration)

```
function ComputeFirst(grammar, nullable):
    first = {}

    // Initialize: FIRST(terminal) = {terminal}
    for each terminal t:
        first[t.ID] = {t.ID}

    // Initialize: FIRST(nonterminal) = {}
    for each nonterminal nt:
        first[nt.ID] = {}

    // Iterate until no changes
    changed = true
    while changed:
        changed = false
        for each rule (A → X₁ X₂ … Xₙ):
            for i = 0 to n-1:
                Xᵢ = rule.RHS[i]
                for each t in first[Xᵢ.ID]:
                    if t not in first[A.ID]:
                        first[A.ID].add(t)
                        changed = true
                if Xᵢ.ID not in nullable:
                    break  // stop scanning RHS

    return first
```

**Termination**: FIRST sets grow monotonically and are bounded by the total
number of terminals. The loop terminates.

### FIRST of a sequence

```
function FirstOfSequence(seq, first, nullable):
    result = {}
    for each sym in seq:
        result = result ∪ first[sym.ID]
        if sym.ID not in nullable:
            break
    return result
```

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/analysis/nullable.go` |
| Create | `internal/analysis/first.go` |
| Create | `internal/analysis/analysis_test.go` |

## Unit tests (pass/fail gate)

We use three tiny grammars for testing:

**Grammar A** (basic arithmetic):
```
expr → expr PLUS term
expr → term
term → NUM
```
Expected: nullable = {} (nothing is nullable). FIRST(expr) = {NUM}. FIRST(term) = {NUM}. FIRST(PLUS) = {PLUS}. FIRST(NUM) = {NUM}.

**Grammar B** (with epsilon rule):
```
list → list item
list → (empty)
item → WORD
```
Expected: nullable = {list}. FIRST(list) = {WORD}. FIRST(item) = {WORD}.

**Grammar C** (chained nullable):
```
S → A B c
A → a
A → (empty)
B → b
B → (empty)
```
Expected: nullable = {A, B}. FIRST(S) = {a, b, c}. FIRST(A) = {a}. FIRST(B) = {b}.

### Test 1: Nullable — nothing nullable

```go
func TestNullableNone(t *testing.T) {
    g := buildGrammarA() // expr → expr PLUS term | term; term → NUM
    nullable := ComputeNullable(g)
    for _, nt := range g.Symbols.Nonterminals() {
        if nullable[nt.ID] {
            t.Errorf("%s should not be nullable", nt.Name)
        }
    }
}
```

### Test 2: Nullable — epsilon rule

```go
func TestNullableEpsilon(t *testing.T) {
    g := buildGrammarB() // list → list item | ε; item → WORD
    nullable := ComputeNullable(g)
    list, _ := g.Symbols.Lookup("list")
    item, _ := g.Symbols.Lookup("item")
    if !nullable[list.ID] {
        t.Error("list should be nullable")
    }
    if nullable[item.ID] {
        t.Error("item should not be nullable")
    }
}
```

### Test 3: Nullable — chained

```go
func TestNullableChained(t *testing.T) {
    g := buildGrammarC() // S → A B c; A → a | ε; B → b | ε
    nullable := ComputeNullable(g)
    a, _ := g.Symbols.Lookup("A")
    b, _ := g.Symbols.Lookup("B")
    s, _ := g.Symbols.Lookup("S")
    if !nullable[a.ID] {
        t.Error("A should be nullable")
    }
    if !nullable[b.ID] {
        t.Error("B should be nullable")
    }
    if nullable[s.ID] {
        t.Error("S should not be nullable (c is terminal)")
    }
}
```

### Test 4: FIRST — basic

```go
func TestFirstBasic(t *testing.T) {
    g := buildGrammarA()
    nullable := ComputeNullable(g)
    first := ComputeFirst(g, nullable)

    expr, _ := g.Symbols.Lookup("expr")
    num, _ := g.Symbols.Lookup("NUM")

    if !first[expr.ID][num.ID] {
        t.Error("FIRST(expr) should contain NUM")
    }
    plus, _ := g.Symbols.Lookup("PLUS")
    if first[expr.ID][plus.ID] {
        t.Error("FIRST(expr) should not contain PLUS")
    }
}
```

### Test 5: FIRST — with nullable prefix

```go
func TestFirstNullablePrefix(t *testing.T) {
    g := buildGrammarC()
    nullable := ComputeNullable(g)
    first := ComputeFirst(g, nullable)

    s, _ := g.Symbols.Lookup("S")
    a, _ := g.Symbols.Lookup("a")
    b, _ := g.Symbols.Lookup("b")
    c, _ := g.Symbols.Lookup("c")

    // FIRST(S) should be {a, b, c} because A and B are both nullable
    if !first[s.ID][a.ID] {
        t.Error("FIRST(S) should contain a")
    }
    if !first[s.ID][b.ID] {
        t.Error("FIRST(S) should contain b")
    }
    if !first[s.ID][c.ID] {
        t.Error("FIRST(S) should contain c")
    }
}
```

### Test 6: FirstOfSequence

```go
func TestFirstOfSequence(t *testing.T) {
    g := buildGrammarC()
    nullable := ComputeNullable(g)
    first := ComputeFirst(g, nullable)

    a, _ := g.Symbols.Lookup("A")
    b, _ := g.Symbols.Lookup("B")
    symB, _ := g.Symbols.Lookup("b")

    // FIRST(A B) = {a, b} because A is nullable
    seq := []*grammar.Symbol{a, b}
    result := FirstOfSequence(seq, first, nullable)
    if len(result) != 2 {
        t.Errorf("FIRST(A B) has %d elements, want 2", len(result))
    }
    if !result[symB.ID] {
        t.Error("FIRST(A B) should contain b")
    }
}
```

### Helper: build test grammars programmatically

```go
func buildGrammarA() *grammar.Grammar {
    g := grammar.NewGrammar()
    g.AddTerminal("PLUS")
    g.AddTerminal("NUM")
    g.AddNonterminal("expr")
    g.AddNonterminal("term")
    g.AddRule("expr", []string{"expr", "PLUS", "term"}, "")
    g.AddRule("expr", []string{"term"}, "")
    g.AddRule("term", []string{"NUM"}, "")
    g.Finalize()
    return g
}
```

Note: Use lowercase single-letter terminals `a`, `b`, `c` for Grammar C.
The grammar model allows this — the case convention is a lexer concern for
`.y` files, but the programmatic API accepts any names.

## Common mistakes

1. **Forgetting that terminals are never nullable**: Only nonterminals can
   have epsilon rules. Don't even check terminals.

2. **Not handling empty rules**: A rule with `len(RHS) == 0` makes its LHS
   immediately nullable. Missing this bootstraps the entire computation wrong.

3. **Breaking out of the RHS loop too early or too late**: When computing
   FIRST, stop adding FIRST(Xᵢ) as soon as you hit a non-nullable symbol.
   If you don't break, you'll add too many terminals.

4. **Confusing symbol IDs**: FIRST sets are indexed by symbol ID and contain
   symbol IDs. Don't mix up names and IDs.

5. **Not running to a fixed point**: A single pass is not enough. Nonterminals
   that depend on other nonterminals need multiple iterations.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-08-nullable-first.md and
implement exactly what it describes.

Scope:
- Create internal/analysis/nullable.go with ComputeNullable.
- Create internal/analysis/first.go with ComputeFirst and FirstOfSequence.
- Create internal/analysis/analysis_test.go with all six tests plus the
  helper functions to build test grammars A, B, and C programmatically.
- Both algorithms use fixed-point iteration.
- FIRST sets are maps from symbol ID to sets of terminal IDs.
- Use only the Go standard library.
- Keep code clear and well-commented.
- Leave TODO: "// TODO(lesson-09): FOLLOW sets"

Run `go test ./internal/analysis/...` and ensure all tests pass.
```

## Checkpoint questions

1. Can a terminal ever be nullable? Why or why not?
2. How many iterations of the nullable fixed-point loop are needed for
   Grammar C? Walk through each iteration.
3. Why does `FIRST(S)` in Grammar C contain `c`?
4. What would happen if `FirstOfSequence` didn't check nullable?
5. Is there a grammar where the FIRST fixed-point needs more than 2
   iterations? Give an example.
