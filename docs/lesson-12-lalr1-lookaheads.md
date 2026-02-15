# Lesson 12: LALR(1) Lookaheads and State Merging

**Goal**: Promote the LR(0) automaton to LALR(1) by computing lookahead sets
for each reduce item, using the DeRemer and Pennello approach of propagating
lookaheads through the state graph.

## What you will build in this lesson

- A `ComputeLookaheads(automaton *Automaton, g *grammar.Grammar, first, nullable) *Automaton`
  function that annotates reduce items with lookahead terminal sets
- A `LAItem` struct that extends Item with a lookahead set
- Lookahead propagation: spontaneous generation and propagation links
- The complete LALR(1) state machine ready for table construction

## Key concepts

**Why LR(0) isn't enough**: An LR(0) item `A → α •` says "reduce by rule A"
but doesn't say *when*. It would reduce on any input, causing spurious
conflicts. LALR(1) adds a lookahead set to each reduce item: reduce only when
the next input token is in the lookahead set.

**LALR(1) vs. LR(1)**: Full LR(1) items carry their lookahead from the start,
creating potentially many more states. LALR(1) starts from LR(0) states and
computes lookaheads as a separate step — fewer states, same power for most
grammars.

**Lookahead computation (simplified approach)**:
For each reduce item `A → α •` in state S, the lookahead is FOLLOW(A)
restricted to the context in which state S was entered. The practical approach:

1. **Spontaneous generation**: In state S, if item `A → α • B β` exists and
   `β` produces first terminals, those terminals are spontaneous lookaheads
   for any reduce item of B reached from S.

2. **Propagation**: If `β` is nullable (or empty), lookaheads for A in S
   propagate to lookaheads for B in GOTO(S, ...).

A simpler but sufficient approach for our purposes: **Use FOLLOW sets as the
initial lookahead sets for reduce items, then refine by state-specific
context.** For a basic working LALR(1) parser generator, assigning
FOLLOW(LHS) as the lookahead for each reduce item produces an SLR(1) parser,
which we then refine.

**Our approach**: We implement the standard LALR(1) lookahead propagation
algorithm:

1. For each state and each kernel item, determine which lookaheads are
   generated spontaneously and which propagate from other items.
2. Iterate to a fixed point.

## Inputs / Outputs

### Files to create / modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/lalr/lookahead.go` | Create | Lookahead computation |
| `internal/lalr/lalr_test.go` | Modify | Add lookahead tests |

### Types and functions

```go
// LAItem is an LR(0) item annotated with an LALR(1) lookahead set.
type LAItem struct {
    Item
    Lookahead map[int]bool // set of terminal symbol IDs
}

// LAState is a state with lookahead-annotated items.
type LAState struct {
    ID       int
    Items    []LAItem
    Kernel   []LAItem // kernel items only (for debugging)
}

// ComputeLookaheads annotates the automaton's states with LALR(1) lookaheads.
// It modifies the automaton in place, adding lookahead sets to reduce items.
func ComputeLookaheads(
    automaton *Automaton,
    g *grammar.Grammar,
    nullable map[int]bool,
    first map[int]map[int]bool,
)
```

## Algorithm sketch

### Phase 1: Determine spontaneous lookaheads and propagation links

```
For each state S in automaton:
    For each kernel item I = (A → α • X β) in S:
        // Use a probe: create a temporary LR(1) item with lookahead = {#}
        // where # is a dummy symbol not in the grammar

        J = Closure({(A → α • X β, {#})})  // LR(1) closure with dummy

        For each item (B → γ • , la) in J:  // reduce items in closure
            if la contains real terminals (not #):
                // Spontaneous: add those terminals to lookahead of
                // corresponding reduce item in GOTO(S, ...)
                spontaneous[GOTO(S, nextSym), reduceItem] ∪= la - {#}
            if la contains #:
                // Propagation: lookahead of I propagates to this item
                propagation_links.add(S:I → GOTO_state:reduceItem)
```

### Phase 2: Propagate to fixed point

```
// Seed: $accept → • S has lookahead {$}
state0.kernelItems[$accept → • S].Lookahead = {$}

// Add spontaneous lookaheads
For each (state, item, terminals) in spontaneous:
    item.Lookahead ∪= terminals

// Propagate
changed = true
while changed:
    changed = false
    For each link (srcState:srcItem → dstState:dstItem):
        For each t in srcItem.Lookahead:
            if t not in dstItem.Lookahead:
                dstItem.Lookahead.add(t)
                changed = true
```

**Simpler alternative (SLR-like, with refinement)**:
For a first working version, assign `FOLLOW(LHS)` as the lookahead for every
reduce item. This is the SLR(1) approach. It works for many grammars and is
much simpler. We can upgrade to full LALR(1) if SLR causes problems.

The implementation should support both approaches: SLR as the default with
a flag or method for full LALR(1).

**Termination**: Lookahead sets grow monotonically and are bounded by the
number of terminals. The propagation loop terminates.

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/lalr/lookahead.go` |
| Modify | `internal/lalr/state.go` — add LAState, LAItem |
| Modify | `internal/lalr/lalr_test.go` — add lookahead tests |

## Unit tests (pass/fail gate)

### Test 1: Accept rule gets $ lookahead

```go
func TestAcceptRuleLookahead(t *testing.T) {
    g := buildAugmentedArith()
    nullable := analysis.ComputeNullable(g)
    first := analysis.ComputeFirst(g, nullable)
    automaton := BuildCanonical(g)
    ComputeLookaheads(automaton, g, nullable, first)

    // Find the state containing $accept → expr •
    acceptRule := findAcceptRule(g)
    for _, s := range automaton.States {
        for _, item := range s.LAItems {
            if item.RuleIndex == acceptRule.Index && item.IsReduce(g) {
                if !item.Lookahead[g.EOF.ID] {
                    t.Error("$accept → expr • should have $ in lookahead")
                }
            }
        }
    }
}
```

### Test 2: Reduce item has correct lookahead

```go
func TestReduceItemLookahead(t *testing.T) {
    g := buildAugmentedArith()
    nullable := analysis.ComputeNullable(g)
    first := analysis.ComputeFirst(g, nullable)
    automaton := BuildCanonical(g)
    ComputeLookaheads(automaton, g, nullable, first)

    // Find state with: term → NUM •
    // Its lookahead should be FOLLOW(term) = {PLUS, $}
    termRule := findRule(g, "term", []string{"NUM"})
    plus, _ := g.Symbols.Lookup("PLUS")

    found := false
    for _, s := range automaton.States {
        for _, item := range s.LAItems {
            if item.RuleIndex == termRule.Index && item.IsReduce(g) {
                found = true
                if !item.Lookahead[plus.ID] {
                    t.Error("term → NUM • should have PLUS in lookahead")
                }
                if !item.Lookahead[g.EOF.ID] {
                    t.Error("term → NUM • should have $ in lookahead")
                }
            }
        }
    }
    if !found {
        t.Error("could not find reduce item for term → NUM •")
    }
}
```

### Test 3: Non-reduce items don't need lookaheads

```go
func TestNonReduceNoLookahead(t *testing.T) {
    g := buildAugmentedArith()
    nullable := analysis.ComputeNullable(g)
    first := analysis.ComputeFirst(g, nullable)
    automaton := BuildCanonical(g)
    ComputeLookaheads(automaton, g, nullable, first)

    for _, s := range automaton.States {
        for _, item := range s.LAItems {
            if !item.IsReduce(g) && len(item.Lookahead) > 0 {
                // Non-reduce items may have lookaheads for propagation,
                // but it's not an error. Just verify reduce items have them.
            }
        }
    }
}
```

### Test 4: SLR lookaheads match FOLLOW

```go
func TestSLRLookaheadsMatchFollow(t *testing.T) {
    // For the arithmetic grammar, SLR and LALR(1) should agree
    g := buildAugmentedArith()
    nullable := analysis.ComputeNullable(g)
    first := analysis.ComputeFirst(g, nullable)
    follow := analysis.ComputeFollow(g, nullable, first)
    automaton := BuildCanonical(g)
    ComputeLookaheads(automaton, g, nullable, first)

    for _, s := range automaton.States {
        for _, item := range s.LAItems {
            if item.IsReduce(g) {
                rule := g.Rules[item.RuleIndex]
                if rule.LHS.Name == "$accept" {
                    continue
                }
                // For SLR grammars, lookahead ⊆ FOLLOW(LHS)
                for t := range item.Lookahead {
                    if !follow[rule.LHS.ID][t] {
                        sym := g.Symbols.ByID(t)
                        t.Errorf("lookahead %s for %s not in FOLLOW(%s)",
                            sym.Name, rule.LHS.Name, rule.LHS.Name)
                    }
                }
            }
        }
    }
}
```

### Test 5: Automaton has LAItems after computation

```go
func TestAutomatonHasLAItems(t *testing.T) {
    g := buildAugmentedArith()
    nullable := analysis.ComputeNullable(g)
    first := analysis.ComputeFirst(g, nullable)
    automaton := BuildCanonical(g)
    ComputeLookaheads(automaton, g, nullable, first)

    for _, s := range automaton.States {
        if len(s.LAItems) == 0 {
            t.Errorf("state %d has no LAItems after lookahead computation", s.ID)
        }
    }
}
```

## Common mistakes

1. **Forgetting to seed $accept with $ lookahead**: The augmented start rule's
   reduce item must have `{$}` as its lookahead. Everything propagates from
   there.

2. **Mixing up kernel and non-kernel items**: Kernel items are: (a) the
   initial item `$accept → • S`, and (b) all items where dot > 0. Non-kernel
   items have dot=0 and are added by closure. Lookaheads propagate through
   kernel items.

3. **Not propagating transitively**: Lookahead propagation requires a
   fixed-point loop. One pass is not sufficient for chains of propagation.

4. **Incorrect SLR fallback**: SLR uses FOLLOW(LHS) directly, which is a
   superset of the true LALR(1) lookahead. This can introduce spurious
   conflicts. Start with SLR for simplicity, but be prepared to upgrade.

5. **Confusing state-specific vs. global lookaheads**: The same reduce item
   (same rule, dot at end) might appear in multiple states with *different*
   lookahead sets. Each state has its own context.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-12-lalr1-lookaheads.md and
implement exactly what it describes.

Scope:
- Create internal/lalr/lookahead.go with ComputeLookaheads function.
- Add LAItem and LAState types (or add Lookahead field to existing structures).
- Implement SLR(1) lookahead assignment as the initial approach: for each
  reduce item, set lookahead = FOLLOW(LHS). Mark with a TODO for upgrading
  to full LALR(1) propagation if needed.
- Modify State to carry LAItems after lookahead computation.
- Add five tests to internal/lalr/lalr_test.go.
- Use only the Go standard library.
- Import the analysis package for FOLLOW computation.
- Keep code clear and document the SLR vs LALR(1) tradeoff.

Run `go test ./internal/lalr/...` and ensure all tests pass.
```

## Checkpoint questions

1. What is the difference between SLR(1) and LALR(1) lookaheads?
2. Why does the accept rule's reduce item have only `{$}` as its lookahead?
3. Can two reduce items in the same state have different lookahead sets?
4. What is a "kernel item" and why is the distinction important?
5. Give an example of a grammar where SLR(1) has a conflict but LALR(1) does
   not.
