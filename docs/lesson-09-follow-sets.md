# Lesson 09: FOLLOW Sets

**Goal**: Compute the FOLLOW set for every nonterminal — the set of terminals
that can appear immediately after that nonterminal in any sentential form.

## What you will build in this lesson

- A `ComputeFollow(g *grammar.Grammar, nullable map[int]bool, first map[int]map[int]bool) map[int]map[int]bool` function
- Unit tests with the same tiny grammars from Lesson 08, plus expected FOLLOW sets

## Key concepts

**FOLLOW set**: For a nonterminal `A`, `FOLLOW(A)` is the set of terminals
that can appear immediately to the right of `A` in some sentential form
derived from the start symbol. The end-of-input marker `$` is in
`FOLLOW(start_symbol)`.

**Why we need FOLLOW**: The parser uses FOLLOW sets to decide when to reduce.
If the parser is in a state where it could reduce by rule `A → α`, it should
only do so if the next input token is in `FOLLOW(A)`. (More precisely, LALR
uses per-item lookaheads, but those are computed from FOLLOW-like reasoning.)

**Computing FOLLOW**: For each rule `A → α B β`:
1. Add `FIRST(β)` to `FOLLOW(B)` — whatever can start `β` can follow `B`
2. If `β` is nullable (or empty), add `FOLLOW(A)` to `FOLLOW(B)` — whatever
   can follow `A` can also follow `B` when `β` vanishes

**Seed**: `FOLLOW(start_symbol)` initially contains `$` (EOF).

## Inputs / Outputs

### Files to create

| File | Purpose |
|------|---------|
| `internal/analysis/follow.go` | FOLLOW set computation |
| `internal/analysis/analysis_test.go` | Add FOLLOW tests (same file as Lesson 08) |

### Function signature

```go
// ComputeFollow returns FOLLOW sets indexed by nonterminal symbol ID.
// Each FOLLOW set is a set of terminal symbol IDs.
// The EOF symbol ($) is included in FOLLOW(start) by convention.
func ComputeFollow(
    g *grammar.Grammar,
    nullable map[int]bool,
    first map[int]map[int]bool,
) map[int]map[int]bool
```

## Algorithm sketch

```
function ComputeFollow(grammar, nullable, first):
    follow = {} // map[int]map[int]bool

    // Initialize empty sets for all nonterminals
    for each nonterminal nt:
        follow[nt.ID] = {}

    // Seed: $ ∈ FOLLOW(start)
    follow[grammar.Start.ID][grammar.EOF.ID] = true

    // Fixed-point iteration
    changed = true
    while changed:
        changed = false
        for each rule (A → X₁ X₂ … Xₙ):
            for i = 0 to n-1:
                Xᵢ = rule.RHS[i]
                if Xᵢ is a terminal:
                    continue  // terminals don't have FOLLOW sets

                // β = Xᵢ₊₁ Xᵢ₊₂ … Xₙ
                β = rule.RHS[i+1:]
                firstOfBeta = FirstOfSequence(β, first, nullable)

                // Add FIRST(β) to FOLLOW(Xᵢ)
                for each t in firstOfBeta:
                    if t not in follow[Xᵢ.ID]:
                        follow[Xᵢ.ID][t] = true
                        changed = true

                // If β is nullable (or empty), add FOLLOW(A) to FOLLOW(Xᵢ)
                if allNullable(β, nullable):
                    for each t in follow[A.ID]:
                        if t not in follow[Xᵢ.ID]:
                            follow[Xᵢ.ID][t] = true
                            changed = true

    return follow

function allNullable(symbols, nullable):
    for each s in symbols:
        if s.ID not in nullable:
            return false
    return true  // empty sequence is vacuously nullable
```

**Termination**: FOLLOW sets grow monotonically and are bounded by the number
of terminals + 1 (for $). The loop terminates.

**Invariant**: After the algorithm completes, for every nonterminal B, if
terminal t can appear immediately after B in any derivation from the start
symbol, then t ∈ FOLLOW(B).

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/analysis/follow.go` |
| Modify | `internal/analysis/analysis_test.go` — add FOLLOW tests |

## Unit tests (pass/fail gate)

### Grammar A (arithmetic): `expr → expr PLUS term | term; term → NUM`

Expected FOLLOW sets:
- FOLLOW(expr) = {PLUS, $} — PLUS follows expr in `expr PLUS term`; $ follows because expr is the start
- FOLLOW(term) = {PLUS, $} — term can be followed by PLUS (via `expr → expr PLUS term` and `expr → term`) or by $

### Grammar B (with epsilon): `list → list item | ε; item → WORD`

Expected FOLLOW sets:
- FOLLOW(list) = {WORD, $} — WORD follows list (via `list → list item` where FIRST(item)={WORD}); $ follows because list is start
- FOLLOW(item) = {WORD, $} — item can be followed by WORD (next item) or $ (end)

### Grammar C (chained nullable): `S → A B c; A → a | ε; B → b | ε`

Expected FOLLOW sets:
- FOLLOW(S) = {$}
- FOLLOW(A) = {b, c} — in `S → A B c`, FIRST(B c) = {b, c} since B is nullable
- FOLLOW(B) = {c} — in `S → A B c`, c follows B

### Test 1: FOLLOW for arithmetic grammar

```go
func TestFollowArithmetic(t *testing.T) {
    g := buildGrammarA()
    nullable := ComputeNullable(g)
    first := ComputeFirst(g, nullable)
    follow := ComputeFollow(g, nullable, first)

    expr, _ := g.Symbols.Lookup("expr")
    term, _ := g.Symbols.Lookup("term")
    plus, _ := g.Symbols.Lookup("PLUS")

    // FOLLOW(expr) should contain PLUS and $
    if !follow[expr.ID][plus.ID] {
        t.Error("FOLLOW(expr) should contain PLUS")
    }
    if !follow[expr.ID][g.EOF.ID] {
        t.Error("FOLLOW(expr) should contain $")
    }

    // FOLLOW(term) should contain PLUS and $
    if !follow[term.ID][plus.ID] {
        t.Error("FOLLOW(term) should contain PLUS")
    }
    if !follow[term.ID][g.EOF.ID] {
        t.Error("FOLLOW(term) should contain $")
    }
}
```

### Test 2: FOLLOW with epsilon

```go
func TestFollowEpsilon(t *testing.T) {
    g := buildGrammarB()
    nullable := ComputeNullable(g)
    first := ComputeFirst(g, nullable)
    follow := ComputeFollow(g, nullable, first)

    list, _ := g.Symbols.Lookup("list")
    item, _ := g.Symbols.Lookup("item")
    word, _ := g.Symbols.Lookup("WORD")

    if !follow[list.ID][word.ID] {
        t.Error("FOLLOW(list) should contain WORD")
    }
    if !follow[list.ID][g.EOF.ID] {
        t.Error("FOLLOW(list) should contain $")
    }
    if !follow[item.ID][word.ID] {
        t.Error("FOLLOW(item) should contain WORD")
    }
    if !follow[item.ID][g.EOF.ID] {
        t.Error("FOLLOW(item) should contain $")
    }
}
```

### Test 3: FOLLOW with chained nullable

```go
func TestFollowChainedNullable(t *testing.T) {
    g := buildGrammarC()
    nullable := ComputeNullable(g)
    first := ComputeFirst(g, nullable)
    follow := ComputeFollow(g, nullable, first)

    a, _ := g.Symbols.Lookup("A")
    bNT, _ := g.Symbols.Lookup("B")
    s, _ := g.Symbols.Lookup("S")
    bTerm, _ := g.Symbols.Lookup("b")
    c, _ := g.Symbols.Lookup("c")

    // FOLLOW(S) = {$}
    if !follow[s.ID][g.EOF.ID] {
        t.Error("FOLLOW(S) should contain $")
    }
    if len(follow[s.ID]) != 1 {
        t.Errorf("FOLLOW(S) should have exactly 1 element, got %d", len(follow[s.ID]))
    }

    // FOLLOW(A) = {b, c}
    if !follow[a.ID][bTerm.ID] {
        t.Error("FOLLOW(A) should contain b")
    }
    if !follow[a.ID][c.ID] {
        t.Error("FOLLOW(A) should contain c")
    }

    // FOLLOW(B) = {c}
    if !follow[bNT.ID][c.ID] {
        t.Error("FOLLOW(B) should contain c")
    }
    if len(follow[bNT.ID]) != 1 {
        t.Errorf("FOLLOW(B) should have exactly 1 element, got %d", len(follow[bNT.ID]))
    }
}
```

### Test 4: FOLLOW does not include FIRST(ε) confusion

```go
func TestFollowNoEpsilon(t *testing.T) {
    // FOLLOW sets should only contain terminal IDs, never nonterminal IDs
    g := buildGrammarC()
    nullable := ComputeNullable(g)
    first := ComputeFirst(g, nullable)
    follow := ComputeFollow(g, nullable, first)

    for _, nt := range g.Symbols.Nonterminals() {
        for id := range follow[nt.ID] {
            sym := g.Symbols.ByID(id)
            if sym == nil {
                t.Errorf("FOLLOW(%s) contains unknown ID %d", nt.Name, id)
            } else if sym.Kind != grammar.SymbolTerminal {
                t.Errorf("FOLLOW(%s) contains nonterminal %s", nt.Name, sym.Name)
            }
        }
    }
}
```

## Common mistakes

1. **Forgetting to seed FOLLOW(start) with $**: Without this, the algorithm
   can't propagate $ to other nonterminals.

2. **Not handling empty β**: When a nonterminal is the last symbol in a rule
   (β is empty), FOLLOW(A) must be added to FOLLOW(B). An empty sequence is
   vacuously nullable.

3. **Adding nonterminal IDs to FOLLOW**: FOLLOW sets should contain only
   terminal IDs (and $). This usually happens when `FirstOfSequence` returns
   nonterminal IDs by mistake.

4. **Skipping terminals in the RHS**: When scanning `A → X₁ X₂ ... Xₙ`, only
   compute FOLLOW for nonterminals. Skip terminals.

5. **Not iterating to a fixed point**: One pass through the rules is not
   enough. FOLLOW(A) may update FOLLOW(B) which in turn updates FOLLOW(C).

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-09-follow-sets.md and implement
exactly what it describes.

Scope:
- Create internal/analysis/follow.go with ComputeFollow function.
- Add to internal/analysis/analysis_test.go: four FOLLOW tests using the
  same helper grammars from Lesson 08.
- Add a ByID(id int) *Symbol method to SymbolTable if not already present.
- ComputeFollow must use fixed-point iteration, seed FOLLOW(start) with $,
  and only add terminal IDs to FOLLOW sets.
- Use only the Go standard library.
- Keep code clear. Add comments explaining each step of the algorithm.

Run `go test ./internal/analysis/...` and ensure all tests pass.
```

## Checkpoint questions

1. Why is `$` (EOF) in `FOLLOW(start_symbol)` but not in any FIRST set?
2. In Grammar C, why is `c` in `FOLLOW(A)`?
3. Can `FOLLOW(A)` ever be empty for a reachable nonterminal? Give an
   example or argue why not.
4. If a nonterminal `A` is unreachable, what is `FOLLOW(A)`?
5. How does the FOLLOW computation handle a rule like `A → B C D` where
   both `C` and `D` are nullable?
