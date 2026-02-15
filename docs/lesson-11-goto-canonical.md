# Lesson 11: GOTO and the Canonical Collection

**Goal**: Implement the GOTO function and build the canonical collection of
LR(0) states — the complete set of parser states and their transitions.

## What you will build in this lesson

- A `Goto(items ItemSet, sym *grammar.Symbol, g *grammar.Grammar) ItemSet`
  function
- A `State` struct holding an ItemSet and an integer ID
- A `BuildCanonical(g *grammar.Grammar) *Automaton` function that computes
  all states and transitions
- An `Automaton` struct that holds the states and transition table
- Unit tests verifying the automaton for the arithmetic grammar

## Key concepts

**GOTO function**: Given a state (an item set) and a grammar symbol X,
`GOTO(I, X)` computes the next state. It works by:
1. Finding all items in I where the dot is immediately before X
2. Advancing the dot past X in each such item
3. Taking the closure of the resulting set

The result is a new state (or an existing state, if we've seen that item set
before).

**Canonical LR(0) collection**: The set of all reachable states starting from
the initial state (closure of `{$accept → • S}`). We build it by computing
GOTO for every state and every grammar symbol, adding new states as we
discover them.

**State numbering**: States are numbered sequentially (0, 1, 2, …) in the
order they are discovered. State 0 is always the initial state. Deterministic
ordering is important for reproducible parse tables.

**Transitions**: Each transition is a triple (from_state, symbol, to_state).
We store these in the automaton for use in parse table construction.

## Inputs / Outputs

### Files to create

| File | Purpose |
|------|---------|
| `internal/lalr/goto.go` | GOTO function |
| `internal/lalr/state.go` | State struct |
| `internal/lalr/collection.go` | BuildCanonical, Automaton |
| `internal/lalr/lalr_test.go` | Add tests (same file as Lesson 10) |

### Types defined

```go
// internal/lalr/state.go

type State struct {
    ID    int
    Items ItemSet
}

// internal/lalr/collection.go

type Transition struct {
    From   int // state ID
    Symbol *grammar.Symbol
    To     int // state ID
}

type Automaton struct {
    States      []*State
    Transitions []Transition
    Grammar     *grammar.Grammar
}

func BuildCanonical(g *grammar.Grammar) *Automaton
func Goto(items ItemSet, sym *grammar.Symbol, g *grammar.Grammar) ItemSet
```

## Algorithm sketch

### GOTO

```
function Goto(items, X, grammar):
    advanced = new ItemSet
    for each item in items:
        sym = item.SymbolAfterDot(grammar)
        if sym == X:
            advanced.Add(Item{RuleIndex: item.RuleIndex, Dot: item.Dot + 1})
    return Closure(advanced, grammar)
```

### BuildCanonical

```
function BuildCanonical(grammar):
    // Initial state: closure of {$accept → • start_symbol}
    acceptRule = find rule where LHS.Name == "$accept"
    initialItems = Closure(NewItemSet(Item{acceptRule.Index, 0}), grammar)
    state0 = State{ID: 0, Items: initialItems}

    automaton = Automaton{States: [state0]}
    stateMap = map[itemset_key]int  // maps item sets to state IDs
    stateMap[key(state0.Items)] = 0
    worklist = [0]

    while worklist not empty:
        stateID = worklist.pop_front()
        state = automaton.States[stateID]

        // Find all symbols that appear after a dot
        symbols = unique symbols after dot in state.Items

        for each sym in symbols:
            targetItems = Goto(state.Items, sym, grammar)
            if targetItems is empty:
                continue

            k = key(targetItems)
            if k in stateMap:
                targetID = stateMap[k]
            else:
                targetID = len(automaton.States)
                stateMap[k] = targetID
                newState = State{ID: targetID, Items: targetItems}
                automaton.States = append(automaton.States, newState)
                worklist.append(targetID)

            automaton.Transitions = append(automaton.Transitions,
                Transition{From: stateID, Symbol: sym, To: targetID})

    return automaton
```

**Termination**: The number of possible item sets is finite (bounded by 2^(items count),
practically much smaller). Each unique item set is created at most once.
The worklist processes each state exactly once.

**Invariants**:
- State 0 is always the initial state
- No two states have the same item set
- Every transition connects existing states
- All states are reachable from state 0

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/lalr/goto.go` |
| Create | `internal/lalr/state.go` |
| Create | `internal/lalr/collection.go` |
| Modify | `internal/lalr/lalr_test.go` — add GOTO and canonical collection tests |

## Unit tests (pass/fail gate)

### Grammar: augmented arithmetic
```
Rule 0: $accept → expr
Rule 1: expr → expr PLUS term
Rule 2: expr → term
Rule 3: term → NUM
```

This grammar should produce exactly **8 states** (verify by hand or from a
reference implementation):

| State | Core items |
|-------|------------|
| 0 | $accept → • expr, expr → • expr PLUS term, expr → • term, term → • NUM |
| 1 | $accept → expr •, expr → expr • PLUS term |
| 2 | expr → term • |
| 3 | term → NUM • |
| 4 | expr → expr PLUS • term, term → • NUM |
| 5 | expr → expr PLUS term • |

(Note: the exact count may be 6 for this grammar — count carefully with your implementation.)

### Test 1: GOTO basic

```go
func TestGotoBasic(t *testing.T) {
    g := buildAugmentedArith()
    // Build state 0
    acceptRule := findAcceptRule(g)
    state0 := Closure(NewItemSet(Item{RuleIndex: acceptRule.Index, Dot: 0}), g)

    num, _ := g.Symbols.Lookup("NUM")
    gotoNUM := Goto(state0, num, g)

    // GOTO(state0, NUM) should contain: term → NUM •
    if gotoNUM.Len() == 0 {
        t.Fatal("GOTO(state0, NUM) should not be empty")
    }
    termRule := findRule(g, "term", []string{"NUM"})
    if !gotoNUM.Contains(Item{RuleIndex: termRule.Index, Dot: 1}) {
        t.Error("GOTO(state0, NUM) should contain term → NUM •")
    }
}
```

### Test 2: GOTO returns empty for no match

```go
func TestGotoEmpty(t *testing.T) {
    g := buildAugmentedArith()
    acceptRule := findAcceptRule(g)
    state0 := Closure(NewItemSet(Item{RuleIndex: acceptRule.Index, Dot: 0}), g)

    // PLUS cannot appear after dot in state 0 (no item has • PLUS)
    // Actually, in state 0: expr → • expr PLUS term — dot is before expr, not PLUS
    // So GOTO(state0, PLUS) should be empty... wait, let me check.
    // Hmm, PLUS is not after any dot in state 0. Only expr, term, NUM are.
    plus, _ := g.Symbols.Lookup("PLUS")
    gotoPLUS := Goto(state0, plus, g)
    if gotoPLUS.Len() != 0 {
        t.Errorf("GOTO(state0, PLUS) should be empty, got %d items", gotoPLUS.Len())
    }
}
```

### Test 3: Canonical collection state count

```go
func TestCanonicalStateCount(t *testing.T) {
    g := buildAugmentedArith()
    automaton := BuildCanonical(g)

    // This small grammar should produce exactly 6 states
    // State 0: $accept→•expr, expr→•expr PLUS term, expr→•term, term→•NUM
    // State 1: $accept→expr•, expr→expr•PLUS term
    // State 2: expr→term•
    // State 3: term→NUM•
    // State 4: expr→expr PLUS•term, term→•NUM
    // State 5: expr→expr PLUS term•
    if len(automaton.States) != 6 {
        t.Errorf("state count = %d, want 6", len(automaton.States))
        for _, s := range automaton.States {
            t.Logf("State %d: %v", s.ID, s.Items)
        }
    }
}
```

### Test 4: State 0 is initial

```go
func TestState0IsInitial(t *testing.T) {
    g := buildAugmentedArith()
    automaton := BuildCanonical(g)

    state0 := automaton.States[0]
    acceptRule := findAcceptRule(g)
    if !state0.Items.Contains(Item{RuleIndex: acceptRule.Index, Dot: 0}) {
        t.Error("state 0 should contain $accept → • expr")
    }
}
```

### Test 5: Transitions exist

```go
func TestTransitionsExist(t *testing.T) {
    g := buildAugmentedArith()
    automaton := BuildCanonical(g)

    if len(automaton.Transitions) == 0 {
        t.Error("automaton should have transitions")
    }

    // Find transition from state 0 on NUM
    num, _ := g.Symbols.Lookup("NUM")
    found := false
    for _, tr := range automaton.Transitions {
        if tr.From == 0 && tr.Symbol == num {
            found = true
            break
        }
    }
    if !found {
        t.Error("expected transition from state 0 on NUM")
    }
}
```

### Test 6: No duplicate states

```go
func TestNoDuplicateStates(t *testing.T) {
    g := buildAugmentedArith()
    automaton := BuildCanonical(g)

    for i := 0; i < len(automaton.States); i++ {
        for j := i + 1; j < len(automaton.States); j++ {
            if automaton.States[i].Items.Equal(automaton.States[j].Items) {
                t.Errorf("states %d and %d have identical item sets",
                    automaton.States[i].ID, automaton.States[j].ID)
            }
        }
    }
}
```

## Common mistakes

1. **Not using closure after advancing the dot**: `GOTO(I, X)` is not just
   advancing the dot — you must take the closure of the advanced items.

2. **Using unstable keys for state deduplication**: The item set key must
   be deterministic. Sort items before hashing/comparing. If two item sets
   have the same items in different order, they are the same state.

3. **Forgetting to process all symbols**: For each state, compute GOTO for
   every symbol that appears after a dot. Missing one means missing a state
   or transition.

4. **Creating empty states**: If GOTO(I, X) produces an empty item set,
   don't create a state for it.

5. **Counting states wrong**: It's easy to get off-by-one. Walk through the
   arithmetic grammar by hand to verify your expected count before writing
   the test.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-11-goto-canonical.md and
implement exactly what it describes.

Scope:
- Create internal/lalr/goto.go with the Goto function.
- Create internal/lalr/state.go with the State struct.
- Create internal/lalr/collection.go with the Automaton struct and
  BuildCanonical function.
- BuildCanonical uses a worklist to discover all states, starting from
  state 0 (closure of $accept → • start).
- State deduplication uses sorted item set comparison.
- Add tests to internal/lalr/lalr_test.go: GOTO basic, GOTO empty,
  canonical state count, state 0 initial, transitions exist, no duplicates.
- Use only the Go standard library.
- Keep code simple. Use a map from item-set-key to state ID for dedup.
- Leave TODO: "// TODO(lesson-12): LALR(1) lookaheads"

Run `go test ./internal/lalr/...` and ensure all tests pass.
```

## Checkpoint questions

1. What is the difference between GOTO as a function and a transition in the
   automaton?
2. Why must GOTO take the closure of the advanced items?
3. How many transitions leave state 0 in the arithmetic grammar?
4. What happens if two different states have the same item set?
5. How do you ensure deterministic state numbering?
